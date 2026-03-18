import React, { Suspense, useEffect, useState } from 'react';
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
  X
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { api, type Metrics, type AnalyticsSnapshot, type AnomalyRecord, type ModelStatsResponse, type CapturedRequest, type ModificationRule } from './api/client';
import './index.css';

// Lazy load components
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

type NavTab = 'overview' | 'tunnels' | 'ai_gateway' | 'traffic' | 'settings' | 'api_keys';

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
  const [mlStats, setMlStats] = useState<ModelStatsResponse>({});
  const [history, setHistory] = useState<CapturedRequest[]>([]);
  const [rules, setRules] = useState<ModificationRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  useEffect(() => {
    const checkSession = async () => {
      const storedUser = localStorage.getItem('gorenel_user');
      if (storedUser) {
        setUser(JSON.parse(storedUser));
        setLoading(false);
        return;
      }
      try {
        const data = await api.getMe();
        if (data && data.user) {
          setUser(data.user);
          localStorage.setItem('gorenel_user', JSON.stringify(data.user));
        }
      } catch (err) {
        console.log('No active session');
      } finally {
        setLoading(false);
      }
    };
    checkSession();
  }, []);

  const fetchData = async () => {
    try {
      const [metricsData, analyticsData, tunnelsData, anomaliesData, mlStatsData, historyData, rulesData] = await Promise.all([
        api.getMetrics(),
        api.getAnalytics(),
        api.getTunnels(),
        api.getAnomalies(),
        api.getMLStats(),
        api.getTrafficHistory(),
        api.getModificationRules(),
      ]);
      setMetrics(metricsData);
      setAnalytics(analyticsData);
      setTunnels(tunnelsData.tunnels || []);
      setAnomalies(anomaliesData.anomalies || []);
      setMlStats(mlStatsData || {});
      setHistory(historyData || []);
      setRules(rulesData || []);
      setLoading(false);
    } catch (err) {
      console.error('Core Analytics Sync Failed: Backend unreachable or returning 502.');
      setLoading(false);
    }
  };

  useEffect(() => {
    if (user) {
      fetchData();
      const interval = setInterval(fetchData, 5000);
      return () => clearInterval(interval);
    }
  }, [user]);

  const handleLogout = async () => {
    await api.logout();
    setUser(null);
    localStorage.removeItem('gorenel_user');
  };

  // Handle public share links
  const path = window.location.pathname;
  if (path.startsWith('/share/')) {
    const shareId = path.split('/')[2];
    return <ShareView shareId={shareId} />;
  }

  if (!user) {
    if (!isAuthStarted) {
      return <LandingPage onLogin={() => setIsAuthStarted(true)} />;
    }

    if (authMode === 'register') {
      return (
        <RegisterPage 
          onSwitchToLogin={() => setAuthMode('login')} 
          onRegisterSuccess={(u) => { setUser(u); localStorage.setItem('gorenel_user', JSON.stringify(u)); }} 
        />
      );
    }
    return (
      <LoginPage 
        onSwitchToRegister={() => setAuthMode('register')} 
        onLoginSuccess={(u) => { setUser(u); localStorage.setItem('gorenel_user', JSON.stringify(u)); }} 
      />
    );
  }
  if (loading) return (
    <div className="min-h-screen bg-black flex items-center justify-center">
      <div className="flex flex-col items-center gap-4">
        <Activity className="w-12 h-12 text-primary animate-pulse" />
        <div className="h-1 w-32 bg-white/10 rounded-full overflow-hidden">
          <div className="h-full bg-primary animate-progress origin-left"></div>
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
      className={`w-full flex items-center gap-3 px-4 py-2.5 rounded-xl transition-all duration-300 group ${activeTab === id
        ? 'bg-emerald-500/10 text-emerald-400'
        : 'text-white/40 hover:text-white hover:bg-white/5'
        }`}
    >
      <Icon className={`w-4 h-4 ${activeTab === id ? 'text-emerald-400' : 'group-hover:scale-110 transition-transform'}`} />
      <span className="font-medium text-sm">{t(`common.${id}`, label)}</span>
      {activeTab === id && <ChevronRight className="w-3 h-3 ml-auto opacity-50" />}
    </button>
  );

  return (
    <div className="min-h-screen flex text-white selection:bg-emerald-500/30 font-sans">

      {/* Background Mesh Gradients (Fixed) */}
      <div className="fixed inset-0 z-0 pointer-events-none">
        <div className="absolute top-[20%] left-[10%] w-[30%] h-[30%] bg-emerald-500/5 rounded-full blur-[100px] animate-pulse-slow" />
        <div className="absolute bottom-[20%] right-[10%] w-[30%] h-[30%] bg-blue-600/5 rounded-full blur-[100px] animate-pulse-slow delay-1000" />
      </div>

      {/* Floating Sidebar - Sticky */}
      <aside className="w-64 z-20 p-4 hidden md:block shrink-0">
        <div className="sticky top-4 h-[calc(100vh-2rem)]">
          <div className="h-full bg-[#0A0C10]/60 backdrop-blur-xl border border-white/5 rounded-3xl flex flex-col p-6 shadow-2xl">
            <div className="flex items-center gap-3 mb-10">
              <div className="w-10 h-10 rounded-xl overflow-hidden border border-white/10 shadow-[0_0_20px_rgba(16,185,129,0.2)]">
                <img src="/logo.png" alt="Gorenel Logo" className="w-full h-full object-cover" />
              </div>
              <div className="flex flex-col">
                <span className="font-bold text-lg tracking-tight">GORENEL</span>
              </div>
            </div>

            <nav className="flex-1 space-y-6">
              <div className="space-y-1">
                <div className="px-4 mb-2">
                  <span className="text-[10px] font-bold text-white/30 uppercase tracking-widest">Platform</span>
                </div>
                <NavItem id="overview" icon={LayoutDashboard} label="Overview" />
                <NavItem id="tunnels" icon={Globe} label="Tunnels" />
                <NavItem id="ai_gateway" icon={Cpu} label="AI Gateway" />
              </div>

              <div className="space-y-1">
                <div className="px-4 mb-2">
                  <span className="text-[10px] font-bold text-white/30 uppercase tracking-widest">Developers</span>
                </div>
                <NavItem id="traffic" icon={Microscope} label="Inspector" />
                <NavItem id="api_keys" icon={ShieldCheck} label="API Keys" />
                <NavItem id="settings" icon={Settings} label="Rules" />
              </div>
            </nav>

            <div className="mt-auto pt-6 border-t border-white/5 space-y-4">
              <div className="flex items-center gap-2 text-xs text-white/40 px-2">
                <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                EU-Central-1 • v1.0.0
              </div>

              <button
                onClick={handleLogout}
                className="w-full flex items-center gap-3 px-4 py-2.5 rounded-xl text-white/40 hover:text-white hover:bg-white/5 transition-all text-sm font-medium"
              >
                <LogOut className="w-4 h-4" />
                {t('common.sign_out')}
              </button>
            </div>
          </div>
        </div>
      </aside>
      
      {/* Mobile Menu Overlay */}
      {isMobileMenuOpen && (
        <div 
          className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm md:hidden"
          onClick={() => setIsMobileMenuOpen(false)}
        />
      )}

      {/* Mobile Sidebar (Slide-over) */}
      <aside className={`fixed inset-y-0 left-0 z-50 w-72 bg-[#0A0C10] border-r border-white/5 p-6 transition-transform duration-300 md:hidden ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full'}`}>
        <div className="flex items-center justify-between mb-10">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg overflow-hidden border border-white/10 shadow-[0_0_20px_rgba(16,185,129,0.2)]">
              <img src="/logo.png" alt="Gorenel Logo" className="w-full h-full object-cover" />
            </div>
            <span className="font-bold text-lg tracking-tight">GORENEL</span>
          </div>
          <button 
            onClick={() => setIsMobileMenuOpen(false)}
            className="p-2 hover:bg-white/5 rounded-lg text-white/40 hover:text-white transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <nav className="flex-1 space-y-6">
          <div className="space-y-1">
            <div className="px-4 mb-2">
              <span className="text-[10px] font-bold text-white/30 uppercase tracking-widest">Platform</span>
            </div>
            <NavItem id="overview" icon={LayoutDashboard} label="Overview" />
            <NavItem id="tunnels" icon={Globe} label="Tunnels" />
            <NavItem id="ai_gateway" icon={Cpu} label="AI Gateway" />
          </div>

          <div className="space-y-1">
            <div className="px-4 mb-2">
              <span className="text-[10px] font-bold text-white/30 uppercase tracking-widest">Developers</span>
            </div>
            <NavItem id="traffic" icon={Microscope} label="Inspector" />
            <NavItem id="api_keys" icon={ShieldCheck} label="API Keys" />
            <NavItem id="settings" icon={Settings} label="Rules" />
          </div>
        </nav>

        <div className="mt-auto pt-6 border-t border-white/5 space-y-4 fixed bottom-6 left-6 right-6">
          <button
            onClick={handleLogout}
            className="w-full flex items-center gap-3 px-4 py-2.5 rounded-xl text-white/40 hover:text-white hover:bg-white/5 transition-all text-sm font-medium"
          >
            <LogOut className="w-4 h-4" />
            {t('common.sign_out')}
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 relative z-10 min-w-0">
        <div className="p-6 md:p-8 lg:p-12 max-w-[1600px] mx-auto space-y-10">

          <header className="flex flex-col md:flex-row md:items-center justify-between gap-6 pb-4">
            <div className="flex items-center justify-between">
              <div className="animate-in slide-in-from-bottom-2 duration-500">
                <h2 className="text-2xl md:text-4xl font-bold tracking-tight text-white mb-2">
                  {activeTab === 'overview' && t('dashboard.command_center')}
                  {activeTab === 'tunnels' && t('dashboard.active_tunnels')}
                  {activeTab === 'ai_gateway' && t('dashboard.ai_gateway')}
                  {activeTab === 'traffic' && t('dashboard.traffic_inspector')}
                  {activeTab === 'api_keys' && t('dashboard.api_keys')}
                  {activeTab === 'settings' && t('dashboard.global_rules')}
                </h2>
                <p className="text-sm md:text-lg text-white/50 font-normal max-w-2xl hidden xs:block">
                  {activeTab === 'overview' && t('dashboard.overview_desc')}
                  {activeTab === 'tunnels' && t('dashboard.tunnels_desc')}
                  {activeTab === 'ai_gateway' && t('dashboard.ai_desc')}
                  {activeTab === 'traffic' && t('dashboard.traffic_desc')}
                  {activeTab === 'api_keys' && t('dashboard.keys_desc')}
                  {activeTab === 'settings' && t('dashboard.rules_desc')}
                </p>
              </div>

              <button 
                onClick={() => setIsMobileMenuOpen(true)}
                className="md:hidden p-2 bg-white/5 border border-white/10 rounded-xl hover:bg-white/10 transition-all text-emerald-400"
              >
                <Menu size={20} />
              </button>
            </div>

            <div className="flex items-center gap-4">
              <button
                onClick={() => {
                  const newLang = i18n.language === 'en' ? 'tr' : 'en';
                  i18n.changeLanguage(newLang);
                }}
                className="flex items-center gap-2 px-4 py-2 bg-white/5 border border-white/10 rounded-xl text-xs font-bold hover:bg-white/10 transition-all uppercase"
              >
                <Languages size={14} className="text-emerald-400" />
                {i18n.language.toUpperCase()}
              </button>
              
              <button
                onClick={() => setIsConnectOpen(true)}
                className="btn-primary-premium"
              >
                <span className="text-lg mr-1">+</span> {t('common.connect')}
              </button>
            </div>
          </header>

          <Suspense fallback={<div className="h-96 flex items-center justify-center"><Activity className="w-8 h-8 text-emerald-500 animate-spin" /></div>}>
            {activeTab === 'overview' && (
              <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">

                {/* Metrics Grid - Airy */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                  <MetricCard title="Active Tunnels" value={metrics?.tunnels.active_count || 0} icon={Globe} color="emerald" />
                  <MetricCard title="Total Requests" value={analytics?.total_requests || 0} icon={TrendingUp} color="blue" trend={{ value: 12, isPositive: true }} />
                  <MetricCard title="System Load" value={`${metrics?.system.goroutines || 0}`} icon={Activity} color="violet" />
                  <MetricCard title="Avg Latency" value={`${((analytics?.avg_response_time_ms ?? 0) / 1000000).toFixed(0)} ms`} icon={Activity} color="rose" />
                </div>

                {/* Charts - Floating */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <RealtimeChart data={analytics?.time_series || []} metric="requests" title="Global Requests / Sec" color="#10b981" />
                  <RealtimeChart data={analytics?.time_series || []} metric="avg_latency_ms" title="P95 Latency (ms)" color="#eff6ff" />
                </div>

                {/* Bottom Section */}
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                  <div className="lg:col-span-2">
                    <GeoMap data={analytics?.top_countries || []} />
                  </div>
                  <div className="space-y-6">
                    <AnomalyAlerts anomalies={anomalies} />

                    <div className="p-6 rounded-3xl bg-emerald-500/5 border border-emerald-500/10 flex flex-col justify-center h-full relative overflow-hidden">
                      <div className="absolute inset-0 bg-emerald-500/5 blur-3xl" />
                      <div className="relative z-10">
                        <ShieldCheck className="w-10 h-10 text-emerald-400 mb-4" />
                        <h3 className="text-xl font-bold text-white mb-2">System Secure</h3>
                        <p className="text-white/50 leading-relaxed">
                          Anomaly detection is active. No threats detected in the last 24 hours.
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'tunnels' && (
              <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
                <TunnelsList tunnels={tunnels} onOpenConnect={() => setIsConnectOpen(true)} />
              </div>
            )}

            {activeTab === 'ai_gateway' && (
              <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                  <div className="lg:col-span-2">
                    <ModelComparison stats={mlStats} />
                  </div>
                  <div className="lg:col-span-1">
                    {anomalies.length > 0 ? (
                      <ThreatRadar latestAnomaly={anomalies[0]} />
                    ) : (
                      <ThreatRadar />
                    )}
                  </div>
                </div>
                <div className="p-8 rounded-3xl bg-[#0A0C10]/40 border border-white/5 backdrop-blur-md">
                  <h3 className="font-bold text-lg mb-6">Provider Status</h3>
                  <div className="space-y-4">
                    {['OpenAI', 'Anthropic', 'Local-Llama'].map(p => (
                      <div key={p} className="flex items-center justify-between p-4 rounded-2xl bg-white/[0.02] border border-white/[0.02]">
                        <span className="font-medium">{p}</span>
                        <div className="flex items-center gap-2">
                          <span className="text-[10px] text-emerald-500 font-bold uppercase tracking-wider">Operational</span>
                          <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)]" />
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'traffic' && (
              <div className="animate-in fade-in slide-in-from-bottom-4 duration-500 bg-[#0A0C10]/40 backdrop-blur-md rounded-3xl border border-white/5 overflow-hidden">
                <TrafficInspector history={history} />
              </div>
            )}

            {activeTab === 'api_keys' && (
              <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
                <ApiKeyManager />
              </div>
            )}

            {activeTab === 'settings' && (
              <div className="max-w-4xl animate-in fade-in slide-in-from-bottom-4 duration-500">
                <ModificationRules rules={rules} onRulesChange={fetchData} />
              </div>
            )}
          </Suspense>
        </div>
      </main >

      <ConnectModal isOpen={isConnectOpen} onClose={() => setIsConnectOpen(false)} apiKey={user?.api_key} />
    </div >
  );
}

export default App;
