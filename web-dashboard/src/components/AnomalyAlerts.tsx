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
        case 'GET': return 'bg-blue-500/10 text-blue-400 border border-blue-500/15';
        case 'POST': return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/15';
        case 'PUT': return 'bg-yellow-500/10 text-yellow-400 border border-yellow-500/15';
        case 'DELETE': return 'bg-rose-500/10 text-rose-400 border border-rose-500/15';
        default: return 'bg-white/[0.04] text-white/60 border border-white/[0.08]';
    }
}

function getSeverity(score: number): { label: string; color: string } {
    if (score >= 0.9) return { label: 'Critical', color: 'text-rose-300 bg-rose-500/10 border-rose-500/15' };
    if (score >= 0.7) return { label: 'High', color: 'text-orange-300 bg-orange-500/10 border-orange-500/15' };
    if (score >= 0.5) return { label: 'Medium', color: 'text-yellow-300 bg-yellow-500/10 border-yellow-500/15' };
    return { label: 'Low', color: 'text-sky-300 bg-sky-500/10 border-sky-500/15' };
}

export const AnomalyAlerts: React.FC<Props> = ({ anomalies }) => {
    if (anomalies.length === 0) {
        return (
            <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-5">
                <div className="flex items-center gap-2.5 mb-4">
                    <div className="w-9 h-9 bg-emerald-500/10 rounded-xl flex items-center justify-center border border-emerald-500/15">
                        <Shield className="w-4 h-4 text-emerald-400" />
                    </div>
                    <div>
                        <h3 className="font-medium text-sm text-white">Security Status</h3>
                        <p className="text-[11px] text-white/40">Anomaly detection active</p>
                    </div>
                </div>
                <div className="flex items-center gap-2 p-3 bg-emerald-500/[0.06] rounded-xl border border-emerald-500/15">
                    <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                    <span className="text-xs font-medium text-emerald-300/80">System secure — No anomalies</span>
                </div>
            </div>
        );
    }

    return (
        <div className="rounded-2xl border border-rose-500/15 bg-white/[0.02] p-5">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2.5">
                    <div className="w-9 h-9 bg-rose-500/10 rounded-xl flex items-center justify-center border border-rose-500/15">
                        <AlertTriangle className="w-4 h-4 text-rose-400" />
                    </div>
                    <div>
                        <h3 className="font-medium text-sm text-white">Security Alerts</h3>
                        <p className="text-[11px] text-white/40">ML-detected anomalies</p>
                    </div>
                </div>
                <span className="bg-rose-500/10 text-rose-300 border border-rose-500/15 text-[11px] font-medium px-2 py-1 rounded-lg">
                    {anomalies.length} alert{anomalies.length !== 1 ? 's' : ''}
                </span>
            </div>

            <div className="space-y-2 max-h-72 overflow-y-auto pr-1">
                {anomalies.slice(0, 15).map((anomaly) => {
                    const severity = getSeverity(anomaly.anomaly_score);
                    return (
                        <div
                            key={anomaly.id}
                            className="p-3 rounded-xl border border-white/[0.06] hover:border-rose-500/20 hover:bg-rose-500/[0.03] transition-all duration-200"
                        >
                            <div className="flex items-center justify-between mb-2">
                                <div className="flex items-center gap-2">
                                    <span className={`text-[10px] font-mono font-medium px-1.5 py-0.5 rounded ${getMethodColor(anomaly.method)}`}>
                                        {anomaly.method}
                                    </span>
                                    <span className="text-sm text-white/70 truncate max-w-[180px]">
                                        {anomaly.path}
                                    </span>
                                </div>
                                <span className={`text-[10px] font-medium px-1.5 py-0.5 rounded border ${severity.color}`}>
                                    {severity.label}
                                </span>
                            </div>
                            {anomaly.risk_reason && (
                                <p className="text-xs text-rose-300/70 mb-2 px-2 py-1 bg-rose-500/[0.06] rounded-lg border border-rose-500/15">
                                    {anomaly.risk_reason}
                                </p>
                            )}
                            <div className="flex items-center justify-between text-[11px] text-white/35">
                                <div className="flex items-center gap-2">
                                    <span>{anomaly.subdomain}</span>
                                    <span>{anomaly.client_ip}</span>
                                </div>
                                <div className="flex items-center gap-2">
                                    {anomaly.detected_by && (
                                        <span className="text-[10px] font-medium text-white/50 border border-white/[0.08] px-1.5 py-0.5 rounded">
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
