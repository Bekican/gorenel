import React, { useEffect, useState } from 'react';
import { Shield, Clock, ArrowRight, Bot, Cpu, Zap, Activity } from 'lucide-react';
import { type CapturedRequest, api } from '../api/client';
import { format } from 'date-fns';

interface ShareViewProps {
    shareId: string;
}

export const ShareView: React.FC<ShareViewProps> = ({ shareId }) => {
    const [trace, setTrace] = useState<CapturedRequest | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchTrace = async () => {
            try {
                const data = await api.getSharedTrace(shareId);
                setTrace(data);
            } catch (err) {
                setError('Shared trace not found or expired.');
            } finally {
                setLoading(false);
            }
        };
        fetchTrace();
    }, [shareId]);

    if (loading) return (
        <div className="min-h-screen bg-black flex items-center justify-center">
            <Zap className="w-8 h-8 text-violet-500 animate-pulse" />
        </div>
    );

    if (error || !trace) return (
        <div className="min-h-screen bg-black flex flex-col items-center justify-center gap-4">
            <Shield className="w-12 h-12 text-rose-500/20" />
            <p className="text-white/40 font-black uppercase text-[10px] tracking-widest">{error}</p>
        </div>
    );

    return (
        <div className="min-h-screen bg-black p-8 lg:p-12 selection:bg-violet-500/30">
            <div className="max-w-[1200px] mx-auto space-y-12">
                <header className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <div className="p-3 bg-violet-500/10 rounded-2xl border border-violet-500/20">
                            <Activity className="w-6 h-6 text-violet-400" />
                        </div>
                        <div>
                            <h1 className="text-2xl font-black">Shared Packet Trace</h1>
                            <p className="text-sm text-white/40 font-medium">Temporary secure link • Expires in 24h</p>
                        </div>
                    </div>
                </header>

                <div className="bg-[#0A0C10]/60 backdrop-blur-xl border border-white/5 rounded-[3rem] p-8 lg:p-12 shadow-2xl space-y-12">
                    {/* Header Info */}
                    <div className="flex flex-wrap items-center gap-8 pb-8 border-b border-white/5">
                        <div className="px-4 py-2 bg-white/5 border border-white/5 rounded-full">
                            <span className="text-xs font-mono text-violet-400 font-black">{trace.method}</span>
                        </div>
                        <div className={`px-4 py-2 rounded-full border ${trace.status_code >= 400 ? 'border-rose-500/20 text-rose-400' : 'border-emerald-500/20 text-emerald-400'}`}>
                            <span className="text-xs font-black">HTTP {trace.status_code}</span>
                        </div>
                        <div className="text-sm font-mono text-white/40 truncate flex-1">
                            {trace.subdomain}.gorenel.site{trace.path}
                        </div>
                    </div>

                    {/* AI Inspector if available */}
                    {trace.ai_metadata && (
                        <div className="p-8 bg-violet-500/5 border border-violet-500/10 rounded-[2.5rem] space-y-8">
                            <div className="flex items-center gap-4">
                                <Bot className="w-6 h-6 text-violet-400" />
                                <div>
                                    <h4 className="font-black text-violet-100">AI Protocol Intelligence</h4>
                                    <p className="text-xs text-violet-400/60 font-mono">{trace.ai_metadata.model}</p>
                                </div>
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                                <div className="space-y-4">
                                    <label className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em] flex items-center gap-2">
                                        <Cpu className="w-3.5 h-3.5" /> Context Prompt
                                    </label>
                                    <div className="p-6 bg-black/40 border border-white/5 rounded-3xl text-sm text-white/70 leading-relaxed font-medium">
                                        {trace.ai_metadata.prompt}
                                    </div>
                                </div>
                                <div className="space-y-4">
                                    <label className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em] flex items-center gap-2">
                                        <Bot className="w-3.5 h-3.5 text-violet-400" /> Model Insight
                                    </label>
                                    <div className="p-6 bg-violet-400/5 border border-violet-400/10 rounded-3xl text-sm text-violet-100/70 leading-relaxed font-medium italic">
                                        {trace.ai_metadata.completion}
                                    </div>
                                </div>
                            </div>
                            <div className="flex items-center gap-8 pt-6 border-t border-violet-500/10">
                                <div className="flex flex-col">
                                    <span className="text-[10px] font-black text-white/20 uppercase tracking-widest">Input</span>
                                    <span className="text-sm font-black text-white/60">{trace.ai_metadata.tokens.prompt} tokens</span>
                                </div>
                                <div className="flex flex-col">
                                    <span className="text-[10px] font-black text-white/20 uppercase tracking-widest">Output</span>
                                    <span className="text-sm font-black text-white/60">{trace.ai_metadata.tokens.completion} tokens</span>
                                </div>
                                <div className="ml-auto px-4 py-1.5 bg-violet-500/20 rounded-full text-[10px] font-black uppercase text-violet-400">
                                    {trace.ai_metadata.provider} Protocol
                                </div>
                            </div>
                        </div>
                    )}

                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
                        <section className="space-y-6">
                            <h3 className="text-xs font-black uppercase tracking-[0.3em] text-white/20 flex items-center gap-3">
                                <ArrowRight className="w-4 h-4 text-emerald-400" /> Request Payload
                            </h3>
                            <div className="p-6 bg-black/40 border border-white/5 rounded-3xl space-y-4 font-mono text-xs overflow-auto max-h-[400px]">
                                {Object.entries(trace.req_headers).map(([k, v]) => (
                                    <div key={k} className="flex gap-4">
                                        <span className="text-emerald-400/60 font-bold shrink-0">{k}:</span>
                                        <span className="text-white/40">{v.join(', ')}</span>
                                    </div>
                                ))}
                                {trace.req_body && (
                                    <div className="mt-8 pt-8 border-t border-white/5 text-white/60">
                                        <pre>{atob(trace.req_body)}</pre>
                                    </div>
                                )}
                            </div>
                        </section>

                        <section className="space-y-6">
                            <h3 className="text-xs font-black uppercase tracking-[0.3em] text-white/20 flex items-center gap-3">
                                <Shield className="w-4 h-4 text-blue-400" /> Response Stack
                            </h3>
                            <div className="p-6 bg-black/40 border border-white/5 rounded-3xl space-y-4 font-mono text-xs overflow-auto max-h-[400px]">
                                {Object.entries(trace.resp_headers).map(([k, v]) => (
                                    <div key={k} className="flex gap-4">
                                        <span className="text-blue-400/60 font-bold shrink-0">{k}:</span>
                                        <span className="text-white/40">{v.join(', ')}</span>
                                    </div>
                                ))}
                                {trace.resp_body && (
                                    <div className="mt-8 pt-8 border-t border-white/5 text-white/60">
                                        <pre>{atob(trace.resp_body)}</pre>
                                    </div>
                                )}
                            </div>
                        </section>
                    </div>

                    <footer className="pt-8 border-t border-white/5 flex items-center justify-between text-[10px] font-black text-white/10 uppercase tracking-widest">
                        <div className="flex items-center gap-8">
                            <div className="flex items-center gap-2">
                                <Clock className="w-3.5 h-3.5" />
                                {format(new Date(trace.timestamp), 'yyyy-MM-dd HH:mm:ss.SSS')}
                            </div>
                            <div className="flex items-center gap-2">
                                <Zap className="w-3.5 h-3.5" />
                                {(trace.duration / 1000000).toFixed(2)}ms Processing
                            </div>
                        </div>
                        <div>Generated by Gorenel Intelligent Tunnel</div>
                    </footer>
                </div>
            </div>
        </div>
    );
};
