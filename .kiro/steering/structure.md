---
inclusion: fileMatch
fileMatchPattern: ["**/*.go", "**/go.mod", "**/go.sum"]
---

# Go Code Structure Guidelines

## File Organization Rules

### Critical Limits

- **300 lines maximum per Go file** - split before reaching this limit
- **Single responsibility** - each file has one focused purpose
- **Domain-based organization** - group related functionality together

### Service Structure Pattern

Always split services by operation:

```go
// internal/services/user/
// service.go - Service struct and constructor
type Service struct {
    userRepo     interfaces.UserRepository
    cacheService interfaces.CacheService
    logger       *zap.Logger
}

// create.go - User creation logic only
func (s *Service) CreateUser(ctx context.Context, req *requests.CreateUser) (*responses.UserResponse, error)

// read.go - User retrieval operations only
func (s *Service) GetUser(ctx context.Context, userID string) (*responses.UserResponse, error)
func (s *Service) ListUsers(ctx context.Context, filters *requests.UserFilters) (*responses.UserListResponse, error)

// update.go - User modification operations only
func (s *Service) UpdateUser(ctx context.Context, userID string, req *requests.UpdateUser) (*responses.UserResponse, error)

// delete.go - User deletion logic only
func (s *Service) DeleteUser(ctx context.Context, userID string) error
```

## Naming Conventions (Strict)

- **Files**: snake_case (`user_service.go`, `auth_handler.go`, `role_inheritance_engine.go`)
- **Packages**: lowercase (`users`, `auth`, `roles`, `organizations`)
- **Exported functions**: PascalCase (`CreateUser`, `ValidateToken`, `GetUserRoles`)
- **Private functions**: camelCase (`validateRequest`, `hashPassword`, `buildQuery`)
- **Variables**: camelCase (`userID`, `accessToken`, `organizationHierarchy`)
- **Constants**: UPPER_CASE (`JWT_SECRET`, `MAX_LOGIN_ATTEMPTS`, `DEFAULT_CACHE_TTL`)

## Required Patterns

### Database Operations

```go
// ALWAYS use kisanlink-db for database operations
func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

// Use transactions for multi-step operations
func (s *Service) CreateUserWithProfile(ctx context.Context, req *requests.CreateUserWithProfile) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // Multiple operations within transaction
    })
}
```

### Error Handling

```go
// Use custom error types from pkg/errors/
import "aaa-service/pkg/errors"

func (s *Service) GetUser(ctx context.Context, userID string) (*responses.UserResponse, error) {
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        s.logger.Error("failed to get user", zap.String("userID", userID), zap.Error(err))
        return nil, errors.NewNotFoundError("user not found")
    }
    return responses.ToUserResponse(user), nil
}
```

### Request Validation

```go
// Validate at handler boundaries using struct tags
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Phone    string `json:"phone" validate:"required,phone"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
}

func (h *Handler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, responses.NewErrorResponse("invalid request", err))
        return
    }

    if err := h.validator.Struct(&req); err != nil {
        c.JSON(400, responses.NewValidationErrorResponse(err))
        return
    }
}
```

### Context Usage

```go
// Always use context for timeouts and cancellation
func (s *Service) ProcessUserData(ctx context.Context, userID string) error {
    // Check context before expensive operations
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // Pass context to all downstream calls
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return err
    }

    return s.processUser(ctx, user)
}
```

## Import Organization (Required Order)

```go
package users

import (
    // 1. Standard library
    "context"
    "fmt"
    "time"

    // 2. Third-party packages
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "gorm.io/gorm"

    // 3. Local packages
    "aaa-service/internal/entities/models"
    "aaa-service/internal/entities/requests"
    "aaa-service/internal/entities/responses"
    "aaa-service/pkg/errors"
)
```

## Testing Requirements

```go
// Write tests alongside implementation files
// Use table-driven tests for multiple scenarios
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        request *requests.CreateUser
        want    *responses.UserResponse
        wantErr bool
    }{
        {
            name: "valid user creation",
            request: &requests.CreateUser{
                Email: "test@example.com",
                Name:  "Test User",
            },
            want: &responses.UserResponse{
                Email: "test@example.com",
                Name:  "Test User",
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## When to Split Files

- File approaches 250 lines (before 300 limit)
- Multiple responsibilities in single file
- Difficult to understand or test
- More than 5-7 public methods in a service file
