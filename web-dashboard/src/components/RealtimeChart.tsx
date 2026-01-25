import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { TrendingUp } from 'lucide-react';

interface TimeSeriesData {
  timestamp: string;
  requests: number;
  bytes_in: number;
  bytes_out: number;
  avg_latency_ms: number;
}

interface RealtimeChartProps {
  data: TimeSeriesData[];
  metric: 'requests' | 'avg_latency_ms' | 'bytes_in' | 'bytes_out';
  title: string;
  color: string;
}

export const RealtimeChart: React.FC<RealtimeChartProps> = ({ data, metric, title, color }) => {
  // Format timestamp for display
  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
  };

  // Format value based on metric type
  const formatValue = (value: number) => {
    if (metric === 'avg_latency_ms') {
      return `${value.toFixed(2)}ms`;
    }
    if (metric === 'bytes_in' || metric === 'bytes_out') {
      return formatBytes(value);
    }
    return value.toString();
  };

  return (
    <div className="card">
      <div className="flex items-center gap-2 mb-6">
        <TrendingUp className="w-5 h-5 text-primary-400" />
        <h3 className="text-lg font-semibold text-white">{title}</h3>
      </div>

      <ResponsiveContainer width="100%" height={250}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
          <XAxis
            dataKey="timestamp"
            tickFormatter={formatTime}
            stroke="#64748b"
            style={{ fontSize: '12px' }}
          />
          <YAxis
            stroke="#64748b"
            style={{ fontSize: '12px' }}
            tickFormatter={(value) => {
              if (metric === 'avg_latency_ms') return `${value}ms`;
              if (metric === 'bytes_in' || metric === 'bytes_out') return formatBytes(value);
              return value;
            }}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1e293b',
              border: '1px solid #334155',
              borderRadius: '0.5rem',
            }}
            labelStyle={{ color: '#cbd5e1' }}
            formatter={(value: number | undefined) => [formatValue(value ?? 0), title]}
            labelFormatter={(label) => formatTime(String(label))}
          />
          <Line
            type="monotone"
            dataKey={metric}
            stroke={color}
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
};

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}
