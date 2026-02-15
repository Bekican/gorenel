import React, { useState } from 'react';
import { Plus, Trash2, Settings2, Globe, Server, ShieldPlus } from 'lucide-react';
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
            setNewRule({ path_pattern: '/*', add_headers: {}, remove_headers: [], replace_path: '' });
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
        <div className="card h-full flex flex-col p-6">
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary-50 rounded-lg">
                        <Settings2 className="w-5 h-5 text-primary-600" />
                    </div>
                    <div>
                        <h3 className="text-lg font-semibold text-neutral-900">Traffic Modification</h3>
                        <p className="text-sm text-neutral-500">Real-time request manipulation</p>
                    </div>
                </div>
                {!isAdding && (
                    <button
                        onClick={() => setIsAdding(true)}
                        className="flex items-center gap-2 bg-neutral-900 text-white px-4 py-2 rounded-xl text-sm font-bold shadow-sm transition-all hover:bg-black active:scale-95"
                    >
                        <Plus className="w-4 h-4" />
                        Add Rule
                    </button>
                )}
            </div>

            <div className="flex-1 overflow-auto space-y-4">
                {isAdding && (
                    <div className="bg-neutral-50 border border-primary-200 rounded-2xl p-6 space-y-4 animate-in fade-in slide-in-from-top-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-1.5">
                                <label className="text-xs font-bold text-neutral-500 uppercase flex items-center gap-1.5">
                                    <Globe className="w-3 h-3" /> Path Pattern
                                </label>
                                <input
                                    type="text"
                                    placeholder="/api/*"
                                    className="w-full px-4 py-2 bg-white border border-neutral-200 rounded-xl text-sm focus:ring-2 focus:ring-primary-500 transition-all outline-none"
                                    value={newRule.path_pattern}
                                    onChange={(e) => setNewRule({ ...newRule, path_pattern: e.target.value })}
                                />
                            </div>
                            <div className="space-y-1.5">
                                <label className="text-xs font-bold text-neutral-500 uppercase flex items-center gap-1.5">
                                    <Server className="w-3 h-3" /> Replace Path (Optional)
                                </label>
                                <input
                                    type="text"
                                    placeholder="/v2/api"
                                    className="w-full px-4 py-2 bg-white border border-neutral-200 rounded-xl text-sm focus:ring-2 focus:ring-primary-500 transition-all outline-none"
                                    value={newRule.replace_path}
                                    onChange={(e) => setNewRule({ ...newRule, replace_path: e.target.value })}
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <label className="text-xs font-bold text-neutral-500 uppercase flex items-center gap-1.5">
                                <ShieldPlus className="w-3 h-3" /> Add Headers
                            </label>
                            <div className="flex gap-2">
                                <input
                                    placeholder="Key"
                                    className="flex-1 px-4 py-2 bg-white border border-neutral-200 rounded-xl text-sm outline-none"
                                    value={headerKey}
                                    onChange={(e) => setHeaderKey(e.target.value)}
                                />
                                <input
                                    placeholder="Value"
                                    className="flex-1 px-4 py-2 bg-white border border-neutral-200 rounded-xl text-sm outline-none"
                                    value={headerValue}
                                    onChange={(e) => setHeaderValue(e.target.value)}
                                />
                                <button
                                    onClick={handleAddHeader}
                                    className="p-2 bg-primary-100 text-primary-700 rounded-xl hover:bg-primary-200 transition-all font-bold text-sm px-4"
                                >
                                    Add
                                </button>
                            </div>
                            <div className="flex flex-wrap gap-2">
                                {Object.entries(newRule.add_headers || {}).map(([k, v]) => (
                                    <span key={k} className="bg-primary-50 text-primary-700 px-2.5 py-1 rounded-lg text-xs font-medium border border-primary-100">
                                        {k}: {v}
                                    </span>
                                ))}
                            </div>
                        </div>

                        <div className="flex items-center gap-3 pt-4 border-t border-neutral-200">
                            <button
                                onClick={handleSaveRule}
                                className="flex-1 bg-primary-600 text-white py-2 rounded-xl text-sm font-bold shadow-md hover:bg-primary-700 transition-all"
                            >
                                Save Modification Rule
                            </button>
                            <button
                                onClick={() => setIsAdding(false)}
                                className="px-6 py-2 bg-white border border-neutral-200 text-neutral-600 rounded-xl text-sm font-bold hover:bg-neutral-50 transition-all"
                            >
                                Cancel
                            </button>
                        </div>
                    </div>
                )}

                {rules.length === 0 && !isAdding ? (
                    <div className="text-center py-12 bg-neutral-50 rounded-2xl border border-neutral-100 border-dashed border-2">
                        <Settings2 className="w-12 h-12 text-neutral-400 mx-auto mb-4" />
                        <p className="text-neutral-600 font-medium mb-1">No modification rules</p>
                        <p className="text-sm text-neutral-400">Add a rule to manipulate traffic in real-time</p>
                    </div>
                ) : (
                    rules.map((rule) => (
                        <div key={rule.id} className="bg-white border border-neutral-200 rounded-2xl p-5 hover:border-primary-300 hover:shadow-sm transition-all group">
                            <div className="flex items-start justify-between">
                                <div className="space-y-3 flex-1">
                                    <div className="flex items-center gap-3">
                                        <span className="bg-primary-50 text-primary-700 font-mono text-sm font-bold px-2 py-0.5 rounded border border-primary-100">
                                            {rule.path_pattern}
                                        </span>
                                        {rule.replace_path && (
                                            <span className="text-neutral-400 text-xs">→ {rule.replace_path}</span>
                                        )}
                                    </div>

                                    <div className="flex flex-wrap gap-2">
                                        {Object.entries(rule.add_headers || {}).map(([k, v]) => (
                                            <div key={k} className="flex items-center gap-1 bg-neutral-50 border border-neutral-100 px-2 py-1 rounded-lg text-[10px] font-bold">
                                                <Plus className="w-3 h-3 text-green-600" />
                                                <span className="text-neutral-500 uppercase">{k}</span>
                                                <span className="text-neutral-300">:</span>
                                                <span className="text-neutral-800">{v}</span>
                                            </div>
                                        ))}
                                        {rule.remove_headers?.map((h) => (
                                            <div key={h} className="flex items-center gap-1 bg-red-50 border border-red-100 px-2 py-1 rounded-lg text-[10px] font-bold">
                                                <Trash2 className="w-3 h-3 text-red-600" />
                                                <span className="text-red-500 line-through">{h}</span>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                                <button
                                    onClick={() => handleDeleteRule(rule.id)}
                                    className="p-2 text-neutral-400 hover:bg-red-50 hover:text-red-500 rounded-xl transition-all"
                                >
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};
