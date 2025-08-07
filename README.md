# AAA Service - Identity & Access Management

A comprehensive Identity and Access Management (IAM) service built with Go, featuring role-based access control, permission management, and SpiceDB integration for advanced authorization.

## 🚀 Features

### Core IAM Functionality
- **User Management** - Complete user lifecycle with profile and contact management
- **Role-Based Access Control (RBAC)** - Hierarchical role system with permission inheritance
- **Permission Management** - Fine-grained permissions with resource-based access control
- **Authorization** - SpiceDB integration for real-time permission evaluation
- **Audit Logging** - Comprehensive audit trail for all operations

### Advanced Features
- **Multi-Factor Authentication (MFA)** - Framework ready for MFA implementation
- **Token Management** - JWT-based authentication with refresh tokens
- **Address Management** - Geocoding and address validation
- **Maintenance Mode** - System maintenance with admin bypass
- **Health Monitoring** - Comprehensive health checks and metrics

## 🏗️ Architecture

### Service Layer
```
HTTP API (Gin) → gRPC Services → Business Logic → Data Layer → Database
```

### Key Components
- **HTTP Handlers** - RESTful API endpoints
- **gRPC Services** - Internal service communication
- **SpiceDB** - Authorization and relationship management
- **PostgreSQL/DynamoDB** - Data persistence
- **Redis** - Caching and session management

## 📁 Project Structure

```
aaa-service/
├── cmd/                    # Application entry points
├── config/                 # Configuration management
├── database/              # Database layer
├── docs/                  # Documentation
├── entities/              # Domain models
│   ├── models/           # Data models
│   ├── requests/         # Request DTOs
│   └── responses/        # Response DTOs
├── grpc_server/          # gRPC service implementations
├── handlers/             # HTTP handlers
│   ├── admin/           # Admin operations
│   ├── auth/            # Authentication
│   ├── health/          # Health checks
│   ├── permissions/     # Permission management
│   ├── roles/           # Role management
│   ├── users/           # User management
│   └── addresses/       # Address management
├── interfaces/           # Interface definitions
├── middleware/           # HTTP middleware
├── pb/                   # Generated protobuf files
├── proto/                # Protobuf definitions
├── repositories/         # Data access layer
├── routes/              # Route definitions
├── scripts/             # Build and deployment scripts
├── server/              # Server implementations
├── services/            # Business logic layer
├── test/                # Test files
├── utils/               # Utility functions
├── spicedb_schema.zed  # SpiceDB authorization schema
└── README.md           # This file
```

## 🛠️ Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL (or DynamoDB)
- Redis
- SpiceDB

### Development Setup

1. **Clone and setup**
   ```bash
   git clone <repository-url>
   cd aaa-service
   cp env.example .env
   ```

2. **Start dependencies**
   ```bash
   docker-compose up -d
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Run the service**
   ```bash
   go run cmd/server/main.go
   ```

### Environment Configuration

```bash
# Database Configuration
DB_PROVIDER=postgres  # or dynamodb
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=aaa_service

# SpiceDB Configuration
DB_SPICEDB_ENDPOINT=localhost:50051
DB_SPICEDB_TOKEN=your_token

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your_jwt_secret
JWT_EXPIRY=24h
```

## 🔐 IAM Workflows

### User Management
```bash
# Create user
POST /api/v2/users
{
  "username": "john.doe",
  "email": "john@example.com",
  "password": "secure_password",
  "roles": ["user"]
}

# Get user with roles
GET /api/v2/users/{user_id}

# Update user
PUT /api/v2/users/{user_id}
```

### Role Management
```bash
# Create role
POST /api/v2/roles
{
  "name": "admin",
  "description": "Administrator role",
  "parent_id": null
}

# Assign permissions to role
POST /api/v2/roles/{role_id}/permissions
{
  "permission_ids": ["perm_123", "perm_456"]
}
```

### Permission Management
```bash
# Create permission
POST /api/v2/permissions
{
  "name": "read_users",
  "description": "Read user information",
  "resource": "users",
  "actions": ["read"]
}

# Evaluate permission
POST /api/v2/permissions/evaluate
{
  "user_id": "user_123",
  "resource": "users",
  "action": "read"
}
```

## 🔧 API Endpoints

### Authentication
- `POST /api/v2/auth/login` - User login
- `POST /api/v2/auth/logout` - User logout
- `POST /api/v2/auth/refresh` - Refresh token
- `POST /api/v2/auth/register` - User registration

### User Management
- `GET /api/v2/users` - List users
- `POST /api/v2/users` - Create user
- `GET /api/v2/users/{id}` - Get user
- `PUT /api/v2/users/{id}` - Update user
- `DELETE /api/v2/users/{id}` - Delete user

### Role Management
- `GET /api/v2/roles` - List roles
- `POST /api/v2/roles` - Create role
- `GET /api/v2/roles/{id}` - Get role
- `PUT /api/v2/roles/{id}` - Update role
- `DELETE /api/v2/roles/{id}` - Delete role

### Permission Management
- `GET /api/v2/permissions` - List permissions
- `POST /api/v2/permissions` - Create permission
- `GET /api/v2/permissions/{id}` - Get permission
- `PUT /api/v2/permissions/{id}` - Update permission
- `DELETE /api/v2/permissions/{id}` - Delete permission

### Admin Operations
- `GET /api/v2/admin/health` - Health check
- `GET /api/v2/admin/metrics` - System metrics
- `GET /api/v2/admin/audit-logs` - Audit logs

## 🧪 Testing

### Run Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./handlers -v
```

### Integration Tests
```bash
# Run integration tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## 🚀 Deployment

### Docker Deployment
```bash
# Build image
docker build -t aaa-service .

# Run container
docker run -p 8080:8080 aaa-service
```

### Kubernetes Deployment
```bash
# Apply manifests
kubectl apply -f k8s/
```

## 📊 Monitoring

### Health Checks
- `GET /health` - Basic health check
- `GET /api/v2/admin/health/detailed` - Detailed health status

### Metrics
- `GET /api/v2/admin/metrics` - System metrics
- Prometheus metrics available at `/metrics`

## 🔒 Security Features

### Authentication
- JWT-based authentication
- Refresh token rotation
- Token blacklisting
- Password hashing with bcrypt

### Authorization
- SpiceDB integration for fine-grained permissions
- Role-based access control
- Resource-based permissions
- Real-time permission evaluation

### Audit & Compliance
- Comprehensive audit logging
- User action tracking
- Data access monitoring
- Compliance reporting

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make check` to ensure code quality
6. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For questions and support:
- Create an issue in the GitHub repository
- Check the [documentation](docs/)
- Review the [API documentation](docs/swagger.json)

## 🗺️ Roadmap

### Planned Features
- [ ] Complete MFA implementation
- [ ] Advanced analytics dashboard
- [ ] GraphQL API layer
- [ ] Real-time notifications
- [ ] Advanced audit analytics
- [ ] Multi-tenant support

### Performance Optimizations
- [ ] Connection pooling
- [ ] Query optimization
- [ ] Caching improvements
- [ ] Load balancing
- [ ] Auto-scaling

---

**AAA Service** - Enterprise-grade Identity and Access Management
