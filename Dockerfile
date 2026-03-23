FROM public.ecr.aws/docker/library/alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl bash

# Security: Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy the server binary
COPY gorenel-server .
# .env is optional — env vars are injected via docker-compose
COPY .env* ./

RUN chmod +x gorenel-server

# Create logs directory owned by appuser
RUN mkdir -p logs/batches logs/archives && chown -R appuser:appgroup .

# API, Control, and Proxy ports
EXPOSE 9091 7000 8085

USER appuser

ENTRYPOINT ["./gorenel-server"]
