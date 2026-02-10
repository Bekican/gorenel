import React from 'react';
import { Cpu, Zap, Activity, ShieldCheck, AlertCircle } from 'lucide-react';
import type { ModelStatsResponse } from '../api/client';

interface Props {
    stats: ModelStatsResponse;
}

const modelNames: Record<string, string> = {
    isolation_forest: 'Isolation Forest',
    autoencoder: 'Neural Autoencoder'
};

const modelColors: Record<string, string> = {
    isolation_forest: 'text-blue-500 bg-blue-50',
    autoencoder: 'text-purple-500 bg-purple-50'
};

export const ModelComparison: React.FC<Props> = ({ stats }) => {
    return (
        <div className="bg-white rounded-2xl shadow-sm border border-neutral-100 p-6 h-full">
            <div className="flex items-center gap-3 mb-6">
                <div className="w-10 h-10 bg-neutral-50 rounded-xl flex items-center justify-center">
                    <Cpu className="w-5 h-5 text-neutral-500" />
                </div>
                <div>
                    <h3 className="font-semibold text-neutral-800">Anomaly Engine Analysis</h3>
                    <p className="text-xs text-neutral-500">Dual-model academic comparison</p>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {Object.entries(stats).map(([key, stat]) => (
                    <div key={key} className="p-4 rounded-xl border border-neutral-100 bg-neutral-50/30">
                        <div className="flex items-center justify-between mb-3">
                            <span className={`text-xs font-bold px-2 py-0.5 rounded-full ${modelColors[key] || 'text-neutral-500 bg-neutral-50'}`}>
                                {modelNames[key] || key}
                            </span>
                            {stat.is_trained ? (
                                <ShieldCheck className="w-4 h-4 text-green-500" />
                            ) : (
                                <AlertCircle className="w-4 h-4 text-yellow-500" />
                            )}
                        </div>

                        <div className="space-y-3">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2 text-neutral-500 shrink-0">
                                    <Activity className="w-3.5 h-3.5" />
                                    <span className="text-xs">Detections</span>
                                </div>
                                <span className="text-sm font-semibold text-neutral-800 font-mono">
                                    {stat.total_anomalies}
                                </span>
                            </div>

                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2 text-neutral-500 shrink-0">
                                    <Zap className="w-3.5 h-3.5" />
                                    <span className="text-xs">Latency</span>
                                </div>
                                <span className="text-sm font-semibold text-neutral-800 font-mono">
                                    {stat.avg_inference_ms.toFixed(2)}ms
                                </span>
                            </div>

                            {/* Progress bar visual for confidence/load or similar */}
                            <div className="pt-1">
                                <div className="w-full bg-neutral-200 rounded-full h-1">
                                    <div
                                        className={`h-1 rounded-full ${key === 'autoencoder' ? 'bg-purple-500' : 'bg-blue-500'}`}
                                        style={{ width: `${Math.min((stat.total_anomalies / (stat.total_predictions || 1)) * 500, 100)}%` }}
                                    />
                                </div>
                            </div>
                        </div>
                    </div>
                ))}
            </div>

            {!Object.keys(stats).length && (
                <div className="text-center py-8 text-neutral-400 text-sm italic">
                    Fetching model statistics...
                </div>
            )}
        </div>
    );
};
