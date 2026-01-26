import React from "react";
import type { LucideIcon } from "lucide-react";

interface MetricCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: LucideIcon;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  color?: "blue" | "green" | "purple" | "orange" | "red";
}

const colorClasses = {
  blue: "from-blue-500 to-cyan-500",
  green: "from-green-500 to-emerald-500",
  purple: "from-purple-500 to-pink-500",
  orange: "from-orange-500 to-red-500",
  red: "from-red-500 to-rose-500",
};

const iconBgColors = {
  blue: "bg-blue-50",
  green: "bg-green-50",
  purple: "bg-purple-50",
  orange: "bg-orange-50",
  red: "bg-red-50",
};

const iconTextColors = {
  blue: "text-blue-600",
  green: "text-green-600",
  purple: "text-purple-600",
  orange: "text-orange-600",
  red: "text-red-600",
};

const useCountUp = (end: number | string, duration = 1000) => {
  const [displayValue, setDisplayValue] = React.useState(end);

  React.useEffect(() => {
    let numericEnd = 0;
    let suffix = "";
    let prefix = "";

    if (typeof end === 'number') {
      numericEnd = end;
    } else {
      const match = String(end).match(/([^\d.-]*)([\d,.]+)([^\d]*)/);
      if (match) {
        prefix = match[1];
        const numStr = match[2].replace(/,/g, '');
        numericEnd = parseFloat(numStr);
        suffix = match[3];
      } else {
        setDisplayValue(end);
        return;
      }
    }

    if (isNaN(numericEnd)) {
      setDisplayValue(end);
      return;
    }

    let startTimestamp: number | null = null;
    const step = (timestamp: number) => {
      if (!startTimestamp) startTimestamp = timestamp;
      const progress = Math.min((timestamp - startTimestamp) / duration, 1);
      const ease = progress === 1 ? 1 : 1 - Math.pow(2, -10 * progress);
      const current = numericEnd * ease;

      let formattedCurrent;
      if (Number.isInteger(numericEnd)) {
        formattedCurrent = Math.floor(current).toLocaleString();
      } else {
        formattedCurrent = current.toFixed(2);
      }

      setDisplayValue(`${prefix}${formattedCurrent}${suffix}`);

      if (progress < 1) {
        window.requestAnimationFrame(step);
      }
    };
    window.requestAnimationFrame(step);
  }, [end, duration]);

  return displayValue;
};

export const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  subtitle,
  icon: Icon,
  trend,
  color = "blue",
}) => {
  const animatedValue = useCountUp(value);

  return (
    <div className="metric-card group relative overflow-hidden transform transition-all duration-300 hover:-translate-y-1 hover:shadow-lg animate-fade-in-up">
      <div
        className={`absolute inset-0 bg-gradient-to-br ${colorClasses[color]} opacity-0 group-hover:opacity-5 transition-opacity duration-300`}
      />

      <div className="relative">
        <div className="flex items-start justify-between mb-4">
          <div className={`p-3 rounded-xl ${iconBgColors[color]} shadow-sm group-hover:scale-110 transition-transform duration-300`}>
            <Icon className={`w-6 h-6 ${iconTextColors[color]}`} />
          </div>

          {trend && (
            <div
              className={`flex items-center gap-1 text-sm font-medium ${trend.isPositive ? "text-green-600" : "text-red-600"
                }`}
            >
              {trend.isPositive ? "↑" : "↓"} {Math.abs(trend.value)}%
            </div>
          )}
        </div>

        <div className="space-y-1">
          <p className="text-3xl font-bold text-neutral-900 tracking-tight">{animatedValue}</p>
          <p className="text-sm text-neutral-500 font-medium">{title}</p>
          {subtitle && <p className="text-xs text-neutral-400">{subtitle}</p>}
        </div>
      </div>
    </div>
  );
};
