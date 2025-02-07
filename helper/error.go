package helper

import "fmt"

// CustomError defines a custom error type that includes the code, message, and an optional underlying error
type CustomError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"error,omitempty"` // Holds the original error (optional)
}

// New creates a new CustomError with a message and code
func New(code int, message string, err error) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error implements the error interface and returns a string representation of the CustomError
func (e *CustomError) Error() string {
	if e.Err != nil {
		// If there's an underlying error, include it in the error string
		return fmt.Sprintf("Error %d: %s - %v", e.Code, e.Message, e.Err)
	}
	// If there's no underlying error, return just the message and code
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// Example usage:
func ExampleFunction() {
	// Simulating an original error
	originalErr := fmt.Errorf("a database connection error occurred")

	// Create a custom error with original error, code, and message
	customErr := New(500, "Internal Server Error", originalErr)
	fmt.Println(customErr.Error())
}
