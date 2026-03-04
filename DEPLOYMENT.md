# Deployment Guide

This document provides instructions on how to deploy Gorenel in a production environment.

## Prerequisites

- **Docker & Docker Compose**: For containerized deployment.
- **Go 1.24+**: If building from source.
- **Node.js 18+**: If building the dashboard from source.
- **PostgreSQL**: Primary database.
- **Redis**: Rate limiting and caching.
- **ClickHouse**: High-volume analytics.
- **OAuth Credentials**: Google Cloud Console project with OAuth 2.0 Client ID.

## Environment Variables

The following environment variables MUST be configured for production:

| Variable | Description | Required |
| --- | --- | --- |
| `GO_ENV` | Set to `production`. | Yes |
| `JWT_SECRET` | A secure, random string for signing JWT tokens. | Yes |
| `DB_URL` | PostgreSQL connection string. | Yes |
| `REDIS_ADDR` | Redis server address. | Yes |
| `CLICKHOUSE_ADDR` | ClickHouse native port address. | Yes |
| `CLICKHOUSE_USER` | ClickHouse username. | Yes |
| `CLICKHOUSE_PASSWORD`| ClickHouse password. | Yes |
| `GOOGLE_CLIENT_ID` | Google OAuth Client ID. | Yes |
| `GOOGLE_CLIENT_SECRET`| Google OAuth Client Secret. | Yes |
| `GOOGLE_REDIRECT_URL`| Must match the callback URL in Google Console. | Yes |
| `BASE_DOMAIN` | The base domain for tunnels (e.g., `.gorenel.io`). | Yes |
| `ACME_EMAIL` | Email for Let's Encrypt certificate registration. | Yes |

## Deployment Strategy

### Option 1: Docker Compose (Single Node)

1. Clone the repository.
2. Create a `.env` file based on `.env.example` and fill in the production values.
3. Run the following command:
   ```bash
   docker-compose -f docker-compose.yml up -d
   ```

### Option 2: Kubernetes (Helm)

1. Navigate to the `helm/gorenel` directory.
2. Update the `values.yaml` with your production settings (ingress, secrets, etc.).
3. Deploy using Helm:
   ```bash
   helm install gorenel ./helm/gorenel -f values.yaml
   ```

## Security Recommendations

- **TLS/SSL**: Ensure `BASE_DOMAIN` and `ACME_EMAIL` are correctly set for automatic HTTPS.
- **Secrets Management**: Use Kubernetes Secrets or a secure vault instead of plain environment variables where possible.
- **Database Backups**: Implement regular backup procedures for PostgreSQL and ClickHouse.
- **Monitoring**: Use the built-in monitoring port (`:9091`) to scrape metrics for Prometheus/Grafana.

---
© 2026 Gorenel Team
