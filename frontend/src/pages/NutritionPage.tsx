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
  Calendar as CalendarIcon,
  ChevronLeft,
  ChevronRight,
  X,
  AlertTriangle,
  Search,
} from 'lucide-react';
import { DayMeals, MealItem, CatalogMeal, UserProfile } from '../types';
import { api } from '../api/client';
import { useToast } from '../context/ToastContext';

interface NutritionPageProps {
  initialDay?: number;
  onRefresh?: () => void;
}

export const NutritionPage: React.FC<NutritionPageProps> = ({ initialDay, onRefresh }) => {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);

  // Default to today's ISO day (Monday = 1 ... Sunday = 7)
  const today = new Date();
  const todayIso = today.getDay() === 0 ? 7 : today.getDay();
  const [selectedDay, setSelectedDay] = useState(initialDay || todayIso);

  const [dayMeals, setDayMeals] = useState<DayMeals | null>(null);
  const [user, setUser] = useState<UserProfile | null>(null);
  const [waterIntake, setWaterIntake] = useState(0);
  const [waterTarget, setWaterTarget] = useState(2000);
  const [addingWater, setAddingWater] = useState(false);
  const [togglingId, setTogglingId] = useState<string | null>(null);

  // Unified Calendar state
  const [weekOffset, setWeekOffset] = useState(0);
  const [calendarModalOpen, setCalendarModalOpen] = useState(false);
  const [calendarYear, setCalendarYear] = useState(() => new Date().getFullYear());
  const [calendarMonth, setCalendarMonth] = useState(() => new Date().getMonth());

  // Meal Catalog Modal state
  const [catalogModalOpen, setCatalogModalOpen] = useState(false);
  const [catalogMeals, setCatalogMeals] = useState<CatalogMeal[]>([]);
  const [loadingCatalog, setLoadingCatalog] = useState(false);
  const [catalogSearch, setCatalogSearch] = useState('');
  const [catalogSlot, setCatalogSlot] = useState('');
  const [catalogVeganOnly, setCatalogVeganOnly] = useState(false);
  const [addingCatalogId, setAddingCatalogId] = useState<string | null>(null);

  // Big Red Exclamation Warning state for Vegan user
  const [veganWarningModal, setVeganWarningModal] = useState<{
    mealName: string;
    onConfirm: () => void;
  } | null>(null);

  const todayMidnight = new Date(today.getFullYear(), today.getMonth(), today.getDate());

  const monthsRu = [
    'января', 'февраля', 'марта', 'апреля', 'мая', 'июня',
    'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря'
  ];
  const monthsRuNominative = [
    'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
    'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
  ];
  const monthsShortRu = [
    'янв', 'фев', 'мар', 'апр', 'мая', 'июн',
    'июл', 'авг', 'сен', 'окт', 'ноя', 'дек'
  ];
  const weekdayShort = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

  const getMondayOfWeek = (offset: number) => {
    const d = new Date();
    const isoDay = d.getDay() === 0 ? 7 : d.getDay();
    d.setDate(d.getDate() - (isoDay - 1) + offset * 7);
    d.setHours(0, 0, 0, 0);
    return d;
  };

  const getDateForDay = (dayIndex: number, offset: number) => {
    const monday = getMondayOfWeek(offset);
    const target = new Date(monday);
    target.setDate(monday.getDate() + (dayIndex - 1));
    target.setHours(0, 0, 0, 0);
    return target;
  };

  const isSameDay = (d1: Date, d2: Date) => {
    return (
      d1.getFullYear() === d2.getFullYear() &&
      d1.getMonth() === d2.getMonth() &&
      d1.getDate() === d2.getDate()
    );
  };

  const activeDate = getDateForDay(selectedDay, weekOffset);

  // Week range label
  const mondayDate = getDateForDay(1, weekOffset);
  const sundayDate = getDateForDay(7, weekOffset);
  const weekLabel = `${mondayDate.getDate()} ${monthsShortRu[mondayDate.getMonth()]} – ${sundayDate.getDate()} ${monthsShortRu[sundayDate.getMonth()]} ${sundayDate.getFullYear()}`;

  // Calendar calculations
  const firstDayOfMonth = new Date(calendarYear, calendarMonth, 1);
  const daysInMonth = new Date(calendarYear, calendarMonth + 1, 0).getDate();
  const monthStartOffset = (firstDayOfMonth.getDay() + 6) % 7;

  const handleSelectCalendarDate = (dayNum: number) => {
    const targetDate = new Date(calendarYear, calendarMonth, dayNum);
    const targetIsoDay = targetDate.getDay() === 0 ? 7 : targetDate.getDay();

    const targetMonday = new Date(targetDate);
    targetMonday.setDate(targetDate.getDate() - (targetIsoDay - 1));
    targetMonday.setHours(0, 0, 0, 0);

    const baseMonday = getMondayOfWeek(0);
    const diffWeeks = Math.round((targetMonday.getTime() - baseMonday.getTime()) / (7 * 24 * 3600 * 1000));

    setWeekOffset(diffWeeks);
    setSelectedDay(targetIsoDay);
    setCalendarModalOpen(false);
    showToast(`Выбрана дата: ${dayNum} ${monthsRu[calendarMonth]} ${calendarYear}`);
  };

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

  const fetchCatalogMeals = async () => {
    try {
      setLoadingCatalog(true);
      const res = await api.getCatalogMeals({
        q: catalogSearch || undefined,
        slot: catalogSlot || undefined,
        vegan: catalogVeganOnly || undefined,
      });
      setCatalogMeals(res.meals || []);
    } catch (e: any) {
      showToast(e.message || 'Ошибка загрузки каталога блюд', 'error');
    } finally {
      setLoadingCatalog(false);
    }
  };

  useEffect(() => {
    fetchNutrition(selectedDay);
  }, [selectedDay]);

  useEffect(() => {
    if (catalogModalOpen) {
      fetchCatalogMeals();
    }
  }, [catalogModalOpen, catalogSearch, catalogSlot, catalogVeganOnly]);

  const executeToggleMeal = async (mealId: string, currentEaten: boolean) => {
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

  const handleToggleMeal = (meal: MealItem) => {
    // If not eaten yet, user is vegan, and the dish is NOT vegan -> show big red exclamation confirmation!
    if (!meal.eaten && user?.isVegan && !meal.isVegan) {
      setVeganWarningModal({
        mealName: meal.name,
        onConfirm: () => executeToggleMeal(meal.id, meal.eaten),
      });
    } else {
      executeToggleMeal(meal.id, meal.eaten);
    }
  };

  const executeAddMeal = async (mealId: string, mealName: string) => {
    try {
      setAddingCatalogId(mealId);
      await api.addMealToDay(selectedDay, mealId);
      showToast(`«${mealName}» добавлено в рацион на сегодня!`);
      await fetchNutrition(selectedDay);
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка добавления блюда', 'error');
    } finally {
      setAddingCatalogId(null);
    }
  };

  const handleAddFromCatalog = (m: CatalogMeal) => {
    if (user?.isVegan && !m.isVegan) {
      // Big red exclamation confirmation!
      setVeganWarningModal({
        mealName: m.name,
        onConfirm: () => executeAddMeal(m.id, m.name),
      });
    } else {
      executeAddMeal(m.id, m.name);
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

  const mealSections = [
    { key: 'breakfast', title: 'Завтрак', icon: Coffee, desc: 'Энергетический старт дня' },
    { key: 'lunch', title: 'Обед', icon: Utensils, desc: 'Основной приём пищи и баланс нутриентов' },
    { key: 'snack', title: 'Полдник / Перекус', icon: Cookie, desc: 'Поддержание метаболизма и сытости' },
    { key: 'dinner', title: 'Ужин', icon: Moon, desc: 'Лёгкое восстановление и качественный белок' },
  ];

  return (
    <div className="space-y-6">
      {/* Unified Real-Time Calendar Bar */}
      <div className="rounded-2xl p-4 sm:p-5 border bg-[#121216] border-white/[0.08] shadow-sm space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-white/10 text-white border border-white/20">
              <CalendarIcon className="w-5 h-5 text-white" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-sm font-bold text-white tracking-wide">
                  {weekLabel}
                </span>
                {weekOffset === 0 && (
                  <span className="text-[10px] uppercase font-bold px-2 py-0.5 rounded-md bg-white/15 text-white border border-white/20">
                    Текущая неделя
                  </span>
                )}
              </div>
              <span className="text-xs text-zinc-400">
                Единый календарь питания и нутриентов
              </span>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            {/* Week navigation arrows */}
            <div className="flex items-center bg-[#17171d] rounded-xl border border-white/[0.08] p-1">
              <button
                onClick={() => setWeekOffset(prev => prev - 1)}
                className="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-white/10 transition-colors"
                title="Предыдущая неделя"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
              <button
                onClick={() => {
                  setWeekOffset(0);
                  setSelectedDay(todayIso);
                }}
                className="px-2.5 py-1 text-xs font-semibold text-zinc-300 hover:text-white transition-colors"
              >
                Сегодня
              </button>
              <button
                onClick={() => setWeekOffset(prev => prev + 1)}
                className="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-white/10 transition-colors"
                title="Следующая неделя"
              >
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>

            {/* Calendar modal trigger */}
            <button
              onClick={() => {
                setCalendarYear(activeDate.getFullYear());
                setCalendarMonth(activeDate.getMonth());
                setCalendarModalOpen(true);
              }}
              className="flex items-center gap-2 px-3 py-2 rounded-xl text-xs font-bold bg-[#17171d] hover:bg-[#22222c] text-white border border-white/[0.12] transition-all active:scale-95 shadow-sm"
            >
              <CalendarIcon className="w-4 h-4 text-white" />
              <span>Календарь</span>
            </button>

            {/* Food Catalog trigger */}
            <button
              onClick={() => {
                setCatalogModalOpen(true);
              }}
              className="flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-all active:scale-95 shadow-md shadow-white/5"
            >
              <Utensils className="w-4 h-4 text-zinc-950" />
              <span>Каталог блюд</span>
            </button>
          </div>
        </div>

        {/* 7-Day Real Date Selector Strip */}
        <div className="grid grid-cols-7 gap-1.5 sm:gap-2 pt-1">
          {[1, 2, 3, 4, 5, 6, 7].map(day => {
            const isSelected = selectedDay === day;
            const dayDate = getDateForDay(day, weekOffset);
            const isToday = isSameDay(dayDate, today);
            const isPast = dayDate.getTime() < todayMidnight.getTime() && !isToday;

            return (
              <button
                key={day}
                onClick={() => setSelectedDay(day)}
                className={`relative flex flex-col items-center justify-center py-2.5 px-1 rounded-xl text-xs font-bold transition-all duration-150 ${
                  isSelected
                    ? 'bg-white text-zinc-950 shadow-lg shadow-white/10 scale-[1.02] border border-white'
                    : isToday
                    ? 'bg-zinc-800/90 text-white border border-white/40 hover:bg-zinc-800'
                    : isPast
                    ? 'bg-[#141419]/80 text-zinc-500 hover:text-zinc-300 hover:bg-[#1a1a22] border border-white/[0.04]'
                    : 'bg-[#17171d] text-zinc-400 hover:text-white hover:bg-[#202028] border border-white/[0.06]'
                }`}
              >
                {/* Weekday & Indicator */}
                <div className="flex items-center gap-1 mb-0.5">
                  <span className={`text-[10px] uppercase font-semibold ${isSelected ? 'text-zinc-800' : isPast ? 'text-zinc-500' : 'text-zinc-400'}`}>
                    {weekdayShort[day - 1]}
                  </span>
                  {isToday && (
                    <span className={`w-1.5 h-1.5 rounded-full ${isSelected ? 'bg-zinc-950' : 'bg-white animate-pulse'}`} />
                  )}
                </div>

                {/* Real calendar date number */}
                <span className={`text-sm font-black ${isPast && !isSelected ? 'text-zinc-500' : ''}`}>
                  {dayDate.getDate()} {monthsShortRu[dayDate.getMonth()]}
                </span>

                <span className={`text-[10px] mt-1 font-medium truncate max-w-full px-1 ${
                  isSelected
                    ? 'text-zinc-900 font-bold'
                    : isToday
                    ? 'text-zinc-200'
                    : isPast
                    ? 'text-zinc-600'
                    : 'text-zinc-400'
                }`}>
                  Рацион
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {loading ? (
        <div className="py-20 text-center text-zinc-300">Загрузка рациона...</div>
      ) : dayMeals ? (
        <div className="space-y-6">
          {/* Top Banner: Targets & Macros */}
          <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-white/[0.08] pb-4">
              <div>
                <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                  {dayMeals.weekdayName} • {activeDate.getDate()} {monthsRu[activeDate.getMonth()]} {activeDate.getFullYear()}
                </span>
                <h2 className="text-2xl font-black text-white mt-1">
                  Рацион на день
                </h2>
                {user?.isVegan && (
                  <span className="inline-flex items-center gap-1.5 text-xs font-bold px-2.5 py-0.5 rounded-full bg-white/10 text-white border border-white/20 mt-1">
                    <Leaf className="w-3.5 h-3.5 text-emerald-400" />
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
                  <div className="text-xs text-zinc-400 font-medium">Вода (1 стакан = 330 мл)</div>
                  <div className="text-sm font-bold text-white">
                    {waterIntake} / {waterTarget} мл ({Math.floor(waterIntake / 330)} из {Math.ceil(waterTarget / 330)} стак.)
                  </div>
                </div>
                <div className="flex items-center gap-1.5">
                  <button
                    onClick={() => handleAddWater(-330)}
                    disabled={addingWater || waterIntake <= 0}
                    className="flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-bold bg-[#1c1c24] hover:bg-[#252530] text-zinc-200 border border-white/[0.10] transition-all active:scale-95 disabled:opacity-40"
                    title="Убрать 1 стакан воды (-330 мл)"
                  >
                    <Minus className="w-3.5 h-3.5" />
                    <span>-330</span>
                  </button>
                  <button
                    onClick={() => handleAddWater(330)}
                    disabled={addingWater}
                    className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs font-bold bg-white/10 hover:bg-white/20 text-white border border-white/20 transition-all active:scale-95 disabled:opacity-50"
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
                <span className="text-xs font-semibold text-zinc-400 block mb-1">Калории</span>
                <div className="text-xl font-black text-white">
                  {dayMeals.currentCalories || 0} <span className="text-xs text-zinc-400 font-normal">/ {dayMeals.targetCalories} ккал</span>
                </div>
              </div>
              <div className="p-3.5 rounded-xl bg-[#17171d] border border-white/[0.06]">
                <span className="text-xs font-semibold text-zinc-300 block mb-1">Белки</span>
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
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2.5">
                      <div className="p-2 rounded-xl bg-[#17171d] text-white border border-white/[0.08]">
                        <Icon className="w-4 h-4" />
                      </div>
                      <div>
                        <h3 className="text-base font-bold text-white">{sec.title}</h3>
                        <p className="text-xs text-zinc-400">{sec.desc}</p>
                      </div>
                    </div>

                    <button
                      onClick={() => {
                        setCatalogSlot(sec.key);
                        setCatalogModalOpen(true);
                      }}
                      className="text-xs font-bold text-zinc-300 hover:text-white flex items-center gap-1 px-2.5 py-1 rounded-lg bg-[#17171d] hover:bg-[#202028] border border-white/[0.08] transition-colors"
                    >
                      <Plus className="w-3.5 h-3.5" />
                      <span>Выбрать блюдо</span>
                    </button>
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
                              <h4 className={`text-sm font-bold text-white ${m.eaten ? 'line-through text-zinc-500' : ''}`}>
                                {m.name}
                              </h4>
                              {m.isVegan ? (
                                <span className="p-0.5 rounded bg-emerald-500/20 text-emerald-400 border border-emerald-500/30" title="Веганское блюдо">
                                  <Leaf className="w-3 h-3 text-emerald-400" />
                                </span>
                              ) : user?.isVegan ? (
                                <span className="p-0.5 rounded bg-rose-500/20 text-rose-300 border border-rose-500/30" title="Не веганское блюдо!">
                                  <AlertTriangle className="w-3 h-3 text-rose-400" />
                                </span>
                              ) : null}
                            </div>
                            <p className="text-xs text-zinc-300 leading-relaxed font-normal">
                              {m.description}
                            </p>
                            <div className="flex items-center gap-2 pt-1 text-[11px] text-zinc-400 font-medium">
                              <span className="text-white font-bold">{m.calories} ккал</span>
                              <span>• Б: {m.proteins}г</span>
                              <span>• Ж: {m.fats}г</span>
                              <span>• У: {m.carbs}г</span>
                            </div>
                          </div>

                          <button
                            onClick={() => handleToggleMeal(m)}
                            disabled={togglingId === m.id}
                            className={`p-2 rounded-xl transition-all flex-shrink-0 ${
                              m.eaten
                                ? 'bg-white/15 text-white border border-white/30'
                                : 'bg-[#17171d] text-zinc-400 hover:text-white hover:bg-[#22222c] border border-white/[0.08]'
                            }`}
                            title={m.eaten ? 'Отменить отметку' : 'Отметить съеденным'}
                          >
                            {m.eaten ? (
                              <CheckCircle2 className="w-5 h-5 text-white" />
                            ) : (
                              <Circle className="w-5 h-5" />
                            )}
                          </button>
                        </div>
                      ))
                    ) : (
                      <div className="col-span-2 p-4 rounded-xl bg-[#121216]/60 border border-white/[0.06] text-xs text-zinc-400 text-center flex items-center justify-center gap-2">
                        <span>Блюда еще не добавлены.</span>
                        <button
                          onClick={() => {
                            setCatalogSlot(sec.key);
                            setCatalogModalOpen(true);
                          }}
                          className="text-white underline font-bold"
                        >
                          Выбрать из каталога
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      ) : (
        <div className="py-20 text-center text-zinc-300">Данные рациона недоступны</div>
      )}

      {/* BIG RED EXCLAMATION MARK VEGAN CONFIRMATION MODAL */}
      {veganWarningModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md animate-in fade-in">
          <div className="rounded-3xl p-6 sm:p-8 bg-[#161214] border-2 border-rose-500/70 shadow-2xl max-w-md w-full space-y-5 text-center">
            {/* Big Red Exclamation Sign */}
            <div className="w-20 h-20 rounded-full bg-rose-500/20 border-2 border-rose-500 text-rose-500 flex items-center justify-center mx-auto animate-bounce">
              <AlertTriangle className="w-12 h-12 stroke-[2.5]" />
            </div>

            <div className="space-y-2">
              <span className="text-xs font-black uppercase tracking-widest text-rose-400 block">
                Внимание: Блюдо не является веганским!
              </span>
              <h3 className="text-xl font-black text-white">
                «{veganWarningModal.mealName}»
              </h3>
              <p className="text-xs sm:text-sm text-zinc-300 leading-relaxed font-normal">
                В вашем профиле отмечен <span className="text-emerald-400 font-bold">100% веганский рацион</span>. Данное блюдо содержит продукты животного происхождения (мясо, рыбу, птицу, яйца или лактозу).
              </p>
            </div>

            <div className="p-3.5 rounded-xl bg-[#1d1416] border border-rose-500/30 text-xs text-rose-200">
              Вы уверены, что действительно съели это блюдо и хотите отметить его в рационе?
            </div>

            <div className="flex flex-col sm:flex-row items-center gap-2.5 pt-2">
              <button
                type="button"
                onClick={() => setVeganWarningModal(null)}
                className="w-full sm:w-1/2 py-2.5 rounded-xl text-xs font-bold text-zinc-400 hover:text-white bg-[#1a1a22] border border-white/[0.08] transition-colors"
              >
                Отмена
              </button>
              <button
                type="button"
                onClick={() => {
                  const action = veganWarningModal.onConfirm;
                  setVeganWarningModal(null);
                  action();
                }}
                className="w-full sm:w-1/2 py-2.5 rounded-xl text-xs font-bold bg-rose-600 hover:bg-rose-500 text-white shadow-lg shadow-rose-600/30 transition-all active:scale-95"
              >
                Подтверждаю, съел
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Full Food Catalog Modal */}
      {catalogModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in">
          <div className="rounded-3xl p-6 bg-[#121216] border border-white/20 shadow-2xl max-w-2xl w-full max-h-[85vh] flex flex-col space-y-4">
            <div className="flex items-start justify-between">
              <div>
                <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                  База блюд PulseFit
                </span>
                <h3 className="text-xl font-black text-white mt-0.5">
                  Каталог сбалансированного питания
                </h3>
              </div>
              <button
                onClick={() => setCatalogModalOpen(false)}
                className="p-1 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Filter row inside modal */}
            <div className="space-y-2.5">
              <div className="relative">
                <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" />
                <input
                  type="text"
                  value={catalogSearch}
                  onChange={e => setCatalogSearch(e.target.value)}
                  placeholder="Поиск по названию или ингредиентам..."
                  className="w-full pl-9 pr-4 py-2 rounded-xl bg-[#17171d] border border-white/[0.10] text-white placeholder-zinc-500 text-xs focus:border-white focus:outline-none"
                />
              </div>

              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex items-center gap-1 overflow-x-auto pb-0.5">
                  {[
                    { id: '', label: 'Все' },
                    { id: 'breakfast', label: 'Завтрак' },
                    { id: 'lunch', label: 'Обед' },
                    { id: 'snack', label: 'Перекус' },
                    { id: 'dinner', label: 'Ужин' },
                  ].map(s => (
                    <button
                      key={s.id}
                      onClick={() => setCatalogSlot(s.id)}
                      className={`px-3 py-1 rounded-lg text-xs font-bold whitespace-nowrap transition-all ${
                        catalogSlot === s.id
                          ? 'bg-white text-zinc-950 font-black'
                          : 'bg-[#17171d] text-zinc-300 hover:text-white border border-white/[0.08]'
                      }`}
                    >
                      {s.label}
                    </button>
                  ))}
                </div>

                <button
                  onClick={() => setCatalogVeganOnly(!catalogVeganOnly)}
                  className={`flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-bold transition-all border ${
                    catalogVeganOnly
                      ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40'
                      : 'bg-[#17171d] text-zinc-400 hover:text-white border-white/[0.08]'
                  }`}
                >
                  <Leaf className="w-3.5 h-3.5 text-emerald-400" />
                  <span>Веганские</span>
                </button>
              </div>
            </div>

            {/* Dishes list */}
            <div className="flex-1 overflow-y-auto space-y-2.5 pr-1">
              {loadingCatalog ? (
                <div className="py-12 text-center text-zinc-400 text-xs">Загрузка каталога...</div>
              ) : catalogMeals.length > 0 ? (
                catalogMeals.map(m => (
                  <div
                    key={m.id}
                    className="p-3.5 rounded-2xl bg-[#17171d] border border-white/[0.08] hover:border-white/[0.18] flex items-center justify-between gap-3 transition-colors"
                  >
                    <div className="space-y-0.5">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-bold text-white">{m.name}</span>
                        {m.isVegan ? (
                          <span className="flex items-center gap-1 text-[10px] font-bold px-1.5 py-0.5 rounded bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                            <Leaf className="w-2.5 h-2.5" />
                            Веган
                          </span>
                        ) : user?.isVegan ? (
                          <span className="flex items-center gap-1 text-[10px] font-bold px-1.5 py-0.5 rounded bg-rose-500/20 text-rose-300 border border-rose-500/30">
                            <AlertTriangle className="w-2.5 h-2.5 text-rose-400" />
                            Не веган
                          </span>
                        ) : null}
                      </div>
                      <div className="text-xs text-zinc-400">
                        {m.calories} ккал • Б: {m.proteins}г • Ж: {m.fats}г • У: {m.carbs}г
                      </div>
                      <div className="text-[11px] text-zinc-500 truncate max-w-md">
                        {m.portion}
                      </div>
                    </div>

                    <button
                      onClick={() => handleAddFromCatalog(m)}
                      disabled={addingCatalogId === m.id}
                      className="flex items-center gap-1 px-3 py-1.5 rounded-xl text-xs font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-all flex-shrink-0 active:scale-95 disabled:opacity-50"
                    >
                      <Plus className="w-3.5 h-3.5" />
                      <span>{addingCatalogId === m.id ? 'Добавление...' : 'Добавить'}</span>
                    </button>
                  </div>
                ))
              ) : (
                <div className="py-12 text-center text-zinc-400 text-xs">Блюд не найдено</div>
              )}
            </div>

            <div className="pt-2 border-t border-white/[0.08] flex justify-end">
              <button
                onClick={() => setCatalogModalOpen(false)}
                className="px-4 py-2 rounded-xl text-xs font-bold bg-[#17171d] hover:bg-[#202028] text-white border border-white/[0.12]"
              >
                Закрыть
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Calendar Modal */}
      {calendarModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in">
          <div className="rounded-3xl p-6 bg-[#121216] border border-white/20 shadow-2xl max-w-md w-full space-y-5">
            {/* Modal Header */}
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-black text-white">
                  {monthsRuNominative[calendarMonth]} {calendarYear}
                </h3>
                <p className="text-xs text-zinc-400">
                  Выберите дату для перехода к рациону
                </p>
              </div>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => {
                    if (calendarMonth === 0) {
                      setCalendarMonth(11);
                      setCalendarYear(y => y - 1);
                    } else {
                      setCalendarMonth(m => m - 1);
                    }
                  }}
                  className="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800"
                >
                  <ChevronLeft className="w-5 h-5" />
                </button>
                <button
                  onClick={() => {
                    if (calendarMonth === 11) {
                      setCalendarMonth(0);
                      setCalendarYear(y => y + 1);
                    } else {
                      setCalendarMonth(m => m + 1);
                    }
                  }}
                  className="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800"
                >
                  <ChevronRight className="w-5 h-5" />
                </button>
                <button
                  onClick={() => setCalendarModalOpen(false)}
                  className="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800 ml-1"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>
            </div>

            {/* Weekday headers */}
            <div className="grid grid-cols-7 gap-1 text-center text-xs font-bold text-zinc-400">
              {weekdayShort.map(w => (
                <div key={w} className="py-1">
                  {w}
                </div>
              ))}
            </div>

            {/* Days grid */}
            <div className="grid grid-cols-7 gap-1 text-center">
              {Array.from({ length: monthStartOffset }).map((_, i) => (
                <div key={`empty-${i}`} className="h-11" />
              ))}

              {Array.from({ length: daysInMonth }).map((_, i) => {
                const dayNum = i + 1;
                const dateObj = new Date(calendarYear, calendarMonth, dayNum);
                dateObj.setHours(0, 0, 0, 0);
                const isTodayCell = isSameDay(dateObj, today);
                const isPast = dateObj.getTime() < todayMidnight.getTime() && !isTodayCell;

                return (
                  <button
                    key={dayNum}
                    onClick={() => handleSelectCalendarDate(dayNum)}
                    className={`h-11 rounded-2xl flex flex-col items-center justify-center text-xs transition-all ${
                      isTodayCell
                        ? 'bg-white text-zinc-950 font-black shadow-md'
                        : isPast
                        ? 'text-zinc-600 hover:text-zinc-300 hover:bg-white/5 opacity-60'
                        : 'hover:bg-white/10 text-zinc-200'
                    }`}
                  >
                    <span className="font-bold leading-none">{dayNum}</span>
                    <span className={`w-1 h-1 rounded-full mt-1 ${
                      isTodayCell ? 'bg-zinc-950' : 'bg-white/60'
                    }`} />
                  </button>
                );
              })}
            </div>

            {/* Footer */}
            <div className="pt-2 border-t border-white/[0.08] flex items-center justify-end text-xs text-zinc-400">
              <button
                onClick={() => {
                  setWeekOffset(0);
                  setSelectedDay(todayIso);
                  setCalendarModalOpen(false);
                }}
                className="text-white hover:underline font-bold"
              >
                К сегодня
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
