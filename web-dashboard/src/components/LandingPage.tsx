import React from 'react';
import { ArrowRight, CheckCircle2, Command, Globe, Languages, Lock, Shield, TerminalSquare, Activity, Gauge, Eye, Brain, Server, Sparkles, Clock, Users } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Button } from './ui/Button';

interface LandingPageProps {
    onLogin: () => void;
    isLoggedIn?: boolean;
    onGoToDashboard?: () => void;
}

export const LandingPage: React.FC<LandingPageProps> = ({ onLogin, isLoggedIn, onGoToDashboard }) => {
    const { t, i18n } = useTranslation();
    const isTr = i18n.language === 'tr';

    const toggleLanguage = () => {
        i18n.changeLanguage(isTr ? 'en' : 'tr');
    };

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
                            {isLoggedIn ? 'Dashboard' : (isTr ? 'Ücretsiz Başla' : 'Start free')}
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
                            <div className="inline-flex items-center gap-2 rounded-full border border-emerald-500/20 bg-emerald-500/[0.06] px-3.5 py-1.5 text-xs font-medium text-emerald-300/80">
                                <Sparkles className="w-3.5 h-3.5" />
                                {isTr ? 'Yapay zeka destekli tünel altyapısı' : 'AI-powered tunnel infrastructure'}
                            </div>

                            <h1 className="text-4xl md:text-5xl lg:text-[3.5rem] font-semibold tracking-tight leading-[1.1]">
                                {isTr
                                    ? <>Localhost'unuzu <span className="text-emerald-400">saniyeler</span> içinde dünyaya açın.</>
                                    : <>Ship localhost to the world in <span className="text-emerald-400">seconds</span>.</>
                                }
                            </h1>

                            <p className="text-base md:text-lg text-white/45 leading-relaxed max-w-lg">
                                {isTr
                                    ? 'Güvenli tüneller, sabit URL\'ler, trafik politikaları ve yapay zeka ile anomali tespiti — tek bir CLI komutuyla. Geliştirme, demo ve prodüksiyon için.'
                                    : 'Secure tunnels, stable URLs, traffic policies and AI-powered anomaly detection — with a single CLI command. For dev, demos and production.'
                                }
                            </p>

                            <div className="flex flex-col sm:flex-row gap-3">
                                <Button
                                    variant="primary"
                                    size="lg"
                                    type="button"
                                    onClick={isLoggedIn ? onGoToDashboard : onLogin}
                                >
                                    {isLoggedIn ? (isTr ? 'Dashboard\'a Git' : 'Open Dashboard') : (isTr ? 'Ücretsiz Başlayın' : 'Get started — it\'s free')}
                                    <ArrowRight size={16} />
                                </Button>
                                <Button variant="outline" size="lg" type="button" onClick={onLogin}>
                                    {isTr ? 'Canlı Demo' : 'Live demo'}
                                </Button>
                            </div>

                            {/* Social proof strip */}
                            <div className="flex items-center gap-5 pt-2 text-xs text-white/30">
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
                            </div>
                        </div>

                        {/* Hero Visual - Terminal */}
                        <div className="relative">
                            <div className="absolute -inset-6 bg-gradient-to-r from-emerald-500/10 to-cyan-500/10 rounded-3xl blur-2xl opacity-40" />
                            <div className="relative rounded-2xl border border-white/[0.08] bg-[#0c0e14]/80 backdrop-blur-xl shadow-elevated overflow-hidden">
                                <div className="px-5 py-4 border-b border-white/[0.06] flex items-center justify-between">
                                    <div className="flex items-center gap-2">
                                        <div className="flex gap-1.5">
                                            <div className="w-2.5 h-2.5 rounded-full bg-red-400/40" />
                                            <div className="w-2.5 h-2.5 rounded-full bg-yellow-400/40" />
                                            <div className="w-2.5 h-2.5 rounded-full bg-emerald-400/40" />
                                        </div>
                                        <span className="text-xs font-medium text-white/30 ml-2">terminal</span>
                                    </div>
                                    <div className="flex items-center gap-1.5">
                                        <Activity className="w-3 h-3 text-emerald-400/50" />
                                        <span className="text-[11px] text-emerald-400/50">Live</span>
                                    </div>
                                </div>
                                <div className="p-5 space-y-4">
                                    <div className="font-mono text-xs text-white/55 space-y-2.5">
                                        <div><span className="text-emerald-400/60">$</span> gorenel start --port 3000</div>
                                        <div className="text-white/30 pl-3">
                                            <span className="text-emerald-400/50">✓</span> {isTr ? 'Tünel oluşturuldu' : 'Tunnel established'}
                                        </div>
                                        <div className="text-white/30 pl-3">
                                            <span className="text-emerald-400/50">✓</span> {isTr ? 'SSL sertifikası hazır' : 'SSL certificate ready'}
                                        </div>
                                        <div className="text-white/30 pl-3">
                                            <span className="text-emerald-400/50">✓</span> {isTr ? 'AI anomali tespiti aktif' : 'AI anomaly detection active'}
                                        </div>
                                        <div className="mt-3 pt-3 border-t border-white/[0.06]">
                                            <span className="text-white/25">{isTr ? 'Canlı adresiniz:' : 'Your live URL:'}</span>
                                        </div>
                                        <div className="text-emerald-400 font-semibold text-sm">
                                            https://my-app.gorenel.site
                                        </div>
                                    </div>

                                    <div className="rounded-xl border border-emerald-500/15 bg-emerald-500/[0.04] p-4 mt-3">
                                        <div className="text-[10px] font-medium text-emerald-300/50 mb-2">{isTr ? 'Trafik Politikaları — Aktif' : 'Traffic Policies — Active'}</div>
                                        <div className="grid grid-cols-3 gap-2">
                                            {[
                                                { label: 'KeyAuth', active: true },
                                                { label: 'Rate Limit', active: true },
                                                { label: 'IP Allowlist', active: true },
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

                            <div className="mt-4 rounded-xl border border-white/[0.06] bg-white/[0.02] p-4">
                                <div className="flex items-center gap-2 text-sm font-medium text-white/70 mb-3">
                                    <Command className="w-4 h-4 text-white/40" />
                                    {isTr ? 'Hepsi dahil' : 'Everything included'}
                                </div>
                                <div className="grid grid-cols-2 gap-2 text-xs text-white/40">
                                    {[
                                        isTr ? 'Sabit URL\'ler' : 'Reserved URLs',
                                        isTr ? 'Trafik izleyici' : 'Traffic inspector',
                                        isTr ? 'ML anomali tespiti' : 'ML anomaly detection',
                                        isTr ? 'Tünel başına politika' : 'Per-tunnel policies',
                                    ].map((item) => (
                                        <div key={item} className="inline-flex items-center gap-2"><CheckCircle2 className="w-3.5 h-3.5 text-emerald-400/70" /> {item}</div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Stats strip */}
                <section className="border-t border-b border-white/[0.04] bg-white/[0.01]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-10">
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
                            {[
                                { value: '<50ms', label: isTr ? 'Ortalama Gecikme' : 'Avg Latency', icon: Clock },
                                { value: '256-bit', label: isTr ? 'Uçtan Uca Şifreleme' : 'End-to-End Encryption', icon: Lock },
                                { value: '99.9%', label: isTr ? 'Uptime SLA' : 'Uptime SLA', icon: Server },
                                { value: '3', label: isTr ? 'Saniyede Kurulum' : 'Seconds to Deploy', icon: Gauge },
                            ].map((stat) => (
                                <div key={stat.label} className="space-y-2">
                                    <stat.icon className="w-5 h-5 text-emerald-400/50 mx-auto" />
                                    <div className="text-2xl md:text-3xl font-bold tracking-tight text-white">{stat.value}</div>
                                    <div className="text-xs text-white/35">{stat.label}</div>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>

                {/* How it works */}
                <section className="max-w-6xl mx-auto px-6 md:px-10 py-20 space-y-14">
                    <div className="text-center space-y-3 max-w-2xl mx-auto">
                        <h2 className="text-2xl md:text-3xl font-semibold tracking-tight">
                            {isTr ? '3 adımda canlıya geçin' : 'Go live in 3 steps'}
                        </h2>
                        <p className="text-sm md:text-base text-white/40 leading-relaxed">
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
                                code: isTr ? 'curl -sSL gorenel.site/install.sh | bash' : 'curl -sSL gorenel.site/install.sh | bash',
                                color: 'emerald',
                            },
                            {
                                step: '02',
                                icon: Lock,
                                title: isTr ? 'API key alın' : 'Grab your API key',
                                desc: isTr
                                    ? 'Dashboard\'dan tek tıkla API anahtarınızı oluşturun ve CLI\'ya bağlayın.'
                                    : 'Generate your API key from the dashboard with one click and link it to the CLI.',
                                code: 'gorenel config set api_key gk_*****',
                                color: 'blue',
                            },
                            {
                                step: '03',
                                icon: Globe,
                                title: isTr ? 'Yayına geçin' : 'Go live',
                                desc: isTr
                                    ? 'Localhost\'unuz artık HTTPS ile dünyaya açık. Sabit URL, trafik politikası ve anomali tespiti dahil.'
                                    : 'Your localhost is now live with HTTPS. Reserved URL, traffic policies and anomaly detection included.',
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

                {/* Features Section */}
                <section className="border-t border-white/[0.04]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-20 space-y-14">
                        <div className="text-center space-y-3 max-w-2xl mx-auto">
                            <h2 className="text-2xl md:text-3xl font-semibold tracking-tight">
                                {isTr ? 'Sadece bir tünel değil, tam bir altyapı' : 'Not just a tunnel — a full infrastructure'}
                            </h2>
                            <p className="text-sm md:text-base text-white/40 leading-relaxed">
                                {isTr
                                    ? 'Güvenlik, gözlemlenebilirlik ve yapay zeka — hepsi kutudan çıktığı gibi.'
                                    : 'Security, observability and AI — all out of the box.'
                                }
                            </p>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
                            {[
                                {
                                    icon: Shield,
                                    title: isTr ? 'Uç Nokta Güvenliği' : 'Edge Security',
                                    desc: isTr
                                        ? 'KeyAuth, BasicAuth, IP allowlist ve rate limiter ile her tüneli ayrı ayrı koruyun.'
                                        : 'Protect each tunnel individually with KeyAuth, BasicAuth, IP allowlists and rate limiting.',
                                    color: 'text-emerald-400/70',
                                },
                                {
                                    icon: Eye,
                                    title: isTr ? 'Trafik İzleyici' : 'Traffic Inspector',
                                    desc: isTr
                                        ? 'Tüm HTTP isteklerini gerçek zamanlı izleyin, tek tıkla yeniden çalıştırın ve paylaşılabilir trace oluşturun.'
                                        : 'Monitor all HTTP requests in real-time, replay with one click and create shareable traces.',
                                    color: 'text-blue-400/70',
                                },
                                {
                                    icon: Brain,
                                    title: isTr ? 'AI Anomali Tespiti' : 'AI Anomaly Detection',
                                    desc: isTr
                                        ? 'Isolation Forest ve Autoencoder ile trafiğinizdeki anormal davranışları otomatik tespit edin.'
                                        : 'Automatically detect abnormal behavior in your traffic with Isolation Forest and Autoencoder models.',
                                    color: 'text-violet-400/70',
                                },
                                {
                                    icon: Globe,
                                    title: isTr ? 'Sabit URL\'ler' : 'Reserved URLs',
                                    desc: isTr
                                        ? 'Her cihaz için kalıcı subdomain rezerve edin. Yeniden bağlandığınızda aynı URL\'yi koruyun.'
                                        : 'Reserve permanent subdomains per device. Keep the same URL when you reconnect.',
                                    color: 'text-cyan-400/70',
                                },
                                {
                                    icon: Gauge,
                                    title: isTr ? 'Gerçek Zamanlı Metrikler' : 'Real-Time Metrics',
                                    desc: isTr
                                        ? 'Canlı dashboard ile istek/saniye, gecikme, bant genişliği ve coğrafi dağılımı takip edin.'
                                        : 'Track requests/sec, latency, bandwidth and geo distribution with a live dashboard.',
                                    color: 'text-amber-400/70',
                                },
                                {
                                    icon: TerminalSquare,
                                    title: isTr ? 'CLI-First Deneyim' : 'CLI-First Experience',
                                    desc: isTr
                                        ? 'Servis modu, otomatik yeniden bağlanma ve çapraz platform desteği ile geliştirici dostu.'
                                        : 'Developer-friendly with service mode, auto-reconnect and cross-platform support.',
                                    color: 'text-rose-400/70',
                                },
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

                {/* Use cases */}
                <section className="border-t border-white/[0.04] bg-white/[0.01]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-20 space-y-14">
                        <div className="text-center space-y-3 max-w-2xl mx-auto">
                            <h2 className="text-2xl md:text-3xl font-semibold tracking-tight">
                                {isTr ? 'Her senaryo için kullanın' : 'Built for every scenario'}
                            </h2>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                            {[
                                {
                                    icon: Users,
                                    title: isTr ? 'Webhook Geliştirme' : 'Webhook Development',
                                    desc: isTr
                                        ? 'Stripe, GitHub ve 3. parti webhook\'ları doğrudan localhost\'unuza yönlendirin. Deploy etmeden test edin.'
                                        : 'Route Stripe, GitHub and third-party webhooks directly to your localhost. Test without deploying.',
                                },
                                {
                                    icon: Sparkles,
                                    title: isTr ? 'Müşteri Demoları' : 'Client Demos',
                                    desc: isTr
                                        ? 'Sabit URL ile müşterinize canlı demo gösterin. Her seferinde aynı link, profesyonel görünüm.'
                                        : 'Show live demos with a stable URL. Same link every time, professional appearance.',
                                },
                                {
                                    icon: Server,
                                    title: isTr ? 'Uzak Erişim' : 'Remote Access',
                                    desc: isTr
                                        ? 'Ev sunucunuz, IoT cihazlarınız veya Raspberry Pi\'nize dünyanın her yerinden güvenli erişim.'
                                        : 'Secure access to your home server, IoT devices or Raspberry Pi from anywhere in the world.',
                                },
                                {
                                    icon: Shield,
                                    title: isTr ? 'Güvenlik Testi' : 'Security Testing',
                                    desc: isTr
                                        ? 'AI anomali tespiti ve trafik izleyici ile uygulamanızın güvenliğini sürekli izleyin.'
                                        : 'Continuously monitor your app\'s security with AI anomaly detection and traffic inspector.',
                                },
                            ].map((uc) => (
                                <div key={uc.title} className="flex gap-5 rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-200">
                                    <div className="w-10 h-10 rounded-xl bg-emerald-500/[0.08] border border-emerald-500/20 flex items-center justify-center shrink-0">
                                        <uc.icon className="w-5 h-5 text-emerald-400/70" />
                                    </div>
                                    <div>
                                        <h3 className="text-base font-semibold text-white mb-1.5">{uc.title}</h3>
                                        <p className="text-sm text-white/40 leading-relaxed">{uc.desc}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>

                {/* CTA Section */}
                <section className="border-t border-white/[0.04]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-20">
                        <div className="relative rounded-3xl border border-white/[0.08] bg-gradient-to-br from-emerald-500/[0.08] to-cyan-500/[0.04] overflow-hidden">
                            <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,rgba(16,185,129,0.12),transparent_60%)]" />
                            <div className="relative px-8 md:px-14 py-14 md:py-16 text-center space-y-6">
                                <h2 className="text-3xl md:text-4xl font-semibold tracking-tight">
                                    {isTr
                                        ? <>Localhost'unuz <span className="text-emerald-400">dünyayı</span> bekliyor</>
                                        : <>Your localhost is ready for the <span className="text-emerald-400">world</span></>
                                    }
                                </h2>
                                <p className="text-base text-white/40 max-w-xl mx-auto leading-relaxed">
                                    {isTr
                                        ? 'Ücretsiz başlayın. Kredi kartı gerekmez. 30 saniyede ilk tünelinizi oluşturun.'
                                        : 'Start free. No credit card required. Create your first tunnel in 30 seconds.'
                                    }
                                </p>
                                <div className="flex flex-col sm:flex-row gap-3 justify-center pt-2">
                                    <Button variant="light" size="lg" type="button" onClick={isLoggedIn ? onGoToDashboard : onLogin}>
                                        {isTr ? 'Hemen Başla' : 'Get Started Now'}
                                        <ArrowRight size={16} />
                                    </Button>
                                    <Button variant="outline" size="lg" type="button" onClick={onLogin}>
                                        {isTr ? 'Canlı Demo' : 'View Live Demo'}
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Footer */}
                <footer className="border-t border-white/[0.04]">
                    <div className="max-w-6xl mx-auto px-6 md:px-10 py-8 flex flex-col md:flex-row items-center justify-between gap-4 text-sm text-white/25">
                        <div className="flex items-center gap-3">
                            <div className="w-5 h-5 rounded bg-white/[0.05] border border-white/[0.08] overflow-hidden">
                                <img src="/logo.png" alt="Gorenel" className="w-full h-full object-cover" />
                            </div>
                            <span>&copy; 2026 Gorenel</span>
                        </div>
                        <div className="flex items-center gap-4 text-xs">
                            <span className="flex items-center gap-1.5">
                                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400/60 shadow-[0_0_6px_rgba(16,185,129,0.4)]" />
                                {isTr ? 'Tüm sistemler aktif' : 'All systems operational'}
                            </span>
                            <span className="w-1 h-1 rounded-full bg-white/10" />
                            <span>{isTr ? 'Edge-ready altyapı' : 'Edge-ready infrastructure'}</span>
                        </div>
                    </div>
                </footer>
            </main>
        </div>
    );
};
