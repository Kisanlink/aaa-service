package organizations

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// OrganizationRepository handles database operations for Organization entities
type OrganizationRepository struct {
	*base.BaseFilterableRepository[*models.Organization]
	dbManager db.DBManager
}

// NewOrganizationRepository creates a new OrganizationRepository instance
func NewOrganizationRepository(dbManager db.DBManager) *OrganizationRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.Organization]()
	baseRepo.SetDBManager(dbManager)
	return &OrganizationRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Create creates a new organization using the base repository
func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	return r.BaseFilterableRepository.Create(ctx, org)
}

// GetByID retrieves an organization by ID using the base repository
func (r *OrganizationRepository) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	org := &models.Organization{}
	return r.BaseFilterableRepository.GetByID(ctx, id, org)
}

// Update updates an existing organization using the base repository
func (r *OrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	return r.BaseFilterableRepository.Update(ctx, org)
}

// Delete deletes an organization by ID using the base repository
func (r *OrganizationRepository) Delete(ctx context.Context, id string) error {
	org := &models.Organization{}
	return r.BaseFilterableRepository.Delete(ctx, id, org)
}

// List retrieves organizations with pagination using database-level filtering
func (r *OrganizationRepository) List(ctx context.Context, limit, offset int) ([]*models.Organization, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of organizations using database-level counting
func (r *OrganizationRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.Count(ctx, filter, models.Organization{})
}

// Exists checks if an organization exists by ID using the base repository
func (r *OrganizationRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// SoftDelete soft deletes an organization by ID using the base repository
func (r *OrganizationRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted organization using the base repository
func (r *OrganizationRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// ListWithDeleted retrieves organizations including soft-deleted ones using the base repository
func (r *OrganizationRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.Organization, error) {
	return r.BaseFilterableRepository.ListWithDeleted(ctx, limit, offset)
}

// CountWithDeleted returns count including soft-deleted organizations using the base repository
func (r *OrganizationRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx, &models.Organization{})
}

// ExistsWithDeleted checks if organization exists including soft-deleted ones using the base repository
func (r *OrganizationRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.ExistsWithDeleted(ctx, id)
}

// GetByCreatedBy gets organizations by creator using the base repository
func (r *OrganizationRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Organization, error) {
	return r.BaseFilterableRepository.GetByCreatedBy(ctx, createdBy, limit, offset)
}

// GetByUpdatedBy gets organizations by updater using the base repository
func (r *OrganizationRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.Organization, error) {
	return r.BaseFilterableRepository.GetByUpdatedBy(ctx, updatedBy, limit, offset)
}

// GetByName retrieves an organization by name
func (r *OrganizationRepository) GetByName(ctx context.Context, name string) (*models.Organization, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Build()

	organizations, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization by name: %w", err)
	}

	if len(organizations) == 0 {
		return nil, fmt.Errorf("organization not found with name: %s", name)
	}

	return organizations[0], nil
}

// GetByType retrieves organizations by type
func (r *OrganizationRepository) GetByType(ctx context.Context, orgType string, limit, offset int) ([]*models.Organization, error) {
	filter := base.NewFilterBuilder().
		Where("type", base.OpEqual, orgType).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// ListActive retrieves only active organizations
func (r *OrganizationRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Organization, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetChildren retrieves all child organizations
func (r *OrganizationRepository) GetChildren(ctx context.Context, parentID string) ([]*models.Organization, error) {
	filter := base.NewFilterBuilder().
		Where("parent_id", base.OpEqual, parentID).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetActiveChildren retrieves only active child organizations
func (r *OrganizationRepository) GetActiveChildren(ctx context.Context, parentID string) ([]*models.Organization, error) {
	filter := base.NewFilterBuilder().
		Where("parent_id", base.OpEqual, parentID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetParentHierarchy retrieves the complete parent hierarchy for an organization
func (r *OrganizationRepository) GetParentHierarchy(ctx context.Context, orgID string) ([]*models.Organization, error) {
	var parents []*models.Organization
	currentID := orgID

	for currentID != "" {
		org, err := r.GetByID(ctx, currentID)
		if err != nil || org == nil || org.ParentID == nil {
			break
		}

		parent, err := r.GetByID(ctx, *org.ParentID)
		if err != nil || parent == nil {
			break
		}

		parents = append([]*models.Organization{parent}, parents...)
		currentID = *org.ParentID
	}

	return parents, nil
}

// CountChildren returns the number of child organizations
func (r *OrganizationRepository) CountChildren(ctx context.Context, parentID string) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("parent_id", base.OpEqual, parentID).
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// CountGroups returns the number of groups in an organization
func (r *OrganizationRepository) CountGroups(ctx context.Context, orgID string) (int64, error) {
	// This would need to be implemented based on your group model
	// For now, returning 0 as placeholder
	return 0, nil
}

// CountUsers returns the number of users in an organization
func (r *OrganizationRepository) CountUsers(ctx context.Context, orgID string) (int64, error) {
	// This would need to be implemented based on your user-organization relationship
	// For now, returning 0 as placeholder
	return 0, nil
}

// HasActiveGroups checks if an organization has active groups
func (r *OrganizationRepository) HasActiveGroups(ctx context.Context, orgID string) (bool, error) {
	// This would need to be implemented based on your group model
	// For now, returning false as placeholder
	return false, nil
}

// Search searches for organizations by keyword in name
func (r *OrganizationRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*models.Organization, error) {
	if keyword == "" {
		return r.List(ctx, limit, offset)
	}

	filter := base.NewFilterBuilder().
		Where("name", base.OpContains, "%"+keyword+"%").
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByStatus retrieves organizations by active status
func (r *OrganizationRepository) GetByStatus(ctx context.Context, isActive bool, limit, offset int) ([]*models.Organization, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, isActive).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetRootOrganizations retrieves organizations without parents
func (r *OrganizationRepository) GetRootOrganizations(ctx context.Context, limit, offset int) ([]*models.Organization, error) {
	filter := base.NewFilterBuilder().
		Where("parent_id", base.OpIsNull, nil).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
