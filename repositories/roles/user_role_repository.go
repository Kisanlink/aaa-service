package roles

import (
	"context"
	"fmt"

	"aaa-service/entities/models"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// UserRoleRepository handles database operations for UserRole entities
type UserRoleRepository struct {
	*base.BaseRepository[*models.UserRole]
	dbManager db.DBManager
}

// NewUserRoleRepository creates a new UserRoleRepository
func NewUserRoleRepository(dbManager db.DBManager) *UserRoleRepository {
	return &UserRoleRepository{
		BaseRepository: base.NewBaseRepository[*models.UserRole](),
		dbManager:      dbManager,
	}
}

// Create creates a new user role
func (r *UserRoleRepository) Create(ctx context.Context, userRole *models.UserRole) error {
	if err := userRole.BeforeCreate(); err != nil {
		return fmt.Errorf("failed to prepare user role for creation: %w", err)
	}

	return r.dbManager.Create(ctx, userRole)
}

// GetByID retrieves a user role by ID
func (r *UserRoleRepository) GetByID(ctx context.Context, id string) (*models.UserRole, error) {
	var userRole models.UserRole
	if err := r.dbManager.GetByID(ctx, id, &userRole); err != nil {
		return nil, fmt.Errorf("failed to get user role by ID: %w", err)
	}
	return &userRole, nil
}

// GetByUserID retrieves all roles for a user
func (r *UserRoleRepository) GetByUserID(ctx context.Context, userID string) ([]models.UserRole, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("user_id", db.FilterOpEqual, userID),
		r.dbManager.BuildFilter("is_active", db.FilterOpEqual, true),
	}

	var userRoles []models.UserRole
	if err := r.dbManager.List(ctx, filters, &userRoles); err != nil {
		return nil, fmt.Errorf("failed to get user roles by user ID: %w", err)
	}

	return userRoles, nil
}

// GetByRoleID retrieves all users for a role
func (r *UserRoleRepository) GetByRoleID(ctx context.Context, roleID string) ([]models.UserRole, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("role_id", db.FilterOpEqual, roleID),
		r.dbManager.BuildFilter("is_active", db.FilterOpEqual, true),
	}

	var userRoles []models.UserRole
	if err := r.dbManager.List(ctx, filters, &userRoles); err != nil {
		return nil, fmt.Errorf("failed to get user roles by role ID: %w", err)
	}

	return userRoles, nil
}

// Update updates an existing user role
func (r *UserRoleRepository) Update(ctx context.Context, userRole *models.UserRole) error {
	if err := userRole.BeforeUpdate(); err != nil {
		return fmt.Errorf("failed to prepare user role for update: %w", err)
	}

	return r.dbManager.Update(ctx, userRole)
}

// Delete deletes a user role by ID
func (r *UserRoleRepository) Delete(ctx context.Context, id string) error {
	return r.dbManager.Delete(ctx, id)
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
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return false, nil // User role doesn't exist
	}
	return true, nil
}

// ExistsByUserAndRole checks if a user role assignment exists
func (r *UserRoleRepository) ExistsByUserAndRole(ctx context.Context, userID, roleID string) (bool, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("user_id", db.FilterOpEqual, userID),
		r.dbManager.BuildFilter("role_id", db.FilterOpEqual, roleID),
		r.dbManager.BuildFilter("is_active", db.FilterOpEqual, true),
	}

	var userRoles []models.UserRole
	if err := r.dbManager.List(ctx, filters, &userRoles); err != nil {
		return false, fmt.Errorf("failed to check user role existence: %w", err)
	}

	return len(userRoles) > 0, nil
}

// List retrieves user roles with optional filters
func (r *UserRoleRepository) List(ctx context.Context, filters []db.Filter, limit, offset int) ([]models.UserRole, error) {
	var userRoles []models.UserRole

	if err := r.dbManager.List(ctx, filters, &userRoles); err != nil {
		return nil, fmt.Errorf("failed to list user roles: %w", err)
	}

	return userRoles, nil
}
