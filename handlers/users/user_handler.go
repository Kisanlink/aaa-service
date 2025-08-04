package users

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/entities/requests/users"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body users.CreateUserRequest true "User creation data"
// @Success 201 {object} responses.UserResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users [post]
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
// @Summary Get user by ID
// @Description Retrieve a user by their unique identifier
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} responses.UserResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/{id} [get]
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
// @Summary Update user
// @Description Update an existing user's information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body users.UpdateUserRequest true "User update data"
// @Success 200 {object} responses.UserResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/{id} [put]
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
// @Summary Delete user
// @Description Delete a user by their unique identifier
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	h.logger.Info("Deleting user", zap.String("userID", userID))

	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}

	// Delete user through service
	err := h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("User deleted successfully", zap.String("userID", userID))
	h.responder.SendSuccess(c, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// ListUsers handles GET /users
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Param limit query int false "Number of users to return" default(10)
// @Param offset query int false "Number of users to skip" default(0)
// @Success 200 {object} responses.PaginatedResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users [get]
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
// @Summary Search users
// @Description Search for users based on query parameters
// @Tags users
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Number of users to return" default(10)
// @Param offset query int false "Number of users to skip" default(0)
// @Success 200 {object} responses.PaginatedResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	h.logger.Info("Searching users", zap.String("query", query))

	if query == "" {
		h.responder.SendValidationError(c, []string{"search query is required"})
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
// @Summary Validate user
// @Description Validate a user account
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/{id}/validate [post]
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
// @Summary Assign role to user
// @Description Assign a role to a specific user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param roleId path string true "Role ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/{id}/roles/{roleId} [post]
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

// RemoveRole handles DELETE /users/:id/roles/:roleId
// @Summary Remove role from user
// @Description Remove a role from a specific user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param roleId path string true "Role ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/{id}/roles/{roleId} [delete]
func (h *UserHandler) RemoveRole(c *gin.Context) {
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
	h.responder.SendSuccess(c, http.StatusOK, map[string]string{"message": "Role removed successfully"})
}
