package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pulsefit/internal/models"
)

// JSONResponse отправляет клиенту форматированный JSON-ответ с заголовками.
func JSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("ошибка кодирования JSON-ответа", "error", err)
	}
}

// JSONError отправляет ошибку в стандартном JSON-формате.
func JSONError(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, map[string]any{"error": message})
}

// MapUserProfile маппит доменную модель Profile в структуру для фронтенда.
func MapUserProfile(u *models.User) map[string]any {
	if u == nil {
		return nil
	}
	p := u.Profile
	bmiVal, _, _ := p.BMI()

	var ageVal int
	if p.Age != nil {
		ageVal = *p.Age
	}
	var heightVal float64
	if p.HeightCm != nil {
		heightVal = *p.HeightCm
	}
	var weightVal float64
	if p.WeightKg != nil {
		weightVal = *p.WeightKg
	}

	isVegan := false
	for _, dp := range p.DietPreferences {
		if dp == "vegan" || dp == "веган" || dp == "растительное" {
			isVegan = true
			break
		}
	}

	calTarget := 2200
	proteinTarget := 120
	fatTarget := 65
	carbTarget := 250
	if u.Plan != nil {
		if u.Plan.KcalTarget != nil && *u.Plan.KcalTarget > 0 {
			calTarget = *u.Plan.KcalTarget
		}
		if u.Plan.ProteinG != nil && *u.Plan.ProteinG > 0 {
			proteinTarget = *u.Plan.ProteinG
		}
		if u.Plan.FatG != nil && *u.Plan.FatG > 0 {
			fatTarget = *u.Plan.FatG
		}
		if u.Plan.CarbsG != nil && *u.Plan.CarbsG > 0 {
			carbTarget = *u.Plan.CarbsG
		}
	}

	waterTarget := 2000
	if weightVal > 0 {
		waterTarget = int(weightVal * 35)
	}

	return map[string]any{
		"id":              p.ID,
		"code":            p.Code,
		"name":            p.Name,
		"gender":          p.Sex,
		"age":             ageVal,
		"height":          heightVal,
		"weight":          weightVal,
		"targetWeight":    weightVal,
		"goal":            p.Goal,
		"level":           p.Level,
		"equipment":       p.Equipment,
		"daysPerWeek":     p.WorkoutsPerWeek,
		"workoutDuration": p.SessionMinutes,
		"preferredTime":   p.PreferredTime,
		"pace":            p.Pace,
		"isVegan":         isVegan,
		"healthIssues":    p.HealthFlags,
		"hasHealthIssues": p.MedicalBlocked() || len(p.HealthFlags) > 0,
		"bmi":             bmiVal,
		"calorieTarget":   calTarget,
		"proteinTarget":   proteinTarget,
		"fatTarget":       fatTarget,
		"carbTarget":      carbTarget,
		"waterTarget":     waterTarget,
		"onboarded":       p.Onboarded,
		"planVersion":     p.PlanVersion,
		"createdAt":       p.CreatedAt.Format(time.RFC3339),
		"updatedAt":       p.UpdatedAt.Format(time.RFC3339),
	}
}

// MapWorkoutDay маппит PlanDay в WorkoutDay для SPA.
func MapWorkoutDay(d models.PlanDay, sessions []models.SessionEntry) map[string]any {
	todayStr := time.Now().Format("2006-01-02")
	completed := false
	for _, s := range sessions {
		if (s.DayIndex == d.Index || s.Date == todayStr) && s.Completed {
			completed = true
			break
		}
	}

	var weekdayName string
	if d.Index >= 0 && d.Index < len(models.WeekdayNames) {
		weekdayName = models.WeekdayNames[d.Index]
	} else {
		weekdayName = fmt.Sprintf("День %d", d.Index+1)
	}

	exList := make([]map[string]any, 0, len(d.Exercises))
	for _, ex := range d.Exercises {
		exList = append(exList, map[string]any{
			"id":          ex.ExerciseID,
			"title":       ex.Name,
			"muscle":      strings.Join(ex.Muscles, ", "),
			"equipment":   models.EquipLabel(ex.Equipment),
			"difficulty":  models.LevelLabel(ex.Level),
			"sets":        ex.Sets,
			"reps":        ex.Reps,
			"durationSec": ex.DurationSec,
			"restSec":     ex.RestSec,
			"coachTip":    "Контролируйте движение в негативной фазе и соблюдайте ровное дыхание.",
		})
	}

	return map[string]any{
		"dayIndex":         d.Index + 1,
		"weekdayName":      weekdayName,
		"dayName":          d.Title,
		"focus":            d.Focus,
		"isRest":           d.Rest,
		"estimatedMinutes": d.Minutes,
		"completed":        completed,
		"exercises":        exList,
	}
}

// MapDayMeals маппит MealDay в структуру питания для SPA.
func MapDayMeals(md models.MealDay, eatenMeals map[string]bool, plan *models.Plan) map[string]any {
	var weekdayName string
	if md.Index >= 0 && md.Index < len(models.WeekdayNames) {
		weekdayName = models.WeekdayNames[md.Index]
	} else {
		weekdayName = fmt.Sprintf("День %d", md.Index+1)
	}

	categorized := map[string][]map[string]any{
		"breakfast": {},
		"lunch":     {},
		"snack":     {},
		"dinner":    {},
	}

	currentCals := 0
	currentProteins := 0.0
	currentFats := 0.0
	currentCarbs := 0.0

	for _, m := range md.Meals {
		eaten := false
		if eatenMeals != nil && eatenMeals[m.MealID] {
			eaten = true
			currentCals += m.Kcal
			currentProteins += m.ProteinG
			currentFats += m.FatG
			currentCarbs += m.CarbsG
		}

		item := map[string]any{
			"id":          m.MealID,
			"type":        m.Slot,
			"name":        m.Name,
			"description": m.Portion,
			"calories":    m.Kcal,
			"proteins":    m.ProteinG,
			"fats":        m.FatG,
			"carbs":       m.CarbsG,
			"isVegan":     false,
			"eaten":       eaten,
		}

		switch m.Slot {
		case "breakfast", "завтрак":
			categorized["breakfast"] = append(categorized["breakfast"], item)
		case "lunch", "обед":
			categorized["lunch"] = append(categorized["lunch"], item)
		case "snack", "полдник", "перекус":
			categorized["snack"] = append(categorized["snack"], item)
		case "dinner", "ужин":
			categorized["dinner"] = append(categorized["dinner"], item)
		default:
			categorized["lunch"] = append(categorized["lunch"], item)
		}
	}

	targetCal := 2200
	targetProt := 120
	targetFat := 65
	targetCarb := 250
	if plan != nil {
		if plan.KcalTarget != nil && *plan.KcalTarget > 0 {
			targetCal = *plan.KcalTarget
		}
		if plan.ProteinG != nil && *plan.ProteinG > 0 {
			targetProt = *plan.ProteinG
		}
		if plan.FatG != nil && *plan.FatG > 0 {
			targetFat = *plan.FatG
		}
		if plan.CarbsG != nil && *plan.CarbsG > 0 {
			targetCarb = *plan.CarbsG
		}
	}

	return map[string]any{
		"dayIndex":        md.Index + 1,
		"weekdayName":     weekdayName,
		"targetCalories":  targetCal,
		"targetProteins":  targetProt,
		"targetFats":      targetFat,
		"targetCarbs":     targetCarb,
		"currentCalories": currentCals,
		"currentProteins": int(currentProteins),
		"currentFats":     int(currentFats),
		"currentCarbs":    int(currentCarbs),
		"meals":           categorized,
	}
}

// HandleAPIBootstrap проверяет состояние пользователя и возвращает метаданные.
func (h *Handler) HandleAPIBootstrap(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "не удалось получить профиль")
		return
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"onboarded":  u.Profile.Onboarded,
		"user":       MapUserProfile(u),
		"todayIndex": models.TodayIndex() + 1,
	})
}

// HandleAPIOnboarding принимает JSON анкету и генерирует персональный план.
func (h *Handler) HandleAPIOnboarding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "ошибка сессии")
		return
	}

	var req struct {
		Name            string   `json:"name"`
		Gender          string   `json:"gender"`
		Age             int      `json:"age"`
		Height          float64  `json:"height"`
		Weight          float64  `json:"weight"`
		TargetWeight    float64  `json:"targetWeight"`
		Goal            string   `json:"goal"`
		Level           string   `json:"level"`
		Equipment       string   `json:"equipment"`
		DaysPerWeek     int      `json:"daysPerWeek"`
		WorkoutDuration int      `json:"workoutDuration"`
		PreferredTime   string   `json:"preferredTime"`
		IsVegan         bool     `json:"isVegan"`
		HealthIssues    []string `json:"healthIssues"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "неверный формат запроса: "+err.Error())
		return
	}

	prof := u.Profile
	prof.Name = req.Name
	prof.Sex = req.Gender
	prof.Age = &req.Age
	prof.HeightCm = &req.Height
	prof.WeightKg = &req.Weight
	prof.Goal = req.Goal
	prof.Level = req.Level
	prof.Equipment = req.Equipment
	prof.WorkoutsPerWeek = req.DaysPerWeek
	prof.SessionMinutes = req.WorkoutDuration
	prof.PreferredTime = req.PreferredTime
	prof.Consent = true
	prof.Onboarded = true

	if req.IsVegan {
		prof.DietPreferences = []string{"vegan"}
	} else {
		prof.DietPreferences = nil
	}

	prof.HealthFlags = nil
	for _, hf := range req.HealthIssues {
		switch hf {
		case "joints", "spine":
			prof.HealthFlags = append(prof.HealthFlags, models.HealthInjury)
		case "cardio":
			prof.HealthFlags = append(prof.HealthFlags, models.HealthHeart)
		default:
			prof.HealthFlags = append(prof.HealthFlags, hf)
		}
	}

	if err := h.svc.SaveOnboarding(u, prof); err != nil {
		JSONError(w, http.StatusInternalServerError, "ошибка сохранения онбординга: "+err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"ok":              true,
		"user":            MapUserProfile(u),
		"hasMedicalFlags": prof.MedicalBlocked() || len(prof.HealthFlags) > 0,
	})
}

// HandleAPIDashboard возвращает данные для главной страницы дашборда.
func (h *Handler) HandleAPIDashboard(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "пользователь не найден")
		return
	}

	todayIdx := models.TodayIndex()
	var todayWorkout map[string]any
	var todayMeals map[string]any

	if u.Plan != nil {
		if todayIdx < len(u.Plan.Days) {
			todayWorkout = MapWorkoutDay(u.Plan.Days[todayIdx], u.Progress.Sessions)
		}
		if todayIdx < len(u.Plan.Meals) {
			todayMeals = MapDayMeals(u.Plan.Meals[todayIdx], u.Progress.Meals, u.Plan)
		}
	}

	// Весовая история
	weights := make([]map[string]any, 0, len(u.Progress.Weights))
	for _, wItem := range u.Progress.Weights {
		weights = append(weights, map[string]any{
			"date":   wItem.Date,
			"weight": wItem.Kg,
			"note":   wItem.Note,
		})
	}

	// История тренировок
	sessions := make([]map[string]any, 0, len(u.Progress.Sessions))
	for _, s := range u.Progress.Sessions {
		sessions = append(sessions, map[string]any{
			"date":           s.Date,
			"dayIndex":       s.DayIndex + 1,
			"dayName":        s.Title,
			"focus":          s.Feeling,
			"exercisesCount": 4,
		})
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"user":           MapUserProfile(u),
		"todayWorkout":   todayWorkout,
		"todayMeals":     todayMeals,
		"todayWater":     u.WaterToday(),
		"streak":         len(u.Progress.Sessions) + 1,
		"weightHistory":  weights,
		"workoutHistory": sessions,
		"isGentleMode":   u.Profile.MedicalBlocked() || len(u.Profile.HealthFlags) > 0,
		"reminders":      u.Reminders,
	})
}

// HandleAPIWorkouts возвращает тренировки для запрошенного дня или всей недели.
func (h *Handler) HandleAPIWorkouts(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil || u.Plan == nil {
		JSONError(w, http.StatusNotFound, "план тренировок не найден")
		return
	}

	dayParam := r.URL.Query().Get("day")
	dayIdx := models.TodayIndex()
	if dayParam != "" {
		if parsed, err := strconv.Atoi(dayParam); err == nil && parsed >= 1 && parsed <= 7 {
			dayIdx = parsed - 1
		}
	}

	allDays := make([]map[string]any, 0, len(u.Plan.Days))
	var selectedWorkout map[string]any

	for i, d := range u.Plan.Days {
		mapped := MapWorkoutDay(d, u.Progress.Sessions)
		allDays = append(allDays, mapped)
		if i == dayIdx {
			selectedWorkout = mapped
		}
	}

	if selectedWorkout == nil && len(allDays) > 0 {
		selectedWorkout = allDays[0]
		dayIdx = 0
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"user":        MapUserProfile(u),
		"selectedDay": dayIdx + 1,
		"workout":     selectedWorkout,
		"allDays":     allDays,
	})
}

// HandleAPIWorkoutComplete отмечает тренировку дня завершенной.
func (h *Handler) HandleAPIWorkoutComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "сессия не найдена")
		return
	}

	var req struct {
		Day       int    `json:"day"`
		Completed bool   `json:"completed"`
		Feeling   string `json:"feeling"`
		Energy    int    `json:"energy"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "неверный формат тела запроса")
		return
	}

	dayIdx := req.Day - 1
	if dayIdx < 0 {
		dayIdx = 0
	}

	if err := h.svc.CompleteWorkout(u, dayIdx, req.Completed, req.Feeling, req.Energy); err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "тренировка успешно зафиксирована",
	})
}

// HandleAPIWorkoutSwap заменяет упражнение на альтернативное из пула.
func (h *Handler) HandleAPIWorkoutSwap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "сессия не найдена")
		return
	}

	var req struct {
		Day        int    `json:"day"`
		ExerciseID string `json:"exerciseId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "неверный формат тела запроса")
		return
	}

	dayIdx := req.Day - 1
	if dayIdx < 0 {
		dayIdx = 0
	}

	newEx, err := h.svc.SwapExercise(u, dayIdx, req.ExerciseID)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "ошибка замены: "+err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"ok": true,
		"newExercise": map[string]any{
			"id":          newEx.ExerciseID,
			"title":       newEx.Name,
			"muscle":      strings.Join(newEx.Muscles, ", "),
			"equipment":   models.EquipLabel(newEx.Equipment),
			"difficulty":  models.LevelLabel(newEx.Level),
			"sets":        newEx.Sets,
			"reps":        newEx.Reps,
			"durationSec": newEx.DurationSec,
			"restSec":     newEx.RestSec,
			"coachTip":    "Выполняйте упражнение подконтрольно, акцентируя внимание на рабочей мышечной группе.",
		},
	})
}

// HandleAPINutrition возвращает план питания для дня.
func (h *Handler) HandleAPINutrition(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil || u.Plan == nil {
		JSONError(w, http.StatusNotFound, "рацион не найден")
		return
	}

	dayParam := r.URL.Query().Get("day")
	dayIdx := models.TodayIndex()
	if dayParam != "" {
		if parsed, err := strconv.Atoi(dayParam); err == nil && parsed >= 1 && parsed <= 7 {
			dayIdx = parsed - 1
		}
	}

	if dayIdx >= len(u.Plan.Meals) {
		dayIdx = 0
	}

	mappedMeals := MapDayMeals(u.Plan.Meals[dayIdx], u.Progress.Meals, u.Plan)

	waterTarget := 2000
	if u.Profile.WeightKg != nil && *u.Profile.WeightKg > 0 {
		waterTarget = int(*u.Profile.WeightKg * 35)
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"user":        MapUserProfile(u),
		"selectedDay": dayIdx + 1,
		"dayMeals":    mappedMeals,
		"waterIntake": u.WaterToday(),
		"waterTarget": waterTarget,
	})
}

// HandleAPINutritionToggle переключает статус блюда (съедено/нет).
func (h *Handler) HandleAPINutritionToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "сессия не найдена")
		return
	}

	var req struct {
		Day    int    `json:"day"`
		MealID string `json:"mealId"`
		Eaten  bool   `json:"eaten"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "неверный формат тела запроса")
		return
	}

	dayIdx := req.Day - 1
	if dayIdx < 0 {
		dayIdx = 0
	}

	if err := h.svc.ToggleMeal(u, dayIdx, req.MealID, req.Eaten); err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"ok":    true,
		"eaten": req.Eaten,
	})
}

// HandleAPIWaterAdd добавляет количество выпитой воды (в мл).
func (h *Handler) HandleAPIWaterAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "сессия не найдена")
		return
	}

	var req struct {
		Delta int `json:"delta"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Fallback for form-data
		deltaStr := r.FormValue("delta")
		req.Delta, _ = strconv.Atoi(deltaStr)
	}

	if req.Delta == 0 {
		req.Delta = 330 // default 1 glass
	}

	today := time.Now().Format("2006-01-02")
	total, err := h.svc.LogWater(u, today, req.Delta)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	waterTarget := 2000
	if u.Profile.WeightKg != nil && *u.Profile.WeightKg > 0 {
		waterTarget = int(*u.Profile.WeightKg * 35)
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"ok":      true,
		"current": total,
		"target":  waterTarget,
		"glasses": total / 330,
	})
}

// HandleAPIProgress возвращает данные прогресса: замеры веса и тренировочную активность.
func (h *Handler) HandleAPIProgress(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "сессия не найдена")
		return
	}

	weights := make([]map[string]any, 0, len(u.Progress.Weights))
	for _, wItem := range u.Progress.Weights {
		weights = append(weights, map[string]any{
			"date":   wItem.Date,
			"weight": wItem.Kg,
			"note":   wItem.Note,
		})
	}

	sessions := make([]map[string]any, 0, len(u.Progress.Sessions))
	for _, s := range u.Progress.Sessions {
		sessions = append(sessions, map[string]any{
			"date":           s.Date,
			"dayIndex":       s.DayIndex + 1,
			"dayName":        s.Title,
			"focus":          s.Feeling,
			"exercisesCount": 4,
		})
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"user":           MapUserProfile(u),
		"weightHistory":  weights,
		"workoutHistory": sessions,
		"streak":         len(u.Progress.Sessions) + 1,
	})
}

// HandleAPIProgressWeight сохраняет новый замер веса с заметкой.
func (h *Handler) HandleAPIProgressWeight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "сессия не найдена")
		return
	}

	var req struct {
		Weight float64 `json:"weight"`
		Note   string  `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wStr := r.FormValue("weight")
		req.Weight, _ = strconv.ParseFloat(wStr, 64)
		req.Note = r.FormValue("note")
	}

	if req.Weight < 30 || req.Weight > 300 {
		JSONError(w, http.StatusBadRequest, "вес должен быть в диапазоне от 30 до 300 кг")
		return
	}

	today := time.Now().Format("2006-01-02")
	if err := h.svc.AddWeight(u, today, req.Weight, req.Note); err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"ok": true,
		"entry": map[string]any{
			"date":   today,
			"weight": req.Weight,
			"note":   req.Note,
		},
	})
}

// HandleAPICatalog возвращает упражнения из каталога с фильтрацией.
func (h *Handler) HandleAPICatalog(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	muscle := strings.TrimSpace(r.URL.Query().Get("muscle"))
	equipment := strings.TrimSpace(r.URL.Query().Get("equipment"))
	difficulty := strings.TrimSpace(r.URL.Query().Get("difficulty"))

	exercises := h.svc.Catalog().Exercises
	filtered := make([]map[string]any, 0, len(exercises))

	for _, ex := range exercises {
		if q != "" {
			nameLower := strings.ToLower(ex.Name)
			matches := strings.Contains(nameLower, q)
			for _, m := range ex.Muscles {
				if strings.Contains(strings.ToLower(m), q) {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
		}

		if muscle != "" {
			muscleMatch := false
			for _, m := range ex.Muscles {
				if strings.EqualFold(m, muscle) || strings.Contains(strings.ToLower(m), strings.ToLower(muscle)) {
					muscleMatch = true
					break
				}
			}
			if !muscleMatch {
				continue
			}
		}

		if equipment != "" && !strings.EqualFold(ex.Equipment, equipment) && !strings.Contains(strings.ToLower(ex.Equipment), strings.ToLower(equipment)) {
			continue
		}

		if difficulty != "" && !strings.EqualFold(ex.Level, difficulty) {
			continue
		}

		isGentle := ex.Level == models.LevelBeginner || ex.Equipment == models.EquipNone

		filtered = append(filtered, map[string]any{
			"id":         ex.ID,
			"title":      ex.Name,
			"muscle":     strings.Join(ex.Muscles, ", "),
			"equipment":  models.EquipLabel(ex.Equipment),
			"difficulty": models.LevelLabel(ex.Level),
			"coachTip":   ex.Safety,
			"isGentle":   isGentle,
		})
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"exercises": filtered,
	})
}

// HandleAPICatalogExercise возвращает детальную карточку упражнения.
func (h *Handler) HandleAPICatalogExercise(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	ex, ok := h.svc.Catalog().Exercise(id)
	if !ok {
		JSONError(w, http.StatusNotFound, "упражнение не найдено")
		return
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"exercise": map[string]any{
			"id":         ex.ID,
			"title":      ex.Name,
			"muscle":     strings.Join(ex.Muscles, ", "),
			"equipment":  models.EquipLabel(ex.Equipment),
			"difficulty": models.LevelLabel(ex.Level),
			"coachTip":   ex.Safety,
			"isGentle":   ex.Level == models.LevelBeginner,
		},
	})
}

// HandleAPIProfile возвращает профиль пользователя.
func (h *Handler) HandleAPIProfile(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "сессия не найдена")
		return
	}

	if r.Method == http.MethodGet {
		JSONResponse(w, http.StatusOK, map[string]any{
			"user": MapUserProfile(u),
		})
		return
	}

	// POST update
	var req struct {
		Name            string   `json:"name"`
		Gender          string   `json:"gender"`
		Age             int      `json:"age"`
		Height          float64  `json:"height"`
		Weight          float64  `json:"weight"`
		TargetWeight    float64  `json:"targetWeight"`
		Goal            string   `json:"goal"`
		Level           string   `json:"level"`
		Equipment       string   `json:"equipment"`
		DaysPerWeek     int      `json:"daysPerWeek"`
		WorkoutDuration int      `json:"workoutDuration"`
		PreferredTime   string   `json:"preferredTime"`
		IsVegan         bool     `json:"isVegan"`
		HealthIssues    []string `json:"healthIssues"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "неверный формат тела запроса")
		return
	}

	prof := u.Profile
	if req.Name != "" {
		prof.Name = req.Name
	}
	if req.Gender != "" {
		prof.Sex = req.Gender
	}
	if req.Age > 0 {
		prof.Age = &req.Age
	}
	if req.Height > 0 {
		prof.HeightCm = &req.Height
	}
	if req.Weight > 0 {
		prof.WeightKg = &req.Weight
	}
	if req.Goal != "" {
		prof.Goal = req.Goal
	}
	if req.Level != "" {
		prof.Level = req.Level
	}
	if req.Equipment != "" {
		prof.Equipment = req.Equipment
	}
	if req.DaysPerWeek > 0 {
		prof.WorkoutsPerWeek = req.DaysPerWeek
	}
	if req.WorkoutDuration > 0 {
		prof.SessionMinutes = req.WorkoutDuration
	}
	if req.PreferredTime != "" {
		prof.PreferredTime = req.PreferredTime
	}

	if req.IsVegan {
		prof.DietPreferences = []string{"vegan"}
	} else {
		prof.DietPreferences = nil
	}

	if req.HealthIssues != nil {
		prof.HealthFlags = nil
		for _, hf := range req.HealthIssues {
			switch hf {
			case "joints", "spine":
				prof.HealthFlags = append(prof.HealthFlags, models.HealthInjury)
			case "cardio":
				prof.HealthFlags = append(prof.HealthFlags, models.HealthHeart)
			default:
				prof.HealthFlags = append(prof.HealthFlags, hf)
			}
		}
	}

	if err := h.svc.UpdateProfile(u, prof); err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"ok":   true,
		"user": MapUserProfile(u),
	})
}

// HandleAPIPlanExport возвращает полную 7-дневную программу для печати и экспорта.
func (h *Handler) HandleAPIPlanExport(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil || u.Plan == nil {
		JSONError(w, http.StatusNotFound, "план не найден")
		return
	}

	days := make([]map[string]any, 0, len(u.Plan.Days))
	for _, d := range u.Plan.Days {
		days = append(days, MapWorkoutDay(d, u.Progress.Sessions))
	}

	meals := make([]map[string]any, 0, len(u.Plan.Meals))
	for _, m := range u.Plan.Meals {
		meals = append(meals, MapDayMeals(m, u.Progress.Meals, u.Plan))
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"user":  MapUserProfile(u),
		"days":  days,
		"meals": meals,
	})
}

// HandleAPIDemoSeed мгновенно активирует демонстрационный сценарий.
func (h *Handler) HandleAPIDemoSeed(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		JSONError(w, http.StatusInternalServerError, "сессия не найдена")
		return
	}

	if err := h.svc.SeedDemoUser(u); err != nil {
		JSONError(w, http.StatusInternalServerError, "ошибка сидирования демо: "+err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"ok":       true,
		"redirect": "/dashboard",
	})
}

// HandleAPIDemoReset сбрасывает все данные пользователя.
func (h *Handler) HandleAPIDemoReset(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u != nil {
		_ = h.svc.DeleteUser(u.Profile.ID)
	}

	// Очищаем куку
	http.SetCookie(w, &http.Cookie{
		Name:     "pulsefit_uid",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	JSONResponse(w, http.StatusOK, map[string]any{
		"ok":       true,
		"redirect": "/onboarding",
	})
}

// HandleSPA раздает скомпилированный SPA index.html для любых клиентских путей.
func (h *Handler) HandleSPA(w http.ResponseWriter, r *http.Request) {
	distFile, err := h.assets.Open("dist/index.html")
	if err != nil {
		// Fallback к корню
		h.HandleRoot(w, r)
		return
	}
	defer distFile.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, distFile)
}
