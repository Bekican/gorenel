# Build stage
FROM docker.io/library/golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Create bin directory so the production stage can always copy it (even if empty)
RUN mkdir -p bin

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN go build -o gorenel-server ./cmd/server

# Production stage
FROM public.ecr.aws/docker/library/alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl bash

# Security: Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy the server binary from builder stage
COPY --from=builder /app/gorenel-server .
# Copy client binaries for download (from builder stage to ensure it exists)
COPY --from=builder /app/bin ./bin
# .env is optional — env vars are injected via docker-compose
COPY .env* ./

RUN chmod +x gorenel-server

# Create logs directory owned by appuser
RUN mkdir -p logs/batches logs/archives bin && chown -R appuser:appgroup .

# API, Control, and Proxy ports
EXPOSE 9091 7000 8085

USER appuser

ENTRYPOINT ["./gorenel-server"]
