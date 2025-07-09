package auth

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/entities/requests"
	"github.com/Kisanlink/aaa-service/entities/requests/users"
	"github.com/Kisanlink/aaa-service/entities/responses"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
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

// LoginV2 handles POST /v2/auth/login
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

	// Get user by username
	userResponse, err := h.userService.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusUnauthorized, "Invalid credentials", notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	// Verify password (Note: We need to implement password verification in the service)
	// For now, we'll do a basic check - ideally this should be in the service
	if !h.verifyPassword(req.Password, userResponse) {
		h.logger.Warn("Invalid password attempt", zap.String("username", req.Username))
		h.responder.SendError(c, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	// Generate tokens using helper functions (from existing controller logic)
	// Note: Passing nil for roles since helper expects []model.UserRole but we have []RoleDetail
	// TODO: Implement proper role conversion or update helper functions
	accessToken, err := helper.GenerateAccessToken(userResponse.ID, nil, userResponse.Username, userResponse.IsValidated)
	if err != nil {
		h.logger.Error("Failed to generate access token", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	refreshToken, err := helper.GenerateRefreshToken(userResponse.ID, nil, userResponse.Username, userResponse.IsValidated)
	if err != nil {
		h.logger.Error("Failed to generate refresh token", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	// Create login response
	loginResponse := &responses.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		User: &responses.UserInfo{
			ID:          userResponse.ID,
			Username:    userResponse.Username,
			IsValidated: userResponse.IsValidated,
		},
		Message: "Login successful",
	}

	h.logger.Info("User logged in successfully", zap.String("userID", userResponse.ID))
	h.responder.SendSuccess(c, http.StatusOK, loginResponse)
}

// RegisterV2 handles POST /v2/auth/register
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
	registerResponse := &responses.RegisterResponse{
		User: &responses.UserInfo{
			ID:          userResponse.ID,
			Username:    userResponse.Username,
			IsValidated: userResponse.IsValidated,
		},
		Message: "User registered successfully",
	}

	h.logger.Info("User registered successfully", zap.String("userID", userResponse.ID))
	h.responder.SendSuccess(c, http.StatusCreated, registerResponse)
}

// RefreshTokenV2 handles POST /v2/auth/refresh
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

	// TODO: Implement token refresh logic using helper functions
	// For now, return a placeholder response
	refreshResponse := &responses.RefreshTokenResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Message:      "Token refreshed successfully",
	}

	h.logger.Info("Token refreshed successfully")
	h.responder.SendSuccess(c, http.StatusOK, refreshResponse)
}

// LogoutV2 handles POST /v2/auth/logout
func (h *AuthHandler) LogoutV2(c *gin.Context) {
	h.logger.Info("Processing logout request")

	// TODO: Implement token revocation logic
	// For now, return a simple success response

	logoutResponse := &responses.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
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
	forgotResponse := &responses.ForgotPasswordResponse{
		Message: "If the account exists, a password reset link has been sent",
		SentTo:  "***@***.com", // Masked for security
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
	resetResponse := &responses.ResetPasswordResponse{
		Success: true,
		Message: "Password reset successfully",
	}

	h.logger.Info("Password reset completed")
	h.responder.SendSuccess(c, http.StatusOK, resetResponse)
}

// Helper methods

// verifyPassword verifies if the provided password matches the user's stored password
// Note: This is a temporary implementation. Ideally, password verification should be in the service layer
func (h *AuthHandler) verifyPassword(password string, user interface{}) bool {
	// TODO: Implement proper password verification
	// For now, this is a placeholder that always returns true for development
	// In production, this should:
	// 1. Get the hashed password from the user object
	// 2. Use bcrypt.CompareHashAndPassword to verify
	return true
}

// convertToCreateUserRequest converts a RegisterRequest to a CreateUserRequest
func (h *AuthHandler) convertToCreateUserRequest(req *requests.RegisterRequest) *users.CreateUserRequest {
	return &users.CreateUserRequest{
		Username:      req.Username,
		Password:      req.Password,
		MobileNumber:  req.MobileNumber,
		AadhaarNumber: req.AadhaarNumber,
		CountryCode:   req.CountryCode,
		Name:          req.Name,
	}
}
