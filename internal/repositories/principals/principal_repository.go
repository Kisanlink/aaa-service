package principals

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// PrincipalRepository handles database operations for Principal entities
type PrincipalRepository struct {
	*base.BaseFilterableRepository[*models.Principal]
	dbManager db.DBManager
}

// ServiceRepository handles database operations for Service entities
type ServiceRepository struct {
	*base.BaseFilterableRepository[*models.Service]
	dbManager db.DBManager
}

// NewPrincipalRepository creates a new PrincipalRepository instance
func NewPrincipalRepository(dbManager db.DBManager) *PrincipalRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.Principal]()
	baseRepo.SetDBManager(dbManager)
	return &PrincipalRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// NewServiceRepository creates a new ServiceRepository instance
func NewServiceRepository(dbManager db.DBManager) *ServiceRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.Service]()
	baseRepo.SetDBManager(dbManager)
	return &ServiceRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Principal Repository Methods

// Create creates a new principal using the base repository
func (r *PrincipalRepository) Create(ctx context.Context, principal *models.Principal) error {
	return r.BaseFilterableRepository.Create(ctx, principal)
}

// GetByID retrieves a principal by ID using the base repository
func (r *PrincipalRepository) GetByID(ctx context.Context, id string) (*models.Principal, error) {
	principal := &models.Principal{}
	return r.BaseFilterableRepository.GetByID(ctx, id, principal)
}

// Update updates an existing principal using the base repository
func (r *PrincipalRepository) Update(ctx context.Context, principal *models.Principal) error {
	return r.BaseFilterableRepository.Update(ctx, principal)
}

// Delete deletes a principal by ID using the base repository
func (r *PrincipalRepository) Delete(ctx context.Context, id string) error {
	principal := &models.Principal{}
	return r.BaseFilterableRepository.Delete(ctx, id, principal)
}

// SoftDelete soft deletes a principal by ID using the base repository
func (r *PrincipalRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// GetByUserID retrieves a principal by user ID
func (r *PrincipalRepository) GetByUserID(ctx context.Context, userID string) (*models.Principal, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("is_active", base.OpEqual, true).
		Build()

	principals, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(principals) == 0 {
		return nil, nil
	}

	return principals[0], nil
}

// GetByServiceID retrieves a principal by service ID
func (r *PrincipalRepository) GetByServiceID(ctx context.Context, serviceID string) (*models.Principal, error) {
	filter := base.NewFilterBuilder().
		Where("service_id", base.OpEqual, serviceID).
		Where("is_active", base.OpEqual, true).
		Build()

	principals, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(principals) == 0 {
		return nil, nil
	}

	return principals[0], nil
}

// GetByType retrieves principals by type with pagination
func (r *PrincipalRepository) GetByType(ctx context.Context, principalType string, limit, offset int) ([]*models.Principal, error) {
	filter := base.NewFilterBuilder().
		Where("type", base.OpEqual, principalType).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByOrganization retrieves principals by organization with pagination
func (r *PrincipalRepository) GetByOrganization(ctx context.Context, organizationID string, limit, offset int) ([]*models.Principal, error) {
	filter := base.NewFilterBuilder().
		Where("organization_id", base.OpEqual, organizationID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// List retrieves principals with pagination using database-level filtering
func (r *PrincipalRepository) List(ctx context.Context, limit, offset int) ([]*models.Principal, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// ListActive retrieves only active principals with pagination
func (r *PrincipalRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Principal, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of principals using database-level counting
func (r *PrincipalRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// Exists checks if a principal exists by ID using the base repository
func (r *PrincipalRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// Service Repository Methods

// Create creates a new service using the base repository
func (r *ServiceRepository) Create(ctx context.Context, service *models.Service) error {
	return r.BaseFilterableRepository.Create(ctx, service)
}

// GetByID retrieves a service by ID using the base repository
func (r *ServiceRepository) GetByID(ctx context.Context, id string) (*models.Service, error) {
	service := &models.Service{}
	return r.BaseFilterableRepository.GetByID(ctx, id, service)
}

// Update updates an existing service using the base repository
func (r *ServiceRepository) Update(ctx context.Context, service *models.Service) error {
	return r.BaseFilterableRepository.Update(ctx, service)
}

// Delete deletes a service by ID using the base repository
func (r *ServiceRepository) Delete(ctx context.Context, id string) error {
	service := &models.Service{}
	return r.BaseFilterableRepository.Delete(ctx, id, service)
}

// SoftDelete soft deletes a service by ID using the base repository
func (r *ServiceRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// GetByName retrieves a service by name
func (r *ServiceRepository) GetByName(ctx context.Context, name string) (*models.Service, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("is_active", base.OpEqual, true).
		Build()

	services, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	return services[0], nil
}

// GetByOrganization retrieves services by organization with pagination
func (r *ServiceRepository) GetByOrganization(ctx context.Context, organizationID string, limit, offset int) ([]*models.Service, error) {
	filter := base.NewFilterBuilder().
		Where("organization_id", base.OpEqual, organizationID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByAPIKey retrieves a service by API key hash
func (r *ServiceRepository) GetByAPIKey(ctx context.Context, apiKeyHash string) (*models.Service, error) {
	filter := base.NewFilterBuilder().
		Where("api_key", base.OpEqual, apiKeyHash).
		Where("is_active", base.OpEqual, true).
		Build()

	services, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	return services[0], nil
}

// List retrieves services with pagination using database-level filtering
func (r *ServiceRepository) List(ctx context.Context, limit, offset int) ([]*models.Service, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// ListActive retrieves only active services with pagination
func (r *ServiceRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Service, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of services using database-level counting
func (r *ServiceRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// Exists checks if a service exists by ID using the base repository
func (r *ServiceRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// Enhanced Principal Repository Methods

// GetByName retrieves a principal by name
func (r *PrincipalRepository) GetByName(ctx context.Context, name string) (*models.Principal, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("is_active", base.OpEqual, true).
		Build()

	principals, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(principals) == 0 {
		return nil, nil
	}

	return principals[0], nil
}

// GetByNameAndOrganization retrieves a principal by name within a specific organization
func (r *PrincipalRepository) GetByNameAndOrganization(ctx context.Context, name, organizationID string) (*models.Principal, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("organization_id", base.OpEqual, organizationID).
		Where("is_active", base.OpEqual, true).
		Build()

	principals, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(principals) == 0 {
		return nil, nil
	}

	return principals[0], nil
}

// SearchPrincipals searches principals by keyword with pagination
func (r *PrincipalRepository) SearchPrincipals(ctx context.Context, keyword string, limit, offset int) ([]*models.Principal, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpLike, "%"+keyword+"%").
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetPrincipalsByTypeAndOrganization retrieves principals by type and organization
func (r *PrincipalRepository) GetPrincipalsByTypeAndOrganization(ctx context.Context, principalType, organizationID string, limit, offset int) ([]*models.Principal, error) {
	filter := base.NewFilterBuilder().
		Where("type", base.OpEqual, principalType).
		Where("organization_id", base.OpEqual, organizationID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// CountByType returns the count of principals by type
func (r *PrincipalRepository) CountByType(ctx context.Context, principalType string) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("type", base.OpEqual, principalType).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// CountByOrganization returns the count of principals by organization
func (r *PrincipalRepository) CountByOrganization(ctx context.Context, organizationID string) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("organization_id", base.OpEqual, organizationID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// GetPrincipalsWithMetadata retrieves principals with specific metadata
func (r *PrincipalRepository) GetPrincipalsWithMetadata(ctx context.Context, metadataKey, metadataValue string, limit, offset int) ([]*models.Principal, error) {
	// This would need JSONB query support in the base repository
	// For now, we'll use a simple approach
	filter := base.NewFilterBuilder().
		Where("metadata", base.OpLike, "%"+metadataKey+"%").
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Enhanced Service Repository Methods

// SearchServices searches services by keyword with pagination
func (r *ServiceRepository) SearchServices(ctx context.Context, keyword string, limit, offset int) ([]*models.Service, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpLike, "%"+keyword+"%").
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetServicesByMetadata retrieves services with specific metadata
func (r *ServiceRepository) GetServicesByMetadata(ctx context.Context, metadataKey, metadataValue string, limit, offset int) ([]*models.Service, error) {
	// This would need JSONB query support in the base repository
	// For now, we'll use a simple approach
	filter := base.NewFilterBuilder().
		Where("metadata", base.OpLike, "%"+metadataKey+"%").
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// CountByOrganization returns the count of services by organization
func (r *ServiceRepository) CountByOrganization(ctx context.Context, organizationID string) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("organization_id", base.OpEqual, organizationID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// GetServicesByStatus retrieves services by active status
func (r *ServiceRepository) GetServicesByStatus(ctx context.Context, isActive bool, limit, offset int) ([]*models.Service, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, isActive).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Bulk Operations

// BulkCreatePrincipals creates multiple principals in a transaction
func (r *PrincipalRepository) BulkCreatePrincipals(ctx context.Context, principals []*models.Principal) error {
	// This would need transaction support in the base repository
	// For now, we'll create them one by one
	for _, principal := range principals {
		if err := r.Create(ctx, principal); err != nil {
			return err
		}
	}
	return nil
}

// BulkUpdatePrincipals updates multiple principals in a transaction
func (r *PrincipalRepository) BulkUpdatePrincipals(ctx context.Context, principals []*models.Principal) error {
	// This would need transaction support in the base repository
	// For now, we'll update them one by one
	for _, principal := range principals {
		if err := r.Update(ctx, principal); err != nil {
			return err
		}
	}
	return nil
}

// BulkCreateServices creates multiple services in a transaction
func (r *ServiceRepository) BulkCreateServices(ctx context.Context, services []*models.Service) error {
	// This would need transaction support in the base repository
	// For now, we'll create them one by one
	for _, service := range services {
		if err := r.Create(ctx, service); err != nil {
			return err
		}
	}
	return nil
}

// BulkUpdateServices updates multiple services in a transaction
func (r *ServiceRepository) BulkUpdateServices(ctx context.Context, services []*models.Service) error {
	// This would need transaction support in the base repository
	// For now, we'll update them one by one
	for _, service := range services {
		if err := r.Update(ctx, service); err != nil {
			return err
		}
	}
	return nil
}

// Utility Methods

// IsNameUnique checks if a principal name is unique within an organization
func (r *PrincipalRepository) IsNameUnique(ctx context.Context, name, organizationID string, excludeID string) (bool, error) {
	filterBuilder := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("organization_id", base.OpEqual, organizationID).
		Where("is_active", base.OpEqual, true)

	if excludeID != "" {
		filterBuilder = filterBuilder.Where("id", base.OpNotEqual, excludeID)
	}

	filter := filterBuilder.Build()

	principals, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return false, err
	}

	return len(principals) == 0, nil
}

// IsServiceNameUnique checks if a service name is unique within an organization
func (r *ServiceRepository) IsServiceNameUnique(ctx context.Context, name, organizationID string, excludeID string) (bool, error) {
	filterBuilder := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("organization_id", base.OpEqual, organizationID).
		Where("is_active", base.OpEqual, true)

	if excludeID != "" {
		filterBuilder = filterBuilder.Where("id", base.OpNotEqual, excludeID)
	}

	filter := filterBuilder.Build()

	services, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return false, err
	}

	return len(services) == 0, nil
}

// GetPrincipalStats returns statistics about principals
func (r *PrincipalRepository) GetPrincipalStats(ctx context.Context, organizationID string) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by type
	userCount, err := r.CountByType(ctx, "user")
	if err != nil {
		return nil, err
	}
	stats["user_count"] = userCount

	serviceCount, err := r.CountByType(ctx, "service")
	if err != nil {
		return nil, err
	}
	stats["service_count"] = serviceCount

	// Count by organization if specified
	if organizationID != "" {
		orgCount, err := r.CountByOrganization(ctx, organizationID)
		if err != nil {
			return nil, err
		}
		stats["organization_count"] = orgCount
	}

	// Total count
	totalCount, err := r.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	return stats, nil
}

// GetServiceStats returns statistics about services
func (r *ServiceRepository) GetServiceStats(ctx context.Context, organizationID string) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by organization if specified
	if organizationID != "" {
		orgCount, err := r.CountByOrganization(ctx, organizationID)
		if err != nil {
			return nil, err
		}
		stats["organization_count"] = orgCount
	}

	// Total count
	totalCount, err := r.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	// Active count
	activeFilter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Build()
	activeCount, err := r.BaseFilterableRepository.CountWithFilter(ctx, activeFilter)
	if err != nil {
		return nil, err
	}
	stats["active_count"] = activeCount

	return stats, nil
}
