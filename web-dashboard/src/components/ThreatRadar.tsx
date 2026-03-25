import React from 'react';
import { ShieldAlert, Cpu, Brain, Zap, Fingerprint, ShieldCheck } from 'lucide-react';
import type { AnomalyRecord } from '../api/client';

interface ThreatRadarProps {
    latestAnomaly?: AnomalyRecord;
}

export const ThreatRadar: React.FC<ThreatRadarProps> = ({ latestAnomaly }) => {
    const ifScore = latestAnomaly?.if_score || 0;
    const aeScore = latestAnomaly?.ae_score || 0;
    const aiScore = latestAnomaly?.detected_by?.includes('AI_SECURITY') ? latestAnomaly.anomaly_score : 0;

    const hasAnomaly = latestAnomaly != null;
    const hasThreat = ifScore > 0 || aeScore > 0 || aiScore > 0;

    const R = 80;
    const cx = 100;
    const cy = 110;

    const pIF = { x: cx, y: cy - R * ifScore };
    const pAE = {
        x: cx + R * aeScore * Math.cos(Math.PI / 6),
        y: cy + R * aeScore * Math.sin(Math.PI / 6)
    };
    const pAI = {
        x: cx - R * aiScore * Math.cos(Math.PI / 6),
        y: cy + R * aiScore * Math.sin(Math.PI / 6)
    };

    const radarPath = `${pIF.x},${pIF.y} ${pAE.x},${pAE.y} ${pAI.x},${pAI.y}`;

    return (
        <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 h-full flex flex-col space-y-5 overflow-hidden relative">
            <div className="flex items-center gap-3 relative z-10">
                <div className={`p-2.5 rounded-xl border transition-colors duration-500 ${hasThreat
                        ? 'bg-rose-500/10 border-rose-500/15'
                        : 'bg-emerald-500/10 border-emerald-500/15'
                    }`}>
                    {hasThreat
                        ? <ShieldAlert className="w-4 h-4 text-rose-400" />
                        : <ShieldCheck className="w-4 h-4 text-emerald-400" />
                    }
                </div>
                <div>
                    <h3 className="text-base font-semibold">Threat Radar</h3>
                    <p className="text-xs text-white/35">Model consensus</p>
                </div>
            </div>

            <div className="flex-1 flex flex-col items-center justify-center relative">
                {hasAnomaly ? (
                    <svg viewBox="0 0 200 220" className="w-full max-w-[240px]">
                        <defs>
                            <radialGradient id="sweepGlow" cx="50%" cy="50%" r="50%">
                                <stop offset="0%" stopColor="#f43f5e" stopOpacity="0.2" />
                                <stop offset="100%" stopColor="#f43f5e" stopOpacity="0" />
                            </radialGradient>
                        </defs>

                        <circle cx={cx} cy={cy} r={R} fill="none" stroke="white" strokeWidth="0.5" strokeDasharray="3 3" opacity="0.06" />
                        <circle cx={cx} cy={cy} r={R * 0.66} fill="none" stroke="white" strokeWidth="0.5" strokeDasharray="3 3" opacity="0.04" />
                        <circle cx={cx} cy={cy} r={R * 0.33} fill="none" stroke="white" strokeWidth="0.5" strokeDasharray="3 3" opacity="0.03" />

                        <line x1={cx} y1={cy} x2={cx} y2={cy - R} stroke="white" strokeWidth="0.5" opacity="0.08" />
                        <line x1={cx} y1={cy} x2={cx + R * Math.cos(Math.PI / 6)} y2={cy + R * Math.sin(Math.PI / 6)} stroke="white" strokeWidth="0.5" opacity="0.08" />
                        <line x1={cx} y1={cy} x2={cx - R * Math.cos(Math.PI / 6)} y2={cy + R * Math.sin(Math.PI / 6)} stroke="white" strokeWidth="0.5" opacity="0.08" />

                        <line x1={cx} y1={cy} x2={cx} y2={cy - R} stroke="#f43f5e" strokeWidth="0.5" opacity="0.3">
                            <animateTransform attributeName="transform" type="rotate" from={`0 ${cx} ${cy}`} to={`360 ${cx} ${cy}`} dur="6s" repeatCount="indefinite" />
                        </line>

                        <polygon
                            points={radarPath}
                            fill="rgba(244, 63, 94, 0.1)"
                            stroke="#f43f5e"
                            strokeWidth="1.5"
                            className="transition-all duration-1000 ease-out"
                        >
                            {hasThreat && (
                                <animate attributeName="fill-opacity" values="0.1;0.25;0.1" dur="2s" repeatCount="indefinite" />
                            )}
                        </polygon>

                        {hasThreat && (
                            <>
                                <circle cx={pIF.x} cy={pIF.y} r="2.5" fill="#f43f5e" opacity={ifScore > 0 ? 1 : 0.15} />
                                <circle cx={pAE.x} cy={pAE.y} r="2.5" fill="#f43f5e" opacity={aeScore > 0 ? 1 : 0.15} />
                                <circle cx={pAI.x} cy={pAI.y} r="2.5" fill="#f43f5e" opacity={aiScore > 0 ? 1 : 0.15} />
                            </>
                        )}

                        <text x={cx} y={cy - R - 8} textAnchor="middle" fill="white" fontSize="6" fontWeight="500" opacity="0.25">ISOLATION FOREST</text>
                        <text x={cx + R * Math.cos(Math.PI / 6) + 4} y={cy + R * Math.sin(Math.PI / 6) + 10} textAnchor="start" fill="white" fontSize="6" fontWeight="500" opacity="0.25">AUTOENCODER</text>
                        <text x={cx - R * Math.cos(Math.PI / 6) - 4} y={cy + R * Math.sin(Math.PI / 6) + 10} textAnchor="end" fill="white" fontSize="6" fontWeight="500" opacity="0.25">AI SECURITY</text>
                    </svg>
                ) : (
                    <div className="flex flex-col items-center gap-3 py-6">
                        <div className="p-3 rounded-full bg-emerald-500/[0.05] border border-emerald-500/10">
                            <ShieldCheck className="w-8 h-8 text-emerald-500/30" />
                        </div>
                        <p className="text-xs text-white/25 text-center">
                            No threats detected<br />
                            <span className="text-[10px] text-white/15">Monitoring active</span>
                        </p>
                    </div>
                )}
            </div>

            <div className="grid grid-cols-3 gap-3 pt-4 border-t border-white/[0.04] relative z-10">
                {[
                    { label: 'IF', icon: Cpu, score: ifScore },
                    { label: 'AE', icon: Brain, score: aeScore },
                    { label: 'AI', icon: Fingerprint, score: aiScore },
                ].map(({ label, icon: Icon, score }) => (
                    <div key={label} className="text-center space-y-1">
                        <div className="flex items-center justify-center gap-1 text-[10px] text-white/20">
                            <Icon className="w-3 h-3" /> {label}
                        </div>
                        <p className={`text-sm font-semibold tabular-nums transition-colors duration-500 ${score > 0.7 ? 'text-rose-400' : score > 0 ? 'text-amber-400' : 'text-white/40'}`}>
                            {(score * 100).toFixed(0)}%
                        </p>
                    </div>
                ))}
            </div>
        </div>
    );
};
