import React from 'react';
import { LineChart, Line, XAxis, Tooltip, ResponsiveContainer } from 'recharts';

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
    <div className="h-full min-h-[300px] flex flex-col justify-center">
      <div className="flex items-center gap-2 mb-6 px-2">
        {/* Minimal Header */}
        <div className={`w-2 h-8 rounded-full`} style={{ backgroundColor: color }} />
        <div>
          <h3 className="text-sm font-bold text-white/50 uppercase tracking-widest">{title}</h3>
          {data.length > 0 && (
            <div className="text-2xl font-light text-white">
              {formatValue(data[data.length - 1][metric] as number)}
            </div>
          )}
        </div>
      </div>

      <ResponsiveContainer width="100%" height={200}>
        <LineChart data={data}>
          <defs>
            <linearGradient id={`gradient-${metric}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={color} stopOpacity={0.3} />
              <stop offset="95%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <XAxis
            dataKey="timestamp"
            tickFormatter={formatTime}
            stroke="rgba(255,255,255,0.1)"
            style={{ fontSize: '10px', fontFamily: 'inherit' }}
            tickLine={false}
            axisLine={false}
            dy={10}
            minTickGap={30}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'rgba(10,12,16,0.8)',
              backdropFilter: 'blur(10px)',
              border: '1px solid rgba(255,255,255,0.05)',
              borderRadius: '1rem',
              boxShadow: '0 20px 40px -10px rgba(0, 0, 0, 0.5)',
              color: '#fff',
              padding: '12px 16px'
            }}
            itemStyle={{ color: '#fff', fontSize: '12px', fontWeight: 600 }}
            labelStyle={{ color: 'rgba(255,255,255,0.4)', marginBottom: '8px', fontSize: '10px' }}
            formatter={(value: number | undefined) => [formatValue(value ?? 0), title]}
            labelFormatter={(label) => formatTime(String(label))}
            cursor={{ stroke: 'rgba(255,255,255,0.1)', strokeWidth: 1, strokeDasharray: '4 4' }}
          />
          <Line
            type="monotone"
            dataKey={metric}
            stroke={color}
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4, fill: '#000', stroke: color, strokeWidth: 2 }}
            fill={`url(#gradient-${metric})`}
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
