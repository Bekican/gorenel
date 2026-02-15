import React, { useState } from 'react';
import { Search, Play, Clock, Globe, Shield, Terminal } from 'lucide-react';
import { format } from 'date-fns';
import { type CapturedRequest, api } from '../api/client';

interface TrafficInspectorProps {
    history: CapturedRequest[];
}

export const TrafficInspector: React.FC<TrafficInspectorProps> = ({ history }) => {
    const [selectedId, setSelectedId] = useState<string | null>(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [replaying, setReplaying] = useState<string | null>(null);

    const filteredHistory = history.filter(req =>
        req.path.toLowerCase().includes(searchTerm.toLowerCase()) ||
        req.subdomain.toLowerCase().includes(searchTerm.toLowerCase()) ||
        req.method.toLowerCase().includes(searchTerm.toLowerCase()) ||
        req.status_code.toString().includes(searchTerm)
    );

    const handleReplay = async (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        setReplaying(id);
        try {
            await api.replayRequest(id);
            // Success notification could go here
        } catch (err) {
            console.error('Replay failed:', err);
        } finally {
            setReplaying(null);
        }
    };

    const getStatusBadgeClass = (code: number) => {
        if (code >= 200 && code < 300) return 'bg-green-50 text-green-700 border-green-100';
        if (code >= 400 && code < 500) return 'bg-orange-50 text-orange-700 border-orange-100';
        if (code >= 500) return 'bg-red-50 text-red-700 border-red-100';
        return 'bg-blue-50 text-blue-700 border-blue-100';
    };

    return (
        <div className="card h-full flex flex-col p-0 overflow-hidden">
            <div className="p-6 border-b border-neutral-100">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-primary-50 rounded-lg">
                            <Search className="w-5 h-5 text-primary-600" />
                        </div>
                        <div>
                            <h3 className="text-lg font-semibold text-neutral-900">Traffic Inspector</h3>
                            <p className="text-sm text-neutral-500">Live request history through your tunnels</p>
                        </div>
                    </div>
                </div>

                <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-neutral-400" />
                    <input
                        type="text"
                        placeholder="Filter by path, method, status..."
                        className="w-full pl-10 pr-4 py-2 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:ring-2 focus:ring-primary-500 transition-all outline-none"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
            </div>

            <div className="flex-1 overflow-auto">
                <table className="w-full text-left border-collapse">
                    <thead className="sticky top-0 bg-neutral-50 text-xs font-bold text-neutral-500 uppercase tracking-wider">
                        <tr>
                            <th className="px-6 py-3">Method</th>
                            <th className="px-6 py-3">Status</th>
                            <th className="px-6 py-3">Path</th>
                            <th className="px-6 py-3">Time</th>
                            <th className="px-6 py-3 text-right">Action</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-neutral-100">
                        {filteredHistory.length === 0 ? (
                            <tr>
                                <td colSpan={5} className="px-6 py-12 text-center text-neutral-400">
                                    No requests matching your filters
                                </td>
                            </tr>
                        ) : (
                            filteredHistory.map((req) => (
                                <React.Fragment key={req.id}>
                                    <tr
                                        className={`hover:bg-neutral-50 cursor-pointer transition-colors ${selectedId === req.id ? 'bg-primary-50/30' : ''}`}
                                        onClick={() => setSelectedId(selectedId === req.id ? null : req.id)}
                                    >
                                        <td className="px-6 py-4">
                                            <span className={`text-xs font-bold px-2 py-1 rounded-md ${req.method === 'POST' ? 'bg-blue-50 text-blue-700' :
                                                req.method === 'GET' ? 'bg-emerald-50 text-emerald-700' : 'bg-neutral-100 text-neutral-700'
                                                }`}>
                                                {req.method}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4">
                                            <span className={`text-xs font-medium px-2.5 py-1 rounded-full border ${getStatusBadgeClass(req.status_code)}`}>
                                                {req.status_code}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 truncate max-w-xs font-mono text-sm">
                                            <span className="text-neutral-400 mr-1">{req.subdomain}.</span>
                                            {req.path}
                                        </td>
                                        <td className="px-6 py-4 text-sm text-neutral-500">
                                            {format(new Date(req.timestamp), 'HH:mm:ss')}
                                        </td>
                                        <td className="px-6 py-4 text-right">
                                            <button
                                                onClick={(e) => handleReplay(req.id, e)}
                                                disabled={replaying === req.id}
                                                className="p-1.5 hover:bg-white rounded-lg transition-all text-neutral-400 hover:text-primary-600 border border-transparent hover:border-primary-100 disabled:opacity-50"
                                                title="Replay request"
                                            >
                                                <Play className={`w-4 h-4 ${replaying === req.id ? 'animate-spin' : ''}`} />
                                            </button>
                                        </td>
                                    </tr>

                                    {selectedId === req.id && (
                                        <tr className="bg-neutral-50/50">
                                            <td colSpan={5} className="px-6 py-6 font-mono text-xs">
                                                <div className="grid grid-cols-2 gap-8">
                                                    <div>
                                                        <div className="flex items-center gap-2 mb-3 text-neutral-900 font-bold border-b border-neutral-200 pb-2">
                                                            <div className="p-1 bg-blue-100 rounded">
                                                                <Terminal className="w-3 h-3 text-blue-600" />
                                                            </div>
                                                            Request Headers
                                                        </div>
                                                        <div className="space-y-1">
                                                            {Object.entries(req.req_headers).map(([k, v]) => (
                                                                <div key={k} className="flex gap-2">
                                                                    <span className="text-primary-600 font-bold shrink-0">{k}:</span>
                                                                    <span className="text-neutral-600 break-all">{v.join(', ')}</span>
                                                                </div>
                                                            ))}
                                                        </div>
                                                    </div>
                                                    <div>
                                                        <div className="flex items-center gap-2 mb-3 text-neutral-900 font-bold border-b border-neutral-200 pb-2">
                                                            <div className="p-1 bg-emerald-100 rounded">
                                                                <Shield className="w-3 h-3 text-emerald-600" />
                                                            </div>
                                                            Response Headers
                                                        </div>
                                                        <div className="space-y-1">
                                                            {Object.entries(req.resp_headers).map(([k, v]) => (
                                                                <div key={k} className="flex gap-2">
                                                                    <span className="text-emerald-600 font-bold shrink-0">{k}:</span>
                                                                    <span className="text-neutral-600 break-all">{v.join(', ')}</span>
                                                                </div>
                                                            ))}
                                                        </div>
                                                    </div>
                                                </div>

                                                <div className="mt-6">
                                                    <div className="flex items-center gap-2 mb-2 text-neutral-500 font-bold uppercase text-[10px] tracking-widest">
                                                        Details
                                                    </div>
                                                    <div className="flex gap-4">
                                                        <div className="flex items-center gap-1.5 text-neutral-500">
                                                            <Clock className="w-3 h-3" />
                                                            <span>Duration: {(req.duration / 1000000).toFixed(2)}ms</span>
                                                        </div>
                                                        <div className="flex items-center gap-1.5 text-neutral-500">
                                                            <Globe className="w-3 h-3" />
                                                            <span>Timestamp: {req.timestamp}</span>
                                                        </div>
                                                    </div>
                                                </div>
                                            </td>
                                        </tr>
                                    )}
                                </React.Fragment>
                            ))
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
