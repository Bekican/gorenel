import React from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import { Globe } from 'lucide-react';

interface GeoData {
  key: string;
  count: number;
}

interface GeoMapProps {
  data: GeoData[];
}

const COLORS = [
  '#10b981',
  '#3b82f6',
  '#8b5cf6',
  '#f43f5e',
  '#f59e0b',
  '#06b6d4',
  '#ec4899',
  '#f97316',
];

export const GeoMap: React.FC<GeoMapProps> = ({ data }) => {
  const topCountries = data.slice(0, 8);
  const totalRequests = data.reduce((sum, item) => sum + item.count, 0);

  return (
    <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 h-full">
      <div className="flex items-center gap-2.5 mb-5">
        <div className="p-2 bg-emerald-500/[0.08] rounded-lg border border-emerald-500/[0.1]">
          <Globe className="w-4 h-4 text-emerald-400" />
        </div>
        <h3 className="text-sm font-semibold text-white">Geographic Distribution</h3>
      </div>

      <ResponsiveContainer width="100%" height={280}>
        <BarChart data={topCountries} layout="horizontal">
          <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.03)" vertical={false} />
          <XAxis
            type="number"
            stroke="rgba(255,255,255,0.15)"
            style={{ fontSize: '10px' }}
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            type="category"
            dataKey="key"
            stroke="rgba(255,255,255,0.35)"
            style={{ fontSize: '11px', fontWeight: 500 }}
            width={90}
            tickLine={false}
            axisLine={false}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'rgba(12,14,20,0.95)',
              border: '1px solid rgba(255,255,255,0.06)',
              borderRadius: '0.75rem',
              boxShadow: '0 8px 24px rgba(0, 0, 0, 0.4)',
              color: '#fff',
              fontSize: '12px',
            }}
            itemStyle={{ color: '#fff' }}
            labelStyle={{ color: 'rgba(255,255,255,0.4)' }}
            cursor={{ fill: 'rgba(255,255,255,0.02)' }}
          />
          <Bar dataKey="count" radius={[0, 4, 4, 0]} barSize={18}>
            {topCountries.map((_entry, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} fillOpacity={0.8} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>

      <div className="mt-4 space-y-1.5">
        {topCountries.map((country, index) => {
          const percentage = totalRequests > 0 ? ((country.count / totalRequests) * 100).toFixed(1) : "0.0";
          return (
            <div key={country.key} className="flex items-center justify-between py-1.5 px-2 hover:bg-white/[0.02] rounded-lg transition-colors group">
              <div className="flex items-center gap-2.5">
                <div
                  className="w-2 h-2 rounded-full"
                  style={{ backgroundColor: COLORS[index % COLORS.length] }}
                />
                <span className="text-sm text-white/55 group-hover:text-white/80 transition-colors">{country.key}</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-sm text-white/30 font-mono tabular-nums">{country.count.toLocaleString()}</span>
                <span className="text-xs text-emerald-400/80 w-12 text-right font-medium tabular-nums">{percentage}%</span>
              </div>
            </div>
          );
        })}
      </div>

      {data.length > 8 && (
        <div className="mt-3 pt-3 border-t border-white/[0.04]">
          <p className="text-xs text-white/20 text-center">
            Showing top 8 of {data.length}
          </p>
        </div>
      )}
    </div>
  );
};
