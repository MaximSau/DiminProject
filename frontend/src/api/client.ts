import {
  UserProfile,
  WorkoutDay,
  DayMeals,
  WeightEntry,
  CatalogExercise,
  DashboardResponse,
  PlanExercise,
} from '../types';

async function request<T>(url: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers || {});
  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }
  headers.set('X-Requested-With', 'XMLHttpRequest');

  const res = await fetch(url, {
    ...options,
    headers,
  });

  if (!res.ok) {
    let errMsg = `Ошибка сервера (${res.status})`;
    try {
      const errData = await res.json();
      if (errData && errData.error) errMsg = errData.error;
    } catch {
      // non-JSON error
    }
    throw new Error(errMsg);
  }

  return res.json();
}

export const api = {
  // Bootstrap & User status
  async getBootstrap(): Promise<{ onboarded: boolean; user: UserProfile | null; todayIndex: number }> {
    return request('/api/bootstrap');
  },

  // Onboarding
  async submitOnboarding(data: Record<string, any>): Promise<{ ok: boolean; user: UserProfile; hasMedicalFlags: boolean }> {
    return request('/api/onboarding', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // Dashboard
  async getDashboard(): Promise<DashboardResponse> {
    return request('/api/dashboard');
  },

  // Workouts
  async getWorkouts(day?: number): Promise<{
    user: UserProfile;
    selectedDay: number;
    workout: WorkoutDay;
    allDays: WorkoutDay[];
  }> {
    const q = day !== undefined ? `?day=${day}` : '';
    return request(`/api/workouts${q}`);
  },

  async completeWorkout(day: number, completed = true, feeling = 'good', energy = 8): Promise<{ ok: boolean; message: string }> {
    return request('/api/workouts/complete', {
      method: 'POST',
      body: JSON.stringify({ day, completed, feeling, energy }),
    });
  },

  async swapExercise(day: number, exerciseId: string): Promise<{ ok: boolean; newExercise: PlanExercise }> {
    return request('/api/workouts/swap', {
      method: 'POST',
      body: JSON.stringify({ day, exerciseId }),
    });
  },

  // Nutrition
  async getNutrition(day?: number): Promise<{
    user: UserProfile;
    selectedDay: number;
    dayMeals: DayMeals;
    waterIntake: number;
    waterTarget: number;
  }> {
    const q = day !== undefined ? `?day=${day}` : '';
    return request(`/api/nutrition${q}`);
  },

  async toggleMeal(day: number, mealId: string, eaten: boolean): Promise<{ ok: boolean; eaten: boolean }> {
    return request('/api/nutrition/toggle', {
      method: 'POST',
      body: JSON.stringify({ day, mealId, eaten }),
    });
  },

  async addWater(delta: number): Promise<{ ok: boolean; current: number; target: number; glasses: number }> {
    return request('/api/water/add', {
      method: 'POST',
      body: JSON.stringify({ delta }),
    });
  },

  // Progress
  async getProgress(): Promise<{
    user: UserProfile;
    weightHistory: WeightEntry[];
    workoutHistory: any[];
    streak: number;
  }> {
    return request('/api/progress');
  },

  async addWeight(weight: number, note: string): Promise<{ ok: boolean; entry: WeightEntry }> {
    return request('/api/progress/weight', {
      method: 'POST',
      body: JSON.stringify({ weight, note }),
    });
  },

  // Catalog
  async getCatalog(params: { q?: string; muscle?: string; equipment?: string; difficulty?: string } = {}): Promise<{
    exercises: CatalogExercise[];
    muscles: string[];
    equipments: string[];
  }> {
    const searchParams = new URLSearchParams();
    if (params.q) searchParams.set('q', params.q);
    if (params.muscle) searchParams.set('muscle', params.muscle);
    if (params.equipment) searchParams.set('equipment', params.equipment);
    if (params.difficulty) searchParams.set('difficulty', params.difficulty);
    const qs = searchParams.toString() ? `?${searchParams.toString()}` : '';
    return request(`/api/catalog${qs}`);
  },

  async getExerciseDetail(id: string): Promise<{ exercise: CatalogExercise }> {
    return request(`/api/catalog/exercise?id=${encodeURIComponent(id)}`);
  },

  // Profile
  async getProfile(): Promise<{ user: UserProfile }> {
    return request('/api/profile');
  },

  async updateProfile(profile: Partial<UserProfile>): Promise<{ ok: boolean; user: UserProfile }> {
    return request('/api/profile', {
      method: 'POST',
      body: JSON.stringify(profile),
    });
  },

  // Plan export
  async getPlanExport(): Promise<{ user: UserProfile; days: WorkoutDay[]; meals: DayMeals[] }> {
    return request('/api/plan/export');
  },

  // Demo actions
  async seedDemo(): Promise<{ ok: boolean }> {
    return request('/api/demo/seed', { method: 'POST' });
  },

  async resetDemo(): Promise<{ ok: boolean }> {
    return request('/api/demo/reset', { method: 'POST' });
  },
};
