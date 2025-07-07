package errors

import (
	"fmt"
)

// Validation errors
func NewValidationError(message string) *CustomError {
	return New(ErrorTypeValidation, message)
}

func NewValidationErrorf(format string, args ...interface{}) *CustomError {
	return Newf(ErrorTypeValidation, format, args...)
}

func NewInvalidInputError(field string, value interface{}, message string) *CustomError {
	return New(ErrorTypeInvalidInput, message).
		WithField(field).
		WithValue(value)
}

func NewMissingFieldError(field string) *CustomError {
	return New(ErrorTypeMissingField, fmt.Sprintf("field '%s' is required", field)).
		WithField(field)
}

func NewInvalidFormatError(field string, value interface{}, expectedFormat string) *CustomError {
	return New(ErrorTypeInvalidFormat, fmt.Sprintf("field '%s' has invalid format, expected: %s", field, expectedFormat)).
		WithField(field).
		WithValue(value)
}

// Authentication and Authorization errors
func NewUnauthorizedError(message string) *CustomError {
	return New(ErrorTypeUnauthorized, message).WithHTTPStatus(401)
}

func NewForbiddenError(message string) *CustomError {
	return New(ErrorTypeForbidden, message).WithHTTPStatus(403)
}

func NewAuthenticationError(message string) *CustomError {
	return New(ErrorTypeAuthentication, message).WithHTTPStatus(401)
}

func NewTokenExpiredError() *CustomError {
	return New(ErrorTypeTokenExpired, "token has expired").WithHTTPStatus(401)
}

func NewInvalidTokenError() *CustomError {
	return New(ErrorTypeInvalidToken, "invalid token").WithHTTPStatus(401)
}

// Database errors
func NewNotFoundError(resource string, id string) *CustomError {
	return New(ErrorTypeNotFound, fmt.Sprintf("%s with id '%s' not found", resource, id)).
		WithHTTPStatus(404).
		WithDetails(map[string]interface{}{
			"resource": resource,
			"id":       id,
		})
}

func NewAlreadyExistsError(resource string, field string, value interface{}) *CustomError {
	return New(ErrorTypeAlreadyExists, fmt.Sprintf("%s with %s '%v' already exists", resource, field, value)).
		WithHTTPStatus(409).
		WithField(field).
		WithValue(value).
		WithDetails(map[string]interface{}{
			"resource": resource,
		})
}

func NewDatabaseError(message string, cause error) *CustomError {
	return Wrap(cause, ErrorTypeDatabaseError, message).WithHTTPStatus(500)
}

func NewConnectionError(message string, cause error) *CustomError {
	return Wrap(cause, ErrorTypeConnectionError, message).WithHTTPStatus(503)
}

func NewConstraintViolationError(constraint string, message string) *CustomError {
	return New(ErrorTypeConstraintViolation, message).
		WithHTTPStatus(409).
		WithDetails(map[string]interface{}{
			"constraint": constraint,
		})
}

// Business logic errors
func NewBusinessRuleError(rule string, message string) *CustomError {
	return New(ErrorTypeBusinessRule, message).
		WithHTTPStatus(422).
		WithDetails(map[string]interface{}{
			"rule": rule,
		})
}

func NewInsufficientTokensError(required int, available int) *CustomError {
	return New(ErrorTypeInsufficientTokens, fmt.Sprintf("insufficient tokens: required %d, available %d", required, available)).
		WithHTTPStatus(422).
		WithDetails(map[string]interface{}{
			"required":  required,
			"available": available,
		})
}

func NewUserInactiveError(userID string) *CustomError {
	return New(ErrorTypeUserInactive, "user account is inactive").
		WithHTTPStatus(422).
		WithDetails(map[string]interface{}{
			"user_id": userID,
		})
}

func NewUserBlockedError(userID string) *CustomError {
	return New(ErrorTypeUserBlocked, "user account is blocked").
		WithHTTPStatus(422).
		WithDetails(map[string]interface{}{
			"user_id": userID,
		})
}

// External service errors
func NewExternalServiceError(service string, message string, cause error) *CustomError {
	return Wrap(cause, ErrorTypeExternalService, fmt.Sprintf("%s service error: %s", service, message)).
		WithHTTPStatus(503).
		WithDetails(map[string]interface{}{
			"service": service,
		})
}

func NewTimeoutError(operation string, timeout string) *CustomError {
	return New(ErrorTypeTimeout, fmt.Sprintf("operation '%s' timed out after %s", operation, timeout)).
		WithHTTPStatus(503).
		WithDetails(map[string]interface{}{
			"operation": operation,
			"timeout":   timeout,
		})
}

func NewRateLimitError(limit int, window string) *CustomError {
	return New(ErrorTypeRateLimit, fmt.Sprintf("rate limit exceeded: %d requests per %s", limit, window)).
		WithHTTPStatus(429).
		WithDetails(map[string]interface{}{
			"limit":  limit,
			"window": window,
		})
}

// System errors
func NewInternalError(message string, cause error) *CustomError {
	return Wrap(cause, ErrorTypeInternal, message).WithHTTPStatus(500)
}

func NewConfigurationError(config string, message string) *CustomError {
	return New(ErrorTypeConfiguration, fmt.Sprintf("configuration error for '%s': %s", config, message)).
		WithHTTPStatus(500).
		WithDetails(map[string]interface{}{
			"config": config,
		})
}

func NewNotImplementedError(feature string) *CustomError {
	return New(ErrorTypeNotImplemented, fmt.Sprintf("feature '%s' is not implemented", feature)).
		WithHTTPStatus(501).
		WithDetails(map[string]interface{}{
			"feature": feature,
		})
}

// User-specific errors
func NewUserNotFoundError(userID string) *CustomError {
	return NewNotFoundError("user", userID)
}

func NewUserAlreadyExistsError(field string, value interface{}) *CustomError {
	return NewAlreadyExistsError("user", field, value)
}

func NewUserProfileNotFoundError(userID string) *CustomError {
	return NewNotFoundError("user profile", userID)
}

func NewContactNotFoundError(contactID string) *CustomError {
	return NewNotFoundError("contact", contactID)
}

func NewContactAlreadyExistsError(mobileNumber uint64) *CustomError {
	return NewAlreadyExistsError("contact", "mobile_number", mobileNumber)
}

func NewAddressNotFoundError(addressID string) *CustomError {
	return NewNotFoundError("address", addressID)
}

func NewRoleNotFoundError(roleID string) *CustomError {
	return NewNotFoundError("role", roleID)
}

func NewRoleAlreadyExistsError(name string) *CustomError {
	return NewAlreadyExistsError("role", "name", name)
}

func NewUserRoleNotFoundError(userID string, roleID string) *CustomError {
	return NewNotFoundError("user role", fmt.Sprintf("%s-%s", userID, roleID))
}

func NewUserRoleAlreadyExistsError(userID string, roleID string) *CustomError {
	return NewAlreadyExistsError("user role", "user_id-role_id", fmt.Sprintf("%s-%s", userID, roleID))
}
