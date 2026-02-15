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
    bg: "bg-emerald-500/10",
    text: "text-emerald-500",
    glow: "shadow-[0_0_15px_rgba(16,185,129,0.3)]",
    trend: "text-emerald-400"
  },
  violet: {
    bg: "bg-violet-500/10",
    text: "text-violet-500",
    glow: "shadow-[0_0_15px_rgba(139,92,246,0.3)]",
    trend: "text-violet-400"
  },
  blue: {
    bg: "bg-blue-500/10",
    text: "text-blue-500",
    glow: "shadow-[0_0_15px_rgba(59,130,246,0.3)]",
    trend: "text-blue-400"
  },
  rose: {
    bg: "bg-rose-500/10",
    text: "text-rose-500",
    glow: "shadow-[0_0_15px_rgba(244,63,94,0.3)]",
    trend: "text-rose-400"
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
    <div className="metric-card relative group">
      {/* Background glow on hover */}
      <div className={`absolute -inset-1 rounded-[2.2rem] bg-gradient-to-br transition-all duration-500 opacity-0 group-hover:opacity-10 blur-xl ${style.bg}`} />

      <div className="relative h-full flex flex-col justify-between">
        <div className="flex items-center justify-between mb-8">
          <div className={`p-4 rounded-2xl ${style.bg} ${style.glow} group-hover:scale-110 transition-transform duration-500`}>
            <Icon className={`w-6 h-6 ${style.text}`} />
          </div>

          {trend && (
            <div className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/5 border border-white/5 text-[10px] font-black uppercase tracking-widest ${trend.isPositive ? 'text-primary' : 'text-rose-400'}`}>
              {trend.isPositive ? "↑" : "↓"} {Math.abs(trend.value)}%
            </div>
          )}
        </div>

        <div>
          <h3 className="text-white/40 text-[10px] font-black uppercase tracking-[0.2em] mb-2">{title}</h3>
          <div className="flex items-baseline gap-2">
            <span className="text-4xl font-black tracking-tighter text-gradient">{animatedValue}</span>
          </div>
        </div>
      </div>
    </div>
  );
};
