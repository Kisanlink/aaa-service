package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/internal/entities/responses"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandlerConfig holds configuration for error handling
type ErrorHandlerConfig struct {
	Logger                *zap.Logger
	IncludeStackTrace     bool
	LogSensitiveErrors    bool
	EnableDetailedLogging bool
}

// NewErrorHandlerConfig creates a new error handler configuration
func NewErrorHandlerConfig(logger *zap.Logger) *ErrorHandlerConfig {
	return &ErrorHandlerConfig{
		Logger:                logger,
		IncludeStackTrace:     false, // Never include stack traces in production
		LogSensitiveErrors:    true,  // Always log sensitive errors for security monitoring
		EnableDetailedLogging: true,  // Enable detailed logging for debugging
	}
}

// SecureErrorHandler provides secure error handling that doesn't leak sensitive information
func SecureErrorHandler(config *ErrorHandlerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors[0].Err
			requestID := c.GetString("request_id")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Log the actual error for debugging (server-side only)
			logFields := []zap.Field{
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("request_id", requestID),
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_agent", c.GetHeader("User-Agent")),
				zap.Int("status_code", c.Writer.Status()),
			}

			// Add user context if available
			if userID, exists := c.Get("user_id"); exists {
				if uid, ok := userID.(string); ok {
					logFields = append(logFields, zap.String("user_id", uid))
				}
			}

			// Add additional context for detailed logging
			if config.EnableDetailedLogging {
				logFields = append(logFields,
					zap.String("referer", c.GetHeader("Referer")),
					zap.String("x_forwarded_for", c.GetHeader("X-Forwarded-For")),
					zap.Int64("content_length", c.Request.ContentLength),
				)
			}

			// Create standardized error response
			errorResponse := responses.NewErrorResponseFromError(err, requestID)
			statusCode := errorResponse.GetHTTPStatusCode()

			// Log based on error severity
			switch {
			case statusCode >= 500:
				config.Logger.Error("Internal server error", logFields...)
			case statusCode >= 400 && statusCode < 500:
				if isSensitiveError(err) && config.LogSensitiveErrors {
					config.Logger.Warn("Security-related error", logFields...)
				} else {
					config.Logger.Info("Client error", logFields...)
				}
			default:
				config.Logger.Debug("Request completed with error", logFields...)
			}

			// Send standardized error response
			c.JSON(statusCode, errorResponse.ToJSON())
		}
	}
}

// StructuredErrorHandler provides structured error handling with audit logging
func StructuredErrorHandler(config *ErrorHandlerConfig, auditLogger AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors[0].Err
			requestID := c.GetString("request_id")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Create error context
			errorContext := &ErrorContext{
				Error:     err,
				RequestID: requestID,
				Path:      c.Request.URL.Path,
				Method:    c.Request.Method,
				ClientIP:  c.ClientIP(),
				UserAgent: c.GetHeader("User-Agent"),
				UserID:    getUserIDFromContext(c),
				Timestamp: time.Now().UTC(),
			}

			// Log structured error
			logStructuredError(config.Logger, errorContext)

			// Log security audit if needed
			if isSensitiveError(err) && auditLogger != nil {
				auditLogger.LogSecurityError(c.Request.Context(), errorContext)
			}

			// Create and send response
			errorResponse := responses.NewErrorResponseFromError(err, requestID)
			statusCode := errorResponse.GetHTTPStatusCode()

			c.JSON(statusCode, errorResponse.ToJSON())
		}
	}
}

// ErrorContext holds context information about an error
type ErrorContext struct {
	Error     error
	RequestID string
	Path      string
	Method    string
	ClientIP  string
	UserAgent string
	UserID    string
	Timestamp time.Time
}

// AuditLogger interface for audit logging
type AuditLogger interface {
	LogSecurityError(ctx interface{}, errorContext *ErrorContext)
}

// logStructuredError logs errors with structured information
func logStructuredError(logger *zap.Logger, ctx *ErrorContext) {
	fields := []zap.Field{
		zap.Error(ctx.Error),
		zap.String("request_id", ctx.RequestID),
		zap.String("path", ctx.Path),
		zap.String("method", ctx.Method),
		zap.String("client_ip", ctx.ClientIP),
		zap.String("user_agent", ctx.UserAgent),
		zap.Time("timestamp", ctx.Timestamp),
	}

	if ctx.UserID != "" {
		fields = append(fields, zap.String("user_id", ctx.UserID))
	}

	// Add error type information
	errorType := getErrorType(ctx.Error)
	fields = append(fields, zap.String("error_type", errorType))

	// Log with appropriate level based on error type
	switch errorType {
	case "INTERNAL_ERROR":
		logger.Error("Internal server error occurred", fields...)
	case "UNAUTHORIZED", "FORBIDDEN":
		logger.Warn("Security error occurred", fields...)
	case "VALIDATION_ERROR", "BAD_REQUEST", "NOT_FOUND", "CONFLICT":
		logger.Info("Client error occurred", fields...)
	default:
		logger.Error("Unknown error occurred", fields...)
	}
}

// isSensitiveError checks if an error is security-sensitive
func isSensitiveError(err error) bool {
	switch err.(type) {
	case *errors.UnauthorizedError, *errors.ForbiddenError:
		return true
	default:
		return false
	}
}

// getErrorType returns the error type as a string
func getErrorType(err error) string {
	switch err.(type) {
	case *errors.ValidationError:
		return "VALIDATION_ERROR"
	case *errors.BadRequestError:
		return "BAD_REQUEST"
	case *errors.UnauthorizedError:
		return "UNAUTHORIZED"
	case *errors.ForbiddenError:
		return "FORBIDDEN"
	case *errors.NotFoundError:
		return "NOT_FOUND"
	case *errors.ConflictError:
		return "CONFLICT"
	case *errors.InternalError:
		return "INTERNAL_ERROR"
	default:
		return "GENERIC_ERROR"
	}
}

// getUserIDFromContext extracts user ID from gin context
func getUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	return ""
}

// generateRequestID generates a new request ID if none exists
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// ErrorHandler provides backward compatibility
func ErrorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) > 0 {
		err := c.Errors[0].Err
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = generateRequestID()
		}

		switch e := err.(type) {
		case *helper.CustomError:
			// Handle legacy CustomError with standardized response
			errorResponse := responses.NewErrorResponse("REQUEST_ERROR", e.Message, "REQUEST_ERROR").
				WithRequestID(requestID)
			c.JSON(e.Code, errorResponse.ToJSON())
		default:
			// Default case with standardized response
			errorResponse := responses.NewInternalErrorResponse(requestID)
			c.JSON(http.StatusInternalServerError, errorResponse.ToJSON())
		}
	}
}

// ValidationErrorHandler specifically handles validation errors
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors[0].Err
			requestID := c.GetString("request_id")
			if requestID == "" {
				requestID = generateRequestID()
			}

			if validationErr, ok := err.(*errors.ValidationError); ok {
				response := responses.NewValidationErrorResponse(
					validationErr.Error(),
					validationErr.Details(),
					requestID,
				)
				c.JSON(http.StatusBadRequest, response)
				return
			}

			// Fall back to standard error handling
			errorResponse := responses.NewErrorResponseFromError(err, requestID)
			c.JSON(errorResponse.GetHTTPStatusCode(), errorResponse.ToJSON())
		}
	}
}

// SecurityErrorHandler specifically handles security-related errors
func SecurityErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors[0].Err
			requestID := c.GetString("request_id")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Log security errors with additional context
			if isSensitiveError(err) {
				logger.Warn("Security error detected",
					zap.Error(err),
					zap.String("request_id", requestID),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.GetHeader("User-Agent")),
					zap.String("user_id", getUserIDFromContext(c)),
				)
			}

			errorResponse := responses.NewErrorResponseFromError(err, requestID)
			c.JSON(errorResponse.GetHTTPStatusCode(), errorResponse.ToJSON())
		}
	}
}
