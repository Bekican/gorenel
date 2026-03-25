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
  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
  };

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
    <div className="h-full min-h-[280px] flex flex-col justify-center">
      <div className="flex items-center gap-3 mb-5 px-1">
        <div className="w-1.5 h-7 rounded-full" style={{ backgroundColor: color, opacity: 0.7 }} />
        <div>
          <h3 className="text-xs font-medium text-white/40 uppercase tracking-wider">{title}</h3>
          {data.length > 0 && (
            <div className="text-xl font-semibold text-white mt-0.5 tabular-nums">
              {formatValue(data[data.length - 1][metric] as number)}
            </div>
          )}
        </div>
      </div>

      <ResponsiveContainer width="100%" height={180}>
        <LineChart data={data}>
          <defs>
            <linearGradient id={`gradient-${metric}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={color} stopOpacity={0.15} />
              <stop offset="95%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <XAxis
            dataKey="timestamp"
            tickFormatter={formatTime}
            stroke="rgba(255,255,255,0.06)"
            style={{ fontSize: '10px', fontFamily: 'inherit' }}
            tickLine={false}
            axisLine={false}
            dy={8}
            minTickGap={40}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'rgba(12,14,20,0.95)',
              backdropFilter: 'blur(12px)',
              border: '1px solid rgba(255,255,255,0.06)',
              borderRadius: '0.75rem',
              boxShadow: '0 8px 24px rgba(0, 0, 0, 0.4)',
              color: '#fff',
              padding: '10px 14px',
              fontSize: '12px',
            }}
            itemStyle={{ color: '#fff', fontSize: '12px', fontWeight: 500 }}
            labelStyle={{ color: 'rgba(255,255,255,0.35)', marginBottom: '6px', fontSize: '10px' }}
            formatter={(value: number | undefined) => [formatValue(value ?? 0), title]}
            labelFormatter={(label) => formatTime(String(label))}
            cursor={{ stroke: 'rgba(255,255,255,0.06)', strokeWidth: 1, strokeDasharray: '3 3' }}
          />
          <Line
            type="monotone"
            dataKey={metric}
            stroke={color}
            strokeWidth={1.5}
            dot={false}
            activeDot={{ r: 3, fill: '#080a10', stroke: color, strokeWidth: 2 }}
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
