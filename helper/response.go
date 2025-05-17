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
// @Failure 400 {object} Response "Bad request"
// @Failure 500 {object} Response "Internal server error"
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
		Error:      []string{},
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
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
