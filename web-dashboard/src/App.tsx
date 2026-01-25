import { useEffect, useState } from 'react';
import { Activity, Zap, Users, TrendingUp, AlertCircle } from 'lucide-react';
import { MetricCard } from './components/MetricCard';
import { RealtimeChart } from './components/RealtimeChart';
import { TunnelsList } from './components/TunnelsList';
import { GeoMap } from './components/GeoMap';
import { api, type Metrics, type AnalyticsSnapshot } from './api/client';
import './index.css';

// Mock tunnels data (replace with real data from your API)
const MOCK_TUNNELS = [
  {
    id: '1',
    subdomain: 'abc123.tunnel-project.xyz',
    localPort: 3000,
    publicUrl: 'https://abc123.tunnel-project.xyz',
    status: 'active' as const,
    requestCount: 1234,
    bandwidth: { in: 1024 * 1024 * 2.5, out: 1024 * 1024 * 5.2 },
    startedAt: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
    lastActivity: new Date().toISOString(),
  },
];

function App() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [analytics, setAnalytics] = useState<AnalyticsSnapshot | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch data on mount and set up polling
  useEffect(() => {
    const fetchData = async () => {
      try {
        setError(null);
        const [metricsData, analyticsData] = await Promise.all([
          api.getMetrics(),
          api.getAnalytics(),
        ]);
        setMetrics(metricsData);
        setAnalytics(analyticsData);
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch data. Make sure the server is running on port 9090.');
        setLoading(false);
        console.error('Error fetching data:', err);
      }
    };

    fetchData();
    
    // Poll every 5 seconds
    const interval = setInterval(fetchData, 5000);
    
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-dark-900 flex items-center justify-center">
        <div className="text-center">
          <Activity className="w-12 h-12 text-primary-400 animate-spin mx-auto mb-4" />
          <p className="text-dark-400">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-dark-900 flex items-center justify-center">
        <div className="text-center max-w-md">
          <AlertCircle className="w-12 h-12 text-red-400 mx-auto mb-4" />
          <p className="text-white text-lg mb-2">Connection Error</p>
          <p className="text-dark-400 text-sm">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-dark-900">
      {/* Header */}
      <header className="bg-dark-800 border-b border-dark-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-primary-500 to-cyan-500 rounded-lg flex items-center justify-center">
                <Zap className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-white">Gorenel Dashboard</h1>
                <p className="text-sm text-dark-400">Production Monitoring</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
              <span className="text-sm text-dark-300">Live</span>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Metrics Cards */}
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
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {analytics?.time_series && (
            <>
              <RealtimeChart
                data={analytics.time_series}
                metric="requests"
                title="Request Rate"
                color="#3b82f6"
              />
              <RealtimeChart
                data={analytics.time_series}
                metric="avg_latency_ms"
                title="Average Latency"
                color="#10b981"
              />
            </>
          )}
        </div>

        {/* Tunnels and Geo */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <TunnelsList tunnels={MOCK_TUNNELS} />
          {analytics?.top_countries && (
            <GeoMap data={analytics.top_countries} />
          )}
        </div>

        {/* Footer */}
        <div className="mt-8 text-center text-dark-500 text-sm">
          <p>Gorenel v1.0.0 • Powered by Go & React</p>
        </div>
      </main>
    </div>
  );
}

export default App;