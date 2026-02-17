# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o gorenel-server ./cmd/server/main.go

# Run stage
FROM alpine:latest

# Security: Add CA certificates for HTTPS requests to AI providers
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/gorenel-server .

# Create logs directory
RUN mkdir -p logs/batches logs/archives

# Expose ports
# 7000: Control Port
# 8080: Proxy Port
# 9090: Monitoring Server
EXPOSE 7000 8080 9090

# Standard production environment variables
ENV GO_ENV=production
ENV JWT_SECRET=change_me_in_production

CMD ["./gorenel-server"]
