import React from 'react';
import { AlertTriangle, CheckCircle2, Info } from 'lucide-react';

type Props = {
  variant?: 'error' | 'info' | 'success';
  title?: string;
  children: React.ReactNode;
};

const STYLE: Record<NonNullable<Props['variant']>, string> = {
  error: 'border-rose-500/25 bg-rose-500/10 text-rose-100',
  info: 'border-white/10 bg-white/[0.04] text-white/80',
  success: 'border-emerald-500/25 bg-emerald-500/10 text-emerald-100',
};

export const Alert: React.FC<Props> = ({ variant = 'info', title, children }) => {
  const Icon = variant === 'error' ? AlertTriangle : variant === 'success' ? CheckCircle2 : Info;
  return (
    <div className={`rounded-2xl border px-4 py-3 ${STYLE[variant]} flex gap-3`}>
      <Icon className="w-4 h-4 shrink-0 opacity-90 mt-0.5" />
      <div className="min-w-0">
        {title ? <div className="text-[11px] font-black uppercase tracking-widest opacity-80">{title}</div> : null}
        <div className="text-[12px] font-semibold leading-relaxed break-words">{children}</div>
      </div>
    </div>
  );
};

