# API Response Optimization Design Document

## Overview

This design document outlines the optimization of AAA Service API responses to eliminate redundancy, standardize field naming, and create cleaner, more intuitive response structures. The solution focuses on consistent data representation, payload optimization, and flexible response control through query parameters.

## Architecture

### Response Transformation Layer

The optimization will be implemented through a response transformation layer that sits between the service layer and HTTP handlers. This approach ensures consistent response formatting across all endpoints without requiring changes to core business logic.

```
HTTP Handler → Response Transformer → JSON Serialization → Client
```

**Design Decision Rationale:** Using a transformation layer allows us to maintain backward compatibility while gradually migrating to optimized responses. It also centralizes response formatting logic, making it easier to maintain consistency.

### Field Naming Standardization

All API responses will use consistent snake_case naming conventions, regardless of the underlying database column names or Go struct field names.

**Implementation Strategy:**

- Custom JSON struct tags for all response models
- Consistent naming patterns across all endpoints
- Automated validation to prevent naming inconsistencies

**Design Decision Rationale:** Snake_case is the JSON standard and provides better readability for API consumers. Decoupling response field names from internal struct names allows for better API design flexibility.

## Components and Interfaces

### Response Transformer Interface

```go
type ResponseTransformer interface {
    TransformUser(user *models.User, options TransformOptions) *responses.UserResponse
    TransformRole(role *models.Role, options TransformOptions) *responses.RoleResponse
    TransformUserRole(userRole *models.UserRole, options TransformOptions) *responses.UserRoleResponse
    // add the transformers for other responses also
}

type TransformOptions struct {
    IncludeUser     bool
    IncludeRole     bool
    IncludeProfile  bool
    ExcludeDeleted  bool
    // same
}
```

### Standardized Response Models

#### User Response Structure

```go
type UserResponse struct {
    ID          uint      `json:"id"`
    Username    string    `json:"username"`
    Email       string    `json:"email"`
    Phone       string    `json:"phone,omitempty"`
    IsActive    bool      `json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
    Profile     *UserProfileResponse `json:"profile,omitempty"`
}
```

#### Role Assignment Response Structure

```go
type UserRoleResponse struct {
    ID           uint      `json:"id"`
    UserID       uint      `json:"user_id"`
    RoleID       uint      `json:"role_id"`
    AssignedAt   time.Time `json:"assigned_at"`
    AssignedBy   uint      `json:"assigned_by"`
    IsActive     bool      `json:"is_active"`
    User         *UserResponse `json:"user,omitempty"`
    Role         *RoleResponse `json:"role,omitempty"`
}
```

**Design Decision Rationale:** These structures eliminate redundant user_id fields in nested objects and provide clear, single sources of truth for entity identifiers. Optional nested objects allow for flexible response control.

### Query Parameter Handler

```go
type QueryParameterHandler struct {
    allowedIncludes map[string]bool
    defaultOptions  TransformOptions
}

func (h *QueryParameterHandler) ParseTransformOptions(c *gin.Context) TransformOptions {
    options := h.defaultOptions

    if c.Query("include_user") == "true" {
        options.IncludeUser = true
    }
    if c.Query("include_role") == "true" {
        options.IncludeRole = true
    }

    return options
}
```

**Design Decision Rationale:** Query parameter control allows API consumers to optimize bandwidth usage by requesting only the data they need. This is particularly important for mobile applications and high-traffic scenarios.

## Data Models

### Sensitive Field Exclusion

All response models will explicitly exclude sensitive fields through struct tags and transformation logic:

```go
type User struct {
    // Database fields
    ID       uint   `gorm:"primaryKey"`
    Username string
    Email    string
    Password string `json:"-"` // Excluded from JSON
    MPIN     string `json:"-"` // Excluded from JSON

    // Response transformation will further filter fields
}
```

### Consistent State Representation

The design ensures logical consistency between `is_active` and `deleted_at` fields:

- `is_active: true, deleted_at: null` → Active entity
- `is_active: false, deleted_at: timestamp` → Soft-deleted entity
- Invalid combinations will be normalized during transformation

**Design Decision Rationale:** Consistent state representation eliminates ambiguity and makes the API more predictable for consumers.

### Timestamp Standardization

All timestamps will be represented in UTC format with consistent JSON serialization:

```go
type BaseResponse struct {
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
```

## Error Handling

### Response Validation

The transformation layer will include validation to ensure response consistency:

```go
type ResponseValidator interface {
    ValidateUserResponse(response *UserResponse) error
    ValidateConsistency(responses []interface{}) error
}
```

### Error Response Standardization

Error responses will also follow consistent formatting:

```go
type ErrorResponse struct {
    Error   string            `json:"error"`
    Message string            `json:"message"`
    Code    int              `json:"code"`
    Details map[string]string `json:"details,omitempty"`
}
```

### Authentication Error Handling

The response optimization will include proper authentication validation to ensure invalidated tokens are rejected:

```go
type AuthenticationError struct {
    Code      string `json:"code"`
    Message   string `json:"message"`
    RequestID string `json:"request_id"`
    Success   bool   `json:"success"`
    Timestamp string `json:"timestamp"`
}

// Example 401 response for invalidated token
{
    "code": "AUTHENTICATION_ERROR",
    "message": "Token has been invalidated",
    "request_id": "uuid-here",
    "success": false,
    "timestamp": "2025-01-01T00:00:00Z"
}
```

**Design Decision Rationale:** Consistent error responses improve API usability and make error handling more predictable for consumers. Proper authentication validation ensures security by preventing access with invalidated tokens.

## Testing Strategy

### Response Structure Validation

Automated tests will validate response structures across all endpoints:

1. **Field Naming Tests**: Verify consistent snake_case naming
2. **Structure Consistency Tests**: Ensure identical structures for common entities
3. **Sensitive Data Tests**: Verify exclusion of password and MPIN fields
4. **State Consistency Tests**: Validate is_active/deleted_at relationships

### Integration Testing

```go
func TestResponseConsistency(t *testing.T) {
    // Test that user objects have identical structure across endpoints
    userFromUsersEndpoint := getUserFromUsersAPI()
    userFromRolesEndpoint := getUserFromUserRolesAPI()

    assert.Equal(t, userFromUsersEndpoint.Structure(), userFromRolesEndpoint.Structure())
}
```

### Query Parameter Testing

Tests will verify that query parameters correctly control response content:

```go
func TestQueryParameterControl(t *testing.T) {
    // Test include_user parameter
    responseWithUser := makeRequest("/users/1/roles?include_user=true")
    responseWithoutUser := makeRequest("/users/1/roles")

    assert.NotNil(t, responseWithUser.User)
    assert.Nil(t, responseWithoutUser.User)
}
```

### Authentication Validation Testing

Tests will verify that invalidated tokens are properly rejected:

```go
func TestAuthenticationValidation(t *testing.T) {
    // Test that invalidated tokens return 401
    token := loginAndGetToken()
    logoutWithToken(token)

    response := makeAuthenticatedRequest("/api/v1/users", token)
    assert.Equal(t, 401, response.StatusCode)
    assert.Contains(t, response.Body, "AUTHENTICATION_ERROR")

    // Test that valid tokens still work
    newToken := loginAndGetToken()
    validResponse := makeAuthenticatedRequest("/api/v1/users", newToken)
    assert.Equal(t, 200, validResponse.StatusCode)
}
```

### Performance Testing

Response payload size optimization will be validated through performance tests:

1. **Payload Size Tests**: Measure response size reduction
2. **Serialization Performance**: Ensure transformation doesn't impact performance
3. **Memory Usage Tests**: Verify efficient memory usage during transformation

### Password Management Testing

Tests will verify password change and reset functionality:

```go
func TestPasswordManagement(t *testing.T) {
    // Test password change with valid current password
    token := loginAndGetToken()
    changeResponse := changePassword(token, "oldPassword", "newPassword123!")
    assert.Equal(t, 200, changeResponse.StatusCode)
    assert.True(t, changeResponse.TokensInvalidated)

    // Test password reset flow
    resetRequest := requestPasswordReset("user@example.com")
    assert.Equal(t, 200, resetRequest.StatusCode)
    assert.NotEmpty(t, resetRequest.TransactionID)

    // Test password reset with OTP
    resetResponse := verifyPasswordReset(resetRequest.TransactionID, "123456", "newPassword456!")
    assert.Equal(t, 200, resetResponse.StatusCode)
    assert.True(t, resetResponse.TokensInvalidated)
}

func TestPasswordSecurity(t *testing.T) {
    // Test that old tokens are invalidated after password change
    token := loginAndGetToken()
    changePassword(token, "oldPassword", "newPassword123!")

    // Old token should no longer work
    response := makeAuthenticatedRequest("/api/v1/users", token)
    assert.Equal(t, 401, response.StatusCode)
}
```

**Design Decision Rationale:** Comprehensive testing ensures that optimizations don't break existing functionality and that performance improvements are measurable. Password management testing ensures security requirements are met and token invalidation works correctly.

## Implementation Phases

### Phase 1: Response Transformation Infrastructure

- Implement ResponseTransformer interface
- Create standardized response models
- Add query parameter parsing

### Phase 2: Core Endpoint Migration

- Migrate user management endpoints
- Migrate role management endpoints
- Implement sensitive field exclusion

### Phase 3: Advanced Features

- Add query parameter controls
- Implement response validation
- Add performance monitoring

### Phase 4: Validation and Optimization

- Comprehensive testing
- Performance optimization
- Documentation updates

**Design Decision Rationale:** Phased implementation allows for gradual migration with minimal risk to existing functionality. Each phase builds upon the previous one, ensuring stable progress.

## Aadhaar Verification Integration

### User Validation Endpoint Integration

The API response optimization will include integration with the aadhaar-verification service for enhanced user validation capabilities.

### Aadhaar Verification Service Interface

```go
type AadhaarVerificationService interface {
    GenerateOTP(aadhaarNumber string) (*AadhaarOTPResponse, error)
    VerifyOTP(aadhaarNumber string, otp string) (*AadhaarVerificationResponse, error)
    GetVerificationStatus(userID string) (*VerificationStatusResponse, error)
}

type AadhaarOTPResponse struct {
    TransactionID string    `json:"transaction_id"`
    Message       string    `json:"message"`
    Success       bool      `json:"success"`
    RequestID     string    `json:"request_id"`
    Timestamp     time.Time `json:"timestamp"`
}

type AadhaarVerificationResponse struct {
    UserID        string                 `json:"user_id"`
    AadhaarNumber string                 `json:"aadhaar_number,omitempty"` // Masked for security
    Verified      bool                   `json:"verified"`
    VerifiedAt    *time.Time            `json:"verified_at,omitempty"`
    UserData      *AadhaarUserData      `json:"user_data,omitempty"`
    RequestID     string                 `json:"request_id"`
    Success       bool                   `json:"success"`
    Timestamp     time.Time             `json:"timestamp"`
}

type AadhaarUserData struct {
    Name          string `json:"name"`
    DateOfBirth   string `json:"date_of_birth,omitempty"`
    Gender        string `json:"gender,omitempty"`
    Address       string `json:"address,omitempty"`
    // Other fields as per Aadhaar verification service response
}
```

### Enhanced User Response with Verification Status

```go
type UserResponse struct {
    ID                uint                         `json:"id"`
    Username          string                       `json:"username"`
    Email             string                       `json:"email"`
    Phone             string                       `json:"phone,omitempty"`
    IsActive          bool                         `json:"is_active"`
    CreatedAt         time.Time                    `json:"created_at"`
    UpdatedAt         time.Time                    `json:"updated_at"`
    DeletedAt         *time.Time                   `json:"deleted_at,omitempty"`
    Profile           *UserProfileResponse         `json:"profile,omitempty"`
    VerificationStatus *VerificationStatusResponse `json:"verification_status,omitempty"`
}

type VerificationStatusResponse struct {
    AadhaarVerified   bool       `json:"aadhaar_verified"`
    AadhaarVerifiedAt *time.Time `json:"aadhaar_verified_at,omitempty"`
    KYCStatus         string     `json:"kyc_status"` // PENDING, VERIFIED, REJECTED
    LastVerificationAttempt *time.Time `json:"last_verification_attempt,omitempty"`
}
```

### User Validation Endpoints

New endpoints will be added to support Aadhaar verification:

1. **POST /api/v1/users/{id}/aadhaar/otp** - Generate OTP for Aadhaar verification
2. **POST /api/v1/users/{id}/aadhaar/verify** - Verify OTP and complete Aadhaar verification
3. **GET /api/v1/users/{id}/verification-status** - Get user's verification status

### Password Management Endpoints

New endpoints will be added to support secure password management:

1. **POST /api/v1/users/{id}/password/change** - Change password with current password validation
2. **POST /api/v1/auth/password/reset/request** - Request password reset OTP
3. **POST /api/v1/auth/password/reset/verify** - Reset password using OTP

### Integration with Existing Contact Verification

The Aadhaar verification will complement the existing contact verification system:

```go
type ContactResponse struct {
    ID           uint      `json:"id"`
    UserID       uint      `json:"user_id"`
    Type         string    `json:"type"`
    Value        string    `json:"value"`
    CountryCode  string    `json:"country_code,omitempty"`
    IsPrimary    bool      `json:"is_primary"`
    IsVerified   bool      `json:"is_verified"`
    IsActive     bool      `json:"is_active"`
    VerifiedAt   *time.Time `json:"verified_at,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

### Query Parameter Support for Verification Data

The query parameter system will support verification data inclusion:

- `include_verification=true` - Include verification status in user responses
- `include_aadhaar_data=true` - Include Aadhaar verification details (with proper security controls)
- `include_contacts=true` - Include contact verification status

## Password Management Integration

### Password Management Service Interface

```go
type PasswordManagementService interface {
    ChangePassword(userID string, currentPassword, newPassword string) (*PasswordChangeResponse, error)
    RequestPasswordReset(identifier string) (*PasswordResetOTPResponse, error) // identifier can be email/phone
    VerifyPasswordReset(otp, newPassword string, transactionID string) (*PasswordResetResponse, error)
    ValidatePasswordStrength(password string) (*PasswordValidationResponse, error)
}

type PasswordChangeResponse struct {
    UserID    string    `json:"user_id"`
    Message   string    `json:"message"`
    Success   bool      `json:"success"`
    RequestID string    `json:"request_id"`
    Timestamp time.Time `json:"timestamp"`
    TokensInvalidated bool `json:"tokens_invalidated"`
}

type PasswordResetOTPResponse struct {
    TransactionID string    `json:"transaction_id"`
    Message       string    `json:"message"`
    Success       bool      `json:"success"`
    RequestID     string    `json:"request_id"`
    Timestamp     time.Time `json:"timestamp"`
    ExpiresIn     int       `json:"expires_in"` // seconds
    SentTo        string    `json:"sent_to"`    // masked contact info
}

type PasswordResetResponse struct {
    UserID    string    `json:"user_id"`
    Message   string    `json:"message"`
    Success   bool      `json:"success"`
    RequestID string    `json:"request_id"`
    Timestamp time.Time `json:"timestamp"`
    TokensInvalidated bool `json:"tokens_invalidated"`
}

type PasswordValidationResponse struct {
    Valid         bool     `json:"valid"`
    Score         int      `json:"score"`         // 0-100 strength score
    Requirements  []string `json:"requirements"`  // List of requirements met
    Suggestions   []string `json:"suggestions"`   // Improvement suggestions
}
```

### Password Change Request/Response Models

```go
type ChangePasswordRequest struct {
    CurrentPassword string `json:"current_password" validate:"required"`
    NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type PasswordResetRequest struct {
    Identifier string `json:"identifier" validate:"required"` // email or phone
}

type PasswordResetVerifyRequest struct {
    TransactionID string `json:"transaction_id" validate:"required"`
    OTP          string `json:"otp" validate:"required"`
    NewPassword  string `json:"new_password" validate:"required,min=8"`
}
```

### Security Considerations for Password Management

1. **Password Validation**: Enforce strong password policies
2. **Rate Limiting**: Limit password reset attempts per user/IP
3. **Token Invalidation**: Invalidate all user tokens after password changes
4. **Audit Logging**: Log all password-related activities
5. **OTP Security**: Short expiration times and single-use OTPs

### Integration with Existing Authentication Flow

The password management endpoints will integrate seamlessly with the existing authentication system:

```go
// Enhanced auth response to include password-related metadata
type AuthResponse struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    TokenType    string    `json:"token_type"`
    ExpiresIn    int       `json:"expires_in"`
    Message      string    `json:"message"`
    UserID       string    `json:"user_id"`
    PasswordLastChanged *time.Time `json:"password_last_changed,omitempty"`
    RequiresPasswordChange bool   `json:"requires_password_change,omitempty"`
}
```

**Design Decision Rationale:** Integrating comprehensive password management ensures complete user account security while maintaining the consistent response structure. The separation of change password (authenticated) and reset password (unauthenticated with OTP) flows provides appropriate security levels for different scenarios. Token invalidation after password changes ensures that compromised accounts are properly secured.
