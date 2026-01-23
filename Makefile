# Gorenel - Makefile
# Production-grade build pipeline

# Değişkenler
BINARY_NAME_CLIENT=gorenel
BINARY_NAME_SERVER=gorenel-server
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w"

# Renkli output için
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help build-client build-server build-all clean test run-server run-client install deps

# Default target
help:
	@echo "${GREEN}Gorenel Build System${NC}"
	@echo ""
	@echo "${YELLOW}Available targets:${NC}"
	@echo "  make build-all       - Build both client and server"
	@echo "  make build-client    - Build client binary"
	@echo "  make build-server    - Build server binary"
	@echo "  make install         - Install client to /usr/local/bin"
	@echo "  make run-server      - Run server in development mode"
	@echo "  make run-client      - Run client in development mode"
	@echo "  make test            - Run all tests"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make deps            - Download dependencies"
	@echo ""

# Dependencies
deps:
	@echo "${YELLOW}📦 Downloading dependencies...${NC}"
	go mod download
	go mod tidy
	@echo "${GREEN}✅ Dependencies installed${NC}"

# Build client (multiple platforms)
build-client:
	@echo "${YELLOW}🔨 Building client...${NC}"
	@mkdir -p bin
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME_CLIENT}-linux-amd64 cmd/client/main.go
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME_CLIENT}-darwin-amd64 cmd/client/main.go
	# macOS ARM64 (M1/M2)
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME_CLIENT}-darwin-arm64 cmd/client/main.go
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME_CLIENT}-windows-amd64.exe cmd/client/main.go
	@echo "${GREEN}✅ Client built successfully${NC}"
	@ls -lh bin/${BINARY_NAME_CLIENT}*

# Build server
build-server:
	@echo "${YELLOW}🔨 Building server...${NC}"
	@mkdir -p bin
	# Linux AMD64 (production target)
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME_SERVER} cmd/server/main.go
	# ARM64 (Oracle Cloud Always Free)
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME_SERVER}-arm64 cmd/server/main.go
	@echo "${GREEN}✅ Server built successfully${NC}"
	@ls -lh bin/${BINARY_NAME_SERVER}*

# Build everything
build-all: deps build-client build-server
	@echo "${GREEN}✅ All binaries built${NC}"

# Install client to system
install: build-client
	@echo "${YELLOW}📦 Installing client...${NC}"
	sudo cp bin/${BINARY_NAME_CLIENT}-$(shell uname -s | tr '[:upper:]' '[:lower:]')-$(shell uname -m | sed 's/x86_64/amd64/') /usr/local/bin/${BINARY_NAME_CLIENT}
	sudo chmod +x /usr/local/bin/${BINARY_NAME_CLIENT}
	@echo "${GREEN}✅ Installed to /usr/local/bin/${BINARY_NAME_CLIENT}${NC}"
	@echo "Run '${BINARY_NAME_CLIENT} --help' to get started"

# Run server in development
run-server:
	@echo "${YELLOW}🚀 Starting server in dev mode...${NC}"
	go run cmd/server/main.go

# Run client in development
run-client:
	@echo "${YELLOW}🚀 Starting client in dev mode...${NC}"
	go run cmd/client/main.go start --port 3000 --verbose

# Run tests
test:
	@echo "${YELLOW}🧪 Running tests...${NC}"
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}✅ Tests completed. Coverage report: coverage.html${NC}"

# Clean build artifacts
clean:
	@echo "${YELLOW}🧹 Cleaning...${NC}"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean
	@echo "${GREEN}✅ Cleaned${NC}"

# Docker build (bonus)
docker-build:
	@echo "${YELLOW}🐳 Building Docker image...${NC}"
	docker build -t gorenel-server:${VERSION} -f Dockerfile.server .
	@echo "${GREEN}✅ Docker image built: gorenel-server:${VERSION}${NC}"

# Development loop (auto-reload)
dev-server:
	@echo "${YELLOW}🔄 Starting development server with auto-reload...${NC}"
	@which air > /dev/null || (echo "${RED}❌ Air not found. Install: go install github.com/cosmtrek/air@latest${NC}" && exit 1)
	air -c .air.server.toml

# Docker commands
docker-build-server:
	@echo "${YELLOW}🐳 Building server Docker image...${NC}"
	docker build -f Dockerfile.server -t gorenel-server:${VERSION} .
	@echo "${GREEN}✅ Server image built: gorenel-server:${VERSION}${NC}"

docker-build-client:
	@echo "${YELLOW}🐳 Building client Docker image...${NC}"
	docker build -f Dockerfile.client -t gorenel-client:${VERSION} .
	@echo "${GREEN}✅ Client image built: gorenel-client:${VERSION}${NC}"

docker-compose-up:
	@echo "${YELLOW}🐳 Starting Docker Compose stack...${NC}"
	docker-compose up -d
	@echo "${GREEN}✅ Stack started${NC}"
	@echo "Services:"
	@echo "  - Gorenel Server: http://localhost:8080"
	@echo "  - Monitoring: http://localhost:9090"
	@echo "  - Prometheus: http://localhost:9091"
	@echo "  - Grafana: http://localhost:3001"
	@echo "  - ClickHouse: http://localhost:8123"

docker-compose-down:
	@echo "${YELLOW}🐳 Stopping Docker Compose stack...${NC}"
	docker-compose down
	@echo "${GREEN}✅ Stack stopped${NC}"

docker-compose-logs:
	docker-compose logs -f gorenel-server

# Kubernetes/Helm commands
helm-lint:
	@echo "${YELLOW}🔍 Linting Helm chart...${NC}"
	helm lint helm/gorenel
	@echo "${GREEN}✅ Helm chart is valid${NC}"

helm-template:
	@echo "${YELLOW}📝 Rendering Helm templates...${NC}"
	helm template gorenel helm/gorenel --values helm/gorenel/values.yaml

helm-install:
	@echo "${YELLOW}☸️  Installing Gorenel via Helm...${NC}"
	helm install gorenel helm/gorenel --values helm/gorenel/values.yaml
	@echo "${GREEN}✅ Gorenel installed${NC}"

helm-upgrade:
	@echo "${YELLOW}☸️  Upgrading Gorenel...${NC}"
	helm upgrade gorenel helm/gorenel --values helm/gorenel/values.yaml
	@echo "${GREEN}✅ Gorenel upgraded${NC}"

helm-uninstall:
	@echo "${YELLOW}☸️  Uninstalling Gorenel...${NC}"
	helm uninstall gorenel
	@echo "${GREEN}✅ Gorenel uninstalled${NC}"

# Load testing
load-test-k6:
	@echo "${YELLOW}🧪 Running k6 load test...${NC}"
	k6 run tests/load/k6-test.js

load-test-ab:
	@echo "${YELLOW}🧪 Running Apache Bench test...${NC}"
	bash tests/load/ab-test.sh

stress-test:
	@echo "${YELLOW}🧪 Running progressive stress test...${NC}"
	bash tests/load/stress-test.sh

# Monitoring
grafana-import-dashboard:
	@echo "${YELLOW}📊 Importing Grafana dashboard...${NC}"
	curl -X POST http://admin:admin@localhost:3001/api/dashboards/db \
		-H "Content-Type: application/json" \
		-d @monitoring/grafana/dashboards/gorenel-overview.json
	@echo "${GREEN}✅ Dashboard imported${NC}"

prometheus-reload:
	@echo "${YELLOW}🔄 Reloading Prometheus config...${NC}"
	curl -X POST http://localhost:9091/-/reload
	@echo "${GREEN}✅ Config reloaded${NC}"

# CI/CD
ci-local:
	@echo "${YELLOW}🔄 Running CI pipeline locally...${NC}"
	act -j test

release:
	@echo "${YELLOW}📦 Creating release...${NC}"
	@read -p "Enter version (e.g., v1.0.0): " version; \
	git tag -a $version -m "Release $version"; \
	git push origin $version
	@echo "${GREEN}✅ Release $version created${NC}"

# Help command
help:
	@echo "${GREEN}Gorenel Build System${NC}"
	@echo ""
	@echo "${YELLOW}Available targets:${NC}"
	@echo "  make build-all           - Build both client and server"
	@echo "  make build-client        - Build client binary"
	@echo "  make build-server        - Build server binary"
	@echo "  make install             - Install client to /usr/local/bin"
	@echo "  make run-server          - Run server in development mode"
	@echo "  make run-client          - Run client in development mode"
	@echo "  make test                - Run all tests"
	@echo "  make clean               - Remove build artifacts"
	@echo "  make deps                - Download dependencies"
	@echo ""
	@echo "${YELLOW}Docker commands:${NC}"
	@echo "  make docker-build-server - Build server Docker image"
	@echo "  make docker-build-client - Build client Docker image"
	@echo "  make docker-compose-up   - Start full stack with Docker Compose"
	@echo "  make docker-compose-down - Stop Docker Compose stack"
	@echo ""
	@echo "${YELLOW}Kubernetes/Helm:${NC}"
	@echo "  make helm-lint           - Validate Helm chart"
	@echo "  make helm-install        - Install via Helm"
	@echo "  make helm-upgrade        - Upgrade Helm release"
	@echo "  make helm-uninstall      - Remove Helm release"
	@echo ""
	@echo "${YELLOW}Load Testing:${NC}"
	@echo "  make load-test-k6        - Run k6 load test"
	@echo "  make load-test-ab        - Run Apache Bench test"
	@echo "  make stress-test         - Progressive stress test"
	@echo ""
	@echo "${YELLOW}Monitoring:${NC}"
	@echo "  make grafana-import-dashboard - Import Grafana dashboard"
	@echo "  make prometheus-reload        - Reload Prometheus config"
	@echo ""