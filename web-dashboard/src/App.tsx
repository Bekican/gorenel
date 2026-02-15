```
import React, { Suspense, useEffect, useState } from 'react';
import { 
  Zap, 
  Users, 
  TrendingUp, 
  AlertCircle, 
  LayoutDashboard, 
  Globe, 
  Microscope, 
  Settings, 
  LogOut, 
  Activity,
  ChevronRight,
  ShieldCheck,
  Cpu
} from 'lucide-react';
import { api, type Metrics, type AnalyticsSnapshot, type AnomalyRecord, type ModelStatsResponse, type CapturedRequest, type ModificationRule } from './api/client';
import { ErrorBoundary } from './components/ErrorBoundary';
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
import { LoginPage } from './components/LoginPage';
import { ConnectModal } from './components/ConnectModal';

type NavTab = 'overview' | 'tunnels' | 'ai_gateway' | 'traffic' | 'settings';

function App() {
  const [user, setUser] = useState<any>(null);
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
  const [error, setError] = useState<string | null>(null);

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
      setError('Connection refused. Is the Gorenel server running on port 9090?');
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

  if (!user) return <LoginPage onLoginSuccess={(u) => { setUser(u); localStorage.setItem('gorenel_user', JSON.stringify(u)); }} />;
  if (loading) return (
    <div className="min-h-screen bg-black flex items-center justify-center">
      <div className="flex flex-col items-center gap-4">
        <Zap className="w-12 h-12 text-primary animate-pulse" />
        <div className="h-1 w-32 bg-white/10 rounded-full overflow-hidden">
          <div className="h-full bg-primary animate-progress origin-left"></div>
        </div>
      </div>
    </div>
  );

  const NavItem = ({ id, icon: Icon, label }: { id: NavTab, icon: any, label: string }) => (
    <button
      onClick={() => setActiveTab(id)}
      className={`w - full flex items - center gap - 3 px - 4 py - 3 rounded - 2xl transition - all ${
  activeTab === id
    ? 'bg-primary/10 text-primary glow-emerald'
    : 'text-white/40 hover:text-white hover:bg-white/5'
} `}
    >
      <Icon className={`w - 5 h - 5 ${ activeTab === id ? 'text-primary' : '' } `} />
      <span className="font-bold text-sm tracking-tight">{label}</span>
      {activeTab === id && <ChevronRight className="w-4 h-4 ml-auto" />}
    </button>
  );

  return (
    <div className="min-h-screen bg-black flex text-white selection:bg-primary/30">
      {/* Sidebar */}
      <aside className="w-72 border-r border-white/5 flex flex-col sticky top-0 h-screen p-6 shrink-0 bg-[#050505]">
        <div className="flex items-center gap-3 mb-10 px-2">
          <div className="w-10 h-10 bg-primary rounded-xl flex items-center justify-center shadow-lg shadow-primary/20">
            <Zap className="w-6 h-6 text-black fill-current" />
          </div>
          <div className="flex flex-col">
            <span className="font-black text-xl tracking-tighter">GORENEL</span>
            <span className="text-[10px] text-white/40 font-black tracking-widest uppercase -mt-1">Cloud Gateway</span>
          </div>
        </div>

        <nav className="flex-1 space-y-2">
          <div className="px-4 mb-4">
            <span className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em]">Management</span>
          </div>
          <NavItem id="overview" icon={LayoutDashboard} label="Overview" />
          <NavItem id="tunnels" icon={Globe} label="Standard Tunnels" />
          <NavItem id="ai_gateway" icon={Cpu} label="AI Gateway" />
          
          <div className="mt-8 px-4 mb-4">
            <span className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em]">Developer Tools</span>
          </div>
          <NavItem id="traffic" icon={Microscope} label="Traffic Inspector" />
          <NavItem id="settings" icon={Settings} label="Global Rules" />
        </nav>

        <div className="mt-auto space-y-4">
          <div className="bg-white/5 rounded-3xl p-4 border border-white/5">
            <div className="flex items-center gap-3 mb-2">
              <div className="w-2 h-2 rounded-full bg-primary animate-pulse shadow-[0_0_8px_rgba(16,185,129,1)]" />
              <span className="text-xs font-bold text-white/50">System Online</span>
            </div>
            <p className="text-[10px] text-white/30 leading-relaxed font-medium">Node: EU-Central-1<br/>Agent Version: 1.0.0-gold</p>
          </div>
          
          <button 
            onClick={handleLogout}
            className="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-white/40 hover:text-red-400 hover:bg-red-400/10 transition-all group"
          >
            <LogOut className="w-5 h-5 group-hover:rotate-12 transition-transform" />
            <span className="font-bold text-sm">Sign Out</span>
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto p-12">
        {error && (
          <div className="mb-8 p-4 bg-red-500/10 border border-red-500/20 rounded-2xl flex items-center gap-3 text-red-500">
            <AlertCircle className="w-5 h-5" />
            <span className="text-sm font-bold">{error}</span>
          </div>
        )}

        <header className="flex items-center justify-between mb-12">
           <div>
              <h2 className="text-4xl font-black tracking-tight mb-2 text-gradient">
                {activeTab === 'overview' && "Command Center"}
                {activeTab === 'tunnels' && "Traffic Tunnels"}
                {activeTab === 'ai_gateway' && "AI Intelligence"}
                {activeTab === 'traffic' && "Deep Inspection"}
                {activeTab === 'settings' && "Modification Rules"}
              </h2>
              <p className="text-white/40 font-medium">
                {activeTab === 'overview' && "Real-time system integrity and global metrics."}
                {activeTab === 'tunnels' && "Manage and monitor your public endpoints."}
                {activeTab === 'ai_gateway' && "Unified interface for LLM proxying and caching."}
                {activeTab === 'traffic' && "Detailed request/response logs and replay tools."}
                {activeTab === 'settings' && "Define dynamic traffic manipulation logic."}
              </p>
           </div>
           
           <div className="flex items-center gap-4">
              <button 
                onClick={() => setIsConnectOpen(true)}
                className="btn-primary-premium"
              >
                Connect Localhost <ChevronRight className="w-4 h-4" />
              </button>
           </div>
        </header>

        <Suspense fallback={<div className="flex items-center justify-center p-20"><Activity className="w-8 h-8 text-primary animate-spin" /></div>}>
           {activeTab === 'overview' && (
              <div className="space-y-12">
                 <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                    <MetricCard title="Active Tunnels" value={metrics?.tunnels.active_count || 0} icon={Globe} color="emerald" />
                    <MetricCard title="Total Requests" value={analytics?.total_requests || 0} icon={TrendingUp} color="emerald" trend={{value: 12, isPositive: true}} />
                    <MetricCard title="System Load" value={metrics?.system.goroutines || 0} icon={Activity} color="emerald" />
                    <MetricCard title="Avg Latency" value={`${ ((analytics?.avg_response_time_ms ?? 0) / 1000000).toFixed(0) } ms`} icon={Zap} color="emerald" />
                 </div>

                 <div className="grid grid-cols-2 gap-8">
                    <RealtimeChart data={analytics?.time_series || []} metric="requests" title="Global Requests / Sec" color="#10b981" />
                    <RealtimeChart data={analytics?.time_series || []} metric="avg_latency_ms" title="P95 Response Latency" color="#8b5cf6" />
                 </div>

                 <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                    <div className="lg:col-span-2">
                       <GeoMap data={analytics?.top_countries || []} />
                    </div>
                    <div className="space-y-8">
                       <AnomalyAlerts anomalies={anomalies} />
                       <div className="card bg-gradient-to-br from-primary/10 to-transparent">
                          <div className="flex items-center gap-3 mb-4">
                             <ShieldCheck className="w-6 h-6 text-primary" />
                             <h3 className="font-black">Safe Shield Active</h3>
                          </div>
                          <p className="text-xs text-white/50 leading-relaxed font-medium">Automatic anomaly detection is protecting 100% of your incoming traffic using a distributed consensus model.</p>
                       </div>
                    </div>
                 </div>
              </div>
           )}

           {activeTab === 'tunnels' && (
              <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
                 <TunnelsList tunnels={tunnels} onOpenConnect={() => setIsConnectOpen(true)} />
              </div>
           )}

           {activeTab === 'ai_gateway' && (
              <div className="space-y-12 animate-in fade-in slide-in-from-bottom-4 duration-500">
                  <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                      <div className="lg:col-span-2">
                         <ModelComparison stats={mlStats} />
                      </div>
                      <div className="card space-y-6">
                         <h3 className="font-black text-xl">Provider Heath</h3>
                         <div className="space-y-4">
                            {['OpenAI', 'Anthropic', 'Local-Llama'].map(p => (
                               <div key={p} className="flex items-center justify-between p-3 bg-white/5 rounded-2xl border border-white/5">
                                  <span className="font-bold text-sm">{p}</span>
                                  <div className="flex items-center gap-2">
                                     <span className="text-[10px] text-primary font-black uppercase tracking-widest">Active</span>
                                     <div className="w-1.5 h-1.5 rounded-full bg-primary shadow-[0_0_4px_rgba(16,185,129,1)]" />
                                  </div>
                               </div>
                            ))}
                         </div>
                      </div>
                  </div>
              </div>
           )}

           {activeTab === 'traffic' && (
              <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
                 <TrafficInspector history={history} />
              </div>
           )}

           {activeTab === 'settings' && (
              <div className="max-w-4xl animate-in fade-in slide-in-from-bottom-4 duration-500">
                 <ModificationRules rules={rules} onRulesChange={fetchData} />
              </div>
           )}
        </Suspense>

        <ConnectModal isOpen={isConnectOpen} onClose={() => setIsConnectOpen(false)} />
      </main>
    </div>
  );
}

export default App;
```