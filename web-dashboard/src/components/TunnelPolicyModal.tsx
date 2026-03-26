import React, { useEffect, useMemo, useState } from 'react';
import { Shield, KeyRound, Copy, RotateCcw, X, Lock, Gauge, ArrowRightLeft, CornerDownRight } from 'lucide-react';
import { api, type Tunnel } from '../api/client';
import { Tooltip } from './ui/Tooltip';
import { Tabs } from './ui/Tabs';
import { Button } from './ui/Button';
import { Input } from './ui/Input';
import { Alert } from './ui/Alert';

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
  const [tab, setTab] = useState<'access' | 'limits' | 'rewrite'>('access');

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
    setTab('access');
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

  const ToggleRow = ({ label, desc, checked, onChange, icon: Icon }: { label: string; desc: string; checked: boolean; onChange: (v: boolean) => void; icon: any }) => (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-3">
        <div className="p-2 rounded-lg bg-white/[0.03] border border-white/[0.06]">
          <Icon className="w-4 h-4 text-white/50" />
        </div>
        <div>
          <div className="text-sm font-medium text-white">{label}</div>
          <div className="text-[11px] text-white/35">{desc}</div>
        </div>
      </div>
      <button
        type="button"
        onClick={() => onChange(!checked)}
        className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors duration-200 ${checked ? 'bg-emerald-500' : 'bg-white/10'}`}
      >
        <span className={`inline-block h-3.5 w-3.5 rounded-full bg-white transition-transform duration-200 ${checked ? 'translate-x-[18px]' : 'translate-x-[3px]'}`} />
      </button>
    </div>
  );

  const SectionLabel = ({ children }: { children: React.ReactNode }) => (
    <div className="text-[11px] font-medium text-white/25 uppercase tracking-wider">{children}</div>
  );

  return (
    <div className="fixed inset-0 z-[120] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
      <div
        className="bg-[#0d0f14] border border-white/[0.08] w-full max-w-2xl rounded-2xl shadow-modal overflow-hidden animate-scale-in relative max-h-[90vh] flex flex-col"
        role="dialog"
        aria-modal="true"
        aria-label="Tunnel security policy"
      >
        <button
          onClick={onClose}
          className="absolute top-4 right-4 z-10 p-1.5 hover:bg-white/[0.06] rounded-lg transition-all text-white/40 hover:text-white"
          aria-label="Close"
        >
          <X className="w-4 h-4" />
        </button>

        <div className="p-7 space-y-6 overflow-y-auto">
          <div className="flex items-start gap-4">
            <div className="w-11 h-11 bg-emerald-500/10 rounded-xl flex items-center justify-center border border-emerald-500/15 shrink-0">
              <Shield className="w-5 h-5 text-emerald-400" />
            </div>
            <div className="min-w-0">
              <h2 className="text-lg font-semibold text-white tracking-tight">Security Policy</h2>
              <p className="text-sm text-white/40">
                {tunnel.subdomain}.gorenel.site — protect traffic before it reaches your server.
              </p>
            </div>
          </div>

          {toast && <Alert variant="success" title="Done">{toast}</Alert>}
          {actionError && <Alert variant="error" title="Error">{actionError}</Alert>}

          <Tabs
            value={tab}
            onChange={setTab}
            tabs={[
              { id: 'access', label: 'Access' },
              { id: 'limits', label: 'Limits' },
              { id: 'rewrite', label: 'Rewrite' },
            ]}
          />

          {tab === 'access' && (
            <div className="space-y-4">
              {/* KeyAuth */}
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5 space-y-4">
                <ToggleRow label="KeyAuth (X-TOKEN)" desc="Token gate for public requests." checked={keyEnabled} onChange={setKeyEnabled} icon={KeyRound} />

                <div className="flex flex-wrap gap-2">
                  <Tooltip label="Generates a new token and enables KeyAuth.">
                    <Button type="button" onClick={rotateToken} disabled={loading} variant="secondary" size="sm">
                      <RotateCcw className="w-3.5 h-3.5" />
                      {freshToken ? 'Rotate again' : 'Generate token'}
                    </Button>
                  </Tooltip>
                  {freshToken && (
                    <Button type="button" onClick={() => copy(freshToken)} variant="outline" size="sm">
                      <Copy className="w-3.5 h-3.5" /> Copy token
                    </Button>
                  )}
                </div>

                <div className="rounded-lg border border-white/[0.06] bg-black/20 p-3 space-y-1.5">
                  <SectionLabel>Curl example</SectionLabel>
                  <button
                    type="button"
                    onClick={() => curlExample && copy(curlExample)}
                    className="w-full text-left font-mono text-[11px] text-emerald-400/70 break-all hover:text-emerald-300 transition leading-relaxed"
                  >
                    {curlExample || '—'}
                  </button>
                </div>
              </div>

              {/* IP Allowlist */}
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5 space-y-4">
                <ToggleRow label="IP allowlist" desc="Allow only specific IPs or CIDRs." checked={ipEnabled} onChange={setIpEnabled} icon={Shield} />

                <div className="space-y-2">
                  <SectionLabel>Entries (space/comma separated)</SectionLabel>
                  <textarea
                    value={ipText}
                    onChange={(e) => setIpText(e.target.value)}
                    placeholder="1.2.3.4, 10.0.0.0/24"
                    className="w-full h-24 px-3.5 py-3 bg-black/20 border border-white/[0.06] rounded-xl text-xs font-mono focus:ring-1 focus:ring-emerald-500/20 focus:border-emerald-500/30 transition-all outline-none text-white/70 resize-none disabled:opacity-40"
                    disabled={!ipEnabled}
                  />
                  {ipEnabled && allowlist.length > 0 && (
                    <div className="flex flex-wrap gap-1.5 pt-1">
                      {allowlist.slice(0, 8).map((v) => (
                        <span key={v} className="inline-flex items-center gap-1 rounded-md border border-white/[0.08] bg-white/[0.03] px-2 py-0.5 text-[10px] font-mono text-white/50">
                          {v}
                        </span>
                      ))}
                      {allowlist.length > 8 && (
                        <span className="text-[10px] text-white/20">+{allowlist.length - 8} more</span>
                      )}
                    </div>
                  )}
                </div>
              </div>

              {/* Basic Auth */}
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5 space-y-4">
                <ToggleRow label="Basic Auth" desc="Browser-native username/password." checked={basicEnabled} onChange={setBasicEnabled} icon={Lock} />
                <div className="grid grid-cols-2 gap-3">
                  <Input value={basicUser} onChange={(e) => setBasicUser(e.target.value)} disabled={!basicEnabled} placeholder="username" className="text-xs font-mono" />
                  <Input value={basicPass} onChange={(e) => setBasicPass(e.target.value)} disabled={!basicEnabled} placeholder="password" type="password" className="text-xs font-mono" />
                </div>
                <p className="text-[11px] text-white/30">Password stored hashed, never shown again.</p>
              </div>
            </div>
          )}

          {tab === 'limits' && (
            <div className="space-y-4">
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5 space-y-4">
                <ToggleRow label="Rate limit" desc="Per-tunnel, per-client IP sliding window." checked={rateEnabled} onChange={setRateEnabled} icon={Gauge} />
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1.5">
                    <SectionLabel>Requests</SectionLabel>
                    <Input value={String(rateReq)} onChange={(e) => setRateReq(Number(e.target.value || 0))} disabled={!rateEnabled} type="number" min={1} className="text-xs font-mono" />
                  </div>
                  <div className="space-y-1.5">
                    <SectionLabel>Window (seconds)</SectionLabel>
                    <Input value={String(rateWin)} onChange={(e) => setRateWin(Number(e.target.value || 0))} disabled={!rateEnabled} type="number" min={1} className="text-xs font-mono" />
                  </div>
                </div>
              </div>

              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5">
                <ToggleRow label="HTTPS redirect" desc="Redirect HTTP to HTTPS automatically." checked={httpsRedirect} onChange={setHttpsRedirect} icon={Lock} />
              </div>
            </div>
          )}

          {tab === 'rewrite' && (
            <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5 space-y-4">
              <div className="flex items-center gap-3 mb-1">
                <div className="p-2 rounded-lg bg-white/[0.03] border border-white/[0.06]">
                  <ArrowRightLeft className="w-4 h-4 text-white/50" />
                </div>
                <div>
                  <div className="text-sm font-medium text-white">Rewrite &amp; headers</div>
                  <div className="text-[11px] text-white/35">Request/response shaping per tunnel.</div>
                </div>
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
                <div className="lg:col-span-1 space-y-2">
                  <SectionLabel><CornerDownRight size={11} className="inline mr-1" />Path prefix</SectionLabel>
                  <Input value={pathPrefix} onChange={(e) => setPathPrefix(e.target.value)} placeholder="/api" className="text-xs font-mono" />
                  <SectionLabel>Replace path from/to</SectionLabel>
                  <Input value={replaceFrom} onChange={(e) => setReplaceFrom(e.target.value)} placeholder="/v1" className="text-xs font-mono" />
                  <Input value={replaceTo} onChange={(e) => setReplaceTo(e.target.value)} placeholder="/v2" className="text-xs font-mono" />
                </div>
                <div className="lg:col-span-2 grid grid-cols-1 md:grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <SectionLabel>Add request headers</SectionLabel>
                    <textarea value={addReqHeadersText} onChange={(e) => setAddReqHeadersText(e.target.value)} className="w-full h-20 px-3.5 py-2.5 bg-black/20 border border-white/[0.06] rounded-xl text-xs font-mono outline-none text-white/70 resize-none" placeholder="Key: Value" />
                    <SectionLabel>Remove request headers</SectionLabel>
                    <Input value={removeReqHeadersText} onChange={(e) => setRemoveReqHeadersText(e.target.value)} className="text-xs font-mono" placeholder="Header-Name" />
                  </div>
                  <div className="space-y-2">
                    <SectionLabel>Add response headers</SectionLabel>
                    <textarea value={addRespHeadersText} onChange={(e) => setAddRespHeadersText(e.target.value)} className="w-full h-20 px-3.5 py-2.5 bg-black/20 border border-white/[0.06] rounded-xl text-xs font-mono outline-none text-white/70 resize-none" placeholder="Key: Value" />
                    <SectionLabel>Remove response headers</SectionLabel>
                    <Input value={removeRespHeadersText} onChange={(e) => setRemoveRespHeadersText(e.target.value)} className="text-xs font-mono" placeholder="Header-Name" />
                  </div>
                </div>
              </div>
            </div>
          )}

          <div className="flex flex-col sm:flex-row gap-2.5 sm:justify-end pt-3 border-t border-white/[0.04]">
            <Button type="button" variant="outline" size="md" onClick={onClose} disabled={loading}>
              Close
            </Button>
            <Button type="button" variant="primary" size="md" onClick={save} disabled={loading}>
              <Shield className="w-3.5 h-3.5" /> Save policy
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};
