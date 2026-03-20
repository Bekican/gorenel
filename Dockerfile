# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/server ./cmd/server
COPY cmd/client ./cmd/client
COPY internal ./internal
COPY pkg ./pkg

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o gorenel-server ./cmd/server/main.go

# Build the windows client
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o gorenel-windows-amd64.exe ./cmd/client/main.go

# Run stage
FROM alpine:latest

# Security: Add CA certificates for HTTPS requests to AI providers
RUN apk --no-cache add ca-certificates tzdata curl

# Security: Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy the binary from builder
COPY --from=builder /app/gorenel-server .
COPY --from=builder /app/gorenel-windows-amd64.exe .

# Create logs directory owned by appuser
RUN mkdir -p logs/batches logs/archives && chown -R appuser:appgroup .

# Expose ports
# 7000: Control Port
# 8085: Proxy Port
# 9091: Monitoring Server
EXPOSE 7000 8085 9091

# Security: Run as non-root user
USER appuser

CMD ["./gorenel-server"]
