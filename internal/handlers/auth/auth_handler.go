package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/helper"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests/users"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/responses"
	userResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
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

// setAuthCookies sets HTTP-only cookies for access and refresh tokens
// This provides secure cookie-based authentication while maintaining backward compatibility
// with JSON response tokens for other clients
func (h *AuthHandler) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	// Determine if we're in a secure context (HTTPS)
	secure := isSecureContext()

	// Cookie path - root so it works for all API endpoints
	path := "/"

	// Set access token cookie (1 hour expiry)
	c.SetCookie(
		"auth_token", // name - matches middleware expectation
		accessToken,  // value
		3600,         // maxAge in seconds (1 hour)
		path,         // path
		"",           // domain (empty = current domain)
		secure,       // secure flag
		true,         // httpOnly
	)

	// Set refresh token cookie (7 days expiry)
	c.SetCookie(
		"refresh_token", // name
		refreshToken,    // value
		604800,          // maxAge in seconds (7 days)
		path,            // path
		"",              // domain
		secure,          // secure flag
		true,            // httpOnly
	)

	// Set SameSite attribute via header for better browser compatibility
	// Gin's SetCookie doesn't support SameSite directly in older versions
	c.Header("Set-Cookie", c.Writer.Header().Get("Set-Cookie"))
}

// clearAuthCookies removes auth cookies on logout
func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	secure := isSecureContext()
	path := "/"

	// Clear access token cookie
	c.SetCookie("auth_token", "", -1, path, "", secure, true)

	// Clear refresh token cookie
	c.SetCookie("refresh_token", "", -1, path, "", secure, true)
}

// isSecureContext determines if cookies should be set with Secure flag
func isSecureContext() bool {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	return env == "production" || env == "prod" || env == "staging"
}

// Login handles POST /api/v1/auth/login with MPIN support
//
//	@Summary		User login with MPIN support
//	@Description	Authenticate user with phone number and either password or MPIN. Returns comprehensive user information including roles, profile, and contacts based on request flags.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.LoginRequest			true	"Login credentials with optional flags for additional data"
//	@Success		200		{object}	responses.LoginSuccessResponse	"Successful login with tokens and user info"
//	@Failure		400		{object}	responses.ErrorResponseSwagger	"Invalid request data or validation error"
//	@Failure		401		{object}	responses.ErrorResponseSwagger	"Invalid credentials or authentication failed"
//	@Failure		500		{object}	responses.ErrorResponseSwagger	"Internal server error"
//	@Router			/api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
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

	// Convert user roles for token generation with complete Role data
	var userRoles []models.UserRole
	for _, roleDetail := range userResponse.Roles {
		userRole := models.NewUserRole(roleDetail.UserID, roleDetail.RoleID)
		userRole.SetID(roleDetail.ID)
		userRole.IsActive = roleDetail.IsActive

		// Populate the Role relationship from the response data
		// Note: Scope is not available in users.RoleDetail, will default to empty RoleScope
		role := models.NewRole(roleDetail.Role.Name, roleDetail.Role.Description, "")
		role.SetID(roleDetail.Role.ID)
		role.IsActive = roleDetail.Role.IsActive
		userRole.Role = *role

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

	// Set HTTP-only cookies for browser clients (backward compatible - JSON response still sent)
	h.setAuthCookies(c, accessToken, refreshToken)

	h.logger.Info("User logged in successfully",
		zap.String("userID", userResponse.ID),
		zap.String("method", authMethod),
		zap.Int("role_count", len(userResponse.Roles)))
	h.responder.SendSuccess(c, http.StatusOK, loginResponse)
}

// Register handles POST /api/v1/auth/register
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
//	@Router			/api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
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

// RefreshToken handles POST /api/v1/auth/refresh
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
//	@Router			/api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
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

	// Convert user roles for token generation with complete Role data
	var userRoles []models.UserRole
	for _, roleDetail := range userResponse.Roles {
		userRole := models.NewUserRole(roleDetail.UserID, roleDetail.RoleID)
		userRole.SetID(roleDetail.ID)
		userRole.IsActive = roleDetail.IsActive

		// Populate the Role relationship from the response data
		// Note: Scope is not available in users.RoleDetail, will default to empty RoleScope
		role := models.NewRole(roleDetail.Role.Name, roleDetail.Role.Description, "")
		role.SetID(roleDetail.Role.ID)
		role.IsActive = roleDetail.Role.IsActive
		userRole.Role = *role

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

	// Set HTTP-only cookies for browser clients (backward compatible - JSON response still sent)
	h.setAuthCookies(c, newAccessToken, newRefreshToken)

	h.logger.Info("Token refreshed successfully", zap.String("userID", userID))
	h.responder.SendSuccess(c, http.StatusOK, refreshResponse)
}

// Logout handles POST /api/v1/auth/logout
//
//	@Summary		User logout
//	@Description	Logout user and invalidate tokens (placeholder implementation)
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.LogoutSuccessResponse	"Logged out successfully"
//	@Failure		500	{object}	responses.ErrorResponseSwagger	"Internal server error"
//	@Router			/api/v1/auth/logout [post]
//	@Security		Bearer
func (h *AuthHandler) Logout(c *gin.Context) {
	h.logger.Info("Processing logout request")

	// Clear HTTP-only auth cookies
	h.clearAuthCookies(c)

	// TODO: Implement token revocation logic
	// For now, return a simple success response

	logoutResponse := map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	}

	h.logger.Info("User logged out successfully")
	h.responder.SendSuccess(c, http.StatusOK, logoutResponse)
}

// ForgotPassword handles POST /api/v1/auth/forgot-password
// Sends a 6-digit OTP via SMS for password reset
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
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

	// Initiate password reset - this will create an OTP and send via SMS
	tokenID, err := h.userService.InitiatePasswordReset(
		c.Request.Context(),
		req.PhoneNumber,
		req.CountryCode,
		req.Username,
		req.Email,
	)
	if err != nil {
		h.logger.Error("Failed to initiate password reset", zap.Error(err))
		// Still return success to prevent user enumeration
	}

	// For security, always return the same response whether user exists or not
	forgotResponse := map[string]interface{}{
		"message": "If the account exists, a password reset code has been sent via SMS",
	}

	// Include token_id for the reset step (required for OTP verification)
	if tokenID != "" {
		forgotResponse["token_id"] = tokenID
	}

	// Mask the phone number in response
	if req.PhoneNumber != nil && len(*req.PhoneNumber) >= 4 {
		maskedPhone := "XXXX-XXX-" + (*req.PhoneNumber)[len(*req.PhoneNumber)-4:]
		forgotResponse["sent_to"] = maskedPhone
	}

	h.logger.Info("Forgot password request processed")
	h.responder.SendSuccess(c, http.StatusOK, forgotResponse)
}

// ResetPassword handles POST /api/v1/auth/reset-password
// Verifies OTP and sets new password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
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

	// Reset the password using token ID and OTP
	err := h.userService.ResetPassword(c.Request.Context(), req.TokenID, req.OTP, req.NewPassword)
	if err != nil {
		h.logger.Error("Failed to reset password", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "Failed to reset password. OTP may be invalid or expired.", err)
		return
	}

	resetResponse := map[string]interface{}{
		"success": true,
		"message": "Password reset successfully. You can now login with your new password.",
	}

	h.logger.Info("Password reset completed")
	h.responder.SendSuccess(c, http.StatusOK, resetResponse)
}

// Helper methods

// SetMPin handles POST /api/v1/auth/set-mpin
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
//	@Router			/api/v1/auth/set-mpin [post]
//	@Security		Bearer
func (h *AuthHandler) SetMPin(c *gin.Context) {
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

// UpdateMPin handles POST /api/v1/auth/update-mpin
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
//	@Router			/api/v1/auth/update-mpin [post]
//	@Security		Bearer
func (h *AuthHandler) UpdateMPin(c *gin.Context) {
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
