FROM public.ecr.aws/docker/library/alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl bash

# Security: Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

# Copy ONLY the server binary and necessary files
COPY gorenel-server .
COPY .env .
COPY bin/ ./bin/
# Copy config or other required dirs if needed
COPY configs/ ./configs/

RUN chmod +x gorenel-server

# Create logs directory owned by appuser
RUN mkdir -p logs/batches logs/archives && chown -R appuser:appgroup .

# API, Control, and Proxy ports
EXPOSE 9091 7000 8085

USER appuser

ENTRYPOINT ["./gorenel-server"]
