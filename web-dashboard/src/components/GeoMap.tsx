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
        <Globe className="w-5 h-5 text-primary-400" />
        <h3 className="text-lg font-semibold text-white">Geographic Distribution</h3>
      </div>

      {/* Bar Chart */}
      <ResponsiveContainer width="100%" height={300}>
        <BarChart data={topCountries} layout="horizontal">
          <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
          <XAxis 
            type="number" 
            stroke="#64748b"
            style={{ fontSize: '12px' }}
          />
          <YAxis 
            type="category" 
            dataKey="key" 
            stroke="#64748b"
            style={{ fontSize: '12px' }}
            width={100}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1e293b',
              border: '1px solid #334155',
              borderRadius: '0.5rem',
            }}
            labelStyle={{ color: '#cbd5e1' }}
          />
          <Bar dataKey="count" radius={[0, 4, 4, 0]}>
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
            <div key={country.key} className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: COLORS[index % COLORS.length] }}
                />
                <span className="text-sm text-dark-300">{country.key}</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-sm text-dark-400">{country.count.toLocaleString()}</span>
                <span className="text-xs text-dark-500 w-12 text-right">{percentage}%</span>
              </div>
            </div>
          );
        })}
      </div>

      {data.length > 8 && (
        <div className="mt-4 pt-4 border-t border-dark-700">
          <p className="text-xs text-dark-500 text-center">
            Showing top 8 countries out of {data.length}
          </p>
        </div>
      )}
    </div>
  );
};