# AAA Service Project Structure

## File Organization Principles

### File Size Limit

- **Maximum 300 lines per file**: All Go files must stay under 300 lines for readability
- **Single Responsibility**: Each file has one clear, focused purpose
- **Split when growing**: Break large files into logical modules before hitting the limit

### Package Structure

```
aaa-service/
├── cmd/server/              # Application entry point
├── internal/                # Private application code
│   ├── config/             # Configuration management
│   ├── entities/           # Data models and DTOs
│   │   ├── models/        # Domain models
│   │   ├── requests/      # Request DTOs (organized by domain)
│   │   └── responses/     # Response DTOs (organized by domain)
│   ├── handlers/          # HTTP handlers (one per domain)
│   ├── services/          # Business logic (split by domain and operation)
│   ├── repositories/      # Data access layer (using kisanlink-db)
│   ├── middleware/        # HTTP middleware
│   └── grpc_server/       # gRPC server implementation
├── pkg/                    # Public API packages
│   ├── client/            # HTTP client for external use
│   ├── models/            # Public data models
│   ├── proto/             # Protocol buffer definitions
│   └── errors/            # Error types and handling
├── migrations/             # Database migrations and seeds
├── docs/                   # Documentation and OpenAPI specs
└── utils/                  # Shared utilities
```

## Code Organization Patterns

### Service Layer Structure

Services are split into focused files by operation:

```
internal/services/user/
├── service.go          # Service struct and constructor
├── create.go           # User creation logic
├── read.go             # User retrieval operations
├── update.go           # User modification operations
└── delete.go           # User deletion logic
```

### Handler Organization

Handlers are organized by domain with clear separation:

```
internal/handlers/
├── auth/               # Authentication endpoints
├── users/              # User management endpoints
├── roles/              # Role management endpoints
├── organizations/      # Organization endpoints
└── health/             # Health check endpoints
```

### Request/Response Structure

DTOs are organized by domain and operation:

```
internal/entities/requests/
├── auth_requests.go    # Authentication requests
├── users/              # User-related requests
│   ├── create_user.go
│   ├── update_user.go
│   └── create_user_profile.go
└── roles/              # Role-related requests
    ├── create_role.go
    ├── assign_role.go
    └── update_role.go
```

## Naming Conventions

### Files and Packages

- Use snake_case for file names: `user_service.go`, `auth_handler.go`
- Use lowercase for package names: `users`, `auth`, `roles`
- Group related functionality in domain packages

### Functions and Methods

- Use PascalCase for exported functions: `CreateUser`, `ValidateToken`
- Use camelCase for private functions: `validateRequest`, `hashPassword`
- Use descriptive names that indicate purpose: `VerifyUserPasswordByPhone`

### Variables and Constants

- Use camelCase for variables: `userID`, `accessToken`
- Use UPPER_CASE for constants: `JWT_SECRET`, `MAX_LOGIN_ATTEMPTS`
- Use meaningful names over abbreviations

## Dependency Management

### Interface-Based Design

- Define interfaces in `internal/interfaces/interfaces.go`
- Use dependency injection in service constructors
- Keep interfaces focused and minimal

### Repository Pattern

- Use kisanlink-db for all database operations
- Implement repository interfaces for each domain
- Leverage kisanlink-db's multi-backend support

### Service Dependencies

```go
type Service struct {
    userRepo     interfaces.UserRepository
    roleRepo     interfaces.RoleRepository
    cacheService interfaces.CacheService
    logger       *zap.Logger
    validator    interfaces.Validator
}
```

## Code Quality Standards

### Error Handling

- Use custom error types from `pkg/errors/`
- Provide meaningful error messages
- Log errors with appropriate context using Zap

### Validation

- Validate at request boundaries (handlers)
- Use struct tags for validation rules
- Implement custom validation methods when needed

### Testing

- Write tests alongside implementation files
- Use table-driven tests for multiple scenarios
- Mock dependencies using interfaces

## Efficiency Guidelines

### Database Operations

- Use kisanlink-db for consistent database access
- Implement proper connection pooling
- Use transactions for multi-step operations

### Caching Strategy

- Cache frequently accessed data in Redis
- Use appropriate TTL values
- Implement cache invalidation patterns

### Performance Considerations

- Use context for request timeouts
- Implement proper pagination for list operations
- Use database indexes for query optimization

## Migration and Refactoring

### When to Split Files

- File approaches 250 lines (before 300 limit)
- Multiple responsibilities in single file
- Difficult to understand or test

### Refactoring Guidelines

- Move common functionality to kisanlink-db
- Extract reusable components to shared packages
- Maintain backward compatibility in public APIs

### Import Organization

- Standard library imports first
- Third-party imports second
- Local imports last
- Use goimports for automatic formatting
