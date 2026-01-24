import axios from 'axios'
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:9090'

const apiClient = axios.create({
    baseURL : API_BASE_URL,
    timeout : 10000,
    headers : {
        'Content-Type' : 'application/json',
    },
});

//typelar
export interface HealthStatus{
    status : string,
    uptime : string,
    time : string,
}

export interface Metrics{
    tunnels : {
        active_count : number;
    };
    requestes: {
        total : number;
        active_connections : number;
    };
    bandwidth : {
        bytes_in : number;
        bytes_out : number;
    };
    websocket ?: {
        connections : number;
        messages : number;
    };
    system : {
        goroutine : number;
        memory_alloc : string;
        memory_sys : string;
    }
    uptime_seconds : number;
}

//snapshot
export interface AnalyticsSnapshot{
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



export interface SystemInfo{
    version : string;
    go_version : string;
    platform : string;
    avg_latency_ms : number;
}


//api işlemleri
export const api = {
    //api sağlığı -> health
    getHealth : async(): Promise<HealthStatus> => {
        const {data} = await apiClient.get<HealthStatus>('/health')
        return data;
    },

    //system metrics
    getMetrics: async(): Promise<Metrics> => {
        const{data} = await apiClient.get<Metrics>('/metrics');
        return data;
    },

    //analytics verilerin
    getAnalytics : async(): Promise<AnalyticsSnapshot>=>{
        const{data} = await apiClient.get<AnalyticsSnapshot>('/api/analytics/realtime');
        return data;
    },

    //sistem bilgisi
    getInfo : async(): Promise<SystemInfo> =>{
        const {data} = await apiClient.get<SystemInfo>('/info');
        return data;
    },
};



//web-socket
export class RealtimeClient {
  private ws: WebSocket | null = null;
  private reconnectInterval = 3000;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(private url: string, private onMessage: (data: any) => void) {}

  connect() {
    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
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
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected, reconnecting...');
        this.reconnect();
      };
    } catch (error) {
      console.error('Failed to connect WebSocket:', error);
      this.reconnect();
    }
  }

  private reconnect() {
    if (this.reconnectTimer) return;
    
    this.reconnectTimer = setTimeout(() => {
      console.log('Attempting to reconnect...');
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

  send(data: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }
}

export default apiClient;
