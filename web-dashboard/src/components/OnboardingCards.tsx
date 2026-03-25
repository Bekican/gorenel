import React from 'react';
import { ArrowRight, KeyRound, Globe, Shield, TerminalSquare, Gauge, Lock } from 'lucide-react';

type Props = {
  onGoTunnels: () => void;
  onGoReservations: () => void;
};

export const OnboardingCards: React.FC<Props> = ({ onGoTunnels, onGoReservations }) => {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div className="rounded-3xl border border-white/10 bg-white/[0.02] p-6 space-y-3">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-2xl bg-white/[0.04] border border-white/10">
            <Globe className="w-4 h-4 text-primary" />
          </div>
          <div className="text-sm font-black text-white">Reserve a stable URL</div>
        </div>
        <p className="text-xs text-white/45 leading-relaxed">
          Create a reserved subdomain for a device/customer and keep it stable across reconnects.
        </p>
        <button onClick={onGoReservations} className="inline-flex items-center gap-2 text-xs font-black text-primary hover:text-emerald-300">
          Open Reservations <ArrowRight size={14} />
        </button>
      </div>

      <div className="rounded-3xl border border-white/10 bg-white/[0.02] p-6 space-y-3">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-2xl bg-white/[0.04] border border-white/10">
            <Shield className="w-4 h-4 text-white/70" />
          </div>
          <div className="text-sm font-black text-white">Lock down public traffic</div>
        </div>
        <div className="grid grid-cols-2 gap-2 text-[11px] text-white/45">
          <div className="inline-flex items-center gap-2"><KeyRound size={14} className="text-white/40" /> KeyAuth</div>
          <div className="inline-flex items-center gap-2"><Lock size={14} className="text-white/40" /> Basic Auth</div>
          <div className="inline-flex items-center gap-2"><Gauge size={14} className="text-white/40" /> Rate limit</div>
          <div className="inline-flex items-center gap-2"><TerminalSquare size={14} className="text-white/40" /> Rewrite/Headers</div>
        </div>
        <button onClick={onGoTunnels} className="inline-flex items-center gap-2 text-xs font-black text-primary hover:text-emerald-300">
          Open Tunnels & Security <ArrowRight size={14} />
        </button>
      </div>

      <div className="rounded-3xl border border-white/10 bg-white/[0.02] p-6 space-y-3">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-2xl bg-white/[0.04] border border-white/10">
            <TerminalSquare className="w-4 h-4 text-white/70" />
          </div>
          <div className="text-sm font-black text-white">Run always-on</div>
        </div>
        <p className="text-xs text-white/45 leading-relaxed">
          Install Gorenel as a service so tunnels survive terminal closes and auto-reconnect on failure.
        </p>
        <div className="rounded-2xl border border-white/10 bg-black/40 p-3 font-mono text-[11px] text-white/60">
          gorenel service install
        </div>
      </div>
    </div>
  );
};

