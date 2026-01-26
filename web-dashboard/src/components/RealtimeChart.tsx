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
        <div className="p-2 bg-primary-50 rounded-lg">
          <TrendingUp className="w-5 h-5 text-primary-600" />
        </div>
        <h3 className="text-lg font-semibold text-neutral-900">{title}</h3>
      </div>

      <ResponsiveContainer width="100%" height={250}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" vertical={false} />
          <XAxis
            dataKey="timestamp"
            tickFormatter={formatTime}
            stroke="#9ca3af"
            style={{ fontSize: '12px' }}
            tickLine={false}
            axisLine={false}
            dy={10}
          />
          <YAxis
            stroke="#9ca3af"
            style={{ fontSize: '12px' }}
            tickLine={false}
            axisLine={false}
            tickFormatter={(value) => {
              if (metric === 'avg_latency_ms') return `${value}ms`;
              if (metric === 'bytes_in' || metric === 'bytes_out') return formatBytes(value);
              return value;
            }}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: '#ffffff',
              border: '1px solid #e5e7eb',
              borderRadius: '0.5rem',
              boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
              color: '#111827'
            }}
            itemStyle={{ color: '#111827' }}
            labelStyle={{ color: '#6b7280', marginBottom: '4px' }}
            formatter={(value: number | undefined) => [formatValue(value ?? 0), title]}
            labelFormatter={(label) => formatTime(String(label))}
          />
          <Line
            type="monotone"
            dataKey={metric}
            stroke={color}
            strokeWidth={3}
            dot={false}
            activeDot={{ r: 6, strokeWidth: 0 }}
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
