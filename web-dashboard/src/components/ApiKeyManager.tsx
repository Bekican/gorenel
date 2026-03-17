import React, { useState, useEffect } from 'react';
import { Key, Plus, Copy, Trash2, Shield, Clock, AlertTriangle, CheckCircle2 } from 'lucide-react';
import { api } from '../api/client';

interface APIKey {
    key: string;
    created_at: string;
    usage_count: number;
    rate_limit: number;
}

export const ApiKeyManager: React.FC = () => {
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
            setError('Failed to load API keys');
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
        if (!confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) return;
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

    if (loading) return <div className="p-8 text-center text-gray-400">Loading security keys...</div>;

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                        <Shield className="text-blue-500" /> API Keys
                    </h2>
                    <p className="text-gray-400">Manage your secure tunnel access credentials</p>
                </div>
                <button
                    onClick={handleCreateKey}
                    className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-all"
                >
                    <Plus size={18} /> Generate New Key
                </button>
            </div>

            {showNewKey && (
                <div className="bg-green-900/20 border border-green-500/50 p-4 rounded-xl flex items-center justify-between gap-4 animate-in fade-in slide-in-from-top-4">
                    <div className="flex-1">
                        <p className="text-green-400 text-sm font-semibold mb-1">New Key Generated Successfully!</p>
                        <p className="text-white font-mono bg-black/40 p-2 rounded border border-white/10 break-all">{showNewKey}</p>
                        <p className="text-gray-400 text-xs mt-2">Make sure to copy it now. For security reasons, we won't show it again.</p>
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
                    <div key={k.key} className="bg-zinc-900 border border-white/5 p-5 rounded-2xl hover:border-white/10 transition-all group">
                        <div className="flex justify-between items-start">
                            <div className="flex gap-4">
                                <div className="p-3 bg-zinc-800 rounded-xl text-blue-400">
                                    <Key size={24} />
                                </div>
                                <div>
                                    <h3 className="text-white font-semibold flex items-center gap-2">
                                        Tunnel Key • {k.key.substring(0, 12)}...
                                    </h3>
                                    <div className="flex items-center gap-4 mt-2 text-xs text-gray-500">
                                        <span className="flex items-center gap-1">
                                            <Clock size={12} /> {new Date(k.created_at).toLocaleDateString()}
                                        </span>
                                        <span className="flex items-center gap-1">
                                            <CheckCircle2 size={12} className="text-green-500" /> {k.usage_count} Requests
                                        </span>
                                        <span className="px-2 py-0.5 bg-blue-500/10 text-blue-400 rounded-full border border-blue-500/20">
                                            Limit: {k.rate_limit}/min
                                        </span>
                                    </div>
                                </div>
                            </div>
                            <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                <button
                                    onClick={() => copyToClipboard(k.key)}
                                    className="p-2 hover:bg-zinc-800 text-gray-400 hover:text-white rounded-lg transition-colors"
                                    title="Copy Key"
                                >
                                    <Copy size={18} />
                                </button>
                                <button
                                    onClick={() => handleDeleteKey(k.key)}
                                    className="p-2 hover:bg-red-900/20 text-gray-400 hover:text-red-500 rounded-lg transition-colors"
                                    title="Revoke Key"
                                >
                                    <Trash2 size={18} />
                                </button>
                            </div>
                        </div>
                    </div>
                ))}

                {keys.length === 0 && !loading && (
                    <div className="text-center py-12 bg-zinc-900/50 border border-dashed border-white/10 rounded-2xl">
                        <Shield className="mx-auto text-zinc-700 mb-4" size={48} />
                        <p className="text-gray-500">No API keys found. Create one to start tunneling.</p>
                    </div>
                )}
            </div>
        </div>
    );
};
