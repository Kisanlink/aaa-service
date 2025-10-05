# Enhanced JWT Context Implementation

## Overview

The AAA service now includes comprehensive organizational context in JWT tokens, providing rich information about user roles, organizations, groups, permissions, and scopes. This enhancement enables fine-grained authorization and multi-tenant support.

## JWT Token Structure

### Access Token Claims

```json
{
  // Standard JWT claims
  "sub": "user_id",
  "iss": "aaa-service",
  "aud": "aaa-clients",
  "iat": 1640995200,
  "nbf": 1640995200,
  "exp": 1640998800,
  "jti": "unique_token_id",

  // Enhanced user context
  "user_context": {
    "id": "user_id",
    "username": "john_doe",
    "phone_number": "1234567890",
    "country_code": "+91",
    "is_validated": true,
    "status": "active",
    "roles": [
      {
        "id": "role_id",
        "name": "admin",
        "scope": "ORG",
        "organization_id": "org_123",
        "group_id": null,
        "is_active": true,
        "organization": {
          "id": "org_123",
          "name": "Acme Corp"
        }
      }
    ],
    "organizations": [
      {
        "id": "org_123",
        "name": "Acme Corp"
      }
    ],
    "groups": [
      {
        "id": "group_456",
        "name": "Engineering Team",
        "organization_id": "org_123"
      }
    ]
  },

  // Security and session information
  "session_id": "session_unique_id",
  "token_type": "access",
  "token_version": "2.0",

  // Authorization context
  "permissions": [
    "user:read",
    "user:write",
    "org:manage"
  ],
  "scopes": [
    "role:admin",
    "org:org_123",
    "group:group_456",
    "scope:ORG"
  ],
  "tenant_context": {
    "organizations": ["org_123"],
    "groups": ["group_456"]
  },

  // Legacy fields (for backward compatibility)
  "user_id": "user_id",
  "username": "john_doe",
  "isvalidate": true,
  "roleIds": [...]
}
```

### Refresh Token Claims

Refresh tokens contain minimal information for security:

```json
{
  "sub": "user_id",
  "iss": "aaa-service",
  "aud": "aaa-clients",
  "iat": 1640995200,
  "exp": 1641600000,
  "jti": "unique_refresh_token_id",
  "token_type": "refresh",
  "token_version": "2.0",
  "session_id": "session_unique_id",
  "user_id": "user_id",
  "username": "john_doe",
  "isvalidate": true
}
```

## Usage Examples

### 1. Generating Tokens with Enhanced Context

```go
// In your service layer
userRoles := []models.UserRole{
    // ... populated user roles with organization/group context
}

accessToken, err := helper.GenerateAccessToken(
    userID,
    userRoles,
    username,
    isValidated,
)

refreshToken, err := helper.GenerateRefreshToken(
    userID,
    userRoles,
    username,
    isValidated,
)
```

### 2. Validating Tokens with Context

```go
// Basic validation (backward compatible)
userID, err := helper.ValidateToken(tokenString)

// Enhanced validation with full context
tokenContext, err := helper.ValidateTokenWithContext(tokenString)
if err != nil {
    return err
}

// Access user information
userID := tokenContext.UserID
permissions := tokenContext.Permissions
organizations := helper.GetUserOrganizations(tokenContext)
```

### 3. Authorization Checks

```go
// Check specific permission
if helper.HasPermission(tokenContext, "user:write") {
    // User can write users
}

// Check organization access
if helper.HasOrganizationAccess(tokenContext, "org_123") {
    // User has access to organization
}

// Check role
if helper.HasRole(tokenContext, "admin") {
    // User has admin role
}

// Check scope
if helper.HasScope(tokenContext, "org:org_123") {
    // User has organization scope
}
```

### 4. Middleware Integration

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // Validate token with full context
        tokenContext, err := helper.ValidateTokenWithContext(tokenString)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Check if token is expired
        if helper.IsTokenExpired(tokenContext) {
            c.JSON(401, gin.H{"error": "Token expired"})
            c.Abort()
            return
        }

        // Set context for downstream handlers
        c.Set("user_id", tokenContext.UserID)
        c.Set("token_context", tokenContext)
        c.Set("user_context", tokenContext.UserContext)

        c.Next()
    }
}
```

### 5. Organization-Scoped Operations

```go
func GetOrganizationUsers(c *gin.Context) {
    orgID := c.Param("org_id")
    tokenContext := c.MustGet("token_context").(*helper.TokenContext)

    // Check if user has access to this organization
    if !helper.HasOrganizationAccess(tokenContext, orgID) {
        c.JSON(403, gin.H{"error": "Access denied to organization"})
        return
    }

    // Check specific permission
    if !helper.HasPermission(tokenContext, "user:read") {
        c.JSON(403, gin.H{"error": "Insufficient permissions"})
        return
    }

    // Proceed with operation
    users, err := userService.GetUsersByOrganization(c.Request.Context(), orgID)
    // ...
}
```

## Key Features

### 1. Multi-Tenant Support

- Organization-scoped roles and permissions
- Group-based access control
- Hierarchical organization structure support

### 2. Enhanced Security

- Unique token identifiers (JTI)
- Session tracking
- Token versioning
- Comprehensive audit trail

### 3. Fine-Grained Authorization

- Permission-based access control
- Scope-based authorization
- Role hierarchy support
- Time-bound access (future enhancement)

### 4. Backward Compatibility

- Legacy fields maintained
- Existing validation functions work
- Gradual migration path

### 5. Performance Optimized

- Minimal refresh token payload
- Efficient permission checking
- Cached organization/group lookups

## Migration Guide

### From Legacy Tokens

1. **Update Token Generation**:

   ```go
   // Old way
   token, err := helper.GenerateAccessToken(userID, roles, username, validated)

   // New way (same signature, enhanced payload)
   token, err := helper.GenerateAccessToken(userID, roles, username, validated)
   ```

2. **Update Token Validation**:

   ```go
   // Old way (still works)
   userID, err := helper.ValidateToken(tokenString)

   // New way (enhanced)
   tokenContext, err := helper.ValidateTokenWithContext(tokenString)
   ```

3. **Update Authorization Logic**:

   ```go
   // Old way
   if hasRole(userRoles, "admin") {
       // ...
   }

   // New way
   if helper.HasRole(tokenContext, "admin") {
       // ...
   }
   ```

## Best Practices

1. **Use Enhanced Validation**: Always use `ValidateTokenWithContext` for new implementations
2. **Check Permissions**: Use specific permission checks rather than role checks
3. **Validate Organization Access**: Always verify organization/group access for scoped operations
4. **Handle Token Expiration**: Check token expiration and handle refresh appropriately
5. **Secure Refresh Tokens**: Store refresh tokens securely and implement proper rotation

## Security Considerations

1. **Token Size**: Enhanced tokens are larger; consider compression for network efficiency
2. **Sensitive Data**: Avoid including sensitive information in JWT payload
3. **Token Rotation**: Implement proper refresh token rotation
4. **Scope Validation**: Always validate scopes for cross-organization operations
5. **Audit Logging**: Log all token-based authorization decisions

## Future Enhancements

1. **Time-Bound Roles**: Support for roles with start/end dates
2. **Dynamic Permissions**: Runtime permission evaluation
3. **Token Compression**: Compress large token payloads
4. **Distributed Validation**: Support for distributed token validation
5. **Advanced Scoping**: More granular scope definitions
