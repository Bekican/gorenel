import { useTranslation } from 'react-i18next';
import { Copy, ExternalLink, Server, Plus, Zap, ArrowDown, ArrowUp, ArrowRight, Shield, KeyRound, Lock } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import type { Tunnel, TunnelSessionHistory } from '../api/client';
import React, { useState } from 'react';
import { TunnelPolicyModal } from './TunnelPolicyModal';
import { OnboardingCards } from './OnboardingCards';
import { Button } from './ui/Button';

interface TunnelsListProps {
  tunnels: Tunnel[];
  historySessions?: TunnelSessionHistory[];
  onOpenConnect: () => void;
  onGoReservations?: () => void;
}

export const TunnelsList: React.FC<TunnelsListProps> = ({ tunnels, historySessions = [], onOpenConnect, onGoReservations }) => {
  const { t } = useTranslation();
  const [policyTunnel, setPolicyTunnel] = useState<Tunnel | null>(null);
  const [isPolicyOpen, setIsPolicyOpen] = useState(false);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const getStatusInfo = (status: string) => {
    switch (status) {
      case 'active':
        return { text: 'Active', pulse: true, color: 'text-emerald-400', glow: 'bg-emerald-400 shadow-[0_0_6px_rgba(16,185,129,0.6)]' };
      case 'idle':
        return { text: 'Idle', pulse: false, color: 'text-yellow-400', glow: 'bg-yellow-400' };
      case 'error':
        return { text: 'Error', pulse: true, color: 'text-rose-400', glow: 'bg-rose-400' };
      default:
        return { text: 'Offline', pulse: false, color: 'text-white/25', glow: 'bg-white/15' };
    }
  };

  return (
    <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 md:p-8 space-y-6">
      <TunnelPolicyModal open={isPolicyOpen} onClose={() => setIsPolicyOpen(false)} tunnel={policyTunnel} onUpdated={() => {}} />

      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3 min-w-0">
          <div className="p-3 bg-emerald-500/10 rounded-xl border border-emerald-500/15 shrink-0">
            <Server className="w-5 h-5 text-emerald-400" />
          </div>
          <div className="min-w-0">
            <h3 className="text-lg font-semibold">Managed Tunnels</h3>
            <p className="text-sm text-white/35">{tunnels.length} active endpoints</p>
          </div>
        </div>

        <Button type="button" variant="primary" size="md" onClick={onOpenConnect} className="w-full sm:w-auto">
          <Plus className="w-4 h-4" /> {t('tunnels.cta', 'New Tunnel')}
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-3">
        {tunnels.length === 0 ? (
          <div className="space-y-5">
            <OnboardingCards
              onGoTunnels={onOpenConnect}
              onGoReservations={() => (onGoReservations ? onGoReservations() : onOpenConnect())}
            />
            <div className="py-14 px-6 border border-dashed border-white/[0.06] rounded-2xl bg-white/[0.01] relative overflow-hidden">
              <div className="absolute top-0 right-0 p-8 opacity-[0.02]">
                <Zap className="w-48 h-48 text-emerald-500" />
              </div>
              
              <div className="relative z-10 text-center max-w-lg mx-auto space-y-6">
                <div className="space-y-3">
                  <div className="w-14 h-14 bg-emerald-500/10 rounded-xl flex items-center justify-center mx-auto border border-emerald-500/15">
                    <Zap className="w-7 h-7 text-emerald-400 animate-pulse" />
                  </div>
                  <h4 className="text-xl font-semibold text-white">{t('tunnels.empty_title')}</h4>
                  <p className="text-white/35 text-sm leading-relaxed">
                    {t('tunnels.empty_subtitle')}
                  </p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-left">
                  {[
                    { step: 1, title: t('tunnels.step1_title'), desc: t('tunnels.step1_desc') },
                    { step: 2, title: t('tunnels.step2_title'), desc: t('tunnels.step2_desc') },
                    { step: 3, title: t('tunnels.step3_title'), desc: t('tunnels.step3_desc') }
                  ].map((s) => (
                    <div key={s.step} className="p-4 rounded-xl bg-white/[0.03] border border-white/[0.06] space-y-1.5">
                      <span className="text-[10px] font-semibold text-emerald-400 uppercase tracking-wider">Step {s.step}</span>
                      <h5 className="font-medium text-sm text-white">{s.title}</h5>
                      <p className="text-[11px] text-white/30 leading-relaxed">{s.desc}</p>
                    </div>
                  ))}
                </div>

                <Button type="button" variant="primary" size="lg" onClick={onOpenConnect}>
                  {t('tunnels.cta')} <ArrowRight className="ml-1 w-4 h-4" />
                </Button>
              </div>
            </div>
          </div>
        ) : (
          tunnels.map((tunnel) => {
            const status = getStatusInfo(tunnel.status);
            const keyOn = !!tunnel.policy?.key_auth_enabled;
            const ipOn = !!tunnel.policy?.ip_allowlist_enabled;
            const basicOn = !!tunnel.policy?.basic_auth_enabled;
            const rlOn = !!tunnel.policy?.rate_limit_enabled;
            return (
              <div
                key={tunnel.id}
                className="group relative bg-white/[0.02] border border-white/[0.06] rounded-xl p-5 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-200"
              >
                <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                  <div className="flex items-start gap-3">
                    <div className="w-10 h-10 rounded-xl bg-white/[0.04] border border-white/[0.06] flex items-center justify-center font-semibold text-lg text-emerald-400 shrink-0">
                      {tunnel.subdomain.charAt(0).toUpperCase()}
                    </div>
                    <div className="space-y-1.5">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="font-semibold text-base truncate max-w-[260px] sm:max-w-none">{tunnel.subdomain}.gorenel.site</span>
                        <div className={`flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-white/[0.03] border border-white/[0.06] text-[10px] font-medium ${status.color}`}>
                          <div className={`w-1 h-1 rounded-full ${status.glow} ${status.pulse ? 'animate-pulse' : ''}`} />
                          {status.text}
                        </div>
                        {(keyOn || ipOn || basicOn || rlOn) && (
                          <div className="hidden sm:flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-white/[0.03] border border-white/[0.06] text-[10px] font-medium text-white/40">
                            <Shield className="w-2.5 h-2.5" /> Secured
                          </div>
                        )}
                      </div>
                      <div className="flex flex-wrap items-center gap-1.5">
                        {keyOn && (
                          <span className="inline-flex items-center gap-1 rounded-md border border-emerald-500/15 bg-emerald-500/[0.06] px-2 py-0.5 text-[10px] font-medium text-emerald-300/80">
                            <KeyRound className="w-2.5 h-2.5" /> KeyAuth
                          </span>
                        )}
                        {ipOn && (
                          <span className="inline-flex items-center gap-1 rounded-md border border-violet-400/15 bg-violet-400/[0.06] px-2 py-0.5 text-[10px] font-medium text-violet-300/80">
                            <Lock className="w-2.5 h-2.5" /> IP Allowlist
                          </span>
                        )}
                        {!keyOn && !ipOn && (
                          <span className="inline-flex items-center gap-1 rounded-md border border-white/[0.06] bg-white/[0.02] px-2 py-0.5 text-[10px] font-medium text-white/25">
                            <Shield className="w-2.5 h-2.5" /> Open
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-4 text-xs text-white/30">
                        <span>Local: <span className="text-white/55">127.0.0.1:{tunnel.localPort}</span></span>
                        <span className="flex items-center gap-1">
                          <Zap className="w-3 h-3" />
                          <span className="text-white/55">{tunnel.requestCount.toLocaleString()} hits</span>
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="flex flex-col xs:flex-row items-stretch sm:items-center gap-3 lg:flex-row">
                    <div className="flex items-center gap-4 text-xs min-w-0">
                      <div className="text-center sm:text-left">
                        <span className="block text-white/20 mb-0.5 text-[10px]">In</span>
                        <span className="flex items-center gap-1 text-emerald-400 font-medium tabular-nums">
                          <ArrowDown className="w-3 h-3 shrink-0" /> {formatBytes(tunnel.bandwidth.in)}
                        </span>
                      </div>
                      <div className="w-px h-6 bg-white/[0.04]" />
                      <div className="text-center sm:text-left">
                        <span className="block text-white/20 mb-0.5 text-[10px]">Out</span>
                        <span className="flex items-center gap-1 text-violet-400 font-medium tabular-nums">
                          <ArrowUp className="w-3 h-3 shrink-0" /> {formatBytes(tunnel.bandwidth.out)}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-1.5 shrink-0">
                      <Button
                        type="button"
                        onClick={() => { setPolicyTunnel(tunnel); setIsPolicyOpen(true); }}
                        variant="secondary"
                        size="sm"
                      >
                        <Shield className="w-3.5 h-3.5" /> Policy
                      </Button>
                      <button
                        type="button"
                        onClick={() => copyToClipboard(tunnel.publicUrl)}
                        className="p-2 bg-white/[0.04] border border-white/[0.06] rounded-lg text-white/35 hover:text-white hover:bg-white/[0.07] transition-all"
                        title="Copy URL"
                      >
                        <Copy className="w-3.5 h-3.5" />
                      </button>
                      <a
                        href={tunnel.publicUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="p-2 bg-white/[0.04] border border-white/[0.06] rounded-lg text-white/35 hover:text-emerald-400 hover:border-emerald-500/15 hover:bg-emerald-500/[0.06] transition-all"
                      >
                        <ExternalLink className="w-3.5 h-3.5" />
                      </a>
                    </div>
                  </div>
                </div>

                <div className="mt-4 pt-3 border-t border-white/[0.04] flex items-center justify-between text-[10px] text-white/15">
                  <span>ID: {tunnel.id}</span>
                  <span>Started {formatDistanceToNow(new Date(tunnel.startedAt), { addSuffix: true })}</span>
                </div>
              </div>
            );
          })
        )}
      </div>

      {historySessions.length > 0 && (
        <div className="pt-4 border-t border-white/[0.04] mt-6">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-sm font-medium text-white/40">Recent sessions</h4>
            <span className="text-[11px] text-white/20 tabular-nums">{historySessions.length} records</span>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-2.5">
            {historySessions.slice(0, 6).map((session) => (
              <div
                key={session.id}
                className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4 hover:border-white/[0.1] hover:bg-white/[0.03] transition-all duration-200"
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium text-sm text-white">{session.subdomain}.gorenel.site</span>
                  <span className="text-[10px] text-white/30">{session.tunnel_type || 'http'}</span>
                </div>
                <div className="grid grid-cols-3 gap-2 text-xs">
                  <div>
                    <div className="text-white/25 mb-0.5">Requests</div>
                    <div className="text-white/70 font-medium tabular-nums">{session.request_count.toLocaleString()}</div>
                  </div>
                  <div>
                    <div className="text-white/25 mb-0.5">Avg RPS</div>
                    <div className="text-white/70 font-medium tabular-nums">{session.avg_rps.toFixed(2)}</div>
                  </div>
                  <div>
                    <div className="text-white/25 mb-0.5">Traffic</div>
                    <div className="text-white/70 font-medium tabular-nums">{formatBytes(session.bytes_out || 0)}</div>
                  </div>
                </div>
                <div className="mt-2 text-[10px] text-white/20">
                  Started {formatDistanceToNow(new Date(session.started_at), { addSuffix: true })}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}
