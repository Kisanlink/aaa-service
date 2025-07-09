# Makefile for AAA Service
.PHONY: help install-tools lint test test-unit test-integration test-coverage build clean fmt imports security tidy docker pre-commit setup-hooks

# Variables
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*' -not -path './docs/*')
BINARY_NAME := aaa-service
BUILD_DIR := bin
COVERAGE_DIR := coverage
DOCKER_IMAGE := aaa-service:latest

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

## help: Show this help message
help:
	@echo "$(BLUE)AAA Service Makefile Commands:$(NC)"
	@echo ""
	@grep -E '^## [a-zA-Z_-]+:' $(MAKEFILE_LIST) | \
		sed 's/## //' | \
		awk -F: '{printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

## install-tools: Install required development tools
install-tools:
	@echo "$(BLUE)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@which pre-commit > /dev/null || (echo "$(RED)Please install pre-commit: pip install pre-commit$(NC)" && exit 1)
	@echo "$(GREEN)Tools installed successfully$(NC)"

## setup-hooks: Install pre-commit hooks
setup-hooks: install-tools
	@echo "$(BLUE)Setting up pre-commit hooks...$(NC)"
	@pre-commit install
	@pre-commit install --hook-type commit-msg
	@echo "$(GREEN)Pre-commit hooks installed$(NC)"

## fmt: Format Go code
fmt:
	@echo "$(BLUE)Formatting Go code...$(NC)"
	@gofmt -s -w $(GO_FILES)
	@echo "$(GREEN)Code formatted$(NC)"

## imports: Fix and organize imports
imports:
	@echo "$(BLUE)Fixing imports...$(NC)"
	@goimports -w $(GO_FILES)
	@echo "$(GREEN)Imports fixed$(NC)"

## tidy: Tidy Go modules
tidy:
	@echo "$(BLUE)Tidying Go modules...$(NC)"
	@go mod tidy
	@go mod verify
	@echo "$(GREEN)Modules tidied$(NC)"

## lint: Run all linters
lint:
	@echo "$(BLUE)Running linters...$(NC)"
	@golangci-lint run --timeout=5m
	@echo "$(GREEN)Linting completed$(NC)"

## security: Run security analysis
security:
	@echo "$(BLUE)Running security analysis...$(NC)"
	@gosec -quiet ./...
	@echo "$(GREEN)Security analysis completed$(NC)"

## build: Build the application
build:
	@echo "$(BLUE)Building application...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## test-unit: Run unit tests
test-unit:
	@echo "$(BLUE)Running unit tests...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -race -short -coverprofile=$(COVERAGE_DIR)/unit.out -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_DIR)/unit.out -o $(COVERAGE_DIR)/unit.html
	@echo "$(GREEN)Unit tests completed$(NC)"
	@echo "$(YELLOW)Coverage report: $(COVERAGE_DIR)/unit.html$(NC)"

## test-integration: Run integration tests
test-integration:
	@echo "$(BLUE)Running integration tests...$(NC)"
	@echo "$(YELLOW)Starting test database...$(NC)"
	@docker-compose -f docker-compose.test.yml up -d --wait
	@sleep 5
	@mkdir -p $(COVERAGE_DIR)
	@go test -race -tags=integration -coverprofile=$(COVERAGE_DIR)/integration.out -covermode=atomic ./test/integration/...
	@go tool cover -html=$(COVERAGE_DIR)/integration.out -o $(COVERAGE_DIR)/integration.html
	@echo "$(GREEN)Integration tests completed$(NC)"
	@echo "$(YELLOW)Coverage report: $(COVERAGE_DIR)/integration.html$(NC)"
	@echo "$(YELLOW)Stopping test database...$(NC)"
	@docker-compose -f docker-compose.test.yml down

## test-coverage: Run all tests and generate combined coverage
test-coverage: test-unit test-integration
	@echo "$(BLUE)Generating combined coverage report...$(NC)"
	@echo 'mode: atomic' > $(COVERAGE_DIR)/combined.out
	@tail -n +2 $(COVERAGE_DIR)/unit.out >> $(COVERAGE_DIR)/combined.out
	@tail -n +2 $(COVERAGE_DIR)/integration.out >> $(COVERAGE_DIR)/combined.out
	@go tool cover -html=$(COVERAGE_DIR)/combined.out -o $(COVERAGE_DIR)/combined.html
	@go tool cover -func=$(COVERAGE_DIR)/combined.out | tail -1
	@echo "$(GREEN)Combined coverage report: $(COVERAGE_DIR)/combined.html$(NC)"

## test: Run all tests (unit only for pre-commit)
test: test-unit

## benchmark: Run benchmarks
benchmark:
	@echo "$(BLUE)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...
	@echo "$(GREEN)Benchmarks completed$(NC)"

## docker: Build Docker image
docker:
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build -t $(DOCKER_IMAGE) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE)$(NC)"

## clean: Clean build artifacts and coverage reports
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR) $(COVERAGE_DIR)
	@go clean -cache -testcache -modcache
	@echo "$(GREEN)Clean completed$(NC)"

## pre-commit: Run all pre-commit checks manually
pre-commit: fmt imports tidy lint security test-unit build
	@echo "$(GREEN)All pre-commit checks passed!$(NC)"

## ci: Run all CI checks (includes integration tests)
ci: fmt imports tidy lint security test-coverage build
	@echo "$(GREEN)All CI checks passed!$(NC)"

## dev-setup: Complete development environment setup
dev-setup: install-tools setup-hooks
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@cp .env.example .env 2>/dev/null || echo "No .env.example found"
	@echo "$(GREEN)Development environment setup completed!$(NC)"
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Update .env file with your configuration"
	@echo "  2. Run 'make test' to verify everything works"
	@echo "  3. Start coding!"

## run: Run the application locally
run: build
	@echo "$(BLUE)Starting AAA service...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

## docs: Generate documentation
docs:
	@echo "$(BLUE)Generating documentation...$(NC)"
	@go doc -all ./... > docs/API.md
	@echo "$(GREEN)Documentation generated$(NC)"
