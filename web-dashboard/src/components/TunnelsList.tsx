import { useTranslation } from 'react-i18next';
import { Copy, ExternalLink, Server, Plus, Zap, ArrowDown, ArrowUp, ArrowRight, Shield, KeyRound, Lock } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import type { Tunnel, TunnelSessionHistory } from '../api/client';
import React, { useState } from 'react';
import { TunnelPolicyModal } from './TunnelPolicyModal';

interface TunnelsListProps {
  tunnels: Tunnel[];
  historySessions?: TunnelSessionHistory[];
  onOpenConnect: () => void;
}

export const TunnelsList: React.FC<TunnelsListProps> = ({ tunnels, historySessions = [], onOpenConnect }) => {
  const { t } = useTranslation();
  const [policyTunnel, setPolicyTunnel] = useState<Tunnel | null>(null);
  const [isPolicyOpen, setIsPolicyOpen] = useState(false);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const getStatusInfo = (status: string) => {
    switch (status) {
      case 'active':
        return { text: 'Active', pulse: true, color: 'text-primary', glow: 'bg-primary shadow-[0_0_8px_rgba(16,185,129,1)]' };
      case 'idle':
        return { text: 'Idle', pulse: false, color: 'text-yellow-400', glow: 'bg-yellow-400' };
      case 'error':
        return { text: 'Error', pulse: true, color: 'text-rose-500', glow: 'bg-rose-500' };
      default:
        return { text: 'Offline', pulse: false, color: 'text-white/20', glow: 'bg-white/10' };
    }
  };

  return (
    <div className="card space-y-8">
      <TunnelPolicyModal open={isPolicyOpen} onClose={() => setIsPolicyOpen(false)} tunnel={policyTunnel} onUpdated={() => {}} />
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-4 min-w-0">
          <div className="p-4 bg-primary/10 rounded-2xl shadow-[0_0_20px_rgba(16,185,129,0.1)] shrink-0">
            <Server className="w-6 h-6 text-primary" />
          </div>
          <div className="min-w-0">
            <h3 className="text-xl font-black">Managed Tunnels</h3>
            <p className="text-sm text-white/40 font-medium">{tunnels.length} Active Cloud Endpoints</p>
          </div>
        </div>

        <button
          type="button"
          onClick={onOpenConnect}
          className="btn-primary-premium text-sm py-2.5 w-full sm:w-auto justify-center transition-transform duration-200 hover:brightness-110 active:scale-[0.98]"
        >
          <Plus className="w-4 h-4" /> {t('tunnels.cta', 'New Tunnel')}
        </button>
      </div>

      <div className="grid grid-cols-1 gap-4">
        {tunnels.length === 0 ? (
          <div className="py-16 px-8 border-2 border-dashed border-white/5 rounded-[3rem] bg-white/[0.01] relative overflow-hidden group">
            <div className="absolute top-0 right-0 p-12 opacity-[0.02] group-hover:opacity-[0.05] transition-opacity">
               <Zap className="w-64 h-64 text-primary" />
            </div>
            
            <div className="relative z-10 text-center max-w-2xl mx-auto space-y-10">
              <div className="space-y-4">
                <div className="w-20 h-20 bg-primary/10 rounded-[2rem] flex items-center justify-center mx-auto shadow-2xl ring-1 ring-primary/20">
                  <Zap className="w-10 h-10 text-primary animate-pulse" />
                </div>
                <h4 className="text-3xl font-black text-white tracking-tight">{t('tunnels.empty_title')}</h4>
                <p className="text-white/40 font-medium text-lg leading-relaxed">
                  {t('tunnels.empty_subtitle')}
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 text-left">
                {[
                  { step: 1, title: t('tunnels.step1_title'), desc: t('tunnels.step1_desc') },
                  { step: 2, title: t('tunnels.step2_title'), desc: t('tunnels.step2_desc') },
                  { step: 3, title: t('tunnels.step3_title'), desc: t('tunnels.step3_desc') }
                ].map((s) => (
                  <div key={s.step} className="p-6 rounded-3xl bg-white/5 border border-white/5 space-y-2">
                    <span className="text-xs font-black text-primary uppercase tracking-widest">Step 0{s.step}</span>
                    <h5 className="font-bold text-white">{s.title}</h5>
                    <p className="text-xs text-white/30 leading-relaxed">{s.desc}</p>
                  </div>
                ))}
              </div>

              <div className="pt-4">
                <button 
                  onClick={onOpenConnect}
                  className="btn-primary-premium px-10 py-4 text-lg"
                >
                  {t('tunnels.cta')} <ArrowRight className="ml-2 w-5 h-5" />
                </button>
              </div>
            </div>
          </div>
        ) : (
          tunnels.map((tunnel) => {
            const status = getStatusInfo(tunnel.status);
            const keyOn = !!tunnel.policy?.key_auth_enabled;
            const ipOn = !!tunnel.policy?.ip_allowlist_enabled;
            return (
              <div
                key={tunnel.id}
                className="group relative bg-white/[0.02] border border-white/5 rounded-3xl p-5 sm:p-6 hover:bg-white/[0.04] hover:border-white/10 transition-all duration-300"
              >
                <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-5 lg:gap-6">
                  <div className="flex items-start gap-4">
                    <div className="w-12 h-12 rounded-2xl bg-black border border-white/5 flex items-center justify-center font-black text-xl text-primary shrink-0">
                      {tunnel.subdomain.charAt(0).toUpperCase()}
                    </div>
                    <div className="space-y-1">
                      <div className="flex items-center gap-3">
                        <span className="font-black text-lg tracking-tight">{tunnel.subdomain}.gorenel.site</span>
                        <div className={`flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-white/5 border border-white/5 text-[10px] font-black uppercase tracking-widest ${status.color}`}>
                          <div className={`w-1 h-1 rounded-full ${status.glow} ${status.pulse ? 'animate-pulse' : ''}`} />
                          {status.text}
                        </div>
                      </div>
                      <div className="flex flex-wrap items-center gap-2 pt-1">
                        {keyOn && (
                          <span className="inline-flex items-center gap-1.5 rounded-full border border-emerald-500/20 bg-emerald-500/10 px-2.5 py-1 text-[10px] font-black uppercase tracking-widest text-emerald-200/90">
                            <KeyRound className="w-3 h-3" /> KeyAuth
                          </span>
                        )}
                        {ipOn && (
                          <span className="inline-flex items-center gap-1.5 rounded-full border border-violet-400/20 bg-violet-400/10 px-2.5 py-1 text-[10px] font-black uppercase tracking-widest text-violet-200/90">
                            <Lock className="w-3 h-3" /> IP Allowlist
                          </span>
                        )}
                        {!keyOn && !ipOn && (
                          <span className="inline-flex items-center gap-1.5 rounded-full border border-white/10 bg-white/[0.03] px-2.5 py-1 text-[10px] font-black uppercase tracking-widest text-white/30">
                            <Shield className="w-3 h-3" /> Open
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-4 text-xs font-bold text-white/30 tracking-tight">
                        <span className="flex items-center gap-1.5">
                          <div className="w-1 h-1 rounded-full bg-white/10" />
                          Local: <span className="text-white/60">127.0.0.1:{tunnel.localPort}</span>
                        </span>
                        <span className="flex items-center gap-1.5">
                          <Zap className="w-3 h-3" />
                          <span className="text-white/60">{tunnel.requestCount.toLocaleString()} Hits</span>
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="flex flex-col xs:flex-row xs:flex-wrap items-stretch sm:items-center gap-4 sm:gap-6 lg:flex-row">
                    <div className="flex items-center justify-between sm:justify-start gap-4 text-xs font-black tracking-widest uppercase min-w-0 flex-1">
                      <div className="text-center sm:text-left min-w-0">
                        <span className="block text-white/20 mb-1">Traffic In</span>
                        <span className="flex items-center justify-center sm:justify-start gap-1 text-emerald-400 tabular-nums">
                          <ArrowDown className="w-3 h-3 shrink-0" /> {formatBytes(tunnel.bandwidth.in)}
                        </span>
                      </div>
                      <div className="hidden sm:block w-px h-8 bg-white/5 shrink-0" aria-hidden />
                      <div className="text-center sm:text-left min-w-0">
                        <span className="block text-white/20 mb-1">Traffic Out</span>
                        <span className="flex items-center justify-center sm:justify-start gap-1 text-violet-400 tabular-nums">
                          <ArrowUp className="w-3 h-3 shrink-0" /> {formatBytes(tunnel.bandwidth.out)}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 justify-end sm:justify-start shrink-0">
                      <button
                        type="button"
                        onClick={() => {
                          setPolicyTunnel(tunnel);
                          setIsPolicyOpen(true);
                        }}
                        className="px-4 py-3 bg-white/5 border border-white/5 rounded-2xl text-white/50 hover:text-white hover:bg-white/10 transition-all duration-200 active:scale-95 text-xs font-black uppercase tracking-widest flex items-center gap-2"
                        title="Security Policy"
                      >
                        <Shield className="w-4 h-4" /> Security
                      </button>
                      <button
                        type="button"
                        onClick={() => copyToClipboard(tunnel.publicUrl)}
                        className="p-3 bg-white/5 border border-white/5 rounded-2xl text-white/40 hover:text-white hover:bg-white/10 transition-all duration-200 active:scale-95"
                        title="Copy Public URL"
                      >
                        <Copy className="w-4 h-4" />
                      </button>
                      <a
                        href={tunnel.publicUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="p-3 bg-white/5 border border-white/5 rounded-2xl text-white/40 hover:text-primary hover:border-primary/20 hover:bg-primary/10 transition-all duration-200 active:scale-95"
                      >
                        <ExternalLink className="w-4 h-4" />
                      </a>
                    </div>
                  </div>
                </div>

                <div className="mt-6 pt-4 border-t border-white/5 flex items-center justify-between text-[10px] font-black uppercase tracking-widest text-white/10">
                  <span>Session ID: {tunnel.id}</span>
                  <span>Started {formatDistanceToNow(new Date(tunnel.startedAt), { addSuffix: true })}</span>
                </div>
              </div>
            );
          })
        )}
      </div>

      {historySessions.length > 0 && (
        <div className="pt-4 border-t border-white/5 mt-8">
          <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between mb-4">
            <h4 className="text-sm font-black uppercase tracking-widest text-white/40">Recent Session History</h4>
            <span className="text-[10px] text-white/30 font-semibold tabular-nums">{historySessions.length} records</span>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {historySessions.slice(0, 6).map((session) => (
              <div
                key={session.id}
                className="rounded-2xl border border-white/5 bg-white/[0.02] p-4 transition-all duration-200 hover:border-white/10 hover:bg-white/[0.035]"
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="font-semibold text-white">{session.subdomain}.gorenel.site</span>
                  <span className="text-[10px] uppercase tracking-widest text-white/40">{session.tunnel_type || 'http'}</span>
                </div>
                <div className="grid grid-cols-3 gap-2 text-xs">
                  <div>
                    <div className="text-white/30">Requests</div>
                    <div className="text-white font-semibold">{session.request_count.toLocaleString()}</div>
                  </div>
                  <div>
                    <div className="text-white/30">Avg RPS</div>
                    <div className="text-white font-semibold">{session.avg_rps.toFixed(2)}</div>
                  </div>
                  <div>
                    <div className="text-white/30">Traffic</div>
                    <div className="text-white font-semibold">{formatBytes(session.bytes_out || 0)}</div>
                  </div>
                </div>
                <div className="mt-2 text-[10px] text-white/30">
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
