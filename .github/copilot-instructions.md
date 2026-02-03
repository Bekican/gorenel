# Gorenel Codebase Instructions for AI Agents

## Project Overview

**Gorenel** is a production-grade tunnel system providing secure subdomain-based HTTP tunneling with real-time analytics, event streaming, and monitoring capabilities. It uses a client-server architecture with a TypeScript+Vite frontend dashboard.

### Architecture Layers
- **Control Plane** (Port 7000): Client registration and tunnel management via Yamux multiplexing
- **Data Plane** (Port 8080): HTTP proxy routing tunneled requests to local services
- **Monitoring** (Port 9090): Prometheus metrics and health checks
- **Frontend Dashboard**: React+TypeScript for analytics and tunnel visualization

---

## Critical Architectural Patterns

### 1. **Event-Driven Architecture**
The core pattern is an `EventStream` pub-sub system:
- `EventStream` (in `internal/server/events.go`) publishes `RequestEvent` for every HTTP request
- Multiple `EventConsumer` implementations subscribe: `BatchLogger`, `DataArchiver`, `AnalyticsEngine`
- Events include: subdomain, method, status code, response time, geolocation, session tracking
- **Why**: Decouples concerns and enables extensible observability without modifying core request handler

**Key file**: [internal/server/events.go](internal/server/events.go) - DO NOT bypass the event system; all request tracking flows through here.

### 2. **Yamux Multiplexing for Tunnel Management**
- Clients establish single TCP connection to control port `:7000`
- Connection upgraded to Yamux session for multiplexed streams
- TunnelManager maps subdomains → Yamux sessions
- HTTP proxy routes incoming requests to appropriate tunnel stream

**Key file**: [internal/server/tunnel_manager.go](internal/server/tunnel_manager.go) - Thread-safe RWMutex-protected maps.

### 3. **Protocol-First Communication**
Binary message protocol in [internal/protocol/messages.go](internal/protocol/messages.go):
- Message types: `REGISTER`, `REGISTERED`, `ERROR`, `PING`, `PONG`
- Uses helper functions: `protocol.ReadMessage()`, `protocol.WriteMessage()`
- All control-plane communication must use these helpers, not raw I/O

### 4. **Error Handling Pattern**
Custom `AppError` type in [pkg/errors/errors_definition.go](pkg/errors/errors_definition.go):
- Typed errors: `TypeValidation`, `TypeNotFound`, `TypeUnauthorized`, `TypeInternal`, `TypeConflict`
- Always includes HTTP status code and `FieldErrors` map
- Stack trace support via `StackTrace()` method
- Must be used for all API responses, not generic `error`

---

## Project-Specific Conventions

### Go Code Patterns
- **Naming**: Turkish comments for domain logic (tunnel operations, event handling) alongside English comments for generic patterns
- **Concurrency**: Always use `sync.RWMutex` for read-heavy structures (TunnelManager, RateLimiter). Defer unlock immediately.
- **Logging**: Use injected `*log.Logger` or `zap.Logger` (configured in [pkg/logger/logger.go](pkg/logger/logger.go)), never `fmt.Print`
- **Interfaces**: Define consumers via interfaces (`EventConsumer`, rate limit behavior) for testability

### Testing Strategy
- **Unit tests**: In `tests/unit/` using table-driven patterns with `testify/assert`
- **Integration tests**: In `tests/integration/` with helper in [tests/test_helpers.go](tests/test_helpers.go)
- **Test server creation**: Use `helper.CreateTestServer()` and `helper.CreateYamuxPair()`
- **Run tests**: `make test` (includes coverage report) or `make test-unit`
- Tests must be tagged with `if testing.Short() { t.Skip() }` for CI optimization

### TypeScript/React Frontend Patterns
- Located in [web-dashboard/](web-dashboard/) - React 19 + TypeScript + Vite
- ESLint configured, React Compiler enabled
- Use API client in [src/api/client.ts](web-dashboard/src/api/client.ts) for backend communication
- Components: RealtimeChart, MetricCard, TunnelsList, GeoMap

---

## Build & Development Workflows

### Building
```bash
make build-all       # Builds client & server for multiple platforms
make build-server    # Linux AMD64 (production), ARM64 (Oracle Cloud)
make build-client    # Linux, macOS (Intel/ARM), Windows
```
Binary locations: `bin/gorenel*` (client), `bin/gorenel-server*` (server)

### Running
```bash
make run-server      # go run cmd/server/main.go
make run-client      # go run cmd/client/main.go start --port 3000 --verbose
```

### Testing & Validation
```bash
make test            # Run all with coverage.html report
make test-unit       # Unit tests only
```

### Deployment
- **Docker**: [Dockerfile.server](Dockerfile.server), [Dockerfile.client](Dockerfile.client)
- **Compose**: [docker-compose.yml](docker-compose.yml) includes server, Prometheus, Grafana
- **Kubernetes**: Helm chart in [helm/gorenel/](helm/gorenel/) with HA config (3+ replicas, pod anti-affinity, autoscaling)

---

## Key Dependencies & External Integration Points

### Go Dependencies
- **hashicorp/yamux**: TCP multiplexing for tunnel streams
- **spf13/cobra**: CLI framework for client commands
- **go.uber.org/zap**: Structured logging
- **stretchr/testify**: Testing assertions
- **google/uuid**: Subdomain generation

### External Services
- **Geolocation Service**: [internal/server/geolocation.go](internal/server/geolocation.go) - looks up IP geolocation
- **Prometheus Monitoring**: [monitoring/prometheus.yml](monitoring/prometheus.yml) - scrapes metrics endpoint
- **Rate Limiting**: Token bucket algorithm in [internal/server/rate_limiter.go](internal/server/rate_limiter.go) for per-subdomain limits

### Protocol Details
- Base domain: `tunnel.local` (configurable in [internal/protocol/constants.go](internal/protocol/constants.go))
- Control port: `:7000`, Proxy port: `:8080`, Monitoring: `:9090`
- Timeouts: Handshake 10s, Read/Write 30s

---

## Common Pitfalls & Anti-Patterns

❌ **DO NOT**:
- Bypass EventStream for request tracking (always publish events)
- Use direct file I/O for logs (use BatchLogger or Archiver via EventStream)
- Create new HTTP handlers without error wrapping in AppError
- Use raw Yamux stream without protocol message helpers
- Mix sync primitives (use consistent RWMutex pattern)

✅ **DO**:
- Add new event consumers by implementing `EventConsumer` interface and calling `es.Subscribe()`
- Always close resources: `defer closer.Close()` for streams, loggers, archivers
- Test integration flows with Yamux pairs (see tunnel_flow_test.go example)
- Use `protocol.ReadMessage`/`WriteMessage` for all control plane I/O
- Return typed `AppError` from handlers for consistent API responses

---

## Adding Features: Common Patterns

### Add New Event Consumer
1. Implement `EventConsumer` interface with `Consume(*RequestEvent)` and `Name()` methods
2. Register in [cmd/server/main.go](cmd/server/main.go): `eventStream.Subscribe(myConsumer)`
3. Subscribe in init phase (before accepting connections)

### Add New Tunnel Feature
1. Define message type in [internal/protocol/constants.go](internal/protocol/constants.go)
2. Implement read/write in [internal/protocol/messages.go](internal/protocol/messages.go)
3. Handle in appropriate handler (auth in [internal/server/auth.go](internal/server/auth.go), routing in [internal/server/http_proxy.go](internal/server/http_proxy.go))
4. Return `AppError` on failures
5. Add integration test in `tests/integration/`

### Add Monitoring Metrics
1. Publish events through EventStream (events automatically tracked)
2. Or expose Prometheus endpoint handler in monitoring package
3. Add dashboard in [helm/gorenel/templates/](helm/gorenel/templates/) or web-dashboard components

---

## Version & Build Metadata

- **Go Version**: 1.24.12 (toolchain specified in go.mod)
- **Build Variables**: `VERSION` (git tag) and `BUILD_TIME` injected via `-ldflags` in Makefile
- **Multi-platform**: Cross-compile to Linux (AMD64/ARM64), macOS (Intel/ARM), Windows

---

## Quick Reference: File Navigation

| Purpose | File(s) |
|---------|---------|
| Server entry point | [cmd/server/main.go](cmd/server/main.go) |
| Request event model | [internal/server/events.go](internal/server/events.go) |
| Tunnel lifecycle | [internal/server/tunnel_manager.go](internal/server/tunnel_manager.go) |
| HTTP routing | [internal/server/http_proxy.go](internal/server/http_proxy.go) |
| Authentication | [internal/server/auth.go](internal/server/auth.go) |
| Error definitions | [pkg/errors/errors_definition.go](pkg/errors/errors_definition.go) |
| Logger setup | [pkg/logger/logger.go](pkg/logger/logger.go) |
| Protocol messages | [internal/protocol/messages.go](internal/protocol/messages.go) |
| Integration tests | [tests/integration/tunnel_flow_test.go](tests/integration/tunnel_flow_test.go) |
| Frontend API | [web-dashboard/src/api/client.ts](web-dashboard/src/api/client.ts) |
