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
        case 'GET': return 'bg-blue-100 text-blue-700';
        case 'POST': return 'bg-green-100 text-green-700';
        case 'PUT': return 'bg-yellow-100 text-yellow-700';
        case 'DELETE': return 'bg-red-100 text-red-700';
        default: return 'bg-neutral-100 text-neutral-700';
    }
}

function getSeverity(score: number): { label: string; color: string } {
    if (score >= 0.9) return { label: 'Critical', color: 'text-red-600 bg-red-50 border-red-200' };
    if (score >= 0.7) return { label: 'High', color: 'text-orange-600 bg-orange-50 border-orange-200' };
    return { label: 'Medium', color: 'text-yellow-600 bg-yellow-50 border-yellow-200' };
}

export const AnomalyAlerts: React.FC<Props> = ({ anomalies }) => {
    if (anomalies.length === 0) {
        return (
            <div className="bg-white rounded-2xl shadow-sm border border-neutral-100 p-6">
                <div className="flex items-center gap-3 mb-4">
                    <div className="w-10 h-10 bg-green-50 rounded-xl flex items-center justify-center">
                        <Shield className="w-5 h-5 text-green-500" />
                    </div>
                    <div>
                        <h3 className="font-semibold text-neutral-800">Security Status</h3>
                        <p className="text-xs text-neutral-500">Anomaly detection active</p>
                    </div>
                </div>
                <div className="flex items-center gap-2 p-4 bg-green-50 rounded-xl border border-green-100">
                    <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                    <span className="text-sm font-medium text-green-700">System secure — No anomalies detected</span>
                </div>
            </div>
        );
    }

    return (
        <div className="bg-white rounded-2xl shadow-sm border border-red-100 p-6">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-red-50 rounded-xl flex items-center justify-center">
                        <AlertTriangle className="w-5 h-5 text-red-500" />
                    </div>
                    <div>
                        <h3 className="font-semibold text-neutral-800">Security Alerts</h3>
                        <p className="text-xs text-neutral-500">Anomalies detected by ML engine</p>
                    </div>
                </div>
                <span className="bg-red-100 text-red-700 text-xs font-semibold px-3 py-1.5 rounded-full animate-pulse">
                    {anomalies.length} alert{anomalies.length !== 1 ? 's' : ''}
                </span>
            </div>

            <div className="space-y-2 max-h-80 overflow-y-auto pr-1">
                {anomalies.slice(0, 15).map((anomaly) => {
                    const severity = getSeverity(anomaly.anomaly_score);
                    return (
                        <div
                            key={anomaly.id}
                            className="p-3 rounded-xl border border-neutral-100 hover:border-red-200 hover:bg-red-50/30 transition-all duration-200"
                        >
                            <div className="flex items-center justify-between mb-2">
                                <div className="flex items-center gap-2">
                                    <span className={`text-xs font-mono font-semibold px-2 py-0.5 rounded ${getMethodColor(anomaly.method)}`}>
                                        {anomaly.method}
                                    </span>
                                    <span className="text-sm font-medium text-neutral-700 truncate max-w-[200px]">
                                        {anomaly.path}
                                    </span>
                                </div>
                                <span className={`text-xs font-semibold px-2 py-0.5 rounded border ${severity.color}`}>
                                    {severity.label}
                                </span>
                            </div>
                            <div className="flex items-center justify-between text-xs text-neutral-500">
                                <div className="flex items-center gap-3">
                                    <span>{anomaly.subdomain}</span>
                                    <span>{anomaly.client_ip}</span>
                                </div>
                                <div className="flex items-center gap-3">
                                    {anomaly.detected_by && (
                                        <span className="text-[10px] font-bold uppercase tracking-wider text-neutral-400 border border-neutral-200 px-1.5 py-0.5 rounded">
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
