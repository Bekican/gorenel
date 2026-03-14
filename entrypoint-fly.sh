#!/bin/bash
set -e

# Start Docker daemon in the background
dockerd-entrypoint.sh &

# Wait for Docker to be ready
echo "Waiting for Docker daemon to start..."
until docker info >/dev/null 2>&1; do
  echo "Still waiting for Docker..."
  sleep 1
done

echo "Docker daemon is ready!"

# Run docker-compose
echo "Starting Gorenel services..."
docker-compose up --build
