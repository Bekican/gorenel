import React, { Suspense, useEffect, useState } from 'react';
import { Activity, Zap, Users, TrendingUp, AlertCircle } from 'lucide-react';
import { api, type Metrics, type AnalyticsSnapshot, type AnomalyRecord, type ModelStatsResponse } from './api/client';
import { ErrorBoundary } from './components/ErrorBoundary';
import './index.css';

// Lazy load components
const MetricCard = React.lazy(() => import('./components/MetricCard').then(module => ({ default: module.MetricCard })));
const RealtimeChart = React.lazy(() => import('./components/RealtimeChart').then(module => ({ default: module.RealtimeChart })));
const TunnelsList = React.lazy(() => import('./components/TunnelsList').then(module => ({ default: module.TunnelsList })));
const GeoMap = React.lazy(() => import('./components/GeoMap').then(module => ({ default: module.GeoMap })));
const AnomalyAlerts = React.lazy(() => import('./components/AnomalyAlerts').then(module => ({ default: module.AnomalyAlerts })));
const ModelComparison = React.lazy(() => import('./components/ModelComparison').then(module => ({ default: module.ModelComparison })));
import { LoginPage } from './components/LoginPage';
import { LogOut } from 'lucide-react';

// Main App Component
function App() {
  const [user, setUser] = useState<any>(null);
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [analytics, setAnalytics] = useState<AnalyticsSnapshot | null>(null);
  const [tunnels, setTunnels] = useState<any[]>([]);
  const [anomalies, setAnomalies] = useState<AnomalyRecord[]>([]);
  const [mlStats, setMlStats] = useState<ModelStatsResponse>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const checkSession = async () => {
      // 1. Check LocalStorage first
      const storedUser = localStorage.getItem('gorenel_user');
      if (storedUser) {
        setUser(JSON.parse(storedUser));
        setLoading(false);
        return;
      }

      // 2. If nothing in LocalStorage, try API (for handling redirects)
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

  const handleLoginSuccess = (userData: any) => {
    setUser(userData);
    localStorage.setItem('gorenel_user', JSON.stringify(userData));
    setLoading(true); // Trigger re-fetch for dashboard data
  };

  const handleLogout = async () => {
    await api.logout();
    setUser(null);
    localStorage.removeItem('gorenel_user');
  };

  // Fetch data on mount and set up polling
  useEffect(() => {
    const fetchData = async () => {
      try {
        setError(null);
        const [metricsData, analyticsData, tunnelsData, anomaliesData, mlStatsData] = await Promise.all([
          api.getMetrics(),
          api.getAnalytics(),
          api.getTunnels(),
          api.getAnomalies(),
          api.getMLStats(),
        ]);
        setMetrics(metricsData);
        setAnalytics(analyticsData);
        setTunnels(tunnelsData.tunnels || []);
        setAnomalies(anomaliesData.anomalies || []);
        setMlStats(mlStatsData || {});
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch data. Make sure the server is running on port 9090.');
        setLoading(false);
        console.error('Error fetching data:', err);
      }
    };

    fetchData();

    // Poll every 5 seconds if logged in
    let interval: any;
    if (user) {
      interval = setInterval(fetchData, 5000);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [user]);

  if (!user) {
    return <LoginPage onLoginSuccess={handleLoginSuccess} />;
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-neutral-50 flex items-center justify-center">
        <div className="text-center">
          <Activity className="w-12 h-12 text-primary-500 animate-spin mx-auto mb-4" />
          <p className="text-neutral-500">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-neutral-50 flex items-center justify-center">
        <div className="text-center max-w-md">
          <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <p className="text-neutral-900 text-lg mb-2">Connection Error</p>
          <p className="text-neutral-500 text-sm">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-neutral-50 font-sans text-neutral-900">
      {/* Header */}
      <header className="bg-white border-b border-neutral-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-primary-500 to-primary-600 rounded-xl shadow-sm flex items-center justify-center">
                <Zap className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-neutral-900 tracking-tight">Gorenel Dashboard</h1>
                <p className="text-sm text-neutral-500">Production Monitoring</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2 bg-green-50 px-3 py-1.5 rounded-full border border-green-100">
                <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                <span className="text-sm font-medium text-green-700">Live System</span>
              </div>
              <button
                onClick={handleLogout}
                className="p-2 text-neutral-400 hover:text-red-500 hover:bg-red-50 rounded-xl transition-all"
                title="Çıkış Yap"
              >
                <LogOut className="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Metrics Cards */}
        <Suspense fallback={
          <div className="flex justify-center p-8">
            <Activity className="w-8 h-8 text-primary-500 animate-spin" />
          </div>
        }>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            <MetricCard
              title="Active Tunnels"
              value={metrics?.tunnels.active_count || 0}
              subtitle="Currently running"
              icon={Activity}
              color="blue"
            />
            <MetricCard
              title="Total Requests"
              value={(analytics?.total_requests || 0).toLocaleString()}
              subtitle="All time"
              icon={TrendingUp}
              color="green"
              trend={{ value: 12.5, isPositive: true }}
            />
            <MetricCard
              title="Active Connections"
              value={metrics?.requests.active_connections || 0}
              subtitle="Current connections"
              icon={Users}
              color="purple"
            />
            <MetricCard
              title="Avg Response Time"
              value={`${((analytics?.avg_response_time_ms ?? 0) / 1000000).toFixed(2)}ms`}
              subtitle="P95 latency"
              icon={Zap}
              color="orange"
            />
          </div>

          {/* Charts Row */}
          <ErrorBoundary>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8 animate-fade-in-up" style={{ animationDelay: '200ms', animationFillMode: 'both' }}>
              {analytics?.time_series && (
                <>
                  <RealtimeChart
                    data={analytics.time_series}
                    metric="requests"
                    title="Request Rate"
                    color="#10b981"
                  />
                  <RealtimeChart
                    data={analytics.time_series}
                    metric="avg_latency_ms"
                    title="Average Latency"
                    color="#f59e0b"
                  />
                </>
              )}
            </div>
          </ErrorBoundary>

          {/* Tunnels and Geo */}
          <ErrorBoundary>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8 animate-fade-in-up" style={{ animationDelay: '400ms', animationFillMode: 'both' }}>
              <TunnelsList tunnels={tunnels} />
              {analytics?.top_countries && (
                <GeoMap data={analytics.top_countries} />
              )}
            </div>
          </ErrorBoundary>

          {/* Security & ML Models */}
          <ErrorBoundary>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8 animate-fade-in-up" style={{ animationDelay: '600ms', animationFillMode: 'both' }}>
              <ModelComparison stats={mlStats} />
              <AnomalyAlerts anomalies={anomalies} />
            </div>
          </ErrorBoundary>

          {/* Footer */}
          <div className="mt-12 border-t border-neutral-200 pt-8 text-center text-neutral-400 text-sm">
            <p>Gorenel v1.0.0 • Powered by Go & React</p>
          </div>
        </Suspense>
      </main>
    </div>
  );
}

export default App;