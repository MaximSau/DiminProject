export interface UserProfile {
  id: string;
  code: string;
  name: string;
  gender: string;
  age: number;
  height: number;
  weight: number;
  targetWeight: number;
  goal: string;
  level: string;
  equipment: string;
  daysPerWeek: number;
  workoutDuration: number;
  preferredTime: string;
  pace: string;
  isVegan: boolean;
  healthIssues: string[];
  hasHealthIssues: boolean;
  bmi: number;
  bmr: number;
  tdee: number;
  calorieTarget: number;
  proteinTarget: number;
  fatTarget: number;
  carbTarget: number;
  waterTarget: number;
  onboarded: boolean;
  planVersion: int;
  createdAt: string;
  updatedAt: string;
}

export type int = number;

export interface PlanExercise {
  id: string;
  title: string;
  muscle: string;
  equipment: string;
  difficulty: string;
  sets: number;
  reps: string;
  durationSec: number;
  restSec: number;
  coachTip: string;
  isGentle?: boolean;
  isContraindicated?: boolean;
}

export interface WorkoutDay {
  dayIndex: number;
  weekdayName: string;
  dayName: string;
  focus: string;
  isRest: boolean;
  estimatedMinutes: number;
  completed: boolean;
  notes?: string;
  exercises: PlanExercise[];
}

export interface MealItem {
  id: string;
  type: string;
  name: string;
  description: string;
  calories: number;
  proteins: number;
  fats: number;
  carbs: number;
  isVegan: boolean;
  eaten: boolean;
}

export interface DayMeals {
  dayIndex: number;
  weekdayName: string;
  targetCalories: number;
  targetProteins: number;
  targetFats: number;
  targetCarbs: number;
  currentCalories: number;
  currentProteins: number;
  currentFats: number;
  currentCarbs: number;
  meals: {
    breakfast: MealItem[];
    lunch: MealItem[];
    snack: MealItem[];
    dinner: MealItem[];
  };
}

export interface WeightEntry {
  date: string;
  weight: number;
  note: string;
}

export interface WorkoutHistoryEntry {
  date: string;
  dayIndex: number;
  dayName: string;
  focus: string;
  exercisesCount: number;
}

export interface Reminder {
  id: string;
  type: string;
  title: string;
  time: string;
  enabled: boolean;
}

export interface CatalogExercise {
  id: string;
  title: string;
  muscle: string;
  secondaryMuscles?: string[];
  equipment: string;
  difficulty: string;
  description?: string;
  coachTip: string;
  isGentle?: boolean;
  isContraindicated?: boolean;
}

export interface CatalogMeal {
  id: string;
  name: string;
  slot: string;
  calories: number;
  proteins: number;
  fats: number;
  carbs: number;
  portion: string;
  ingredients: string[];
  prep?: string;
  tags?: string[];
  isVegan: boolean;
}

export interface DashboardResponse {
  user: UserProfile;
  todayWorkout: WorkoutDay;
  todayMeals: DayMeals;
  todayWater: number;
  streak: number;
  weightHistory: WeightEntry[];
  workoutHistory: WorkoutHistoryEntry[];
  completedDates?: string[];
  isGentleMode: boolean;
  reminders: Reminder[];
}
