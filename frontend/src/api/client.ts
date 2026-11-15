import {
  UserProfile,
  WorkoutDay,
  DayMeals,
  WeightEntry,
  CatalogExercise,
  CatalogMeal,
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
      if (errData && errData.notOnboarded) {
        window.dispatchEvent(new CustomEvent("pulsefit:reset"));
      }
    } catch {
      // non-JSON error
    }
    if (res.status === 404 || res.status === 401) {
      window.dispatchEvent(new CustomEvent("pulsefit:reset"));
    }
    throw new Error(errMsg);
  }

  const data = await res.json();
  if (data && data.notOnboarded) {
    window.dispatchEvent(new CustomEvent("pulsefit:reset"));
  }
  return data;
}

export const api = {
  // Bootstrap & User status
  async getBootstrap(): Promise<{ onboarded: boolean; user: UserProfile | null; todayIndex: number }> {
    return request('/api/bootstrap');
  },

  // Onboarding
  async submitOnboarding(data: {
    name?: string;
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
    isVegan: boolean;
    healthIssues: string[];
  }): Promise<{ ok: boolean; user: UserProfile }> {
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
    workout: WorkoutDay;
    allDays: WorkoutDay[];
    completedDates: string[];
    user: UserProfile;
    selectedDay: number;
  }> {
    const qs = day ? `?day=${day}` : '';
    return request(`/api/workouts${qs}`);
  },

  async completeWorkout(day: number, date?: string, completed: boolean = true): Promise<{ ok: boolean; completed: boolean; date?: string }> {
    return request('/api/workouts/complete', {
      method: 'POST',
      body: JSON.stringify({ day, date, completed }),
    });
  },

  async swapExercise(day: number, exerciseId: string): Promise<{ ok: boolean; newExercise: PlanExercise }> {
    return request('/api/workouts/swap', {
      method: 'POST',
      body: JSON.stringify({ day, exerciseId }),
    });
  },

  async rescheduleWorkout(fromDay: number, toDay: number): Promise<{ ok: boolean; fromDay: number; toDay: number }> {
    return request('/api/workouts/reschedule', {
      method: 'POST',
      body: JSON.stringify({ fromDay, toDay }),
    });
  },

  async saveWorkoutNote(day: number, notes: string): Promise<{ ok: boolean }> {
    return request('/api/workouts/note', {
      method: 'POST',
      body: JSON.stringify({ day, notes }),
    });
  },

  // Nutrition
  async getNutrition(day?: number): Promise<{
    dayMeals: DayMeals;
    user: UserProfile;
    waterIntake: number;
    waterTarget: number;
    selectedDay: number;
  }> {
    const qs = day ? `?day=${day}` : '';
    return request(`/api/nutrition${qs}`);
  },

  async toggleMeal(day: number, mealId: string, eaten: boolean): Promise<{ ok: boolean; eaten: boolean }> {
    return request('/api/nutrition/toggle', {
      method: 'POST',
      body: JSON.stringify({ day, mealId, eaten }),
    });
  },

  async addMealToDay(day: number, mealId: string): Promise<{ ok: boolean; item: any }> {
    return request('/api/nutrition/add', {
      method: 'POST',
      body: JSON.stringify({ day, mealId }),
    });
  },

  async addWater(delta: number = 330): Promise<{ ok: boolean; current: number; delta: number }> {
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
  async getCatalog(params: {
    q?: string;
    muscle?: string;
    equipment?: string;
    difficulty?: string;
  } = {}): Promise<{ exercises: CatalogExercise[] }> {
    const searchParams = new URLSearchParams();
    if (params.q) searchParams.set('q', params.q);
    if (params.muscle) searchParams.set('muscle', params.muscle);
    if (params.equipment) searchParams.set('equipment', params.equipment);
    if (params.difficulty) searchParams.set('difficulty', params.difficulty);
    const qs = searchParams.toString() ? `?${searchParams.toString()}` : '';
    return request(`/api/catalog${qs}`);
  },

  async getCatalogMeals(params: {
    q?: string;
    slot?: string;
    vegan?: boolean;
  } = {}): Promise<{ meals: CatalogMeal[] }> {
    const searchParams = new URLSearchParams();
    if (params.q) searchParams.set('q', params.q);
    if (params.slot) searchParams.set('slot', params.slot);
    if (params.vegan) searchParams.set('vegan', 'true');
    const qs = searchParams.toString() ? `?${searchParams.toString()}` : '';
    return request(`/api/catalog/meals${qs}`);
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
    const res = await request<{ ok: boolean }>('/api/demo/reset', { method: 'POST' });
    window.dispatchEvent(new CustomEvent("pulsefit:reset"));
    return res;
  },
};
