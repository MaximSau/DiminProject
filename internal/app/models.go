package app

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
	ID        string    `json:"id"`
	Code      string    `json:"code"` // 8-символьный код для синхронизации без пароля
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name            string   `json:"name"`
	Age             *int     `json:"age,omitempty"`
	HeightCm        *float64 `json:"height_cm,omitempty"`
	WeightKg        *float64 `json:"weight_kg,omitempty"`
	Sex             string   `json:"sex,omitempty"` // male | female | other
	Goal            string   `json:"goal"`          // lose | gain
	Pace            string   `json:"pace"`          // gentle | moderate | steady
	Level           string   `json:"level"`         // beginner | intermediate | advanced
	WorkoutsPerWeek int      `json:"workouts_per_week"`
	SessionMinutes  int      `json:"session_minutes"`
	Equipment       string   `json:"equipment"`
	PreferredTime   string   `json:"preferred_time"`
	ActivityLevel   string   `json:"activity_level"`
	DietPreferences []string `json:"diet_preferences,omitempty"`
	Allergies       []string `json:"allergies,omitempty"`
	AvoidFoods      string   `json:"avoid_foods,omitempty"`
	HealthFlags     []string `json:"health_flags,omitempty"`
	HealthNotes     string   `json:"health_notes,omitempty"`
	Consent         bool     `json:"consent"`
	Onboarded       bool     `json:"onboarded"`
	PlanVersion     int      `json:"plan_version"`
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

func levelRank(l string) int {
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

// Вспомогательные функции форматирования и лейблов

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
