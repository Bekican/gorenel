import React from "react";
import type { LucideIcon } from "lucide-react";

interface MetricCardProps {
  title: string;
  value: string | number;
  icon: LucideIcon;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  color?: "emerald" | "violet" | "blue" | "rose";
}

const colorStyles = {
  emerald: {
    text: "text-emerald-400",
    bg: "bg-emerald-500/5",
    border: "border-emerald-500/10",
    glow: "shadow-[0_0_30px_-10px_rgba(16,185,129,0.2)]"
  },
  violet: {
    text: "text-violet-400",
    bg: "bg-violet-500/5",
    border: "border-violet-500/10",
    glow: "shadow-[0_0_30px_-10px_rgba(139,92,246,0.2)]"
  },
  blue: {
    text: "text-blue-400",
    bg: "bg-blue-500/5",
    border: "border-blue-500/10",
    glow: "shadow-[0_0_30px_-10px_rgba(59,130,246,0.2)]"
  },
  rose: {
    text: "text-rose-400",
    bg: "bg-rose-500/5",
    border: "border-rose-500/10",
    glow: "shadow-[0_0_30px_-10px_rgba(244,63,94,0.2)]"
  },
};

const useCountUp = (end: number | string, duration = 1500) => {
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
        formattedCurrent = current.toFixed(0);
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
  icon: Icon,
  trend,
  color = "emerald",
}) => {
  const animatedValue = useCountUp(value);
  const style = colorStyles[color];

  return (
    <div className={`metric-card relative group hover:-translate-y-1 ${style.glow}`}>
      <div className="relative z-10">
        <div className="flex justify-between items-start mb-6">
          <div className={`p-3 rounded-xl ${style.bg} ${style.border} border`}>
            <Icon className={`w-5 h-5 ${style.text}`} />
          </div>
          {trend && (
            <div className={`flex items-center gap-1 text-xs font-semibold px-2 py-1 rounded-lg ${trend.isPositive ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400'}`}>
              {trend.isPositive ? "↑" : "↓"} {Math.abs(trend.value)}%
            </div>
          )}
        </div>

        <div className="space-y-1">
          <h3 className="text-white/40 text-xs font-medium uppercase tracking-wider">{title}</h3>
          <div className="text-4xl font-light tracking-tight text-white">{animatedValue}</div>
        </div>
      </div>

      {/* Dynamic Background Gradient */}
      <div className={`absolute inset-0 bg-gradient-to-br from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none`} />
    </div>
  );
};
