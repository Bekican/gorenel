import React, { useState } from 'react';
import { Lock, Mail, Loader2, AlertCircle, ArrowRight, ShieldCheck } from 'lucide-react';
import { api } from '../api/client';

interface LoginPageProps {
    onLoginSuccess: (user: any) => void;
}

export const LoginPage: React.FC<LoginPageProps> = ({ onLoginSuccess }) => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError(null);

        try {
            const data = await api.login({ email, password });

            if (data.redirect_url) {
                window.location.href = data.redirect_url;
            } else if (data.user) {
                onLoginSuccess(data.user);
            }
        } catch (err: any) {
            setError(err.response?.data || 'Access Denied. Please verify your credentials.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-[#0c0c0c] flex items-center justify-center p-6 relative overflow-hidden">
            {/* Background Glows */}
            <div className="absolute top-1/4 left-1/4 w-[500px] h-[500px] bg-primary/20 blur-[120px] rounded-full animate-pulse" />
            <div className="absolute bottom-1/4 right-1/4 w-[400px] h-[400px] bg-violet-500/10 blur-[100px] rounded-full" />

            <div className="max-w-md w-full relative z-10 space-y-12">
                {/* Logo Section */}
                <div className="text-center space-y-4">
                    <div className="inline-flex overflow-hidden w-24 h-24 bg-white/5 rounded-[2rem] border border-white/10 shadow-[0_0_50px_rgba(16,185,129,0.15)] mb-4">
                        <img src="/logo.png" alt="Gorenel" className="w-full h-full object-cover" />
                    </div>
                    <div className="space-y-1">
                        <h1 className="text-5xl font-black tracking-tighter text-white">
                            GORENEL<span className="text-primary text-6xl leading-[0]">.</span>
                        </h1>
                        <p className="text-[10px] font-black uppercase tracking-[0.4em] text-white/20">Secure Tunneling Interface</p>
                    </div>
                </div>

                {/* Glassmorphic Form Container */}
                <div className="glass rounded-[3rem] p-10 space-y-8 relative group overflow-hidden border-white/5">
                    <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-primary/50 to-transparent opacity-50" />

                    <form onSubmit={handleSubmit} className="space-y-6">
                        {error && (
                            <div className="bg-rose-500/10 border border-rose-500/20 text-rose-400 px-5 py-4 rounded-2xl flex items-center gap-3 animate-bounce">
                                <AlertCircle className="w-5 h-5 flex-shrink-0" />
                                <p className="text-xs font-black uppercase tracking-widest">{error}</p>
                            </div>
                        )}

                        <div className="space-y-3">
                            <label className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em] ml-2">Endpoint Identity</label>
                            <div className="relative group/input">
                                <Mail className="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-white/10 group-focus-within/input:text-primary transition-colors" />
                                <input
                                    type="email"
                                    required
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    className="w-full bg-black border border-white/5 rounded-2xl py-5 pl-14 pr-6 text-white font-medium placeholder:text-white/10 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/50 transition-all selection:bg-primary/30"
                                    placeholder="operator@gorenel.net"
                                />
                            </div>
                        </div>

                        <div className="space-y-3">
                            <label className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em] ml-2">Access Key</label>
                            <div className="relative group/input">
                                <Lock className="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-white/10 group-focus-within/input:text-primary transition-colors" />
                                <input
                                    type="password"
                                    required
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    className="w-full bg-black border border-white/5 rounded-2xl py-5 pl-14 pr-6 text-white font-medium placeholder:text-white/10 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/50 transition-all selection:bg-primary/30"
                                    placeholder="••••••••"
                                />
                            </div>
                        </div>

                        <button
                            type="submit"
                            disabled={loading}
                            className="btn-primary-premium w-full py-5 rounded-2xl flex items-center justify-center gap-3 active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed group"
                        >
                            {loading ? (
                                <>
                                    <Loader2 className="w-5 h-5 animate-spin" />
                                    <span className="font-black uppercase tracking-widest text-xs">Synchronizing...</span>
                                </>
                            ) : (
                                <>
                                    <span className="font-black uppercase tracking-widest text-xs">Establish Session</span>
                                    <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                                </>
                            )}
                        </button>
                    </form>

                    {/* Footer Info */}
                    {import.meta.env.VITE_SHOW_DEMO === 'true' && (
                        <div className="pt-8 border-t border-white/5 space-y-4">
                            <div className="flex items-center justify-center gap-2">
                                <div className="w-1.5 h-1.5 rounded-full bg-primary animate-pulse" />
                                <p className="text-[10px] font-black text-white/10 uppercase tracking-widest text-center">
                                    Operator credentials: <span className="text-white/30 font-mono lower-case">demo@gorenel.net</span>
                                </p>
                            </div>
                        </div>
                    )}
                </div>

                {/* Sub-footer */}
                <div className="flex flex-col items-center gap-4">
                    <div className="flex items-center gap-4 text-white/10">
                        <div className="flex items-center gap-1.5">
                            <ShieldCheck className="w-3 h-3" />
                            <span className="text-[10px] font-black uppercase tracking-widest">Encrypted</span>
                        </div>
                        <div className="w-1 h-1 bg-white/5 rounded-full" />
                        <div className="flex items-center gap-1.5">
                            <ShieldCheck className="w-3 h-3 text-primary" />
                            <span className="text-[10px] font-black uppercase tracking-widest">Ultra Low Latency</span>
                        </div>
                    </div>
                    <p className="text-center font-black text-[10px] uppercase tracking-[0.5em] text-white/5">
                        Gorenel Gateway v1.0.0
                    </p>
                </div>
            </div>
        </div>
    );
};
