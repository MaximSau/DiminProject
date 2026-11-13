package app

import "pulsefit/internal/models"

const (
	GoalLose = models.GoalLose
	GoalGain = models.GoalGain

	LevelBeginner     = models.LevelBeginner
	LevelIntermediate = models.LevelIntermediate
	LevelAdvanced     = models.LevelAdvanced

	EquipNone = models.EquipNone
	EquipHome = models.EquipHome
	EquipGym  = models.EquipGym

	TimeMorning = models.TimeMorning
	TimeDay     = models.TimeDay
	TimeEvening = models.TimeEvening

	HealthNone     = models.HealthNone
	HealthInjury   = models.HealthInjury
	HealthChronic  = models.HealthChronic
	HealthHeart    = models.HealthHeart
	HealthPregnant = models.HealthPregnant
	HealthSurgery  = models.HealthSurgery
	HealthOther    = models.HealthOther
)

type Exercise = models.Exercise
type Meal = models.Meal
type Plan = models.Plan
type PlanDay = models.PlanDay
type PlanExercise = models.PlanExercise
type MealDay = models.MealDay
type MealItem = models.MealItem
type Routine = models.Routine
type Nutrition = models.Nutrition

var WeekdayNames = models.WeekdayNames
var TodayIndex = models.TodayIndex

var GoalLabel = models.GoalLabel
var GoalShort = models.GoalShort
var LevelLabel = models.LevelLabel
var EquipLabel = models.EquipLabel
var ActivityLabel = models.ActivityLabel
var TimeLabel = models.TimeLabel
var PaceLabel = models.PaceLabel
var HealthLabel = models.HealthLabel
var EquipmentRank = models.EquipmentRank
var LevelRank = models.LevelRank
var BodyDataNote = models.BodyDataNote
var EstimateNutrition = models.EstimateNutrition
var WaterHint = models.WaterHint
