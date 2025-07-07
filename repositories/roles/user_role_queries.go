package roles

import (
	"context"
	"fmt"

	"aaa-service/entities/models"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

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

// Deactivate deactivates a user role assignment
func (r *UserRoleRepository) Deactivate(ctx context.Context, id string) error {
	var userRole models.UserRole
	if err := r.dbManager.GetByID(ctx, id, &userRole); err != nil {
		return fmt.Errorf("failed to get user role for deactivation: %w", err)
	}

	userRole.IsActive = false
	return r.Update(ctx, &userRole)
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
