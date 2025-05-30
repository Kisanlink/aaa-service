package helper

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response represents a standard API response format
// @name Response
// @Description Standard API response structure
// @Success 200 {object} Response "Successful operation"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
type Response struct {
	// HTTP status code
	// @example 200
	StatusCode int `json:"status_code"`

	// Indicates if the request was successfully processed
	// @example true
	Success bool `json:"success"`

	// Human-readable message about the response
	// @example "Request processed successfully"
	Message string `json:"message"`

	// The actual data payload (can be any type)
	// @example {"id": 1, "name": "John Doe"}
	Data interface{} `json:"data"`

	// List of error messages (if any)
	// @example ["Invalid email format", "Password too short"]
	Error []string `json:"error"`

	// Timestamp of when the response was generated
	// @example "2023-05-15T10:00:00Z"
	Timestamp string `json:"timestamp"`
}

// SendSuccessResponse sends a standardized success response
func SendSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := Response{
		StatusCode: statusCode,
		Success:    true,
		Message:    message,
		Data:       data,
		Error:      nil, // Initialize as nil instead of empty array
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// ErrorResponse represents an API error response format with array of error messages and nil data
// @name ErrorResponse
// @Description Standard API error response structure with multiple error messages
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
type ErrorResponse struct {
	// HTTP status code
	// @example 400
	StatusCode int `json:"status_code"`

	// Indicates if the request was successfully processed (always false for error responses)
	// @example false
	Success bool `json:"success"`

	// Human-readable summary message about the error
	// @example "Validation failed"
	Message string `json:"message"`

	// The actual data payload (nil for error responses)
	// @example null
	Data interface{} `json:"data"`

	// List of error messages describing what went wrong
	// @example ["Invalid email format", "Password must be at least 8 characters"]
	Errors []string `json:"errors"`

	// Timestamp of when the response was generated
	// @example "2023-05-15T10:00:00Z"
	Timestamp string `json:"timestamp"`
}

// SendErrorResponse sends a standardized error response
func SendErrorResponse(w http.ResponseWriter, statusCode int, errors []string) {
	response := Response{
		StatusCode: statusCode,
		Success:    false,
		Message:    "Error occurred",
		Data:       nil,
		Error:      errors,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
