import React, { useState, useEffect } from 'react';
import {
  Printer,
  ArrowLeft,
  Dumbbell,
  Clock,
  Activity,
} from 'lucide-react';
import { WorkoutDay, DayMeals, UserProfile } from '../types';
import { api } from '../api/client';
import { useToast } from '../context/ToastContext';

interface PlanExportPageProps {
  onBack: () => void;
}

export const PlanExportPage: React.FC<PlanExportPageProps> = ({ onBack }) => {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);
  const [user, setUser] = useState<UserProfile | null>(null);
  const [days, setDays] = useState<WorkoutDay[]>([]);
  const [meals, setMeals] = useState<DayMeals[]>([]);

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        const res = await api.getPlanExport();
        setUser(res.user);
        setDays(res.days || []);
        setMeals(res.meals || []);
      } catch (e: any) {
        showToast(e.message || 'Ошибка экспорта программы', 'error');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  return (
    <div className="space-y-6 max-w-5xl mx-auto">
      {/* Action buttons (hidden on print) */}
      <div className="flex items-center justify-between print:hidden">
        <button
          onClick={onBack}
          className="flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-bold text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Назад к панели</span>
        </button>

        <button
          onClick={() => window.print()}
          className="flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-bold bg-white hover:bg-zinc-200 text-zinc-950 shadow-md transition-all active:scale-95"
        >
          <Printer className="w-4 h-4" />
          <span>Распечатать / Сохранить в PDF</span>
        </button>
      </div>

      {loading ? (
        <div className="py-20 text-center text-zinc-400">Формирование документа программы...</div>
      ) : user ? (
        <div className="rounded-2xl p-8 border bg-[#121216] border-white/[0.08] space-y-8 print:border-none print:bg-white print:text-black">
          {/* Header */}
          <div className="border-b border-white/[0.08] pb-6 print:border-slate-300">
            <div className="flex items-center gap-2 text-zinc-400 print:text-zinc-700 text-xs font-bold uppercase tracking-wider mb-1">
              <Activity className="w-4 h-4 text-white print:text-black" />
              PulseFit Athletic Program
            </div>
            <h1 className="text-3xl font-black text-white print:text-black">
              Персональный план: {user.name || 'Атлет'}
            </h1>
            <div className="flex flex-wrap items-center gap-4 text-xs text-zinc-400 print:text-slate-600 mt-2">
              <span>Цель: <strong className="text-white print:text-black">{user.goal}</strong></span>
              <span>• Уровень: <strong className="text-white print:text-black">{user.level}</strong></span>
              <span>• Калораж: <strong className="text-white print:text-black">{user.calorieTarget} ккал</strong></span>
              <span>• Вода: <strong className="text-white print:text-black">{user.waterTarget} мл</strong></span>
            </div>
          </div>

          {/* 7-Day Workouts */}
          <div className="space-y-4">
            <h2 className="text-xl font-bold text-white print:text-black flex items-center gap-2">
              <Dumbbell className="w-5 h-5 text-white print:text-black" />
              Расписание тренировок (7 дней)
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {days.map(d => (
                <div
                  key={d.dayIndex}
                  className="rounded-xl p-4 border bg-[#17171d] border-white/[0.06] print:bg-slate-50 print:border-slate-300 space-y-2"
                >
                  <div className="flex items-center justify-between border-b border-white/[0.08]/60 print:border-slate-200 pb-2">
                    <span className="text-xs font-bold uppercase tracking-wider text-zinc-400 print:text-slate-700">
                      {d.weekdayName} • День {d.dayIndex}
                    </span>
                    <span className="text-xs text-zinc-400 print:text-slate-600">
                      {d.isRest ? 'Отдых' : `~${d.estimatedMinutes} мин`}
                    </span>
                  </div>
                  <div className="text-sm font-bold text-white print:text-black">
                    {d.isRest ? 'День восстановления' : d.dayName}
                  </div>
                  <div className="text-xs text-zinc-400 print:text-slate-600">
                    Фокус: {d.focus}
                  </div>

                  {!d.isRest && d.exercises && (
                    <div className="pt-2 space-y-1">
                      {d.exercises.map((ex, i) => (
                        <div key={i} className="text-xs text-zinc-300 print:text-slate-700 flex justify-between">
                          <span>{i + 1}. {ex.title}</span>
                          <span className="font-semibold text-white print:text-black">{ex.sets} × {ex.reps}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
};
