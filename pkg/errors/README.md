# Error Handling System

This package provides a comprehensive error handling system that mimics Python's try-except pattern while maintaining Go's idiomatic error handling approach.

## Features

- **Custom Error Types**: Predefined error types for different scenarios
- **Error Wrapping**: Chain errors with additional context
- **Try-Catch Pattern**: Similar to Python's try-except functionality
- **Panic Recovery**: Automatic panic recovery with error conversion
- **HTTP Middleware**: Built-in HTTP error handling middleware
- **Error Handlers**: Register custom handlers for specific error types
- **Structured Logging**: Integration with structured logging

## Error Types

### Validation Errors
- `ErrorTypeValidation` - General validation errors
- `ErrorTypeInvalidInput` - Invalid input data
- `ErrorTypeMissingField` - Required fields missing
- `ErrorTypeInvalidFormat` - Incorrect data format

### Authentication & Authorization
- `ErrorTypeUnauthorized` - Unauthorized access
- `ErrorTypeForbidden` - Forbidden access
- `ErrorTypeAuthentication` - Authentication failures
- `ErrorTypeTokenExpired` - Expired tokens
- `ErrorTypeInvalidToken` - Invalid tokens

### Database Errors
- `ErrorTypeNotFound` - Resource not found
- `ErrorTypeAlreadyExists` - Resource already exists
- `ErrorTypeDatabaseError` - Database operation failures
- `ErrorTypeConnectionError` - Database connection issues
- `ErrorTypeConstraintViolation` - Database constraint violations

### Business Logic Errors
- `ErrorTypeBusinessRule` - Business rule violations
- `ErrorTypeInsufficientTokens` - Insufficient tokens
- `ErrorTypeUserInactive` - Inactive user accounts
- `ErrorTypeUserBlocked` - Blocked user accounts

### External Service Errors
- `ErrorTypeExternalService` - External service failures
- `ErrorTypeTimeout` - Operation timeouts
- `ErrorTypeRateLimit` - Rate limiting

### System Errors
- `ErrorTypeInternal` - Internal server errors
- `ErrorTypeConfiguration` - Configuration errors
- `ErrorTypeNotImplemented` - Unimplemented features

## Usage Examples

### Basic Error Creation

```go
// Create a simple error
err := errors.New(ErrorTypeValidation, "Invalid input data")

// Create error with formatting
err := errors.Newf(ErrorTypeNotFound, "User with ID %s not found", userID)

// Create domain-specific errors
err := errors.NewUserNotFoundError(userID)
err := errors.NewValidationError("username", "Username is required")
err := errors.NewInsufficientTokensError(100, 50)
```

### Error Wrapping

```go
// Wrap an existing error
dbErr := fmt.Errorf("connection timeout")
wrappedErr := errors.Wrap(dbErr, ErrorTypeDatabaseError, "Failed to create user")

// Add additional context
wrappedErr = wrappedErr.WithField("user_id").
    WithValue("12345").
    WithDetails(map[string]interface{}{
        "operation": "create_user",
        "retry_count": 3,
    })
```

### Try-Catch Pattern

```go
// Simple error handling
err := errors.Try(ctx, func() error {
    // Your code here
    if somethingWrong {
        return errors.NewValidationError("field", "Invalid value")
    }
    return nil
})

// Error handling with return values
result, err := errors.TryWithResult(ctx, func() (string, error) {
    // Your code here
    if somethingWrong {
        return "", errors.NewNotFoundError("resource", "id")
    }
    return "success", nil
})
```

### Panic Recovery

```go
// Recover from panics
defer errors.Recover(ctx)

// Recover with default result
result := errors.RecoverWithResult(ctx, "default_value")
```

### Error Type Checking

```go
// Check error types
if errors.IsErrorType(err, ErrorTypeValidation) {
    // Handle validation error
} else if errors.IsErrorType(err, ErrorTypeNotFound) {
    // Handle not found error
}

// Get error type
errorType := errors.GetErrorType(err)
```

### HTTP Error Handling

```go
// Create error handler
handler := errors.NewErrorHandler()

// Register custom handlers
handler.RegisterHandler(ErrorTypeValidation, func(ctx context.Context, err error) error {
    logger.Warn("Validation error occurred", zap.Error(err))
    return err
})

// Create middleware
middleware := errors.NewErrorMiddleware(logger, handler)

// Use in HTTP handlers
mux.HandleFunc("/users", errors.HandleWithErrorAndRecovery(func(w http.ResponseWriter, r *http.Request) error {
    // Your handler logic
    return errors.NewUserNotFoundError(userID)
}))
```

### Service Layer Example

```go
type UserService struct {
    logger *zap.Logger
}

func (s *UserService) CreateUser(ctx context.Context, username, password string) error {
    return errors.Try(ctx, func() error {
        // Validate input
        if username == "" {
            return errors.NewMissingFieldError("username")
        }

        if len(password) < 8 {
            return errors.NewInvalidInputError("password", password, "Password must be at least 8 characters")
        }

        // Check if user exists
        if s.userExists(username) {
            return errors.NewUserAlreadyExistsError("username", username)
        }

        // Create user
        if err := s.repository.Create(ctx, user); err != nil {
            return errors.Wrap(err, ErrorTypeDatabaseError, "Failed to create user")
        }

        s.logger.Info("User created successfully", zap.String("username", username))
        return nil
    })
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    return errors.TryWithResult(ctx, func() (*User, error) {
        if userID == "" {
            return nil, errors.NewMissingFieldError("user_id")
        }

        user, err := s.repository.GetByID(ctx, userID)
        if err != nil {
            return nil, errors.NewUserNotFoundError(userID)
        }

        return user, nil
    })
}
```

### Repository Layer Example

```go
type UserRepository struct {
    dbManager *db.Manager
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    return errors.TryWithResult(ctx, func() (*User, error) {
        var user User
        err := r.dbManager.GetDB().WithContext(ctx).
            Where("id = ?", id).
            First(&user).Error

        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return nil, errors.NewUserNotFoundError(id)
            }
            return nil, errors.Wrap(err, ErrorTypeDatabaseError, "Failed to get user")
        }

        return &user, nil
    })
}
```

### HTTP Handler Example

```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) error {
    // Parse request
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return errors.NewInvalidInputError("body", nil, "Invalid JSON format")
    }

    // Validate request
    if err := req.Validate(); err != nil {
        return err
    }

    // Create user
    if err := h.userService.CreateUser(r.Context(), req.Username, req.Password); err != nil {
        return err
    }

    // Return success response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})

    return nil
}
```

## Error Response Format

When using the HTTP middleware, errors are returned in the following JSON format:

```json
{
  "error": "VALIDATION_ERROR",
  "type": "VALIDATION_ERROR",
  "message": "field 'username' is required",
  "field": "username",
  "value": null,
  "details": {
    "resource": "user"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "path": "/api/v1/users",
  "method": "POST",
  "request_id": "req-12345"
}
```

## Best Practices

1. **Use Domain-Specific Errors**: Use the provided domain-specific error constructors
2. **Wrap External Errors**: Always wrap external library errors with context
3. **Add Relevant Details**: Include field names, values, and additional context
4. **Handle Errors at Boundaries**: Handle errors at service boundaries, not deep in the code
5. **Use Try-Catch Pattern**: Use the Try/TryWithResult functions for consistent error handling
6. **Register Error Handlers**: Register custom handlers for specific error types
7. **Log Errors Appropriately**: Use structured logging with error context

## Integration with Existing Code

To integrate this error handling system with existing code:

1. Replace `fmt.Errorf` with appropriate error constructors
2. Wrap external library errors with context
3. Use `errors.Try` and `errors.TryWithResult` for consistent error handling
4. Register error handlers for specific error types
5. Use the HTTP middleware for web applications
6. Update error checking to use `errors.IsErrorType`
