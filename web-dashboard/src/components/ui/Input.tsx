import React from 'react';

type Props = React.InputHTMLAttributes<HTMLInputElement> & {
  leftIcon?: React.ReactNode;
};

export const Input: React.FC<Props> = ({ leftIcon, className = '', ...rest }) => {
  return (
    <div className="relative group">
      {leftIcon ? (
        <div className="absolute left-4 top-1/2 -translate-y-1/2 text-white/25 group-focus-within:text-emerald-400 transition-colors">
          {leftIcon}
        </div>
      ) : null}
      <input
        {...rest}
        className={[
          'w-full bg-black/40 border border-white/10 rounded-2xl py-3.5',
          leftIcon ? 'pl-11 pr-4' : 'px-4',
          'text-sm font-semibold text-white/85 placeholder:text-white/15 outline-none',
          'focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/40 transition',
          className,
        ].join(' ')}
      />
    </div>
  );
};

