package user_profiles

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
	baseRepo := base.NewBaseFilterableRepository[*models.UserProfile]()
	baseRepo.SetDBManager(dbManager)
	return &UserProfileRepository{
		BaseFilterableRepository: baseRepo,
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

// Update updates an existing user profile using the base repository
func (r *UserProfileRepository) Update(ctx context.Context, profile *models.UserProfile) error {
	return r.BaseFilterableRepository.Update(ctx, profile)
}

// Delete deletes a user profile by ID using the base repository
func (r *UserProfileRepository) Delete(ctx context.Context, id string) error {
	profile := &models.UserProfile{}
	return r.BaseFilterableRepository.Delete(ctx, id, profile)
}

// List retrieves user profiles with pagination using database-level filtering
func (r *UserProfileRepository) List(ctx context.Context, limit, offset int) ([]*models.UserProfile, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of user profiles using database-level counting
func (r *UserProfileRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.Count(ctx, filter, models.UserProfile{})
}

// Exists checks if a user profile exists by ID using the base repository
func (r *UserProfileRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// SoftDelete soft deletes a user profile by ID using the base repository
func (r *UserProfileRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted user profile using the base repository
func (r *UserProfileRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// ListWithDeleted retrieves user profiles including soft-deleted ones using the base repository
func (r *UserProfileRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.UserProfile, error) {
	return r.BaseFilterableRepository.ListWithDeleted(ctx, limit, offset)
}

// CountWithDeleted returns count including soft-deleted user profiles using the base repository
func (r *UserProfileRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx, &models.UserProfile{})
}

// ExistsWithDeleted checks if user profile exists including soft-deleted ones using the base repository
func (r *UserProfileRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.ExistsWithDeleted(ctx, id)
}

// GetByCreatedBy gets user profiles by creator using the base repository
func (r *UserProfileRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.UserProfile, error) {
	return r.BaseFilterableRepository.GetByCreatedBy(ctx, createdBy, limit, offset)
}

// GetByUpdatedBy gets user profiles by updater using the base repository
func (r *UserProfileRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.UserProfile, error) {
	return r.BaseFilterableRepository.GetByUpdatedBy(ctx, updatedBy, limit, offset)
}

// GetByUserID retrieves a user profile by user ID
func (r *UserProfileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Build()

	profiles, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile by user ID: %w", err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("user profile not found with user ID: %s", userID)
	}

	return profiles[0], nil
}

// GetByAadhaarNumber retrieves a user profile by Aadhaar number
func (r *UserProfileRepository) GetByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*models.UserProfile, error) {
	filter := base.NewFilterBuilder().
		Where("aadhaar_number", base.OpEqual, aadhaarNumber).
		Build()

	profiles, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile by Aadhaar number: %w", err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("user profile not found with Aadhaar number: %s", aadhaarNumber)
	}

	return profiles[0], nil
}
