FROM alpine:latest

# Security: Add CA certificates for HTTPS requests to AI providers
RUN apk --no-cache add ca-certificates tzdata curl

# Security: Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy the PRE-BUILT binaries from the context
COPY gorenel-server .
COPY web-dashboard/public/downloads/gorenel-windows-amd64.exe .

# Create logs directory owned by appuser
RUN mkdir -p logs/batches logs/archives && chown -R appuser:appgroup .

# Expose ports
EXPOSE 7000 8085 9091

# Security: Run as non-root user
USER appuser

CMD ["./gorenel-server"]
