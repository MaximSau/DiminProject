import React, { useState, useEffect } from 'react';
import {
  User,
  Settings,
  Dumbbell,
  Clock,
  ShieldAlert,
  Leaf,
  CheckCircle2,
  Trash2,
  Save,
} from 'lucide-react';
import { UserProfile } from '../types';
import { api } from '../api/client';
import { useToast } from '../context/ToastContext';

interface ProfilePageProps {
  onRefresh?: () => void;
  onNavigate?: (tab: string) => void;
}

export const ProfilePage: React.FC<ProfilePageProps> = ({ onRefresh, onNavigate }) => {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [resetting, setResetting] = useState(false);

  // Form fields (strings to avoid sticky leading zero on backspace)
  const [name, setName] = useState('');
  const [gender, setGender] = useState('male');
  const [age, setAge] = useState('25');
  const [height, setHeight] = useState('175');
  const [weight, setWeight] = useState('75');
  const [targetWeight, setTargetWeight] = useState('70');
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

  const fetchProfile = async () => {
    try {
      setLoading(true);
      const res = await api.getProfile();
      const u = res.user;
      if (u) {
        setName(u.name || '');
        setGender(u.gender || 'male');
        setAge(String(u.age || 25));
        setHeight(String(u.height || 175));
        setWeight(String(u.weight || 75));
        setTargetWeight(String(u.targetWeight || 70));
        setGoal(u.goal || 'weight_loss');
        setLevel(u.level || 'beginner');
        setEquipment(u.equipment || 'bodyweight');
        setDaysPerWeek(u.daysPerWeek || 3);
        setWorkoutDuration(u.workoutDuration || 45);
        setPreferredTime(u.preferredTime || 'evening');
        setIsVegan(!!u.isVegan);
        setHealthIssues(u.healthIssues || []);
      }
    } catch (e: any) {
      showToast(e.message || 'Ошибка загрузки профиля', 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  const handleToggleHealth = (val: string) => {
    setHealthIssues(prev =>
      prev.includes(val) ? prev.filter(x => x !== val) : [...prev, val]
    );
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();

    const numAge = Number(age);
    const numHeight = Number(height);
    const numWeight = parseFloat(weight);
    const numTarget = parseFloat(targetWeight);

    if (!numAge || numAge < 14 || numAge > 100) {
      showToast('Пожалуйста, укажите корректный возраст (от 14 до 100 лет)', 'error');
      return;
    }
    if (!numHeight || numHeight < 100 || numHeight > 250) {
      showToast('Пожалуйста, укажите корректный рост (от 100 до 250 см)', 'error');
      return;
    }
    if (!numWeight || numWeight < 30 || numWeight > 300) {
      showToast('Пожалуйста, укажите корректный вес (от 30 до 300 кг)', 'error');
      return;
    }
    if (!numTarget || numTarget < 30 || numTarget > 300) {
      showToast('Пожалуйста, укажите целевой вес (от 30 до 300 кг)', 'error');
      return;
    }

    try {
      setSaving(true);
      await api.updateProfile({
        name,
        gender,
        age: numAge,
        height: numHeight,
        weight: numWeight,
        targetWeight: numTarget,
        goal,
        level,
        equipment,
        daysPerWeek: Number(daysPerWeek),
        workoutDuration: Number(workoutDuration),
        preferredTime,
        isVegan,
        healthIssues,
      });
      showToast('Профиль сохранён и программа тренировок обновлена!');
      if (onRefresh) onRefresh();
    } catch (e: any) {
      showToast(e.message || 'Ошибка сохранения профиля', 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleReset = async () => {
    if (!window.confirm('Вы уверены, что хотите сбросить все персональные данные и начать заново?')) {
      return;
    }
    try {
      setResetting(true);
      await api.resetDemo();
      showToast('Данные сброшены. Перенаправление на создание профиля...');
    } catch (e: any) {
      showToast(e.message || 'Ошибка сброса', 'error');
    } finally {
      setResetting(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Header */}
      <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-2xl bg-white/10 text-white border border-white/20">
            <User className="w-6 h-6 text-white" />
          </div>
          <div>
            <h2 className="text-2xl font-black text-white">Мой профиль и параметры</h2>
            <p className="text-xs sm:text-sm text-zinc-400">
              Единый центр управления программой тренировок, питанием и целями
            </p>
          </div>
        </div>
      </div>

      {loading ? (
        <div className="py-20 text-center text-zinc-400">Загрузка профиля...</div>
      ) : (
        <form onSubmit={handleSave} className="space-y-6">
          {/* Section 1: Personal Measurements */}
          <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-4">
            <h3 className="text-base font-bold text-white border-b border-white/[0.08] pb-3 flex items-center gap-2">
              <Settings className="w-4 h-4 text-white" />
              Личные параметры
            </h3>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">Имя</label>
                <input
                  type="text"
                  value={name}
                  onChange={e => setName(e.target.value)}
                  placeholder="Ваше имя"
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">Пол</label>
                <select
                  value={gender}
                  onChange={e => setGender(e.target.value)}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none"
                >
                  <option value="male">Мужской</option>
                  <option value="female">Женский</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">Возраст (лет)</label>
                <input
                  type="number"
                  value={age}
                  onChange={e => setAge(cleanNumber(e.target.value))}
                  placeholder="25"
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">Рост (см)</label>
                <input
                  type="number"
                  value={height}
                  onChange={e => setHeight(cleanNumber(e.target.value))}
                  placeholder="175"
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">Текущий вес (кг)</label>
                <input
                  type="number"
                  step="0.1"
                  value={weight}
                  onChange={e => setWeight(cleanNumber(e.target.value))}
                  placeholder="75"
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">Целевой вес (кг)</label>
                <input
                  type="number"
                  step="0.1"
                  value={targetWeight}
                  onChange={e => setTargetWeight(cleanNumber(e.target.value))}
                  placeholder="70"
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none"
                />
              </div>
            </div>
          </div>

          {/* Section 2: Goals & Schedule */}
          <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-5">
            <h3 className="text-base font-bold text-white border-b border-white/[0.08] pb-3 flex items-center gap-2">
              <Dumbbell className="w-4 h-4 text-white" />
              Цель и расписание тренировок
            </h3>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">Цель программы</label>
                <select
                  value={goal}
                  onChange={e => setGoal(e.target.value)}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none"
                >
                  <option value="weight_loss">Похудение и сжигание жира</option>
                  <option value="muscle_gain">Набор сухой мышечной массы</option>
                  <option value="recomposition">Рельеф и поддержание формы</option>
                  <option value="endurance">Выносливость и тонус</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">Уровень подготовки</label>
                <select
                  value={level}
                  onChange={e => setLevel(e.target.value)}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none"
                >
                  <option value="beginner">Начинающий (до 6 месяцев)</option>
                  <option value="intermediate">Средний (1–2 года тренировок)</option>
                  <option value="advanced">Продвинутый (атлетический стаж)</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">Инвентарь</label>
                <select
                  value={equipment}
                  onChange={e => setEquipment(e.target.value)}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none"
                >
                  <option value="bodyweight">Только вес тела (дома/на улице)</option>
                  <option value="dumbbells">Гантели / эспандеры</option>
                  <option value="gym">Тренажерный зал (полный набор)</option>
                </select>
              </div>

              {/* 1-7 days per week selector */}
              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">
                  Частота тренировок: <span className="text-white font-bold">{daysPerWeek} дней/нед</span>
                </label>
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

              {/* Duration up to 90 min */}
              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">
                  Длительность тренировки: <span className="text-white font-bold">{workoutDuration} мин</span>
                </label>
                <div className="flex gap-1.5">
                  {[30, 45, 60, 75, 90].map(dur => (
                    <button
                      type="button"
                      key={dur}
                      onClick={() => setWorkoutDuration(dur)}
                      className={`flex-1 py-2 rounded-xl text-xs font-bold transition-all ${
                        workoutDuration === dur
                          ? 'bg-white text-zinc-950 shadow-sm'
                          : 'bg-[#17171d] text-zinc-400 hover:text-white border border-white/[0.08]'
                      }`}
                    >
                      {dur}м
                    </button>
                  ))}
                </div>
              </div>

              {/* Preferred time */}
              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-1.5">
                  Окно тренировки (биологический ритм)
                </label>
                <select
                  value={preferredTime}
                  onChange={e => setPreferredTime(e.target.value)}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#17171d] border border-white/[0.12] text-white text-sm focus:border-white focus:outline-none"
                >
                  <option value="morning">Утро (07:00 – 09:00)</option>
                  <option value="day">День (12:00 – 15:00)</option>
                  <option value="evening">Вечер (18:00 – 21:00)</option>
                </select>
              </div>
            </div>
          </div>

          {/* Section 3: Nutrition & Health */}
          <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-4">
            <h3 className="text-base font-bold text-white border-b border-white/[0.08] pb-3 flex items-center gap-2">
              <Leaf className="w-4 h-4 text-white" />
              Особенности питания и здоровье
            </h3>

            <div className="space-y-4">
              <label className="flex items-center gap-3 p-3.5 rounded-xl bg-[#17171d] border border-white/[0.08] cursor-pointer hover:border-zinc-700">
                <input
                  type="checkbox"
                  checked={isVegan}
                  onChange={e => setIsVegan(e.target.checked)}
                  className="w-4 h-4 rounded text-white focus:ring-white bg-zinc-800 border-zinc-700"
                />
                <div>
                  <div className="text-sm font-bold text-white flex items-center gap-1.5">
                    <Leaf className="w-4 h-4 text-zinc-300" />
                    Веганский рацион
                  </div>
                  <div className="text-xs text-zinc-400">
                    Исключить мясо, рыбу, яйца и молочные продукты. Меню составляется из растительных источников белка (тофу, бобовые, киноа).
                  </div>
                </div>
              </label>

              <div>
                <label className="block text-xs font-semibold text-zinc-400 mb-2">
                  Медицинские ограничения (активируют щадящий режим программы)
                </label>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
                  {[
                    { id: 'joints', label: 'Суставы и колени' },
                    { id: 'spine', label: 'Спина / позвоночник' },
                    { id: 'cardio', label: 'Сердце / давление' },
                  ].map(h => (
                    <button
                      type="button"
                      key={h.id}
                      onClick={() => handleToggleHealth(h.id)}
                      className={`p-3 rounded-xl text-left border text-xs font-bold transition-all ${
                        healthIssues.includes(h.id)
                          ? 'bg-amber-500/15 border-amber-500/40 text-amber-300'
                          : 'bg-[#17171d] border-white/[0.08] text-zinc-400 hover:text-white'
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
            </div>
          </div>

          {/* Unified Single Action Button */}
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4 pt-2">
            <button
              type="submit"
              disabled={saving}
              className="w-full sm:w-auto flex items-center justify-center gap-2 px-8 py-3.5 rounded-xl text-sm font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-all shadow-lg active:scale-95 disabled:opacity-50"
            >
              <Save className="w-4 h-4" />
              <span>{saving ? 'Сохранение...' : 'Сохранить профиль и обновить программу'}</span>
            </button>

            <button
              type="button"
              onClick={handleReset}
              disabled={resetting}
              className="w-full sm:w-auto flex items-center justify-center gap-2 px-4 py-3 rounded-xl text-xs font-bold text-rose-400 hover:text-rose-300 hover:bg-rose-950/30 border border-rose-500/20 transition-colors"
            >
              <Trash2 className="w-3.5 h-3.5" />
              <span>Сбросить данные профиля</span>
            </button>
          </div>
        </form>
      )}
    </div>
  );
};
