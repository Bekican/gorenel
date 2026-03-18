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
echo "Starting resilient socat proxy to bridge Fly.io IPv6 to Docker IPv4..."
while true; do
  socat TCP6-LISTEN:4001,fork,reuseaddr TCP4:127.0.0.1:4000
  echo "socat crashed with exit code $?. Restarting in 1s..."
  sleep 1
done &

# Check if we have pre-built images or if we need to build
if [[ "$(docker images -q gorenel-server 2> /dev/null)" == "" ]]; then
  echo "Images not found. Building services sequentially to prevent OOM..."
  docker-compose build gorenel-server
  docker-compose build ml-engine
  docker-compose build gorenel-dashboard
else
  echo "Images found. Skipping build step for fast boot."
fi

echo "Starting Gorenel services..."
docker-compose up --remove-orphans
