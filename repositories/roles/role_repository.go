package roles

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// RoleRepository handles role-related database operations
type RoleRepository struct {
	*base.BaseFilterableRepository[*models.Role]
	dbManager db.DBManager
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(dbManager db.DBManager) *RoleRepository {
	return &RoleRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*models.Role](),
		dbManager:                dbManager,
	}
}

// Create creates a new role in the database
func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.dbManager.Create(ctx, role)
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(ctx context.Context, id string) (*models.Role, error) {
	role := &models.Role{}
	err := r.dbManager.GetByID(ctx, id, role)
	if err != nil {
		return nil, err
	}
	return role, nil
}

// Update updates an existing role
func (r *RoleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.dbManager.Update(ctx, role)
}

// Delete deletes a role by ID
func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	return r.dbManager.Delete(ctx, id)
}

// List retrieves all roles with pagination
func (r *RoleRepository) List(ctx context.Context, limit, offset int) ([]*models.Role, error) {
	var roles []models.Role
	err := r.dbManager.List(ctx, []db.Filter{}, &roles)
	if err != nil {
		return nil, err
	}

	// Convert []models.Role to []*models.Role
	result := make([]*models.Role, len(roles))
	for i := range roles {
		result[i] = &roles[i]
	}

	return result, nil
}

// Count returns the total number of roles
func (r *RoleRepository) Count(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.Count(ctx)
}

// GetByName retrieves a role by name
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("name", db.FilterOpEqual, name),
	}

	var roles []models.Role
	err := r.dbManager.List(ctx, filters, &roles)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("role not found")
	}

	return &roles[0], nil
}

// GetActive retrieves all active roles with pagination
func (r *RoleRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.Role, error) {
	// For now, we'll consider all non-deleted roles as active
	// In the future, you might want to add an "active" field to the Role model
	filters := []db.Filter{
		// Add any active filters here if Role model has an active field
	}

	var roles []models.Role
	err := r.dbManager.List(ctx, filters, &roles)
	if err != nil {
		return nil, fmt.Errorf("failed to get active roles: %w", err)
	}

	// Convert []models.Role to []*models.Role
	result := make([]*models.Role, len(roles))
	for i := range roles {
		result[i] = &roles[i]
	}

	return result, nil
}

// Search searches roles by keyword
func (r *RoleRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Role, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("name", db.FilterOpContains, query),
	}

	var roles []models.Role
	err := r.dbManager.List(ctx, filters, &roles)
	if err != nil {
		return nil, fmt.Errorf("failed to search roles: %w", err)
	}

	// Convert []models.Role to []*models.Role
	result := make([]*models.Role, len(roles))
	for i := range roles {
		result[i] = &roles[i]
	}

	return result, nil
}

// ExistsByName checks if a role exists with the given name
func (r *RoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	_, err := r.GetByName(ctx, name)
	if err != nil {
		return false, nil // Role doesn't exist
	}
	return true, nil
}
