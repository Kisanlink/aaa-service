package principals

import (
	"net/http"
	"strconv"

	principalRequests "github.com/Kisanlink/aaa-service/internal/entities/requests/principals"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/services/principals"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for principal and service operations
type Handler struct {
	principalService *principals.Service
	responder        interfaces.Responder
	logger           *zap.Logger
}

// NewPrincipalHandler creates a new principal handler instance
func NewPrincipalHandler(
	principalService *principals.Service,
	responder interfaces.Responder,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		principalService: principalService,
		responder:        responder,
		logger:           logger,
	}
}

// CreatePrincipal handles POST /principals
func (h *Handler) CreatePrincipal(c *gin.Context) {
	var req principalRequests.CreatePrincipalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind create principal request", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request format", err)
		return
	}

	response, err := h.principalService.CreatePrincipal(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create principal", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusCreated, response)
}

// CreateService handles POST /services
func (h *Handler) CreateService(c *gin.Context) {
	var req principalRequests.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind create service request", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request format", err)
		return
	}

	response, err := h.principalService.CreateService(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create service", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusCreated, response)
}

// GetPrincipal handles GET /principals/:id
func (h *Handler) GetPrincipal(c *gin.Context) {
	principalID := c.Param("id")
	if principalID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "principal ID is required", nil)
		return
	}

	response, err := h.principalService.GetPrincipal(c.Request.Context(), principalID)
	if err != nil {
		h.logger.Error("Failed to get principal", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// GetService handles GET /services/:id
func (h *Handler) GetService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "service ID is required", nil)
		return
	}

	response, err := h.principalService.GetService(c.Request.Context(), serviceID)
	if err != nil {
		h.logger.Error("Failed to get service", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// UpdatePrincipal handles PUT /principals/:id
func (h *Handler) UpdatePrincipal(c *gin.Context) {
	principalID := c.Param("id")
	if principalID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "principal ID is required", nil)
		return
	}

	var req principalRequests.UpdatePrincipalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind update principal request", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request format", err)
		return
	}

	response, err := h.principalService.UpdatePrincipal(c.Request.Context(), principalID, &req)
	if err != nil {
		h.logger.Error("Failed to update principal", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// DeletePrincipal handles DELETE /principals/:id
func (h *Handler) DeletePrincipal(c *gin.Context) {
	principalID := c.Param("id")
	if principalID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "principal ID is required", nil)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.responder.SendError(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	err := h.principalService.DeletePrincipal(c.Request.Context(), principalID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete principal", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, nil)
}

// DeleteService handles DELETE /services/:id
func (h *Handler) DeleteService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "service ID is required", nil)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.responder.SendError(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	err := h.principalService.DeleteService(c.Request.Context(), serviceID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete service", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, nil)
}

// ListPrincipals handles GET /principals
func (h *Handler) ListPrincipals(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	principalType := c.Query("type")
	organizationID := c.Query("organization_id")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Cap limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	response, err := h.principalService.ListPrincipals(c.Request.Context(), limit, offset, principalType, organizationID)
	if err != nil {
		h.logger.Error("Failed to list principals", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// ListServices handles GET /services
func (h *Handler) ListServices(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	organizationID := c.Query("organization_id")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Cap limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	response, err := h.principalService.ListServices(c.Request.Context(), limit, offset, organizationID)
	if err != nil {
		h.logger.Error("Failed to list services", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// GenerateAPIKey handles POST /services/generate-api-key
func (h *Handler) GenerateAPIKey(c *gin.Context) {
	apiKey, err := h.principalService.GenerateAPIKey()
	if err != nil {
		h.logger.Error("Failed to generate API key", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "failed to generate API key", err)
		return
	}

	response := map[string]string{
		"api_key": apiKey,
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// handleServiceError converts service errors to appropriate HTTP responses
func (h *Handler) handleServiceError(c *gin.Context, err error) {
	// This would need to be implemented based on your error types
	// For now, we'll return a generic internal server error
	h.responder.SendError(c, http.StatusInternalServerError, "internal server error", err)
}
