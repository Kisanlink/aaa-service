package errors

import (
	"fmt"
	"strings"
)

// Error types for different scenarios
type ValidationError struct {
	message string
	details []string
}

type NotFoundError struct {
	message string
}

type ConflictError struct {
	message string
}

type UnauthorizedError struct {
	message string
}

type ForbiddenError struct {
	message string
}

type InternalError struct {
	message string
	err     error
}

type BadRequestError struct {
	message string
	details []string
}

// Error method implementations
func (e *ValidationError) Error() string {
	if len(e.details) > 0 {
		return fmt.Sprintf("%s: %s", e.message, strings.Join(e.details, "; "))
	}
	return e.message
}

func (e *NotFoundError) Error() string {
	return e.message
}

func (e *ConflictError) Error() string {
	return e.message
}

func (e *UnauthorizedError) Error() string {
	return e.message
}

func (e *ForbiddenError) Error() string {
	return e.message
}

func (e *InternalError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.message, e.err)
	}
	return e.message
}

func (e *BadRequestError) Error() string {
	if len(e.details) > 0 {
		return fmt.Sprintf("%s: %s", e.message, strings.Join(e.details, "; "))
	}
	return e.message
}

// Details methods for error details
func (e *ValidationError) Details() []string {
	return e.details
}

func (e *BadRequestError) Details() []string {
	return e.details
}

// Constructor functions
func NewValidationError(message string, details ...string) *ValidationError {
	return &ValidationError{
		message: sanitizeErrorMessage(message),
		details: sanitizeErrorDetails(details),
	}
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		message: sanitizeErrorMessage(message),
	}
}

func NewConflictError(message string) *ConflictError {
	return &ConflictError{
		message: sanitizeErrorMessage(message),
	}
}

func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{
		message: sanitizeErrorMessage(message),
	}
}

func NewForbiddenError(message string) *ForbiddenError {
	return &ForbiddenError{
		message: sanitizeErrorMessage(message),
	}
}

func NewInternalError(err error) *InternalError {
	return &InternalError{
		message: "Internal server error",
		err:     err, // Internal errors are not exposed to clients
	}
}

func NewBadRequestError(message string, details ...string) *BadRequestError {
	return &BadRequestError{
		message: sanitizeErrorMessage(message),
		details: sanitizeErrorDetails(details),
	}
}

// Security error constructors that don't leak sensitive information
func NewSecureUnauthorizedError() *UnauthorizedError {
	return &UnauthorizedError{
		message: "Authentication required",
	}
}

func NewSecureForbiddenError() *ForbiddenError {
	return &ForbiddenError{
		message: "Access denied",
	}
}

func NewSecureNotFoundError(resourceType string) *NotFoundError {
	return &NotFoundError{
		message: fmt.Sprintf("%s not found", resourceType),
	}
}

func NewSecureValidationError() *ValidationError {
	return &ValidationError{
		message: "Invalid input provided",
		details: []string{},
	}
}

// Authentication specific errors that don't leak information
func NewAuthenticationFailedError() *UnauthorizedError {
	return &UnauthorizedError{
		message: "Invalid credentials",
	}
}

func NewAccountLockedError() *UnauthorizedError {
	return &UnauthorizedError{
		message: "Account temporarily locked due to multiple failed attempts",
	}
}

func NewTokenExpiredError() *UnauthorizedError {
	return &UnauthorizedError{
		message: "Token has expired",
	}
}

func NewInvalidTokenError() *UnauthorizedError {
	return &UnauthorizedError{
		message: "Invalid token",
	}
}

// Rate limiting errors
func NewRateLimitError(retryAfter string) *BadRequestError {
	return &BadRequestError{
		message: "Rate limit exceeded. Please try again later.",
		details: []string{fmt.Sprintf("retry_after: %s seconds", retryAfter)},
	}
}

// sanitizeErrorMessage removes sensitive information from error messages
func sanitizeErrorMessage(message string) string {
	// Remove potential sensitive patterns
	sensitivePatterns := []string{
		"password", "token", "secret", "key", "hash",
		"database", "sql", "connection", "server",
		"internal", "system", "debug", "trace",
	}

	lowerMessage := strings.ToLower(message)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerMessage, pattern) {
			// Return a generic message if sensitive information is detected
			return "An error occurred while processing your request"
		}
	}

	// Remove file paths and stack traces
	if strings.Contains(message, "/") || strings.Contains(message, "\\") {
		return "An error occurred while processing your request"
	}

	return message
}

// sanitizeErrorDetails removes sensitive information from error details
func sanitizeErrorDetails(details []string) []string {
	var sanitized []string
	for _, detail := range details {
		sanitizedDetail := sanitizeErrorMessage(detail)
		if sanitizedDetail != "An error occurred while processing your request" {
			sanitized = append(sanitized, sanitizedDetail)
		}
	}
	return sanitized
}

// IsErrorType helper functions
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

func IsConflictError(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}

func IsUnauthorizedError(err error) bool {
	_, ok := err.(*UnauthorizedError)
	return ok
}

func IsForbiddenError(err error) bool {
	_, ok := err.(*ForbiddenError)
	return ok
}

func IsInternalError(err error) bool {
	_, ok := err.(*InternalError)
	return ok
}

func IsBadRequestError(err error) bool {
	_, ok := err.(*BadRequestError)
	return ok
}

// GetErrorCode returns the appropriate HTTP status code for an error
func GetErrorCode(err error) int {
	switch err.(type) {
	case *ValidationError:
		return 400
	case *BadRequestError:
		return 400
	case *UnauthorizedError:
		return 401
	case *ForbiddenError:
		return 403
	case *NotFoundError:
		return 404
	case *ConflictError:
		return 409
	case *InternalError:
		return 500
	default:
		return 500
	}
}

// GetErrorType returns the error type as a string
func GetErrorType(err error) string {
	switch err.(type) {
	case *ValidationError:
		return "VALIDATION_ERROR"
	case *BadRequestError:
		return "BAD_REQUEST"
	case *UnauthorizedError:
		return "UNAUTHORIZED"
	case *ForbiddenError:
		return "FORBIDDEN"
	case *NotFoundError:
		return "NOT_FOUND"
	case *ConflictError:
		return "CONFLICT"
	case *InternalError:
		return "INTERNAL_ERROR"
	default:
		return "GENERIC_ERROR"
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *ValidationError:
		return NewValidationError(fmt.Sprintf("%s: %s", message, e.message), e.details...)
	case *NotFoundError:
		return NewNotFoundError(fmt.Sprintf("%s: %s", message, e.message))
	case *ConflictError:
		return NewConflictError(fmt.Sprintf("%s: %s", message, e.message))
	case *UnauthorizedError:
		return NewUnauthorizedError(fmt.Sprintf("%s: %s", message, e.message))
	case *ForbiddenError:
		return NewForbiddenError(fmt.Sprintf("%s: %s", message, e.message))
	case *InternalError:
		return NewInternalError(fmt.Errorf("%s: %w", message, e.err))
	case *BadRequestError:
		return NewBadRequestError(fmt.Sprintf("%s: %s", message, e.message), e.details...)
	default:
		return NewInternalError(fmt.Errorf("%s: %w", message, err))
	}
}

// CombineErrors combines multiple errors into a single error
func CombineErrors(errors ...error) error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return errors[0]
	}

	var messages []string
	var details []string

	for _, err := range errors {
		if err != nil {
			messages = append(messages, err.Error())

			// Extract details from validation errors
			if validationErr, ok := err.(*ValidationError); ok {
				details = append(details, validationErr.details...)
			}
		}
	}

	if len(details) > 0 {
		return NewValidationError(strings.Join(messages, "; "), details...)
	}

	return NewBadRequestError(strings.Join(messages, "; "))
}
