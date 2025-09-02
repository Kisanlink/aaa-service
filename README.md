# AAA Service

A comprehensive Authentication, Authorization, and Accounting service built with Go, featuring role-based access control, permission management, and PostgreSQL RBAC for advanced authorization.

## Features

- **Authentication** - JWT-based authentication with refresh tokens
- **Authorization** - PostgreSQL RBAC integration for real-time permission evaluation
- **Accounting** - Comprehensive audit logging and event tracking
- **User Management** - Complete user lifecycle management
- **Role Management** - Hierarchical roles with organization and group scoping
- **Permission System** - Fine-grained permissions with resource-level control
- **Multi-tenancy** - Organization and group-based isolation
- **API Gateway** - RESTful HTTP and gRPC interfaces
- **Caching** - Redis-based caching for performance
- **Database Support** - PostgreSQL primary, DynamoDB and S3 optional

## Architecture

The service follows a clean architecture pattern with:

- **Handlers** - HTTP and gRPC request handling
- **Services** - Business logic and authorization
- **Repositories** - Data access layer
- **Models** - Domain entities and data structures
- **Middleware** - Authentication, authorization, and audit
- **Database** - PostgreSQL with kisanlink-db manager

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin (HTTP), gRPC
- **Database**: PostgreSQL (primary), DynamoDB, S3
- **Cache**: Redis
- **Authorization**: PostgreSQL RBAC
- **Logging**: Zap
- **Validation**: Custom validator
- **Testing**: Go testing framework

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose

### Environment Setup

1. Copy the environment file:

   ```bash
   cp env.example .env
   ```

2. Update the database configuration in `.env`:

   ```bash
   DB_POSTGRES_HOST=localhost
   DB_POSTGRES_PORT=5432
   DB_POSTGRES_USER=postgres
   DB_POSTGRES_PASSWORD=your_password
   DB_POSTGRES_DBNAME=kisanlink
   ```

3. Start the required services:

   ```bash
   docker-compose up -d postgres redis
   ```

4. Run the service:
   ```bash
   go run cmd/server/main.go
   ```

### Database Setup

The service will automatically:

- Connect to PostgreSQL using kisanlink-db manager
- Run migrations if `AAA_AUTO_MIGRATE=true`
- Seed initial data if `AAA_RUN_SEED=true`

## API Documentation

### Interactive Documentation

- **Swagger UI**: Available at `/swagger/index.html` when `AAA_ENABLE_DOCS=true`
- **OpenAPI Spec**: Available at `/swagger/doc.json` and `/swagger/swagger.yaml`

### Comprehensive Guides

- **[API Usage Examples](docs/API_EXAMPLES.md)** - Complete examples for all endpoints including:

  - Enhanced login with password and MPIN support
  - Role assignment and management
  - MPIN management (set/update)
  - User lifecycle management
  - Error handling scenarios

- **[Error Response Documentation](docs/ERROR_RESPONSES.md)** - Detailed error handling guide:
  - Standardized error response format
  - Authentication and authorization errors
  - Validation error examples
  - Role management error scenarios
  - MPIN management error cases
  - System error responses

### Key API Features

- **Enhanced Authentication**: Login with phone/password or phone/MPIN
- **Role Management**: Assign/remove roles with comprehensive audit logging
- **MPIN Support**: Set and update MPIN for secure mobile authentication
- **User Management**: Complete user lifecycle including soft deletion
- **Comprehensive Error Handling**: Consistent error responses with detailed context

### Additional Resources

- **gRPC**: Reflection enabled for development tools
- **Health Check**: `/health` endpoint for monitoring
- **Audit Logging**: All operations logged with request tracing

## Configuration

### Database Configuration

The service uses kisanlink-db for database management:

```bash
# Primary backend (gorm for PostgreSQL)
DB_PRIMARY_BACKEND=gorm

# PostgreSQL settings
DB_POSTGRES_HOST=localhost
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=postgres
DB_POSTGRES_PASSWORD=your_password
DB_POSTGRES_DBNAME=kisanlink

# Optional: DynamoDB and S3
DB_DYNAMO_REGION=us-east-1
DB_S3_REGION=us-east-1
```

### Authorization Configuration

PostgreSQL RBAC is used for authorization:

- No external authorization service required
- Permissions stored in PostgreSQL tables
- Real-time permission evaluation
- Support for hierarchical roles and resources

## Development

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests with test database
docker-compose -f docker-compose.test.yml up -d
go test -tags=integration ./...
```

### Code Structure

```
aaa-service/
├── cmd/server/          # Main application entry point
├── internal/            # Private application code
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP and gRPC handlers
│   ├── services/       # Business logic
│   ├── repositories/   # Data access layer
│   ├── middleware/     # HTTP middleware
│   └── routes/         # Route definitions
├── pkg/                # Public packages
├── migrations/         # Database migrations
└── docs/              # Documentation and schemas
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.
