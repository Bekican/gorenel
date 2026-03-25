## Incident Runbook

### Fast triage

- Confirm **dashboard**: `GET /` (should return 200)
- Confirm **backend**: `GET /health` (should return 200)
- Confirm **API**: `GET /api/tunnels` (should return 401 if not logged in, 200 if logged in)

### Common symptoms

- **502 from gorenel.site**
  - Check Fly machine state and logs: `fly status`, `fly logs --no-tail`
  - Verify dashboard container can reach backend (`gorenel-dashboard` logs print a “Backend reachable” line)

- **401 loops in dashboard**
  - Clear local session and re-login (UI auto-clears localStorage on 401)

- **Tunnel endpoints not responding**
  - Verify tunnel is registered in `/api/tunnels`
  - Check `gorenel-server` logs for policy denies (KeyAuth/IP/BasicAuth/rate limit)

### Rollback

- Deploy previous image tag from Fly releases.

### Post-incident

- Write a short summary (timeline + root cause + fix + preventive action).

