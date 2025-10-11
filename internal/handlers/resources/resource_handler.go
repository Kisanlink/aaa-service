package resources

import (
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/services/resources"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResourceHandler handles resource-related HTTP requests
type ResourceHandler struct {
	resourceService resources.ResourceService
	validator       interfaces.Validator
	responder       interfaces.Responder
	logger          *zap.Logger
}

// NewResourceHandler creates a new ResourceHandler instance
func NewResourceHandler(
	resourceService resources.ResourceService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *ResourceHandler {
	return &ResourceHandler{
		resourceService: resourceService,
		validator:       validator,
		responder:       responder,
		logger:          logger.Named("resource_handler"),
	}
}

// getRequestID extracts the request ID from the Gin context
func (h *ResourceHandler) getRequestID(c *gin.Context) string {
	if reqID, exists := c.Get("request_id"); exists {
		if id, ok := reqID.(string); ok {
			return id
		}
	}
	return c.GetString("request_id")
}

// parseIntParam safely parses an integer parameter with a default value
func (h *ResourceHandler) parseIntParam(c *gin.Context, param string, defaultValue int) int {
	value := c.DefaultQuery(param, "")
	if value == "" {
		return defaultValue
	}

	// Try to parse the string to int
	var intValue int
	if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
		h.logger.Warn("Invalid integer parameter",
			zap.String("param", param),
			zap.String("value", value),
			zap.Error(err))
		return defaultValue
	}

	return intValue
}
