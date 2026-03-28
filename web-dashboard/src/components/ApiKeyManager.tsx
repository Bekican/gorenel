import React, { useState, useEffect, useCallback } from 'react';
import { Key, Plus, Copy, Trash2, Shield, Clock, AlertTriangle, CheckCircle2, Activity, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { api } from '../api/client';
import { ConfirmDialog } from './ui/ConfirmDialog';
import { Tooltip } from './ui/Tooltip';
import { Button } from './ui/Button';
import { tunnelQuickCommandFull } from '../lib/tunnelQuickCommand';

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

  const fetchKeys = useCallback(async () => {
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
  }, [t]);

  useEffect(() => {
    fetchKeys();
  }, [fetchKeys]);

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

  const installCmd = (key: string, isWindows: boolean) => {
    if (typeof window === 'undefined') return '';
    return tunnelQuickCommandFull({
      apiKey: key,
      os: isWindows ? 'windows' : 'unix',
      hostname: window.location.hostname,
      protocol: window.location.protocol,
    });
  };

  if (loading) {
    return (
      <div className="flex min-h-[200px] flex-col items-center justify-center gap-3 rounded-2xl border border-white/[0.06] bg-white/[0.02]">
        <Activity className="h-8 w-8 animate-spin text-emerald-500/50" aria-hidden />
        <p className="text-sm text-white/35">{t('api_keys_manager.loading')}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
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
          className="fixed bottom-6 left-1/2 z-[90] flex -translate-x-1/2 items-center gap-2.5 rounded-xl border border-emerald-500/20 bg-[#0c1210]/95 px-5 py-3 text-sm text-emerald-100 shadow-elevated backdrop-blur-md animate-fade-in-up"
        >
          <CheckCircle2 className="h-4 w-4 shrink-0 text-emerald-400" aria-hidden />
          {toast}
        </div>
      )}

      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h2 className="flex items-center gap-2.5 text-xl font-semibold tracking-tight text-white md:text-2xl">
            <span className="flex h-10 w-10 items-center justify-center rounded-xl border border-emerald-500/15 bg-emerald-500/10">
              <Shield className="h-5 w-5 text-emerald-400" aria-hidden />
            </span>
            {t('api_keys_manager.title')}
          </h2>
          <p className="max-w-lg text-sm text-white/40 mt-1">{t('api_keys_manager.subtitle')}</p>
        </div>
        <Button type="button" variant="primary" size="md" onClick={handleCreateKey}>
          <Plus size={16} />
          {t('api_keys_manager.generate')}
        </Button>
      </div>

      {actionError && (
        <div className="flex items-center gap-2.5 rounded-xl border border-rose-500/20 bg-rose-500/[0.06] px-4 py-3 text-sm text-rose-200">
          <AlertTriangle size={16} className="shrink-0 text-rose-400" />
          <span className="flex-1">{actionError}</span>
          <button type="button" onClick={() => setActionError(null)} className="p-1 text-rose-300/60 hover:bg-white/[0.06] rounded-lg transition">
            <X size={14} />
          </button>
        </div>
      )}

      <div className="rounded-2xl border border-white/[0.06] bg-gradient-to-br from-emerald-500/[0.04] via-transparent to-blue-500/[0.03] p-px">
        <div className="flex flex-col gap-5 rounded-[calc(1rem-1px)] bg-[#0a0c11]/90 p-5 md:flex-row md:items-center md:gap-6 md:p-8">
          <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl border border-white/[0.06] bg-gradient-to-br from-emerald-500/10 to-blue-500/[0.06]">
            <Key className="h-6 w-6 text-emerald-400/80" />
          </div>
          <div className="min-w-0">
            <h3 className="text-base font-semibold text-white">{t('api_keys_manager.onboarding_title')}</h3>
            <p className="text-sm text-white/40 mt-0.5">{t('api_keys_manager.onboarding_desc')}</p>
          </div>
        </div>
      </div>

      {showNewKey && (
        <div className="rounded-2xl border border-emerald-500/20 bg-emerald-500/[0.04] p-5 md:p-6 animate-fade-in-up">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="min-w-0 flex-1 space-y-2">
              <p className="text-sm font-medium text-emerald-400">{t('api_keys_manager.success')}</p>
              <code className="block break-all rounded-xl border border-white/[0.08] bg-black/30 px-4 py-3 font-mono text-sm text-white/85">
                {showNewKey}
              </code>
              <p className="text-xs text-white/30">{t('api_keys_manager.security_notice')}</p>
            </div>
            <div className="flex shrink-0 gap-2">
              <Button type="button" onClick={() => copyToClipboard(showNewKey)} variant="primary" size="sm">
                <Copy size={14} /> {t('api_keys_manager.copy_label')}
              </Button>
              <Button type="button" onClick={() => setShowNewKey(null)} variant="outline" size="sm">
                {t('api_keys_manager.dismiss')}
              </Button>
            </div>
          </div>
        </div>
      )}

      {error && (
        <div className="flex items-center gap-2.5 rounded-xl border border-rose-500/20 bg-rose-500/[0.06] px-4 py-3 text-sm text-rose-200">
          <AlertTriangle size={16} className="shrink-0" />
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 gap-4">
        {keys.map((k) => (
          <article
            key={k.key}
            className="group/card rounded-2xl border border-white/[0.06] bg-white/[0.02] p-5 md:p-6 hover:border-white/[0.1] hover:bg-white/[0.03] transition-all duration-200"
          >
            <div className="space-y-5">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="flex min-w-0 gap-3">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-white/[0.06] bg-white/[0.03] text-emerald-400/80">
                    <Key size={18} />
                  </div>
                  <div className="min-w-0 space-y-1">
                    <h3 className="text-sm font-medium text-white">
                      <span className="text-white/50">{t('api_keys_manager.key_type_tunnel')}</span>
                      <span className="mx-1.5 text-white/15">&middot;</span>
                      <span className="font-mono text-emerald-400/80">{k.key.substring(0, 14)}...</span>
                    </h3>
                    <div className="flex items-center gap-4 text-xs text-white/30">
                      <span className="inline-flex items-center gap-1"><Clock size={11} /> {new Date(k.created_at).toLocaleDateString()}</span>
                      <span className="inline-flex items-center gap-1"><CheckCircle2 size={11} className="text-emerald-500/60" /> {k.usage_count.toLocaleString()} {t('api_keys_manager.requests_label')}</span>
                    </div>
                  </div>
                </div>
                <Tooltip label={t('api_keys_manager.tooltip_revoke')}>
                  <button
                    type="button"
                    onClick={() => setRevokeKey(k.key)}
                    className="rounded-lg p-2 text-white/25 transition hover:bg-rose-500/10 hover:text-rose-400"
                  >
                    <Trash2 size={16} />
                  </button>
                </Tooltip>
              </div>

              <div className="space-y-3 border-t border-white/[0.04] pt-4">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <span className="inline-flex items-center gap-1.5 text-[10px] font-medium text-emerald-500/60">
                    <span className="h-1 w-1 rounded-full bg-emerald-500 shadow-[0_0_4px_rgba(16,185,129,0.5)]" />
                    {t('api_keys_manager.magic_command_section')}
                  </span>
                  <div className="flex flex-wrap gap-1">
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
                        className="rounded-md border border-white/[0.06] bg-white/[0.03] px-2.5 py-1 text-[10px] font-medium text-white/45 transition hover:border-emerald-500/20 hover:bg-emerald-500/[0.08] hover:text-emerald-300"
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
                  className="group/cmd relative w-full rounded-xl border border-white/[0.06] bg-black/30 p-3.5 text-left font-mono text-[11px] leading-relaxed text-white/60 transition hover:border-emerald-500/15 hover:bg-black/40"
                  onClick={() => copyToClipboard(installCmd(k.key, true))}
                >
                  <div className="flex items-start gap-2.5 pr-4">
                    <span className="shrink-0 text-emerald-500/70">$</span>
                    <code className="min-w-0 flex-1 break-all">{installCmd(k.key, true)}</code>
                  </div>
                  <div className="pointer-events-none absolute right-3 top-1/2 flex -translate-y-1/2 items-center gap-1 rounded-md bg-emerald-500 px-2 py-1 text-[10px] font-semibold text-[#080a10] opacity-0 transition group-hover/cmd:opacity-100">
                    <Copy size={11} /> {t('api_keys_manager.copy_label')}
                  </div>
                </button>
                <p className="text-[11px] text-white/25">{t('api_keys_manager.magic_command_note')}</p>
              </div>
            </div>
          </article>
        ))}

        {keys.length === 0 && !loading && (
          <div className="py-14 text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-xl border border-dashed border-white/[0.08] bg-white/[0.02]">
              <Shield className="text-white/15" size={24} />
            </div>
            <p className="text-sm text-white/35">{t('api_keys_manager.empty')}</p>
          </div>
        )}
      </div>
    </div>
  );
};
