package organizations

import (
	"fmt"
	"net/http"
	"strconv"

	orgRequests "github.com/Kisanlink/aaa-service/internal/entities/requests/organizations"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for organization operations
type Handler struct {
	orgService interfaces.OrganizationService
	logger     *zap.Logger
	responder  interfaces.Responder
}

// NewOrganizationHandler creates a new organization handler instance
func NewOrganizationHandler(
	orgService interfaces.OrganizationService,
	logger *zap.Logger,
	responder interfaces.Responder,
) *Handler {
	return &Handler{
		orgService: orgService,
		logger:     logger,
		responder:  responder,
	}
}

// CreateOrganization handles POST /organizations
// @Summary Create a new organization
// @Description Create a new organization with the provided information
// @Tags organizations
// @Accept json
// @Produce json
// @Param organization body organizations.CreateOrganizationRequest true "Organization creation data"
// @Success 201 {object} organizationResponses.OrganizationResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/organizations [post]
func (h *Handler) CreateOrganization(c *gin.Context) {
	var req orgRequests.CreateOrganizationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for organization creation", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Create organization
	org, err := h.orgService.CreateOrganization(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create organization", zap.Error(err))

		switch {
		case errors.IsValidationError(err):
			h.responder.SendValidationError(c, []string{err.Error()})
		case errors.IsConflictError(err):
			h.responder.SendError(c, http.StatusConflict, "conflict", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("Organization created successfully",
		zap.String("created_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusCreated, org)
}

// GetOrganization handles GET /organizations/:id
// @Summary Get organization by ID
// @Description Retrieve an organization by its ID
// @Tags organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} organizationResponses.OrganizationResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/organizations/{id} [get]
func (h *Handler) GetOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "bad request", nil)
		return
	}

	org, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to retrieve organization", zap.Error(err), zap.String("org_id", orgID))

		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, org)
}

// UpdateOrganization handles PUT /organizations/:id
// @Summary Update organization
// @Description Update an existing organization with the provided information
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param organization body organizations.UpdateOrganizationRequest true "Organization update data"
// @Success 200 {object} organizationResponses.OrganizationResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/organizations/{id} [put]
func (h *Handler) UpdateOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "bad request", nil)
		return
	}

	var req orgRequests.UpdateOrganizationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for organization update", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Update organization
	org, err := h.orgService.UpdateOrganization(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("Failed to update organization", zap.Error(err), zap.String("org_id", orgID))

		switch {
		case errors.IsValidationError(err):
			h.responder.SendValidationError(c, []string{err.Error()})
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "not found", err)
		case errors.IsConflictError(err):
			h.responder.SendError(c, http.StatusConflict, "conflict", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("Organization updated successfully",
		zap.String("org_id", orgID),
		zap.String("updated_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusOK, org)
}

// DeleteOrganization handles DELETE /organizations/:id
// @Summary Delete organization
// @Description Delete an organization by its ID
// @Tags organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/organizations/{id} [delete]
func (h *Handler) DeleteOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", fmt.Errorf("organization ID is required"))
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	// Delete organization
	err := h.orgService.DeleteOrganization(c.Request.Context(), orgID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete organization", zap.Error(err), zap.String("org_id", orgID))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		case errors.IsValidationError(err):
			h.responder.SendError(c, http.StatusBadRequest, "validation failed", err)
		default:
			h.responder.SendError(c, http.StatusInternalServerError, "failed to delete organization", err)
		}
		return
	}

	h.logger.Info("Organization deleted successfully",
		zap.String("org_id", orgID),
		zap.String("deleted_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusOK, "organization deleted successfully")
}

// ListOrganizations handles GET /organizations
// @Summary List organizations
// @Description Retrieve a list of organizations with pagination
// @Tags organizations
// @Produce json
// @Param limit query int false "Number of organizations to return (default: 10, max: 100)"
// @Param offset query int false "Number of organizations to skip (default: 0)"
// @Param include_inactive query bool false "Include inactive organizations (default: false)"
// @Success 200 {array} organizationResponses.OrganizationResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/organizations [get]
func (h *Handler) ListOrganizations(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	includeInactiveStr := c.DefaultQuery("include_inactive", "false")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	includeInactive, err := strconv.ParseBool(includeInactiveStr)
	if err != nil {
		includeInactive = false
	}

	// Cap limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	// List organizations
	orgs, err := h.orgService.ListOrganizations(c.Request.Context(), limit, offset, includeInactive)
	if err != nil {
		h.logger.Error("Failed to list organizations", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "failed to list organizations", nil)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, orgs)
}

// GetOrganizationHierarchy handles GET /organizations/:id/hierarchy
// @Summary Get organization hierarchy
// @Description Retrieve the hierarchical structure of an organization
// @Tags organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} organizationResponses.OrganizationHierarchyResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/organizations/{id}/hierarchy [get]
func (h *Handler) GetOrganizationHierarchy(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	hierarchy, err := h.orgService.GetOrganizationHierarchy(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to retrieve organization hierarchy", zap.Error(err), zap.String("org_id", orgID))

		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendError(c, http.StatusInternalServerError, "failed to retrieve organization hierarchy", err)
		}
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, hierarchy)
}

// ActivateOrganization handles POST /organizations/:id/activate
// @Summary Activate organization
// @Description Activate an inactive organization
// @Tags organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/organizations/{id}/activate [post]
func (h *Handler) ActivateOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	// Activate organization
	err := h.orgService.ActivateOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to activate organization", zap.Error(err), zap.String("org_id", orgID))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		case errors.IsValidationError(err):
			h.responder.SendError(c, http.StatusBadRequest, "validation failed", err)
		default:
			h.responder.SendError(c, http.StatusInternalServerError, "failed to activate organization", err)
		}
		return
	}

	h.logger.Info("Organization activated successfully",
		zap.String("org_id", orgID),
		zap.String("activated_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusOK, "organization activated successfully")
}

// DeactivateOrganization handles POST /organizations/:id/deactivate
// @Summary Deactivate organization
// @Description Deactivate an active organization
// @Tags organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/organizations/{id}/deactivate [post]
func (h *Handler) DeactivateOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	// Deactivate organization
	err := h.orgService.DeactivateOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to deactivate organization", zap.Error(err), zap.String("org_id", orgID))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		case errors.IsValidationError(err):
			h.responder.SendError(c, http.StatusBadRequest, "validation failed", err)
		default:
			h.responder.SendError(c, http.StatusInternalServerError, "failed to deactivate organization", err)
		}
		return
	}

	h.logger.Info("Organization deactivated successfully",
		zap.String("org_id", orgID),
		zap.String("deactivated_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusOK, "organization deactivated successfully")
}

// GetOrganizationStats handles GET /organizations/:id/stats
// @Summary Get organization statistics
// @Description Retrieve statistics and metrics for an organization
// @Tags organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} organizationResponses.OrganizationStatsResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/organizations/{id}/stats [get]
func (h *Handler) GetOrganizationStats(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	stats, err := h.orgService.GetOrganizationStats(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to retrieve organization stats", zap.Error(err), zap.String("org_id", orgID))

		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendError(c, http.StatusInternalServerError, "failed to retrieve organization stats", err)
		}
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, stats)
}
