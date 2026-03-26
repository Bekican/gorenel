import React from 'react';
import { ArrowRight, CheckCircle2, Command, Globe, Languages, Lock, Shield, TerminalSquare, Activity } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Button } from './ui/Button';

interface LandingPageProps {
    onLogin: () => void;
    isLoggedIn?: boolean;
    onGoToDashboard?: () => void;
}

export const LandingPage: React.FC<LandingPageProps> = ({ onLogin, isLoggedIn, onGoToDashboard }) => {
    const { t, i18n } = useTranslation();

    const toggleLanguage = () => {
        const newLang = i18n.language === 'en' ? 'tr' : 'en';
        i18n.changeLanguage(newLang);
    };

    return (
        <div className="min-h-screen bg-[#080a10] text-white font-sans">
            {/* Ambient background */}
            <div className="fixed inset-0 pointer-events-none overflow-hidden">
                <div className="absolute -top-32 left-1/4 w-[600px] h-[600px] bg-emerald-500/[0.06] rounded-full blur-[150px]" />
                <div className="absolute bottom-0 right-1/4 w-[500px] h-[500px] bg-blue-500/[0.04] rounded-full blur-[140px]" />
                <div className="absolute inset-0 bg-gradient-to-b from-transparent via-transparent to-[#080a10]/80" />
                <div className="absolute inset-0 opacity-[0.15] [background-image:radial-gradient(circle_at_1px_1px,rgba(255,255,255,0.04)_1px,transparent_0)] [background-size:24px_24px]" />
            </div>

            {/* Navigation */}
            <nav className="sticky top-0 z-50 border-b border-white/[0.04] bg-[#080a10]/80 backdrop-blur-xl">
                <div className="max-w-6xl mx-auto px-6 md:px-10 py-3.5 flex items-center justify-between">
                    <div className="flex items-center gap-2.5">
                        <div className="w-8 h-8 rounded-lg bg-white/[0.05] border border-white/[0.08] overflow-hidden">
                            <img src="/logo.png" alt="Gorenel" className="w-full h-full object-cover" />
                        </div>
                        <span className="font-semibold tracking-tight text-white">Gorenel</span>
                    </div>
                    <div className="flex items-center gap-2.5">
                        <button
                            onClick={toggleLanguage}
                            className="flex items-center gap-1.5 px-2.5 py-1.5 bg-white/[0.04] border border-white/[0.08] rounded-lg text-xs font-medium hover:bg-white/[0.07] transition-all"
                            type="button"
                        >
                            <Languages size={13} className="text-emerald-400/70" />
                            {i18n.language.toUpperCase()}
                        </button>
                        <Button variant="ghost" size="sm" type="button" onClick={onLogin}>
                            {t('common.login')}
                        </Button>
                        <Button variant="primary" size="sm" type="button" onClick={isLoggedIn ? onGoToDashboard : onLogin}>
                            {isLoggedIn ? 'Dashboard' : 'Get started'}
                            <ArrowRight size={14} />
                        </Button>
                    </div>
                </div>
            </nav>

            <main className="relative z-10">
                {/* Hero */}
                <section className="max-w-6xl mx-auto px-6 md:px-10 pt-20 md:pt-28 pb-20">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
                        <div className="space-y-8">
                            <div className="inline-flex items-center gap-2 rounded-full border border-white/[0.08] bg-white/[0.02] px-3 py-1.5 text-xs font-medium text-white/50">
                                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 shadow-[0_0_10px_rgba(16,185,129,0.5)]" />
                                Secure tunnels &middot; Reserved URLs &middot; Traffic policy
                            </div>

                            <h1 className="text-4xl md:text-5xl lg:text-[3.5rem] font-semibold tracking-tight leading-[1.1]">
                                {i18n.language === 'tr' ? 'Güvenli tüneller.' : 'Secure tunnels.'}
                                <span className="block text-white/50 mt-1">
                                    {i18n.language === 'tr' ? 'Rezervasyonlu URL ve traffic policy ile.' : 'With reserved URLs and traffic policy.'}
                                </span>
                            </h1>

                            <p className="text-base md:text-lg text-white/40 leading-relaxed max-w-lg">
                                {t('landing.description')}
                            </p>

                            <div className="flex flex-col sm:flex-row gap-3">
                                <Button
                                    variant="primary"
                                    size="lg"
                                    type="button"
                                    onClick={isLoggedIn ? onGoToDashboard : onLogin}
                                >
                                    {isLoggedIn ? 'Open dashboard' : (i18n.language === 'tr' ? 'Başlayın' : 'Get started')}
                                    <ArrowRight size={16} />
                                </Button>
                                <Button
                                    variant="outline"
                                    size="lg"
                                    type="button"
                                    onClick={onLogin}
                                >
                                    {i18n.language === 'tr' ? 'Demo / Giriş' : 'View demo'}
                                </Button>
                            </div>

                            <div className="grid grid-cols-2 gap-3 pt-2">
                                {[
                                    { icon: <Globe className="w-4 h-4 text-emerald-400/70" />, title: 'Reserved URLs', desc: 'Stable subdomains per device.' },
                                    { icon: <Shield className="w-4 h-4 text-blue-400/70" />, title: 'Traffic policy', desc: 'Auth, allowlist, rate limit.' },
                                    { icon: <Lock className="w-4 h-4 text-violet-400/70" />, title: 'Edge security', desc: 'BasicAuth + KeyAuth at edge.' },
                                    { icon: <TerminalSquare className="w-4 h-4 text-amber-400/70" />, title: 'Always-on', desc: 'Service mode, auto-reconnect.' },
                                ].map((f) => (
                                    <div key={f.title} className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-200 group">
                                        <div className="flex items-center gap-2 text-sm font-medium text-white/80 group-hover:text-white transition-colors">
                                            {f.icon}
                                            {f.title}
                                        </div>
                                        <div className="mt-1.5 text-xs text-white/35 leading-relaxed">{f.desc}</div>
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* Hero Visual - Traffic Flow */}
                        <div className="relative">
                            <div className="absolute -inset-6 bg-gradient-to-r from-emerald-500/10 to-blue-500/10 rounded-3xl blur-2xl opacity-40" />
                            <div className="relative rounded-2xl border border-white/[0.08] bg-[#0c0e14]/80 backdrop-blur-xl shadow-elevated overflow-hidden">
                                <div className="px-5 py-4 border-b border-white/[0.06] flex items-center justify-between">
                                    <div className="flex items-center gap-2">
                                        <div className="flex gap-1.5">
                                            <div className="w-2.5 h-2.5 rounded-full bg-white/[0.06]" />
                                            <div className="w-2.5 h-2.5 rounded-full bg-white/[0.06]" />
                                            <div className="w-2.5 h-2.5 rounded-full bg-white/[0.06]" />
                                        </div>
                                        <span className="text-xs font-medium text-white/30 ml-2">Traffic policy</span>
                                    </div>
                                    <div className="flex items-center gap-1.5">
                                        <Activity className="w-3 h-3 text-emerald-400/50" />
                                        <span className="text-[11px] text-white/25">Live</span>
                                    </div>
                                </div>
                                <div className="p-5 space-y-4">
                                    <div className="rounded-xl border border-white/[0.06] bg-black/20 p-4">
                                        <div className="grid grid-cols-3 gap-3 text-[11px] font-medium text-white/30">
                                            <div className="flex items-center gap-2"><span className="w-1.5 h-1.5 rounded-full bg-white/10" /> Internet</div>
                                            <div className="flex items-center gap-2 justify-center"><span className="w-1.5 h-1.5 rounded-full bg-emerald-400/60 shadow-[0_0_6px_rgba(16,185,129,0.4)]" /> Gorenel edge</div>
                                            <div className="flex items-center gap-2 justify-end"><span className="w-1.5 h-1.5 rounded-full bg-white/10" /> Your service</div>
                                        </div>

                                        <div className="mt-4 grid grid-cols-3 gap-3 items-center">
                                            <div className="rounded-lg border border-white/[0.06] bg-white/[0.02] p-3">
                                                <div className="text-[10px] font-medium text-white/30 mb-1.5">Request</div>
                                                <div className="font-mono text-xs text-white/60">GET /api/users</div>
                                            </div>
                                            <div className="rounded-lg border border-emerald-500/15 bg-emerald-500/[0.06] p-3">
                                                <div className="text-[10px] font-medium text-emerald-300/60 mb-1.5">Policies</div>
                                                <ul className="space-y-1 text-[11px] text-emerald-100/60">
                                                    <li className="flex items-center justify-between"><span>KeyAuth</span><span className="text-emerald-400/70">✓</span></li>
                                                    <li className="flex items-center justify-between"><span>IP allowlist</span><span className="text-emerald-400/70">✓</span></li>
                                                    <li className="flex items-center justify-between"><span>Rate limit</span><span className="text-emerald-400/70">✓</span></li>
                                                </ul>
                                            </div>
                                            <div className="rounded-lg border border-white/[0.06] bg-white/[0.02] p-3">
                                                <div className="text-[10px] font-medium text-white/30 mb-1.5">Forward</div>
                                                <div className="font-mono text-xs text-white/60">localhost:3000</div>
                                            </div>
                                        </div>
                                    </div>

                                    <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4">
                                        <div className="text-[10px] font-medium text-white/30 mb-2">Quick start</div>
                                        <div className="font-mono text-xs text-white/55 space-y-1">
                                            <div><span className="text-emerald-400/50">$</span> gorenel config set api_key gk_********</div>
                                            <div><span className="text-emerald-400/50">$</span> gorenel start --port 3000 --subdomain my-device-01</div>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div className="mt-4 rounded-xl border border-white/[0.06] bg-white/[0.02] p-4">
                                <div className="flex items-center gap-2 text-sm font-medium text-white/70 mb-3">
                                    <Command className="w-4 h-4 text-white/40" />
                                    Built-in capabilities
                                </div>
                                <div className="grid grid-cols-2 gap-2 text-xs text-white/40">
                                    <div className="inline-flex items-center gap-2"><CheckCircle2 className="w-3.5 h-3.5 text-emerald-400/70" /> Reserved URLs</div>
                                    <div className="inline-flex items-center gap-2"><CheckCircle2 className="w-3.5 h-3.5 text-emerald-400/70" /> Inspector + replay</div>
                                    <div className="inline-flex items-center gap-2"><CheckCircle2 className="w-3.5 h-3.5 text-emerald-400/70" /> ML anomaly signals</div>
                                    <div className="inline-flex items-center gap-2"><CheckCircle2 className="w-3.5 h-3.5 text-emerald-400/70" /> Per-tunnel policy</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Features Section */}
                <section className="border-t border-white/[0.04]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-20 space-y-12">
                        <div className="text-center space-y-3 max-w-xl mx-auto">
                            <h2 className="text-2xl md:text-3xl font-semibold tracking-tight">Built for production tunnels</h2>
                            <p className="text-sm md:text-base text-white/40 leading-relaxed">
                                Secure, observable, and fast by default — with deep debugging when you need it.
                            </p>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
                            {[
                                { icon: Shield, title: 'Edge security', desc: 'KeyAuth, BasicAuth, allowlists, redirect, limits.', color: 'text-emerald-400/70' },
                                { icon: Command, title: 'Developer workflow', desc: 'CLI-first, reserved URLs, service mode.', color: 'text-blue-400/70' },
                                { icon: Globe, title: 'Operations', desc: 'Region preference, runbooks, hardened timeouts.', color: 'text-violet-400/70' },
                            ].map((c) => (
                                <div key={c.title} className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-7 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-200 group">
                                    <div className="w-10 h-10 rounded-xl bg-white/[0.04] border border-white/[0.06] flex items-center justify-center mb-4 group-hover:border-white/[0.1] transition-colors">
                                        <c.icon className={`w-5 h-5 ${c.color}`} />
                                    </div>
                                    <h3 className="text-base font-semibold text-white mb-2">{c.title}</h3>
                                    <p className="text-sm text-white/40 leading-relaxed">{c.desc}</p>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>

                {/* Footer */}
                <footer className="border-t border-white/[0.04]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-8 flex flex-col md:flex-row items-center justify-between gap-4 text-sm text-white/25">
                        <span>&copy; 2026 Gorenel</span>
                        <div className="flex items-center gap-4 text-xs">
                            <span>Edge-ready</span>
                            <span className="w-1 h-1 rounded-full bg-white/10" />
                            <span>Secure by default</span>
                        </div>
                    </div>
                </footer>
            </main>
        </div>
    );
};
