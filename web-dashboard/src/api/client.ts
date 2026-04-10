import axios from 'axios';

function defaultApiBaseUrl(): string {
  // VITE_API_URL is preferred (build-time).
  const fromEnv = import.meta.env?.VITE_API_URL;
  const isProd = typeof window !== 'undefined' && 
                 (window.location.hostname === 'gorenel.site' || 
                  window.location.hostname.endsWith('.gorenel.site') ||
                  window.location.hostname.endsWith('.vercel.app') ||
                  window.location.hostname.endsWith('.fly.dev'));

  // If we are in production but the environment variable is pointing to localhost, ignore it.
  if (fromEnv && typeof fromEnv === 'string' && fromEnv.trim() !== '') {
    const trimmed = fromEnv.trim();
    if (isProd && (trimmed.includes('localhost') || trimmed.includes('127.0.0.1'))) {
      // Use relative paths instead of hardcoded local IP in production.
    } else {
      return trimmed;
    }
  }

  // Runtime fallback:
  if (typeof window === 'undefined') return '';
  const host = window.location.hostname;

  // If we are on a tunnel subdomain (e.g. user.gorenel.site), we must hit the main API.
  // We point to gorenel.site, which Vercel will then proxy to api.gorenel.site.
  if (host.endsWith('.gorenel.site') && host !== 'gorenel.site' && host !== 'api.gorenel.site') {
    return 'https://gorenel.site';
  }

  // Same for fly.dev fallbacks
  if (host.endsWith('.fly.dev') && host !== 'gorenel-app.fly.dev') {
    return 'https://gorenel-app.fly.dev';
  }
  
  // Default to relative paths (works with Vercel rewrites or Vite proxy)
  return '';
}


const API_BASE_URL = defaultApiBaseUrl();

export const AUTH_EVENTS = {
  unauthorized: 'gorenel:unauthorized',
} as const;

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

apiClient.interceptors.response.use(
  (res) => res,
  (err) => {
    const status = err?.response?.status;
    if (status === 401) {
      // Let the app react (clear stale local session, stop polling, show login).
      if (typeof window !== 'undefined') {
        window.dispatchEvent(new CustomEvent(AUTH_EVENTS.unauthorized));
      }
    }
    return Promise.reject(err);
  },
);

// Types
export interface HealthStatus {
  status: string;
  uptime: string;
  time: string;
}

export interface UserSession {
  id?: string;
  email: string;
  name?: string;
  api_key?: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterPayload {
  name: string;
  email: string;
  password: string;
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
  policy?: {
    key_auth_enabled?: boolean;
    ip_allowlist_enabled?: boolean;
    basic_auth_enabled?: boolean;
    basic_auth_username?: string;
    https_redirect_enabled?: boolean;
    rate_limit_enabled?: boolean;
    rate_limit_requests?: number;
    rate_limit_window_s?: number;
    path_prefix?: string;
    replace_path_from?: string;
    replace_path_to?: string;
  };
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

export interface TunnelSessionHistory {
  id: string;
  user_id: string;
  subdomain: string;
  tunnel_type: string;
  local_port: number;
  public_url: string;
  started_at: string;
  ended_at?: string | null;
  request_count: number;
  bytes_in: number;
  bytes_out: number;
  avg_rps: number;
}

export interface TunnelHistoryResponse {
  sessions: TunnelSessionHistory[];
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
  detected_by: string;
  if_score?: number;
  ae_score?: number;
}

export interface AnomaliesResponse {
  anomalies: AnomalyRecord[];
  count: number;
}

export interface APIKeyRecord {
  key: string;
  created_at: string;
  usage_count: number;
  rate_limit: number;
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
  is_security_risk?: boolean;
  risk_score?: number;
  risk_reason?: string;
}

export interface ModificationRule {
  id: string;
  path_pattern: string;
  add_headers?: Record<string, string>;
  remove_headers?: string[];
  replace_path?: string;
  delay_ms?: number;
  status_code?: number;
  mock_body?: string;
}

export type ModelStatsResponse = Record<string, ModelStat>;
export interface MLStatsEnvelope {
  stats: ModelStatsResponse;
  active_tunnels: number;
  ml_up: boolean;
  last_prediction_at?: string | null;
}

export interface ReservedSubdomain {
  subdomain: string;
  user_id: string;
  assigned_api_key_hash?: string | null;
  created_at: string;
  last_used_at?: string | null;
}

export interface ReservationsResponse {
  reservations: ReservedSubdomain[];
  count: number;
}

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
  login: async (credentials: LoginCredentials): Promise<{ user?: UserSession; redirect_url?: string }> => {
    const { data } = await apiClient.post<{ user?: UserSession; redirect_url?: string }>('/api/login', credentials);
    return data;
  },

  socialLogin: async (provider: string): Promise<{ redirect_url?: string }> => {
    const { data } = await apiClient.get<{ redirect_url?: string }>(`/api/login?provider=${provider}`);
    return data;
  },

  register: async (userData: RegisterPayload): Promise<{ user?: UserSession }> => {
    const { data } = await apiClient.post<{ user?: UserSession }>('/api/register', userData);
    return data;
  },

  getMe: async (): Promise<{ user?: UserSession }> => {
    const { data } = await apiClient.get<{ user?: UserSession }>('/api/me');
    return data;
  },

  getTunnels: async (): Promise<TunnelsResponse> => {
    const { data } = await apiClient.get<TunnelsResponse>('/api/tunnels');
    return data;
  },

  // Tunnel Policy
  rotateTunnelToken: async (subdomain: string): Promise<{ token: string }> => {
    const { data } = await apiClient.post<{ token: string }>(`/api/tunnel-policy/${encodeURIComponent(subdomain)}/rotate`);
    return data;
  },

  updateTunnelPolicy: async (
    subdomain: string,
    body: {
      key_auth_enabled?: boolean;
      ip_allowlist_enabled?: boolean;
      ip_allowlist?: string[];
      basic_auth_enabled?: boolean;
      basic_auth_username?: string;
      basic_auth_password?: string;
      https_redirect_enabled?: boolean;
      rate_limit_enabled?: boolean;
      rate_limit_requests?: number;
      rate_limit_window_s?: number;
      add_request_headers?: Record<string, string>;
      remove_request_headers?: string[];
      add_response_headers?: Record<string, string>;
      remove_response_headers?: string[];
      path_prefix?: string;
      replace_path_from?: string;
      replace_path_to?: string;
    },
  ): Promise<void> => {
    await apiClient.put(`/api/tunnel-policy/${encodeURIComponent(subdomain)}`, body);
  },

  getTunnelHistory: async (): Promise<TunnelHistoryResponse> => {
    const { data } = await apiClient.get<TunnelHistoryResponse>('/api/tunnels?history=1');
    return data;
  },

  // Reservations
  listReservations: async (): Promise<ReservationsResponse> => {
    const { data } = await apiClient.get<ReservationsResponse>('/api/reservations');
    return data;
  },
  reserveSubdomain: async (subdomain: string): Promise<ReservedSubdomain> => {
    const { data } = await apiClient.post<ReservedSubdomain>('/api/reservations', { subdomain });
    return data;
  },
  releaseSubdomain: async (subdomain: string): Promise<void> => {
    await apiClient.delete(`/api/reservations/${encodeURIComponent(subdomain)}`);
  },
  assignReservationToKey: async (subdomain: string, apiKey: string | null): Promise<void> => {
    await apiClient.put(`/api/reservations/${encodeURIComponent(subdomain)}/assign`, { api_key: apiKey || '' });
  },

  getAnomalies: async (): Promise<AnomaliesResponse> => {
    const { data } = await apiClient.get<AnomaliesResponse>('/api/anomalies');
    return data;
  },

  getMLStats: async (): Promise<MLStatsEnvelope> => {
    const { data } = await apiClient.get<MLStatsEnvelope>('/api/ml/stats');
    // Back-compat: if server returns the old map shape, normalize to envelope.
    const isEnvelope = !!(data && typeof data === 'object' && 'stats' in data);
    if (isEnvelope) return data;
    return {
      stats: (data && typeof data === 'object' ? (data as Record<string, ModelStat>) : {}) || {},
      active_tunnels: 0,
      ml_up: false,
      last_prediction_at: null,
    };
  },

  listAPIKeys: async (): Promise<APIKeyRecord[]> => {
    const { data } = await apiClient.get<APIKeyRecord[]>('/api/keys');
    return data;
  },

  createAPIKey: async () => {
    const { data } = await apiClient.post<{ key: string }>('/api/keys');
    return data;
  },

  deleteAPIKey: async (key: string) => {
    await apiClient.delete('/api/keys', { params: { key } });
  },

  logout: async () => {
    await apiClient.post('/api/logout');
    // Clear any local user data just in case, though the cookie is the main one
    localStorage.removeItem('gorenel_user');
  },

  // Traffic Inspection
  getTrafficHistory: async (): Promise<CapturedRequest[]> => {
    const { data } = await apiClient.get<CapturedRequest[]>('/api/inspector/history');
    return data;
  },

  replayRequest: async (id: string): Promise<unknown> => {
    const { data } = await apiClient.post<unknown>(`/api/inspector/replay?id=${encodeURIComponent(id)}`);
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