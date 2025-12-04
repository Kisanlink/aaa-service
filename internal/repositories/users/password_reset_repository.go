package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// PasswordResetRepository handles password reset token operations
type PasswordResetRepository struct {
	dbManager db.DBManager
	userRepo  *UserRepository
}

// NewPasswordResetRepository creates a new PasswordResetRepository
func NewPasswordResetRepository(dbManager db.DBManager, userRepo *UserRepository) *PasswordResetRepository {
	return &PasswordResetRepository{
		dbManager: dbManager,
		userRepo:  userRepo,
	}
}

// GenerateToken generates a secure random token
func (r *PasswordResetRepository) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateResetToken creates a new password reset token with auto-generated value
func (r *PasswordResetRepository) CreateResetToken(ctx context.Context, userID string, expiryDuration time.Duration) (*models.PasswordResetToken, error) {
	// Generate token
	token, err := r.GenerateToken()
	if err != nil {
		return nil, err
	}

	return r.CreateResetTokenWithValue(ctx, userID, token, expiryDuration)
}

// CreateResetTokenWithValue creates a new password reset token with a provided value
// This is useful for storing hashed OTPs instead of randomly generated tokens
func (r *PasswordResetRepository) CreateResetTokenWithValue(ctx context.Context, userID, tokenValue string, expiryDuration time.Duration) (*models.PasswordResetToken, error) {
	// Create reset token model
	resetToken := &models.PasswordResetToken{
		UserID:    userID,
		Token:     tokenValue,
		ExpiresAt: time.Now().Add(expiryDuration),
		Used:      false,
	}

	// Save to database
	if err := r.dbManager.Create(ctx, resetToken); err != nil {
		return nil, fmt.Errorf("failed to create password reset token: %w", err)
	}

	return resetToken, nil
}

// GetTokenByValue retrieves a reset token by its value
func (r *PasswordResetRepository) GetTokenByValue(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken

	// Use the userRepo's getDB method to get database connection
	db, err := r.userRepo.getDB(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	err = db.WithContext(ctx).
		Where("token = ?", token).
		First(&resetToken).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to get reset token: %w", err)
	}

	return &resetToken, nil
}

// GetTokenByID retrieves a reset token by its ID
func (r *PasswordResetRepository) GetTokenByID(ctx context.Context, tokenID string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken

	// Use the userRepo's getDB method to get database connection
	db, err := r.userRepo.getDB(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	err = db.WithContext(ctx).
		Where("id = ?", tokenID).
		First(&resetToken).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to get reset token: %w", err)
	}

	return &resetToken, nil
}

// MarkTokenAsUsed marks a token as used
func (r *PasswordResetRepository) MarkTokenAsUsed(ctx context.Context, tokenID string) error {
	var resetToken models.PasswordResetToken
	if err := r.dbManager.GetByID(ctx, tokenID, &resetToken); err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	resetToken.MarkAsUsed()

	if err := r.dbManager.Update(ctx, &resetToken); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// InvalidateUserTokens invalidates all existing tokens for a user
func (r *PasswordResetRepository) InvalidateUserTokens(ctx context.Context, userID string) error {
	// Use the userRepo's getDB method to get database connection
	db, err := r.userRepo.getDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	err = db.WithContext(ctx).
		Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND used = ?", userID, false).
		Update("used", true).Error

	if err != nil {
		return fmt.Errorf("failed to invalidate user tokens: %w", err)
	}

	return nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (r *PasswordResetRepository) CleanupExpiredTokens(ctx context.Context) error {
	// Use the userRepo's getDB method to get database connection
	db, err := r.userRepo.getDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	err = db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.PasswordResetToken{}).Error

	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return nil
}
