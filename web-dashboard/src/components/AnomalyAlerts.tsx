import React from 'react';
import { AlertTriangle, Shield, Clock } from 'lucide-react';

interface AnomalyRecord {
    id: string;
    timestamp: string;
    subdomain: string;
    method: string;
    path: string;
    client_ip: string;
    anomaly_score: number;
    detected_by?: string;
    risk_reason?: string;
}

interface Props {
    anomalies: AnomalyRecord[];
}

function timeAgo(dateStr: string): string {
    const diff = Date.now() - new Date(dateStr).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'just now';
    if (mins < 60) return `${mins}m ago`;
    const hrs = Math.floor(mins / 60);
    if (hrs < 24) return `${hrs}h ago`;
    return `${Math.floor(hrs / 24)}d ago`;
}

function getMethodColor(method: string): string {
    switch (method.toUpperCase()) {
        case 'GET': return 'bg-blue-500/10 text-blue-300 border border-blue-500/20';
        case 'POST': return 'bg-emerald-500/10 text-emerald-300 border border-emerald-500/20';
        case 'PUT': return 'bg-yellow-500/10 text-yellow-300 border border-yellow-500/20';
        case 'DELETE': return 'bg-rose-500/10 text-rose-300 border border-rose-500/20';
        default: return 'bg-white/5 text-white/70 border border-white/10';
    }
}

function getSeverity(score: number): { label: string; color: string } {
    if (score >= 0.9) return { label: 'Critical', color: 'text-rose-300 bg-rose-500/10 border-rose-500/20' };
    if (score >= 0.7) return { label: 'High', color: 'text-orange-300 bg-orange-500/10 border-orange-500/20' };
    return { label: 'Medium', color: 'text-yellow-300 bg-yellow-500/10 border-yellow-500/20' };
}

export const AnomalyAlerts: React.FC<Props> = ({ anomalies }) => {
    if (anomalies.length === 0) {
        return (
            <div className="rounded-2xl border border-white/10 bg-[#0A0C10]/70 p-6">
                <div className="flex items-center gap-3 mb-4">
                    <div className="w-10 h-10 bg-emerald-500/10 rounded-xl flex items-center justify-center border border-emerald-500/20">
                        <Shield className="w-5 h-5 text-emerald-400" />
                    </div>
                    <div>
                        <h3 className="font-semibold text-white">Security Status</h3>
                        <p className="text-xs text-white/50">Anomaly detection active</p>
                    </div>
                </div>
                <div className="flex items-center gap-2 p-4 bg-emerald-500/10 rounded-xl border border-emerald-500/20">
                    <div className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                    <span className="text-sm font-medium text-emerald-300">System secure — No anomalies detected</span>
                </div>
            </div>
        );
    }

    return (
        <div className="rounded-2xl border border-rose-500/20 bg-[#0A0C10]/70 p-6">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-rose-500/10 rounded-xl flex items-center justify-center border border-rose-500/20">
                        <AlertTriangle className="w-5 h-5 text-rose-400" />
                    </div>
                    <div>
                        <h3 className="font-semibold text-white">Security Alerts</h3>
                        <p className="text-xs text-white/50">Anomalies detected by ML engine</p>
                    </div>
                </div>
                <span className="bg-rose-500/10 text-rose-300 border border-rose-500/20 text-xs font-semibold px-3 py-1.5 rounded-full animate-pulse">
                    {anomalies.length} alert{anomalies.length !== 1 ? 's' : ''}
                </span>
            </div>

            <div className="space-y-2 max-h-80 overflow-y-auto pr-1">
                {anomalies.slice(0, 15).map((anomaly) => {
                    const severity = getSeverity(anomaly.anomaly_score);
                    return (
                        <div
                            key={anomaly.id}
                            className="p-3 rounded-xl border border-white/10 hover:border-rose-500/30 hover:bg-rose-500/5 transition-all duration-200"
                        >
                            <div className="flex items-center justify-between mb-2">
                                <div className="flex items-center gap-2">
                                    <span className={`text-xs font-mono font-semibold px-2 py-0.5 rounded ${getMethodColor(anomaly.method)}`}>
                                        {anomaly.method}
                                    </span>
                                    <span className="text-sm font-medium text-white/80 truncate max-w-[200px]">
                                        {anomaly.path}
                                    </span>
                                </div>
                                <span className={`text-xs font-semibold px-2 py-0.5 rounded border ${severity.color}`}>
                                    {severity.label}
                                </span>
                            </div>
                            {anomaly.risk_reason && (
                                <p className="text-xs text-rose-300 font-medium mb-2 px-2 py-1 bg-rose-500/10 rounded-lg border border-rose-500/20">
                                    ⚠ {anomaly.risk_reason}
                                </p>
                            )}
                            <div className="flex items-center justify-between text-xs text-white/50">
                                <div className="flex items-center gap-3">
                                    <span>{anomaly.subdomain}</span>
                                    <span>{anomaly.client_ip}</span>
                                </div>
                                <div className="flex items-center gap-3">
                                    {anomaly.detected_by && (
                                        <span className="text-[10px] font-bold uppercase tracking-wider text-white/60 border border-white/10 px-1.5 py-0.5 rounded">
                                            {anomaly.detected_by}
                                        </span>
                                    )}
                                    <div className="flex items-center gap-1">
                                        <Clock className="w-3 h-3" />
                                        <span>{timeAgo(anomaly.timestamp)}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

export type { AnomalyRecord };
