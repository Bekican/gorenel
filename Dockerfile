# Build stage
FROM docker.io/library/golang:1.25-alpine AS builder

RUN apk add --no-cache git make bash

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build server binary
RUN go build -o gorenel-server ./cmd/server

# Build ALL client binaries for every platform + generate SHA256 checksums
RUN mkdir -p bin && \
    GOOS=linux   GOARCH=amd64 go build -ldflags "-s -w" -o bin/gorenel-linux-amd64       cmd/client/main.go && \
    GOOS=linux   GOARCH=arm64 go build -ldflags "-s -w" -o bin/gorenel-linux-arm64       cmd/client/main.go && \
    GOOS=darwin  GOARCH=amd64 go build -ldflags "-s -w" -o bin/gorenel-darwin-amd64      cmd/client/main.go && \
    GOOS=darwin  GOARCH=arm64 go build -ldflags "-s -w" -o bin/gorenel-darwin-arm64      cmd/client/main.go && \
    GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/gorenel-windows-amd64.exe cmd/client/main.go && \
    GOOS=windows GOARCH=arm64 go build -ldflags "-s -w" -o bin/gorenel-windows-arm64.exe cmd/client/main.go && \
    cd bin && for f in gorenel-*; do sha256sum "$f" > "$f.sha256"; done

# Production stage
FROM public.ecr.aws/docker/library/alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl bash

# Security: Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy the server binary from builder stage
COPY --from=builder /app/gorenel-server .
# Copy client binaries + checksums for download
COPY --from=builder /app/bin ./bin
# .env is optional — env vars are injected via docker-compose
COPY .env* ./

RUN chmod +x gorenel-server

# Create logs directory owned by appuser
RUN mkdir -p logs/batches logs/archives && chown -R appuser:appgroup .

# API, Control, and Proxy ports
EXPOSE 9091 7000 8085

USER appuser

ENTRYPOINT ["./gorenel-server"]
