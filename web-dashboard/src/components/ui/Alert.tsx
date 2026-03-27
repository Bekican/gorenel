import React from 'react';
import { AlertTriangle, CheckCircle2, Info } from 'lucide-react';

type Props = {
  variant?: 'error' | 'info' | 'success';
  title?: string;
  children: React.ReactNode;
};

const STYLE: Record<NonNullable<Props['variant']>, { container: string; icon: string }> = {
  error: {
    container: 'border-rose-500/20 bg-rose-500/[0.06] text-rose-100',
    icon: 'text-rose-400',
  },
  info: {
    container: 'border-white/[0.08] bg-white/[0.03] text-white/75',
    icon: 'text-white/50',
  },
  success: {
    container: 'border-emerald-500/20 bg-emerald-500/[0.06] text-emerald-100',
    icon: 'text-emerald-400',
  },
};

export const Alert: React.FC<Props> = ({ variant = 'info', title, children }) => {
  const Icon = variant === 'error' ? AlertTriangle : variant === 'success' ? CheckCircle2 : Info;
  const style = STYLE[variant];
  const live = variant === 'error' ? { role: 'alert' as const, 'aria-live': 'assertive' as const } : {}
  return (
    <div className={`rounded-xl border px-4 py-3 ${style.container} flex gap-3 text-sm`} {...live}>
      <Icon className={`w-4 h-4 shrink-0 mt-0.5 ${style.icon}`} />
      <div className="min-w-0 flex-1">
        {title ? <div className="text-xs font-semibold mb-0.5">{title}</div> : null}
        <div className="text-[13px] leading-relaxed break-words opacity-80">{children}</div>
      </div>
    </div>
  );
};
