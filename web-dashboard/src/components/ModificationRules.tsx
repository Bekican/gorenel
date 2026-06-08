import React, { useState } from 'react';
import { Plus, Trash2, Settings2, Globe, Server, ShieldPlus, X, ChevronRight, Hash, Activity, ShieldAlert, Timer, Zap } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { type ModificationRule, api } from '../api/client';
import { Button } from './ui/Button';
import { Input } from './ui/Input';

interface ModificationRulesProps {
    rules: ModificationRule[];
    onRulesChange: () => void;
}

export const ModificationRules: React.FC<ModificationRulesProps> = ({ rules, onRulesChange }) => {
    const { t } = useTranslation();
    const [isAdding, setIsAdding] = useState(false);
    const [newRule, setNewRule] = useState<Partial<ModificationRule>>({
        path_pattern: '/*',
        add_headers: {},
        remove_headers: [],
        replace_path: '',
        delay_ms: 0,
        status_code: 0,
        mock_body: '',
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
            setNewRule({ path_pattern: '/*', add_headers: {}, remove_headers: [], replace_path: '', delay_ms: 0, status_code: 0, mock_body: '' });
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
        <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 md:p-8 space-y-6">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-emerald-500/10 rounded-xl border border-emerald-500/15">
                        <Zap className="w-5 h-5 text-emerald-400" />
                    </div>
                    <div>
                        <h3 className="text-lg font-semibold">{t('modification_rules.title')}</h3>
                        <p className="text-sm text-white/35">{t('modification_rules.subtitle')}</p>
                    </div>
                </div>
                {!isAdding && (
                    <Button type="button" variant="primary" size="md" onClick={() => setIsAdding(true)}>
                        <Plus className="w-4 h-4" /> {t('modification_rules.add_btn')}
                    </Button>
                )}
            </div>

            {!isAdding && rules.length === 0 && (
                <div className="bg-white/[0.02] border border-white/[0.06] rounded-xl p-5 flex items-start gap-3">
                    <div className="p-2 bg-violet-500/10 rounded-lg shrink-0">
                        <Settings2 className="w-4 h-4 text-violet-400" />
                    </div>
                    <div>
                        <h4 className="text-sm font-medium text-white/80">{t('modification_rules.onboarding_title')}</h4>
                        <p className="text-xs text-white/35 leading-relaxed mt-0.5">
                            {t('modification_rules.onboarding_desc')}
                        </p>
                    </div>
                </div>
            )}

            <div className="space-y-3">
                {isAdding && (
                    <div className="bg-white/[0.02] border border-white/[0.08] rounded-xl p-6 space-y-5 animate-fade-in-up">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div className="space-y-1.5">
                                <label className="text-[11px] font-medium text-white/30 flex items-center gap-1.5">
                                    <Globe className="w-3 h-3" /> {t('modification_rules.path_trigger')}
                                </label>
                                <Input
                                    type="text"
                                    placeholder="/api/v1/*"
                                    value={newRule.path_pattern}
                                    onChange={(e) => setNewRule({ ...newRule, path_pattern: e.target.value })}
                                    className="font-mono text-sm"
                                />
                                <p className="text-[10px] text-white/20">{t('modification_rules.path_trigger_desc')}</p>
                            </div>
                            <div className="space-y-1.5">
                                <label className="text-[11px] font-medium text-white/30 flex items-center gap-1.5">
                                    <Server className="w-3 h-3" /> {t('modification_rules.rewrite_target')}
                                </label>
                                <Input
                                    type="text"
                                    placeholder="/internal/forward"
                                    value={newRule.replace_path}
                                    onChange={(e) => setNewRule({ ...newRule, replace_path: e.target.value })}
                                    className="font-mono text-sm"
                                />
                            </div>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div className="space-y-1.5">
                                <label className="text-[11px] font-medium text-white/30 flex items-center gap-1.5">
                                    <Activity className="w-3 h-3 text-rose-400" /> {t('modification_rules.chaos_delay')}
                                </label>
                                <Input
                                    type="number"
                                    placeholder="500"
                                    value={newRule.delay_ms || ''}
                                    onChange={(e) => setNewRule({ ...newRule, delay_ms: parseInt(e.target.value) || 0 })}
                                    className="font-mono text-sm"
                                />
                            </div>
                            <div className="space-y-1.5">
                                <label className="text-[11px] font-medium text-white/30 flex items-center gap-1.5">
                                    <ShieldAlert className="w-3 h-3 text-amber-400" /> {t('modification_rules.status_override')}
                                </label>
                                <Input
                                    type="number"
                                    placeholder="500"
                                    value={newRule.status_code || ''}
                                    onChange={(e) => setNewRule({ ...newRule, status_code: parseInt(e.target.value) || 0 })}
                                    className="font-mono text-sm"
                                />
                            </div>
                        </div>

                        <div className="space-y-1.5">
                            <label className="text-[11px] font-medium text-white/30 flex items-center gap-1.5">
                                <Zap className="w-3 h-3 text-violet-400" /> {t('modification_rules.mock_body')}
                            </label>
                            <textarea
                                placeholder='{ "message": "Intercepted by Gorenel Edge" }'
                                className="w-full h-28 px-4 py-3 bg-white/[0.03] border border-white/[0.08] rounded-xl text-xs font-mono focus:ring-1 focus:ring-emerald-500/20 focus:border-emerald-500/30 transition-all outline-none text-white/70 resize-none"
                                value={newRule.mock_body || ''}
                                onChange={(e) => setNewRule({ ...newRule, mock_body: e.target.value })}
                            />
                            <p className="text-[10px] text-white/20">{t('modification_rules.mock_body_desc')}</p>
                        </div>

                        <div className="space-y-2">
                            <label className="text-[11px] font-medium text-white/30 flex items-center gap-1.5">
                                <ShieldPlus className="w-3 h-3" /> {t('modification_rules.header_injections')}
                            </label>
                            <div className="flex gap-2">
                                <Input placeholder="Header Key" value={headerKey} onChange={(e) => setHeaderKey(e.target.value)} className="flex-1 font-mono text-sm" />
                                <Input placeholder="Value" value={headerValue} onChange={(e) => setHeaderValue(e.target.value)} className="flex-1 font-mono text-sm" />
                                <Button type="button" onClick={handleAddHeader} variant="secondary" size="md">Add</Button>
                            </div>
                            <div className="flex flex-wrap gap-1.5">
                                {Object.entries(newRule.add_headers || {}).map(([k, v]) => (
                                    <span key={k} className="flex items-center gap-1.5 bg-emerald-500/[0.08] text-emerald-400 px-2.5 py-1 rounded-md text-[11px] font-mono border border-emerald-500/15">
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

                        <div className="flex items-center gap-2.5 pt-4 border-t border-white/[0.04]">
                            <Button type="button" onClick={handleSaveRule} variant="primary" size="md" className="flex-1">
                                {t('modification_rules.deploy_btn')}
                            </Button>
                            <Button type="button" onClick={() => setIsAdding(false)} variant="outline" size="md">
                                {t('modification_rules.abort_btn')}
                            </Button>
                        </div>
                    </div>
                )}

                {rules.length === 0 && !isAdding ? (
                    <div className="text-center py-16 bg-white/[0.01] rounded-2xl border border-dashed border-white/[0.06]">
                        <Zap className="w-10 h-10 text-white/[0.06] mx-auto mb-4" />
                        <h4 className="font-medium text-white/30 mb-1">{t('modification_rules.zero_rules')}</h4>
                        <p className="text-sm text-white/20 max-w-sm mx-auto">{t('modification_rules.zero_rules_desc')}</p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 gap-3">
                        {rules.map((rule) => (
                            <div key={rule.id} className="group relative bg-white/[0.02] border border-white/[0.06] rounded-xl p-5 hover:bg-white/[0.04] hover:border-white/[0.1] transition-all duration-200">
                                <div className="flex items-start justify-between">
                                    <div className="flex-1 space-y-3">
                                        <div className="flex items-center gap-3 flex-wrap">
                                            <div className="flex items-center gap-1.5 px-2.5 py-1 bg-black/20 border border-white/[0.06] rounded-lg">
                                                <Hash className="w-3 h-3 text-white/20" />
                                                <span className="text-sm font-mono text-emerald-400">{rule.path_pattern}</span>
                                            </div>
                                            {rule.replace_path && (
                                                <>
                                                    <ChevronRight className="w-3.5 h-3.5 text-white/15" />
                                                    <div className="flex items-center gap-1.5 px-2.5 py-1 bg-black/20 border border-white/[0.06] rounded-lg">
                                                        <span className="text-sm font-mono text-white/40">{rule.replace_path}</span>
                                                    </div>
                                                </>
                                            )}
                                        </div>

                                        <div className="flex flex-wrap gap-1.5">
                                            {Object.entries(rule.add_headers || {}).map(([k, v]) => (
                                                <div key={k} className="flex items-center gap-1.5 bg-emerald-500/[0.06] border border-emerald-500/15 px-2 py-0.5 rounded-md text-[10px] font-mono text-emerald-400">
                                                    <Plus className="w-2.5 h-2.5" /> {k}: {v as string}
                                                </div>
                                            ))}
                                            {rule.remove_headers?.map((h) => (
                                                <div key={h} className="flex items-center gap-1.5 bg-rose-500/[0.06] border border-rose-500/15 px-2 py-0.5 rounded-md text-[10px] font-mono text-rose-400">
                                                    <Trash2 className="w-2.5 h-2.5" /> <span className="line-through opacity-60">{h}</span>
                                                </div>
                                            ))}
                                            {rule.delay_ms && rule.delay_ms > 0 && (
                                                <div className="flex items-center gap-1.5 bg-rose-400/[0.06] border border-rose-400/15 px-2 py-0.5 rounded-md text-[10px] font-medium text-rose-400">
                                                    <Timer className="w-2.5 h-2.5" /> {rule.delay_ms}ms
                                                </div>
                                            )}
                                            {rule.status_code && rule.status_code > 0 && (
                                                <div className="flex items-center gap-1.5 bg-amber-400/[0.06] border border-amber-400/15 px-2 py-0.5 rounded-md text-[10px] font-medium text-amber-400">
                                                    <ShieldAlert className="w-2.5 h-2.5" /> HTTP {rule.status_code}
                                                </div>
                                            )}
                                            {rule.mock_body && (
                                                <div className="flex items-center gap-1.5 bg-violet-400/[0.08] border border-violet-400/15 px-2 py-0.5 rounded-md text-[10px] font-medium text-violet-400">
                                                    <Zap className="w-2.5 h-2.5" /> Mock active
                                                </div>
                                            )}
                                        </div>
                                    </div>

                                    <button
                                        onClick={() => handleDeleteRule(rule.id)}
                                        className="p-2 text-white/20 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition-all opacity-0 group-hover:opacity-100"
                                        title="Delete rule"
                                    >
                                        <Trash2 className="w-4 h-4" />
                                    </button>
                                </div>
                                <div className="mt-3 pt-3 border-t border-white/[0.04] flex items-center justify-between text-[10px] text-white/15">
                                    <span className="font-mono">ID: {rule.id}</span>
                                    <div className="flex items-center gap-1.5">
                                        <div className="w-1 h-1 rounded-full bg-emerald-500 shadow-[0_0_4px_rgba(16,185,129,0.5)]" />
                                        <span className="text-emerald-500/50">{t('modification_rules.active_modifier')}</span>
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
