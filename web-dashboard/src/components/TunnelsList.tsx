import React from 'react';
import { Activity, Copy, ExternalLink } from 'lucide-react';
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
}

export const TunnelsList: React.FC<TunnelsListProps> = ({ tunnels }) => {
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
        <div>
          <h3 className="text-lg font-semibold text-white">Active Tunnels</h3>
          <p className="text-sm text-dark-400">{tunnels.length} tunnel(s) running</p>
        </div>
        <div className="flex items-center gap-2">
          <Activity className="w-5 h-5 text-green-400 animate-pulse" />
          <span className="text-sm text-green-400 font-medium">Live</span>
        </div>
      </div>

      <div className="space-y-3">
        {tunnels.length === 0 ? (
          <div className="text-center py-12">
            <Activity className="w-12 h-12 text-dark-600 mx-auto mb-4" />
            <p className="text-dark-400 mb-2">No active tunnels</p>
            <p className="text-sm text-dark-500">
              Start a tunnel from the CLI to see it here
            </p>
          </div>
        ) : (
          tunnels.map((tunnel) => (
            <div
              key={tunnel.id}
              className="bg-dark-900 border border-dark-700 rounded-lg p-4 hover:border-primary-500/50 transition-colors duration-200"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-mono text-primary-400 font-semibold">
                      {tunnel.subdomain}
                    </span>
                    <span className={getStatusColor(tunnel.status)}>
                      {tunnel.status}
                    </span>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-dark-400">
                    <span>localhost:{tunnel.localPort}</span>
                    <span>•</span>
                    <span>{tunnel.requestCount} requests</span>
                  </div>
                </div>
                
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => copyToClipboard(tunnel.publicUrl)}
                    className="p-2 hover:bg-dark-700 rounded-lg transition-colors"
                    title="Copy URL"
                  >
                    <Copy className="w-4 h-4 text-dark-400 hover:text-white" />
                  </button>
                  <a
                    href={tunnel.publicUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-2 hover:bg-dark-700 rounded-lg transition-colors"
                    title="Open in new tab"
                  >
                    <ExternalLink className="w-4 h-4 text-dark-400 hover:text-white" />
                  </a>
                </div>
              </div>

              <div className="flex items-center justify-between text-xs text-dark-500">
                <span>
                  Started {formatDistanceToNow(new Date(tunnel.startedAt), { addSuffix: true })}
                </span>
                <div className="flex items-center gap-4">
                  <span>↓ {formatBytes(tunnel.bandwidth.in)}</span>
                  <span>↑ {formatBytes(tunnel.bandwidth.out)}</span>
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