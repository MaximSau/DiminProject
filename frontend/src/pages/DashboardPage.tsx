import React, { useState } from 'react';
import {
  Flame,
  Droplets,
  Dumbbell,
  Clock,
  ShieldAlert,
  ChevronRight,
  Plus,
  Minus,
  Moon,
  Sun,
  Activity,
  CheckCircle2,
  FileText,
} from 'lucide-react';
import { DashboardResponse } from '../types';
import { api } from '../api/client';
import { useToast } from '../context/ToastContext';

interface DashboardPageProps {
  data: DashboardResponse;
  onNavigate: (tab: string, params?: any) => void;
  onRefresh: () => void;
}

export const DashboardPage: React.FC<DashboardPageProps> = ({ data, onNavigate, onRefresh }) => {
  const { user, todayWorkout, todayMeals, todayWater, streak, isGentleMode } = data;
  const { showToast } = useToast();
  const [addingWater, setAddingWater] = useState(false);

  const handleAddWater = async (delta: number) => {
    try {
      setAddingWater(true);
      await api.addWater(delta);
      showToast(delta > 0 ? `+${delta} мл воды добавлено` : `${delta} мл воды списано`);
      onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка обновления водного баланса', 'error');
    } finally {
      setAddingWater(false);
    }
  };

  const calTarget = user.calorieTarget || 2200;
  const calCurrent = todayMeals?.currentCalories || 0;
  const calPercent = Math.min(100, Math.round((calCurrent / calTarget) * 100));

  const waterTarget = user.waterTarget || 2000;
  const waterCurrent = todayWater || 0;
  const waterPercent = Math.min(100, Math.round((waterCurrent / waterTarget) * 100));
  const waterGlasses = Math.floor(waterCurrent / 330);
  const targetGlasses = Math.ceil(waterTarget / 330);

  return (
    <div className="space-y-6">
      {/* Gentle Medical Mode Banner */}
      {isGentleMode && (
        <div className="rounded-2xl p-4 sm:p-5 border bg-[#121216] border-amber-500/40 text-zinc-100 flex items-start gap-4 shadow-lg backdrop-blur-sm">
          <div className="p-2.5 rounded-xl bg-amber-500/20 text-amber-400 border border-amber-500/30 flex-shrink-0 mt-0.5">
            <ShieldAlert className="w-5 h-5 text-amber-400" />
          </div>
          <div className="space-y-1">
            <h4 className="font-bold text-white text-sm sm:text-base flex items-center gap-2">
              Активирован щадящий режим: восстановление и мобильность
            </h4>
            <p className="text-xs sm:text-sm text-zinc-300 leading-relaxed font-normal">
              В вашем профиле отмечены медицинские особенности. Программа автоматически адаптирована: исключены ударные осевые нагрузки, прыжки и высокоинтенсивные берпи. Перед началом занятий обязательно проконсультируйтесь с лечащим врачом.
            </p>
          </div>
        </div>
      )}

      {/* Top Welcome & KPI row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Calories Card */}
        <div className="rounded-2xl p-5 border bg-[#121216] border-white/[0.08] shadow-sm relative overflow-hidden group hover:border-white/[0.18] transition-colors">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">Калории за день</span>
            <div className="p-2 rounded-xl bg-rose-500/15 text-rose-400 border border-rose-500/20">
              <Flame className="w-4 h-4" />
            </div>
          </div>
          <div className="flex items-baseline gap-2 mb-2">
            <span className="text-2xl sm:text-3xl font-black text-white">{calCurrent}</span>
            <span className="text-xs font-semibold text-zinc-400">/ {calTarget} ккал</span>
          </div>
          <div className="w-full bg-[#1c1c24] h-2 rounded-full overflow-hidden border border-white/[0.04]">
            <div
              className="bg-gradient-to-r from-rose-500 to-amber-400 h-full rounded-full transition-all duration-500"
              style={{ width: `${calPercent}%` }}
            />
          </div>
          <div className="mt-2.5 text-[11px] text-zinc-400 flex justify-between font-medium">
            <span>{calPercent}% выполнено</span>
            <span>Осталось {Math.max(0, calTarget - calCurrent)} ккал</span>
          </div>
        </div>

        {/* Water Card with +330 and -330 ml buttons */}
        <div className="rounded-2xl p-5 border bg-[#121216] border-white/[0.08] shadow-sm relative overflow-hidden group hover:border-white/[0.18] transition-colors">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">Водный баланс</span>
            <div className="p-2 rounded-xl bg-sky-500/15 text-sky-400 border border-sky-500/20">
              <Droplets className="w-4 h-4" />
            </div>
          </div>
          <div className="flex items-baseline justify-between mb-2">
            <div className="flex items-baseline gap-2">
              <span className="text-2xl sm:text-3xl font-black text-white">{waterCurrent}</span>
              <span className="text-xs font-semibold text-zinc-400">/ {waterTarget} мл</span>
            </div>
            <span className="text-xs font-bold px-2 py-0.5 rounded-md bg-sky-500/15 text-sky-300 border border-sky-500/30">
              {waterGlasses} / {targetGlasses} стак.
            </span>
          </div>
          <div className="w-full bg-[#1c1c24] h-2 rounded-full overflow-hidden border border-white/[0.04]">
            <div
              className="bg-gradient-to-r from-sky-500 to-sky-400 h-full rounded-full transition-all duration-500"
              style={{ width: `${waterPercent}%` }}
            />
          </div>
          <div className="mt-3 flex items-center justify-between">
            <span className="text-[11px] text-zinc-400 font-medium">1 стак. = 330 мл</span>
            <div className="flex items-center gap-1.5">
              <button
                onClick={() => handleAddWater(-330)}
                disabled={addingWater || waterCurrent <= 0}
                className="flex items-center gap-1 px-2 py-1 rounded-lg text-xs font-bold bg-[#1c1c24] hover:bg-[#252530] text-zinc-200 border border-white/[0.10] transition-all active:scale-95 disabled:opacity-40"
                title="Убрать 1 стакан воды (-330 мл)"
              >
                <Minus className="w-3.5 h-3.5" />
                <span>-330</span>
              </button>

              <button
                onClick={() => handleAddWater(330)}
                disabled={addingWater}
                className="flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-bold bg-white/10 hover:bg-white/20 text-white border border-white/20 transition-all active:scale-95 disabled:opacity-50"
                title="Добавить 1 стакан воды (+330 мл)"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>+330</span>
              </button>
            </div>
          </div>
        </div>

        {/* Workout Activity Card */}
        <div className="rounded-2xl p-5 border bg-[#121216] border-white/[0.08] shadow-sm relative overflow-hidden group hover:border-white/[0.18] transition-colors">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">Тренировка сегодня</span>
            <div className="p-2 rounded-xl bg-white/10 text-white border border-white/20">
              <Dumbbell className="w-4 h-4 text-white" />
            </div>
          </div>
          <div className="mb-2">
            <div className="text-lg font-bold text-white truncate">
              {todayWorkout?.isRest ? 'День восстановления' : todayWorkout?.dayName || 'Тренировка'}
            </div>
            <div className="text-xs text-zinc-400 truncate font-medium">
              {todayWorkout?.focus || 'Отдых и растяжка'}
            </div>
          </div>
          <div className="mt-4 flex items-center justify-between text-xs">
            <span className="text-zinc-400 flex items-center gap-1.5 font-medium">
              <Clock className="w-3.5 h-3.5 text-zinc-400" />
              ~{todayWorkout?.estimatedMinutes || user.workoutDuration || 45} мин
            </span>
            <button
              onClick={() => onNavigate('workouts')}
              className="text-white font-bold hover:underline flex items-center gap-0.5"
            >
              Смотреть
              <ChevronRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Streak & Goal Card */}
        <div className="rounded-2xl p-5 border bg-[#121216] border-white/[0.08] shadow-sm relative overflow-hidden group hover:border-white/[0.18] transition-colors">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">Серия и Цель</span>
            <div className="p-2 rounded-xl bg-amber-500/15 text-amber-400 border border-amber-500/20">
              <Activity className="w-4 h-4" />
            </div>
          </div>
          <div className="flex items-baseline gap-2 mb-1">
            <span className="text-2xl sm:text-3xl font-black text-white">{streak || 1}</span>
            <span className="text-xs font-semibold text-zinc-300">дней подряд 🔥</span>
          </div>
          <div className="text-xs text-zinc-400 font-medium">
            ИМТ: <span className="font-bold text-white">{user.bmi?.toFixed(1) || '22.5'}</span> • Вес: {user.weight} кг → {user.targetWeight} кг
          </div>
          <div className="mt-3 flex items-center gap-2">
            <span className="text-[11px] px-2 py-0.5 rounded-md bg-white/10 text-white border border-white/20 font-semibold">
              {user.goal === 'weight_loss' ? 'Похудение' : user.goal === 'muscle_gain' ? 'Набор массы' : 'Поддержание формы'}
            </span>
            <span className="text-[11px] px-2 py-0.5 rounded-md bg-[#1c1c24] text-zinc-300 border border-white/[0.08] font-semibold">
              {user.daysPerWeek || 3} тр/нед
            </span>
          </div>
        </div>
      </div>

      {/* Main Grid: Today's Workout detail + Daily Schedule (Мой распорядок дня) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Today's Workout detailed card */}
        <div className="lg:col-span-2 rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-5">
          <div className="flex items-center justify-between border-b border-white/[0.08] pb-4">
            <div>
              <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                {todayWorkout?.weekdayName || 'Сегодня'}
              </span>
              <h3 className="text-xl font-bold text-white mt-0.5">
                {todayWorkout?.dayName || 'Программа тренировки'}
              </h3>
              <p className="text-xs sm:text-sm text-zinc-400">
                Фокус: <span className="font-medium text-white">{todayWorkout?.focus || 'Комплексная проработка'}</span>
              </p>
            </div>
            <button
              onClick={() => onNavigate('workouts')}
              className="flex items-center gap-2 px-4 py-2 rounded-xl text-xs sm:text-sm font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-all shadow-md shadow-white/5 active:scale-95"
            >
              <span>К тренировке</span>
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>

          {todayWorkout?.exercises && todayWorkout.exercises.length > 0 ? (
            <div className="space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {todayWorkout.exercises.slice(0, 4).map((ex, idx) => (
                  <div
                    key={ex.id || idx}
                    className="rounded-xl p-3.5 border bg-[#17171d] border-white/[0.06] hover:border-white/[0.15] transition-colors flex items-center justify-between"
                  >
                    <div className="space-y-0.5 truncate pr-2">
                      <span className="text-xs font-bold text-white truncate block">{ex.title}</span>
                      <span className="text-[11px] text-zinc-400 font-medium">{ex.muscle} • {ex.equipment}</span>
                    </div>
                    <div className="text-right flex-shrink-0">
                      <span className="text-xs font-bold text-white bg-white/10 px-2 py-0.5 rounded border border-white/20">
                        {ex.sets} × {ex.reps}
                      </span>
                    </div>
                  </div>
                ))}
              </div>

              {/* Note preview if present */}
              {todayWorkout.notes && (
                <div className="p-3.5 rounded-xl bg-[#17171d] border border-white/[0.08] flex items-start gap-2.5">
                  <FileText className="w-4 h-4 text-zinc-400 flex-shrink-0 mt-0.5" />
                  <div className="text-xs text-zinc-300">
                    <span className="font-bold text-white block mb-0.5">Заметка к тренировке:</span>
                    {todayWorkout.notes}
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="p-8 text-center rounded-xl bg-[#17171d] border border-white/[0.06]">
              <CheckCircle2 className="w-8 h-8 text-zinc-400 mx-auto mb-2 opacity-90" />
              <div className="text-sm font-semibold text-white">Сегодня день восстановления</div>
              <p className="text-xs text-zinc-400 mt-1 max-w-sm mx-auto">
                Отдохните, уделите время прогулке на свежем воздухе или легкой растяжке для регенерации мышц.
              </p>
            </div>
          )}
        </div>

        {/* Bio-rhythm / "Мой распорядок дня" card */}
        <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-4">
          <div className="flex items-center justify-between border-b border-white/[0.08] pb-3">
            <div className="flex items-center gap-2.5">
              <Clock className="w-5 h-5 text-sky-400" />
              <h3 className="text-base font-bold text-white">Мой распорядок дня</h3>
            </div>
            <button
              onClick={() => onNavigate('profile')}
              className="text-xs text-zinc-400 hover:text-white font-medium"
            >
              Настроить
            </button>
          </div>
          <p className="text-xs text-zinc-400 leading-relaxed font-normal">
            Ваш оптимизированный биологический ритм для поддержания высокой энергии и восстановления:
          </p>

          <div className="space-y-3 pt-1">
            <div className="flex items-center justify-between p-3 rounded-xl bg-[#17171d] border border-white/[0.06]">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-white/10 text-white">
                  <Sun className="w-4 h-4 text-white" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">Подъем и стакан воды</div>
                  <div className="text-[11px] text-zinc-400">Запуск метаболизма</div>
                </div>
              </div>
              <span className="text-xs font-bold text-white">07:00</span>
            </div>

            <div className="flex items-center justify-between p-3 rounded-xl bg-[#17171d] border border-white/[0.06]">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-white/10 text-white border border-white/20">
                  <Dumbbell className="w-4 h-4 text-white" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">Окно тренировки</div>
                  <div className="text-[11px] text-zinc-400">Пик физической активности</div>
                </div>
              </div>
              <span className="text-xs font-bold text-white">
                {user.preferredTime === 'morning' ? '07:30' : user.preferredTime === 'day' ? '13:00' : '18:30'}
              </span>
            </div>

            <div className="flex items-center justify-between p-3 rounded-xl bg-[#17171d] border border-white/[0.06]">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-sky-500/15 text-sky-400 border border-sky-500/20">
                  <Moon className="w-4 h-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">Отбой и восстановление</div>
                  <div className="text-[11px] text-zinc-400">8 часов качественного сна</div>
                </div>
              </div>
              <span className="text-xs font-bold text-sky-300">23:00</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
