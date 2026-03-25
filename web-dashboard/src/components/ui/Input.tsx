import React from 'react';

type Props = React.InputHTMLAttributes<HTMLInputElement> & {
  leftIcon?: React.ReactNode;
};

export const Input: React.FC<Props> = ({ leftIcon, className = '', ...rest }) => {
  return (
    <div className="relative group">
      {leftIcon ? (
        <div className="absolute left-3.5 top-1/2 -translate-y-1/2 text-white/25 group-focus-within:text-emerald-400/80 transition-colors duration-200">
          {leftIcon}
        </div>
      ) : null}
      <input
        {...rest}
        className={[
          'w-full bg-white/[0.03] border border-white/[0.08] rounded-xl py-3',
          leftIcon ? 'pl-10 pr-4' : 'px-4',
          'text-sm font-medium text-white/90 placeholder:text-white/20 outline-none',
          'hover:border-white/[0.12] focus:border-emerald-500/40 focus:ring-1 focus:ring-emerald-500/20',
          'transition-all duration-200',
          'disabled:opacity-40 disabled:pointer-events-none',
          className,
        ].join(' ')}
      />
    </div>
  );
};
