# AAA Service Error Response Documentation

This document provides comprehensive documentation for all error responses in the AAA Service API, including enhanced login, role management, and MPIN management endpoints.

## Table of Contents

1. [Error Response Format](#error-response-format)
2. [HTTP Status Codes](#http-status-codes)
3. [Error Types](#error-types)
4. [Authentication Errors](#authentication-errors)
5. [Role Management Errors](#role-management-errors)
6. [MPIN Management Errors](#mpin-management-errors)
7. [User Management Errors](#user-management-errors)
8. [Validation Errors](#validation-errors)
9. [System Errors](#system-errors)

## Error Response Format

All error responses follow a consistent structure:

```json
{
  "success": false,
  "error": "ERROR_TYPE",
  "message": "Human-readable error message",
  "code": 400,
  "details": {
    "additional": "error details"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Fields Description

- **success**: Always `false` for error responses
- **error**: Machine-readable error type identifier
- **message**: Human-readable error description
- **code**: HTTP status code
- **details**: Additional error context (optional)
- **timestamp**: ISO 8601 timestamp when the error occurred
- **request_id**: Unique request identifier for tracing (optional)

## HTTP Status Codes

| Code | Status                | Description                                       |
| ---- | --------------------- | ------------------------------------------------- |
| 400  | Bad Request           | Invalid request data or validation error          |
| 401  | Unauthorized          | Authentication required or invalid credentials    |
| 403  | Forbidden             | Insufficient permissions for the requested action |
| 404  | Not Found             | Requested resource does not exist                 |
| 409  | Conflict              | Resource conflict (e.g., duplicate assignment)    |
| 422  | Unprocessable Entity  | Valid request format but semantic errors          |
| 429  | Too Many Requests     | Rate limit exceeded                               |
| 500  | Internal Server Error | Unexpected server error                           |
| 503  | Service Unavailable   | Service temporarily unavailable                   |

## Error Types

### VALIDATION_ERROR

Invalid input data or request format violations.

### AUTHENTICATION_ERROR

Authentication failures including invalid credentials.

### AUTHORIZATION_ERROR

Insufficient permissions for the requested operation.

### NOT_FOUND_ERROR

Requested resource does not exist.

### CONFLICT_ERROR

Resource conflicts such as duplicate assignments.

### RATE_LIMIT_EXCEEDED

Too many requests within the rate limit window.

### INTERNAL_SERVER_ERROR

Unexpected system errors.

### SERVICE_UNAVAILABLE

Service temporarily unavailable.

## Authentication Errors

### Invalid Login Credentials

**Endpoint:** `POST /api/v1/auth/login`

**Error Response:**

```json
{
  "success": false,
  "error": "AUTHENTICATION_ERROR",
  "message": "Invalid credentials",
  "code": 401,
  "details": {
    "authentication_method": "password",
    "attempts_remaining": 2
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Invalid MPIN

**Endpoint:** `POST /api/v1/auth/login`

**Error Response:**

```json
{
  "success": false,
  "error": "AUTHENTICATION_ERROR",
  "message": "Invalid MPIN",
  "code": 401,
  "details": {
    "authentication_method": "mpin",
    "attempts_remaining": 2,
    "lockout_duration": "15 minutes"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### User Not Found

**Endpoint:** `POST /api/v1/auth/login`

**Error Response:**

```json
{
  "success": false,
  "error": "AUTHENTICATION_ERROR",
  "message": "Invalid credentials",
  "code": 401,
  "details": {
    "reason": "user_not_found"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Account Locked

**Endpoint:** `POST /api/v1/auth/login`

**Error Response:**

```json
{
  "success": false,
  "error": "AUTHENTICATION_ERROR",
  "message": "Account temporarily locked due to multiple failed attempts",
  "code": 401,
  "details": {
    "locked_until": "2024-01-01T12:15:00Z",
    "reason": "multiple_failed_attempts",
    "unlock_duration": "15 minutes"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Invalid Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

**Error Response:**

```json
{
  "success": false,
  "error": "AUTHENTICATION_ERROR",
  "message": "Invalid refresh token",
  "code": 401,
  "details": {
    "reason": "token_expired",
    "expired_at": "2024-01-01T11:00:00Z"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Missing Authentication

**Endpoint:** Protected endpoints without Authorization header

**Error Response:**

```json
{
  "success": false,
  "error": "AUTHENTICATION_ERROR",
  "message": "Authentication required",
  "code": 401,
  "details": {
    "required_header": "Authorization",
    "format": "Bearer <token>"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## Role Management Errors

### Role Not Found

**Endpoint:** `POST /api/v1/users/{user_id}/roles`

**Error Response:**

```json
{
  "success": false,
  "error": "NOT_FOUND_ERROR",
  "message": "Role not found",
  "code": 404,
  "details": {
    "resource": "role",
    "resource_id": "ROLE123456789"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### User Not Found for Role Assignment

**Endpoint:** `POST /api/v1/users/{user_id}/roles`

**Error Response:**

```json
{
  "success": false,
  "error": "NOT_FOUND_ERROR",
  "message": "User not found",
  "code": 404,
  "details": {
    "resource": "user",
    "resource_id": "USER123456789"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Role Already Assigned

**Endpoint:** `POST /api/v1/users/{user_id}/roles`

**Error Response:**

```json
{
  "success": false,
  "error": "CONFLICT_ERROR",
  "message": "Role already assigned to user",
  "code": 409,
  "details": {
    "user_id": "USER123456789",
    "role_id": "ROLE123456789",
    "existing_assignment_id": "USERROLE123456789",
    "assigned_at": "2024-01-01T10:00:00Z"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Role Assignment Not Found

**Endpoint:** `DELETE /api/v1/users/{user_id}/roles/{role_id}`

**Error Response:**

```json
{
  "success": false,
  "error": "NOT_FOUND_ERROR",
  "message": "Role assignment not found",
  "code": 404,
  "details": {
    "user_id": "USER123456789",
    "role_id": "ROLE123456789",
    "resource": "user_role_assignment"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Insufficient Permissions for Role Management

**Endpoint:** `POST /api/v1/users/{user_id}/roles`

**Error Response:**

```json
{
  "success": false,
  "error": "AUTHORIZATION_ERROR",
  "message": "Insufficient permissions to assign roles",
  "code": 403,
  "details": {
    "required_permission": "roles:assign",
    "user_permissions": ["users:read", "users:update"],
    "missing_permissions": ["roles:assign"]
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Role Name Conflict

**Endpoint:** `POST /api/v1/roles`

**Error Response:**

```json
{
  "success": false,
  "error": "CONFLICT_ERROR",
  "message": "Role name already exists",
  "code": 409,
  "details": {
    "field": "name",
    "value": "admin",
    "existing_role_id": "ROLE987654321"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## MPIN Management Errors

### MPIN Not Set

**Endpoint:** `POST /api/v1/auth/update-mpin`

**Error Response:**

```json
{
  "success": false,
  "error": "NOT_FOUND_ERROR",
  "message": "MPIN not set for user",
  "code": 404,
  "details": {
    "user_id": "USER123456789",
    "action": "update_mpin",
    "suggestion": "Use set-mpin endpoint to set initial MPIN"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Invalid Current MPIN

**Endpoint:** `POST /api/v1/auth/update-mpin`

**Error Response:**

```json
{
  "success": false,
  "error": "AUTHENTICATION_ERROR",
  "message": "Invalid current MPIN",
  "code": 401,
  "details": {
    "field": "current_mpin",
    "attempts_remaining": 2,
    "lockout_warning": "Account will be locked after 3 failed attempts"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Invalid Password for MPIN Set

**Endpoint:** `POST /api/v1/auth/set-mpin`

**Error Response:**

```json
{
  "success": false,
  "error": "AUTHENTICATION_ERROR",
  "message": "Invalid password",
  "code": 401,
  "details": {
    "field": "password",
    "required_for": "mpin_verification"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### MPIN Format Validation Error

**Endpoint:** `POST /api/v1/auth/set-mpin`

**Error Response:**

```json
{
  "success": false,
  "error": "VALIDATION_ERROR",
  "message": "Invalid MPIN format",
  "code": 400,
  "details": {
    "field": "mpin",
    "errors": [
      "MPIN must be 4 or 6 digits",
      "MPIN must contain only numeric characters"
    ],
    "provided_length": 3,
    "allowed_lengths": [4, 6]
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Same MPIN Error

**Endpoint:** `POST /api/v1/auth/update-mpin`

**Error Response:**

```json
{
  "success": false,
  "error": "VALIDATION_ERROR",
  "message": "New MPIN must be different from current MPIN",
  "code": 400,
  "details": {
    "field": "new_mpin",
    "reason": "same_as_current"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## User Management Errors

### User Not Found for Deletion

**Endpoint:** `DELETE /api/v1/users/{user_id}`

**Error Response:**

```json
{
  "success": false,
  "error": "NOT_FOUND_ERROR",
  "message": "User not found",
  "code": 404,
  "details": {
    "resource": "user",
    "resource_id": "USER123456789"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### User Already Deleted

**Endpoint:** `DELETE /api/v1/users/{user_id}`

**Error Response:**

```json
{
  "success": false,
  "error": "CONFLICT_ERROR",
  "message": "User is already deleted",
  "code": 409,
  "details": {
    "user_id": "USER123456789",
    "deleted_at": "2024-01-01T10:00:00Z",
    "deleted_by": "ADMIN987654321"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Cannot Delete Admin User

**Endpoint:** `DELETE /api/v1/users/{user_id}`

**Error Response:**

```json
{
  "success": false,
  "error": "AUTHORIZATION_ERROR",
  "message": "Cannot delete users with critical admin roles",
  "code": 403,
  "details": {
    "user_id": "USER123456789",
    "critical_roles": ["super_admin", "system_admin"],
    "reason": "protection_policy"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### User Registration Conflict

**Endpoint:** `POST /api/v1/auth/register`

**Error Response:**

```json
{
  "success": false,
  "error": "CONFLICT_ERROR",
  "message": "User already exists with this phone number",
  "code": 409,
  "details": {
    "field": "phone_number",
    "value": "+1234567890",
    "existing_user_id": "USER987654321"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## Validation Errors

### Missing Required Fields

**Endpoint:** `POST /api/v1/auth/login`

**Error Response:**

```json
{
  "success": false,
  "error": "VALIDATION_ERROR",
  "message": "Missing required fields",
  "code": 400,
  "details": {
    "missing_fields": ["phone_number", "country_code"],
    "provided_fields": ["password"]
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Invalid Field Format

**Endpoint:** `POST /api/v1/auth/register`

**Error Response:**

```json
{
  "success": false,
  "error": "VALIDATION_ERROR",
  "message": "Invalid field format",
  "code": 400,
  "details": {
    "field_errors": {
      "phone_number": "Phone number must be 10 digits",
      "country_code": "Country code must start with + and contain 1-4 digits",
      "password": "Password must be at least 8 characters long"
    }
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Invalid JSON Format

**Endpoint:** Any endpoint with JSON body

**Error Response:**

```json
{
  "success": false,
  "error": "VALIDATION_ERROR",
  "message": "Invalid JSON format",
  "code": 400,
  "details": {
    "json_error": "invalid character '}' looking for beginning of object key string",
    "position": 45
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Authentication Method Missing

**Endpoint:** `POST /api/v1/auth/login`

**Error Response:**

```json
{
  "success": false,
  "error": "VALIDATION_ERROR",
  "message": "Either password or MPIN is required",
  "code": 400,
  "details": {
    "available_methods": ["password", "mpin"],
    "provided_methods": []
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## System Errors

### Database Connection Error

**Error Response:**

```json
{
  "success": false,
  "error": "INTERNAL_SERVER_ERROR",
  "message": "Database connection error",
  "code": 500,
  "details": {
    "error_id": "ERR123456789",
    "component": "database",
    "retry_after": 30
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Service Unavailable

**Error Response:**

```json
{
  "success": false,
  "error": "SERVICE_UNAVAILABLE",
  "message": "Service temporarily unavailable",
  "code": 503,
  "details": {
    "reason": "maintenance",
    "estimated_duration": "30 minutes",
    "retry_after": 1800
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Rate Limit Exceeded

**Error Response:**

```json
{
  "success": false,
  "error": "RATE_LIMIT_EXCEEDED",
  "message": "Too many requests. Please try again later.",
  "code": 429,
  "details": {
    "limit": 100,
    "window": "1 hour",
    "reset_time": "2024-01-01T13:00:00Z",
    "retry_after": 3600
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Timeout Error

**Error Response:**

```json
{
  "success": false,
  "error": "INTERNAL_SERVER_ERROR",
  "message": "Request timeout",
  "code": 500,
  "details": {
    "timeout_duration": "30 seconds",
    "component": "external_service",
    "suggestion": "Retry with smaller request size"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## Error Handling Best Practices

### Client-Side Error Handling

1. **Always check the `success` field** to determine if the request was successful
2. **Use the `error` field** for programmatic error handling
3. **Display the `message` field** to users for human-readable errors
4. **Check `details`** for additional context and specific error information
5. **Implement retry logic** for 5xx errors with exponential backoff
6. **Store `request_id`** for support and debugging purposes

### Example Error Handling Code

```javascript
async function handleApiResponse(response) {
  const data = await response.json();

  if (data.success) {
    return data.data;
  }

  // Handle specific error types
  switch (data.error) {
    case "AUTHENTICATION_ERROR":
      // Redirect to login or refresh token
      handleAuthError(data);
      break;

    case "VALIDATION_ERROR":
      // Show validation errors to user
      showValidationErrors(data.details);
      break;

    case "RATE_LIMIT_EXCEEDED":
      // Implement retry with delay
      scheduleRetry(data.details.retry_after);
      break;

    case "INTERNAL_SERVER_ERROR":
      // Log error and show generic message
      logError(data.request_id, data.details);
      showGenericError();
      break;

    default:
      // Handle unknown errors
      showGenericError(data.message);
  }

  throw new Error(data.message);
}
```

### Logging and Monitoring

- **Log all error responses** with their `request_id` for debugging
- **Monitor error rates** by error type and endpoint
- **Set up alerts** for high error rates or specific critical errors
- **Track authentication failures** for security monitoring
- **Monitor rate limit hits** to adjust limits or identify abuse

### Support and Debugging

When contacting support, always include:

- The `request_id` from the error response
- The full error response JSON
- The request that caused the error
- Timestamp of when the error occurred
- Any relevant user or session information
