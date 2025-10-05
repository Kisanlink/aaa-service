package groups

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// GroupRepository handles database operations for Group entities
type GroupRepository struct {
	*base.BaseFilterableRepository[*models.Group]
	dbManager db.DBManager
}

// NewGroupRepository creates a new GroupRepository instance
func NewGroupRepository(dbManager db.DBManager) *GroupRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.Group]()
	baseRepo.SetDBManager(dbManager)
	return &GroupRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Create creates a new group using the base repository
func (r *GroupRepository) Create(ctx context.Context, group *models.Group) error {
	return r.BaseFilterableRepository.Create(ctx, group)
}

// GetByID retrieves a group by ID using the base repository
func (r *GroupRepository) GetByID(ctx context.Context, id string) (*models.Group, error) {
	group := &models.Group{}
	return r.BaseFilterableRepository.GetByID(ctx, id, group)
}

// Update updates an existing group using the base repository
func (r *GroupRepository) Update(ctx context.Context, group *models.Group) error {
	return r.BaseFilterableRepository.Update(ctx, group)
}

// Delete deletes a group by ID using the base repository
func (r *GroupRepository) Delete(ctx context.Context, id string) error {
	group := &models.Group{}
	return r.BaseFilterableRepository.Delete(ctx, id, group)
}

// List retrieves groups with pagination using database-level filtering
func (r *GroupRepository) List(ctx context.Context, limit, offset int) ([]*models.Group, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of groups using database-level counting
func (r *GroupRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// Exists checks if a group exists by ID using the base repository
func (r *GroupRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// SoftDelete soft deletes a group by ID using the base repository
func (r *GroupRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted group using the base repository
func (r *GroupRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// ListWithDeleted retrieves groups including soft-deleted ones using the base repository
func (r *GroupRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.Group, error) {
	return r.BaseFilterableRepository.ListWithDeleted(ctx, limit, offset)
}

// CountWithDeleted returns count including soft-deleted groups using the base repository
func (r *GroupRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx)
}

// ExistsWithDeleted checks if group exists including soft-deleted ones using the base repository
func (r *GroupRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.ExistsWithDeleted(ctx, id)
}

// GetByCreatedBy gets groups by creator using the base repository
func (r *GroupRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Group, error) {
	return r.BaseFilterableRepository.GetByCreatedBy(ctx, createdBy, limit, offset)
}

// GetByUpdatedBy gets groups by updater using the base repository
func (r *GroupRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.Group, error) {
	return r.BaseFilterableRepository.GetByUpdatedBy(ctx, updatedBy, limit, offset)
}

// GetByName retrieves a group by name
func (r *GroupRepository) GetByName(ctx context.Context, name string) (*models.Group, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Build()

	groups, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get group by name: %w", err)
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("group not found with name: %s", name)
	}

	return groups[0], nil
}

// GetByServiceName retrieves groups by service name
func (r *GroupRepository) GetByServiceName(ctx context.Context, serviceName string, limit, offset int) ([]*models.Group, error) {
	filter := base.NewFilterBuilder().
		Where("service_name", base.OpEqual, serviceName).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByType retrieves groups by type
func (r *GroupRepository) GetByType(ctx context.Context, groupType string, limit, offset int) ([]*models.Group, error) {
	filter := base.NewFilterBuilder().
		Where("type", base.OpEqual, groupType).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByNameAndOrganization retrieves a group by name within a specific organization
func (r *GroupRepository) GetByNameAndOrganization(ctx context.Context, name, organizationID string) (*models.Group, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("organization_id", base.OpEqual, organizationID).
		Build()

	groups, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, nil
	}

	return groups[0], nil
}

// GetByOrganization retrieves groups by organization with pagination
func (r *GroupRepository) GetByOrganization(ctx context.Context, organizationID string, limit, offset int, includeInactive bool) ([]*models.Group, error) {
	filter := base.NewFilterBuilder().
		Where("organization_id", base.OpEqual, organizationID).
		Limit(limit, offset).
		Build()

	if !includeInactive {
		filter = base.NewFilterBuilder().
			Where("organization_id", base.OpEqual, organizationID).
			Where("is_active", base.OpEqual, true).
			Limit(limit, offset).
			Build()
	}

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// ListActive retrieves only active groups with pagination
func (r *GroupRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Group, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetChildren retrieves all child groups of a parent group
func (r *GroupRepository) GetChildren(ctx context.Context, parentID string) ([]*models.Group, error) {
	filter := base.NewFilterBuilder().
		Where("parent_id", base.OpEqual, parentID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// HasActiveMembers checks if a group has any active members
func (r *GroupRepository) HasActiveMembers(ctx context.Context, groupID string) (bool, error) {
	// This would need to check the group_memberships table
	// For now, we'll use a simple count approach
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("is_active", base.OpEqual, true).
		Build()

	count, err := r.BaseFilterableRepository.CountWithFilter(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateMembership creates a new group membership
// Note: This method is deprecated. Use GroupMembershipRepository directly for better separation of concerns.
func (r *GroupRepository) CreateMembership(ctx context.Context, membership *models.GroupMembership) error {
	// Create a temporary GroupMembershipRepository instance
	membershipRepo := NewGroupMembershipRepository(r.dbManager)
	return membershipRepo.Create(ctx, membership)
}

// UpdateMembership updates an existing group membership
// Note: This method is deprecated. Use GroupMembershipRepository directly for better separation of concerns.
func (r *GroupRepository) UpdateMembership(ctx context.Context, membership *models.GroupMembership) error {
	// Create a temporary GroupMembershipRepository instance
	membershipRepo := NewGroupMembershipRepository(r.dbManager)
	return membershipRepo.Update(ctx, membership)
}

// GetMembership retrieves a specific group membership
// Note: This method is deprecated. Use GroupMembershipRepository directly for better separation of concerns.
func (r *GroupRepository) GetMembership(ctx context.Context, groupID, principalID string) (*models.GroupMembership, error) {
	// Create a temporary GroupMembershipRepository instance
	membershipRepo := NewGroupMembershipRepository(r.dbManager)
	return membershipRepo.GetByGroupAndPrincipal(ctx, groupID, principalID)
}

// GetGroupMembers retrieves all members of a group with pagination
// Note: This method is deprecated. Use GroupMembershipRepository directly for better separation of concerns.
func (r *GroupRepository) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) ([]*models.GroupMembership, error) {
	// Create a temporary GroupMembershipRepository instance
	membershipRepo := NewGroupMembershipRepository(r.dbManager)
	return membershipRepo.GetByGroupID(ctx, groupID, limit, offset)
}
