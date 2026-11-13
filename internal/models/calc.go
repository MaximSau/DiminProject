package models

import (
	"math"
	"strings"
)

// Nutrition — ориентировочные суточные ориентиры по калориям и БЖУ.
type Nutrition struct {
	Kcal     int
	ProteinG int
	FatG     int
	CarbsG   int
}

// EstimateNutrition считает BMR по Миффлину — Сан Жеору, умножает на
// коэффициент активности и добавляет безопасную поправку под цель.
func EstimateNutrition(p Profile) Nutrition {
	if !p.HasBodyData() {
		return Nutrition{}
	}
	age := float64(*p.Age)
	height := *p.HeightCm
	weight := *p.WeightKg

	// BMR: для мужчин и женщин разный коэффициент.
	var bmr float64
	switch p.Sex {
	case "male":
		bmr = 10*weight + 6.25*height - 5*age + 5
	case "female":
		bmr = 10*weight + 6.25*height - 5*age - 161
	default:
		// Пол не указан — берём среднее между формулами.
		bmr = 10*weight + 6.25*height - 5*age - 78
	}

	// Коэффициент повседневной активности.
	factor := 1.375 // низкая активность + 3 тренировки
	switch p.ActivityLevel {
	case "medium":
		factor = 1.55
	case "high":
		factor = 1.725
	}
	tdee := bmr * factor

	// Безопасная поправка под цель: не более 15–20% и без экстремальных значений.
	var target float64
	switch p.Goal {
	case GoalGain:
		target = tdee * surplusFactor(p.Pace)
	default:
		target = tdee * deficitFactor(p.Pace)
	}

	// Нижняя граница: не ниже базального обмена, и не ниже 1200/1500 ккал
	floor := 1200.0
	if p.Sex == "male" {
		floor = 1500.0
	}
	if target < floor {
		target = floor
	}
	if target < bmr {
		target = bmr
	}

	kcal := int(math.Round(target / 10 * 10))
	return Nutrition{
		Kcal:     kcal,
		ProteinG: int(math.Round(weight * proteinPerKg(p.Goal) / 5 * 5)),
		FatG:     int(math.Round(target * 0.27 / 9 / 5 * 5)),
		CarbsG:   int(math.Round(target * 0.45 / 4 / 5 * 5)),
	}
}

func deficitFactor(pace string) float64 {
	switch pace {
	case "gentle":
		return 0.90
	case "steady":
		return 0.80
	default:
		return 0.85
	}
}

func surplusFactor(pace string) float64 {
	switch pace {
	case "gentle":
		return 1.05
	case "steady":
		return 1.15
	default:
		return 1.10
	}
}

func proteinPerKg(goal string) float64 {
	if goal == GoalGain {
		return 1.8
	}
	return 1.6
}

// WaterHint — ориентир по воде, если известен вес.
func WaterHint(p Profile) string {
	if p.WeightKg == nil {
		return ""
	}
	ml := int(math.Round(*p.WeightKg * 30 / 50 * 50))
	return "Ориентир по воде — около " + itoa(ml) + " мл в день, включая жидкость из еды. В жару и на тренировке — больше."
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b strings.Builder
	var rev []byte
	for n > 0 {
		rev = append(rev, byte('0'+n%10))
		n /= 10
	}
	for i := len(rev) - 1; i >= 0; i-- {
		b.WriteByte(rev[i])
	}
	return b.String()
}
