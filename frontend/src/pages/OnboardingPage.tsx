import React, { useState } from 'react';
import {
  Activity,
  Sparkles,
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

  // Fields (stored as string to avoid sticky zeros on backspace)
  const [name, setName] = useState('');
  const [gender, setGender] = useState('male');
  const [age, setAge] = useState('25');
  const [height, setHeight] = useState('178');
  const [weight, setWeight] = useState('78');
  const [targetWeight, setTargetWeight] = useState('72');
  const [goal, setGoal] = useState('weight_loss');
  const [level, setLevel] = useState('beginner');
  const [equipment, setEquipment] = useState('bodyweight');
  const [daysPerWeek, setDaysPerWeek] = useState(3);
  const [workoutDuration, setWorkoutDuration] = useState(45);
  const [preferredTime, setPreferredTime] = useState('evening');
  const [isVegan, setIsVegan] = useState(false);
  const [healthIssues, setHealthIssues] = useState<string[]>([]);

  const cleanNumber = (val: string) => {
    if (val === '') return '';
    if (val.startsWith('0') && val.length > 1 && !val.startsWith('0.')) {
      return val.replace(/^0+/, '');
    }
    return val;
  };

  const handleToggleHealth = (val: string) => {
    setHealthIssues(prev =>
      prev.includes(val) ? prev.filter(x => x !== val) : [...prev, val]
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const parsedAge = parseInt(age, 10);
    const parsedHeight = parseFloat(height);
    const parsedWeight = parseFloat(weight);
    const parsedTarget = parseFloat(targetWeight);

    if (isNaN(parsedAge) || parsedAge < 10 || parsedAge > 120) {
      showToast('Укажите корректный возраст (от 10 до 120 лет)', 'error');
      return;
    }
    if (isNaN(parsedHeight) || parsedHeight < 100 || parsedHeight > 250) {
      showToast('Укажите корректный рост (от 100 до 250 см)', 'error');
      return;
    }
    if (isNaN(parsedWeight) || parsedWeight < 30 || parsedWeight > 300) {
      showToast('Укажите корректный вес (от 30 до 300 кг)', 'error');
      return;
    }
    if (isNaN(parsedTarget) || parsedTarget < 30 || parsedTarget > 300) {
      showToast('Укажите корректный целевой вес (от 30 до 300 кг)', 'error');
      return;
    }

    try {
      setSubmitting(true);
      await api.submitOnboarding({
        name: name.trim() || undefined,
        gender,
        age: parsedAge,
        height: parsedHeight,
        weight: parsedWeight,
        targetWeight: parsedTarget,
        goal,
        level,
        equipment,
        daysPerWeek,
        workoutDuration,
        preferredTime,
        isVegan,
        healthIssues,
      });

      showToast('Персональный план успешно сгенерирован!');
      onCompleted();
    } catch (err: any) {
      showToast(err.message || 'Ошибка сохранения плана', 'error');
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
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/10 border border-white/20 text-white text-xs font-bold uppercase tracking-wider mb-2">
            <Activity className="w-4 h-4 animate-pulse text-white" />
            Интеллектуальный тренер PulseFit
          </div>
          <h1 className="text-3xl sm:text-4xl font-black text-white tracking-tight">
            Создайте вашу персональную программу
          </h1>
          <p className="text-xs sm:text-sm text-zinc-400 max-w-lg mx-auto">
            Ответьте на несколько вопросов, чтобы система рассчитала ваш дефицит/профицит, подобрала сплит и составила план питания.
          </p>

          {/* Quick Demo button */}
          <div className="pt-2">
            <button
              type="button"
              onClick={handleDemo}
              disabled={seedingDemo}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold uppercase tracking-wider bg-white/10 hover:bg-white/20 text-white border border-white/20 transition-all hover:scale-105 active:scale-95 disabled:opacity-50"
            >
              <Sparkles className="w-3.5 h-3.5 text-zinc-300" />
              <span>{seedingDemo ? 'Загрузка сценария...' : 'Запустить демонстрационный сценарий (Демо)'}</span>
            </button>
          </div>
        </div>

        {/* Onboarding Form */}
        <form onSubmit={handleSubmit} className="rounded-2xl p-6 sm:p-8 border bg-[#121216] border-white/[0.08] shadow-2xl space-y-6 backdrop-blur-md">
          {/* Name input */}
          <div className="space-y-1.5">
            <label className="block text-xs font-bold uppercase tracking-wider text-zinc-400">
              Ваше имя / Фамилия
            </label>
            <input
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="Фамилия Имя"
              className="w-full px-4 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white placeholder-zinc-500 text-sm focus:border-white focus:outline-none"
            />
          </div>

          {/* Goal selection */}
          <div className="space-y-2">
            <label className="block text-xs font-bold uppercase tracking-wider text-zinc-400">
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
                      ? 'bg-white text-zinc-950 shadow-md font-black'
                      : 'bg-[#17171d] text-zinc-300 hover:text-white border border-white/[0.08]'
                  }`}
                >
                  {g.label}
                </button>
              ))}
            </div>
          </div>

          {/* Body stats without sticky zeros */}
          <div className="space-y-2">
            <label className="block text-xs font-bold uppercase tracking-wider text-zinc-400">
              Параметры тела
            </label>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
              <div>
                <span className="block text-[11px] text-zinc-400 mb-1">Возраст</span>
                <input
                  type="number"
                  value={age}
                  onChange={e => setAge(cleanNumber(e.target.value))}
                  placeholder="0"
                  className="w-full px-3 py-2 rounded-xl bg-[#17171d] border border-white/[0.12] text-white font-bold text-sm focus:border-white focus:outline-none placeholder-zinc-500"
                  required
                />
              </div>
              <div>
                <span className="block text-[11px] text-zinc-400 mb-1">Рост (см)</span>
                <input
                  type="number"
                  value={height}
                  onChange={e => setHeight(cleanNumber(e.target.value))}
                  placeholder="0"
                  className="w-full px-3 py-2 rounded-xl bg-[#17171d] border border-white/[0.12] text-white font-bold text-sm focus:border-white focus:outline-none placeholder-zinc-500"
                  required
                />
              </div>
              <div>
                <span className="block text-[11px] text-zinc-400 mb-1">Вес (кг)</span>
                <input
                  type="number"
                  step="0.1"
                  value={weight}
                  onChange={e => setWeight(cleanNumber(e.target.value))}
                  placeholder="0"
                  className="w-full px-3 py-2 rounded-xl bg-[#17171d] border border-white/[0.12] text-white font-bold text-sm focus:border-white focus:outline-none placeholder-zinc-500"
                  required
                />
              </div>
              <div>
                <span className="block text-[11px] text-zinc-400 mb-1">Цель (кг)</span>
                <input
                  type="number"
                  step="0.1"
                  value={targetWeight}
                  onChange={e => setTargetWeight(cleanNumber(e.target.value))}
                  placeholder="0"
                  className="w-full px-3 py-2 rounded-xl bg-[#17171d] border border-white/[0.12] text-white font-bold text-sm focus:border-white focus:outline-none placeholder-zinc-500"
                  required
                />
              </div>
            </div>
          </div>

          {/* Schedule & Days */}
          <div className="space-y-3">
            <div>
              <div className="flex justify-between items-center mb-1.5">
                <label className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                  Частота тренировок в неделю
                </label>
                <span className="text-xs font-bold text-white">{daysPerWeek} дней</span>
              </div>
              <div className="flex gap-1.5">
                {[1, 2, 3, 4, 5, 6, 7].map(d => (
                  <button
                    type="button"
                    key={d}
                    onClick={() => setDaysPerWeek(d)}
                    className={`flex-1 py-2 rounded-xl text-xs font-bold transition-all ${
                      daysPerWeek === d
                        ? 'bg-white text-zinc-950 shadow-sm'
                        : 'bg-[#17171d] text-zinc-400 hover:text-white border border-white/[0.08]'
                    }`}
                  >
                    {d}
                  </button>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
              <div>
                <label className="block text-xs text-zinc-400 mb-1 font-medium">Инвентарь</label>
                <select
                  value={equipment}
                  onChange={e => setEquipment(e.target.value)}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-xs font-semibold focus:border-white focus:outline-none"
                >
                  <option value="bodyweight">Свой вес (дома/на турниках)</option>
                  <option value="dumbbells">Гантели / эспандеры</option>
                  <option value="gym">Тренажерный зал</option>
                </select>
              </div>

              <div>
                <label className="block text-xs text-zinc-400 mb-1 font-medium">Время для тренировок</label>
                <select
                  value={preferredTime}
                  onChange={e => setPreferredTime(e.target.value)}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-xs font-semibold focus:border-white focus:outline-none"
                >
                  <option value="morning">Утро (07:00 – 09:00)</option>
                  <option value="day">День (12:00 – 14:00)</option>
                  <option value="evening">Вечер (18:00 – 21:00)</option>
                </select>
              </div>
            </div>

            <div>
              <div className="flex justify-between items-center mb-1.5">
                <label className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                  Длительность тренировки
                </label>
                <span className="text-xs font-bold text-white">{workoutDuration} минут</span>
              </div>
              <input
                type="range"
                min="30"
                max="90"
                step="5"
                value={workoutDuration}
                onChange={e => setWorkoutDuration(parseInt(e.target.value, 10))}
                className="w-full accent-white h-1.5 bg-[#17171d] rounded-lg cursor-pointer"
              />
              <div className="flex justify-between text-[11px] text-zinc-400 mt-1">
                <span>30 мин</span>
                <span>45 мин</span>
                <span>60 мин</span>
                <span>90 мин</span>
              </div>
            </div>
          </div>

          {/* Vegan diet toggle */}
          <div className="p-4 rounded-xl bg-[#17171d] border border-white/[0.08] flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                <Leaf className="w-5 h-5 text-emerald-400" />
              </div>
              <div>
                <span className="text-sm font-bold text-white block">Веганский рацион</span>
                <span className="text-xs text-zinc-400 font-normal">
                  Только 100% растительная пища без мяса, птицы, рыбы, яиц и лактозы
                </span>
              </div>
            </div>
            <button
              type="button"
              onClick={() => setIsVegan(!isVegan)}
              className={`w-12 h-6 rounded-full transition-colors relative p-0.5 border ${
                isVegan ? 'bg-white border-white' : 'bg-zinc-800 border-zinc-700'
              }`}
            >
              <div
                className={`w-5 h-5 rounded-full transition-transform ${
                  isVegan ? 'translate-x-6 bg-zinc-950' : 'translate-x-0 bg-zinc-400'
                }`}
              />
            </button>
          </div>

          {/* Health issues checkboxes */}
          <div className="space-y-2">
            <div className="flex items-center gap-1.5">
              <ShieldAlert className="w-4 h-4 text-amber-400" />
              <label className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                Медицинские ограничения и здоровье
              </label>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {[
                { id: 'joints', label: 'Боли в коленях / суставах' },
                { id: 'back', label: 'Проблемы со спиной / позвоночником' },
                { id: 'hypertension', label: 'Гипертония / давление' },
                { id: 'injuries', label: 'Недавние травмы' },
              ].map(h => (
                <button
                  type="button"
                  key={h.id}
                  onClick={() => handleToggleHealth(h.id)}
                  className={`p-3 rounded-xl text-xs font-semibold text-left border transition-all flex items-center justify-between ${
                    healthIssues.includes(h.id)
                      ? 'bg-amber-500/15 border-amber-500/40 text-amber-300'
                      : 'bg-[#17171d] border-white/[0.08] text-zinc-400 hover:text-white'
                  }`}
                >
                  <span>{h.label}</span>
                  {healthIssues.includes(h.id) && (
                    <CheckCircle2 className="w-4 h-4 text-amber-400" />
                  )}
                </button>
              ))}
            </div>
          </div>

          {/* Submit CTA */}
          <button
            type="submit"
            disabled={submitting}
            className="w-full py-4 rounded-xl text-sm font-black bg-white hover:bg-zinc-200 text-zinc-950 transition-all shadow-xl shadow-white/10 flex items-center justify-center gap-2 active:scale-95 disabled:opacity-50"
          >
            <span>{submitting ? 'Генерация персональной программы...' : 'Сгенерировать программу тренировок'}</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </form>
      </div>
    </div>
  );
};
