import React, { useState } from 'react';
import { Plus, Trash2, Settings2, Globe, Server, ShieldPlus, X, ChevronRight, Hash, Activity, ShieldAlert, Timer } from 'lucide-react';
import { type ModificationRule, api } from '../api/client';

interface ModificationRulesProps {
    rules: ModificationRule[];
    onRulesChange: () => void;
}

export const ModificationRules: React.FC<ModificationRulesProps> = ({ rules, onRulesChange }) => {
    const [isAdding, setIsAdding] = useState(false);
    const [newRule, setNewRule] = useState<Partial<ModificationRule>>({
        path_pattern: '/*',
        add_headers: {},
        remove_headers: [],
        replace_path: '',
        delay_ms: 0,
        status_code: 0,
    });

    const [headerKey, setHeaderKey] = useState('');
    const [headerValue, setHeaderValue] = useState('');

    const handleAddHeader = () => {
        if (!headerKey) return;
        setNewRule({
            ...newRule,
            add_headers: {
                ...(newRule.add_headers || {}),
                [headerKey]: headerValue,
            }
        });
        setHeaderKey('');
        setHeaderValue('');
    };

    const handleSaveRule = async () => {
        if (!newRule.path_pattern) return;
        try {
            await api.addModificationRule(newRule);
            setIsAdding(false);
            setNewRule({ path_pattern: '/*', add_headers: {}, remove_headers: [], replace_path: '', delay_ms: 0, status_code: 0 });
            onRulesChange();
        } catch (err) {
            console.error('Failed to save rule:', err);
        }
    };

    const handleDeleteRule = async (id: string) => {
        try {
            await api.deleteModificationRule(id);
            onRulesChange();
        } catch (err) {
            console.error('Failed to delete rule:', err);
        }
    };

    return (
        <div className="card h-full flex flex-col p-8 space-y-8">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-white/5 rounded-2xl border border-white/10 shadow-[0_0_20px_rgba(255,255,255,0.05)]">
                        <Settings2 className="w-5 h-5 text-white/60" />
                    </div>
                    <div>
                        <h3 className="text-xl font-black">Dynamic Modifiers</h3>
                        <p className="text-sm text-white/40 font-medium">Global traffic manipulation rules</p>
                    </div>
                </div>
                {!isAdding && (
                    <button
                        onClick={() => setIsAdding(true)}
                        className="btn-primary-premium text-sm"
                    >
                        <Plus className="w-4 h-4" /> Add Logic Rule
                    </button>
                )}
            </div>

            <div className="flex-1 space-y-4">
                {isAdding && (
                    <div className="bg-white/[0.02] border border-white/10 rounded-3xl p-8 space-y-8 animate-in fade-in zoom-in-95 duration-300">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                            <div className="space-y-2">
                                <label className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em] flex items-center gap-2">
                                    <Globe className="w-3 h-3" /> Path Trigger
                                </label>
                                <input
                                    type="text"
                                    placeholder="/api/v1/*"
                                    className="w-full px-5 py-4 bg-black border border-white/5 rounded-2xl text-sm font-mono focus:ring-2 focus:ring-primary/20 focus:border-primary/50 transition-all outline-none text-white"
                                    value={newRule.path_pattern}
                                    onChange={(e) => setNewRule({ ...newRule, path_pattern: e.target.value })}
                                />
                            </div>
                            <div className="space-y-2">
                                <label className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em] flex items-center gap-2">
                                    <Server className="w-3 h-3" /> rewrite target (Optional)
                                </label>
                                <input
                                    type="text"
                                    placeholder="/internal/forward"
                                    className="w-full px-5 py-4 bg-black border border-white/5 rounded-2xl text-sm font-mono focus:ring-2 focus:ring-primary/20 focus:border-primary/50 transition-all outline-none text-white"
                                    value={newRule.replace_path}
                                    onChange={(e) => setNewRule({ ...newRule, replace_path: e.target.value })}
                                />
                            </div>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                            <div className="space-y-2">
                                <label className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em] flex items-center gap-2">
                                    <Activity className="w-3 h-3 text-rose-500" /> Chaos Delay (ms)
                                </label>
                                <input
                                    type="number"
                                    placeholder="500"
                                    className="w-full px-5 py-4 bg-black border border-white/5 rounded-2xl text-sm font-mono focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500/50 transition-all outline-none text-white"
                                    value={newRule.delay_ms || ''}
                                    onChange={(e) => setNewRule({ ...newRule, delay_ms: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="space-y-2">
                                <label className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em] flex items-center gap-2">
                                    <ShieldAlert className="w-3 h-3 text-amber-500" /> Status Override
                                </label>
                                <input
                                    type="number"
                                    placeholder="500"
                                    className="w-full px-5 py-4 bg-black border border-white/5 rounded-2xl text-sm font-mono focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500/50 transition-all outline-none text-white"
                                    value={newRule.status_code || ''}
                                    onChange={(e) => setNewRule({ ...newRule, status_code: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                        </div>

                        <div className="space-y-4">
                            <label className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em] flex items-center gap-2">
                                <ShieldPlus className="w-3 h-3" /> header injections
                            </label>
                            <div className="flex gap-4">
                                <input
                                    placeholder="Header Key"
                                    className="flex-1 px-5 py-4 bg-black border border-white/5 rounded-2xl text-sm font-mono outline-none focus:border-white/20"
                                    value={headerKey}
                                    onChange={(e) => setHeaderKey(e.target.value)}
                                />
                                <input
                                    placeholder="Value"
                                    className="flex-1 px-5 py-4 bg-black border border-white/5 rounded-2xl text-sm font-mono outline-none focus:border-white/20"
                                    value={headerValue}
                                    onChange={(e) => setHeaderValue(e.target.value)}
                                />
                                <button
                                    onClick={handleAddHeader}
                                    className="px-6 bg-white/5 border border-white/10 rounded-2xl text-[10px] font-black uppercase tracking-widest hover:bg-white/10 transition-all"
                                >
                                    Inject
                                </button>
                            </div>
                            <div className="flex flex-wrap gap-2">
                                {Object.entries(newRule.add_headers || {}).map(([k, v]) => (
                                    <span key={k} className="flex items-center gap-2 bg-primary/10 text-primary px-3 py-1.5 rounded-full text-[10px] font-black uppercase tracking-widest border border-primary/20">
                                        {k}: {v as string}
                                        <button onClick={() => {
                                            const next = { ...newRule.add_headers };
                                            delete next[k];
                                            setNewRule({ ...newRule, add_headers: next });
                                        }}><X className="w-3 h-3" /></button>
                                    </span>
                                ))}
                            </div>
                        </div>

                        <div className="flex items-center gap-4 pt-8 border-t border-white/5">
                            <button
                                onClick={handleSaveRule}
                                className="flex-1 btn-primary-premium py-4"
                            >
                                Deploy Execution Rule
                            </button>
                            <button
                                onClick={() => setIsAdding(false)}
                                className="px-8 py-4 bg-white/5 border border-white/5 text-white/40 rounded-2xl text-xs font-black uppercase tracking-widest hover:bg-white/10 transition-all"
                            >
                                Abort
                            </button>
                        </div>
                    </div>
                )}

                {rules.length === 0 && !isAdding ? (
                    <div className="text-center py-24 bg-white/[0.01] rounded-[2.5rem] border-2 border-white/5 border-dashed">
                        <Settings2 className="w-12 h-12 text-white/10 mx-auto mb-6" />
                        <h4 className="font-black text-xl text-white/40 mb-2">Zero Modifiers Active</h4>
                        <p className="text-sm text-white/20 font-medium max-w-sm mx-auto">Traffic is flowing without modifications. Define logic rules to intercept and alter packets in real-time.</p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 gap-4">
                        {rules.map((rule) => (
                            <div key={rule.id} className="group relative bg-white/[0.02] border border-white/5 rounded-[2rem] p-6 hover:bg-white/[0.04] hover:border-white/10 transition-all duration-300">
                                <div className="flex items-start justify-between">
                                    <div className="flex-1 space-y-6">
                                        <div className="flex items-center gap-4">
                                            <div className="flex items-center gap-2 px-3 py-1 bg-black border border-white/5 rounded-lg">
                                                <Hash className="w-3 h-3 text-white/20" />
                                                <span className="text-sm font-mono font-black text-primary">{rule.path_pattern}</span>
                                            </div>
                                            {rule.replace_path && (
                                                <>
                                                    <ChevronRight className="w-4 h-4 text-white/10" />
                                                    <div className="flex items-center gap-2 px-3 py-1 bg-black border border-white/5 rounded-lg">
                                                        <span className="text-sm font-mono font-black text-white/40">{rule.replace_path}</span>
                                                    </div>
                                                </>
                                            )}
                                        </div>

                                        <div className="flex flex-wrap gap-2">
                                            {Object.entries(rule.add_headers || {}).map(([k, v]) => (
                                                <div key={k} className="flex items-center gap-2 bg-emerald-500/5 border border-emerald-500/20 px-3 py-1.5 rounded-full text-[10px] font-black uppercase tracking-widest text-emerald-400">
                                                    <Plus className="w-3 h-3" />
                                                    <span>{k}</span>
                                                    <span className="text-white/20">:</span>
                                                    <span className="text-white/60">{v as string}</span>
                                                </div>
                                            ))}
                                            {rule.remove_headers?.map((h) => (
                                                <div key={h} className="flex items-center gap-2 bg-rose-500/5 border border-rose-500/20 px-3 py-1.5 rounded-full text-[10px] font-black uppercase tracking-widest text-rose-400">
                                                    <Trash2 className="w-3 h-3" />
                                                    <span className="line-through opacity-40">{h}</span>
                                                </div>
                                            ))}
                                            {rule.delay_ms && rule.delay_ms > 0 && (
                                                <div className="flex items-center gap-2 bg-rose-400/5 border border-rose-400/20 px-3 py-1.5 rounded-full text-[10px] font-black uppercase tracking-widest text-rose-400">
                                                    <Timer className="w-3 h-3" />
                                                    <span>Delay: {rule.delay_ms}ms</span>
                                                </div>
                                            )}
                                            {rule.status_code && rule.status_code > 0 && (
                                                <div className="flex items-center gap-2 bg-amber-400/5 border border-amber-400/20 px-3 py-1.5 rounded-full text-[10px] font-black uppercase tracking-widest text-amber-400">
                                                    <ShieldAlert className="w-3 h-3" />
                                                    <span>HTTP {rule.status_code}</span>
                                                </div>
                                            )}
                                        </div>
                                    </div>

                                    <div className="flex flex-col items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <button
                                            onClick={() => handleDeleteRule(rule.id)}
                                            className="p-3 bg-rose-500/10 border border-rose-500/20 text-rose-500 rounded-2xl hover:bg-rose-500 hover:text-white transition-all shadow-lg hover:shadow-rose-500/20"
                                            title="Delete Persistence"
                                        >
                                            <Trash2 className="w-4 h-4" />
                                        </button>
                                        <span className="text-[8px] font-black text-white/10 uppercase tracking-widest">Delete</span>
                                    </div>
                                </div>
                                <div className="mt-6 pt-4 border-t border-white/5 flex items-center justify-between">
                                    <span className="text-[10px] font-black text-white/10 uppercase tracking-widest font-mono">ID: {rule.id}</span>
                                    <div className="flex items-center gap-2">
                                        <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" />
                                        <span className="text-[10px] font-black text-emerald-500/60 uppercase tracking-widest">Active Modifier</span>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};
