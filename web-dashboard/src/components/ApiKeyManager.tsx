import React, { useState, useEffect, useCallback } from 'react';
import { Key, Plus, Copy, Trash2, Shield, Clock, AlertTriangle, CheckCircle2, Activity, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { api } from '../api/client';
import { ConfirmDialog } from './ui/ConfirmDialog';
import { Tooltip } from './ui/Tooltip';

interface APIKey {
  key: string;
  created_at: string;
  usage_count: number;
  rate_limit: number;
}

export const ApiKeyManager: React.FC = () => {
  const { t } = useTranslation();
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showNewKey, setShowNewKey] = useState<string | null>(null);
  const [revokeKey, setRevokeKey] = useState<string | null>(null);
  const [toast, setToast] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const showToast = useCallback((message: string) => {
    setToast(message);
  }, []);

  useEffect(() => {
    if (!toast) return;
    const id = window.setTimeout(() => setToast(null), 2800);
    return () => window.clearTimeout(id);
  }, [toast]);

  useEffect(() => {
    fetchKeys();
  }, []);

  const fetchKeys = async () => {
    try {
      setLoading(true);
      const data = await api.listAPIKeys();
      setKeys(data);
      setError(null);
    } catch (err) {
      setError(t('common.error_fetch', 'Failed to load API keys'));
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateKey = async () => {
    setActionError(null);
    try {
      const result = await api.createAPIKey();
      setShowNewKey(result.key);
      fetchKeys();
    } catch {
      setActionError(t('api_keys_manager.create_failed'));
    }
  };

  const confirmRevoke = async () => {
    if (!revokeKey) return;
    const k = revokeKey;
    setRevokeKey(null);
    setActionError(null);
    try {
      await api.deleteAPIKey(k);
      fetchKeys();
      showToast(t('api_keys_manager.revoke_success'));
    } catch {
      setActionError(t('api_keys_manager.delete_failed'));
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    showToast(t('api_keys_manager.copied_toast'));
  };

  const installCmd = (key: string, isWindows: boolean) =>
    isWindows
      ? `powershell -ExecutionPolicy ByPass -Command "iwr -useb https://gorenel.site/install.ps1 | iex; gorenel config set api_key ${key}; gorenel connect --port 3000"`
      : `curl -sSL https://gorenel.site/install.sh | bash -s --; gorenel config set api_key ${key}; gorenel connect --port 3000`;

  if (loading) {
    return (
      <div className="flex min-h-[280px] flex-col items-center justify-center gap-4 rounded-3xl border border-white/[0.04] bg-white/[0.02]">
        <Activity className="h-10 w-10 animate-spin text-emerald-500/80" aria-hidden />
        <p className="text-sm font-medium text-white/45">{t('api_keys_manager.loading')}</p>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <ConfirmDialog
        open={!!revokeKey}
        title={t('api_keys_manager.revoke_title')}
        description={t('api_keys_manager.revoke_confirm')}
        confirmLabel={t('api_keys_manager.revoke_cta')}
        cancelLabel={t('api_keys_manager.revoke_cancel')}
        onConfirm={confirmRevoke}
        onCancel={() => setRevokeKey(null)}
        variant="danger"
      />

      {toast && (
        <div
          role="status"
          className="fixed bottom-6 left-1/2 z-[90] flex -translate-x-1/2 items-center gap-3 rounded-2xl border border-emerald-500/25 bg-[#0c1210]/95 px-6 py-3.5 text-sm font-semibold text-emerald-100 shadow-[0_12px_40px_rgba(0,0,0,0.45)] backdrop-blur-md animate-in fade-in slide-in-from-bottom-4 duration-300"
        >
          <CheckCircle2 className="h-4 w-4 shrink-0 text-emerald-400" aria-hidden />
          {toast}
        </div>
      )}

      <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
        <div className="space-y-1">
          <h2 className="flex items-center gap-3 text-2xl font-bold tracking-tight text-white md:text-3xl">
            <span className="flex h-11 w-11 items-center justify-center rounded-2xl border border-emerald-500/15 bg-emerald-500/10 shadow-[0_0_24px_-4px_rgba(16,185,129,0.25)]">
              <Shield className="h-5 w-5 text-emerald-400" aria-hidden />
            </span>
            {t('api_keys_manager.title')}
          </h2>
          <p className="max-w-xl text-sm leading-relaxed text-white/55 md:text-base">{t('api_keys_manager.subtitle')}</p>
        </div>
        <button
          type="button"
          onClick={handleCreateKey}
          className="btn-primary-premium inline-flex shrink-0 items-center justify-center gap-2 px-5 py-3 text-sm font-bold"
        >
          <Plus size={18} strokeWidth={2.5} />
          {t('api_keys_manager.generate')}
        </button>
      </div>

      {actionError && (
        <div className="flex items-center gap-3 rounded-2xl border border-rose-500/25 bg-rose-500/10 px-4 py-3 text-sm text-rose-200/95">
          <AlertTriangle size={18} className="shrink-0 text-rose-400" />
          <span>{actionError}</span>
          <button
            type="button"
            onClick={() => setActionError(null)}
            className="ml-auto rounded-lg p-1 text-rose-300/80 transition hover:bg-white/10"
            aria-label={t('api_keys_manager.dismiss')}
          >
            <X size={16} />
          </button>
        </div>
      )}

      <div className="relative overflow-hidden rounded-[2rem] border border-white/[0.06] bg-gradient-to-br from-emerald-500/[0.07] via-[#0A0C10] to-blue-500/[0.06] p-px shadow-[0_0_0_1px_rgba(255,255,255,0.03)]">
        <div className="flex flex-col gap-6 rounded-[1.95rem] bg-[#080a0d]/90 p-6 md:flex-row md:items-center md:gap-8 md:p-10">
          <div className="flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl border border-white/[0.06] bg-gradient-to-br from-emerald-500/15 to-blue-500/10 shadow-inner">
            <Key className="h-8 w-8 text-emerald-400/90" />
          </div>
          <div className="min-w-0 space-y-2">
            <h3 className="text-lg font-bold tracking-tight text-white md:text-xl">{t('api_keys_manager.onboarding_title')}</h3>
            <p className="text-sm font-medium leading-relaxed text-white/50">{t('api_keys_manager.onboarding_desc')}</p>
          </div>
        </div>
      </div>

      {showNewKey && (
        <div className="relative overflow-hidden rounded-3xl border border-emerald-500/25 bg-[#0a1210]/80 p-6 shadow-[0_0_40px_-12px_rgba(16,185,129,0.35)] animate-in fade-in slide-in-from-top-3 duration-300 md:p-8">
          <div className="absolute -right-4 -top-4 h-32 w-32 rounded-full bg-emerald-500/10 blur-3xl" aria-hidden />
          <div className="relative flex flex-col gap-6 sm:flex-row sm:items-start sm:justify-between">
            <div className="min-w-0 flex-1 space-y-3">
              <p className="text-sm font-semibold text-emerald-400/95">{t('api_keys_manager.success')}</p>
              <code className="block break-all rounded-2xl border border-white/10 bg-black/50 px-4 py-3 font-mono text-sm text-white/90">
                {showNewKey}
              </code>
              <p className="text-xs leading-relaxed text-white/40">{t('api_keys_manager.security_notice')}</p>
            </div>
            <div className="flex shrink-0 gap-2 sm:flex-col">
              <button
                type="button"
                onClick={() => copyToClipboard(showNewKey)}
                className="inline-flex items-center justify-center gap-2 rounded-xl border border-emerald-500/30 bg-emerald-500/15 px-4 py-2.5 text-sm font-semibold text-emerald-100 transition hover:bg-emerald-500/25"
              >
                <Copy size={18} />
                {t('api_keys_manager.copy_label')}
              </button>
              <button
                type="button"
                onClick={() => setShowNewKey(null)}
                className="rounded-xl border border-white/10 px-4 py-2.5 text-sm font-semibold text-white/55 transition hover:bg-white/[0.04] hover:text-white/90"
              >
                {t('api_keys_manager.dismiss')}
              </button>
            </div>
          </div>
        </div>
      )}

      {error && (
        <div className="flex items-center gap-3 rounded-2xl border border-rose-500/30 bg-rose-500/[0.08] px-4 py-3 text-sm text-rose-200">
          <AlertTriangle size={20} className="shrink-0" />
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 gap-5">
        {keys.map((k) => (
          <article
            key={k.key}
            className="group/card relative overflow-hidden rounded-3xl border border-white/[0.06] bg-gradient-to-b from-white/[0.04] to-transparent p-6 shadow-[0_24px_80px_-32px_rgba(0,0,0,0.8)] transition-all duration-300 hover:border-white/[0.1] md:p-8"
          >
            <div className="pointer-events-none absolute -right-8 -top-8 h-40 w-40 rounded-full bg-emerald-500/[0.06] blur-3xl transition-opacity duration-500 group-hover/card:opacity-100" />
            <div className="relative space-y-6">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div className="flex min-w-0 gap-4">
                  <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border border-white/[0.06] bg-white/[0.03] text-emerald-400/90">
                    <Key size={22} strokeWidth={2} />
                  </div>
                  <div className="min-w-0 space-y-2">
                    <h3 className="text-base font-bold text-white md:text-lg">
                      <span className="text-white/60">{t('api_keys_manager.key_type_tunnel')}</span>
                      <span className="mx-2 text-white/25">·</span>
                      <span className="font-mono text-sm text-emerald-400/95 md:text-base">{k.key.substring(0, 14)}…</span>
                    </h3>
                    <div className="flex flex-wrap items-center gap-x-5 gap-y-1 text-xs text-white/40">
                      <span className="inline-flex items-center gap-1.5">
                        <Clock size={12} className="text-white/30" />
                        {new Date(k.created_at).toLocaleDateString()}
                      </span>
                      <span className="inline-flex items-center gap-1.5">
                        <CheckCircle2 size={12} className="text-emerald-500/80" />
                        {k.usage_count.toLocaleString()} {t('api_keys_manager.requests_label')}
                      </span>
                    </div>
                  </div>
                </div>
                <Tooltip label={t('api_keys_manager.tooltip_revoke')}>
                  <button
                    type="button"
                    onClick={() => setRevokeKey(k.key)}
                    className="rounded-xl border border-transparent p-2.5 text-white/35 transition hover:border-rose-500/25 hover:bg-rose-500/10 hover:text-rose-400 focus:outline-none focus-visible:ring-2 focus-visible:ring-rose-500/40"
                  >
                    <Trash2 size={18} />
                  </button>
                </Tooltip>
              </div>

              <div className="space-y-4 border-t border-white/[0.05] pt-6">
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <span className="inline-flex items-center gap-2 text-[10px] font-bold uppercase tracking-[0.2em] text-emerald-500/80">
                    <span className="h-1 w-1 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)]" />
                    {t('api_keys_manager.magic_command_section')}
                  </span>
                  <div className="flex flex-wrap gap-1.5">
                    {(
                      [
                        ['win', t('api_keys_manager.os_windows'), true],
                        ['linux', t('api_keys_manager.os_linux'), false],
                        ['mac', t('api_keys_manager.os_mac'), false],
                      ] as const
                    ).map(([id, label, isWindows]) => (
                      <button
                        key={id}
                        type="button"
                        className="rounded-lg border border-white/[0.08] bg-white/[0.03] px-3 py-1.5 text-[11px] font-semibold text-white/60 transition hover:border-emerald-500/30 hover:bg-emerald-500/10 hover:text-emerald-100"
                        onClick={(e) => {
                          e.stopPropagation();
                          copyToClipboard(installCmd(k.key, isWindows));
                        }}
                      >
                        {label}
                      </button>
                    ))}
                  </div>
                </div>

                <button
                  type="button"
                  className="group/cmd relative w-full rounded-2xl border border-white/[0.06] bg-black/50 p-4 text-left font-mono text-[11px] leading-relaxed text-white/75 transition hover:border-emerald-500/20 hover:bg-black/70 md:text-xs"
                  onClick={() =>
                    copyToClipboard(
                      `powershell -ExecutionPolicy ByPass -Command "iwr -useb https://gorenel.site/install.ps1 | iex; gorenel config set api_key ${k.key}; gorenel connect --port 3000"`,
                    )
                  }
                >
                  <div className="flex items-start gap-3 pr-4">
                    <span className="shrink-0 text-emerald-500/90">$</span>
                    <code className="min-w-0 flex-1 break-all">
                      powershell -ExecutionPolicy ByPass -Command &quot;iwr -useb https://gorenel.site/install.ps1 | iex; gorenel config set api_key {k.key}; gorenel connect --port 3000&quot;
                    </code>
                  </div>
                  <div className="pointer-events-none absolute right-3 top-1/2 flex -translate-y-1/2 items-center gap-1.5 rounded-lg bg-emerald-500/90 px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide text-[#020408] opacity-0 shadow-lg transition group-hover/cmd:opacity-100">
                    <Copy size={12} />
                    {t('api_keys_manager.copy_label')}
                  </div>
                </button>
                <p className="text-[11px] leading-relaxed text-white/35">{t('api_keys_manager.magic_command_note')}</p>
              </div>
            </div>
          </article>
        ))}

        {keys.length === 0 && !loading && (
          <div className="py-16 text-center">
            <div className="mx-auto mb-5 flex h-16 w-16 items-center justify-center rounded-2xl border border-dashed border-white/10 bg-white/[0.02]">
              <Shield className="text-white/20" size={28} />
            </div>
            <p className="text-sm text-white/45">{t('api_keys_manager.empty')}</p>
          </div>
        )}
      </div>
    </div>
  );
};
