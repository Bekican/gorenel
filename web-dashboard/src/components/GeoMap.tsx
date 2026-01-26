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
  '#3b82f6', // blue
  '#10b981', // green
  '#f59e0b', // amber
  '#ef4444', // red
  '#8b5cf6', // purple
  '#ec4899', // pink
  '#06b6d4', // cyan
  '#f97316', // orange
];

export const GeoMap: React.FC<GeoMapProps> = ({ data }) => {
  // Get top 8 countries
  const topCountries = data.slice(0, 8);
  const totalRequests = data.reduce((sum, item) => sum + item.count, 0);

  return (
    <div className="card">
      <div className="flex items-center gap-2 mb-6">
        <div className="p-2 bg-primary-50 rounded-lg">
          <Globe className="w-5 h-5 text-primary-600" />
        </div>
        <h3 className="text-lg font-semibold text-neutral-900">Geographic Distribution</h3>
      </div>

      {/* Bar Chart */}
      <ResponsiveContainer width="100%" height={300}>
        <BarChart data={topCountries} layout="horizontal">
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" vertical={false} />
          <XAxis
            type="number"
            stroke="#9ca3af"
            style={{ fontSize: '12px' }}
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            type="category"
            dataKey="key"
            stroke="#64748b"
            style={{ fontSize: '12px' }}
            width={100}
            tickLine={false}
            axisLine={false}
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
            labelStyle={{ color: '#6b7280' }}
            cursor={{ fill: 'transparent' }}
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
          const percentage = ((country.count / totalRequests) * 100).toFixed(1);
          return (
            <div key={country.key} className="flex items-center justify-between p-2 hover:bg-neutral-50 rounded-lg transition-colors">
              <div className="flex items-center gap-3">
                <div
                  className="w-2.5 h-2.5 rounded-full"
                  style={{ backgroundColor: COLORS[index % COLORS.length] }}
                />
                <span className="text-sm text-neutral-700 font-medium">{country.key}</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-sm text-neutral-600 font-mono">{country.count.toLocaleString()}</span>
                <span className="text-xs text-neutral-400 w-12 text-right bg-neutral-100 px-1 py-0.5 rounded">{percentage}%</span>
              </div>
            </div>
          );
        })}
      </div>

      {data.length > 8 && (
        <div className="mt-4 pt-4 border-t border-neutral-100">
          <p className="text-xs text-neutral-500 text-center">
            Showing top 8 countries out of {data.length}
          </p>
        </div>
      )}
    </div>
  );
};