# Variables
TEST_DIR = ./test
BINARY_NAME = build
SRC_FILE=./main.go
MIGRATE_MAIN=./main.go

# Run tests
test:
	@echo "Running Go tests..."
	@go test -v $(TEST_DIR)/...

# Run the application with Air (ensure Air is installed in GOPATH)
air:
	@$(shell go env GOPATH)/bin/air || air

# Run the application normally
run:
	@go run $(SRC_FILE)

# Build the application binary
build:
	@go build -o $(BINARY_NAME) $(SRC_FILE)

# Run database migrations (all or specific tables)
migrate:
	@echo "Running database migrations..."
	@go run -tags=migrate $(MIGRATE_MAIN) migrate

# Run migrations for specific tables (e.g., make migrate-tables table1 table2)
migrate-tables:
	@echo "Running migrations for specified tables..."
	@go run -tags=migrate $(MIGRATE_MAIN) migrate $(filter-out $@,$(MAKECMDGOALS))

# Reset all migrations
reset:
	@echo "Resetting all database tables..."
	@go run -tags=migrate $(MIGRATE_MAIN) reset

# Reset specific tables (e.g., make reset-tables table1 table2)
reset-tables:
	@echo "Resetting specified database tables..."
	@go run -tags=migrate $(MIGRATE_MAIN) reset $(filter-out $@,$(MAKECMDGOALS))

swagger:
	@echo "Generating Swagger documentation..."
	@swag init 
	@echo "Swagger docs generated in ./docs"

# Clean up generated files (Linux/macOS)
clean-linux:
	@rm -f $(BINARY_NAME)

# Clean up generated files (Windows)
clean-windows:
	@if exist $(BINARY_NAME) del $(BINARY_NAME)

# Format the Go code
fmt:
	@go fmt ./...

# Install dependencies
deps:
	@go mod tidy

# List all available commands
help:
	@echo "Available commands:"
	@echo "  test            - Run all tests"
	@echo "  air             - Run the application with Air"
	@echo "  run             - Run the application normally"
	@echo "  build           - Build the application binary"
	@echo "  migrate         - Run all database migrations"
	@echo "  migrate-tables  - Run migrations for specific tables (e.g., make migrate-tables table1 table2)"
	@echo "  reset           - Reset all database tables"
	@echo "  reset-tables    - Reset specific tables (e.g., make reset-tables table1 table2)"
	@echo "  clean-linux     - Clean up generated files (Linux/macOS)"
	@echo "  clean-win       - Clean up generated files (Windows)"
	@echo "  fmt             - Format Go code"
	@echo "  deps            - Install dependencies"

# Handle target arguments
%:
	@:

.PHONY: test air run build migrate migrate-tables reset reset-tables clean-linux clean-windows fmt deps help