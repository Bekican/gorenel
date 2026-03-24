import React from 'react';

type TooltipProps = {
  label: string;
  children: React.ReactNode;
  side?: 'top' | 'bottom';
};

/** Native `title` yerine tema ile uyumlu, yumuşak geçişli ipucu */
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
        className={`pointer-events-none absolute ${pos} z-[60] min-w-max max-w-[240px] rounded-xl border border-white/10 bg-[#0c0f14]/95 px-3 py-2 text-center text-[11px] font-medium leading-snug text-white/90 shadow-[0_8px_32px_rgba(0,0,0,0.5)] backdrop-blur-md transition-all duration-200 ease-out opacity-0 scale-95 invisible group-hover/tooltip:opacity-100 group-hover/tooltip:scale-100 group-hover/tooltip:visible group-focus-within/tooltip:opacity-100 group-focus-within/tooltip:scale-100 group-focus-within/tooltip:visible`}
      >
        {label}
      </span>
    </span>
  );
};
