import React, { useEffect, useState } from 'react';
import {
    ArrowRight,
    CheckCircle2,
    Command,
    Globe,
    Languages,
    Lock,
    Shield,
    TerminalSquare,
    Activity,
    Gauge,
    Eye,
    Brain,
    Server,
    Sparkles,
    Clock,
    Zap,
    Code2,
    GitBranch,
    Layers,
    BarChart3,
    Check,
    X as XIcon,
    Network,
    Cpu,
    Monitor
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Button } from './ui/Button';

interface LandingPageProps {
    onLogin: () => void;
    isLoggedIn?: boolean;
    onGoToDashboard?: () => void;
}

/* ---------- Animated counter ---------- */
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

export const LandingPage: React.FC<LandingPageProps> = ({ onLogin, isLoggedIn, onGoToDashboard }) => {
    const { t, i18n } = useTranslation();
    const isTr = i18n.language === 'tr';

    const toggleLanguage = () => {
        i18n.changeLanguage(isTr ? 'en' : 'tr');
    };

    const handleCTA = isLoggedIn ? onGoToDashboard : onLogin;

    return (
        <div className="min-h-screen bg-[#080a10] text-white font-sans">
            {/* Ambient background */}
            <div className="fixed inset-0 pointer-events-none overflow-hidden">
                <div className="absolute -top-32 left-1/4 w-[600px] h-[600px] bg-emerald-500/[0.07] rounded-full blur-[150px]" />
                <div className="absolute top-1/3 right-0 w-[400px] h-[400px] bg-cyan-500/[0.04] rounded-full blur-[120px]" />
                <div className="absolute bottom-0 right-1/4 w-[500px] h-[500px] bg-blue-500/[0.04] rounded-full blur-[140px]" />
                <div className="absolute inset-0 bg-gradient-to-b from-transparent via-transparent to-[#080a10]/80" />
                <div className="absolute inset-0 opacity-[0.12] [background-image:radial-gradient(circle_at_1px_1px,rgba(255,255,255,0.04)_1px,transparent_0)] [background-size:24px_24px]" />
            </div>

            {/* ─── Navigation ─── */}
            <nav className="sticky top-0 z-50 border-b border-white/[0.04] bg-[#080a10]/80 backdrop-blur-xl">
                <div className="max-w-6xl mx-auto px-6 md:px-10 py-3.5 flex items-center justify-between">
                    <div className="flex items-center gap-2.5">
                        <div className="w-8 h-8 rounded-lg bg-white/[0.05] border border-white/[0.08] overflow-hidden shadow-glow-emerald">
                            <img src="/logo.png" alt="Gorenel" width="256" height="256" className="w-full h-full object-cover" />
                        </div>
                        <span className="font-bold tracking-tight text-white text-[15px]">Gorenel</span>
                        <span className="hidden sm:inline-flex ml-2 text-[10px] font-medium px-2 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
                            v1.0
                        </span>
                    </div>
                    <div className="hidden md:flex items-center gap-6 text-sm text-white/50">
                        <a href="#features" className="hover:text-white transition-colors">{isTr ? 'Özellikler' : 'Features'}</a>
                        <a href="#how-it-works" className="hover:text-white transition-colors">{isTr ? 'Nasıl Çalışır' : 'How it works'}</a>
                        <a href="#pricing" className="hover:text-white transition-colors">{isTr ? 'Neden Ücretsiz?' : 'Why Free?'}</a>
                        <a href="#comparison" className="hover:text-white transition-colors">{isTr ? 'Karşılaştırma' : 'Compare'}</a>
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
                        <Button variant="primary" size="sm" type="button" onClick={handleCTA}>
                            {isLoggedIn ? 'Dashboard' : (isTr ? 'Ücretsiz Başla' : 'Start free')}
                            <ArrowRight size={14} />
                        </Button>
                    </div>
                </div>
            </nav>

            <main className="relative z-10">
                {/* ═══════════════════════  HERO  ═══════════════════════ */}
                <section className="max-w-6xl mx-auto px-6 md:px-10 pt-20 md:pt-28 pb-20">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
                        <div className="space-y-8">
                            {/* Badge */}
                            <div className="inline-flex items-center gap-2 rounded-full border border-emerald-500/20 bg-emerald-500/[0.06] px-3.5 py-1.5 text-xs font-medium text-emerald-300/80 animate-fade-in">
                                <Sparkles className="w-3.5 h-3.5" />
                                {isTr ? 'Ngrok alternatifi — Yapay zeka destekli' : 'Ngrok alternative — AI-powered'}
                            </div>

                            {/* Headline */}
                            <h1 className="text-4xl md:text-5xl lg:text-[3.5rem] font-bold tracking-tight leading-[1.08]">
                                {isTr
                                    ? <>Localhost'unuzu <span className="text-gradient-accent">saniyeler</span> içinde <br className="hidden md:block" />dünyaya açın.</>
                                    : <>Ship localhost to <br className="hidden md:block" />the world in <span className="text-gradient-accent">seconds</span>.</>
                                }
                            </h1>

                            {/* Subheadline */}
                            <p className="text-base md:text-lg text-white/65 leading-relaxed max-w-lg">
                                {isTr
                                    ? 'Güvenli tüneller, sabit URL\'ler, trafik politikaları ve yapay zeka ile anomali tespiti — tek bir CLI komutuyla. Ngrok\'tan daha hızlı, daha güvenli ve tamamen açık kaynak.'
                                    : 'Secure tunnels, stable URLs, traffic policies and AI-powered anomaly detection — with a single CLI command. Faster than ngrok, more secure, and fully open-source.'
                                }
                            </p>

                            {/* CTA Buttons */}
                            <div className="flex flex-col sm:flex-row gap-3">
                                <Button variant="primary" size="lg" type="button" onClick={handleCTA}>
                                    {isLoggedIn
                                        ? (isTr ? 'Dashboard\'a Git' : 'Open Dashboard')
                                        : (isTr ? 'Ücretsiz Başlayın' : 'Get started — it\'s free')
                                    }
                                    <ArrowRight size={16} />
                                </Button>
                                <Button variant="outline" size="lg" type="button" onClick={onLogin}>
                                    <TerminalSquare size={15} />
                                    {isTr ? 'Canlı Demo' : 'Live demo'}
                                </Button>
                            </div>

                            {/* Trust signals */}
                            <div className="flex flex-wrap items-center gap-x-5 gap-y-2 pt-2 text-xs text-white/55">
                                <div className="flex items-center gap-1.5">
                                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400/60" />
                                    {isTr ? 'Kredi kartı gerekmez' : 'No credit card required'}
                                </div>
                                <div className="flex items-center gap-1.5">
                                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400/60" />
                                    {isTr ? '30 saniyede kurulum' : 'Setup in 30 seconds'}
                                </div>
                                <div className="flex items-center gap-1.5">
                                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400/60" />
                                    {isTr ? 'Açık kaynak CLI' : 'Open-source CLI'}
                                </div>
                                <div className="flex items-center gap-1.5">
                                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400/60" />
                                    {isTr ? 'KVKK uyumlu' : 'GDPR compliant'}
                                </div>
                            </div>
                        </div>

                        {/* Hero Visual — Terminal */}
                        <div className="relative">
                            <div className="absolute -inset-6 bg-gradient-to-r from-emerald-500/10 to-cyan-500/10 rounded-3xl blur-2xl opacity-40 animate-pulse-slow" />
                            <div className="relative rounded-2xl border border-white/[0.08] bg-[#0c0e14]/80 backdrop-blur-xl shadow-elevated overflow-hidden">
                                <div className="px-5 py-4 border-b border-white/[0.06] flex items-center justify-between">
                                    <div className="flex items-center gap-2">
                                        <div className="flex gap-1.5">
                                            <div className="w-2.5 h-2.5 rounded-full bg-red-400/40" />
                                            <div className="w-2.5 h-2.5 rounded-full bg-yellow-400/40" />
                                            <div className="w-2.5 h-2.5 rounded-full bg-emerald-400/40" />
                                        </div>
                                        <span className="text-xs font-medium text-white/55 ml-2">terminal</span>
                                    </div>
                                    <div className="flex items-center gap-1.5">
                                        <Activity className="w-3 h-3 text-emerald-400/50 animate-pulse" />
                                        <span className="text-[11px] text-emerald-400/50 font-medium">Live</span>
                                    </div>
                                </div>
                                <div className="p-5 space-y-4">
                                    <div className="font-mono text-xs text-white/75 space-y-2.5">
                                        <div><span className="text-emerald-400/60">$</span> gorenel start --port 3000</div>
                                        <div className="text-white/60 pl-3">
                                            <span className="text-emerald-400/50">✓</span> {isTr ? 'Tünel oluşturuldu' : 'Tunnel established'}
                                            <span className="text-white/45 ml-2">12ms</span>
                                        </div>
                                        <div className="text-white/60 pl-3">
                                            <span className="text-emerald-400/50">✓</span> {isTr ? 'SSL sertifikası hazır' : 'SSL certificate ready'}
                                            <span className="text-white/45 ml-2">auto</span>
                                        </div>
                                        <div className="text-white/60 pl-3">
                                            <span className="text-emerald-400/50">✓</span> {isTr ? 'AI anomali tespiti aktif' : 'AI anomaly detection active'}
                                            <span className="text-white/45 ml-2">2 models</span>
                                        </div>
                                        <div className="text-white/60 pl-3">
                                            <span className="text-emerald-400/50">✓</span> {isTr ? 'Rate limiter ayarlandı' : 'Rate limiter configured'}
                                            <span className="text-white/45 ml-2">100 req/s</span>
                                        </div>
                                        <div className="mt-3 pt-3 border-t border-white/[0.06]">
                                            <span className="text-white/55">{isTr ? 'Canlı adresiniz:' : 'Your live URL:'}</span>
                                        </div>
                                        <div className="text-emerald-400 font-semibold text-sm">
                                            https://my-app.gorenel.site
                                        </div>
                                    </div>

                                    {/* Live traffic simulation */}
                                    <div className="rounded-xl border border-emerald-500/15 bg-emerald-500/[0.04] p-4 mt-3">
                                        <div className="text-[10px] font-medium text-emerald-300/50 mb-2.5 flex items-center gap-1.5">
                                            <Shield className="w-3 h-3" />
                                            {isTr ? 'Güvenlik Katmanları — Aktif' : 'Security Layers — Active'}
                                        </div>
                                        <div className="grid grid-cols-3 gap-2">
                                            {[
                                                { label: 'KeyAuth', active: true },
                                                { label: 'Rate Limit', active: true },
                                                { label: 'AI Monitor', active: true },
                                            ].map((p) => (
                                                <div key={p.label} className="flex items-center gap-1.5 text-[11px] text-emerald-100/50">
                                                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400/70 shadow-[0_0_6px_rgba(16,185,129,0.5)]" />
                                                    {p.label}
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            </div>

                            {/* Feature badges below terminal */}
                            <div className="mt-4 rounded-xl border border-white/[0.06] bg-white/[0.02] p-4">
                                <div className="flex items-center gap-2 text-sm font-medium text-white/70 mb-3">
                                    <Command className="w-4 h-4 text-white/40" />
                                    {isTr ? 'Tek komutla hepsi dahil' : 'Everything included in one command'}
                                </div>
                                <div className="grid grid-cols-2 gap-2 text-xs text-white/40">
                                    {(isTr
                                        ? ['Sabit URL\'ler', 'Trafik izleyici', 'ML anomali tespiti', 'Tünel politikaları', 'Gerçek zamanlı metrikler', 'Otomatik SSL']
                                        : ['Reserved URLs', 'Traffic inspector', 'ML anomaly detection', 'Tunnel policies', 'Real-time metrics', 'Auto SSL']
                                    ).map((item) => (
                                        <div key={item} className="inline-flex items-center gap-2"><CheckCircle2 className="w-3.5 h-3.5 text-emerald-400/70 shrink-0" /> {item}</div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* ═══════════════════════  STATS / SOCIAL PROOF  ═══════════════════════ */}
                <section className="border-t border-b border-white/[0.04] bg-white/[0.01]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-12">
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
                            {[
                                { value: 50, suffix: 'ms', prefix: '<', label: isTr ? 'Ortalama Gecikme' : 'Avg Latency', icon: Clock },
                                { value: 256, suffix: '-bit', label: isTr ? 'Uçtan Uca Şifreleme' : 'End-to-End Encryption', icon: Lock },
                                { value: 99.9, suffix: '%', label: isTr ? 'Uptime SLA' : 'Uptime SLA', icon: Server },
                                { value: 3, suffix: 's', label: isTr ? 'Kurulum Süresi' : 'Time to Deploy', icon: Gauge },
                            ].map((stat) => (
                                <div key={stat.label} className="space-y-2 group">
                                    <stat.icon className="w-5 h-5 text-emerald-400/50 mx-auto group-hover:text-emerald-400/80 transition-colors" />
                                    <div className="text-2xl md:text-3xl font-bold tracking-tight text-white">
                                        {typeof stat.value === 'number' && stat.value > 10
                                            ? <AnimatedNumber target={stat.value} prefix={stat.prefix} suffix={stat.suffix} />
                                            : <>{stat.prefix}{stat.value}{stat.suffix}</>
                                        }
                                    </div>
                                    <div className="text-xs text-white/35">{stat.label}</div>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>

                {/* ═══════════════════════  HOW IT WORKS  ═══════════════════════ */}
                <section id="how-it-works" className="max-w-6xl mx-auto px-6 md:px-10 py-24 space-y-14">
                    <div className="text-center space-y-3 max-w-2xl mx-auto">
                        <div className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400/60 mb-2">
                            <Zap className="w-3.5 h-3.5" />
                            {isTr ? 'HIZLI BAŞLANGIÇ' : 'QUICK START'}
                        </div>
                        <h2 className="text-2xl md:text-4xl font-bold tracking-tight">
                            {isTr ? '3 adımda canlıya geçin' : 'Go live in 3 steps'}
                        </h2>
                        <p className="text-sm md:text-base text-white/40 leading-relaxed max-w-lg mx-auto">
                            {isTr
                                ? 'Karmaşık konfigürasyonlara son. Bir komut, bir URL, sonsuz güç.'
                                : 'No complex configurations. One command, one URL, unlimited power.'
                            }
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        {[
                            {
                                step: '01',
                                icon: TerminalSquare,
                                title: isTr ? 'CLI\'yı kurun' : 'Install the CLI',
                                desc: isTr
                                    ? 'Tek satır komutla Windows, macOS ve Linux için kurun. 10 saniye sürer.'
                                    : 'One-line install for Windows, macOS and Linux. Takes 10 seconds.',
                                code: 'curl -sSL gorenel.site/install.sh | bash',
                                color: 'emerald',
                            },
                            {
                                step: '02',
                                icon: Lock,
                                title: isTr ? 'API key alın' : 'Grab your API key',
                                desc: isTr
                                    ? 'Dashboard\'dan tek tıkla API anahtarınızı oluşturun ve CLI\'ya bağlayın.'
                                    : 'Generate your API key from the dashboard with one click and link it.',
                                code: 'gorenel config set api_key gk_*****',
                                color: 'blue',
                            },
                            {
                                step: '03',
                                icon: Globe,
                                title: isTr ? 'Yayına geçin' : 'Go live',
                                desc: isTr
                                    ? 'Localhost\'unuz artık HTTPS ile dünyaya açık. AI anomali tespiti otomatik aktif.'
                                    : 'Your localhost is now live with HTTPS. AI anomaly detection activates automatically.',
                                code: 'gorenel start --port 3000',
                                color: 'violet',
                            },
                        ].map((item) => (
                            <div key={item.step} className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-7 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-300 group relative overflow-hidden">
                                <div className="absolute top-4 right-5 text-5xl font-black text-white/[0.03] select-none group-hover:text-white/[0.06] transition-colors">{item.step}</div>
                                <div className={`w-10 h-10 rounded-xl bg-${item.color}-500/10 border border-${item.color}-500/20 flex items-center justify-center mb-5 group-hover:border-${item.color}-500/30 transition-colors`}>
                                    <item.icon className={`w-5 h-5 text-${item.color}-400/70`} />
                                </div>
                                <h3 className="text-base font-semibold text-white mb-2">{item.title}</h3>
                                <p className="text-sm text-white/40 leading-relaxed mb-4">{item.desc}</p>
                                <div className="rounded-lg bg-black/30 border border-white/[0.06] px-3.5 py-2.5 font-mono text-xs text-white/40">
                                    <span className="text-emerald-400/50">$</span> {item.code}
                                </div>
                            </div>
                        ))}
                    </div>
                </section>

                {/* ═══════════════════════  FEATURES  ═══════════════════════ */}
                <section id="features" className="border-t border-white/[0.04]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-24 space-y-14">
                        <div className="text-center space-y-3 max-w-2xl mx-auto">
                            <div className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400/60 mb-2">
                                <Layers className="w-3.5 h-3.5" />
                                {isTr ? 'ÖZELLİKLER' : 'FEATURES'}
                            </div>
                            <h2 className="text-2xl md:text-4xl font-bold tracking-tight">
                                {isTr ? 'Sadece bir tünel değil, tam bir platform' : 'Not just a tunnel — a complete platform'}
                            </h2>
                            <p className="text-sm md:text-base text-white/40 leading-relaxed max-w-lg mx-auto">
                                {isTr
                                    ? 'Güvenlik, gözlemlenebilirlik ve yapay zeka — hepsi kutudan çıktığı gibi. Ekstra kurulum yok.'
                                    : 'Security, observability and AI — all out of the box. Zero extra configuration.'
                                }
                            </p>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
                            {[
                                {
                                    icon: Shield,
                                    title: isTr ? 'Uç Nokta Güvenliği' : 'Edge Security',
                                    desc: isTr
                                        ? 'KeyAuth, BasicAuth, IP allowlist ve rate limiter ile her tüneli ayrı ayrı koruyun. Sıfır güven mimarisi.'
                                        : 'Protect each tunnel with KeyAuth, BasicAuth, IP allowlists and rate limiting. Zero-trust architecture.',
                                    color: 'text-emerald-400/70',
                                    badge: isTr ? 'Sıfır Güven' : 'Zero Trust',
                                },
                                {
                                    icon: Eye,
                                    title: isTr ? 'Trafik İzleyici' : 'Traffic Inspector',
                                    desc: isTr
                                        ? 'Tüm HTTP isteklerini gerçek zamanlı izleyin, tek tıkla yeniden çalıştırın ve paylaşılabilir trace oluşturun.'
                                        : 'Monitor all HTTP requests in real-time, replay with one click and create shareable traces.',
                                    color: 'text-blue-400/70',
                                    badge: isTr ? 'Gerçek Zamanlı' : 'Real-Time',
                                },
                                {
                                    icon: Brain,
                                    title: isTr ? 'AI Anomali Tespiti' : 'AI Anomaly Detection',
                                    desc: isTr
                                        ? 'Isolation Forest ve Autoencoder ile trafiğinizdeki anormal davranışları 1ms\'de tespit edin.'
                                        : 'Detect abnormal traffic patterns in 1ms with dual Isolation Forest + Autoencoder ML models.',
                                    color: 'text-violet-400/70',
                                    badge: isTr ? 'Çift Model' : 'Dual Model',
                                },
                                {
                                    icon: Globe,
                                    title: isTr ? 'Sabit URL\'ler' : 'Reserved URLs',
                                    desc: isTr
                                        ? 'Her proje için kalıcı subdomain rezerve edin. Yeniden bağlandığınızda aynı URL\'yi koruyun.'
                                        : 'Reserve permanent subdomains per project. Keep the same URL when you reconnect.',
                                    color: 'text-cyan-400/70',
                                    badge: isTr ? 'Kalıcı' : 'Persistent',
                                },
                                {
                                    icon: BarChart3,
                                    title: isTr ? 'Gerçek Zamanlı Metrikler' : 'Real-Time Metrics',
                                    desc: isTr
                                        ? 'Canlı dashboard ile istek/saniye, gecikme, bant genişliği ve coğrafi dağılımı takip edin.'
                                        : 'Track requests/sec, latency, bandwidth and geo distribution with a live dashboard.',
                                    color: 'text-amber-400/70',
                                    badge: 'Dashboard',
                                },
                                {
                                    icon: Code2,
                                    title: isTr ? 'CLI-First Deneyim' : 'CLI-First Experience',
                                    desc: isTr
                                        ? 'Servis modu, otomatik yeniden bağlanma ve cross-platform desteği. Windows, macOS, Linux.'
                                        : 'Service mode, auto-reconnect and cross-platform support. Windows, macOS, Linux.',
                                    color: 'text-rose-400/70',
                                    badge: isTr ? 'Çapraz Platform' : 'Cross-Platform',
                                },
                            ].map((c) => (
                                <div key={c.title} className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-7 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-300 group relative overflow-hidden">
                                    <div className="flex items-center justify-between mb-4">
                                        <div className="w-10 h-10 rounded-xl bg-white/[0.04] border border-white/[0.06] flex items-center justify-center group-hover:border-white/[0.1] transition-colors">
                                            <c.icon className={`w-5 h-5 ${c.color}`} />
                                        </div>
                                        {c.badge && (
                                            <span className="text-[10px] font-medium px-2 py-0.5 rounded-full bg-white/[0.04] border border-white/[0.06] text-white/30">
                                                {c.badge}
                                            </span>
                                        )}
                                    </div>
                                    <h3 className="text-base font-semibold text-white mb-2">{c.title}</h3>
                                    <p className="text-sm text-white/40 leading-relaxed">{c.desc}</p>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>

                {/* ═══════════════════════  AI SECTION (Differentiator)  ═══════════════════════ */}
                <section className="border-t border-white/[0.04] bg-gradient-to-b from-transparent to-emerald-500/[0.02]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-24">
                        <div className="grid grid-cols-1 lg:grid-cols-2 gap-14 items-center">
                            <div className="space-y-8">
                                <div className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400/60">
                                    <Brain className="w-3.5 h-3.5" />
                                    {isTr ? 'YAPAY ZEKA MOTORU' : 'AI ENGINE'}
                                </div>
                                <h2 className="text-2xl md:text-4xl font-bold tracking-tight leading-tight">
                                    {isTr
                                        ? <>Trafiğinizi <span className="text-emerald-400">yapay zeka</span> ile koruyun</>
                                        : <>Protect your traffic with <span className="text-emerald-400">artificial intelligence</span></>
                                    }
                                </h2>
                                <p className="text-base text-white/40 leading-relaxed max-w-lg">
                                    {isTr
                                        ? 'Gorenel, her isteği iki farklı ML modeli ile analiz eder. Isolation Forest nokta anomalileri yakalar, Autoencoder karmaşık desenleri tespit eder. Hiçbir rakip bu seviyede yapay zeka sunmuyor.'
                                        : 'Gorenel analyzes every request with two distinct ML models. Isolation Forest catches point anomalies, Autoencoder detects complex patterns. No competitor offers this level of AI protection.'
                                    }
                                </p>
                                <div className="space-y-3">
                                    {(isTr
                                        ? [
                                            { label: '~1ms çıkarım süresi', desc: 'Gerçek zamanlı koruma, sıfır gecikme' },
                                            { label: 'Çift model konsensüs', desc: 'İki model aynı anda çalışır, maksimum kapsama' },
                                            { label: 'Otomatik model eğitimi', desc: 'Trafiğinize göre sürekli öğrenir ve adapte olur' },
                                        ]
                                        : [
                                            { label: '~1ms inference time', desc: 'Real-time protection, zero added latency' },
                                            { label: 'Dual model consensus', desc: 'Two models run in parallel for max coverage' },
                                            { label: 'Auto model training', desc: 'Continuously learns and adapts to your traffic' },
                                        ]
                                    ).map((item) => (
                                        <div key={item.label} className="flex items-start gap-3">
                                            <div className="mt-0.5 w-5 h-5 rounded-md bg-emerald-500/15 flex items-center justify-center shrink-0">
                                                <Check className="w-3 h-3 text-emerald-400" />
                                            </div>
                                            <div>
                                                <div className="text-sm font-medium text-white/80">{item.label}</div>
                                                <div className="text-xs text-white/35">{item.desc}</div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>

                            {/* AI Visual */}
                            <div className="relative">
                                <div className="absolute -inset-4 bg-gradient-to-br from-emerald-500/10 to-violet-500/5 rounded-3xl blur-2xl opacity-50" />
                                <div className="relative rounded-2xl border border-white/[0.08] bg-[#0c0e14]/80 backdrop-blur-xl p-6 space-y-4">
                                    <div className="flex items-center justify-between">
                                        <h4 className="text-sm font-semibold text-white/80 flex items-center gap-2">
                                            <Cpu className="w-4 h-4 text-emerald-400/60" />
                                            {isTr ? 'Model Karşılaştırması' : 'Model Comparison'}
                                        </h4>
                                        <span className="text-[10px] font-medium px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400/70 border border-emerald-500/20">
                                            {isTr ? 'Canlı' : 'Live'}
                                        </span>
                                    </div>
                                    <div className="space-y-3">
                                        {[
                                            { name: 'Isolation Forest', speed: '~1ms', type: isTr ? 'Ağaç tabanlı' : 'Tree-based', accuracy: '96.2%', color: 'emerald' },
                                            { name: 'Autoencoder', speed: '~5ms', type: isTr ? 'Sinir ağı' : 'Neural network', accuracy: '98.7%', color: 'violet' },
                                        ].map((m) => (
                                            <div key={m.name} className="rounded-xl bg-white/[0.03] border border-white/[0.06] p-4">
                                                <div className="flex items-center justify-between mb-2">
                                                    <span className="text-sm font-medium text-white/80">{m.name}</span>
                                                    <span className={`text-[10px] font-medium text-${m.color}-400/70`}>{m.accuracy}</span>
                                                </div>
                                                <div className="flex items-center gap-4 text-[11px] text-white/35">
                                                    <span>{m.type}</span>
                                                    <span className="w-1 h-1 rounded-full bg-white/10" />
                                                    <span>{m.speed} {isTr ? 'çıkarım' : 'inference'}</span>
                                                </div>
                                                <div className="mt-2 h-1 rounded-full bg-white/[0.06] overflow-hidden">
                                                    <div className={`h-full bg-${m.color}-500/50 rounded-full`} style={{ width: m.accuracy }} />
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                    <div className="pt-3 border-t border-white/[0.06] flex items-center justify-between text-xs text-white/30">
                                        <span className="flex items-center gap-1.5">
                                            <Activity className="w-3 h-3 text-emerald-400/50" />
                                            {isTr ? 'Konsensüs motoru aktif' : 'Consensus engine active'}
                                        </span>
                                        <span>{isTr ? 'Son 24 saat' : 'Last 24 hours'}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* ═══════════════════════  COMPETITOR COMPARISON  ═══════════════════════ */}
                <section id="comparison" className="border-t border-white/[0.04]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-24 space-y-14">
                        <div className="text-center space-y-3 max-w-2xl mx-auto">
                            <div className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400/60 mb-2">
                                <Network className="w-3.5 h-3.5" />
                                {isTr ? 'KARŞILAŞTIRMA' : 'COMPARISON'}
                            </div>
                            <h2 className="text-2xl md:text-4xl font-bold tracking-tight">
                                {isTr ? 'Neden Gorenel?' : 'Why Gorenel?'}
                            </h2>
                            <p className="text-sm md:text-base text-white/40 leading-relaxed max-w-lg mx-auto">
                                {isTr
                                    ? 'Popüler tünelleme çözümleriyle özellik karşılaştırması.'
                                    : 'Feature comparison with popular tunneling solutions.'
                                }
                            </p>
                        </div>

                        <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] overflow-hidden">
                            <div className="overflow-x-auto">
                                <table className="w-full text-sm">
                                    <thead>
                                        <tr className="border-b border-white/[0.06]">
                                            <th className="text-left p-5 text-white/50 font-medium">{isTr ? 'Özellik' : 'Feature'}</th>
                                            <th className="p-5 text-center">
                                                <div className="flex items-center justify-center gap-2">
                                                    <div className="w-5 h-5 rounded bg-emerald-500/20 flex items-center justify-center">
                                                        <Zap className="w-3 h-3 text-emerald-400" />
                                                    </div>
                                                    <span className="font-bold text-emerald-400">Gorenel</span>
                                                </div>
                                            </th>
                                            <th className="p-5 text-center text-white/40 font-medium">ngrok</th>
                                            <th className="p-5 text-center text-white/40 font-medium">Cloudflare Tunnel</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-white/[0.04]">
                                        {[
                                            { feature: isTr ? 'AI Anomali Tespiti' : 'AI Anomaly Detection', gorenel: true, ngrok: false, cf: false },
                                            { feature: isTr ? 'Çift ML Modeli' : 'Dual ML Models', gorenel: true, ngrok: false, cf: false },
                                            { feature: isTr ? 'Sabit URL\'ler' : 'Reserved URLs', gorenel: true, ngrok: true, cf: true },
                                            { feature: isTr ? 'Trafik İzleyici' : 'Traffic Inspector', gorenel: true, ngrok: true, cf: false },
                                            { feature: isTr ? 'Gerçek Zamanlı Metrikler' : 'Real-Time Metrics', gorenel: true, ngrok: true, cf: true },
                                            { feature: isTr ? 'Açık Kaynak CLI' : 'Open-Source CLI', gorenel: true, ngrok: false, cf: false },
                                            { feature: isTr ? 'Kendi Sunucunuzda Çalıştırma' : 'Self-Hosted Option', gorenel: true, ngrok: false, cf: false },
                                            { feature: isTr ? 'Tünel Başına Politika' : 'Per-Tunnel Policies', gorenel: true, ngrok: false, cf: true },
                                            { feature: isTr ? 'Ücretsiz Plan' : 'Free Plan', gorenel: true, ngrok: true, cf: true },
                                        ].map((row) => (
                                            <tr key={row.feature} className="hover:bg-white/[0.02] transition-colors">
                                                <td className="p-4 pl-5 text-white/60 font-medium">{row.feature}</td>
                                                <td className="p-4 text-center">
                                                    {row.gorenel
                                                        ? <Check className="w-5 h-5 text-emerald-400 mx-auto" />
                                                        : <XIcon className="w-4 h-4 text-white/15 mx-auto" />
                                                    }
                                                </td>
                                                <td className="p-4 text-center">
                                                    {row.ngrok
                                                        ? <Check className="w-5 h-5 text-white/30 mx-auto" />
                                                        : <XIcon className="w-4 h-4 text-white/15 mx-auto" />
                                                    }
                                                </td>
                                                <td className="p-4 text-center">
                                                    {row.cf
                                                        ? <Check className="w-5 h-5 text-white/30 mx-auto" />
                                                        : <XIcon className="w-4 h-4 text-white/15 mx-auto" />
                                                    }
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                </section>

                {/* ═══════════════════════  USE CASES  ═══════════════════════ */}
                <section className="border-t border-white/[0.04] bg-white/[0.01]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-24 space-y-14">
                        <div className="text-center space-y-3 max-w-2xl mx-auto">
                            <div className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400/60 mb-2">
                                <Monitor className="w-3.5 h-3.5" />
                                {isTr ? 'KULLANIM ALANLARI' : 'USE CASES'}
                            </div>
                            <h2 className="text-2xl md:text-4xl font-bold tracking-tight">
                                {isTr ? 'Her senaryo için tasarlandı' : 'Built for every scenario'}
                            </h2>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                            {[
                                {
                                    icon: GitBranch,
                                    title: isTr ? 'Webhook Geliştirme' : 'Webhook Development',
                                    desc: isTr
                                        ? 'Stripe, GitHub ve 3. parti webhook\'ları doğrudan localhost\'unuza yönlendirin. Deploy etmeden gerçek verilerle test edin.'
                                        : 'Route Stripe, GitHub and third-party webhooks directly to localhost. Test with real data without deploying.',
                                    tag: isTr ? 'Geliştiriciler' : 'Developers',
                                },
                                {
                                    icon: Sparkles,
                                    title: isTr ? 'Müşteri Demoları' : 'Client Demos',
                                    desc: isTr
                                        ? 'Sabit URL ile müşterinize canlı demo gösterin. Her seferinde aynı link, profesyonel görünüm.'
                                        : 'Show live demos with a stable URL. Same link every time, professional appearance.',
                                    tag: isTr ? 'Satış Ekipleri' : 'Sales Teams',
                                },
                                {
                                    icon: Server,
                                    title: isTr ? 'IoT & Uzak Erişim' : 'IoT & Remote Access',
                                    desc: isTr
                                        ? 'Ev sunucunuz, IoT cihazlarınız veya Raspberry Pi\'nize dünyanın her yerinden güvenli erişim sağlayın.'
                                        : 'Secure access to your home server, IoT devices or Raspberry Pi from anywhere in the world.',
                                    tag: 'IoT',
                                },
                                {
                                    icon: Shield,
                                    title: isTr ? 'Güvenlik Testi' : 'Security Testing',
                                    desc: isTr
                                        ? 'AI anomali tespiti ve trafik izleyici ile uygulamanızın güvenliğini sürekli izleyin ve denetleyin.'
                                        : 'Continuously monitor and audit your app\'s security with AI anomaly detection and traffic inspector.',
                                    tag: isTr ? 'DevSecOps' : 'DevSecOps',
                                },
                            ].map((uc) => (
                                <div key={uc.title} className="flex gap-5 rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-300 group">
                                    <div className="w-10 h-10 rounded-xl bg-emerald-500/[0.08] border border-emerald-500/20 flex items-center justify-center shrink-0 group-hover:border-emerald-500/30 transition-colors">
                                        <uc.icon className="w-5 h-5 text-emerald-400/70" />
                                    </div>
                                    <div className="flex-1">
                                        <div className="flex items-center gap-2 mb-1.5">
                                            <h3 className="text-base font-semibold text-white">{uc.title}</h3>
                                            <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-white/[0.04] text-white/25">{uc.tag}</span>
                                        </div>
                                        <p className="text-sm text-white/40 leading-relaxed">{uc.desc}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>

                {/* ═══════════════════════  FREE & OPEN SOURCE  ═══════════════════════ */}
                <section id="pricing" className="border-t border-white/[0.04]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-24 space-y-14">
                        <div className="text-center space-y-3 max-w-2xl mx-auto">
                            <div className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400/60 mb-2">
                                <Sparkles className="w-3.5 h-3.5" />
                                {isTr ? 'TAMAMEN ÜCRETSİZ' : 'COMPLETELY FREE'}
                            </div>
                            <h2 className="text-2xl md:text-4xl font-bold tracking-tight">
                                {isTr ? 'Her şey ücretsiz. Gerçekten.' : 'Everything is free. Really.'}
                            </h2>
                            <p className="text-sm md:text-base text-white/40 leading-relaxed max-w-lg mx-auto">
                                {isTr
                                    ? 'Gizli ücret yok, premium duvarı yok. Tüm özellikler — AI anomali tespiti, trafik politikaları, sabit URL\'ler — herkese açık.'
                                    : 'No hidden fees, no paywall. Every feature — AI anomaly detection, traffic policies, reserved URLs — is available to everyone.'
                                }
                            </p>
                        </div>

                        <div className="max-w-3xl mx-auto">
                            <div className="rounded-2xl border-2 border-emerald-500/30 bg-emerald-500/[0.04] p-8 md:p-10 relative shadow-[0_0_40px_-12px_rgba(16,185,129,0.2)]">
                                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                                    <span className="px-3 py-1 text-[11px] font-bold bg-emerald-500 text-[#080a10] rounded-full">
                                        {isTr ? '100% ÜCRETSİZ' : '100% FREE'}
                                    </span>
                                </div>
                                <div className="text-center mb-8">
                                    <div className="mb-3">
                                        <span className="text-5xl md:text-6xl font-bold text-white">$0</span>
                                        <span className="text-lg text-white/30 ml-2">{isTr ? 'sonsuza dek' : 'forever'}</span>
                                    </div>
                                    <p className="text-sm text-white/40">
                                        {isTr
                                            ? 'Açık kaynak proje. Tüm özellikler dahil, sınırsız kullanım.'
                                            : 'Open-source project. All features included, unlimited usage.'
                                        }
                                    </p>
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 mb-8">
                                    {(isTr
                                        ? [
                                            'Sınırsız tünel',
                                            'Sabit subdomain\'ler',
                                            'AI anomali tespiti',
                                            'Trafik politikaları',
                                            'Trafik izleyici',
                                            'Gerçek zamanlı metrikler',
                                            'Çift ML modeli',
                                            'Otomatik SSL',
                                            'Self-hosted seçeneği',
                                            'Rate limiting',
                                            'GeoLocation',
                                            'Açık kaynak CLI',
                                        ]
                                        : [
                                            'Unlimited tunnels',
                                            'Reserved subdomains',
                                            'AI anomaly detection',
                                            'Traffic policies',
                                            'Traffic inspector',
                                            'Real-time metrics',
                                            'Dual ML models',
                                            'Auto SSL',
                                            'Self-hosted option',
                                            'Rate limiting',
                                            'GeoLocation',
                                            'Open-source CLI',
                                        ]
                                    ).map((f) => (
                                        <div key={f} className="flex items-center gap-2.5 text-sm text-white/60">
                                            <Check className="w-4 h-4 text-emerald-400 shrink-0" />
                                            {f}
                                        </div>
                                    ))}
                                </div>
                                <div className="flex justify-center">
                                    <Button variant="primary" size="lg" type="button" onClick={handleCTA} className="px-10">
                                        {isLoggedIn
                                            ? (isTr ? 'Dashboard\'a Git' : 'Open Dashboard')
                                            : (isTr ? 'Hemen Başla — Ücretsiz' : 'Get Started — It\'s Free')
                                        }
                                        <ArrowRight size={16} />
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* ═══════════════════════  ARCHITECTURE  ═══════════════════════ */}
                <section className="border-t border-white/[0.04] bg-white/[0.01]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-24">
                        <div className="text-center space-y-3 max-w-2xl mx-auto mb-14">
                            <div className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400/60 mb-2">
                                <Layers className="w-3.5 h-3.5" />
                                {isTr ? 'MİMARİ' : 'ARCHITECTURE'}
                            </div>
                            <h2 className="text-2xl md:text-4xl font-bold tracking-tight">
                                {isTr ? 'Prodüksiyon düzeyinde altyapı' : 'Production-grade infrastructure'}
                            </h2>
                            <p className="text-sm md:text-base text-white/40 leading-relaxed max-w-lg mx-auto">
                                {isTr
                                    ? 'Go backend, Python ML servisi, React dashboard ve PostgreSQL + Redis + ClickHouse veri katmanı.'
                                    : 'Go backend, Python ML service, React dashboard and PostgreSQL + Redis + ClickHouse data layer.'
                                }
                            </p>
                        </div>

                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                            {[
                                { icon: Server, label: 'Go Backend', desc: isTr ? 'Yüksek performanslı proxy' : 'High-performance proxy', color: 'emerald' },
                                { icon: Brain, label: 'Python ML', desc: isTr ? 'Çift model AI motoru' : 'Dual-model AI engine', color: 'violet' },
                                { icon: Monitor, label: 'React Dashboard', desc: isTr ? 'Gerçek zamanlı panel' : 'Real-time dashboard', color: 'blue' },
                                { icon: Layers, label: 'Data Layer', desc: 'PostgreSQL · Redis · ClickHouse', color: 'cyan' },
                            ].map((item) => (
                                <div key={item.label} className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-5 text-center hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-300 group">
                                    <div className={`w-10 h-10 rounded-xl bg-${item.color}-500/10 border border-${item.color}-500/20 flex items-center justify-center mx-auto mb-3 group-hover:border-${item.color}-500/30 transition-colors`}>
                                        <item.icon className={`w-5 h-5 text-${item.color}-400/70`} />
                                    </div>
                                    <h4 className="text-sm font-semibold text-white/80 mb-0.5">{item.label}</h4>
                                    <p className="text-[11px] text-white/30">{item.desc}</p>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>

                {/* ═══════════════════════  FINAL CTA  ═══════════════════════ */}
                <section className="border-t border-white/[0.04]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-24">
                        <div className="relative rounded-3xl border border-white/[0.08] bg-gradient-to-br from-emerald-500/[0.08] to-cyan-500/[0.04] overflow-hidden">
                            <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,rgba(16,185,129,0.12),transparent_60%)]" />
                            <div className="absolute top-0 right-0 w-96 h-96 bg-emerald-500/5 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
                            <div className="relative px-8 md:px-14 py-14 md:py-20 text-center space-y-6">
                                <div className="inline-flex items-center gap-2 rounded-full border border-emerald-500/20 bg-emerald-500/[0.08] px-3 py-1 text-xs font-medium text-emerald-300/70 mb-2">
                                    <Zap className="w-3.5 h-3.5" />
                                    {isTr ? '30 saniyede başlayın' : 'Start in 30 seconds'}
                                </div>
                                <h2 className="text-3xl md:text-5xl font-bold tracking-tight leading-tight">
                                    {isTr
                                        ? <>Localhost'unuz <span className="text-emerald-400">dünyayı</span> bekliyor</>
                                        : <>Your localhost is ready <br className="hidden md:block" />for the <span className="text-emerald-400">world</span></>
                                    }
                                </h2>
                                <p className="text-base md:text-lg text-white/65 max-w-xl mx-auto leading-relaxed">
                                    {isTr
                                        ? 'Ücretsiz başlayın. Kredi kartı gerekmez. İlk tünelinizi 30 saniyede oluşturun ve AI korumasının farkını yaşayın.'
                                        : 'Start free. No credit card required. Create your first tunnel in 30 seconds and experience AI-powered protection.'
                                    }
                                </p>
                                <div className="flex flex-col sm:flex-row gap-3 justify-center pt-4">
                                    <Button variant="light" size="lg" type="button" onClick={handleCTA}>
                                        {isTr ? 'Hemen Başla' : 'Get Started Now'}
                                        <ArrowRight size={16} />
                                    </Button>
                                    <Button variant="outline" size="lg" type="button" onClick={onLogin}>
                                        {isTr ? 'Canlı Demo' : 'View Live Demo'}
                                    </Button>
                                </div>
                                <p className="text-xs text-white/50 pt-2">
                                    {isTr
                                        ? '✦ Tüm özellikler ücretsiz — açık kaynak proje'
                                        : '✦ All features free — open-source project'
                                    }
                                </p>
                            </div>
                        </div>
                    </div>
                </section>

                {/* ═══════════════════════  FOOTER  ═══════════════════════ */}
                <footer className="border-t border-white/[0.04]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-10">
                        <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-10">
                            <div className="space-y-3">
                                <div className="flex items-center gap-2.5">
                                    <div className="w-6 h-6 rounded-md bg-white/[0.05] border border-white/[0.08] overflow-hidden">
                                        <img src="/logo.png" alt="Gorenel" width="256" height="256" className="w-full h-full object-cover" />
                                    </div>
                                    <span className="font-bold text-sm text-white">Gorenel</span>
                                </div>
                                <p className="text-xs text-white/55 leading-relaxed max-w-[200px]">
                                    {isTr
                                        ? 'AI destekli yeni nesil tünelleme platformu.'
                                        : 'AI-powered next-gen tunneling platform.'
                                    }
                                </p>
                            </div>
                            <div className="space-y-3">
                                <h4 className="text-xs font-semibold text-white/65 uppercase tracking-wider">{isTr ? 'Ürün' : 'Product'}</h4>
                                <div className="space-y-2 text-xs text-white/55">
                                    <div className="hover:text-white/75 cursor-pointer transition-colors">{isTr ? 'Özellikler' : 'Features'}</div>
                                    <div className="hover:text-white/75 cursor-pointer transition-colors">{isTr ? 'Neden Ücretsiz?' : 'Why Free?'}</div>
                                    <div className="hover:text-white/75 cursor-pointer transition-colors">{isTr ? 'Değişiklik Günlüğü' : 'Changelog'}</div>
                                </div>
                            </div>
                            <div className="space-y-3">
                                <h4 className="text-xs font-semibold text-white/65 uppercase tracking-wider">{isTr ? 'Geliştirici' : 'Developer'}</h4>
                                <div className="space-y-2 text-xs text-white/55">
                                    <div className="hover:text-white/75 cursor-pointer transition-colors">{isTr ? 'Dokümantasyon' : 'Documentation'}</div>
                                    <div className="hover:text-white/75 cursor-pointer transition-colors">API Reference</div>
                                    <div className="hover:text-white/75 cursor-pointer transition-colors">GitHub</div>
                                </div>
                            </div>
                            <div className="space-y-3">
                                <h4 className="text-xs font-semibold text-white/65 uppercase tracking-wider">{isTr ? 'Şirket' : 'Company'}</h4>
                                <div className="space-y-2 text-xs text-white/55">
                                    <div className="hover:text-white/75 cursor-pointer transition-colors">{isTr ? 'Hakkımızda' : 'About'}</div>
                                    <div className="hover:text-white/75 cursor-pointer transition-colors">{isTr ? 'Gizlilik Politikası' : 'Privacy Policy'}</div>
                                    <div className="hover:text-white/75 cursor-pointer transition-colors">{isTr ? 'Hizmet Koşulları' : 'Terms of Service'}</div>
                                </div>
                            </div>
                        </div>
                        <div className="pt-6 border-t border-white/[0.04] flex flex-col md:flex-row items-center justify-between gap-4 text-xs text-white/50">
                            <span>&copy; 2026 Gorenel. {isTr ? 'Tüm hakları saklıdır.' : 'All rights reserved.'}</span>
                            <div className="flex items-center gap-4">
                                <span className="flex items-center gap-1.5">
                                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400/60 shadow-[0_0_6px_rgba(16,185,129,0.4)]" />
                                    {isTr ? 'Tüm sistemler aktif' : 'All systems operational'}
                                </span>
                                <span className="w-1 h-1 rounded-full bg-white/10" />
                                <span>EU-Central-1</span>
                            </div>
                        </div>
                    </div>
                </footer>
            </main>
        </div>
    );
};
