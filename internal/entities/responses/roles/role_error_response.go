package roles

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/responses"
)

// RoleErrorType represents different types of role-related errors
type RoleErrorType string

const (
	RoleErrorTypeNotFound            RoleErrorType = "ROLE_NOT_FOUND"
	RoleErrorTypeUserNotFound        RoleErrorType = "USER_NOT_FOUND"
	RoleErrorTypeDuplicateAssignment RoleErrorType = "DUPLICATE_ROLE_ASSIGNMENT"
	RoleErrorTypeAssignmentNotFound  RoleErrorType = "ROLE_ASSIGNMENT_NOT_FOUND"
	RoleErrorTypeInvalidRole         RoleErrorType = "INVALID_ROLE"
	RoleErrorTypeInactiveRole        RoleErrorType = "INACTIVE_ROLE"
	RoleErrorTypePermissionDenied    RoleErrorType = "PERMISSION_DENIED"
	RoleErrorTypeValidationFailed    RoleErrorType = "VALIDATION_FAILED"
)

// RoleErrorResponse represents role-specific error responses
type RoleErrorResponse struct {
	responses.ErrorResponse
	ErrorType RoleErrorType          `json:"error_type"`
	UserID    string                 `json:"user_id,omitempty"`
	RoleID    string                 `json:"role_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// NewRoleErrorResponse creates a new role-specific error response
func NewRoleErrorResponse(errorType RoleErrorType, message string, code int) *RoleErrorResponse {
	return &RoleErrorResponse{
		ErrorResponse: responses.ErrorResponse{
			Error:     string(errorType),
			Message:   message,
			Code:      strconv.Itoa(code),
			Timestamp: time.Now(),
		},
		ErrorType: errorType,
	}
}

// NewRoleNotFoundError creates a role not found error
func NewRoleNotFoundError(roleID string) *RoleErrorResponse {
	err := NewRoleErrorResponse(
		RoleErrorTypeNotFound,
		fmt.Sprintf("Role with ID '%s' not found", roleID),
		404,
	)
	err.RoleID = roleID
	return err
}

// NewUserNotFoundError creates a user not found error for role operations
func NewUserNotFoundError(userID string) *RoleErrorResponse {
	err := NewRoleErrorResponse(
		RoleErrorTypeUserNotFound,
		fmt.Sprintf("User with ID '%s' not found", userID),
		404,
	)
	err.UserID = userID
	return err
}

// NewDuplicateRoleAssignmentError creates a duplicate role assignment error
func NewDuplicateRoleAssignmentError(userID, roleID string) *RoleErrorResponse {
	err := NewRoleErrorResponse(
		RoleErrorTypeDuplicateAssignment,
		fmt.Sprintf("Role '%s' is already assigned to user '%s'", roleID, userID),
		409,
	)
	err.UserID = userID
	err.RoleID = roleID
	return err
}

// NewRoleAssignmentNotFoundError creates a role assignment not found error
func NewRoleAssignmentNotFoundError(userID, roleID string) *RoleErrorResponse {
	err := NewRoleErrorResponse(
		RoleErrorTypeAssignmentNotFound,
		fmt.Sprintf("Role assignment not found for user '%s' and role '%s'", userID, roleID),
		404,
	)
	err.UserID = userID
	err.RoleID = roleID
	return err
}

// NewInactiveRoleError creates an inactive role error
func NewInactiveRoleError(roleID string) *RoleErrorResponse {
	err := NewRoleErrorResponse(
		RoleErrorTypeInactiveRole,
		fmt.Sprintf("Role '%s' is inactive and cannot be assigned", roleID),
		400,
	)
	err.RoleID = roleID
	return err
}

// NewRoleValidationError creates a role validation error
func NewRoleValidationError(field, message string) *RoleErrorResponse {
	err := NewRoleErrorResponse(
		RoleErrorTypeValidationFailed,
		fmt.Sprintf("Validation failed for field '%s': %s", field, message),
		400,
	)
	if err.Details == nil {
		err.Details = make(map[string]interface{})
	}
	err.Details["field"] = field
	err.Details["validation_message"] = message
	return err
}

// NewRolePermissionDeniedError creates a permission denied error for role operations
func NewRolePermissionDeniedError(operation string) *RoleErrorResponse {
	return NewRoleErrorResponse(
		RoleErrorTypePermissionDenied,
		fmt.Sprintf("Permission denied for role operation: %s", operation),
		403,
	)
}

// WithUserID adds user ID to the error response
func (r *RoleErrorResponse) WithUserID(userID string) *RoleErrorResponse {
	r.UserID = userID
	return r
}

// WithRoleID adds role ID to the error response
func (r *RoleErrorResponse) WithRoleID(roleID string) *RoleErrorResponse {
	r.RoleID = roleID
	return r
}

// WithDetails adds additional details to the error response
func (r *RoleErrorResponse) WithDetails(details map[string]interface{}) *RoleErrorResponse {
	if r.Details == nil {
		r.Details = make(map[string]interface{})
	}
	for k, v := range details {
		r.Details[k] = v
	}
	return r
}

// WithRequestID adds request ID to the error response
func (r *RoleErrorResponse) WithRequestID(requestID string) *RoleErrorResponse {
	r.ErrorResponse.RequestID = requestID
	return r
}

// GetType returns the response type
func (r *RoleErrorResponse) GetType() string {
	return "RoleErrorResponse"
}

// IsSuccess returns whether the response indicates success
func (r *RoleErrorResponse) IsSuccess() bool {
	return false
}

// GetProtocol returns the transport protocol
func (r *RoleErrorResponse) GetProtocol() string {
	return "http"
}

// GetOperation returns the operation
func (r *RoleErrorResponse) GetOperation() string {
	return "error"
}

// GetVersion returns the API version
func (r *RoleErrorResponse) GetVersion() string {
	return "v2"
}

// GetResponseID returns the response ID
func (r *RoleErrorResponse) GetResponseID() string {
	return ""
}

// GetHeaders returns response headers
func (r *RoleErrorResponse) GetHeaders() map[string][]string {
	return nil
}

// GetBody returns the response body
func (r *RoleErrorResponse) GetBody() interface{} {
	return r
}

// GetContext returns response context
func (r *RoleErrorResponse) GetContext() map[string]interface{} {
	return nil
}

// ToProto converts to protocol buffer format
func (r *RoleErrorResponse) ToProto() interface{} {
	return nil
}

// String returns a string representation
func (r *RoleErrorResponse) String() string {
	return fmt.Sprintf("RoleErrorResponse{Type: %s, Message: %s, UserID: %s, RoleID: %s}",
		r.ErrorType, r.ErrorResponse.Message, r.UserID, r.RoleID)
}
