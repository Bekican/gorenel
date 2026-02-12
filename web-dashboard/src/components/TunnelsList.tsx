import React from 'react';
import { Activity, Copy, ExternalLink, Server, Plus } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

interface Tunnel {
  id: string;
  subdomain: string;
  localPort: number;
  publicUrl: string;
  status: 'active' | 'idle' | 'error';
  requestCount: number;
  bandwidth: {
    in: number;
    out: number;
  };
  startedAt: string;
  lastActivity: string;
}

interface TunnelsListProps {
  tunnels: Tunnel[];
  onOpenConnect: () => void;
}

export const TunnelsList: React.FC<TunnelsListProps> = ({ tunnels, onOpenConnect }) => {
  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    // Could add toast notification here
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'badge-success';
      case 'idle':
        return 'badge-warning';
      case 'error':
        return 'badge-error';
      default:
        return 'badge';
    }
  };

  return (
    <div className="card">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary-50 rounded-lg">
            <Server className="w-5 h-5 text-primary-600" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-neutral-900">Active Tunnels</h3>
            <p className="text-sm text-neutral-500">{tunnels.length} tunnel(s) running</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={onOpenConnect}
            className="flex items-center gap-2 bg-primary-600 hover:bg-primary-700 text-white px-4 py-2 rounded-xl text-sm font-bold shadow-sm transition-all active:scale-95"
          >
            <Plus className="w-4 h-4" />
            New Tunnel
          </button>
          <div className="flex items-center gap-2 bg-green-50 px-2.5 py-1 rounded-full border border-green-100">
            <Activity className="w-4 h-4 text-green-600 animate-pulse" />
            <span className="text-xs text-green-700 font-medium">Live</span>
          </div>
        </div>
      </div>

      <div className="space-y-3">
        {tunnels.length === 0 ? (
          <div className="text-center py-12 bg-neutral-50 rounded-lg border border-neutral-100 dashed border-2">
            <Activity className="w-12 h-12 text-neutral-400 mx-auto mb-4" />
            <p className="text-neutral-600 font-medium mb-1">No active tunnels</p>
            <p className="text-sm text-neutral-400">
              Start a tunnel from the CLI to see it here
            </p>
          </div>
        ) : (
          tunnels.map((tunnel) => (
            <div
              key={tunnel.id}
              className="bg-white border border-neutral-200 rounded-xl p-4 hover:shadow-md hover:border-primary-300 transition-all duration-200 group"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-1">
                    <span className="font-mono text-primary-700 font-bold bg-primary-50 px-2 py-0.5 rounded text-sm">
                      {tunnel.subdomain}
                    </span>
                    <span className={getStatusColor(tunnel.status)}>
                      {tunnel.status}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 text-sm text-neutral-500 mt-2">
                    <span className="flex items-center gap-1">
                      <span className="w-1.5 h-1.5 rounded-full bg-neutral-300"></span>
                      localhost:{tunnel.localPort}
                    </span>
                    <span className="text-neutral-300">•</span>
                    <span className="flex items-center gap-1">
                      {tunnel.requestCount.toLocaleString()} requests
                    </span>
                  </div>
                </div>

                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={() => copyToClipboard(tunnel.publicUrl)}
                    className="p-2 hover:bg-neutral-100 rounded-lg transition-colors text-neutral-400 hover:text-neutral-700"
                    title="Copy URL"
                  >
                    <Copy className="w-4 h-4" />
                  </button>
                  <a
                    href={tunnel.publicUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-2 hover:bg-neutral-100 rounded-lg transition-colors text-neutral-400 hover:text-neutral-700"
                    title="Open in new tab"
                  >
                    <ExternalLink className="w-4 h-4" />
                  </a>
                </div>
              </div>

              <div className="flex items-center justify-between text-xs text-neutral-400 pt-3 border-t border-neutral-100 mt-3">
                <span>
                  Started {formatDistanceToNow(new Date(tunnel.startedAt), { addSuffix: true })}
                </span>
                <div className="flex items-center gap-3 font-mono">
                  <span className="flex items-center gap-1 text-emerald-600 bg-emerald-50 px-1.5 py-0.5 rounded">
                    ↓ {formatBytes(tunnel.bandwidth.in)}
                  </span>
                  <span className="flex items-center gap-1 text-blue-600 bg-blue-50 px-1.5 py-0.5 rounded">
                    ↑ {formatBytes(tunnel.bandwidth.out)}
                  </span>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}