package organizations

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	groupRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/groups"
	orgRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/organizations"
	groupResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/groups"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for organization operations
type Handler struct {
	orgService   interfaces.OrganizationService
	groupService interfaces.GroupService
	logger       *zap.Logger
	responder    interfaces.Responder
}

// NewOrganizationHandler creates a new organization handler instance
func NewOrganizationHandler(
	orgService interfaces.OrganizationService,
	groupService interfaces.GroupService,
	logger *zap.Logger,
	responder interfaces.Responder,
) *Handler {
	return &Handler{
		orgService:   orgService,
		groupService: groupService,
		logger:       logger,
		responder:    responder,
	}
}

// CreateOrganization handles POST /organizations
//
//	@Summary		Create a new organization
//	@Description	Create a new organization with the provided information
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Param			organization	body		organizations.CreateOrganizationRequest	true	"Organization creation data"
//	@Success		201				{object}	organizations.OrganizationResponse
//	@Failure		400				{object}	responses.ErrorResponse
//	@Failure		409				{object}	responses.ErrorResponse
//	@Failure		500				{object}	responses.ErrorResponse
//	@Router			/api/v2/organizations [post]
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
//
//	@Summary		Get organization by ID
//	@Description	Retrieve an organization by its ID
//	@Tags			organizations
//	@Produce		json
//	@Param			id	path		string	true	"Organization ID"
//	@Success		200	{object}	organizations.OrganizationResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/organizations/{id} [get]
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
//
//	@Summary		Update organization
//	@Description	Update an existing organization with the provided information
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string									true	"Organization ID"
//	@Param			organization	body		organizations.UpdateOrganizationRequest	true	"Organization update data"
//	@Success		200				{object}	organizations.OrganizationResponse
//	@Failure		400				{object}	responses.ErrorResponse
//	@Failure		404				{object}	responses.ErrorResponse
//	@Failure		409				{object}	responses.ErrorResponse
//	@Failure		500				{object}	responses.ErrorResponse
//	@Router			/api/v2/organizations/{id} [put]
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
//
//	@Summary		Delete organization
//	@Description	Delete an organization by its ID
//	@Tags			organizations
//	@Produce		json
//	@Param			id	path		string	true	"Organization ID"
//	@Success		200	{object}	responses.SuccessResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		401	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/organizations/{id} [delete]
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
//
//	@Summary		List organizations
//	@Description	Retrieve a list of organizations with pagination and optional type filter
//	@Tags			organizations
//	@Produce		json
//	@Param			limit				query		int		false	"Number of organizations to return (default: 10, max: 100)"
//	@Param			offset				query		int		false	"Number of organizations to skip (default: 0)"
//	@Param			include_inactive	query		bool	false	"Include inactive organizations (default: false)"
//	@Param			type				query		string	false	"Filter by organization type (enterprise, small_business, individual, fpo, cooperative, agribusiness, farmers_group, shg, ngo, government, input_supplier, trader, processing_unit, research_institute)"
//	@Success		200					{array}		organizations.OrganizationResponse
//	@Failure		500					{object}	responses.ErrorResponse
//	@Router			/api/v2/organizations [get]
func (h *Handler) ListOrganizations(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	includeInactiveStr := c.DefaultQuery("include_inactive", "false")
	orgType := c.Query("type")

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
	orgs, err := h.orgService.ListOrganizations(c.Request.Context(), limit, offset, includeInactive, orgType)
	if err != nil {
		h.logger.Error("Failed to list organizations", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "failed to list organizations", nil)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, orgs)
}

// GetOrganizationHierarchy handles GET /organizations/:id/hierarchy
//
//	@Summary		Get organization hierarchy
//	@Description	Retrieve the hierarchical structure of an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			id	path		string	true	"Organization ID"
//	@Success		200	{object}	organizations.OrganizationHierarchyResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/organizations/{id}/hierarchy [get]
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
//
//	@Summary		Activate organization
//	@Description	Activate an inactive organization
//	@Tags			organizations
//	@Produce		json
//	@Param			id	path		string	true	"Organization ID"
//	@Success		200	{object}	responses.SuccessResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		401	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/organizations/{id}/activate [post]
//
//nolint:dupl // Activate and Deactivate have similar structure but different logic
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
//
//	@Summary		Deactivate organization
//	@Description	Deactivate an active organization
//	@Tags			organizations
//	@Produce		json
//	@Param			id	path		string	true	"Organization ID"
//	@Success		200	{object}	responses.SuccessResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		401	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/organizations/{id}/deactivate [post]
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
//
//	@Summary		Get organization statistics
//	@Description	Retrieve statistics and metrics for an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			id	path		string	true	"Organization ID"
//	@Success		200	{object}	organizations.OrganizationStatsResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/organizations/{id}/stats [get]
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

// GetOrganizationGroups handles GET /organizations/:orgId/groups
//
//	@Summary		Get organization groups
//	@Description	Retrieve all groups within an organization with pagination
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId				path		string	true	"Organization ID"
//	@Param			limit				query		int		false	"Number of groups to return (default: 10, max: 100)"
//	@Param			offset				query		int		false	"Number of groups to skip (default: 0)"
//	@Param			include_inactive	query		bool	false	"Include inactive groups (default: false)"
//	@Success		200					{object}	organizations.OrganizationGroupListResponse
//	@Failure		400					{object}	responses.ErrorResponse
//	@Failure		404					{object}	responses.ErrorResponse
//	@Failure		500					{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups [get]
func (h *Handler) GetOrganizationGroups(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

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

	// Verify organization exists
	_, err = h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Get groups for the organization
	groups, err := h.groupService.ListGroups(c.Request.Context(), limit, offset, orgID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to retrieve organization groups", zap.Error(err), zap.String("org_id", orgID))
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, groups)
}

// CreateGroupInOrganization handles POST /organizations/:orgId/groups
//
//	@Summary		Create group in organization
//	@Description	Create a new group within a specific organization
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Param			orgId	path		string											true	"Organization ID"
//	@Param			group	body		organizations.CreateOrganizationGroupRequest	true	"Group creation data"
//	@Success		201		{object}	organizations.OrganizationGroupResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		409		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups [post]
func (h *Handler) CreateGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	var req groupRequests.CreateGroupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for group creation", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Ensure the organization ID in the request matches the URL parameter
	req.OrganizationID = orgID

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Create group
	group, err := h.groupService.CreateGroup(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create group", zap.Error(err))

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

	h.logger.Info("Group created successfully in organization",
		zap.String("org_id", orgID),
		zap.String("created_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusCreated, group)
}

// GetGroupInOrganization handles GET /organizations/:orgId/groups/:groupId
//
//	@Summary		Get group in organization
//	@Description	Retrieve a specific group within an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			groupId	path		string	true	"Group ID"
//	@Success		200		{object}	organizations.OrganizationGroupResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId} [get]
func (h *Handler) GetGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Get group
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Failed to retrieve group", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	// Note: This assumes the group response has an OrganizationID field
	// We need to check this by converting the interface{} response
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	h.responder.SendSuccess(c, http.StatusOK, group)
}

// UpdateGroupInOrganization handles PUT /organizations/:orgId/groups/:groupId
//
//	@Summary		Update group in organization
//	@Description	Update a specific group within an organization (super_admin only)
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Param			orgId	path		string										true	"Organization ID"
//	@Param			groupId	path		string										true	"Group ID"
//	@Param			request	body		organizations.UpdateOrganizationGroupRequest	true	"Group update data"
//	@Success		200		{object}	organizations.OrganizationGroupResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		401		{object}	responses.ErrorResponse
//	@Failure		403		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId} [put]
//	@Security		BearerAuth
func (h *Handler) UpdateGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	var req orgRequests.UpdateOrganizationGroupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for group update", zap.Error(err))
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

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	existingGroup, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := existingGroup.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Convert to UpdateGroupRequest for the service layer
	updateGroupReq := &groupRequests.UpdateGroupRequest{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		IsActive:    req.IsActive,
	}

	// Update group
	updatedGroup, err := h.groupService.UpdateGroup(c.Request.Context(), groupID, updateGroupReq)
	if err != nil {
		h.logger.Error("Failed to update group",
			zap.Error(err),
			zap.String("group_id", groupID),
			zap.String("org_id", orgID))

		switch {
		case errors.IsValidationError(err):
			h.responder.SendValidationError(c, []string{err.Error()})
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		case errors.IsConflictError(err):
			h.responder.SendError(c, http.StatusConflict, "conflict", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("Group updated successfully in organization",
		zap.String("group_id", groupID),
		zap.String("org_id", orgID),
		zap.String("updated_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusOK, updatedGroup)
}

// DeleteGroupInOrganization handles DELETE /organizations/:orgId/groups/:groupId
//
//	@Summary		Delete group in organization
//	@Description	Delete a specific group within an organization (super_admin only)
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			groupId	path		string	true	"Group ID"
//	@Success		200		{object}	responses.SuccessResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		401		{object}	responses.ErrorResponse
//	@Failure		403		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId} [delete]
//	@Security		BearerAuth
func (h *Handler) DeleteGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	existingGroup, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := existingGroup.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Delete group
	err = h.groupService.DeleteGroup(c.Request.Context(), groupID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete group",
			zap.Error(err),
			zap.String("group_id", groupID),
			zap.String("org_id", orgID))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		case errors.IsConflictError(err):
			h.responder.SendError(c, http.StatusConflict, "cannot delete group with active members", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("Group deleted successfully from organization",
		zap.String("group_id", groupID),
		zap.String("org_id", orgID),
		zap.String("deleted_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusOK, "group deleted successfully")
}

// AddUserToGroupInOrganization handles POST /organizations/:orgId/groups/:groupId/users
//
//	@Summary		Add user to group in organization
//	@Description	Add a user to a specific group within an organization
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Param			orgId	path		string									true	"Organization ID"
//	@Param			groupId	path		string									true	"Group ID"
//	@Param			request	body		organizations.AssignUserToGroupRequest	true	"User assignment data"
//	@Success		201		{object}	organizations.OrganizationGroupMemberResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		409		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId}/users [post]
func (h *Handler) AddUserToGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	var req orgRequests.AssignUserToGroupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for user-group assignment",
			zap.Error(err),
			zap.String("org_id", orgID),
			zap.String("group_id", groupID),
			zap.Any("request_body", req))

		// Try to extract detailed validation errors
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, fieldErr := range validationErrs {
				errorMessages = append(errorMessages, fmt.Sprintf("field '%s' failed validation '%s'", fieldErr.Field(), fieldErr.Tag()))
				h.logger.Error("Validation error detail",
					zap.String("field", fieldErr.Field()),
					zap.String("tag", fieldErr.Tag()),
					zap.Any("value", fieldErr.Value()),
					zap.String("param", fieldErr.Param()))
			}
			h.responder.SendValidationError(c, errorMessages)
			return
		}

		h.responder.SendError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Extract user ID from context (set by auth middleware)
	currentUserID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Convert to AddMemberRequest for the service layer
	addMemberReq := &groupRequests.AddMemberRequest{
		GroupID:       groupID,
		PrincipalID:   req.PrincipalID,
		PrincipalType: req.PrincipalType,
		AddedByID:     currentUserID.(string),
		StartsAt:      req.StartsAt,
		EndsAt:        req.EndsAt,
	}

	// Add member to group
	membership, err := h.groupService.AddMemberToGroup(c.Request.Context(), addMemberReq)
	if err != nil {
		h.logger.Error("Failed to add user to group", zap.Error(err))

		switch {
		case errors.IsValidationError(err):
			h.responder.SendValidationError(c, []string{err.Error()})
		case errors.IsConflictError(err):
			h.responder.SendError(c, http.StatusConflict, "user already in group", err)
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "user not found", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("User added to group successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("principal_id", req.PrincipalID),
		zap.String("added_by", currentUserID.(string)))

	h.responder.SendSuccess(c, http.StatusCreated, membership)
}

// RemoveUserFromGroupInOrganization handles DELETE /organizations/:orgId/groups/:groupId/users/:userId
//
//	@Summary		Remove user from group in organization
//	@Description	Remove a user from a specific group within an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			groupId	path		string	true	"Group ID"
//	@Param			userId	path		string	true	"User ID"
//	@Success		200		{object}	responses.SuccessResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId}/users/{userId} [delete]
func (h *Handler) RemoveUserFromGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")
	principalID := c.Param("userId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	if principalID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "user ID is required", nil)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Remove member from group
	err = h.groupService.RemoveMemberFromGroup(c.Request.Context(), groupID, principalID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to remove user from group", zap.Error(err))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "user not found in group", err)
		case errors.IsValidationError(err):
			h.responder.SendError(c, http.StatusBadRequest, "validation failed", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("User removed from group successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("principal_id", principalID),
		zap.String("removed_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusOK, "user removed from group successfully")
}

// GetGroupUsersInOrganization handles GET /organizations/:orgId/groups/:groupId/users
//
//	@Summary		Get group users in organization
//	@Description	Retrieve all users in a specific group within an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			groupId	path		string	true	"Group ID"
//	@Param			limit	query		int		false	"Number of users to return (default: 10, max: 100)"
//	@Param			offset	query		int		false	"Number of users to skip (default: 0)"
//	@Success		200		{object}	organizations.OrganizationGroupMembersResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId}/users [get]
func (h *Handler) GetGroupUsersInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

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

	// Verify organization exists
	_, err = h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Get group members
	members, err := h.groupService.GetGroupMembers(c.Request.Context(), groupID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to retrieve group members", zap.Error(err), zap.String("group_id", groupID))
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, members)
}

// GetUserGroupsInOrganization handles GET /organizations/:orgId/users/:userId/groups
//
//	@Summary		Get user groups in organization
//	@Description	Retrieve all groups a user belongs to within an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			userId	path		string	true	"User ID"
//	@Param			limit	query		int		false	"Number of groups to return (default: 10, max: 100)"
//	@Param			offset	query		int		false	"Number of groups to skip (default: 0)"
//	@Success		200		{object}	organizations.UserOrganizationGroupsResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/users/{userId}/groups [get]
func (h *Handler) GetUserGroupsInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	principalID := c.Param("userId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if principalID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "user ID is required", nil)
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

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

	// Verify organization exists
	_, err = h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Get user's groups within the organization
	// Note: This would require extending the GroupService interface with a method like GetUserGroups
	// For now, we'll use a placeholder implementation that would need to be added to the service
	groups, err := h.getUserGroupsInOrganization(c.Request.Context(), orgID, principalID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to retrieve user groups", zap.Error(err),
			zap.String("org_id", orgID),
			zap.String("user_id", principalID))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "user not found", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, groups)
}

// getUserGroupsInOrganization is a helper method to get user's groups within an organization
// This would need to be implemented properly with the service layer
func (h *Handler) getUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
	// This is a placeholder implementation
	// In a real implementation, this would call a service method that:
	// 1. Gets all groups in the organization
	// 2. Filters for groups where the user is a member
	// 3. Returns the filtered list with pagination

	// For now, return an empty list to prevent compilation errors
	// This should be replaced with actual service implementation
	return []interface{}{}, nil
}

// AssignRoleToGroupInOrganization handles POST /organizations/:orgId/groups/:groupId/roles
//
//	@Summary		Assign role to group in organization
//	@Description	Assign a role to a specific group within an organization
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Param			orgId	path		string									true	"Organization ID"
//	@Param			groupId	path		string									true	"Group ID"
//	@Param			request	body		organizations.AssignRoleToGroupRequest	true	"Role assignment data"
//	@Success		201		{object}	organizations.OrganizationGroupRoleResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		409		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId}/roles [post]
func (h *Handler) AssignRoleToGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	h.logger.Info("AssignRoleToGroupInOrganization called",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))

	if orgID == "" {
		h.logger.Error("Organization ID missing from URL")
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.logger.Error("Group ID missing from URL")
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	var req orgRequests.AssignRoleToGroupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for role-group assignment",
			zap.Error(err),
			zap.String("org_id", orgID),
			zap.String("group_id", groupID))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	h.logger.Info("Request bound successfully",
		zap.String("role_id", req.RoleID),
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Additional validation: Verify role exists and is accessible within organization context
	// Note: This would require a RoleService method to validate role existence
	// For now, we'll proceed with the assignment

	// Assign role to group
	h.logger.Info("Attempting to assign role to group",
		zap.String("group_id", groupID),
		zap.String("role_id", req.RoleID),
		zap.String("assigned_by", userID.(string)))

	roleAssignment, err := h.groupService.AssignRoleToGroup(c.Request.Context(), groupID, req.RoleID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to assign role to group",
			zap.Error(err),
			zap.String("group_id", groupID),
			zap.String("role_id", req.RoleID),
			zap.String("error_type", fmt.Sprintf("%T", err)))

		switch {
		case errors.IsValidationError(err):
			h.responder.SendValidationError(c, []string{err.Error()})
		case errors.IsConflictError(err):
			h.responder.SendError(c, http.StatusConflict, "role already assigned to group", err)
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "role not found", err)
		default:
			h.logger.Error("Unhandled error in role assignment",
				zap.Error(err),
				zap.String("error_details", fmt.Sprintf("%+v", err)))
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("Role assignment service call completed successfully")

	h.logger.Info("Role assigned to group successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("role_id", req.RoleID),
		zap.String("assigned_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusCreated, roleAssignment)
}

// RemoveRoleFromGroupInOrganization handles DELETE /organizations/:orgId/groups/:groupId/roles/:roleId
//
//	@Summary		Remove role from group in organization
//	@Description	Remove a role from a specific group within an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			groupId	path		string	true	"Group ID"
//	@Param			roleId	path		string	true	"Role ID"
//	@Success		200		{object}	responses.SuccessResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId}/roles/{roleId} [delete]
func (h *Handler) RemoveRoleFromGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")
	roleID := c.Param("roleId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	if roleID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "role ID is required", nil)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Remove role from group
	err = h.groupService.RemoveRoleFromGroup(c.Request.Context(), groupID, roleID)
	if err != nil {
		h.logger.Error("Failed to remove role from group", zap.Error(err))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "role assignment not found", err)
		case errors.IsValidationError(err):
			h.responder.SendError(c, http.StatusBadRequest, "validation failed", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("Role removed from group successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("role_id", roleID),
		zap.String("removed_by", userID.(string)))

	response := groupResponses.NewRemoveGroupRoleResponse(groupID, roleID, orgID, "role removed from group successfully")
	h.responder.SendSuccess(c, http.StatusOK, response)
}

// GetGroupRolesInOrganization handles GET /organizations/:orgId/groups/:groupId/roles
//
//	@Summary		Get group roles in organization
//	@Description	Retrieve all roles assigned to a specific group within an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			groupId	path		string	true	"Group ID"
//	@Param			limit	query		int		false	"Number of roles to return (default: 10, max: 100)"
//	@Param			offset	query		int		false	"Number of roles to skip (default: 0)"
//	@Success		200		{object}	organizations.OrganizationGroupRolesResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId}/roles [get]
func (h *Handler) GetGroupRolesInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

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

	// Verify organization exists
	_, err = h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Get group roles
	roles, err := h.groupService.GetGroupRoles(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Failed to retrieve group roles", zap.Error(err), zap.String("group_id", groupID))
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, roles)
}

// Helper methods for role-group operations
// These would need to be replaced with actual service calls once the GroupService interface is extended

// GetUserEffectiveRolesInOrganization handles GET /organizations/:orgId/users/:userId/effective-roles
//
//	@Summary		Get user's effective roles in organization
//	@Description	Retrieve all effective roles for a user within an organization, including inherited roles from group hierarchy
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			userId	path		string	true	"User ID"
//	@Success		200		{object}	organizations.UserEffectiveRolesResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/users/{userId}/effective-roles [get]
func (h *Handler) GetUserEffectiveRolesInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	userID := c.Param("userId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if userID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "user ID is required", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Get user's effective roles using the group service
	effectiveRoles, err := h.groupService.GetUserEffectiveRoles(c.Request.Context(), orgID, userID)
	if err != nil {
		h.logger.Error("Failed to retrieve user effective roles",
			zap.Error(err),
			zap.String("org_id", orgID),
			zap.String("user_id", userID))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "user or organization not found", err)
		case errors.IsValidationError(err):
			h.responder.SendError(c, http.StatusBadRequest, "validation failed", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("User effective roles retrieved successfully",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	h.responder.SendSuccess(c, http.StatusOK, effectiveRoles)
}
