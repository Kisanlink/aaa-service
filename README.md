# AAA Service - Refactored Architecture

A comprehensive Authentication, Authorization, and Accounting (AAA) service built with Go, featuring a clean architecture with proper separation of concerns, dependency injection, and high performance.

## 🏗️ Architecture Overview

The service follows a clean, layered architecture with the following components:

```
┌─────────────────┐
│   HTTP Server   │  ← Gin-based HTTP server with middleware
├─────────────────┤
│    Handlers     │  ← HTTP request handlers
├─────────────────┤
│    Services     │  ← Business logic layer
├─────────────────┤
│  Repositories   │  ← Data access layer
├─────────────────┤
│   Database      │  ← PostgreSQL + Redis
└─────────────────┘
```

### Key Features

- **Interface-based Design**: All components use interfaces for better testability and flexibility
- **Dependency Injection**: Clean dependency management using constructor injection
- **Caching Layer**: Redis-based caching for improved performance
- **Comprehensive Validation**: Input validation with custom validators
- **Error Handling**: Structured error handling with custom error types
- **Middleware Stack**: Request ID, logging, CORS, rate limiting, and security headers
- **Health Checks**: Built-in health and readiness endpoints
- **Graceful Shutdown**: Proper shutdown handling for all components

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Redis 6+
- Docker (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd aaa-service
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**

   Create a `.env` file in the root directory with the following variables:
   ```bash
   # Database Configuration
   DB_PRIMARY_BACKEND=gorm
   DB_POSTGRES_HOST=localhost
   DB_POSTGRES_PORT=5432
   DB_POSTGRES_USER=aaa_user
   DB_POSTGRES_PASSWORD=aaa_password
   DB_POSTGRES_DBNAME=aaa_service
   DB_POSTGRES_SSLMODE=disable
   DB_POSTGRES_MAX_CONNS=10
   DB_POSTGRES_IDLE_CONNS=5

   # SpiceDB Configuration (optional)
   DB_SPICEDB_ENDPOINT=localhost:50051
   DB_SPICEDB_TOKEN=your-secret-key-here

   # Redis Configuration
   REDIS_HOST=localhost
   REDIS_PORT=6379

   # Server Configuration
   PORT=8080
   LOG_LEVEL=info
   ```

4. **Run the service**
   ```bash
   go run main.go
   ```

### Docker Setup

```bash
# Start dependencies
docker-compose up -d postgres redis

# Run the service
docker-compose up aaa-service
```

## 📁 Project Structure

```
aaa-service/
├── config/                 # Configuration management
├── entities/              # Domain entities and models
│   ├── models/           # Database models
│   ├── requests/         # Request DTOs
│   └── responses/        # Response DTOs
├── handlers/             # HTTP request handlers
├── interfaces/           # Interface definitions
├── middleware/           # HTTP middleware
├── repositories/         # Data access layer
├── server/              # HTTP server setup
├── services/            # Business logic layer
├── utils/               # Utility functions
├── pkg/                 # Internal packages
│   └── errors/          # Custom error types
└── main.go              # Application entry point
```

## 🔧 Configuration

### Database Configuration

The service supports multiple database backends through the `kisanlink-db` package:

- **PostgreSQL**: Primary database for user data
- **Redis**: Caching layer for improved performance
- **SpiceDB**: Authorization database (optional)

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | PostgreSQL user | `postgres` |
| `POSTGRES_PASSWORD` | PostgreSQL password | - |
| `POSTGRES_DB` | PostgreSQL database | `aaa_service` |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | - |
| `REDIS_DB` | Redis database | `0` |

## 🛠️ API Endpoints

### Health Checks

- `GET /health` - Health check endpoint
- `GET /ready` - Readiness check endpoint

### User Management

- `POST /api/v1/users/` - Create a new user
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `GET /api/v1/users/` - List users with pagination
- `GET /api/v1/users/search` - Search users
- `POST /api/v1/users/:id/validate` - Validate user
- `POST /api/v1/users/:id/roles/:roleId` - Assign role to user
- `DELETE /api/v1/users/:id/roles/:roleId` - Remove role from user

### Address Management

- `POST /api/v1/addresses/` - Create a new address
- `GET /api/v1/addresses/:id` - Get address by ID
- `PUT /api/v1/addresses/:id` - Update address
- `DELETE /api/v1/addresses/:id` - Delete address
- `GET /api/v1/addresses/search` - Search addresses

### Role Management

- `POST /api/v1/roles/` - Create a new role
- `GET /api/v1/roles/:id` - Get role by ID
- `PUT /api/v1/roles/:id` - Update role
- `DELETE /api/v1/roles/:id` - Delete role
- `GET /api/v1/roles/` - List roles with pagination

## 🔒 Security Features

### Authentication & Authorization

- JWT-based authentication
- Role-based access control (RBAC)
- Token validation and refresh
- Secure password hashing with bcrypt

### Security Headers

- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- X-XSS-Protection: 1; mode=block
- Strict-Transport-Security
- Content-Security-Policy

### Rate Limiting

- IP-based rate limiting
- Configurable limits per endpoint
- Token bucket algorithm

## 📊 Performance Features

### Caching Strategy

- **User Data**: Cached for 5 minutes
- **Role Data**: Cached for 5 minutes
- **Address Data**: Cached for 5 minutes
- **Search Results**: Cached for 2 minutes

### Database Optimization

- Connection pooling
- Prepared statements
- Indexed queries
- Efficient pagination

### Request Handling

- Request timeouts (30 seconds)
- Graceful shutdown
- Connection reuse
- Response compression

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./services/...

# Run integration tests
go test -tags=integration ./...
```

### Test Structure

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete workflows

## 📈 Monitoring & Logging

### Logging

- Structured logging with Zap
- Request/response logging
- Error tracking with context
- Performance metrics

### Metrics

- Request duration
- Error rates
- Cache hit/miss ratios
- Database connection status

### Health Monitoring

- Database connectivity
- Cache connectivity
- Service status
- Resource usage

## 🔄 Deployment

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aaa-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: aaa-service
  template:
    metadata:
      labels:
        app: aaa-service
    spec:
      containers:
      - name: aaa-service
        image: aaa-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: POSTGRES_HOST
          valueFrom:
            configMapKeyRef:
              name: aaa-service-config
              key: postgres_host
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Code Style

- Follow Go conventions
- Use interfaces for dependency injection
- Write comprehensive tests
- Document public APIs
- Use meaningful commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:

- Create an issue in the repository
- Check the documentation
- Review the API examples

## 🔄 Migration Guide

### From Old Architecture

The refactored service maintains backward compatibility while providing:

- Better performance through caching
- Improved error handling
- Enhanced security features
- Comprehensive logging
- Better testability

### Breaking Changes

- Updated API response format
- New authentication requirements
- Changed database schema
- Modified configuration structure

## 📚 Additional Resources

- [API Documentation](docs/api.md)
- [Architecture Guide](docs/architecture.md)
- [Deployment Guide](docs/deployment.md)
- [Troubleshooting](docs/troubleshooting.md)
