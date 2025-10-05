---
inclusion: always
---

# AAA Service Development Guide

## Service Overview

AAA Service is an enterprise-grade Authentication, Authorization, and Accounting service providing JWT-based auth, PostgreSQL RBAC with hierarchical roles, comprehensive audit logging, and multi-tenant organization management.

## Technology Stack

- **Go 1.24+** with Gin (HTTP), gRPC, PostgreSQL 17+, Redis 7+, GORM via kisanlink-db
- **Key Dependencies**: kisanlink-db (custom DB manager), golang-jwt/jwt/v4, go.uber.org/zap, swaggo/swag
- **Tools**: Docker, golangci-lint, Make for automation

## Architecture Principles

- **Clean Architecture**: Domain-driven design with interface-based dependency injection
- **File Size Limit**: Maximum 300 lines per Go file - split before reaching limit
- **Single Responsibility**: Each file has one focused purpose
- **Repository Pattern**: Use kisanlink-db for all database operations

## Project Structure

```
internal/
├── config/          # Configuration management
├── entities/        # Models, requests (by domain), responses (by domain)
├── handlers/        # HTTP handlers (one per domain)
├── services/        # Business logic (split by domain/operation)
├── repositories/    # Data access via kisanlink-db
├── middleware/      # HTTP middleware
└── grpc_server/     # gRPC implementation

pkg/                 # Public API packages
migrations/          # Database migrations and seeds
```

## Code Standards

### Naming Conventions

- Files: `snake_case` (user_service.go, auth_handler.go)
- Packages: lowercase (users, auth, roles)
- Exported functions: PascalCase (CreateUser, ValidateToken)
- Private functions: camelCase (validateRequest, hashPassword)
- Variables: camelCase (userID, accessToken)
- Constants: UPPER_CASE (JWT_SECRET, MAX_LOGIN_ATTEMPTS)

### Service Organization

Split services by operation:

```
internal/services/user/
├── service.go    # Service struct and constructor
├── create.go     # Creation logic
├── read.go       # Retrieval operations
├── update.go     # Modification operations
└── delete.go     # Deletion logic
```

### Error Handling

- Use custom error types from `pkg/errors/`
- Log errors with context using Zap
- Provide meaningful error messages

### Testing

- Write tests alongside implementation
- Use table-driven tests for multiple scenarios
- Mock dependencies using interfaces

## Development Commands

```bash
# Development
make run                    # Run locally
go run cmd/server/main.go   # Alternative run

# Testing
make test                   # All tests
make test-coverage          # With coverage
go test ./...              # Alternative test

# Quality
make format                 # Format code
make lint                   # Run linter
make docs                   # Generate Swagger docs

# Docker
docker-compose up -d        # Local environment
```

## Environment Variables

Use `AAA_` prefix: `AAA_JWT_SECRET` (required), `AAA_AUTO_MIGRATE`, `AAA_RUN_SEED`, `AAA_ENABLE_DOCS`
Database: `DB_POSTGRES_*`, Redis: `REDIS_*`

## Key Implementation Rules

1. **Always use kisanlink-db** for database operations
2. **Validate at request boundaries** (handlers) using struct tags
3. **Use context** for request timeouts and cancellation
4. **Implement proper caching** with Redis and appropriate TTL
5. **Follow import order**: standard library, third-party, local
6. **Split files before 250 lines** to stay under 300 limit
7. **Use interfaces** for dependency injection and testing
