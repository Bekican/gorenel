import React, { useState, useEffect } from 'react';
import { Key, Plus, Copy, Trash2, Shield, Clock, AlertTriangle, CheckCircle2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { api } from '../api/client';

interface APIKey {
    key: string;
    created_at: string;
    usage_count: number;
    rate_limit: number;
}

export const ApiKeyManager: React.FC = () => {
    const { t } = useTranslation();
    const [keys, setKeys] = useState<APIKey[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [showNewKey, setShowNewKey] = useState<string | null>(null);

    useEffect(() => {
        fetchKeys();
    }, []);

    const fetchKeys = async () => {
        try {
            setLoading(true);
            const data = await api.listAPIKeys();
            setKeys(data);
            setError(null);
        } catch (err) {
            setError(t('common.error_fetch', 'Failed to load API keys'));
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateKey = async () => {
        try {
            const result = await api.createAPIKey();
            setShowNewKey(result.key);
            fetchKeys();
        } catch (err) {
            alert('Failed to create API key');
        }
    };

    const handleDeleteKey = async (key: string) => {
        if (!confirm(t('api_keys_manager.revoke_confirm'))) return;
        try {
            await api.deleteAPIKey(key);
            fetchKeys();
        } catch (err) {
            alert('Failed to delete API key');
        }
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        alert('Copied to clipboard!');
    };

    if (loading) return <div className="p-8 text-center text-gray-400">{t('api_keys_manager.loading')}</div>;

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                        <Shield className="text-blue-500" /> {t('api_keys_manager.title')}
                    </h2>
                    <p className="text-gray-400">{t('api_keys_manager.subtitle')}</p>
                </div>
                <button
                    onClick={handleCreateKey}
                    className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-all"
                >
                    <Plus size={18} /> {t('api_keys_manager.generate')}
                </button>
            </div>

            {/* Onboarding Info Card */}
            <div className="p-1 rounded-[2.5rem] bg-gradient-to-r from-blue-500/20 to-indigo-500/20">
                <div className="bg-[#0A0C10] rounded-[2.4rem] p-6 md:p-8 relative overflow-hidden group">
                    <div className="absolute top-0 right-0 p-8 opacity-5 pointer-events-none group-hover:scale-110 transition-transform duration-1000">
                        <Shield className="w-32 h-32 text-blue-500" />
                    </div>
                    <div className="relative z-10 flex flex-col md:flex-row items-center gap-6">
                        <div className="w-16 h-16 rounded-2xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center shrink-0">
                            <Key className="w-8 h-8 text-blue-400" />
                        </div>
                        <div>
                            <h3 className="text-xl font-black mb-2">{t('api_keys_manager.onboarding_title')}</h3>
                            <p className="text-white/50 text-sm font-medium leading-relaxed max-w-2xl">
                                {t('api_keys_manager.onboarding_desc')}
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            {showNewKey && (
                <div className="bg-green-900/20 border border-green-500/50 p-4 rounded-xl flex items-center justify-between gap-4 animate-in fade-in slide-in-from-top-4">
                    <div className="flex-1">
                        <p className="text-green-400 text-sm font-semibold mb-1">{t('api_keys_manager.success')}</p>
                        <p className="text-white font-mono bg-black/40 p-2 rounded border border-white/10 break-all">{showNewKey}</p>
                        <p className="text-gray-400 text-xs mt-2">{t('api_keys_manager.security_notice')}</p>
                    </div>
                    <div className="flex gap-2">
                        <button onClick={() => copyToClipboard(showNewKey)} className="p-2 bg-green-600 text-white rounded-lg hover:bg-green-500">
                            <Copy size={20} />
                        </button>
                        <button onClick={() => setShowNewKey(null)} className="p-2 text-gray-400 hover:text-white">
                            Close
                        </button>
                    </div>
                </div>
            )}

            {error && (
                <div className="bg-red-900/20 border border-red-500/50 p-4 rounded-xl text-red-400 flex items-center gap-3">
                    <AlertTriangle size={20} /> {error}
                </div>
            )}

            <div className="grid grid-cols-1 gap-4">
                {keys.map((k) => (
                    <div key={k.key} className="bg-zinc-900 border border-white/5 p-5 rounded-3xl hover:border-white/10 transition-all group overflow-hidden relative">
                        {/* Background Decoration */}
                        <div className="absolute -right-4 -top-4 w-24 h-24 bg-blue-500/5 blur-3xl rounded-full pointer-events-none" />
                        
                        <div className="flex flex-col gap-4 relative">
                            <div className="flex justify-between items-start">
                                <div className="flex gap-4">
                                    <div className="p-3 bg-zinc-800 rounded-2xl text-blue-400">
                                        <Key size={24} />
                                    </div>
                                    <div>
                                        <h3 className="text-white font-bold flex items-center gap-2">
                                            Tunnel Key • <span className="text-blue-400">{k.key.substring(0, 12)}...</span>
                                        </h3>
                                        <div className="flex items-center gap-4 mt-1 text-xs text-gray-500">
                                            <span className="flex items-center gap-1">
                                                <Clock size={12} /> {new Date(k.created_at).toLocaleDateString()}
                                            </span>
                                            <span className="flex items-center gap-1">
                                                <CheckCircle2 size={12} className="text-green-500" /> {k.usage_count} Requests
                                            </span>
                                        </div>
                                    </div>
                                </div>
                                <div className="flex gap-2">
                                    <button
                                        onClick={() => handleDeleteKey(k.key)}
                                        className="p-2 hover:bg-red-900/20 text-gray-500 hover:text-red-500 rounded-xl transition-all"
                                        title="Revoke Key"
                                    >
                                        <Trash2 size={18} />
                                    </button>
                                </div>
                            </div>

                            {/* Magic Command Section */}
                            <div className="mt-2 space-y-3">
                                <div className="flex items-center justify-between">
                                    <span className="text-[10px] font-black uppercase tracking-widest text-blue-400 flex items-center gap-2">
                                        <div className="w-1 h-1 rounded-full bg-blue-400 animate-pulse" />
                                        Magic Command / Büyülü Komut
                                    </span>
                                    <div className="flex gap-1 text-[10px]">
                                        {['Windows', 'Linux', 'Mac'].map((os) => (
                                            <button 
                                                key={os}
                                                className="px-2 py-0.5 rounded-md bg-zinc-800 text-gray-400 hover:text-white hover:bg-zinc-700 transition-all"
                                                onClick={(e) => {
                                                    const cmd = os === 'Windows' 
                                                        ? `powershell -ExecutionPolicy ByPass -Command "iwr -useb https://gorenel.site/install.ps1 | iex; gorenel connect --key ${k.key}"`
                                                        : `curl -sSL https://gorenel.site/install.sh | bash -s -- connect --key ${k.key}`;
                                                    copyToClipboard(cmd);
                                                    e.stopPropagation();
                                                }}
                                            >
                                                {os}
                                            </button>
                                        ))}
                                    </div>
                                </div>
                                
                                <div 
                                    className="bg-black/60 border border-white/5 rounded-2xl p-4 font-mono text-[11px] group/cmd cursor-pointer hover:bg-black/80 transition-all relative"
                                    onClick={() => copyToClipboard(`powershell -ExecutionPolicy ByPass -Command "iwr -useb https://gorenel.site/install.ps1 | iex; gorenel connect --key ${k.key}"`)}
                                >
                                    <div className="flex items-center gap-3 text-white/70 overflow-hidden">
                                        <span className="text-blue-500 shrink-0">$</span>
                                        <code className="truncate">powershell -ExecutionPolicy ByPass -Command "iwr -useb https://gorenel.site/install.ps1 | iex; gorenel connect --key {k.key} --port 3000"</code>
                                    </div>
                                    <div className="absolute right-3 top-1/2 -translate-y-1/2 opacity-0 group-hover/cmd:opacity-100 transition-all bg-blue-600 px-2 py-1 rounded-lg text-white font-sans font-bold flex items-center gap-1 shadow-lg shadow-blue-500/20">
                                        <Copy size={12} /> Kopyala
                                    </div>
                                </div>
                                <p className="text-[10px] text-gray-600 italic">
                                    * Bu komut istemciyi otomatik indirir, kurar ve tüneli başlatır. Hiçbir şeye dokunmanıza gerek yok.
                                </p>
                            </div>
                        </div>
                    </div>
                ))}

                {keys.length === 0 && !loading && (
                    <div className="text-center py-12 bg-zinc-900/50 border border-dashed border-white/10 rounded-2xl">
                        <Shield className="mx-auto text-zinc-700 mb-4" size={48} />
                        <p className="text-gray-500">{t('api_keys_manager.empty')}</p>
                    </div>
                )}
            </div>
        </div>
    );
};
