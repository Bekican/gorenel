import React from 'react';

type Props = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'outline' | 'light';
  size?: 'sm' | 'md' | 'lg';
};

const VARIANT: Record<NonNullable<Props['variant']>, string> = {
  primary:
    'bg-emerald-500 text-[#080a10] hover:bg-emerald-400 shadow-[0_0_0_1px_rgba(16,185,129,0.15),0_1px_2px_rgba(0,0,0,0.2),0_0_16px_-6px_rgba(16,185,129,0.35)]',
  secondary:
    'bg-white/[0.05] text-white/80 hover:bg-white/[0.08] hover:text-white border border-white/[0.08] hover:border-white/[0.12]',
  ghost:
    'bg-transparent text-white/60 hover:text-white hover:bg-white/[0.06]',
  danger:
    'bg-rose-500/10 text-rose-300 hover:bg-rose-500/15 border border-rose-500/20 hover:border-rose-500/30',
  outline:
    'bg-transparent text-white/80 hover:bg-white/[0.05] border border-white/[0.1] hover:border-white/[0.15]',
  light:
    'bg-white text-[#080a10] hover:bg-white/95 shadow-[0_1px_3px_rgba(0,0,0,0.2),0_8px_24px_rgba(0,0,0,0.25)]',
};

const SIZE: Record<NonNullable<Props['size']>, string> = {
  sm: 'h-8 px-3 text-xs gap-1.5 rounded-lg',
  md: 'h-10 px-4 text-sm gap-2 rounded-xl',
  lg: 'h-11 px-5 text-sm gap-2 rounded-xl',
};

export const Button: React.FC<Props> = ({ variant = 'secondary', size = 'md', className = '', ...rest }) => {
  return (
    <button
      {...rest}
      className={[
        'inline-flex items-center justify-center font-semibold transition-all duration-200 active:scale-[0.98] disabled:opacity-40 disabled:pointer-events-none select-none',
        VARIANT[variant],
        SIZE[size],
        className,
      ].join(' ')}
    />
  );
};
