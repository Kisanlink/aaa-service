package responses

import (
	"time"

	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Success   bool                   `json:"success"`
	Error     string                 `json:"error"`
	Message   string                 `json:"message"`
	Code      string                 `json:"code,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Code      string    `json:"code"`
	Errors    []string  `json:"errors"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

// SecurityErrorResponse represents a security-related error response
type SecurityErrorResponse struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Code      string    `json:"code"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
	// Note: No details for security errors to prevent information leakage
}

// NewErrorResponse creates a new error response
func NewErrorResponse(error, message string, code string) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Error:     error,
		Message:   message,
		Code:      code,
		Timestamp: time.Now().UTC(),
	}
}

// NewErrorResponseFromError creates an error response from a custom error
func NewErrorResponseFromError(err error, requestID string) *ErrorResponse {
	if err == nil {
		return NewErrorResponse("UNKNOWN_ERROR", "An unknown error occurred", "UNKNOWN_ERROR").
			WithRequestID(requestID)
	}

	switch e := err.(type) {
	case *errors.ValidationError:
		response := &ValidationErrorResponse{
			Success:   false,
			Error:     "VALIDATION_ERROR",
			Message:   e.Error(),
			Code:      "VALIDATION_ERROR",
			Errors:    e.Details(),
			Timestamp: time.Now().UTC(),
			RequestID: requestID,
		}
		// Convert to ErrorResponse for consistency
		return &ErrorResponse{
			Success:   false,
			Error:     response.Error,
			Message:   response.Message,
			Code:      response.Code,
			Details:   map[string]interface{}{"errors": response.Errors},
			Timestamp: response.Timestamp,
			RequestID: response.RequestID,
		}

	case *errors.BadRequestError:
		return NewErrorResponse("BAD_REQUEST", e.Error(), "BAD_REQUEST").
			WithRequestID(requestID).
			WithDetails(map[string]interface{}{"errors": e.Details()})

	case *errors.UnauthorizedError:
		return NewSecurityErrorResponse("UNAUTHORIZED", e.Error(), "UNAUTHORIZED", requestID)

	case *errors.ForbiddenError:
		return NewSecurityErrorResponse("FORBIDDEN", e.Error(), "FORBIDDEN", requestID)

	case *errors.NotFoundError:
		return NewErrorResponse("NOT_FOUND", e.Error(), "NOT_FOUND").
			WithRequestID(requestID)

	case *errors.ConflictError:
		return NewErrorResponse("CONFLICT", e.Error(), "CONFLICT").
			WithRequestID(requestID)

	case *errors.InternalError:
		// Never expose internal error details
		return NewErrorResponse("INTERNAL_ERROR", "An internal server error occurred", "INTERNAL_ERROR").
			WithRequestID(requestID)

	default:
		return NewErrorResponse("GENERIC_ERROR", err.Error(), "GENERIC_ERROR").
			WithRequestID(requestID)
	}
}

// NewSecurityErrorResponse creates a security error response (no details)
func NewSecurityErrorResponse(error, message, code, requestID string) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Error:     error,
		Message:   message,
		Code:      code,
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
		// Intentionally no Details field for security errors
	}
}

// NewValidationErrorResponse creates a validation error response
func NewValidationErrorResponse(message string, validationErrors []string, requestID string) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Success:   false,
		Error:     "VALIDATION_ERROR",
		Message:   message,
		Code:      "VALIDATION_ERROR",
		Errors:    validationErrors,
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	}
}

// GetType returns the response type
func (r *ErrorResponse) GetType() string {
	return "error"
}

// IsSuccess returns whether the response indicates success
func (r *ErrorResponse) IsSuccess() bool {
	return false
}

// WithDetails adds details to the error response
func (r *ErrorResponse) WithDetails(details map[string]interface{}) *ErrorResponse {
	r.Details = details
	return r
}

// WithRequestID adds a request ID to the error response
func (r *ErrorResponse) WithRequestID(requestID string) *ErrorResponse {
	r.RequestID = requestID
	return r
}

// WithTimestamp sets a custom timestamp
func (r *ErrorResponse) WithTimestamp(timestamp time.Time) *ErrorResponse {
	r.Timestamp = timestamp
	return r
}

// ToJSON returns the error response as a map for JSON serialization
func (r *ErrorResponse) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"success":   r.Success,
		"error":     r.Error,
		"message":   r.Message,
		"timestamp": r.Timestamp.Format(time.RFC3339),
	}

	if r.Code != "" {
		result["code"] = r.Code
	}

	if r.RequestID != "" {
		result["request_id"] = r.RequestID
	}

	if r.Details != nil && len(r.Details) > 0 {
		result["details"] = r.Details
	}

	return result
}

// GetHTTPStatusCode returns the appropriate HTTP status code for the error
func (r *ErrorResponse) GetHTTPStatusCode() int {
	switch r.Code {
	case "VALIDATION_ERROR", "BAD_REQUEST":
		return 400
	case "UNAUTHORIZED":
		return 401
	case "FORBIDDEN":
		return 403
	case "NOT_FOUND":
		return 404
	case "CONFLICT":
		return 409
	case "RATE_LIMIT_EXCEEDED":
		return 429
	case "INTERNAL_ERROR":
		return 500
	default:
		return 500
	}
}

// Common error response constructors for consistency

// NewBadRequestResponse creates a bad request error response
func NewBadRequestResponse(message, requestID string) *ErrorResponse {
	return NewErrorResponse("BAD_REQUEST", message, "BAD_REQUEST").WithRequestID(requestID)
}

// NewUnauthorizedResponse creates an unauthorized error response
func NewUnauthorizedResponse(message, requestID string) *ErrorResponse {
	return NewSecurityErrorResponse("UNAUTHORIZED", message, "UNAUTHORIZED", requestID)
}

// NewForbiddenResponse creates a forbidden error response
func NewForbiddenResponse(message, requestID string) *ErrorResponse {
	return NewSecurityErrorResponse("FORBIDDEN", message, "FORBIDDEN", requestID)
}

// NewNotFoundResponse creates a not found error response
func NewNotFoundResponse(message, requestID string) *ErrorResponse {
	return NewErrorResponse("NOT_FOUND", message, "NOT_FOUND").WithRequestID(requestID)
}

// NewConflictResponse creates a conflict error response
func NewConflictResponse(message, requestID string) *ErrorResponse {
	return NewErrorResponse("CONFLICT", message, "CONFLICT").WithRequestID(requestID)
}

// NewInternalErrorResponse creates an internal error response
func NewInternalErrorResponse(requestID string) *ErrorResponse {
	return NewErrorResponse("INTERNAL_ERROR", "An internal server error occurred", "INTERNAL_ERROR").WithRequestID(requestID)
}

// NewRateLimitResponse creates a rate limit error response
func NewRateLimitResponse(message, retryAfter, requestID string) *ErrorResponse {
	details := map[string]interface{}{}
	if retryAfter != "" {
		details["retry_after"] = retryAfter
	}
	return NewErrorResponse("RATE_LIMIT_EXCEEDED", message, "RATE_LIMIT_EXCEEDED").
		WithRequestID(requestID).
		WithDetails(details)
}
