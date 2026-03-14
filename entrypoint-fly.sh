#!/bin/bash
set -e

# Start Docker daemon in the background with explicit DNS
dockerd-entrypoint.sh --dns 8.8.8.8 --dns 8.8.4.4 &

# Wait for Docker to be ready
echo "Waiting for Docker daemon to start..."
until docker info >/dev/null 2>&1; do
  echo "Still waiting for Docker..."
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

# Run docker-compose
echo "Starting Gorenel services..."
docker-compose up --build
