package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	userResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/users"
	"github.com/Kisanlink/aaa-service/v2/internal/services/sms"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// OTP expiry duration for password reset
const passwordResetOTPExpiry = 10 * time.Minute

// PasswordResetService holds dependencies for password reset operations
type PasswordResetService struct {
	userRepo interface {
		GetByID(ctx context.Context, id string) (*models.User, error)
		GetByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (*models.User, error)
		GetByUsername(ctx context.Context, username string) (*models.User, error)
		GetByEmail(ctx context.Context, email string) (*models.User, error)
		Update(ctx context.Context, user *models.User) error
	}
	resetTokenRepo *users.PasswordResetRepository
	emailService   interface{} // TODO: Define email service interface
	logger         *zap.Logger
	tokenExpiry    time.Duration
}

// SetPasswordResetRepository sets the password reset repository
func (s *Service) SetPasswordResetRepository(repo *users.PasswordResetRepository) {
	// This will be called during service initialization
	// For now, we'll store it in a way that the password reset methods can access
}

// InitiatePasswordReset creates a password reset OTP and sends it via SMS
// Returns the token ID (not the OTP) for use in the reset step
func (s *Service) InitiatePasswordReset(ctx context.Context, phoneNumber, countryCode, username, email *string) (string, error) {
	s.logger.Info("Initiating password reset",
		zap.Any("phone", phoneNumber),
		zap.Any("username", username),
		zap.Any("email", email))

	// Find user by provided identifier
	var user *models.User
	var err error

	if phoneNumber != nil && countryCode != nil {
		user, err = s.userRepo.GetByPhoneNumber(ctx, *phoneNumber, *countryCode)
	} else if username != nil {
		user, err = s.userRepo.GetByUsername(ctx, *username)
	} else if email != nil {
		user, err = s.userRepo.GetByEmail(ctx, *email)
	} else {
		return "", fmt.Errorf("at least one identifier (phone, username, or email) must be provided")
	}

	if err != nil {
		// Don't reveal whether user exists for security
		s.logger.Warn("User not found for password reset", zap.Error(err))
		return "", nil // Return success but don't send anything
	}

	// Create password reset repository instance (temporary until proper DI)
	// TODO: This should be injected via constructor
	// Cast userRepo to access dbManager
	userRepoImpl, ok := s.userRepo.(*users.UserRepository)
	if !ok {
		return "", fmt.Errorf("invalid user repository implementation")
	}
	resetTokenRepo := users.NewPasswordResetRepository(userRepoImpl.GetDBManager(), userRepoImpl)

	// Invalidate any existing tokens for this user
	if err := resetTokenRepo.InvalidateUserTokens(ctx, user.ID); err != nil {
		s.logger.Error("Failed to invalidate existing tokens", zap.Error(err))
		// Continue anyway
	}

	// Generate 6-digit OTP
	otp, err := sms.GenerateOTP()
	if err != nil {
		s.logger.Error("Failed to generate OTP", zap.Error(err))
		return "", fmt.Errorf("failed to generate password reset code")
	}

	// Hash the OTP before storing (for security)
	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash OTP", zap.Error(err))
		return "", fmt.Errorf("failed to process password reset code")
	}

	// Create reset token with hashed OTP (10 minute expiry)
	resetToken, err := resetTokenRepo.CreateResetTokenWithValue(ctx, user.ID, string(hashedOTP), passwordResetOTPExpiry)
	if err != nil {
		s.logger.Error("Failed to create reset token", zap.Error(err))
		return "", fmt.Errorf("failed to create password reset token")
	}

	s.logger.Info("Password reset OTP created",
		zap.String("user_id", user.ID),
		zap.String("token_id", resetToken.GetID()))

	// Send OTP via SMS if phone number is available and SMS service is configured
	if phoneNumber != nil && countryCode != nil && s.smsService != nil {
		// Format phone number as E.164 (e.g., +919876543210)
		// Country code may already include '+' prefix
		cc := *countryCode
		if !strings.HasPrefix(cc, "+") {
			cc = "+" + cc
		}
		fullPhoneNumber := fmt.Sprintf("%s%s", cc, *phoneNumber)
		if err := s.smsService.SendOTP(ctx, fullPhoneNumber, otp); err != nil {
			s.logger.Error("Failed to send password reset OTP via SMS",
				zap.String("user_id", user.ID),
				zap.Error(err))
			// Don't fail the operation - token was created, user can retry SMS
		} else {
			s.logger.Info("Password reset OTP sent via SMS",
				zap.String("user_id", user.ID),
				zap.String("phone_masked", models.MaskPhoneNumber(*phoneNumber)))
		}
	} else if s.smsService == nil {
		s.logger.Warn("SMS service not configured, OTP not sent")
	}

	// Return token ID (not the OTP) - user receives OTP via SMS
	return resetToken.GetID(), nil
}

// ResetPassword validates the OTP and updates the user's password
// tokenID: The ID returned from InitiatePasswordReset
// otp: The 6-digit OTP sent via SMS
// newPassword: The new password to set
func (s *Service) ResetPassword(ctx context.Context, tokenID, otp, newPassword string) error {
	s.logger.Info("Processing password reset", zap.String("token_id", tokenID))

	// Validate inputs
	if tokenID == "" {
		return fmt.Errorf("token ID is required")
	}
	if otp == "" {
		return fmt.Errorf("OTP is required")
	}
	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Create password reset repository instance (temporary until proper DI)
	// TODO: This should be injected via constructor
	// Cast userRepo to access dbManager
	userRepoImpl, ok := s.userRepo.(*users.UserRepository)
	if !ok {
		return fmt.Errorf("invalid user repository implementation")
	}
	resetTokenRepo := users.NewPasswordResetRepository(userRepoImpl.GetDBManager(), userRepoImpl)

	// Get token by ID
	resetToken, err := resetTokenRepo.GetTokenByID(ctx, tokenID)
	if err != nil {
		s.logger.Error("Invalid token ID", zap.Error(err))
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check if token is valid (not used, not expired)
	if !resetToken.IsValid() {
		s.logger.Warn("Reset token is not valid",
			zap.Bool("used", resetToken.Used),
			zap.Bool("expired", resetToken.IsExpired()))
		return fmt.Errorf("invalid or expired reset token")
	}

	// Verify OTP against stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(resetToken.Token), []byte(otp)); err != nil {
		s.logger.Warn("Invalid OTP provided",
			zap.String("token_id", tokenID),
			zap.Error(err))
		return fmt.Errorf("invalid OTP")
	}

	// Get user - Initialize with BaseModel to allow GORM to scan into it
	user := &models.User{
		BaseModel: &base.BaseModel{},
	}
	_, err = s.userRepo.GetByID(ctx, resetToken.UserID, user)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return fmt.Errorf("user not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return fmt.Errorf("failed to process password")
	}

	// Update user password and clear force-change flag
	user.Password = string(hashedPassword)
	user.MustChangePassword = false
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update password", zap.Error(err))
		return fmt.Errorf("failed to update password")
	}

	// Mark token as used
	if err := resetTokenRepo.MarkTokenAsUsed(ctx, resetToken.GetID()); err != nil {
		s.logger.Error("Failed to mark token as used", zap.Error(err))
		// Don't fail the operation, password was already updated
	}

	s.logger.Info("Password reset successful", zap.String("user_id", user.ID))
	return nil
}

// ResetPasswordWithToken validates the reset token (legacy method) and updates the user's password
// Deprecated: Use ResetPassword with OTP instead
func (s *Service) ResetPasswordWithToken(ctx context.Context, token, newPassword string) error {
	s.logger.Info("Processing password reset with token", zap.String("token", token[:10]+"..."))

	// Validate inputs
	if token == "" {
		return fmt.Errorf("reset token is required")
	}
	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Create password reset repository instance (temporary until proper DI)
	userRepoImpl, ok := s.userRepo.(*users.UserRepository)
	if !ok {
		return fmt.Errorf("invalid user repository implementation")
	}
	resetTokenRepo := users.NewPasswordResetRepository(userRepoImpl.GetDBManager(), userRepoImpl)

	// Get and validate token by value
	resetToken, err := resetTokenRepo.GetTokenByValue(ctx, token)
	if err != nil {
		s.logger.Error("Invalid reset token", zap.Error(err))
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check if token is valid
	if !resetToken.IsValid() {
		s.logger.Warn("Reset token is not valid",
			zap.Bool("used", resetToken.Used),
			zap.Bool("expired", resetToken.IsExpired()))
		return fmt.Errorf("invalid or expired reset token")
	}

	// Get user
	user := &models.User{
		BaseModel: &base.BaseModel{},
	}
	_, err = s.userRepo.GetByID(ctx, resetToken.UserID, user)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return fmt.Errorf("user not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return fmt.Errorf("failed to process password")
	}

	// Update user password and clear force-change flag
	user.Password = string(hashedPassword)
	user.MustChangePassword = false
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update password", zap.Error(err))
		return fmt.Errorf("failed to update password")
	}

	// Mark token as used
	if err := resetTokenRepo.MarkTokenAsUsed(ctx, resetToken.GetID()); err != nil {
		s.logger.Error("Failed to mark token as used", zap.Error(err))
	}

	s.logger.Info("Password reset successful", zap.String("user_id", user.ID))
	return nil
}

// GetUserByEmail retrieves a user by email address
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user by email", zap.String("email", email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Failed to get user by email", zap.Error(err))
		return nil, fmt.Errorf("user not found")
	}

	return s.convertToUserResponse(user), nil
}

// convertToUserResponse converts a User model to UserResponse
func (s *Service) convertToUserResponse(user *models.User) *userResponses.UserResponse {
	if user == nil {
		return nil
	}

	return &userResponses.UserResponse{
		ID:                 user.ID,
		Username:           user.Username,
		PhoneNumber:        user.PhoneNumber,
		CountryCode:        user.CountryCode,
		IsValidated:        user.IsValidated,
		MustChangePassword: user.MustChangePassword,
		Status:             user.Status,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}
}
