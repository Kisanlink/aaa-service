# AAA Service Architecture

## Project Structure

The AAA (Authentication, Authorization, and Accounting) service follows Go project layout standards with clear separation between internal and public APIs.

```
aaa-service/
├── cmd/                    # Main applications
│   └── server/            # HTTP and gRPC server
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── database/         # Database connection and management
│   ├── entities/         # Data models and DTOs
│   │   ├── models/      # Domain models
│   │   ├── requests/    # Request DTOs
│   │   └── responses/   # Response DTOs
│   ├── grpc_server/     # gRPC server implementation
│   ├── handlers/        # HTTP handlers
│   │   ├── auth/       # Authentication handlers
│   │   ├── users/      # User management handlers
│   │   ├── roles/      # Role management handlers
│   │   └── ...         # Other domain handlers
│   ├── interfaces/      # Interface definitions
│   ├── middleware/      # HTTP middleware
│   ├── repositories/    # Data access layer
│   │   ├── users/      # User repository
│   │   ├── roles/      # Role repository
│   │   └── addresses/  # Address repository
│   ├── routes/         # Route definitions (split into domain-specific files)
│   └── services/       # Business logic layer
│       ├── auth/       # Authentication service (split into focused files)
│       ├── user/       # User service (split into CRUD operations)
│       └── ...         # Other domain services
├── pkg/                   # Public API and libraries
│   ├── client/           # HTTP client for external consumption
│   ├── models/           # Public data models
│   ├── proto/            # Protocol buffer definitions
│   ├── pb/               # Generated protobuf code
│   └── errors/           # Error types and handling
├── utils/                 # Shared utilities
├── docs/                  # Documentation and OpenAPI specs
├── test/                  # Test files and test data
└── scripts/               # Build and deployment scripts
```

## Design Principles

### 1. File Size Limit
- **Maximum 300 lines per file**: All Go files are kept under 300 lines to maintain readability and focus
- **Single Responsibility**: Each file has a clear, single purpose
- **Logical Grouping**: Related functionality is grouped together

### 2. Package Organization

#### Internal Packages (`internal/`)
- **Private to the service**: Cannot be imported by external applications
- **Business Logic**: Contains all service-specific implementation
- **Domain-Driven**: Organized by business domains (auth, users, roles, etc.)

#### Public Packages (`pkg/`)
- **External Consumption**: Can be imported by other services
- **Client Libraries**: HTTP client for service integration
- **Data Models**: Public API contracts
- **Generated Code**: Protocol buffer generated files

### 3. Service Layer Architecture

#### User Service Structure
The user service is split into focused files:
- `service.go`: Service struct and constructor
- `create.go`: User creation logic
- `read.go`: User retrieval operations
- `update.go`: User modification operations
- `delete.go`: User deletion logic

#### Auth Service Structure
Similar modular approach for authentication:
- `service.go`: Service setup and configuration
- `login.go`: Login and authentication logic
- `tokens.go`: Token management (JWT, refresh tokens)
- `mfa.go`: Multi-factor authentication
- `session.go`: Session management

### 4. Route Organization

Routes are split by domain:
- `setup.go`: Main route configuration
- `auth_routes.go`: Authentication endpoints
- `user_routes.go`: User management endpoints
- `health_routes.go`: Health check endpoints
- `admin_routes.go`: Administrative endpoints

## Key Features

### 1. Modular Design
- Each domain is self-contained
- Clear interfaces between layers
- Easy to test and maintain

### 2. Scalable Architecture
- Services can be split into microservices easily
- Clear API boundaries
- Standardized error handling

### 3. Developer Experience
- Clear file organization
- Consistent naming conventions
- Comprehensive documentation

### 4. External Integration
- Public client library in `pkg/client/`
- Well-defined data models in `pkg/models/`
- OpenAPI/Swagger documentation

## Usage Examples

### Using the Public Client
```go
import "github.com/Kisanlink/aaa-service/pkg/client"
import "github.com/Kisanlink/aaa-service/pkg/models"

// Create client
client := client.NewClient(&client.Config{
    BaseURL: "http://localhost:8080",
    APIKey:  "your-api-key",
})

// Login user
loginReq := &models.LoginRequest{
    PhoneNumber: "+1234567890",
    CountryCode: "US",
    Password:    "password123",
}
response, err := client.Login(ctx, loginReq)
```

### Extending the Service
1. **Add new domain**: Create new packages under `internal/`
2. **Add new endpoints**: Create route files under `internal/routes/`
3. **Add public models**: Add to `pkg/models/` for external consumption
4. **Keep files focused**: Split large files when they exceed 300 lines

## Build and Development

```bash
# Build the service
go build -o bin/aaa-server cmd/server/main.go

# Run tests
go test ./...

# Generate Swagger docs
swag init -g cmd/server/main.go -o docs --parseInternal

# Update imports after refactoring
./scripts/update_imports.sh
```

## Migration Guide

If migrating from the old structure:
1. Update import paths using the provided script
2. Split large files into focused modules
3. Move public APIs to `pkg/` directory
4. Update CI/CD pipelines to reflect new structure
