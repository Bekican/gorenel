import React, { useEffect } from 'react';
import { AlertTriangle } from 'lucide-react';

type ConfirmDialogProps = {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  cancelLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
  variant?: 'danger' | 'default';
};

export const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  open,
  title,
  description,
  confirmLabel,
  cancelLabel,
  onConfirm,
  onCancel,
  variant = 'danger',
}) => {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCancel();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, onCancel]);

  if (!open) return null;

  const confirmBtn =
    variant === 'danger'
      ? 'bg-rose-500/90 text-white hover:bg-rose-400 focus-visible:ring-rose-400/40 shadow-[0_0_24px_-4px_rgba(244,63,94,0.45)]'
      : 'bg-emerald-500 text-[#020408] hover:bg-emerald-400 focus-visible:ring-emerald-400/40 shadow-[0_0_24px_-4px_rgba(16,185,129,0.4)]';

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6">
      <button
        type="button"
        className="absolute inset-0 bg-black/70 backdrop-blur-sm animate-in fade-in duration-200"
        aria-label="Close"
        onClick={onCancel}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="confirm-dialog-title"
        className="relative z-[101] w-full max-w-md animate-in fade-in zoom-in-95 duration-200"
      >
        <div className="rounded-3xl border border-white/[0.08] bg-[#0A0C10]/95 p-6 shadow-[0_24px_80px_rgba(0,0,0,0.5)] backdrop-blur-xl sm:p-8">
          <div className="flex gap-4">
            <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border border-rose-500/20 bg-rose-500/10">
              <AlertTriangle className="h-6 w-6 text-rose-400" aria-hidden />
            </div>
            <div className="min-w-0 flex-1 space-y-2">
              <h2 id="confirm-dialog-title" className="text-lg font-bold tracking-tight text-white">
                {title}
              </h2>
              <p className="text-sm leading-relaxed text-white/55">{description}</p>
            </div>
          </div>
          <div className="mt-8 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
            <button
              type="button"
              onClick={onCancel}
              className="rounded-xl border border-white/10 bg-white/[0.03] px-4 py-2.5 text-sm font-semibold text-white/80 transition hover:bg-white/[0.06] focus:outline-none focus-visible:ring-2 focus-visible:ring-white/20"
            >
              {cancelLabel}
            </button>
            <button
              type="button"
              onClick={onConfirm}
              className={`rounded-xl px-4 py-2.5 text-sm font-bold transition focus:outline-none focus-visible:ring-2 ${confirmBtn}`}
            >
              {confirmLabel}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
