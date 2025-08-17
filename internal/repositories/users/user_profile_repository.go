package users

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// UserProfileRepository handles database operations for UserProfile entities
type UserProfileRepository struct {
	*base.BaseFilterableRepository[*models.UserProfile]
	dbManager db.DBManager
}

// NewUserProfileRepository creates a new UserProfileRepository instance
func NewUserProfileRepository(dbManager db.DBManager) *UserProfileRepository {
	return &UserProfileRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*models.UserProfile](),
		dbManager:                dbManager,
	}
}

// Create creates a new user profile using the base repository
func (r *UserProfileRepository) Create(ctx context.Context, profile *models.UserProfile) error {
	return r.BaseFilterableRepository.Create(ctx, profile)
}

// GetByID retrieves a user profile by ID using the base repository
func (r *UserProfileRepository) GetByID(ctx context.Context, id string) (*models.UserProfile, error) {
	profile := &models.UserProfile{}
	return r.BaseFilterableRepository.GetByID(ctx, id, profile)
}

// GetByUserID retrieves a user profile by user ID using database-level filtering
func (r *UserProfileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Build()

	profiles, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile by user ID: %w", err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found with user ID: %s", userID)
	}

	return profiles[0], nil
}

// Update updates an existing user profile using the base repository
func (r *UserProfileRepository) Update(ctx context.Context, profile *models.UserProfile) error {
	return r.BaseFilterableRepository.Update(ctx, profile)
}

// Delete deletes a user profile by ID using the base repository
func (r *UserProfileRepository) Delete(ctx context.Context, id string) error {
	profile := &models.UserProfile{}
	return r.BaseFilterableRepository.Delete(ctx, id, profile)
}

// List retrieves a list of user profiles with pagination using database-level filtering
func (r *UserProfileRepository) List(ctx context.Context, limit, offset int) ([]*models.UserProfile, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// ExistsByUserID checks if a user profile exists for the given user ID
func (r *UserProfileRepository) ExistsByUserID(ctx context.Context, userID string) (bool, error) {
	_, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return false, nil // Profile doesn't exist
	}
	return true, nil
}
