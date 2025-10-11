package permissions

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// PermissionRepository handles database operations for Permission entities
type PermissionRepository struct {
	*base.BaseFilterableRepository[*models.Permission]
	dbManager db.DBManager
}

// NewPermissionRepository creates a new PermissionRepository instance
func NewPermissionRepository(dbManager db.DBManager) *PermissionRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.Permission]()
	baseRepo.SetDBManager(dbManager)
	return &PermissionRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Create creates a new permission in the database
func (r *PermissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	return r.BaseFilterableRepository.Create(ctx, permission)
}

// GetByID retrieves a permission by ID
func (r *PermissionRepository) GetByID(ctx context.Context, id string) (*models.Permission, error) {
	permission := &models.Permission{}
	return r.BaseFilterableRepository.GetByID(ctx, id, permission)
}

// Update updates an existing permission
func (r *PermissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	return r.BaseFilterableRepository.Update(ctx, permission)
}

// Delete deletes a permission by ID (hard delete)
func (r *PermissionRepository) Delete(ctx context.Context, id string) error {
	permission := &models.Permission{}
	return r.BaseFilterableRepository.Delete(ctx, id, permission)
}

// SoftDelete soft deletes a permission by ID
func (r *PermissionRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted permission
func (r *PermissionRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// List retrieves permissions with pagination
func (r *PermissionRepository) List(ctx context.Context, limit, offset int) ([]*models.Permission, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of permissions
func (r *PermissionRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.Count(ctx, filter, models.Permission{})
}

// Exists checks if a permission exists by ID
func (r *PermissionRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// ListWithDeleted retrieves permissions including soft-deleted ones
func (r *PermissionRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.Permission, error) {
	return r.BaseFilterableRepository.ListWithDeleted(ctx, limit, offset)
}

// CountWithDeleted returns count including soft-deleted permissions
func (r *PermissionRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx, &models.Permission{})
}

// ExistsWithDeleted checks if permission exists including soft-deleted ones
func (r *PermissionRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.ExistsWithDeleted(ctx, id)
}

// GetByCreatedBy gets permissions by creator
func (r *PermissionRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Permission, error) {
	return r.BaseFilterableRepository.GetByCreatedBy(ctx, createdBy, limit, offset)
}

// GetByUpdatedBy gets permissions by updater
func (r *PermissionRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.Permission, error) {
	return r.BaseFilterableRepository.GetByUpdatedBy(ctx, updatedBy, limit, offset)
}
