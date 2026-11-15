import React, { useState, useEffect } from 'react';
import {
  Dumbbell,
  Clock,
  RotateCw,
  CheckCircle2,
  Info,
  ShieldAlert,
  Calendar,
  ChevronLeft,
  ChevronRight,
  ArrowRightLeft,
  FileText,
  Save,
  Sparkles,
  X,
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
  const [selectedDay, setSelectedDay] = useState(initialDay || 1);
  const [workout, setWorkout] = useState<WorkoutDay | null>(null);
  const [allDays, setAllDays] = useState<WorkoutDay[]>([]);
  const [user, setUser] = useState<UserProfile | null>(null);
  const [swappingId, setSwappingId] = useState<string | null>(null);
  const [completing, setCompleting] = useState(false);
  const [detailExercise, setDetailExercise] = useState<PlanExercise | null>(null);

  // Calendar state
  const [weekOffset, setWeekOffset] = useState(0);
  const [calendarModalOpen, setCalendarModalOpen] = useState(false);
  const [calendarYear, setCalendarYear] = useState(() => new Date().getFullYear());
  const [calendarMonth, setCalendarMonth] = useState(() => new Date().getMonth());

  // Workout rescheduling modal
  const [rescheduleModalOpen, setRescheduleModalOpen] = useState(false);
  const [rescheduleTargetDay, setRescheduleTargetDay] = useState<number>(1);
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

  const handleComplete = async () => {
    try {
      setCompleting(true);
      await api.completeWorkout(selectedDay, true);
      showToast('Тренировка отмечена как выполненная! Отличная работа!');
      if (workout) {
        setWorkout({ ...workout, completed: true });
      }
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка завершения тренировки', 'error');
    } finally {
      setCompleting(false);
    }
  };

  const handleReschedule = async () => {
    if (rescheduleTargetDay === selectedDay) {
      showToast('Выберите день, отличный от текущего', 'error');
      return;
    }
    try {
      setRescheduling(true);
      await api.rescheduleWorkout(selectedDay, rescheduleTargetDay);
      showToast(`Тренировка перенесена: День ${selectedDay} ⇄ День ${rescheduleTargetDay}`);
      setRescheduleModalOpen(false);
      await fetchWorkout(rescheduleTargetDay);
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

  // Calendar date math
  const today = new Date();
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
    return target;
  };

  const isSameDay = (d1: Date, d2: Date) => {
    return (
      d1.getFullYear() === d2.getFullYear() &&
      d1.getMonth() === d2.getMonth() &&
      d1.getDate() === d2.getDate()
    );
  };

  // Week range label (e.g., "10 ноя – 16 ноя 2026")
  const mondayDate = getDateForDay(1, weekOffset);
  const sundayDate = getDateForDay(7, weekOffset);
  const weekLabel = `${mondayDate.getDate()} ${monthsShortRu[mondayDate.getMonth()]} – ${sundayDate.getDate()} ${monthsShortRu[sundayDate.getMonth()]} ${sundayDate.getFullYear()}`;

  // Big iOS Calendar calculations
  const firstDayOfMonth = new Date(calendarYear, calendarMonth, 1);
  const daysInMonth = new Date(calendarYear, calendarMonth + 1, 0).getDate();
  // Monday = 0 ... Sunday = 6
  const monthStartOffset = (firstDayOfMonth.getDay() + 6) % 7;

  const handleSelectCalendarDate = (dayNum: number) => {
    const targetDate = new Date(calendarYear, calendarMonth, dayNum);
    const targetIsoDay = targetDate.getDay() === 0 ? 7 : targetDate.getDay();

    // Compute Monday of clicked date
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

  return (
    <div className="space-y-6">
      {/* Top Calendar Bar: Week navigation, dates, iOS calendar trigger */}
      <div className="rounded-2xl p-4 sm:p-5 border bg-[#121216] border-white/[0.08] shadow-sm space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-white/10 text-white border border-white/20">
              <Calendar className="w-5 h-5 text-white" />
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

            {/* Big iOS Calendar modal trigger */}
            <button
              onClick={() => {
                setCalendarYear(today.getFullYear());
                setCalendarMonth(today.getMonth());
                setCalendarModalOpen(true);
              }}
              className="flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-bold bg-white/10 hover:bg-white/20 text-white border border-white/20 transition-all active:scale-95 shadow-sm"
            >
              <Calendar className="w-4 h-4 text-white" />
              <span>Календарь iOS</span>
            </button>
          </div>
        </div>

        {/* 7-Day Real Date Selector Strip */}
        <div className="grid grid-cols-7 gap-1.5 sm:gap-2 pt-1">
          {[1, 2, 3, 4, 5, 6, 7].map(day => {
            const isSelected = selectedDay === day;
            const dayData = allDays.find(d => d.dayIndex === day);
            const isRest = dayData?.isRest;
            const isDone = dayData?.completed;
            const dayDate = getDateForDay(day, weekOffset);
            const isToday = isSameDay(dayDate, today);

            return (
              <button
                key={day}
                onClick={() => setSelectedDay(day)}
                className={`relative flex flex-col items-center justify-center py-2.5 px-1 rounded-xl text-xs font-bold transition-all duration-150 ${
                  isSelected
                    ? 'bg-white text-zinc-950 shadow-lg shadow-white/10 scale-[1.02] border border-white'
                    : isToday
                    ? 'bg-zinc-800/80 text-white border border-white/40 hover:bg-zinc-800'
                    : 'bg-[#17171d] text-zinc-400 hover:text-white hover:bg-[#202028] border border-white/[0.06]'
                }`}
              >
                {/* Indicator dot if completed or training */}
                <div className="flex items-center gap-1 mb-0.5">
                  <span className={`text-[10px] uppercase font-semibold ${isSelected ? 'text-zinc-700' : 'text-zinc-400'}`}>
                    {weekdayShort[day - 1]}
                  </span>
                  {isToday && (
                    <span className={`w-1.5 h-1.5 rounded-full ${isSelected ? 'bg-zinc-950' : 'bg-white animate-pulse'}`} />
                  )}
                </div>

                {/* Real calendar date number */}
                <span className="text-sm font-black">
                  {dayDate.getDate()} {monthsShortRu[dayDate.getMonth()]}
                </span>

                <span className={`text-[10px] mt-1 font-medium truncate max-w-full px-1 ${
                  isSelected
                    ? 'text-zinc-800 font-bold'
                    : isDone
                    ? 'text-white'
                    : isRest
                    ? 'text-zinc-500'
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
                  {workout.weekdayName} • День {selectedDay}
                </span>
                {workout.completed && (
                  <span className="flex items-center gap-1 text-[11px] font-bold px-2.5 py-0.5 rounded-md bg-white/10 text-white border border-white/20">
                    <CheckCircle2 className="w-3 h-3 text-white" />
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
                  const target = selectedDay === 7 ? 1 : selectedDay + 1;
                  setRescheduleTargetDay(target);
                  setRescheduleModalOpen(true);
                }}
                className="flex items-center gap-1.5 px-4 py-2.5 rounded-xl text-xs font-bold bg-[#17171d] hover:bg-[#22222c] text-zinc-200 hover:text-white border border-white/[0.12] transition-all active:scale-95"
                title="Перенести тренировку на другой день"
              >
                <ArrowRightLeft className="w-4 h-4 text-zinc-300" />
                <span>Перенести</span>
              </button>

              {!workout.isRest && (
                <button
                  onClick={handleComplete}
                  disabled={completing || workout.completed}
                  className={`flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold transition-all shadow-md ${
                    workout.completed
                      ? 'bg-zinc-800 text-zinc-300 border border-white/20 cursor-default'
                      : 'bg-white hover:bg-zinc-200 text-zinc-950 shadow-white/10 active:scale-95'
                  }`}
                >
                  <CheckCircle2 className="w-4 h-4" />
                  <span>{workout.completed ? 'Выполнена' : 'Отметить как выполненную'}</span>
                </button>
              )}
            </div>
          </div>

          {/* Gentle mode alert if active */}
          {user?.hasHealthIssues && (
            <div className="rounded-xl p-3.5 border bg-[#121216] border-amber-500/40 text-zinc-200 text-xs flex items-center gap-3">
              <ShieldAlert className="w-4 h-4 text-amber-400 flex-shrink-0" />
              <span>
                Щадящий режим: упражнения подобраны без осевой ударной нагрузки на позвоночник и суставы.
              </span>
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
                  Упражнения программы ({workout.exercises?.length || 0})
                </h3>
                <span className="text-xs text-zinc-400">
                  Если упражнение не подходит — нажмите «Заменить»
                </span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {workout.exercises?.map((ex, idx) => (
                  <div
                    key={ex.id || idx}
                    className="rounded-2xl p-5 border bg-[#121216] border-white/[0.08] shadow-sm hover:border-white/[0.20] transition-all flex flex-col justify-between gap-4 group"
                  >
                    <div>
                      <div className="flex items-start justify-between gap-2 mb-2">
                        <div>
                          <span className="text-[11px] font-bold uppercase tracking-wider text-zinc-400 block mb-0.5">
                            Упражнение {idx + 1}
                          </span>
                          <h4 className="text-base font-bold text-white group-hover:text-zinc-200 transition-colors">
                            {ex.title}
                          </h4>
                        </div>
                        <span className="text-xs font-bold px-2.5 py-1 rounded-lg bg-white/10 text-white border border-white/20 flex-shrink-0">
                          {ex.sets} × {ex.reps}
                        </span>
                      </div>

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
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold bg-[#17171d] hover:bg-[#252530] text-zinc-200 hover:text-white border border-white/[0.12] transition-all active:scale-95 disabled:opacity-50"
                      >
                        <RotateCw className={`w-3.5 h-3.5 ${swappingId === ex.id ? 'animate-spin text-white' : ''}`} />
                        <span>{swappingId === ex.id ? 'Подбор...' : 'Заменить'}</span>
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* DEDICATED WORKOUT NOTES CARD ("Где писать заметки") */}
          <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-4">
            <div className="flex items-center justify-between border-b border-white/[0.08] pb-3">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-xl bg-white/10 text-white border border-white/20">
                  <FileText className="w-4 h-4 text-white" />
                </div>
                <div>
                  <h3 className="text-base font-bold text-white">
                    Заметки к тренировке (День {selectedDay})
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

          {/* Big iOS-Style Calendar Modal */}
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

                {/* Days grid iOS style */}
                <div className="grid grid-cols-7 gap-1 text-center">
                  {/* Empty cells */}
                  {Array.from({ length: monthStartOffset }).map((_, i) => (
                    <div key={`empty-${i}`} className="h-10" />
                  ))}

                  {/* Month days */}
                  {Array.from({ length: daysInMonth }).map((_, i) => {
                    const dayNum = i + 1;
                    const dateObj = new Date(calendarYear, calendarMonth, dayNum);
                    const isTodayCell = isSameDay(dateObj, today);
                    const isoDay = dateObj.getDay() === 0 ? 7 : dateObj.getDay();
                    const dayData = allDays.find(d => d.dayIndex === isoDay);
                    const isRestDay = dayData?.isRest;

                    return (
                      <button
                        key={dayNum}
                        onClick={() => handleSelectCalendarDate(dayNum)}
                        className={`h-11 rounded-2xl flex flex-col items-center justify-center text-xs transition-all ${
                          isTodayCell
                            ? 'bg-white text-zinc-950 font-black shadow-md'
                            : 'hover:bg-white/10 text-zinc-200'
                        }`}
                      >
                        <span className="font-bold leading-none">{dayNum}</span>
                        {/* Dot indicator */}
                        <span className={`w-1 h-1 rounded-full mt-1 ${
                          isTodayCell
                            ? 'bg-zinc-950'
                            : isRestDay
                            ? 'bg-zinc-600'
                            : 'bg-white'
                        }`} />
                      </button>
                    );
                  })}
                </div>

                {/* Footer notes */}
                <div className="pt-2 border-t border-white/[0.08] flex items-center justify-between text-xs text-zinc-400">
                  <div className="flex items-center gap-3">
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
                    К сегодняшнему дню
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Reschedule Modal ("Перенести тренировку") */}
          {rescheduleModalOpen && (
            <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in">
              <div className="rounded-2xl p-6 bg-[#121216] border border-white/20 shadow-2xl max-w-md w-full space-y-4">
                <div className="flex items-start justify-between">
                  <div>
                    <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                      Перенос тренировки
                    </span>
                    <h3 className="text-lg font-black text-white mt-1">
                      Перенести День {selectedDay} ({workout.dayName})
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
                  Выберите день недели, с которым вы хотите поменять местами текущую программу. Все упражнения и статус сохранятся.
                </p>

                <div className="space-y-2 pt-1">
                  {[1, 2, 3, 4, 5, 6, 7].map(d => {
                    if (d === selectedDay) return null;
                    const dData = allDays.find(item => item.dayIndex === d);
                    const isSelected = rescheduleTargetDay === d;

                    return (
                      <button
                        key={d}
                        type="button"
                        onClick={() => setRescheduleTargetDay(d)}
                        className={`w-full p-3 rounded-xl border text-left flex items-center justify-between transition-all ${
                          isSelected
                            ? 'bg-white/10 text-white border-white'
                            : 'bg-[#17171d] text-zinc-300 border-white/[0.08] hover:border-white/20'
                        }`}
                      >
                        <div>
                          <div className="text-xs font-bold text-white">
                            День {d} ({weekdayShort[d - 1]})
                          </div>
                          <div className="text-[11px] text-zinc-400">
                            {dData ? (dData.isRest ? 'День восстановления' : dData.dayName) : 'Программа'}
                          </div>
                        </div>
                        {isSelected && (
                          <span className="text-xs font-bold px-2 py-0.5 rounded bg-white text-zinc-950">
                            Выбран
                          </span>
                        )}
                      </button>
                    );
                  })}
                </div>

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
                    onClick={handleReschedule}
                    disabled={rescheduling}
                    className="flex items-center gap-2 px-5 py-2.5 rounded-xl text-xs font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-all active:scale-95 disabled:opacity-50"
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
