package utils

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
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
func (r *Responder) SendSuccess(c interface{}, statusCode int, data interface{}) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		r.logger.Error("Invalid context type for responder")
		return
	}

	response := gin.H{
		"success":   true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      data,
	}

	// Add request ID if available
	if requestID := ginCtx.GetString("request_id"); requestID != "" {
		response["request_id"] = requestID
	}

	ginCtx.JSON(statusCode, response)
}

// SendError sends an error response
func (r *Responder) SendError(c interface{}, statusCode int, message string, err error) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		r.logger.Error("Invalid context type for responder")
		return
	}

	response := gin.H{
		"success":   false,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"message":   message,
	}

	// Add request ID if available
	if requestID := ginCtx.GetString("request_id"); requestID != "" {
		response["request_id"] = requestID
	}

	// Add error details based on error type
	if err != nil {
		switch e := err.(type) {
		case *errors.ValidationError:
			response["code"] = "VALIDATION_ERROR"
			response["details"] = e.Details()
			statusCode = http.StatusBadRequest
		case *errors.NotFoundError:
			response["code"] = "NOT_FOUND"
			response["details"] = e.Error()
			statusCode = http.StatusNotFound
		case *errors.ConflictError:
			response["code"] = "CONFLICT"
			response["details"] = e.Error()
			statusCode = http.StatusConflict
		case *errors.UnauthorizedError:
			response["code"] = "UNAUTHORIZED"
			response["details"] = e.Error()
			statusCode = http.StatusUnauthorized
		case *errors.ForbiddenError:
			response["code"] = "FORBIDDEN"
			response["details"] = e.Error()
			statusCode = http.StatusForbidden
		case *errors.InternalError:
			response["code"] = "INTERNAL_ERROR"
			response["details"] = "An internal server error occurred"
			statusCode = http.StatusInternalServerError
		default:
			response["code"] = "GENERIC_ERROR"
			response["details"] = err.Error()
		}
	} else {
		response["code"] = "GENERIC_ERROR"
	}

	// Log the error
	r.logger.Error("HTTP error response",
		"status_code", statusCode,
		"message", message,
		"error", err,
		"path", ginCtx.Request.URL.Path,
		"method", ginCtx.Request.Method,
		"ip", ginCtx.ClientIP(),
	)

	ginCtx.JSON(statusCode, response)
}

// SendValidationError sends a validation error response
func (r *Responder) SendValidationError(c interface{}, errors []interface{}) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		r.logger.Error("Invalid context type for responder")
		return
	}

	response := gin.H{
		"success":   false,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"code":      "VALIDATION_ERROR",
		"message":   "Validation failed",
		"errors":    errors,
	}

	// Add request ID if available
	if requestID := ginCtx.GetString("request_id"); requestID != "" {
		response["request_id"] = requestID
	}

	// Log validation errors
	r.logger.Warn("Validation error response",
		"errors", errors,
		"path", ginCtx.Request.URL.Path,
		"method", ginCtx.Request.Method,
		"ip", ginCtx.ClientIP(),
	)

	ginCtx.JSON(http.StatusBadRequest, response)
}

// SendCreated sends a 201 Created response
func (r *Responder) SendCreated(c interface{}, data interface{}) {
	r.SendSuccess(c, http.StatusCreated, data)
}

// SendNoContent sends a 204 No Content response
func (r *Responder) SendNoContent(c interface{}) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		r.logger.Error("Invalid context type for responder")
		return
	}

	ginCtx.Status(http.StatusNoContent)
}

// SendUnauthorized sends a 401 Unauthorized response
func (r *Responder) SendUnauthorized(c interface{}, message string) {
	if message == "" {
		message = "Unauthorized access"
	}
	r.SendError(c, http.StatusUnauthorized, message, errors.NewUnauthorizedError(message))
}

// SendForbidden sends a 403 Forbidden response
func (r *Responder) SendForbidden(c interface{}, message string) {
	if message == "" {
		message = "Access forbidden"
	}
	r.SendError(c, http.StatusForbidden, message, errors.NewForbiddenError(message))
}

// SendNotFound sends a 404 Not Found response
func (r *Responder) SendNotFound(c interface{}, message string) {
	if message == "" {
		message = "Resource not found"
	}
	r.SendError(c, http.StatusNotFound, message, errors.NewNotFoundError(message))
}

// SendConflict sends a 409 Conflict response
func (r *Responder) SendConflict(c interface{}, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	r.SendError(c, http.StatusConflict, message, errors.NewConflictError(message))
}

// SendInternalError sends a 500 Internal Server Error response
func (r *Responder) SendInternalError(c interface{}, message string, err error) {
	if message == "" {
		message = "Internal server error"
	}
	r.SendError(c, http.StatusInternalServerError, message, errors.NewInternalError(err))
}

// SendBadRequest sends a 400 Bad Request response
func (r *Responder) SendBadRequest(c interface{}, message string, err error) {
	if message == "" {
		message = "Bad request"
	}
	r.SendError(c, http.StatusBadRequest, message, err)
}

// SendPaginatedResponse sends a paginated response
func (r *Responder) SendPaginatedResponse(c interface{}, data interface{}, total, limit, offset int) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		r.logger.Error("Invalid context type for responder")
		return
	}

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
	if requestID := ginCtx.GetString("request_id"); requestID != "" {
		response["request_id"] = requestID
	}

	ginCtx.JSON(http.StatusOK, response)
}

// SendFile sends a file response
func (r *Responder) SendFile(c interface{}, filepath, filename string) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		r.logger.Error("Invalid context type for responder")
		return
	}

	ginCtx.FileAttachment(filepath, filename)
}

// SendRedirect sends a redirect response
func (r *Responder) SendRedirect(c interface{}, location string, permanent bool) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		r.logger.Error("Invalid context type for responder")
		return
	}

	statusCode := http.StatusFound // 302
	if permanent {
		statusCode = http.StatusMovedPermanently // 301
	}

	ginCtx.Redirect(statusCode, location)
}

// SendJSON sends a JSON response with custom status code
func (r *Responder) SendJSON(c interface{}, statusCode int, data interface{}) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		r.logger.Error("Invalid context type for responder")
		return
	}

	ginCtx.JSON(statusCode, data)
}

// SendXML sends an XML response
func (r *Responder) SendXML(c interface{}, statusCode int, data interface{}) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		r.logger.Error("Invalid context type for responder")
		return
	}

	ginCtx.XML(statusCode, data)
}

// SendYAML sends a YAML response
func (r *Responder) SendYAML(c interface{}, statusCode int, data interface{}) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		r.logger.Error("Invalid context type for responder")
		return
	}

	ginCtx.YAML(statusCode, data)
}
