# Response Transformation Infrastructure

This document describes the response transformation infrastructure implemented for the AAA Service, which provides standardized, consistent, and optimized API responses.

## Overview

The response transformation infrastructure consists of several key components that work together to:

1. **Standardize Response Structures**: Ensure all API responses follow consistent field naming (snake_case) and structure
2. **Control Response Content**: Allow clients to specify which nested objects to include/exclude via query parameters
3. **Validate Response Consistency**: Automatically validate responses for consistency and security
4. **Optimize Performance**: Reduce payload sizes by excluding unnecessary data

## Core Components

### 1. ResponseTransformer Interface

The `ResponseTransformer` interface provides methods to transform domain models into standardized response structures.

```go
type ResponseTransformer interface {
    TransformUser(user *models.User, options TransformOptions) interface{}
    TransformUsers(users []models.User, options TransformOptions) interface{}
    TransformRole(role *models.Role, options TransformOptions) interface{}
    // ... other transformation methods
}
```

**Key Features:**

- Transforms models to standardized response structures
- Applies transformation options for conditional field inclusion
- Masks sensitive data (e.g., Aadhaar numbers)
- Ensures consistent state representation

### 2. QueryParameterHandler Interface

The `QueryParameterHandler` interface manages query parameter parsing and validation.

```go
type QueryParameterHandler interface {
    ParseTransformOptions(c *gin.Context) TransformOptions
    ValidateQueryParameters(c *gin.Context) error
    GetDefaultOptions() TransformOptions
}
```

**Supported Query Parameters:**

| Parameter            | Type    | Description                         | Example                    |
| -------------------- | ------- | ----------------------------------- | -------------------------- |
| `include_profile`    | boolean | Include user profile data           | `?include_profile=true`    |
| `include_contacts`   | boolean | Include user contacts               | `?include_contacts=1`      |
| `include_role`       | boolean | Include role information            | `?include_role=yes`        |
| `include_user`       | boolean | Include user data in role responses | `?include_user=true`       |
| `include_address`    | boolean | Include address information         | `?include_address=true`    |
| `exclude_deleted`    | boolean | Exclude soft-deleted records        | `?exclude_deleted=true`    |
| `exclude_inactive`   | boolean | Exclude inactive records            | `?exclude_inactive=true`   |
| `only_active_roles`  | boolean | Only include active roles           | `?only_active_roles=true`  |
| `mask_sensitive`     | boolean | Mask sensitive data                 | `?mask_sensitive=true`     |
| `include_timestamps` | boolean | Include timestamp fields            | `?include_timestamps=true` |

**Legacy Support:**

- `include=profile,contacts,role` - Comma-separated list of includes
- `include=all` - Include all available nested objects

### 3. ResponseValidator Interface

The `ResponseValidator` interface ensures response consistency and security.

```go
type ResponseValidator interface {
    ValidateUserResponse(response interface{}) error
    ValidateRoleResponse(response interface{}) error
    ValidateUserRoleResponse(response interface{}) error
    ValidateResponseConsistency(responses []interface{}) error
    ValidateNoSensitiveData(response interface{}) error
}
```

**Validation Features:**

- Field naming consistency (snake_case)
- State consistency (is_active vs deleted_at)
- Sensitive data exclusion
- Response structure consistency

### 4. TransformOptions Structure

The `TransformOptions` struct controls how responses are transformed.

```go
type TransformOptions struct {
    // Include flags for nested objects
    IncludeProfile   bool
    IncludeContacts  bool
    IncludeRole      bool
    IncludeUser      bool
    IncludeAddress   bool

    // Exclusion flags
    ExcludeDeleted   bool
    ExcludeInactive  bool
    OnlyActiveRoles  bool

    // Field control
    MaskSensitiveData bool
    IncludeTimestamps bool
}
```

## Standardized Response Structures

### User Response

```json
{
  "id": "user123",
  "phone_number": "1234567890",
  "country_code": "+91",
  "username": "testuser",
  "is_validated": true,
  "is_active": true,
  "status": "active",
  "tokens": 100,
  "has_mpin": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "deleted_at": null,
  "profile": {
    "id": "profile123",
    "user_id": "user123",
    "name": "Test User",
    "aadhaar_number": "XXXX-XXXX-1234"
  },
  "contacts": [...],
  "roles": [...]
}
```

### Role Response

```json
{
  "id": "role123",
  "name": "admin",
  "description": "Administrator role",
  "scope": "organization",
  "is_active": true,
  "version": 1,
  "organization_id": "org123",
  "group_id": null,
  "parent_id": null,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "deleted_at": null,
  "permissions": [...],
  "children": [...]
}
```

### Paginated Response

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  },
  "request_id": "req-123",
  "metadata": {
    "processing_time_ms": 45,
    "api_version": "v1"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "success": true
}
```

## Usage Examples

### Basic Handler Implementation

```go
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")

    // Get transformation options from middleware
    options := middleware.GetTransformOptions(c)

    // Get user from service
    user, err := h.userService.GetUserByID(c.Request.Context(), userID)
    if err != nil {
        h.responseUtils.SendErrorResponse(c, http.StatusInternalServerError,
            "USER_FETCH_ERROR", "Failed to retrieve user", nil)
        return
    }

    // Transform and send response
    transformedUser := h.transformer.TransformUser(user, options)
    h.responseUtils.SendSuccessResponse(c, http.StatusOK, transformedUser,
        "User retrieved successfully")
}
```

### Middleware Setup

```go
func SetupRoutes(router *gin.Engine, queryHandler interfaces.QueryParameterHandler, logger interfaces.Logger) {
    // Create response transformation middleware
    rtMiddleware := middleware.NewResponseTransformationMiddleware(queryHandler, logger)

    // Apply middleware to routes
    userRoutes := router.Group("/api/v1/users")
    userRoutes.Use(rtMiddleware.ValidateQueryParameters())
    {
        userRoutes.GET("/:id", handler.GetUser)
        userRoutes.GET("", handler.ListUsers)
    }
}
```

## Query Parameter Examples

### Include Specific Data

```bash
# Include user profile
GET /api/v1/users/123?include_profile=true

# Include multiple nested objects
GET /api/v1/users/123?include_profile=true&include_contacts=true&include_role=true

# Legacy format
GET /api/v1/users/123?include=profile,contacts,role
```

### Exclude Data

```bash
# Exclude deleted users
GET /api/v1/users?exclude_deleted=true

# Exclude inactive users
GET /api/v1/users?exclude_inactive=true

# Only active roles
GET /api/v1/users/123/roles?only_active_roles=true
```

### Pagination and Filtering

```bash
# Paginated results
GET /api/v1/users?limit=50&offset=100

# Page-based pagination
GET /api/v1/users?page=3&limit=20

# Search and filter
GET /api/v1/users?search=john&status=active&is_validated=true
```

## Security Features

### Sensitive Data Masking

The infrastructure automatically masks sensitive data:

- **Aadhaar Numbers**: `123456789012` → `XXXX-XXXX-9012`
- **Passwords**: Never included in responses
- **MPIN**: Never included in responses (only `has_mpin` flag)

### Validation

All responses are validated for:

- **Field Naming**: Ensures snake_case consistency
- **State Consistency**: Validates is_active vs deleted_at logic
- **Sensitive Data**: Prevents accidental exposure of sensitive fields
- **Structure Consistency**: Ensures consistent response structures

## Performance Optimizations

### Conditional Loading

- Only load and include requested nested objects
- Reduce database queries for unused data
- Minimize response payload sizes

### Caching Support

The infrastructure supports response caching:

```go
// Cache transformed responses
cacheKey := fmt.Sprintf("user:%s:options:%s", userID, optionsHash)
if cached := cache.Get(cacheKey); cached != nil {
    return cached
}

transformed := transformer.TransformUser(user, options)
cache.Set(cacheKey, transformed, 300) // 5 minutes TTL
```

## Testing

### Unit Tests

```go
func TestResponseTransformer_TransformUser(t *testing.T) {
    transformer := NewResponseTransformer()

    user := &models.User{
        PhoneNumber: "1234567890",
        CountryCode: "+91",
        // ... other fields
    }

    options := interfaces.TransformOptions{
        IncludeProfile: true,
        MaskSensitiveData: true,
    }

    result := transformer.TransformUser(user, options)

    // Assertions
    assert.NotNil(t, result)
    // ... more assertions
}
```

### Integration Tests

```go
func TestUserHandler_GetUser(t *testing.T) {
    // Setup test server with middleware
    router := gin.New()
    rtMiddleware := middleware.NewResponseTransformationMiddleware(queryHandler, logger)
    router.Use(rtMiddleware.ValidateQueryParameters())
    router.GET("/users/:id", handler.GetUser)

    // Test request with query parameters
    req := httptest.NewRequest("GET", "/users/123?include_profile=true", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Assertions
    assert.Equal(t, http.StatusOK, w.Code)
    // ... validate response structure
}
```

## Migration Guide

### From Existing Endpoints

1. **Add Middleware**: Apply `ValidateQueryParameters()` middleware to routes
2. **Update Handlers**: Use `GetTransformOptions()` to get transformation options
3. **Transform Responses**: Use `ResponseTransformer` to transform data
4. **Update Tests**: Add tests for query parameter functionality

### Backward Compatibility

- Legacy `include` parameter is supported
- Default options maintain existing behavior
- Gradual migration path available

## Best Practices

1. **Always Use Middleware**: Apply query parameter validation middleware to all routes
2. **Validate Responses**: Use ResponseValidator to ensure consistency
3. **Cache Transformed Data**: Cache expensive transformations when possible
4. **Test Query Parameters**: Include query parameter tests in your test suite
5. **Document Parameters**: Document supported query parameters in API documentation
6. **Monitor Performance**: Monitor response times and payload sizes

## Troubleshooting

### Common Issues

1. **Invalid Query Parameters**: Check parameter names and values
2. **Missing Nested Data**: Verify include flags are set correctly
3. **Sensitive Data Exposure**: Ensure MaskSensitiveData is enabled
4. **Performance Issues**: Check if unnecessary nested objects are being loaded

### Debug Logging

Enable debug logging to troubleshoot issues:

```go
logger.Debug("Transform options", zap.Any("options", options))
logger.Debug("Response validation", zap.Any("response", response))
```

## Future Enhancements

1. **GraphQL-style Field Selection**: More granular field selection
2. **Response Compression**: Automatic response compression
3. **Advanced Caching**: Intelligent cache invalidation
4. **Metrics Collection**: Response transformation metrics
5. **Schema Validation**: JSON schema validation for responses
