export type TabItem<T extends string> = {
  id: T;
  label: string;
  hint?: string;
};

type Props<T extends string> = {
  value: T;
  onChange: (v: T) => void;
  tabs: TabItem<T>[];
};

export function Tabs<T extends string>({ value, onChange, tabs }: Props<T>) {
  return (
    <div className="inline-flex gap-1 rounded-xl border border-white/[0.06] bg-white/[0.02] p-1">
      {tabs.map((t) => {
        const active = t.id === value;
        return (
          <button
            key={t.id}
            type="button"
            onClick={() => onChange(t.id)}
            className={[
              'px-4 py-2 rounded-lg text-xs font-semibold transition-all duration-200',
              active
                ? 'bg-white/[0.1] text-white shadow-sm'
                : 'text-white/40 hover:text-white/70 hover:bg-white/[0.04]',
            ].join(' ')}
          >
            {t.label}
          </button>
        );
      })}
    </div>
  );
}
