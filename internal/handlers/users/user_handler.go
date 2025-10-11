package users

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests/users"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/responses"

	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// references to satisfy imports for Swagger comment parsing
var (
	_ responses.ErrorResponse
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService interfaces.UserService
	roleService interfaces.RoleService
	validator   interfaces.Validator
	responder   interfaces.Responder
	logger      *zap.Logger
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(
	userService interfaces.UserService,
	roleService interfaces.RoleService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		roleService: roleService,
		validator:   validator,
		responder:   responder,
		logger:      logger,
	}
}

// CreateUser handles POST /users
//
//	@Summary		Create a new user
//	@Description	Create a new user with the provided information
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			user	body		users.CreateUserRequest	true	"User creation data"
//	@Success		201		{object}	responses.UserDetailResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		409		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v2/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	h.logger.Info("Creating user")

	var req users.CreateUserRequest
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

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Struct validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Create user through service
	userResponse, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		if validationErr, ok := err.(*errors.ValidationError); ok {
			h.responder.SendValidationError(c, []string{validationErr.Error()})
			return
		}
		if conflictErr, ok := err.(*errors.ConflictError); ok {
			h.responder.SendError(c, http.StatusConflict, conflictErr.Error(), conflictErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("User created successfully", zap.String("userID", userResponse.ID))
	h.responder.SendSuccess(c, http.StatusCreated, userResponse)
}

// GetUserByID handles GET /users/:id
//
//	@Summary		Get user by ID
//	@Description	Retrieve a user by their unique identifier
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	responses.UserDetailResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	h.logger.Info("Getting user by ID", zap.String("userID", userID))

	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}

	// Get user through service
	userResponse, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("User retrieved successfully", zap.String("userID", userID))
	h.responder.SendSuccess(c, http.StatusOK, userResponse)
}

// UpdateUser handles PUT /users/:id
//
//	@Summary		Update user
//	@Description	Update an existing user's information
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"User ID"
//	@Param			user	body		users.UpdateUserRequest	true	"User update data"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v2/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	h.logger.Info("Updating user", zap.String("userID", userID))

	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}

	var req users.UpdateUserRequest
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

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Struct validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Update user through service
	// Note: Setting userID in context since current service implementation expects it there
	type userIDKey struct{}
	ctx := context.WithValue(c.Request.Context(), userIDKey{}, userID)
	userResponse, err := h.userService.UpdateUser(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		if validationErr, ok := err.(*errors.ValidationError); ok {
			h.responder.SendValidationError(c, []string{validationErr.Error()})
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("User updated successfully", zap.String("userID", userID))
	h.responder.SendSuccess(c, http.StatusOK, userResponse)
}

// DeleteUser handles DELETE /users/:id
//
//	@Summary		Delete user
//	@Description	Soft delete a user by their unique identifier with proper cascade handling
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		409	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	h.logger.Info("Deleting user with enhanced cascade handling", zap.String("userID", userID))

	// Validate user ID parameter
	if userID == "" {
		h.logger.Warn("Delete user request missing user ID")
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}

	// Validate user ID format (basic validation)
	if err := h.validator.ValidateUserID(userID); err != nil {
		h.logger.Warn("Invalid user ID format", zap.String("userID", userID), zap.Error(err))
		h.responder.SendValidationError(c, []string{"invalid user ID format"})
		return
	}

	// Get the actor (who is performing the deletion) from context
	// This could come from JWT claims or authentication middleware
	actorID := "system" // Default to system if no actor found
	if claims, exists := c.Get("user_claims"); exists {
		if userClaims, ok := claims.(map[string]interface{}); ok {
			if id, exists := userClaims["user_id"]; exists {
				if idStr, ok := id.(string); ok {
					actorID = idStr
				}
			}
		}
	}

	// Perform enhanced soft delete with transaction support
	err := h.userService.SoftDeleteUserWithCascade(c.Request.Context(), userID, actorID)
	if err != nil {
		h.logger.Error("Failed to delete user",
			zap.String("userID", userID),
			zap.String("actorID", actorID),
			zap.Error(err))

		// Handle specific error types with appropriate HTTP status codes
		switch e := err.(type) {
		case *errors.NotFoundError:
			h.responder.SendError(c, http.StatusNotFound, "User not found", e)
			return
		case *errors.ConflictError:
			h.responder.SendError(c, http.StatusConflict, "Cannot delete user due to constraints", e)
			return
		case *errors.ValidationError:
			h.responder.SendValidationError(c, []string{e.Error()})
			return
		case *errors.ForbiddenError:
			h.responder.SendError(c, http.StatusForbidden, "Insufficient permissions to delete user", e)
			return
		default:
			h.responder.SendInternalError(c, err)
			return
		}
	}

	h.logger.Info("User deleted successfully with cascade cleanup",
		zap.String("userID", userID),
		zap.String("actorID", actorID))

	// Return detailed success response
	response := map[string]interface{}{
		"message":    "User deleted successfully",
		"user_id":    userID,
		"deleted_by": actorID,
		"deleted_at": "now", // Could be actual timestamp if needed
		"type":       "soft_delete",
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// ListUsers handles GET /users
//
//	@Summary		List users
//	@Description	Get a paginated list of users
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int	false	"Number of users to return"	default(10)
//	@Param			offset	query		int	false	"Number of users to skip"	default(0)
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v2/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	h.logger.Info("Listing users")

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

	// Get users through service
	result, err := h.userService.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Users listed successfully")
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// SearchUsers handles GET /users/search
//
//	@Summary		Search users
//	@Description	Search for users based on query parameters
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			q		query		string	false	"Search query"
//	@Param			query	query		string	false	"Search query (alternative parameter)"
//	@Param			limit	query		int		false	"Number of users to return"	default(10)
//	@Param			offset	query		int		false	"Number of users to skip"	default(0)
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v2/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	// Accept both 'q' and 'query' parameters for flexibility
	query := c.Query("q")
	if query == "" {
		query = c.Query("query")
	}

	h.logger.Info("Searching users", zap.String("query", query))

	if query == "" {
		h.responder.SendValidationError(c, []string{"search query is required (use 'q' or 'query' parameter)"})
		return
	}

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

	// Search users through service
	result, err := h.userService.SearchUsers(c.Request.Context(), query, limit, offset)
	if err != nil {
		h.logger.Error("Failed to search users", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Users search completed", zap.String("query", query))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// ValidateUser handles POST /users/:id/validate
//
//	@Summary		Validate user
//	@Description	Validate a user account
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Failure		409	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/api/v2/users/{id}/validate [post]
func (h *UserHandler) ValidateUser(c *gin.Context) {
	userID := c.Param("id")
	h.logger.Info("Validating user", zap.String("userID", userID))

	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}

	// Validate user through service
	err := h.userService.ValidateUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to validate user", zap.Error(err))
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

	h.logger.Info("User validated successfully", zap.String("userID", userID))
	h.responder.SendSuccess(c, http.StatusOK, map[string]string{"message": "User validated successfully"})
}

// AssignRole handles POST /users/:id/roles/:roleId
//
//	@Summary		Assign role to user
//	@Description	Assign a role to a specific user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"User ID"
//	@Param			roleId	path		string	true	"Role ID"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		409		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v2/users/{id}/roles/{roleId} [post]
func (h *UserHandler) AssignRole(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("roleId")
	h.logger.Info("Assigning role to user", zap.String("userID", userID), zap.String("roleID", roleID))

	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}
	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}

	// Assign role through service
	err := h.roleService.AssignRoleToUser(c.Request.Context(), userID, roleID)
	if err != nil {
		h.logger.Error("Failed to assign role", zap.Error(err))
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

	h.logger.Info("Role assigned successfully", zap.String("userID", userID), zap.String("roleID", roleID))
	h.responder.SendSuccess(c, http.StatusOK, map[string]string{"message": "Role assigned successfully"})
}

// RemoveRole handles DELETE /users/:id/roles/:roleId/legacy (legacy endpoint)
//
//	@Summary		Remove role from user (legacy)
//	@Description	Remove a role from a specific user using legacy endpoint
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"User ID"
//	@Param			roleId	path		string	true	"Role ID"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v2/users/{id}/roles/{roleId}/legacy [delete]
func (h *UserHandler) RemoveRole(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("roleId")
	h.logger.Info("Removing role from user (legacy endpoint)", zap.String("userID", userID), zap.String("roleID", roleID))

	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}
	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}

	// Remove role through service
	err := h.roleService.RemoveRoleFromUser(c.Request.Context(), userID, roleID)
	if err != nil {
		h.logger.Error("Failed to remove role", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Role removed successfully", zap.String("userID", userID), zap.String("roleID", roleID))
	h.responder.SendSuccess(c, http.StatusOK, map[string]string{"message": "Role removed successfully"})
}

// GetUserRoles handles GET /users/:id/roles
//
//	@Summary		Get user roles
//	@Description	Get all roles assigned to a specific user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/users/{id}/roles [get]
func (h *UserHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	h.logger.Info("Getting user roles", zap.String("userID", userID))

	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}

	// Get user roles through service
	roles, err := h.roleService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user roles", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("User roles retrieved successfully", zap.String("userID", userID), zap.Int("roleCount", len(roles)))
	h.responder.SendSuccess(c, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"roles":   roles,
		"count":   len(roles),
	})
}

// AssignRoleToUser handles POST /users/:id/roles
//
//	@Summary		Assign role to user
//	@Description	Assign a role to a specific user using request body
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"User ID"
//	@Param			role	body		map[string]string	true	"Role assignment data"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		409		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v2/users/{id}/roles [post]
func (h *UserHandler) AssignRoleToUser(c *gin.Context) {
	userID := c.Param("id")
	h.logger.Info("Assigning role to user", zap.String("userID", userID))

	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}

	var req struct {
		RoleID string `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	if req.RoleID == "" {
		h.responder.SendValidationError(c, []string{"role_id is required"})
		return
	}

	// Assign role through service
	err := h.roleService.AssignRoleToUser(c.Request.Context(), userID, req.RoleID)
	if err != nil {
		h.logger.Error("Failed to assign role", zap.Error(err))
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

	h.logger.Info("Role assigned successfully", zap.String("userID", userID), zap.String("roleID", req.RoleID))
	h.responder.SendSuccess(c, http.StatusOK, map[string]interface{}{
		"message": "Role assigned successfully",
		"user_id": userID,
		"role_id": req.RoleID,
	})
}

// RemoveRoleFromUser handles DELETE /users/:id/roles/:roleId
//
//	@Summary		Remove role from user
//	@Description	Remove a role from a specific user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"User ID"
//	@Param			roleId	path		string	true	"Role ID"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v2/users/{id}/roles/{roleId} [delete]
func (h *UserHandler) RemoveRoleFromUser(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("roleId")
	h.logger.Info("Removing role from user", zap.String("userID", userID), zap.String("roleID", roleID))

	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}
	if roleID == "" {
		h.responder.SendValidationError(c, []string{"role ID is required"})
		return
	}

	// Remove role through service
	err := h.roleService.RemoveRoleFromUser(c.Request.Context(), userID, roleID)
	if err != nil {
		h.logger.Error("Failed to remove role", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Role removed successfully", zap.String("userID", userID), zap.String("roleID", roleID))
	h.responder.SendSuccess(c, http.StatusOK, map[string]interface{}{
		"message": "Role removed successfully",
		"user_id": userID,
		"role_id": roleID,
	})
}
