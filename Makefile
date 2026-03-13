# DriftGuard Makefile
# Common development and deployment tasks

.PHONY: help build test run docker-clean release

# Default target
help:
	@echo "DriftGuard - AI Agent Behavior Degradation Monitoring System"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - Build the binary"
	@echo "  make test         - Run all tests"
	@echo "  make test-unit    - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make run          - Run locally"
	@echo "  make docker-up    - Start Docker environment"
	@echo "  make docker-down  - Stop Docker environment"
	@echo "  make docker-clean - Clean Docker containers"
	@echo "  make release      - Build release binaries"
	@echo "  make lint         - Run linters"
	@echo "  make fmt          - Format code"
	@echo "  make coverage     - Generate coverage report"
	@echo ""

# Build
build:
	@echo "Building driftguard..."
	go build -o driftguard ./cmd/main.go
	@echo "✅ Build complete: ./driftguard"

# Test
test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	go test ./internal/... ./pkg/... -v -race

test-integration:
	@echo "Running integration tests..."
	./tests/integration-test.sh

# Run locally
run:
	@echo "Starting driftguard..."
	go run cmd/main.go -config config.json

# Docker
docker-up:
	@echo "Starting Docker environment..."
	docker compose up -d
	@echo "✅ Services started:"
	@echo "   - DriftGuard API: http://localhost:8080"
	@echo "   - Sidecar: http://localhost:8081"
	@echo "   - Prometheus: http://localhost:9090"
	@echo "   - Grafana: http://localhost:3000 (admin/driftguard)"

docker-down:
	@echo "Stopping Docker environment..."
	docker compose down

docker-clean:
	@echo "Cleaning Docker containers..."
	docker compose down -v
	docker system prune -f

docker-logs:
	docker compose logs -f $(service)

docker-restart:
	docker compose restart

# Release
release:
	@echo "Building release..."
	goreleaser release --snapshot --rm-dist
	@echo "✅ Release build complete in ./dist/"

# Lint
lint:
	@echo "Running linters..."
	go vet ./...
	golangci-lint run

# Format
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Coverage
coverage:
	@echo "Generating coverage report..."
	go test ./... -coverprofile=coverage.txt -covermode=atomic
	go tool cover -html=coverage.txt -o coverage.html
	@echo "✅ Coverage report: coverage.html"
	@echo "Open with: open coverage.html"

# Clean
clean:
	@echo "Cleaning..."
	rm -f driftguard
	rm -f coverage.txt coverage.html
	rm -rf dist/
	@echo "✅ Clean complete"

# Dev - watch and rebuild
dev:
	@echo "Starting development mode..."
	air -c .air.toml

# Init - setup development environment
init:
	@echo "Setting up development environment..."
	go mod download
	docker compose pull
	@echo "✅ Setup complete"
	@echo "Run 'make docker-up' to start services"
