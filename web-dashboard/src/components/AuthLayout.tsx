import React from 'react';
import { ShieldCheck } from 'lucide-react';

interface AuthLayoutProps {
  children: React.ReactNode;
  title: string;
  subtitle: string;
}

export const AuthLayout: React.FC<AuthLayoutProps> = ({ children, title, subtitle }) => {
  return (
    <div className="min-h-screen bg-[#020408] flex items-center justify-center p-6 relative overflow-hidden">
      {/* Dynamic Animated Background */}
      <div className="fixed inset-0 z-0">
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-emerald-500/10 rounded-full blur-[120px] animate-pulse-slow" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-600/10 rounded-full blur-[120px] animate-pulse-slow delay-1000" />
        <div className="absolute top-[30%] right-[20%] w-[20%] h-[20%] bg-violet-600/5 rounded-full blur-[80px] animate-pulse-slow delay-2000" />
      </div>

      <div className="max-w-md w-full relative z-10">
        {/* Logo Section */}
        <div className="text-center mb-10 group cursor-default">
          <div className="inline-flex w-20 h-20 bg-white/5 rounded-3xl border border-white/10 shadow-[0_0_40px_rgba(16,185,129,0.1)] mb-6 transition-all duration-500 group-hover:scale-105 group-hover:shadow-[0_0_60px_rgba(16,185,129,0.2)] overflow-hidden">
            <img src="/logo.png" alt="Gorenel Logo" className="w-full h-full object-cover" />
          </div>
          <div className="space-y-2">
            <h1 className="text-4xl font-black tracking-tighter text-white">
              {title}<span className="text-emerald-500 leading-none">.</span>
            </h1>
            <p className="text-[11px] font-bold uppercase tracking-[0.4em] text-white/30 truncate">
              {subtitle}
            </p>
          </div>
        </div>

        {/* Glass Container */}
        <div className="glass rounded-[2.5rem] p-8 md:p-10 shadow-2xl relative overflow-hidden group">
          <div className="absolute top-0 left-0 w-full h-[2px] bg-gradient-to-r from-transparent via-emerald-500/30 to-transparent" />
          {children}
        </div>

        {/* Footer Metrics/Info */}
        <div className="mt-10 flex flex-col items-center gap-6">
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-2 opacity-30 hover:opacity-100 transition-opacity">
              <ShieldCheck className="w-3.5 h-3.5 text-emerald-500" />
              <span className="text-[10px] font-black uppercase tracking-widest">End-to-End Encrypted</span>
            </div>
            <div className="w-1 h-1 bg-white/10 rounded-full" />
            <div className="flex items-center gap-2 opacity-30 hover:opacity-100 transition-opacity">
              <ShieldCheck className="w-3.5 h-3.5 text-blue-500" />
              <span className="text-[10px] font-black uppercase tracking-widest">Global Edge Network</span>
            </div>
          </div>
          
          <div className="flex flex-col items-center gap-2">
            <p className="text-[10px] font-bold uppercase tracking-[0.6em] text-white/5">
              Powered by Gorenel Architecture
            </p>
            <div className="h-[1px] w-8 bg-white/5" />
          </div>
        </div>
      </div>
    </div>
  );
};
