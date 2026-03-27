import React from 'react';
import { ShieldCheck, Zap, Globe } from 'lucide-react';
import { Button } from './ui/Button';

interface AuthLayoutProps {
  children: React.ReactNode;
  title: string;
  subtitle: string;
  topRight?: React.ReactNode;
}

export const AuthLayout: React.FC<AuthLayoutProps> = ({ children, title, subtitle, topRight }) => {
  return (
    <div className="min-h-screen bg-[#080a10] grid lg:grid-cols-[1.1fr_1fr]">
      {/* Left - Brand Panel */}
      <div className="relative hidden lg:block overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-[#0a0e18] to-[#080a10]" />
        <div className="absolute inset-0">
          <div className="absolute top-[10%] left-[10%] w-[400px] h-[400px] bg-emerald-500/[0.06] rounded-full blur-[120px]" />
          <div className="absolute bottom-[10%] right-[10%] w-[300px] h-[300px] bg-blue-500/[0.04] rounded-full blur-[100px]" />
        </div>

        <div className="relative z-10 p-10 lg:p-12 h-full flex flex-col">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-xl bg-white/[0.05] border border-white/[0.08] overflow-hidden flex items-center justify-center">
                <img src="/logo.png" alt="Gorenel" width="256" height="256" className="w-full h-full object-cover" />
              </div>
              <span className="font-semibold text-white tracking-tight">Gorenel</span>
            </div>
            <div className="opacity-80">{topRight}</div>
          </div>

          <div className="mt-auto mb-auto space-y-8 max-w-md">
            <div className="inline-flex items-center gap-2 rounded-full border border-white/[0.08] bg-white/[0.03] px-3 py-1 text-[11px] font-medium tracking-wide text-white/50">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 shadow-[0_0_8px_rgba(16,185,129,0.5)]" />
              Secure tunnels &middot; Reserved URLs &middot; Policy engine
            </div>

            <h1 className="text-3xl lg:text-4xl font-semibold tracking-tight text-white leading-tight">
              {title}
            </h1>
            <p className="text-sm text-white/40 leading-relaxed max-w-sm">{subtitle}</p>

            <div className="grid grid-cols-2 gap-3 pt-2">
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4 space-y-2">
                <Globe className="w-4 h-4 text-emerald-400/70" />
                <div className="text-xs font-medium text-white/70">Reserved URLs</div>
                <div className="text-[11px] text-white/35 leading-relaxed">Stable subdomains per device</div>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4 space-y-2">
                <ShieldCheck className="w-4 h-4 text-blue-400/70" />
                <div className="text-xs font-medium text-white/70">Per-tunnel policies</div>
                <div className="text-[11px] text-white/35 leading-relaxed">Auth, allowlist, rate limit</div>
              </div>
            </div>

            <div className="rounded-xl border border-white/[0.06] bg-black/30 p-4 font-mono text-xs text-white/50">
              <span className="text-emerald-400/60">$</span> gorenel start --subdomain my-device-01 --key-auth &lt;TOKEN&gt;
            </div>
          </div>

          <div className="flex items-center gap-6 text-[11px] font-medium text-white/20">
            <span className="inline-flex items-center gap-1.5"><ShieldCheck className="w-3.5 h-3.5 text-emerald-500/50" /> Encrypted</span>
            <span className="inline-flex items-center gap-1.5"><Zap className="w-3.5 h-3.5 text-blue-500/50" /> Edge-ready</span>
          </div>
        </div>
      </div>

      {/* Right - Form Panel */}
      <div className="relative flex items-center justify-center p-6 lg:p-12">
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute top-[10%] right-[20%] w-[300px] h-[300px] bg-emerald-500/[0.03] rounded-full blur-[100px]" />
        </div>

        <div className="relative z-10 max-w-[420px] w-full">
          <div className="lg:hidden flex items-center justify-between mb-8">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-xl bg-white/[0.05] border border-white/[0.08] overflow-hidden">
                <img src="/logo.png" alt="Gorenel" width="256" height="256" className="w-full h-full object-cover" />
              </div>
              <span className="font-semibold text-white tracking-tight">Gorenel</span>
            </div>
            {topRight ? <Button variant="ghost" size="sm" type="button">{topRight}</Button> : null}
          </div>

          <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] backdrop-blur-xl p-8 shadow-elevated relative overflow-hidden">
            <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-emerald-500/20 to-transparent" />
            {children}
          </div>
        </div>
      </div>
    </div>
  );
};
