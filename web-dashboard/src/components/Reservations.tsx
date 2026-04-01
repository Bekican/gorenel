import React, { useEffect, useMemo, useState } from 'react';
import { Plus, Trash2, Link2, Unlink2, RefreshCcw, Sparkles, Globe, ShieldCheck } from 'lucide-react';
import { api, type ReservedSubdomain } from '../api/client';
import { useTranslation } from 'react-i18next';
import { Button } from './ui/Button';
import { Input } from './ui/Input';
import { Alert } from './ui/Alert';

export const Reservations: React.FC = () => {
  const { t } = useTranslation();
  const [items, setItems] = useState<ReservedSubdomain[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newSub, setNewSub] = useState('');
  const [assignKey, setAssignKey] = useState<Record<string, string>>({});

  const sorted = useMemo(() => {
    return [...items].sort((a, b) => (a.created_at < b.created_at ? 1 : -1));
  }, [items]);

  const refresh = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listReservations();
      setItems(res.reservations || []);
    } catch {
      setError('Failed to load reservations');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    refresh();
  }, []);

  const reserve = async () => {
    const s = newSub.trim();
    if (!s) return;
    setError(null);
    try {
      await api.reserveSubdomain(s);
      setNewSub('');
      await refresh();
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } };
      setError((typeof err.response?.data === 'string' && err.response.data) || 'Reserve failed');
    }
  };

  const release = async (subdomain: string) => {
    setError(null);
    try {
      await api.releaseSubdomain(subdomain);
      await refresh();
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } };
      setError((typeof err.response?.data === 'string' && err.response.data) || 'Release failed');
    }
  };

  const assign = async (subdomain: string, apiKey: string | null) => {
    setError(null);
    try {
      await api.assignReservationToKey(subdomain, apiKey);
      await refresh();
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } };
      setError((typeof err.response?.data === 'string' && err.response.data) || 'Assign failed');
    }
  };

  return (
    <div className="space-y-5">
      <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 md:p-8 space-y-4">
        <div className="flex items-center justify-between gap-4">
          <div>
            <h3 className="text-lg font-semibold text-white">{t('reservations.title')}</h3>
            <p className="text-sm text-white/35">
              {t('reservations.subtitle')}
            </p>
          </div>
          <Button type="button" variant="secondary" size="sm" onClick={refresh}>
            <RefreshCcw size={13} /> {t('common.loading')}
          </Button>
        </div>

        <div className="relative overflow-hidden bg-violet-500/[0.03] border border-violet-500/10 rounded-2xl p-6 group">
          <div className="relative z-10 flex flex-col md:flex-row md:items-start justify-between gap-6">
            <div className="space-y-2">
              <h4 className="text-sm font-semibold text-violet-400 flex items-center gap-2">
                <Sparkles className="w-4 h-4" />
                {t('reservations.onboarding_title')}
              </h4>
              <p className="text-xs text-white/40 leading-relaxed max-w-xl">
                {t('reservations.onboarding_desc')}
              </p>
            </div>
            <div className="grid grid-cols-2 gap-3 shrink-0">
               <div className="p-3 bg-white/[0.02] border border-white/[0.04] rounded-xl space-y-1">
                 <Globe className="w-3.5 h-3.5 text-violet-400/50" />
                 <div className="text-[10px] font-medium text-white/40">{t('reservations.stable_endpoints')}</div>
               </div>
               <div className="p-3 bg-white/[0.02] border border-white/[0.04] rounded-xl space-y-1">
                 <ShieldCheck className="w-3.5 h-3.5 text-emerald-400/50" />
                 <div className="text-[10px] font-medium text-white/40">{t('reservations.reserve_btn')}</div>
               </div>
            </div>
          </div>
          <div className="mt-4 flex items-center gap-2 text-[10px] text-white/20 bg-white/[0.02] w-fit px-2.5 py-1.5 rounded-lg border border-white/[0.04]">
            <span className="font-mono text-violet-400/60">Tip:</span> {t('reservations.usage_guide')}
          </div>
          <div className="absolute -right-8 -bottom-8 w-32 h-32 bg-violet-500/5 blur-[40px] rounded-full group-hover:bg-violet-500/10 transition-colors duration-500" />
        </div>

        <div className="flex flex-col md:flex-row gap-2.5">
          <Input
            value={newSub}
            onChange={(e) => setNewSub(e.target.value)}
            placeholder="my-device-01"
            className="flex-1"
          />
          <Button onClick={reserve} variant="primary" size="lg" type="button" className="md:w-auto w-full">
            <Plus size={16} /> Reserve
          </Button>
        </div>

        {error && <Alert variant="error" title="Action failed">{error}</Alert>}

        <Alert variant="info" title="How it works">
          Reserve a subdomain, then start with <span className="font-mono text-white/70">--subdomain</span>. Assign to an API key for exclusive use.
        </Alert>
      </div>

      <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 md:p-8">
        {loading ? (
          <div className="text-sm text-white/35">Loading...</div>
        ) : sorted.length === 0 ? (
          <div className="text-sm text-white/35">No reserved subdomains yet.</div>
        ) : (
          <div className="space-y-3">
            {sorted.map((r) => (
              <div key={r.subdomain} className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5 hover:border-white/[0.1] transition-colors">
                <div className="flex flex-col md:flex-row md:items-center gap-3 justify-between">
                  <div className="min-w-0">
                    <div className="font-medium text-white">{r.subdomain}.gorenel.site</div>
                    <div className="text-[11px] text-white/30">
                      Created: {new Date(r.created_at).toLocaleString()} {r.last_used_at ? `· Last used: ${new Date(r.last_used_at).toLocaleString()}` : ''}
                    </div>
                  </div>
                  <div className="flex flex-col sm:flex-row gap-2 sm:items-center shrink-0">
                    <Input
                      value={assignKey[r.subdomain] ?? ''}
                      onChange={(e) => setAssignKey((p) => ({ ...p, [r.subdomain]: e.target.value }))}
                      placeholder="API key (optional)"
                      className="w-full sm:w-[260px] text-xs"
                    />
                    <div className="flex gap-1.5">
                      <Button type="button" onClick={() => assign(r.subdomain, (assignKey[r.subdomain] ?? '').trim() || null)} variant="secondary" size="sm">
                        <Link2 size={13} /> Assign
                      </Button>
                      <Button type="button" onClick={() => assign(r.subdomain, null)} variant="ghost" size="sm">
                        <Unlink2 size={13} />
                      </Button>
                      <Button type="button" onClick={() => release(r.subdomain)} variant="danger" size="sm">
                        <Trash2 size={13} />
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
