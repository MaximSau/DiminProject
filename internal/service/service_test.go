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


func TestPlanDaysAndGoalAdaptation(t *testing.T) {
	cat, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}

	// 1. Проверяем генерацию от 1 до 7 дней в неделю
	for daysCount := 1; daysCount <= 7; daysCount++ {
		p := models.Profile{
			Goal:            models.GoalLose,
			WorkoutsPerWeek: daysCount,
			SessionMinutes:  45,
			Level:           models.LevelBeginner,
			Equipment:       models.EquipNone,
		}
		plan, err := cat.BuildPlan(p)
		if err != nil {
			t.Fatalf("BuildPlan failed for %d days: %v", daysCount, err)
		}
		trainingDays := 0
		for _, d := range plan.Days {
			if !d.Rest {
				trainingDays++
			}
		}
		if trainingDays != daysCount {
			t.Errorf("expected %d training days, got %d", daysCount, trainingDays)
		}
	}

	// 2. Проверяем адаптацию GoalLose: кардио-финишер присутствует, повторения 12-15
	pLose := models.Profile{
		Goal:            models.GoalLose,
		WorkoutsPerWeek: 3,
		SessionMinutes:  45,
		Level:           models.LevelBeginner,
		Equipment:       models.EquipNone,
	}
	planLose, err := cat.BuildPlan(pLose)
	if err != nil {
		t.Fatalf("BuildPlan lose failed: %v", err)
	}
	hasCardio := false
	for _, d := range planLose.Days {
		if !d.Rest {
			for _, ex := range d.Exercises {
				if ex.Kind == "cardio" {
					hasCardio = true
				}
				if ex.Kind == "strength" && ex.Reps != "12–15" {
					t.Errorf("expected GoalLose strength reps 12–15, got %s", ex.Reps)
				}
			}
		}
	}
	if !hasCardio {
		t.Errorf("expected cardio in GoalLose workout plan")
	}

	// 3. Проверяем адаптацию GoalGain: отдых 75 сек, повторения 8-10
	pGain := models.Profile{
		Goal:            models.GoalGain,
		WorkoutsPerWeek: 3,
		SessionMinutes:  45,
		Level:           models.LevelBeginner,
		Equipment:       models.EquipNone,
	}
	planGain, err := cat.BuildPlan(pGain)
	if err != nil {
		t.Fatalf("BuildPlan gain failed: %v", err)
	}
	for _, d := range planGain.Days {
		if !d.Rest {
			for _, ex := range d.Exercises {
				if ex.Kind == "strength" {
					if ex.RestSec != 75 {
						t.Errorf("expected GoalGain rest 75, got %d", ex.RestSec)
					}
					if ex.Reps != "8–10" {
						t.Errorf("expected GoalGain reps 8–10, got %s", ex.Reps)
					}
				}
			}
		}
	}

	// 4. Проверяем фильтрацию веганства
	pVegan := models.Profile{
		Goal:            models.GoalLose,
		WorkoutsPerWeek: 3,
		SessionMinutes:  45,
		DietPreferences: []string{"vegan"},
	}
	planVegan, err := cat.BuildPlan(pVegan)
	if err != nil {
		t.Fatalf("BuildPlan vegan failed: %v", err)
	}
	for _, d := range planVegan.Meals {
		for _, m := range d.Meals {
			catMeal, ok := cat.Meal(m.MealID)
			if ok {
				isVegan := false
				for _, tag := range catMeal.Tags {
					if tag == "vegan" {
						isVegan = true
					}
				}
				if !isVegan {
					t.Errorf("expected vegan meal, but %s is not vegan", m.Name)
				}
			}
		}
	}
}
