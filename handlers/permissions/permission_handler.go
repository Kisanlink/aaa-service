package permissions

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PermissionHandler handles permission-related HTTP requests
type PermissionHandler struct {
	validator interfaces.Validator
	responder interfaces.Responder
	logger    *zap.Logger
}

// NewPermissionHandler creates a new PermissionHandler instance
func NewPermissionHandler(
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *PermissionHandler {
	return &PermissionHandler{
		validator: validator,
		responder: responder,
		logger:    logger,
	}
}

// CreatePermissionRequest represents a request to create a permission
type CreatePermissionRequest struct {
	Resource string   `json:"resource" validate:"required"`
	Effect   string   `json:"effect" validate:"required"`
	Actions  []string `json:"actions" validate:"required"`
}

// Validate validates the CreatePermissionRequest
func (r *CreatePermissionRequest) Validate() error {
	if r.Resource == "" {
		return errors.NewValidationError("resource is required")
	}
	if r.Effect == "" {
		return errors.NewValidationError("effect is required")
	}
	if len(r.Actions) == 0 {
		return errors.NewValidationError("at least one action is required")
	}
	return nil
}

// UpdatePermissionRequest represents a request to update a permission
type UpdatePermissionRequest struct {
	ID       string   `json:"id" validate:"required"`
	Resource *string  `json:"resource,omitempty"`
	Effect   *string  `json:"effect,omitempty"`
	Actions  []string `json:"actions,omitempty"`
}

// Validate validates the UpdatePermissionRequest
func (r *UpdatePermissionRequest) Validate() error {
	if r.ID == "" {
		return errors.NewValidationError("permission ID is required")
	}
	return nil
}

// CreatePermissionV2 handles POST /v2/permissions
func (h *PermissionHandler) CreatePermissionV2(c *gin.Context) {
	h.logger.Info("Creating permission")

	var req CreatePermissionRequest
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
	permission := models.NewPermission(req.Resource, req.Effect, req.Actions)

	// TODO: Create permission through service when PermissionService is available
	// For now, return mock response
	result := map[string]interface{}{
		"id":       permission.ID,
		"resource": permission.Resource,
		"effect":   permission.Effect,
		"actions":  permission.Actions,
		"message":  "Permission created successfully",
	}

	h.logger.Info("Permission created successfully", zap.String("permissionID", permission.ID))
	h.responder.SendSuccess(c, http.StatusCreated, result)
}

// GetPermissionV2 handles GET /v2/permissions/:id
func (h *PermissionHandler) GetPermissionV2(c *gin.Context) {
	permissionID := c.Param("id")
	h.logger.Info("Getting permission by ID", zap.String("permissionID", permissionID))

	if permissionID == "" {
		h.responder.SendValidationError(c, []string{"permission ID is required"})
		return
	}

	// TODO: Get permission through service when PermissionService is available
	// For now, return mock response
	result := map[string]interface{}{
		"id":       permissionID,
		"resource": "example_resource",
		"effect":   "allow",
		"actions":  []string{"read", "write"},
		"message":  "Permission retrieved successfully",
	}

	h.logger.Info("Permission retrieved successfully", zap.String("permissionID", permissionID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// UpdatePermissionV2 handles PUT /v2/permissions/:id
func (h *PermissionHandler) UpdatePermissionV2(c *gin.Context) {
	permissionID := c.Param("id")
	h.logger.Info("Updating permission", zap.String("permissionID", permissionID))

	if permissionID == "" {
		h.responder.SendValidationError(c, []string{"permission ID is required"})
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Set permission ID from URL parameter
	req.ID = permissionID

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// TODO: Update permission through service when PermissionService is available
	// For now, return mock response
	result := map[string]interface{}{
		"id":      permissionID,
		"message": "Permission updated successfully",
	}

	h.logger.Info("Permission updated successfully", zap.String("permissionID", permissionID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// DeletePermissionV2 handles DELETE /v2/permissions/:id
func (h *PermissionHandler) DeletePermissionV2(c *gin.Context) {
	permissionID := c.Param("id")
	h.logger.Info("Deleting permission", zap.String("permissionID", permissionID))

	if permissionID == "" {
		h.responder.SendValidationError(c, []string{"permission ID is required"})
		return
	}

	// TODO: Delete permission through service when PermissionService is available
	// For now, return mock response
	result := map[string]interface{}{
		"message": "Permission deleted successfully",
	}

	h.logger.Info("Permission deleted successfully", zap.String("permissionID", permissionID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// ListPermissionsV2 handles GET /v2/permissions
func (h *PermissionHandler) ListPermissionsV2(c *gin.Context) {
	h.logger.Info("Listing permissions")

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

	// TODO: List permissions through service when PermissionService is available
	// For now, return mock response
	result := map[string]interface{}{
		"permissions": []interface{}{},
		"total":       0,
		"limit":       limit,
		"offset":      offset,
		"message":     "Permissions listed successfully",
	}

	h.logger.Info("Permissions listed successfully")
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// EvaluatePermissionV2 handles POST /v2/permissions/evaluate
func (h *PermissionHandler) EvaluatePermissionV2(c *gin.Context) {
	h.logger.Info("Evaluating permission")

	var req struct {
		UserID   string `json:"user_id" validate:"required"`
		Resource string `json:"resource" validate:"required"`
		Action   string `json:"action" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Basic validation
	if req.UserID == "" || req.Resource == "" || req.Action == "" {
		h.responder.SendValidationError(c, []string{"user_id, resource, and action are required"})
		return
	}

	// TODO: Evaluate permission through service when PermissionService is available
	// For now, return mock response
	result := map[string]interface{}{
		"allowed":  true, // Mock allowed response
		"user_id":  req.UserID,
		"resource": req.Resource,
		"action":   req.Action,
		"message":  "Permission evaluated successfully",
	}

	h.logger.Info("Permission evaluated successfully",
		zap.String("userID", req.UserID),
		zap.String("resource", req.Resource),
		zap.String("action", req.Action))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// GrantTemporaryPermissionV2 handles POST /v2/permissions/temporary
func (h *PermissionHandler) GrantTemporaryPermissionV2(c *gin.Context) {
	h.logger.Info("Granting temporary permission")

	var req struct {
		UserID    string   `json:"user_id" validate:"required"`
		Resource  string   `json:"resource" validate:"required"`
		Actions   []string `json:"actions" validate:"required"`
		ExpiresAt string   `json:"expires_at" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Basic validation
	if req.UserID == "" || req.Resource == "" || len(req.Actions) == 0 || req.ExpiresAt == "" {
		h.responder.SendValidationError(c, []string{"user_id, resource, actions, and expires_at are required"})
		return
	}

	// TODO: Grant temporary permission through service when PermissionService is available
	// For now, return mock response
	result := map[string]interface{}{
		"permission_id": "temp_perm_" + req.UserID + "_" + req.Resource,
		"user_id":       req.UserID,
		"resource":      req.Resource,
		"actions":       req.Actions,
		"expires_at":    req.ExpiresAt,
		"message":       "Temporary permission granted successfully",
	}

	h.logger.Info("Temporary permission granted successfully",
		zap.String("userID", req.UserID),
		zap.String("resource", req.Resource))
	h.responder.SendSuccess(c, http.StatusCreated, result)
}
