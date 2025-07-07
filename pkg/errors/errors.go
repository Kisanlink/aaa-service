package errors

import (
	"fmt"
	"runtime"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Validation errors
	ErrorTypeValidation    ErrorType = "VALIDATION_ERROR"
	ErrorTypeInvalidInput  ErrorType = "INVALID_INPUT"
	ErrorTypeMissingField  ErrorType = "MISSING_FIELD"
	ErrorTypeInvalidFormat ErrorType = "INVALID_FORMAT"

	// Authentication and Authorization errors
	ErrorTypeUnauthorized   ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden      ErrorType = "FORBIDDEN"
	ErrorTypeAuthentication ErrorType = "AUTHENTICATION_ERROR"
	ErrorTypeTokenExpired   ErrorType = "TOKEN_EXPIRED"
	ErrorTypeInvalidToken   ErrorType = "INVALID_TOKEN"

	// Database errors
	ErrorTypeNotFound            ErrorType = "NOT_FOUND"
	ErrorTypeAlreadyExists       ErrorType = "ALREADY_EXISTS"
	ErrorTypeDatabaseError       ErrorType = "DATABASE_ERROR"
	ErrorTypeConnectionError     ErrorType = "CONNECTION_ERROR"
	ErrorTypeConstraintViolation ErrorType = "CONSTRAINT_VIOLATION"

	// Business logic errors
	ErrorTypeBusinessRule       ErrorType = "BUSINESS_RULE_VIOLATION"
	ErrorTypeInsufficientTokens ErrorType = "INSUFFICIENT_TOKENS"
	ErrorTypeUserInactive       ErrorType = "USER_INACTIVE"
	ErrorTypeUserBlocked        ErrorType = "USER_BLOCKED"

	// External service errors
	ErrorTypeExternalService ErrorType = "EXTERNAL_SERVICE_ERROR"
	ErrorTypeTimeout         ErrorType = "TIMEOUT"
	ErrorTypeRateLimit       ErrorType = "RATE_LIMIT"

	// System errors
	ErrorTypeInternal       ErrorType = "INTERNAL_ERROR"
	ErrorTypeConfiguration  ErrorType = "CONFIGURATION_ERROR"
	ErrorTypeNotImplemented ErrorType = "NOT_IMPLEMENTED"
)

// CustomError represents a custom error with additional context
type CustomError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code,omitempty"`
	Field      string                 `json:"field,omitempty"`
	Value      interface{}            `json:"value,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StackTrace []string               `json:"stack_trace,omitempty"`
	Cause      error                  `json:"cause,omitempty"`
	HTTPStatus int                    `json:"http_status,omitempty"`
}

// Error implements the error interface
func (e *CustomError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *CustomError) Unwrap() error {
	return e.Cause
}

// Is checks if the error is of a specific type
func (e *CustomError) Is(target error) bool {
	if targetError, ok := target.(*CustomError); ok {
		return e.Type == targetError.Type
	}
	return false
}

// WithField adds a field to the error
func (e *CustomError) WithField(field string) *CustomError {
	e.Field = field
	return e
}

// WithValue adds a value to the error
func (e *CustomError) WithValue(value interface{}) *CustomError {
	e.Value = value
	return e
}

// WithDetails adds details to the error
func (e *CustomError) WithDetails(details map[string]interface{}) *CustomError {
	e.Details = details
	return e
}

// WithCause adds a cause to the error
func (e *CustomError) WithCause(cause error) *CustomError {
	e.Cause = cause
	return e
}

// WithHTTPStatus adds HTTP status to the error
func (e *CustomError) WithHTTPStatus(status int) *CustomError {
	e.HTTPStatus = status
	return e
}

// WithStackTrace adds stack trace to the error
func (e *CustomError) WithStackTrace() *CustomError {
	e.StackTrace = getStackTrace()
	return e
}

// New creates a new custom error
func New(errorType ErrorType, message string) *CustomError {
	return &CustomError{
		Type:    errorType,
		Message: message,
	}
}

// Newf creates a new custom error with formatted message
func Newf(errorType ErrorType, format string, args ...interface{}) *CustomError {
	return &CustomError{
		Type:    errorType,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errorType ErrorType, message string) *CustomError {
	return &CustomError{
		Type:    errorType,
		Message: message,
		Cause:   err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, errorType ErrorType, format string, args ...interface{}) *CustomError {
	return &CustomError{
		Type:    errorType,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
	}
}

// getStackTrace returns the current stack trace
func getStackTrace() []string {
	var stack []string
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	return stack
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errorType ErrorType) bool {
	var customErr *CustomError
	if err != nil && err.Error() != "" {
		if customErr, ok := err.(*CustomError); ok {
			return customErr.Type == errorType
		}
	}
	return false
}

// GetErrorType returns the error type if it's a custom error
func GetErrorType(err error) ErrorType {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Type
	}
	return ErrorTypeInternal
}

// GetHTTPStatus returns the HTTP status for an error
func GetHTTPStatus(err error) int {
	if customErr, ok := err.(*CustomError); ok {
		if customErr.HTTPStatus != 0 {
			return customErr.HTTPStatus
		}
		return getDefaultHTTPStatus(customErr.Type)
	}
	return 500
}

// getDefaultHTTPStatus returns the default HTTP status for an error type
func getDefaultHTTPStatus(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeValidation, ErrorTypeInvalidInput, ErrorTypeMissingField, ErrorTypeInvalidFormat:
		return 400
	case ErrorTypeUnauthorized, ErrorTypeAuthentication, ErrorTypeTokenExpired, ErrorTypeInvalidToken:
		return 401
	case ErrorTypeForbidden:
		return 403
	case ErrorTypeNotFound:
		return 404
	case ErrorTypeAlreadyExists, ErrorTypeConstraintViolation:
		return 409
	case ErrorTypeBusinessRule, ErrorTypeInsufficientTokens, ErrorTypeUserInactive, ErrorTypeUserBlocked:
		return 422
	case ErrorTypeRateLimit:
		return 429
	case ErrorTypeTimeout, ErrorTypeExternalService:
		return 503
	case ErrorTypeNotImplemented:
		return 501
	default:
		return 500
	}
}
