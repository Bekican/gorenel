import React from 'react';

type TooltipProps = {
  label: string;
  children: React.ReactNode;
  side?: 'top' | 'bottom';
};

export const Tooltip: React.FC<TooltipProps> = ({ label, children, side = 'top' }) => {
  const pos =
    side === 'top'
      ? 'bottom-full left-1/2 -translate-x-1/2 mb-2 origin-bottom'
      : 'top-full left-1/2 -translate-x-1/2 mt-2 origin-top';

  return (
    <span className="group/tooltip relative inline-flex focus-within:outline-none">
      {children}
      <span
        role="tooltip"
        className={`pointer-events-none absolute ${pos} z-[60] min-w-max max-w-[220px] rounded-lg border border-white/[0.08] bg-[#141619]/95 px-3 py-1.5 text-center text-xs font-medium leading-snug text-white/85 shadow-elevated backdrop-blur-lg transition-all duration-150 ease-out opacity-0 scale-95 invisible group-hover/tooltip:opacity-100 group-hover/tooltip:scale-100 group-hover/tooltip:visible group-focus-within/tooltip:opacity-100 group-focus-within/tooltip:scale-100 group-focus-within/tooltip:visible`}
      >
        {label}
      </span>
    </span>
  );
};
