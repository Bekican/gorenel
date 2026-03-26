#!/bin/bash
set -e

# Start Docker daemon
echo "Cleaning up potential stale Docker socket or port bindings..."
rm -f /var/run/docker.sock
# Kill any processes on common ports to avoid "port already allocated"
for port in 4000 7000 7005 8123 5000 6379 5432 9091 8085; do
  fuser -k ${port}/tcp || true
done

echo "Starting Docker daemon (direct)..."
dockerd --dns 1.1.1.1 --dns 8.8.8.8 --mtu 1400 &

# Wait for Docker to be ready
echo "Waiting for Docker daemon to start..."
MAX_RETRIES=120
COUNT=0
until docker info >/dev/null 2>&1; do
  COUNT=$((COUNT + 1))
  if [ $((COUNT % 10)) -eq 0 ]; then
    echo "Still waiting for Docker ($COUNT/$MAX_RETRIES)..."
  fi
  if [ $COUNT -gt $MAX_RETRIES ]; then
    echo "ERROR: Docker daemon failed to start after $MAX_RETRIES seconds."
    # Attempt recovery: kill and try again
    pkill dockerd || true
    sleep 2
    dockerd --dns 1.1.1.1 --dns 8.8.8.8 --mtu 1400 &
    sleep 5
    if docker info >/dev/null 2>&1; then break; fi
    exit 1
  fi
  sleep 1
done

echo "Docker daemon is ready!"

# Verify docker-compose
if ! command -v docker-compose &> /dev/null; then
    echo "docker-compose not found, or not in PATH. Checking alternative..."
    if command -v docker &> /dev/null && docker compose version &> /dev/null; then
        alias docker-compose="docker compose"
        echo "Using 'docker compose' alias."
    else
        echo "ERROR: Docker Compose is missing."
        exit 1
    fi
fi

# Control Port: Fly.io routes external:7000 -> internal_port:7005 (set in fly.toml)
# Docker maps 7005 -> container:7000 (set in docker-compose.yml)
# No socat needed for control port


# Start Gorenel services
echo "Stopping any old containers..."
docker-compose down --remove-orphans || true
# Light cleanup: only remove stopped containers and dangling images (NOT all images/volumes)
docker system prune -f || true

echo "Starting Gorenel services..."
if ! docker-compose up -d --build --remove-orphans; then
    echo "ERROR: Docker Compose failed to start services."
    docker-compose ps
    docker-compose logs --tail=50
    exit 1
fi

# Wait for gorenel-server to be reachable internally
echo "Waiting for Backend (9091) to be reachable from Dashboard..."
for i in {1..30}; do
  # Resolve dashboard container ID dynamically (name differs by compose project name)
  DASHBOARD_CID="$(docker-compose ps -q gorenel-dashboard 2>/dev/null || true)"

  # Check if dashboard container can reach server health endpoint
  if [ -n "$DASHBOARD_CID" ] && docker exec "$DASHBOARD_CID" curl -s http://gorenel-server:9091/health > /dev/null; then
    echo "SUCCESS: Backend is reachable from Dashboard!"
    break
  fi

  if [ -z "$DASHBOARD_CID" ]; then
    echo "Dashboard container not ready yet ($i/30)..."
  else
    echo "Backend not reachable yet ($i/30)... Check logs for 'port already allocated' if this fails."
  fi
  sleep 3
done

# Diagnostic: Check all containers status
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# The script should not exit, so we tail logs or just wait
echo "Gorenel is up! Tailing logs..."
docker-compose logs -f

