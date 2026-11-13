package routes

import (
	"io/fs"
	"net/http"

	transport "pulsefit/internal/transport/http"
)

// RegisterRoutes регистрирует все HTTP-эндпоинты и статическую раздачу на ServeMux.
func RegisterRoutes(mux *http.ServeMux, h *transport.Handler, assets fs.FS) {
	// Раздача статических файлов (CSS, SVG и т.д.)
	if staticSub, err := fs.Sub(assets, "static"); err == nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
	}

	// Системные и PWA эндпоинты
	mux.HandleFunc("/healthz", h.HandleHealthz)
	mux.HandleFunc("/manifest.json", h.HandleManifest)
	mux.HandleFunc("/sw.js", h.HandleServiceWorker)

	// Основные страницы и онбординг
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
