# 🔷 Gorenel — Intelligent Reverse Proxy with ML-Powered Anomaly Detection

> A production-grade, multi-protocol tunneling system with real-time anomaly detection using dual ML models (Isolation Forest + Autoencoder).

---

## 🏗️ Architecture

```mermaid
graph TB
    subgraph Client
        CLI["Gorenel CLI"]
    end

    subgraph Server["Go Backend (:8080 / :9090)"]
        CP["Control Port\n(TCP :3000)"]
        HP["HTTP Proxy\n(:8080)"]
        MON["Monitoring API\n(:9090)"]
        TM["Tunnel Manager"]
        ES["Event Stream"]
        AS["Anomaly Store"]
        RL["Rate Limiter\n(Sliding Window)"]
    end

    subgraph ML["Python ML Service (:5000)"]
        FE["Feature Engineer"]
        IF["Isolation Forest"]
        AE["Autoencoder"]
        MR["Model Registry"]
    end

    subgraph Dashboard["React Dashboard (:5173)"]
        MC["Metric Cards"]
        RC["Realtime Charts"]
        AM["Anomaly Alerts"]
        MComp["Model Comparison"]
    end

    CLI -- "yamux" --> CP
    CP --> TM
    TM --> HP
    HP --> RL
    HP -- "async" --> ML
    HP --> ES
    ES --> AS
    MON --> AS
    Dashboard --> MON
    FE --> IF
    FE --> AE
    MR --> FE
```

## ⚡ Quick Start

### Prerequisites
- **Go** 1.21+
- **Python** 3.10+ (with pip)
- **Node.js** 18+

### 1. Start the Go Server
```bash
cd gorenel
go run cmd/server/main.go
```

### 2. Start the ML Service
```bash
cd services/ml
pip install -r requirements.txt
python app.py
```

### 3. Start the Dashboard
```bash
cd web-dashboard
npm install
npm run dev
```

### 4. Connect a Client
```bash
go run cmd/client/main.go connect --port 3001 --type http
```

## 🧠 Dual-Model Anomaly Detection

| Feature | Isolation Forest | Autoencoder |
|---------|-----------------|-------------|
| **Type** | Tree-based | Neural Network |
| **Approach** | Isolates outliers via random partitioning | Measures reconstruction error |
| **Speed** | ~1ms inference | ~5ms inference |
| **Strengths** | Point anomalies, small datasets | Complex patterns, temporal drift |

Both models run in parallel. The consensus engine flags anomalies when **any** model detects an issue, giving maximum coverage.

## 📡 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Server health check |
| `/metrics` | GET | System metrics (tunnels, connections, bandwidth) |
| `/analytics` | GET | Real-time analytics snapshot |
| `/api/tunnels` | GET | List active tunnels |
| `/api/anomalies` | GET | Recent anomaly detections |
| `/api/ml/stats` | GET | ML model statistics |
| `/api/login` | POST | User login |
| `/api/register` | POST | User registration |
| `/api/me` | GET | Current user info (JWT required) |

### ML Service API (`:5000`)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | ML service health |
| `/predict` | POST | Single-model prediction |
| `/predict/compare` | POST | Dual-model comparison |
| `/train` | POST | Trigger model training |
| `/stats` | GET | Model statistics |

## 🛡️ Production Hardening (Phase 9)

- ✅ **Structured Logging**: All `log.Printf` replaced with `zap` — JSON output, typed fields, log levels
- ✅ **Graceful Shutdown**: `SIGINT`/`SIGTERM` handlers for clean resource cleanup
- ✅ **Panic Recovery**: Middleware catches panics and returns 500 instead of crashing
- ✅ **Error Boundaries**: React ErrorBoundary components prevent dashboard white-screens

## 🧪 Testing

```bash
# Run all Go tests with race detector
go test -v -race ./...

# Run ML load test
python services/ml/load_tester.py --requests 1000 --concurrency 10
```

## 📁 Project Structure

```
gorenel/
├── cmd/
│   ├── server/main.go          # Server entry point
│   └── client/cmd/start.go     # CLI client
├── internal/
│   ├── server/                  # Core proxy, tunnels, analytics
│   ├── middleware/              # Auth, rate-limit, panic recovery
│   ├── ml/                      # ML client (Go → Python bridge)
│   └── protocol/               # Wire protocol constants
├── services/
│   └── ml/                      # Python ML service
│       ├── app.py               # Flask API
│       ├── model_registry.py    # Dual-model management
│       └── feature_engineering.py
├── web-dashboard/               # React + Vite dashboard
│   └── src/
│       ├── components/          # UI components
│       └── api/client.ts        # API client
└── tests/                       # Stress & integration tests
```

## 📜 License

MIT — Built as an academic research & SaaS demo project.





## Client Command Standard

Use config-first flow:
```bash
go run cmd/client/main.go config set api_key YOUR_API_KEY
go run cmd/client/main.go connect --port 3001 --type http
```
