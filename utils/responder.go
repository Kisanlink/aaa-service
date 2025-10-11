package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/responses"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Responder implements the Responder interface
type Responder struct {
	logger interfaces.Logger
}

// NewResponder creates a new Responder instance
func NewResponder(logger interfaces.Logger) interfaces.Responder {
	return &Responder{
		logger: logger,
	}
}

// SendSuccess sends a successful response
func (r *Responder) SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	// Propagate context/token info in headers
	if requestID := c.GetString("request_id"); requestID != "" {
		c.Header("X-Request-Id", requestID)
	}
	if uid, ok := c.Get("user_id"); ok {
		if s, ok := uid.(string); ok && s != "" {
			c.Header("X-User-Id", s)
		}
	}
	if authz := c.GetHeader("Authorization"); authz != "" {
		c.Header("X-Authorization", authz)
	}

	response := gin.H{
		"success":   true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      data,
	}

	// Add request ID if available
	if requestID := c.GetString("request_id"); requestID != "" {
		response["request_id"] = requestID
	}

	c.JSON(statusCode, response)
}

// SendError sends an error response using standardized error format
func (r *Responder) SendError(c *gin.Context, statusCode int, message string, err error) {
	// Propagate context/token info in headers
	if requestID := c.GetString("request_id"); requestID != "" {
		c.Header("X-Request-Id", requestID)
	}
	if uid, ok := c.Get("user_id"); ok {
		if s, ok := uid.(string); ok && s != "" {
			c.Header("X-User-Id", s)
		}
	}
	if authz := c.GetHeader("Authorization"); authz != "" {
		c.Header("X-Authorization", authz)
	}

	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	var errorResponse interface{}

	// Create standardized error response
	if err != nil {
		// Use the new error response system
		switch e := err.(type) {
		case *errors.ValidationError:
			response := &responses.ValidationErrorResponse{
				Success:   false,
				Error:     "VALIDATION_ERROR",
				Message:   message,
				Code:      "VALIDATION_ERROR",
				Errors:    e.Details(),
				Timestamp: time.Now().UTC(),
				RequestID: requestID,
			}
			errorResponse = response
			statusCode = http.StatusBadRequest
		case *errors.BadRequestError:
			errorResponse = responses.NewBadRequestResponse(message, requestID).
				WithDetails(map[string]interface{}{"errors": e.Details()}).ToJSON()
			statusCode = http.StatusBadRequest
		case *errors.UnauthorizedError:
			errorResponse = responses.NewUnauthorizedResponse(message, requestID).ToJSON()
			statusCode = http.StatusUnauthorized
		case *errors.ForbiddenError:
			errorResponse = responses.NewForbiddenResponse(message, requestID).ToJSON()
			statusCode = http.StatusForbidden
		case *errors.NotFoundError:
			errorResponse = responses.NewNotFoundResponse(message, requestID).ToJSON()
			statusCode = http.StatusNotFound
		case *errors.ConflictError:
			errorResponse = responses.NewConflictResponse(message, requestID).ToJSON()
			statusCode = http.StatusConflict
		case *errors.InternalError:
			errorResponse = responses.NewInternalErrorResponse(requestID).ToJSON()
			statusCode = http.StatusInternalServerError
		default:
			errorResponse = responses.NewErrorResponse("GENERIC_ERROR", message, "GENERIC_ERROR").
				WithRequestID(requestID).ToJSON()
		}
	} else {
		// No specific error provided
		errorResponse = responses.NewErrorResponse("GENERIC_ERROR", message, "GENERIC_ERROR").
			WithRequestID(requestID).ToJSON()
	}

	// Log the error with structured logging
	logFields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Int("status", statusCode),
		zap.String("message", message),
		zap.String("client_ip", c.ClientIP()),
	}

	if err != nil {
		logFields = append(logFields, zap.Error(err))
	}

	if userID := getUserIDFromContext(c); userID != "" {
		logFields = append(logFields, zap.String("user_id", userID))
	}

	// Log with appropriate level based on status code
	switch {
	case statusCode >= 500:
		r.logger.Error("HTTP server error response", logFields...)
	case statusCode >= 400:
		r.logger.Warn("HTTP client error response", logFields...)
	default:
		r.logger.Info("HTTP error response", logFields...)
	}

	c.JSON(statusCode, errorResponse)
}

// SendErrorWithContext sends an error response with additional context
func (r *Responder) SendErrorWithContext(c *gin.Context, statusCode int, message string, err error, context map[string]interface{}) {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	var errorResponse *responses.ErrorResponse

	if err != nil {
		errorResponse = responses.NewErrorResponseFromError(err, requestID)
	} else {
		errorResponse = responses.NewErrorResponse("GENERIC_ERROR", message, "GENERIC_ERROR").
			WithRequestID(requestID)
	}

	// Add additional context
	if context != nil {
		if errorResponse.Details == nil {
			errorResponse.Details = make(map[string]interface{})
		}
		for k, v := range context {
			errorResponse.Details[k] = v
		}
	}

	// Set headers
	if requestID != "" {
		c.Header("X-Request-Id", requestID)
	}
	if uid, ok := c.Get("user_id"); ok {
		if s, ok := uid.(string); ok && s != "" {
			c.Header("X-User-Id", s)
		}
	}

	// Log with context
	logFields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Int("status", statusCode),
		zap.String("message", message),
		zap.Any("context", context),
	}

	if err != nil {
		logFields = append(logFields, zap.Error(err))
	}

	r.logger.Error("HTTP error response with context", logFields...)

	c.JSON(statusCode, errorResponse.ToJSON())
}

// SendValidationError sends a validation error response
func (r *Responder) SendValidationError(c *gin.Context, errors []string) {
	if requestID := c.GetString("request_id"); requestID != "" {
		c.Header("X-Request-Id", requestID)
	}
	if uid, ok := c.Get("user_id"); ok {
		if s, ok := uid.(string); ok && s != "" {
			c.Header("X-User-Id", s)
		}
	}
	if authz := c.GetHeader("Authorization"); authz != "" {
		c.Header("X-Authorization", authz)
	}

	response := gin.H{
		"success":   false,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"code":      "VALIDATION_ERROR",
		"message":   "Validation failed",
		"errors":    errors,
	}

	// Add request ID if available
	if requestID := c.GetString("request_id"); requestID != "" {
		response["request_id"] = requestID
	}

	// Log validation errors
	r.logger.Warn("Validation error response",
		zap.Strings("errors", errors),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("ip", c.ClientIP()),
	)

	c.JSON(http.StatusBadRequest, response)
}

// SendInternalError sends a 500 Internal Server Error response
func (r *Responder) SendInternalError(c *gin.Context, err error) {
	message := "Internal server error"
	r.SendError(c, http.StatusInternalServerError, message, errors.NewInternalError(err))
}

// Additional utility methods (not part of the interface but useful)

// SendCreated sends a 201 Created response
func (r *Responder) SendCreated(c *gin.Context, data interface{}) {
	r.SendSuccess(c, http.StatusCreated, data)
}

// SendNoContent sends a 204 No Content response
func (r *Responder) SendNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// SendUnauthorized sends a 401 Unauthorized response
func (r *Responder) SendUnauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized access"
	}
	r.SendError(c, http.StatusUnauthorized, message, errors.NewUnauthorizedError(message))
}

// SendForbidden sends a 403 Forbidden response
func (r *Responder) SendForbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Access forbidden"
	}
	r.SendError(c, http.StatusForbidden, message, errors.NewForbiddenError(message))
}

// SendNotFound sends a 404 Not Found response
func (r *Responder) SendNotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	r.SendError(c, http.StatusNotFound, message, errors.NewNotFoundError(message))
}

// SendConflict sends a 409 Conflict response
func (r *Responder) SendConflict(c *gin.Context, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	r.SendError(c, http.StatusConflict, message, errors.NewConflictError(message))
}

// SendBadRequest sends a 400 Bad Request response
func (r *Responder) SendBadRequest(c *gin.Context, message string, err error) {
	if message == "" {
		message = "Bad request"
	}
	r.SendError(c, http.StatusBadRequest, message, err)
}

// SendPaginatedResponse sends a paginated response
func (r *Responder) SendPaginatedResponse(c *gin.Context, data interface{}, total, limit, offset int) {
	response := gin.H{
		"success":   true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      data,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
			"pages":  (total + limit - 1) / limit, // Calculate total pages
		},
	}

	// Add request ID if available
	if requestID := c.GetString("request_id"); requestID != "" {
		response["request_id"] = requestID
	}

	c.JSON(http.StatusOK, response)
}

// SendFile sends a file response
func (r *Responder) SendFile(c *gin.Context, filepath, filename string) {
	c.FileAttachment(filepath, filename)
}

// SendRedirect sends a redirect response
func (r *Responder) SendRedirect(c *gin.Context, location string, permanent bool) {
	statusCode := http.StatusFound // 302
	if permanent {
		statusCode = http.StatusMovedPermanently // 301
	}

	c.Redirect(statusCode, location)
}

// SendJSON sends a JSON response with custom status code
func (r *Responder) SendJSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// SendXML sends an XML response
func (r *Responder) SendXML(c *gin.Context, statusCode int, data interface{}) {
	c.XML(statusCode, data)
}

// SendYAML sends a YAML response
func (r *Responder) SendYAML(c *gin.Context, statusCode int, data interface{}) {
	c.YAML(statusCode, data)
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
