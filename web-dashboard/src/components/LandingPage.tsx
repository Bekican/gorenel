import React, { useState, useEffect } from 'react';
import { Shield, BarChart3, ArrowRight, Command, Globe, Zap, CheckCircle2, Play } from 'lucide-react';

interface LandingPageProps {
    onLogin: () => void;
}

export const LandingPage: React.FC<LandingPageProps> = ({ onLogin }) => {
    const [typedText, setTypedText] = useState('');
    const fullText = "gorenel start --port 3000";

    useEffect(() => {
        let i = 0;
        const interval = setInterval(() => {
            setTypedText(fullText.slice(0, i + 1));
            i++;
            if (i > fullText.length) clearInterval(interval);
        }, 100);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="min-h-screen bg-[#020408] text-white selection:bg-emerald-500/30 overflow-hidden font-sans">
            {/* Background Mesh Gradients */}
            <div className="fixed inset-0 z-0 pointer-events-none">
                <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-emerald-500/5 rounded-full blur-[120px] animate-pulse-slow" />
                <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-600/5 rounded-full blur-[120px] animate-pulse-slow delay-1000" />
            </div>

            {/* Navigation */}
            <nav className="fixed top-0 left-0 right-0 z-50 flex items-center justify-between px-6 py-4 md:px-12 backdrop-blur-md border-b border-white/5 bg-[#020408]/50">
                <div className="flex items-center gap-2">
                    <div className="w-8 h-8 bg-gradient-to-br from-emerald-400 to-cyan-500 rounded-lg flex items-center justify-center shadow-[0_0_15px_rgba(52,211,153,0.3)]">
                        <Zap className="w-5 h-5 text-black fill-current" />
                    </div>
                    <span className="font-bold text-xl tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-white to-white/60">Gorenel</span>
                </div>

                <div className="flex items-center gap-6">
                    <a href="#" className="hidden md:block text-sm text-white/60 hover:text-white transition-colors">Documentation</a>
                    <a href="#" className="hidden md:block text-sm text-white/60 hover:text-white transition-colors">Pricing</a>
                    <button
                        onClick={onLogin}
                        className="px-5 py-2 bg-white text-black text-sm font-semibold rounded-full hover:bg-neutral-200 transition-all active:scale-95 shadow-[0_0_20px_rgba(255,255,255,0.1)]"
                    >
                        Dashboard {`->`}
                    </button>
                </div>
            </nav>

            {/* Hero Section */}
            <main className="relative z-10 pt-32 pb-20 px-6 md:px-12 max-w-7xl mx-auto flex flex-col md:flex-row items-center gap-16">

                {/* Left: Content */}
                <div className="flex-1 space-y-8 animate-fade-in-up">
                    <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-emerald-500/20 bg-emerald-500/10 text-emerald-400 text-xs font-medium">
                        <span className="relative flex h-2 w-2">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                        </span>
                        v2.0 is now live
                    </div>

                    <h1 className="text-5xl md:text-7xl font-bold tracking-tight leading-[1.1]">
                        One gateway for <br />
                        <span className="text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 via-cyan-400 to-blue-500 animate-gradient-x">everything.</span>
                    </h1>

                    <p className="text-lg text-white/50 max-w-xl leading-relaxed">
                        Securely expose your localhost to the world. Load balancing, traffic inspection, and AI gateway features included. No config required.
                    </p>

                    <div className="flex items-center gap-4 pt-4">
                        <button
                            onClick={onLogin}
                            className="group px-8 py-3.5 bg-emerald-500 hover:bg-emerald-400 text-black font-bold rounded-xl transition-all shadow-[0_0_40px_-10px_rgba(16,185,129,0.4)] flex items-center gap-2"
                        >
                            Get Started
                            <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                        </button>
                        <button className="px-8 py-3.5 bg-white/5 border border-white/10 hover:bg-white/10 text-white font-medium rounded-xl transition-all flex items-center gap-2">
                            <Play className="w-4 h-4 fill-current" /> Watch Demo
                        </button>
                    </div>

                    <div className="pt-8 flex items-center gap-8 text-white/30 text-sm font-medium">
                        <span className="flex items-center gap-2"><CheckCircle2 className="w-4 h-4 text-emerald-500" /> Free for developers</span>
                        <span className="flex items-center gap-2"><CheckCircle2 className="w-4 h-4 text-emerald-500" /> No credit card</span>
                    </div>
                </div>

                {/* Right: Interactive Terminal */}
                <div className="flex-1 w-full max-w-lg relative animate-fade-in-up delay-100 group perspective-1000">
                    {/* Glow backing */}
                    <div className="absolute -inset-4 bg-gradient-to-r from-emerald-500/20 to-blue-500/20 rounded-3xl blur-2xl opacity-50 group-hover:opacity-70 transition-opacity duration-500" />

                    <div className="relative bg-[#0A0C10] border border-white/10 rounded-2xl shadow-2xl overflow-hidden transform transition-transform duration-500 hover:rotate-y-2 hover:rotate-x-2">
                        {/* Window Controls */}
                        <div className="px-4 py-3 border-b border-white/5 flex items-center justify-between bg-white/5">
                            <div className="flex items-center gap-2">
                                <div className="w-3 h-3 rounded-full bg-red-500/20 border border-red-500/50" />
                                <div className="w-3 h-3 rounded-full bg-yellow-500/20 border border-yellow-500/50" />
                                <div className="w-3 h-3 rounded-full bg-green-500/20 border border-green-500/50" />
                            </div>
                            <div className="text-xs text-white/30 font-mono">bash — 80x24</div>
                        </div>

                        {/* Terminal Content */}
                        <div className="p-6 font-mono text-sm space-y-4 min-h-[300px]">
                            <div className="flex items-center gap-2 text-white/50">
                                <span className="text-emerald-400">➜</span>
                                <span>~</span>
                                <span className="text-white">{typedText}<span className="animate-pulse">_</span></span>
                            </div>

                            {typedText === fullText && (
                                <>
                                    <div className="space-y-1 animate-fade-in-up">
                                        <div className="text-white/70">Tunnel Status: <span className="text-emerald-400 font-bold">Online</span></div>
                                        <div className="text-white/70">Version: <span className="text-white">2.0.0</span></div>
                                        <div className="text-white/70">Region: <span className="text-white">eu-central-1</span></div>
                                        <div className="text-white/70">Latency: <span className="text-white">24ms</span></div>
                                    </div>

                                    <div className="p-3 bg-white/5 border border-white/10 rounded-lg animate-fade-in-up delay-100">
                                        <div className="text-xs text-white/40 uppercase tracking-wider mb-1">Public URL</div>
                                        <div className="text-emerald-400 font-bold hover:underline cursor-pointer">
                                            https://random-name.gorenel.net
                                        </div>
                                    </div>

                                    <div className="pt-2 text-white/30 text-xs animate-fade-in-up delay-200">
                                        Ctrl+C to stop
                                    </div>
                                </>
                            )}
                        </div>
                    </div>
                </div>
            </main>

            {/* Features Bento Grid */}
            <section className="relative z-10 py-20 px-6 md:px-12 border-t border-white/5 bg-white/[0.02]">
                <div className="max-w-7xl mx-auto space-y-12">
                    <div className="text-center space-y-4">
                        <h2 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-br from-white to-white/60">Everything you need to ship.</h2>
                        <p className="text-white/50 max-w-2xl mx-auto">
                            Built for modern engineering teams who care about speed, security, and developer experience.
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        {/* Card 1: Traffic */}
                        <div className="col-span-1 md:col-span-2 bg-[#0A0C10] border border-white/5 rounded-3xl p-8 hover:border-white/10 transition-all group overflow-hidden relative">
                            <div className="absolute top-0 right-0 p-8 opacity-10 group-hover:opacity-20 transition-opacity">
                                <Globe className="w-48 h-48 text-emerald-500" />
                            </div>
                            <div className="relative z-10 space-y-4">
                                <div className="w-12 h-12 bg-white/5 rounded-2xl flex items-center justify-center border border-white/10">
                                    <Command className="w-6 h-6 text-emerald-400" />
                                </div>
                                <h3 className="text-xl font-bold">Global Traffic Routing</h3>
                                <p className="text-white/50 max-w-md">
                                    Instantly route traffic to any local service. Load balance between multiple regions with zero configuration throughout our edge network.
                                </p>
                            </div>
                        </div>

                        {/* Card 2: Security */}
                        <div className="bg-[#0A0C10] border border-white/5 rounded-3xl p-8 hover:border-white/10 transition-all group overflow-hidden relative">
                            <div className="relative z-10 space-y-4">
                                <div className="w-12 h-12 bg-white/5 rounded-2xl flex items-center justify-center border border-white/10">
                                    <Shield className="w-6 h-6 text-blue-400" />
                                </div>
                                <h3 className="text-xl font-bold">Zero Trust Security</h3>
                                <p className="text-white/50">
                                    Built-in authentication (SSO, OAuth) and IP allowlisting. Secure your endpoints before they even reach your server.
                                </p>
                            </div>
                        </div>

                        {/* Card 3: Observability */}
                        <div className="bg-[#0A0C10] border border-white/5 rounded-3xl p-8 hover:border-white/10 transition-all group overflow-hidden relative">
                            <div className="relative z-10 space-y-4">
                                <div className="w-12 h-12 bg-white/5 rounded-2xl flex items-center justify-center border border-white/10">
                                    <BarChart3 className="w-6 h-6 text-purple-400" />
                                </div>
                                <h3 className="text-xl font-bold">Real-time Metrics</h3>
                                <p className="text-white/50">
                                    Monitor latency, error rates, and traffic volume in real-time. Debug requests with built-in traffic inspector.
                                </p>
                            </div>
                        </div>

                        {/* Card 4: AI Gateway */}
                        <div className="col-span-1 md:col-span-2 bg-[#0A0C10] border border-white/5 rounded-3xl p-8 hover:border-white/10 transition-all group overflow-hidden relative">
                            <div className="relative z-10 space-y-4">
                                <div className="w-12 h-12 bg-white/5 rounded-2xl flex items-center justify-center border border-white/10">
                                    <Zap className="w-6 h-6 text-yellow-400" />
                                </div>
                                <h3 className="text-xl font-bold">AI Gateway Included</h3>
                                <p className="text-white/50 max-w-md">
                                    Unified API for OpenAI, Anthropic, and Local LLMs. Cache responses, rate limit users, and track token usage automatically.
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* Footer */}
            <footer className="py-12 px-6 md:px-12 border-t border-white/5 text-center text-white/30 text-sm">
                <p>&copy; 2026 Gorenel Cloud Gateway. Designed for builders.</p>
            </footer>
        </div>
    );
};
