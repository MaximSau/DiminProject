import React, { useState } from 'react';
import {
  Activity,
  Dumbbell,
  Apple,
  TrendingUp,
  BookOpen,
  User,
  Sparkles,
  Sun,
  Moon,
  Menu,
  X,
  Printer,
} from 'lucide-react';
import { useTheme } from '../context/ThemeContext';
import { useToast } from '../context/ToastContext';
import { api } from '../api/client';

interface NavbarProps {
  currentTab: string;
  onNavigate: (tab: string) => void;
  onRefresh?: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ currentTab, onNavigate, onRefresh }) => {
  const { theme, toggleTheme } = useTheme();
  const { showToast } = useToast();
  const [seeding, setSeeding] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const handleDemoSeed = async () => {
    try {
      setSeeding(true);
      await api.seedDemo();
      showToast('Демо-сценарий успешно активирован!');
      if (onRefresh) onRefresh();
      onNavigate('dashboard');
    } catch (e: any) {
      showToast(e.message || 'Ошибка активации демо', 'error');
    } finally {
      setSeeding(false);
    }
  };

  const navItems = [
    { id: 'dashboard', label: 'Дашборд', icon: Activity },
    { id: 'workouts', label: 'Тренировки', icon: Dumbbell },
    { id: 'nutrition', label: 'Питание', icon: Apple },
    { id: 'progress', label: 'Прогресс', icon: TrendingUp },
    { id: 'catalog', label: 'Каталог', icon: BookOpen },
    { id: 'profile', label: 'Профиль', icon: User },
  ];

  return (
    <header className="sticky top-0 z-40 w-full border-b backdrop-blur-md transition-colors duration-200 bg-[#08080a]/90 border-white/[0.08] dark:bg-[#08080a]/90 dark:border-white/[0.08]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Logo */}
        <button
          onClick={() => onNavigate('dashboard')}
          className="flex items-center gap-3 group text-left focus:outline-none"
        >
          <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-emerald-500 to-sky-400 p-0.5 shadow-lg shadow-emerald-500/20 group-hover:scale-105 transition-transform duration-200">
            <div className="w-full h-full bg-[#08080a] rounded-[10px] flex items-center justify-center">
              <Activity className="w-5 h-5 text-emerald-400 animate-pulse" />
            </div>
          </div>
          <div>
            <span className="text-xl font-black tracking-tight bg-gradient-to-r from-emerald-400 via-sky-300 to-white bg-clip-text text-transparent">
              PulseFit
            </span>
            <span className="block text-[10px] uppercase tracking-widest font-semibold text-slate-400 -mt-1">
              Pro Fitness
            </span>
          </div>
        </button>

        {/* Desktop Navigation */}
        <nav className="hidden md:flex items-center gap-1.5 lg:gap-2">
          {navItems.map(item => {
            const Icon = item.icon;
            const isActive = currentTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => onNavigate(item.id)}
                className={`flex items-center gap-3 px-3.5 py-2 rounded-xl text-sm font-semibold transition-all duration-150 ${
                  isActive
                    ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 shadow-sm'
                    : 'text-slate-300 hover:text-white hover:bg-slate-800/50 dark:hover:bg-slate-800/60 dark:hover:text-white light:text-slate-600 light:hover:text-slate-900 light:hover:bg-slate-100'
                }`}
              >
                <Icon className={`w-4 h-4 ${isActive ? 'text-emerald-400' : 'text-slate-400'}`} />
                <span>{item.label}</span>
              </button>
            );
          })}
        </nav>

        {/* Right action buttons: Demo, Print, Theme, Mobile toggle */}
        <div className="flex items-center gap-2 sm:gap-3">
          {/* Plan export link */}
          <button
            onClick={() => onNavigate('export')}
            title="Экспорт / Печать программы"
            className="p-2 rounded-xl text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 dark:hover:bg-slate-800/60 transition-colors"
          >
            <Printer className="w-4 h-4" />
          </button>

          {/* Theme switcher */}
          <button
            onClick={toggleTheme}
            title={theme === 'dark' ? 'Включить светлую тему' : 'Включить тёмную тему'}
            className="p-2 rounded-xl text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 dark:hover:bg-slate-800/60 transition-colors"
          >
            {theme === 'dark' ? (
              <Sun className="w-4 h-4 text-amber-400" />
            ) : (
              <Moon className="w-4 h-4 text-sky-400" />
            )}
          </button>

          {/* Demo Button - strictly preserved for project defense */}
          <button
            id="btn-demo-topbar"
            onClick={handleDemoSeed}
            disabled={seeding}
            className="flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-bold uppercase tracking-wider bg-gradient-to-r from-white/10 to-white/15 hover:from-white/20 hover:to-white/25 text-white border border-white/20 shadow-sm hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50"
          >
            <Sparkles className="w-3.5 h-3.5 text-emerald-400 animate-spin" style={{ animationDuration: '4s' }} />
            <span>{seeding ? 'Загрузка...' : 'Демо'}</span>
          </button>

          {/* Mobile menu trigger */}
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800/60"
          >
            {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
          </button>
        </div>
      </div>

      {/* Mobile Menu Dropdown */}
      {mobileMenuOpen && (
        <div className="md:hidden border-b border-white/[0.08] bg-[#08080a] px-4 pt-2 pb-4 space-y-1">
          {navItems.map(item => {
            const Icon = item.icon;
            const isActive = currentTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => {
                  onNavigate(item.id);
                  setMobileMenuOpen(false);
                }}
                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-semibold text-left transition-colors ${
                  isActive
                    ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'
                    : 'text-slate-400 hover:text-white hover:bg-slate-800/40'
                }`}
              >
                <Icon className="w-4 h-4 text-emerald-400" />
                <span>{item.label}</span>
              </button>
            );
          })}
        </div>
      )}
    </header>
  );
};
