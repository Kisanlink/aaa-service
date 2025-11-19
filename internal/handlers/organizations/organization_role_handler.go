package organizations

import (
	"fmt"
	"net/http"
	"strconv"

	orgRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/organizations"
	groupResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/groups"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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
