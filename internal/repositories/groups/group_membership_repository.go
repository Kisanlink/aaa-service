package groups

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	pkgErrors "github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// GroupMembershipRepository handles database operations for GroupMembership entities
type GroupMembershipRepository struct {
	*base.BaseFilterableRepository[*models.GroupMembership]
	dbManager db.DBManager
}

// NewGroupMembershipRepository creates a new GroupMembershipRepository instance
func NewGroupMembershipRepository(dbManager db.DBManager) *GroupMembershipRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.GroupMembership]()
	baseRepo.SetDBManager(dbManager)
	return &GroupMembershipRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Create creates a new group membership using the base repository
func (r *GroupMembershipRepository) Create(ctx context.Context, membership *models.GroupMembership) error {
	return r.BaseFilterableRepository.Create(ctx, membership)
}

// GetByID retrieves a group membership by ID using the base repository
func (r *GroupMembershipRepository) GetByID(ctx context.Context, id string) (*models.GroupMembership, error) {
	membership := &models.GroupMembership{}
	return r.BaseFilterableRepository.GetByID(ctx, id, membership)
}

// Update updates an existing group membership using the base repository
func (r *GroupMembershipRepository) Update(ctx context.Context, membership *models.GroupMembership) error {
	return r.BaseFilterableRepository.Update(ctx, membership)
}

// UpdateWithVersion updates a group membership with optimistic locking
// Returns OptimisticLockError if version mismatch occurs
func (r *GroupMembershipRepository) UpdateWithVersion(ctx context.Context, membership *models.GroupMembership, expectedVersion int) error {
	// Get database connection through dbManager
	var db *gorm.DB
	if postgresMgr, ok := r.dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		var err error
		db, err = postgresMgr.GetDB(ctx, false) // Write operation
		if err != nil {
			return fmt.Errorf("failed to get database connection: %w", err)
		}
	} else {
		return fmt.Errorf("database manager does not support GetDB method")
	}

	// First, get the current version
	var current models.GroupMembership
	if err := db.Where("id = ?", membership.ID).First(&current).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkgErrors.NewNotFoundError(fmt.Sprintf("group membership not found with id: %s", membership.ID))
		}
		return fmt.Errorf("failed to get current group membership: %w", err)
	}

	// Check if version matches
	if current.Version != expectedVersion {
		return pkgErrors.NewOptimisticLockError("group_membership", membership.ID, expectedVersion, current.Version)
	}

	// Perform the update with version check and increment
	result := db.Model(&models.GroupMembership{}).
		Where("id = ? AND version = ?", membership.ID, expectedVersion).
		Updates(map[string]interface{}{
			"group_id":       membership.GroupID,
			"principal_id":   membership.PrincipalID,
			"principal_type": membership.PrincipalType,
			"starts_at":      membership.StartsAt,
			"ends_at":        membership.EndsAt,
			"is_active":      membership.IsActive,
			"added_by_id":    membership.AddedByID,
			"metadata":       membership.Metadata,
			"version":        gorm.Expr("version + 1"),
			"updated_at":     gorm.Expr("NOW()"),
			"updated_by":     membership.UpdatedBy,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update group membership: %w", result.Error)
	}

	// Double-check that exactly one row was affected
	if result.RowsAffected == 0 {
		// Re-fetch to get current version for accurate error reporting
		if err := db.Where("id = ?", membership.ID).First(&current).Error; err == nil {
			return pkgErrors.NewOptimisticLockError("group_membership", membership.ID, expectedVersion, current.Version)
		}
		return fmt.Errorf("no rows updated for group membership: %s", membership.ID)
	}

	// Update the model's version to reflect the new state
	membership.Version = expectedVersion + 1

	return nil
}

// Delete deletes a group membership by ID using the base repository
func (r *GroupMembershipRepository) Delete(ctx context.Context, id string) error {
	membership := &models.GroupMembership{}
	return r.BaseFilterableRepository.Delete(ctx, id, membership)
}

// GetByGroupID retrieves all memberships for a specific group
func (r *GroupMembershipRepository) GetByGroupID(ctx context.Context, groupID string, limit, offset int) ([]*models.GroupMembership, error) {
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// CountByGroupID returns the count of active memberships for a specific group
func (r *GroupMembershipRepository) CountByGroupID(ctx context.Context, groupID string) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// GetByPrincipalID retrieves all memberships for a specific principal (user/service)
func (r *GroupMembershipRepository) GetByPrincipalID(ctx context.Context, principalID string, limit, offset int) ([]*models.GroupMembership, error) {
	filter := base.NewFilterBuilder().
		Where("principal_id", base.OpEqual, principalID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByGroupAndPrincipal retrieves a specific group membership
func (r *GroupMembershipRepository) GetByGroupAndPrincipal(ctx context.Context, groupID, principalID string) (*models.GroupMembership, error) {
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("principal_id", base.OpEqual, principalID).
		Where("is_active", base.OpEqual, true).
		Build()

	memberships, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(memberships) == 0 {
		return nil, nil
	}

	return memberships[0], nil
}

// GetUserDirectGroups retrieves all groups that a user is directly a member of within an organization
func (r *GroupMembershipRepository) GetUserDirectGroups(ctx context.Context, orgID, userID string) ([]*models.Group, error) {
	// For now, we'll use a simpler approach by getting memberships first, then groups
	// This can be optimized later with custom SQL queries if needed

	// Get all memberships for the user
	filter := base.NewFilterBuilder().
		Where("principal_id", base.OpEqual, userID).
		Where("principal_type", base.OpEqual, "user").
		Where("is_active", base.OpEqual, true).
		Build()

	memberships, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user memberships: %w", err)
	}

	if len(memberships) == 0 {
		return []*models.Group{}, nil
	}

	// Create a temporary group repository to get group details
	groupRepo := NewGroupRepository(r.dbManager)

	var groups []*models.Group
	for _, membership := range memberships {
		// Check if membership is currently effective
		if !membership.IsEffective(time.Now()) {
			continue
		}

		// Get the group and check if it belongs to the specified organization
		group, err := groupRepo.GetByID(ctx, membership.GroupID)
		if err != nil {
			continue // Skip if group not found
		}

		if group != nil && group.IsActive && group.OrganizationID == orgID {
			groups = append(groups, group)
		}
	}

	return groups, nil
}

// GetEffectiveMemberships retrieves all effective (currently active) memberships for a group
func (r *GroupMembershipRepository) GetEffectiveMemberships(ctx context.Context, groupID string, limit, offset int) ([]*models.GroupMembership, error) {
	now := time.Now()

	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	memberships, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter for time-effective memberships
	effectiveMemberships := make([]*models.GroupMembership, 0, len(memberships))
	for _, membership := range memberships {
		if membership.IsEffective(now) {
			effectiveMemberships = append(effectiveMemberships, membership)
		}
	}

	return effectiveMemberships, nil
}

// GetUserGroupsInOrganization retrieves all groups a user belongs to within a specific organization
func (r *GroupMembershipRepository) GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) ([]*models.Group, error) {
	// Get all groups for the user first, then apply pagination
	allGroups, err := r.GetUserDirectGroups(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	// Apply pagination manually
	start := offset
	end := offset + limit

	if start >= len(allGroups) {
		return []*models.Group{}, nil
	}

	if end > len(allGroups) {
		end = len(allGroups)
	}

	return allGroups[start:end], nil
}

// CountUserGroupsInOrganization counts all groups a user belongs to within a specific organization
func (r *GroupMembershipRepository) CountUserGroupsInOrganization(ctx context.Context, orgID, userID string) (int64, error) {
	groups, err := r.GetUserDirectGroups(ctx, orgID, userID)
	if err != nil {
		return 0, err
	}

	return int64(len(groups)), nil
}

// ExistsByGroupAndPrincipal checks if a membership exists between a group and principal
func (r *GroupMembershipRepository) ExistsByGroupAndPrincipal(ctx context.Context, groupID, principalID string) (bool, error) {
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("principal_id", base.OpEqual, principalID).
		Where("is_active", base.OpEqual, true).
		Build()

	count, err := r.BaseFilterableRepository.CountWithFilter(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeactivateMembership deactivates a group membership instead of deleting it
func (r *GroupMembershipRepository) DeactivateMembership(ctx context.Context, groupID, principalID string) error {
	membership, err := r.GetByGroupAndPrincipal(ctx, groupID, principalID)
	if err != nil {
		return err
	}

	if membership == nil {
		return fmt.Errorf("membership not found for group %s and principal %s", groupID, principalID)
	}

	membership.IsActive = false
	return r.Update(ctx, membership)
}

// ActivateMembership activates a previously deactivated group membership
func (r *GroupMembershipRepository) ActivateMembership(ctx context.Context, groupID, principalID string) error {
	// Find even inactive memberships
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("principal_id", base.OpEqual, principalID).
		Build()

	memberships, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return err
	}

	if len(memberships) == 0 {
		return fmt.Errorf("membership not found for group %s and principal %s", groupID, principalID)
	}

	membership := memberships[0]
	membership.IsActive = true
	return r.Update(ctx, membership)
}

// GetGroupMembersWithDetails retrieves group members with their user details
func (r *GroupMembershipRepository) GetGroupMembersWithDetails(ctx context.Context, groupID string, limit, offset int) ([]map[string]interface{}, error) {
	// Get effective memberships first
	memberships, err := r.GetEffectiveMemberships(ctx, groupID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert to map format for compatibility
	results := make([]map[string]interface{}, len(memberships))
	for i, membership := range memberships {
		results[i] = map[string]interface{}{
			"membership_id":         membership.ID,
			"principal_id":          membership.PrincipalID,
			"principal_type":        membership.PrincipalType,
			"starts_at":             membership.StartsAt,
			"ends_at":               membership.EndsAt,
			"is_active":             membership.IsActive,
			"added_by_id":           membership.AddedByID,
			"membership_created_at": membership.CreatedAt,
			// User details would need to be fetched separately if needed
			"username":       nil,
			"first_name":     nil,
			"last_name":      nil,
			"email":          nil,
			"user_is_active": nil,
		}
	}

	return results, nil
}

// List retrieves group memberships with pagination using database-level filtering
func (r *GroupMembershipRepository) List(ctx context.Context, limit, offset int) ([]*models.GroupMembership, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of group memberships using database-level counting
func (r *GroupMembershipRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.Count(ctx, filter, models.GroupMembership{})
}

// Exists checks if a group membership exists by ID using the base repository
func (r *GroupMembershipRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// SoftDelete soft deletes a group membership by ID using the base repository
func (r *GroupMembershipRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted group membership using the base repository
func (r *GroupMembershipRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}
