package errors

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Code      string                 `json:"code,omitempty"`
	Field     string                 `json:"field,omitempty"`
	Value     interface{}            `json:"value,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
	Path      string                 `json:"path,omitempty"`
	Method    string                 `json:"method,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// ErrorMiddleware provides HTTP error handling middleware
type ErrorMiddleware struct {
	logger  *zap.Logger
	handler *ErrorHandler
}

// NewErrorMiddleware creates a new error middleware
func NewErrorMiddleware(logger *zap.Logger, handler *ErrorHandler) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger:  logger,
		handler: handler,
	}
}

// HandleError handles errors and returns appropriate HTTP responses
func (em *ErrorMiddleware) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Handle the error using the error handler
	em.handler.Handle(r.Context(), err)

	// Get error details
	errorType := GetErrorType(err)
	httpStatus := GetHTTPStatus(err)

	// Create error response
	errorResp := ErrorResponse{
		Error:     errorType,
		Type:      string(errorType),
		Message:   err.Error(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
		Method:    r.Method,
	}

	// Add request ID if available
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		errorResp.RequestID = requestID
	}

	// Add custom error details if available
	if customErr, ok := err.(*CustomError); ok {
		errorResp.Code = customErr.Code
		errorResp.Field = customErr.Field
		errorResp.Value = customErr.Value
		errorResp.Details = customErr.Details
	}

	// Log the error
	em.logger.Error("HTTP error occurred",
		zap.Error(err),
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method),
		zap.Int("status", httpStatus),
		zap.String("error_type", string(errorType)),
	)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	// Write error response
	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		em.logger.Error("Failed to encode error response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Middleware returns an HTTP middleware function
func (em *ErrorMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer that captures panics
		responseWriter := &panicResponseWriter{
			ResponseWriter:  w,
			errorMiddleware: em,
		}

		defer func() {
			if r := recover(); r != nil {
				panicErr := NewInternalError("panic occurred", nil).
					WithDetails(map[string]interface{}{
						"panic_value": r,
					})
				em.HandleError(w, r, panicErr)
			}
		}()

		next.ServeHTTP(responseWriter, r)
	})
}

// panicResponseWriter wraps http.ResponseWriter to capture panics
type panicResponseWriter struct {
	http.ResponseWriter
	errorMiddleware *ErrorMiddleware
	statusWritten   bool
}

func (prw *panicResponseWriter) WriteHeader(statusCode int) {
	if !prw.statusWritten {
		prw.ResponseWriter.WriteHeader(statusCode)
		prw.statusWritten = true
	}
}

func (prw *panicResponseWriter) Write(data []byte) (int, error) {
	if !prw.statusWritten {
		prw.WriteHeader(http.StatusOK)
	}
	return prw.ResponseWriter.Write(data)
}

// Global error middleware instance
var globalErrorMiddleware *ErrorMiddleware

// SetGlobalErrorMiddleware sets the global error middleware
func SetGlobalErrorMiddleware(logger *zap.Logger, handler *ErrorHandler) {
	globalErrorMiddleware = NewErrorMiddleware(logger, handler)
}

// GetGlobalErrorMiddleware returns the global error middleware
func GetGlobalErrorMiddleware() *ErrorMiddleware {
	return globalErrorMiddleware
}

// HandleHTTPError handles HTTP errors using the global middleware
func HandleHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	if globalErrorMiddleware != nil {
		globalErrorMiddleware.HandleError(w, r, err)
	} else {
		// Fallback to basic error handling
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ErrorHandlerFunc is a function that handles HTTP requests and may return an error
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// HandleWithError wraps an HTTP handler function with error handling
func HandleWithError(handler ErrorHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			HandleHTTPError(w, r, err)
		}
	}
}

// HandleWithRecovery wraps an HTTP handler function with panic recovery
func HandleWithRecovery(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				panicErr := NewInternalError("panic occurred", nil).
					WithDetails(map[string]interface{}{
						"panic_value": r,
					})
				HandleHTTPError(w, r, panicErr)
			}
		}()
		handler(w, r)
	}
}

// HandleWithErrorAndRecovery wraps an HTTP handler function with both error handling and panic recovery
func HandleWithErrorAndRecovery(handler ErrorHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				panicErr := NewInternalError("panic occurred", nil).
					WithDetails(map[string]interface{}{
						"panic_value": r,
					})
				HandleHTTPError(w, r, panicErr)
			}
		}()

		if err := handler(w, r); err != nil {
			HandleHTTPError(w, r, err)
		}
	}
}
