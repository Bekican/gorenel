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
        <div className="card h-full p-8 flex flex-col space-y-8 overflow-hidden relative">
            {/* Header */}
            <div className="flex items-center gap-4 relative z-10">
                <div className={`p-3 rounded-2xl border shadow-[0_0_20px_rgba(244,63,94,0.1)] transition-colors duration-500 ${hasThreat
                        ? 'bg-rose-500/10 border-rose-500/20'
                        : 'bg-emerald-500/10 border-emerald-500/20'
                    }`}>
                    {hasThreat
                        ? <ShieldAlert className="w-5 h-5 text-rose-500" />
                        : <ShieldCheck className="w-5 h-5 text-emerald-500" />
                    }
                </div>
                <div>
                    <h3 className="text-xl font-black">Threat Radar</h3>
                    <p className="text-sm text-white/40 font-medium">Model Consensus</p>
                </div>
            </div>

            {/* Radar Area */}
            <div className="flex-1 flex flex-col items-center justify-center relative">
                {hasAnomaly ? (
                    <svg viewBox="0 0 200 220" className="w-full max-w-[280px] drop-shadow-[0_0_30px_rgba(244,63,94,0.15)]">
                        {/* Defs: Pulse + Sweep animations */}
                        <defs>
                            <radialGradient id="sweepGlow" cx="50%" cy="50%" r="50%">
                                <stop offset="0%" stopColor="#f43f5e" stopOpacity="0.3" />
                                <stop offset="100%" stopColor="#f43f5e" stopOpacity="0" />
                            </radialGradient>
                        </defs>

                        {/* Background rings */}
                        <circle cx={cx} cy={cy} r={R} fill="none" stroke="white" strokeWidth="0.5" strokeDasharray="4 4" opacity="0.1" />
                        <circle cx={cx} cy={cy} r={R * 0.66} fill="none" stroke="white" strokeWidth="0.5" strokeDasharray="4 4" opacity="0.07" />
                        <circle cx={cx} cy={cy} r={R * 0.33} fill="none" stroke="white" strokeWidth="0.5" strokeDasharray="4 4" opacity="0.04" />

                        {/* Axes */}
                        <line x1={cx} y1={cy} x2={cx} y2={cy - R} stroke="white" strokeWidth="0.5" opacity="0.15" />
                        <line x1={cx} y1={cy} x2={cx + R * Math.cos(Math.PI / 6)} y2={cy + R * Math.sin(Math.PI / 6)} stroke="white" strokeWidth="0.5" opacity="0.15" />
                        <line x1={cx} y1={cy} x2={cx - R * Math.cos(Math.PI / 6)} y2={cy + R * Math.sin(Math.PI / 6)} stroke="white" strokeWidth="0.5" opacity="0.15" />

                        {/* Radar Sweep Line */}
                        <line
                            x1={cx} y1={cy} x2={cx} y2={cy - R}
                            stroke="#f43f5e" strokeWidth="1" opacity="0.4"
                        >
                            <animateTransform
                                attributeName="transform"
                                type="rotate"
                                from={`0 ${cx} ${cy}`}
                                to={`360 ${cx} ${cy}`}
                                dur="4s"
                                repeatCount="indefinite"
                            />
                        </line>

                        {/* Sweep glow trail */}
                        <circle cx={cx} cy={cy} r={R * 0.8} fill="url(#sweepGlow)" opacity="0.3">
                            <animateTransform
                                attributeName="transform"
                                type="rotate"
                                from={`0 ${cx} ${cy}`}
                                to={`360 ${cx} ${cy}`}
                                dur="4s"
                                repeatCount="indefinite"
                            />
                        </circle>

                        {/* Radar Polygon */}
                        <polygon
                            points={radarPath}
                            fill="rgba(244, 63, 94, 0.15)"
                            stroke="#f43f5e"
                            strokeWidth="2"
                            className="transition-all duration-1000 ease-out"
                        >
                            {hasThreat && (
                                <animate
                                    attributeName="fill-opacity"
                                    values="0.15;0.35;0.15"
                                    dur="2s"
                                    repeatCount="indefinite"
                                />
                            )}
                        </polygon>

                        {/* Score dots on vertices */}
                        {hasThreat && (
                            <>
                                <circle cx={pIF.x} cy={pIF.y} r="3" fill="#f43f5e" opacity={ifScore > 0 ? 1 : 0.2} />
                                <circle cx={pAE.x} cy={pAE.y} r="3" fill="#f43f5e" opacity={aeScore > 0 ? 1 : 0.2} />
                                <circle cx={pAI.x} cy={pAI.y} r="3" fill="#f43f5e" opacity={aiScore > 0 ? 1 : 0.2} />
                            </>
                        )}

                        {/* Labels */}
                        <text x={cx} y={cy - R - 10} textAnchor="middle" fill="white" fontSize="7" fontWeight="bold" opacity="0.35">ISOLATION FOREST</text>
                        <text x={cx + R * Math.cos(Math.PI / 6) + 5} y={cy + R * Math.sin(Math.PI / 6) + 12} textAnchor="start" fill="white" fontSize="7" fontWeight="bold" opacity="0.35">AUTOENCODER</text>
                        <text x={cx - R * Math.cos(Math.PI / 6) - 5} y={cy + R * Math.sin(Math.PI / 6) + 12} textAnchor="end" fill="white" fontSize="7" fontWeight="bold" opacity="0.35">AI SECURITY</text>
                    </svg>
                ) : (
                    /* Empty State */
                    <div className="flex flex-col items-center gap-4 py-8">
                        <div className="p-4 rounded-full bg-emerald-500/5 border border-emerald-500/10">
                            <ShieldCheck className="w-10 h-10 text-emerald-500/40" />
                        </div>
                        <p className="text-sm text-white/30 font-medium text-center">
                            No threats detected<br />
                            <span className="text-xs text-white/15">Radar is monitoring...</span>
                        </p>
                    </div>
                )}

                <div className="absolute inset-0 bg-gradient-radial from-rose-500/5 to-transparent pointer-events-none" />
            </div>

            {/* Score Footer */}
            <div className="grid grid-cols-3 gap-4 pt-4 border-t border-white/5 relative z-10">
                <div className="text-center space-y-1">
                    <div className="flex items-center justify-center gap-1 text-[8px] font-black text-white/20 uppercase tracking-widest">
                        <Cpu className="w-3 h-3" /> IF
                    </div>
                    <p className={`text-sm font-black transition-colors duration-500 ${ifScore > 0.7 ? 'text-rose-500' : ifScore > 0 ? 'text-amber-400' : 'text-white/60'}`}>
                        {(ifScore * 100).toFixed(0)}%
                    </p>
                </div>
                <div className="text-center space-y-1">
                    <div className="flex items-center justify-center gap-1 text-[8px] font-black text-white/20 uppercase tracking-widest">
                        <Brain className="w-3 h-3" /> AE
                    </div>
                    <p className={`text-sm font-black transition-colors duration-500 ${aeScore > 0.7 ? 'text-rose-500' : aeScore > 0 ? 'text-amber-400' : 'text-white/60'}`}>
                        {(aeScore * 100).toFixed(0)}%
                    </p>
                </div>
                <div className="text-center space-y-1">
                    <div className="flex items-center justify-center gap-1 text-[8px] font-black text-white/20 uppercase tracking-widest">
                        <Fingerprint className="w-3 h-3" /> AI
                    </div>
                    <p className={`text-sm font-black transition-colors duration-500 ${aiScore > 0.7 ? 'text-rose-500' : aiScore > 0 ? 'text-amber-400' : 'text-white/60'}`}>
                        {(aiScore * 100).toFixed(0)}%
                    </p>
                </div>
            </div>

            {/* Background icon */}
            <div className="absolute top-0 right-0 p-4 opacity-[0.03]">
                <Zap className="w-24 h-24 text-rose-500 rotate-12" />
            </div>
        </div>
    );
};
