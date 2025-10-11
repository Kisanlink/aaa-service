package permissions

import (
	"net/http"

	reqPermissions "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/permissions"
	respPermissions "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/permissions"
	permissionService "github.com/Kisanlink/aaa-service/v2/internal/services/permissions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetPermission handles GET /api/v2/permissions/:id
//
//	@Summary		Get permission by ID
//	@Description	Retrieve a permission by its unique identifier
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Permission ID"
//	@Success		200	{object}	respPermissions.PermissionResponse
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/api/v2/permissions/{id} [get]
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	permissionID := c.Param("id")
	h.logger.Info("Getting permission by ID", zap.String("permissionID", permissionID))

	if permissionID == "" {
		h.responder.SendValidationError(c, []string{"permission ID is required"})
		return
	}

	// Get permission through service
	permission, err := h.permissionService.GetPermissionByID(c.Request.Context(), permissionID)
	if err != nil {
		h.logger.Error("Failed to get permission", zap.Error(err), zap.String("permissionID", permissionID))
		h.responder.SendError(c, http.StatusNotFound, "Permission not found", err)
		return
	}

	// Convert to response
	response := respPermissions.NewPermissionResponse(permission)

	h.logger.Info("Permission retrieved successfully", zap.String("permissionID", permissionID))
	h.responder.SendSuccess(c, http.StatusOK, response)
}

// ListPermissions handles GET /api/v2/permissions
//
//	@Summary		List permissions
//	@Description	Get a paginated list of permissions with optional filters
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			role_id		query		string	false	"Role ID filter"
//	@Param			resource_id	query		string	false	"Resource ID filter"
//	@Param			action_id	query		string	false	"Action ID filter"
//	@Param			is_active	query		boolean	false	"Active status filter"
//	@Param			search		query		string	false	"Search term"
//	@Param			limit		query		int		false	"Number of permissions to return"	default(10)
//	@Param			offset		query		int		false	"Number of permissions to skip"		default(0)
//	@Success		200			{object}	respPermissions.PermissionListResponse
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/api/v2/permissions [get]
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	h.logger.Info("Listing permissions")

	// Parse query parameters
	var req reqPermissions.QueryPermissionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Failed to bind query parameters", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Query validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Build filter
	filter := &permissionService.PermissionFilter{
		ResourceID: req.ResourceID,
		ActionID:   req.ActionID,
		IsActive:   req.IsActive,
		Limit:      req.GetLimit(),
		Offset:     req.GetOffset(),
	}

	// Get permissions from service
	permissions, err := h.permissionService.ListPermissions(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list permissions", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to retrieve permissions", err)
		return
	}

	// Calculate total count (simplified - in production, service should provide this)
	total := len(permissions)

	// Calculate page number
	page := (req.GetOffset() / req.GetLimit()) + 1

	// Create response
	response := respPermissions.NewPermissionListResponse(
		permissions,
		page,
		req.GetLimit(),
		total,
		h.getRequestID(c),
	)

	h.logger.Info("Permissions listed successfully",
		zap.Int("count", len(permissions)),
		zap.Int("total", total))

	c.JSON(http.StatusOK, response)
}

// GetRolePermissions handles GET /api/v2/roles/:id/permissions
//
//	@Summary		Get role permissions
//	@Description	Get all permissions assigned to a specific role
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Role ID"
//	@Success		200	{object}	respPermissions.PermissionListResponse
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/api/v2/roles/{id}/permissions [get]
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	h.logger.Info("Getting permissions for role", zap.String("roleID", roleID))

	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}

	// Get permissions for role
	permissions, err := h.permissionService.GetPermissionsForRole(c.Request.Context(), roleID)
	if err != nil {
		h.logger.Error("Failed to get role permissions", zap.Error(err), zap.String("roleID", roleID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to get role permissions", err)
		return
	}

	// Create response
	response := respPermissions.NewPermissionListResponse(
		permissions,
		1,
		len(permissions),
		len(permissions),
		h.getRequestID(c),
	)

	h.logger.Info("Role permissions retrieved successfully",
		zap.String("roleID", roleID),
		zap.Int("count", len(permissions)))

	c.JSON(http.StatusOK, response)
}
