package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"pulsefit/internal/models"
	"pulsefit/internal/service"
)

// PageData модель данных для рендеринга HTML-страниц PulseFit.
type PageData struct {
	ActiveTab   string
	User        *models.User
	AI          map[string]string
	Error       string
	Notice      string
	Extra       any
	SelectedDay int
	TodayIdx    int
	WeightChart template.HTML
}

// Handler реализует транспортный HTTP-слой приложения PulseFit.
type Handler struct {
	mu        sync.Mutex
	svc       *service.Service
	assets    fs.FS
	templates map[string]*template.Template
}

// NewHandler инициализирует обработчики и компилирует шаблоны интерфейса.
func NewHandler(svc *service.Service, assets fs.FS) (*Handler, error) {
	h := &Handler{
		svc:       svc,
		assets:    assets,
		templates: make(map[string]*template.Template),
	}

	if err := h.initTemplates(); err != nil {
		return nil, fmt.Errorf("ошибка компиляции шаблонов: %w", err)
	}

	return h, nil
}

func (h *Handler) Service() *service.Service {
	return h.svc
}

func (h *Handler) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"eq": func(a, b any) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
		"ne": func(a, b any) bool {
			return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"containsStr": func(slice []string, val string) bool {
			for _, item := range slice {
				if item == val {
					return true
				}
			}
			return false
		},
		"formatDate": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("02.01.2006 15:04")
		},
		"goalLabel":     models.GoalLabel,
		"goalShort":     models.GoalShort,
		"levelLabel":    models.LevelLabel,
		"equipLabel":    models.EquipLabel,
		"activityLabel": models.ActivityLabel,
		"timeLabel":     models.TimeLabel,
		"paceLabel":     models.PaceLabel,
		"healthLabel":   models.HealthLabel,
		"todayIndex":    models.TodayIndex,
		"weekdayName": func(idx int) string {
			if idx >= 0 && idx < len(models.WeekdayNames) {
				return models.WeekdayNames[idx]
			}
			return fmt.Sprintf("День %d", idx+1)
		},
		"isMealEaten": func(meals map[string]bool, dayIndex int, mealID string) bool {
			if meals == nil {
				return false
			}
			key := fmt.Sprintf("%d:%s", dayIndex, mealID)
			return meals[key]
		},
		"isSessionCompleted": func(sessions []models.SessionEntry, dayIndex int) bool {
			for _, s := range sessions {
				if s.DayIndex == dayIndex && s.Completed {
					return true
				}
			}
			return false
		},
		"userBMI": func(p models.Profile) map[string]any {
			bmi, label, status := p.BMI()
			minW, maxW := p.IdealWeightRange()
			return map[string]any{
				"value":  bmi,
				"label":  label,
				"status": status,
				"minW":   minW,
				"maxW":   maxW,
			}
		},
		"waterToday": func(u *models.User) int {
			if u == nil {
				return 0
			}
			return u.WaterToday()
		},
		"waterPercent": func(u *models.User) int {
			if u == nil {
				return 0
			}
			cur := u.WaterToday()
			target := u.Profile.WaterTargetMl()
			if target <= 0 {
				return 0
			}
			pct := cur * 100 / target
			if pct > 100 {
				return 100
			}
			return pct
		},
		"waterTarget": func(p models.Profile) int {
			return p.WaterTargetMl()
		},
		"waterGlasses": func(u *models.User) int {
			if u == nil {
				return 0
			}
			return (u.WaterToday() + 165) / 330
		},
		"waterTargetGlasses": func(p models.Profile) int {
			target := p.WaterTargetMl()
			g := (target + 165) / 330
			if g < 1 {
				return 1
			}
			return g
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"renderStars": func(count int) template.HTML {
			var b strings.Builder
			for i := 1; i <= 5; i++ {
				if i <= count {
					b.WriteString("★")
				} else {
					b.WriteString("☆")
				}
			}
			return template.HTML(b.String())
		},
	}
}

func (h *Handler) initTemplates() error {
	pages := []string{
		"onboarding",
		"medical_warning",
		"dashboard",
		"workouts",
		"exercise_detail",
		"nutrition",
		"progress",
		"profile",
		"catalog",
		"plan_export",
		"404",
		"500",
	}

	for _, page := range pages {
		t := template.New("layout.html").Funcs(h.templateFuncs())
		var err error
		t, err = t.ParseFS(h.assets, "templates/layout.html", "templates/"+page+".html")
		if err != nil {
			return fmt.Errorf("шаблон %s: %w", page, err)
		}
		h.templates[page] = t
	}
	return nil
}

func (h *Handler) RenderWithStatus(w http.ResponseWriter, page string, data any, status int) {
	t, ok := h.templates[page]
	if !ok {
		slog.Error("шаблон не найден в кэше", "page", page)
		http.Error(w, fmt.Sprintf("Шаблон %s не найден", page), http.StatusInternalServerError)
		return
	}
	buf := new(bytes.Buffer)
	if err := t.ExecuteTemplate(buf, "layout.html", data); err != nil {
		slog.Error("ошибка рендеринга шаблона", "page", page, "error", err)
		http.Error(w, "Ошибка формирования страницы", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status > 0 {
		w.WriteHeader(status)
	}
	_, _ = buf.WriteTo(w)
}

func (h *Handler) Render(w http.ResponseWriter, page string, data any) {
	h.RenderWithStatus(w, page, data, http.StatusOK)
}

func (h *Handler) Render404(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	h.RenderWithStatus(w, "404", PageData{
		ActiveTab: "dashboard",
		User:      u,
	}, http.StatusNotFound)
}

func (h *Handler) Render500(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	h.RenderWithStatus(w, "500", PageData{
		ActiveTab: "dashboard",
		User:      u,
	}, http.StatusInternalServerError)
}

func (h *Handler) GetOrCreateUser(w http.ResponseWriter, r *http.Request) *models.User {
	var uid string
	if cookie, err := r.Cookie("pulsefit_uid"); err == nil && cookie.Value != "" {
		uid = cookie.Value
	}

	u, isNew, err := h.svc.GetOrCreateUser(uid)
	if err != nil || u == nil {
		return nil
	}

	if isNew || uid == "" || uid != u.Profile.ID {
		http.SetCookie(w, &http.Cookie{
			Name:     "pulsefit_uid",
			Value:    u.Profile.ID,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   365 * 24 * 3600,
		})
	}
	return u
}

func (h *Handler) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","app":"pulsefit"}`))
}

func (h *Handler) HandleManifest(w http.ResponseWriter, r *http.Request) {
	manifest := `{
  "name": "PulseFit — Персональный фитнес и питание",
  "short_name": "PulseFit",
  "description": "Умные тренировки, персональный план питания и контроль прогресса",
  "start_url": "/dashboard",
  "display": "standalone",
  "background_color": "#111512",
  "theme_color": "#2e7d32",
  "icons": [
    {
      "src": "/static/icon.svg",
      "sizes": "any",
      "type": "image/svg+xml",
      "purpose": "any maskable"
    }
  ]
}`
	w.Header().Set("Content-Type", "application/manifest+json")
	_, _ = w.Write([]byte(manifest))
}

func (h *Handler) HandleServiceWorker(w http.ResponseWriter, r *http.Request) {
	sw := `const CACHE_NAME = 'pulsefit-v2.2-nocache';
const STATIC_ASSETS = [
  '/static/app.css',
  '/manifest.json'
];

self.addEventListener('install', (e) => {
  e.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS)).then(() => self.skipWaiting())
  );
});

self.addEventListener('activate', (e) => {
  e.waitUntil(
    caches.keys().then((keys) => Promise.all(
      keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key))
    )).then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', (e) => {
  if (e.request.url.includes('/static/')) {
    e.respondWith(
      fetch(e.request).catch(() => caches.match(e.request))
    );
  }
});
`
	w.Header().Set("Content-Type", "application/javascript")
	_, _ = w.Write([]byte(sw))
}

func (h *Handler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.Render404(w, r)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
		return
	}

	if !u.Profile.Onboarded {
		if u.Profile.MedicalBlocked() {
			http.Redirect(w, r, "/medical-warning", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) HandleOnboarding(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		h.Render(w, "onboarding", PageData{
			ActiveTab: "profile",
			User:      u,
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Неверные данные формы", http.StatusBadRequest)
		return
	}

	consent := r.FormValue("consent") == "1" || r.FormValue("consent") == "on" || r.FormValue("consent") == "true"
	if !consent {
		h.Render(w, "onboarding", PageData{
			ActiveTab: "profile",
			User:      u,
			Error:     "Для формирования плана необходимо дать согласие на использование сведений для расчёта.",
		})
		return
	}

	p := u.Profile
	p.Name = strings.TrimSpace(r.FormValue("name"))
	p.Goal = r.FormValue("goal")
	if p.Goal != models.GoalGain && p.Goal != models.GoalLose {
		p.Goal = models.GoalLose
	}
	p.Pace = r.FormValue("pace")
	if p.Pace == "" {
		p.Pace = "moderate"
	}

	if ageStr := strings.TrimSpace(r.FormValue("age")); ageStr != "" {
		if age, err := strconv.Atoi(ageStr); err == nil && age >= 14 && age <= 100 {
			p.Age = &age
		} else {
			p.Age = nil
		}
	} else {
		p.Age = nil
	}

	if hStr := strings.TrimSpace(r.FormValue("height_cm")); hStr != "" {
		if ht, err := strconv.ParseFloat(hStr, 64); err == nil && ht >= 100 && ht <= 250 {
			p.HeightCm = &ht
		} else {
			p.HeightCm = nil
		}
	} else {
		p.HeightCm = nil
	}

	if wStr := strings.TrimSpace(r.FormValue("weight_kg")); wStr != "" {
		if wt, err := strconv.ParseFloat(wStr, 64); err == nil && wt >= 30 && wt <= 300 {
			p.WeightKg = &wt
		} else {
			p.WeightKg = nil
		}
	} else {
		p.WeightKg = nil
	}

	p.Sex = r.FormValue("sex")
	p.Level = r.FormValue("level")
	if p.Level == "" {
		p.Level = models.LevelBeginner
	}

	wpw, _ := strconv.Atoi(r.FormValue("workouts_per_week"))
	if wpw < 1 || wpw > 7 {
		wpw = 3
	}
	p.WorkoutsPerWeek = wpw

	dur, _ := strconv.Atoi(r.FormValue("session_minutes"))
	if dur <= 0 {
		dur = 35
	}
	p.SessionMinutes = dur

	p.Equipment = r.FormValue("equipment")
	if p.Equipment == "" {
		p.Equipment = models.EquipNone
	}

	p.PreferredTime = r.FormValue("preferred_time")
	p.ActivityLevel = r.FormValue("activity_level")
	p.DietPreferences = r.Form["diet_preferences"]
	p.Allergies = r.Form["allergies"]
	p.AvoidFoods = strings.TrimSpace(r.FormValue("avoid_foods"))

	p.HealthFlags = r.Form["health_flags"]
	p.HealthNotes = strings.TrimSpace(r.FormValue("health_notes"))
	p.Consent = true
	p.UpdatedAt = time.Now()

	u.Profile = p

	if p.MedicalBlocked() {
		u.Profile.Onboarded = false
		_ = h.svc.SaveUser(u)
		http.Redirect(w, r, "/medical-warning", http.StatusSeeOther)
		return
	}

	if err := h.svc.SaveOnboarding(u, p); err != nil {
		slog.Error("ошибка сохранения онбординга", "error", err)
		h.Render(w, "onboarding", PageData{
			ActiveTab: "profile",
			User:      u,
			Error:     "Не удалось составить план под указанные параметры. Пожалуйста, проверьте введённые данные.",
		})
		return
	}

	if p.WeightKg != nil {
		_ = h.svc.AddWeight(u, time.Now().Format("2006-01-02"), *p.WeightKg, "Стартовый вес при заполнении анкеты")
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) HandleMedicalWarning(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost && r.FormValue("proceed_anyway") == "1" {
		plan, err := h.svc.Catalog().BuildPlan(u.Profile)
		if err == nil {
			u.Plan = plan
			u.Profile.Onboarded = true
			_ = h.svc.SaveUser(u)
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	h.Render(w, "medical_warning", PageData{
		ActiveTab: "profile",
		User:      u,
	})
}

func (h *Handler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil || !u.Profile.Onboarded || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	todayIdx := models.TodayIndex()
	var todayWorkout *models.PlanDay
	if todayIdx >= 0 && todayIdx < len(u.Plan.Days) {
		todayWorkout = &u.Plan.Days[todayIdx]
	}

	var todayMeals *models.MealDay
	if todayIdx >= 0 && todayIdx < len(u.Plan.Meals) {
		todayMeals = &u.Plan.Meals[todayIdx]
	}

	completedCount := 0
	for _, s := range u.Progress.Sessions {
		if s.Completed {
			completedCount++
		}
	}

	type dashExtra struct {
		TodayWorkout   *models.PlanDay
		TodayMeals     *models.MealDay
		CompletedCount int
		LastWeight     *float64
		WaterToday     int
		WaterTarget    int
		WaterPercent   int
	}

	var lastWeight *float64
	if len(u.Progress.Weights) > 0 {
		wVal := u.Progress.Weights[len(u.Progress.Weights)-1].Kg
		lastWeight = &wVal
	} else if u.Profile.WeightKg != nil {
		lastWeight = u.Profile.WeightKg
	}

	wtCur := u.WaterToday()
	wtTarget := u.Profile.WaterTargetMl()
	wtPct := 0
	if wtTarget > 0 {
		wtPct = wtCur * 100 / wtTarget
		if wtPct > 100 {
			wtPct = 100
		}
	}

	h.Render(w, "dashboard", PageData{
		ActiveTab: "dashboard",
		User:      u,
		TodayIdx:  todayIdx,
		Notice:    r.URL.Query().Get("notice"),
		Extra: dashExtra{
			TodayWorkout:   todayWorkout,
			TodayMeals:     todayMeals,
			CompletedCount: completedCount,
			LastWeight:     lastWeight,
			WaterToday:     wtCur,
			WaterTarget:    wtTarget,
			WaterPercent:   wtPct,
		},
	})
}

func (h *Handler) HandleWorkouts(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil || !u.Profile.Onboarded || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	dayIdx := models.TodayIndex()
	if dStr := r.URL.Query().Get("day"); dStr != "" {
		if d, err := strconv.Atoi(dStr); err == nil && d >= 0 && d < 7 {
			dayIdx = d
		}
	}

	var curDay models.PlanDay
	if len(u.Plan.Days) > 0 {
		if dayIdx < 0 || dayIdx >= len(u.Plan.Days) {
			dayIdx = 0
		}
		curDay = u.Plan.Days[dayIdx]
	}

	isCompleted := false
	for _, s := range u.Progress.Sessions {
		if s.DayIndex == dayIdx && s.Completed {
			isCompleted = true
			break
		}
	}

	type workoutsExtra struct {
		CurrentDay  models.PlanDay
		IsCompleted bool
	}

	h.Render(w, "workouts", PageData{
		ActiveTab:   "workouts",
		User:        u,
		SelectedDay: dayIdx,
		TodayIdx:    models.TodayIndex(),
		Notice:      r.URL.Query().Get("notice"),
		Extra: workoutsExtra{
			CurrentDay:  curDay,
			IsCompleted: isCompleted,
		},
	})
}

func (h *Handler) HandleWorkoutComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u == nil || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	dayIdx, _ := strconv.Atoi(r.FormValue("day"))
	if dayIdx < 0 || dayIdx >= 7 {
		dayIdx = models.TodayIndex()
	}

	energy, _ := strconv.Atoi(r.FormValue("energy"))
	if energy < 1 || energy > 5 {
		energy = 4
	}
	feeling := strings.TrimSpace(r.FormValue("feeling"))
	if feeling == "" {
		feeling = "Тренировка выполнена по плану"
	}

	_ = h.svc.CompleteWorkout(u, dayIdx, true, feeling, energy)
	http.Redirect(w, r, fmt.Sprintf("/workouts?day=%d&notice=Тренировка+отмечена+выполненной!+Отличная+работа.", dayIdx), http.StatusSeeOther)
}

func (h *Handler) HandleWorkoutSwap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u == nil || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	dayIdx, _ := strconv.Atoi(r.FormValue("day"))
	exID := r.FormValue("exercise_id")

	if dayIdx < 0 || dayIdx >= len(u.Plan.Days) || exID == "" {
		http.Redirect(w, r, "/workouts", http.StatusSeeOther)
		return
	}

	newEx, err := h.svc.SwapExercise(u, dayIdx, exID)
	if err == nil && newEx != nil {
		http.Redirect(w, r, fmt.Sprintf("/workouts?day=%d&notice=Упражнение+заменено+на+«%s»", dayIdx, newEx.Name), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/workouts?day=%d", dayIdx), http.StatusSeeOther)
}

func (h *Handler) HandleExerciseDetail(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	id := r.URL.Query().Get("id")
	dayStr := r.URL.Query().Get("day")
	dayIdx, _ := strconv.Atoi(dayStr)

	ex, ok := h.svc.Catalog().Exercise(id)
	if !ok {
		http.Error(w, "Упражнение не найдено", http.StatusNotFound)
		return
	}

	type exExtra struct {
		Exercise models.Exercise
		DayIndex int
	}

	h.Render(w, "exercise_detail", PageData{
		ActiveTab:   "workouts",
		User:        u,
		SelectedDay: dayIdx,
		Extra: exExtra{
			Exercise: ex,
			DayIndex: dayIdx,
		},
	})
}

func (h *Handler) HandleNutrition(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil || !u.Profile.Onboarded || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	dayIdx := models.TodayIndex()
	if dStr := r.URL.Query().Get("day"); dStr != "" {
		if d, err := strconv.Atoi(dStr); err == nil && d >= 0 && d < 7 {
			dayIdx = d
		}
	}

	var curMeals models.MealDay
	if len(u.Plan.Meals) > 0 {
		if dayIdx < 0 || dayIdx >= len(u.Plan.Meals) {
			dayIdx = 0
		}
		curMeals = u.Plan.Meals[dayIdx]
	}

	type nutrExtra struct {
		CurrentMeals models.MealDay
	}

	h.Render(w, "nutrition", PageData{
		ActiveTab:   "nutrition",
		User:        u,
		SelectedDay: dayIdx,
		TodayIdx:    models.TodayIndex(),
		Notice:      r.URL.Query().Get("notice"),
		Extra: nutrExtra{
			CurrentMeals: curMeals,
		},
	})
}

func (h *Handler) HandleNutritionToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/nutrition", http.StatusSeeOther)
		return
	}

	dayIdx, _ := strconv.Atoi(r.FormValue("day"))
	mealID := r.FormValue("meal_id")
	if mealID == "" {
		http.Redirect(w, r, "/nutrition", http.StatusSeeOther)
		return
	}

	key := fmt.Sprintf("%d:%s", dayIdx, mealID)
	currentState := u.Progress.Meals[key]
	_ = h.svc.ToggleMeal(u, dayIdx, mealID, !currentState)

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		strings.Contains(r.Header.Get("Accept"), "application/json") ||
		r.FormValue("ajax") == "1" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":     true,
			"eaten":  !currentState,
			"mealId": mealID,
			"day":    dayIdx,
		})
		return
	}

	returnURL := r.FormValue("return")
	if returnURL == "dashboard" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/nutrition?day=%d", dayIdx), http.StatusSeeOther)
}

func (h *Handler) HandleNutritionSwap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u == nil || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	dayIdx, _ := strconv.Atoi(r.FormValue("day"))
	mealID := r.FormValue("meal_id")

	if dayIdx < 0 || dayIdx >= len(u.Plan.Meals) || mealID == "" {
		http.Redirect(w, r, "/nutrition", http.StatusSeeOther)
		return
	}

	day := &u.Plan.Meals[dayIdx]
	mealIdx := -1
	var targetMeal models.MealItem
	for i, m := range day.Meals {
		if m.MealID == mealID {
			mealIdx = i
			targetMeal = m
			break
		}
	}

	if mealIdx >= 0 {
		slots := []string{targetMeal.Slot}
		poolMap := h.svc.Catalog().SuitableMeals(slots, u.Profile.DietPreferences, u.Profile.Allergies, u.Profile.AvoidFoods)
		pool := poolMap[targetMeal.Slot]

		var candidates []models.Meal
		for _, m := range pool {
			if m.ID != mealID {
				candidates = append(candidates, m)
			}
		}

		if len(candidates) > 0 {
			newMeal := candidates[int(time.Now().UnixNano())%len(candidates)]
			day.Meals[mealIdx] = models.MealItem{
				MealID:      newMeal.ID,
				Name:        newMeal.Name,
				Slot:        newMeal.Slot,
				Kcal:        newMeal.Kcal,
				ProteinG:    newMeal.ProteinG,
				FatG:        newMeal.FatG,
				CarbsG:      newMeal.CarbsG,
				Portion:     newMeal.Portion,
				Ingredients: newMeal.Ingredients,
				Prep:        newMeal.Prep,
				Swaps:       newMeal.Swaps,
				Demo:        true,
			}
			_ = h.svc.SaveUser(u)
			http.Redirect(w, r, fmt.Sprintf("/nutrition?day=%d&notice=Блюдо+заменено+на+«%s»", dayIdx, newMeal.Name), http.StatusSeeOther)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/nutrition?day=%d", dayIdx), http.StatusSeeOther)
}

func (h *Handler) HandleProgress(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil || !u.Profile.Onboarded {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	chartSVG := generateWeightSVG(u.Progress.Weights)

	h.Render(w, "progress", PageData{
		ActiveTab:   "progress",
		User:        u,
		Notice:      r.URL.Query().Get("notice"),
		WeightChart: template.HTML(chartSVG),
	})
}

func (h *Handler) HandleProgressWeight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/progress", http.StatusSeeOther)
		return
	}

	wStr := strings.TrimSpace(r.FormValue("weight_kg"))
	wt, err := strconv.ParseFloat(wStr, 64)
	if err != nil || wt < 30 || wt > 300 {
		http.Redirect(w, r, "/progress?notice=Пожалуйста,+укажите+корректный+вес", http.StatusSeeOther)
		return
	}

	dateStr := strings.TrimSpace(r.FormValue("date"))
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	note := strings.TrimSpace(r.FormValue("note"))

	_ = h.svc.AddWeight(u, dateStr, wt, note)
	http.Redirect(w, r, "/progress?notice=Запись+веса+сохранена!", http.StatusSeeOther)
}

func (h *Handler) HandleProgressCheckin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/progress", http.StatusSeeOther)
		return
	}

	mood, _ := strconv.Atoi(r.FormValue("mood"))
	if mood < 1 || mood > 5 {
		mood = 4
	}
	energy, _ := strconv.Atoi(r.FormValue("energy"))
	if energy < 1 || energy > 5 {
		energy = 4
	}
	nutrition, _ := strconv.Atoi(r.FormValue("nutrition"))
	if nutrition < 1 || nutrition > 5 {
		nutrition = 4
	}
	sleep, _ := strconv.ParseFloat(r.FormValue("sleep"), 64)
	if sleep <= 0 {
		sleep = 7.5
	}
	note := strings.TrimSpace(r.FormValue("note"))

	_ = h.svc.AddCheckin(u, time.Now().Format("2006-01-02"), mood, energy, nutrition, sleep, note)
	http.Redirect(w, r, "/progress?notice=Чек-ин+самочувствия+сохранён!", http.StatusSeeOther)
}

func (h *Handler) HandleWaterAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	delta, _ := strconv.Atoi(r.FormValue("delta"))
	today := time.Now().Format("2006-01-02")
	cur, _ := h.svc.LogWater(u, today, delta)

	target := u.Profile.WaterTargetMl()
	pct := 0
	if target > 0 {
		pct = cur * 100 / target
		if pct > 100 {
			pct = 100
		}
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		strings.Contains(r.Header.Get("Accept"), "application/json") ||
		r.FormValue("ajax") == "1" {
		w.Header().Set("Content-Type", "application/json")
		glasses := (cur + 165) / 330
		targetGlasses := (target + 165) / 330
		if targetGlasses < 1 {
			targetGlasses = 1
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":             true,
			"current":        cur,
			"target":         target,
			"percent":        pct,
			"glasses":        glasses,
			"target_glasses": targetGlasses,
		})
		return
	}

	http.Redirect(w, r, "/dashboard#water-section", http.StatusSeeOther)
}

func (h *Handler) HandleCatalog(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	type catExtra struct {
		Exercises []models.Exercise
		Meals     []models.Meal
	}
	h.Render(w, "catalog", PageData{
		ActiveTab: "catalog",
		User:      u,
		Extra: catExtra{
			Exercises: h.svc.Catalog().Exercises,
			Meals:     h.svc.Catalog().Meals,
		},
	})
}

func (h *Handler) HandlePlanExport(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil || !u.Profile.Onboarded || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}
	h.Render(w, "plan_export", PageData{
		ActiveTab: "workouts",
		User:      u,
	})
}

func (h *Handler) HandleChat(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/catalog", http.StatusMovedPermanently)
}

func (h *Handler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		h.Render(w, "profile", PageData{
			ActiveTab: "profile",
			User:      u,
			Notice:    r.URL.Query().Get("notice"),
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	_ = r.ParseForm()
	p := u.Profile
	p.Name = strings.TrimSpace(r.FormValue("name"))
	p.Goal = r.FormValue("goal")
	p.Pace = r.FormValue("pace")

	if ageStr := strings.TrimSpace(r.FormValue("age")); ageStr != "" {
		if age, err := strconv.Atoi(ageStr); err == nil && age >= 14 && age <= 100 {
			p.Age = &age
		}
	}
	if hStr := strings.TrimSpace(r.FormValue("height_cm")); hStr != "" {
		if ht, err := strconv.ParseFloat(hStr, 64); err == nil && ht >= 100 && ht <= 250 {
			p.HeightCm = &ht
		}
	}
	if wStr := strings.TrimSpace(r.FormValue("weight_kg")); wStr != "" {
		if wt, err := strconv.ParseFloat(wStr, 64); err == nil && wt >= 30 && wt <= 300 {
			p.WeightKg = &wt
		}
	}

	p.Sex = r.FormValue("sex")
	p.Level = r.FormValue("level")
	wpw, _ := strconv.Atoi(r.FormValue("workouts_per_week"))
	if wpw >= 1 && wpw <= 7 {
		p.WorkoutsPerWeek = wpw
	}
	dur, _ := strconv.Atoi(r.FormValue("session_minutes"))
	if dur >= 15 && dur <= 120 {
		p.SessionMinutes = dur
	}
	p.Equipment = r.FormValue("equipment")
	p.PreferredTime = r.FormValue("preferred_time")
	p.ActivityLevel = r.FormValue("activity_level")
	p.DietPreferences = r.Form["diet_preferences"]
	p.Allergies = r.Form["allergies"]
	p.AvoidFoods = strings.TrimSpace(r.FormValue("avoid_foods"))

	p.HealthFlags = r.Form["health_flags"]
	p.HealthNotes = strings.TrimSpace(r.FormValue("health_notes"))

	_ = h.svc.UpdateProfile(u, p)
	http.Redirect(w, r, "/profile?notice=Данные+профиля+успешно+обновлены", http.StatusSeeOther)
}

func (h *Handler) HandleProfileSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	code := strings.ToUpper(strings.TrimSpace(r.FormValue("code")))
	if code == "" {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	if targetUser, ok := h.svc.GetUserByCode(code); ok {
		http.SetCookie(w, &http.Cookie{
			Name:     "pulsefit_uid",
			Value:    targetUser.Profile.ID,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   365 * 24 * 3600,
		})
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/profile?notice=Профиль+с+таким+кодом+не+найден", http.StatusSeeOther)
}

func (h *Handler) HandleRemindersUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	_ = r.ParseForm()
	rems := make([]models.Reminder, len(u.Reminders))
	copy(rems, u.Reminders)
	for i := range rems {
		rem := &rems[i]
		enabledKey := fmt.Sprintf("enabled_%s", rem.ID)
		timeKey := fmt.Sprintf("time_%s", rem.ID)
		daysKey := fmt.Sprintf("days_%s", rem.ID)

		rem.Enabled = r.FormValue(enabledKey) == "1" || r.FormValue(enabledKey) == "on"
		if t := strings.TrimSpace(r.FormValue(timeKey)); t != "" {
			rem.Time = t
		}
		if d := strings.TrimSpace(r.FormValue(daysKey)); d != "" {
			rem.Days = d
		}
	}

	_ = h.svc.SaveReminders(u, rems)
	http.Redirect(w, r, "/profile?notice=Настройки+напоминаний+сохранены", http.StatusSeeOther)
}

func (h *Handler) HandlePlanRecreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if err := h.svc.RecreatePlan(u); err != nil {
		slog.Error("ошибка пересоздания плана", "error", err)
		http.Redirect(w, r, "/profile?notice=Не+удалось+пересоздать+план", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/dashboard?notice=Персональный+план+успешно+обновлён!", http.StatusSeeOther)
}

func (h *Handler) HandlePlanChangeGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	newGoal := r.FormValue("goal")
	_ = h.svc.ChangeGoal(u, newGoal)
	http.Redirect(w, r, "/dashboard?notice=Цель+изменена+на+«"+models.GoalLabel(u.Profile.Goal)+"»!", http.StatusSeeOther)
}

func (h *Handler) HandleDemoSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if err := h.svc.SeedDemoUser(u); err != nil {
		slog.Error("ошибка заполнения демо профиля", "error", err)
		http.Redirect(w, r, "/dashboard?notice=Ошибка+загрузки+демо", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/dashboard?notice=Загружен+демонстрационный+сценарий+с+заполненным+прогрессом!", http.StatusSeeOther)
}

func (h *Handler) HandleDemoReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := h.GetOrCreateUser(w, r)
	if u != nil {
		_ = h.svc.DeleteUser(u.Profile.ID)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "pulsefit_uid",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
}

// generateWeightSVG строит аккуратный SVG график изменения веса.
func generateWeightSVG(entries []models.WeightEntry) string {
	if len(entries) < 2 {
		return `<div class="empty muted" style="padding: 20px 0;">Добавьте хотя бы две записи веса, чтобы увидеть график изменений.</div>`
	}

	width := 420.0
	height := 160.0
	padX := 35.0
	padY := 25.0

	minKg := entries[0].Kg
	maxKg := entries[0].Kg
	for _, e := range entries {
		if e.Kg < minKg {
			minKg = e.Kg
		}
		if e.Kg > maxKg {
			maxKg = e.Kg
		}
	}

	if maxKg-minKg < 1.0 {
		minKg -= 1.0
		maxKg += 1.0
	} else {
		minKg -= 0.5
		maxKg += 0.5
	}

	type pt struct {
		x, y float64
		kg   float64
		date string
	}

	points := make([]pt, len(entries))
	stepX := (width - 2*padX) / float64(len(entries)-1)

	for i, e := range entries {
		x := padX + float64(i)*stepX
		ratio := (e.Kg - minKg) / (maxKg - minKg)
		y := height - padY - ratio*(height-2*padY)
		points[i] = pt{x: x, y: y, kg: e.Kg, date: e.Date}
	}

	var pathBuilder strings.Builder
	for i, p := range points {
		if i == 0 {
			pathBuilder.WriteString(fmt.Sprintf("M %.1f %.1f", p.x, p.y))
		} else {
			pathBuilder.WriteString(fmt.Sprintf(" L %.1f %.1f", p.x, p.y))
		}
	}

	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg viewBox="0 0 %.0f %.0f" class="weight-svg" style="width: 100%%; height: auto; display: block;">`, width, height))

	// Горизонтальные направляющие линии
	svg.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="var(--border)" stroke-dasharray="3,3" />`, padX, padY, width-padX, padY))
	svg.WriteString(fmt.Sprintf(`<text x="5" y="%.1f" font-size="10" fill="var(--muted)">%.1f</text>`, padY+4, maxKg))

	svg.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="var(--border)" stroke-dasharray="3,3" />`, padX, height-padY, width-padX, height-padY))
	svg.WriteString(fmt.Sprintf(`<text x="5" y="%.1f" font-size="10" fill="var(--muted)">%.1f</text>`, height-padY+4, minKg))

	// Заливка под графиком
	var areaBuilder strings.Builder
	areaBuilder.WriteString(pathBuilder.String())
	areaBuilder.WriteString(fmt.Sprintf(" L %.1f %.1f L %.1f %.1f Z", points[len(points)-1].x, height-padY, points[0].x, height-padY))
	svg.WriteString(fmt.Sprintf(`<path d="%s" fill="var(--accent-soft)" opacity="0.4" />`, areaBuilder.String()))

	// Основная линия тренда
	svg.WriteString(fmt.Sprintf(`<path d="%s" fill="none" stroke="var(--accent)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />`, pathBuilder.String()))

	// Точки и подписи
	for _, p := range points {
		svg.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="4" fill="var(--accent)" stroke="var(--card)" stroke-width="2" />`, p.x, p.y))
		shortDate := p.date
		if len(shortDate) >= 5 {
			shortDate = shortDate[5:] // MM-DD
		}
		svg.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" font-size="10" fill="var(--muted)" text-anchor="middle">%.1f</text>`, p.x, p.y-8, p.kg))
		svg.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" font-size="9" fill="var(--muted)" text-anchor="middle">%s</text>`, p.x, height-8, shortDate))
	}

	svg.WriteString(`</svg>`)
	return svg.String()
}
