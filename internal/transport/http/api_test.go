package http_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"testing"

	"pulsefit/internal/app"
)

func TestRESTAPIEndpoints(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pulsefit_api_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	assets := os.DirFS("../../../web")
	appInstance, err := app.New(app.Config{
		DataDir: tmpDir,
		Assets:  assets,
	})
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}
	defer appInstance.Close()

	srv := httptest.NewServer(appInstance.Handler())
	defer srv.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar err: %v", err)
	}
	client := &http.Client{Jar: jar}

	// 1. GET /api/bootstrap (before onboarding)
	resp, err := client.Get(srv.URL + "/api/bootstrap")
	if err != nil {
		t.Fatalf("GET /api/bootstrap failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for bootstrap, got %d", resp.StatusCode)
	}
	var bootData map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&bootData)
	resp.Body.Close()
	if bootData["onboarded"] != false {
		t.Errorf("expected onboarded false, got %v", bootData["onboarded"])
	}

	// 2. POST /api/onboarding
	onboardingPayload := map[string]any{
		"name":            "Михаил",
		"gender":          "male",
		"age":             29,
		"height":          180.0,
		"weight":          80.0,
		"targetWeight":    75.0,
		"goal":            "weight_loss",
		"level":           "beginner",
		"equipment":       "bodyweight",
		"daysPerWeek":     3,
		"workoutDuration": 45,
		"preferredTime":   "evening",
		"isVegan":         true,
		"healthIssues":    []string{"joints"},
	}
	bodyJSON, _ := json.Marshal(onboardingPayload)
	resp, err = client.Post(srv.URL+"/api/onboarding", "application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		t.Fatalf("POST /api/onboarding failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 for onboarding, got %d: %s", resp.StatusCode, string(body))
	}
	var onbRes map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&onbRes)
	resp.Body.Close()
	if onbRes["ok"] != true || onbRes["hasMedicalFlags"] != true {
		t.Errorf("expected ok=true and hasMedicalFlags=true, got %v", onbRes)
	}

	// 3. GET /api/dashboard
	resp, err = client.Get(srv.URL + "/api/dashboard")
	if err != nil {
		t.Fatalf("GET /api/dashboard failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for dashboard, got %d", resp.StatusCode)
	}
	var dashData map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&dashData)
	resp.Body.Close()
	if dashData["isGentleMode"] != true {
		t.Errorf("expected gentle mode true due to joints issue, got %v", dashData["isGentleMode"])
	}

	// 4. GET /api/workouts
	resp, err = client.Get(srv.URL + "/api/workouts?day=1")
	if err != nil {
		t.Fatalf("GET /api/workouts failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for workouts, got %d", resp.StatusCode)
	}
	var workoutsData map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&workoutsData)
	resp.Body.Close()
	if workoutsData["selectedDay"].(float64) != 1 {
		t.Errorf("expected selectedDay 1, got %v", workoutsData["selectedDay"])
	}

	// 4b. POST /api/workouts/note
	notePayload := map[string]any{"day": 1, "note": "Жим лежа 75кг 4х8"}
	nJSON, _ := json.Marshal(notePayload)
	resp, err = client.Post(srv.URL+"/api/workouts/note", "application/json", bytes.NewReader(nJSON))
	if err != nil {
		t.Fatalf("POST /api/workouts/note failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for workout note, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 4c. POST /api/workouts/reschedule
	reschedPayload := map[string]any{"fromDay": 1, "toDay": 2}
	rsJSON, _ := json.Marshal(reschedPayload)
	resp, err = client.Post(srv.URL+"/api/workouts/reschedule", "application/json", bytes.NewReader(rsJSON))
	if err != nil {
		t.Fatalf("POST /api/workouts/reschedule failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for workout reschedule, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 5. POST /api/workouts/complete
	completePayload := map[string]any{"day": 1, "completed": true, "feeling": "great", "energy": 9}
	cJSON, _ := json.Marshal(completePayload)
	resp, err = client.Post(srv.URL+"/api/workouts/complete", "application/json", bytes.NewReader(cJSON))
	if err != nil {
		t.Fatalf("POST /api/workouts/complete failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for workout complete, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 6. POST /api/water/add
	waterPayload := map[string]any{"delta": 330}
	wJSON, _ := json.Marshal(waterPayload)
	resp, err = client.Post(srv.URL+"/api/water/add", "application/json", bytes.NewReader(wJSON))
	if err != nil {
		t.Fatalf("POST /api/water/add failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for water add, got %d", resp.StatusCode)
	}
	var waterRes map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&waterRes)
	resp.Body.Close()
	if waterRes["ok"] != true || waterRes["current"].(float64) < 330 {
		t.Errorf("unexpected water add result: %v", waterRes)
	}

	// 6b. POST /api/water/add with negative delta (subtract 330ml glass)
	waterSubPayload := map[string]any{"delta": -330}
	wsJSON, _ := json.Marshal(waterSubPayload)
	resp, err = client.Post(srv.URL+"/api/water/add", "application/json", bytes.NewReader(wsJSON))
	if err != nil {
		t.Fatalf("POST /api/water/add negative failed: %v", err)
	}
	var waterSubRes map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&waterSubRes)
	resp.Body.Close()
	if waterSubRes["ok"] != true || waterSubRes["current"].(float64) != 0 {
		t.Errorf("unexpected water subtract result, expected 0: %v", waterSubRes)
	}

	// 7. GET /api/nutrition
	resp, err = client.Get(srv.URL + "/api/nutrition?day=1")
	if err != nil {
		t.Fatalf("GET /api/nutrition failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for nutrition, got %d", resp.StatusCode)
	}
	var nutData map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&nutData)
	resp.Body.Close()
	if nutData["dayMeals"] == nil {
		t.Errorf("expected dayMeals to be present")
	}

	// 8. POST /api/progress/weight
	weightPayload := map[string]any{"weight": 79.2, "note": "после кардио, чувствую прогресс"}
	wtJSON, _ := json.Marshal(weightPayload)
	resp, err = client.Post(srv.URL+"/api/progress/weight", "application/json", bytes.NewReader(wtJSON))
	if err != nil {
		t.Fatalf("POST /api/progress/weight failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for progress weight, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 9. GET /api/catalog
	resp, err = client.Get(srv.URL + "/api/catalog?q=приседания")
	if err != nil {
		t.Fatalf("GET /api/catalog failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for catalog, got %d", resp.StatusCode)
	}
	var catData map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&catData)
	resp.Body.Close()
	if catData["exercises"] == nil {
		t.Errorf("expected exercises array")
	}

	// 9b. GET /api/catalog/meals
	resp, err = client.Get(srv.URL + "/api/catalog/meals?slot=breakfast")
	if err != nil {
		t.Fatalf("GET /api/catalog/meals failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for catalog meals, got %d", resp.StatusCode)
	}
	var mealsCatData map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&mealsCatData)
	resp.Body.Close()
	if mealsCatData["meals"] == nil {
		t.Errorf("expected meals array")
	}

	// 9c. POST /api/nutrition/add
	addReq, _ := json.Marshal(map[string]any{"day": 1, "mealId": "oatmeal-berries-chia"})
	resp, err = client.Post(srv.URL+"/api/nutrition/add", "application/json", bytes.NewReader(addReq))
	if err != nil {
		t.Fatalf("POST /api/nutrition/add failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for nutrition add, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 10. GET /api/plan/export
	resp, err = client.Get(srv.URL + "/api/plan/export")
	if err != nil {
		t.Fatalf("GET /api/plan/export failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for plan export, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 11. POST /api/demo/seed
	resp, err = client.Post(srv.URL+"/api/demo/seed", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/demo/seed failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for demo seed, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 12. POST /api/demo/reset
	resp, err = client.Post(srv.URL+"/api/demo/reset", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/demo/reset failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for demo reset, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 13. Verify /api/workouts returns notOnboarded: true after reset
	resp, err = client.Get(srv.URL + "/api/workouts")
	if err != nil {
		t.Fatalf("GET /api/workouts after reset failed: %v", err)
	}
	var afterReset map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&afterReset)
	resp.Body.Close()
	if afterReset["notOnboarded"] != true {
		t.Errorf("expected notOnboarded: true after reset, got %v", afterReset)
	}
}
