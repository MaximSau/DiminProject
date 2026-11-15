import React, { useState, useEffect } from 'react';
import { Navbar } from './components/Navbar';
import { DashboardPage } from './pages/DashboardPage';
import { WorkoutsPage } from './pages/WorkoutsPage';
import { NutritionPage } from './pages/NutritionPage';
import { ProgressPage } from './pages/ProgressPage';
import { CatalogPage } from './pages/CatalogPage';
import { ProfilePage } from './pages/ProfilePage';
import { OnboardingPage } from './pages/OnboardingPage';
import { PlanExportPage } from './pages/PlanExportPage';
import { DashboardResponse } from './types';
import { api } from './api/client';
import { ThemeProvider } from './context/ThemeContext';
import { ToastProvider, useToast } from './context/ToastContext';

const AppContent: React.FC = () => {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);
  const [onboarded, setOnboarded] = useState(false);
  const [dashboardData, setDashboardData] = useState<DashboardResponse | null>(null);

  // Read current tab from URL hash if available (e.g., #workouts)
  const getInitialTab = () => {
    const hash = window.location.hash.replace('#', '');
    const valid = ['dashboard', 'workouts', 'nutrition', 'progress', 'catalog', 'profile', 'export'];
    return valid.includes(hash) ? hash : 'dashboard';
  };

  const [currentTab, setCurrentTab] = useState<string>(getInitialTab);

  const loadBootstrap = async () => {
    try {
      setLoading(true);
      const boot = await api.getBootstrap();
      if (!boot.onboarded) {
        setOnboarded(false);
        return;
      }
      setOnboarded(true);
      const dash = await api.getDashboard();
      setDashboardData(dash);
    } catch (e: any) {
      console.error('Bootstrap error', e);
      // fallback or prompt
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadBootstrap();

    const handleHash = () => {
      const hash = window.location.hash.replace('#', '');
      const valid = ['dashboard', 'workouts', 'nutrition', 'progress', 'catalog', 'profile', 'export'];
      if (valid.includes(hash)) {
        setCurrentTab(hash);
      }
    };
    window.addEventListener('hashchange', handleHash);
    return () => window.removeEventListener('hashchange', handleHash);
  }, []);

  const handleNavigate = (tab: string) => {
    setCurrentTab(tab);
    window.location.hash = tab;
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleRefresh = async () => {
    try {
      const dash = await api.getDashboard();
      setDashboardData(dash);
    } catch (e) {
      console.error('Refresh error', e);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-[#090d16] text-white">
        <div className="w-10 h-10 border-4 border-emerald-500/20 border-t-emerald-500 rounded-full animate-spin mb-4" />
        <span className="text-sm font-semibold tracking-wider uppercase text-slate-400">
          Загрузка PulseFit...
        </span>
      </div>
    );
  }

  if (!onboarded) {
    return (
      <OnboardingPage
        onCompleted={() => {
          loadBootstrap();
        }}
      />
    );
  }

  return (
    <div className="min-h-screen flex flex-col bg-[#090d16] text-slate-100 transition-colors duration-200">
      <Navbar
        currentTab={currentTab}
        onNavigate={handleNavigate}
        onRefresh={handleRefresh}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
        {currentTab === 'dashboard' && dashboardData && (
          <DashboardPage
            data={dashboardData}
            onNavigate={handleNavigate}
            onRefresh={handleRefresh}
          />
        )}
        {currentTab === 'workouts' && (
          <WorkoutsPage onRefresh={handleRefresh} />
        )}
        {currentTab === 'nutrition' && (
          <NutritionPage onRefresh={handleRefresh} />
        )}
        {currentTab === 'progress' && (
          <ProgressPage onRefresh={handleRefresh} />
        )}
        {currentTab === 'catalog' && (
          <CatalogPage />
        )}
        {currentTab === 'profile' && (
          <ProfilePage
            onRefresh={handleRefresh}
            onNavigate={handleNavigate}
          />
        )}
        {currentTab === 'export' && (
          <PlanExportPage onBack={() => handleNavigate('dashboard')} />
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-800/80 py-6 text-center text-xs text-slate-500 print:hidden">
        <div className="max-w-7xl mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-2">
          <span>PulseFit Pro © 2026. Профессиональный фитнес-трекер и рацион.</span>
          <span>Автономная архитектура • SQLite Pure-Go • React + Tailwind SPA</span>
        </div>
      </footer>
    </div>
  );
};

export const App: React.FC = () => {
  return (
    <ThemeProvider>
      <ToastProvider>
        <AppContent />
      </ToastProvider>
    </ThemeProvider>
  );
};

export default App;
