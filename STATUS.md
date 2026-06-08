## Gorenel Status

This repository includes a full tunnel platform:

- **Control plane (WebSocket)**: `wss://gorenel.site/tunnel/connect`
- **HTTP Proxy (tunnel edge)**: `https://<subdomain>.gorenel.site`
- **Monitoring/API**: `https://gorenel.site/api/*`

### Components

- **gorenel-server**: Go server (proxy + monitoring)
- **web-dashboard**: React dashboard (served by Nginx, proxies `/api` to monitoring)
- **redis/postgres/clickhouse**: queues, auth data, analytics
- **ml-engine**: anomaly detection + model stats

### Health checks

- **Main**: `GET /` (served by dashboard)
- **Backend**: `GET /health`

### Multi-region (Fly.io)

- **Primary region** is configured in `fly.toml` (`primary_region = "fra"`).
- **Scale to multiple regions** using Fly Machines (example):

```bash
fly scale count 2 --region fra
fly scale count 2 --region ams
```

### Client region preference

The CLI supports `--region` which sets `Fly-Prefer-Region` for the control-plane WebSocket connection.
Example:

```bash
gorenel start --region fra
```

