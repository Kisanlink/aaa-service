# AAA Service API Usage Examples

This document provides comprehensive examples for using the AAA Service API endpoints, including the enhanced login functionality, role management, and MPIN management features.

## Table of Contents

1. [Authentication Endpoints](#authentication-endpoints)
2. [Role Management Endpoints](#role-management-endpoints)
3. [MPIN Management Endpoints](#mpin-management-endpoints)
4. [User Management Endpoints](#user-management-endpoints)
5. [Error Handling Examples](#error-handling-examples)

## Authentication Endpoints

### Enhanced Login with Password

**Endpoint:** `POST /api/v1/auth/login`

**Description:** Login with phone number and password, optionally including additional user data.

```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+1234567890",
    "country_code": "US",
    "password": "securePassword123",
    "include_profile": true,
    "include_roles": true,
    "include_contacts": false
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": "USER123456789",
      "username": "john_doe",
      "phone_number": "+1234567890",
      "country_code": "US",
      "is_validated": true,
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z",
      "tokens": 100,
      "has_mpin": true,
      "roles": [
        {
          "id": "USERROLE123456789",
          "user_id": "USER123456789",
          "role_id": "ROLE123456789",
          "role": {
            "id": "ROLE123456789",
            "name": "user",
            "description": "Standard user role",
            "scope": "organization",
            "is_active": true,
            "version": 1
          },
          "is_active": true
        }
      ],
      "profile": {
        "id": "PROFILE123456789",
        "name": "John Doe",
        "date_of_birth": "1990-01-01",
        "address": {
          "id": "ADDR123456789",
          "house": "123",
          "street": "Main Street",
          "district": "Metro",
          "state": "State",
          "country": "Country",
          "pincode": "123456"
        }
      }
    }
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Enhanced Login with MPIN

**Endpoint:** `POST /api/v1/auth/login`

**Description:** Login with phone number and MPIN instead of password.

```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+1234567890",
    "country_code": "US",
    "mpin": "1234",
    "include_roles": true
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": "USER123456789",
      "username": "john_doe",
      "phone_number": "+1234567890",
      "country_code": "US",
      "is_validated": true,
      "has_mpin": true,
      "roles": [
        {
          "id": "USERROLE123456789",
          "user_id": "USER123456789",
          "role_id": "ROLE123456789",
          "role": {
            "id": "ROLE123456789",
            "name": "user",
            "description": "Standard user role",
            "scope": "organization",
            "is_active": true,
            "version": 1
          },
          "is_active": true
        }
      ]
    }
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### User Registration

**Endpoint:** `POST /api/v1/auth/register`

**Description:** Register a new user account with optional profile information.

```bash
curl -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+1234567890",
    "country_code": "US",
    "password": "securePassword123",
    "username": "john_doe",
    "name": "John Doe",
    "aadhaar_number": "123456789012"
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": "USER123456789",
      "username": "john_doe",
      "phone_number": "+1234567890",
      "country_code": "US",
      "is_validated": false,
      "status": "active",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z",
      "tokens": 0,
      "has_mpin": false
    }
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Token Refresh

**Endpoint:** `POST /api/v1/auth/refresh`

**Description:** Refresh access token using refresh token and MPIN verification.

```bash
curl -X POST "http://localhost:8080/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "mpin": "1234"
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### User Logout

**Endpoint:** `POST /api/v1/auth/logout`

**Description:** Logout user and invalidate tokens.

```bash
curl -X POST "http://localhost:8080/api/v1/auth/logout" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

**Response:**

```json
{
  "success": true,
  "message": "Logout successful",
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## Role Management Endpoints

### Assign Role to User

**Endpoint:** `POST /api/v1/users/{user_id}/roles`

**Description:** Assign a role to an existing user.

```bash
curl -X POST "http://localhost:8080/api/v1/users/USER123456789/roles" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": "ROLE123456789"
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "Role assigned successfully",
  "user_id": "USER123456789",
  "role": {
    "id": "ROLE123456789",
    "name": "moderator",
    "description": "Moderator role with limited admin access",
    "scope": "organization",
    "is_active": true,
    "version": 1
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Remove Role from User

**Endpoint:** `DELETE /api/v1/users/{user_id}/roles/{role_id}`

**Description:** Remove a role assignment from a user.

```bash
curl -X DELETE "http://localhost:8080/api/v1/users/USER123456789/roles/ROLE123456789" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

**Response:**

```json
{
  "success": true,
  "message": "Role removed successfully",
  "user_id": "USER123456789",
  "role_id": "ROLE123456789",
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Create Role

**Endpoint:** `POST /api/v1/roles`

**Description:** Create a new role.

```bash
curl -X POST "http://localhost:8080/api/v1/roles" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "content_moderator",
    "description": "Content moderation role with specific permissions"
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "Role created successfully",
  "data": {
    "id": "ROLE987654321",
    "name": "content_moderator",
    "description": "Content moderation role with specific permissions",
    "scope": "organization",
    "is_active": true,
    "version": 1,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Get Role Details

**Endpoint:** `GET /api/v1/roles/{role_id}`

**Description:** Retrieve detailed information about a specific role.

```bash
curl -X GET "http://localhost:8080/api/v1/roles/ROLE123456789" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

**Response:**

```json
{
  "success": true,
  "message": "Role retrieved successfully",
  "data": {
    "role": {
      "id": "ROLE123456789",
      "name": "moderator",
      "description": "Moderator role with limited admin access",
      "scope": "organization",
      "is_active": true,
      "version": 1,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### List Roles

**Endpoint:** `GET /api/v1/roles`

**Description:** Get a paginated list of roles.

```bash
curl -X GET "http://localhost:8080/api/v1/roles?limit=10&offset=0" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

**Response:**

```json
{
  "success": true,
  "message": "Roles retrieved successfully",
  "data": {
    "roles": [
      {
        "id": "ROLE123456789",
        "name": "admin",
        "description": "Administrator role with full access",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      },
      {
        "id": "ROLE987654321",
        "name": "user",
        "description": "Standard user role",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "pagination": {
    "page": 1,
    "per_page": 10,
    "total": 2,
    "total_pages": 1
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## MPIN Management Endpoints

### Set MPIN

**Endpoint:** `POST /api/v1/auth/set-mpin`

**Description:** Set MPIN for the authenticated user with password verification.

```bash
curl -X POST "http://localhost:8080/api/v1/auth/set-mpin" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "mpin": "1234",
    "password": "securePassword123"
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "MPIN set successfully",
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Update MPIN

**Endpoint:** `POST /api/v1/auth/update-mpin`

**Description:** Update existing MPIN with current MPIN verification.

```bash
curl -X POST "http://localhost:8080/api/v1/auth/update-mpin" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "current_mpin": "1234",
    "new_mpin": "5678"
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "MPIN updated successfully",
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## User Management Endpoints

### Delete User

**Endpoint:** `DELETE /api/v1/users/{user_id}`

**Description:** Soft delete a user account with proper cascade handling.

```bash
curl -X DELETE "http://localhost:8080/api/v1/users/USER123456789" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

**Response:**

```json
{
  "success": true,
  "message": "User deleted successfully",
  "data": {
    "user_id": "USER123456789",
    "deleted_at": "2024-01-01T12:00:00Z",
    "deleted_by": "ADMIN123456789"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Get User Details

**Endpoint:** `GET /api/v1/users/{user_id}`

**Description:** Retrieve detailed user information including roles and profile.

```bash
curl -X GET "http://localhost:8080/api/v1/users/USER123456789?include_roles=true&include_profile=true" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

**Response:**

```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "user": {
      "id": "USER123456789",
      "username": "john_doe",
      "phone_number": "+1234567890",
      "country_code": "US",
      "is_validated": true,
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z",
      "tokens": 100,
      "has_mpin": true,
      "roles": [
        {
          "id": "USERROLE123456789",
          "user_id": "USER123456789",
          "role_id": "ROLE123456789",
          "role": {
            "id": "ROLE123456789",
            "name": "user",
            "description": "Standard user role",
            "scope": "organization",
            "is_active": true,
            "version": 1
          },
          "is_active": true
        }
      ],
      "profile": {
        "id": "PROFILE123456789",
        "name": "John Doe",
        "date_of_birth": "1990-01-01",
        "aadhaar_number": "123456789012"
      }
    }
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## Error Handling Examples

### Validation Error

**Request:** Invalid login credentials format

```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "invalid",
    "country_code": "",
    "password": "123"
  }'
```

**Response:**

```json
{
  "success": false,
  "error": "VALIDATION_ERROR",
  "message": "Invalid input data",
  "code": 400,
  "details": {
    "errors": [
      "phone number must be 10 digits",
      "country code is required",
      "password must be at least 8 characters long"
    ]
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Authentication Error

**Request:** Invalid credentials

```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+1234567890",
    "country_code": "US",
    "password": "wrongPassword"
  }'
```

**Response:**

```json
{
  "success": false,
  "error": "AUTHENTICATION_ERROR",
  "message": "Invalid credentials",
  "code": 401,
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Authorization Error

**Request:** Insufficient permissions for role assignment

```bash
curl -X POST "http://localhost:8080/api/v1/users/USER123456789/roles" \
  -H "Authorization: Bearer invalid_or_insufficient_token" \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": "ROLE123456789"
  }'
```

**Response:**

```json
{
  "success": false,
  "error": "AUTHORIZATION_ERROR",
  "message": "Insufficient permissions to assign roles",
  "code": 403,
  "details": {
    "required_permission": "roles:assign",
    "user_permissions": ["users:read"]
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Not Found Error

**Request:** Role assignment to non-existent user

```bash
curl -X POST "http://localhost:8080/api/v1/users/NONEXISTENT/roles" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": "ROLE123456789"
  }'
```

**Response:**

```json
{
  "success": false,
  "error": "NOT_FOUND_ERROR",
  "message": "User not found",
  "code": 404,
  "details": {
    "resource": "user",
    "resource_id": "NONEXISTENT"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Conflict Error

**Request:** Assigning already assigned role

```bash
curl -X POST "http://localhost:8080/api/v1/users/USER123456789/roles" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": "ROLE123456789"
  }'
```

**Response:**

```json
{
  "success": false,
  "error": "CONFLICT_ERROR",
  "message": "Role already assigned to user",
  "code": 409,
  "details": {
    "user_id": "USER123456789",
    "role_id": "ROLE123456789",
    "existing_assignment_id": "USERROLE123456789"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

### Internal Server Error

**Request:** System error during processing

**Response:**

```json
{
  "success": false,
  "error": "INTERNAL_SERVER_ERROR",
  "message": "An internal error occurred while processing the request",
  "code": 500,
  "details": {
    "error_id": "ERR123456789",
    "support_message": "Please contact support with this error ID"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## Authentication Headers

All protected endpoints require authentication using Bearer tokens:

```bash
-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Rate Limiting

The API implements rate limiting. If you exceed the rate limit, you'll receive:

```json
{
  "success": false,
  "error": "RATE_LIMIT_EXCEEDED",
  "message": "Too many requests. Please try again later.",
  "code": 429,
  "details": {
    "limit": 100,
    "window": "1 hour",
    "retry_after": 3600
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123456789"
}
```

## Content Types

All requests should use `Content-Type: application/json` header and send JSON payloads.

## Response Format

All responses follow a consistent format:

- `success`: Boolean indicating if the request was successful
- `message`: Human-readable message describing the result
- `data`: Response data (present on successful requests)
- `error`: Error type (present on failed requests)
- `code`: HTTP status code (present on failed requests)
- `details`: Additional error details (present on failed requests)
- `timestamp`: ISO 8601 timestamp of the response
- `request_id`: Unique identifier for request tracing (optional)
