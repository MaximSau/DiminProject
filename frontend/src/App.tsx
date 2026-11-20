import React, { useState, useEffect } from 'react';
import { Navbar } from './components/Navbar';
import { DashboardPage } from './pages/DashboardPage';
import { WorkoutsPage } from './pages/WorkoutsPage';
import { NutritionPage } from './pages/NutritionPage';
import { ProgressPage } from './pages/ProgressPage';
import { CatalogPage } from './pages/CatalogPage';
import { ProfilePage } from './pages/ProfilePage';
import { OnboardingPage } from './pages/OnboardingPage';
import { DashboardResponse } from './types';
import { api } from './api/client';
import { ThemeProvider } from './context/ThemeContext';
import { ToastProvider, useToast } from './context/ToastContext';

const AppContent: React.FC = () => {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);
  const [onboarded, setOnboarded] = useState(false);
  const [dashboardData, setDashboardData] = useState<DashboardResponse | null>(null);

  // Read current tab from pathname or URL hash (e.g., /workouts or #workouts)
  const getInitialTab = () => {
    const path = window.location.pathname.replace(/^\//, '').split('/')[0];
    const hash = window.location.hash.replace('#', '');
    const valid = ['dashboard', 'workouts', 'nutrition', 'progress', 'catalog', 'profile'];
    if (valid.includes(path)) return path;
    if (valid.includes(hash)) return hash;
    return 'dashboard';
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
      setOnboarded(false);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadBootstrap();

    const handleHash = () => {
      const hash = window.location.hash.replace('#', '');
      const valid = ['dashboard', 'workouts', 'nutrition', 'progress', 'catalog', 'profile'];
      if (valid.includes(hash)) {
        setCurrentTab(hash);
      }
    };

    const handleReset = () => {
      setOnboarded(false);
      setDashboardData(null);
      window.location.hash = 'dashboard';
      setCurrentTab('dashboard');
    };

    window.addEventListener('hashchange', handleHash);
    window.addEventListener('pulsefit:reset', handleReset);
    return () => {
      window.removeEventListener('hashchange', handleHash);
      window.removeEventListener('pulsefit:reset', handleReset);
    };
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
      <div className="min-h-screen flex flex-col items-center justify-center bg-[#08080a] text-white">
        <div className="w-10 h-10 border-4 border-white/20 border-t-white rounded-full animate-spin mb-4" />
        <span className="text-sm font-semibold tracking-wider uppercase text-zinc-400">
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
    <div className="min-h-screen flex flex-col bg-[#08080a] text-zinc-100 transition-colors duration-200">
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
      </main>

      {/* Footer */}
      <footer className="border-t border-white/[0.08] py-6 text-center text-xs text-zinc-500 print:hidden">
        <div className="max-w-7xl mx-auto px-4 flex items-center justify-center">
          <span>PulseFit Pro © 2026. Профессиональный фитнес-трекер и рацион.</span>
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
