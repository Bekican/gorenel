import React, { useEffect, useMemo, useState } from 'react';
import { Shield, KeyRound, Copy, RotateCcw, X, CheckCircle2, AlertTriangle, Lock, Gauge, ArrowRightLeft, CornerDownRight, Plus } from 'lucide-react';
import { api, type Tunnel } from '../api/client';
import { Tooltip } from './ui/Tooltip';

type Props = {
  open: boolean;
  onClose: () => void;
  tunnel: Tunnel | null;
  onUpdated: () => void;
};

function splitAllowlist(text: string): string[] {
  return text
    .split(/[\s,]+/g)
    .map((s) => s.trim())
    .filter(Boolean);
}

export const TunnelPolicyModal: React.FC<Props> = ({ open, onClose, tunnel, onUpdated }) => {
  const [toast, setToast] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const [ipText, setIpText] = useState('');
  const [ipEnabled, setIpEnabled] = useState(false);
  const [keyEnabled, setKeyEnabled] = useState(false);
  const [freshToken, setFreshToken] = useState<string | null>(null);
  const [basicEnabled, setBasicEnabled] = useState(false);
  const [basicUser, setBasicUser] = useState('');
  const [basicPass, setBasicPass] = useState('');
  const [httpsRedirect, setHttpsRedirect] = useState(false);
  const [rateEnabled, setRateEnabled] = useState(false);
  const [rateReq, setRateReq] = useState(60);
  const [rateWin, setRateWin] = useState(60);
  const [pathPrefix, setPathPrefix] = useState('');
  const [replaceFrom, setReplaceFrom] = useState('');
  const [replaceTo, setReplaceTo] = useState('');
  const [addReqHeadersText, setAddReqHeadersText] = useState('');
  const [removeReqHeadersText, setRemoveReqHeadersText] = useState('');
  const [addRespHeadersText, setAddRespHeadersText] = useState('');
  const [removeRespHeadersText, setRemoveRespHeadersText] = useState('');

  const currentPolicy = tunnel?.policy || {};

  useEffect(() => {
    if (!open || !tunnel) return;
    setActionError(null);
    setFreshToken(null);
    setIpEnabled(!!currentPolicy.ip_allowlist_enabled);
    setKeyEnabled(!!currentPolicy.key_auth_enabled);
    setBasicEnabled(!!currentPolicy.basic_auth_enabled);
    setBasicUser(currentPolicy.basic_auth_username || '');
    setBasicPass('');
    setHttpsRedirect(!!currentPolicy.https_redirect_enabled);
    setRateEnabled(!!currentPolicy.rate_limit_enabled);
    setRateReq(currentPolicy.rate_limit_requests || 60);
    setRateWin(currentPolicy.rate_limit_window_s || 60);
    setPathPrefix(currentPolicy.path_prefix || '');
    setReplaceFrom(currentPolicy.replace_path_from || '');
    setReplaceTo(currentPolicy.replace_path_to || '');
    setIpText('');
    setAddReqHeadersText('');
    setRemoveReqHeadersText('');
    setAddRespHeadersText('');
    setRemoveRespHeadersText('');
  }, [open, tunnel?.subdomain]);

  useEffect(() => {
    if (!toast) return;
    const id = window.setTimeout(() => setToast(null), 2800);
    return () => window.clearTimeout(id);
  }, [toast]);

  const allowlist = useMemo(() => splitAllowlist(ipText), [ipText]);

  const parseHeaderMap = (text: string): Record<string, string> => {
    const out: Record<string, string> = {};
    text
      .split('\n')
      .map((l) => l.trim())
      .filter(Boolean)
      .forEach((line) => {
        const idx = line.indexOf(':');
        if (idx <= 0) return;
        const k = line.slice(0, idx).trim();
        const v = line.slice(idx + 1).trim();
        if (k) out[k] = v;
      });
    return out;
  };
  const parseHeaderList = (text: string): string[] =>
    text
      .split(/[\s,]+/g)
      .map((s) => s.trim())
      .filter(Boolean);

  if (!open || !tunnel) return null;

  const copy = (text: string) => {
    navigator.clipboard.writeText(text);
    setToast('Copied to clipboard');
  };

  const rotateToken = async () => {
    setActionError(null);
    setLoading(true);
    try {
      const res = await api.rotateTunnelToken(tunnel.subdomain);
      setFreshToken(res.token);
      setKeyEnabled(true);
      onUpdated();
      setToast('New token generated');
    } catch (e) {
      console.error(e);
      setActionError('Failed to generate token. Are you logged in?');
    } finally {
      setLoading(false);
    }
  };

  const save = async () => {
    setActionError(null);
    setLoading(true);
    try {
      await api.updateTunnelPolicy(tunnel.subdomain, {
        key_auth_enabled: keyEnabled,
        ip_allowlist_enabled: ipEnabled,
        ip_allowlist: ipEnabled ? allowlist : [],
        basic_auth_enabled: basicEnabled,
        basic_auth_username: basicUser || undefined,
        basic_auth_password: basicPass || undefined,
        https_redirect_enabled: httpsRedirect,
        rate_limit_enabled: rateEnabled,
        rate_limit_requests: rateEnabled ? rateReq : undefined,
        rate_limit_window_s: rateEnabled ? rateWin : undefined,
        path_prefix: pathPrefix || undefined,
        replace_path_from: replaceFrom || undefined,
        replace_path_to: replaceTo || undefined,
        add_request_headers: parseHeaderMap(addReqHeadersText),
        remove_request_headers: parseHeaderList(removeReqHeadersText),
        add_response_headers: parseHeaderMap(addRespHeadersText),
        remove_response_headers: parseHeaderList(removeRespHeadersText),
      });
      onUpdated();
      setToast('Policy saved');
      if (!keyEnabled) setFreshToken(null);
      setBasicPass('');
    } catch (e) {
      console.error(e);
      setActionError('Failed to save policy. Check IP/CIDR formats.');
    } finally {
      setLoading(false);
    }
  };

  const curlExample =
    freshToken && tunnel.publicUrl
      ? `curl "${tunnel.publicUrl}" -H "X-TOKEN: ${freshToken}"`
      : tunnel.publicUrl
        ? `curl "${tunnel.publicUrl}" -H "X-TOKEN: <TOKEN>"`
        : '';

  return (
    <div className="fixed inset-0 z-[120] flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-in fade-in duration-200">
      <div
        className="bg-[#0A0A0A] border border-white/10 w-full max-w-2xl rounded-3xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200 relative"
        role="dialog"
        aria-modal="true"
        aria-label="Tunnel security policy"
      >
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-2 hover:bg-white/10 rounded-full transition-all text-white/50 hover:text-white"
          aria-label="Close"
        >
          <X className="w-5 h-5" />
        </button>

        <div className="p-8 space-y-8">
          <div className="flex items-start gap-5">
            <div className="w-14 h-14 bg-primary/15 rounded-2xl flex items-center justify-center ring-1 ring-primary/35 shadow-[0_0_30px_-10px_rgba(16,185,129,0.28)]">
              <Shield className="w-7 h-7 text-primary" />
            </div>
            <div className="min-w-0">
              <h2 className="text-2xl font-black text-white tracking-tight">Security Policy</h2>
              <p className="text-white/50 text-sm font-medium">
                {tunnel.subdomain}.gorenel.site — protect public traffic before it reaches your local server.
              </p>
            </div>
          </div>

          {toast && (
            <div className="flex items-center gap-3 rounded-2xl border border-emerald-500/25 bg-[#0c1210]/95 px-5 py-3 text-sm font-semibold text-emerald-100 shadow-[0_12px_40px_rgba(0,0,0,0.45)] backdrop-blur-md animate-in fade-in slide-in-from-top-2 duration-300">
              <CheckCircle2 className="h-4 w-4 shrink-0 text-emerald-400" aria-hidden />
              {toast}
            </div>
          )}

          {actionError && (
            <div className="flex items-center gap-3 rounded-2xl border border-rose-500/25 bg-rose-500/10 px-4 py-3 text-sm text-rose-200/95">
              <AlertTriangle size={18} className="shrink-0 text-rose-400" />
              <span className="min-w-0">{actionError}</span>
            </div>
          )}

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="rounded-3xl border border-white/10 bg-white/[0.03] p-6 space-y-5">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-2xl bg-white/[0.04] border border-white/10">
                    <KeyRound className="w-4 h-4 text-white/70" />
                  </div>
                  <div>
                    <div className="text-sm font-black text-white tracking-tight">KeyAuth (X-TOKEN)</div>
                    <div className="text-[11px] text-white/40 font-medium">Simple token gate for public requests.</div>
                  </div>
                </div>
                <label className="inline-flex items-center gap-2 cursor-pointer select-none">
                  <input
                    type="checkbox"
                    checked={keyEnabled}
                    onChange={(e) => setKeyEnabled(e.target.checked)}
                    className="h-4 w-4 accent-emerald-400"
                  />
                  <span className="text-[11px] font-black uppercase tracking-widest text-white/50">
                    {keyEnabled ? 'ON' : 'OFF'}
                  </span>
                </label>
              </div>

              <div className="flex flex-wrap gap-2">
                <Tooltip label="Generates a new token and enables KeyAuth. Token is shown once.">
                  <button
                    type="button"
                    onClick={rotateToken}
                    disabled={loading}
                    className="inline-flex items-center justify-center gap-2 rounded-xl border border-emerald-500/25 bg-emerald-500/10 px-4 py-2.5 text-sm font-bold text-emerald-100 transition hover:bg-emerald-500/20 disabled:opacity-50"
                  >
                    <RotateCcw className="w-4 h-4" />
                    {freshToken ? 'Rotate again' : 'Generate / Rotate'}
                  </button>
                </Tooltip>

                {freshToken && (
                  <button
                    type="button"
                    onClick={() => copy(freshToken)}
                    className="inline-flex items-center justify-center gap-2 rounded-xl border border-white/10 bg-white/[0.03] px-4 py-2.5 text-sm font-bold text-white/80 transition hover:bg-white/[0.06]"
                  >
                    <Copy className="w-4 h-4" />
                    Copy token
                  </button>
                )}
              </div>

              <div className="rounded-2xl border border-white/10 bg-black/40 p-4 space-y-2">
                <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/30">Curl example</div>
                <button
                  type="button"
                  onClick={() => curlExample && copy(curlExample)}
                  className="w-full text-left font-mono text-[11px] text-primary break-all hover:text-emerald-300 transition"
                >
                  {curlExample || '—'}
                </button>
                <div className="text-[11px] text-white/35 font-medium">
                  Your clients must send header <span className="font-mono text-white/70">X-TOKEN</span>.
                </div>
              </div>
            </div>

            <div className="rounded-3xl border border-white/10 bg-white/[0.03] p-6 space-y-5">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-sm font-black text-white tracking-tight">IP allowlist</div>
                  <div className="text-[11px] text-white/40 font-medium">Allow only specific IPs or CIDRs.</div>
                </div>
                <label className="inline-flex items-center gap-2 cursor-pointer select-none">
                  <input
                    type="checkbox"
                    checked={ipEnabled}
                    onChange={(e) => setIpEnabled(e.target.checked)}
                    className="h-4 w-4 accent-emerald-400"
                  />
                  <span className="text-[11px] font-black uppercase tracking-widest text-white/50">
                    {ipEnabled ? 'ON' : 'OFF'}
                  </span>
                </label>
              </div>

              <div className="space-y-2">
                <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/30">
                  Entries (space/comma separated)
                </div>
                <textarea
                  value={ipText}
                  onChange={(e) => setIpText(e.target.value)}
                  placeholder="1.2.3.4, 10.0.0.0/24"
                  className="w-full h-28 px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono focus:ring-2 focus:ring-primary/20 focus:border-primary/50 transition-all outline-none text-white/80 resize-none disabled:opacity-50"
                  disabled={!ipEnabled}
                />
                <div className="flex flex-wrap gap-2 pt-1">
                  {ipEnabled && allowlist.slice(0, 8).map((v) => (
                    <span
                      key={v}
                      className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.03] px-3 py-1 text-[10px] font-black uppercase tracking-widest text-white/60"
                    >
                      <Plus className="w-3 h-3 text-white/20" />
                      {v}
                    </span>
                  ))}
                  {ipEnabled && allowlist.length > 8 && (
                    <span className="text-[10px] font-black uppercase tracking-widest text-white/25">
                      +{allowlist.length - 8} more
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="rounded-3xl border border-white/10 bg-white/[0.03] p-6 space-y-5">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-2xl bg-white/[0.04] border border-white/10">
                    <Lock className="w-4 h-4 text-white/70" />
                  </div>
                  <div>
                    <div className="text-sm font-black text-white tracking-tight">Basic Auth</div>
                    <div className="text-[11px] text-white/40 font-medium">Browser-native username/password gate.</div>
                  </div>
                </div>
                <label className="inline-flex items-center gap-2 cursor-pointer select-none">
                  <input
                    type="checkbox"
                    checked={basicEnabled}
                    onChange={(e) => setBasicEnabled(e.target.checked)}
                    className="h-4 w-4 accent-emerald-400"
                  />
                  <span className="text-[11px] font-black uppercase tracking-widest text-white/50">
                    {basicEnabled ? 'ON' : 'OFF'}
                  </span>
                </label>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <input
                  value={basicUser}
                  onChange={(e) => setBasicUser(e.target.value)}
                  disabled={!basicEnabled}
                  placeholder="username"
                  className="w-full px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80 disabled:opacity-50"
                />
                <input
                  value={basicPass}
                  onChange={(e) => setBasicPass(e.target.value)}
                  disabled={!basicEnabled}
                  placeholder="password (set to update)"
                  type="password"
                  className="w-full px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80 disabled:opacity-50"
                />
              </div>
              <div className="text-[11px] text-white/35 font-medium">
                Password is stored hashed and never shown again.
              </div>
            </div>

            <div className="rounded-3xl border border-white/10 bg-white/[0.03] p-6 space-y-5">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-2xl bg-white/[0.04] border border-white/10">
                    <Gauge className="w-4 h-4 text-white/70" />
                  </div>
                  <div>
                    <div className="text-sm font-black text-white tracking-tight">Rate limit</div>
                    <div className="text-[11px] text-white/40 font-medium">Per-tunnel, per-client IP sliding window.</div>
                  </div>
                </div>
                <label className="inline-flex items-center gap-2 cursor-pointer select-none">
                  <input
                    type="checkbox"
                    checked={rateEnabled}
                    onChange={(e) => setRateEnabled(e.target.checked)}
                    className="h-4 w-4 accent-emerald-400"
                  />
                  <span className="text-[11px] font-black uppercase tracking-widest text-white/50">
                    {rateEnabled ? 'ON' : 'OFF'}
                  </span>
                </label>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1">
                  <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/30">Requests</div>
                  <input
                    value={rateReq}
                    onChange={(e) => setRateReq(Number(e.target.value || 0))}
                    disabled={!rateEnabled}
                    type="number"
                    min={1}
                    className="w-full px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80 disabled:opacity-50"
                  />
                </div>
                <div className="space-y-1">
                  <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/30">Window (s)</div>
                  <input
                    value={rateWin}
                    onChange={(e) => setRateWin(Number(e.target.value || 0))}
                    disabled={!rateEnabled}
                    type="number"
                    min={1}
                    className="w-full px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80 disabled:opacity-50"
                  />
                </div>
              </div>
              <label className="inline-flex items-center gap-2 cursor-pointer select-none pt-2">
                <input
                  type="checkbox"
                  checked={httpsRedirect}
                  onChange={(e) => setHttpsRedirect(e.target.checked)}
                  className="h-4 w-4 accent-emerald-400"
                />
                <span className="text-[11px] font-black uppercase tracking-widest text-white/50">
                  HTTPS redirect
                </span>
              </label>
            </div>
          </div>

          <div className="rounded-3xl border border-white/10 bg-white/[0.03] p-6 space-y-5">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-2xl bg-white/[0.04] border border-white/10">
                <ArrowRightLeft className="w-4 h-4 text-white/70" />
              </div>
              <div>
                <div className="text-sm font-black text-white tracking-tight">Rewrite & headers</div>
                <div className="text-[11px] text-white/40 font-medium">Lightweight request/response shaping per tunnel.</div>
              </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
              <div className="lg:col-span-1 space-y-2">
                <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/30 flex items-center gap-2"><CornerDownRight size={12} /> Path prefix</div>
                <input
                  value={pathPrefix}
                  onChange={(e) => setPathPrefix(e.target.value)}
                  placeholder="/api"
                  className="w-full px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80"
                />
                <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/30">Replace path from/to</div>
                <input
                  value={replaceFrom}
                  onChange={(e) => setReplaceFrom(e.target.value)}
                  placeholder="/v1"
                  className="w-full px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80"
                />
                <input
                  value={replaceTo}
                  onChange={(e) => setReplaceTo(e.target.value)}
                  placeholder="/v2"
                  className="w-full px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80"
                />
              </div>
              <div className="lg:col-span-2 grid grid-cols-1 md:grid-cols-2 gap-3">
                <div className="space-y-2">
                  <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/30">Add request headers (one per line: Key: Value)</div>
                  <textarea value={addReqHeadersText} onChange={(e) => setAddReqHeadersText(e.target.value)} className="w-full h-24 px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80 resize-none" />
                  <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/30">Remove request headers (space/comma)</div>
                  <input value={removeReqHeadersText} onChange={(e) => setRemoveReqHeadersText(e.target.value)} className="w-full px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80" />
                </div>
                <div className="space-y-2">
                  <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/30">Add response headers (one per line: Key: Value)</div>
                  <textarea value={addRespHeadersText} onChange={(e) => setAddRespHeadersText(e.target.value)} className="w-full h-24 px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80 resize-none" />
                  <div className="text-[10px] font-black uppercase tracking-[0.2em] text-white/30">Remove response headers (space/comma)</div>
                  <input value={removeRespHeadersText} onChange={(e) => setRemoveRespHeadersText(e.target.value)} className="w-full px-4 py-3 bg-black/50 border border-white/10 rounded-2xl text-xs font-mono outline-none text-white/80" />
                </div>
              </div>
            </div>
          </div>

          <div className="flex flex-col sm:flex-row gap-3 sm:justify-end pt-2 border-t border-white/5">
            <button
              type="button"
              onClick={onClose}
              className="rounded-xl border border-white/10 bg-white/[0.03] px-5 py-3 text-sm font-semibold text-white/80 transition hover:bg-white/[0.06]"
              disabled={loading}
            >
              Close
            </button>
            <button
              type="button"
              onClick={save}
              disabled={loading}
              className="btn-primary-premium px-6 py-3 text-sm font-bold inline-flex items-center justify-center gap-2 disabled:opacity-50"
            >
              <Shield className="w-4 h-4" /> Save policy
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

