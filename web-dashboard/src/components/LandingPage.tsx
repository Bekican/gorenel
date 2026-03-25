import React from 'react';
import { ArrowRight, CheckCircle2, Command, Globe, Languages, Lock, Shield, TerminalSquare, Zap } from 'lucide-react';
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
        <div className="min-h-screen bg-[#020408] text-white selection:bg-emerald-500/30 font-sans">
            <div className="fixed inset-0 pointer-events-none">
                <div className="absolute -top-24 -left-24 w-[520px] h-[520px] bg-emerald-500/10 rounded-full blur-[140px]" />
                <div className="absolute -bottom-24 -right-24 w-[520px] h-[520px] bg-blue-600/10 rounded-full blur-[160px]" />
                <div className="absolute inset-0 bg-gradient-to-b from-black/0 via-black/35 to-black/85" />
                <div className="absolute inset-0 opacity-[0.25] [background-image:radial-gradient(circle_at_1px_1px,rgba(255,255,255,0.06)_1px,transparent_0)] [background-size:22px_22px]" />
            </div>

            <nav className="sticky top-0 z-50 border-b border-white/5 bg-[#020408]/70 backdrop-blur-xl">
                <div className="max-w-7xl mx-auto px-6 md:px-12 py-4 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="w-9 h-9 rounded-2xl bg-white/5 border border-white/10 overflow-hidden shadow-[0_0_20px_rgba(16,185,129,0.15)]">
                            <img src="/logo.png" alt="Gorenel" className="w-full h-full object-cover" />
                        </div>
                        <div className="font-black tracking-tight text-white">Gorenel</div>
                    </div>
                    <div className="flex items-center gap-3">
                        <button
                            onClick={toggleLanguage}
                            className="flex items-center gap-2 px-3 py-2 bg-white/5 border border-white/10 rounded-xl text-[11px] font-black hover:bg-white/10 transition-all uppercase"
                            type="button"
                        >
                            <Languages size={14} className="text-emerald-400" />
                            {i18n.language.toUpperCase()}
                        </button>
                        <Button variant="outline" size="sm" type="button" onClick={onLogin}>
                            {t('common.login')}
                        </Button>
                        <Button variant="light" size="sm" type="button" onClick={isLoggedIn ? onGoToDashboard : onLogin}>
                            {isLoggedIn ? 'Open dashboard' : 'Start building'}
                            <ArrowRight size={14} />
                        </Button>
                    </div>
                </div>
            </nav>

            <main className="relative z-10">
                <section className="max-w-7xl mx-auto px-6 md:px-12 pt-16 md:pt-24 pb-14">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
                        <div className="space-y-8">
                            <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.02] px-3 py-1 text-[11px] font-black tracking-[0.22em] text-white/60">
                                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 shadow-[0_0_16px_rgba(16,185,129,0.65)]" />
                                SECURE TUNNELS · RESERVED URLS · TRAFFIC POLICY
                            </div>

                            <h1 className="text-4xl md:text-6xl font-black tracking-tight leading-[1.05]">
                                {i18n.language === 'tr' ? 'Güvenli tüneller.' : 'Secure tunnels.'}
                                <span className="block text-white/65">
                                    {i18n.language === 'tr' ? 'Rezervasyonlu URL ve traffic policy ile.' : 'With reserved URLs and traffic policy.'}
                                </span>
                            </h1>

                            <p className="text-base md:text-lg text-white/45 leading-relaxed max-w-xl">
                                {t('landing.description')}
                            </p>

                            <div className="flex flex-col sm:flex-row gap-3">
                                <Button
                                    variant="light"
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

                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
                                {[
                                    { icon: <Globe className="w-4 h-4 text-white/70" />, title: 'Reserved URLs', desc: 'Stable subdomains per device/customer.' },
                                    { icon: <Shield className="w-4 h-4 text-white/70" />, title: 'Traffic policy', desc: 'Auth, allowlist, rate limit, rewrite.' },
                                    { icon: <Lock className="w-4 h-4 text-white/70" />, title: 'Edge security', desc: 'BasicAuth + KeyAuth before your origin.' },
                                    { icon: <TerminalSquare className="w-4 h-4 text-white/70" />, title: 'Always-on', desc: 'Service mode with auto-reconnect.' },
                                ].map((f) => (
                                    <div key={f.title} className="rounded-3xl border border-white/10 bg-white/[0.015] p-4 hover:bg-white/[0.03] transition-colors">
                                        <div className="flex items-center gap-2 text-sm font-black text-white">
                                            {f.icon}
                                            {f.title}
                                        </div>
                                        <div className="mt-1 text-[12px] text-white/45 font-semibold leading-relaxed">{f.desc}</div>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div className="relative">
                            <div className="absolute -inset-4 bg-gradient-to-r from-emerald-500/20 to-blue-500/20 rounded-[2.5rem] blur-2xl opacity-60" />
                            <div className="relative rounded-[2.5rem] border border-white/10 bg-[#0A0C10]/70 backdrop-blur-xl shadow-2xl overflow-hidden">
                                <div className="p-5 border-b border-white/10 flex items-center justify-between">
                                    <div className="text-[11px] font-black uppercase tracking-[0.28em] text-white/40">Traffic policy</div>
                                    <div className="text-[11px] font-black text-white/30">Flow</div>
                                </div>
                                <div className="p-6 space-y-4">
                                    <div className="rounded-3xl border border-white/10 bg-black/35 p-5">
                                        <div className="grid grid-cols-3 gap-4 text-[11px] font-black uppercase tracking-[0.24em] text-white/35">
                                            <div className="flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-white/15" /> Internet</div>
                                            <div className="flex items-center gap-2 justify-center"><span className="w-2 h-2 rounded-full bg-emerald-400/70 shadow-[0_0_12px_rgba(16,185,129,0.45)]" /> Gorenel edge</div>
                                            <div className="flex items-center gap-2 justify-end"><span className="w-2 h-2 rounded-full bg-white/15" /> Your service</div>
                                        </div>

                                        <div className="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4 items-center">
                                            <div className="rounded-2xl border border-white/10 bg-white/[0.02] p-4">
                                                <div className="text-[10px] font-black uppercase tracking-[0.24em] text-white/35">Request</div>
                                                <div className="mt-2 font-mono text-[12px] text-white/70">GET /api/users</div>
                                            </div>
                                            <div className="rounded-2xl border border-emerald-500/20 bg-emerald-500/10 p-4">
                                                <div className="text-[10px] font-black uppercase tracking-[0.24em] text-emerald-100/70">Policies</div>
                                                <ul className="mt-2 space-y-1 text-[12px] font-semibold text-emerald-50/80">
                                                    <li className="flex items-center justify-between"><span>KeyAuth / BasicAuth</span><span className="text-emerald-200/70">✓</span></li>
                                                    <li className="flex items-center justify-between"><span>IP allowlist</span><span className="text-emerald-200/70">✓</span></li>
                                                    <li className="flex items-center justify-between"><span>Rate limit</span><span className="text-emerald-200/70">✓</span></li>
                                                    <li className="flex items-center justify-between"><span>Rewrite / headers</span><span className="text-emerald-200/70">✓</span></li>
                                                </ul>
                                            </div>
                                            <div className="rounded-2xl border border-white/10 bg-white/[0.02] p-4">
                                                <div className="text-[10px] font-black uppercase tracking-[0.24em] text-white/35">Forward</div>
                                                <div className="mt-2 font-mono text-[12px] text-white/70">localhost:3000</div>
                                            </div>
                                        </div>
                                    </div>

                                    <div className="rounded-2xl border border-white/10 bg-white/[0.02] p-4">
                                        <div className="text-[10px] font-black uppercase tracking-[0.24em] text-white/35">Quick start</div>
                                        <div className="mt-2 font-mono text-[12px] text-white/70">
                                            <span className="text-white/50">$</span> gorenel config set api_key gk_********
                                            <br />
                                            <span className="text-white/50">$</span> gorenel start --port 3000 --subdomain my-device-01
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div className="mt-6 rounded-3xl border border-white/10 bg-white/[0.015] p-5">
                                <div className="flex items-center gap-2 text-sm font-black">
                                    <Command className="w-4 h-4 text-white/70" />
                                    Built-in capabilities
                                </div>
                                <div className="mt-2 grid grid-cols-1 sm:grid-cols-2 gap-2 text-[12px] text-white/45 font-semibold">
                                    <div className="inline-flex items-center gap-2"><CheckCircle2 className="w-4 h-4 text-emerald-400" /> Reserved URLs</div>
                                    <div className="inline-flex items-center gap-2"><CheckCircle2 className="w-4 h-4 text-emerald-400" /> Inspector + replay</div>
                                    <div className="inline-flex items-center gap-2"><CheckCircle2 className="w-4 h-4 text-emerald-400" /> ML anomaly signals</div>
                                    <div className="inline-flex items-center gap-2"><CheckCircle2 className="w-4 h-4 text-emerald-400" /> Per-tunnel traffic policy</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                <section className="border-t border-white/5 bg-white/[0.015]">
                    <div className="max-w-7xl mx-auto px-6 md:px-12 py-16 space-y-10">
                        <div className="text-center space-y-3">
                            <div className="text-3xl md:text-4xl font-black tracking-tight">Built for production tunnels</div>
                            <div className="text-sm md:text-base text-white/45 font-semibold max-w-2xl mx-auto">
                                Secure, observable, and fast by default — with knobs to go deep when debugging.
                            </div>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                            {[
                                { icon: Shield, title: 'Edge security', desc: 'KeyAuth, BasicAuth, allowlists, redirect, limits.' },
                                { icon: Command, title: 'Developer workflow', desc: 'CLI-first, reserved URLs, service mode.' },
                                { icon: Globe, title: 'Operations', desc: 'Region preference + runbooks + hardened timeouts.' },
                            ].map((c) => (
                                <div key={c.title} className="rounded-[2.5rem] border border-white/10 bg-[#0A0C10]/60 backdrop-blur-xl p-8">
                                    <div className="w-12 h-12 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center">
                                        <c.icon className="w-6 h-6 text-primary" />
                                    </div>
                                    <div className="mt-4 text-lg font-black">{c.title}</div>
                                    <div className="mt-2 text-sm text-white/45 font-semibold leading-relaxed">{c.desc}</div>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>

                <footer className="border-t border-white/5">
                    <div className="max-w-7xl mx-auto px-6 md:px-12 py-10 flex flex-col md:flex-row items-center justify-between gap-4 text-sm text-white/30">
                        <div className="font-semibold">&copy; 2026 Gorenel</div>
                        <div className="flex items-center gap-3 text-[12px] font-semibold">
                            <span className="text-white/35">Edge-ready</span>
                            <span className="w-1 h-1 rounded-full bg-white/10" />
                            <span className="text-white/35">Secure by default</span>
                        </div>
                    </div>
                </footer>
            </main>
        </div>
    );
};

