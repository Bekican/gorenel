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

const modelColors: Record<string, string> = {
    isolation_forest: 'text-blue-400 bg-blue-400/10 border-blue-400/20 glow-blue',
    autoencoder: 'text-violet-400 bg-violet-400/10 border-violet-400/20 glow-violet'
};

export const ModelComparison: React.FC<Props> = ({ ml }) => {
    const { t } = useTranslation();
    const stats: ModelStatsResponse = ml?.stats || {};
    const statusLabel = !ml?.ml_up ? 'Offline' : (ml?.active_tunnels ?? 0) > 0 ? 'Live' : 'Idle';
    const statusStyle = !ml?.ml_up
        ? 'bg-rose-500/10 border-rose-500/20 text-rose-300'
        : (ml?.active_tunnels ?? 0) > 0
            ? 'bg-primary/10 border-primary/20 text-primary'
            : 'bg-white/5 border-white/10 text-white/50';
    return (
        <div className="card h-full p-8 flex flex-col space-y-8">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-white/5 rounded-2xl border border-white/10 shadow-[0_0_20px_rgba(255,255,255,0.05)]">
                        <BrainCircuit className="w-5 h-5 text-white/60" />
                    </div>
                    <div>
                        <h3 className="text-xl font-black">{t('ai_gateway.onboarding_title')}</h3>
                        <p className="text-sm text-white/40 font-medium">{t('ai_gateway.onboarding_desc')}</p>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <div className={`px-3 py-1 border rounded-full flex items-center gap-2 ${statusStyle}`}>
                        <div className={`w-1 h-1 rounded-full ${(ml?.active_tunnels ?? 0) > 0 && ml?.ml_up ? 'bg-primary animate-pulse' : 'bg-white/30'}`} />
                        <span className="text-[10px] font-black uppercase tracking-widest">{statusLabel}</span>
                    </div>
                </div>
            </div>

            <div className="flex-1 grid grid-cols-1 lg:grid-cols-2 gap-6">
                {Object.entries(stats).map(([key, stat]) => (
                    <div key={key} className="relative group overflow-hidden bg-white/[0.02] border border-white/5 rounded-[2rem] p-6 hover:bg-white/[0.05] hover:border-white/10 transition-all duration-300">
                        <div className="relative z-10 space-y-6">
                            <div className="flex items-center justify-between">
                                <span className={`text-[10px] font-black px-3 py-1 rounded-full border uppercase tracking-widest ${modelColors[key] || 'text-white/40 bg-white/5 border-white/10'}`}>
                                    {modelNames[key] || key}
                                </span>
                                {stat.is_trained ? (
                                    <div className="p-2 bg-emerald-500/10 rounded-xl">
                                        <ShieldCheck className="w-4 h-4 text-emerald-500" />
                                    </div>
                                ) : (
                                    <div className="p-2 bg-amber-500/10 rounded-xl animate-pulse">
                                        <AlertCircle className="w-4 h-4 text-amber-500" />
                                    </div>
                                )}
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                                <div className="p-4 bg-black border border-white/5 rounded-2xl space-y-1">
                                    <div className="flex items-center gap-2 text-white/20">
                                        <Activity className="w-3 h-3" />
                                        <span className="text-[8px] font-black uppercase tracking-widest">Detections</span>
                                    </div>
                                    <p className="text-xl font-black text-white selection:bg-primary/30">{stat.total_anomalies}</p>
                                </div>
                                <div className="p-4 bg-black border border-white/5 rounded-2xl space-y-1">
                                    <div className="flex items-center gap-2 text-white/20">
                                        <Zap className="w-3 h-3" />
                                        <span className="text-[8px] font-black uppercase tracking-widest">Latency</span>
                                    </div>
                                    <p className="text-xl font-black text-white selection:bg-primary/30">{stat.avg_inference_ms.toFixed(2)}<span className="text-xs text-white/20 ml-1">ms</span></p>
                                </div>
                            </div>

                            <div className="space-y-2">
                                <div className="flex justify-between items-end">
                                    <span className="text-[10px] font-black text-white/20 uppercase tracking-widest">Heuristic confidence</span>
                                    <span className="text-[10px] font-black text-primary uppercase tracking-widest">
                                        {Math.round(Math.min((stat.total_anomalies / (stat.total_predictions || 1)) * 500, 100))}%
                                    </span>
                                </div>
                                <div className="w-full bg-white/5 rounded-full h-1.5 overflow-hidden">
                                    <div
                                        className={`h-full rounded-full transition-all duration-1000 ${key === 'autoencoder' ? 'bg-violet-400 glow-violet' : 'bg-blue-400 glow-blue'}`}
                                        style={{ width: `${Math.min((stat.total_anomalies / (stat.total_predictions || 1)) * 500, 100)}%` }}
                                    />
                                </div>
                            </div>
                        </div>

                        {/* Decorative background accent */}
                        <div className={`absolute -right-8 -bottom-8 w-32 h-32 blur-[60px] opacity-10 rounded-full ${key === 'autoencoder' ? 'bg-violet-500' : 'bg-blue-500'}`} />
                    </div>
                ))}
            </div>

            {!Object.keys(stats).length && (
                <div className="flex-1 flex flex-col items-center justify-center space-y-4 border-2 border-white/5 border-dashed rounded-[3rem]">
                    <BarChart3 className="w-12 h-12 text-white/5" />
                    <p className="text-[10px] font-black text-white/10 uppercase tracking-[0.3em]">Awaiting signal calibration...</p>
                </div>
            )}
        </div>
    );
};
