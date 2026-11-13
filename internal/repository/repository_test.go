package repository

import (
	"os"
	"testing"

	"pulsefit/internal/models"
)

func TestRepositoryLifecycle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pulsefit_repo_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := tmpDir + "/test.db"
	repo, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer repo.Close()

	if repo.DB() == nil {
		t.Fatalf("expected non-nil sqlx.DB")
	}

	// 1. Создание нового пользователя
	u, err := repo.NewUser()
	if err != nil {
		t.Fatalf("NewUser failed: %v", err)
	}
	if u.Profile.ID == "" || u.Profile.Code == "" {
		t.Fatalf("invalid user profile: %+v", u.Profile)
	}

	// 2. Заполнение данных и сохранение
	age := 25
	h := 178.0
	w := 75.0
	u.Profile.Name = "Тестовый Пользователь"
	u.Profile.Age = &age
	u.Profile.HeightCm = &h
	u.Profile.WeightKg = &w
	u.Profile.Goal = models.GoalLose
	u.Profile.Pace = "moderate"
	u.Profile.Level = models.LevelBeginner
	u.Profile.WorkoutsPerWeek = 3
	u.Profile.SessionMinutes = 40
	u.Profile.Equipment = models.EquipHome
	u.Profile.Onboarded = true
	u.Profile.DietPreferences = []string{"high-protein"}
	u.Profile.Allergies = []string{"peanuts"}
	u.Profile.HealthFlags = []string{models.HealthNone}

	u.Reminders = []models.Reminder{
		{ID: "rem-1", Kind: "workout", Title: "Тренировка", Time: "19:00", Days: "daily", Enabled: true},
	}
	u.Progress.Weights = []models.WeightEntry{
		{Date: "2026-11-10", Kg: 76.0, Note: "Первое взвешивание"},
		{Date: "2026-11-13", Kg: 75.0, Note: "Прогресс"},
	}
	u.Progress.Sessions = []models.SessionEntry{
		{Date: "2026-11-12", DayIndex: 0, Title: "Понедельник", Completed: true, Feeling: "Отлично", Energy: 5},
	}
	u.Progress.Checkins = []models.Checkin{
		{Date: "2026-11-12", Mood: 5, Energy: 4, Nutrition: 5, Sleep: 8.0, Note: "Хороший день"},
	}
	u.Progress.Meals["0:oatmeal"] = true
	u.Progress.Water["2026-11-13"] = 1500

	if err := repo.Save(u); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 3. Чтение по ID
	loaded := repo.User(u.Profile.ID)
	if loaded == nil {
		t.Fatalf("User not found by ID")
	}
	if loaded.Profile.Name != "Тестовый Пользователь" {
		t.Errorf("expected name 'Тестовый Пользователь', got %s", loaded.Profile.Name)
	}
	if len(loaded.Reminders) != 1 || loaded.Reminders[0].Title != "Тренировка" {
		t.Errorf("reminders not restored properly")
	}
	if len(loaded.Progress.Weights) != 2 || loaded.Progress.Weights[1].Kg != 75.0 {
		t.Errorf("weights not restored properly")
	}
	if !loaded.Progress.Meals["0:oatmeal"] {
		t.Errorf("meals not restored properly")
	}
	if loaded.Progress.Water["2026-11-13"] != 1500 {
		t.Errorf("water logs not restored properly")
	}

	// 4. Поиск по Code
	byCode, ok := repo.ByCode(u.Profile.Code)
	if !ok || byCode.Profile.ID != u.Profile.ID {
		t.Errorf("ByCode failed to find user %s", u.Profile.Code)
	}

	// 5. Количество и удаление
	if count := repo.Count(); count < 1 {
		t.Errorf("expected at least 1 user, got %d", count)
	}

	if err := repo.Delete(u.Profile.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if deleted := repo.User(u.Profile.ID); deleted != nil {
		t.Errorf("expected user to be deleted")
	}
}
