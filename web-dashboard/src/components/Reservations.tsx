import React, { useEffect, useMemo, useState } from 'react';
import { Plus, Trash2, Link2, Unlink2, RefreshCcw } from 'lucide-react';
import { api, type ReservedSubdomain } from '../api/client';

export const Reservations: React.FC = () => {
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
    } catch (e) {
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
    } catch (e: any) {
      setError(e?.response?.data || 'Reserve failed');
    }
  };

  const release = async (subdomain: string) => {
    setError(null);
    try {
      await api.releaseSubdomain(subdomain);
      await refresh();
    } catch (e: any) {
      setError(e?.response?.data || 'Release failed');
    }
  };

  const assign = async (subdomain: string, apiKey: string | null) => {
    setError(null);
    try {
      await api.assignReservationToKey(subdomain, apiKey);
      await refresh();
    } catch (e: any) {
      setError(e?.response?.data || 'Assign failed');
    }
  };

  return (
    <div className="space-y-6">
      <div className="card p-6 md:p-8 space-y-4">
        <div className="flex items-center justify-between gap-4">
          <div>
            <h3 className="text-xl font-black tracking-tight text-white">Reserved URLs</h3>
            <p className="text-sm text-white/40 font-medium">
              Reserve stable subdomains for devices/customers. Use them via CLI <span className="font-mono text-white/60">--subdomain</span>.
            </p>
          </div>
          <button type="button" onClick={refresh} className="rounded-xl border border-white/10 bg-white/[0.03] px-4 py-2 text-xs font-bold text-white/60 hover:bg-white/[0.06]">
            <span className="inline-flex items-center gap-2"><RefreshCcw size={14} /> Refresh</span>
          </button>
        </div>

        <div className="flex flex-col md:flex-row gap-3">
          <input
            value={newSub}
            onChange={(e) => setNewSub(e.target.value)}
            placeholder="my-device-01"
            className="flex-1 rounded-2xl border border-white/10 bg-black/40 px-4 py-3 text-sm text-white/80 focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
          />
          <button onClick={reserve} className="btn-primary-premium inline-flex items-center justify-center gap-2 px-5 py-3 text-sm font-black">
            <Plus size={18} /> Reserve
          </button>
        </div>

        {error && (
          <div className="rounded-2xl border border-rose-500/25 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">
            {error}
          </div>
        )}
      </div>

      <div className="card p-6 md:p-8">
        {loading ? (
          <div className="text-sm text-white/40">Loading…</div>
        ) : sorted.length === 0 ? (
          <div className="text-sm text-white/40">No reserved subdomains yet.</div>
        ) : (
          <div className="space-y-4">
            {sorted.map((r) => (
              <div key={r.subdomain} className="rounded-3xl border border-white/10 bg-white/[0.02] p-5 md:p-6">
                <div className="flex flex-col md:flex-row md:items-center gap-4 justify-between">
                  <div className="min-w-0">
                    <div className="font-black text-white truncate">{r.subdomain}.gorenel.site</div>
                    <div className="text-xs text-white/35">
                      Created: {new Date(r.created_at).toLocaleString()} {r.last_used_at ? `· Last used: ${new Date(r.last_used_at).toLocaleString()}` : ''}
                    </div>
                  </div>
                  <div className="flex flex-col sm:flex-row gap-2 sm:items-center">
                    <input
                      value={assignKey[r.subdomain] ?? ''}
                      onChange={(e) => setAssignKey((p) => ({ ...p, [r.subdomain]: e.target.value }))}
                      placeholder="Assign to API key (optional)"
                      className="w-full sm:w-[320px] rounded-2xl border border-white/10 bg-black/40 px-4 py-2.5 text-xs text-white/80 focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
                    />
                    <button
                      type="button"
                      onClick={() => assign(r.subdomain, (assignKey[r.subdomain] ?? '').trim() || null)}
                      className="inline-flex items-center justify-center gap-2 rounded-2xl border border-emerald-500/25 bg-emerald-500/10 px-4 py-2.5 text-xs font-black text-emerald-200 hover:bg-emerald-500/15"
                    >
                      <Link2 size={14} /> Assign
                    </button>
                    <button
                      type="button"
                      onClick={() => assign(r.subdomain, null)}
                      className="inline-flex items-center justify-center gap-2 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-2.5 text-xs font-black text-white/60 hover:bg-white/[0.06]"
                    >
                      <Unlink2 size={14} /> Unassign
                    </button>
                    <button
                      type="button"
                      onClick={() => release(r.subdomain)}
                      className="inline-flex items-center justify-center gap-2 rounded-2xl border border-rose-500/25 bg-rose-500/10 px-4 py-2.5 text-xs font-black text-rose-200 hover:bg-rose-500/15"
                    >
                      <Trash2 size={14} /> Release
                    </button>
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

