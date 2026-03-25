import React from 'react';
import { ArrowRight, KeyRound, Globe, Shield, TerminalSquare, Gauge, Lock } from 'lucide-react';

type Props = {
  onGoTunnels: () => void;
  onGoReservations: () => void;
};

export const OnboardingCards: React.FC<Props> = ({ onGoTunnels, onGoReservations }) => {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
      <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5 space-y-3 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-200 group">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-emerald-500/[0.08] border border-emerald-500/[0.12]">
            <Globe className="w-4 h-4 text-emerald-400" />
          </div>
          <span className="text-sm font-medium text-white">Reserve a stable URL</span>
        </div>
        <p className="text-xs text-white/35 leading-relaxed">
          Create a reserved subdomain for a device/customer and keep it stable across reconnects.
        </p>
        <button onClick={onGoReservations} className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400 hover:text-emerald-300 transition-colors">
          Open Reservations <ArrowRight size={13} />
        </button>
      </div>

      <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5 space-y-3 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-200 group">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-white/[0.04] border border-white/[0.06]">
            <Shield className="w-4 h-4 text-white/50" />
          </div>
          <span className="text-sm font-medium text-white">Lock down traffic</span>
        </div>
        <div className="grid grid-cols-2 gap-1.5 text-[11px] text-white/35">
          <div className="inline-flex items-center gap-1.5"><KeyRound size={12} className="text-white/30" /> KeyAuth</div>
          <div className="inline-flex items-center gap-1.5"><Lock size={12} className="text-white/30" /> Basic Auth</div>
          <div className="inline-flex items-center gap-1.5"><Gauge size={12} className="text-white/30" /> Rate limit</div>
          <div className="inline-flex items-center gap-1.5"><TerminalSquare size={12} className="text-white/30" /> Rewrite</div>
        </div>
        <button onClick={onGoTunnels} className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400 hover:text-emerald-300 transition-colors">
          Security settings <ArrowRight size={13} />
        </button>
      </div>

      <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5 space-y-3 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-200 group">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-white/[0.04] border border-white/[0.06]">
            <TerminalSquare className="w-4 h-4 text-white/50" />
          </div>
          <span className="text-sm font-medium text-white">Run always-on</span>
        </div>
        <p className="text-xs text-white/35 leading-relaxed">
          Install as a service so tunnels survive closes and auto-reconnect on failure.
        </p>
        <div className="rounded-lg border border-white/[0.06] bg-black/20 p-2.5 font-mono text-[11px] text-white/50">
          gorenel service install
        </div>
      </div>
    </div>
  );
};
