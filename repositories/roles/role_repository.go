package roles

import (
	"context"
	"errors"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// RoleRepository handles database operations for Role entities
type RoleRepository struct {
	base.Repository[models.Role]
	dbManager *db.Manager
}

// NewRoleRepository creates a new RoleRepository instance
func NewRoleRepository(dbManager *db.Manager) *RoleRepository {
	return &RoleRepository{
		Repository: base.NewRepository[models.Role](dbManager),
		dbManager:  dbManager,
	}
}

// Create creates a new role
func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.Repository.Create(ctx, role)
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(ctx context.Context, id string) (*models.Role, error) {
	var role models.Role
	err := r.dbManager.GetDB().WithContext(ctx).
		Where("id = ?", id).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, base.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

// GetByName retrieves a role by name
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.dbManager.GetDB().WithContext(ctx).
		Where("name = ?", name).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, base.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

// Update updates an existing role
func (r *RoleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.Repository.Update(ctx, role)
}

// Delete deletes a role by ID
func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	return r.Repository.Delete(ctx, id)
}

// List retrieves a list of roles with pagination
func (r *RoleRepository) List(ctx context.Context, filters *base.Filters) ([]*models.Role, error) {
	var roles []*models.Role
	query := r.dbManager.GetDB().WithContext(ctx)

	if filters != nil {
		query = filters.Apply(query)
	}

	err := query.Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// ExistsByName checks if a role exists with the given name
func (r *RoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.dbManager.GetDB().WithContext(ctx).
		Model(&models.Role{}).
		Where("name = ?", name).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
