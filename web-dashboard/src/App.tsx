import React, { Suspense, useCallback, useEffect, useState } from 'react';
import {
  TrendingUp,
  LayoutDashboard,
  Globe,
  Microscope,
  Settings,
  LogOut,
  Activity,
  ChevronRight,
  ShieldCheck,
  Cpu,
  Languages,
  Menu,
  X,
  Zap
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { api, AUTH_EVENTS, type Metrics, type AnalyticsSnapshot, type AnomalyRecord, type MLStatsEnvelope, type CapturedRequest, type ModificationRule, type TunnelSessionHistory } from './api/client';
import './index.css';
import { Button } from './components/ui/Button';

const MetricCard = React.lazy(() => import('./components/MetricCard').then(module => ({ default: module.MetricCard })));
const RealtimeChart = React.lazy(() => import('./components/RealtimeChart').then(module => ({ default: module.RealtimeChart })));
const TunnelsList = React.lazy(() => import('./components/TunnelsList').then(module => ({ default: module.TunnelsList })));
const GeoMap = React.lazy(() => import('./components/GeoMap').then(module => ({ default: module.GeoMap })));
const AnomalyAlerts = React.lazy(() => import('./components/AnomalyAlerts').then(module => ({ default: module.AnomalyAlerts })));
const ModelComparison = React.lazy(() => import('./components/ModelComparison').then(module => ({ default: module.ModelComparison })));
const TrafficInspector = React.lazy(() => import('./components/TrafficInspector').then(module => ({ default: module.TrafficInspector })));
const ModificationRules = React.lazy(() => import('./components/ModificationRules').then(module => ({ default: module.ModificationRules })));
const ThreatRadar = React.lazy(() => import('./components/ThreatRadar').then(module => ({ default: module.ThreatRadar })));
const ApiKeyManager = React.lazy(() => import('./components/ApiKeyManager').then(module => ({ default: module.ApiKeyManager })));
import { LoginPage } from './components/LoginPage';
import { RegisterPage } from './components/RegisterPage';
import { LandingPage } from './components/LandingPage';
import { ConnectModal } from './components/ConnectModal';
import { ShareView } from './components/ShareView';
const Reservations = React.lazy(() => import('./components/Reservations').then(module => ({ default: module.Reservations })));

type NavTab = 'overview' | 'tunnels' | 'ai_gateway' | 'traffic' | 'settings' | 'api_keys' | 'reservations';

function App() {
  const { t, i18n } = useTranslation();
  const [user, setUser] = useState<any>(null);
  const [isAuthStarted, setIsAuthStarted] = useState(false);
  const [authMode, setAuthMode] = useState<'login' | 'register'>('login');
  const [activeTab, setActiveTab] = useState<NavTab>('overview');
  const [isConnectOpen, setIsConnectOpen] = useState(false);
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [analytics, setAnalytics] = useState<AnalyticsSnapshot | null>(null);
  const [tunnels, setTunnels] = useState<any[]>([]);
  const [anomalies, setAnomalies] = useState<AnomalyRecord[]>([]);
  const [mlStats, setMlStats] = useState<MLStatsEnvelope>({ stats: {}, active_tunnels: 0, ml_up: false, last_prediction_at: null });
  const [history, setHistory] = useState<CapturedRequest[]>([]);
  const [rules, setRules] = useState<ModificationRule[]>([]);
  const [tunnelHistory, setTunnelHistory] = useState<TunnelSessionHistory[]>([]);
  const [loading, setLoading] = useState(true);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

  const clearLocalSession = useCallback(() => {
    localStorage.removeItem('gorenel_user');
    setUser(null);
    setIsAuthStarted(true);
    setAuthMode('login');
  }, []);

  useEffect(() => {
    const checkSession = async () => {
      const storedUser = localStorage.getItem('gorenel_user');
      if (storedUser) {
        setUser(JSON.parse(storedUser));
        setIsAuthStarted(true); // Skip landing page for returning users
        setLoading(false);
        return;
      }
      try {
        const data = await api.getMe();
        if (data && data.user) {
          setUser(data.user);
          localStorage.setItem('gorenel_user', JSON.stringify(data.user));
          setIsAuthStarted(true); // Skip landing page for cookie-based sessions
        }
      } catch (err) {
        console.log('No active session');
        localStorage.removeItem('gorenel_user');
      } finally {
        setLoading(false);
      }
    };
    checkSession();
  }, []);

  useEffect(() => {
    const onUnauthorized = () => {
      clearLocalSession();
    };
    window.addEventListener(AUTH_EVENTS.unauthorized, onUnauthorized as EventListener);
    return () => window.removeEventListener(AUTH_EVENTS.unauthorized, onUnauthorized as EventListener);
  }, [clearLocalSession]);

  const fetchData = useCallback(async () => {
    const settled = await Promise.allSettled([
      api.getMetrics(),
      api.getAnalytics(),
      api.getTunnels(),
      api.getAnomalies(),
      api.getMLStats(),
      api.getTrafficHistory(),
      api.getModificationRules(),
      api.getTunnelHistory(),
    ]);

    const [
      metricsRes,
      analyticsRes,
      tunnelsRes,
      anomaliesRes,
      mlStatsRes,
      historyRes,
      rulesRes,
      tunnelHistoryRes,
    ] = settled;

    if (metricsRes.status === 'fulfilled') setMetrics(metricsRes.value);
    if (analyticsRes.status === 'fulfilled') setAnalytics(analyticsRes.value);
    if (tunnelsRes.status === 'fulfilled') setTunnels(tunnelsRes.value.tunnels || []);
    if (anomaliesRes.status === 'fulfilled') setAnomalies(anomaliesRes.value.anomalies || []);
    if (mlStatsRes.status === 'fulfilled') setMlStats(mlStatsRes.value || { stats: {}, active_tunnels: 0, ml_up: false, last_prediction_at: null });
    if (historyRes.status === 'fulfilled') setHistory(historyRes.value || []);
    if (rulesRes.status === 'fulfilled') setRules(rulesRes.value || []);
    if (tunnelHistoryRes.status === 'fulfilled') setTunnelHistory(tunnelHistoryRes.value.sessions || []);

    const failed = settled.filter((r) => r.status === 'rejected');
    failed.forEach((r) => {
      if (r.status === 'rejected') console.warn('API partial failure:', r.reason);
    });

    const anyUnauthorized = failed.some((r: any) => r?.reason?.response?.status === 401);
    if (anyUnauthorized) {
      clearLocalSession();
      return;
    }

    const metricsDown = metricsRes.status === 'rejected';
    const analyticsDown = analyticsRes.status === 'rejected';
    if (metricsDown && analyticsDown) {
      console.error('Core Analytics Sync Failed: Backend unreachable or returning 5xx.');
      setApiError(t('common.error_fetch', 'Failed to fetch latest data from backend.'));
    } else {
      setApiError(null);
    }
    setLoading(false);
  }, [t, clearLocalSession]);

  useEffect(() => {
    if (user) {
      fetchData();
      const interval = setInterval(fetchData, 12000);
      return () => clearInterval(interval);
    }
  }, [user, fetchData]);

  const handleLogout = useCallback(async () => {
    await api.logout();
    setUser(null);
    localStorage.removeItem('gorenel_user');
    setIsAuthStarted(false);
  }, []);

  const path = window.location.pathname;
  if (path.startsWith('/share/')) {
    const shareId = path.split('/')[2];
    return <ShareView shareId={shareId} />;
  }

  // Handle OAuth callback redirect (/dashboard?login=success)
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('login') === 'success' && !user) {
      // OAuth just completed, check session
      api.getMe().then((data) => {
        if (data?.user) {
          setUser(data.user);
          localStorage.setItem('gorenel_user', JSON.stringify(data.user));
          setIsAuthStarted(true);
          // Clean URL
          window.history.replaceState({}, '', '/');
        }
      }).catch(() => {});
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  if (!user || !isAuthStarted) {
    if (!isAuthStarted) {
      return (
        <LandingPage 
          onLogin={() => { setAuthMode('login'); setIsAuthStarted(true); }} 
          isLoggedIn={!!user}
          onGoToDashboard={() => setIsAuthStarted(true)}
        />
      );
    }

    if (authMode === 'register') {
      return (
        <RegisterPage 
          onSwitchToLogin={() => setAuthMode('login')} 
          onRegisterSuccess={(u) => { setUser(u); localStorage.setItem('gorenel_user', JSON.stringify(u)); setIsAuthStarted(true); }} 
        />
      );
    }
    return (
      <LoginPage 
        onSwitchToRegister={() => setAuthMode('register')} 
        onLoginSuccess={(u) => { setUser(u); localStorage.setItem('gorenel_user', JSON.stringify(u)); setIsAuthStarted(true); }} 
      />
    );
  }

  if (loading) return (
    <div className="min-h-screen bg-[#080a10] flex items-center justify-center">
      <div className="flex flex-col items-center gap-4">
        <Activity className="w-8 h-8 text-emerald-500/70 animate-pulse" />
        <div className="h-1 w-24 bg-white/[0.06] rounded-full overflow-hidden">
          <div className="h-full w-1/2 bg-emerald-500/50 rounded-full animate-shimmer" />
        </div>
      </div>
    </div>
  );

  const NavItem = ({ id, icon: Icon, label }: { id: NavTab, icon: any, label: string }) => (
    <button
      onClick={() => {
        setActiveTab(id);
        setIsMobileMenuOpen(false);
      }}
      className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-200 group ${activeTab === id
        ? 'bg-white/[0.06] text-white'
        : 'text-white/40 hover:text-white/70 hover:bg-white/[0.03]'
        }`}
    >
      <div className={`p-1.5 rounded-lg transition-colors ${activeTab === id ? 'bg-emerald-500/15 text-emerald-400' : 'text-white/30 group-hover:text-white/50'}`}>
        <Icon size={16} />
      </div>
      <div className="flex flex-col items-start overflow-hidden text-left">
        <span className="font-medium text-[13px] leading-tight">{t(`common.${id}`, label)}</span>
      </div>
      {activeTab === id && <ChevronRight className="w-3 h-3 ml-auto opacity-40 shrink-0" />}
    </button>
  );

  const navSections = (
    <>
      <div className="space-y-0.5">
        <div className="px-3 mb-2">
          <span className="text-[11px] font-medium text-white/20 uppercase tracking-wider">Platform</span>
        </div>
        <NavItem id="overview" icon={LayoutDashboard} label="Overview" />
        <NavItem id="tunnels" icon={Globe} label="Tunnels" />
        <NavItem id="ai_gateway" icon={Cpu} label="AI Gateway" />
      </div>
      <div className="space-y-0.5">
        <div className="px-3 mb-2">
          <span className="text-[11px] font-medium text-white/20 uppercase tracking-wider">Developers</span>
        </div>
        <NavItem id="traffic" icon={Microscope} label="Inspector" />
        <NavItem id="api_keys" icon={ShieldCheck} label="API Keys" />
        <NavItem id="reservations" icon={Globe} label="Reservations" />
        <NavItem id="settings" icon={Settings} label="Rules" />
      </div>
    </>
  );

  return (
    <div className="min-h-screen flex text-white font-sans">
      {/* Ambient gradients */}
      <div className="fixed inset-0 z-0 pointer-events-none">
        <div className="absolute top-0 left-[15%] w-[400px] h-[400px] bg-emerald-500/[0.03] rounded-full blur-[120px] animate-pulse-slow" />
        <div className="absolute bottom-0 right-[15%] w-[400px] h-[400px] bg-blue-500/[0.02] rounded-full blur-[120px] animate-pulse-slow" />
      </div>

      {/* Desktop Sidebar */}
      <aside className="w-60 z-20 p-3 hidden md:block shrink-0">
        <div className="sticky top-3 h-[calc(100vh-1.5rem)]">
          <div className="h-full bg-[#0c0e14]/80 backdrop-blur-xl border border-white/[0.04] rounded-2xl flex flex-col p-4">
            <div className="flex items-center gap-2.5 mb-8 px-1">
              <div className="w-8 h-8 rounded-lg overflow-hidden border border-white/[0.08] shadow-glow-emerald">
                <img src="/logo.png" alt="Gorenel Logo" className="w-full h-full object-cover" />
              </div>
              <span className="font-semibold text-[15px] tracking-tight">GORENEL</span>
            </div>

            <nav className="flex-1 space-y-6 overflow-y-auto">
              {navSections}
            </nav>

            <div className="mt-4 pt-4 border-t border-white/[0.04] space-y-3 shrink-0">
              <div className="flex items-center gap-2 text-[11px] text-white/25 px-2">
                <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                EU-Central-1 &middot; v1.0.0
              </div>

              <button
                onClick={handleLogout}
                className="w-full flex items-center gap-2.5 px-3 py-2.5 rounded-xl text-white/35 hover:text-white/70 hover:bg-white/[0.04] transition-all duration-200 text-sm group"
              >
                <div className="p-1.5 rounded-lg bg-white/[0.03] group-hover:bg-rose-500/10 group-hover:text-rose-400 transition-colors">
                  <LogOut className="w-3.5 h-3.5" />
                </div>
                {t('common.sign_out')}
              </button>
            </div>
          </div>
        </div>
      </aside>
      
      {/* Mobile overlay */}
      {isMobileMenuOpen && (
        <div 
          className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm md:hidden animate-fade-in"
          onClick={() => setIsMobileMenuOpen(false)}
        />
      )}

      {/* Mobile sidebar */}
      <aside
        className={`fixed inset-y-0 left-0 z-50 w-64 bg-[#0c0e14] border-r border-white/[0.04] p-5 transition-transform duration-300 ease-out md:hidden ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full'}`}
        role="dialog"
        aria-modal="true"
        aria-label={t('common.mobile_navigation', 'Mobile navigation')}
      >
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center gap-2.5">
            <div className="w-7 h-7 rounded-lg overflow-hidden border border-white/[0.08]">
              <img src="/logo.png" alt="Gorenel Logo" className="w-full h-full object-cover" />
            </div>
            <span className="font-semibold tracking-tight">GORENEL</span>
          </div>
          <button 
            onClick={() => setIsMobileMenuOpen(false)}
            className="p-1.5 hover:bg-white/[0.05] rounded-lg text-white/40 hover:text-white transition-colors"
            aria-label={t('common.close_menu', 'Close menu')}
          >
            <X size={18} />
          </button>
        </div>

        <nav className="flex-1 space-y-6 overflow-y-auto my-4">
          {navSections}
        </nav>

        <div className="mt-auto pt-4 border-t border-white/[0.04]">
          <button
            onClick={handleLogout}
            className="w-full flex items-center gap-2.5 px-3 py-2.5 rounded-xl text-white/35 hover:text-white/70 hover:bg-white/[0.04] transition-all text-sm group"
          >
            <div className="p-1.5 rounded-lg bg-white/[0.03] group-hover:bg-rose-500/10 group-hover:text-rose-400 transition-colors">
              <LogOut className="w-3.5 h-3.5" />
            </div>
            {t('common.sign_out')}
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 relative z-10 min-w-0">
        <div className="p-5 md:p-8 lg:p-10 max-w-[1440px] mx-auto space-y-8">

          {/* Header */}
          <header className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="flex items-center justify-between gap-4">
              <div className="min-w-0">
                <h2 className="text-xl md:text-2xl font-semibold tracking-tight text-white">
                  {activeTab === 'overview' && t('dashboard.command_center')}
                  {activeTab === 'tunnels' && t('dashboard.active_tunnels')}
                  {activeTab === 'ai_gateway' && t('dashboard.ai_gateway')}
                  {activeTab === 'traffic' && t('dashboard.traffic_inspector')}
                  {activeTab === 'api_keys' && t('dashboard.api_keys')}
                  {activeTab === 'reservations' && t('dashboard.reservations', 'Reservations')}
                  {activeTab === 'settings' && t('dashboard.global_rules')}
                </h2>
                <p className="text-sm text-white/35 mt-0.5 hidden sm:block">
                  {activeTab === 'overview' && t('dashboard.overview_desc')}
                  {activeTab === 'tunnels' && t('dashboard.tunnels_desc')}
                  {activeTab === 'ai_gateway' && t('dashboard.ai_desc')}
                  {activeTab === 'traffic' && t('dashboard.traffic_desc')}
                  {activeTab === 'api_keys' && t('dashboard.keys_desc')}
                  {activeTab === 'reservations' && t('dashboard.reservations_desc', 'Reserve stable subdomains and bind them to API keys.')}
                  {activeTab === 'settings' && t('dashboard.rules_desc')}
                </p>
              </div>

              <button 
                onClick={() => setIsMobileMenuOpen(true)}
                className="md:hidden p-2 bg-white/[0.04] border border-white/[0.08] rounded-lg hover:bg-white/[0.07] transition-all text-white/60"
                aria-label={t('common.open_menu', 'Open menu')}
              >
                <Menu size={18} />
              </button>
            </div>

            <div className="flex items-center gap-2.5">
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={() => {
                  const newLang = i18n.language === 'en' ? 'tr' : 'en';
                  i18n.changeLanguage(newLang);
                }}
              >
                <Languages size={13} className="text-emerald-400/70" />
                {i18n.language.toUpperCase()}
              </Button>
              
              <Button type="button" variant="primary" size="sm" onClick={() => setIsConnectOpen(true)}>
                <Zap size={14} />
                {t('common.connect')}
              </Button>
            </div>
          </header>

          {apiError && (
            <div className="rounded-xl border border-rose-500/20 bg-rose-500/[0.06] px-4 py-3 text-sm text-rose-300 flex items-center gap-2" role="alert">
              <div className="w-1.5 h-1.5 rounded-full bg-rose-400 animate-pulse" />
              {apiError}
            </div>
          )}

          <Suspense fallback={<div className="h-64 flex items-center justify-center"><Activity className="w-6 h-6 text-emerald-500/50 animate-spin" /></div>}>
            {activeTab === 'overview' && (
              <div className="space-y-6 animate-fade-in-up">
                {/* Metrics */}
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                  <MetricCard title={t('dashboard.active_tunnels_metric', 'Active Tunnels')} value={metrics?.tunnels.active_count || 0} icon={Globe} color="emerald" />
                  <MetricCard title={t('dashboard.total_requests_metric', 'Total Requests')} value={analytics?.total_requests || 0} icon={TrendingUp} color="blue" />
                  <MetricCard title={t('dashboard.system_load_metric', 'System Load')} value={`${metrics?.system.goroutines || 0}`} icon={Activity} color="violet" />
                  <MetricCard title={t('dashboard.avg_latency_metric', 'Avg Latency')} value={`${((analytics?.avg_response_time_ms ?? 0) / 1000000).toFixed(0)} ms`} icon={Activity} color="rose" />
                </div>

                {/* Quick Start */}
                <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-7 md:p-8 relative overflow-hidden">
                  <div className="absolute top-0 right-0 p-8 opacity-[0.03] pointer-events-none">
                    <Zap className="w-48 h-48 text-emerald-500" />
                  </div>
                  
                  <div className="relative z-10 space-y-6">
                    <div>
                      <h3 className="text-lg md:text-xl font-semibold mb-1.5">{t('dashboard.quick_start')}</h3>
                      <p className="text-sm text-white/40 max-w-xl">{t('landing.description')}</p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
                      {[1, 2, 3].map((step) => (
                        <div key={step} className="space-y-3">
                          <div className="w-9 h-9 rounded-lg bg-white/[0.04] border border-white/[0.06] flex items-center justify-center font-semibold text-sm text-emerald-400">
                            {step}
                          </div>
                          <div>
                            <h4 className="font-medium text-white/90 text-sm mb-0.5">{t(`dashboard.step_${step}`)}</h4>
                            <p className="text-xs text-white/35 leading-relaxed">{t(`dashboard.step_${step}_desc`)}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                {/* Charts */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                  <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-5">
                    <RealtimeChart data={analytics?.time_series || []} metric="requests" title={t('dashboard.global_requests_chart', 'Global Requests / Sec')} color="#10b981" />
                  </div>
                  <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-5">
                    <RealtimeChart data={analytics?.time_series || []} metric="avg_latency_ms" title={t('dashboard.latency_chart', 'P95 Latency (ms)')} color="#60a5fa" />
                  </div>
                </div>

                {/* Bottom Section */}
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
                  <div className="lg:col-span-2">
                    <GeoMap data={analytics?.top_countries || []} />
                  </div>
                  <div className="space-y-4">
                    <AnomalyAlerts anomalies={anomalies} />

                    <div className="p-5 rounded-2xl bg-emerald-500/[0.04] border border-emerald-500/[0.08] relative overflow-hidden">
                      <div className="absolute inset-0 bg-emerald-500/[0.02] blur-2xl" />
                      <div className="relative z-10">
                        <ShieldCheck className="w-8 h-8 text-emerald-400/70 mb-3" />
                        <h3 className="text-base font-semibold text-white mb-1">{t('dashboard.system_secure')}</h3>
                        <p className="text-sm text-white/40 leading-relaxed">
                          {t('dashboard.no_threats')}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'tunnels' && (
              <div className="animate-fade-in-up">
                <TunnelsList
                  tunnels={tunnels}
                  historySessions={tunnelHistory}
                  onOpenConnect={() => setIsConnectOpen(true)}
                  onGoReservations={() => setActiveTab('reservations')}
                />
              </div>
            )}

            {activeTab === 'ai_gateway' && (
              <div className="space-y-6 animate-fade-in-up">
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
                  <div className="lg:col-span-2">
                    <ModelComparison ml={mlStats} />
                  </div>
                  <div className="lg:col-span-1">
                    {anomalies.length > 0 ? (
                      <ThreatRadar latestAnomaly={anomalies[0]} />
                    ) : (
                      <ThreatRadar />
                    )}
                  </div>
                </div>
                <div className="p-6 rounded-2xl bg-white/[0.02] border border-white/[0.06]">
                  <h3 className="font-semibold text-sm mb-4">{t('common.provider_status', 'Provider Status')}</h3>
                  <div className="space-y-2">
                    {['OpenAI', 'Anthropic', 'Local-Llama'].map(p => (
                      <div key={p} className="flex items-center justify-between p-3.5 rounded-xl bg-white/[0.02] border border-white/[0.04] hover:border-white/[0.08] transition-colors">
                        <span className="text-sm font-medium text-white/80">{p}</span>
                        <div className="flex items-center gap-2">
                          <span className="text-[11px] text-emerald-400/80 font-medium">{t('common.operational')}</span>
                          <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_6px_rgba(16,185,129,0.6)]" />
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'traffic' && (
              <div className="animate-fade-in-up rounded-2xl border border-white/[0.06] bg-white/[0.02] overflow-hidden">
                <TrafficInspector history={history} />
              </div>
            )}

            {activeTab === 'api_keys' && (
              <div className="animate-fade-in-up">
                <ApiKeyManager />
              </div>
            )}

            {activeTab === 'reservations' && (
              <div className="animate-fade-in-up">
                <Reservations />
              </div>
            )}

            {activeTab === 'settings' && (
              <div className="max-w-4xl animate-fade-in-up">
                <ModificationRules rules={rules} onRulesChange={fetchData} />
              </div>
            )}
          </Suspense>
        </div>
      </main>

      <ConnectModal isOpen={isConnectOpen} onClose={() => setIsConnectOpen(false)} apiKey={user?.api_key} />
    </div>
  );
}

export default App;
