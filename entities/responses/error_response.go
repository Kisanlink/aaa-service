package responses

import "time"

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Message   string                 `json:"message"`
	Code      int                    `json:"code,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(error, message string, code int) *ErrorResponse {
	return &ErrorResponse{
		Error:     error,
		Message:   message,
		Code:      code,
		Timestamp: time.Now(),
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
