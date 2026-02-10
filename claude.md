# Gorenel – Persistent AI Context

## Architecture Rules

### System Overview
Gorenel is a tunneling system (like ngrok) with ML-powered anomaly detection, written in Go (backend) + Python (ML) + React/TS (dashboard).

### Backend (Go)
- **3-Port Architecture:** Control (7000), Proxy (8080), Monitoring (9090)
- **Multiplexing:** Yamux protocol for client-server connections
- **Package Structure:** `internal/server/` for core logic, `internal/protocol/` for constants, `pkg/` for shared utilities
- **Concurrency Pattern:** `sync.RWMutex` for shared state, `sync/atomic` for counters, goroutines for async work
- **Error Handling:** `pkg/errors` has error wrapper; monitoring handlers use `serverErrors.ErrorWrapper`
- **Auth:** JWT-based (`pkg/auth`), middleware in `internal/middleware/RequireAuth`
- **Rate Limiting:** IP/Subdomain-based via `internal/limiter`

### ML Service (Python)
- **Flask** REST API on port 5000
- **Models live in:** `services/ml/models/`
- **Active model:** `IsolationForestAnomalyDetector` (scikit-learn)
- **Skeleton exists:** `AutoencoderAnomalyDetector` (TensorFlow/Keras) — not yet wired into app.py
- **Feature pipeline:** `feature_engineering.py` → extract_features → transform → numpy array
- **Training:** POST `/train` (currently uses synthetic data)
- **Prediction:** POST `/predict` with `{ "data": { method, path, response_time, status_code, ... } }`

### Frontend (React + TypeScript + Vite)
- **Components are lazy-loaded** via `React.lazy()`
- **API client:** `src/api/client.ts` with axios, all endpoints typed
- **Polling:** 5-second interval in `App.tsx` when user is logged in
- **Styling:** Utility classes (not Tailwind — vanilla CSS in `index.css`)

### Naming Conventions
- Go: PascalCase for exports, camelCase for internal
- Files: snake_case in Go, PascalCase for React components
- API paths: `/api/kebab-case`
- Turkish comments are acceptable (this is a Turkish developer's project)

---

## Known Mistakes

### 1. BaseDomain mismatch
- **Problem:** `constants.go` had `.tunnel.local`, simulator used `.gorenel.io`
- **Fix:** Changed BaseDomain to `.gorenel.io`
- **Rule:** Never change BaseDomain without updating ALL consumers (simulator, client, tests)

### 2. Analytics defer placement
- **Problem:** Analytics events were only published for successful requests
- **Fix:** Moved `defer { publishEvent(...) }` to the TOP of `ServeHTTP`, before any early returns
- **Rule:** Always place analytics/telemetry defer blocks at the very start of handlers

### 3. tunnelsHandler signature bug
- **Problem:** Had `**http.Request` (double pointer) — compiled but was never callable
- **Fix:** Replaced with proper `*http.Request` signature  
- **Rule:** Always verify handler signatures match `http.HandlerFunc` or `func(w, r)`

### 4. Duplicate error logging
- **Problem:** Same error logged twice in http_proxy.go
- **Rule:** One log statement per error, at the point of handling

---

## Constraints

### Performance
- ML inference must not block HTTP proxy (use `PredictAsync` with goroutines)
- AnomalyStore capped at 100 records max (in-memory, no persistence needed for MVP)
- Dashboard polling at 5s intervals — don't overload monitoring server

### Security  
- All `/api/` endpoints behind CORS middleware
- Auth endpoints: `/api/login`, `/api/register`, `/api/me` (JWT)
- Rate limiting on all metric/analytics endpoints
- Never expose internal server errors to clients

### Cost / Resources
- No external databases required for MVP (in-memory stores)
- Redis is optional (publisher exists but fails gracefully)
- ML service is lightweight Flask — no GPU required (Isolation Forest is CPU-only)
- Autoencoder will need TensorFlow — heavier dependency

### Academic Requirements
- This is a **3rd year CS project** — code must be explainable and well-documented
- Autoencoder comparison with Isolation Forest is a key deliverable
- Performance benchmarks (inference latency) needed for thesis
- Dashboard must visually demonstrate anomaly detection working

---

## Current State (Feb 2026)

### ✅ Done
- HTTP/TCP/UDP tunneling with Yamux
- JWT auth + rate limiting
- Traffic inspection & modification
- Real-time analytics dashboard
- Isolation Forest ML model
- Anomaly store + `/api/anomalies` endpoint
- AnomalyAlerts React component

### 🔲 Remaining
- Wire Autoencoder into ML service (model exists, not integrated)
- Dual-model comparison endpoint (`/compare`)
- Dashboard: show both models' results side-by-side
- Performance benchmarking (latency measurement)
- Comprehensive test coverage
- Production hardening (graceful shutdown, health checks)
