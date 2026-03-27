import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Search, Play, Clock, Globe, Shield, Terminal, Filter, ArrowRight, Sparkles, Bot, Cpu, Share2, Activity } from 'lucide-react';
import { format } from 'date-fns';
import { type CapturedRequest, api } from '../api/client';

interface TrafficInspectorProps {
    history: CapturedRequest[];
}

export const TrafficInspector: React.FC<TrafficInspectorProps> = ({ history }) => {
    const { t } = useTranslation();
    const [selectedId, setSelectedId] = useState<string | null>(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [replaying, setReplaying] = useState<string | null>(null);

    const filteredHistory = history.filter(req =>
        req.path.toLowerCase().includes(searchTerm.toLowerCase()) ||
        req.subdomain.toLowerCase().includes(searchTerm.toLowerCase()) ||
        req.method.toLowerCase().includes(searchTerm.toLowerCase()) ||
        req.status_code.toString().includes(searchTerm)
    );
    const hasActiveFilter = searchTerm.trim().length > 0;

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
        if (code >= 200 && code < 300) return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/15';
        if (code >= 400 && code < 500) return 'text-yellow-400 bg-yellow-400/10 border-yellow-400/15';
        if (code >= 500) return 'text-rose-400 bg-rose-500/10 border-rose-500/15';
        return 'text-blue-400 bg-blue-400/10 border-blue-400/15';
    };

    return (
        <div className="flex flex-col h-[calc(100dvh-220px)] sm:h-[calc(100vh-220px)] min-h-[520px] overflow-hidden">
            <div className="p-6 border-b border-white/[0.04] space-y-4 shrink-0">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="p-2.5 bg-white/[0.04] rounded-xl border border-white/[0.06]">
                            <Filter className="w-4 h-4 text-white/50" />
                        </div>
                        <div>
                            <h3 className="text-base font-semibold">{t('traffic_inspector.title')}</h3>
                            <p className="text-sm text-white/35">{t('traffic_inspector.subtitle')}</p>
                        </div>
                    </div>
                    <div className="flex items-center gap-1.5 px-3 py-1.5 bg-emerald-500/10 text-emerald-400 rounded-lg text-[10px] font-medium border border-emerald-500/15">
                        <div className="w-1 h-1 rounded-full bg-emerald-400 animate-pulse" />
                        {t('traffic_inspector.capturing')}
                    </div>
                </div>

                <div className="relative group">
                    <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-white/15 group-focus-within:text-emerald-400/60 transition-colors" />
                    <input
                        type="text"
                        placeholder={t('traffic_inspector.search_placeholder')}
                        className="w-full pl-10 pr-4 py-3 bg-white/[0.02] border border-white/[0.06] rounded-xl text-sm font-medium focus:ring-1 focus:ring-emerald-500/20 focus:border-emerald-500/30 transition-all outline-none text-white"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
            </div>
            
            <div className="flex-1 overflow-y-auto relative">
                {/* Mobile: card list */}
                <div className="sm:hidden divide-y divide-white/[0.03]">
                    {filteredHistory.length === 0 ? (
                        <div className="px-6 py-16 text-center">
                            <div className="max-w-sm mx-auto space-y-4">
                                <div className="w-14 h-14 bg-emerald-500/[0.06] border border-emerald-500/15 rounded-xl flex items-center justify-center mx-auto">
                                    <Activity className="w-7 h-7 text-emerald-400/50" />
                                </div>
                                <div className="space-y-1.5">
                                    <h4 className="text-base font-semibold text-white/60">
                                        {hasActiveFilter ? t('traffic_inspector.no_results_title', 'No matching requests') : t('traffic_inspector.listening_title')}
                                    </h4>
                                    <p className="text-sm text-white/25">
                                        {hasActiveFilter ? t('traffic_inspector.no_results_desc', 'Try a different search.') : t('traffic_inspector.listening_desc')}
                                    </p>
                                </div>
                            </div>
                        </div>
                    ) : (
                        filteredHistory.map((req) => {
                            const open = selectedId === req.id;
                            return (
                                <div key={req.id} className="px-5 py-4">
                                    <button
                                        type="button"
                                        onClick={() => setSelectedId(open ? null : req.id)}
                                        className="w-full text-left rounded-xl border border-white/[0.06] bg-white/[0.02] hover:bg-white/[0.03] hover:border-white/[0.1] transition-colors p-4"
                                    >
                                        <div className="flex items-center justify-between gap-2">
                                            <span className={`text-[10px] font-medium px-2 py-0.5 rounded border ${req.method === 'POST'
                                                ? 'border-blue-500/20 text-blue-400 bg-blue-500/[0.06]'
                                                : req.method === 'GET'
                                                    ? 'border-emerald-500/20 text-emerald-400 bg-emerald-500/[0.06]'
                                                    : 'border-white/[0.08] text-white/40 bg-white/[0.03]'
                                                }`}>
                                                {req.method}
                                            </span>
                                            <span className={`text-[10px] font-medium px-2 py-0.5 rounded-md border ${getStatusStyles(req.status_code)}`}>
                                                {req.status_code}
                                            </span>
                                        </div>

                                        <div className="mt-2 flex items-center gap-2 min-w-0">
                                            <span className="text-white/15 text-[11px] shrink-0">{req.subdomain}.</span>
                                            <span className="text-sm font-mono text-white/55 truncate">{req.path}</span>
                                            {req.ai_metadata && (
                                                <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-violet-500/10 border border-violet-500/15 rounded text-[9px] font-medium text-violet-400 shrink-0">
                                                    <Sparkles className="w-2.5 h-2.5" /> AI
                                                </span>
                                            )}
                                        </div>

                                        <div className="mt-3 flex items-center justify-between text-[11px] text-white/30">
                                            <span className="tabular-nums">{format(new Date(req.timestamp), 'HH:mm:ss')}</span>
                                            <button
                                                onClick={(e) => handleReplay(req.id, e)}
                                                disabled={replaying === req.id}
                                                className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white/[0.03] border border-white/[0.06] rounded-lg text-white/35 hover:text-emerald-400 hover:border-emerald-500/15 hover:bg-emerald-500/[0.06] transition-all disabled:opacity-30"
                                                title="Replay"
                                            >
                                                <Play className={`w-3.5 h-3.5 ${replaying === req.id ? 'animate-spin' : ''}`} />
                                                Replay
                                            </button>
                                        </div>
                                    </button>

                                    {open && (
                                        <div className="mt-3 rounded-xl border border-white/[0.06] bg-white/[0.02] p-4 space-y-3">
                                            <div className="flex items-center justify-between text-[11px] text-white/35">
                                                <span className="font-medium text-white/60">Details</span>
                                                <span className="tabular-nums">{(req.duration / 1000000).toFixed(2)}ms</span>
                                            </div>
                                            <div className="space-y-1.5 font-mono text-[11px] text-white/45">
                                                <div className="break-all"><span className="text-white/25">Host:</span> {req.subdomain}.gorenel.site</div>
                                                <div className="break-all"><span className="text-white/25">Path:</span> {req.path}</div>
                                            </div>
                                            <div className="flex gap-2">
                                                <button
                                                    className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 bg-violet-500/10 hover:bg-violet-500/15 text-violet-400 border border-violet-500/15 rounded-xl text-xs font-medium transition-all"
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
                                                    <Share2 className="w-3.5 h-3.5" />
                                                    Share
                                                </button>
                                                <button
                                                    className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 bg-white/[0.04] hover:bg-white/[0.07] rounded-xl text-xs font-medium text-white/50 transition-all"
                                                    onClick={() => setSelectedId(null)}
                                                >
                                                    {t('traffic_inspector.collapse')}
                                                </button>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            );
                        })
                    )}
                </div>

                {/* Desktop/tablet: table */}
                <div className="hidden sm:block overflow-x-auto">
                <table className="w-full min-w-[720px] text-left border-separate border-spacing-0">
                    <thead className="sticky top-0 bg-[#0c0e14]/90 backdrop-blur-lg z-10">
                        <tr className="text-[10px] font-medium text-white/20 uppercase tracking-wider">
                            <th className="px-6 py-3">{t('traffic_inspector.method')}</th>
                            <th className="px-6 py-3">{t('traffic_inspector.status')}</th>
                            <th className="px-6 py-3">{t('traffic_inspector.path')}</th>
                            <th className="px-6 py-3">{t('traffic_inspector.time')}</th>
                            <th className="px-6 py-3 text-right">{t('traffic_inspector.actions')}</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-white/[0.03]">
                        {filteredHistory.length === 0 ? (
                            <tr>
                                <td colSpan={5} className="px-6 py-24 text-center">
                                    <div className="max-w-sm mx-auto space-y-4">
                                        <div className="w-14 h-14 bg-emerald-500/[0.06] border border-emerald-500/15 rounded-xl flex items-center justify-center mx-auto">
                                           <Activity className="w-7 h-7 text-emerald-400/50" />
                                        </div>
                                        <div className="space-y-1.5">
                                            <h4 className="text-base font-semibold text-white/60">
                                                {hasActiveFilter ? t('traffic_inspector.no_results_title', 'No matching requests') : t('traffic_inspector.listening_title')}
                                            </h4>
                                            <p className="text-sm text-white/25">
                                                {hasActiveFilter ? t('traffic_inspector.no_results_desc', 'Try a different search.') : t('traffic_inspector.listening_desc')}
                                            </p>
                                        </div>
                                    </div>
                                </td>
                            </tr>
                        ) : (
                            filteredHistory.map((req) => (
                                <React.Fragment key={req.id}>
                                    <tr
                                        className={`hover:bg-white/[0.02] cursor-pointer transition-colors ${selectedId === req.id ? 'bg-emerald-500/[0.03]' : ''}`}
                                        onClick={() => setSelectedId(selectedId === req.id ? null : req.id)}
                                    >
                                        <td className="px-6 py-4">
                                            <span className={`text-[10px] font-medium px-2 py-0.5 rounded border ${req.method === 'POST' ? 'border-blue-500/20 text-blue-400 bg-blue-500/[0.06]' :
                                                req.method === 'GET' ? 'border-emerald-500/20 text-emerald-400 bg-emerald-500/[0.06]' :
                                                    'border-white/[0.08] text-white/40 bg-white/[0.03]'
                                                }`}>
                                                {req.method}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4">
                                            <span className={`text-[10px] font-medium px-2 py-0.5 rounded-md border ${getStatusStyles(req.status_code)}`}>
                                                {req.status_code}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4">
                                            <div className="flex items-center gap-1.5 max-w-md">
                                                <span className="text-white/15 text-xs shrink-0">{req.subdomain}.</span>
                                                <span className="text-sm font-mono text-white/50 truncate">{req.path}</span>
                                                {req.ai_metadata && (
                                                    <div className="flex items-center gap-1 px-1.5 py-0.5 bg-violet-500/10 border border-violet-500/15 rounded text-[9px] font-medium text-violet-400">
                                                        <Sparkles className="w-2.5 h-2.5" /> AI
                                                    </div>
                                                )}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 text-xs text-white/20 tabular-nums">
                                            {format(new Date(req.timestamp), 'HH:mm:ss.SSS')}
                                        </td>
                                        <td className="px-6 py-4 text-right">
                                            <button
                                                onClick={(e) => handleReplay(req.id, e)}
                                                disabled={replaying === req.id}
                                                className="p-2 bg-white/[0.03] border border-white/[0.06] rounded-lg text-white/30 hover:text-emerald-400 hover:border-emerald-500/15 hover:bg-emerald-500/[0.06] transition-all disabled:opacity-20"
                                                title="Replay"
                                            >
                                                <Play className={`w-3.5 h-3.5 ${replaying === req.id ? 'animate-spin' : ''}`} />
                                            </button>
                                        </td>
                                    </tr>

                                    {selectedId === req.id && (
                                        <tr>
                                            <td colSpan={5} className="px-6 py-0">
                                                <div className="bg-white/[0.02] border border-white/[0.04] rounded-xl p-6 mb-3 animate-fade-in-up space-y-6">

                                                    {req.ai_metadata && (
                                                        <div className="p-5 bg-violet-500/[0.04] border border-violet-500/10 rounded-xl space-y-4">
                                                            <div className="flex items-center justify-between">
                                                                <div className="flex items-center gap-3">
                                                                    <Bot className="w-4 h-4 text-violet-400" />
                                                                    <div>
                                                                        <h4 className="font-medium text-sm text-violet-100">{t('traffic_inspector.ai_inspector')}</h4>
                                                                        <p className="text-[11px] text-violet-400/40">{t('traffic_inspector.ai_desc')}</p>
                                                                    </div>
                                                                </div>
                                                                <div className="flex items-center gap-4 text-xs">
                                                                    <div className="text-right">
                                                                        <div className="text-[10px] text-white/20">{t('traffic_inspector.model')}</div>
                                                                        <div className="font-mono text-white/50">{req.ai_metadata.model}</div>
                                                                    </div>
                                                                    <div className="h-6 w-px bg-white/[0.04]" />
                                                                    <div className="text-right">
                                                                        <div className="text-[10px] text-white/20">{t('traffic_inspector.tokens')}</div>
                                                                        <div className="font-mono text-violet-400">{req.ai_metadata.tokens.total}</div>
                                                                    </div>
                                                                </div>
                                                            </div>

                                                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                                <div className="space-y-1.5">
                                                                    <label className="text-[10px] text-white/20 flex items-center gap-1.5">
                                                                        <Cpu className="w-3 h-3" /> {t('traffic_inspector.prompt')} ({req.ai_metadata.tokens.prompt} tokens)
                                                                    </label>
                                                                    <div className="p-3 bg-black/30 border border-white/[0.04] rounded-lg text-xs text-white/55 leading-relaxed max-h-48 overflow-auto whitespace-pre-wrap">
                                                                        {req.ai_metadata.prompt}
                                                                    </div>
                                                                </div>
                                                                <div className="space-y-1.5">
                                                                    <label className="text-[10px] text-white/20 flex items-center gap-1.5">
                                                                        <Terminal className="w-3 h-3 text-emerald-400" /> {t('traffic_inspector.completion')} ({req.ai_metadata.tokens.completion} tokens)
                                                                    </label>
                                                                    <div className="p-3 bg-emerald-500/[0.04] border border-emerald-500/10 rounded-lg text-xs text-emerald-100/55 leading-relaxed max-h-48 overflow-auto whitespace-pre-wrap italic">
                                                                        {req.ai_metadata.completion || "Streaming in progress..."}
                                                                    </div>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    )}

                                                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                                                        <div className="space-y-4">
                                                            <div className="flex items-center justify-between pb-3 border-b border-white/[0.04]">
                                                                <div className="flex items-center gap-2">
                                                                    <ArrowRight className="w-3.5 h-3.5 text-blue-400" />
                                                                    <span className="text-xs font-medium text-white/30">{t('traffic_inspector.req_frame')}</span>
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
                                                                    className="text-[10px] font-medium text-blue-400 hover:text-white transition-colors"
                                                                >
                                                                    Copy as curl
                                                                </button>
                                                            </div>
                                                            <div className="space-y-1.5 font-mono text-xs max-h-40 overflow-auto">
                                                                {Object.entries(req.req_headers).map(([k, v]) => (
                                                                    <div key={k} className="flex gap-3 py-1 px-2 rounded hover:bg-white/[0.02]">
                                                                        <span className="text-blue-400/60 shrink-0 w-28 truncate">{k}:</span>
                                                                        <span className="text-white/30 break-all">{v.join(', ')}</span>
                                                                    </div>
                                                                ))}
                                                            </div>
                                                            {req.req_body && (
                                                                <div className="p-3 rounded-lg bg-black/20 border border-white/[0.04] font-mono text-[10px] text-white/45 overflow-auto max-h-40">
                                                                    <pre className="whitespace-pre-wrap">
                                                                        {(() => {
                                                                            try {
                                                                                const decoded = atob(req.req_body);
                                                                                try { return JSON.stringify(JSON.parse(decoded), null, 2); } catch { return decoded; }
                                                                            } catch { return req.req_body; }
                                                                        })()}
                                                                    </pre>
                                                                </div>
                                                            )}
                                                        </div>

                                                        <div className="space-y-4">
                                                            <div className="flex items-center gap-2 pb-3 border-b border-white/[0.04]">
                                                                <Shield className="w-3.5 h-3.5 text-emerald-400" />
                                                                <span className="text-xs font-medium text-white/30">{t('traffic_inspector.resp_stack')}</span>
                                                            </div>
                                                            <div className="space-y-1.5 font-mono text-xs max-h-40 overflow-auto">
                                                                {Object.entries(req.resp_headers).map(([k, v]) => (
                                                                    <div key={k} className="flex gap-3 py-1 px-2 rounded hover:bg-white/[0.02]">
                                                                        <span className="text-emerald-400/60 shrink-0 w-28 truncate">{k}:</span>
                                                                        <span className="text-white/30 break-all">{v.join(', ')}</span>
                                                                    </div>
                                                                ))}
                                                            </div>
                                                            {req.resp_body && (
                                                                <div className="p-3 rounded-lg bg-black/20 border border-white/[0.04] font-mono text-[10px] text-white/45 overflow-auto max-h-40">
                                                                    <pre className="whitespace-pre-wrap">
                                                                        {(() => {
                                                                            try {
                                                                                const decoded = atob(req.resp_body);
                                                                                try { return JSON.stringify(JSON.parse(decoded), null, 2); } catch { return decoded; }
                                                                            } catch { return req.resp_body; }
                                                                        })()}
                                                                    </pre>
                                                                </div>
                                                            )}
                                                        </div>
                                                    </div>

                                                    <div className="flex items-center gap-6 pt-4 border-t border-white/[0.04]">
                                                        <div className="flex items-center gap-2">
                                                            <Clock className="w-3.5 h-3.5 text-white/15" />
                                                            <div>
                                                                <span className="text-[10px] text-white/15">Latency</span>
                                                                <span className="text-sm font-medium text-white/50 ml-1.5 tabular-nums">{(req.duration / 1000000).toFixed(2)}ms</span>
                                                            </div>
                                                        </div>
                                                        <div className="flex items-center gap-2">
                                                            <Globe className="w-3.5 h-3.5 text-white/15" />
                                                            <span className="text-sm text-white/40">{req.subdomain}.gorenel.site</span>
                                                        </div>
                                                        <div className="ml-auto flex items-center gap-2">
                                                            <button
                                                                className="flex items-center gap-1.5 px-4 py-2 bg-violet-500/10 hover:bg-violet-500/15 text-violet-400 border border-violet-500/15 rounded-xl text-xs font-medium transition-all"
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
                                                                <Share2 className="w-3.5 h-3.5" />
                                                                {t('traffic_inspector.share')}
                                                            </button>
                                                            <button
                                                                className="flex items-center gap-1.5 px-4 py-2 bg-white/[0.04] hover:bg-white/[0.07] rounded-xl text-xs font-medium text-white/50 transition-all"
                                                                onClick={() => setSelectedId(null)}
                                                            >
                                                                {t('traffic_inspector.collapse')}
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
        </div>
    );
};
