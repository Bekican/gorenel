import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Types
export interface HealthStatus {
  status: string;
  uptime: string;
  time: string;
}

export interface Metrics {
  tunnels: {
    active_count: number;
  };
  requests: {
    total: number;
    active_connections: number;
  };
  bandwidth: {
    bytes_in: number;
    bytes_out: number;
  };
  websocket?: {
    connections: number;
    messages: number;
  };
  system: {
    goroutines: number;
    memory_alloc: string;
    memory_sys: string;
  };
  uptime_seconds: number;
}

export interface AnalyticsSnapshot {
  timestamp: string;
  total_requests: number;
  avg_response_time_ms: number;
  top_paths: Array<{ key: string; count: number }>;
  top_countries: Array<{ key: string; count: number }>;
  top_user_agents: Array<{ key: string; count: number }>;
  status_code_distribution: Record<number, number>;
  time_series: Array<{
    timestamp: string;
    requests: number;
    bytes_in: number;
    bytes_out: number;
    avg_latency_ms: number;
  }>;
}

export interface SystemInfo {
  version: string;
  go_version: string;
  platform: string;
  start_time: string;
}

export interface Tunnel {
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

export interface TunnelsResponse {
  tunnels: Tunnel[];
  count: number;
}

export interface AnomalyRecord {
  id: string;
  timestamp: string;
  subdomain: string;
  method: string;
  path: string;
  client_ip: string;
  anomaly_score: number;
}

export interface AnomaliesResponse {
  anomalies: AnomalyRecord[];
  count: number;
}

export interface ModelResult {
  model: string;
  is_anomaly: boolean;
  anomaly_score: number;
  prediction: string;
  inference_ms: number;
}

export interface ConsensusResult {
  any_anomaly: boolean;
  all_agree: boolean;
  flagged_by: string[];
  models_compared: number;
}

export interface MLComparisonResponse {
  models: Record<string, ModelResult>;
  consensus: ConsensusResult;
}

export interface ModelStat {
  is_trained: boolean;
  total_predictions: number;
  total_anomalies: number;
  avg_inference_ms: number;
}

export interface CapturedRequest {
  id: string;
  subdomain: string;
  method: string;
  path: string;
  req_headers: Record<string, string[]>;
  req_body: string; // Base64 or string
  resp_headers: Record<string, string[]>;
  resp_body: string;
  status_code: number;
  timestamp: string;
  duration: number; // in nanoseconds
  ai_metadata?: AIMetadata;
}

export interface AIMetadata {
  model: string;
  provider: string;
  prompt: string;
  completion: string;
  tokens: {
    prompt: number;
    completion: number;
    total: number;
  };
}

export interface ModificationRule {
  id: string;
  path_pattern: string;
  add_headers?: Record<string, string>;
  remove_headers?: string[];
  replace_path?: string;
  delay_ms?: number;
  status_code?: number;
}

export type ModelStatsResponse = Record<string, ModelStat>;

// API Functions
export const api = {
  // Health check
  getHealth: async (): Promise<HealthStatus> => {
    const { data } = await apiClient.get<HealthStatus>('/health');
    return data;
  },

  // System metrics
  getMetrics: async (): Promise<Metrics> => {
    const { data } = await apiClient.get<Metrics>('/metrics');
    return data;
  },

  // Analytics data
  getAnalytics: async (): Promise<AnalyticsSnapshot> => {
    const { data } = await apiClient.get<AnalyticsSnapshot>('/api/analytics/realtime');
    return data;
  },

  // System info
  getInfo: async (): Promise<SystemInfo> => {
    const { data } = await apiClient.get<SystemInfo>('/info');
    return data;
  },

  // Auth Functions
  login: async (credentials: any) => {
    const { data } = await apiClient.post('/api/login', credentials);
    return data;
  },

  getMe: async () => {
    const { data } = await apiClient.get('/api/me');
    return data;
  },

  getTunnels: async (): Promise<TunnelsResponse> => {
    const { data } = await apiClient.get<TunnelsResponse>('/api/tunnels');
    return data;
  },

  getAnomalies: async (): Promise<AnomaliesResponse> => {
    const { data } = await apiClient.get<AnomaliesResponse>('/api/anomalies');
    return data;
  },

  getMLStats: async (): Promise<ModelStatsResponse> => {
    const { data } = await apiClient.get<ModelStatsResponse>('/api/ml/stats');
    return data;
  },

  logout: async () => {
    // We can clear cookies here or handle server-side
    document.cookie = "auth_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
  },

  // Traffic Inspection
  getTrafficHistory: async (): Promise<CapturedRequest[]> => {
    const { data } = await apiClient.get<CapturedRequest[]>('/api/inspector/history');
    return data;
  },

  replayRequest: async (id: string): Promise<any> => {
    const { data } = await apiClient.post(`/api/inspector/replay?id=${id}`);
    return data;
  },

  // Traffic Modification Rules
  getModificationRules: async (): Promise<ModificationRule[]> => {
    const { data } = await apiClient.get<ModificationRule[]>('/api/inspector/rules');
    return data;
  },

  addModificationRule: async (rule: Partial<ModificationRule>): Promise<ModificationRule> => {
    const { data } = await apiClient.post<ModificationRule>('/api/inspector/rules', rule);
    return data;
  },

  deleteModificationRule: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/inspector/rules?id=${id}`);
  },

  // Trace Sharing
  shareTrace: async (id: string): Promise<{ share_id: string; url: string }> => {
    const { data } = await apiClient.post<{ share_id: string; url: string }>(`/api/shares?id=${id}`);
    return data;
  },

  getSharedTrace: async (shareId: string): Promise<CapturedRequest> => {
    const { data } = await apiClient.get<CapturedRequest>(`/api/shares/${shareId}`);
    return data;
  },
};

// WebSocket for real-time updates
export class RealtimeClient {
  private ws: WebSocket | null = null;
  private reconnectInterval = 3000;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  private url: string;
  private onMessage: (data: unknown) => void;

  constructor(url: string, onMessage: (data: unknown) => void) {
    this.url = url;
    this.onMessage = onMessage;
  }

  connect() {
    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log("WebSocket connected");
        if (this.reconnectTimer) {
          clearTimeout(this.reconnectTimer);
          this.reconnectTimer = null;
        }
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.onMessage(data);
        } catch (error) {
          console.error("Failed to parse WebSocket message:", error);
        }
      };

      this.ws.onerror = (error) => {
        console.error("WebSocket error:", error);
      };

      this.ws.onclose = () => {
        console.log("WebSocket disconnected, reconnecting...");
        this.reconnect();
      };
    } catch (error) {
      console.error("Failed to connect WebSocket:", error);
      this.reconnect();
    }
  }

  private reconnect() {
    if (this.reconnectTimer) return;

    this.reconnectTimer = setTimeout(() => {
      console.log("Attempting to reconnect...");
      this.connect();
    }, this.reconnectInterval);
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  send(data: unknown) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }
}


export default apiClient;