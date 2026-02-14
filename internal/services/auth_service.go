package services

import (
	"context"
	"fmt"
	"time"

	configPkg "github.com/Kisanlink/aaa-service/v2/internal/config"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	jwt "github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthService provides authentication services
type AuthService struct {
	userRepository     interfaces.UserRepository
	roleService        interfaces.RoleService
	userRoleRepository interfaces.UserRoleRepository
	cacheService       interfaces.CacheService
	authzService       *AuthorizationService
	auditService       *AuditService

	logger        *zap.Logger
	validator     interfaces.Validator
	jwtSecret     []byte
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
	jwtCfg        *configPkg.JWTConfig
}

// AuthServiceConfig contains configuration for AuthService
type AuthServiceConfig struct {
	JWTSecret     string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo interfaces.UserRepository,
	roleService interfaces.RoleService,
	userRoleRepo interfaces.UserRoleRepository,
	cacheService interfaces.CacheService,
	authzService *AuthorizationService,
	auditService *AuditService,
	cfg *AuthServiceConfig,
	logger *zap.Logger,
	validator interfaces.Validator,
	jwtCfg *configPkg.JWTConfig,
) (*AuthService, error) {
	return &AuthService{
		userRepository:     userRepo,
		roleService:        roleService,
		userRoleRepository: userRoleRepo,
		cacheService:       cacheService,
		authzService:       authzService,
		auditService:       auditService,

		logger:        logger,
		validator:     validator,
		jwtSecret:     []byte(cfg.JWTSecret),
		tokenExpiry:   cfg.TokenExpiry,
		refreshExpiry: cfg.RefreshExpiry,
		jwtCfg:        jwtCfg,
	}, nil
}

// LoginRequest represents a login request
type LoginRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	CountryCode string `json:"country_code" validate:"required"`
	Password    string `json:"password" validate:"required"`
	MFACode     string `json:"mfa_code,omitempty"`
}

// UsernameLoginRequest represents a username-based login request
type UsernameLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int64       `json:"expires_in"`
	User         models.User `json:"user"`
	Permissions  []string    `json:"permissions"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	PhoneNumber        string   `json:"phone_number" validate:"required"`
	CountryCode        string   `json:"country_code" validate:"required"`
	Password           string   `json:"password" validate:"required,min=8"`
	Username           *string  `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email              *string  `json:"email,omitempty" validate:"omitempty,email"`
	FullName           *string  `json:"full_name,omitempty" validate:"omitempty,min=1,max=100"`
	RoleIDs            []string `json:"role_ids,omitempty"`
	MustChangePassword bool     `json:"must_change_password,omitempty"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID      string             `json:"user_id"`
	Username    string             `json:"username"`
	IsValidated bool               `json:"is_validated"`
	Roles       []*models.UserRole `json:"roles"`
	Permissions []string           `json:"permissions"`
	TokenType   string             `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// Login authenticates a user and returns JWT tokens
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, errors.NewValidationError("invalid login request", err.Error())
	}

	// Get user by phone number
	user, err := s.userRepository.GetByPhoneNumber(ctx, req.PhoneNumber, req.CountryCode)
	if err != nil {
		s.logger.Warn("Login attempt with invalid phone number", zap.String("phone", req.CountryCode+req.PhoneNumber))
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Security check: Ensure user is not deleted
	if user.DeletedAt != nil {
		s.logger.Warn("Login attempt for deleted user", zap.String("user_id", user.ID))
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn("Login attempt with invalid password", zap.String("user_id", user.ID))
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Check if user is active
	if user.Status != nil && *user.Status != "active" {
		s.logger.Warn("Login attempt for inactive user", zap.String("user_id", user.ID), zap.String("status", *user.Status))
		return nil, errors.NewUnauthorizedError("account is not active")
	}

	// Validate MFA if required and provided
	if req.MFACode != "" {
		if err := s.validateMFA(ctx, user.ID, req.MFACode); err != nil {
			s.logger.Warn("MFA validation failed",
				zap.String("user_id", user.ID),
				zap.Error(err))
			return nil, errors.NewUnauthorizedError("invalid MFA code")
		}
		s.logger.Info("MFA validation successful", zap.String("user_id", user.ID))
	}

	// Get user roles and permissions
	userRoles, err := s.userRoleRepository.GetByUserID(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get user roles", zap.String("user_id", user.ID), zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to retrieve user roles: %w", err))
	}

	// Get effective permissions from database
	permissions, err := s.authzService.GetUserPermissions(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get user permissions", zap.String("user_id", user.ID), zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to retrieve user permissions: %w", err))
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user, userRoles, permissions)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.String("user_id", user.ID), zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate access token: %w", err))
	}

	refreshToken, err := s.generateRefreshToken(user, userRoles, permissions)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.String("user_id", user.ID), zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate refresh token: %w", err))
	}

	// Store refresh token in cache
	cacheKey := fmt.Sprintf("refresh_token:%s", user.ID)
	if err := s.cacheService.Set(cacheKey, refreshToken, int(s.refreshExpiry.Seconds())); err != nil {
		s.logger.Warn("Failed to cache refresh token", zap.String("user_id", user.ID), zap.Error(err))
	}

	// Audit login
	if s.auditService != nil {
		s.auditService.LogUserAction(ctx, user.ID, "login", "user", user.ID, map[string]interface{}{
			"username": user.Username,
		})
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	s.logger.Info("User logged in successfully", zap.String("user_id", user.ID), zap.String("username", username))

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenExpiry.Seconds()),
		User:         *user,
		Permissions:  permissions,
	}, nil
}

// LoginWithUsername authenticates a user by username and returns JWT tokens
func (s *AuthService) LoginWithUsername(ctx context.Context, req *UsernameLoginRequest) (*LoginResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, errors.NewValidationError("invalid login request", err.Error())
	}

	// Get user by username
	user, err := s.userRepository.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Warn("Login attempt with invalid username", zap.String("username", req.Username))
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Security check: Ensure user is not deleted
	if user.DeletedAt != nil {
		s.logger.Warn("Login attempt for deleted user", zap.String("user_id", user.ID))
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn("Login attempt with invalid password", zap.String("user_id", user.ID))
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Check if user is active
	if user.Status != nil && *user.Status != "active" {
		s.logger.Warn("Login attempt for inactive user", zap.String("user_id", user.ID), zap.String("status", *user.Status))
		return nil, errors.NewUnauthorizedError("account is not active")
	}

	// Validate MFA if required and provided
	if req.MFACode != "" {
		if err := s.validateMFA(ctx, user.ID, req.MFACode); err != nil {
			s.logger.Warn("MFA validation failed",
				zap.String("user_id", user.ID),
				zap.Error(err))
			return nil, errors.NewUnauthorizedError("invalid MFA code")
		}
		s.logger.Info("MFA validation successful", zap.String("user_id", user.ID))
	}

	// Get user roles and permissions
	userRoles, err := s.userRoleRepository.GetByUserID(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get user roles", zap.String("user_id", user.ID), zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to retrieve user roles: %w", err))
	}

	// Get effective permissions from database
	permissions, err := s.authzService.GetUserPermissions(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get user permissions", zap.String("user_id", user.ID), zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to retrieve user permissions: %w", err))
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user, userRoles, permissions)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.String("user_id", user.ID), zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate access token: %w", err))
	}

	refreshToken, err := s.generateRefreshToken(user, userRoles, permissions)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.String("user_id", user.ID), zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate refresh token: %w", err))
	}

	// Store refresh token in cache
	cacheKey := fmt.Sprintf("refresh_token:%s", user.ID)
	if err := s.cacheService.Set(cacheKey, refreshToken, int(s.refreshExpiry.Seconds())); err != nil {
		s.logger.Warn("Failed to cache refresh token", zap.String("user_id", user.ID), zap.Error(err))
	}

	// Audit login
	if s.auditService != nil {
		s.auditService.LogUserAction(ctx, user.ID, "login", "user", user.ID, map[string]interface{}{
			"username": user.Username,
		})
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	s.logger.Info("User logged in successfully via username", zap.String("user_id", user.ID), zap.String("username", username))

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenExpiry.Seconds()),
		User:         *user,
		Permissions:  permissions,
	}, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*LoginResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, errors.NewValidationError("invalid registration request", err.Error())
	}

	// Check if user already exists by phone number
	existingUser, err := s.userRepository.GetByPhoneNumber(ctx, req.PhoneNumber, req.CountryCode)
	if err == nil && existingUser != nil {
		return nil, errors.NewConflictError("user with this phone number already exists")
	}

	// Check if username is taken (if provided)
	if req.Username != nil && *req.Username != "" {
		existingUser, err = s.userRepository.GetByUsername(ctx, *req.Username)
		if err == nil && existingUser != nil {
			return nil, errors.NewConflictError("username already taken")
		}
	}

	// Note: The current interface doesn't have GetByEmail method
	// In a production system, you'd need to add this method to the interface

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Errorf("failed to hash password: %w", err))
	}

	// Create user using the model's constructor
	user := models.NewUser(req.PhoneNumber, req.CountryCode, string(hashedPassword))
	if req.Username != nil && *req.Username != "" {
		user.Username = req.Username
	}
	user.IsValidated = false
	user.MustChangePassword = req.MustChangePassword
	status := "active"
	user.Status = &status

	if err := s.userRepository.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to create user: %w", err))
	}

	// Assign default roles or specified roles
	if len(req.RoleIDs) == 0 {
		// Assign default user role
		defaultRole, err := s.roleService.GetRoleByName(ctx, "user")
		if err != nil {
			s.logger.Warn("Default user role not found", zap.Error(err))
		} else {
			req.RoleIDs = []string{defaultRole.ID}
		}
	}

	// Assign roles to user
	for _, roleID := range req.RoleIDs {
		if err := s.roleService.AssignRoleToUser(ctx, user.ID, roleID); err != nil {
			s.logger.Error("Failed to assign role to user", zap.String("user_id", user.ID), zap.String("role_id", roleID), zap.Error(err))
		}
	}

	// Assign roles to user in database
	for _, roleID := range req.RoleIDs {
		if err := s.authzService.AssignRoleToUser(ctx, user.ID, roleID); err != nil {
			s.logger.Error("Failed to assign role to user",
				zap.String("user_id", user.ID),
				zap.String("role_id", roleID),
				zap.Error(err))
		}
	}

	// Audit registration
	if s.auditService != nil {
		s.auditService.LogUserAction(ctx, user.ID, "register", "user", user.ID, map[string]interface{}{
			"username": user.Username,
			"roles":    req.RoleIDs,
		})
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	s.logger.Info("User registered successfully", zap.String("user_id", user.ID), zap.String("username", username))

	// Return successful registration without auto-login for now
	return &LoginResponse{
		User: *user,
		// We'll add tokens later when login is fixed
	}, nil
}

// RefreshToken refreshes an access token using a refresh token and mPin
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string, mPin string) (*LoginResponse, error) {
	// Validate refresh token
	claims, err := s.validateToken(refreshToken)
	if err != nil {
		return nil, errors.NewUnauthorizedError("invalid refresh token")
	}

	if claims.TokenType != "refresh" {
		return nil, errors.NewUnauthorizedError("invalid token type")
	}

	// Check if refresh token exists in cache
	cacheKey := fmt.Sprintf("refresh_token:%s", claims.UserID)
	cachedToken, exists := s.cacheService.Get(cacheKey)
	if !exists || cachedToken != refreshToken {
		return nil, errors.NewUnauthorizedError("refresh token has been revoked")
	}

	// Get user
	user, err := s.userRepository.GetByID(ctx, claims.UserID, &models.User{})
	if err != nil {
		return nil, errors.NewUnauthorizedError("user not found")
	}

	// Security check: Ensure user is not deleted
	if user.DeletedAt != nil {
		s.logger.Warn("Token refresh attempt for deleted user", zap.String("user_id", user.ID))
		return nil, errors.NewUnauthorizedError("user not found")
	}

	// Verify mPin
	if user.MPin == nil || *user.MPin == "" {
		return nil, errors.NewUnauthorizedError("mPin not set for this user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.MPin), []byte(mPin)); err != nil {
		s.logger.Warn("Invalid mPin during token refresh", zap.String("user_id", user.ID))
		return nil, errors.NewUnauthorizedError("invalid mPin")
	}

	// Check if user is still active
	if user.Status != nil && *user.Status != "active" {
		return nil, errors.NewUnauthorizedError("account is not active")
	}

	// Get updated user roles and permissions
	userRoles, err := s.userRoleRepository.GetByUserID(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get user roles during refresh", zap.String("user_id", user.ID), zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to retrieve user roles: %w", err))
	}

	permissions, err := s.authzService.GetUserPermissions(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get user permissions during refresh", zap.String("user_id", user.ID), zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to retrieve user permissions: %w", err))
	}

	// Generate new tokens
	accessToken, err := s.generateAccessToken(user, userRoles, permissions)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate new access token: %w", err))
	}

	newRefreshToken, err := s.generateRefreshToken(user, userRoles, permissions)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Errorf("failed to generate new refresh token: %w", err))
	}

	// Update refresh token in cache
	if err := s.cacheService.Set(cacheKey, newRefreshToken, int(s.refreshExpiry.Seconds())); err != nil {
		s.logger.Warn("Failed to update refresh token in cache", zap.String("user_id", user.ID), zap.Error(err))
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenExpiry.Seconds()),
		User:         *user,
		Permissions:  permissions,
	}, nil
}

// Logout invalidates user tokens
func (s *AuthService) Logout(ctx context.Context, userID string) error {
	// Remove refresh token from cache
	cacheKey := fmt.Sprintf("refresh_token:%s", userID)
	if err := s.cacheService.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete refresh token from cache", zap.String("user_id", userID), zap.Error(err))
	}

	// TODO: Add access token to blacklist with expiration time

	// Audit logout
	if s.auditService != nil {
		s.auditService.LogUserAction(ctx, userID, "logout", "user", userID, nil)
	}

	s.logger.Info("User logged out", zap.String("user_id", userID))
	return nil
}

// ValidateToken validates a JWT token and returns claims
func (s *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	return s.validateToken(tokenString)
}

// generateAccessToken generates a JWT access token
func (s *AuthService) generateAccessToken(user *models.User, roles []*models.UserRole, permissions []string) (string, error) {
	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	now := time.Now()
	iat := now.Add(-s.jwtCfg.Leeway / 2)
	nbf := now.Add(-s.jwtCfg.Leeway / 2)
	exp := now.Add(s.jwtCfg.TTL)

	claims := &TokenClaims{
		UserID:      user.ID,
		Username:    username,
		IsValidated: user.IsValidated,
		Roles:       roles,
		Permissions: permissions,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: exp},
			IssuedAt:  &jwt.NumericDate{Time: iat},
			NotBefore: &jwt.NumericDate{Time: nbf},
			Subject:   user.ID,
			Issuer:    s.jwtCfg.Issuer,
			Audience:  []string{s.jwtCfg.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

// generateRefreshToken generates a JWT refresh token
func (s *AuthService) generateRefreshToken(user *models.User, roles []*models.UserRole, permissions []string) (string, error) {
	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	now := time.Now()
	iat := now.Add(-s.jwtCfg.Leeway / 2)
	nbf := now.Add(-s.jwtCfg.Leeway / 2)
	exp := now.Add(s.refreshExpiry)

	claims := &TokenClaims{
		UserID:      user.ID,
		Username:    username,
		IsValidated: user.IsValidated,
		Roles:       roles,
		Permissions: permissions,
		TokenType:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: exp},
			IssuedAt:  &jwt.NumericDate{Time: iat},
			NotBefore: &jwt.NumericDate{Time: nbf},
			Subject:   user.ID,
			Issuer:    s.jwtCfg.Issuer,
			Audience:  []string{s.jwtCfg.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

// validateToken validates a JWT token and returns claims
func (s *AuthService) validateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Use the same secret as token generation (jwtCfg.Secret)
		return []byte(s.jwtCfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// validateMFA validates a multi-factor authentication code
func (s *AuthService) validateMFA(ctx context.Context, userID, mfaCode string) error {
	// Get MFA settings for the user from cache or database
	cacheKey := fmt.Sprintf("mfa:%s", userID)
	mfaData, exists := s.cacheService.Get(cacheKey)
	if !exists {
		// In practice, you'd fetch MFA configuration from database
		// For now, return an error indicating MFA is not set up
		return fmt.Errorf("MFA not configured for user")
	}

	// Parse MFA data (in practice, this would be a structured object)
	mfaConfig, ok := mfaData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid MFA configuration")
	}

	mfaType := mfaConfig["type"].(string)

	switch mfaType {
	case "totp":
		// Validate TOTP code
		return s.validateTOTP(ctx, userID, mfaCode, mfaConfig)
	case "sms":
		// Validate SMS code
		return s.validateSMSCode(ctx, userID, mfaCode, mfaConfig)
	case "email":
		// Validate email code
		return s.validateEmailCode(ctx, userID, mfaCode, mfaConfig)
	default:
		return fmt.Errorf("unsupported MFA type: %s", mfaType)
	}
}

// validateTOTP validates a Time-based One-Time Password
func (s *AuthService) validateTOTP(ctx context.Context, userID, code string, config map[string]interface{}) error {
	// This would integrate with a TOTP library like github.com/pquerna/otp
	// For now, implement a basic validation

	// In a real implementation, you'd:
	// 1. Get the user's TOTP secret from secure storage
	// 2. Generate the expected code for the current time window
	// 3. Compare with the provided code (allowing for time skew)

	// Mock implementation - accept any 6-digit code
	if len(code) != 6 {
		return fmt.Errorf("TOTP code must be 6 digits")
	}

	// Log the MFA attempt for security auditing
	s.auditService.LogSecurityEvent(ctx, userID, "mfa_totp_validation", "authentication", true, map[string]interface{}{
		"mfa_type": "totp",
		"success":  true,
	})

	return nil
}

// validateSMSCode validates an SMS-based MFA code
func (s *AuthService) validateSMSCode(ctx context.Context, userID, code string, config map[string]interface{}) error {
	// Get the expected code from cache (stored when SMS was sent)
	cacheKey := fmt.Sprintf("sms_code:%s", userID)
	expectedCode, exists := s.cacheService.Get(cacheKey)
	if !exists {
		return fmt.Errorf("SMS code expired or not found")
	}

	if code != expectedCode.(string) {
		s.auditService.LogSecurityEvent(ctx, userID, "mfa_sms_validation", "authentication", false, map[string]interface{}{
			"mfa_type": "sms",
			"success":  false,
			"reason":   "invalid_code",
		})
		return fmt.Errorf("invalid SMS code")
	}

	// Remove the code from cache after successful validation
	if err := s.cacheService.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete cache key", zap.String("key", cacheKey), zap.Error(err))
	}

	s.auditService.LogSecurityEvent(ctx, userID, "mfa_sms_validation", "authentication", true, map[string]interface{}{
		"mfa_type": "sms",
		"success":  true,
	})

	return nil
}

// validateEmailCode validates an email-based MFA code
func (s *AuthService) validateEmailCode(ctx context.Context, userID, code string, config map[string]interface{}) error {
	// Similar to SMS validation
	cacheKey := fmt.Sprintf("email_code:%s", userID)
	expectedCode, exists := s.cacheService.Get(cacheKey)
	if !exists {
		return fmt.Errorf("email code expired or not found")
	}

	if code != expectedCode.(string) {
		s.auditService.LogSecurityEvent(ctx, userID, "mfa_email_validation", "authentication", false, map[string]interface{}{
			"mfa_type": "email",
			"success":  false,
			"reason":   "invalid_code",
		})
		return fmt.Errorf("invalid email code")
	}

	// Remove the code from cache after successful validation
	if err := s.cacheService.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete cache key", zap.String("key", cacheKey), zap.Error(err))
	}

	s.auditService.LogSecurityEvent(ctx, userID, "mfa_email_validation", "authentication", true, map[string]interface{}{
		"mfa_type": "email",
		"success":  true,
	})

	return nil
}

// SetMPin sets the user's mPin (only if not already set)
func (s *AuthService) SetMPin(ctx context.Context, userID string, mPin string, currentPassword string) error {
	// Get user
	user, err := s.userRepository.GetByID(ctx, userID, &models.User{})
	if err != nil {
		return errors.NewNotFoundError("user not found")
	}

	// Security check: Ensure user is not deleted
	if user.DeletedAt != nil {
		s.logger.Warn("SetMPin attempt for deleted user", zap.String("user_id", userID))
		return errors.NewNotFoundError("user not found")
	}

	// Check if mPin is already set
	if user.MPin != nil && *user.MPin != "" {
		return errors.NewConflictError("mPin is already set. Use update-mpin endpoint to change it")
	}

	// Verify current password is required
	if currentPassword == "" {
		return errors.NewValidationError("current password is required to set mPin")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		s.logger.Warn("Invalid password during mPin setup", zap.String("user_id", userID))
		return errors.NewUnauthorizedError("invalid password")
	}

	// Hash the mPin
	hashedMPin, err := bcrypt.GenerateFromPassword([]byte(mPin), bcrypt.DefaultCost)
	if err != nil {
		return errors.NewInternalError(fmt.Errorf("failed to hash mPin: %w", err))
	}

	// Update user's mPin
	user.SetMPin(string(hashedMPin))
	if err := s.userRepository.Update(ctx, user); err != nil {
		return errors.NewInternalError(fmt.Errorf("failed to update mPin: %w", err))
	}

	// Audit the action
	if s.auditService != nil {
		s.auditService.LogUserAction(ctx, userID, "set_mpin", "user", userID, map[string]interface{}{
			"action": "mPin set",
		})
	}

	s.logger.Info("mPin set successfully", zap.String("user_id", userID))
	return nil
}

// UpdateMPin updates the user's existing mPin
func (s *AuthService) UpdateMPin(ctx context.Context, userID string, currentMPin string, newMPin string) error {
	// Get user
	user, err := s.userRepository.GetByID(ctx, userID, &models.User{})
	if err != nil {
		return errors.NewNotFoundError("user not found")
	}

	// Security check: Ensure user is not deleted
	if user.DeletedAt != nil {
		s.logger.Warn("UpdateMPin attempt for deleted user", zap.String("user_id", userID))
		return errors.NewNotFoundError("user not found")
	}

	// Check if mPin is set
	if user.MPin == nil || *user.MPin == "" {
		return errors.NewNotFoundError("mPin not set for this user. Use set-mpin endpoint first")
	}

	// Verify current mPin
	if err := bcrypt.CompareHashAndPassword([]byte(*user.MPin), []byte(currentMPin)); err != nil {
		s.logger.Warn("Invalid current mPin during update", zap.String("user_id", userID))
		return errors.NewUnauthorizedError("invalid current mPin")
	}

	// Hash the new mPin
	hashedMPin, err := bcrypt.GenerateFromPassword([]byte(newMPin), bcrypt.DefaultCost)
	if err != nil {
		return errors.NewInternalError(fmt.Errorf("failed to hash new mPin: %w", err))
	}

	// Update user's mPin
	user.SetMPin(string(hashedMPin))
	if err := s.userRepository.Update(ctx, user); err != nil {
		return errors.NewInternalError(fmt.Errorf("failed to update mPin: %w", err))
	}

	// Audit the action
	if s.auditService != nil {
		s.auditService.LogUserAction(ctx, userID, "update_mpin", "user", userID, map[string]interface{}{
			"action": "mPin updated",
		})
	}

	s.logger.Info("mPin updated successfully", zap.String("user_id", userID))
	return nil
}

// VerifyMPin verifies a user's mPin
func (s *AuthService) VerifyMPin(ctx context.Context, userID string, mPin string) error {
	// Get user
	user, err := s.userRepository.GetByID(ctx, userID, &models.User{})
	if err != nil {
		return errors.NewNotFoundError("user not found")
	}

	// Security check: Ensure user is not deleted
	if user.DeletedAt != nil {
		s.logger.Warn("VerifyMPin attempt for deleted user", zap.String("user_id", userID))
		return errors.NewNotFoundError("user not found")
	}

	// Check if mPin is set
	if user.MPin == nil || *user.MPin == "" {
		return errors.NewUnauthorizedError("mPin not set for this user")
	}

	// Verify mPin
	if err := bcrypt.CompareHashAndPassword([]byte(*user.MPin), []byte(mPin)); err != nil {
		s.logger.Warn("Invalid mPin verification attempt", zap.String("user_id", userID))
		return errors.NewUnauthorizedError("invalid mPin")
	}

	return nil
}

// CheckSystemPermission checks if user has system-level permissions
func (s *AuthService) CheckSystemPermission(ctx context.Context, userID, permission string) bool {
	// Check permission using authorization service
	perm := &Permission{
		UserID:     userID,
		Resource:   "system",
		ResourceID: "main",
		Action:     permission,
	}

	result, err := s.authzService.CheckPermission(ctx, perm)
	if err != nil {
		s.logger.Warn("Failed to check system permission",
			zap.String("user_id", userID),
			zap.String("permission", permission),
			zap.Error(err))

		// Fallback: Check if user has admin role in database
		return s.checkAdminRoleFallback(ctx, userID)
	}

	return result.Allowed
}

// checkAdminRoleFallback checks if user has admin role in database
func (s *AuthService) checkAdminRoleFallback(ctx context.Context, userID string) bool {
	// Get user roles from database
	userRoles, err := s.userRoleRepository.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get user roles for fallback check",
			zap.String("user_id", userID),
			zap.Error(err))
		return false
	}

	// Check if user has admin role
	for _, userRole := range userRoles {
		if userRole.Role.Name == "admin" || userRole.Role.Name == "super_admin" {
			return true
		}
	}

	return false
}
