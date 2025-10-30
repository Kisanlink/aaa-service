package roles

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
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
	return r.BaseFilterableRepository.Count(ctx, filter, models.UserRole{})
}

// CountWithDeleted returns count including soft-deleted user roles
func (r *UserRoleRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx, &models.UserRole{})
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

// GetActiveRolesByUserID retrieves all active user roles with role details preloaded
func (r *UserRoleRepository) GetActiveRolesByUserID(ctx context.Context, userID string) ([]*models.UserRole, error) {
	// Get the GORM DB instance for complex queries with preloading
	db, err := r.getDB(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	var userRoles []models.UserRole
	err = db.WithContext(ctx).
		Preload("Role").
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&userRoles).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get active roles for user %s: %w", userID, err)
	}

	// Filter out roles that are not active and convert to pointers
	activeUserRoles := make([]*models.UserRole, 0, len(userRoles))
	for i := range userRoles {
		if userRoles[i].Role.IsActive {
			activeUserRoles = append(activeUserRoles, &userRoles[i])
		}
	}

	return activeUserRoles, nil
}

// AssignRole creates a new user role assignment with transaction support and validation
func (r *UserRoleRepository) AssignRole(ctx context.Context, userID, roleID string) error {
	// Check if role is already assigned
	exists, err := r.IsRoleAssigned(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to check existing role assignment: %w", err)
	}
	if exists {
		return fmt.Errorf("role %s is already assigned to user %s", roleID, userID)
	}

	// Create new user role assignment
	userRole := models.NewUserRole(userID, roleID)

	// Use GORM transaction for consistency
	db, err := r.getDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(userRole).Error; err != nil {
			return fmt.Errorf("failed to create user role assignment: %w", err)
		}
		return nil
	})
}

// RemoveRole removes a user role assignment with proper constraint handling
func (r *UserRoleRepository) RemoveRole(ctx context.Context, userID, roleID string) error {
	// Find the user role assignment
	userRole, err := r.GetByUserAndRole(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("user role assignment not found for user %s and role %s: %w", userID, roleID, err)
	}

	// Use GORM transaction for consistency
	db, err := r.getDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Soft delete by setting is_active to false instead of hard delete
		// This maintains audit trail and handles foreign key constraints
		userRole.IsActive = false
		if err := tx.Save(userRole).Error; err != nil {
			return fmt.Errorf("failed to deactivate user role assignment: %w", err)
		}
		return nil
	})
}

// IsRoleAssigned checks if a role is currently assigned to a user (active assignment)
func (r *UserRoleRepository) IsRoleAssigned(ctx context.Context, userID, roleID string) (bool, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	userRoles, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check role assignment: %w", err)
	}

	return len(userRoles) > 0, nil
}

// getDB is a helper method to get the database connection from the database manager
func (r *UserRoleRepository) getDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
	// Try to get the database from the database manager
	if postgresMgr, ok := r.dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		return postgresMgr.GetDB(ctx, readOnly)
	}

	return nil, fmt.Errorf("database manager does not support GetDB method")
}

// DeleteBySourceGroup deletes all user_roles inherited from a specific group
// Returns the number of records deleted
func (r *UserRoleRepository) DeleteBySourceGroup(ctx context.Context, userID, groupID string) (int, error) {
	db, err := r.getDB(ctx, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get database: %w", err)
	}

	result := db.Where("user_id = ? AND source_group_id = ?", userID, groupID).Delete(&models.UserRole{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete inherited roles: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}
