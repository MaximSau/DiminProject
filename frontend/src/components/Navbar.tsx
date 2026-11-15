import React, { useState } from 'react';
import {
  Activity,
  Dumbbell,
  Apple,
  TrendingUp,
  BookOpen,
  User,
  Sparkles,
  Menu,
  X,
} from 'lucide-react';
import { useToast } from '../context/ToastContext';
import { api } from '../api/client';

interface NavbarProps {
  currentTab: string;
  onNavigate: (tab: string) => void;
  onRefresh?: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ currentTab, onNavigate, onRefresh }) => {
  const { showToast } = useToast();
  const [seeding, setSeeding] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const handleDemoSeed = async () => {
    try {
      setSeeding(true);
      await api.seedDemo();
      showToast('Демонстрационный профиль загружен!');
      if (onRefresh) onRefresh();
      onNavigate('dashboard');
    } catch (e: any) {
      showToast(e.message || 'Ошибка запуска демо', 'error');
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
    <header className="sticky top-0 z-40 w-full border-b border-white/[0.08] bg-[#08080a]/90 backdrop-blur-md transition-colors duration-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand Logo & Name */}
        <div
          onClick={() => onNavigate('dashboard')}
          className="flex items-center gap-3 cursor-pointer group select-none"
        >
          <div className="w-9 h-9 rounded-xl bg-white/10 text-white border border-white/20 flex items-center justify-center shadow-sm group-hover:scale-105 transition-all duration-200">
            <Activity className="w-5 h-5 text-white" />
          </div>
          <div>
            <div className="flex items-center gap-1.5">
              <span className="font-black text-lg tracking-tight text-white group-hover:text-zinc-200 transition-colors">
                PulseFit
              </span>
              <span className="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded bg-white/10 text-zinc-300 border border-white/20">
                PRO
              </span>
            </div>
            <span className="text-[10px] text-zinc-400 block -mt-0.5 font-medium">
              Персональный тренинг & рацион
            </span>
          </div>
        </div>

        {/* Desktop Navigation Links */}
        <nav className="hidden md:flex items-center gap-1">
          {navItems.map(item => {
            const Icon = item.icon;
            const isActive = currentTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => onNavigate(item.id)}
                className={`flex items-center gap-2 px-3 py-2 rounded-xl text-xs font-bold transition-all duration-150 ${
                  isActive
                    ? 'bg-white text-zinc-950 shadow-md shadow-white/5 font-black scale-[1.02]'
                    : 'text-zinc-400 hover:text-white hover:bg-zinc-800/60'
                }`}
              >
                <Icon className={`w-4 h-4 ${isActive ? 'text-zinc-950' : 'text-zinc-400'}`} />
                <span>{item.label}</span>
              </button>
            );
          })}
        </nav>

        {/* Right action buttons: Demo, Mobile toggle */}
        <div className="flex items-center gap-2 sm:gap-3">
          {/* Demo Button - strictly preserved for project defense */}
          <button
            id="btn-demo-topbar"
            onClick={handleDemoSeed}
            disabled={seeding}
            className="flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-bold uppercase tracking-wider bg-white/10 hover:bg-white/20 text-white border border-white/20 shadow-sm hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50"
          >
            <Sparkles className="w-3.5 h-3.5 text-zinc-300" />
            <span>{seeding ? 'Загрузка...' : 'Демо'}</span>
          </button>

          {/* Mobile menu trigger */}
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden p-2 rounded-xl text-zinc-400 hover:text-white hover:bg-zinc-800/60"
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
                    ? 'bg-white/10 text-white border border-white/20'
                    : 'text-zinc-400 hover:text-white hover:bg-zinc-800/40'
                }`}
              >
                <Icon className={`w-4 h-4 ${isActive ? 'text-white' : 'text-zinc-400'}`} />
                <span>{item.label}</span>
              </button>
            );
          })}
        </div>
      )}
    </header>
  );
};
