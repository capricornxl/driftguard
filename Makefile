.PHONY: all build test clean docker docker-up docker-down lint coverage help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=driftguard
BINARY_UNIX=$(BINARY_NAME)_unix

# Main package
MAIN_PACKAGE=./cmd/main.go

# Build flags
VERSION?=0.1.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Docker
DOCKER_IMAGE=driftguard
DOCKER_TAG?=$(VERSION)

all: clean deps lint test build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: bin/$(BINARY_NAME)"

## build-linux: Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_UNIX) $(MAIN_PACKAGE)

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## benchmark: Run benchmark tests
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem -run=^a ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		$(GOTEST) -v ./...; \
	fi

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

## tidy: Tidy modules
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

## docker: Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .

## docker-up: Start Docker Compose stack
docker-up:
	@echo "Starting Docker Compose stack..."
	docker-compose up -d
	@echo "Stack started. Access:"
	@echo "  - API: http://localhost:8080"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Grafana: http://localhost:3000 (admin/admin123)"

## docker-down: Stop Docker Compose stack
docker-down:
	@echo "Stopping Docker Compose stack..."
	docker-compose down

## docker-logs: View Docker Compose logs
docker-logs:
	docker-compose logs -f

## run: Run the application locally
run:
	@echo "Running application..."
	$(GOCMD) run $(MAIN_PACKAGE)

## install: Install the binary
install:
	@echo "Installing..."
	$(GOCMD) install $(LDFLAGS) $(MAIN_PACKAGE)

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## security: Run security scanner
security:
	@echo "Running security scanner..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

## generate: Run go generate
generate:
	@echo "Running go generate..."
	$(GOCMD) generate ./...

## help: Show this help
help:
	@echo "DriftGuard - AI Agent Drift Detection System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'

.DEFAULT_GOAL := help