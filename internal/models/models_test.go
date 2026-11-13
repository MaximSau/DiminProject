package models

import (
	"strings"
	"testing"
)

func TestModels(t *testing.T) {
	// 1. BMI calculation
	h := 175.0
	w := 70.0
	p := Profile{
		HeightCm: &h,
		WeightKg: &w,
	}

	bmi, label, status := p.BMI()
	if bmi <= 0 || status != "normal" || !strings.Contains(label, "Нормальный") {
		t.Errorf("unexpected BMI results: %f, %s, %s", bmi, label, status)
	}

	minW, maxW := p.IdealWeightRange()
	if minW <= 0 || maxW <= minW {
		t.Errorf("invalid ideal weight range: %f - %f", minW, maxW)
	}

	// 2. Water Target
	water := p.WaterTargetMl()
	if water < 1500 || water > 4000 {
		t.Errorf("invalid water target: %d", water)
	}

	// 3. Medical flags
	p.HealthFlags = []string{HealthInjury}
	if !p.MedicalBlocked() {
		t.Errorf("expected medical block for injury")
	}

	p.HealthFlags = []string{HealthNone}
	if p.MedicalBlocked() {
		t.Errorf("expected no block for HealthNone")
	}

	// 4. Labels
	if GoalLabel(GoalLose) != "Похудение" {
		t.Errorf("unexpected GoalLabel")
	}
	if LevelLabel(LevelBeginner) != "Новичок" {
		t.Errorf("unexpected LevelLabel")
	}
	if EquipLabel(EquipNone) != "Без оборудования" {
		t.Errorf("unexpected EquipLabel")
	}
}
