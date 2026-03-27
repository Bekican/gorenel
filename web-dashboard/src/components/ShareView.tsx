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
            } catch {
                setError('Shared trace not found or expired.');
            } finally {
                setLoading(false);
            }
        };
        fetchTrace();
    }, [shareId]);

    if (loading) return (
        <div className="min-h-screen bg-[#080a10] flex items-center justify-center">
            <Zap className="w-6 h-6 text-violet-400 animate-pulse" />
        </div>
    );

    if (error || !trace) return (
        <div className="min-h-screen bg-[#080a10] flex flex-col items-center justify-center gap-3">
            <Shield className="w-10 h-10 text-rose-500/20" />
            <p className="text-sm text-white/30">{error}</p>
        </div>
    );

    return (
        <div className="min-h-screen bg-[#080a10] p-6 md:p-10">
            <div className="max-w-[1100px] mx-auto space-y-8">
                <header className="flex items-center gap-3">
                    <div className="p-2.5 bg-violet-500/10 rounded-xl border border-violet-500/15">
                        <Activity className="w-5 h-5 text-violet-400" />
                    </div>
                    <div>
                        <h1 className="text-xl font-semibold">Shared Packet Trace</h1>
                        <p className="text-sm text-white/35">Temporary link &middot; Expires in 24h</p>
                    </div>
                </header>

                <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 md:p-10 space-y-8">
                    <div className="flex flex-wrap items-center gap-4 pb-5 border-b border-white/[0.04]">
                        <span className="px-3 py-1.5 bg-white/[0.04] border border-white/[0.06] rounded-lg text-xs font-mono text-violet-400 font-medium">{trace.method}</span>
                        <span className={`px-3 py-1.5 rounded-lg border text-xs font-medium ${trace.status_code >= 400 ? 'border-rose-500/15 text-rose-400 bg-rose-500/[0.06]' : 'border-emerald-500/15 text-emerald-400 bg-emerald-500/[0.06]'}`}>
                            HTTP {trace.status_code}
                        </span>
                        <span className="text-sm font-mono text-white/35 truncate flex-1">
                            {trace.subdomain}.gorenel.site{trace.path}
                        </span>
                    </div>

                    {trace.ai_metadata && (
                        <div className="p-6 bg-violet-500/[0.04] border border-violet-500/10 rounded-2xl space-y-6">
                            <div className="flex items-center gap-3">
                                <Bot className="w-5 h-5 text-violet-400" />
                                <div>
                                    <h4 className="font-semibold text-violet-100">AI Protocol Intelligence</h4>
                                    <p className="text-xs text-violet-400/50 font-mono">{trace.ai_metadata.model}</p>
                                </div>
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                                <div className="space-y-2">
                                    <label className="text-[11px] font-medium text-white/25 flex items-center gap-1.5">
                                        <Cpu className="w-3 h-3" /> Context Prompt
                                    </label>
                                    <div className="p-4 bg-black/30 border border-white/[0.04] rounded-xl text-sm text-white/60 leading-relaxed">
                                        {trace.ai_metadata.prompt}
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-[11px] font-medium text-white/25 flex items-center gap-1.5">
                                        <Bot className="w-3 h-3 text-violet-400" /> Model Output
                                    </label>
                                    <div className="p-4 bg-violet-400/[0.04] border border-violet-400/10 rounded-xl text-sm text-violet-100/60 leading-relaxed italic">
                                        {trace.ai_metadata.completion}
                                    </div>
                                </div>
                            </div>
                            <div className="flex items-center gap-6 pt-4 border-t border-violet-500/10 text-xs">
                                <div><span className="text-white/20">Input:</span> <span className="text-white/50 font-medium">{trace.ai_metadata.tokens.prompt} tokens</span></div>
                                <div><span className="text-white/20">Output:</span> <span className="text-white/50 font-medium">{trace.ai_metadata.tokens.completion} tokens</span></div>
                                <div className="ml-auto px-2.5 py-1 bg-violet-500/15 rounded-md text-[10px] font-medium text-violet-400">
                                    {trace.ai_metadata.provider}
                                </div>
                            </div>
                        </div>
                    )}

                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                        <section className="space-y-3">
                            <h3 className="text-xs font-medium text-white/25 flex items-center gap-2">
                                <ArrowRight className="w-3.5 h-3.5 text-emerald-400" /> Request
                            </h3>
                            <div className="p-5 bg-black/20 border border-white/[0.04] rounded-xl space-y-3 font-mono text-xs overflow-auto max-h-[360px]">
                                {Object.entries(trace.req_headers).map(([k, v]) => (
                                    <div key={k} className="flex gap-3">
                                        <span className="text-emerald-400/50 shrink-0">{k}:</span>
                                        <span className="text-white/35">{v.join(', ')}</span>
                                    </div>
                                ))}
                                {trace.req_body && (
                                    <div className="mt-4 pt-4 border-t border-white/[0.04] text-white/50">
                                        <pre className="whitespace-pre-wrap">{atob(trace.req_body)}</pre>
                                    </div>
                                )}
                            </div>
                        </section>

                        <section className="space-y-3">
                            <h3 className="text-xs font-medium text-white/25 flex items-center gap-2">
                                <Shield className="w-3.5 h-3.5 text-blue-400" /> Response
                            </h3>
                            <div className="p-5 bg-black/20 border border-white/[0.04] rounded-xl space-y-3 font-mono text-xs overflow-auto max-h-[360px]">
                                {Object.entries(trace.resp_headers).map(([k, v]) => (
                                    <div key={k} className="flex gap-3">
                                        <span className="text-blue-400/50 shrink-0">{k}:</span>
                                        <span className="text-white/35">{v.join(', ')}</span>
                                    </div>
                                ))}
                                {trace.resp_body && (
                                    <div className="mt-4 pt-4 border-t border-white/[0.04] text-white/50">
                                        <pre className="whitespace-pre-wrap">{atob(trace.resp_body)}</pre>
                                    </div>
                                )}
                            </div>
                        </section>
                    </div>

                    <footer className="pt-5 border-t border-white/[0.04] flex items-center justify-between text-[11px] text-white/15">
                        <div className="flex items-center gap-5">
                            <span className="flex items-center gap-1.5"><Clock className="w-3 h-3" /> {format(new Date(trace.timestamp), 'yyyy-MM-dd HH:mm:ss.SSS')}</span>
                            <span className="flex items-center gap-1.5"><Zap className="w-3 h-3" /> {(trace.duration / 1000000).toFixed(2)}ms</span>
                        </div>
                        <span>Gorenel Tunnel</span>
                    </footer>
                </div>
            </div>
        </div>
    );
};
