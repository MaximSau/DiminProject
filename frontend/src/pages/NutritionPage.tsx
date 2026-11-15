import React, { useState, useEffect } from 'react';
import {
  Apple,
  Flame,
  Droplets,
  Plus,
  CheckCircle2,
  Circle,
  Leaf,
  Coffee,
  Utensils,
  Cookie,
  Moon,
  RotateCw,
} from 'lucide-react';
import { DayMeals, MealItem, UserProfile } from '../types';
import { api } from '../api/client';
import { useToast } from '../context/ToastContext';

interface NutritionPageProps {
  initialDay?: number;
  onRefresh?: () => void;
}

export const NutritionPage: React.FC<NutritionPageProps> = ({ initialDay, onRefresh }) => {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);
  const [selectedDay, setSelectedDay] = useState(initialDay || 1);
  const [dayMeals, setDayMeals] = useState<DayMeals | null>(null);
  const [user, setUser] = useState<UserProfile | null>(null);
  const [waterIntake, setWaterIntake] = useState(0);
  const [waterTarget, setWaterTarget] = useState(2000);
  const [addingWater, setAddingWater] = useState(false);
  const [togglingId, setTogglingId] = useState<string | null>(null);

  const fetchNutrition = async (day: number) => {
    try {
      setLoading(true);
      const res = await api.getNutrition(day);
      setDayMeals(res.dayMeals);
      setUser(res.user);
      setWaterIntake(res.waterIntake);
      setWaterTarget(res.waterTarget || 2000);
      setSelectedDay(res.selectedDay);
    } catch (e: any) {
      showToast(e.message || 'Ошибка загрузки питания', 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNutrition(selectedDay);
  }, [selectedDay]);

  const handleToggleMeal = async (mealId: string, currentEaten: boolean) => {
    try {
      setTogglingId(mealId);
      const nextEaten = !currentEaten;
      await api.toggleMeal(selectedDay, mealId, nextEaten);
      showToast(nextEaten ? 'Блюдо отмечено съеденным' : 'Отметка снята');

      // Update state locally
      if (dayMeals) {
        const updateList = (items: MealItem[] = []) =>
          items.map(m => (m.id === mealId ? { ...m, eaten: nextEaten } : m));

        setDayMeals({
          ...dayMeals,
          meals: {
            breakfast: updateList(dayMeals.meals.breakfast),
            lunch: updateList(dayMeals.meals.lunch),
            snack: updateList(dayMeals.meals.snack),
            dinner: updateList(dayMeals.meals.dinner),
          },
        });
      }
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка изменения статуса блюда', 'error');
    } finally {
      setTogglingId(null);
    }
  };

  const handleAddWater = async (delta = 330) => {
    try {
      setAddingWater(true);
      const res = await api.addWater(delta);
      setWaterIntake(res.current);
      showToast(`+${delta} мл воды добавлено`);
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка добавления воды', 'error');
    } finally {
      setAddingWater(false);
    }
  };

  const weekdayShort = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

  const mealSections = [
    { key: 'breakfast', title: 'Завтрак', icon: Coffee, desc: 'Энергетический старт дня' },
    { key: 'lunch', title: 'Обед', icon: Utensils, desc: 'Основной приём пищи и баланс нутриентов' },
    { key: 'snack', title: 'Полдник / Перекус', icon: Cookie, desc: 'Поддержание метаболизма и сытости' },
    { key: 'dinner', title: 'Ужин', icon: Moon, desc: 'Лёгкое восстановление и качественный белок' },
  ];

  return (
    <div className="space-y-6">
      {/* 7-Day Selector Tabs */}
      <div className="grid grid-cols-7 gap-1.5 sm:gap-2 p-1.5 rounded-2xl bg-[#111827]/80 border border-slate-800 backdrop-blur-sm">
        {[1, 2, 3, 4, 5, 6, 7].map(day => (
          <button
            key={day}
            onClick={() => setSelectedDay(day)}
            className={`flex flex-col items-center justify-center py-2.5 px-1 rounded-xl text-xs font-bold transition-all duration-150 ${
              selectedDay === day
                ? 'bg-emerald-500 text-slate-950 shadow-md shadow-emerald-500/20 scale-[1.02]'
                : 'text-slate-400 hover:text-white hover:bg-slate-800/60'
            }`}
          >
            <span className="text-[10px] uppercase font-semibold opacity-75">
              {weekdayShort[day - 1]}
            </span>
            <span className="text-sm font-black mt-0.5">День {day}</span>
          </button>
        ))}
      </div>

      {loading ? (
        <div className="py-20 text-center text-slate-400">Загрузка рациона...</div>
      ) : dayMeals ? (
        <div className="space-y-6">
          {/* Top Banner: Targets & Macros */}
          <div className="rounded-2xl p-6 border bg-[#111827]/80 border-slate-800 shadow-sm space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800/80 pb-4">
              <div>
                <span className="text-xs font-bold uppercase tracking-wider text-emerald-400">
                  {dayMeals.weekdayName} • День {selectedDay}
                </span>
                <h2 className="text-2xl font-black text-white mt-1">
                  Рацион на день
                </h2>
                {user?.isVegan && (
                  <span className="inline-flex items-center gap-1.5 text-xs font-bold px-2.5 py-0.5 rounded-full bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 mt-1">
                    <Leaf className="w-3.5 h-3.5" />
                    Веганское меню (100% растительный рацион)
                  </span>
                )}
              </div>

              {/* Water in 330ml glasses */}
              <div className="flex items-center gap-3 bg-slate-900/60 p-3 rounded-xl border border-slate-800">
                <div className="p-2 rounded-lg bg-sky-500/15 text-sky-400">
                  <Droplets className="w-5 h-5" />
                </div>
                <div>
                  <div className="text-xs text-slate-400 font-medium">Вода (1 стакан = 330 мл)</div>
                  <div className="text-sm font-bold text-white">
                    {waterIntake} / {waterTarget} мл ({Math.floor(waterIntake / 330)} из {Math.ceil(waterTarget / 330)} стак.)
                  </div>
                </div>
                <button
                  onClick={() => handleAddWater(330)}
                  disabled={addingWater}
                  className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs font-bold bg-sky-500/20 hover:bg-sky-500/30 text-sky-300 border border-sky-500/40 transition-all active:scale-95 disabled:opacity-50"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>+330 мл</span>
                </button>
              </div>
            </div>

            {/* Macros Bar */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div className="p-3.5 rounded-xl bg-slate-900/60 border border-slate-800/80">
                <span className="text-xs font-semibold text-slate-400 block mb-1">Калории</span>
                <div className="text-xl font-black text-white">
                  {dayMeals.currentCalories || 0} <span className="text-xs text-slate-400 font-normal">/ {dayMeals.targetCalories} ккал</span>
                </div>
              </div>
              <div className="p-3.5 rounded-xl bg-slate-900/60 border border-slate-800/80">
                <span className="text-xs font-semibold text-emerald-400 block mb-1">Белки</span>
                <div className="text-xl font-black text-white">
                  {dayMeals.targetProteins} г
                </div>
              </div>
              <div className="p-3.5 rounded-xl bg-slate-900/60 border border-slate-800/80">
                <span className="text-xs font-semibold text-amber-400 block mb-1">Жиры</span>
                <div className="text-xl font-black text-white">
                  {dayMeals.targetFats} г
                </div>
              </div>
              <div className="p-3.5 rounded-xl bg-slate-900/60 border border-slate-800/80">
                <span className="text-xs font-semibold text-sky-400 block mb-1">Углеводы</span>
                <div className="text-xl font-black text-white">
                  {dayMeals.targetCarbs} г
                </div>
              </div>
            </div>
          </div>

          {/* Meals list by Section */}
          <div className="space-y-6">
            {mealSections.map(sec => {
              const items: MealItem[] = (dayMeals.meals as any)?.[sec.key] || [];
              const Icon = sec.icon;

              return (
                <div key={sec.key} className="space-y-3">
                  <div className="flex items-center gap-2.5">
                    <div className="p-2 rounded-xl bg-slate-800/80 text-emerald-400 border border-slate-700/60">
                      <Icon className="w-4 h-4" />
                    </div>
                    <div>
                      <h3 className="text-base font-bold text-white">{sec.title}</h3>
                      <p className="text-xs text-slate-400">{sec.desc}</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    {items.length > 0 ? (
                      items.map(m => (
                        <div
                          key={m.id}
                          className={`rounded-2xl p-4 border transition-all duration-200 flex items-start justify-between gap-3 ${
                            m.eaten
                              ? 'bg-slate-900/40 border-slate-800/60 opacity-60'
                              : 'bg-[#111827]/80 border-slate-800 hover:border-slate-700 shadow-sm'
                          }`}
                        >
                          <div className="space-y-1">
                            <div className="flex items-center gap-2">
                              <h4 className={`text-sm font-bold text-white ${m.eaten ? 'line-through text-slate-400' : ''}`}>
                                {m.name}
                              </h4>
                              {m.isVegan && (
                                <span className="p-0.5 rounded bg-emerald-500/10 text-emerald-400" title="Веган">
                                  <Leaf className="w-3 h-3" />
                                </span>
                              )}
                            </div>
                            <p className="text-xs text-slate-400 leading-relaxed">
                              {m.description}
                            </p>
                            <div className="flex items-center gap-2 pt-1 text-[11px] text-slate-400 font-medium">
                              <span className="text-emerald-400 font-bold">{m.calories} ккал</span>
                              <span>• Б: {m.proteins}г</span>
                              <span>• Ж: {m.fats}г</span>
                              <span>• У: {m.carbs}г</span>
                            </div>
                          </div>

                          <button
                            onClick={() => handleToggleMeal(m.id, m.eaten)}
                            disabled={togglingId === m.id}
                            className={`p-2 rounded-xl transition-all flex-shrink-0 ${
                              m.eaten
                                ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40'
                                : 'bg-slate-800/80 text-slate-400 hover:text-white hover:bg-slate-700 border border-slate-700'
                            }`}
                            title={m.eaten ? 'Отменить отметку' : 'Отметить съеденным'}
                          >
                            {m.eaten ? (
                              <CheckCircle2 className="w-5 h-5 text-emerald-400" />
                            ) : (
                              <Circle className="w-5 h-5" />
                            )}
                          </button>
                        </div>
                      ))
                    ) : (
                      <div className="col-span-2 p-4 rounded-xl bg-slate-900/30 border border-slate-800/40 text-xs text-slate-500 text-center">
                        Блюда не указаны
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      ) : (
        <div className="py-20 text-center text-slate-400">Данные рациона недоступны</div>
      )}
    </div>
  );
};
