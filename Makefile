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