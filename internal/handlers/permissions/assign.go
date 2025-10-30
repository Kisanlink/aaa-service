package permissions

import (
	"net/http"

	reqRoleAssignments "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/role_assignments"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AssignPermissionsToRole handles POST /api/v1/roles/:id/permissions
//
//	@Summary		Assign permissions to role
//	@Description	Assign one or more permissions to a role
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string											true	"Role ID"
//	@Param			request	body		reqRoleAssignments.AssignPermissionsToRoleRequest	true	"Assignment request"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v1/roles/{id}/permissions [post]
func (h *PermissionHandler) AssignPermissionsToRole(c *gin.Context) {
	roleID := c.Param("id")
	h.logger.Info("Assigning permissions to role", zap.String("roleID", roleID))

	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}

	var req reqRoleAssignments.AssignPermissionsToRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Get assigned_by from context (authenticated user)
	assignedBy := c.GetString("user_id")
	if assignedBy == "" {
		assignedBy = "system"
	}

	// Assign permissions through service
	if err := h.roleAssignmentService.AssignPermissionsToRole(c.Request.Context(), roleID, req.PermissionIDs, assignedBy); err != nil {
		h.logger.Error("Failed to assign permissions", zap.Error(err), zap.String("roleID", roleID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to assign permissions", err)
		return
	}

	h.logger.Info("Permissions assigned successfully",
		zap.String("roleID", roleID),
		zap.Int("count", len(req.PermissionIDs)))

	response := map[string]interface{}{
		"role_id":              roleID,
		"assigned_permissions": req.PermissionIDs,
		"count":                len(req.PermissionIDs),
		"message":              "Permissions assigned successfully",
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// RevokePermissionFromRole handles DELETE /api/v1/roles/:id/permissions/:permId
//
//	@Summary		Revoke permission from role
//	@Description	Revoke a specific permission from a role
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"Role ID"
//	@Param			permId	path		string	true	"Permission ID"
//	@Success		204		{object}	nil
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v1/roles/{id}/permissions/{permId} [delete]
func (h *PermissionHandler) RevokePermissionFromRole(c *gin.Context) {
	roleID := c.Param("id")
	permissionID := c.Param("permId")
	h.logger.Info("Revoking permission from role",
		zap.String("roleID", roleID),
		zap.String("permissionID", permissionID))

	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}
	if permissionID == "" {
		h.responder.SendValidationError(c, []string{"permission ID is required"})
		return
	}

	// Get revoked_by from context (authenticated user)
	revokedBy := c.GetString("user_id")
	if revokedBy == "" {
		revokedBy = "system"
	}

	// Revoke permission through service
	if err := h.roleAssignmentService.RevokePermissionFromRole(c.Request.Context(), roleID, permissionID, revokedBy); err != nil {
		h.logger.Error("Failed to revoke permission", zap.Error(err),
			zap.String("roleID", roleID),
			zap.String("permissionID", permissionID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to revoke permission", err)
		return
	}

	h.logger.Info("Permission revoked successfully",
		zap.String("roleID", roleID),
		zap.String("permissionID", permissionID))

	c.Status(http.StatusNoContent)
}

// AssignResourcesToRole handles POST /api/v1/roles/:id/resources
//
//	@Summary		Assign resources to role
//	@Description	Assign resource-action combinations to a role
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string											true	"Role ID"
//	@Param			request	body		reqRoleAssignments.AssignResourcesToRoleRequest	true	"Assignment request"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v1/roles/{id}/resources [post]
func (h *PermissionHandler) AssignResourcesToRole(c *gin.Context) {
	roleID := c.Param("id")
	h.logger.Info("Assigning resources to role", zap.String("roleID", roleID))

	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}

	var req reqRoleAssignments.AssignResourcesToRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Get assigned_by from context (authenticated user)
	assignedBy := c.GetString("user_id")
	if assignedBy == "" {
		assignedBy = "system"
	}

	// Convert request assignments to service type
	var assignments []interface{}
	for _, a := range req.Assignments {
		assignments = append(assignments, struct {
			ResourceType string
			ResourceID   string
			Actions      []string
		}{
			ResourceType: a.ResourceType,
			ResourceID:   a.ResourceID,
			Actions:      a.Actions,
		})
	}

	// Assign resources through service (using individual assignment calls)
	assignedCount := 0
	for _, assignment := range req.Assignments {
		for _, action := range assignment.Actions {
			err := h.roleAssignmentService.AssignResourceActionToRole(
				c.Request.Context(),
				roleID,
				assignment.ResourceType,
				assignment.ResourceID,
				action,
				assignedBy,
			)
			if err != nil {
				h.logger.Warn("Failed to assign resource action",
					zap.Error(err),
					zap.String("roleID", roleID),
					zap.String("resourceID", assignment.ResourceID),
					zap.String("action", action))
			} else {
				assignedCount++
			}
		}
	}

	h.logger.Info("Resources assigned successfully",
		zap.String("roleID", roleID),
		zap.Int("assigned", assignedCount))

	response := map[string]interface{}{
		"role_id":         roleID,
		"assigned_count":  assignedCount,
		"total_requested": len(req.Assignments),
		"message":         "Resources assigned successfully",
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// RevokeResourceFromRole handles DELETE /api/v1/roles/:id/resources/:resId
//
//	@Summary		Revoke resource from role
//	@Description	Revoke a specific resource from a role
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"Role ID"
//	@Param			resId	path		string	true	"Resource ID"
//	@Param			action	query		string	false	"Specific action to revoke"
//	@Success		204		{object}	nil
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v1/roles/{id}/resources/{resId} [delete]
func (h *PermissionHandler) RevokeResourceFromRole(c *gin.Context) {
	roleID := c.Param("id")
	resourceID := c.Param("resId")
	action := c.DefaultQuery("action", "")

	h.logger.Info("Revoking resource from role",
		zap.String("roleID", roleID),
		zap.String("resourceID", resourceID),
		zap.String("action", action))

	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}
	if resourceID == "" {
		h.responder.SendValidationError(c, []string{"resource ID is required"})
		return
	}

	// Get revoked_by from context (authenticated user)
	revokedBy := c.GetString("user_id")
	if revokedBy == "" {
		revokedBy = "system"
	}

	// TODO: Service needs RevokeResourceActionFromRole with resourceID param only
	// For now, return not implemented
	h.responder.SendError(c, http.StatusNotImplemented, "Resource revocation not yet implemented", nil)
}

// GetRoleResources handles GET /api/v1/roles/:id/resources
//
//	@Summary		Get role resources
//	@Description	Get all resource-action combinations assigned to a role
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Role ID"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/api/v1/roles/{id}/resources [get]
func (h *PermissionHandler) GetRoleResources(c *gin.Context) {
	roleID := c.Param("id")
	h.logger.Info("Getting resources for role", zap.String("roleID", roleID))

	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}

	// Get resources for role
	resources, err := h.roleAssignmentService.GetRoleResources(c.Request.Context(), roleID)
	if err != nil {
		h.logger.Error("Failed to get role resources", zap.Error(err), zap.String("roleID", roleID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to get role resources", err)
		return
	}

	// Convert to response
	response := map[string]interface{}{
		"role_id":   roleID,
		"resources": resources,
		"count":     len(resources),
	}

	h.logger.Info("Role resources retrieved successfully",
		zap.String("roleID", roleID),
		zap.Int("count", len(resources)))

	h.responder.SendSuccess(c, http.StatusOK, response)
}
