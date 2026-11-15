import React, { useState, useEffect, useMemo } from 'react';
import {
  Dumbbell,
  RotateCw,
  CheckCircle2,
  Info,
  ShieldAlert,
  Calendar as CalendarIcon,
  ChevronLeft,
  ChevronRight,
  ArrowRightLeft,
  FileText,
  Save,
  Sparkles,
  X,
  AlertTriangle,
} from 'lucide-react';
import { WorkoutDay, PlanExercise, UserProfile } from '../types';
import { api } from '../api/client';
import { useToast } from '../context/ToastContext';

interface WorkoutsPageProps {
  initialDay?: number;
  onRefresh?: () => void;
}

export const WorkoutsPage: React.FC<WorkoutsPageProps> = ({ initialDay, onRefresh }) => {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);
  const today = new Date();
  const todayIso = today.getDay() === 0 ? 7 : today.getDay();
  const [selectedDay, setSelectedDay] = useState(initialDay || todayIso);
  const [workout, setWorkout] = useState<WorkoutDay | null>(null);
  const [allDays, setAllDays] = useState<WorkoutDay[]>([]);
  const [completedDates, setCompletedDates] = useState<string[]>([]);
  const [user, setUser] = useState<UserProfile | null>(null);
  const [swappingId, setSwappingId] = useState<string | null>(null);
  const [completing, setCompleting] = useState(false);
  const [detailExercise, setDetailExercise] = useState<PlanExercise | null>(null);

  // Calendar state
  const [weekOffset, setWeekOffset] = useState(0);
  const [calendarModalOpen, setCalendarModalOpen] = useState(false);
  const [calendarYear, setCalendarYear] = useState(() => new Date().getFullYear());
  const [calendarMonth, setCalendarMonth] = useState(() => new Date().getMonth());

  // Reschedule calendar modal
  const [rescheduleModalOpen, setRescheduleModalOpen] = useState(false);
  const [rescheduleYear, setRescheduleYear] = useState(() => new Date().getFullYear());
  const [rescheduleMonth, setRescheduleMonth] = useState(() => new Date().getMonth());
  const [rescheduleTargetDate, setRescheduleTargetDate] = useState<{
    dateStr: string;
    dayNum: number;
    month: number;
    year: number;
    isoDay: number;
  } | null>(null);
  const [rescheduling, setRescheduling] = useState(false);

  // Dedicated notes state
  const [noteText, setNoteText] = useState('');
  const [savingNote, setSavingNote] = useState(false);

  const fetchWorkout = async (day: number) => {
    try {
      setLoading(true);
      const res = await api.getWorkouts(day);
      setWorkout(res.workout);
      setAllDays(res.allDays || []);
      setCompletedDates(res.completedDates || []);
      setUser(res.user);
      setSelectedDay(res.selectedDay);
      setNoteText(res.workout?.notes || '');
    } catch (e: any) {
      showToast(e.message || 'Ошибка загрузки тренировок', 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWorkout(selectedDay);
  }, [selectedDay]);

  // Calendar date math
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

  const formatDateKey = (d: Date): string => {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  };

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

  // Active workout date calculations
  const activeDate = getDateForDay(selectedDay, weekOffset);
  const activeDateKey = formatDateKey(activeDate);
  const isWorkoutCompletedOnActiveDate = completedDates.includes(activeDateKey);

  // Week range label
  const mondayDate = getDateForDay(1, weekOffset);
  const sundayDate = getDateForDay(7, weekOffset);
  const weekLabel = `${mondayDate.getDate()} ${monthsShortRu[mondayDate.getMonth()]} – ${sundayDate.getDate()} ${monthsShortRu[sundayDate.getMonth()]} ${sundayDate.getFullYear()}`;

  // Big Calendar calculations
  const firstDayOfMonth = new Date(calendarYear, calendarMonth, 1);
  const daysInMonth = new Date(calendarYear, calendarMonth + 1, 0).getDate();
  const monthStartOffset = (firstDayOfMonth.getDay() + 6) % 7;

  // Reschedule Calendar calculations
  const rescheduleFirstDay = new Date(rescheduleYear, rescheduleMonth, 1);
  const rescheduleDaysCount = new Date(rescheduleYear, rescheduleMonth + 1, 0).getDate();
  const rescheduleStartOffset = (rescheduleFirstDay.getDay() + 6) % 7;

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

  const handleSwap = async (exerciseId: string) => {
    try {
      setSwappingId(exerciseId);
      const res = await api.swapExercise(selectedDay, exerciseId);
      showToast(`Упражнение заменено на «${res.newExercise.title}»`);
      if (workout) {
        setWorkout({
          ...workout,
          exercises: workout.exercises.map(ex => (ex.id === exerciseId ? res.newExercise : ex)),
        });
      }
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка замены упражнения', 'error');
    } finally {
      setSwappingId(null);
    }
  };

  const handleToggleComplete = async () => {
    const targetStatus = !isWorkoutCompletedOnActiveDate;
    try {
      setCompleting(true);
      await api.completeWorkout(selectedDay, activeDateKey, targetStatus);
      if (targetStatus) {
        setCompletedDates(prev => (prev.includes(activeDateKey) ? prev : [...prev, activeDateKey]));
        showToast(`Тренировка за ${activeDate.getDate()} ${monthsRu[activeDate.getMonth()]} выполнена! 🔥`);
      } else {
        setCompletedDates(prev => prev.filter(d => d !== activeDateKey));
        showToast('Отметка о выполнении снята');
      }
      if (workout) {
        setWorkout({ ...workout, completed: targetStatus });
      }
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка изменения статуса тренировки', 'error');
    } finally {
      setCompleting(false);
    }
  };

  const handleExecuteReschedule = async () => {
    if (!rescheduleTargetDate) {
      showToast('Выберите свободную дату в календаре для переноса', 'error');
      return;
    }
    if (rescheduleTargetDate.isoDay === selectedDay) {
      showToast('Выберите день, отличный от текущей тренировки', 'error');
      return;
    }
    try {
      setRescheduling(true);
      await api.rescheduleWorkout(selectedDay, rescheduleTargetDate.isoDay);
      showToast(`Тренировка перенесена на ${rescheduleTargetDate.dayNum} ${monthsRu[rescheduleTargetDate.month]}!`);
      setRescheduleModalOpen(false);

      // Compute week offset for the target date to focus on it
      const targetDateObj = new Date(rescheduleTargetDate.year, rescheduleTargetDate.month, rescheduleTargetDate.dayNum);
      const targetIso = rescheduleTargetDate.isoDay;
      const targetMonday = new Date(targetDateObj);
      targetMonday.setDate(targetDateObj.getDate() - (targetIso - 1));
      targetMonday.setHours(0, 0, 0, 0);
      const baseMonday = getMondayOfWeek(0);
      const diffWeeks = Math.round((targetMonday.getTime() - baseMonday.getTime()) / (7 * 24 * 3600 * 1000));

      setWeekOffset(diffWeeks);
      setSelectedDay(targetIso);
      await fetchWorkout(targetIso);
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка переноса тренировки', 'error');
    } finally {
      setRescheduling(false);
    }
  };

  const handleSaveNote = async () => {
    try {
      setSavingNote(true);
      await api.saveWorkoutNote(selectedDay, noteText);
      showToast('Заметка к тренировке сохранена!');
      if (workout) {
        setWorkout({ ...workout, notes: noteText });
      }
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка сохранения заметки', 'error');
    } finally {
      setSavingNote(false);
    }
  };

  // Sort exercises for gentle mode:
  // Recommended exercises first, contraindicated ones with clear warnings placed at the end
  const sortedExercises = useMemo(() => {
    if (!workout?.exercises) return [];
    if (!user?.hasHealthIssues) return workout.exercises;

    const recommended = workout.exercises.filter(ex => !ex.isContraindicated);
    const contraindicated = workout.exercises.filter(ex => ex.isContraindicated);
    return [...recommended, ...contraindicated];
  }, [workout?.exercises, user?.hasHealthIssues]);

  return (
    <div className="space-y-6">
      {/* Top Calendar Bar: Week navigation, dates, Calendar modal trigger */}
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
                Календарь тренировок в реальном времени
              </span>
            </div>
          </div>

          <div className="flex items-center gap-2">
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
                  const todayIso = today.getDay() === 0 ? 7 : today.getDay();
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
              className="flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-bold bg-white/10 hover:bg-white/20 text-white border border-white/20 transition-all active:scale-95 shadow-sm"
            >
              <CalendarIcon className="w-4 h-4 text-white" />
              <span>Календарь</span>
            </button>
          </div>
        </div>

        {/* 7-Day Real Date Selector Strip */}
        <div className="grid grid-cols-7 gap-1.5 sm:gap-2 pt-1">
          {[1, 2, 3, 4, 5, 6, 7].map(day => {
            const isSelected = selectedDay === day;
            const dayData = allDays.find(d => d.dayIndex === day);
            const isRest = dayData?.isRest;
            const dayDate = getDateForDay(day, weekOffset);
            const dayDateKey = formatDateKey(dayDate);
            const isDone = completedDates.includes(dayDateKey);
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

                {/* Completion or Status tag */}
                <span className={`text-[10px] mt-1 font-medium truncate max-w-full px-1 flex items-center gap-0.5 ${
                  isSelected
                    ? 'text-zinc-900 font-bold'
                    : isDone
                    ? 'text-emerald-400 font-bold'
                    : isRest
                    ? 'text-zinc-500'
                    : isPast
                    ? 'text-zinc-600'
                    : 'text-zinc-300'
                }`}>
                  {isDone ? '✓ Готово' : isRest ? 'Отдых' : 'Тренинг'}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {loading ? (
        <div className="py-20 text-center text-zinc-400">Загрузка тренировки...</div>
      ) : workout ? (
        <div className="space-y-6">
          {/* Header Card */}
          <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                  {workout.weekdayName} • {activeDate.getDate()} {monthsRu[activeDate.getMonth()]} {activeDate.getFullYear()}
                </span>
                {isWorkoutCompletedOnActiveDate && (
                  <span className="flex items-center gap-1 text-[11px] font-bold px-2.5 py-0.5 rounded-md bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                    <CheckCircle2 className="w-3 h-3 text-emerald-400" />
                    Выполнено
                  </span>
                )}
              </div>
              <h2 className="text-2xl font-black text-white mt-1">
                {workout.isRest ? 'День восстановления и отдыха' : workout.dayName}
              </h2>
              <p className="text-xs sm:text-sm text-zinc-400 mt-0.5">
                Фокус: <span className="text-zinc-200 font-medium">{workout.focus}</span> • Примерное время: ~{workout.estimatedMinutes} мин
              </p>
            </div>

            <div className="flex items-center gap-2.5">
              {/* Reschedule Button */}
              <button
                onClick={() => {
                  setRescheduleYear(activeDate.getFullYear());
                  setRescheduleMonth(activeDate.getMonth());
                  setRescheduleTargetDate(null);
                  setRescheduleModalOpen(true);
                }}
                className="flex items-center gap-1.5 px-4 py-2.5 rounded-xl text-xs font-bold bg-[#17171d] hover:bg-[#22222c] text-zinc-200 hover:text-white border border-white/[0.12] transition-all active:scale-95"
                title="Перенести тренировку на другую дату"
              >
                <ArrowRightLeft className="w-4 h-4 text-zinc-300" />
                <span>Перенести</span>
              </button>

              {!workout.isRest && (
                <button
                  onClick={handleToggleComplete}
                  disabled={completing}
                  className={`flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold transition-all shadow-md ${
                    isWorkoutCompletedOnActiveDate
                      ? 'bg-zinc-800 hover:bg-zinc-700 text-zinc-200 border border-white/20 active:scale-95'
                      : 'bg-white hover:bg-zinc-200 text-zinc-950 shadow-white/10 active:scale-95'
                  }`}
                >
                  <CheckCircle2 className={`w-4 h-4 ${isWorkoutCompletedOnActiveDate ? 'text-emerald-400' : 'text-zinc-950'}`} />
                  <span>
                    {isWorkoutCompletedOnActiveDate
                      ? '✓ Выполнено • Отменить'
                      : 'Отметить как выполненную'}
                  </span>
                </button>
              )}
            </div>
          </div>

          {/* Gentle mode alert if active */}
          {user?.hasHealthIssues && (
            <div className="rounded-xl p-4 border bg-[#121216] border-amber-500/40 text-zinc-200 text-xs flex items-start gap-3">
              <ShieldAlert className="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
              <div className="space-y-1">
                <span className="font-bold text-white block">
                  Включен щадящий режим тренировок
                </span>
                <p className="text-zinc-300 leading-relaxed font-normal">
                  Рекомендованные безопасные упражнения отображаются первыми с зеленой отметкой. Упражнения с осевой нагрузкой помечены знаком предупреждения — тренер настоятельно рекомендует заменить их на щадящие аналоги.
                </p>
              </div>
            </div>
          )}

          {/* Exercise list or Rest Day screen */}
          {workout.isRest ? (
            <div className="rounded-2xl p-12 border bg-[#121216] border-white/[0.08] text-center max-w-xl mx-auto space-y-4">
              <div className="w-14 h-14 rounded-2xl bg-white/10 border border-white/20 flex items-center justify-center mx-auto text-white">
                <Sparkles className="w-7 h-7" />
              </div>
              <h3 className="text-xl font-bold text-white">Время восстановить силы</h3>
              <p className="text-sm text-zinc-300 leading-relaxed">
                Мышцы растут и адаптируются именно во время отдыха. Проведите легкую пешую прогулку, пейте достаточно воды и соблюдайте режим сна.
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-base font-bold text-white flex items-center gap-2">
                  <Dumbbell className="w-4 h-4 text-white" />
                  Упражнения программы ({sortedExercises.length})
                </h3>
                <span className="text-xs text-zinc-400">
                  Если упражнение не подходит — нажмите «Заменить»
                </span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {sortedExercises.map((ex, idx) => {
                  const isContra = ex.isContraindicated && user?.hasHealthIssues;

                  return (
                    <div
                      key={ex.id || idx}
                      className={`rounded-2xl p-5 border shadow-sm transition-all flex flex-col justify-between gap-4 group ${
                        isContra
                          ? 'bg-[#161214] border-rose-500/40 hover:border-rose-500/70'
                          : 'bg-[#121216] border-white/[0.08] hover:border-white/[0.20]'
                      }`}
                    >
                      <div>
                        {/* Header badge & status */}
                        <div className="flex items-start justify-between gap-2 mb-2">
                          <div>
                            <div className="flex items-center gap-2 mb-1">
                              <span className="text-[11px] font-bold uppercase tracking-wider text-zinc-400 block">
                                Упражнение {idx + 1}
                              </span>
                              {user?.hasHealthIssues && (
                                isContra ? (
                                  <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-rose-500/20 text-rose-300 border border-rose-500/40">
                                    ⚠️ Нельзя выполнять при ограничениях
                                  </span>
                                ) : (
                                  <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                                    ✓ Можно выполнять
                                  </span>
                                )
                              )}
                            </div>
                            <h4 className={`text-base font-bold transition-colors ${isContra ? 'text-rose-100' : 'text-white group-hover:text-zinc-200'}`}>
                              {ex.title}
                            </h4>
                          </div>
                          <span className={`text-xs font-bold px-2.5 py-1 rounded-lg border flex-shrink-0 ${
                            isContra
                              ? 'bg-rose-500/15 text-rose-200 border-rose-500/30'
                              : 'bg-white/10 text-white border-white/20'
                          }`}>
                            {ex.sets} × {ex.reps}
                          </span>
                        </div>

                        {/* Muscle & equipment pills */}
                        <div className="flex flex-wrap items-center gap-1.5 mt-2">
                          <span className="text-[11px] px-2 py-0.5 rounded-md bg-[#1c1c24] text-zinc-300 font-medium border border-white/[0.06]">
                            {ex.muscle}
                          </span>
                          <span className="text-[11px] px-2 py-0.5 rounded-md bg-[#1c1c24] text-zinc-300 font-medium border border-white/[0.06]">
                            {ex.equipment}
                          </span>
                          <span className="text-[11px] px-2 py-0.5 rounded-md bg-[#1c1c24] text-zinc-400 font-medium border border-white/[0.06]">
                            Отдых {ex.restSec || 60}с
                          </span>
                        </div>

                        {/* Contraindication Warning Banner */}
                        {isContra && (
                          <div className="mt-3 p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-200 text-xs flex items-start gap-2.5">
                            <AlertTriangle className="w-4 h-4 text-rose-400 flex-shrink-0 mt-0.5" />
                            <div className="space-y-0.5">
                              <span className="font-bold text-rose-300 block">
                                Не рекомендуется тренером!
                              </span>
                              <p className="text-[11px] text-rose-200/90 leading-relaxed font-normal">
                                Создает осевую или ударную нагрузку на позвоночник/суставы. Нажмите «Заменить», чтобы подобрать безопасную альтернативу.
                              </p>
                            </div>
                          </div>
                        )}

                        {ex.coachTip && (
                          <p className="text-xs text-zinc-200 mt-3 line-clamp-2 leading-relaxed bg-[#17171d] p-2.5 rounded-xl border border-white/[0.08]">
                            💡 <span className="text-white font-medium">Совет тренера:</span> {ex.coachTip}
                          </p>
                        )}
                      </div>

                      <div className="pt-2 border-t border-white/[0.08] flex items-center justify-between">
                        <button
                          onClick={() => setDetailExercise(ex)}
                          className="text-xs font-bold text-zinc-300 hover:text-white flex items-center gap-1 transition-colors"
                        >
                          <Info className="w-3.5 h-3.5 text-zinc-400" />
                          Техника
                        </button>

                        <button
                          onClick={() => handleSwap(ex.id)}
                          disabled={swappingId === ex.id}
                          className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold transition-all active:scale-95 disabled:opacity-50 ${
                            isContra
                              ? 'bg-rose-500/20 hover:bg-rose-500/30 text-rose-200 border border-rose-500/40'
                              : 'bg-[#17171d] hover:bg-[#252530] text-zinc-200 hover:text-white border border-white/[0.12]'
                          }`}
                        >
                          <RotateCw className={`w-3.5 h-3.5 ${swappingId === ex.id ? 'animate-spin text-white' : ''}`} />
                          <span>
                            {swappingId === ex.id
                              ? 'Подбор...'
                              : isContra
                              ? 'Заменить на безопасное'
                              : 'Заменить'}
                          </span>
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* DEDICATED WORKOUT NOTES CARD */}
          <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-4">
            <div className="flex items-center justify-between border-b border-white/[0.08] pb-3">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-xl bg-white/10 text-white border border-white/20">
                  <FileText className="w-4 h-4 text-white" />
                </div>
                <div>
                  <h3 className="text-base font-bold text-white">
                    Заметки к тренировке ({workout.dayName})
                  </h3>
                  <p className="text-xs text-zinc-400">
                    Запишите рабочие веса, самочувствие, пульс, количество повторений или ощущения
                  </p>
                </div>
              </div>

              {workout.notes && (
                <span className="text-[11px] font-semibold text-zinc-400">
                  Сохранено
                </span>
              )}
            </div>

            <div className="space-y-3">
              <textarea
                value={noteText}
                onChange={e => setNoteText(e.target.value)}
                rows={3}
                placeholder="Например: Жим лежа 80 кг × 8 повторений. Пульс в норме, чувствуется прогресс в плечевом поясе..."
                className="w-full px-4 py-3 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none placeholder-zinc-500 transition-colors"
              />

              <div className="flex items-center justify-between">
                <span className="text-xs text-zinc-500">
                  Заметки синхронизируются с историей тренировок
                </span>
                <button
                  onClick={handleSaveNote}
                  disabled={savingNote}
                  className="flex items-center gap-2 px-5 py-2 rounded-xl text-xs font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-all shadow-sm active:scale-95 disabled:opacity-50"
                >
                  <Save className="w-3.5 h-3.5" />
                  <span>{savingNote ? 'Сохранение...' : 'Сохранить заметку'}</span>
                </button>
              </div>
            </div>
          </div>

          {/* Exercise Detail Modal */}
          {detailExercise && (
            <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in">
              <div className="rounded-2xl p-6 bg-[#121216] border border-white/20 shadow-2xl max-w-lg w-full space-y-4">
                <div className="flex items-start justify-between">
                  <div>
                    <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                      Техника выполнения
                    </span>
                    <h3 className="text-xl font-bold text-white mt-1">{detailExercise.title}</h3>
                  </div>
                  <button
                    onClick={() => setDetailExercise(null)}
                    className="p-1 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800"
                  >
                    <X className="w-5 h-5" />
                  </button>
                </div>

                <div className="flex gap-2">
                  <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-[#1c1c24] text-zinc-200 border border-white/[0.08]">
                    Мышцы: {detailExercise.muscle}
                  </span>
                  <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-[#1c1c24] text-zinc-200 border border-white/[0.08]">
                    Инвентарь: {detailExercise.equipment}
                  </span>
                </div>

                <div className="p-4 rounded-xl bg-[#17171d] border border-white/[0.08] text-xs sm:text-sm text-zinc-300 leading-relaxed space-y-2">
                  <div className="font-bold text-white flex items-center gap-1.5">
                    <Sparkles className="w-4 h-4 text-white" />
                    Рекомендации по выполнению:
                  </div>
                  <p>{detailExercise.coachTip || 'Выполняйте упражнение в подконтрольном темпе, сохраняя ровное дыхание.'}</p>
                </div>

                <div className="flex items-center justify-between text-xs text-zinc-400 pt-2 border-t border-white/[0.08]">
                  <span>Рекомендуемый сет: {detailExercise.sets} подходов по {detailExercise.reps}</span>
                  <button
                    onClick={() => setDetailExercise(null)}
                    className="px-4 py-2 rounded-xl text-xs font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-colors"
                  >
                    Понятно
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Standard Calendar Modal */}
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
                      Выберите дату для перехода к тренировке
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
                  {/* Empty cells */}
                  {Array.from({ length: monthStartOffset }).map((_, i) => (
                    <div key={`empty-${i}`} className="h-11" />
                  ))}

                  {/* Month days */}
                  {Array.from({ length: daysInMonth }).map((_, i) => {
                    const dayNum = i + 1;
                    const dateObj = new Date(calendarYear, calendarMonth, dayNum);
                    dateObj.setHours(0, 0, 0, 0);
                    const isTodayCell = isSameDay(dateObj, today);
                    const cellKey = formatDateKey(dateObj);
                    const isDone = completedDates.includes(cellKey);
                    const isPast = dateObj.getTime() < todayMidnight.getTime() && !isTodayCell;
                    const isoDay = dateObj.getDay() === 0 ? 7 : dateObj.getDay();
                    const dayData = allDays.find(d => d.dayIndex === isoDay);
                    const isRestDay = dayData?.isRest;

                    return (
                      <button
                        key={dayNum}
                        onClick={() => handleSelectCalendarDate(dayNum)}
                        className={`h-11 rounded-2xl flex flex-col items-center justify-center text-xs transition-all relative ${
                          isTodayCell
                            ? 'bg-white text-zinc-950 font-black shadow-md'
                            : isDone
                            ? 'bg-emerald-500/15 text-emerald-300 border border-emerald-500/30 hover:bg-emerald-500/25'
                            : isPast
                            ? 'text-zinc-600 hover:text-zinc-300 hover:bg-white/5 opacity-60'
                            : 'hover:bg-white/10 text-zinc-200'
                        }`}
                      >
                        <span className="font-bold leading-none">{dayNum}</span>
                        {/* Dot indicator or checkmark */}
                        {isDone ? (
                          <span className="text-[10px] font-black text-emerald-400 leading-none mt-0.5">✓</span>
                        ) : (
                          <span className={`w-1 h-1 rounded-full mt-1 ${
                            isTodayCell
                              ? 'bg-zinc-950'
                              : isRestDay
                              ? 'bg-zinc-600'
                              : 'bg-white'
                          }`} />
                        )}
                      </button>
                    );
                  })}
                </div>

                {/* Footer notes */}
                <div className="pt-2 border-t border-white/[0.08] flex items-center justify-between text-xs text-zinc-400">
                  <div className="flex items-center gap-3">
                    <span className="flex items-center gap-1.5">
                      <span className="text-emerald-400 font-bold">✓</span> Выполнено
                    </span>
                    <span className="flex items-center gap-1.5">
                      <span className="w-2 h-2 rounded-full bg-white" /> Тренинг
                    </span>
                    <span className="flex items-center gap-1.5">
                      <span className="w-2 h-2 rounded-full bg-zinc-600" /> Отдых
                    </span>
                  </div>
                  <button
                    onClick={() => {
                      setWeekOffset(0);
                      const todayIso = today.getDay() === 0 ? 7 : today.getDay();
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

          {/* Reschedule Modal: Real Calendar Picker */}
          {rescheduleModalOpen && (
            <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in">
              <div className="rounded-3xl p-6 bg-[#121216] border border-white/20 shadow-2xl max-w-lg w-full space-y-4">
                <div className="flex items-start justify-between">
                  <div>
                    <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                      Перенос тренировки по календарю
                    </span>
                    <h3 className="text-lg font-black text-white mt-1">
                      Перенести «{workout.dayName}»
                    </h3>
                  </div>
                  <button
                    onClick={() => setRescheduleModalOpen(false)}
                    className="p-1 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800"
                  >
                    <X className="w-5 h-5" />
                  </button>
                </div>

                <p className="text-xs text-zinc-300 leading-relaxed">
                  Выберите свободную дату в календаре (день отдыха). Если на день уже назначена тренировка — перенести на него нельзя.
                </p>

                {/* Month navigation */}
                <div className="flex items-center justify-between bg-[#17171d] p-3 rounded-2xl border border-white/[0.08]">
                  <span className="text-sm font-bold text-white">
                    {monthsRuNominative[rescheduleMonth]} {rescheduleYear}
                  </span>
                  <div className="flex items-center gap-1">
                    <button
                      type="button"
                      onClick={() => {
                        if (rescheduleMonth === 0) {
                          setRescheduleMonth(11);
                          setRescheduleYear(y => y - 1);
                        } else {
                          setRescheduleMonth(m => m - 1);
                        }
                      }}
                      className="p-1 rounded-lg text-zinc-400 hover:text-white hover:bg-white/10"
                    >
                      <ChevronLeft className="w-4 h-4" />
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        if (rescheduleMonth === 11) {
                          setRescheduleMonth(0);
                          setRescheduleYear(y => y + 1);
                        } else {
                          setRescheduleMonth(m => m + 1);
                        }
                      }}
                      className="p-1 rounded-lg text-zinc-400 hover:text-white hover:bg-white/10"
                    >
                      <ChevronRight className="w-4 h-4" />
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

                {/* Days grid with occupied / free indicators */}
                <div className="grid grid-cols-7 gap-1 text-center max-h-60 overflow-y-auto p-1">
                  {/* Empty cells */}
                  {Array.from({ length: rescheduleStartOffset }).map((_, i) => (
                    <div key={`re-empty-${i}`} className="h-12" />
                  ))}

                  {/* Month days */}
                  {Array.from({ length: rescheduleDaysCount }).map((_, i) => {
                    const dayNum = i + 1;
                    const dateObj = new Date(rescheduleYear, rescheduleMonth, dayNum);
                    dateObj.setHours(0, 0, 0, 0);
                    const isoDay = dateObj.getDay() === 0 ? 7 : dateObj.getDay();
                    const dayData = allDays.find(d => d.dayIndex === isoDay);
                    const isOccupied = dayData && !dayData.isRest;
                    const isCurrent = isoDay === selectedDay;
                    const dKey = formatDateKey(dateObj);
                    const isSelected = rescheduleTargetDate?.dateStr === dKey;

                    // A date can be selected ONLY if it's not occupied and not current day
                    const canSelect = !isOccupied && !isCurrent;

                    return (
                      <button
                        key={dayNum}
                        type="button"
                        disabled={!canSelect}
                        onClick={() => {
                          if (canSelect) {
                            setRescheduleTargetDate({
                              dateStr: dKey,
                              dayNum,
                              month: rescheduleMonth,
                              year: rescheduleYear,
                              isoDay,
                            });
                          }
                        }}
                        className={`h-12 rounded-xl flex flex-col items-center justify-center p-1 text-xs transition-all relative ${
                          isSelected
                            ? 'bg-white text-zinc-950 font-black shadow-lg scale-105 border border-white'
                            : isCurrent
                            ? 'bg-zinc-800 text-zinc-400 opacity-60 cursor-not-allowed border border-white/[0.05]'
                            : isOccupied
                            ? 'bg-rose-950/20 text-zinc-500 border border-rose-500/20 cursor-not-allowed opacity-50'
                            : 'bg-[#17171d] text-zinc-200 hover:bg-emerald-500/20 hover:border-emerald-500/40 border border-white/[0.06] cursor-pointer'
                        }`}
                        title={
                          isCurrent
                            ? 'Текущая тренировка'
                            : isOccupied
                            ? `Занято: ${dayData?.dayName}`
                            : 'Свободный день отдыха (можно перенести сюда)'
                        }
                      >
                        <span className="font-bold leading-none">{dayNum}</span>
                        <span className={`text-[9px] mt-0.5 truncate max-w-full font-semibold ${
                          isSelected
                            ? 'text-zinc-900'
                            : isCurrent
                            ? 'text-zinc-400'
                            : isOccupied
                            ? 'text-rose-400'
                            : 'text-emerald-400'
                        }`}>
                          {isCurrent ? 'Текущая' : isOccupied ? 'Занято' : 'Свободно'}
                        </span>
                      </button>
                    );
                  })}
                </div>

                {/* Selected target info banner */}
                {rescheduleTargetDate ? (
                  <div className="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-xs flex items-center justify-between">
                    <div>
                      <span className="font-bold text-white block">
                        Выбрана дата переноса:
                      </span>
                      <span>
                        {rescheduleTargetDate.dayNum} {monthsRu[rescheduleTargetDate.month]} {rescheduleTargetDate.year} ({weekdayShort[rescheduleTargetDate.isoDay - 1]})
                      </span>
                    </div>
                    <span className="px-2.5 py-1 rounded-lg bg-emerald-500/20 text-emerald-300 text-[11px] font-bold">
                      День свободен
                    </span>
                  </div>
                ) : (
                  <div className="p-3 rounded-xl bg-[#17171d] border border-white/[0.06] text-zinc-400 text-xs">
                    Нажмите на любую зеленую дату («Свободно»), чтобы запланировать тренировку на этот день.
                  </div>
                )}

                <div className="flex items-center justify-end gap-2.5 pt-3 border-t border-white/[0.08]">
                  <button
                    type="button"
                    onClick={() => setRescheduleModalOpen(false)}
                    className="px-4 py-2 rounded-xl text-xs font-semibold text-zinc-400 hover:text-white"
                  >
                    Отмена
                  </button>
                  <button
                    type="button"
                    onClick={handleExecuteReschedule}
                    disabled={rescheduling || !rescheduleTargetDate}
                    className="flex items-center gap-2 px-5 py-2.5 rounded-xl text-xs font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-all active:scale-95 disabled:opacity-40"
                  >
                    <ArrowRightLeft className="w-3.5 h-3.5" />
                    <span>{rescheduling ? 'Перенос...' : 'Подтвердить перенос'}</span>
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      ) : (
        <div className="py-20 text-center text-zinc-400">Тренировка не найдена</div>
      )}
    </div>
  );
};
