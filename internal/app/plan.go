package app

import (
	"fmt"
	"strings"
	"time"
)

// Plan — персональный план: тренировки, питание, режим и пояснения.
//
// Plan.Source = "demo" означает, что план собран детерминированным
// генератором на демонстрационном каталоге, а не ИИ.
type Plan struct {
	ID        string    `json:"id"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Source    string    `json:"source"` // demo | ai

	Goal        string `json:"goal"`
	Explanation string `json:"explanation"`

	Days    []PlanDay `json:"days"`
	Meals   []MealDay `json:"meals"`
	Routine Routine   `json:"routine"`

	KcalTarget *int   `json:"kcal_target,omitempty"`
	ProteinG   *int   `json:"protein_g,omitempty"`
	FatG       *int   `json:"fat_g,omitempty"`
	CarbsG     *int   `json:"carbs_g,omitempty"`
	CalcNote   string `json:"calc_note,omitempty"`

	Progression string `json:"progression"`
}

// PlanDay — день недели с тренировкой или днём отдыха.
type PlanDay struct {
	Index     int            `json:"index"` // 0 = понедельник
	Title     string         `json:"title"`
	Focus     string         `json:"focus"`
	Rest      bool           `json:"rest"`
	Minutes   int            `json:"minutes"`
	Exercises []PlanExercise `json:"exercises,omitempty"`
	Notes     string         `json:"notes,omitempty"`
}

// PlanExercise — конкретное упражнение внутри тренировки.
type PlanExercise struct {
	ExerciseID  string   `json:"exercise_id"`
	Name        string   `json:"name"`
	Muscles     []string `json:"muscles"`
	Equipment   string   `json:"equipment"`
	Level       string   `json:"level"`
	Kind        string   `json:"kind"`
	Sets        int      `json:"sets"`
	Reps        string   `json:"reps"`
	RestSec     int      `json:"rest_sec"`
	DurationSec int      `json:"duration_sec"`
	Demo        bool     `json:"demo"`
}

// MealDay — питание на один день.
type MealDay struct {
	Index int        `json:"index"`
	Meals []MealItem `json:"meals"`
}

// MealItem — блюдо в плане питания.
type MealItem struct {
	MealID      string   `json:"meal_id"`
	Name        string   `json:"name"`
	Slot        string   `json:"slot"`
	Kcal        int      `json:"kcal"`
	ProteinG    float64  `json:"protein_g"`
	FatG        float64  `json:"fat_g"`
	CarbsG      float64  `json:"carbs_g"`
	Portion     string   `json:"portion,omitempty"`
	Ingredients []string `json:"ingredients,omitempty"`
	Prep        string   `json:"prep,omitempty"`
	Swaps       []string `json:"swaps,omitempty"`
	Demo        bool     `json:"demo"`
}

// Routine — режим дня: сон, вода, активность, восстановление.
type Routine struct {
	Sleep    string   `json:"sleep"`
	Water    string   `json:"water"`
	Activity string   `json:"activity"`
	Recovery []string `json:"recovery"`
}

// WeekdayNames — названия дней недели по индексу (0 = понедельник).
var WeekdayNames = []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}

// TodayIndex — индекс сегодняшнего дня в неделе плана.
func TodayIndex() int {
	wd := int(time.Now().Weekday())
	if wd == 0 { // воскресенье в Go = 0
		return 6
	}
	return wd - 1
}

// BuildPlan собирает детерминированный демонстрационный план.
func (c *Catalog) BuildPlan(p Profile) (*Plan, error) {
	if p.Goal == "" {
		p.Goal = GoalLose
	}
	if p.WorkoutsPerWeek <= 0 {
		p.WorkoutsPerWeek = 3
	}
	if p.SessionMinutes <= 0 {
		p.SessionMinutes = 40
	}
	if p.Level == "" {
		p.Level = LevelBeginner
	}
	if p.Equipment == "" {
		p.Equipment = EquipNone
	}

	pool := c.SuitableExercises(p.Equipment, p.Level)
	if len(pool) == 0 {
		return nil, fmt.Errorf("в каталоге нет упражнений под оборудование %q", p.Equipment)
	}

	days := c.buildWeek(p, pool)
	meals, kcal, protein, fat, carbs := c.buildMeals(p)

	plan := &Plan{
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

// buildWeek распределяет тренировки и отдых по дням недели.
func (c *Catalog) buildWeek(p Profile, pool []Exercise) []PlanDay {
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

	days := make([]PlanDay, 7)
	strengthCursor := 0
	for i := 0; i < 7; i++ {
		days[i].Index = i
		days[i].Title = WeekdayNames[i]
		if !isTraining[i] {
			days[i].Rest = true
			days[i].Title = WeekdayNames[i] + " — отдых"
			days[i].Notes = "Отдых — часть плана: мышцам нужно восстановление. Подойдёт спокойная прогулка."
			continue
		}

		var exs []PlanExercise
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
				exs = append([]PlanExercise{toPlanExercise(mobility[0])}, exs...)
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

func toPlanExercise(ex Exercise) PlanExercise {
	return PlanExercise{
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

func estimateMinutes(exs []PlanExercise) int {
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

func focusFor(exs []PlanExercise) string {
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

func sessionNote(p Profile) string {
	return fmt.Sprintf("Разминка 5 минут, затем упражнения. Отдых между подходами указан у каждого упражнения. Ориентировочная длительность — %d минут.", p.SessionMinutes)
}

// buildMeals собирает питание на неделю из каталога и оценивает калорийность.
func (c *Catalog) buildMeals(p Profile) ([]MealDay, *int, *int, *int, *int) {
	slots := []string{"breakfast", "lunch", "dinner", "snack"}
	bySlot := c.SuitableMeals(slots, p.DietPreferences, p.Allergies, p.AvoidFoods)

	days := make([]MealDay, 7)
	for d := 0; d < 7; d++ {
		days[d].Index = d
		for si, slot := range slots {
			list := bySlot[slot]
			if len(list) == 0 {
				continue
			}
			m := list[(d+si)%len(list)]
			days[d].Meals = append(days[d].Meals, MealItem{
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
	n := EstimateNutrition(p)
	return days, &n.Kcal, &n.ProteinG, &n.FatG, &n.CarbsG
}

func planExplanation(p Profile) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("План собран под цель «%s», уровень «%s» и оборудование «%s». ",
		GoalShort(p.Goal), strings.ToLower(LevelLabel(p.Level)), strings.ToLower(EquipLabel(p.Equipment))))
	b.WriteString(fmt.Sprintf("Тренировок в неделю: %d, по %d минут. ", p.WorkoutsPerWeek, p.SessionMinutes))
	b.WriteString("Нагрузка растёт постепенно: сначала техника и регулярность, потом вес и повторения. ")
	b.WriteString("Дни отдыха — обязательная часть плана, а не пропуск. ")
	b.WriteString("Это демонстрационный план из встроенного каталога, а не персональная рекомендация ИИ. ")
	b.WriteString("Рекомендации не являются медицинской консультацией.")
	return b.String()
}

func calcPlanNote(p Profile) string {
	if !p.HasBodyData() {
		return p.BodyHint() + " " + BodyDataNote()
	}
	return "Калорийность и БЖУ посчитаны по формуле Миффлина — Сан Жеора с поправкой на активность. " +
		"Это ориентир, а не точное предписание; " + BodyDataNote()
}

func progressionNote(p Profile) string {
	if p.Level == LevelBeginner {
		return "Первые 2–3 недели — привыкание: делайте упражнения в комфортной амплитуде и не гонитесь за весом. " +
			"Когда станет легко, добавьте 1–2 повторения или 1 подход. Вес повышайте не чаще раза в неделю и небольшими шагами."
	}
	return "Увеличивайте нагрузку постепенно: +1–2 повторения, затем небольшой шаг по весу. " +
		"Если техника «плывёт» или появилась боль — вернитесь на шаг назад. Отдых и сон важны не меньше тренировок."
}

func buildRoutine(p Profile) Routine {
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
	return Routine{Sleep: sleep, Water: water, Activity: activity, Recovery: recovery}
}
