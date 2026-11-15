import React, { useState, useEffect } from 'react';
import {
  Dumbbell,
  Clock,
  RotateCw,
  CheckCircle2,
  Info,
  ShieldAlert,
  ChevronRight,
  Flame,
  Award,
  Sparkles,
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

  const fetchWorkout = async (day: number) => {
    try {
      setLoading(true);
      const res = await api.getWorkouts(day);
      setWorkout(res.workout);
      setAllDays(res.allDays || []);
      setUser(res.user);
      setSelectedDay(res.selectedDay);
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
      // Update workout exercise locally
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

  const weekdayShort = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

  return (
    <div className="space-y-6">
      {/* 7-Day Selector Tabs */}
      <div className="grid grid-cols-7 gap-1.5 sm:gap-2 p-1.5 rounded-2xl bg-[#111827]/80 border border-white/[0.08] backdrop-blur-sm">
        {[1, 2, 3, 4, 5, 6, 7].map(day => {
          const isSelected = selectedDay === day;
          const dayData = allDays.find(d => d.dayIndex === day);
          const isRest = dayData?.isRest;
          const isDone = dayData?.completed;

          return (
            <button
              key={day}
              onClick={() => setSelectedDay(day)}
              className={`flex flex-col items-center justify-center py-2.5 px-1 rounded-xl text-xs font-bold transition-all duration-150 ${
                isSelected
                  ? 'bg-emerald-500 text-slate-950 shadow-md shadow-emerald-500/20 scale-[1.02]'
                  : 'text-slate-300 hover:text-white hover:bg-slate-800/60'
              }`}
            >
              <span className="text-[10px] uppercase font-semibold opacity-75">
                {weekdayShort[day - 1]}
              </span>
              <span className="text-sm font-black mt-0.5">День {day}</span>
              <span className="text-[10px] mt-0.5 font-medium">
                {isDone ? '✓ Готово' : isRest ? 'Отдых' : 'Тренинг'}
              </span>
            </button>
          );
        })}
      </div>

      {loading ? (
        <div className="py-20 text-center text-slate-400">Загрузка тренировки...</div>
      ) : workout ? (
        <div className="space-y-6">
          {/* Header Card */}
          <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-bold uppercase tracking-wider text-emerald-400">
                  {workout.weekdayName} • День {selectedDay}
                </span>
                {workout.completed && (
                  <span className="flex items-center gap-1 text-[11px] font-bold px-2 py-0.5 rounded-md bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                    <CheckCircle2 className="w-3 h-3" />
                    Выполнено
                  </span>
                )}
              </div>
              <h2 className="text-2xl font-black text-white mt-1">
                {workout.isRest ? 'День восстановления и отдыха' : workout.dayName}
              </h2>
              <p className="text-xs sm:text-sm text-slate-400 mt-0.5">
                Фокус: <span className="text-slate-200 font-medium">{workout.focus}</span> • Примерное время: ~{workout.estimatedMinutes} мин
              </p>
            </div>

            {!workout.isRest && (
              <button
                onClick={handleComplete}
                disabled={completing || workout.completed}
                className={`flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold transition-all shadow-lg ${
                  workout.completed
                    ? 'bg-slate-800 text-emerald-400 border border-emerald-500/30 cursor-default'
                    : 'bg-emerald-500 hover:bg-emerald-400 text-slate-950 shadow-emerald-500/20 active:scale-95'
                }`}
              >
                <CheckCircle2 className="w-4 h-4" />
                <span>{workout.completed ? 'Тренировка выполнена' : 'Отметить как выполненную'}</span>
              </button>
            )}
          </div>

          {/* Gentle mode alert if active */}
          {user?.hasHealthIssues && (
            <div className="rounded-xl p-3.5 border bg-[#121216] border-amber-500/40 text-slate-200 text-xs flex items-center gap-3">
              <ShieldAlert className="w-4 h-4 text-amber-400 flex-shrink-0" />
              <span>
                Щадящий режим: упражнения подобраны без осевой ударной нагрузки на позвоночник и суставы.
              </span>
            </div>
          )}

          {/* Exercise list or Rest Day screen */}
          {workout.isRest ? (
            <div className="rounded-2xl p-12 border bg-[#121216] border-white/[0.08] text-center max-w-xl mx-auto space-y-4">
              <div className="w-14 h-14 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center mx-auto text-emerald-400">
                <Sparkles className="w-7 h-7" />
              </div>
              <h3 className="text-xl font-bold text-white">Время восстановить силы</h3>
              <p className="text-sm text-slate-200 leading-relaxed">
                Мышцы растут и адаптируются именно во время отдыха. Проведите легкую пешую прогулку, пейте достаточно воды и соблюдайте режим сна.
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-base font-bold text-white flex items-center gap-2">
                  <Dumbbell className="w-4 h-4 text-emerald-400" />
                  Упражнения программы ({workout.exercises?.length || 0})
                </h3>
                <span className="text-xs text-slate-400">
                  Если упражнение не подходит — нажмите «Заменить»
                </span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {workout.exercises?.map((ex, idx) => (
                  <div
                    key={ex.id || idx}
                    className="rounded-2xl p-5 border bg-[#121216] border-white/[0.08] shadow-sm hover:border-slate-700 transition-all flex flex-col justify-between gap-4 group"
                  >
                    <div>
                      <div className="flex items-start justify-between gap-2 mb-2">
                        <div>
                          <span className="text-[11px] font-bold uppercase tracking-wider text-emerald-400 block mb-0.5">
                            Упражнение {idx + 1}
                          </span>
                          <h4 className="text-base font-bold text-white group-hover:text-emerald-300 transition-colors">
                            {ex.title}
                          </h4>
                        </div>
                        <span className="text-xs font-bold px-2.5 py-1 rounded-lg bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 flex-shrink-0">
                          {ex.sets} × {ex.reps}
                        </span>
                      </div>

                      <div className="flex flex-wrap items-center gap-1.5 mt-2">
                        <span className="text-[11px] px-2 py-0.5 rounded-md bg-slate-800 text-slate-300 font-medium">
                          {ex.muscle}
                        </span>
                        <span className="text-[11px] px-2 py-0.5 rounded-md bg-slate-800 text-slate-300 font-medium">
                          {ex.equipment}
                        </span>
                        <span className="text-[11px] px-2 py-0.5 rounded-md bg-slate-800/60 text-slate-400 font-medium">
                          Отдых {ex.restSec || 60}с
                        </span>
                      </div>

                      {ex.coachTip && (
                        <p className="text-xs text-slate-200 mt-3 line-clamp-2 leading-relaxed bg-slate-900/60 p-2.5 rounded-xl border border-white/[0.08]/60">
                          💡 <span className="text-slate-300 font-medium">Совет тренера:</span> {ex.coachTip}
                        </p>
                      )}
                    </div>

                    <div className="pt-2 border-t border-white/[0.08]/80 flex items-center justify-between">
                      <button
                        onClick={() => setDetailExercise(ex)}
                        className="text-xs font-bold text-slate-300 hover:text-white flex items-center gap-1 transition-colors"
                      >
                        <Info className="w-3.5 h-3.5 text-slate-500" />
                        Техника
                      </button>

                      <button
                        onClick={() => handleSwap(ex.id)}
                        disabled={swappingId === ex.id}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold bg-slate-800 hover:bg-slate-700 text-slate-200 hover:text-white border border-slate-700/80 transition-all active:scale-95 disabled:opacity-50"
                      >
                        <RotateCw className={`w-3.5 h-3.5 ${swappingId === ex.id ? 'animate-spin text-emerald-400' : ''}`} />
                        <span>{swappingId === ex.id ? 'Подбор...' : 'Заменить'}</span>
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Exercise Detail Modal */}
          {detailExercise && (
            <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in">
              <div className="rounded-2xl p-6 bg-[#121216] border border-white/20 shadow-2xl max-w-lg w-full space-y-4">
                <div className="flex items-start justify-between">
                  <div>
                    <span className="text-xs font-bold uppercase tracking-wider text-emerald-400">
                      Техника выполнения
                    </span>
                    <h3 className="text-xl font-bold text-white mt-1">{detailExercise.title}</h3>
                  </div>
                  <button
                    onClick={() => setDetailExercise(null)}
                    className="p-1 rounded-lg text-slate-300 hover:text-white hover:bg-slate-800"
                  >
                    ✕
                  </button>
                </div>

                <div className="flex gap-2">
                  <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-slate-800 text-slate-200">
                    Мышцы: {detailExercise.muscle}
                  </span>
                  <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-slate-800 text-slate-200">
                    Инвентарь: {detailExercise.equipment}
                  </span>
                </div>

                <div className="p-4 rounded-xl bg-slate-900 border border-white/[0.08] text-xs sm:text-sm text-slate-300 leading-relaxed space-y-2">
                  <div className="font-bold text-white flex items-center gap-1.5">
                    <Sparkles className="w-4 h-4 text-emerald-400" />
                    Рекомендации по выполнению:
                  </div>
                  <p>{detailExercise.coachTip || 'Выполняйте упражнение в подконтрольном темпе, сохраняя ровное дыхание.'}</p>
                </div>

                <div className="flex items-center justify-between text-xs text-slate-400 pt-2 border-t border-white/[0.08]">
                  <span>Рекомендуемый сет: {detailExercise.sets} подходов по {detailExercise.reps}</span>
                  <button
                    onClick={() => setDetailExercise(null)}
                    className="px-4 py-2 rounded-xl text-xs font-bold bg-emerald-500 hover:bg-emerald-400 text-slate-950 transition-colors"
                  >
                    Понятно
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      ) : (
        <div className="py-20 text-center text-slate-400">Тренировка не найдена</div>
      )}
    </div>
  );
};
