package service

import (
	"os"
	"testing"

	"pulsefit/internal/models"
	"pulsefit/internal/repository"
)

func TestServiceFullFlow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pulsefit_service_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := tmpDir + "/test_service.db"
	repo, err := repository.Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer repo.Close()

	cat, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}

	svc := New(repo, cat)

	// 1. Создание пользователя
	u, isNew, err := svc.GetOrCreateUser("")
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}
	if !isNew || u == nil {
		t.Fatalf("expected new user")
	}

	// 2. Демо сценарий
	if err := svc.SeedDemoUser(u); err != nil {
		t.Fatalf("SeedDemoUser failed: %v", err)
	}
	if u.Profile.Name != "Алексей" || u.Plan == nil {
		t.Errorf("demo user not populated properly")
	}

	// 3. Отметка тренировки
	if err := svc.CompleteWorkout(u, 0, true, "Отличное начало", 5); err != nil {
		t.Fatalf("CompleteWorkout failed: %v", err)
	}
	foundSession := false
	for _, s := range u.Progress.Sessions {
		if s.DayIndex == 0 && s.Completed {
			foundSession = true
			break
		}
	}
	if !foundSession {
		t.Errorf("expected session for day 0 to be completed")
	}

	// 4. Замена упражнения
	if len(u.Plan.Days[0].Exercises) > 0 {
		oldExID := u.Plan.Days[0].Exercises[0].ExerciseID
		newEx, err := svc.SwapExercise(u, 0, oldExID)
		if err != nil {
			t.Fatalf("SwapExercise failed: %v", err)
		}
		if newEx.ExerciseID == oldExID {
			t.Errorf("expected new exercise ID to differ from old")
		}
	}

	// 5. Логирование воды
	currentWater, err := svc.LogWater(u, "2026-11-13", 250)
	if err != nil {
		t.Fatalf("LogWater failed: %v", err)
	}
	if currentWater < 250 {
		t.Errorf("expected water >= 250, got %d", currentWater)
	}

	// 6. Смена цели
	if err := svc.ChangeGoal(u, models.GoalGain); err != nil {
		t.Fatalf("ChangeGoal failed: %v", err)
	}
	if u.Profile.Goal != models.GoalGain {
		t.Errorf("expected goal %s, got %s", models.GoalGain, u.Profile.Goal)
	}

	// 7. Прогресс веса
	if err := svc.AddWeight(u, "2026-11-13", 79.0, "Замер"); err != nil {
		t.Fatalf("AddWeight failed: %v", err)
	}
	if u.Profile.WeightKg == nil || *u.Profile.WeightKg != 79.0 {
		t.Errorf("expected weight 79.0")
	}
	if err := svc.DeleteWeight(u, "2026-11-13"); err != nil {
		t.Fatalf("DeleteWeight failed: %v", err)
	}
}
