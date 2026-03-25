import React from 'react';
import { ShieldCheck } from 'lucide-react';
import { Button } from './ui/Button';

interface AuthLayoutProps {
  children: React.ReactNode;
  title: string;
  subtitle: string;
  topRight?: React.ReactNode;
}

export const AuthLayout: React.FC<AuthLayoutProps> = ({ children, title, subtitle, topRight }) => {
  return (
    <div className="min-h-screen bg-[#020408] grid lg:grid-cols-2">
      <div className="relative hidden lg:block border-r border-white/5 overflow-hidden">
        <div className="absolute inset-0">
          <div className="absolute -top-24 -left-24 w-[520px] h-[520px] bg-emerald-500/10 rounded-full blur-[120px]" />
          <div className="absolute -bottom-24 -right-24 w-[520px] h-[520px] bg-blue-600/10 rounded-full blur-[140px]" />
          <div className="absolute inset-0 bg-gradient-to-b from-transparent via-black/40 to-black/80" />
        </div>
        <div className="relative z-10 p-10 h-full flex flex-col">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-2xl bg-white/5 border border-white/10 overflow-hidden">
                <img src="/logo.png" alt="Gorenel" className="w-full h-full object-cover" />
              </div>
              <div className="text-sm font-black tracking-tight text-white">Gorenel</div>
            </div>
            <div className="opacity-80">{topRight}</div>
          </div>

          <div className="mt-16 space-y-6 max-w-md">
            <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.03] px-3 py-1 text-[11px] font-black tracking-widest text-white/60">
              <ShieldCheck className="w-3.5 h-3.5 text-primary" />
              SECURE TUNNELS, RESERVED URLS, POLICY ENGINE
            </div>
            <h1 className="text-4xl font-black tracking-tight text-white leading-tight">
              {title}
            </h1>
            <p className="text-sm text-white/45 leading-relaxed">{subtitle}</p>

            <div className="grid grid-cols-2 gap-3 pt-6">
              <div className="rounded-3xl border border-white/10 bg-white/[0.02] p-4">
                <div className="text-[10px] font-black uppercase tracking-widest text-white/40">Reserved URLs</div>
                <div className="mt-2 text-xs text-white/60 font-semibold">Stable subdomains per device/customer</div>
              </div>
              <div className="rounded-3xl border border-white/10 bg-white/[0.02] p-4">
                <div className="text-[10px] font-black uppercase tracking-widest text-white/40">Per-tunnel Policies</div>
                <div className="mt-2 text-xs text-white/60 font-semibold">Auth, allowlist, rate limit, rewrite</div>
              </div>
            </div>

            <div className="pt-6">
              <div className="rounded-2xl border border-white/10 bg-black/40 p-4 font-mono text-[12px] text-white/65">
                gorenel start --subdomain my-device-01 --key-auth &lt;TOKEN&gt;
              </div>
              <div className="pt-2 text-[11px] text-white/35">
                Install always-on: <span className="font-mono text-white/60">gorenel service install</span>
              </div>
            </div>
          </div>

          <div className="mt-auto pt-10 flex items-center gap-4 text-[10px] font-black uppercase tracking-widest text-white/25">
            <span className="inline-flex items-center gap-2"><ShieldCheck className="w-3.5 h-3.5 text-emerald-500" /> Encrypted</span>
            <span className="w-1 h-1 rounded-full bg-white/10" />
            <span className="inline-flex items-center gap-2"><ShieldCheck className="w-3.5 h-3.5 text-blue-500" /> Edge-ready</span>
          </div>
        </div>
      </div>

      <div className="relative flex items-center justify-center p-6">
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute top-[-10%] right-[-10%] w-[40%] h-[40%] bg-emerald-500/6 rounded-full blur-[120px]" />
          <div className="absolute bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/6 rounded-full blur-[120px]" />
        </div>
        <div className="relative z-10 max-w-md w-full">
          <div className="lg:hidden flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-2xl bg-white/5 border border-white/10 overflow-hidden">
                <img src="/logo.png" alt="Gorenel" className="w-full h-full object-cover" />
              </div>
              <div className="text-sm font-black tracking-tight text-white">Gorenel</div>
            </div>
            {topRight ? <Button variant="ghost" size="sm" type="button">{topRight}</Button> : null}
          </div>

          <div className="glass rounded-[2.5rem] p-8 md:p-10 shadow-2xl relative overflow-hidden">
            <div className="absolute top-0 left-0 w-full h-[2px] bg-gradient-to-r from-transparent via-emerald-500/30 to-transparent" />
            {children}
          </div>
        </div>
      </div>
    </div>
  );
};
