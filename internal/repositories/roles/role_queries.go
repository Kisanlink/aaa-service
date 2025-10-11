package roles

import (
	"context"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// GetByName is implemented in role_repository.go

// ListAll retrieves all roles
func (r *RoleRepository) ListAll(ctx context.Context, limit, offset int) ([]*models.Role, error) {
	return r.List(ctx, limit, offset)
}

// SearchByName searches roles by name
func (r *RoleRepository) SearchByName(ctx context.Context, name string, limit, offset int) ([]*models.Role, error) {
	// For now, we'll use the basic List method and filter in-memory
	// In production, this should be implemented with proper database filtering
	allRoles, err := r.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Filter roles by name
	var filteredRoles []*models.Role
	for _, role := range allRoles {
		if role != nil && strings.Contains(strings.ToLower(role.Name), strings.ToLower(name)) {
			filteredRoles = append(filteredRoles, role)
		}
	}

	return filteredRoles, nil
}

// ExistsByName is implemented in role_repository.go
