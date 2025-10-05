package utils

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/responses"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResponseUtils provides utilities for standardized API responses
type ResponseUtils struct {
	transformer interfaces.ResponseTransformer
	validator   interfaces.ResponseValidator
	logger      interfaces.Logger
}

// NewResponseUtils creates a new ResponseUtils instance
func NewResponseUtils(
	transformer interfaces.ResponseTransformer,
	validator interfaces.ResponseValidator,
	logger interfaces.Logger,
) *ResponseUtils {
	return &ResponseUtils{
		transformer: transformer,
		validator:   validator,
		logger:      logger,
	}
}

// SendSuccessResponse sends a standardized success response
func (ru *ResponseUtils) SendSuccessResponse(c *gin.Context, statusCode int, data interface{}, message string) {
	response := responses.StandardSuccessResponse{
		Message:   message,
		Data:      data,
		RequestID: ru.getRequestID(c),
		Metadata:  ru.getResponseMetadata(c),
		Timestamp: time.Now().UTC(),
		Success:   true,
	}

	c.JSON(statusCode, response)
}

// SendPaginatedResponse sends a standardized paginated response
func (ru *ResponseUtils) SendPaginatedResponse(c *gin.Context, data interface{}, total int, page, limit int) {
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	pagination := responses.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	response := responses.StandardPaginatedResponse{
		Data:       data,
		Pagination: pagination,
		RequestID:  ru.getRequestID(c),
		Metadata:   ru.getResponseMetadata(c),
		Timestamp:  time.Now().UTC(),
		Success:    true,
	}

	c.JSON(http.StatusOK, response)
}

// SendErrorResponse sends a standardized error response
func (ru *ResponseUtils) SendErrorResponse(c *gin.Context, statusCode int, errorCode, message string, details map[string]interface{}) {
	response := responses.StandardErrorResponse{
		Error:     errorCode,
		Message:   message,
		Code:      errorCode,
		RequestID: ru.getRequestID(c),
		Details:   details,
		Timestamp: time.Now().UTC(),
		Success:   false,
	}

	ru.logger.Error("API Error Response",
		zap.Int("status_code", statusCode),
		zap.String("error_code", errorCode),
		zap.String("message", message),
		zap.String("path", c.Request.URL.Path),
		zap.Any("details", details),
	)

	c.JSON(statusCode, response)
}

// SendValidationErrorResponse sends a standardized validation error response
func (ru *ResponseUtils) SendValidationErrorResponse(c *gin.Context, validationErrors []string) {
	details := map[string]interface{}{
		"validation_errors": validationErrors,
	}

	ru.SendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Request validation failed", details)
}

// SendTransformedResponse sends a response with data transformation and validation
func (ru *ResponseUtils) SendTransformedResponse(c *gin.Context, data interface{}, options interfaces.TransformOptions, message string) {
	// Transform the data if transformer is available
	if ru.transformer != nil {
		// Note: This is a simplified approach. In practice, you'd need to determine
		// the data type and call the appropriate transform method
		data = ru.transformData(data, options)
	}

	// Validate the response if validator is available
	if ru.validator != nil {
		if err := ru.validateResponse(data); err != nil {
			ru.logger.Warn("Response validation failed",
				zap.String("error", err.Error()),
				zap.String("path", c.Request.URL.Path),
			)
			// Continue anyway, but log the validation failure
		}
	}

	ru.SendSuccessResponse(c, http.StatusOK, data, message)
}

// Helper methods

// getRequestID extracts or generates a request ID
func (ru *ResponseUtils) getRequestID(c *gin.Context) *string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return &requestID
	}

	if requestID := c.GetString("request_id"); requestID != "" {
		return &requestID
	}

	return nil
}

// getResponseMetadata extracts metadata for the response
func (ru *ResponseUtils) getResponseMetadata(c *gin.Context) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Add processing time if available
	if startTime, exists := c.Get("start_time"); exists {
		if start, ok := startTime.(time.Time); ok {
			processingTime := time.Since(start)
			metadata["processing_time_ms"] = processingTime.Milliseconds()
		}
	}

	// Add API version if available
	if version := c.GetHeader("API-Version"); version != "" {
		metadata["api_version"] = version
	}

	// Add user context if available
	if userID := c.GetString("user_id"); userID != "" {
		metadata["user_id"] = userID
	}

	return metadata
}

// transformData applies transformation to data based on its type
func (ru *ResponseUtils) transformData(data interface{}, options interfaces.TransformOptions) interface{} {
	// This is a simplified implementation. In practice, you'd need more sophisticated
	// type detection and transformation logic
	return data
}

// validateResponse validates the response structure
func (ru *ResponseUtils) validateResponse(data interface{}) error {
	// Check for sensitive data
	if err := ru.validator.ValidateNoSensitiveData(data); err != nil {
		return err
	}

	// Add more validation as needed
	return nil
}

// CreatePaginationFromOffset creates pagination info from offset-based parameters
func CreatePaginationFromOffset(total, limit, offset int) responses.PaginationInfo {
	page := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	return responses.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    offset+limit < total,
		HasPrev:    offset > 0,
	}
}

// ValidateTransformOptions validates transformation options
func ValidateTransformOptions(options interfaces.TransformOptions) error {
	// Add validation logic for transform options if needed
	// For now, all options are valid
	return nil
}

// MergeTransformOptions merges two transform options, with override taking precedence
func MergeTransformOptions(base, override interfaces.TransformOptions) interfaces.TransformOptions {
	// Create a copy of base options
	merged := base

	// Override with non-zero values from override
	// Note: This is a simplified merge. You might want more sophisticated logic
	if override.IncludeProfile {
		merged.IncludeProfile = true
	}
	if override.IncludeContacts {
		merged.IncludeContacts = true
	}
	if override.IncludeRole {
		merged.IncludeRole = true
	}
	if override.IncludeUser {
		merged.IncludeUser = true
	}
	if override.IncludeAddress {
		merged.IncludeAddress = true
	}

	return merged
}
