import React, { useState, useEffect } from 'react';
import {
  BookOpen,
  Search,
  Sparkles,
  Info,
  ShieldAlert,
  CheckCircle2,
} from 'lucide-react';
import { CatalogExercise } from '../types';
import { api } from '../api/client';
import { useToast } from '../context/ToastContext';

export const CatalogPage: React.FC = () => {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);
  const [exercises, setExercises] = useState<CatalogExercise[]>([]);
  const [search, setSearch] = useState('');
  const [muscle, setMuscle] = useState('');
  const [equipment, setEquipment] = useState('');
  const [difficulty] = useState('');
  const [selectedEx, setSelectedEx] = useState<CatalogExercise | null>(null);

  const fetchCatalog = async () => {
    try {
      setLoading(true);
      const res = await api.getCatalog({
        q: search || undefined,
        muscle: muscle || undefined,
        equipment: equipment || undefined,
        difficulty: difficulty || undefined,
      });
      setExercises(res.exercises || []);
    } catch (e: any) {
      showToast(e.message || 'Ошибка загрузки каталога', 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCatalog();
  }, [search, muscle, equipment, difficulty]);

  const musclesList = [
    { id: '', label: 'Все группы' },
    { id: 'Грудь', label: 'Грудь' },
    { id: 'Спина', label: 'Спина' },
    { id: 'Ноги', label: 'Ноги' },
    { id: 'Плечи', label: 'Плечи' },
    { id: 'Руки', label: 'Руки' },
    { id: 'Пресс', label: 'Пресс' },
    { id: 'Кардио', label: 'Кардио' },
  ];

  const equipList = [
    { id: '', label: 'Любой инвентарь' },
    { id: 'Вес тела', label: 'Свой вес' },
    { id: 'Гантели', label: 'Гантели' },
    { id: 'Турник', label: 'Турник' },
    { id: 'Зал', label: 'Тренажерный зал' },
  ];

  return (
    <div className="space-y-6">
      {/* Header and Search */}
      <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-4">
        <div>
          <span className="text-xs font-bold uppercase tracking-wider text-emerald-400">
            База знаний и атлетический справочник
          </span>
          <h2 className="text-2xl font-black text-white mt-0.5">
            Каталог упражнений
          </h2>
          <p className="text-xs sm:text-sm text-slate-300 mt-1 font-normal">
            Библиотека проверенных движений с правильной биомеханикой, понятными инструкциями и подсказками тренера.
          </p>
        </div>

        {/* Search bar */}
        <div className="relative">
          <Search className="w-5 h-5 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            type="text"
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder="Поиск по названию или мышцам (например, приседания, планка, жим)..."
            className="w-full pl-11 pr-4 py-3 rounded-xl bg-[#17171d] border border-white/[0.12] text-white placeholder-slate-400 text-sm focus:border-emerald-500 focus:outline-none transition-colors"
          />
        </div>

        {/* Filter chips */}
        <div className="space-y-3 pt-1">
          {/* Muscle chips */}
          <div className="flex items-center gap-1.5 overflow-x-auto pb-1 no-scrollbar">
            {musclesList.map(m => (
              <button
                key={m.id}
                onClick={() => setMuscle(m.id)}
                className={`px-3 py-1.5 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
                  muscle === m.id
                    ? 'bg-emerald-500 text-slate-950 shadow-sm'
                    : 'bg-[#17171d] text-slate-200 hover:text-white hover:bg-[#202028] border border-white/[0.08]'
                }`}
              >
                {m.label}
              </button>
            ))}
          </div>

          {/* Equipment chips */}
          <div className="flex items-center gap-1.5 overflow-x-auto pb-1 no-scrollbar">
            {equipList.map(eq => (
              <button
                key={eq.id}
                onClick={() => setEquipment(eq.id)}
                className={`px-3 py-1 rounded-lg text-[11px] font-semibold whitespace-nowrap transition-all ${
                  equipment === eq.id
                    ? 'bg-sky-500 text-slate-950 shadow-sm'
                    : 'bg-[#17171d] text-slate-300 hover:text-white hover:bg-[#202028] border border-white/[0.06]'
                }`}
              >
                {eq.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Exercises Grid */}
      {loading ? (
        <div className="py-20 text-center text-slate-300">Поиск упражнений...</div>
      ) : exercises.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {exercises.map(ex => (
            <div
              key={ex.id}
              onClick={() => setSelectedEx(ex)}
              className="rounded-2xl p-5 border bg-[#121216] border-white/[0.08] hover:border-white/[0.22] shadow-sm transition-all duration-200 cursor-pointer flex flex-col justify-between group hover:scale-[1.01]"
            >
              <div className="space-y-3">
                <div className="flex items-start justify-between gap-2">
                  <h3 className="text-base font-bold text-white group-hover:text-emerald-400 transition-colors">
                    {ex.title}
                  </h3>
                  {ex.isGentle && (
                    <span className="p-1 rounded bg-amber-500/20 text-amber-300 border border-amber-500/30 flex-shrink-0" title="Щадящее упражнение">
                      <ShieldAlert className="w-3.5 h-3.5" />
                    </span>
                  )}
                </div>

                <div className="flex flex-wrap items-center gap-1.5">
                  <span className="text-[11px] font-semibold px-2 py-0.5 rounded-md bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                    {ex.muscle}
                  </span>
                  <span className="text-[11px] font-semibold px-2 py-0.5 rounded-md bg-[#1c1c24] text-slate-200 border border-white/[0.08]">
                    {ex.equipment}
                  </span>
                  <span className="text-[11px] font-semibold px-2 py-0.5 rounded-md bg-[#1c1c24] text-slate-300 border border-white/[0.06]">
                    {ex.difficulty}
                  </span>
                </div>

                {/* Highly readable Coach Tip */}
                {ex.coachTip && (
                  <div className="text-xs text-slate-200 bg-[#17171d] p-3 rounded-xl border border-white/[0.08] leading-relaxed">
                    <span className="text-emerald-400 font-bold block mb-0.5">💡 Техника и безопасность:</span>
                    <span className="font-normal text-slate-300">{ex.coachTip}</span>
                  </div>
                )}
              </div>

              <div className="pt-3 mt-3 border-t border-white/[0.08] flex items-center justify-between text-xs font-bold text-emerald-400 group-hover:text-emerald-300 transition-colors">
                <span>Подробный разбор</span>
                <Info className="w-4 h-4" />
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="py-20 text-center rounded-2xl border border-white/[0.08] bg-[#121216] space-y-2">
          <BookOpen className="w-8 h-8 text-slate-400 mx-auto opacity-80" />
          <div className="text-sm font-semibold text-white">Упражнений не найдено</div>
          <p className="text-xs text-slate-300">Попробуйте изменить поисковый запрос или сбросить фильтры.</p>
        </div>
      )}

      {/* Exercise Detail Modal */}
      {selectedEx && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in">
          <div className="rounded-2xl p-6 bg-[#121216] border border-white/20 shadow-2xl max-w-lg w-full space-y-4">
            <div className="flex items-start justify-between">
              <div>
                <span className="text-xs font-bold uppercase tracking-wider text-emerald-400">
                  {selectedEx.muscle} • {selectedEx.equipment}
                </span>
                <h3 className="text-xl font-bold text-white mt-1">{selectedEx.title}</h3>
              </div>
              <button
                onClick={() => setSelectedEx(null)}
                className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-[#202028]"
              >
                ✕
              </button>
            </div>

            <div className="flex gap-2">
              <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                Сложность: {selectedEx.difficulty}
              </span>
              {selectedEx.isGentle && (
                <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-amber-500/20 text-amber-300 border border-amber-500/30">
                  Подходит для щадящего режима
                </span>
              )}
            </div>

            <div className="p-4 rounded-xl bg-[#17171d] border border-white/[0.10] text-xs sm:text-sm text-slate-200 leading-relaxed space-y-2">
              <div className="font-bold text-white flex items-center gap-1.5">
                <Sparkles className="w-4 h-4 text-emerald-400" />
                Биомеханика и совет тренера:
              </div>
              <p className="font-normal text-slate-300">
                {selectedEx.coachTip || 'Выполняйте движение плавно, сохраняя контроль над целевыми мышечными группами и дыханием.'}
              </p>
            </div>

            <div className="flex justify-end pt-2 border-t border-white/[0.08]">
              <button
                onClick={() => setSelectedEx(null)}
                className="px-5 py-2.5 rounded-xl text-xs font-bold bg-emerald-500 hover:bg-emerald-400 text-slate-950 transition-colors shadow-md shadow-emerald-500/20 active:scale-95"
              >
                Закрыть карточку
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
