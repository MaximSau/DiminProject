import React, { useState } from 'react';
import {
  Flame,
  Droplets,
  Dumbbell,
  Clock,
  ShieldAlert,
  ChevronRight,
  Plus,
  Moon,
  Sun,
  Activity,
  CheckCircle2,
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
  const { user, todayWorkout, todayMeals, todayWater, streak, isGentleMode, reminders } = data;
  const { showToast } = useToast();
  const [addingWater, setAddingWater] = useState(false);

  const handleAddWater = async (delta = 330) => {
    try {
      setAddingWater(true);
      await api.addWater(delta);
      showToast(`+${delta} мл воды добавлено`);
      onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка добавления воды', 'error');
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
        <div className="rounded-2xl p-4 sm:p-5 border bg-amber-500/10 border-amber-500/30 text-amber-200 flex items-start gap-4 shadow-lg backdrop-blur-sm">
          <div className="p-2.5 rounded-xl bg-amber-500/20 text-amber-300 flex-shrink-0 mt-0.5">
            <ShieldAlert className="w-5 h-5" />
          </div>
          <div className="space-y-1">
            <h4 className="font-bold text-amber-300 text-sm sm:text-base flex items-center gap-2">
              Активирован щадящий режим: восстановление и мобильность
            </h4>
            <p className="text-xs sm:text-sm text-amber-200/90 leading-relaxed">
              В вашем профиле отмечены медицинские особенности. Программа автоматически адаптирована: исключены ударные осевые нагрузки, прыжки и высокоинтенсивные берпи. Перед началом занятий обязательно проконсультируйтесь с лечащим врачом.
            </p>
          </div>
        </div>
      )}

      {/* Top Welcome & KPI row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Calories Card */}
        <div className="rounded-2xl p-5 border bg-[#111827]/80 border-slate-800 shadow-sm relative overflow-hidden group hover:border-slate-700 transition-colors">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Калории за день</span>
            <div className="p-2 rounded-xl bg-rose-500/15 text-rose-400">
              <Flame className="w-4 h-4" />
            </div>
          </div>
          <div className="flex items-baseline gap-2 mb-2">
            <span className="text-2xl sm:text-3xl font-black text-white">{calCurrent}</span>
            <span className="text-xs font-semibold text-slate-400">/ {calTarget} ккал</span>
          </div>
          <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
            <div
              className="bg-gradient-to-r from-rose-500 to-amber-400 h-full rounded-full transition-all duration-500"
              style={{ width: `${calPercent}%` }}
            />
          </div>
          <div className="mt-2 text-[11px] text-slate-400 flex justify-between font-medium">
            <span>{calPercent}% выполнено</span>
            <span>Осталось {Math.max(0, calTarget - calCurrent)} ккал</span>
          </div>
        </div>

        {/* Water Card (in 330ml glasses as requested) */}
        <div className="rounded-2xl p-5 border bg-[#111827]/80 border-slate-800 shadow-sm relative overflow-hidden group hover:border-slate-700 transition-colors">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Водный баланс</span>
            <div className="p-2 rounded-xl bg-sky-500/15 text-sky-400">
              <Droplets className="w-4 h-4" />
            </div>
          </div>
          <div className="flex items-baseline justify-between mb-2">
            <div className="flex items-baseline gap-2">
              <span className="text-2xl sm:text-3xl font-black text-white">{waterCurrent}</span>
              <span className="text-xs font-semibold text-slate-400">/ {waterTarget} мл</span>
            </div>
            <span className="text-xs font-bold px-2 py-0.5 rounded-md bg-sky-500/10 text-sky-400 border border-sky-500/20">
              {waterGlasses} / {targetGlasses} стак.
            </span>
          </div>
          <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
            <div
              className="bg-gradient-to-r from-sky-500 to-emerald-400 h-full rounded-full transition-all duration-500"
              style={{ width: `${waterPercent}%` }}
            />
          </div>
          <div className="mt-3 flex items-center justify-between">
            <span className="text-[11px] text-slate-400">1 стакан = 330 мл</span>
            <button
              onClick={() => handleAddWater(330)}
              disabled={addingWater}
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-bold bg-sky-500/20 hover:bg-sky-500/30 text-sky-300 border border-sky-500/30 transition-all active:scale-95 disabled:opacity-50"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>+330 мл</span>
            </button>
          </div>
        </div>

        {/* Workout Activity Card */}
        <div className="rounded-2xl p-5 border bg-[#111827]/80 border-slate-800 shadow-sm relative overflow-hidden group hover:border-slate-700 transition-colors">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Тренировка сегодня</span>
            <div className="p-2 rounded-xl bg-emerald-500/15 text-emerald-400">
              <Dumbbell className="w-4 h-4" />
            </div>
          </div>
          <div className="mb-2">
            <div className="text-lg font-bold text-white truncate">
              {todayWorkout?.isRest ? 'День восстановления' : todayWorkout?.dayName || 'Тренировка'}
            </div>
            <div className="text-xs text-slate-400 truncate">
              {todayWorkout?.focus || 'Отдых и растяжка'}
            </div>
          </div>
          <div className="mt-4 flex items-center justify-between text-xs">
            <span className="text-slate-400 flex items-center gap-1.5">
              <Clock className="w-3.5 h-3.5 text-slate-500" />
              ~{todayWorkout?.estimatedMinutes || user.workoutDuration || 45} мин
            </span>
            <button
              onClick={() => onNavigate('workouts')}
              className="text-emerald-400 font-bold hover:underline flex items-center gap-0.5"
            >
              Смотреть
              <ChevronRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Streak & Goal Card */}
        <div className="rounded-2xl p-5 border bg-[#111827]/80 border-slate-800 shadow-sm relative overflow-hidden group hover:border-slate-700 transition-colors">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Серия и Цель</span>
            <div className="p-2 rounded-xl bg-amber-500/15 text-amber-400">
              <Activity className="w-4 h-4" />
            </div>
          </div>
          <div className="flex items-baseline gap-2 mb-1">
            <span className="text-2xl sm:text-3xl font-black text-amber-400">{streak || 1}</span>
            <span className="text-xs font-semibold text-slate-400">дней подряд 🔥</span>
          </div>
          <div className="text-xs text-slate-300 font-medium">
            ИМТ: <span className="font-bold text-white">{user.bmi?.toFixed(1) || '22.5'}</span> • Вес: {user.weight} кг → {user.targetWeight} кг
          </div>
          <div className="mt-3 flex items-center gap-2">
            <span className="text-[11px] px-2 py-0.5 rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-semibold">
              {user.goal === 'weight_loss' ? 'Похудение' : user.goal === 'muscle_gain' ? 'Набор массы' : 'Поддержание формы'}
            </span>
            <span className="text-[11px] px-2 py-0.5 rounded-md bg-slate-800 text-slate-300 font-semibold">
              {user.daysPerWeek || 3} тр/нед
            </span>
          </div>
        </div>
      </div>

      {/* Main Grid: Today's Workout detail + Daily Schedule (Мой распорядок дня) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Today's Workout detailed card */}
        <div className="lg:col-span-2 rounded-2xl p-6 border bg-[#111827]/80 border-slate-800 shadow-sm space-y-5">
          <div className="flex items-center justify-between border-b border-slate-800/80 pb-4">
            <div>
              <span className="text-xs font-bold uppercase tracking-wider text-emerald-400">
                {todayWorkout?.weekdayName || 'Сегодня'}
              </span>
              <h3 className="text-xl font-bold text-white mt-0.5">
                {todayWorkout?.dayName || 'Программа тренировки'}
              </h3>
              <p className="text-xs sm:text-sm text-slate-400">
                Фокус: {todayWorkout?.focus || 'Комплексная проработка'}
              </p>
            </div>
            <button
              onClick={() => onNavigate('workouts')}
              className="flex items-center gap-2 px-4 py-2 rounded-xl text-xs sm:text-sm font-bold bg-emerald-500 hover:bg-emerald-400 text-slate-950 transition-all shadow-md shadow-emerald-500/20"
            >
              <span>К тренировке</span>
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>

          {todayWorkout?.exercises && todayWorkout.exercises.length > 0 ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {todayWorkout.exercises.slice(0, 4).map((ex, idx) => (
                <div
                  key={ex.id || idx}
                  className="rounded-xl p-3.5 border bg-slate-900/60 border-slate-800/80 hover:border-slate-700 transition-colors flex items-center justify-between"
                >
                  <div className="space-y-0.5 truncate pr-2">
                    <span className="text-xs font-bold text-white truncate block">{ex.title}</span>
                    <span className="text-[11px] text-slate-400">{ex.muscle} • {ex.equipment}</span>
                  </div>
                  <div className="text-right flex-shrink-0">
                    <span className="text-xs font-bold text-emerald-400">{ex.sets} × {ex.reps}</span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="p-8 text-center rounded-xl bg-slate-900/40 border border-slate-800/60">
              <CheckCircle2 className="w-8 h-8 text-emerald-400 mx-auto mb-2 opacity-80" />
              <div className="text-sm font-semibold text-white">Сегодня день восстановления</div>
              <p className="text-xs text-slate-400 mt-1 max-w-sm mx-auto">
                Отдохните, уделите время прогулке на свежем воздухе или легкой растяжке для регенерации мышц.
              </p>
            </div>
          )}
        </div>

        {/* Bio-rhythm / "Мой распорядок дня" card */}
        <div className="rounded-2xl p-6 border bg-[#111827]/80 border-slate-800 shadow-sm space-y-4">
          <div className="flex items-center justify-between border-b border-slate-800/80 pb-3">
            <div className="flex items-center gap-2.5">
              <Clock className="w-5 h-5 text-sky-400" />
              <h3 className="text-base font-bold text-white">Мой распорядок дня</h3>
            </div>
            <button
              onClick={() => onNavigate('profile')}
              className="text-xs text-slate-400 hover:text-slate-200"
            >
              Настроить
            </button>
          </div>
          <p className="text-xs text-slate-400 leading-relaxed">
            Ваш оптимизированный биологический ритм для поддержания высокой энергии и восстановления:
          </p>

          <div className="space-y-3 pt-1">
            <div className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800/80">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-amber-500/15 text-amber-400">
                  <Sun className="w-4 h-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">Подъем и стакан воды</div>
                  <div className="text-[11px] text-slate-400">Запуск метаболизма</div>
                </div>
              </div>
              <span className="text-xs font-bold text-amber-300">07:00</span>
            </div>

            <div className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800/80">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-emerald-500/15 text-emerald-400">
                  <Dumbbell className="w-4 h-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">Окно тренировки</div>
                  <div className="text-[11px] text-slate-400">Пик физической активности</div>
                </div>
              </div>
              <span className="text-xs font-bold text-emerald-400">
                {user.preferredTime === 'morning' ? '07:30' : user.preferredTime === 'day' ? '13:00' : '18:30'}
              </span>
            </div>

            <div className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800/80">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-sky-500/15 text-sky-400">
                  <Moon className="w-4 h-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">Отбой и восстановление</div>
                  <div className="text-[11px] text-slate-400">8 часов качественного сна</div>
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
