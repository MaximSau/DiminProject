package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"pulsefit/internal/models"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Repository интерфейс хранилища данных.
type Repository interface {
	NewUser() (*models.User, error)
	User(id string) *models.User
	ByCode(code string) (*models.User, bool)
	Save(u *models.User) error
	Delete(id string) error
	Count() int
	Close() error
	DB() *sqlx.DB
}

// SQLiteRepository реализует Repository на базе SQLite, SQLx, Squirrel и Goose.
type SQLiteRepository struct {
	mu sync.RWMutex
	db *sqlx.DB
	sq squirrel.StatementBuilderType
}

// Open открывает SQLite базу данных, настраивает пулы, PRAGMA и накатывает миграции Goose.
func Open(path string) (*SQLiteRepository, error) {
	if strings.HasSuffix(path, ".json") {
		path = strings.TrimSuffix(path, ".json") + ".db"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("ошибка создания каталога базы: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", path)
	stdDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия SQLite: %w", err)
	}

	stdDB.SetMaxOpenConns(25)
	stdDB.SetMaxIdleConns(10)
	stdDB.SetConnMaxLifetime(time.Hour)

	// Настройка WAL-режима и тайм-аута для конкурентной записи
	_, _ = stdDB.Exec("PRAGMA journal_mode = WAL;")
	_, _ = stdDB.Exec("PRAGMA busy_timeout = 5000;")
	_, _ = stdDB.Exec("PRAGMA foreign_keys = ON;")
	_, _ = stdDB.Exec("PRAGMA synchronous = NORMAL;")

	// Автоматические миграции через Goose
	if err := RunMigrations(stdDB); err != nil {
		_ = stdDB.Close()
		return nil, fmt.Errorf("ошибка накатывания goose миграций: %w", err)
	}

	dbx := sqlx.NewDb(stdDB, "sqlite")
	repo := &SQLiteRepository{
		db: dbx,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
	}

	slog.Info("хранилище SQLite успешно инициализировано (SQLx + Squirrel + Goose)", "path", path)
	return repo, nil
}

func (r *SQLiteRepository) DB() *sqlx.DB {
	return r.db
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

func (r *SQLiteRepository) NewUser() (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, err := randomHex(16)
	if err != nil {
		return nil, err
	}

	code, err := r.generateUniqueCode()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	query, args, err := r.sq.Insert("users").
		Columns("id", "code", "created_at", "updated_at").
		Values(id, code, now, now).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("построение squirrel запроса: %w", err)
	}

	if _, err := r.db.Exec(query, args...); err != nil {
		slog.Error("ошибка создания пользователя в БД", "error", err, "id", id)
		return nil, err
	}

	u := &models.User{}
	u.Profile.ID = id
	u.Profile.Code = code
	u.Profile.CreatedAt = now
	u.Profile.UpdatedAt = now
	u.Progress.Meals = make(map[string]bool)
	u.Progress.Water = make(map[string]int)

	slog.Debug("создан новый пользователь", "id", id, "code", code)
	return u, nil
}

func (r *SQLiteRepository) User(id string) *models.User {
	if id == "" {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. users
	qUser, argsUser, err := r.sq.Select("code", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		slog.Error("squirrel select user query error", "error", err)
		return nil
	}

	var code string
	var createdAt, updatedAt time.Time
	if err := r.db.QueryRow(qUser, argsUser...).Scan(&code, &createdAt, &updatedAt); err != nil {
		return nil
	}

	u := &models.User{}
	u.Profile.ID = id
	u.Profile.Code = code
	u.Profile.CreatedAt = createdAt
	u.Profile.UpdatedAt = updatedAt
	u.Progress.Meals = make(map[string]bool)
	u.Progress.Water = make(map[string]int)

	// 2. profiles
	qProf, argsProf, err := r.sq.Select(
		"name", "age", "height_cm", "weight_kg", "sex", "goal", "pace", "level",
		"workouts_per_week", "session_minutes", "equipment", "preferred_time",
		"activity_level", "diet_preferences", "allergies", "avoid_foods",
		"health_flags", "health_notes", "consent", "onboarded", "plan_version",
	).From("profiles").Where(squirrel.Eq{"user_id": id}).ToSql()

	if err == nil {
		var (
			name, sex, goal, pace, level, equip, prefTime, actLevel, avoidFoods, healthNotes string
			dietJSON, allergJSON, healthFlagsJSON                                            sql.NullString
			age                                                                              sql.NullInt64
			heightCm, weightKg                                                               sql.NullFloat64
			wpw, smin, planVer, consent, onboarded                                           int
		)
		row := r.db.QueryRow(qProf, argsProf...)
		if err := row.Scan(
			&name, &age, &heightCm, &weightKg, &sex, &goal, &pace, &level,
			&wpw, &smin, &equip, &prefTime, &actLevel, &dietJSON, &allergJSON, &avoidFoods,
			&healthFlagsJSON, &healthNotes, &consent, &onboarded, &planVer,
		); err == nil {
			u.Profile.Name = name
			if age.Valid {
				a := int(age.Int64)
				u.Profile.Age = &a
			}
			if heightCm.Valid {
				h := heightCm.Float64
				u.Profile.HeightCm = &h
			}
			if weightKg.Valid {
				w := weightKg.Float64
				u.Profile.WeightKg = &w
			}
			u.Profile.Sex = sex
			u.Profile.Goal = goal
			u.Profile.Pace = pace
			u.Profile.Level = level
			u.Profile.WorkoutsPerWeek = wpw
			u.Profile.SessionMinutes = smin
			u.Profile.Equipment = equip
			u.Profile.PreferredTime = prefTime
			u.Profile.ActivityLevel = actLevel
			u.Profile.AvoidFoods = avoidFoods
			u.Profile.HealthNotes = healthNotes
			u.Profile.Consent = consent == 1
			u.Profile.Onboarded = onboarded == 1
			u.Profile.PlanVersion = planVer

			if dietJSON.Valid && dietJSON.String != "" {
				_ = json.Unmarshal([]byte(dietJSON.String), &u.Profile.DietPreferences)
			}
			if allergJSON.Valid && allergJSON.String != "" {
				_ = json.Unmarshal([]byte(allergJSON.String), &u.Profile.Allergies)
			}
			if healthFlagsJSON.Valid && healthFlagsJSON.String != "" {
				_ = json.Unmarshal([]byte(healthFlagsJSON.String), &u.Profile.HealthFlags)
			}
		}
	}

	// 3. plans
	qPlan, argsPlan, err := r.sq.Select("plan_json").From("plans").Where(squirrel.Eq{"user_id": id}).ToSql()
	if err == nil {
		var planJSON string
		if err := r.db.QueryRow(qPlan, argsPlan...).Scan(&planJSON); err == nil && planJSON != "" {
			var pl models.Plan
			if err := json.Unmarshal([]byte(planJSON), &pl); err == nil {
				u.Plan = &pl
			}
		}
	}

	// 4. reminders
	qRem, argsRem, err := r.sq.Select("id", "kind", "title", "time", "days", "enabled", "note").
		From("reminders").
		Where(squirrel.Eq{"user_id": id}).
		OrderBy("id ASC").
		ToSql()
	if err == nil {
		rRows, err := r.db.Query(qRem, argsRem...)
		if err == nil {
			defer rRows.Close()
			for rRows.Next() {
				var rem models.Reminder
				var en int
				if err := rRows.Scan(&rem.ID, &rem.Kind, &rem.Title, &rem.Time, &rem.Days, &en, &rem.Note); err == nil {
					rem.Enabled = en == 1
					u.Reminders = append(u.Reminders, rem)
				}
			}
		}
	}

	// 5. weights
	qWeight, argsWeight, err := r.sq.Select("date", "kg", "note").
		From("weights").
		Where(squirrel.Eq{"user_id": id}).
		OrderBy("date ASC", "id ASC").
		ToSql()
	if err == nil {
		wRows, err := r.db.Query(qWeight, argsWeight...)
		if err == nil {
			defer wRows.Close()
			for wRows.Next() {
				var w models.WeightEntry
				if err := wRows.Scan(&w.Date, &w.Kg, &w.Note); err == nil {
					u.Progress.Weights = append(u.Progress.Weights, w)
				}
			}
		}
	}

	// 6. sessions
	qSess, argsSess, err := r.sq.Select("date", "day_index", "title", "completed", "feeling", "energy").
		From("sessions").
		Where(squirrel.Eq{"user_id": id}).
		OrderBy("id ASC").
		ToSql()
	if err == nil {
		sRows, err := r.db.Query(qSess, argsSess...)
		if err == nil {
			defer sRows.Close()
			for sRows.Next() {
				var se models.SessionEntry
				var comp int
				if err := sRows.Scan(&se.Date, &se.DayIndex, &se.Title, &comp, &se.Feeling, &se.Energy); err == nil {
					se.Completed = comp == 1
					u.Progress.Sessions = append(u.Progress.Sessions, se)
				}
			}
		}
	}

	// 7. checkins
	qCheck, argsCheck, err := r.sq.Select("date", "mood", "energy", "nutrition", "sleep", "note").
		From("checkins").
		Where(squirrel.Eq{"user_id": id}).
		OrderBy("id ASC").
		ToSql()
	if err == nil {
		cRows, err := r.db.Query(qCheck, argsCheck...)
		if err == nil {
			defer cRows.Close()
			for cRows.Next() {
				var ch models.Checkin
				if err := cRows.Scan(&ch.Date, &ch.Mood, &ch.Energy, &ch.Nutrition, &ch.Sleep, &ch.Note); err == nil {
					u.Progress.Checkins = append(u.Progress.Checkins, ch)
				}
			}
		}
	}

	// 8. eaten_meals
	qMeal, argsMeal, err := r.sq.Select("meal_key").From("eaten_meals").Where(squirrel.Eq{"user_id": id}).ToSql()
	if err == nil {
		mRows, err := r.db.Query(qMeal, argsMeal...)
		if err == nil {
			defer mRows.Close()
			for mRows.Next() {
				var k string
				if err := mRows.Scan(&k); err == nil {
					u.Progress.Meals[k] = true
				}
			}
		}
	}

	// 9. water_logs
	qWater, argsWater, err := r.sq.Select("date", "ml").From("water_logs").Where(squirrel.Eq{"user_id": id}).ToSql()
	if err == nil {
		wtRows, err := r.db.Query(qWater, argsWater...)
		if err == nil {
			defer wtRows.Close()
			for wtRows.Next() {
				var dt string
				var ml int
				if err := wtRows.Scan(&dt, &ml); err == nil {
					u.Progress.Water[dt] = ml
				}
			}
		}
	}

	// 10. chat_messages
	qChat, argsChat, err := r.sq.Select("id", "role", "text", "at", "kind", "suggestions").
		From("chat_messages").
		Where(squirrel.Eq{"user_id": id}).
		OrderBy("at ASC").
		ToSql()
	if err == nil {
		chRows, err := r.db.Query(qChat, argsChat...)
		if err == nil {
			defer chRows.Close()
			for chRows.Next() {
				var cm models.ChatMessage
				var suggJSON sql.NullString
				if err := chRows.Scan(&cm.ID, &cm.Role, &cm.Text, &cm.At, &cm.Kind, &suggJSON); err == nil {
					if suggJSON.Valid && suggJSON.String != "" {
						_ = json.Unmarshal([]byte(suggJSON.String), &cm.Suggestions)
					}
					u.Chat = append(u.Chat, cm)
				}
			}
		}
	}

	return u
}

func (r *SQLiteRepository) ByCode(code string) (*models.User, bool) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, false
	}
	r.mu.RLock()
	query, args, err := r.sq.Select("id").From("users").Where(squirrel.Eq{"code": code}).ToSql()
	if err != nil {
		r.mu.RUnlock()
		return nil, false
	}
	var id string
	err = r.db.QueryRow(query, args...).Scan(&id)
	r.mu.RUnlock()
	if err != nil {
		return nil, false
	}
	u := r.User(id)
	return u, u != nil
}

func (r *SQLiteRepository) Save(u *models.User) error {
	if u == nil || u.Profile.ID == "" {
		return errors.New("пустой пользователь")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("ошибка открытия транзакции: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now()
	u.Profile.UpdatedAt = now

	// 1. users (upsert)
	usersUpsert := `
		INSERT INTO users (id, code, created_at, updated_at) VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET updated_at = excluded.updated_at`
	if _, err = tx.Exec(usersUpsert, u.Profile.ID, u.Profile.Code, u.Profile.CreatedAt, now); err != nil {
		slog.Error("ошибка сохранения user", "error", err, "id", u.Profile.ID)
		return err
	}

	// 2. profiles
	dietJSON, _ := json.Marshal(u.Profile.DietPreferences)
	allergJSON, _ := json.Marshal(u.Profile.Allergies)
	healthFlagsJSON, _ := json.Marshal(u.Profile.HealthFlags)

	var ageVal any
	if u.Profile.Age != nil {
		ageVal = *u.Profile.Age
	}
	var hVal any
	if u.Profile.HeightCm != nil {
		hVal = *u.Profile.HeightCm
	}
	var wVal any
	if u.Profile.WeightKg != nil {
		wVal = *u.Profile.WeightKg
	}

	consentInt := 0
	if u.Profile.Consent {
		consentInt = 1
	}
	onboardedInt := 0
	if u.Profile.Onboarded {
		onboardedInt = 1
	}

	profUpsert := `
		INSERT INTO profiles (
			user_id, name, age, height_cm, weight_kg, sex, goal, pace, level,
			workouts_per_week, session_minutes, equipment, preferred_time, activity_level,
			diet_preferences, allergies, avoid_foods, health_flags, health_notes,
			consent, onboarded, plan_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			name = excluded.name, age = excluded.age, height_cm = excluded.height_cm,
			weight_kg = excluded.weight_kg, sex = excluded.sex, goal = excluded.goal,
			pace = excluded.pace, level = excluded.level, workouts_per_week = excluded.workouts_per_week,
			session_minutes = excluded.session_minutes, equipment = excluded.equipment,
			preferred_time = excluded.preferred_time, activity_level = excluded.activity_level,
			diet_preferences = excluded.diet_preferences, allergies = excluded.allergies,
			avoid_foods = excluded.avoid_foods, health_flags = excluded.health_flags,
			health_notes = excluded.health_notes, consent = excluded.consent,
			onboarded = excluded.onboarded, plan_version = excluded.plan_version`

	if _, err = tx.Exec(
		profUpsert,
		u.Profile.ID, u.Profile.Name, ageVal, hVal, wVal, u.Profile.Sex, u.Profile.Goal, u.Profile.Pace, u.Profile.Level,
		u.Profile.WorkoutsPerWeek, u.Profile.SessionMinutes, u.Profile.Equipment, u.Profile.PreferredTime, u.Profile.ActivityLevel,
		string(dietJSON), string(allergJSON), u.Profile.AvoidFoods, string(healthFlagsJSON), u.Profile.HealthNotes,
		consentInt, onboardedInt, u.Profile.PlanVersion,
	); err != nil {
		slog.Error("ошибка сохранения profile", "error", err, "id", u.Profile.ID)
		return err
	}

	// 3. plans
	if u.Plan != nil {
		planRaw, _ := json.Marshal(u.Plan)
		planUpsert := `
			INSERT INTO plans (user_id, plan_json) VALUES (?, ?)
			ON CONFLICT(user_id) DO UPDATE SET plan_json = excluded.plan_json`
		if _, err = tx.Exec(planUpsert, u.Profile.ID, string(planRaw)); err != nil {
			slog.Error("ошибка сохранения plan", "error", err, "id", u.Profile.ID)
			return err
		}
	}

	// 4. reminders
	delRemQuery, delRemArgs, _ := r.sq.Delete("reminders").Where(squirrel.Eq{"user_id": u.Profile.ID}).ToSql()
	if _, err = tx.Exec(delRemQuery, delRemArgs...); err != nil {
		return err
	}
	for _, rem := range u.Reminders {
		en := 0
		if rem.Enabled {
			en = 1
		}
		insRem, insArgs, err := r.sq.Insert("reminders").
			Columns("id", "user_id", "kind", "title", "time", "days", "enabled", "note").
			Values(rem.ID, u.Profile.ID, rem.Kind, rem.Title, rem.Time, rem.Days, en, rem.Note).
			ToSql()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(insRem, insArgs...); err != nil {
			return err
		}
	}

	// 5. weights
	delWQuery, delWArgs, _ := r.sq.Delete("weights").Where(squirrel.Eq{"user_id": u.Profile.ID}).ToSql()
	if _, err = tx.Exec(delWQuery, delWArgs...); err != nil {
		return err
	}
	for _, w := range u.Progress.Weights {
		insW, insArgs, err := r.sq.Insert("weights").
			Columns("user_id", "date", "kg", "note").
			Values(u.Profile.ID, w.Date, w.Kg, w.Note).
			ToSql()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(insW, insArgs...); err != nil {
			return err
		}
	}

	// 6. sessions
	delSessQuery, delSessArgs, _ := r.sq.Delete("sessions").Where(squirrel.Eq{"user_id": u.Profile.ID}).ToSql()
	if _, err = tx.Exec(delSessQuery, delSessArgs...); err != nil {
		return err
	}
	for _, s := range u.Progress.Sessions {
		comp := 0
		if s.Completed {
			comp = 1
		}
		insSess, insArgs, err := r.sq.Insert("sessions").
			Columns("user_id", "date", "day_index", "title", "completed", "feeling", "energy").
			Values(u.Profile.ID, s.Date, s.DayIndex, s.Title, comp, s.Feeling, s.Energy).
			ToSql()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(insSess, insArgs...); err != nil {
			return err
		}
	}

	// 7. checkins
	delCheckQuery, delCheckArgs, _ := r.sq.Delete("checkins").Where(squirrel.Eq{"user_id": u.Profile.ID}).ToSql()
	if _, err = tx.Exec(delCheckQuery, delCheckArgs...); err != nil {
		return err
	}
	for _, c := range u.Progress.Checkins {
		insCheck, insArgs, err := r.sq.Insert("checkins").
			Columns("user_id", "date", "mood", "energy", "nutrition", "sleep", "note").
			Values(u.Profile.ID, c.Date, c.Mood, c.Energy, c.Nutrition, c.Sleep, c.Note).
			ToSql()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(insCheck, insArgs...); err != nil {
			return err
		}
	}

	// 8. eaten_meals
	delMealQuery, delMealArgs, _ := r.sq.Delete("eaten_meals").Where(squirrel.Eq{"user_id": u.Profile.ID}).ToSql()
	if _, err = tx.Exec(delMealQuery, delMealArgs...); err != nil {
		return err
	}
	for k, eaten := range u.Progress.Meals {
		if eaten {
			insMeal, insArgs, err := r.sq.Insert("eaten_meals").
				Columns("user_id", "meal_key").
				Values(u.Profile.ID, k).
				ToSql()
			if err != nil {
				return err
			}
			if _, err = tx.Exec(insMeal, insArgs...); err != nil {
				return err
			}
		}
	}

	// 9. water_logs
	delWaterQuery, delWaterArgs, _ := r.sq.Delete("water_logs").Where(squirrel.Eq{"user_id": u.Profile.ID}).ToSql()
	if _, err = tx.Exec(delWaterQuery, delWaterArgs...); err != nil {
		return err
	}
	for dt, ml := range u.Progress.Water {
		if ml > 0 {
			insWater, insArgs, err := r.sq.Insert("water_logs").
				Columns("user_id", "date", "ml").
				Values(u.Profile.ID, dt, ml).
				ToSql()
			if err != nil {
				return err
			}
			if _, err = tx.Exec(insWater, insArgs...); err != nil {
				return err
			}
		}
	}

	// 10. chat_messages
	delChatQuery, delChatArgs, _ := r.sq.Delete("chat_messages").Where(squirrel.Eq{"user_id": u.Profile.ID}).ToSql()
	if _, err = tx.Exec(delChatQuery, delChatArgs...); err != nil {
		return err
	}
	for _, cm := range u.Chat {
		suggJSON, _ := json.Marshal(cm.Suggestions)
		insChat, insArgs, err := r.sq.Insert("chat_messages").
			Columns("id", "user_id", "role", "text", "at", "kind", "suggestions").
			Values(cm.ID, u.Profile.ID, cm.Role, cm.Text, cm.At, cm.Kind, string(suggJSON)).
			ToSql()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(insChat, insArgs...); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		slog.Error("ошибка коммита транзакции сохранения пользователя", "error", err, "id", u.Profile.ID)
		return err
	}

	slog.Debug("пользователь успешно сохранен", "id", u.Profile.ID)
	return nil
}

func (r *SQLiteRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query, args, err := r.sq.Delete("users").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(query, args...)
	if err != nil {
		slog.Error("ошибка удаления пользователя", "error", err, "id", id)
	} else {
		slog.Info("пользователь успешно удален из БД", "id", id)
	}
	return err
}

func (r *SQLiteRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query, args, err := r.sq.Select("count(*)").From("users").ToSql()
	if err != nil {
		return 0
	}
	var count int
	_ = r.db.QueryRow(query, args...).Scan(&count)
	return count
}

func (r *SQLiteRepository) generateUniqueCode() (string, error) {
	for i := 0; i < 30; i++ {
		v, err := randomHex(4)
		if err != nil {
			return "", err
		}
		code := strings.ToUpper(v)
		query, args, err := r.sq.Select("1").From("users").Where(squirrel.Eq{"code": code}).ToSql()
		if err != nil {
			return "", err
		}
		var exists int
		err = r.db.QueryRow(query, args...).Scan(&exists)
		if err == sql.ErrNoRows {
			return code, nil
		}
	}
	return "", errors.New("не удалось сгенерировать уникальный код профиля")
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
