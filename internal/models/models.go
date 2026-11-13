package models

import (
	"math"
	"strings"
	"time"
)

// Константы целей
const (
	GoalLose = "lose" // похудение
	GoalGain = "gain" // набор веса
)

// Константы уровней подготовки
const (
	LevelBeginner     = "beginner"
	LevelIntermediate = "intermediate"
	LevelAdvanced     = "advanced"
)

// Константы оборудования
const (
	EquipNone = "none" // без оборудования
	EquipHome = "home" // домашний инвентарь
	EquipGym  = "gym"  // спортзал
)

// Константы времени тренировок
const (
	TimeMorning = "morning"
	TimeDay     = "day"
	TimeEvening = "evening"
)

// Флаги здоровья / ограничений
const (
	HealthNone     = "none"     // нет ограничений
	HealthInjury   = "injury"   // недавние травмы
	HealthChronic  = "chronic"  // хронические заболевания
	HealthHeart    = "heart"    // болезни сердца / гипертония
	HealthPregnant = "pregnant" // беременность / послеродовой период
	HealthSurgery  = "surgery"  // недавние операции
	HealthOther    = "other"    // другое
)

// Profile — анкета и цели пользователя.
type Profile struct {
	ID        string    `json:"id" db:"user_id"`
	Code      string    `json:"code" db:"code"` // 8-символьный код для синхронизации без пароля
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	Name            string   `json:"name" db:"name"`
	Age             *int     `json:"age,omitempty" db:"age"`
	HeightCm        *float64 `json:"height_cm,omitempty" db:"height_cm"`
	WeightKg        *float64 `json:"weight_kg,omitempty" db:"weight_kg"`
	Sex             string   `json:"sex,omitempty" db:"sex"` // male | female | other
	Goal            string   `json:"goal" db:"goal"`         // lose | gain
	Pace            string   `json:"pace" db:"pace"`         // gentle | moderate | steady
	Level           string   `json:"level" db:"level"`       // beginner | intermediate | advanced
	WorkoutsPerWeek int      `json:"workouts_per_week" db:"workouts_per_week"`
	SessionMinutes  int      `json:"session_minutes" db:"session_minutes"`
	Equipment       string   `json:"equipment" db:"equipment"`
	PreferredTime   string   `json:"preferred_time" db:"preferred_time"`
	ActivityLevel   string   `json:"activity_level" db:"activity_level"`
	DietPreferences []string `json:"diet_preferences,omitempty"`
	Allergies       []string `json:"allergies,omitempty"`
	AvoidFoods      string   `json:"avoid_foods,omitempty" db:"avoid_foods"`
	HealthFlags     []string `json:"health_flags,omitempty"`
	HealthNotes     string   `json:"health_notes,omitempty" db:"health_notes"`
	Consent         bool     `json:"consent" db:"consent"`
	Onboarded       bool     `json:"onboarded" db:"onboarded"`
	PlanVersion     int      `json:"plan_version" db:"plan_version"`
}

// MedicalBlocked возвращает true, если у пользователя есть серьезные ограничения.
func (p Profile) MedicalBlocked() bool {
	for _, f := range p.HealthFlags {
		switch f {
		case HealthInjury, HealthChronic, HealthHeart, HealthPregnant, HealthSurgery:
			return true
		case HealthOther:
			if strings.TrimSpace(p.HealthNotes) != "" {
				return true
			}
		}
	}
	return false
}

// HasBodyData — достаточно ли данных для расчёта калорий.
func (p Profile) HasBodyData() bool {
	return p.Age != nil && p.HeightCm != nil && p.WeightKg != nil
}

// BodyHint возвращает пояснение, каких параметров не хватает для расчета.
func (p Profile) BodyHint() string {
	var missing []string
	if p.Age == nil {
		missing = append(missing, "возраст")
	}
	if p.HeightCm == nil {
		missing = append(missing, "рост")
	}
	if p.WeightKg == nil {
		missing = append(missing, "вес")
	}
	if len(missing) == 0 {
		return ""
	}
	return "Для точного расчёта калорий не хватает: " + strings.Join(missing, ", ") + "."
}

func BodyDataNote() string {
	return "Потребности в питании индивидуальны, расчёты носят ориентировочный характер."
}

// BMI рассчитывает индекс массы тела и возвращает (индекс, текстовый статус, тип статуса).
func (p Profile) BMI() (float64, string, string) {
	if p.HeightCm == nil || p.WeightKg == nil || *p.HeightCm <= 0 {
		return 0, "", ""
	}
	hM := *p.HeightCm / 100.0
	bmi := *p.WeightKg / (hM * hM)
	bmi = math.Round(bmi*10) / 10
	switch {
	case bmi < 18.5:
		return bmi, "Дефицит массы тела", "underweight"
	case bmi < 25.0:
		return bmi, "Нормальный диапазон веса", "normal"
	case bmi < 30.0:
		return bmi, "Избыточная масса тела", "overweight"
	default:
		return bmi, "Высокий индекс массы (30+)", "obese"
	}
}

// IdealWeightRange рассчитывает здоровый диапазон веса для указанного роста (ИМТ 18.5 - 24.9).
func (p Profile) IdealWeightRange() (float64, float64) {
	if p.HeightCm == nil || *p.HeightCm <= 0 {
		return 0, 0
	}
	hM := *p.HeightCm / 100.0
	minKg := 18.5 * (hM * hM)
	maxKg := 24.9 * (hM * hM)
	return math.Round(minKg*10) / 10, math.Round(maxKg*10) / 10
}

// WaterTargetMl рассчитывает суточную норму чистой воды в мл (30–35 мл на 1 кг веса).
func (p Profile) WaterTargetMl() int {
	if p.WeightKg == nil || *p.WeightKg <= 0 {
		return 2000
	}
	target := int(*p.WeightKg * 32.5)
	if target < 1500 {
		target = 1500
	}
	if target > 4000 {
		target = 4000
	}
	return ((target + 25) / 50) * 50
}

func EquipmentRank(e string) int {
	switch e {
	case EquipNone:
		return 0
	case EquipHome:
		return 1
	case EquipGym:
		return 2
	default:
		return 0
	}
}

func LevelRank(l string) int {
	switch l {
	case LevelBeginner:
		return 0
	case LevelIntermediate:
		return 1
	case LevelAdvanced:
		return 2
	default:
		return 0
	}
}

func GoalLabel(g string) string {
	switch g {
	case GoalLose:
		return "Похудение"
	case GoalGain:
		return "Набор веса"
	default:
		return "Здоровый образ жизни"
	}
}

func GoalShort(g string) string {
	switch g {
	case GoalLose:
		return "похудение"
	case GoalGain:
		return "набор веса"
	default:
		return "поддержание формы"
	}
}

func LevelLabel(l string) string {
	switch l {
	case LevelBeginner:
		return "Новичок"
	case LevelIntermediate:
		return "Средний"
	case LevelAdvanced:
		return "Продвинутый"
	default:
		return "Любой"
	}
}

func EquipLabel(e string) string {
	switch e {
	case EquipNone:
		return "Без оборудования"
	case EquipHome:
		return "Домашний инвентарь"
	case EquipGym:
		return "Полноценный спортзал"
	default:
		return "Любое"
	}
}

func ActivityLabel(a string) string {
	switch a {
	case "sedentary":
		return "Сидячий (минимум движений)"
	case "light":
		return "Легкая (прогулки, быт)"
	case "moderate":
		return "Умеренная (тренировки 3-5 раз)"
	case "high":
		return "Высокая (тяжелый физ. труд)"
	default:
		return "Средняя"
	}
}

func TimeLabel(t string) string {
	switch t {
	case TimeMorning:
		return "Утро"
	case TimeDay:
		return "День"
	case TimeEvening:
		return "Вечер"
	default:
		return "В любое время"
	}
}

func PaceLabel(p string) string {
	switch p {
	case "gentle":
		return "Спокойный (~10% дефицит/профицит)"
	case "moderate":
		return "Умеренный (~15% дефицит/профицит)"
	case "steady":
		return "Активный (~20% дефицит/профицит)"
	default:
		return "Умеренный темп"
	}
}

func HealthLabel(h string) string {
	switch h {
	case HealthNone:
		return "Нет ограничений"
	case HealthInjury:
		return "Недавние травмы"
	case HealthChronic:
		return "Хронические заболевания"
	case HealthHeart:
		return "Болезни сердца / гипертония"
	case HealthPregnant:
		return "Беременность / восстановление"
	case HealthSurgery:
		return "Недавние операции"
	case HealthOther:
		return "Другие противопоказания"
	default:
		return h
	}
}

// User — всё, что приложение хранит про одного человека.
type User struct {
	Profile   Profile    `json:"profile"`
	Plan      *Plan      `json:"plan,omitempty"`
	Reminders []Reminder `json:"reminders,omitempty"`
	Progress  Progress   `json:"progress"`
}

// WaterToday возвращает количество выпитой воды сегодня в миллилитрах.
func (u *User) WaterToday() int {
	if u.Progress.Water == nil {
		return 0
	}
	today := time.Now().Format("2006-01-02")
	return u.Progress.Water[today]
}

// Reminder — настраиваемое напоминание.
type Reminder struct {
	ID      string `json:"id" db:"id"`
	Kind    string `json:"kind" db:"kind"` // workout | meal | water | sleep
	Title   string `json:"title" db:"title"`
	Time    string `json:"time" db:"time"` // HH:MM
	Days    string `json:"days" db:"days"` // daily | workouts
	Enabled bool   `json:"enabled" db:"enabled"`
	Note    string `json:"note,omitempty" db:"note"`
}

type WeightEntry struct {
	Date string  `json:"date" db:"date"`
	Kg   float64 `json:"kg" db:"kg"`
	Note string  `json:"note,omitempty" db:"note"`
}

type SessionEntry struct {
	Date      string `json:"date" db:"date"`
	DayIndex  int    `json:"day_index" db:"day_index"`
	Title     string `json:"title" db:"title"`
	Completed bool   `json:"completed" db:"completed"`
	Feeling   string `json:"feeling,omitempty" db:"feeling"`
	Energy    int    `json:"energy,omitempty" db:"energy"`
}

// Checkin — самочувствие, энергия и регулярность питания за день.
type Checkin struct {
	Date      string  `json:"date" db:"date"`
	Mood      int     `json:"mood" db:"mood"`
	Energy    int     `json:"energy" db:"energy"`
	Nutrition int     `json:"nutrition" db:"nutrition"`
	Sleep     float64 `json:"sleep,omitempty" db:"sleep"`
	Note      string  `json:"note,omitempty" db:"note"`
}

type Progress struct {
	Weights  []WeightEntry   `json:"weights,omitempty"`
	Sessions []SessionEntry  `json:"sessions,omitempty"`
	Checkins []Checkin       `json:"checkins,omitempty"`
	Meals    map[string]bool `json:"meals,omitempty"`
	Water    map[string]int  `json:"water,omitempty"` // date -> ml
}

// Plan — персональный план: тренировки, питание, режим и пояснения.
type Plan struct {
	ID        string    `json:"id"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Source    string    `json:"source"` // demo

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

// Exercise — запись из демонстрационного каталога упражнений.
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
