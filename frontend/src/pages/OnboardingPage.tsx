import React, { useState } from 'react';
import {
  Activity,
  Sparkles,
  Flame,
  ArrowRight,
  ShieldAlert,
  Leaf,
  CheckCircle2,
} from 'lucide-react';
import { api } from '../api/client';
import { useToast } from '../context/ToastContext';

interface OnboardingPageProps {
  onCompleted: () => void;
}

export const OnboardingPage: React.FC<OnboardingPageProps> = ({ onCompleted }) => {
  const { showToast } = useToast();
  const [submitting, setSubmitting] = useState(false);
  const [seedingDemo, setSeedingDemo] = useState(false);

  // Fields
  const [name, setName] = useState('');
  const [gender, setGender] = useState('male');
  const [age, setAge] = useState(25);
  const [height, setHeight] = useState(178);
  const [weight, setWeight] = useState(78);
  const [targetWeight, setTargetWeight] = useState(72);
  const [goal, setGoal] = useState('weight_loss');
  const [level, setLevel] = useState('beginner');
  const [equipment, setEquipment] = useState('bodyweight');
  const [daysPerWeek, setDaysPerWeek] = useState(3);
  const [workoutDuration, setWorkoutDuration] = useState(45);
  const [preferredTime, setPreferredTime] = useState('evening');
  const [isVegan, setIsVegan] = useState(false);
  const [healthIssues, setHealthIssues] = useState<string[]>([]);

  const handleToggleHealth = (val: string) => {
    setHealthIssues(prev =>
      prev.includes(val) ? prev.filter(x => x !== val) : [...prev, val]
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setSubmitting(true);
      await api.submitOnboarding({
        name: name || 'Атлет',
        gender,
        age: Number(age),
        height: Number(height),
        weight: Number(weight),
        targetWeight: Number(targetWeight),
        goal,
        level,
        equipment,
        daysPerWeek: Number(daysPerWeek),
        workoutDuration: Number(workoutDuration),
        preferredTime,
        isVegan,
        healthIssues,
      });
      showToast('Ваша персональная программа сформирована!');
      onCompleted();
    } catch (e: any) {
      showToast(e.message || 'Ошибка генерации программы', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDemo = async () => {
    try {
      setSeedingDemo(true);
      await api.seedDemo();
      showToast('Демонстрационный профиль загружен!');
      onCompleted();
    } catch (e: any) {
      showToast(e.message || 'Ошибка активации демо', 'error');
    } finally {
      setSeedingDemo(false);
    }
  };

  return (
    <div className="min-h-screen py-8 px-4 sm:px-6 lg:px-8 flex items-center justify-center bg-[#08080a]">
      <div className="max-w-2xl w-full space-y-6">
        {/* Brand header */}
        <div className="text-center space-y-2">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-bold uppercase tracking-wider mb-2">
            <Activity className="w-4 h-4 animate-pulse" />
            Интеллектуальный тренер PulseFit
          </div>
          <h1 className="text-3xl sm:text-4xl font-black text-white tracking-tight">
            Создайте вашу персональную программу
          </h1>
          <p className="text-xs sm:text-sm text-slate-400 max-w-lg mx-auto">
            Ответьте на несколько вопросов, чтобы система рассчитала ваш дефицит/профицит, подобрала сплит и составила план питания.
          </p>

          {/* Quick Demo button */}
          <div className="pt-2">
            <button
              type="button"
              onClick={handleDemo}
              disabled={seedingDemo}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold uppercase tracking-wider bg-amber-500/20 hover:bg-amber-500/30 text-amber-300 border border-amber-500/40 transition-all hover:scale-105 active:scale-95 disabled:opacity-50"
            >
              <Sparkles className="w-3.5 h-3.5 text-amber-400" />
              <span>{seedingDemo ? 'Загрузка сценария...' : 'Запустить демонстрационный сценарий (Демо)'}</span>
            </button>
          </div>
        </div>

        {/* Onboarding Form */}
        <form onSubmit={handleSubmit} className="rounded-2xl p-6 sm:p-8 border bg-[#121216] border-white/[0.08] shadow-2xl space-y-6 backdrop-blur-md">
          {/* Goal selection */}
          <div className="space-y-2">
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-400">
              Главная цель
            </label>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
              {[
                { id: 'weight_loss', label: 'Похудение' },
                { id: 'muscle_gain', label: 'Набор массы' },
                { id: 'recomposition', label: 'Рельеф' },
                { id: 'endurance', label: 'Тонус' },
              ].map(g => (
                <button
                  type="button"
                  key={g.id}
                  onClick={() => setGoal(g.id)}
                  className={`p-3 rounded-xl text-xs font-bold transition-all text-center ${
                    goal === g.id
                      ? 'bg-emerald-500 text-slate-950 shadow-md shadow-emerald-500/20'
                      : 'bg-slate-900 text-slate-300 hover:text-white border border-white/[0.08]'
                  }`}
                >
                  {g.label}
                </button>
              ))}
            </div>
          </div>

          {/* Body stats */}
          <div className="space-y-2">
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-400">
              Параметры тела
            </label>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
              <div>
                <span className="block text-[11px] text-slate-400 mb-1">Возраст</span>
                <input
                  type="number"
                  min="14"
                  max="90"
                  value={age}
                  onChange={e => setAge(Number(e.target.value))}
                  className="w-full px-3 py-2 rounded-xl bg-[#17171d] border border-white/[0.12] text-white font-bold text-sm focus:border-emerald-500 focus:outline-none"
                  required
                />
              </div>
              <div>
                <span className="block text-[11px] text-slate-400 mb-1">Рост (см)</span>
                <input
                  type="number"
                  min="120"
                  max="230"
                  value={height}
                  onChange={e => setHeight(Number(e.target.value))}
                  className="w-full px-3 py-2 rounded-xl bg-[#17171d] border border-white/[0.12] text-white font-bold text-sm focus:border-emerald-500 focus:outline-none"
                  required
                />
              </div>
              <div>
                <span className="block text-[11px] text-slate-400 mb-1">Вес (кг)</span>
                <input
                  type="number"
                  step="0.1"
                  min="35"
                  max="250"
                  value={weight}
                  onChange={e => setWeight(Number(e.target.value))}
                  className="w-full px-3 py-2 rounded-xl bg-[#17171d] border border-white/[0.12] text-white font-bold text-sm focus:border-emerald-500 focus:outline-none"
                  required
                />
              </div>
              <div>
                <span className="block text-[11px] text-slate-400 mb-1">Цель (кг)</span>
                <input
                  type="number"
                  step="0.1"
                  min="35"
                  max="250"
                  value={targetWeight}
                  onChange={e => setTargetWeight(Number(e.target.value))}
                  className="w-full px-3 py-2 rounded-xl bg-[#17171d] border border-white/[0.12] text-white font-bold text-sm focus:border-emerald-500 focus:outline-none"
                  required
                />
              </div>
            </div>
          </div>

          {/* Schedule & Days */}
          <div className="space-y-3">
            <div>
              <div className="flex justify-between items-center mb-1.5">
                <label className="text-xs font-bold uppercase tracking-wider text-slate-400">
                  Частота тренировок в неделю
                </label>
                <span className="text-xs font-bold text-emerald-400">{daysPerWeek} дней</span>
              </div>
              <div className="flex gap-1.5">
                {[1, 2, 3, 4, 5, 6, 7].map(d => (
                  <button
                    type="button"
                    key={d}
                    onClick={() => setDaysPerWeek(d)}
                    className={`flex-1 py-2 rounded-xl text-xs font-bold transition-all ${
                      daysPerWeek === d
                        ? 'bg-emerald-500 text-slate-950 shadow-sm'
                        : 'bg-slate-900 text-slate-400 hover:text-white border border-white/[0.08]'
                    }`}
                  >
                    {d}
                  </button>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
              <div>
                <label className="block text-xs text-slate-400 mb-1 font-medium">Инвентарь</label>
                <select
                  value={equipment}
                  onChange={e => setEquipment(e.target.value)}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-xs font-semibold focus:border-emerald-500 focus:outline-none"
                >
                  <option value="bodyweight">Свой вес (дома/на турниках)</option>
                  <option value="dumbbells">Гантели / эспандеры</option>
                  <option value="gym">Тренажерный зал</option>
                </select>
              </div>

              <div>
                <label className="block text-xs text-slate-400 mb-1 font-medium">Время для тренировок</label>
                <select
                  value={preferredTime}
                  onChange={e => setPreferredTime(e.target.value)}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-xs font-semibold focus:border-emerald-500 focus:outline-none"
                >
                  <option value="morning">Утро (07:00 – 09:00)</option>
                  <option value="day">День (12:00 – 15:00)</option>
                  <option value="evening">Вечер (18:00 – 21:00)</option>
                </select>
              </div>
            </div>
          </div>

          {/* Vegan diet toggle */}
          <label className="flex items-center gap-3 p-3.5 rounded-xl bg-[#17171d] border border-white/[0.08] cursor-pointer hover:border-slate-700">
            <input
              type="checkbox"
              checked={isVegan}
              onChange={e => setIsVegan(e.target.checked)}
              className="w-4 h-4 rounded text-emerald-500 bg-slate-800 border-slate-700"
            />
            <div>
              <span className="text-xs font-bold text-white flex items-center gap-1.5">
                <Leaf className="w-3.5 h-3.5 text-emerald-400" />
                Веганский рацион
              </span>
              <span className="text-[11px] text-slate-400 block">
                Исключить продукты животного происхождения из меню.
              </span>
            </div>
          </label>

          {/* Health issues */}
          <div>
            <label className="block text-xs text-slate-400 mb-1.5 font-medium">
              Ограничения по здоровью (включают щадящий режим восстановления)
            </label>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
              {[
                { id: 'joints', label: 'Суставы/колени' },
                { id: 'spine', label: 'Спина' },
                { id: 'cardio', label: 'Сердце/давление' },
              ].map(h => (
                <button
                  type="button"
                  key={h.id}
                  onClick={() => handleToggleHealth(h.id)}
                  className={`p-2.5 rounded-xl text-left border text-xs font-bold transition-all ${
                    healthIssues.includes(h.id)
                      ? 'bg-amber-500/15 border-amber-500/40 text-amber-300'
                      : 'bg-slate-900/60 border-white/[0.08] text-slate-400'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span>{h.label}</span>
                    {healthIssues.includes(h.id) && <CheckCircle2 className="w-3.5 h-3.5 text-amber-400" />}
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Submit */}
          <button
            type="submit"
            disabled={submitting}
            className="w-full flex items-center justify-center gap-2 py-3.5 rounded-xl text-sm font-bold bg-emerald-500 hover:bg-emerald-400 text-slate-950 transition-all shadow-lg shadow-emerald-500/25 active:scale-95 disabled:opacity-50"
          >
            <span>{submitting ? 'Создание программы...' : 'Сформировать программу тренировок'}</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </form>
      </div>
    </div>
  );
};
