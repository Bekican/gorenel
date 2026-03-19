import React from 'react';
import { Copy, ExternalLink, Server, Plus, Zap, ArrowDown, ArrowUp, ArrowRight } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

interface Tunnel {
  id: string;
  subdomain: string;
  localPort: number;
  publicUrl: string;
  status: 'active' | 'idle' | 'error';
  requestCount: number;
  bandwidth: {
    in: number;
    out: number;
  };
  startedAt: string;
  lastActivity: string;
}

interface TunnelsListProps {
  tunnels: Tunnel[];
  onOpenConnect: () => void;
}

export const TunnelsList: React.FC<TunnelsListProps> = ({ tunnels, onOpenConnect }) => {
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
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="p-4 bg-primary/10 rounded-2xl shadow-[0_0_20px_rgba(16,185,129,0.1)]">
            <Server className="w-6 h-6 text-primary" />
          </div>
          <div>
            <h3 className="text-xl font-black">Managed Tunnels</h3>
            <p className="text-sm text-white/40 font-medium">{tunnels.length} Active Cloud Endpoints</p>
          </div>
        </div>

        <button
          onClick={onOpenConnect}
          className="btn-primary-premium text-sm py-2.5"
        >
          <Plus className="w-4 h-4" /> New Tunnel
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
                <h4 className="text-3xl font-black text-white tracking-tight">Ready to launch?</h4>
                <p className="text-white/40 font-medium text-lg leading-relaxed">
                  Your local services are just one command away from the global edge. Follow these 3 steps to go live.
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 text-left">
                {[
                  { step: 1, title: 'Get the CLI', desc: 'Download the lightweight binary for your OS.' },
                  { step: 2, title: 'Copy Login', desc: 'Grab your unique connection command.' },
                  { step: 3, title: 'Go Online', desc: 'Paste into terminal and watch the magic happen.' }
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
                  Start First Tunnel <ArrowRight className="ml-2 w-5 h-5" />
                </button>
              </div>
            </div>
          </div>
        ) : (
          tunnels.map((tunnel) => {
            const status = getStatusInfo(tunnel.status);
            return (
              <div
                key={tunnel.id}
                className="group relative bg-white/[0.02] border border-white/5 rounded-3xl p-6 hover:bg-white/[0.04] hover:border-white/10 transition-all duration-300"
              >
                <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
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

                  <div className="flex items-center gap-6">
                    <div className="flex items-center gap-4 text-xs font-black tracking-widest uppercase">
                      <div className="text-center">
                        <span className="block text-white/20 mb-1">Traffic In</span>
                        <span className="flex items-center gap-1 text-emerald-400">
                          <ArrowDown className="w-3 h-3" /> {formatBytes(tunnel.bandwidth.in)}
                        </span>
                      </div>
                      <div className="w-px h-8 bg-white/5" />
                      <div className="text-center">
                        <span className="block text-white/20 mb-1">Traffic Out</span>
                        <span className="flex items-center gap-1 text-violet-400">
                          <ArrowUp className="w-3 h-3" /> {formatBytes(tunnel.bandwidth.out)}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => copyToClipboard(tunnel.publicUrl)}
                        className="p-3 bg-white/5 border border-white/5 rounded-2xl text-white/40 hover:text-white hover:bg-white/10 transition-all"
                        title="Copy Public URL"
                      >
                        <Copy className="w-4 h-4" />
                      </button>
                      <a
                        href={tunnel.publicUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="p-3 bg-white/5 border border-white/5 rounded-2xl text-white/40 hover:text-primary hover:border-primary/20 hover:bg-primary/10 transition-all"
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
