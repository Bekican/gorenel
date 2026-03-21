FROM alpine:latest

# Install Docker and dependencies
RUN apk --no-cache add ca-certificates tzdata curl bash docker docker-compose socat

# Security: Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy the ENTIRE project into the machine
# This is necessary because docker-compose --build runs INSIDE the machine
COPY . .

# Ensure binaries are executable
RUN chmod +x gorenel-server entrypoint-fly.sh

# Create logs directory owned by appuser
RUN mkdir -p logs/batches logs/archives && chown -R appuser:appgroup .

# Expose ports
EXPOSE 7000 4001 9091

# entrypoint-fly.sh MUST run as root to start dockerd
USER root

ENTRYPOINT ["./entrypoint-fly.sh"]
