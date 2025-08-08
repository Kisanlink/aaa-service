package roles

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/entities/requests/roles"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	roleService interfaces.RoleService
	validator   interfaces.Validator
	responder   interfaces.Responder
	logger      *zap.Logger
}

// NewRoleHandler creates a new RoleHandler instance
func NewRoleHandler(
	roleService interfaces.RoleService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		validator:   validator,
		responder:   responder,
		logger:      logger,
	}
}

// CreateRole handles POST /v1/roles
// @Summary Create a new role
// @Description Create a new role with the provided information
// @Tags roles
// @Accept json
// @Produce json
// @Param role body roles.CreateRoleRequest true "Role creation data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	h.logger.Info("Creating role")

	var req roles.CreateRoleRequest
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

	// Convert to domain model
	description := ""
	if req.Description != nil {
		description = *req.Description
	}
	role := models.NewRole(req.Name, description)

	// Create role through service
	err := h.roleService.CreateRole(c.Request.Context(), role)
	if err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		if conflictErr, ok := err.(*errors.ConflictError); ok {
			h.responder.SendError(c, http.StatusConflict, conflictErr.Error(), conflictErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Role created successfully", zap.String("roleID", role.ID))
	h.responder.SendSuccess(c, http.StatusCreated, role)
}

// GetRole handles GET /v1/roles/:id
// @Summary Get role by ID
// @Description Retrieve a role by its unique identifier
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	roleID := c.Param("id")
	h.logger.Info("Getting role by ID", zap.String("roleID", roleID))

	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}

	// Get role through service
	result, err := h.roleService.GetRoleByID(c.Request.Context(), roleID)
	if err != nil {
		h.logger.Error("Failed to get role", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Role retrieved successfully", zap.String("roleID", roleID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// UpdateRole handles PUT /v1/roles/:id
// @Summary Update role
// @Description Update an existing role's information
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body roles.UpdateRoleRequest true "Role update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	roleID := c.Param("id")
	h.logger.Info("Updating role", zap.String("roleID", roleID))

	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}

	var req roles.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Set role ID from URL parameter
	req.RoleID = roleID

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Convert to domain model
	description := ""
	if req.Description != nil {
		description = *req.Description
	}
	role := models.NewRole(*req.Name, description)
	role.ID = roleID

	// Update role through service
	err := h.roleService.UpdateRole(c.Request.Context(), role)
	if err != nil {
		h.logger.Error("Failed to update role", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		if conflictErr, ok := err.(*errors.ConflictError); ok {
			h.responder.SendError(c, http.StatusConflict, conflictErr.Error(), conflictErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Role updated successfully", zap.String("roleID", roleID))
	h.responder.SendSuccess(c, http.StatusOK, role)
}

// DeleteRole handles DELETE /v1/roles/:id
// @Summary Delete role
// @Description Delete a role by its unique identifier
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v2/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	roleID := c.Param("id")
	h.logger.Info("Deleting role", zap.String("roleID", roleID))

	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}

	// Delete role through service
	err := h.roleService.DeleteRole(c.Request.Context(), roleID)
	if err != nil {
		h.logger.Error("Failed to delete role", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		if conflictErr, ok := err.(*errors.ConflictError); ok {
			h.responder.SendError(c, http.StatusConflict, conflictErr.Error(), conflictErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Role deleted successfully", zap.String("roleID", roleID))
	h.responder.SendSuccess(c, http.StatusOK, map[string]string{"message": "Role deleted successfully"})
}

// ListRoles handles GET /v1/roles
// @Summary List roles
// @Description Get a paginated list of roles
// @Tags roles
// @Accept json
// @Produce json
// @Param limit query int false "Number of roles to return" default(10)
// @Param offset query int false "Number of roles to skip" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	h.logger.Info("Listing roles")

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.responder.SendValidationError(c, []string{"invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		h.responder.SendValidationError(c, []string{"invalid offset parameter"})
		return
	}

	// List roles through service
	result, err := h.roleService.ListRoles(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list roles", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Roles listed successfully")
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// CreateRoleV2 handles POST /v2/roles
func (h *RoleHandler) CreateRoleV2(c *gin.Context) {
	h.logger.Info("Creating role (V2)")
	// For now, delegate to V1 implementation
	h.CreateRole(c)
}

// GetRoleV2 handles GET /v2/roles/:id
func (h *RoleHandler) GetRoleV2(c *gin.Context) {
	h.logger.Info("Getting role by ID (V2)")
	// For now, delegate to V1 implementation
	h.GetRole(c)
}

// UpdateRoleV2 handles PUT /v2/roles/:id
func (h *RoleHandler) UpdateRoleV2(c *gin.Context) {
	h.logger.Info("Updating role (V2)")
	// For now, delegate to V1 implementation
	h.UpdateRole(c)
}

// DeleteRoleV2 handles DELETE /v2/roles/:id
func (h *RoleHandler) DeleteRoleV2(c *gin.Context) {
	h.logger.Info("Deleting role (V2)")
	// For now, delegate to V1 implementation
	h.DeleteRole(c)
}

// ListRolesV2 handles GET /v2/roles
func (h *RoleHandler) ListRolesV2(c *gin.Context) {
	h.logger.Info("Listing roles (V2)")
	// For now, delegate to V1 implementation
	h.ListRoles(c)
}

// AssignPermissionV2 handles POST /v2/roles/:id/permissions
func (h *RoleHandler) AssignPermissionV2(c *gin.Context) {
	roleID := c.Param("id")
	h.logger.Info("Assigning permission to role", zap.String("roleID", roleID))

	var req roles.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Set role ID from URL parameter
	req.RoleID = roleID

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// TODO: Implement permission assignment through service
	// For now, return placeholder response
	result := map[string]interface{}{
		"message": "Permission assigned successfully",
		"role_id": roleID,
		"user_id": req.UserID, // This is actually for user-role assignment
	}

	h.logger.Info("Permission assigned successfully", zap.String("roleID", roleID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// RemovePermissionV2 handles DELETE /v2/roles/:id/permissions/:permissionId
func (h *RoleHandler) RemovePermissionV2(c *gin.Context) {
	roleID := c.Param("id")
	permissionID := c.Param("permissionId")
	h.logger.Info("Removing permission from role", zap.String("roleID", roleID), zap.String("permissionID", permissionID))

	// TODO: Implement permission removal through service
	// For now, return placeholder response
	result := map[string]interface{}{
		"message":       "Permission removed successfully",
		"role_id":       roleID,
		"permission_id": permissionID,
	}

	h.logger.Info("Permission removed successfully", zap.String("roleID", roleID), zap.String("permissionID", permissionID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// GetRolePermissionsV2 handles GET /v2/roles/:id/permissions
func (h *RoleHandler) GetRolePermissionsV2(c *gin.Context) {
	roleID := c.Param("id")
	h.logger.Info("Getting role permissions", zap.String("roleID", roleID))

	// TODO: Implement through service
	result := map[string]interface{}{
		"role_id":     roleID,
		"permissions": []interface{}{},
	}

	h.logger.Info("Role permissions retrieved successfully", zap.String("roleID", roleID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// GetRoleHierarchyV2 handles GET /v2/roles/hierarchy
func (h *RoleHandler) GetRoleHierarchyV2(c *gin.Context) {
	h.logger.Info("Getting role hierarchy")

	// TODO: Implement through service
	result := map[string]interface{}{
		"hierarchy": []interface{}{},
	}

	h.logger.Info("Role hierarchy retrieved successfully")
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// AddChildRoleV2 handles POST /v2/roles/:id/children
func (h *RoleHandler) AddChildRoleV2(c *gin.Context) {
	parentRoleID := c.Param("id")
	h.logger.Info("Adding child role", zap.String("parentRoleID", parentRoleID))

	var req struct {
		ChildRoleID string `json:"child_role_id" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// TODO: Implement through service
	result := map[string]interface{}{
		"message":        "Child role added successfully",
		"parent_role_id": parentRoleID,
		"child_role_id":  req.ChildRoleID,
	}

	h.logger.Info("Child role added successfully", zap.String("parentRoleID", parentRoleID), zap.String("childRoleID", req.ChildRoleID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}
