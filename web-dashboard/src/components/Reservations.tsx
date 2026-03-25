import React, { useEffect, useMemo, useState } from 'react';
import { Plus, Trash2, Link2, Unlink2, RefreshCcw } from 'lucide-react';
import { api, type ReservedSubdomain } from '../api/client';
import { Button } from './ui/Button';
import { Input } from './ui/Input';
import { Alert } from './ui/Alert';

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
          <Button type="button" variant="secondary" size="sm" onClick={refresh}>
            <RefreshCcw size={14} /> Refresh
          </Button>
        </div>

        <div className="flex flex-col md:flex-row gap-3">
          <Input
            value={newSub}
            onChange={(e) => setNewSub(e.target.value)}
            placeholder="my-device-01"
            className="flex-1"
          />
          <Button onClick={reserve} variant="primary" size="lg" type="button" className="md:w-auto w-full">
            <Plus size={18} /> Reserve
          </Button>
        </div>

        {error && (
          <Alert variant="error" title="Action failed">{error}</Alert>
        )}

        <Alert variant="info" title="How it works">
          Reserve a subdomain here, then start your tunnel with <span className="font-mono text-white/80">--subdomain</span>. If you assign it to an API key, only that key can use it.
        </Alert>
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
                    <Input
                      value={assignKey[r.subdomain] ?? ''}
                      onChange={(e) => setAssignKey((p) => ({ ...p, [r.subdomain]: e.target.value }))}
                      placeholder="Assign to API key (optional)"
                      className="w-full sm:w-[320px] text-xs"
                    />
                    <Button
                      type="button"
                      onClick={() => assign(r.subdomain, (assignKey[r.subdomain] ?? '').trim() || null)}
                      variant="secondary"
                      size="sm"
                    >
                      <Link2 size={14} /> Assign
                    </Button>
                    <Button
                      type="button"
                      onClick={() => assign(r.subdomain, null)}
                      variant="ghost"
                      size="sm"
                    >
                      <Unlink2 size={14} /> Unassign
                    </Button>
                    <Button
                      type="button"
                      onClick={() => release(r.subdomain)}
                      variant="danger"
                      size="sm"
                    >
                      <Trash2 size={14} /> Release
                    </Button>
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

