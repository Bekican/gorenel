import React, { useEffect, useState } from 'react';
import {
    ArrowRight,
    CheckCircle2,
    Command,
    Globe,
    Languages,
    Shield,
    Activity,
    Eye,
    Brain,
    Server,
    Sparkles,
    Clock,
    Zap,
    Code2,
    Layers,
    Check,
    X as XIcon,
    Network,
    Cpu,
    Monitor,
    Copy,
    ExternalLink,
    Github,
    AlertCircle
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Button } from './ui/Button';

/* ─── SHARED COMPONENTS ─── */

const AnimatedNumber: React.FC<{ target: number; suffix?: string; prefix?: string; duration?: number }> = ({
    target, suffix = '', prefix = '', duration = 2000
}) => {
    const [current, setCurrent] = useState(0);
    useEffect(() => {
        const start = performance.now();
        const step = (ts: number) => {
            const elapsed = ts - start;
            const progress = Math.min(elapsed / duration, 1);
            const eased = 1 - Math.pow(1 - progress, 3);
            setCurrent(Math.round(target * eased));
            if (progress < 1) requestAnimationFrame(step);
        };
        requestAnimationFrame(step);
    }, [target, duration]);
    return <>{prefix}{current.toLocaleString()}{suffix}</>;
};

const CodeBlock: React.FC<{ code: string; language?: string }> = ({ code }) => {
    const [copied, setCopied] = useState(false);
    const copy = () => {
        navigator.clipboard.writeText(code);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };
    return (
        <div className="relative group rounded-xl border border-white/[0.08] bg-[#0c0e14] p-4 font-mono text-[13px] overflow-hidden">
            <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                    onClick={copy}
                    className="p-1.5 rounded-md bg-white/[0.05] border border-white/[0.1] hover:bg-white/[0.1] text-white/50 hover:text-white transition-all"
                >
                    {copied ? <Check size={14} className="text-emerald-400" /> : <Copy size={14} />}
                </button>
            </div>
            <div className="flex items-center gap-3">
                <span className="text-emerald-500/50 select-none">$</span>
                <code className="text-white/80">{code}</code>
            </div>
        </div>
    );
};

const FeatureCard: React.FC<{ icon: any; title: string; desc: string; color: string; badge?: string }> = ({
    icon: Icon, title, desc, color, badge
}) => (
    <div className="group relative rounded-2xl border border-white/[0.06] bg-white/[0.02] p-8 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-300">
        <div className="flex items-center justify-between mb-6">
            <div className={`w-12 h-12 rounded-xl bg-${color}-500/10 border border-${color}-500/20 flex items-center justify-center group-hover:scale-110 transition-transform duration-300`}>
                <Icon className={`w-6 h-6 text-${color}-400/80`} />
            </div>
            {badge && (
                <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-1 rounded-md bg-white/[0.05] text-white/30 border border-white/[0.08]">
                    {badge}
                </span>
            )}
        </div>
        <h3 className="text-lg font-bold text-white mb-3">{title}</h3>
        <p className="text-sm text-white/40 leading-relaxed">{desc}</p>
    </div>
);

const ComparisonRow: React.FC<{ feature: string; gorenel: boolean; ngrok: boolean; cf: boolean }> = ({
    feature, gorenel, ngrok, cf
}) => (
    <tr className="border-b border-white/[0.04] hover:bg-white/[0.01] transition-colors">
        <td className="py-4 px-6 text-sm font-medium text-white/60">{feature}</td>
        <td className="py-4 px-6 text-center">
            {gorenel ? <div className="flex justify-center"><CheckCircle2 size={18} className="text-emerald-400" /></div> : <XIcon size={16} className="text-white/10 mx-auto" />}
        </td>
        <td className="py-4 px-6 text-center">
            {ngrok ? <Check size={18} className="text-white/30 mx-auto" /> : <XIcon size={16} className="text-white/10 mx-auto" />}
        </td>
        <td className="py-4 px-6 text-center">
            {cf ? <Check size={18} className="text-white/30 mx-auto" /> : <XIcon size={16} className="text-white/10 mx-auto" />}
        </td>
    </tr>
);

/* ─── MAIN PAGE ─── */

interface LandingPageProps {
    onLogin: () => void;
    isLoggedIn?: boolean;
    onGoToDashboard?: () => void;
}

export const LandingPage: React.FC<LandingPageProps> = ({ onLogin, isLoggedIn, onGoToDashboard }) => {
    const { t, i18n } = useTranslation();
    const isTr = i18n.language === 'tr';
    const isWindowsClient = typeof navigator !== 'undefined' && /win/i.test(navigator.platform || '');
    const handleCTA = isLoggedIn ? onGoToDashboard : onLogin;

    return (
        <div className="min-h-screen bg-[#080a10] text-white selection:bg-emerald-500/30 selection:text-white">
            {/* Ambient Background */}
            <div className="fixed inset-0 pointer-events-none z-0">
                <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full h-[800px] bg-[radial-gradient(ellipse_at_center,rgba(16,185,129,0.08),transparent_70%)]" />
                <div className="absolute inset-0 [background-image:radial-gradient(circle_at_1px_1px,rgba(255,255,255,0.03)_1px,transparent_0)] [background-size:32px_32px]" />
                <div className="absolute inset-0 bg-gradient-to-b from-transparent via-[#080a10]/50 to-[#080a10]" />
            </div>

            {/* ─── Header ─── */}
            <nav className="sticky top-0 z-[100] border-b border-white/[0.04] bg-[#080a10]/80 backdrop-blur-md">
                <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center">
                            <img src="/logo.png" alt="Gorenel" className="w-5 h-5 object-contain" />
                        </div>
                        <span className="font-black tracking-tight text-lg">Gorenel</span>
                        <div className="px-2 py-0.5 rounded-full bg-white/[0.05] border border-white/[0.08] text-[10px] font-bold text-white/40 uppercase tracking-widest ml-1">v1.2</div>
                    </div>

                    <div className="hidden lg:flex items-center gap-8 text-[13px] font-medium text-white/50">
                        <a href="#features" className="hover:text-white transition-colors">{t('landing.trust_ai')}</a>
                        <a href="#how-it-works" className="hover:text-white transition-colors">{t('landing.how_it_works_title')}</a>
                        <a href="#comparison" className="hover:text-white transition-colors">{t('landing.comparison_title')}</a>
                        {/* GitHub Icon with Star count (mocked for now) */}
                        <a href="https://github.com/bekican/gorenel" className="flex items-center gap-2 hover:text-white transition-colors border-l border-white/10 pl-8">
                            <Github size={16} />
                            <span>GitHub</span>
                        </a>
                    </div>

                    <div className="flex items-center gap-3">
                        <button
                            onClick={() => i18n.changeLanguage(isTr ? 'en' : 'tr')}
                            className="p-2 rounded-lg bg-white/[0.03] border border-white/[0.06] hover:bg-white/[0.06] transition-all text-white/50"
                        >
                            <Languages size={16} />
                        </button>
                        <Button variant="ghost" size="sm" onClick={onLogin} className="text-white/60 hover:text-white">
                            {t('common.login')}
                        </Button>
                        <Button variant="primary" size="sm" onClick={handleCTA} className="font-bold shadow-lg shadow-emerald-500/10">
                            {isLoggedIn ? t('common.to_dashboard') : t('landing.cta_primary')}
                            <ArrowRight size={14} className="ml-1" />
                        </Button>
                    </div>
                </div>
            </nav>

            <main className="relative z-10">
                {/* ─── Hero Section ─── */}
                <section className="relative pt-24 pb-16 lg:pt-32 lg:pb-32 overflow-hidden">
                    <div className="max-w-7xl mx-auto px-6 grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
                        <div className="relative z-10 space-y-8 text-center lg:text-left">
                            <div className="inline-flex items-center gap-2 rounded-full border border-emerald-500/20 bg-emerald-500/[0.08] px-4 py-2 text-[11px] font-black uppercase tracking-[0.2em] text-emerald-400 animate-pulse-slow">
                                <Sparkles size={12} />
                                {t('landing.hero_badge')}
                            </div>

                            <h1 className="text-5xl lg:text-[4.5rem] font-black tracking-tighter leading-[1] text-gradient">
                                {t('landing.title')} <br />
                                <span className="text-gradient-accent">{t('landing.title_accent')}</span>
                            </h1>

                            <p className="text-lg lg:text-xl text-white/50 leading-relaxed max-w-xl mx-auto lg:mx-0">
                                {t('landing.subtitle')}
                            </p>

                            <div className="flex flex-col sm:flex-row items-center justify-center lg:justify-start gap-4 pt-4">
                                <Button variant="primary" size="lg" onClick={handleCTA} className="h-14 px-10 text-base font-bold w-full sm:w-auto">
                                    {isLoggedIn ? t('common.to_dashboard') : t('landing.cta_primary')}
                                    <ArrowRight size={18} className="ml-2" />
                                </Button>
                                <a
                                    href="https://github.com/bekican/gorenel"
                                    target="_blank"
                                    className="flex items-center gap-2 h-14 px-8 rounded-xl border border-white/[0.08] bg-white/[0.02] hover:bg-white/[0.05] transition-all font-bold w-full sm:w-auto justify-center"
                                >
                                    <Github size={20} />
                                    {t('landing.cta_secondary')}
                                </a>
                            </div>

                            {/* Trust Bar */}
                            <div className="flex flex-wrap items-center justify-center lg:justify-start gap-8 pt-8 opacity-40">
                                {[
                                    { icon: Shield, label: t('landing.trust_ai') },
                                    { icon: Github, label: t('landing.trust_open_source') },
                                    { icon: Server, label: t('landing.trust_self_host') }
                                ].map((item, idx) => (
                                    <div key={idx} className="flex items-center gap-2.5 text-[11px] font-bold uppercase tracking-widest">
                                        <item.icon size={16} className="text-emerald-500" />
                                        {item.label}
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* Hero Visual: Premium Terminal/UI Mockup */}
                        <div className="relative group perspective">
                            <div className="absolute -inset-10 bg-gradient-to-tr from-emerald-500/20 to-blue-500/10 rounded-full blur-[100px] opacity-50 group-hover:opacity-70 transition-opacity duration-1000" />

                            <div className="relative rounded-2xl border border-white/[0.1] bg-[#0c0e14]/90 backdrop-blur-2xl shadow-[0_32px_64px_-16px_rgba(0,0,0,0.5)] overflow-hidden transform-gpu lg:rotate-[-2deg] lg:group-hover:rotate-0 transition-transform duration-700">
                                <div className="h-10 border-b border-white/[0.08] bg-white/[0.02] flex items-center px-4 justify-between">
                                    <div className="flex gap-2">
                                        <div className="w-3 h-3 rounded-full bg-red-500/20 border border-red-500/40" />
                                        <div className="w-3 h-3 rounded-full bg-yellow-500/20 border border-yellow-500/40" />
                                        <div className="w-3 h-3 rounded-full bg-emerald-500/20 border border-emerald-500/40" />
                                    </div>
                                    <div className="text-[10px] uppercase font-bold tracking-widest text-white/30">gorenel-cli — session</div>
                                    <div className="flex items-center gap-2">
                                        <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                                        <span className="text-[9px] font-bold text-emerald-500/60">LIVE</span>
                                    </div>
                                </div>
                                <div className="p-8 space-y-6 font-mono text-[13px]">
                                    <div className="flex gap-3">
                                        <span className="text-emerald-500 select-none">❯</span>
                                        <span className="text-white/80">gorenel expose 3000</span>
                                    </div>
                                    <div className="space-y-2 pl-6">
                                        <p className="text-white/40 flex items-center gap-2">
                                            <Check size={14} className="text-emerald-500" />
                                            Establishing secure bridge to edge infrastructure...
                                        </p>
                                        <p className="text-white/40 flex items-center gap-2">
                                            <Check size={14} className="text-emerald-500" />
                                            AI Monitoring consensus: <span className="text-emerald-500/60 font-bold uppercase text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20">Active</span>
                                        </p>
                                        <p className="text-white/40 flex items-center gap-2">
                                            <Check size={14} className="text-emerald-500" />
                                            Provisioning SSL certificate (RSA 2048)
                                        </p>
                                    </div>
                                    <div className="pt-4 border-t border-white/[0.04]">
                                        <p className="text-white/60 mb-1">Tunnel established successfully!</p>
                                        <div className="p-4 rounded-xl border border-emerald-500/20 bg-emerald-500/[0.03] flex items-center justify-between">
                                            <div className="space-y-1">
                                                <p className="text-[11px] font-bold text-emerald-500/50 uppercase tracking-wider">Public Tunnel URL</p>
                                                <p className="text-emerald-400 font-bold text-base select-all">https://dev-session.gorenel.site</p>
                                            </div>
                                            <ExternalLink size={20} className="text-emerald-500/40" />
                                        </div>
                                    </div>
                                    <div className="grid grid-cols-2 gap-4 pt-4">
                                        <div className="rounded-xl border border-white/[0.05] bg-white/[0.02] p-3">
                                            <p className="text-[9px] font-bold text-white/30 uppercase tracking-widest mb-1">Stability</p>
                                            <p className="text-xs font-bold text-blue-400">99.99% Uptime</p>
                                        </div>
                                        <div className="rounded-xl border border-white/[0.05] bg-white/[0.02] p-3">
                                            <p className="text-[9px] font-bold text-white/30 uppercase tracking-widest mb-1">Security</p>
                                            <p className="text-xs font-bold text-violet-400 px-1">E2E AES-256</p>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* ─── Stats / Trust Indicators ─── */}
                <section className="border-y border-white/[0.04] bg-white/[0.01]">
                    <div className="max-w-7xl mx-auto px-6 py-12 md:py-20">
                        <div className="grid grid-cols-2 lg:grid-cols-4 gap-8 md:gap-12">
                            {[
                                { label: 'Avg Latency', val: 12, suffix: 'ms', icon: Clock },
                                { label: 'Traffic Security', val: 256, suffix: '-bit', icon: Shield },
                                { label: 'Inference Time', val: 1.2, suffix: 'ms', icon: Brain },
                                { label: 'Active Tunnels', val: 50, suffix: 'K+', icon: Globe }
                            ].map((stat, i) => (
                                <div key={i} className="text-center space-y-3">
                                    <div className="w-10 h-10 rounded-xl bg-white/[0.03] border border-white/[0.06] flex items-center justify-center mx-auto text-white/40">
                                        <stat.icon size={20} strokeWidth={1.5} />
                                    </div>
                                    <div className="text-4xl font-black tabular-nums">
                                        <AnimatedNumber target={stat.val} suffix={stat.suffix} />
                                    </div>
                                    <div className="text-[11px] font-black uppercase tracking-[0.2em] text-white/30">{stat.label}</div>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>

                {/* ─── How It Works (Quick Start) ─── */}
                <section id="how-it-works" className="py-24 lg:py-40">
                    <div className="max-w-7xl mx-auto px-6">
                        <div className="flex flex-col lg:flex-row gap-20 items-center">
                            <div className="flex-1 space-y-10">
                                <div className="space-y-6">
                                    <div className="inline-flex items-center gap-2 text-[11px] font-black uppercase tracking-[0.2em] text-blue-400">
                                        <Zap size={14} />
                                        {t('landing.how_it_works_title')}
                                    </div>
                                    <h2 className="text-4xl lg:text-5xl font-black tracking-tight leading-[1.1]">
                                        {t('landing.how_it_works_subtitle')}
                                    </h2>
                                </div>

                                <div className="space-y-8">
                                    {[
                                        { id: '01', title: t('landing.how_it_works_step1'), desc: 'One-line install for macOS, Linux, and Windows.' },
                                        { id: '02', title: t('landing.how_it_works_step2'), desc: 'Authenticate with your dashboard API key.' },
                                        { id: '03', title: t('landing.how_it_works_step3'), desc: 'Deploy your local port to a secure public URL instantly.' }
                                    ].map((step) => (
                                        <div key={step.id} className="flex gap-6 group">
                                            <div className="flex flex-col items-center">
                                                <div className="w-10 h-10 rounded-full border-2 border-white/10 bg-white/[0.02] flex items-center justify-center text-sm font-black group-hover:border-emerald-500/50 transition-colors">
                                                    {step.id}
                                                </div>
                                                <div className="flex-1 w-[2px] bg-gradient-to-b from-white/10 to-transparent mt-2" />
                                            </div>
                                            <div className="pb-8">
                                                <h3 className="text-xl font-bold text-white mb-2">{step.title}</h3>
                                                <p className="text-white/40 leading-relaxed text-sm max-w-sm">{step.desc}</p>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>

                            <div className="flex-1 w-full lg:w-auto">
                                <div className="rounded-3xl border border-white/[0.08] bg-[#0c0e14]/50 p-8 lg:p-12 space-y-8 border-gradient">
                                    <div className="space-y-4">
                                        <div className="flex items-center justify-between">
                                            <span className="text-[10px] font-black uppercase tracking-widest text-white/30">1. Install Gorenel</span>
                                            <div className="flex gap-1.5 grayscale opacity-30">
                                                <Monitor size={14} />
                                                <Layers size={14} />
                                            </div>
                                        </div>
                                        <CodeBlock 
                                            code={isWindowsClient 
                                                ? `iwr -useb ${typeof window !== 'undefined' ? window.location.host : 'gorenel.site'}/install.ps1 | iex` 
                                                : `curl -sSL ${typeof window !== 'undefined' ? window.location.host : 'gorenel.site'}/install.sh | bash`} 
                                        />
                                    </div>

                                    <div className="space-y-4">
                                        <span className="text-[10px] font-black uppercase tracking-widest text-white/30">2. Command line auth</span>
                                        <CodeBlock code="gorenel config set api_key gk_*****" />
                                    </div>

                                    <div className="space-y-4">
                                        <span className="text-[10px] font-black uppercase tracking-widest text-white/30">3. Connect port 3000</span>
                                        <CodeBlock code="gorenel expose 3000" />
                                    </div>

                                    <div className="pt-6 border-t border-white/[0.04] flex items-center gap-4 text-white/30">
                                        <AlertCircle size={16} />
                                        <p className="text-[11px] leading-relaxed">
                                            SSL certificates, DDoS protection, and AI anomaly detection <br />
                                            are enabled automatically for every session.
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* ─── AI Deep Dive ─── */}
                <section className="relative py-24 lg:py-40 border-t border-white/[0.04] overflow-hidden">
                    <div className="absolute top-1/2 left-0 -translate-y-1/2 w-[500px] h-[500px] bg-emerald-500/[0.05] rounded-full blur-[120px] pointer-events-none" />
                    <div className="max-w-7xl mx-auto px-6 grid grid-cols-1 lg:grid-cols-2 gap-20 items-center">
                        <div className="relative order-2 lg:order-1">
                            <div className="absolute -inset-10 bg-gradient-to-tr from-violet-500/10 to-emerald-500/10 rounded-full blur-[100px] opacity-20" />
                            <div className="relative rounded-2xl border border-white/[0.1] bg-[#0c0e14]/80 p-8 space-y-8 shadow-2xl">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-3">
                                        <div className="w-10 h-10 rounded-xl bg-violet-500/10 border border-violet-500/20 flex items-center justify-center">
                                            <Brain className="w-5 h-5 text-violet-400" />
                                        </div>
                                        <div>
                                            <p className="text-[11px] font-black uppercase tracking-widest text-white/30">Intelligence Model</p>
                                            <p className="text-sm font-bold text-white">Isolation Forest Consensus</p>
                                        </div>
                                    </div>
                                    <span className="px-2 py-1 rounded bg-emerald-500/10 text-emerald-400 text-[10px] font-bold border border-emerald-500/20">99.2% Conf.</span>
                                </div>
                                <div className="space-y-4">
                                    <div className="space-y-2">
                                        <div className="flex justify-between text-[11px] text-white/40 font-bold uppercase tracking-widest">
                                            <span>Traffic Anomaly Detection</span>
                                            <span>Active</span>
                                        </div>
                                        <div className="h-2 rounded-full bg-white/[0.05] overflow-hidden">
                                            <div className="h-full bg-emerald-500 w-[92%] animate-pulse" />
                                        </div>
                                    </div>
                                    <div className="space-y-2">
                                        <div className="flex justify-between text-[11px] text-white/40 font-bold uppercase tracking-widest">
                                            <span>Behavioral Autoencoder</span>
                                            <span>Calibrating</span>
                                        </div>
                                        <div className="h-2 rounded-full bg-white/[0.05] overflow-hidden">
                                            <div className="h-full bg-violet-500 w-[74%]" />
                                        </div>
                                    </div>
                                </div>
                                <div className="flex items-center gap-4 p-4 rounded-xl bg-white/[0.02] border border-white/[0.06] text-xs text-white/50 italic leading-relaxed">
                                    <Command size={18} className="text-violet-400 shrink-0" />
                                    "Model detected a structural pattern shift in Header Frames. Blocking potential SQLi attempt locally at the edge."
                                </div>
                            </div>
                        </div>

                        <div className="space-y-10 order-1 lg:order-2">
                            <div className="space-y-6">
                                <div className="inline-flex items-center gap-2 text-[11px] font-black uppercase tracking-[0.2em] text-violet-400">
                                    <Cpu size={14} />
                                    {t('landing.ai_title')}
                                </div>
                                <h2 className="text-4xl lg:text-5xl font-black tracking-tight leading-[1.1]">
                                    {t('landing.ai_subtitle')}
                                </h2>
                                <p className="text-lg text-white/40 leading-relaxed max-w-xl">
                                    {t('landing.ai_desc')}
                                </p>
                            </div>

                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                {[
                                    { label: 'Real-time Inference', detail: 'Sub-millisecond analysis' },
                                    { label: 'Layer 7 Protection', detail: 'Deep packet inspection' },
                                    { label: 'Protocol Aware', detail: 'Understands HTTP/AI traffic' },
                                    { label: 'Offline Analysis', detail: 'History saved locally' }
                                ].map((item, idx) => (
                                    <div key={idx} className="flex items-start gap-3 p-4 rounded-2xl bg-white/[0.02] border border-white/[0.06]">
                                        <div className="mt-1 w-5 h-5 rounded bg-violet-500/20 border border-violet-500/20 flex items-center justify-center shrink-0">
                                            <Check size={12} className="text-violet-400" />
                                        </div>
                                        <div>
                                            <p className="text-sm font-bold text-white/90">{item.label}</p>
                                            <p className="text-[11px] text-white/30">{item.detail}</p>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </section>

                {/* ─── Feature Grid ─── */}
                <section id="features" className="py-24 lg:py-40 bg-[radial-gradient(ellipse_at_center,rgba(59,130,246,0.02),transparent_70%)]">
                    <div className="max-w-7xl mx-auto px-6 space-y-20">
                        <div className="text-center space-y-6 max-w-2xl mx-auto">
                            <div className="inline-flex items-center gap-2 text-[11px] font-black uppercase tracking-[0.2em] text-cyan-400">
                                <Layers size={14} />
                                {t('landing.features_title')}
                            </div>
                            <h2 className="text-4xl lg:text-5xl font-black tracking-tight leading-[1.1]">
                                {t('landing.features_subtitle')}
                            </h2>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            <FeatureCard
                                icon={Globe}
                                color="emerald"
                                badge="Persistent"
                                title={t('landing.trust_self_host')}
                                desc="Reserve custom subdomains that never change. Build integrations that stay stable even after your terminal closes."
                            />
                            <FeatureCard
                                icon={Eye}
                                color="blue"
                                badge="Real-time"
                                title="Traffic Inspector"
                                desc="Intercept and replay HTTP frames in real-time. Debug webhooks and APIs without leaving your dashboard."
                            />
                            <FeatureCard
                                icon={Shield}
                                color="violet"
                                badge="Edge Auth"
                                title="Zero-Trust Proxy"
                                desc="Add Key-Auth, Basic-Auth, or IP Allow-listing to your localhost in one click. Production security for local dev."
                            />
                            <FeatureCard
                                icon={Zap}
                                color="amber"
                                badge="Zero Latency"
                                title="Smart CDN Routing"
                                desc="Gorenel automatically routes your traffic through the global edge nearest to you for zero-added latency."
                            />
                            <FeatureCard
                                icon={Activity}
                                color="rose"
                                badge="Live Stats"
                                title="Real-time Metrics"
                                desc="Monitor throughput, latency, and system load with a live-streaming CLI and Web interface."
                            />
                            <FeatureCard
                                icon={Code2}
                                color="cyan"
                                badge="Developer-First"
                                title="Scriptable CLI"
                                desc="Native binaries for every platform. Export and manage configurations with simple JSON/YAML files."
                            />
                        </div>
                    </div>
                </section>

                {/* ─── Comparison Section ─── */}
                <section id="comparison" className="py-24 lg:py-40 border-t border-white/[0.04]">
                    <div className="max-w-7xl mx-auto px-6 space-y-16">
                        <div className="text-center space-y-6 max-w-2xl mx-auto">
                            <div className="inline-flex items-center gap-2 text-[11px] font-black uppercase tracking-[0.2em] text-white/30">
                                <Github size={14} />
                                {t('landing.comparison_title')}
                            </div>
                            <h2 className="text-4xl lg:text-5xl font-black tracking-tight leading-[1.1]">
                                {t('landing.comparison_subtitle')}
                            </h2>
                        </div>

                        <div className="max-w-4xl mx-auto rounded-3xl border border-white/[0.08] bg-white/[0.01] overflow-hidden shadow-2xl">
                            <table className="w-full">
                                <thead>
                                    <tr className="border-b border-white/[0.08] bg-white/[0.02]">
                                        <th className="py-6 px-6 text-left text-[11px] font-black uppercase tracking-widest text-white/30">Feature Capability</th>
                                        <th className="py-6 px-6 text-center">
                                            <div className="flex flex-col items-center">
                                                <div className="w-8 h-8 rounded-lg bg-emerald-500/20 flex items-center justify-center mb-2">
                                                    <img src="/logo.png" className="w-4 h-4" alt="Gorenel" />
                                                </div>
                                                <span className="text-sm font-black text-emerald-400">Gorenel</span>
                                            </div>
                                        </th>
                                        <th className="py-6 px-6 text-center text-sm font-black text-white/40">ngrok</th>
                                        <th className="py-6 px-6 text-center text-sm font-black text-white/40">Cloudflare</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <ComparisonRow feature="AI Protocol Intelligence" gorenel ngrok={false} cf={false} />
                                    <ComparisonRow feature="Real-time Traffic Inspector" gorenel ngrok={true} cf={false} />
                                    <ComparisonRow feature="Dynamic Edge Rules (Mock/Auth)" gorenel ngrok={false} cf={true} />
                                    <ComparisonRow feature="Open Source CLI Engine" gorenel ngrok={false} cf={false} />
                                    <ComparisonRow feature="Reserved URLs (Free Plan)" gorenel ngrok={true} cf={true} />
                                    <ComparisonRow feature="Self-Hostable Backend" gorenel ngrok={false} cf={false} />
                                    <ComparisonRow feature="Unlimited Local Tunnels" gorenel ngrok={false} cf={true} />
                                </tbody>
                            </table>
                        </div>
                    </div>
                </section>

                {/* ─── Usage Cases Section ─── */}
                <section className="py-24 lg:py-40 border-t border-white/[0.04] bg-white/[0.01]">
                    <div className="max-w-7xl mx-auto px-6 space-y-16 text-center">
                        <div className="space-y-6 max-w-2xl mx-auto">
                            <div className="inline-flex items-center gap-2 text-[11px] font-black uppercase tracking-[0.2em] text-blue-400">
                                <Network size={14} />
                                {t('landing.use_cases_title')}
                            </div>
                            <h2 className="text-4xl lg:text-5xl font-black tracking-tight">Built for every scenario</h2>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                            {[
                                { icon: Code2, title: 'Webhook Debugging', desc: t('landing.use_cases_webhooks') },
                                { icon: Monitor, title: 'Client Presentations', desc: t('landing.use_cases_demos') },
                                { icon: Cpu, title: 'IoT & Remote Access', desc: t('landing.use_cases_iot') },
                                { icon: Brain, title: 'AI/LLM Gateway Proxy', desc: t('landing.use_cases_ai') }
                            ].map((useCase, idx) => (
                                <div key={idx} className="group p-8 rounded-3xl border border-white/[0.06] bg-[#0c0e14] hover:border-blue-500/30 transition-all text-left flex gap-6 items-start">
                                    <div className="w-12 h-12 rounded-2xl bg-white/[0.03] border border-white/[0.06] flex items-center justify-center shrink-0 group-hover:scale-110 transition-transform">
                                        <useCase.icon size={24} className="text-white/40 group-hover:text-blue-400" />
                                    </div>
                                    <div className="space-y-2">
                                        <h3 className="text-xl font-bold text-white">{useCase.title}</h3>
                                        <p className="text-sm text-white/40 leading-relaxed">{useCase.desc}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>

                {/* ─── Community / Free Section ─── */}
                <section id="pricing" className="py-24 lg:py-40 bg-gradient-to-t from-emerald-500/[0.03] to-transparent">
                    <div className="max-w-7xl mx-auto px-6">
                        <div className="rounded-[40px] border border-emerald-500/20 bg-emerald-500/[0.02] p-8 lg:p-24 text-center space-y-10 relative overflow-hidden">
                            <div className="absolute top-0 left-0 w-full h-full bg-[radial-gradient(circle_at_top,rgba(16,185,129,0.05),transparent_70%)] pointer-events-none" />

                            <div className="space-y-6 relative z-10">
                                <div className="inline-flex items-center gap-2 text-[11px] font-black uppercase tracking-[0.2em] text-emerald-400">
                                    <Languages size={14} />
                                    {t('landing.pricing_title')}
                                </div>
                                <h2 className="text-5xl lg:text-7xl font-black tracking-tighter">
                                    {t('landing.pricing_subtitle')}
                                </h2>
                                <p className="text-lg text-white/40 max-w-2xl mx-auto italic">
                                    "We believe developers shouldn't pay to share their work. Gorenel is, and always will be, free for the community."
                                </p>
                            </div>

                            <div className="flex flex-col items-center gap-8 relative z-10">
                                <Button variant="primary" size="lg" onClick={handleCTA} className="h-16 px-16 text-lg font-black shadow-2xl shadow-emerald-500/20">
                                    {t('landing.cta_primary')}
                                    <ArrowRight size={20} className="ml-2" />
                                </Button>

                                <div className="flex flex-wrap justify-center gap-x-12 gap-y-6 opacity-40">
                                    {[
                                        'Unlimited Access',
                                        'No Pro Tier',
                                        'E2E Encryption',
                                        'Commercial Ready'
                                    ].map((f, i) => (
                                        <div key={i} className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest">
                                            <CheckCircle2 size={14} className="text-emerald-500" />
                                            {f}
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* ─── Footer ─── */}
                <footer className="pt-24 pb-12 border-t border-white/[0.04]">
                    <div className="max-w-7xl mx-auto px-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-12 mb-20">
                        <div className="lg:col-span-2 space-y-8">
                            <div className="flex items-center gap-3">
                                <div className="w-8 h-8 rounded-lg bg-white/[0.05] border border-white/[0.08] flex items-center justify-center">
                                    <img src="/logo.png" className="w-5 h-5 opacity-80" alt="Gorenel" />
                                </div>
                                <span className="font-black text-xl">Gorenel</span>
                            </div>
                            <p className="text-sm text-white/30 max-w-xs leading-relaxed italic">
                                The AI-powered tunneling infrastructure for secure, high-performance localhost deployments.
                            </p>
                            <div className="flex gap-4">
                                <Button variant="ghost" size="sm" className="w-10 h-10 p-0 rounded-xl bg-white/[0.02] border border-white/[0.06] hover:text-white text-white/30">
                                    <Github size={20} />
                                </Button>
                                <Button variant="ghost" size="sm" className="w-10 h-10 p-0 rounded-xl bg-white/[0.02] border border-white/[0.06] hover:text-white text-white/30">
                                    <Languages size={20} />
                                </Button>
                            </div>
                        </div>

                        <div className="space-y-6">
                            <h4 className="text-[10px] font-black uppercase tracking-widest text-white/20">Product</h4>
                            <ul className="space-y-4 text-sm text-white/50">
                                <li><a href="#features" className="hover:text-emerald-400 transition-colors">Features</a></li>
                                <li><a href="#comparison" className="hover:text-emerald-400 transition-colors">Compare</a></li>
                                <li><a href="#" className="hover:text-emerald-400 transition-colors">Reserved Domains</a></li>
                                <li><a href="#" className="hover:text-emerald-400 transition-colors">Pricing</a></li>
                            </ul>
                        </div>

                        <div className="space-y-6">
                            <h4 className="text-[10px] font-black uppercase tracking-widest text-white/20">Resources</h4>
                            <ul className="space-y-4 text-sm text-white/50">
                                <li><a href="#" className="hover:text-emerald-400 transition-colors">Documentation</a></li>
                                <li><a href="#" className="hover:text-emerald-400 transition-colors">CLI Reference</a></li>
                                <li><a href="#" className="hover:text-emerald-400 transition-colors">Security Audit</a></li>
                                <li><a href="#" className="hover:text-emerald-400 transition-colors">API Keys</a></li>
                            </ul>
                        </div>

                        <div className="space-y-6">
                            <h4 className="text-[10px] font-black uppercase tracking-widest text-white/20">Legal</h4>
                            <ul className="space-y-4 text-sm text-white/50">
                                <li><a href="#" className="hover:text-emerald-400 transition-colors">Privacy Policy</a></li>
                                <li><a href="#" className="hover:text-emerald-400 transition-colors">Terms of Service</a></li>
                                <li><a href="#" className="hover:text-emerald-400 transition-colors">GDPR</a></li>
                            </ul>
                        </div>
                    </div>

                    <div className="max-w-7xl mx-auto px-6 pt-12 border-t border-white/[0.04] flex flex-col md:flex-row justify-between items-center gap-6">
                        <p className="text-[10px] font-bold uppercase tracking-widest text-white/20">© 2026 Core Infrastructure. Built for Humanity.</p>
                        <div className="flex items-center gap-6">
                            <div className="flex items-center gap-2">
                                <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" />
                                <span className="text-[10px] font-bold uppercase tracking-widest text-white/30">All systems green</span>
                            </div>
                            <span className="text-[10px] font-bold text-white/10 select-none tracking-widest">v1.2.4-stable</span>
                        </div>
                    </div>
                </footer>
            </main>
        </div>
    );
};
