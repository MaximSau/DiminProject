package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMedicalBlocked(t *testing.T) {
	// 1. Без противопоказаний
	pSafe := Profile{
		HealthFlags: []string{HealthNone},
	}
	if pSafe.MedicalBlocked() {
		t.Errorf("ожидалось отсутствие блокировки для безопасного профиля")
	}

	// 2. С травмой
	pInjury := Profile{
		HealthFlags: []string{HealthInjury},
	}
	if !pInjury.MedicalBlocked() {
		t.Errorf("ожидалась медицинская блокировка при указании травмы")
	}

	// 3. С сердечно-сосудистыми ограничениями
	pHeart := Profile{
		HealthFlags: []string{HealthHeart},
	}
	if !pHeart.MedicalBlocked() {
		t.Errorf("ожидалась медицинская блокировка при указании болезней сердца")
	}
}

func TestBuildPlan(t *testing.T) {
	cat, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}

	// План для похудения, новичок, дома
	age := 30
	h := 175.0
	w := 80.0
	p := Profile{
		Goal:            GoalLose,
		Pace:            "moderate",
		Level:           LevelBeginner,
		WorkoutsPerWeek: 3,
		SessionMinutes:  35,
		Equipment:       EquipHome,
		Age:             &age,
		HeightCm:        &h,
		WeightKg:        &w,
		Sex:             "male",
	}

	plan, err := cat.BuildPlan(p)
	if err != nil {
		t.Fatalf("BuildPlan failed: %v", err)
	}

	if plan.Goal != GoalLose {
		t.Errorf("ожидалась цель %s, получено %s", GoalLose, plan.Goal)
	}
	if len(plan.Days) != 7 {
		t.Errorf("ожидалось 7 дней в неделе, получено %d", len(plan.Days))
	}
	if len(plan.Meals) != 7 {
		t.Errorf("ожидалось 7 дней питания, получено %d", len(plan.Meals))
	}
	if plan.KcalTarget == nil || *plan.KcalTarget <= 0 {
		t.Errorf("ожидался рассчитанный ориентир калорий")
	}

	// Проверяем количество тренировочных дней
	trainingDays := 0
	for _, d := range plan.Days {
		if !d.Rest {
			trainingDays++
			if len(d.Exercises) == 0 {
				t.Errorf("тренировочный день %s не содержит упражнений", d.Title)
			}
		}
	}
	if trainingDays != 3 {
		t.Errorf("ожидалось 3 тренировочных дня, получено %d", trainingDays)
	}
}

func TestBMICalculation(t *testing.T) {
	// Случай 1: Нормальный вес
	h1 := 180.0
	w1 := 72.0
	pNorm := Profile{HeightCm: &h1, WeightKg: &w1}
	bmi, label, status := pNorm.BMI()
	if bmi < 22.0 || bmi > 22.5 {
		t.Errorf("ожидался ИМТ ~22.2, получено %.2f", bmi)
	}
	if status != "normal" || !strings.Contains(label, "Нормальный") {
		t.Errorf("ожидался статус normal, получено %s (%s)", status, label)
	}

	// Случай 2: Дефицит массы
	wUnder := 50.0
	pUnder := Profile{HeightCm: &h1, WeightKg: &wUnder}
	_, _, statusUnder := pUnder.BMI()
	if statusUnder != "underweight" {
		t.Errorf("ожидался статус underweight, получено %s", statusUnder)
	}

	// Случай 3: Избыточная масса
	wOver := 88.0
	pOver := Profile{HeightCm: &h1, WeightKg: &wOver}
	_, _, statusOver := pOver.BMI()
	if statusOver != "overweight" {
		t.Errorf("ожидался статус overweight, получено %s", statusOver)
	}

	// Случай 4: Ожирение
	wObese := 110.0
	pObese := Profile{HeightCm: &h1, WeightKg: &wObese}
	_, _, statusObese := pObese.BMI()
	if statusObese != "obese" {
		t.Errorf("ожидался статус obese, получено %s", statusObese)
	}

	// Случай 5: Пустые параметры (nil)
	pEmpty := Profile{}
	bmiZero, _, _ := pEmpty.BMI()
	if bmiZero != 0 {
		t.Errorf("ожидался 0 для пустого профиля, получено %.2f", bmiZero)
	}

	// Проверка IdealWeightRange
	minW, maxW := pNorm.IdealWeightRange()
	if minW < 59 || minW > 61 {
		t.Errorf("ожидался minW ~60 кг, получено %.1f", minW)
	}
	if maxW < 80 || maxW > 82 {
		t.Errorf("ожидался maxW ~80.7 кг, получено %.1f", maxW)
	}
}

func TestWaterTargetCalculation(t *testing.T) {
	// Вес 70 кг -> 70 * 30-35 мл = 2100 - 2450 мл
	w := 70.0
	p := Profile{WeightKg: &w}
	target := p.WaterTargetMl()
	if target < 2000 || target > 2500 {
		t.Errorf("ожидалась норма воды в диапазоне 2000-2500 мл, получено %d", target)
	}

	// Без указания веса — дефолт 2000 мл
	pEmpty := Profile{}
	if pEmpty.WaterTargetMl() != 2000 {
		t.Errorf("ожидался дефолт 2000 мл для пустого профиля, получено %d", pEmpty.WaterTargetMl())
	}
}

func TestEstimateNutritionSafetyFloor(t *testing.T) {
	age := 25
	h := 160.0
	w := 50.0
	p := Profile{
		Sex:           "female",
		Age:           &age,
		HeightCm:      &h,
		WeightKg:      &w,
		Goal:          GoalLose,
		Pace:          "steady",
		ActivityLevel: "sedentary",
	}

	nut := EstimateNutrition(p)
	if nut.Kcal < 1200 {
		t.Errorf("ожидалась защита нижней границы калорий (floor >= 1200), получено %d", nut.Kcal)
	}
	if nut.ProteinG <= 0 || nut.FatG <= 0 || nut.CarbsG <= 0 {
		t.Errorf("ожидались положительные значения БЖУ, получено %+v", nut)
	}

	pEmpty := Profile{}
	nutEmpty := EstimateNutrition(pEmpty)
	if nutEmpty.Kcal != 0 {
		t.Errorf("ожидались нулевые калории для пустого профиля, получено %d", nutEmpty.Kcal)
	}
}

func TestCatalogSwapAndSearch(t *testing.T) {
	cat, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}

	if len(cat.Exercises) == 0 {
		t.Fatalf("каталог упражнений пуст")
	}
	firstEx := cat.Exercises[0]
	foundEx, ok := cat.Exercise(firstEx.ID)
	if !ok || foundEx.ID != firstEx.ID {
		t.Errorf("Exercise lookup failed for id %s", firstEx.ID)
	}

	suitable := cat.SuitableExercises(EquipHome, LevelBeginner)
	if len(suitable) == 0 {
		t.Errorf("SuitableExercises returned 0 items for home/beginner")
	}

	if len(cat.Meals) == 0 {
		t.Fatalf("каталог блюд пуст")
	}
	firstMeal := cat.Meals[0]
	foundMeal, ok := cat.Meal(firstMeal.ID)
	if !ok || foundMeal.ID != firstMeal.ID {
		t.Errorf("Meal lookup failed for id %s", firstMeal.ID)
	}

	mealsMap := cat.SuitableMeals([]string{"breakfast", "lunch"}, []string{"vegetarian"}, nil, "")
	if len(mealsMap) == 0 {
		t.Errorf("SuitableMeals failed to return meals")
	}
}

func TestSQLitePersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pulsefit_sqlite_*")
	if err != nil {
		t.Fatalf("temp dir failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := tmpDir + "/test_store.db"
	st, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("OpenStore failed: %v", err)
	}

	u, err := st.NewUser()
	if err != nil {
		t.Fatalf("NewUser failed: %v", err)
	}

	age := 27
	h := 180.0
	w := 76.5
	u.Profile.Name = "Дмитрий"
	u.Profile.Age = &age
	u.Profile.HeightCm = &h
	u.Profile.WeightKg = &w
	u.Profile.Goal = GoalGain
	u.Profile.Onboarded = true

	u.Progress.Weights = append(u.Progress.Weights, WeightEntry{
		Date: "2026-11-01",
		Kg:   76.0,
		Note: "Старт",
	})
	u.Progress.Meals["0:oatmeal-banana"] = true

	if err := st.Save(u); err != nil {
		t.Fatalf("Save user failed: %v", err)
	}
	st.Close()

	st2, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("Re-open store failed: %v", err)
	}
	defer st2.Close()

	loaded := st2.User(u.Profile.ID)
	if loaded == nil {
		t.Fatalf("User not found by ID")
	}
	if loaded.Profile.Name != "Дмитрий" {
		t.Errorf("expected name Дмитрий, got %s", loaded.Profile.Name)
	}
	if loaded.Profile.WeightKg == nil || *loaded.Profile.WeightKg != 76.5 {
		t.Errorf("expected weight 76.5")
	}
	if len(loaded.Progress.Weights) != 1 || loaded.Progress.Weights[0].Kg != 76.0 {
		t.Errorf("expected 1 weight entry of 76.0")
	}
	if !loaded.Progress.Meals["0:oatmeal-banana"] {
		t.Errorf("expected eaten meal flag to persist in SQLite")
	}

	byCode, ok := st2.ByCode(u.Profile.Code)
	if !ok || byCode.Profile.ID != u.Profile.ID {
		t.Errorf("expected to find user by unique code %s", u.Profile.Code)
	}
}

func setupTestApp(t *testing.T) (*App, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "pulsefit_e2e_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	assets := os.DirFS("../../web")
	appInstance, err := New(Config{
		DataDir: tmpDir,
		Assets:  assets,
	})
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to create test app: %v", err)
	}

	return appInstance, tmpDir
}

func TestHTTPRoutesAndFlow(t *testing.T) {
	appInstance, tmpDir := setupTestApp(t)
	defer os.RemoveAll(tmpDir)
	defer appInstance.Close()

	handler := appInstance.Middleware(appInstance.Handler())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar failed: %v", err)
	}
	client := &http.Client{Jar: jar}

	// 1. GET /healthz
	resp, err := client.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for /healthz, got %d", resp.StatusCode)
	}
	var health map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&health)
	resp.Body.Close()
	if health["status"] != "ok" || health["app"] != "pulsefit" {
		t.Errorf("unexpected healthz response: %v", health)
	}

	// 2. GET /manifest.json
	resp, err = client.Get(ts.URL + "/manifest.json")
	if err != nil {
		t.Fatalf("GET /manifest.json failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for /manifest.json, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 3. GET /sw.js
	resp, err = client.Get(ts.URL + "/sw.js")
	if err != nil {
		t.Fatalf("GET /sw.js failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for /sw.js, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 4. GET /non-existent-page -> 404
	resp, err = client.Get(ts.URL + "/non-existent-page-404")
	if err != nil {
		t.Fatalf("GET /non-existent-page failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for unknown route, got %d", resp.StatusCode)
	}
	body404, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body404), "Страница не найдена") {
		t.Errorf("expected custom 404 page content")
	}

	// 5. POST /demo/seed
	resp, err = client.Post(ts.URL+"/demo/seed", "application/x-www-form-urlencoded", nil)
	if err != nil {
		t.Fatalf("POST /demo/seed failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after following redirect on demo/seed, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 6. GET /dashboard
	resp, err = client.Get(ts.URL + "/dashboard")
	if err != nil {
		t.Fatalf("GET /dashboard failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for /dashboard, got %d", resp.StatusCode)
	}
	dashBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(dashBody), "PulseFit") {
		t.Errorf("dashboard should contain brand name PulseFit")
	}

	// 7. POST /water/add via AJAX
	waterReq, _ := http.NewRequest("POST", ts.URL+"/water/add", strings.NewReader("delta=250"))
	waterReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	waterReq.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err = client.Do(waterReq)
	if err != nil {
		t.Fatalf("POST /water/add failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for AJAX water add, got %d", resp.StatusCode)
	}
	var waterResp map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&waterResp)
	resp.Body.Close()
	if waterResp["ok"] != true || waterResp["current"].(float64) < 250 {
		t.Errorf("unexpected water add AJAX response: %v", waterResp)
	}

	// 8. POST /nutrition/toggle via AJAX
	mealReq, _ := http.NewRequest("POST", ts.URL+"/nutrition/toggle", strings.NewReader("day=0&meal_id=test-meal"))
	mealReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mealReq.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err = client.Do(mealReq)
	if err != nil {
		t.Fatalf("POST /nutrition/toggle failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for AJAX nutrition toggle, got %d", resp.StatusCode)
	}
	var mealResp map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&mealResp)
	resp.Body.Close()
	if mealResp["ok"] != true || mealResp["eaten"] != true {
		t.Errorf("unexpected meal toggle response: %v", mealResp)
	}

	// 9. GET /catalog
	resp, err = client.Get(ts.URL + "/catalog")
	if err != nil {
		t.Fatalf("GET /catalog failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for /catalog, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 10. GET /plan/export
	resp, err = client.Get(ts.URL + "/plan/export")
	if err != nil {
		t.Fatalf("GET /plan/export failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for /plan/export, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 11. POST /progress/weight
	form := url.Values{"weight_kg": {"75.8"}, "date": {"2026-11-09"}, "note": {"тест"}}
	resp, err = client.PostForm(ts.URL+"/progress/weight", form)
	if err != nil {
		t.Fatalf("POST /progress/weight failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 following redirect for /progress/weight, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 12. GET /chat -> redirects to /catalog (База знаний и FAQ)
	resp, err = client.Get(ts.URL+"/chat")
	if err != nil {
		t.Fatalf("GET /chat failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 following redirect to /catalog for /chat, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestOnboardingAndPlanRecreateFlow(t *testing.T) {
	appInstance, tmpDir := setupTestApp(t)
	defer os.RemoveAll(tmpDir)
	defer appInstance.Close()

	handler := appInstance.Middleware(appInstance.Handler())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// 1. Отправка анкеты с противопоказаниями -> редирект на /medical-warning
	medicalForm := url.Values{
		"name":         {"Алексей"},
		"goal":         {GoalLose},
		"health_flags": {HealthInjury},
		"consent":      {"1"},
	}
	resp, err := client.PostForm(ts.URL+"/onboarding", medicalForm)
	if err != nil {
		t.Fatalf("POST /onboarding medical failed: %v", err)
	}
	if !strings.Contains(resp.Request.URL.Path, "medical-warning") {
		t.Errorf("expected redirect to medical-warning, got %s", resp.Request.URL.Path)
	}
	resp.Body.Close()

	// 2. Отправка здоровой анкеты -> создание плана и переход на /dashboard
	healthyForm := url.Values{
		"name":               {"Алексей"},
		"goal":               {GoalLose},
		"pace":               {"moderate"},
		"level":              {LevelBeginner},
		"equipment":          {EquipHome},
		"workouts_per_week":  {"3"},
		"session_minutes":    {"30"},
		"age":                {"28"},
		"height_cm":          {"178"},
		"weight_kg":          {"82.5"},
		"sex":                {"male"},
		"activity_level":     {"medium"},
		"preferred_time":     {TimeEvening},
		"health_flags":       {HealthNone},
		"consent":            {"1"},
	}
	resp, err = client.PostForm(ts.URL+"/onboarding", healthyForm)
	if err != nil {
		t.Fatalf("POST /onboarding healthy failed: %v", err)
	}
	if !strings.Contains(resp.Request.URL.Path, "dashboard") {
		t.Errorf("expected redirect to dashboard, got %s", resp.Request.URL.Path)
	}
	resp.Body.Close()

	// 3. Замена упражнения в плане
	swapExForm := url.Values{"day": {"0"}, "exercise_id": {"ex-1"}}
	resp, err = client.PostForm(ts.URL+"/workouts/swap", swapExForm)
	if err != nil {
		t.Fatalf("POST /workouts/swap failed: %v", err)
	}
	resp.Body.Close()

	// 4. Замена блюда в меню
	swapMealForm := url.Values{"day": {"0"}, "meal_id": {"m1"}}
	resp, err = client.PostForm(ts.URL+"/nutrition/swap", swapMealForm)
	if err != nil {
		t.Fatalf("POST /nutrition/swap failed: %v", err)
	}
	resp.Body.Close()

	// 5. Переключение цели (lose -> gain)
	resp, err = client.PostForm(ts.URL+"/plan/change-goal", url.Values{"goal": {GoalGain}})
	if err != nil {
		t.Fatalf("POST /plan/change-goal failed: %v", err)
	}
	resp.Body.Close()

	// 6. Пересоздание плана
	resp, err = client.PostForm(ts.URL+"/plan/recreate", nil)
	if err != nil {
		t.Fatalf("POST /plan/recreate failed: %v", err)
	}
	resp.Body.Close()

	// 7. Обновление напоминаний
	remForm := url.Values{
		"enabled_rem-1": {"1"},
		"time_rem-1":    {"19:00"},
	}
	resp, err = client.PostForm(ts.URL+"/reminders/update", remForm)
	if err != nil {
		t.Fatalf("POST /reminders/update failed: %v", err)
	}
	resp.Body.Close()
}

func TestSecurityHeadersAndRequestID(t *testing.T) {
	appInstance, tmpDir := setupTestApp(t)
	defer os.RemoveAll(tmpDir)
	defer appInstance.Close()

	handler := appInstance.Middleware(appInstance.Handler())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("missing X-Content-Type-Options")
	}
	if resp.Header.Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("missing X-Frame-Options")
	}
	if resp.Header.Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Errorf("missing Referrer-Policy")
	}
	if resp.Header.Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("missing X-XSS-Protection")
	}
	if resp.Header.Get("Content-Security-Policy") == "" {
		t.Errorf("missing Content-Security-Policy")
	}
	if resp.Header.Get("X-Request-ID") == "" {
		t.Errorf("missing X-Request-ID header")
	}
}

func TestPanicRecovery(t *testing.T) {
	appInstance, tmpDir := setupTestApp(t)
	defer os.RemoveAll(tmpDir)
	defer appInstance.Close()

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("тестовая паника для проверки recovery middleware")
	})

	wrapped := appInstance.Middleware(panicHandler)
	ts := httptest.NewServer(wrapped)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/cause-panic")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 status on panic, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Что-то пошло не так") {
		t.Errorf("expected custom 500 error page, got %s", string(body))
	}
}

func TestConcurrentWritesAndNoDeadlock(t *testing.T) {
	appInstance, tmpDir := setupTestApp(t)
	defer os.RemoveAll(tmpDir)
	defer appInstance.Close()

	handler := appInstance.Middleware(appInstance.Handler())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	u, err := appInstance.store.NewUser()
	if err != nil {
		t.Fatalf("NewUser failed: %v", err)
	}
	u.Profile.Onboarded = true
	_ = appInstance.store.Save(u)

	const workers = 20
	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}

			req, err := http.NewRequest("POST", ts.URL+"/water/add", strings.NewReader("delta=250&ajax=1"))
			if err != nil {
				errCh <- err
				return
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("X-Requested-With", "XMLHttpRequest")
			req.AddCookie(&http.Cookie{Name: "pulsefit_uid", Value: u.Profile.ID})
			resp, err := client.Do(req)
			if err != nil {
				errCh <- fmt.Errorf("water worker %d: %w", workerID, err)
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("water worker %d status %d", workerID, resp.StatusCode)
				return
			}

			mealReq, _ := http.NewRequest("POST", ts.URL+"/nutrition/toggle", strings.NewReader(fmt.Sprintf("day=0&meal_id=meal-%d", workerID)))
			mealReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			mealReq.Header.Set("X-Requested-With", "XMLHttpRequest")
			mealReq.AddCookie(&http.Cookie{Name: "pulsefit_uid", Value: u.Profile.ID})
			mResp, err := client.Do(mealReq)
			if err != nil {
				errCh <- fmt.Errorf("meal worker %d: %w", workerID, err)
				return
			}
			mResp.Body.Close()
			if mResp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("meal worker %d status %d", workerID, mResp.StatusCode)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent test failure: %v", err)
	}

	reloaded := appInstance.store.User(u.Profile.ID)
	if reloaded == nil {
		t.Fatalf("failed to reload user after concurrent test")
	}
	expectedWater := workers * 250
	if reloaded.WaterToday() != expectedWater {
		t.Errorf("expected water %d, got %d", expectedWater, reloaded.WaterToday())
	}
}

func TestStoreCountAndDelete(t *testing.T) {
	appInstance, tmpDir := setupTestApp(t)
	defer os.RemoveAll(tmpDir)
	defer appInstance.Close()

	initialCount := appInstance.store.Count()
	u1, err := appInstance.store.NewUser()
	if err != nil {
		t.Fatalf("NewUser failed: %v", err)
	}

	if count := appInstance.store.Count(); count != initialCount+1 {
		t.Errorf("expected count %d, got %d", initialCount+1, count)
	}

	// Find by code
	byCode, ok := appInstance.store.ByCode(u1.Profile.Code)
	if !ok || byCode == nil || byCode.Profile.ID != u1.Profile.ID {
		t.Errorf("ByCode failed to find user with code %s", u1.Profile.Code)
	}

	// Find non-existent code
	_, ok = appInstance.store.ByCode("NONEXISTINGCODE")
	if ok {
		t.Errorf("ByCode should return false for unknown code")
	}

	// Delete
	if err := appInstance.store.Delete(u1.Profile.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if count := appInstance.store.Count(); count != initialCount {
		t.Errorf("expected count %d after delete, got %d", initialCount, count)
	}

	if appInstance.store.User(u1.Profile.ID) != nil {
		t.Errorf("user still exists after Delete")
	}
}
