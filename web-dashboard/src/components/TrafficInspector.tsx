import React, { useState } from 'react';
import { Search, Play, Clock, Globe, Shield, Terminal, Filter, ArrowRight, Sparkles, Bot, Cpu, Share2 } from 'lucide-react';
import { format } from 'date-fns';
import { type CapturedRequest, api } from '../api/client';

interface TrafficInspectorProps {
    history: CapturedRequest[];
}

export const TrafficInspector: React.FC<TrafficInspectorProps> = ({ history }) => {
    const [selectedId, setSelectedId] = useState<string | null>(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [replaying, setReplaying] = useState<string | null>(null);

    const filteredHistory = history.filter(req =>
        req.path.toLowerCase().includes(searchTerm.toLowerCase()) ||
        req.subdomain.toLowerCase().includes(searchTerm.toLowerCase()) ||
        req.method.toLowerCase().includes(searchTerm.toLowerCase()) ||
        req.status_code.toString().includes(searchTerm)
    );

    const handleReplay = async (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        setReplaying(id);
        try {
            await api.replayRequest(id);
        } catch (err) {
            console.error('Replay failed:', err);
        } finally {
            setReplaying(null);
        }
    };

    const getStatusStyles = (code: number) => {
        if (code >= 200 && code < 300) return 'text-primary bg-primary/10 border-primary/20 glow-emerald';
        if (code >= 400 && code < 500) return 'text-yellow-400 bg-yellow-400/10 border-yellow-400/20';
        if (code >= 500) return 'text-rose-500 bg-rose-500/10 border-rose-500/20 glow-rose';
        return 'text-blue-400 bg-blue-400/10 border-blue-400/20';
    };

    return (
        <div className="card p-0 flex flex-col h-[calc(100vh-250px)]">
            <div className="p-8 border-b border-white/5 space-y-6">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <div className="p-3 bg-white/5 rounded-2xl border border-white/10">
                            <Filter className="w-5 h-5 text-white/60" />
                        </div>
                        <div>
                            <h3 className="text-xl font-black">Packet Streams</h3>
                            <p className="text-sm text-white/40 font-medium">Real-time inspection of tunnel traffic</p>
                        </div>
                    </div>
                    <div className="flex items-center gap-2">
                        <span className="flex items-center gap-2 px-4 py-2 bg-primary/10 text-primary rounded-full text-[10px] font-black uppercase tracking-widest border border-primary/20">
                            <div className="w-1.5 h-1.5 rounded-full bg-primary animate-pulse shadow-[0_0_8px_rgba(16,185,129,1)]" />
                            Capturing Live
                        </span>
                    </div>
                </div>

                <div className="relative group">
                    <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-white/20 group-focus-within:text-primary transition-colors" />
                    <input
                        type="text"
                        placeholder="Search by path, method, or status..."
                        className="w-full pl-12 pr-4 py-4 bg-white/[0.03] border border-white/5 rounded-2xl text-sm font-medium focus:ring-2 focus:ring-primary/20 focus:border-primary/50 transition-all outline-none text-white selection:bg-primary/30"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
            </div>

            <div className="flex-1 overflow-auto">
                <table className="w-full text-left border-collapse">
                    <thead className="sticky top-0 bg-[#0c0c0c]/80 backdrop-blur-md z-10 border-b border-white/5">
                        <tr className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em]">
                            <th className="px-8 py-4">Method</th>
                            <th className="px-8 py-4">Status</th>
                            <th className="px-8 py-4">Cloud Path</th>
                            <th className="px-8 py-4">Time</th>
                            <th className="px-8 py-4 text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-white/[0.03]">
                        {filteredHistory.length === 0 ? (
                            <tr>
                                <td colSpan={5} className="px-8 py-20 text-center">
                                    <div className="space-y-2">
                                        <Terminal className="w-10 h-10 text-white/5 mx-auto" />
                                        <p className="text-white/20 font-black uppercase text-[10px] tracking-widest">No matching logs found</p>
                                    </div>
                                </td>
                            </tr>
                        ) : (
                            filteredHistory.map((req) => (
                                <React.Fragment key={req.id}>
                                    <tr
                                        className={`hover:bg-white/[0.02] cursor-pointer transition-all ${selectedId === req.id ? 'bg-primary/5' : ''}`}
                                        onClick={() => setSelectedId(selectedId === req.id ? null : req.id)}
                                    >
                                        <td className="px-8 py-5">
                                            <span className={`text-[10px] font-black px-2 py-1 rounded-md border ${req.method === 'POST' ? 'border-blue-500/30 text-blue-400 bg-blue-500/5' :
                                                req.method === 'GET' ? 'border-primary/30 text-primary bg-primary/5' :
                                                    'border-white/10 text-white/40'
                                                }`}>
                                                {req.method}
                                            </span>
                                        </td>
                                        <td className="px-8 py-5">
                                            <span className={`text-[10px] font-black px-3 py-1 rounded-full border ${getStatusStyles(req.status_code)}`}>
                                                {req.status_code}
                                            </span>
                                        </td>
                                        <td className="px-8 py-5">
                                            <div className="flex items-center gap-2 max-w-md">
                                                <span className="text-white/20 font-bold shrink-0">{req.subdomain}.</span>
                                                <span className="text-sm font-mono text-white/60 truncate">{req.path}</span>
                                                {req.ai_metadata && (
                                                    <div className="flex items-center gap-1.5 px-2 py-0.5 bg-violet-500/10 border border-violet-500/20 rounded-md text-[9px] font-black uppercase text-violet-400 animate-pulse">
                                                        <Sparkles className="w-2.5 h-2.5" />
                                                        AI {req.ai_metadata.provider}
                                                    </div>
                                                )}
                                            </div>
                                        </td>
                                        <td className="px-8 py-5 text-xs font-bold text-white/20">
                                            {format(new Date(req.timestamp), 'HH:mm:ss.SSS')}
                                        </td>
                                        <td className="px-8 py-5 text-right">
                                            <button
                                                onClick={(e) => handleReplay(req.id, e)}
                                                disabled={replaying === req.id}
                                                className="p-2.5 bg-white/5 border border-white/5 rounded-xl text-white/40 hover:text-primary hover:border-primary/20 hover:bg-primary/10 transition-all disabled:opacity-20"
                                                title="Replay Frame"
                                            >
                                                <Play className={`w-4 h-4 ${replaying === req.id ? 'animate-spin' : ''}`} />
                                            </button>
                                        </td>
                                    </tr>

                                    {selectedId === req.id && (
                                        <tr>
                                            <td colSpan={5} className="px-8 py-0">
                                                <div className="bg-white/[0.02] border-x border-b border-white/5 rounded-b-3xl p-8 mb-4 animate-in slide-in-from-top-2 duration-300">

                                                    {req.ai_metadata && (
                                                        <div className="mb-8 p-6 bg-violet-500/5 border border-violet-500/10 rounded-[2rem] space-y-6">
                                                            <div className="flex items-center justify-between">
                                                                <div className="flex items-center gap-4">
                                                                    <div className="p-3 bg-violet-500/10 rounded-2xl border border-violet-500/20">
                                                                        <Bot className="w-5 h-5 text-violet-400" />
                                                                    </div>
                                                                    <div>
                                                                        <h4 className="text-lg font-black text-violet-100">AI Intelligence Inspector</h4>
                                                                        <p className="text-xs text-violet-400/60 font-medium">Protocol-aware LLM payload analysis</p>
                                                                    </div>
                                                                </div>
                                                                <div className="flex items-center gap-4">
                                                                    <div className="text-right">
                                                                        <div className="text-[10px] font-black text-white/20 uppercase tracking-widest">Model</div>
                                                                        <div className="text-sm font-mono font-black text-white/60">{req.ai_metadata.model}</div>
                                                                    </div>
                                                                    <div className="h-8 w-px bg-white/5" />
                                                                    <div className="text-right">
                                                                        <div className="text-[10px] font-black text-white/20 uppercase tracking-widest">Total Tokens</div>
                                                                        <div className="text-sm font-mono font-black text-violet-400">{req.ai_metadata.tokens.total}</div>
                                                                    </div>
                                                                </div>
                                                            </div>

                                                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                                                <div className="space-y-3">
                                                                    <label className="text-[10px] font-black text-white/20 uppercase tracking-widest flex items-center gap-2">
                                                                        <Cpu className="w-3 h-3" /> Input Prompt ({req.ai_metadata.tokens.prompt} tokens)
                                                                    </label>
                                                                    <div className="p-4 bg-black/40 border border-white/5 rounded-2xl text-xs text-white/70 font-medium leading-relaxed max-h-60 overflow-auto whitespace-pre-wrap">
                                                                        {req.ai_metadata.prompt}
                                                                    </div>
                                                                </div>
                                                                <div className="space-y-3">
                                                                    <label className="text-[10px] font-black text-white/20 uppercase tracking-widest flex items-center gap-2">
                                                                        <Terminal className="w-3 h-3 text-emerald-500" /> Model Completion ({req.ai_metadata.tokens.completion} tokens)
                                                                    </label>
                                                                    <div className="p-4 bg-emerald-500/5 border border-emerald-500/10 rounded-2xl text-xs text-emerald-100/70 font-medium leading-relaxed max-h-60 overflow-auto whitespace-pre-wrap italic">
                                                                        {req.ai_metadata.completion || "Streaming completion in progress or inaccessible..."}
                                                                    </div>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    )}

                                                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
                                                        <div className="space-y-6">
                                                            <div className="flex items-center justify-between border-b border-white/5 pb-4">
                                                                <div className="flex items-center gap-3">
                                                                    <div className="p-2 bg-blue-500/10 rounded-lg">
                                                                        <ArrowRight className="w-4 h-4 text-blue-400" />
                                                                    </div>
                                                                    <span className="font-black text-xs uppercase tracking-widest text-white/40">Request Frame</span>
                                                                </div>
                                                                <button
                                                                    onClick={(e) => {
                                                                        e.stopPropagation();
                                                                        let curl = `curl -X ${req.method} "${req.subdomain}.gorenel.site${req.path}" ${Object.entries(req.req_headers).map(([k, v]) => `-H "${k}: ${v.join(', ')}"`).join(' ')}`;
                                                                        if (req.req_body) {
                                                                            try {
                                                                                const decoded = atob(req.req_body);
                                                                                curl += ` -d '${decoded.replace(/'/g, "'\\''")}'`;
                                                                            } catch {
                                                                                curl += ` -d '${req.req_body.replace(/'/g, "'\\''")}'`;
                                                                            }
                                                                        }
                                                                        navigator.clipboard.writeText(curl);
                                                                    }}
                                                                    className="text-[10px] font-black uppercase text-blue-400 hover:text-white transition-colors"
                                                                >
                                                                    Copy as Curl
                                                                </button>
                                                            </div>
                                                            <div className="space-y-4">
                                                                <div className="space-y-2 font-mono text-xs max-h-48 overflow-auto pr-2">
                                                                    {Object.entries(req.req_headers).map(([k, v]) => (
                                                                        <div key={k} className="flex gap-4 p-2 rounded-lg hover:bg-white/[0.02] transition-colors">
                                                                            <span className="text-blue-400 font-bold shrink-0 w-32 truncate">{k}:</span>
                                                                            <span className="text-white/40 break-all">{v.join(', ')}</span>
                                                                        </div>
                                                                    ))}
                                                                </div>
                                                                {req.req_body && (
                                                                    <div className="mt-4 p-4 rounded-2xl bg-black/40 border border-white/5 font-mono text-[10px] text-white/60 overflow-hidden">
                                                                        <div className="mb-2 text-white/20 font-black uppercase tracking-widest">Payload</div>
                                                                        <pre className="overflow-auto max-h-48 scrollbar-hide">
                                                                            {(() => {
                                                                                try {
                                                                                    const decoded = atob(req.req_body);
                                                                                    try {
                                                                                        return JSON.stringify(JSON.parse(decoded), null, 2);
                                                                                    } catch {
                                                                                        return decoded;
                                                                                    }
                                                                                } catch {
                                                                                    return req.req_body;
                                                                                }
                                                                            })()}
                                                                        </pre>
                                                                    </div>
                                                                )}
                                                            </div>
                                                        </div>

                                                        <div className="space-y-6">
                                                            <div className="flex items-center justify-between border-b border-white/5 pb-4">
                                                                <div className="flex items-center gap-3">
                                                                    <div className="p-2 bg-emerald-500/10 rounded-lg">
                                                                        <Shield className="w-4 h-4 text-emerald-400" />
                                                                    </div>
                                                                    <span className="font-black text-xs uppercase tracking-widest text-white/40">Response Stack</span>
                                                                </div>
                                                            </div>
                                                            <div className="space-y-4">
                                                                <div className="space-y-2 font-mono text-xs max-h-48 overflow-auto pr-2">
                                                                    {Object.entries(req.resp_headers).map(([k, v]) => (
                                                                        <div key={k} className="flex gap-4 p-2 rounded-lg hover:bg-white/[0.02] transition-colors">
                                                                            <span className="text-emerald-400 font-bold shrink-0 w-32 truncate">{k}:</span>
                                                                            <span className="text-white/40 break-all">{v.join(', ')}</span>
                                                                        </div>
                                                                    ))}
                                                                </div>
                                                                {req.resp_body && (
                                                                    <div className="mt-4 p-4 rounded-2xl bg-black/40 border border-white/5 font-mono text-[10px] text-white/60 overflow-hidden">
                                                                        <div className="mb-2 text-white/20 font-black uppercase tracking-widest">Body Output</div>
                                                                        <pre className="overflow-auto max-h-48 scrollbar-hide">
                                                                            {(() => {
                                                                                try {
                                                                                    const decoded = atob(req.resp_body);
                                                                                    try {
                                                                                        return JSON.stringify(JSON.parse(decoded), null, 2);
                                                                                    } catch {
                                                                                        return decoded;
                                                                                    }
                                                                                } catch {
                                                                                    return req.resp_body;
                                                                                }
                                                                            })()}
                                                                        </pre>
                                                                    </div>
                                                                )}
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div className="mt-12 flex items-center gap-8 pt-8 border-t border-white/5">
                                                        <div className="flex items-center gap-3">
                                                            <Clock className="w-4 h-4 text-white/10" />
                                                            <div className="flex flex-col">
                                                                <span className="text-[10px] font-black text-white/10 uppercase tracking-widest">Latency</span>
                                                                <span className="text-sm font-black text-white/60">{(req.duration / 1000000).toFixed(2)}ms</span>
                                                            </div>
                                                        </div>
                                                        <div className="flex items-center gap-3">
                                                            <Globe className="w-4 h-4 text-white/10" />
                                                            <div className="flex flex-col">
                                                                <span className="text-[10px] font-black text-white/10 uppercase tracking-widest">Route</span>
                                                                <span className="text-sm font-black text-white/60">{req.subdomain}.gorenel.site</span>
                                                            </div>
                                                        </div>
                                                        <div className="ml-auto flex items-center gap-4">
                                                            <button
                                                                className="flex items-center gap-2 px-6 py-3 bg-violet-500/10 hover:bg-violet-500/20 text-violet-400 border border-violet-500/20 rounded-2xl text-xs font-black uppercase tracking-widest transition-all"
                                                                onClick={async (e) => {
                                                                    e.stopPropagation();
                                                                    try {
                                                                        const { share_id } = await api.shareTrace(req.id);
                                                                        const shareUrl = `${window.location.origin}/share/${share_id}`;
                                                                        await navigator.clipboard.writeText(shareUrl);
                                                                        alert('Share link copied to clipboard!');
                                                                    } catch (err) {
                                                                        console.error('Sharing failed:', err);
                                                                    }
                                                                }}
                                                            >
                                                                <Share2 className="w-4 h-4" />
                                                                Share Trace
                                                            </button>
                                                            <button
                                                                className="flex items-center gap-2 px-6 py-3 bg-white/5 hover:bg-white/10 rounded-2xl text-xs font-black uppercase tracking-widest transition-all"
                                                                onClick={() => setSelectedId(null)}
                                                            >
                                                                Collapse Inspector
                                                            </button>
                                                        </div>
                                                    </div>
                                                </div>
                                            </td>
                                        </tr>
                                    )}
                                </React.Fragment>
                            ))
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
