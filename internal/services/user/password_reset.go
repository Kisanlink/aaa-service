package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	userResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/users"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

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

// InitiatePasswordReset creates a password reset token and sends it via email/SMS
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

	// Create new reset token (valid for 1 hour by default)
	resetToken, err := resetTokenRepo.CreateResetToken(ctx, user.ID, time.Hour)
	if err != nil {
		s.logger.Error("Failed to create reset token", zap.Error(err))
		return "", fmt.Errorf("failed to create password reset token")
	}

	// TODO: Send email with reset link
	// For now, return the token (in production, this would be sent via email)
	s.logger.Info("Password reset token created",
		zap.String("user_id", user.ID),
		zap.String("token", resetToken.Token))

	// TODO: Implement email service integration
	// emailService.SendPasswordResetEmail(user.Email, resetToken.Token)

	return resetToken.Token, nil
}

// ResetPassword validates the reset token and updates the user's password
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	s.logger.Info("Processing password reset", zap.String("token", token[:10]+"..."))

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
	// TODO: This should be injected via constructor
	// Cast userRepo to access dbManager
	userRepoImpl, ok := s.userRepo.(*users.UserRepository)
	if !ok {
		return fmt.Errorf("invalid user repository implementation")
	}
	resetTokenRepo := users.NewPasswordResetRepository(userRepoImpl.GetDBManager(), userRepoImpl)

	// Get and validate token
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
	var user models.User
	_, err = s.userRepo.GetByID(ctx, resetToken.UserID, &user)
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

	// Update user password
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, &user); err != nil {
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
		ID:          user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		CountryCode: user.CountryCode,
		IsValidated: user.IsValidated,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
