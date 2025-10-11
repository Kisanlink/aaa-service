package auth

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/v2/helper"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests/users"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/responses"
	userResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService interfaces.UserService
	validator   interfaces.Validator
	responder   interfaces.Responder
	logger      *zap.Logger
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(
	userService interfaces.UserService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		validator:   validator,
		responder:   responder,
		logger:      logger,
	}
}

// LoginV2 handles POST /v2/auth/login with enhanced MPIN support
//
//	@Summary		Enhanced user login with MPIN support
//	@Description	Authenticate user with phone number and either password or MPIN. Returns comprehensive user information including roles, profile, and contacts based on request flags.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.LoginRequest			true	"Login credentials with optional flags for additional data"
//	@Success		200		{object}	responses.LoginSuccessResponse	"Successful login with tokens and user info"
//	@Failure		400		{object}	responses.ErrorResponseSwagger	"Invalid request data or validation error"
//	@Failure		401		{object}	responses.ErrorResponseSwagger	"Invalid credentials or authentication failed"
//	@Failure		500		{object}	responses.ErrorResponseSwagger	"Internal server error"
//	@Router			/api/v2/auth/login [post]
func (h *AuthHandler) LoginV2(c *gin.Context) {
	h.logger.Info("Processing login request")

	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind login request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Login request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Struct validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	var userResponse *userResponses.UserResponse
	var authMethod string
	var err error

	// Check if refresh token flow is being used
	if req.HasRefreshToken() {
		// Flow 1: Refresh token + mPin
		h.logger.Info("Processing refresh token login flow")

		// Validate refresh token and get user ID
		var userID string
		userID, err = helper.ValidateToken(req.GetRefreshToken())
		if err != nil {
			h.logger.Error("Invalid refresh token", zap.Error(err))
			h.responder.SendError(c, http.StatusUnauthorized, "Invalid refresh token", err)
			return
		}

		// Verify mPin
		err = h.userService.VerifyMPin(c.Request.Context(), userID, req.GetMPin())
		if err != nil {
			h.logger.Error("Failed to verify mPin", zap.Error(err))
			h.responder.SendError(c, http.StatusUnauthorized, "Invalid mPin", err)
			return
		}

		// Get user details
		userResponse, err = h.userService.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			h.logger.Error("Failed to get user for refresh token login", zap.Error(err))
			h.responder.SendInternalError(c, err)
			return
		}
		authMethod = "refresh_token_mpin"
	} else {
		// Flow 2 & 3: Phone + password OR Phone + mPin
		h.logger.Info("Processing credential-based login flow")

		var password, mpin *string
		if req.HasPassword() {
			pwd := req.GetPassword()
			password = &pwd
			authMethod = "password"
		}
		if req.HasMPin() {
			mp := req.GetMPin()
			mpin = &mp
			authMethod = "mpin"
		}

		userResponse, err = h.userService.VerifyUserCredentials(c.Request.Context(), req.PhoneNumber, req.CountryCode, password, mpin)
		if err != nil {
			h.logger.Error("Failed to verify user credentials", zap.Error(err))
			if notFoundErr, ok := err.(*errors.NotFoundError); ok {
				h.responder.SendError(c, http.StatusUnauthorized, "Invalid credentials", notFoundErr)
				return
			}
			if unauthorizedErr, ok := err.(*errors.UnauthorizedError); ok {
				h.responder.SendError(c, http.StatusUnauthorized, "Invalid credentials", unauthorizedErr)
				return
			}
			if badRequestErr, ok := err.(*errors.BadRequestError); ok {
				h.responder.SendError(c, http.StatusBadRequest, badRequestErr.Error(), badRequestErr)
				return
			}
			h.responder.SendInternalError(c, err)
			return
		}
	}

	// Convert user roles for token generation
	var userRoles []models.UserRole
	for _, role := range userResponse.Roles {
		userRole := models.NewUserRole(role.UserID, role.RoleID)
		userRole.SetID(role.ID)
		userRole.IsActive = role.IsActive
		userRoles = append(userRoles, *userRole)
	}

	// Get user's organizations and groups for JWT context
	organizations, err := h.userService.GetUserOrganizations(c.Request.Context(), userResponse.ID)
	if err != nil {
		h.logger.Warn("Failed to get user organizations, proceeding with empty list",
			zap.String("user_id", userResponse.ID),
			zap.Error(err))
		organizations = []map[string]interface{}{}
	}

	groups, err := h.userService.GetUserGroups(c.Request.Context(), userResponse.ID)
	if err != nil {
		h.logger.Warn("Failed to get user groups, proceeding with empty list",
			zap.String("user_id", userResponse.ID),
			zap.Error(err))
		groups = []map[string]interface{}{}
	}

	// Convert organizations and groups to helper types
	orgContexts := make([]helper.OrganizationContext, len(organizations))
	for i, org := range organizations {
		orgContexts[i] = helper.OrganizationContext{
			ID:   org["id"].(string),
			Name: org["name"].(string),
		}
	}

	groupContexts := make([]helper.GroupContext, len(groups))
	for i, group := range groups {
		groupContexts[i] = helper.GroupContext{
			ID:             group["id"].(string),
			Name:           group["name"].(string),
			OrganizationID: group["organization_id"].(string),
		}
	}

	// Generate tokens with full context including organizations and groups
	username := ""
	if userResponse.Username != nil {
		username = *userResponse.Username
	}
	accessToken, err := helper.GenerateAccessTokenWithContext(
		userResponse.ID,
		userRoles,
		username,
		userResponse.PhoneNumber,
		userResponse.CountryCode,
		userResponse.IsValidated,
		orgContexts,
		groupContexts,
	)
	if err != nil {
		h.logger.Error("Failed to generate access token", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	refreshToken, err := helper.GenerateRefreshToken(userResponse.ID, userRoles, username, userResponse.IsValidated)
	if err != nil {
		h.logger.Error("Failed to generate refresh token", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	// Convert user service response to auth response format
	authUserInfo := h.convertToAuthUserInfo(userResponse)

	// Create enhanced login response
	loginResponse := &responses.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		User:         authUserInfo,
		Message:      "Login successful",
	}

	h.logger.Info("User logged in successfully",
		zap.String("userID", userResponse.ID),
		zap.String("method", authMethod),
		zap.Int("role_count", len(userResponse.Roles)))
	h.responder.SendSuccess(c, http.StatusOK, loginResponse)
}

// RegisterV2 handles POST /v2/auth/register
//
//	@Summary		Register new user account
//	@Description	Create a new user account with phone number, password, and optional profile information
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.RegisterRequest			true	"User registration data"
//	@Success		201		{object}	responses.RegisterSuccessResponse	"User registered successfully"
//	@Failure		400		{object}	responses.ErrorResponseSwagger		"Invalid request data or validation error"
//	@Failure		409		{object}	responses.ErrorResponseSwagger		"User already exists or conflict error"
//	@Failure		500		{object}	responses.ErrorResponseSwagger		"Internal server error"
//	@Router			/api/v2/auth/register [post]
func (h *AuthHandler) RegisterV2(c *gin.Context) {
	h.logger.Info("Processing registration request")

	var req requests.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind register request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Register request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Struct validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Create user using the user service
	// Convert RegisterRequest to CreateUserRequest
	createUserReq := h.convertToCreateUserRequest(&req)
	userResponse, err := h.userService.CreateUser(c.Request.Context(), createUserReq)
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

	// Create register response
	username := ""
	if userResponse.Username != nil {
		username = *userResponse.Username
	}
	registerResponse := &responses.RegisterResponse{
		User: &responses.UserInfo{
			ID:          userResponse.ID,
			Username:    &username,
			PhoneNumber: userResponse.PhoneNumber,
			CountryCode: userResponse.CountryCode,
			IsValidated: userResponse.IsValidated,
		},
		Message: "User registered successfully",
	}

	h.logger.Info("User registered successfully", zap.String("userID", userResponse.ID))
	h.responder.SendSuccess(c, http.StatusCreated, registerResponse)
}

// RefreshTokenV2 handles POST /v2/auth/refresh
//
//	@Summary		Refresh access token using MPIN
//	@Description	Generate new access and refresh tokens using existing refresh token and MPIN verification
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.RefreshTokenRequest			true	"Refresh token and MPIN"
//	@Success		200		{object}	responses.RefreshTokenSuccessResponse	"Token refreshed successfully"
//	@Failure		400		{object}	responses.ErrorResponseSwagger			"Invalid request data"
//	@Failure		401		{object}	responses.ErrorResponseSwagger			"Invalid refresh token or MPIN"
//	@Failure		500		{object}	responses.ErrorResponseSwagger			"Internal server error"
//	@Router			/api/v2/auth/refresh [post]
func (h *AuthHandler) RefreshTokenV2(c *gin.Context) {
	h.logger.Info("Processing token refresh request")

	var req requests.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind refresh token request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Refresh token request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate refresh token and get user ID
	userID, err := helper.ValidateToken(req.RefreshToken)
	if err != nil {
		h.logger.Error("Invalid refresh token", zap.Error(err))
		h.responder.SendError(c, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Verify mPin
	err = h.userService.VerifyMPin(c.Request.Context(), userID, req.MPin)
	if err != nil {
		h.logger.Error("Failed to verify mPin", zap.Error(err))
		h.responder.SendError(c, http.StatusUnauthorized, "Invalid mPin", err)
		return
	}

	// Get user details
	userResponse, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user for token refresh", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	// Convert user roles for token generation
	var userRoles []models.UserRole
	for _, role := range userResponse.Roles {
		userRole := models.NewUserRole(role.UserID, role.RoleID)
		userRole.SetID(role.ID)
		userRole.IsActive = role.IsActive
		userRole.Role = models.Role{
			BaseModel:   &base.BaseModel{},
			Name:        role.Role.Name,
			Description: role.Role.Description,
			IsActive:    role.Role.IsActive,
		}
		userRole.Role.SetID(role.Role.ID)
		userRoles = append(userRoles, *userRole)
	}

	// Get user's organizations and groups for JWT context
	organizations, err := h.userService.GetUserOrganizations(c.Request.Context(), userResponse.ID)
	if err != nil {
		h.logger.Warn("Failed to get user organizations, proceeding with empty list",
			zap.String("user_id", userResponse.ID),
			zap.Error(err))
		organizations = []map[string]interface{}{}
	}

	groups, err := h.userService.GetUserGroups(c.Request.Context(), userResponse.ID)
	if err != nil {
		h.logger.Warn("Failed to get user groups, proceeding with empty list",
			zap.String("user_id", userResponse.ID),
			zap.Error(err))
		groups = []map[string]interface{}{}
	}

	// Convert organizations and groups to helper types
	orgContexts := make([]helper.OrganizationContext, len(organizations))
	for i, org := range organizations {
		orgContexts[i] = helper.OrganizationContext{
			ID:   org["id"].(string),
			Name: org["name"].(string),
		}
	}

	groupContexts := make([]helper.GroupContext, len(groups))
	for i, group := range groups {
		groupContexts[i] = helper.GroupContext{
			ID:             group["id"].(string),
			Name:           group["name"].(string),
			OrganizationID: group["organization_id"].(string),
		}
	}

	// Generate new tokens with full context including organizations and groups
	username := ""
	if userResponse.Username != nil {
		username = *userResponse.Username
	}
	newAccessToken, err := helper.GenerateAccessTokenWithContext(
		userResponse.ID,
		userRoles,
		username,
		userResponse.PhoneNumber,
		userResponse.CountryCode,
		userResponse.IsValidated,
		orgContexts,
		groupContexts,
	)
	if err != nil {
		h.logger.Error("Failed to generate new access token", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	newRefreshToken, err := helper.GenerateRefreshToken(userResponse.ID, userRoles, username, userResponse.IsValidated)
	if err != nil {
		h.logger.Error("Failed to generate new refresh token", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	refreshResponse := &responses.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Message:      "Token refreshed successfully",
	}

	h.logger.Info("Token refreshed successfully", zap.String("userID", userID))
	h.responder.SendSuccess(c, http.StatusOK, refreshResponse)
}

// LogoutV2 handles POST /v2/auth/logout
//
//	@Summary		User logout
//	@Description	Logout user and invalidate tokens (placeholder implementation)
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.LogoutSuccessResponse	"Logged out successfully"
//	@Failure		500	{object}	responses.ErrorResponseSwagger	"Internal server error"
//	@Router			/api/v2/auth/logout [post]
//	@Security		Bearer
func (h *AuthHandler) LogoutV2(c *gin.Context) {
	h.logger.Info("Processing logout request")

	// TODO: Implement token revocation logic
	// For now, return a simple success response

	logoutResponse := map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	}

	h.logger.Info("User logged out successfully")
	h.responder.SendSuccess(c, http.StatusOK, logoutResponse)
}

// ForgotPasswordV2 handles POST /v2/auth/forgot-password
func (h *AuthHandler) ForgotPasswordV2(c *gin.Context) {
	h.logger.Info("Processing forgot password request")

	var req requests.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind forgot password request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Forgot password request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// TODO: Implement forgot password logic
	forgotResponse := map[string]interface{}{
		"message": "If the account exists, a password reset link has been sent",
		"sent_to": "***@***.com", // Masked for security
	}

	h.logger.Info("Forgot password request processed")
	h.responder.SendSuccess(c, http.StatusOK, forgotResponse)
}

// ResetPasswordV2 handles POST /v2/auth/reset-password
func (h *AuthHandler) ResetPasswordV2(c *gin.Context) {
	h.logger.Info("Processing reset password request")

	var req requests.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind reset password request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Reset password request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// TODO: Implement password reset logic
	resetResponse := map[string]interface{}{
		"success": true,
		"message": "Password reset successfully",
	}

	h.logger.Info("Password reset completed")
	h.responder.SendSuccess(c, http.StatusOK, resetResponse)
}

// Helper methods

// SetMPinV2 handles POST /v2/auth/set-mpin
//
//	@Summary		Set or update user's mPin
//	@Description	Set or update mPin for secure refresh token validation
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.SetMPinRequest	true	"Set mPin request"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		401		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v2/auth/set-mpin [post]
//	@Security		Bearer
func (h *AuthHandler) SetMPinV2(c *gin.Context) {
	h.logger.Info("Processing set mPin request")

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.responder.SendError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req requests.SetMPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind set mPin request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Set mPin request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Struct validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Verify current password first
	userIDStr, ok := userID.(string)
	if !ok {
		h.responder.SendInternalError(c, fmt.Errorf("invalid user ID type"))
		return
	}

	// Set mPin with password verification
	err := h.userService.SetMPin(c.Request.Context(), userIDStr, req.MPin, req.Password)
	if err != nil {
		h.logger.Error("Failed to set mPin", zap.Error(err))
		if unauthorizedErr, ok := err.(*errors.UnauthorizedError); ok {
			h.responder.SendError(c, http.StatusUnauthorized, "Invalid password", unauthorizedErr)
			return
		}
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

	response := map[string]any{
		"success": true,
		"message": "mPin set successfully",
	}

	h.logger.Info("mPin set successfully", zap.String("userID", userIDStr))
	h.responder.SendSuccess(c, http.StatusOK, response)
}

// UpdateMPinV2 handles POST /v2/auth/update-mpin
//
//	@Summary		Update user's existing mPin
//	@Description	Update existing mPin with current mPin verification
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.UpdateMPinRequest	true	"Update mPin request"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		401		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v2/auth/update-mpin [post]
//	@Security		Bearer
func (h *AuthHandler) UpdateMPinV2(c *gin.Context) {
	h.logger.Info("Processing update mPin request")

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.responder.SendError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req requests.UpdateMPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind update mPin request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Update mPin request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Struct validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		h.responder.SendInternalError(c, fmt.Errorf("invalid user ID type"))
		return
	}

	// Update mPin using service method that verifies current mPin
	err := h.userService.UpdateMPin(c.Request.Context(), userIDStr, req.CurrentMPin, req.NewMPin)
	if err != nil {
		h.logger.Error("Failed to update mPin", zap.Error(err))
		if unauthorizedErr, ok := err.(*errors.UnauthorizedError); ok {
			h.responder.SendError(c, http.StatusUnauthorized, "Invalid current mPin", unauthorizedErr)
			return
		}
		if validationErr, ok := err.(*errors.ValidationError); ok {
			h.responder.SendValidationError(c, []string{validationErr.Error()})
			return
		}
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, "User not found or mPin not set", notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	response := map[string]any{
		"success": true,
		"message": "mPin updated successfully",
	}

	h.logger.Info("mPin updated successfully", zap.String("userID", userIDStr))
	h.responder.SendSuccess(c, http.StatusOK, response)
}

// convertToCreateUserRequest converts a RegisterRequest to a CreateUserRequest
func (h *AuthHandler) convertToCreateUserRequest(req *requests.RegisterRequest) *users.CreateUserRequest {
	return &users.CreateUserRequest{
		PhoneNumber:   req.PhoneNumber,
		CountryCode:   req.CountryCode,
		Password:      req.Password,
		Username:      req.Username,
		AadhaarNumber: req.AadhaarNumber,
		Name:          req.Name,
	}
}

// convertToAuthUserInfo converts user service response to auth response format
func (h *AuthHandler) convertToAuthUserInfo(userResponse *userResponses.UserResponse) *responses.UserInfo {
	// Convert roles from user service format to auth response format
	authRoles := make([]responses.UserRoleDetail, len(userResponse.Roles))
	for i, role := range userResponse.Roles {
		authRoles[i] = responses.UserRoleDetail{
			ID:       role.ID,
			UserID:   role.UserID,
			RoleID:   role.RoleID,
			IsActive: role.IsActive,
			Role: responses.RoleDetail{
				ID:          role.Role.ID,
				Name:        role.Role.Name,
				Description: role.Role.Description,
				IsActive:    role.Role.IsActive,
			},
		}
	}

	return &responses.UserInfo{
		ID:          userResponse.ID,
		PhoneNumber: userResponse.PhoneNumber,
		CountryCode: userResponse.CountryCode,
		Username:    userResponse.Username,
		IsValidated: userResponse.IsValidated,
		CreatedAt:   userResponse.CreatedAt,
		UpdatedAt:   userResponse.UpdatedAt,
		Tokens:      userResponse.Tokens,
		HasMPin:     userResponse.HasMPin,
		Roles:       authRoles,
	}
}
