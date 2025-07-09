package users

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// UserProfileRepository handles database operations for UserProfile entities
type UserProfileRepository struct {
	dbManager db.DBManager
}

// NewUserProfileRepository creates a new UserProfileRepository instance
func NewUserProfileRepository(dbManager db.DBManager) *UserProfileRepository {
	return &UserProfileRepository{
		dbManager: dbManager,
	}
}

// Create creates a new user profile
func (r *UserProfileRepository) Create(ctx context.Context, profile *models.UserProfile) error {
	if err := profile.BeforeCreate(); err != nil {
		return fmt.Errorf("failed to prepare profile for creation: %w", err)
	}
	return r.dbManager.Create(ctx, profile)
}

// GetByID retrieves a user profile by ID
func (r *UserProfileRepository) GetByID(ctx context.Context, id string) (*models.UserProfile, error) {
	var profile models.UserProfile
	if err := r.dbManager.GetByID(ctx, id, &profile); err != nil {
		return nil, fmt.Errorf("failed to get profile by ID: %w", err)
	}
	return &profile, nil
}

// GetByUserID retrieves a user profile by user ID
func (r *UserProfileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("user_id", db.FilterOpEqual, userID),
	}

	var profiles []models.UserProfile
	if err := r.dbManager.List(ctx, filters, &profiles); err != nil {
		return nil, fmt.Errorf("failed to get profile by user ID: %w", err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found with user ID: %s", userID)
	}

	return &profiles[0], nil
}

// Update updates an existing user profile
func (r *UserProfileRepository) Update(ctx context.Context, profile *models.UserProfile) error {
	if err := profile.BeforeUpdate(); err != nil {
		return fmt.Errorf("failed to prepare profile for update: %w", err)
	}
	return r.dbManager.Update(ctx, profile)
}

// Delete deletes a user profile by ID
func (r *UserProfileRepository) Delete(ctx context.Context, id string) error {
	return r.dbManager.Delete(ctx, id)
}

// List retrieves a list of user profiles with pagination
func (r *UserProfileRepository) List(ctx context.Context, filters []db.Filter, limit, offset int) ([]*models.UserProfile, error) {
	var profiles []models.UserProfile
	if err := r.dbManager.List(ctx, filters, &profiles); err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	// Convert []models.UserProfile to []*models.UserProfile
	results := make([]*models.UserProfile, len(profiles))
	for i, profile := range profiles {
		results[i] = &profile
	}

	return results, nil
}

// ExistsByUserID checks if a user profile exists for the given user ID
func (r *UserProfileRepository) ExistsByUserID(ctx context.Context, userID string) (bool, error) {
	_, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return false, nil // Profile doesn't exist
	}
	return true, nil
}
