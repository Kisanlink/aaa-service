# AAA Service Technology Stack

## Core Technologies

- **Language**: Go 1.24+ (using latest toolchain)
- **HTTP Framework**: Gin for REST API endpoints
- **gRPC**: Google gRPC for high-performance service-to-service communication
- **Database**: PostgreSQL 17+ (primary), with optional DynamoDB and S3 support
- **Cache**: Redis 7+ for performance optimization
- **ORM**: GORM for PostgreSQL interactions via kisanlink-db manager
- **Authentication**: JWT tokens with golang-jwt/jwt/v4
- **Logging**: Zap (go.uber.org/zap) for structured logging
- **Validation**: Custom validator with go-playground/validator/v10
- **Documentation**: Swagger/OpenAPI with swaggo/swag

## Key Dependencies

- **kisanlink-db**: Custom database manager for multi-backend support
- **gin-gonic/gin**: HTTP web framework
- **gorm.io/gorm**: ORM for database operations
- **go-redis/redis/v8**: Redis client
- **google.golang.org/grpc**: gRPC framework
- **google.golang.org/protobuf**: Protocol buffer support

## Development Tools

- **Docker**: Multi-stage builds with Alpine Linux base
- **Docker Compose**: Local development environment with PostgreSQL and Redis
- **golangci-lint**: Code linting and quality checks
- **swag**: API documentation generation
- **Make**: Build automation and common tasks

## Common Commands

### Development

```bash
# Setup development environment
make setup

# Run the service locally
make run
# or
go run cmd/server/main.go

# Build the application
make build
# or
go build -o bin/aaa-server cmd/server/main.go
```

### Testing

```bash
# Run all tests
make test
# or
go test ./...

# Run tests with coverage
make test-coverage

# Run integration tests
make test-e2e

# Run specific module tests
make test-farmers
```

### Docker

```bash
# Build Docker image
make docker-build

# Run with Docker Compose
docker-compose up -d

# Run tests with test database
docker-compose -f docker-compose.test.yml up -d
```

### Documentation

```bash
# Generate Swagger documentation
make docs
# or
swag init -g cmd/server/main.go

# View docs at http://localhost:8080/swagger/index.html
```

### Code Quality

```bash
# Format code
make format

# Run linter
make lint

# Run all quality checks
make check
```

## Environment Configuration

The service uses environment variables with the `AAA_` prefix:

- `AAA_JWT_SECRET`: JWT signing secret (required)
- `AAA_AUTO_MIGRATE`: Auto-run database migrations
- `AAA_RUN_SEED`: Seed initial data
- `AAA_ENABLE_DOCS`: Enable Swagger documentation
- Database configuration via `DB_POSTGRES_*` variables
- Redis configuration via `REDIS_*` variables

## Build System

- Uses Go modules for dependency management
- Multi-stage Docker builds for optimized production images
- Makefile provides standardized commands for all operations
- Supports both local development and containerized deployment
