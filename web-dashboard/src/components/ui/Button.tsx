import React from 'react';

type Props = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'outline' | 'light';
  size?: 'sm' | 'md' | 'lg';
};

const VARIANT: Record<NonNullable<Props['variant']>, string> = {
  primary: 'bg-emerald-500 text-[#020408] hover:bg-emerald-400 shadow-[0_0_20px_-6px_rgba(16,185,129,0.55)]',
  secondary: 'bg-white/[0.04] text-white/85 hover:bg-white/[0.07] border border-white/10',
  ghost: 'bg-transparent text-white/70 hover:text-white hover:bg-white/[0.06] border border-white/10',
  danger: 'bg-rose-500/10 text-rose-200 hover:bg-rose-500/15 border border-rose-500/25',
  outline:
    'bg-transparent text-white/85 hover:bg-white/[0.06] border border-white/15 shadow-[0_0_0_1px_rgba(255,255,255,0.02)_inset]',
  light:
    'bg-white text-[#020408] hover:bg-white/90 shadow-[0_12px_40px_rgba(0,0,0,0.35)]',
};

const SIZE: Record<NonNullable<Props['size']>, string> = {
  sm: 'h-9 px-3 text-xs rounded-xl',
  md: 'h-11 px-4 text-sm rounded-2xl',
  lg: 'h-12 px-5 text-sm rounded-2xl',
};

export const Button: React.FC<Props> = ({ variant = 'secondary', size = 'md', className = '', ...rest }) => {
  return (
    <button
      {...rest}
      className={[
        'inline-flex items-center justify-center gap-2 font-black tracking-tight transition active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed',
        VARIANT[variant],
        SIZE[size],
        className,
      ].join(' ')}
    />
  );
};

