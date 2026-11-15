package routes

import (
	"io/fs"
	"net/http"

	transport "pulsefit/internal/transport/http"
)

// RegisterRoutes регистрирует все HTTP-эндпоинты, REST API и статическую раздачу на ServeMux.
func RegisterRoutes(mux *http.ServeMux, h *transport.Handler, assets fs.FS) {
	// Раздача статических файлов (CSS, SVG и т.д.)
	if staticSub, err := fs.Sub(assets, "static"); err == nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
	}

	// Раздача скомпилированных ассетов React SPA (Vite / React dist)
	if distSub, err := fs.Sub(assets, "dist"); err == nil {
		if assetsSub, err := fs.Sub(distSub, "assets"); err == nil {
			mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsSub))))
		}
	}

	// Системные и PWA эндпоинты
	mux.HandleFunc("/healthz", h.HandleHealthz)
	mux.HandleFunc("/manifest.json", h.HandleManifest)
	mux.HandleFunc("/sw.js", h.HandleServiceWorker)

	// REST API для React SPA фронтенда
	mux.HandleFunc("/api/bootstrap", h.HandleAPIBootstrap)
	mux.HandleFunc("/api/onboarding", h.HandleAPIOnboarding)
	mux.HandleFunc("/api/dashboard", h.HandleAPIDashboard)
	mux.HandleFunc("/api/workouts", h.HandleAPIWorkouts)
	mux.HandleFunc("/api/workouts/complete", h.HandleAPIWorkoutComplete)
	mux.HandleFunc("/api/workouts/swap", h.HandleAPIWorkoutSwap)
	mux.HandleFunc("/api/workouts/reschedule", h.HandleAPIWorkoutReschedule)
	mux.HandleFunc("/api/workouts/note", h.HandleAPIWorkoutNote)
	mux.HandleFunc("/api/nutrition", h.HandleAPINutrition)
	mux.HandleFunc("/api/nutrition/toggle", h.HandleAPINutritionToggle)
	mux.HandleFunc("/api/nutrition/add", h.HandleAPINutritionAdd)
	mux.HandleFunc("/api/water/add", h.HandleAPIWaterAdd)
	mux.HandleFunc("/api/progress", h.HandleAPIProgress)
	mux.HandleFunc("/api/progress/weight", h.HandleAPIProgressWeight)
	mux.HandleFunc("/api/catalog", h.HandleAPICatalog)
	mux.HandleFunc("/api/catalog/exercise", h.HandleAPICatalogExercise)
	mux.HandleFunc("/api/catalog/meals", h.HandleAPICatalogMeals)
	mux.HandleFunc("/api/profile", h.HandleAPIProfile)
	mux.HandleFunc("/api/plan/export", h.HandleAPIPlanExport)
	mux.HandleFunc("/api/demo/seed", h.HandleAPIDemoSeed)
	mux.HandleFunc("/api/demo/reset", h.HandleAPIDemoReset)

	// Маршрут прямого доступа к React SPA
	mux.HandleFunc("/spa", h.HandleSPA)
	mux.HandleFunc("/spa/", h.HandleSPA)
	mux.HandleFunc("/app", h.HandleSPA)
	mux.HandleFunc("/app/", h.HandleSPA)

	// Основные страницы и онбординг (SSR)
	mux.HandleFunc("/", h.HandleRoot)
	mux.HandleFunc("/onboarding", h.HandleOnboarding)
	mux.HandleFunc("/medical-warning", h.HandleMedicalWarning)
	mux.HandleFunc("/dashboard", h.HandleDashboard)

	// Тренировки и упражнения
	mux.HandleFunc("/workouts", h.HandleWorkouts)
	mux.HandleFunc("/workouts/complete", h.HandleWorkoutComplete)
	mux.HandleFunc("/workouts/swap", h.HandleWorkoutSwap)
	mux.HandleFunc("/exercise", h.HandleExerciseDetail)

	// Питание и гидратация
	mux.HandleFunc("/nutrition", h.HandleNutrition)
	mux.HandleFunc("/nutrition/toggle", h.HandleNutritionToggle)
	mux.HandleFunc("/nutrition/swap", h.HandleNutritionSwap)
	mux.HandleFunc("/water/add", h.HandleWaterAdd)

	// Прогресс, вес и самочувствие
	mux.HandleFunc("/progress", h.HandleProgress)
	mux.HandleFunc("/progress/weight", h.HandleProgressWeight)
	mux.HandleFunc("/progress/checkin", h.HandleProgressCheckin)

	// Каталог, база знаний и экспорт
	mux.HandleFunc("/catalog", h.HandleCatalog)
	mux.HandleFunc("/plan/export", h.HandlePlanExport)
	mux.HandleFunc("/chat", h.HandleChat)

	// Профиль и настройки
	mux.HandleFunc("/profile", h.HandleProfile)
	mux.HandleFunc("/profile/switch", h.HandleProfileSwitch)
	mux.HandleFunc("/reminders/update", h.HandleRemindersUpdate)
	mux.HandleFunc("/plan/recreate", h.HandlePlanRecreate)
	mux.HandleFunc("/plan/change-goal", h.HandlePlanChangeGoal)

	// Демонстрационный сценарий
	mux.HandleFunc("/demo/seed", h.HandleDemoSeed)
	mux.HandleFunc("/demo/reset", h.HandleDemoReset)
}
