import React from 'react';
import { ShieldAlert, Cpu, Brain, Zap, Fingerprint } from 'lucide-react';
import type { AnomalyRecord } from '../api/client';

interface ThreatRadarProps {
    latestAnomaly?: AnomalyRecord;
}

export const ThreatRadar: React.FC<ThreatRadarProps> = ({ latestAnomaly }) => {
    // Default values for the radar if no anomaly is selected
    const ifScore = latestAnomaly?.if_score || 0;
    const aeScore = latestAnomaly?.ae_score || 0;
    const aiScore = latestAnomaly?.detected_by?.includes('AI_SECURITY') ? latestAnomaly.anomaly_score : 0;

    // Radar Chart Points calculation
    // Points: 
    // IF: (0, -R * ifScore)
    // AE: (R * aeScore * cos(30), R * aeScore * sin(30))
    // AI: (-R * aiScore * cos(30), R * aiScore * sin(30))
    // R = 100 for max score

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
            <div className="flex items-center gap-4 relative z-10">
                <div className="p-3 bg-rose-500/10 rounded-2xl border border-rose-500/20 shadow-[0_0_20px_rgba(244,63,94,0.1)]">
                    <ShieldAlert className="w-5 h-5 text-rose-500" />
                </div>
                <div>
                    <h3 className="text-xl font-black">Threat Radar</h3>
                    <p className="text-sm text-white/40 font-medium">Model Consensus Consensus</p>
                </div>
            </div>

            <div className="flex-1 flex flex-col items-center justify-center relative">
                {/* SVG Radar */}
                <svg viewBox="0 0 200 220" className="w-full max-w-[280px] drop-shadow-[0_0_30px_rgba(244,63,94,0.15)]">
                    {/* Background Hexagon/Triangle */}
                    <circle cx={cx} cy={cy} r={R} fill="none" stroke="white" strokeWidth="0.5" strokeDasharray="4 4" opacity="0.1" />
                    <circle cx={cx} cy={cy} r={R / 2} fill="none" stroke="white" strokeWidth="0.5" strokeDasharray="4 4" opacity="0.05" />

                    {/* Axes */}
                    <line x1={cx} y1={cy} x2={cx} y2={cy - R} stroke="white" strokeWidth="0.5" opacity="0.2" />
                    <line x1={cx} y1={cy} x2={cx + R * Math.cos(Math.PI / 6)} y2={cy + R * Math.sin(Math.PI / 6)} stroke="white" strokeWidth="0.5" opacity="0.2" />
                    <line x1={cx} y1={cy} x2={cx - R * Math.cos(Math.PI / 6)} y2={cy + R * Math.sin(Math.PI / 6)} stroke="white" strokeWidth="0.5" opacity="0.2" />

                    {/* Radar Polygon */}
                    <polygon
                        points={radarPath}
                        fill="rgba(244, 63, 94, 0.2)"
                        stroke="#f43f5e"
                        strokeWidth="2"
                        className="transition-all duration-1000 ease-out"
                    />

                    {/* Labels */}
                    <text x={cx} y={cy - R - 10} textAnchor="middle" fill="white" fontSize="8" fontWeight="bold" opacity="0.4">ISOLATION FOREST</text>
                    <text x={cx + R * Math.cos(Math.PI / 6) + 5} y={cy + R * Math.sin(Math.PI / 6) + 10} textAnchor="start" fill="white" fontSize="8" fontWeight="bold" opacity="0.4">AUTOENCODER</text>
                    <text x={cx - R * Math.cos(Math.PI / 6) - 5} y={cy + R * Math.sin(Math.PI / 6) + 10} textAnchor="end" fill="white" fontSize="8" fontWeight="bold" opacity="0.4">AI SECURITY</text>
                </svg>

                <div className="absolute inset-0 bg-gradient-radial from-rose-500/5 to-transparent pointer-none" />
            </div>

            <div className="grid grid-cols-3 gap-4 pt-4 border-t border-white/5 relative z-10">
                <div className="text-center space-y-1">
                    <div className="flex items-center justify-center gap-1 text-[8px] font-black text-white/20 uppercase tracking-widest">
                        <Cpu className="w-3 h-3" /> IF
                    </div>
                    <p className={`text-sm font-black ${ifScore > 0.7 ? 'text-rose-500' : 'text-white/60'}`}>{(ifScore * 100).toFixed(0)}%</p>
                </div>
                <div className="text-center space-y-1">
                    <div className="flex items-center justify-center gap-1 text-[8px] font-black text-white/20 uppercase tracking-widest">
                        <Brain className="w-3 h-3" /> AE
                    </div>
                    <p className={`text-sm font-black ${aeScore > 0.7 ? 'text-rose-500' : 'text-white/60'}`}>{(aeScore * 100).toFixed(0)}%</p>
                </div>
                <div className="text-center space-y-1">
                    <div className="flex items-center justify-center gap-1 text-[8px] font-black text-white/20 uppercase tracking-widest">
                        <Fingerprint className="w-3 h-3" /> AI
                    </div>
                    <p className={`text-sm font-black ${aiScore > 0.7 ? 'text-rose-500' : 'text-white/60'}`}>{(aiScore * 100).toFixed(0)}%</p>
                </div>
            </div>

            {/* Background scanner animation */}
            <div className="absolute top-0 right-0 p-4 opacity-5">
                <Zap className="w-24 h-24 text-rose-500 rotate-12" />
            </div>
        </div>
    );
};
