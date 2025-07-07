package users

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/entities/requests/users"
	"github.com/Kisanlink/aaa-service/entities/responses/users"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/Kisanlink/aaa-service/utils"
	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService *services.UserService
	validator   *utils.Validator
	responder   *utils.Responder
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService *services.UserService, validator *utils.Validator, responder *utils.Responder) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator,
		responder:   responder,
	}
}

// CreateUser handles user creation requests
// @Summary Create a new user
// @Description Creates a new user account with the provided details
// @Tags Users
// @Accept json
// @Produce json
// @Param name body string true "User's full name"
// @Param email body string true "User's email address"
// @Param phone body string false "User's phone number"
// @Param status body string true "User's status (active/inactive)" Enums(active,inactive)
// @Success 201 {object} users.UserResponse "User created successfully" {id: string, name: string, email: string, phone: string, status: string, created_at: string, updated_at: string}
// @Success 200 {object} users.UserResponse "User already exists but was updated"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body or validation failed" {code: string, message: string, details: string}
// @Failure 409 {object} utils.ErrorResponse "User already exists and cannot be updated" {code: string, message: string, details: string}
// @Failure 422 {object} utils.ErrorResponse "Validation error" {code: string, message: string, details: []ValidationError}
// @Failure 500 {object} utils.ErrorResponse "Internal server error" {code: string, message: string}
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Parse and validate request
	var req users.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responder.SendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.responder.SendError(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Set request metadata
	req.SetProtocol("http")
	req.SetOperation("post")
	req.SetVersion("v1")
	req.SetRequestID(c.GetString("request_id"))
	req.SetHeaders(c.Request.Header)
	req.SetContext(map[string]interface{}{
		"user_agent": c.GetHeader("User-Agent"),
		"ip_address": c.ClientIP(),
	})

	// Create user
	ctx := context.Background()
	response, err := h.userService.CreateUser(ctx, &req)
	if err != nil {
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	// Send response
	h.responder.SendSuccess(c, http.StatusCreated, response)
}

// GetUserByID handles user retrieval by ID
// @Summary Get user by ID
// @Description Retrieves a user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} users.UserResponse "User retrieved successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid user ID"
// @Failure 404 {object} utils.ErrorResponse "User not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	// Validate user ID format
	if err := h.validator.ValidateUserID(userID); err != nil {
		h.responder.SendError(c, http.StatusBadRequest, "Invalid user ID format", err)
		return
	}

	// Get user
	ctx := context.Background()
	response, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to retrieve user", err)
		return
	}

	// Send response
	h.responder.SendSuccess(c, http.StatusOK, response)
}

// UpdateUser handles user update requests
// @Summary Update user
// @Description Updates an existing user's information
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body users.UpdateUserRequest true "User update request"
// @Success 200 {object} users.UserResponse "User updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body or validation failed"
// @Failure 404 {object} utils.ErrorResponse "User not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	// Parse and validate request
	var req users.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responder.SendError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set user ID from path parameter
	req.SetUserID(userID)

	// Validate request
	if err := req.Validate(); err != nil {
		h.responder.SendError(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Set request metadata
	req.SetProtocol("http")
	req.SetOperation("put")
	req.SetVersion("v1")
	req.SetRequestID(c.GetString("request_id"))
	req.SetHeaders(c.Request.Header)
	req.SetContext(map[string]interface{}{
		"user_agent": c.GetHeader("User-Agent"),
		"ip_address": c.ClientIP(),
	})

	// Update user
	ctx := context.Background()
	response, err := h.userService.UpdateUser(ctx, &req)
	if err != nil {
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to update user", err)
		return
	}

	// Send response
	h.responder.SendSuccess(c, http.StatusOK, response)
}

// DeleteUser handles user deletion requests
// @Summary Delete user
// @Description Soft deletes a user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} users.UserResponse "User deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid user ID"
// @Failure 404 {object} utils.ErrorResponse "User not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	// Validate user ID format
	if err := h.validator.ValidateUserID(userID); err != nil {
		h.responder.SendError(c, http.StatusBadRequest, "Invalid user ID format", err)
		return
	}

	// Delete user
	ctx := context.Background()
	response, err := h.userService.DeleteUser(ctx, userID)
	if err != nil {
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	// Send response
	h.responder.SendSuccess(c, http.StatusOK, response)
}

// ListUsers handles user listing requests
// @Summary List users
// @Description Retrieves a list of users with optional filtering and pagination
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 10, max: 100)"
// @Param status query string false "Filter by status (active, inactive)"
// @Param search query string false "Search by name or email"
// @Success 200 {object} users.UsersListResponse "Users retrieved successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse query parameters
	filters, err := h.validator.ParseListFilters(c)
	if err != nil {
		h.responder.SendError(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// List users
	ctx := context.Background()
	response, err := h.userService.ListUsers(ctx, filters)
	if err != nil {
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to list users", err)
		return
	}

	// Send response
	h.responder.SendSuccess(c, http.StatusOK, response)
}
