package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config конфигурация запуска приложения.
type Config struct {
	DataDir string
	Assets  fs.FS
}

// App основной контроллер веб-приложения PulseFit.
type App struct {
	store     *Store
	catalog   *Catalog
	assets    fs.FS
	templates map[string]*template.Template
	mux       *http.ServeMux
	mu        sync.RWMutex
}

// New создаёт и настраивает сервер приложения.
func New(cfg Config) (*App, error) {
	cat, err := LoadCatalog()
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки каталога: %w", err)
	}

	storePath := fmt.Sprintf("%s/store.json", cfg.DataDir)
	st, err := OpenStore(storePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия хранилища: %w", err)
	}

	a := &App{
		store:     st,
		catalog:   cat,
		assets:    cfg.Assets,
		templates: make(map[string]*template.Template),
		mux:       http.NewServeMux(),
	}

	if err := a.initTemplates(); err != nil {
		return nil, fmt.Errorf("ошибка компиляции шаблонов: %w", err)
	}

	a.setupRoutes()
	return a, nil
}

// Handler возвращает HTTP-обработчик.
func (a *App) Handler() http.Handler {
	return a.mux
}

func (a *App) templateFuncs() template.FuncMap {
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
		"goalLabel":     GoalLabel,
		"goalShort":     GoalShort,
		"levelLabel":    LevelLabel,
		"equipLabel":    EquipLabel,
		"activityLabel": ActivityLabel,
		"timeLabel":     TimeLabel,
		"paceLabel":     PaceLabel,
		"healthLabel":   HealthLabel,
		"todayIndex":    TodayIndex,
		"weekdayName": func(idx int) string {
			if idx >= 0 && idx < len(WeekdayNames) {
				return WeekdayNames[idx]
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
		"isSessionCompleted": func(sessions []SessionEntry, dayIndex int) bool {
			for _, s := range sessions {
				if s.DayIndex == dayIndex && s.Completed {
					return true
				}
			}
			return false
		},
		"userBMI": func(p Profile) map[string]any {
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
		"waterToday": func(u *User) int {
			if u == nil {
				return 0
			}
			return u.WaterToday()
		},
		"waterPercent": func(u *User) int {
			if u == nil {
				return 0
			}
			cur := u.WaterToday()
			target := u.Profile.WaterTargetMl()
			if target <= 0 {
				target = 2000
			}
			pct := cur * 100 / target
			if pct > 100 {
				return 100
			}
			return pct
		},
		"waterTarget": func(p Profile) int {
			return p.WaterTargetMl()
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

func (a *App) initTemplates() error {
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
		t := template.New("layout.html").Funcs(a.templateFuncs())
		var err error
		t, err = t.ParseFS(a.assets, "templates/layout.html", "templates/"+page+".html")
		if err != nil {
			return fmt.Errorf("шаблон %s: %w", page, err)
		}
		a.templates[page] = t
	}
	return nil
}

func (a *App) renderWithStatus(w http.ResponseWriter, page string, data any, status int) {
	t, ok := a.templates[page]
	if !ok {
		http.Error(w, fmt.Sprintf("Шаблон %s не найден", page), http.StatusInternalServerError)
		return
	}
	buf := new(bytes.Buffer)
	if err := t.ExecuteTemplate(buf, "layout.html", data); err != nil {
		log.Printf("ошибка рендеринга шаблона %s: %v", page, err)
		http.Error(w, "Ошибка формирования страницы", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status > 0 {
		w.WriteHeader(status)
	}
	_, _ = buf.WriteTo(w)
}

func (a *App) render(w http.ResponseWriter, page string, data any) {
	a.renderWithStatus(w, page, data, http.StatusOK)
}

func (a *App) render404(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	a.renderWithStatus(w, "404", PageData{
		ActiveTab: "dashboard",
		User:      u,
	}, http.StatusNotFound)
}

func (a *App) render500(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	a.renderWithStatus(w, "500", PageData{
		ActiveTab: "dashboard",
		User:      u,
	}, http.StatusInternalServerError)
}

// Close освобождает ресурсы приложения (закрывает SQLite).
func (a *App) Close() error {
	if a.store != nil {
		return a.store.Close()
	}
	return nil
}

// PageData передаётся в шаблоны.
type PageData struct {
	ActiveTab   string
	User        *User
	AI          map[string]string
	Error       string
	Notice      string
	Extra       any
	SelectedDay int
	TodayIdx    int
	WeightChart template.HTML
}

func (a *App) getOrCreateUser(w http.ResponseWriter, r *http.Request) *User {
	var uid string
	if cookie, err := r.Cookie("pulsefit_uid"); err == nil && cookie.Value != "" {
		uid = cookie.Value
	}

	if uid != "" {
		if u := a.store.User(uid); u != nil {
			return u
		}
	}

	// Если передан код в URL (?code=...)
	if code := r.URL.Query().Get("code"); code != "" {
		if u, ok := a.store.ByCode(code); ok {
			http.SetCookie(w, &http.Cookie{
				Name:     "pulsefit_uid",
				Value:    u.Profile.ID,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   365 * 24 * 3600,
			})
			return u
		}
	}

	// Создаём нового пользователя
	u, err := a.store.NewUser()
	if err != nil {
		log.Printf("ошибка создания пользователя: %v", err)
		return nil
	}

	// Инициализируем стандартные напоминания
	u.Reminders = []Reminder{
		{ID: "rem-1", Kind: "workout", Title: "Тренировка по плану", Time: "18:30", Days: "workouts", Enabled: true, Note: "Небольшая разминка и основная часть"},
		{ID: "rem-2", Kind: "meal", Title: "Время пообедать", Time: "13:00", Days: "daily", Enabled: true, Note: "Спокойный приём пищи без спешки"},
		{ID: "rem-3", Kind: "water", Title: "Стакан воды", Time: "11:00", Days: "daily", Enabled: true, Note: "Поддерживайте водный баланс"},
		{ID: "rem-4", Kind: "sleep", Title: "Подготовка ко сну", Time: "22:30", Days: "daily", Enabled: true, Note: "Отложите экран за 30 мин до сна"},
	}
	u.Progress.Meals = make(map[string]bool)
	u.Progress.Water = make(map[string]int)
	_ = a.store.Save(u)

	http.SetCookie(w, &http.Cookie{
		Name:     "pulsefit_uid",
		Value:    u.Profile.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   365 * 24 * 3600,
	})
	return u
}

func (a *App) setupRoutes() {
	// Статика
	staticSub, err := fs.Sub(a.assets, "static")
	if err == nil {
		a.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
	}

	a.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","app":"pulsefit"}`))
	})
	a.mux.HandleFunc("/", a.handleRoot)
	a.mux.HandleFunc("/manifest.json", a.handleManifest)
	a.mux.HandleFunc("/sw.js", a.handleServiceWorker)
	a.mux.HandleFunc("/onboarding", a.handleOnboarding)
	a.mux.HandleFunc("/medical-warning", a.handleMedicalWarning)
	a.mux.HandleFunc("/dashboard", a.handleDashboard)
	a.mux.HandleFunc("/workouts", a.handleWorkouts)
	a.mux.HandleFunc("/workouts/complete", a.handleWorkoutComplete)
	a.mux.HandleFunc("/workouts/swap", a.handleWorkoutSwap)
	a.mux.HandleFunc("/exercise", a.handleExerciseDetail)
	a.mux.HandleFunc("/nutrition", a.handleNutrition)
	a.mux.HandleFunc("/nutrition/toggle", a.handleNutritionToggle)
	a.mux.HandleFunc("/nutrition/swap", a.handleNutritionSwap)
	a.mux.HandleFunc("/progress", a.handleProgress)
	a.mux.HandleFunc("/progress/weight", a.handleProgressWeight)
	a.mux.HandleFunc("/progress/checkin", a.handleProgressCheckin)
	a.mux.HandleFunc("/water/add", a.handleWaterAdd)
	a.mux.HandleFunc("/catalog", a.handleCatalog)
	a.mux.HandleFunc("/plan/export", a.handlePlanExport)
	a.mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/catalog", http.StatusMovedPermanently)
	})
	a.mux.HandleFunc("/profile", a.handleProfile)
	a.mux.HandleFunc("/profile/switch", a.handleProfileSwitch)
	a.mux.HandleFunc("/reminders/update", a.handleRemindersUpdate)
	a.mux.HandleFunc("/plan/recreate", a.handlePlanRecreate)
	a.mux.HandleFunc("/plan/change-goal", a.handlePlanChangeGoal)
	a.mux.HandleFunc("/demo/seed", a.handleDemoSeed)
	a.mux.HandleFunc("/demo/reset", a.handleDemoReset)
}

func (a *App) handleManifest(w http.ResponseWriter, r *http.Request) {
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

func (a *App) handleServiceWorker(w http.ResponseWriter, r *http.Request) {
	sw := `const CACHE_NAME = 'pulsefit-v2.1-nocache';
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
  // Network first: всегда запрашивать актуальный контент
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

func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		a.render404(w, r)
		return
	}
	u := a.getOrCreateUser(w, r)
	if u == nil {
		http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
		return
	}

	if !u.Profile.Onboarded {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}
	if u.Profile.MedicalBlocked() {
		http.Redirect(w, r, "/medical-warning", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (a *App) handleOnboarding(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	if u == nil {
		http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		a.render(w, "onboarding", PageData{
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
		a.render(w, "onboarding", PageData{
			ActiveTab: "profile",
			User:      u,
						Error:     "Для формирования плана необходимо дать согласие на использование сведений для расчёта.",
		})
		return
	}

	p := u.Profile
	p.Name = strings.TrimSpace(r.FormValue("name"))
	p.Goal = r.FormValue("goal")
	if p.Goal != GoalGain && p.Goal != GoalLose {
		p.Goal = GoalLose
	}
	p.Pace = r.FormValue("pace")
	if p.Pace == "" {
		p.Pace = "moderate"
	}

	// Физические параметры (необязательные)
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
		if h, err := strconv.ParseFloat(hStr, 64); err == nil && h >= 100 && h <= 250 {
			p.HeightCm = &h
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
		p.Level = LevelBeginner
	}

	wpw, _ := strconv.Atoi(r.FormValue("workouts_per_week"))
	if wpw < 1 || wpw > 6 {
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
		p.Equipment = EquipNone
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

	// Проверка на медицинские противопоказания
	if p.MedicalBlocked() {
		u.Profile.Onboarded = false
		_ = a.store.Save(u)
		http.Redirect(w, r, "/medical-warning", http.StatusSeeOther)
		return
	}

	// Формируем план
	plan, err := a.catalog.BuildPlan(u.Profile)
	if err != nil {
		log.Printf("ошибка генерации плана: %v", err)
		a.render(w, "onboarding", PageData{
			ActiveTab: "profile",
			User:      u,
						Error:     "Не удалось составить план под указанные параметры. Пожалуйста, проверьте введённые данные.",
		})
		return
	}

	u.Plan = plan
	u.Profile.Onboarded = true
	u.Profile.PlanVersion = plan.Version

	// Если указан вес, запишем его в прогресс
	if p.WeightKg != nil {
		todayStr := time.Now().Format("2006-01-02")
		u.Progress.Weights = append(u.Progress.Weights, WeightEntry{
			Date: todayStr,
			Kg:   *p.WeightKg,
			Note: "Стартовый вес при заполнении анкеты",
		})
	}

	_ = a.store.Save(u)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (a *App) handleMedicalWarning(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	if u == nil {
		http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
		return
	}

	// Позволяем пользователю продолжить под свою ответственность при желании
	if r.Method == http.MethodPost && r.FormValue("proceed_anyway") == "1" {
		plan, err := a.catalog.BuildPlan(u.Profile)
		if err == nil {
			u.Plan = plan
			u.Profile.Onboarded = true
			_ = a.store.Save(u)
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	a.render(w, "medical_warning", PageData{
		ActiveTab: "profile",
		User:      u,
			})
}

func (a *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	if u == nil || !u.Profile.Onboarded || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	todayIdx := TodayIndex()
	var todayWorkout *PlanDay
	if todayIdx >= 0 && todayIdx < len(u.Plan.Days) {
		todayWorkout = &u.Plan.Days[todayIdx]
	}

	var todayMeals *MealDay
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
		TodayWorkout   *PlanDay
		TodayMeals     *MealDay
		CompletedCount int
		LastWeight     *float64
		WaterToday     int
		WaterTarget    int
		WaterPercent   int
	}

	var lastWeight *float64
	if len(u.Progress.Weights) > 0 {
		w := u.Progress.Weights[len(u.Progress.Weights)-1].Kg
		lastWeight = &w
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

	a.render(w, "dashboard", PageData{
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

func (a *App) handleWaterAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	u := a.getOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	delta, _ := strconv.Atoi(r.FormValue("delta"))
	today := time.Now().Format("2006-01-02")
	if u.Progress.Water == nil {
		u.Progress.Water = make(map[string]int)
	}

	cur := u.Progress.Water[today] + delta
	if cur < 0 {
		cur = 0
	}
	if cur > 10000 {
		cur = 10000
	}
	u.Progress.Water[today] = cur
	_ = a.store.Save(u)

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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"current": cur,
			"target":  target,
			"percent": pct,
		})
		return
	}

	http.Redirect(w, r, "/dashboard#water-section", http.StatusSeeOther)
}

func (a *App) handleCatalog(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	type catExtra struct {
		Exercises []Exercise
		Meals     []Meal
	}
	a.render(w, "catalog", PageData{
		ActiveTab: "catalog",
		User:      u,
				Extra: catExtra{
			Exercises: a.catalog.Exercises,
			Meals:     a.catalog.Meals,
		},
	})
}

func (a *App) handlePlanExport(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	if u == nil || !u.Profile.Onboarded || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}
	a.render(w, "plan_export", PageData{
		ActiveTab: "workouts",
		User:      u,
			})
}

func (a *App) handleWorkouts(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	if u == nil || !u.Profile.Onboarded || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	dayIdx := TodayIndex()
	if dStr := r.URL.Query().Get("day"); dStr != "" {
		if d, err := strconv.Atoi(dStr); err == nil && d >= 0 && d < 7 {
			dayIdx = d
		}
	}

	var curDay PlanDay
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
		CurrentDay  PlanDay
		IsCompleted bool
	}

	a.render(w, "workouts", PageData{
		ActiveTab:   "workouts",
		User:        u,
				SelectedDay: dayIdx,
		TodayIdx:    TodayIndex(),
		Notice:      r.URL.Query().Get("notice"),
		Extra: workoutsExtra{
			CurrentDay:  curDay,
			IsCompleted: isCompleted,
		},
	})
}

func (a *App) handleWorkoutComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := a.getOrCreateUser(w, r)
	if u == nil || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	dayIdx, _ := strconv.Atoi(r.FormValue("day"))
	if dayIdx < 0 || dayIdx >= 7 {
		dayIdx = TodayIndex()
	}

	energy, _ := strconv.Atoi(r.FormValue("energy"))
	if energy < 1 || energy > 5 {
		energy = 4
	}
	feeling := strings.TrimSpace(r.FormValue("feeling"))
	if feeling == "" {
		feeling = "Тренировка выполнена по плану"
	}

	title := "Тренировка"
	if dayIdx < len(u.Plan.Days) {
		title = u.Plan.Days[dayIdx].Title
	}

	entry := SessionEntry{
		Date:      time.Now().Format("2006-01-02"),
		DayIndex:  dayIdx,
		Title:     title,
		Completed: true,
		Feeling:   feeling,
		Energy:    energy,
	}

	u.Progress.Sessions = append(u.Progress.Sessions, entry)
	_ = a.store.Save(u)

	http.Redirect(w, r, fmt.Sprintf("/workouts?day=%d&notice=Тренировка+отмечена+выполненной!+Отличная+работа.", dayIdx), http.StatusSeeOther)
}

func (a *App) handleWorkoutSwap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := a.getOrCreateUser(w, r)
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

	day := &u.Plan.Days[dayIdx]
	exIdx := -1
	var targetEx PlanExercise
	for i, ex := range day.Exercises {
		if ex.ExerciseID == exID {
			exIdx = i
			targetEx = ex
			break
		}
	}

	if exIdx >= 0 {
		pool := a.catalog.SuitableExercises(u.Profile.Equipment, u.Profile.Level)
		var candidates []Exercise
		for _, c := range pool {
			if c.ID != exID && (c.Kind == targetEx.Kind || len(targetEx.Kind) == 0) {
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

		if len(candidates) > 0 {
			newEx := candidates[int(time.Now().UnixNano())%len(candidates)]
			day.Exercises[exIdx] = toPlanExercise(newEx)
			_ = a.store.Save(u)
			http.Redirect(w, r, fmt.Sprintf("/workouts?day=%d&notice=Упражнение+заменено+на+«%s»", dayIdx, newEx.Name), http.StatusSeeOther)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/workouts?day=%d", dayIdx), http.StatusSeeOther)
}

func (a *App) handleExerciseDetail(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	id := r.URL.Query().Get("id")
	dayStr := r.URL.Query().Get("day")
	dayIdx, _ := strconv.Atoi(dayStr)

	ex, ok := a.catalog.Exercise(id)
	if !ok {
		http.Error(w, "Упражнение не найдено", http.StatusNotFound)
		return
	}

	type exExtra struct {
		Exercise Exercise
		DayIndex int
	}

	a.render(w, "exercise_detail", PageData{
		ActiveTab:   "workouts",
		User:        u,
				SelectedDay: dayIdx,
		Extra: exExtra{
			Exercise: ex,
			DayIndex: dayIdx,
		},
	})
}

func (a *App) handleNutrition(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	if u == nil || !u.Profile.Onboarded || u.Plan == nil {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	dayIdx := TodayIndex()
	if dStr := r.URL.Query().Get("day"); dStr != "" {
		if d, err := strconv.Atoi(dStr); err == nil && d >= 0 && d < 7 {
			dayIdx = d
		}
	}

	var curMeals MealDay
	if len(u.Plan.Meals) > 0 {
		if dayIdx < 0 || dayIdx >= len(u.Plan.Meals) {
			dayIdx = 0
		}
		curMeals = u.Plan.Meals[dayIdx]
	}

	type nutrExtra struct {
		CurrentMeals MealDay
	}

	a.render(w, "nutrition", PageData{
		ActiveTab:   "nutrition",
		User:        u,
				SelectedDay: dayIdx,
		TodayIdx:    TodayIndex(),
		Notice:      r.URL.Query().Get("notice"),
		Extra: nutrExtra{
			CurrentMeals: curMeals,
		},
	})
}

func (a *App) handleNutritionToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	u := a.getOrCreateUser(w, r)
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

	if u.Progress.Meals == nil {
		u.Progress.Meals = make(map[string]bool)
	}

	key := fmt.Sprintf("%d:%s", dayIdx, mealID)
	u.Progress.Meals[key] = !u.Progress.Meals[key]
	_ = a.store.Save(u)

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		strings.Contains(r.Header.Get("Accept"), "application/json") ||
		r.FormValue("ajax") == "1" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":     true,
			"eaten":  u.Progress.Meals[key],
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

func (a *App) handleNutritionSwap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := a.getOrCreateUser(w, r)
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
	var targetMeal MealItem
	for i, m := range day.Meals {
		if m.MealID == mealID {
			mealIdx = i
			targetMeal = m
			break
		}
	}

	if mealIdx >= 0 {
		slots := []string{targetMeal.Slot}
		poolMap := a.catalog.SuitableMeals(slots, u.Profile.DietPreferences, u.Profile.Allergies, u.Profile.AvoidFoods)
		pool := poolMap[targetMeal.Slot]

		var candidates []Meal
		for _, m := range pool {
			if m.ID != mealID {
				candidates = append(candidates, m)
			}
		}

		if len(candidates) > 0 {
			newMeal := candidates[int(time.Now().UnixNano())%len(candidates)]
			day.Meals[mealIdx] = MealItem{
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
			_ = a.store.Save(u)
			http.Redirect(w, r, fmt.Sprintf("/nutrition?day=%d&notice=Блюдо+заменено+на+«%s»", dayIdx, newMeal.Name), http.StatusSeeOther)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/nutrition?day=%d", dayIdx), http.StatusSeeOther)
}

func (a *App) handleProgress(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	if u == nil || !u.Profile.Onboarded {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	chartSVG := generateWeightSVG(u.Progress.Weights)

	a.render(w, "progress", PageData{
		ActiveTab:   "progress",
		User:        u,
				Notice:      r.URL.Query().Get("notice"),
		WeightChart: template.HTML(chartSVG),
	})
}

func (a *App) handleProgressWeight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := a.getOrCreateUser(w, r)
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

	u.Progress.Weights = append(u.Progress.Weights, WeightEntry{
		Date: dateStr,
		Kg:   wt,
		Note: note,
	})
	u.Profile.WeightKg = &wt
	_ = a.store.Save(u)

	http.Redirect(w, r, "/progress?notice=Запись+веса+сохранена!", http.StatusSeeOther)
}

func (a *App) handleProgressCheckin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := a.getOrCreateUser(w, r)
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

	entry := Checkin{
		Date:      time.Now().Format("2006-01-02"),
		Mood:      mood,
		Energy:    energy,
		Nutrition: nutrition,
		Sleep:     sleep,
		Note:      note,
	}

	u.Progress.Checkins = append(u.Progress.Checkins, entry)
	_ = a.store.Save(u)

	http.Redirect(w, r, "/progress?notice=Чек-ин+самочувствия+сохранён!", http.StatusSeeOther)
}



func (a *App) handleProfile(w http.ResponseWriter, r *http.Request) {
	u := a.getOrCreateUser(w, r)
	if u == nil {
		http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		a.render(w, "profile", PageData{
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
		if h, err := strconv.ParseFloat(hStr, 64); err == nil && h >= 100 && h <= 250 {
			p.HeightCm = &h
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
	if wpw >= 1 && wpw <= 6 {
		p.WorkoutsPerWeek = wpw
	}
	dur, _ := strconv.Atoi(r.FormValue("session_minutes"))
	if dur > 0 {
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

	u.Profile = p
	_ = a.store.Save(u)

	http.Redirect(w, r, "/profile?notice=Данные+профиля+успешно+обновлены", http.StatusSeeOther)
}

func (a *App) handleProfileSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	code := strings.ToUpper(strings.TrimSpace(r.FormValue("code")))
	if code == "" {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	if targetUser, ok := a.store.ByCode(code); ok {
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

func (a *App) handleRemindersUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := a.getOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	_ = r.ParseForm()
	for i := range u.Reminders {
		rem := &u.Reminders[i]
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

	_ = a.store.Save(u)
	http.Redirect(w, r, "/profile?notice=Настройки+напоминаний+сохранены", http.StatusSeeOther)
}

func (a *App) handlePlanRecreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := a.getOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if u.Profile.MedicalBlocked() {
		http.Redirect(w, r, "/medical-warning", http.StatusSeeOther)
		return
	}

	plan, err := a.catalog.BuildPlan(u.Profile)
	if err != nil {
		log.Printf("ошибка пересоздания плана: %v", err)
		http.Redirect(w, r, "/profile?notice=Не+удалось+пересоздать+план", http.StatusSeeOther)
		return
	}

	u.Plan = plan
	u.Profile.PlanVersion = plan.Version
	_ = a.store.Save(u)

	http.Redirect(w, r, "/dashboard?notice=Персональный+план+успешно+обновлён!", http.StatusSeeOther)
}

func (a *App) handlePlanChangeGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := a.getOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	newGoal := r.FormValue("goal")
	if newGoal != GoalGain && newGoal != GoalLose {
		if u.Profile.Goal == GoalLose {
			newGoal = GoalGain
		} else {
			newGoal = GoalLose
		}
	}

	u.Profile.Goal = newGoal
	plan, err := a.catalog.BuildPlan(u.Profile)
	if err == nil {
		u.Plan = plan
		u.Profile.PlanVersion = plan.Version
	}
	_ = a.store.Save(u)

	http.Redirect(w, r, "/dashboard?notice=Цель+изменена+на+«"+GoalLabel(newGoal)+"»!", http.StatusSeeOther)
}

func (a *App) handleDemoSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := a.getOrCreateUser(w, r)
	if u == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	age := 28
	height := 176.0
	weight := 78.5
	u.Profile.Name = "Алексей"
	u.Profile.Age = &age
	u.Profile.HeightCm = &height
	u.Profile.WeightKg = &weight
	u.Profile.Sex = "male"
	u.Profile.Goal = GoalLose
	u.Profile.Pace = "moderate"
	u.Profile.Level = LevelBeginner
	u.Profile.WorkoutsPerWeek = 3
	u.Profile.SessionMinutes = 35
	u.Profile.Equipment = EquipHome
	u.Profile.PreferredTime = "evening"
	u.Profile.ActivityLevel = "medium"
	u.Profile.DietPreferences = []string{}
	u.Profile.Allergies = []string{}
	u.Profile.HealthFlags = []string{HealthNone}
	u.Profile.Consent = true
	u.Profile.Onboarded = true

	plan, _ := a.catalog.BuildPlan(u.Profile)
	u.Plan = plan

	now := time.Now()
	u.Progress.Weights = []WeightEntry{
		{Date: now.AddDate(0, 0, -21).Format("2006-01-02"), Kg: 80.2, Note: "Старт программы"},
		{Date: now.AddDate(0, 0, -14).Format("2006-01-02"), Kg: 79.5, Note: "Чувствую лёгкость"},
		{Date: now.AddDate(0, 0, -7).Format("2006-01-02"), Kg: 78.9, Note: "Хороший темп"},
		{Date: now.Format("2006-01-02"), Kg: 78.5, Note: "Сегодняшнее взвешивание"},
	}

	u.Progress.Sessions = []SessionEntry{
		{Date: now.AddDate(0, 0, -4).Format("2006-01-02"), DayIndex: 0, Title: "Понедельник", Completed: true, Feeling: "Отличная разминка, техника понятна", Energy: 5},
		{Date: now.AddDate(0, 0, -2).Format("2006-01-02"), DayIndex: 2, Title: "Среда", Completed: true, Feeling: "Мышцы слегка ноют, но самочувствие бодрое", Energy: 4},
	}

	u.Progress.Checkins = []Checkin{
		{Date: now.AddDate(0, 0, -1).Format("2006-01-02"), Mood: 5, Energy: 4, Nutrition: 4, Sleep: 8.0, Note: "Выпил норму воды, спал крепко"},
	}

	if u.Progress.Meals == nil {
		u.Progress.Meals = make(map[string]bool)
	}
	todayIdx := TodayIndex()
	if plan != nil && todayIdx < len(plan.Meals) && len(plan.Meals[todayIdx].Meals) > 0 {
		u.Progress.Meals[fmt.Sprintf("%d:%s", todayIdx, plan.Meals[todayIdx].Meals[0].MealID)] = true
	}

	if u.Progress.Water == nil {
		u.Progress.Water = make(map[string]int)
	}
	u.Progress.Water[now.Format("2006-01-02")] = 1750



	_ = a.store.Save(u)
	http.Redirect(w, r, "/dashboard?notice=Загружен+демонстрационный+сценарий+с+заполненным+прогрессом!", http.StatusSeeOther)
}

func (a *App) handleDemoReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	u := a.getOrCreateUser(w, r)
	if u != nil {
		_ = a.store.Delete(u.Profile.ID)
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
func generateWeightSVG(entries []WeightEntry) string {
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
