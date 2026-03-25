import { Zap, Activity, ShieldCheck, AlertCircle, BarChart3, BrainCircuit } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { MLStatsEnvelope, ModelStatsResponse } from '../api/client';

interface Props {
    ml: MLStatsEnvelope;
}

const modelNames: Record<string, string> = {
    isolation_forest: 'Isolation Forest',
    autoencoder: 'Neural Autoencoder'
};

const modelColors: Record<string, { badge: string; bar: string; glow: string }> = {
    isolation_forest: {
        badge: 'text-blue-400 bg-blue-400/10 border-blue-400/15',
        bar: 'bg-blue-400',
        glow: 'bg-blue-500',
    },
    autoencoder: {
        badge: 'text-violet-400 bg-violet-400/10 border-violet-400/15',
        bar: 'bg-violet-400',
        glow: 'bg-violet-500',
    }
};

export const ModelComparison: React.FC<Props> = ({ ml }) => {
    const { t } = useTranslation();
    const stats: ModelStatsResponse = ml?.stats || {};
    const statusLabel = !ml?.ml_up ? 'Offline' : (ml?.active_tunnels ?? 0) > 0 ? 'Live' : 'Idle';
    const statusStyle = !ml?.ml_up
        ? 'bg-rose-500/10 border-rose-500/15 text-rose-300'
        : (ml?.active_tunnels ?? 0) > 0
            ? 'bg-emerald-500/10 border-emerald-500/15 text-emerald-400'
            : 'bg-white/[0.04] border-white/[0.08] text-white/40';

    return (
        <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 md:p-8 h-full flex flex-col space-y-6">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-white/[0.04] rounded-xl border border-white/[0.06]">
                        <BrainCircuit className="w-5 h-5 text-white/50" />
                    </div>
                    <div>
                        <h3 className="text-lg font-semibold">{t('ai_gateway.onboarding_title')}</h3>
                        <p className="text-sm text-white/35">{t('ai_gateway.onboarding_desc')}</p>
                    </div>
                </div>
                <div className={`px-2.5 py-1 border rounded-lg flex items-center gap-1.5 ${statusStyle}`}>
                    <div className={`w-1 h-1 rounded-full ${(ml?.active_tunnels ?? 0) > 0 && ml?.ml_up ? 'bg-emerald-400 animate-pulse' : 'bg-white/20'}`} />
                    <span className="text-[10px] font-medium">{statusLabel}</span>
                </div>
            </div>

            <div className="flex-1 grid grid-cols-1 lg:grid-cols-2 gap-4">
                {Object.entries(stats).map(([key, stat]) => {
                    const colors = modelColors[key] || { badge: 'text-white/40 bg-white/[0.04] border-white/[0.08]', bar: 'bg-white/30', glow: 'bg-white/20' };
                    const confidence = Math.min((stat.total_anomalies / (stat.total_predictions || 1)) * 500, 100);
                    return (
                        <div key={key} className="relative overflow-hidden bg-white/[0.02] border border-white/[0.06] rounded-xl p-5 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-200">
                            <div className="relative z-10 space-y-5">
                                <div className="flex items-center justify-between">
                                    <span className={`text-[10px] font-medium px-2.5 py-1 rounded-md border ${colors.badge}`}>
                                        {modelNames[key] || key}
                                    </span>
                                    {stat.is_trained ? (
                                        <div className="p-1.5 bg-emerald-500/10 rounded-lg">
                                            <ShieldCheck className="w-3.5 h-3.5 text-emerald-500" />
                                        </div>
                                    ) : (
                                        <div className="p-1.5 bg-amber-500/10 rounded-lg animate-pulse">
                                            <AlertCircle className="w-3.5 h-3.5 text-amber-500" />
                                        </div>
                                    )}
                                </div>

                                <div className="grid grid-cols-2 gap-3">
                                    <div className="p-3 bg-black/20 border border-white/[0.04] rounded-lg space-y-1">
                                        <div className="flex items-center gap-1.5 text-white/20">
                                            <Activity className="w-3 h-3" />
                                            <span className="text-[10px] font-medium">Detections</span>
                                        </div>
                                        <p className="text-lg font-semibold text-white tabular-nums">{stat.total_anomalies}</p>
                                    </div>
                                    <div className="p-3 bg-black/20 border border-white/[0.04] rounded-lg space-y-1">
                                        <div className="flex items-center gap-1.5 text-white/20">
                                            <Zap className="w-3 h-3" />
                                            <span className="text-[10px] font-medium">Latency</span>
                                        </div>
                                        <p className="text-lg font-semibold text-white tabular-nums">{stat.avg_inference_ms.toFixed(2)}<span className="text-xs text-white/25 ml-0.5">ms</span></p>
                                    </div>
                                </div>

                                <div className="space-y-2">
                                    <div className="flex justify-between items-center">
                                        <span className="text-[10px] text-white/25">Confidence</span>
                                        <span className="text-[10px] font-medium text-emerald-400 tabular-nums">{Math.round(confidence)}%</span>
                                    </div>
                                    <div className="w-full bg-white/[0.04] rounded-full h-1 overflow-hidden">
                                        <div
                                            className={`h-full rounded-full transition-all duration-1000 ${colors.bar}`}
                                            style={{ width: `${confidence}%` }}
                                        />
                                    </div>
                                </div>
                            </div>

                            <div className={`absolute -right-6 -bottom-6 w-24 h-24 blur-[50px] opacity-[0.06] rounded-full ${colors.glow}`} />
                        </div>
                    );
                })}
            </div>

            {!Object.keys(stats).length && (
                <div className="flex-1 flex flex-col items-center justify-center space-y-3 border border-dashed border-white/[0.06] rounded-2xl py-12">
                    <BarChart3 className="w-10 h-10 text-white/[0.06]" />
                    <p className="text-xs text-white/15">Awaiting signal calibration...</p>
                </div>
            )}
        </div>
    );
};
