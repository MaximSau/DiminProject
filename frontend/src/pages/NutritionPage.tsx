import React, { useState, useEffect } from 'react';
import {
  Droplets,
  Plus,
  Minus,
  CheckCircle2,
  Circle,
  Leaf,
  Coffee,
  Utensils,
  Cookie,
  Moon,
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

  const handleAddWater = async (delta: number) => {
    try {
      setAddingWater(true);
      const res = await api.addWater(delta);
      setWaterIntake(res.current);
      showToast(delta > 0 ? `+${delta} мл воды добавлено` : `${delta} мл воды списано`);
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка обновления водного баланса', 'error');
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
      <div className="grid grid-cols-7 gap-1.5 sm:gap-2 p-1.5 rounded-2xl bg-[#121216] border border-white/[0.08] backdrop-blur-sm">
        {[1, 2, 3, 4, 5, 6, 7].map(day => (
          <button
            key={day}
            onClick={() => setSelectedDay(day)}
            className={`flex flex-col items-center justify-center py-2.5 px-1 rounded-xl text-xs font-bold transition-all duration-150 ${
              selectedDay === day
                ? 'bg-emerald-500 text-slate-950 shadow-md shadow-emerald-500/20 scale-[1.02]'
                : 'text-slate-300 hover:text-white hover:bg-[#1a1a22]'
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
        <div className="py-20 text-center text-slate-300">Загрузка рациона...</div>
      ) : dayMeals ? (
        <div className="space-y-6">
          {/* Top Banner: Targets & Macros */}
          <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-white/[0.08] pb-4">
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

              {/* Water in 330ml glasses with + and - buttons */}
              <div className="flex items-center gap-3 bg-[#17171d] p-3 rounded-xl border border-white/[0.08]">
                <div className="p-2 rounded-lg bg-sky-500/15 text-sky-400 border border-sky-500/20">
                  <Droplets className="w-5 h-5" />
                </div>
                <div>
                  <div className="text-xs text-slate-300 font-medium">Вода (1 стакан = 330 мл)</div>
                  <div className="text-sm font-bold text-white">
                    {waterIntake} / {waterTarget} мл ({Math.floor(waterIntake / 330)} из {Math.ceil(waterTarget / 330)} стак.)
                  </div>
                </div>
                <div className="flex items-center gap-1.5">
                  <button
                    onClick={() => handleAddWater(-330)}
                    disabled={addingWater || waterIntake <= 0}
                    className="flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-bold bg-[#1c1c24] hover:bg-[#252530] text-slate-200 border border-white/[0.10] transition-all active:scale-95 disabled:opacity-40"
                    title="Убрать 1 стакан воды (-330 мл)"
                  >
                    <Minus className="w-3.5 h-3.5" />
                    <span>-330</span>
                  </button>
                  <button
                    onClick={() => handleAddWater(330)}
                    disabled={addingWater}
                    className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs font-bold bg-sky-500/20 hover:bg-sky-500/30 text-sky-300 border border-sky-500/40 transition-all active:scale-95 disabled:opacity-50"
                    title="Добавить 1 стакан воды (+330 мл)"
                  >
                    <Plus className="w-3.5 h-3.5" />
                    <span>+330</span>
                  </button>
                </div>
              </div>
            </div>

            {/* Macros Bar */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div className="p-3.5 rounded-xl bg-[#17171d] border border-white/[0.06]">
                <span className="text-xs font-semibold text-slate-300 block mb-1">Калории</span>
                <div className="text-xl font-black text-white">
                  {dayMeals.currentCalories || 0} <span className="text-xs text-slate-400 font-normal">/ {dayMeals.targetCalories} ккал</span>
                </div>
              </div>
              <div className="p-3.5 rounded-xl bg-[#17171d] border border-white/[0.06]">
                <span className="text-xs font-semibold text-emerald-400 block mb-1">Белки</span>
                <div className="text-xl font-black text-white">
                  {dayMeals.targetProteins} г
                </div>
              </div>
              <div className="p-3.5 rounded-xl bg-[#17171d] border border-white/[0.06]">
                <span className="text-xs font-semibold text-amber-400 block mb-1">Жиры</span>
                <div className="text-xl font-black text-white">
                  {dayMeals.targetFats} г
                </div>
              </div>
              <div className="p-3.5 rounded-xl bg-[#17171d] border border-white/[0.06]">
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
                    <div className="p-2 rounded-xl bg-[#17171d] text-emerald-400 border border-white/[0.08]">
                      <Icon className="w-4 h-4" />
                    </div>
                    <div>
                      <h3 className="text-base font-bold text-white">{sec.title}</h3>
                      <p className="text-xs text-slate-300">{sec.desc}</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    {items.length > 0 ? (
                      items.map(m => (
                        <div
                          key={m.id}
                          className={`rounded-2xl p-4 border transition-all duration-200 flex items-start justify-between gap-3 ${
                            m.eaten
                              ? 'bg-[#121216]/50 border-white/[0.04] opacity-50'
                              : 'bg-[#121216] border-white/[0.08] hover:border-white/[0.18] shadow-sm'
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
                            <p className="text-xs text-slate-300 leading-relaxed font-normal">
                              {m.description}
                            </p>
                            <div className="flex items-center gap-2 pt-1 text-[11px] text-slate-300 font-medium">
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
                                : 'bg-[#17171d] text-slate-300 hover:text-white hover:bg-[#22222c] border border-white/[0.08]'
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
                      <div className="col-span-2 p-4 rounded-xl bg-[#121216]/60 border border-white/[0.06] text-xs text-slate-400 text-center">
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
        <div className="py-20 text-center text-slate-300">Данные рациона недоступны</div>
      )}
    </div>
  );
};
