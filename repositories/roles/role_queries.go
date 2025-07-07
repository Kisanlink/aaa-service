package roles

import (
	"context"
	"fmt"

	"aaa-service/entities/models"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// GetByName retrieves a role by name
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("name", db.FilterOpEqual, name),
	}

	var roles []models.Role
	if err := r.dbManager.List(ctx, filters, &roles); err != nil {
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("role not found with name: %s", name)
	}

	return &roles[0], nil
}

// ListAll retrieves all roles
func (r *RoleRepository) ListAll(ctx context.Context, limit, offset int) ([]models.Role, error) {
	return r.List(ctx, []db.Filter{}, limit, offset)
}

// SearchByName searches roles by name
func (r *RoleRepository) SearchByName(ctx context.Context, name string, limit, offset int) ([]models.Role, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("name", db.FilterOpContains, name),
	}

	return r.List(ctx, filters, limit, offset)
}

// ExistsByName checks if a role exists by name
func (r *RoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	_, err := r.GetByName(ctx, name)
	if err != nil {
		return false, nil // Role doesn't exist
	}
	return true, nil
}
