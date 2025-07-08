package users

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// UserTokenRepository handles token-related operations for User entities
type UserTokenRepository struct {
	dbManager db.DBManager
}

// NewUserTokenRepository creates a new UserTokenRepository
func NewUserTokenRepository(dbManager db.DBManager) *UserTokenRepository {
	return &UserTokenRepository{
		dbManager: dbManager,
	}
}

// UpdateTokens updates user tokens
func (r *UserTokenRepository) UpdateTokens(ctx context.Context, userID string, tokens int) error {
	var user models.User
	if err := r.dbManager.GetByID(ctx, userID, &user); err != nil {
		return fmt.Errorf("failed to get user for token update: %w", err)
	}

	user.Tokens = tokens
	return r.dbManager.Update(ctx, &user)
}

// DeductTokens deducts tokens from user account
func (r *UserTokenRepository) DeductTokens(ctx context.Context, userID string, amount int) error {
	var user models.User
	if err := r.dbManager.GetByID(ctx, userID, &user); err != nil {
		return fmt.Errorf("failed to get user for token deduction: %w", err)
	}

	if !user.DeductTokens(amount) {
		return fmt.Errorf("insufficient tokens for user %s", userID)
	}

	return r.dbManager.Update(ctx, &user)
}

// AddTokens adds tokens to user account
func (r *UserTokenRepository) AddTokens(ctx context.Context, userID string, amount int) error {
	var user models.User
	if err := r.dbManager.GetByID(ctx, userID, &user); err != nil {
		return fmt.Errorf("failed to get user for token addition: %w", err)
	}

	user.AddTokens(amount)
	return r.dbManager.Update(ctx, &user)
}

// ValidateUser validates a user's Aadhaar
func (r *UserTokenRepository) ValidateUser(ctx context.Context, userID string) error {
	var user models.User
	if err := r.dbManager.GetByID(ctx, userID, &user); err != nil {
		return fmt.Errorf("failed to get user for validation: %w", err)
	}

	user.ValidateAadhaar()
	return r.dbManager.Update(ctx, &user)
}
