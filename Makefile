# AAA Service Makefile
# Provides common development and deployment commands

.PHONY: help build test clean docker-build docker-run dev setup lint format check coverage docs

# Default target
help: ## Show this help message
	@echo "AAA Service - Identity & Access Management"
	@echo "=========================================="
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development commands
setup: ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	go mod tidy
	@echo "✅ Development environment ready"

build: ## Build the application
	@echo "Building AAA Service..."
	go build -o bin/aaa-server cmd/server/main.go
	@echo "✅ Build complete: bin/aaa-server"

test: ## Run tests
	@echo "Running tests..."
	go test ./... -v

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

check: lint test ## Run all quality checks

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t aaa-service:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 aaa-service:latest

docker-clean: ## Clean Docker containers and images
	@echo "Cleaning Docker resources..."
	docker system prune -f
	docker image prune -f

# Database commands
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@echo "TODO: Implement migration script"

db-seed: ## Seed database with test data
	@echo "Seeding database..."
	@echo "TODO: Implement seed script"

# Development server
dev: ## Run development server
	@echo "Starting development server..."
	go run cmd/server/main.go

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	swag init -g cmd/server/main.go
	@echo "✅ Documentation generated"

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "✅ Cleanup complete"

# Production
prod-build: ## Build for production
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/aaa-server cmd/server/main.go

# Security
security-scan: ## Run security scan
	@echo "Running security scan..."
	gosec ./...
	@echo "✅ Security scan complete"

# Performance
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. ./...

# Monitoring
health-check: ## Check service health
	@echo "Checking service health..."
	curl -f http://localhost:8080/health || echo "Service not running"

# Dependencies
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

deps-check: ## Check for outdated dependencies
	@echo "Checking for outdated dependencies..."
	go list -u -m all

# Git hooks
install-hooks: ## Install git hooks
	@echo "Installing git hooks..."
	cp scripts/pre-commit.sh .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "✅ Git hooks installed"

# Quick development workflow
dev-setup: setup install-hooks ## Complete development setup
	@echo "✅ Development setup complete"

dev-start: dev-setup dev ## Start development environment

# Helpers
version: ## Show version information
	@echo "AAA Service"
	@echo "Version: $(shell git describe --tags --always --dirty)"
	@echo "Commit: $(shell git rev-parse HEAD)"
	@echo "Date: $(shell date)"

# Default target
.DEFAULT_GOAL := help
