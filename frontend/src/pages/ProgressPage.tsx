import React, { useState, useEffect } from 'react';
import {
  TrendingUp,
  Scale,
  Plus,
  Calendar,
  Dumbbell,
  FileText,
  Flame,
  Award,
} from 'lucide-react';
import { WeightEntry, UserProfile, WorkoutHistoryEntry } from '../types';
import { api } from '../api/client';
import { useToast } from '../context/ToastContext';

interface ProgressPageProps {
  onRefresh?: () => void;
}

export const ProgressPage: React.FC<ProgressPageProps> = ({ onRefresh }) => {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);
  const [user, setUser] = useState<UserProfile | null>(null);
  const [weightHistory, setWeightHistory] = useState<WeightEntry[]>([]);
  const [workoutHistory, setWorkoutHistory] = useState<WorkoutHistoryEntry[]>([]);
  const [streak, setStreak] = useState(0);

  // Form states
  const [inputWeight, setInputWeight] = useState('');
  const [inputNote, setInputNote] = useState('');
  const [submittingWeight, setSubmittingWeight] = useState(false);

  const fetchProgress = async () => {
    try {
      setLoading(true);
      const res = await api.getProgress();
      setUser(res.user);
      setWeightHistory(res.weightHistory || []);
      setWorkoutHistory(res.workoutHistory || []);
      setStreak(res.streak || 0);

      // Pre-fill weight input with user's current weight or last entry
      if (res.weightHistory && res.weightHistory.length > 0) {
        setInputWeight(String(res.weightHistory[0].weight));
      } else if (res.user?.weight) {
        setInputWeight(String(res.user.weight));
      }
    } catch (e: any) {
      showToast(e.message || 'Ошибка загрузки прогресса', 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProgress();
  }, []);

  const handleAddWeight = async (e: React.FormEvent) => {
    e.preventDefault();
    const w = parseFloat(inputWeight);
    if (!w || isNaN(w) || w < 30 || w > 300) {
      showToast('Введите корректный вес от 30 до 300 кг', 'error');
      return;
    }

    try {
      setSubmittingWeight(true);
      const res = await api.addWeight(w, inputNote.trim());
      showToast('Замер веса успешно сохранён!');
      setInputNote('');
      // Update locally
      setWeightHistory([res.entry, ...weightHistory]);
      if (user) {
        setUser({ ...user, weight: w });
      }
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка сохранения замера', 'error');
    } finally {
      setSubmittingWeight(false);
    }
  };

  // SVG Chart calculation
  const sortedEntries = [...weightHistory].reverse();
  const weights = sortedEntries.map(e => e.weight);
  const minW = weights.length ? Math.min(...weights, user?.targetWeight || 70) - 2 : 65;
  const maxW = weights.length ? Math.max(...weights, user?.targetWeight || 85) + 2 : 85;

  const chartPoints = sortedEntries.map((e, idx) => {
    const x = sortedEntries.length === 1 ? 50 : (idx / (sortedEntries.length - 1)) * 100;
    const y = 100 - ((e.weight - minW) / (maxW - minW || 1)) * 100;
    return `${x},${y}`;
  }).join(' ');

  return (
    <div className="space-y-6">
      {/* KPI Top Row */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="rounded-2xl p-5 border bg-[#111827]/80 border-slate-800 shadow-sm flex items-center justify-between">
          <div>
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Текущий вес</span>
            <div className="text-2xl sm:text-3xl font-black text-white mt-1">
              {user?.weight || '--'} <span className="text-xs font-semibold text-slate-400">кг</span>
            </div>
            <div className="text-[11px] text-slate-400 mt-1">
              Цель: <span className="font-semibold text-emerald-400">{user?.targetWeight || '--'} кг</span>
            </div>
          </div>
          <div className="p-3 rounded-xl bg-emerald-500/15 text-emerald-400">
            <Scale className="w-5 h-5" />
          </div>
        </div>

        <div className="rounded-2xl p-5 border bg-[#111827]/80 border-slate-800 shadow-sm flex items-center justify-between">
          <div>
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Серия активности</span>
            <div className="text-2xl sm:text-3xl font-black text-amber-400 mt-1">
              {streak} <span className="text-xs font-semibold text-slate-400">дней подряд</span>
            </div>
            <div className="text-[11px] text-slate-400 mt-1">
              Дисциплина формирует результат
            </div>
          </div>
          <div className="p-3 rounded-xl bg-amber-500/15 text-amber-400">
            <Flame className="w-5 h-5" />
          </div>
        </div>

        <div className="rounded-2xl p-5 border bg-[#111827]/80 border-slate-800 shadow-sm flex items-center justify-between">
          <div>
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">ИМТ и статус</span>
            <div className="text-2xl sm:text-3xl font-black text-white mt-1">
              {user?.bmi?.toFixed(1) || '22.0'}
            </div>
            <div className="text-[11px] text-emerald-400 font-semibold mt-1">
              Нормальный весовой диапазон
            </div>
          </div>
          <div className="p-3 rounded-xl bg-sky-500/15 text-sky-400">
            <TrendingUp className="w-5 h-5" />
          </div>
        </div>
      </div>

      {loading ? (
        <div className="py-20 text-center text-slate-400">Загрузка данных прогресса...</div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Left Column: Weight Logging & History */}
          <div className="space-y-6">
            {/* New Weight Entry Form */}
            <div className="rounded-2xl p-6 border bg-[#111827]/80 border-slate-800 shadow-sm space-y-4">
              <div className="flex items-center gap-2.5 border-b border-slate-800/80 pb-3">
                <Scale className="w-5 h-5 text-emerald-400" />
                <h3 className="text-base font-bold text-white">Внести замер веса</h3>
              </div>

              <form onSubmit={handleAddWeight} className="space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-semibold text-slate-400 mb-1.5">
                      Вес (кг)
                    </label>
                    <input
                      type="number"
                      step="0.1"
                      value={inputWeight}
                      onChange={e => setInputWeight(e.target.value)}
                      placeholder="Например, 75.4"
                      className="w-full px-3.5 py-2.5 rounded-xl bg-slate-900 border border-slate-700/80 text-white font-bold text-base focus:border-emerald-500 focus:outline-none"
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-slate-400 mb-1.5">
                      Заметка к замеру (видна в истории!)
                    </label>
                    <input
                      type="text"
                      value={inputNote}
                      onChange={e => setInputNote(e.target.value)}
                      placeholder="Напр., после пробежки, легкость"
                      className="w-full px-3.5 py-2.5 rounded-xl bg-slate-900 border border-slate-700/80 text-white text-xs sm:text-sm focus:border-emerald-500 focus:outline-none"
                    />
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={submittingWeight}
                  className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl text-xs sm:text-sm font-bold bg-emerald-500 hover:bg-emerald-400 text-slate-950 transition-all shadow-md shadow-emerald-500/20 disabled:opacity-50"
                >
                  <Plus className="w-4 h-4" />
                  <span>{submittingWeight ? 'Сохранение...' : 'Зафиксировать вес'}</span>
                </button>
              </form>
            </div>

            {/* SVG Weight Dynamics Chart */}
            <div className="rounded-2xl p-6 border bg-[#111827]/80 border-slate-800 shadow-sm space-y-4">
              <div className="flex items-center justify-between border-b border-slate-800/80 pb-3">
                <div className="flex items-center gap-2">
                  <TrendingUp className="w-5 h-5 text-emerald-400" />
                  <h3 className="text-base font-bold text-white">Динамика веса</h3>
                </div>
                <span className="text-xs text-slate-400">
                  {sortedEntries.length} замеров
                </span>
              </div>

              {sortedEntries.length > 1 ? (
                <div className="relative pt-2 pb-1">
                  <div className="h-44 w-full flex items-end">
                    <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="w-full h-full overflow-visible">
                      <polyline
                        fill="none"
                        stroke="#22c55e"
                        strokeWidth="2.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        points={chartPoints}
                      />
                    </svg>
                  </div>
                  <div className="flex justify-between text-[11px] text-slate-400 mt-2 font-medium">
                    <span>Старт: {sortedEntries[0].weight} кг</span>
                    <span>Сейчас: {sortedEntries[sortedEntries.length - 1].weight} кг</span>
                  </div>
                </div>
              ) : (
                <div className="p-6 text-center text-xs text-slate-400 bg-slate-900/30 rounded-xl border border-slate-800/60">
                  Добавьте хотя бы 2 замера для построения интерактивного графика динамики.
                </div>
              )}
            </div>

            {/* Weight History Journal with visible notes */}
            <div className="rounded-2xl p-6 border bg-[#111827]/80 border-slate-800 shadow-sm space-y-3">
              <div className="flex items-center gap-2 border-b border-slate-800/80 pb-3">
                <FileText className="w-5 h-5 text-sky-400" />
                <h3 className="text-base font-bold text-white">История замеров веса и заметки</h3>
              </div>

              {weightHistory.length > 0 ? (
                <div className="space-y-2.5 max-h-72 overflow-y-auto pr-1">
                  {weightHistory.map((item, idx) => (
                    <div
                      key={idx}
                      className="p-3 rounded-xl bg-slate-900/60 border border-slate-800/80 flex items-start justify-between gap-3"
                    >
                      <div className="space-y-0.5">
                        <div className="text-sm font-bold text-white flex items-center gap-2">
                          <span>{item.weight} кг</span>
                          <span className="text-[11px] text-slate-400 font-normal">
                            • {item.date}
                          </span>
                        </div>
                        {item.note && (
                          <p className="text-xs text-slate-300 bg-slate-800/60 px-2 py-1 rounded-md mt-1 border border-slate-700/40">
                            📝 {item.note}
                          </p>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-6 text-xs text-slate-400">
                  Замеров пока нет. Зафиксируйте первый вес выше!
                </div>
              )}
            </div>
          </div>

          {/* Right Column: Workout History Activity */}
          <div className="rounded-2xl p-6 border bg-[#111827]/80 border-slate-800 shadow-sm space-y-4">
            <div className="flex items-center gap-2.5 border-b border-slate-800/80 pb-3">
              <Dumbbell className="w-5 h-5 text-emerald-400" />
              <h3 className="text-base font-bold text-white">История выполненных тренировок</h3>
            </div>

            {workoutHistory.length > 0 ? (
              <div className="space-y-3">
                {workoutHistory.map((w, idx) => (
                  <div
                    key={idx}
                    className="p-4 rounded-xl bg-slate-900/60 border border-slate-800/80 flex items-center justify-between gap-3 hover:border-slate-700 transition-colors"
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-bold uppercase tracking-wider text-emerald-400">
                          {w.date}
                        </span>
                        <span className="text-xs px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-semibold">
                          Выполнено
                        </span>
                      </div>
                      <h4 className="text-sm font-bold text-white">
                        {w.dayName || `Тренировка ${w.dayIndex || ''}`}
                      </h4>
                      <p className="text-xs text-slate-400">
                        Фокус: {w.focus || 'Комплекс упражнений'}
                      </p>
                    </div>
                    <div className="text-right flex-shrink-0">
                      <span className="text-xs font-bold text-slate-300 block">
                        {w.exercisesCount || 4} упр.
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="p-8 text-center rounded-xl bg-slate-900/40 border border-slate-800/60 space-y-2">
                <Award className="w-8 h-8 text-slate-500 mx-auto opacity-75" />
                <div className="text-sm font-semibold text-white">История пока пуста</div>
                <p className="text-xs text-slate-400 max-w-xs mx-auto">
                  Перейдите во вкладку «Тренировки» и нажмите «Отметить как выполненную», чтобы зафиксировать вашу активность.
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
