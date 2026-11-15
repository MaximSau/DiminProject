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

func (c *Catalog) MealsBySlot(slot string) []models.Meal {
	var out []models.Meal
	for _, m := range c.Meals {
		if m.Slot == slot {
			out = append(out, m)
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

// matchesPrefs проверяет соответствие диетическим фильтрам строго (веганство, вегетарианство, лактоза, глютен).
func matchesPrefs(m models.Meal, allowed map[string]bool) bool {
	if allowed["vegan"] && !hasTag(m.Tags, "vegan") {
		return false
	}
	if allowed["vegetarian"] && !hasTag(m.Tags, "vegetarian") && !hasTag(m.Tags, "vegan") {
		return false
	}
	if allowed["lactose-free"] && hasTag(m.Tags, "lactose") {
		return false
	}
	if allowed["gluten-free"] && hasTag(m.Tags, "gluten") {
		return false
	}
	return true
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
	if p.WorkoutsPerWeek < 1 {
		p.WorkoutsPerWeek = 3
	}
	if p.WorkoutsPerWeek > 7 {
		p.WorkoutsPerWeek = 7
	}
	if p.SessionMinutes <= 0 {
		p.SessionMinutes = 45
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

// Хелперы категорий мышц для построения сплитов без повторов одних и тех же ног.
func hasAnyMuscle(ex models.Exercise, needles ...string) bool {
	for _, m := range ex.Muscles {
		mLow := strings.ToLower(m)
		for _, n := range needles {
			if strings.Contains(mLow, strings.ToLower(n)) {
				return true
			}
		}
	}
	return false
}

func filterPush(pool []models.Exercise) []models.Exercise {
	var out []models.Exercise
	for _, ex := range pool {
		if hasAnyMuscle(ex, "грудь", "трицепс", "плечи", "дельта") && ex.Kind == "strength" {
			out = append(out, ex)
		}
	}
	return out
}

func filterPull(pool []models.Exercise) []models.Exercise {
	var out []models.Exercise
	for _, ex := range pool {
		if hasAnyMuscle(ex, "спина", "широчайшие", "бицепс", "ромбовидные", "трапеции") && ex.Kind == "strength" {
			out = append(out, ex)
		}
	}
	return out
}

func filterLegs(pool []models.Exercise) []models.Exercise {
	var out []models.Exercise
	for _, ex := range pool {
		if hasAnyMuscle(ex, "ноги", "квадрицепс", "ягодицы", "икры", "задняя поверхность", "таз") && ex.Kind == "strength" {
			out = append(out, ex)
		}
	}
	return out
}

func (c *Catalog) buildWeek(p models.Profile, pool []models.Exercise) []models.PlanDay {
	count := p.WorkoutsPerWeek
	if count < 1 {
		count = 1
	}
	if count > 7 {
		count = 7
	}
	trainingDays := spreadDays(count)
	isTraining := map[int]bool{}
	for _, d := range trainingDays {
		isTraining[d] = true
	}

	pushPool := filterPush(pool)
	pullPool := filterPull(pool)
	legsPool := filterLegs(pool)
	corePool := c.ExercisesByKind(pool, "core")
	cardioPool := c.ExercisesByKind(pool, "cardio")
	mobilityPool := c.ExercisesByKind(pool, "mobility")

	// Если пулы пусты из-за ограничений, fallback на общие strength
	allStrength := c.ExercisesByKind(pool, "strength")
	if len(pushPool) == 0 {
		pushPool = allStrength
	}
	if len(pullPool) == 0 {
		pullPool = allStrength
	}
	if len(legsPool) == 0 {
		legsPool = allStrength
	}

	// Количество упражнений масштабируется от длительности сессии
	targetExCount := p.SessionMinutes / 10
	if targetExCount < 3 {
		targetExCount = 3
	}
	if targetExCount > 9 {
		targetExCount = 9
	}

	days := make([]models.PlanDay, 7)
	trainingSeq := 0
	pushCursor, pullCursor, legsCursor, coreCursor, cardioCursor := 0, 0, 0, 0, 0

	for i := 0; i < 7; i++ {
		days[i].Index = i
		days[i].Title = models.WeekdayNames[i]

		if !isTraining[i] {
			days[i].Rest = true
			days[i].Title = models.WeekdayNames[i] + " — отдых"
			days[i].Focus = "Восстановление и прогулка"
			days[i].Notes = daySessionNote(p, "Отдых", true)
			continue
		}

		// Выбираем фокус сплита на этот тренировочный день
		splitType, focusLabel := determineSplitDay(count, trainingSeq)
		trainingSeq++

		var sessionExs []models.PlanExercise

		// 1. Разминка / мобильность в начале сессии
		if len(mobilityPool) > 0 {
			sessionExs = append(sessionExs, toPlanExerciseGoal(mobilityPool[i%len(mobilityPool)], p.Goal))
		}

		// 2. Основная часть в зависимости от типа сплита
		switch splitType {
		case "push":
			// Жимовой день: грудь, плечи, трицепс
			n := targetExCount - 1
			if p.Goal == models.GoalLose && len(cardioPool) > 0 {
				n--
			}
			if len(corePool) > 0 {
				n--
			}
			for k := 0; k < n && len(pushPool) > 0; k++ {
				sessionExs = append(sessionExs, toPlanExerciseGoal(pushPool[pushCursor%len(pushPool)], p.Goal))
				pushCursor++
			}
		case "pull":
			// Тяговый день: спина, бицепс
			n := targetExCount - 1
			if p.Goal == models.GoalLose && len(cardioPool) > 0 {
				n--
			}
			if len(corePool) > 0 {
				n--
			}
			for k := 0; k < n && len(pullPool) > 0; k++ {
				sessionExs = append(sessionExs, toPlanExerciseGoal(pullPool[pullCursor%len(pullPool)], p.Goal))
				pullCursor++
			}
		case "legs":
			// День ног и ягодиц
			n := targetExCount - 1
			if p.Goal == models.GoalLose && len(cardioPool) > 0 {
				n--
			}
			if len(corePool) > 0 {
				n--
			}
			for k := 0; k < n && len(legsPool) > 0; k++ {
				sessionExs = append(sessionExs, toPlanExerciseGoal(legsPool[legsCursor%len(legsPool)], p.Goal))
				legsCursor++
			}
		case "upper":
			// Верх тела: баланс push + pull
			half := (targetExCount - 2) / 2
			if half < 1 {
				half = 1
			}
			for k := 0; k < half && len(pushPool) > 0; k++ {
				sessionExs = append(sessionExs, toPlanExerciseGoal(pushPool[pushCursor%len(pushPool)], p.Goal))
				pushCursor++
			}
			for k := 0; k < half && len(pullPool) > 0; k++ {
				sessionExs = append(sessionExs, toPlanExerciseGoal(pullPool[pullCursor%len(pullPool)], p.Goal))
				pullCursor++
			}
		case "lower":
			// Низ тела + кор
			n := targetExCount - 2
			if n < 2 {
				n = 2
			}
			for k := 0; k < n && len(legsPool) > 0; k++ {
				sessionExs = append(sessionExs, toPlanExerciseGoal(legsPool[legsCursor%len(legsPool)], p.Goal))
				legsCursor++
			}
		case "fullbody":
			// Всё тело сбалансированно
			if len(pushPool) > 0 {
				sessionExs = append(sessionExs, toPlanExerciseGoal(pushPool[pushCursor%len(pushPool)], p.Goal))
				pushCursor++
			}
			if len(pullPool) > 0 {
				sessionExs = append(sessionExs, toPlanExerciseGoal(pullPool[pullCursor%len(pullPool)], p.Goal))
				pullCursor++
			}
			if len(legsPool) > 0 {
				sessionExs = append(sessionExs, toPlanExerciseGoal(legsPool[legsCursor%len(legsPool)], p.Goal))
				legsCursor++
			}
		case "hiit_core":
			// Кардио-интенсив и кор
			for k := 0; k < 2 && len(cardioPool) > 0; k++ {
				sessionExs = append(sessionExs, toPlanExerciseGoal(cardioPool[cardioCursor%len(cardioPool)], p.Goal))
				cardioCursor++
			}
		default:
			// Общий день
			if len(allStrength) > 0 {
				sessionExs = append(sessionExs, toPlanExerciseGoal(allStrength[pushCursor%len(allStrength)], p.Goal))
				pushCursor++
			}
		}

		// 3. Блок мышц кора и пресса
		if len(corePool) > 0 && splitType != "hiit_core" {
			sessionExs = append(sessionExs, toPlanExerciseGoal(corePool[coreCursor%len(corePool)], p.Goal))
			coreCursor++
		}

		// 4. Для цели Похудение (GoalLose) обязательно добавляем кардио-финишер
		if p.Goal == models.GoalLose && len(cardioPool) > 0 && splitType != "hiit_core" {
			sessionExs = append(sessionExs, toPlanExerciseGoal(cardioPool[cardioCursor%len(cardioPool)], p.Goal))
			cardioCursor++
		}

		days[i].Minutes = estimateMinutes(sessionExs)
		days[i].Exercises = sessionExs
		days[i].Focus = focusLabel
		days[i].Notes = daySessionNote(p, focusLabel, false)
	}

	return days
}

// determineSplitDay распределяет тренировочные фокусы по дням недели.
func determineSplitDay(totalWorkouts, seq int) (string, string) {
	switch totalWorkouts {
	case 1:
		return "fullbody", "Фулбоди (все группы мышц)"
	case 2:
		if seq == 0 {
			return "upper", "Верх тела (Грудь, Спина, Плечи)"
		}
		return "lower", "Низ тела (Ноги, Ягодицы, Кор)"
	case 3:
		switch seq % 3 {
		case 0:
			return "push", "Push — Жимовой день (Грудь, Плечи, Трицепс)"
		case 1:
			return "pull", "Pull — Тяговый день (Спина, Бицепс, Пресс)"
		default:
			return "legs", "Legs — День ног и ягодиц"
		}
	case 4:
		switch seq % 4 {
		case 0:
			return "push", "Верх тела — Жимовой акцент"
		case 1:
			return "lower", "Низ тела — Квадрицепс и кор"
		case 2:
			return "pull", "Верх тела — Тяговый акцент (Спина, Бицепс)"
		default:
			return "legs", "Низ тела — Ягодицы и бицепс бедра"
		}
	case 5:
		switch seq % 5 {
		case 0:
			return "push", "Грудь и Трицепс"
		case 1:
			return "pull", "Спина и Бицепс"
		case 2:
			return "legs", "Ноги и Ягодицы"
		case 3:
			return "upper", "Плечи и Мышцы кора"
		default:
			return "hiit_core", "Метаболическое кардио и Пресс"
		}
	case 6:
		switch seq % 6 {
		case 0:
			return "push", "Push 1 (Грудь, Плечи, Трицепс)"
		case 1:
			return "pull", "Pull 1 (Спина, Бицепс)"
		case 2:
			return "legs", "Legs 1 (Ноги, Икры)"
		case 3:
			return "push", "Push 2 (Объёмный жимовой)"
		case 4:
			return "pull", "Pull 2 (Широчайшие и Кор)"
		default:
			return "legs", "Legs 2 (Ягодицы и Задняя поверхность)"
		}
	case 7:
		switch seq % 7 {
		case 0:
			return "push", "Push (Грудь, Плечи)"
		case 1:
			return "pull", "Pull (Спина, Бицепс)"
		case 2:
			return "legs", "Legs (Ноги, Ягодицы)"
		case 3:
			return "upper", "Верх тела (Умеренная нагрузка)"
		case 4:
			return "lower", "Низ тела и Кор"
		case 5:
			return "hiit_core", "HIIT Кардио и Пресс"
		default:
			return "fullbody", "Активное восстановление и Мобильность"
		}
	default:
		return "fullbody", "Общая подготовка"
	}
}

func daySessionNote(p models.Profile, focus string, isRest bool) string {
	if isRest {
		if p.Goal == models.GoalLose {
			return "День восстановления — мышцы адаптируются, а дефицит калорий сжигает жир. Рекомендуется 30–45 минут лёгкой ходьбы на свежем воздухе и соблюдение водного баланса (330 мл на стакан)."
		}
		return "День восстановления — мышечные волокна восстанавливаются и растут в покое. Обеспечьте качественный сон 8 часов и достаточное поступление белка."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Фокус: %s. Ориентировочная длительность — %d мин. ", focus, p.SessionMinutes))
	if p.Goal == models.GoalLose {
		b.WriteString("Жиросжигающий интервальный темп с коротким отдыхом (30–45 сек) и обязательным кардио-финишером. ")
	} else if p.Goal == models.GoalGain {
		b.WriteString("Силовой гипертрофический объём с акцентом на базовые движения и отдых 60–90 сек для максимального восстановления сил. ")
	}

	switch p.PreferredTime {
	case models.TimeMorning:
		b.WriteString("Утренний совет: начните с суставной разминки и выпейте 1 стакан чистой воды (330 мл) перед стартом.")
	case models.TimeDay:
		b.WriteString("Дневной совет: пик мышечной силы — держите строгую технику и прогрессируйте в рабочих подходах.")
	case models.TimeEvening:
		b.WriteString("Вечерний совет: завершите тренировку за 2.5–3 часа до сна и сделайте плавную растяжку для снижения пульса.")
	default:
		b.WriteString("Следите за техникой дыхания и пейте воду небольшими глотками между подходами.")
	}
	return b.String()
}

func spreadDays(n int) []int {
	switch n {
	case 1:
		return []int{2} // Среда
	case 2:
		return []int{1, 4} // Вторник, Пятница
	case 3:
		return []int{0, 2, 4} // Понедельник, Среда, Пятница
	case 4:
		return []int{0, 2, 4, 6} // Понедельник, Среда, Пятница, Воскресенье
	case 5:
		return []int{0, 1, 3, 4, 5} // Понедельник, Вторник, Четверг, Пятница, Суббота
	case 6:
		return []int{0, 1, 2, 3, 4, 5} // Понедельник — Суббота
	case 7:
		return []int{0, 1, 2, 3, 4, 5, 6} // Каждый день
	default:
		return []int{0, 2, 4}
	}
}

func toPlanExerciseGoal(ex models.Exercise, goal string) models.PlanExercise {
	pe := models.PlanExercise{
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

	if goal == models.GoalLose {
		if pe.Kind == "strength" {
			pe.Sets = 3
			pe.Reps = "12–15"
			pe.RestSec = 40
		} else if pe.Kind == "cardio" {
			pe.Sets = 3
			pe.RestSec = 30
			if pe.DurationSec == 0 {
				pe.Reps = "15–20"
			}
		} else if pe.Kind == "core" {
			pe.RestSec = 30
		}
	} else if goal == models.GoalGain {
		if pe.Kind == "strength" {
			pe.Sets = 4
			pe.Reps = "8–10"
			pe.RestSec = 75
		} else if pe.Kind == "core" {
			pe.RestSec = 50
		}
	}

	return pe
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

func (c *Catalog) buildMeals(p models.Profile) ([]models.MealDay, *int, *int, *int, *int) {
	slots := []string{"breakfast", "lunch", "dinner", "snack"}
	bySlot := c.SuitableMeals(slots, p.DietPreferences, p.Allergies, p.AvoidFoods)

	days := make([]models.MealDay, 7)
	for d := 0; d < 7; d++ {
		days[d].Index = d
		for si, slot := range slots {
			list := bySlot[slot]
			if len(list) == 0 {
				// Защитный fallback на общее меню слота
				list = c.MealsBySlot(slot)
			}
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
	b.WriteString(fmt.Sprintf("План собран под цель «%s», уровень «%s» и инвентарь «%s». ",
		models.GoalLabel(p.Goal), strings.ToLower(models.LevelLabel(p.Level)), strings.ToLower(models.EquipLabel(p.Equipment))))
	b.WriteString(fmt.Sprintf("Тренировок в неделю: %d, ориентировочно по %d минут. ", p.WorkoutsPerWeek, p.SessionMinutes))
	if p.Goal == models.GoalLose {
		b.WriteString("Программа включает жиросжигающий интервальный тренинг, кардио-финишеры и контроль водного баланса. ")
	} else if p.Goal == models.GoalGain {
		b.WriteString("Программа сфокусирована на гипертрофии, базовых многосуставных движениях и достаточном времени отдыха между сетами. ")
	}
	b.WriteString("Нагрузка дозирована по принципу суперкомпенсации: дни отдыха обязательны для адаптации мышц.")
	return b.String()
}

func calcPlanNote(p models.Profile) string {
	if !p.HasBodyData() {
		return p.BodyHint() + " " + models.BodyDataNote()
	}
	return "Калорийность и БЖУ посчитаны по формуле Миффлина — Сан Жеора с поправкой на физическую активность. " +
		"Это научно выверенный ориентир; " + models.BodyDataNote()
}

func progressionNote(p models.Profile) string {
	if p.Level == models.LevelBeginner {
		return "Первые 2–3 недели — адаптация нейромышечной связи: делайте упражнения в комфортной амплитуде и следите за дыханием. " +
			"Когда повторения станут даваться легко, увеличьте количество повторов на 2–3 или добавьте небольшой вес."
	}
	return "Увеличивайте нагрузку по принципу прогрессивной перегрузки: сначала повторения в подходе, затем рабочий вес отягощения. " +
		"Если техника начинает ломаться — вернитесь на шаг назад. Сон и питание определяют 80% успеха."
}

func buildRoutine(p models.Profile) models.Routine {
	var sleep string
	switch p.PreferredTime {
	case models.TimeEvening:
		sleep = "Старайтесь завершать тренировку за 2.5–3 часа до сна, чтобы уровень кортизола и температура тела снизились к ночи. Спите 7–9 часов в проветренной комнате."
	default:
		sleep = "Спите 7–9 часов. Качественный глубокий сон критически важен: в медленной фазе сна вырабатывается соматотропин, восстанавливающий мышечные ткани."
	}

	targetMl := p.WaterTargetMl()
	glasses := (targetMl + 165) / 330
	if glasses < 1 {
		glasses = 6
	}
	water := fmt.Sprintf("Суточная норма чистой воды: ~%d стаканов (~%d мл, из расчёта 330 мл на стакан). Держите стакан на виду, выпивайте стакан после пробуждения и перед тренировкой.", glasses, targetMl)

	var activity string
	if p.Goal == models.GoalLose {
		activity = "Ежедневная внетренировочная активность (NEAT) — ключевой фактор дефицита калорий. Проходите 8 000–10 000 шагов в день и чаще двигайтесь."
	} else {
		activity = "При наборе массы избегайте изнуряющего кардио в дни отдыха: 20–30 минут спокойной прогулки улучшат кровоток без расхода мышечного гликогена."
	}

	recovery := []string{
		"Разминка 5 минут перед тренировкой разогревает суставы и синовиальную жидкость, предотвращая травмы.",
		"Заминка и статическая растяжка 5 минут после тренировки нормализуют тонус мышц и успокаивают дыхание.",
		"Регулярность важнее разовых рекордов: держите темп плана и не пропускайте дни отдыха.",
	}
	switch p.PreferredTime {
	case models.TimeMorning:
		recovery = append(recovery, "Утренний график тренировок: выпейте стакан тёплой воды за 20 мин до тренировки, а плотный белковый завтрак запланируйте после занятия.")
	case models.TimeDay:
		recovery = append(recovery, "Дневной график тренировок: планируйте приём пищи со сложными углеводами за 1.5–2 часа до физической нагрузки.")
	case models.TimeEvening:
		recovery = append(recovery, "Вечерний график тренировок: лёгкий белковый ужин с клетчаткой после тренировки обеспечит аминокислоты для ночного восстановления.")
	}

	return models.Routine{Sleep: sleep, Water: water, Activity: activity, Recovery: recovery}
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
