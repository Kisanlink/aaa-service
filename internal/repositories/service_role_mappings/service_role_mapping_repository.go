package service_role_mappings

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// ServiceRoleMappingRepository handles service_role_mapping database operations
type ServiceRoleMappingRepository struct {
	*base.BaseFilterableRepository[*models.ServiceRoleMapping]
	dbManager db.DBManager
}

// NewServiceRoleMappingRepository creates a new service role mapping repository
func NewServiceRoleMappingRepository(dbManager db.DBManager) *ServiceRoleMappingRepository {
	repo := &ServiceRoleMappingRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*models.ServiceRoleMapping](),
		dbManager:                dbManager,
	}

	repo.BaseFilterableRepository.SetDBManager(dbManager)
	return repo
}

// Create creates a new service role mapping
func (r *ServiceRoleMappingRepository) Create(ctx context.Context, mapping *models.ServiceRoleMapping) error {
	return r.BaseFilterableRepository.Create(ctx, mapping)
}

// GetByID retrieves a service role mapping by ID
func (r *ServiceRoleMappingRepository) GetByID(ctx context.Context, id string) (*models.ServiceRoleMapping, error) {
	mapping := &models.ServiceRoleMapping{}
	return r.BaseFilterableRepository.GetByID(ctx, id, mapping)
}

// Update updates an existing service role mapping
func (r *ServiceRoleMappingRepository) Update(ctx context.Context, mapping *models.ServiceRoleMapping) error {
	return r.BaseFilterableRepository.Update(ctx, mapping)
}

// Delete hard deletes a service role mapping
func (r *ServiceRoleMappingRepository) Delete(ctx context.Context, id string) error {
	mapping := &models.ServiceRoleMapping{}
	return r.BaseFilterableRepository.Delete(ctx, id, mapping)
}

// GetByServiceID retrieves all mappings for a specific service
func (r *ServiceRoleMappingRepository) GetByServiceID(ctx context.Context, serviceID string) ([]*models.ServiceRoleMapping, error) {
	filter := base.NewFilterBuilder().
		Where("service_id", base.OpEqual, serviceID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByRoleID retrieves all mappings for a specific role
func (r *ServiceRoleMappingRepository) GetByRoleID(ctx context.Context, roleID string) ([]*models.ServiceRoleMapping, error) {
	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByServiceAndRole retrieves a specific service-role mapping
func (r *ServiceRoleMappingRepository) GetByServiceAndRole(ctx context.Context, serviceID, roleID string) (*models.ServiceRoleMapping, error) {
	filter := base.NewFilterBuilder().
		Where("service_id", base.OpEqual, serviceID).
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	mappings, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get service-role mapping: %w", err)
	}

	if len(mappings) == 0 {
		return nil, fmt.Errorf("service-role mapping not found")
	}

	return mappings[0], nil
}

// ExistsByServiceAndRole checks if a service-role mapping exists
func (r *ServiceRoleMappingRepository) ExistsByServiceAndRole(ctx context.Context, serviceID, roleID string) (bool, error) {
	filter := base.NewFilterBuilder().
		Where("service_id", base.OpEqual, serviceID).
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	mappings, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return false, err
	}

	return len(mappings) > 0, nil
}

// GetByServiceName retrieves all mappings for a specific service name
func (r *ServiceRoleMappingRepository) GetByServiceName(ctx context.Context, serviceName string) ([]*models.ServiceRoleMapping, error) {
	filter := base.NewFilterBuilder().
		Where("service_name", base.OpEqual, serviceName).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// DeactivateByServiceID deactivates all mappings for a specific service
func (r *ServiceRoleMappingRepository) DeactivateByServiceID(ctx context.Context, serviceID string) error {
	mappings, err := r.GetByServiceID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get mappings for service: %w", err)
	}

	for _, mapping := range mappings {
		mapping.Deactivate()
		if err := r.Update(ctx, mapping); err != nil {
			return fmt.Errorf("failed to deactivate mapping: %w", err)
		}
	}

	return nil
}

// DeleteByServiceID deletes all mappings for a specific service
func (r *ServiceRoleMappingRepository) DeleteByServiceID(ctx context.Context, serviceID string) error {
	mappings, err := r.GetByServiceID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get mappings for service: %w", err)
	}

	for _, mapping := range mappings {
		if err := r.Delete(ctx, mapping.ID); err != nil {
			return fmt.Errorf("failed to delete mapping: %w", err)
		}
	}

	return nil
}

// List retrieves all active service role mappings with pagination
func (r *ServiceRoleMappingRepository) List(ctx context.Context, limit, offset int) ([]*models.ServiceRoleMapping, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of active service role mappings
func (r *ServiceRoleMappingRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Count(ctx, filter, models.ServiceRoleMapping{})
}

// UpsertMapping creates or updates a service-role mapping idempotently
func (r *ServiceRoleMappingRepository) UpsertMapping(ctx context.Context, serviceID, serviceName, roleID string) (*models.ServiceRoleMapping, error) {
	// Check if mapping exists
	existing, err := r.GetByServiceAndRole(ctx, serviceID, roleID)
	if err == nil && existing != nil {
		// Mapping exists, increment version and update
		existing.IncrementVersion()
		if err := r.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update existing mapping: %w", err)
		}
		return existing, nil
	}

	// Mapping doesn't exist, create new one
	newMapping := models.NewServiceRoleMapping(serviceID, serviceName, roleID)
	if err := r.Create(ctx, newMapping); err != nil {
		return nil, fmt.Errorf("failed to create new mapping: %w", err)
	}

	return newMapping, nil
}
