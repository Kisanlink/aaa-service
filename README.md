# AAA Service

A refactored Authentication, Authorization, and Accounting (AAA) service that uses the kisanlink-db package for database operations and follows a clean architecture pattern.

## Project Structure

```
aaa-service/
├── config/
│   └── database.go              # Database configuration and initialization
├── entities/
│   ├── models/                  # Database models (pure domain entities)
│   │   ├── user.go             # User database model
│   │   ├── address.go          # Address database model
│   │   ├── role.go             # Role and Permission models
│   │   └── user_role.go        # UserRole relationship model
│   ├── requests/               # Request DTOs implementing Request interface
│   │   ├── request.go          # Base Request interface
│   │   └── users/              # User-related request DTOs
│   │       ├── create_user.go  # CreateUserRequest
│   │       └── update_user.go  # UpdateUserRequest
│   └── responses/              # Response DTOs implementing Response 
│       ├── response.go         # Base Response interface
│       └── users/              # User-related response DTOs
│           └── user_response.go # UserResponse and related types
├── repositories/               # Data access layer
│   ├── users/
│   │   ├── user_repository.go  # Basic CRUD operations
│   │   └── user_queries.go     # Complex query methods
│   ├── addresses/
│   │   ├── address_repository.go # Basic CRUD operations
│   │   └── address_queries.go    # Complex query methods
│   └── roles/
│       ├── role_repository.go     # Basic CRUD operations
│       ├── role_queries.go        # Complex query methods
│       ├── user_role_repository.go # Basic CRUD operations
│       └── user_role_queries.go   # Complex query methods
├── go.mod                      # Go module file
├── main.go                     # Application entry point
└── README.md                   # This file
```

## Architecture Principles

### 1. Clean Separation of Concerns
- **Models**: Pure database entities with business logic
- **Requests**: Input validation and data transfer objects
- **Responses**: Output formatting and data transfer objects
- **Repositories**: Data access layer with base repository pattern

### 2. Interface-Based Design
- All request DTOs implement the `Request` interface
- All response DTOs implement the `Response` interface
- Repositories use the base repository pattern from kisanlink-db

### 3. Manageable File Sizes
- Each file is kept under 200 lines for maintainability
- Complex operations are split into separate files
- Clear separation between basic CRUD and complex queries

## Key Features

### Database Models
- All models extend `base.BaseModel` from kisanlink-db
- Automatic ID generation using hash-based identifiers
- Built-in audit fields (created_at, updated_at, created_by, etc.)
- Soft delete support
- GORM integration for database operations

### Request Validation
- Comprehensive validation in request DTOs
- Custom validation rules for business logic
- Clear error messages for validation failures

### Response Formatting
- Consistent response structure
- Proper JSON serialization
- Relationship handling (nested objects)

### Repository Pattern
- Base repository functionality from kisanlink-db
- Custom query methods for specific business needs
- Support for multiple database backends (PostgreSQL, DynamoDB, SpiceDB)

## Usage Example

```go
// Initialize database manager
dbManager, err := config.NewDatabaseManager(logger)
if err != nil {
    log.Fatal("Failed to initialize database manager", err)
}

// Initialize repositories
userRepo := users.NewUserRepository(dbManager.GetPostgresManager())

// Create a new user
user := models.NewUser("testuser", "password123", 9876543210)
user.Name = &[]string{"Test User"}[0]

if err := userRepo.Create(ctx, user); err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Convert to response format
userResponse := &responses.UserResponse{}
userResponse.FromModel(user)
```

## Environment Configuration

The service uses environment variables for configuration:

```bash
# Database Configuration
DB_PRIMARY_BACKEND=gorm
DB_POSTGRES_HOST=localhost
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=postgres
DB_POSTGRES_PASSWORD=password
DB_POSTGRES_DBNAME=kisanlink
DB_POSTGRES_SSLMODE=disable
DB_POSTGRES_MAX_CONNS=10
DB_POSTGRES_IDLE_CONNS=5

# Logging
DB_LOG_LEVEL=info
```

## Dependencies

- **kisanlink-db**: Core database functionality and base models
- **GORM**: ORM for database operations
- **PostgreSQL**: Primary database backend
- **zap**: Structured logging
- **gin**: HTTP framework (for future API endpoints)

## Development

### Adding New Models
1. Create the model in `entities/models/`
2. Extend `base.BaseModel`
3. Implement required interface methods
4. Add repository files in `repositories/`

### Adding New Request DTOs
1. Create the request in `entities/requests/`
2. Implement the `Request` interface
3. Add comprehensive validation logic

### Adding New Response DTOs
1. Create the response in `entities/responses/`
2. Implement the `Response` interface
3. Add conversion methods from models

## Benefits of This Structure

1. **Maintainability**: Small, focused files that are easy to understand
2. **Testability**: Clear separation makes unit testing straightforward
3. **Scalability**: Easy to add new features without affecting existing code
4. **Consistency**: Standardized patterns across all components
5. **Reusability**: Base classes and interfaces promote code reuse
6. **Type Safety**: Strong typing throughout the application 