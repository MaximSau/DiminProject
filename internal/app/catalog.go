package app

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed data/exercises.json data/meals.json
var catalogFiles embed.FS

// Exercise — запись из демонстрационного каталога упражнений.
//
// Каталог вшит в бинарник и помечен как демонстрационный: это не живая
// интеграция с базой упражнений, и интерфейс обязан это показывать.
type Exercise struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Muscles     []string `json:"muscles"`
	Equipment   string   `json:"equipment"`
	Level       string   `json:"level"`
	Kind        string   `json:"kind"`
	Sets        int      `json:"sets"`
	Reps        string   `json:"reps"`
	RestSec     int      `json:"rest_sec"`
	DurationSec int      `json:"duration_sec"`
	Tempo       string   `json:"tempo"`
	Start       string   `json:"start"`
	Steps       []string `json:"steps"`
	Breathing   string   `json:"breathing"`
	Mistakes    []string `json:"mistakes"`
	Safety      string   `json:"safety"`
	Easier      string   `json:"easier"`
	Harder      string   `json:"harder"`
	Video       string   `json:"video"`
}

// Meal — блюдо из демонстрационного каталога.
type Meal struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slot        string   `json:"slot"`
	Kcal        int      `json:"kcal"`
	ProteinG    float64  `json:"protein_g"`
	FatG        float64  `json:"fat_g"`
	CarbsG      float64  `json:"carbs_g"`
	Portion     string   `json:"portion"`
	Ingredients []string `json:"ingredients"`
	Prep        string   `json:"prep"`
	Swaps       []string `json:"swaps"`
	Tags        []string `json:"tags"`
	Demo        bool     `json:"demo"`
}

// Catalog — загруженные каталоги с индексами для выборок.
type Catalog struct {
	Exercises []Exercise
	Meals     []Meal

	exByID   map[string]Exercise
	mealByID map[string]Meal
}

// LoadCatalog читает вшитые JSON-каталоги и строит индексы.
func LoadCatalog() (*Catalog, error) {
	c := &Catalog{
		exByID:   map[string]Exercise{},
		mealByID: map[string]Meal{},
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

// Exercise возвращает упражнение по идентификатору.
func (c *Catalog) Exercise(id string) (Exercise, bool) {
	ex, ok := c.exByID[id]
	return ex, ok
}

// Meal возвращает блюдо по идентификатору.
func (c *Catalog) Meal(id string) (Meal, bool) {
	m, ok := c.mealByID[id]
	return m, ok
}

// SuitableExercises отбирает упражнения под оборудование и уровень.
// Упражнения выше уровня тоже попадают в список, но помечаются сложными.
func (c *Catalog) SuitableExercises(equipment, level string) []Exercise {
	maxEquip := EquipmentRank(equipment)
	maxLevel := levelRank(level)
	var out []Exercise
	for _, ex := range c.Exercises {
		if EquipmentRank(ex.Equipment) > maxEquip {
			continue
		}
		if levelRank(ex.Level) > maxLevel+1 {
			continue
		}
		out = append(out, ex)
	}
	return out
}

// ExercisesByKind — упражнения нужного типа (strength, cardio, core, mobility).
func (c *Catalog) ExercisesByKind(pool []Exercise, kind string) []Exercise {
	var out []Exercise
	for _, ex := range pool {
		if ex.Kind == kind {
			out = append(out, ex)
		}
	}
	return out
}

// SuitableMeals отбирает блюда с учётом диетических предпочтений и аллергий.
// Предпочтения — разрешённые теги, аллергии/«не ем» — запрещённые теги и слова.
//
// Логика намеренно простая и консервативная: если данных не хватает,
// блюдо лучше исключить, чем предложить небезопасный вариант.
func (c *Catalog) SuitableMeals(slots []string, prefs, allergies []string, avoid string) map[string][]Meal {
	wantSlots := map[string]bool{}
	for _, s := range slots {
		wantSlots[s] = true
	}
	blocked := normalizeTags(allergies)
	avoidWords := normalizeWords(avoid)
	allowed := normalizeTags(prefs)

	out := map[string][]Meal{}
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

// conflicts проверяет, попадает ли блюдо под аллергию или запрет.
func conflicts(m Meal, blocked, avoidWords map[string]bool) bool {
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

// matchesPrefs — есть ли пересечение тегов блюда с разрешёнными предпочтениями.
// Предпочтение «без лактозы» (lactose-free) считается выполненным, если у блюда
// нет тега lactose. Аналогично для глютена.
func matchesPrefs(m Meal, allowed map[string]bool) bool {
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
