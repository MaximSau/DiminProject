package service

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"pulsefit/internal/models"
)

//go:embed data/exercises.json data/meals.json
var catalogFiles embed.FS

// Catalog — загруженные каталоги с индексами для выборок.
type Catalog struct {
	Exercises []models.Exercise
	Meals     []models.Meal

	exByID   map[string]models.Exercise
	mealByID map[string]models.Meal
}

// LoadCatalog читает вшитые JSON-каталоги и строит индексы.
func LoadCatalog() (*Catalog, error) {
	c := &Catalog{
		exByID:   map[string]models.Exercise{},
		mealByID: map[string]models.Meal{},
	}

	raw, err := catalogFiles.ReadFile("data/exercises.json")
	if err != nil {
		return nil, fmt.Errorf("каталог упражнений: %w", err)
	}
	if err := json.Unmarshal(raw, &c.Exercises); err != nil {
		return nil, fmt.Errorf("каталог упражнений повреждён: %w", err)
	}
	for _, ex := range c.Exercises {
		if ex.ID == "" {
			return nil, fmt.Errorf("упражнение без id в каталоге")
		}
		c.exByID[ex.ID] = ex
	}

	raw, err = catalogFiles.ReadFile("data/meals.json")
	if err != nil {
		return nil, fmt.Errorf("каталог блюд: %w", err)
	}
	if err := json.Unmarshal(raw, &c.Meals); err != nil {
		return nil, fmt.Errorf("каталог блюд повреждён: %w", err)
	}
	for _, m := range c.Meals {
		if m.ID == "" {
			return nil, fmt.Errorf("блюдо без id в каталоге")
		}
		m.Demo = true
		c.mealByID[m.ID] = m
	}

	sort.SliceStable(c.Exercises, func(i, j int) bool { return c.Exercises[i].Name < c.Exercises[j].Name })
	sort.SliceStable(c.Meals, func(i, j int) bool { return c.Meals[i].Name < c.Meals[j].Name })
	return c, nil
}

func (c *Catalog) Exercise(id string) (models.Exercise, bool) {
	ex, ok := c.exByID[id]
	return ex, ok
}

func (c *Catalog) Meal(id string) (models.Meal, bool) {
	m, ok := c.mealByID[id]
	return m, ok
}

func (c *Catalog) SuitableExercises(equipment, level string) []models.Exercise {
	maxEquip := models.EquipmentRank(equipment)
	maxLevel := models.LevelRank(level)
	var out []models.Exercise
	for _, ex := range c.Exercises {
		if models.EquipmentRank(ex.Equipment) > maxEquip {
			continue
		}
		if models.LevelRank(ex.Level) > maxLevel+1 {
			continue
		}
		out = append(out, ex)
	}
	return out
}

func (c *Catalog) ExercisesByKind(pool []models.Exercise, kind string) []models.Exercise {
	var out []models.Exercise
	for _, ex := range pool {
		if ex.Kind == kind {
			out = append(out, ex)
		}
	}
	return out
}

func (c *Catalog) SuitableMeals(slots []string, prefs, allergies []string, avoid string) map[string][]models.Meal {
	wantSlots := map[string]bool{}
	for _, s := range slots {
		wantSlots[s] = true
	}
	blocked := normalizeTags(allergies)
	avoidWords := normalizeWords(avoid)
	allowed := normalizeTags(prefs)

	out := map[string][]models.Meal{}
	for _, m := range c.Meals {
		if len(wantSlots) > 0 && !wantSlots[m.Slot] {
			continue
		}
		if conflicts(m, blocked, avoidWords) {
			continue
		}
		if len(allowed) > 0 && !matchesPrefs(m, allowed) {
			continue
		}
		out[m.Slot] = append(out[m.Slot], m)
	}
	return out
}

func conflicts(m models.Meal, blocked, avoidWords map[string]bool) bool {
	if len(blocked) == 0 && len(avoidWords) == 0 {
		return false
	}
	if hasOverlap(m.Tags, blocked) {
		return true
	}
	fields := append([]string{m.Name}, m.Ingredients...)
	for _, f := range fields {
		for _, w := range strings.Fields(strings.ToLower(f)) {
			w = strings.Trim(w, ",.;:()[]")
			if avoidWords[w] {
				return true
			}
		}
	}
	return false
}

func matchesPrefs(m models.Meal, allowed map[string]bool) bool {
	for _, tag := range m.Tags {
		if allowed[tag] {
			return true
		}
	}
	if allowed["lactose-free"] && !hasTag(m.Tags, "lactose") {
		return true
	}
	if allowed["gluten-free"] && !hasTag(m.Tags, "gluten") {
		return true
	}
	if allowed["vegetarian"] && (hasTag(m.Tags, "vegetarian") || hasTag(m.Tags, "vegan")) {
		return true
	}
	if allowed["vegan"] && hasTag(m.Tags, "vegan") {
		return true
	}
	return false
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func hasOverlap(tags []string, blocked map[string]bool) bool {
	for _, t := range tags {
		if blocked[t] {
			return true
		}
	}
	return false
}

func normalizeTags(in []string) map[string]bool {
	out := map[string]bool{}
	for _, v := range in {
		v = strings.ToLower(strings.TrimSpace(v))
		if v != "" {
			out[v] = true
		}
	}
	return out
}

func normalizeWords(in string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(in), func(r rune) bool {
		return r == ',' || r == ';' || r == '.' || r == '\n' || r == '/'
	}) {
		w = strings.TrimSpace(w)
		if w != "" {
			out[w] = true
		}
	}
	return out
}

// BuildPlan собирает детерминированный план.
func (c *Catalog) BuildPlan(p models.Profile) (*models.Plan, error) {
	if p.Goal == "" {
		p.Goal = models.GoalLose
	}
	if p.WorkoutsPerWeek <= 0 {
		p.WorkoutsPerWeek = 3
	}
	if p.SessionMinutes <= 0 {
		p.SessionMinutes = 40
	}
	if p.Level == "" {
		p.Level = models.LevelBeginner
	}
	if p.Equipment == "" {
		p.Equipment = models.EquipNone
	}

	pool := c.SuitableExercises(p.Equipment, p.Level)
	if len(pool) == 0 {
		return nil, fmt.Errorf("в каталоге нет упражнений под оборудование %q", p.Equipment)
	}

	days := c.buildWeek(p, pool)
	meals, kcal, protein, fat, carbs := c.buildMeals(p)

	plan := &models.Plan{
		ID:          fmt.Sprintf("plan-%d", time.Now().UnixNano()),
		Version:     p.PlanVersion + 1,
		CreatedAt:   time.Now(),
		Source:      "demo",
		Goal:        p.Goal,
		Explanation: planExplanation(p),
		Days:        days,
		Meals:       meals,
		Routine:     buildRoutine(p),
		KcalTarget:  kcal,
		ProteinG:    protein,
		FatG:        fat,
		CarbsG:      carbs,
		CalcNote:    calcPlanNote(p),
		Progression: progressionNote(p),
	}
	return plan, nil
}

func (c *Catalog) buildWeek(p models.Profile, pool []models.Exercise) []models.PlanDay {
	count := p.WorkoutsPerWeek
	if count > 6 {
		count = 6
	}
	trainingDays := spreadDays(count)
	isTraining := map[int]bool{}
	for _, d := range trainingDays {
		isTraining[d] = true
	}

	strength := c.ExercisesByKind(pool, "strength")
	core := c.ExercisesByKind(pool, "core")
	mobility := c.ExercisesByKind(pool, "mobility")
	cardio := c.ExercisesByKind(pool, "cardio")

	perSession := p.SessionMinutes / 9
	if perSession < 2 {
		perSession = 2
	}
	if perSession > 5 {
		perSession = 5
	}

	days := make([]models.PlanDay, 7)
	strengthCursor := 0
	for i := 0; i < 7; i++ {
		days[i].Index = i
		days[i].Title = models.WeekdayNames[i]
		if !isTraining[i] {
			days[i].Rest = true
			days[i].Title = models.WeekdayNames[i] + " — отдых"
			days[i].Notes = "Отдых — часть плана: мышцам нужно восстановление. Подойдёт спокойная прогулка."
			continue
		}

		var exs []models.PlanExercise
		for k := 0; k < perSession && len(strength) > 0; k++ {
			ex := strength[strengthCursor%len(strength)]
			strengthCursor++
			exs = append(exs, toPlanExercise(ex))
		}
		if len(core) > 0 {
			exs = append(exs, toPlanExercise(core[(i)%len(core)]))
		}
		if i == 0 || len(exs) == 0 {
			if len(mobility) > 0 {
				exs = append([]models.PlanExercise{toPlanExercise(mobility[0])}, exs...)
			}
		}

		days[i].Minutes = estimateMinutes(exs)
		days[i].Exercises = exs
		days[i].Focus = focusFor(exs)
		days[i].Notes = sessionNote(p)
	}

	if len(cardio) > 0 {
		for i := range days {
			if days[i].Rest {
				days[i].Notes += " Можно добавить лёгкое кардио: " + cardio[0].Name + "."
				break
			}
		}
	}
	return days
}

func spreadDays(n int) []int {
	switch n {
	case 1:
		return []int{2}
	case 2:
		return []int{1, 4}
	case 3:
		return []int{0, 2, 4}
	case 4:
		return []int{0, 2, 4, 5}
	case 5:
		return []int{0, 1, 3, 4, 5}
	case 6:
		return []int{0, 1, 2, 3, 4, 5}
	default:
		return []int{0, 2, 4}
	}
}

func toPlanExercise(ex models.Exercise) models.PlanExercise {
	return models.PlanExercise{
		ExerciseID:  ex.ID,
		Name:        ex.Name,
		Muscles:     ex.Muscles,
		Equipment:   ex.Equipment,
		Level:       ex.Level,
		Kind:        ex.Kind,
		Sets:        ex.Sets,
		Reps:        ex.Reps,
		RestSec:     ex.RestSec,
		DurationSec: ex.DurationSec,
		Demo:        true,
	}
}

func estimateMinutes(exs []models.PlanExercise) int {
	total := 5 // разминка
	for _, ex := range exs {
		if ex.DurationSec > 0 {
			total += (ex.DurationSec + ex.RestSec) * ex.Sets / 60
			continue
		}
		total += ex.Sets * (ex.RestSec + 40) / 60
	}
	if total < 15 {
		total = 15
	}
	return total
}

func focusFor(exs []models.PlanExercise) string {
	seen := map[string]bool{}
	var parts []string
	for _, ex := range exs {
		for _, m := range ex.Muscles {
			if !seen[m] {
				seen[m] = true
				parts = append(parts, m)
			}
		}
	}
	if len(parts) == 0 {
		return "Общая подготовка"
	}
	if len(parts) > 3 {
		parts = parts[:3]
	}
	return strings.Join(parts, ", ")
}

func sessionNote(p models.Profile) string {
	return fmt.Sprintf("Разминка 5 минут, затем упражнения. Отдых между подходами указан у каждого упражнения. Ориентировочная длительность — %d минут.", p.SessionMinutes)
}

func (c *Catalog) buildMeals(p models.Profile) ([]models.MealDay, *int, *int, *int, *int) {
	slots := []string{"breakfast", "lunch", "dinner", "snack"}
	bySlot := c.SuitableMeals(slots, p.DietPreferences, p.Allergies, p.AvoidFoods)

	days := make([]models.MealDay, 7)
	for d := 0; d < 7; d++ {
		days[d].Index = d
		for si, slot := range slots {
			list := bySlot[slot]
			if len(list) == 0 {
				continue
			}
			m := list[(d+si)%len(list)]
			days[d].Meals = append(days[d].Meals, models.MealItem{
				MealID:      m.ID,
				Name:        m.Name,
				Slot:        m.Slot,
				Kcal:        m.Kcal,
				ProteinG:    m.ProteinG,
				FatG:        m.FatG,
				CarbsG:      m.CarbsG,
				Portion:     m.Portion,
				Ingredients: m.Ingredients,
				Prep:        m.Prep,
				Swaps:       m.Swaps,
				Demo:        true,
			})
		}
	}

	if !p.HasBodyData() {
		return days, nil, nil, nil, nil
	}
	n := models.EstimateNutrition(p)
	return days, &n.Kcal, &n.ProteinG, &n.FatG, &n.CarbsG
}

func planExplanation(p models.Profile) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("План собран под цель «%s», уровень «%s» и оборудование «%s». ",
		models.GoalShort(p.Goal), strings.ToLower(models.LevelLabel(p.Level)), strings.ToLower(models.EquipLabel(p.Equipment))))
	b.WriteString(fmt.Sprintf("Тренировок в неделю: %d, по %d минут. ", p.WorkoutsPerWeek, p.SessionMinutes))
	b.WriteString("Нагрузка растёт постепенно: сначала техника и регулярность, потом вес и повторения. ")
	b.WriteString("Дни отдыха — обязательная часть плана, а не пропуск. ")
	b.WriteString("Это демонстрационный план из встроенного каталога, а не персональная рекомендация ИИ. ")
	b.WriteString("Рекомендации не являются медицинской консультацией.")
	return b.String()
}

func calcPlanNote(p models.Profile) string {
	if !p.HasBodyData() {
		return p.BodyHint() + " " + models.BodyDataNote()
	}
	return "Калорийность и БЖУ посчитаны по формуле Миффлина — Сан Жеора с поправкой на активность. " +
		"Это ориентир, а не точное предписание; " + models.BodyDataNote()
}

func progressionNote(p models.Profile) string {
	if p.Level == models.LevelBeginner {
		return "Первые 2–3 недели — привыкание: делайте упражнения в комфортной амплитуде и не гонитесь за весом. " +
			"Когда станет легко, добавьте 1–2 повторения или 1 подход. Вес повышайте не чаще раза в неделю и небольшими шагами."
	}
	return "Увеличивайте нагрузку постепенно: +1–2 повторения, затем небольшой шаг по весу. " +
		"Если техника «плывёт» или появилась боль — вернитесь на шаг назад. Отдых и сон важны не меньше тренировок."
}

func buildRoutine(p models.Profile) models.Routine {
	sleep := "Старайтесь спать 7–9 часов. Ложитесь и вставайте примерно в одно время — так проще восстановиться."
	water := "Ориентир по воде — 30 мл на кг веса в день, часть с едой и напитками. В жару и на тренировке — больше."
	if p.WeightKg == nil {
		water = "Пейте регулярно в течение дня, ориентируйтесь на жажду и цвет мочи. Точный объём зависит от веса и активности."
	}
	activity := "Добавьте 20–30 минут спокойной ходьбы в дни отдыха — это помогает восстановлению."
	recovery := []string{
		"Старайтесь завершать тренировку за 2–3 часа до сна.",
		"В дни отдыха не «догоняйте» пропущенное — просто продолжайте план.",
		"Лёгкая растяжка 5–10 минут после тренировки помогает снизить напряжение.",
	}
	return models.Routine{Sleep: sleep, Water: water, Activity: activity, Recovery: recovery}
}
