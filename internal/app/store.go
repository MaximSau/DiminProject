package app

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// User — всё, что приложение хранит про одного человека.
type User struct {
	Profile   Profile       `json:"profile"`
	Plan      *Plan         `json:"plan,omitempty"`
	Reminders []Reminder    `json:"reminders,omitempty"`
	Progress  Progress      `json:"progress"`
	Chat      []ChatMessage `json:"chat,omitempty"`
}

// WaterToday возвращает количество выпитой воды сегодня в миллилитрах.
func (u *User) WaterToday() int {
	if u.Progress.Water == nil {
		return 0
	}
	today := time.Now().Format("2006-01-02")
	return u.Progress.Water[today]
}

// Reminder — настраиваемое напоминание.
type Reminder struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"` // workout | meal | water | sleep
	Title   string `json:"title"`
	Time    string `json:"time"` // HH:MM
	Days    string `json:"days"` // daily | workouts
	Enabled bool   `json:"enabled"`
	Note    string `json:"note,omitempty"`
}

type WeightEntry struct {
	Date string  `json:"date"`
	Kg   float64 `json:"kg"`
	Note string  `json:"note,omitempty"`
}

type SessionEntry struct {
	Date      string `json:"date"`
	DayIndex  int    `json:"day_index"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Feeling   string `json:"feeling,omitempty"`
	Energy    int    `json:"energy,omitempty"`
}

// Checkin — самочувствие, энергия и регулярность питания за день.
type Checkin struct {
	Date      string  `json:"date"`
	Mood      int     `json:"mood"`
	Energy    int     `json:"energy"`
	Nutrition int     `json:"nutrition"`
	Sleep     float64 `json:"sleep,omitempty"`
	Note      string  `json:"note,omitempty"`
}

type Progress struct {
	Weights  []WeightEntry   `json:"weights,omitempty"`
	Sessions []SessionEntry  `json:"sessions,omitempty"`
	Checkins []Checkin       `json:"checkins,omitempty"`
	Meals    map[string]bool `json:"meals,omitempty"`
	Water    map[string]int  `json:"water,omitempty"` // date -> ml
}

// ChatMessage — сообщение в чате с помощником.
type ChatMessage struct {
	ID          string    `json:"id"`
	Role        string    `json:"role"` // user | assistant
	Text        string    `json:"text"`
	At          time.Time `json:"at"`
	Kind        string    `json:"kind,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`
}

// Store — хранилище на базе реляционной СУБД SQLite с поддержкой транзакций.
type Store struct {
	mu sync.RWMutex
	db *sql.DB
}

// OpenStore открывает (или создаёт) базу данных SQLite и накатывает схему таблиц.
func OpenStore(path string) (*Store, error) {
	// Если путь был указан как store.json, переводим в store.db
	if strings.HasSuffix(path, ".json") {
		path = strings.TrimSuffix(path, ".json") + ".db"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия SQLite базы: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	// Форсируем PRAGMA для надёжной работы WAL и отсутствия блокировок при конкурентной записи
	_, _ = db.Exec("PRAGMA journal_mode = WAL;")
	_, _ = db.Exec("PRAGMA busy_timeout = 5000;")
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")
	_, _ = db.Exec("PRAGMA synchronous = NORMAL;")

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ошибка инициализации схемы SQLite: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		code TEXT UNIQUE NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS profiles (
		user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		name TEXT,
		age INTEGER,
		height_cm REAL,
		weight_kg REAL,
		sex TEXT,
		goal TEXT,
		pace TEXT,
		level TEXT,
		workouts_per_week INTEGER,
		session_minutes INTEGER,
		equipment TEXT,
		preferred_time TEXT,
		activity_level TEXT,
		diet_preferences TEXT,
		allergies TEXT,
		avoid_foods TEXT,
		health_flags TEXT,
		health_notes TEXT,
		consent INTEGER,
		onboarded INTEGER,
		plan_version INTEGER
	);

	CREATE TABLE IF NOT EXISTS plans (
		user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		plan_json TEXT
	);

	CREATE TABLE IF NOT EXISTS reminders (
		id TEXT NOT NULL,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		kind TEXT,
		title TEXT,
		time TEXT,
		days TEXT,
		enabled INTEGER,
		note TEXT,
		PRIMARY KEY(user_id, id)
	);

	CREATE TABLE IF NOT EXISTS weights (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		date TEXT NOT NULL,
		kg REAL NOT NULL,
		note TEXT
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		date TEXT NOT NULL,
		day_index INTEGER NOT NULL,
		title TEXT NOT NULL,
		completed INTEGER NOT NULL,
		feeling TEXT,
		energy INTEGER
	);

	CREATE TABLE IF NOT EXISTS checkins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		date TEXT NOT NULL,
		mood INTEGER,
		energy INTEGER,
		nutrition INTEGER,
		sleep REAL,
		note TEXT
	);

	CREATE TABLE IF NOT EXISTS eaten_meals (
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		meal_key TEXT NOT NULL,
		PRIMARY KEY(user_id, meal_key)
	);

	CREATE TABLE IF NOT EXISTS water_logs (
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		date TEXT NOT NULL,
		ml INTEGER NOT NULL,
		PRIMARY KEY(user_id, date)
	);

	CREATE TABLE IF NOT EXISTS chat_messages (
		id TEXT NOT NULL,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role TEXT NOT NULL,
		text TEXT NOT NULL,
		at DATETIME NOT NULL,
		kind TEXT,
		suggestions TEXT,
		PRIMARY KEY(user_id, id)
	);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *Store) NewUser() (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, err := randomHex(16)
	if err != nil {
		return nil, err
	}

	code, err := s.generateUniqueCode()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	_, err = s.db.Exec(`INSERT INTO users (id, code, created_at, updated_at) VALUES (?, ?, ?, ?)`, id, code, now, now)
	if err != nil {
		return nil, err
	}

	u := &User{}
	u.Profile.ID = id
	u.Profile.Code = code
	u.Profile.CreatedAt = now
	u.Profile.UpdatedAt = now
	u.Progress.Meals = make(map[string]bool)
	u.Progress.Water = make(map[string]int)

	return u, nil
}

func (s *Store) User(id string) *User {
	if id == "" {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	var code string
	var createdAt, updatedAt time.Time
	err := s.db.QueryRow(`SELECT code, created_at, updated_at FROM users WHERE id = ?`, id).Scan(&code, &createdAt, &updatedAt)
	if err != nil {
		return nil
	}

	u := &User{}
	u.Profile.ID = id
	u.Profile.Code = code
	u.Profile.CreatedAt = createdAt
	u.Profile.UpdatedAt = updatedAt
	u.Progress.Meals = make(map[string]bool)
	u.Progress.Water = make(map[string]int)

	// Читаем Profile
	var (
		name, sex, goal, pace, level, equip, prefTime, actLevel, avoidFoods, healthNotes string
		dietJSON, allergJSON, healthFlagsJSON                                            sql.NullString
		age                                                                              sql.NullInt64
		heightCm, weightKg                                                               sql.NullFloat64
		wpw, smin, planVer, consent, onboarded                                           int
	)
	row := s.db.QueryRow(`
		SELECT name, age, height_cm, weight_kg, sex, goal, pace, level, workouts_per_week, session_minutes,
		       equipment, preferred_time, activity_level, diet_preferences, allergies, avoid_foods,
		       health_flags, health_notes, consent, onboarded, plan_version
		FROM profiles WHERE user_id = ?`, id)

	if err := row.Scan(&name, &age, &heightCm, &weightKg, &sex, &goal, &pace, &level, &wpw, &smin,
		&equip, &prefTime, &actLevel, &dietJSON, &allergJSON, &avoidFoods,
		&healthFlagsJSON, &healthNotes, &consent, &onboarded, &planVer); err == nil {

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

	// Читаем Plan
	var planJSON string
	if err := s.db.QueryRow(`SELECT plan_json FROM plans WHERE user_id = ?`, id).Scan(&planJSON); err == nil && planJSON != "" {
		var pl Plan
		if err := json.Unmarshal([]byte(planJSON), &pl); err == nil {
			u.Plan = &pl
		}
	}

	// Читаем Reminders
	rRows, err := s.db.Query(`SELECT id, kind, title, time, days, enabled, note FROM reminders WHERE user_id = ? ORDER BY id`, id)
	if err == nil {
		defer rRows.Close()
		for rRows.Next() {
			var rem Reminder
			var en int
			if err := rRows.Scan(&rem.ID, &rem.Kind, &rem.Title, &rem.Time, &rem.Days, &en, &rem.Note); err == nil {
				rem.Enabled = en == 1
				u.Reminders = append(u.Reminders, rem)
			}
		}
	}

	// Читаем Weights
	wRows, err := s.db.Query(`SELECT date, kg, note FROM weights WHERE user_id = ? ORDER BY date ASC, id ASC`, id)
	if err == nil {
		defer wRows.Close()
		for wRows.Next() {
			var w WeightEntry
			if err := wRows.Scan(&w.Date, &w.Kg, &w.Note); err == nil {
				u.Progress.Weights = append(u.Progress.Weights, w)
			}
		}
	}

	// Читаем Sessions
	sRows, err := s.db.Query(`SELECT date, day_index, title, completed, feeling, energy FROM sessions WHERE user_id = ? ORDER BY id ASC`, id)
	if err == nil {
		defer sRows.Close()
		for sRows.Next() {
			var se SessionEntry
			var comp int
			if err := sRows.Scan(&se.Date, &se.DayIndex, &se.Title, &comp, &se.Feeling, &se.Energy); err == nil {
				se.Completed = comp == 1
				u.Progress.Sessions = append(u.Progress.Sessions, se)
			}
		}
	}

	// Читаем Checkins
	cRows, err := s.db.Query(`SELECT date, mood, energy, nutrition, sleep, note FROM checkins WHERE user_id = ? ORDER BY id ASC`, id)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var ch Checkin
			if err := cRows.Scan(&ch.Date, &ch.Mood, &ch.Energy, &ch.Nutrition, &ch.Sleep, &ch.Note); err == nil {
				u.Progress.Checkins = append(u.Progress.Checkins, ch)
			}
		}
	}

	// Читаем Eaten Meals
	mRows, err := s.db.Query(`SELECT meal_key FROM eaten_meals WHERE user_id = ?`, id)
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var k string
			if err := mRows.Scan(&k); err == nil {
				u.Progress.Meals[k] = true
			}
		}
	}

	// Читаем Water Logs
	wtRows, err := s.db.Query(`SELECT date, ml FROM water_logs WHERE user_id = ?`, id)
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

	// Читаем Chat Messages
	chRows, err := s.db.Query(`SELECT id, role, text, at, kind, suggestions FROM chat_messages WHERE user_id = ? ORDER BY at ASC`, id)
	if err == nil {
		defer chRows.Close()
		for chRows.Next() {
			var cm ChatMessage
			var suggJSON sql.NullString
			if err := chRows.Scan(&cm.ID, &cm.Role, &cm.Text, &cm.At, &cm.Kind, &suggJSON); err == nil {
				if suggJSON.Valid && suggJSON.String != "" {
					_ = json.Unmarshal([]byte(suggJSON.String), &cm.Suggestions)
				}
				u.Chat = append(u.Chat, cm)
			}
		}
	}

	return u
}

func (s *Store) ByCode(code string) (*User, bool) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, false
	}
	s.mu.RLock()
	var id string
	err := s.db.QueryRow(`SELECT id FROM users WHERE code = ?`, code).Scan(&id)
	s.mu.RUnlock()
	if err != nil {
		return nil, false
	}
	u := s.User(id)
	return u, u != nil
}

func (s *Store) Save(u *User) error {
	if u == nil || u.Profile.ID == "" {
		return errors.New("пустой пользователь")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now()
	u.Profile.UpdatedAt = now

	// 1. users
	_, err = tx.Exec(`
		INSERT INTO users (id, code, created_at, updated_at) VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET updated_at = ?`,
		u.Profile.ID, u.Profile.Code, u.Profile.CreatedAt, now, now)
	if err != nil {
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

	_, err = tx.Exec(`
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
			onboarded = excluded.onboarded, plan_version = excluded.plan_version`,
		u.Profile.ID, u.Profile.Name, ageVal, hVal, wVal, u.Profile.Sex, u.Profile.Goal, u.Profile.Pace, u.Profile.Level,
		u.Profile.WorkoutsPerWeek, u.Profile.SessionMinutes, u.Profile.Equipment, u.Profile.PreferredTime, u.Profile.ActivityLevel,
		string(dietJSON), string(allergJSON), u.Profile.AvoidFoods, string(healthFlagsJSON), u.Profile.HealthNotes,
		consentInt, onboardedInt, u.Profile.PlanVersion)
	if err != nil {
		return err
	}

	// 3. plans
	if u.Plan != nil {
		planRaw, _ := json.Marshal(u.Plan)
		_, err = tx.Exec(`
			INSERT INTO plans (user_id, plan_json) VALUES (?, ?)
			ON CONFLICT(user_id) DO UPDATE SET plan_json = excluded.plan_json`,
			u.Profile.ID, string(planRaw))
		if err != nil {
			return err
		}
	}

	// 4. reminders
	_, _ = tx.Exec(`DELETE FROM reminders WHERE user_id = ?`, u.Profile.ID)
	for _, r := range u.Reminders {
		en := 0
		if r.Enabled {
			en = 1
		}
		_, err = tx.Exec(`
			INSERT INTO reminders (id, user_id, kind, title, time, days, enabled, note)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			r.ID, u.Profile.ID, r.Kind, r.Title, r.Time, r.Days, en, r.Note)
		if err != nil {
			return err
		}
	}

	// 5. weights
	_, _ = tx.Exec(`DELETE FROM weights WHERE user_id = ?`, u.Profile.ID)
	for _, w := range u.Progress.Weights {
		_, err = tx.Exec(`INSERT INTO weights (user_id, date, kg, note) VALUES (?, ?, ?, ?)`,
			u.Profile.ID, w.Date, w.Kg, w.Note)
		if err != nil {
			return err
		}
	}

	// 6. sessions
	_, _ = tx.Exec(`DELETE FROM sessions WHERE user_id = ?`, u.Profile.ID)
	for _, s := range u.Progress.Sessions {
		comp := 0
		if s.Completed {
			comp = 1
		}
		_, err = tx.Exec(`INSERT INTO sessions (user_id, date, day_index, title, completed, feeling, energy)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			u.Profile.ID, s.Date, s.DayIndex, s.Title, comp, s.Feeling, s.Energy)
		if err != nil {
			return err
		}
	}

	// 7. checkins
	_, _ = tx.Exec(`DELETE FROM checkins WHERE user_id = ?`, u.Profile.ID)
	for _, c := range u.Progress.Checkins {
		_, err = tx.Exec(`INSERT INTO checkins (user_id, date, mood, energy, nutrition, sleep, note)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			u.Profile.ID, c.Date, c.Mood, c.Energy, c.Nutrition, c.Sleep, c.Note)
		if err != nil {
			return err
		}
	}

	// 8. eaten_meals
	_, _ = tx.Exec(`DELETE FROM eaten_meals WHERE user_id = ?`, u.Profile.ID)
	for k, eaten := range u.Progress.Meals {
		if eaten {
			_, err = tx.Exec(`INSERT INTO eaten_meals (user_id, meal_key) VALUES (?, ?)`, u.Profile.ID, k)
			if err != nil {
				return err
			}
		}
	}

	// 9. water_logs
	_, _ = tx.Exec(`DELETE FROM water_logs WHERE user_id = ?`, u.Profile.ID)
	for dt, ml := range u.Progress.Water {
		if ml > 0 {
			_, err = tx.Exec(`INSERT INTO water_logs (user_id, date, ml) VALUES (?, ?, ?)`, u.Profile.ID, dt, ml)
			if err != nil {
				return err
			}
		}
	}

	// 10. chat_messages
	_, _ = tx.Exec(`DELETE FROM chat_messages WHERE user_id = ?`, u.Profile.ID)
	for _, cm := range u.Chat {
		suggJSON, _ := json.Marshal(cm.Suggestions)
		_, err = tx.Exec(`INSERT INTO chat_messages (id, user_id, role, text, at, kind, suggestions)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			cm.ID, u.Profile.ID, cm.Role, cm.Text, cm.At, cm.Kind, string(suggJSON))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var count int
	_ = s.db.QueryRow(`SELECT count(*) FROM users`).Scan(&count)
	return count
}

func (s *Store) generateUniqueCode() (string, error) {
	for i := 0; i < 30; i++ {
		v, err := randomHex(4)
		if err != nil {
			return "", err
		}
		code := strings.ToUpper(v)
		var exists int
		err = s.db.QueryRow(`SELECT 1 FROM users WHERE code = ?`, code).Scan(&exists)
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
