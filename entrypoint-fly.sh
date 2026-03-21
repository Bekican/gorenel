#!/bin/bash
set -e

# Start Docker daemon
echo "Starting Docker daemon (direct)..."
dockerd --dns 8.8.8.8 --dns 8.8.4.4 --mtu 1420 &

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

# Start socat IPv6 to IPv4 proxy with a resilient restart loop
echo "Starting resilient socat proxies to bridge Fly.io IPv6 to Docker IPv4..."
while true; do
  # Bridge external IPv6:4001 to internal Dashboard IPv4:4000
  socat TCP6-LISTEN:4001,fork,reuseaddr,bind=[::] TCP4:127.0.0.1:4000
  sleep 1
done &

# Control Port (7000) bridge - Points to local 7005 (Docker host port)
while true; do
  # Bridge external IPv6:7000 to internal Control Port IPv4:7005
  socat TCP6-LISTEN:7000,fork,reuseaddr,bind=[::] TCP4:127.0.0.1:7005
  sleep 1
done &

# Start Gorenel services (Clean start to avoid stale images)
echo "Cleaning old containers and images..."
docker-compose down || true
docker system prune -f
echo "Starting Gorenel services..."
docker-compose up -d --build --remove-orphans

# The script should not exit, so we tail logs or just wait
echo "Gorenel is up! Tailing logs..."
docker-compose logs -f
