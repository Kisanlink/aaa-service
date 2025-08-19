package roles

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// UserRoleRepository handles database operations for UserRole entities
type UserRoleRepository struct {
	*base.BaseFilterableRepository[*models.UserRole]
	dbManager db.DBManager
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(dbManager db.DBManager) *UserRoleRepository {
	repo := &UserRoleRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*models.UserRole](),
		dbManager:                dbManager,
	}

	// Set the database manager on the BaseFilterableRepository so it can use database-level operations
	repo.BaseFilterableRepository.SetDBManager(dbManager)

	return repo
}

// Create creates a new user role
func (r *UserRoleRepository) Create(ctx context.Context, userRole *models.UserRole) error {
	if err := userRole.BeforeCreate(); err != nil {
		return fmt.Errorf("failed to prepare user role for creation: %w", err)
	}

	return r.BaseFilterableRepository.Create(ctx, userRole)
}

// GetByID retrieves a user role by ID
func (r *UserRoleRepository) GetByID(ctx context.Context, id string, userRole *models.UserRole) (*models.UserRole, error) {
	return r.BaseFilterableRepository.GetByID(ctx, id, userRole)
}

// Update updates an existing user role
func (r *UserRoleRepository) Update(ctx context.Context, userRole *models.UserRole) error {
	if err := userRole.BeforeUpdate(); err != nil {
		return fmt.Errorf("failed to prepare user role for update: %w", err)
	}

	return r.BaseFilterableRepository.Update(ctx, userRole)
}

// Delete deletes a user role by ID
func (r *UserRoleRepository) Delete(ctx context.Context, id string, userRole *models.UserRole) error {
	return r.BaseFilterableRepository.Delete(ctx, id, userRole)
}

// List retrieves user roles with pagination
func (r *UserRoleRepository) List(ctx context.Context, limit, offset int) ([]*models.UserRole, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of user roles
func (r *UserRoleRepository) Count(ctx context.Context) (int64, error) {
	// Create an empty filter to count all user roles
	filter := base.NewFilter()
	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// GetByUserID retrieves all roles for a user using base filterable repository
func (r *UserRoleRepository) GetByUserID(ctx context.Context, userID string) ([]*models.UserRole, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByRoleID retrieves all users for a role using database-level filtering
func (r *UserRoleRepository) GetByRoleID(ctx context.Context, roleID string) ([]*models.UserRole, error) {
	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUserAndRole retrieves a specific user role assignment using database-level filtering
func (r *UserRoleRepository) GetByUserAndRole(ctx context.Context, userID, roleID string) (*models.UserRole, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	userRoles, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	if len(userRoles) == 0 {
		return nil, fmt.Errorf("user role not found")
	}

	return userRoles[0], nil
}

// DeleteByUserAndRole deletes a user role assignment by user ID and role ID
func (r *UserRoleRepository) DeleteByUserAndRole(ctx context.Context, userID, roleID string) error {
	// First find the user role
	userRole, err := r.GetByUserAndRole(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to find user role for deletion: %w", err)
	}

	// Then delete it
	return r.Delete(ctx, userRole.ID, userRole)
}

// GetActiveByUserID retrieves all active roles for a user
func (r *UserRoleRepository) GetActiveByUserID(ctx context.Context, userID string) ([]*models.UserRole, error) {
	// This is the same as GetByUserID since we already filter by is_active = true
	return r.GetByUserID(ctx, userID)
}

// Deactivate deactivates a user role assignment
func (r *UserRoleRepository) Deactivate(ctx context.Context, id string) error {
	var userRole models.UserRole
	if err := r.dbManager.GetByID(ctx, id, &userRole); err != nil {
		return fmt.Errorf("failed to get user role for deactivation: %w", err)
	}

	userRole.IsActive = false
	return r.dbManager.Update(ctx, &userRole)
}

// Exists checks if a user role exists by ID
func (r *UserRoleRepository) Exists(ctx context.Context, id string) (bool, error) {
	userRole := &models.UserRole{}
	_, err := r.GetByID(ctx, id, userRole)
	if err != nil {
		return false, nil // User role doesn't exist
	}
	return true, nil
}

// ExistsByUserAndRole checks if a user role assignment exists
func (r *UserRoleRepository) ExistsByUserAndRole(ctx context.Context, userID, roleID string) (bool, error) {
	_, err := r.GetByUserAndRole(ctx, userID, roleID)
	if err != nil {
		return false, nil // User role doesn't exist
	}
	return true, nil
}
