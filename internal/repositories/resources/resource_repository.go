package resources

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// ResourceRepository handles database operations for Resource entities
type ResourceRepository struct {
	*base.BaseFilterableRepository[*models.Resource]
	dbManager db.DBManager
}

// NewResourceRepository creates a new ResourceRepository instance
func NewResourceRepository(dbManager db.DBManager) *ResourceRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.Resource]()
	baseRepo.SetDBManager(dbManager)
	return &ResourceRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Create creates a new resource using the base repository
func (r *ResourceRepository) Create(ctx context.Context, resource *models.Resource) error {
	return r.BaseFilterableRepository.Create(ctx, resource)
}

// GetByID retrieves a resource by ID using the base repository
func (r *ResourceRepository) GetByID(ctx context.Context, id string) (*models.Resource, error) {
	resource := &models.Resource{}
	return r.BaseFilterableRepository.GetByID(ctx, id, resource)
}

// Update updates an existing resource using the base repository
func (r *ResourceRepository) Update(ctx context.Context, resource *models.Resource) error {
	return r.BaseFilterableRepository.Update(ctx, resource)
}

// Delete deletes a resource by ID using the base repository
func (r *ResourceRepository) Delete(ctx context.Context, id string) error {
	resource := &models.Resource{}
	return r.BaseFilterableRepository.Delete(ctx, id, resource)
}

// List retrieves resources with pagination using database-level filtering
func (r *ResourceRepository) List(ctx context.Context, limit, offset int) ([]*models.Resource, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of resources using database-level counting
func (r *ResourceRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// Exists checks if a resource exists by ID using the base repository
func (r *ResourceRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// SoftDelete soft deletes a resource by ID using the base repository
func (r *ResourceRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted resource using the base repository
func (r *ResourceRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// ListWithDeleted retrieves resources including soft-deleted ones using the base repository
func (r *ResourceRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.Resource, error) {
	return r.BaseFilterableRepository.ListWithDeleted(ctx, limit, offset)
}

// CountWithDeleted returns count including soft-deleted resources using the base repository
func (r *ResourceRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx)
}

// ExistsWithDeleted checks if resource exists including soft-deleted ones using the base repository
func (r *ResourceRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.ExistsWithDeleted(ctx, id)
}

// GetByCreatedBy gets resources by creator using the base repository
func (r *ResourceRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Resource, error) {
	return r.BaseFilterableRepository.GetByCreatedBy(ctx, createdBy, limit, offset)
}

// GetByUpdatedBy gets resources by updater using the base repository
func (r *ResourceRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.Resource, error) {
	return r.BaseFilterableRepository.GetByUpdatedBy(ctx, updatedBy, limit, offset)
}

// GetByName retrieves a resource by name
func (r *ResourceRepository) GetByName(ctx context.Context, name string) (*models.Resource, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Build()

	resources, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource by name: %w", err)
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf("resource not found with name: %s", name)
	}

	return resources[0], nil
}

// GetByServiceName retrieves resources by service name
func (r *ResourceRepository) GetByServiceName(ctx context.Context, serviceName string, limit, offset int) ([]*models.Resource, error) {
	filter := base.NewFilterBuilder().
		Where("service_name", base.OpEqual, serviceName).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByType retrieves resources by type
func (r *ResourceRepository) GetByType(ctx context.Context, resourceType string, limit, offset int) ([]*models.Resource, error) {
	filter := base.NewFilterBuilder().
		Where("type", base.OpEqual, resourceType).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
