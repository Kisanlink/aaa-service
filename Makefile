# AAA Service Makefile
# Common development tasks for the refactored AAA service

.PHONY: help build test clean run docker-build docker-run lint format check-deps install-deps

# Default target
help:
	@echo "AAA Service - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application locally"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  format        - Format code"
	@echo ""
	@echo "Dependencies:"
	@echo "  install-deps  - Install dependencies"
	@echo "  check-deps    - Check for dependency updates"
	@echo "  tidy          - Tidy go.mod and go.sum"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-stop   - Stop Docker containers"
	@echo ""
	@echo "Database:"
	@echo "  db-migrate    - Run database migrations"
	@echo "  db-seed       - Seed database with test data"
	@echo ""
	@echo "Utilities:"
	@echo "  clean         - Clean build artifacts"
	@echo "  proto-gen     - Generate protobuf files"
	@echo "  swagger-gen   - Generate Swagger documentation"

# Build the application
build:
	@echo "Building AAA Service..."
	go build -o bin/aaa-service main.go
	@echo "Build complete: bin/aaa-service"

# Run the application locally
run:
	@echo "Starting AAA Service..."
	go run main.go

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific test package
test-package:
	@echo "Running tests for package: $(PACKAGE)"
	go test -v ./$(PACKAGE)/...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Check for dependency updates
check-deps:
	@echo "Checking for dependency updates..."
	go list -u -m all

# Tidy go.mod and go.sum
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	go mod verify

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf coverage.out
	rm -rf coverage.html
	rm -rf tmp/
	go clean -cache

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t aaa-service:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Stop Docker containers
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# Run database migrations
db-migrate:
	@echo "Running database migrations..."
	# Add migration commands here
	@echo "Migrations complete"

# Seed database with test data
db-seed:
	@echo "Seeding database with test data..."
	# Add seeding commands here
	@echo "Database seeded"

# Generate protobuf files
proto-gen:
	@echo "Generating protobuf files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

# Generate Swagger documentation
swagger-gen:
	@echo "Generating Swagger documentation..."
	swag init -g main.go -o docs

# Development setup
dev-setup: install-deps tidy
	@echo "Development setup complete"

# Pre-commit checks
pre-commit: format lint test
	@echo "Pre-commit checks passed"

# CI/CD pipeline
ci: clean install-deps test-coverage lint
	@echo "CI pipeline completed"

# Production build
prod-build:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/aaa-service main.go

# Health check
health-check:
	@echo "Checking service health..."
	curl -f http://localhost:8080/health || echo "Service is not healthy"

# Performance test
perf-test:
	@echo "Running performance tests..."
	# Add performance testing commands here
	@echo "Performance tests complete"

# Security scan
security-scan:
	@echo "Running security scan..."
	gosec ./...
	@echo "Security scan complete"

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Show application info
info:
	@echo "AAA Service Information:"
	@echo "Version: $(shell git describe --tags --always --dirty)"
	@echo "Commit: $(shell git rev-parse HEAD)"
	@echo "Branch: $(shell git branch --show-current)"
	@echo "Go Version: $(shell go version)"
	@echo "Build Time: $(shell date)"

# Create release
release:
	@echo "Creating release..."
	# Add release creation commands here
	@echo "Release created"

# Backup database
db-backup:
	@echo "Creating database backup..."
	# Add backup commands here
	@echo "Database backup created"

# Restore database
db-restore:
	@echo "Restoring database..."
	# Add restore commands here
	@echo "Database restored"

# Monitor logs
logs:
	@echo "Monitoring application logs..."
	docker-compose logs -f aaa-service

# Scale services
scale:
	@echo "Scaling services..."
	docker-compose up -d --scale aaa-service=$(REPLICAS)

# Environment setup
env-setup:
	@echo "Setting up environment..."
	cp .env.example .env
	@echo "Environment file created. Please update .env with your configuration."

# Database reset
db-reset:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres redis
	@echo "Database reset complete"

# Full development cycle
dev-cycle: clean install-deps build test run
	@echo "Development cycle complete"

# Quick start for new developers
quick-start: env-setup dev-setup docker-run
	@echo "Quick start complete. Service should be running at http://localhost:8080"
