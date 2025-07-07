package users

import (
	"context"
	"errors"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// UserProfileRepository handles database operations for UserProfile entities
type UserProfileRepository struct {
	base.Repository[models.UserProfile]
	dbManager *db.Manager
}

// NewUserProfileRepository creates a new UserProfileRepository instance
func NewUserProfileRepository(dbManager *db.Manager) *UserProfileRepository {
	return &UserProfileRepository{
		Repository: base.NewRepository[models.UserProfile](dbManager),
		dbManager:  dbManager,
	}
}

// Create creates a new user profile
func (r *UserProfileRepository) Create(ctx context.Context, profile *models.UserProfile) error {
	return r.Repository.Create(ctx, profile)
}

// GetByID retrieves a user profile by ID
func (r *UserProfileRepository) GetByID(ctx context.Context, id string) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.dbManager.GetDB().WithContext(ctx).
		Preload("Address").
		Where("id = ?", id).
		First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, base.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

// GetByUserID retrieves a user profile by user ID
func (r *UserProfileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.dbManager.GetDB().WithContext(ctx).
		Preload("Address").
		Where("user_id = ?", userID).
		First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, base.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

// Update updates an existing user profile
func (r *UserProfileRepository) Update(ctx context.Context, profile *models.UserProfile) error {
	return r.Repository.Update(ctx, profile)
}

// Delete deletes a user profile by ID
func (r *UserProfileRepository) Delete(ctx context.Context, id string) error {
	return r.Repository.Delete(ctx, id)
}

// List retrieves a list of user profiles with pagination
func (r *UserProfileRepository) List(ctx context.Context, filters *base.Filters) ([]*models.UserProfile, error) {
	var profiles []*models.UserProfile
	query := r.dbManager.GetDB().WithContext(ctx).Preload("Address")

	if filters != nil {
		query = filters.Apply(query)
	}

	err := query.Find(&profiles).Error
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

// ExistsByUserID checks if a user profile exists for the given user ID
func (r *UserProfileRepository) ExistsByUserID(ctx context.Context, userID string) (bool, error) {
	var count int64
	err := r.dbManager.GetDB().WithContext(ctx).
		Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
