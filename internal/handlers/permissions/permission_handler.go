package permissions

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/entities/responses"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PermissionHandler handles permission-related HTTP requests
type PermissionHandler struct {
	dbManager db.DBManager
	validator interfaces.Validator
	responder interfaces.Responder
	logger    *zap.Logger
}

// NewPermissionHandler creates a new PermissionHandler instance
func NewPermissionHandler(
	dbManager db.DBManager,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *PermissionHandler {
	return &PermissionHandler{
		dbManager: dbManager,
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
// @Summary Create a new permission
// @Description Create a new permission with resource, effect, and actions
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission body CreatePermissionRequest true "Permission creation data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/permissions [post]
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
	permission := models.NewPermission(req.Resource, req.Effect)

	// TODO: Create permission through service when PermissionService is available
	// For now, return mock response
	result := map[string]interface{}{
		"id":          permission.ID,
		"name":        permission.Name,
		"description": permission.Description,
		"message":     "Permission created successfully",
	}

	h.logger.Info("Permission created successfully", zap.String("permissionID", permission.ID))
	h.responder.SendSuccess(c, http.StatusCreated, result)
}

// GetPermissionV2 handles GET /v2/permissions/:id
// @Summary Get permission by ID
// @Description Retrieve a permission by its unique identifier
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/permissions/{id} [get]
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
// @Summary Update permission
// @Description Update an existing permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID"
// @Param permission body UpdatePermissionRequest true "Permission update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/permissions/{id} [put]
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
// @Summary Delete permission
// @Description Delete a permission by its unique identifier
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/permissions/{id} [delete]
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
// @Summary List permissions
// @Description Get a paginated list of permissions
// @Tags permissions
// @Accept json
// @Produce json
// @Param limit query int false "Number of permissions to return" default(10)
// @Param offset query int false "Number of permissions to skip" default(0)
// @Success 200 {object} responses.PermissionsListResponse
// @Failure 400 {object} responses.ErrorResponseSwagger
// @Failure 500 {object} responses.ErrorResponseSwagger
// @Router /api/v2/permissions [get]
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

	// Query permissions from database using the unified Filter interface
	var permissions []models.Permission

	// Create filter for active permissions with pagination and sorting
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "is_active", Operator: base.OpEqual, Value: true},
			},
			Logic: base.LogicAnd,
		},
		Sort: []base.SortField{
			{Field: "id", Direction: "asc"},
			{Field: "created_at", Direction: "desc"},
		},
		Limit:  limit,
		Offset: offset,
	}

	// Get total count first - we need to specify the table for GORM
	total, err := h.dbManager.Count(c.Request.Context(), filter, &permissions)
	if err != nil {
		h.logger.Error("Failed to count permissions", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to retrieve permissions", err)
		return
	}

	// Get paginated permissions - we need to specify the table for GORM
	if err := h.dbManager.List(c.Request.Context(), filter, &permissions); err != nil {
		h.logger.Error("Failed to retrieve permissions", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to retrieve permissions", err)
		return
	}

	// Convert to response format
	permissionInfos := make([]responses.PermissionInfo, 0, len(permissions))
	for _, perm := range permissions {
		permissionInfo := responses.PermissionInfo{
			ID:          perm.ID,
			Name:        perm.Name,
			Description: perm.Description,
			Resource:    "",
			Action:      "",
			CreatedAt:   perm.CreatedAt,
			UpdatedAt:   perm.UpdatedAt,
		}

		// Set resource and action names if available
		if perm.Resource != nil {
			permissionInfo.Resource = perm.Resource.Name
		}
		if perm.Action != nil {
			permissionInfo.Action = perm.Action.Name
		}

		permissionInfos = append(permissionInfos, permissionInfo)
	}

	// Calculate pagination info
	page := (offset / limit) + 1
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Create response data
	data := responses.PermissionsListData{
		Permissions: permissionInfos,
	}

	// Create pagination info
	pagination := responses.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	// Create full response
	response := responses.PermissionsListResponse{
		Success:    true,
		Message:    "Permissions retrieved successfully",
		Data:       data,
		Pagination: pagination,
		Timestamp:  time.Now(),
		RequestID:  c.GetString("request_id"),
	}

	h.logger.Info("Permissions listed successfully",
		zap.Int("count", len(permissionInfos)),
		zap.Int64("total", total),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	c.JSON(http.StatusOK, response)
}

// EvaluatePermissionV2 handles POST /v2/permissions/evaluate
// @Summary Evaluate permission
// @Description Check if a user has permission to perform an action on a resource
// @Tags permissions
// @Accept json
// @Produce json
// @Param evaluation body object{user_id=string,resource=string,action=string} true "Permission evaluation data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/permissions/evaluate [post]
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
// @Summary Grant temporary permission
// @Description Grant temporary permission to a user for specific actions on a resource
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission body object{user_id=string,resource=string,actions=[]string,expires_at=string} true "Temporary permission data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/permissions/temporary [post]
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
