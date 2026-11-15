import React, { useState, useEffect } from 'react';
import {
  BookOpen,
  Search,
  Sparkles,
  Info,
  ShieldAlert,
  AlertTriangle,
  CheckCircle2,
  X,
  Utensils,
  Dumbbell,
  Leaf,
  Plus,
} from 'lucide-react';
import { CatalogExercise, CatalogMeal, UserProfile } from '../types';
import { api } from '../api/client';
import { useToast } from '../context/ToastContext';

export const CatalogPage: React.FC = () => {
  const { showToast } = useToast();
  const [activeTab, setActiveTab] = useState<'exercises' | 'meals'>('exercises');

  // User state for gentle & vegan checks
  const [user, setUser] = useState<UserProfile | null>(null);

  // Exercises state
  const [loadingEx, setLoadingEx] = useState(true);
  const [exercises, setExercises] = useState<CatalogExercise[]>([]);
  const [searchEx, setSearchEx] = useState('');
  const [muscle, setMuscle] = useState('');
  const [equipment, setEquipment] = useState('');
  const [selectedEx, setSelectedEx] = useState<CatalogExercise | null>(null);

  // Meals state
  const [loadingMeals, setLoadingMeals] = useState(false);
  const [meals, setMeals] = useState<CatalogMeal[]>([]);
  const [searchMeal, setSearchMeal] = useState('');
  const [mealSlot, setMealSlot] = useState('');
  const [veganOnly, setVeganOnly] = useState(false);
  const [selectedMeal, setSelectedMeal] = useState<CatalogMeal | null>(null);

  // Vegan warning confirmation modal
  const [veganWarningMeal, setVeganWarningMeal] = useState<CatalogMeal | null>(null);
  const [addingMealId, setAddingMealId] = useState<string | null>(null);

  const todayIso = new Date().getDay() === 0 ? 7 : new Date().getDay();

  useEffect(() => {
    api.getProfile()
      .then(res => setUser(res.user))
      .catch(() => {});
  }, []);

  const fetchExercises = async () => {
    try {
      setLoadingEx(true);
      const res = await api.getCatalog({
        q: searchEx || undefined,
        muscle: muscle || undefined,
        equipment: equipment || undefined,
      });
      setExercises(res.exercises || []);
    } catch (e: any) {
      showToast(e.message || 'Ошибка загрузки каталога упражнений', 'error');
    } finally {
      setLoadingEx(false);
    }
  };

  const fetchMeals = async () => {
    try {
      setLoadingMeals(true);
      const res = await api.getCatalogMeals({
        q: searchMeal || undefined,
        slot: mealSlot || undefined,
        vegan: veganOnly || undefined,
      });
      setMeals(res.meals || []);
    } catch (e: any) {
      showToast(e.message || 'Ошибка загрузки каталога питания', 'error');
    } finally {
      setLoadingMeals(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'exercises') {
      fetchExercises();
    } else {
      fetchMeals();
    }
  }, [activeTab, searchEx, muscle, equipment, searchMeal, mealSlot, veganOnly]);

  const handleAddMealConfirmed = async (m: CatalogMeal) => {
    try {
      setAddingMealId(m.id);
      await api.addMealToDay(todayIso, m.id);
      showToast(`Блюдо «${m.name}» добавлено в сегодняшний рацион!`);
      setVeganWarningMeal(null);
    } catch (e: any) {
      showToast(e.message || 'Ошибка добавления блюда', 'error');
    } finally {
      setAddingMealId(null);
    }
  };

  const handleSelectMeal = (m: CatalogMeal) => {
    if (user?.isVegan && !m.isVegan) {
      // Big red exclamation mark confirmation required!
      setVeganWarningMeal(m);
    } else {
      handleAddMealConfirmed(m);
    }
  };

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

  const slotList = [
    { id: '', label: 'Все приемы' },
    { id: 'breakfast', label: 'Завтрак' },
    { id: 'lunch', label: 'Обед' },
    { id: 'snack', label: 'Перекус' },
    { id: 'dinner', label: 'Ужин' },
  ];

  return (
    <div className="space-y-6">
      {/* Header and Section Tabs */}
      <div className="rounded-2xl p-6 border bg-[#121216] border-white/[0.08] shadow-sm space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">
              База знаний и атлетический справочник
            </span>
            <h2 className="text-2xl font-black text-white mt-0.5">
              Каталог PulseFit
            </h2>
            <p className="text-xs sm:text-sm text-zinc-300 mt-1 font-normal">
              Библиотека всех движений и сбалансированных блюд с пошаговыми инструкциями и подсказками тренера.
            </p>
          </div>

          {/* Switch tabs: Exercises vs Meals */}
          <div className="flex items-center gap-1.5 p-1 rounded-xl bg-[#17171d] border border-white/[0.08] self-start sm:self-center">
            <button
              onClick={() => setActiveTab('exercises')}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold transition-all ${
                activeTab === 'exercises'
                  ? 'bg-white text-zinc-950 shadow-md font-black'
                  : 'text-zinc-400 hover:text-white'
              }`}
            >
              <Dumbbell className="w-4 h-4" />
              <span>Упражнения</span>
            </button>
            <button
              onClick={() => setActiveTab('meals')}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold transition-all ${
                activeTab === 'meals'
                  ? 'bg-white text-zinc-950 shadow-md font-black'
                  : 'text-zinc-400 hover:text-white'
              }`}
            >
              <Utensils className="w-4 h-4" />
              <span>Каталог питания</span>
            </button>
          </div>
        </div>

        {/* Tab 1 Filters: Exercises */}
        {activeTab === 'exercises' && (
          <div className="space-y-3 pt-2">
            <div className="relative">
              <Search className="w-5 h-5 absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-400" />
              <input
                type="text"
                value={searchEx}
                onChange={e => setSearchEx(e.target.value)}
                placeholder="Поиск упражнения по названию или мышцам (например, приседания, становая, жим)..."
                className="w-full pl-11 pr-4 py-3 rounded-xl bg-[#17171d] border border-white/[0.12] text-white placeholder-zinc-500 text-sm focus:border-white focus:outline-none transition-colors"
              />
            </div>

            <div className="flex items-center gap-1.5 overflow-x-auto pb-1 no-scrollbar">
              {musclesList.map(m => (
                <button
                  key={m.id}
                  onClick={() => setMuscle(m.id)}
                  className={`px-3.5 py-1.5 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
                    muscle === m.id
                      ? 'bg-white text-zinc-950 shadow-sm'
                      : 'bg-[#17171d] text-zinc-300 hover:text-white hover:bg-[#202028] border border-white/[0.08]'
                  }`}
                >
                  {m.label}
                </button>
              ))}
            </div>

            <div className="flex items-center gap-1.5 overflow-x-auto pb-1 no-scrollbar">
              {equipList.map(eq => (
                <button
                  key={eq.id}
                  onClick={() => setEquipment(eq.id)}
                  className={`px-3 py-1 rounded-lg text-[11px] font-semibold whitespace-nowrap transition-all ${
                    equipment === eq.id
                      ? 'bg-zinc-200 text-zinc-950 shadow-sm'
                      : 'bg-[#17171d] text-zinc-400 hover:text-white hover:bg-[#202028] border border-white/[0.06]'
                  }`}
                >
                  {eq.label}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Tab 2 Filters: Meals */}
        {activeTab === 'meals' && (
          <div className="space-y-3 pt-2">
            <div className="relative">
              <Search className="w-5 h-5 absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-400" />
              <input
                type="text"
                value={searchMeal}
                onChange={e => setSearchMeal(e.target.value)}
                placeholder="Поиск блюда по названию или ингредиентам (например, овсянка, тофу, лосось, киноа)..."
                className="w-full pl-11 pr-4 py-3 rounded-xl bg-[#17171d] border border-white/[0.12] text-white placeholder-zinc-500 text-sm focus:border-white focus:outline-none transition-colors"
              />
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-1.5 overflow-x-auto pb-1 no-scrollbar">
                {slotList.map(s => (
                  <button
                    key={s.id}
                    onClick={() => setMealSlot(s.id)}
                    className={`px-3.5 py-1.5 rounded-xl text-xs font-bold whitespace-nowrap transition-all ${
                      mealSlot === s.id
                        ? 'bg-white text-zinc-950 shadow-sm'
                        : 'bg-[#17171d] text-zinc-300 hover:text-white hover:bg-[#202028] border border-white/[0.08]'
                    }`}
                  >
                    {s.label}
                  </button>
                ))}
              </div>

              <button
                onClick={() => setVeganOnly(!veganOnly)}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold transition-all border ${
                  veganOnly
                    ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40'
                    : 'bg-[#17171d] text-zinc-400 hover:text-white border-white/[0.08]'
                }`}
              >
                <Leaf className="w-3.5 h-3.5 text-emerald-400" />
                <span>Только веганские</span>
              </button>
            </div>
          </div>
        )}
      </div>

      {/* EXERCISES CONTENT */}
      {activeTab === 'exercises' && (
        loadingEx ? (
          <div className="py-20 text-center text-zinc-400">Поиск упражнений...</div>
        ) : exercises.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {exercises.map(ex => {
              const isContra = ex.isContraindicated && user?.hasHealthIssues;

              return (
                <div
                  key={ex.id}
                  onClick={() => setSelectedEx(ex)}
                  className={`rounded-2xl p-5 border shadow-sm transition-all duration-200 cursor-pointer flex flex-col justify-between group hover:scale-[1.01] ${
                    isContra
                      ? 'bg-[#161214] border-rose-500/40 hover:border-rose-500/70'
                      : 'bg-[#121216] border-white/[0.08] hover:border-white/[0.22]'
                  }`}
                >
                  <div className="space-y-3">
                    <div className="flex items-start justify-between gap-2">
                      <div>
                        <div className="mb-1">
                          {user?.hasHealthIssues ? (
                            isContra ? (
                              <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-rose-500/20 text-rose-300 border border-rose-500/40">
                                ⚠️ Нельзя выполнять при ограничениях
                              </span>
                            ) : (
                              <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                                ✓ Можно выполнять
                              </span>
                            )
                          ) : (
                            ex.isGentle && (
                              <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-amber-500/20 text-amber-300 border border-amber-500/30">
                                <ShieldAlert className="w-3 h-3" />
                                Щадящее
                              </span>
                            )
                          )}
                        </div>
                        <h3 className={`text-base font-bold transition-colors ${isContra ? 'text-rose-100' : 'text-white group-hover:text-zinc-200'}`}>
                          {ex.title}
                        </h3>
                      </div>
                    </div>

                    <div className="flex flex-wrap items-center gap-1.5">
                      <span className="text-[11px] font-semibold px-2 py-0.5 rounded-md bg-[#1c1c24] text-zinc-200 border border-white/[0.08]">
                        {ex.muscle}
                      </span>
                      <span className="text-[11px] font-semibold px-2 py-0.5 rounded-md bg-[#1c1c24] text-zinc-300 border border-white/[0.06]">
                        {ex.equipment}
                      </span>
                      <span className="text-[11px] font-semibold px-2 py-0.5 rounded-md bg-[#1c1c24] text-zinc-400 border border-white/[0.04]">
                        {ex.difficulty}
                      </span>
                    </div>

                    {ex.coachTip && (
                      <div className="text-xs text-zinc-200 bg-[#17171d] p-3 rounded-xl border border-white/[0.08] leading-relaxed">
                        <span className="text-white font-bold block mb-0.5">💡 Техника и безопасность:</span>
                        <span className="font-normal text-zinc-300">{ex.coachTip}</span>
                      </div>
                    )}
                  </div>

                  <div className="pt-3 mt-3 border-t border-white/[0.08] flex items-center justify-between text-xs font-bold text-zinc-300 group-hover:text-white transition-colors">
                    <span>Подробный разбор</span>
                    <Info className="w-4 h-4 text-zinc-400" />
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="py-20 text-center rounded-2xl border border-white/[0.08] bg-[#121216] space-y-2">
            <BookOpen className="w-8 h-8 text-zinc-400 mx-auto opacity-80" />
            <div className="text-sm font-semibold text-white">Упражнений не найдено</div>
            <p className="text-xs text-zinc-400">Попробуйте изменить поисковый запрос или сбросить фильтры.</p>
          </div>
        )
      )}

      {/* MEALS CONTENT */}
      {activeTab === 'meals' && (
        loadingMeals ? (
          <div className="py-20 text-center text-zinc-400">Загрузка каталога блюд...</div>
        ) : meals.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {meals.map(m => (
              <div
                key={m.id}
                className="rounded-2xl p-5 border bg-[#121216] border-white/[0.08] hover:border-white/[0.22] shadow-sm transition-all flex flex-col justify-between gap-4 group"
              >
                <div className="space-y-3">
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-[11px] font-bold uppercase tracking-wider text-zinc-400">
                          {m.slot === 'breakfast' ? 'Завтрак' : m.slot === 'lunch' ? 'Обед' : m.slot === 'snack' ? 'Перекус' : 'Ужин'}
                        </span>
                        {m.isVegan ? (
                          <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                            <Leaf className="w-3 h-3 text-emerald-400" />
                            Веганское
                          </span>
                        ) : (
                          <span className="text-[10px] font-semibold px-2 py-0.5 rounded-md bg-white/5 text-zinc-400 border border-white/[0.06]">
                            Стандартное
                          </span>
                        )}
                      </div>
                      <h3 className="text-base font-bold text-white group-hover:text-zinc-200 transition-colors">
                        {m.name}
                      </h3>
                    </div>
                    <span className="text-xs font-bold px-2.5 py-1 rounded-lg bg-white/10 text-white border border-white/20 flex-shrink-0">
                      {m.calories} ккал
                    </span>
                  </div>

                  {/* Macros */}
                  <div className="flex items-center gap-2 text-xs text-zinc-300 font-medium">
                    <span>Б: {m.proteins}г</span>
                    <span>• Ж: {m.fats}г</span>
                    <span>• У: {m.carbs}г</span>
                  </div>

                  {/* Portion & Ingredients */}
                  <div className="p-3 rounded-xl bg-[#17171d] border border-white/[0.08] text-xs space-y-1.5">
                    <div className="text-zinc-300">
                      <span className="font-bold text-white">Порция:</span> {m.portion}
                    </div>
                    {m.ingredients && m.ingredients.length > 0 && (
                      <div className="text-zinc-400 line-clamp-2">
                        <span className="font-bold text-zinc-300">Состав:</span> {m.ingredients.join(', ')}
                      </div>
                    )}
                  </div>
                </div>

                <div className="pt-3 border-t border-white/[0.08] flex items-center justify-between gap-2">
                  <button
                    onClick={() => setSelectedMeal(m)}
                    className="text-xs font-bold text-zinc-300 hover:text-white flex items-center gap-1 transition-colors"
                  >
                    <Info className="w-3.5 h-3.5 text-zinc-400" />
                    Рецепт
                  </button>

                  <button
                    onClick={() => handleSelectMeal(m)}
                    disabled={addingMealId === m.id}
                    className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl text-xs font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-all shadow-sm active:scale-95 disabled:opacity-50"
                  >
                    <Plus className="w-3.5 h-3.5" />
                    <span>{addingMealId === m.id ? 'Добавление...' : 'В рацион дня'}</span>
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="py-20 text-center rounded-2xl border border-white/[0.08] bg-[#121216] space-y-2">
            <Utensils className="w-8 h-8 text-zinc-400 mx-auto opacity-80" />
            <div className="text-sm font-semibold text-white">Блюд не найдено</div>
            <p className="text-xs text-zinc-400">Попробуйте изменить поисковый запрос или сбросить фильтры.</p>
          </div>
        )
      )}

      {/* BIG RED EXCLAMATION MARK VEGAN WARNING MODAL */}
      {veganWarningMeal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md animate-in fade-in">
          <div className="rounded-3xl p-6 sm:p-8 bg-[#161214] border-2 border-rose-500/70 shadow-2xl max-w-md w-full space-y-5 text-center">
            {/* Big Red Exclamation Sign */}
            <div className="w-20 h-20 rounded-full bg-rose-500/20 border-2 border-rose-500 text-rose-500 flex items-center justify-center mx-auto animate-bounce">
              <AlertTriangle className="w-12 h-12 stroke-[2.5]" />
            </div>

            <div className="space-y-2">
              <span className="text-xs font-black uppercase tracking-widest text-rose-400 block">
                Внимание: Блюдо не является веганским!
              </span>
              <h3 className="text-xl font-black text-white">
                «{veganWarningMeal.name}»
              </h3>
              <p className="text-xs sm:text-sm text-zinc-300 leading-relaxed font-normal">
                В вашем профиле отмечен <span className="text-emerald-400 font-bold">100% веганский рацион</span>. Данное блюдо содержит продукты животного происхождения (мясо, рыбу, птицу, яйца или лактозу).
              </p>
            </div>

            <div className="p-3.5 rounded-xl bg-[#1d1416] border border-rose-500/30 text-xs text-rose-200">
              Вы уверены, что действительно съели это блюдо и хотите учесть его в калориях и БЖУ?
            </div>

            <div className="flex flex-col sm:flex-row items-center gap-2.5 pt-2">
              <button
                type="button"
                onClick={() => setVeganWarningMeal(null)}
                className="w-full sm:w-1/2 py-2.5 rounded-xl text-xs font-bold text-zinc-400 hover:text-white bg-[#1a1a22] border border-white/[0.08] transition-colors"
              >
                Отмена
              </button>
              <button
                type="button"
                onClick={() => handleAddMealConfirmed(veganWarningMeal)}
                className="w-full sm:w-1/2 py-2.5 rounded-xl text-xs font-bold bg-rose-600 hover:bg-rose-500 text-white shadow-lg shadow-rose-600/30 transition-all active:scale-95"
              >
                Подтверждаю, съел
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Exercise Detail Modal */}
      {selectedEx && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in">
          <div className="rounded-2xl p-6 bg-[#121216] border border-white/20 shadow-2xl max-w-lg w-full space-y-4">
            <div className="flex items-start justify-between">
              <div>
                <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                  {selectedEx.muscle} • {selectedEx.equipment}
                </span>
                <h3 className="text-xl font-bold text-white mt-1">{selectedEx.title}</h3>
              </div>
              <button
                onClick={() => setSelectedEx(null)}
                className="p-1 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="flex gap-2">
              <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-[#1c1c24] text-zinc-200 border border-white/[0.08]">
                Сложность: {selectedEx.difficulty}
              </span>
              {user?.hasHealthIssues ? (
                selectedEx.isContraindicated ? (
                  <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-rose-500/20 text-rose-300 border border-rose-500/40">
                    ⚠️ Противопоказано тренером
                  </span>
                ) : (
                  <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-emerald-500/15 text-emerald-300 border border-emerald-500/30">
                    ✓ Безопасно для суставов
                  </span>
                )
              ) : selectedEx.isGentle ? (
                <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-amber-500/20 text-amber-300 border border-amber-500/30">
                  Подходит для щадящего режима
                </span>
              ) : null}
            </div>

            <div className="p-4 rounded-xl bg-[#17171d] border border-white/[0.10] text-xs sm:text-sm text-zinc-200 leading-relaxed space-y-2">
              <div className="font-bold text-white flex items-center gap-1.5">
                <Sparkles className="w-4 h-4 text-white" />
                Биомеханика и совет тренера:
              </div>
              <p className="font-normal text-zinc-300">
                {selectedEx.coachTip || 'Выполняйте движение плавно, сохраняя контроль над целевыми мышечными группами и дыханием.'}
              </p>
            </div>

            <div className="flex justify-end pt-2 border-t border-white/[0.08]">
              <button
                onClick={() => setSelectedEx(null)}
                className="px-5 py-2.5 rounded-xl text-xs font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-colors shadow-md active:scale-95"
              >
                Закрыть карточку
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Meal Detail Modal */}
      {selectedMeal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in">
          <div className="rounded-2xl p-6 bg-[#121216] border border-white/20 shadow-2xl max-w-lg w-full space-y-4">
            <div className="flex items-start justify-between">
              <div>
                <span className="text-xs font-bold uppercase tracking-wider text-zinc-400">
                  {selectedMeal.slot === 'breakfast' ? 'Завтрак' : selectedMeal.slot === 'lunch' ? 'Обед' : selectedMeal.slot === 'snack' ? 'Перекус' : 'Ужин'} • {selectedMeal.calories} ккал
                </span>
                <h3 className="text-xl font-bold text-white mt-1">{selectedMeal.name}</h3>
              </div>
              <button
                onClick={() => setSelectedMeal(null)}
                className="p-1 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="flex gap-2">
              <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-[#1c1c24] text-zinc-200 border border-white/[0.08]">
                Б: {selectedMeal.proteins}г • Ж: {selectedMeal.fats}г • У: {selectedMeal.carbs}г
              </span>
              {selectedMeal.isVegan && (
                <span className="text-xs font-semibold px-2.5 py-1 rounded-md bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 flex items-center gap-1">
                  <Leaf className="w-3 h-3 text-emerald-400" />
                  100% Веган
                </span>
              )}
            </div>

            <div className="p-4 rounded-xl bg-[#17171d] border border-white/[0.10] text-xs sm:text-sm text-zinc-200 leading-relaxed space-y-2">
              <div className="font-bold text-white">Способ приготовления:</div>
              <p className="font-normal text-zinc-300">
                {selectedMeal.prep || 'Сбалансированное блюдо. Готовьте на умеренном огне с минимальным добавлением масла.'}
              </p>
            </div>

            <div className="flex justify-between items-center pt-2 border-t border-white/[0.08]">
              <span className="text-xs text-zinc-400">Порция: {selectedMeal.portion}</span>
              <button
                onClick={() => {
                  const m = selectedMeal;
                  setSelectedMeal(null);
                  handleSelectMeal(m);
                }}
                className="px-5 py-2.5 rounded-xl text-xs font-bold bg-white hover:bg-zinc-200 text-zinc-950 transition-colors shadow-md active:scale-95"
              >
                Добавить в рацион
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
