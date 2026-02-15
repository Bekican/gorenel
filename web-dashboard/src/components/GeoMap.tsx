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
  '#10b981', // emerald
  '#3b82f6', // blue
  '#8b5cf6', // violet
  '#f43f5e', // rose
  '#f59e0b', // amber
  '#06b6d4', // cyan
  '#ec4899', // pink
  '#f97316', // orange
];

export const GeoMap: React.FC<GeoMapProps> = ({ data }) => {
  // Get top 8 countries
  const topCountries = data.slice(0, 8);
  const totalRequests = data.reduce((sum, item) => sum + item.count, 0);

  return (
    <div className="card h-full">
      <div className="flex items-center gap-2 mb-6">
        <div className="p-2 bg-primary/10 rounded-lg">
          <Globe className="w-5 h-5 text-primary" />
        </div>
        <h3 className="text-lg font-bold text-white tracking-tight">Geographic Distribution</h3>
      </div>

      {/* Bar Chart */}
      <ResponsiveContainer width="100%" height={300}>
        <BarChart data={topCountries} layout="horizontal">
          <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" vertical={false} />
          <XAxis
            type="number"
            stroke="rgba(255,255,255,0.3)"
            style={{ fontSize: '10px', fontFamily: 'inherit' }}
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            type="category"
            dataKey="key"
            stroke="rgba(255,255,255,0.5)"
            style={{ fontSize: '11px', fontWeight: 600, fontFamily: 'inherit' }}
            width={100}
            tickLine={false}
            axisLine={false}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: '#09090b',
              border: '1px solid rgba(255,255,255,0.1)',
              borderRadius: '1rem',
              boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.5)',
              color: '#fff'
            }}
            itemStyle={{ color: '#fff' }}
            labelStyle={{ color: 'rgba(255,255,255,0.5)' }}
            cursor={{ fill: 'rgba(255,255,255,0.05)' }}
          />
          <Bar dataKey="count" radius={[0, 4, 4, 0]} barSize={20}>
            {topCountries.map((_entry, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>

      {/* Country List */}
      <div className="mt-6 space-y-2">
        {topCountries.map((country, index) => {
          const percentage = totalRequests > 0 ? ((country.count / totalRequests) * 100).toFixed(1) : "0.0";
          return (
            <div key={country.key} className="flex items-center justify-between p-2 hover:bg-white/5 rounded-lg transition-colors group">
              <div className="flex items-center gap-3">
                <div
                  className="w-2.5 h-2.5 rounded-full shadow-[0_0_8px_rgba(0,0,0,0.5)]"
                  style={{ backgroundColor: COLORS[index % COLORS.length] }}
                />
                <span className="text-sm text-white/70 font-medium group-hover:text-white transition-colors">{country.key}</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-sm text-white/40 font-mono group-hover:text-white/60">{country.count.toLocaleString()}</span>
                <span className="text-xs text-primary w-12 text-right bg-primary/10 px-1 py-0.5 rounded font-bold">{percentage}%</span>
              </div>
            </div>
          );
        })}
      </div>

      {data.length > 8 && (
        <div className="mt-4 pt-4 border-t border-white/5">
          <p className="text-xs text-white/30 text-center uppercase tracking-widest font-bold">
            Showing top 8 of {data.length}
          </p>
        </div>
      )}
    </div>
  );
};
