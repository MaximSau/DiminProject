package service

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"pulsefit/internal/models"
	"pulsefit/internal/repository"
)

// Service инкапсулирует бизнес-логику приложения PulseFit.
type Service struct {
	repo    repository.Repository
	catalog *Catalog
}

// New создаёт новый экземпляр сервисного слоя.
func New(repo repository.Repository, cat *Catalog) *Service {
	return &Service{
		repo:    repo,
		catalog: cat,
	}
}

func (s *Service) Catalog() *Catalog {
	return s.catalog
}

func (s *Service) Repo() repository.Repository {
	return s.repo
}

func (s *Service) GetOrCreateUser(cookieID string) (*models.User, bool, error) {
	if cookieID != "" {
		if u := s.repo.User(cookieID); u != nil {
			return u, false, nil
		}
	}
	u, err := s.repo.NewUser()
	if err != nil {
		slog.Error("ошибка создания пользователя через сервис", "error", err)
		return nil, false, err
	}
	slog.Info("создан новый профиль пользователя", "user_id", u.Profile.ID, "code", u.Profile.Code)
	return u, true, nil
}

func (s *Service) GetUser(id string) *models.User {
	return s.repo.User(id)
}

func (s *Service) GetUserByCode(code string) (*models.User, bool) {
	return s.repo.ByCode(code)
}

func (s *Service) SaveUser(u *models.User) error {
	return s.repo.Save(u)
}

func (s *Service) DeleteUser(id string) error {
	return s.repo.Delete(id)
}

func (s *Service) UsersCount() int {
	return s.repo.Count()
}

func (s *Service) SaveOnboarding(u *models.User, p models.Profile) error {
	if u == nil {
		return errors.New("пользователь не найден")
	}

	p.ID = u.Profile.ID
	p.Code = u.Profile.Code
	p.CreatedAt = u.Profile.CreatedAt
	p.UpdatedAt = time.Now()
	p.Consent = true
	p.Onboarded = true

	u.Profile = p

	plan, err := s.catalog.BuildPlan(p)
		if err != nil {
			slog.Error("ошибка генерации плана при онбординге", "error", err, "user_id", u.Profile.ID)
			return err
		}
		u.Plan = plan
		u.Profile.PlanVersion = plan.Version

		// Базовые напоминания с учётом предпочитаемого времени тренировок
		workoutTime := "18:30"
		switch p.PreferredTime {
		case models.TimeMorning:
			workoutTime = "07:30"
		case models.TimeDay:
			workoutTime = "13:00"
		case models.TimeEvening:
			workoutTime = "18:30"
		}

		u.Reminders = []models.Reminder{
			{ID: "rem-workout", Kind: "workout", Title: "Тренировка по плану", Time: workoutTime, Days: "workouts", Enabled: true, Note: "Разминка и основная часть"},
			{ID: "rem-water", Kind: "water", Title: "Стакан чистой воды (330 мл)", Time: "11:00", Days: "daily", Enabled: true, Note: "Поддерживайте водный баланс"},
			{ID: "rem-sleep", Kind: "sleep", Title: "Подготовка ко сну", Time: "22:30", Days: "daily", Enabled: true, Note: "Отложите экран за 30 минут"},
		}

	return s.repo.Save(u)
}

func (s *Service) UpdateProfile(u *models.User, p models.Profile) error {
	if u == nil {
		return errors.New("пользователь не найден")
	}
	p.ID = u.Profile.ID
	p.Code = u.Profile.Code
	p.CreatedAt = u.Profile.CreatedAt
	p.UpdatedAt = time.Now()
	p.Consent = u.Profile.Consent
	p.Onboarded = u.Profile.Onboarded
	p.PlanVersion = u.Profile.PlanVersion

	u.Profile = p

	// Если пользователь уже прошёл онбординг и нет медицинских блоков,
	// автоматически пересоздаём план, чтобы изменения параметров (дней, цели, оборудования, питания)
	// сразу вступили в силу без ручного сброса
	if u.Profile.Onboarded && !u.Profile.MedicalBlocked() {
		plan, err := s.catalog.BuildPlan(u.Profile)
		if err == nil {
			u.Plan = plan
			u.Profile.PlanVersion = plan.Version
		} else {
			slog.Error("ошибка пересоздания плана при обновлении профиля", "error", err, "user_id", u.Profile.ID)
		}
	}

	// Синхронизируем время тренировочного напоминания с предпочтительным временем, если оно задано
	if p.PreferredTime != "" {
		newTime := "18:30"
		switch p.PreferredTime {
		case models.TimeMorning:
			newTime = "07:30"
		case models.TimeDay:
			newTime = "13:00"
		case models.TimeEvening:
			newTime = "18:30"
		}
		for i := range u.Reminders {
			if u.Reminders[i].Kind == "workout" {
				u.Reminders[i].Time = newTime
			}
		}
	}

	return s.repo.Save(u)
}

func (s *Service) RecreatePlan(u *models.User) error {
	if u == nil {
		return errors.New("пользователь не найден")
	}
	plan, err := s.catalog.BuildPlan(u.Profile)
	if err != nil {
		slog.Error("ошибка создания плана", "error", err, "user_id", u.Profile.ID)
		return err
	}

	u.Plan = plan
	u.Profile.PlanVersion = plan.Version
	return s.repo.Save(u)
}

func (s *Service) ChangeGoal(u *models.User, newGoal string) error {
	if u == nil {
		return errors.New("пользователь не найден")
	}
	if newGoal != models.GoalGain && newGoal != models.GoalLose {
		if u.Profile.Goal == models.GoalLose {
			newGoal = models.GoalGain
		} else {
			newGoal = models.GoalLose
		}
	}
	u.Profile.Goal = newGoal
	plan, err := s.catalog.BuildPlan(u.Profile)
	if err == nil {
		u.Plan = plan
		u.Profile.PlanVersion = plan.Version
	}
	return s.repo.Save(u)
}

func (s *Service) CompleteWorkout(u *models.User, dayIndex int, completed bool, feeling string, energy int) error {
	if u == nil || u.Plan == nil {
		return errors.New("план не найден")
	}
	if dayIndex < 0 || dayIndex >= len(u.Plan.Days) {
		dayIndex = models.TodayIndex()
	}

	dateStr := time.Now().Format("2006-01-02")
	title := u.Plan.Days[dayIndex].Title

	// Проверяем, есть ли уже запись за этот день недели
	found := false
	for i := range u.Progress.Sessions {
		if u.Progress.Sessions[i].DayIndex == dayIndex {
			u.Progress.Sessions[i].Completed = completed
			u.Progress.Sessions[i].Date = dateStr
			u.Progress.Sessions[i].Feeling = feeling
			u.Progress.Sessions[i].Energy = energy
			found = true
			break
		}
	}
	if !found {
		u.Progress.Sessions = append(u.Progress.Sessions, models.SessionEntry{
			Date:      dateStr,
			DayIndex:  dayIndex,
			Title:     title,
			Completed: completed,
			Feeling:   feeling,
			Energy:    energy,
		})
	}

	return s.repo.Save(u)
}

func (s *Service) RescheduleWorkout(u *models.User, fromDay, toDay int) error {
	if u == nil || u.Plan == nil {
		return errors.New("план не найден")
	}
	if fromDay < 0 || fromDay >= len(u.Plan.Days) || toDay < 0 || toDay >= len(u.Plan.Days) {
		return errors.New("неверный индекс дня")
	}
	if fromDay == toDay {
		return nil
	}

	// Swap day contents while keeping their respective Index values intact
	dayA := u.Plan.Days[fromDay]
	dayB := u.Plan.Days[toDay]

	u.Plan.Days[fromDay].Title = dayB.Title
	u.Plan.Days[fromDay].Focus = dayB.Focus
	u.Plan.Days[fromDay].Rest = dayB.Rest
	u.Plan.Days[fromDay].Minutes = dayB.Minutes
	u.Plan.Days[fromDay].Exercises = dayB.Exercises
	u.Plan.Days[fromDay].Notes = dayB.Notes

	u.Plan.Days[toDay].Title = dayA.Title
	u.Plan.Days[toDay].Focus = dayA.Focus
	u.Plan.Days[toDay].Rest = dayA.Rest
	u.Plan.Days[toDay].Minutes = dayA.Minutes
	u.Plan.Days[toDay].Exercises = dayA.Exercises
	u.Plan.Days[toDay].Notes = dayA.Notes

	return s.repo.Save(u)
}

func (s *Service) SaveWorkoutNote(u *models.User, dayIndex int, note string) error {
	if u == nil || u.Plan == nil {
		return errors.New("план не найден")
	}
	if dayIndex < 0 || dayIndex >= len(u.Plan.Days) {
		dayIndex = models.TodayIndex()
	}

	u.Plan.Days[dayIndex].Notes = note

	dateStr := time.Now().Format("2006-01-02")
	found := false
	for i := range u.Progress.Sessions {
		if u.Progress.Sessions[i].DayIndex == dayIndex {
			u.Progress.Sessions[i].Feeling = note
			found = true
			break
		}
	}
	if !found && note != "" {
		u.Progress.Sessions = append(u.Progress.Sessions, models.SessionEntry{
			Date:      dateStr,
			DayIndex:  dayIndex,
			Title:     u.Plan.Days[dayIndex].Title,
			Completed: false,
			Feeling:   note,
			Energy:    8,
		})
	}

	return s.repo.Save(u)
}

func (s *Service) SwapExercise(u *models.User, dayIndex int, currentExerciseID string) (*models.PlanExercise, error) {
	if u == nil || u.Plan == nil {
		return nil, errors.New("план не найден")
	}
	if dayIndex < 0 || dayIndex >= len(u.Plan.Days) {
		return nil, errors.New("неверный индекс дня")
	}

	day := &u.Plan.Days[dayIndex]
	targetIdx := -1
	var currentEx models.PlanExercise
	for i, ex := range day.Exercises {
		if ex.ExerciseID == currentExerciseID {
			targetIdx = i
			currentEx = ex
			break
		}
	}
	if targetIdx == -1 {
		return nil, errors.New("упражнение не найдено в тренировке")
	}

	pool := s.catalog.SuitableExercises(u.Profile.Equipment, u.Profile.Level)
	var candidates []models.Exercise
	for _, c := range pool {
		if c.ID != currentExerciseID && (c.Kind == currentEx.Kind || len(currentEx.Kind) == 0) {
			alreadyInDay := false
			for _, existing := range day.Exercises {
				if existing.ExerciseID == c.ID {
					alreadyInDay = true
					break
				}
			}
			if !alreadyInDay {
				candidates = append(candidates, c)
			}
		}
	}

	if len(candidates) == 0 {
		for _, c := range pool {
			if c.ID != currentExerciseID {
				alreadyInDay := false
				for _, existing := range day.Exercises {
					if existing.ExerciseID == c.ID {
						alreadyInDay = true
						break
					}
				}
				if !alreadyInDay {
					candidates = append(candidates, c)
				}
			}
		}
	}

	if len(candidates) == 0 {
		return nil, errors.New("подходящая замена не найдена в каталоге")
	}

	newEx := toPlanExerciseGoal(candidates[int(time.Now().UnixNano())%len(candidates)], u.Profile.Goal)
	day.Exercises[targetIdx] = newEx
	day.Minutes = estimateMinutes(day.Exercises)
	day.Focus = focusFor(day.Exercises)

	if err := s.repo.Save(u); err != nil {
		return nil, err
	}
	return &newEx, nil
}

func (s *Service) ToggleMeal(u *models.User, dayIndex int, mealID string, eaten bool) error {
	if u == nil {
		return errors.New("пользователь не найден")
	}
	if u.Progress.Meals == nil {
		u.Progress.Meals = make(map[string]bool)
	}
	key := fmt.Sprintf("%d:%s", dayIndex, mealID)
	if eaten {
		u.Progress.Meals[key] = true
	} else {
		delete(u.Progress.Meals, key)
	}
	return s.repo.Save(u)
}

func (s *Service) LogWater(u *models.User, date string, deltaMl int) (int, error) {
	if u == nil {
		return 0, errors.New("пользователь не найден")
	}
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if u.Progress.Water == nil {
		u.Progress.Water = make(map[string]int)
	}
	current := u.Progress.Water[date] + deltaMl
	if current < 0 {
		current = 0
	}
	if current > 6000 {
		current = 6000
	}
	u.Progress.Water[date] = current
	err := s.repo.Save(u)
	return current, err
}

func (s *Service) AddWeight(u *models.User, date string, kg float64, note string) error {
	if u == nil {
		return errors.New("пользователь не найден")
	}
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if kg < 30 || kg > 300 {
		return errors.New("вес должен быть в диапазоне от 30 до 300 кг")
	}

	u.Profile.WeightKg = &kg

	// Обновляем существующую запись за эту дату или добавляем новую
	found := false
	for i := range u.Progress.Weights {
		if u.Progress.Weights[i].Date == date {
			u.Progress.Weights[i].Kg = kg
			u.Progress.Weights[i].Note = note
			found = true
			break
		}
	}
	if !found {
		u.Progress.Weights = append(u.Progress.Weights, models.WeightEntry{
			Date: date,
			Kg:   kg,
			Note: note,
		})
	}

	return s.repo.Save(u)
}

func (s *Service) DeleteWeight(u *models.User, date string) error {
	if u == nil {
		return errors.New("пользователь не найден")
	}
	var filtered []models.WeightEntry
	for _, w := range u.Progress.Weights {
		if w.Date != date {
			filtered = append(filtered, w)
		}
	}
	u.Progress.Weights = filtered
	if len(filtered) > 0 {
		last := filtered[len(filtered)-1].Kg
		u.Profile.WeightKg = &last
	}
	return s.repo.Save(u)
}

func (s *Service) AddCheckin(u *models.User, date string, mood, energy, nutrition int, sleep float64, note string) error {
	if u == nil {
		return errors.New("пользователь не найден")
	}
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	found := false
	for i := range u.Progress.Checkins {
		if u.Progress.Checkins[i].Date == date {
			u.Progress.Checkins[i] = models.Checkin{
				Date:      date,
				Mood:      mood,
				Energy:    energy,
				Nutrition: nutrition,
				Sleep:     sleep,
				Note:      note,
			}
			found = true
			break
		}
	}
	if !found {
		u.Progress.Checkins = append(u.Progress.Checkins, models.Checkin{
			Date:      date,
			Mood:      mood,
			Energy:    energy,
			Nutrition: nutrition,
			Sleep:     sleep,
			Note:      note,
		})
	}

	return s.repo.Save(u)
}

func (s *Service) SaveReminders(u *models.User, reminders []models.Reminder) error {
	if u == nil {
		return errors.New("пользователь не найден")
	}
	u.Reminders = reminders
	return s.repo.Save(u)
}

func (s *Service) SeedDemoUser(u *models.User) error {
	if u == nil {
		return errors.New("пользователь не найден")
	}

	age := 28
	height := 176.0
	weight := 78.5
	u.Profile.Name = "Алексей"
	u.Profile.Age = &age
	u.Profile.HeightCm = &height
	u.Profile.WeightKg = &weight
	u.Profile.Sex = "male"
	u.Profile.Goal = models.GoalLose
	u.Profile.Pace = "moderate"
	u.Profile.Level = models.LevelBeginner
	u.Profile.WorkoutsPerWeek = 3
	u.Profile.SessionMinutes = 45
	u.Profile.Equipment = models.EquipHome
	u.Profile.PreferredTime = "evening"
	u.Profile.ActivityLevel = "medium"
	u.Profile.DietPreferences = []string{}
	u.Profile.Allergies = []string{}
	u.Profile.HealthFlags = []string{models.HealthNone}
	u.Profile.Consent = true
	u.Profile.Onboarded = true

	plan, err := s.catalog.BuildPlan(u.Profile)
	if err != nil {
		return err
	}
	u.Plan = plan
	u.Profile.PlanVersion = plan.Version

	now := time.Now()
	u.Progress.Weights = []models.WeightEntry{
		{Date: now.AddDate(0, 0, -21).Format("2006-01-02"), Kg: 80.2, Note: "Старт программы"},
		{Date: now.AddDate(0, 0, -14).Format("2006-01-02"), Kg: 79.5, Note: "Чувствую лёгкость"},
		{Date: now.AddDate(0, 0, -7).Format("2006-01-02"), Kg: 78.9, Note: "Хороший темп"},
		{Date: now.Format("2006-01-02"), Kg: 78.5, Note: "Сегодняшнее взвешивание"},
	}

	u.Progress.Sessions = []models.SessionEntry{
		{Date: now.AddDate(0, 0, -4).Format("2006-01-02"), DayIndex: 0, Title: "Понедельник", Completed: true, Feeling: "Отличная разминка, техника понятна", Energy: 5},
		{Date: now.AddDate(0, 0, -2).Format("2006-01-02"), DayIndex: 2, Title: "Среда", Completed: true, Feeling: "Мышцы слегка ноют, но самочувствие бодрое", Energy: 4},
	}

	u.Progress.Checkins = []models.Checkin{
		{Date: now.AddDate(0, 0, -1).Format("2006-01-02"), Mood: 5, Energy: 4, Nutrition: 4, Sleep: 8.0, Note: "Выпил норму воды, спал крепко"},
	}

	if u.Progress.Meals == nil {
		u.Progress.Meals = make(map[string]bool)
	}
	todayIdx := models.TodayIndex()
	if plan != nil && todayIdx < len(plan.Meals) && len(plan.Meals[todayIdx].Meals) > 0 {
		u.Progress.Meals[fmt.Sprintf("%d:%s", todayIdx, plan.Meals[todayIdx].Meals[0].MealID)] = true
	}

	if u.Progress.Water == nil {
		u.Progress.Water = make(map[string]int)
	}
	u.Progress.Water[now.Format("2006-01-02")] = 1650

	u.Reminders = []models.Reminder{
		{ID: "rem-workout", Kind: "workout", Title: "Тренировка по плану", Time: "18:30", Days: "workouts", Enabled: true, Note: "Разминка и основная часть"},
		{ID: "rem-water", Kind: "water", Title: "Стакан чистой воды (330 мл)", Time: "11:00", Days: "daily", Enabled: true, Note: "Поддерживайте водный баланс"},
		{ID: "rem-sleep", Kind: "sleep", Title: "Подготовка ко сну", Time: "22:30", Days: "daily", Enabled: true, Note: "Отложите экран за 30 минут"},
	}

	return s.repo.Save(u)
}
