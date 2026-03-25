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
    bg: "bg-emerald-500/[0.08]",
    border: "border-emerald-500/[0.12]",
    glow: "hover:shadow-glow-emerald"
  },
  violet: {
    text: "text-violet-400",
    bg: "bg-violet-500/[0.08]",
    border: "border-violet-500/[0.12]",
    glow: "hover:shadow-glow-blue"
  },
  blue: {
    text: "text-blue-400",
    bg: "bg-blue-500/[0.08]",
    border: "border-blue-500/[0.12]",
    glow: "hover:shadow-glow-blue"
  },
  rose: {
    text: "text-rose-400",
    bg: "bg-rose-500/[0.08]",
    border: "border-rose-500/[0.12]",
    glow: "hover:shadow-glow-rose"
  },
};

const useCountUp = (end: number | string, duration = 1200) => {
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
    <div className={`metric-card group ${style.glow}`}>
      <div className="relative z-10">
        <div className="flex justify-between items-start mb-5">
          <div className={`p-2.5 rounded-xl ${style.bg} ${style.border} border`}>
            <Icon className={`w-4 h-4 ${style.text}`} />
          </div>
          {trend && (
            <div className={`flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-md ${trend.isPositive ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400'}`}>
              {trend.isPositive ? "↑" : "↓"} {Math.abs(trend.value)}%
            </div>
          )}
        </div>

        <div className="space-y-1">
          <h3 className="text-white/35 text-xs font-medium">{title}</h3>
          <div className="text-2xl md:text-3xl font-semibold tracking-tight text-white tabular-nums">{animatedValue}</div>
        </div>
      </div>

      <div className="absolute inset-0 bg-gradient-to-br from-white/[0.02] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none rounded-2xl" />
    </div>
  );
};
