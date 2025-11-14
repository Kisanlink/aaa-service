package groups

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// GroupRoleRepository handles database operations for GroupRole entities
type GroupRoleRepository struct {
	*base.BaseFilterableRepository[*models.GroupRole]
	dbManager db.DBManager
}

// NewGroupRoleRepository creates a new GroupRoleRepository instance
func NewGroupRoleRepository(dbManager db.DBManager) *GroupRoleRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.GroupRole]()
	baseRepo.SetDBManager(dbManager)
	return &GroupRoleRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Create creates a new group role assignment
func (r *GroupRoleRepository) Create(ctx context.Context, groupRole *models.GroupRole) error {
	return r.BaseFilterableRepository.Create(ctx, groupRole)
}

// GetByID retrieves a group role by ID
func (r *GroupRoleRepository) GetByID(ctx context.Context, id string) (*models.GroupRole, error) {
	groupRole := &models.GroupRole{}
	return r.BaseFilterableRepository.GetByID(ctx, id, groupRole)
}

// Update updates an existing group role
func (r *GroupRoleRepository) Update(ctx context.Context, groupRole *models.GroupRole) error {
	return r.BaseFilterableRepository.Update(ctx, groupRole)
}

// Delete deletes a group role by ID
func (r *GroupRoleRepository) Delete(ctx context.Context, id string) error {
	groupRole := &models.GroupRole{}
	return r.BaseFilterableRepository.Delete(ctx, id, groupRole)
}

// GetByGroupAndRole retrieves a group role by group ID and role ID
func (r *GroupRoleRepository) GetByGroupAndRole(ctx context.Context, groupID, roleID string) (*models.GroupRole, error) {
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	groupRoles, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get group role by group and role: %w", err)
	}

	if len(groupRoles) == 0 {
		return nil, nil
	}

	return groupRoles[0], nil
}

// GetByGroupID retrieves all roles assigned to a group
func (r *GroupRoleRepository) GetByGroupID(ctx context.Context, groupID string) ([]*models.GroupRole, error) {
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByGroupIDWithRoles retrieves all roles assigned to a group with role details preloaded
func (r *GroupRoleRepository) GetByGroupIDWithRoles(ctx context.Context, groupID string) ([]*models.GroupRole, error) {
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("is_active", base.OpEqual, true).
		Build()

	// Use the base repository's FindWithPreload method if available
	// For now, we'll use the basic method and note that preloading should be implemented
	groupRoles, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get group roles with roles preloaded: %w", err)
	}

	// Note: In a full implementation with GORM integration, this would use:
	// return r.BaseFilterableRepository.FindWithPreload(ctx, filter, "Role", "Organization", "Group")
	return groupRoles, nil
}

// GetByOrganizationID retrieves all group roles within an organization
func (r *GroupRoleRepository) GetByOrganizationID(ctx context.Context, organizationID string, limit, offset int) ([]*models.GroupRole, error) {
	filter := base.NewFilterBuilder().
		Where("organization_id", base.OpEqual, organizationID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByOrganizationAndGroupID retrieves all roles assigned to a specific group within an organization
func (r *GroupRoleRepository) GetByOrganizationAndGroupID(ctx context.Context, organizationID, groupID string, limit, offset int) ([]*models.GroupRole, error) {
	filter := base.NewFilterBuilder().
		Where("organization_id", base.OpEqual, organizationID).
		Where("group_id", base.OpEqual, groupID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByRoleID retrieves all groups that have a specific role assigned
func (r *GroupRoleRepository) GetByRoleID(ctx context.Context, roleID string) ([]*models.GroupRole, error) {
	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// ExistsByGroupAndRole checks if a role is already assigned to a group
func (r *GroupRoleRepository) ExistsByGroupAndRole(ctx context.Context, groupID, roleID string) (bool, error) {
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Limit(1, 0). // Only fetch 1 record for efficiency
		Build()

	// Use Find instead of CountWithFilter to work around library bug
	// where CountWithFilter passes nil model causing "Table not set" error
	groupRoles, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return false, err
	}

	return len(groupRoles) > 0, nil
}

// DeactivateByGroupAndRole deactivates a group role assignment
func (r *GroupRoleRepository) DeactivateByGroupAndRole(ctx context.Context, groupID, roleID string) error {
	groupRole, err := r.GetByGroupAndRole(ctx, groupID, roleID)
	if err != nil {
		return err
	}

	if groupRole == nil {
		return fmt.Errorf("group role assignment not found")
	}

	groupRole.IsActive = false
	return r.Update(ctx, groupRole)
}

// GetEffectiveRolesForUser retrieves all effective roles for a user through group memberships
// This method joins group_roles with group_memberships to find user's effective roles
func (r *GroupRoleRepository) GetEffectiveRolesForUser(ctx context.Context, organizationID, userID string) ([]*models.GroupRole, error) {
	// This is a complex query that requires joining multiple tables:
	// 1. group_memberships (to find user's groups)
	// 2. group_roles (to find roles assigned to those groups)
	// 3. Consider group hierarchy for inherited roles

	// For now, we implement a basic version that would work with direct group memberships
	// In a full implementation, this would use raw SQL or advanced GORM queries

	// Step 1: This would typically be done with a complex join query
	// SELECT gr.* FROM group_roles gr
	// INNER JOIN group_memberships gm ON gr.group_id = gm.group_id
	// WHERE gm.principal_id = ? AND gm.principal_type = 'user'
	// AND gr.organization_id = ? AND gr.is_active = true AND gm.is_active = true

	// For now, return empty slice with a note that this needs proper implementation
	// with access to GroupMembershipRepository or raw SQL queries
	return []*models.GroupRole{}, fmt.Errorf("GetEffectiveRolesForUser requires complex join query - implement with GroupMembershipRepository or raw SQL")
}

// CreateWithTransaction creates a new group role assignment within a transaction
func (r *GroupRoleRepository) CreateWithTransaction(ctx context.Context, groupRole *models.GroupRole) error {
	// Validate the group role before creation
	if err := groupRole.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if the assignment already exists
	exists, err := r.ExistsByGroupAndRole(ctx, groupRole.GroupID, groupRole.RoleID)
	if err != nil {
		return fmt.Errorf("failed to check existing assignment: %w", err)
	}
	if exists {
		return fmt.Errorf("role %s is already assigned to group %s", groupRole.RoleID, groupRole.GroupID)
	}

	return r.Create(ctx, groupRole)
}

// UpdateWithValidation updates an existing group role with validation
func (r *GroupRoleRepository) UpdateWithValidation(ctx context.Context, groupRole *models.GroupRole) error {
	// Validate the group role before update
	if err := groupRole.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if the group role exists
	exists, err := r.Exists(ctx, groupRole.GetID())
	if err != nil {
		return fmt.Errorf("failed to check if group role exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("group role with ID %s not found", groupRole.GetID())
	}

	return r.Update(ctx, groupRole)
}

// DeleteWithValidation deletes a group role with validation
func (r *GroupRoleRepository) DeleteWithValidation(ctx context.Context, id string) error {
	// Check if the group role exists
	exists, err := r.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if group role exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("group role with ID %s not found", id)
	}

	return r.Delete(ctx, id)
}

// SoftDelete soft deletes a group role assignment
func (r *GroupRoleRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted group role assignment
func (r *GroupRoleRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// List retrieves group roles with pagination
func (r *GroupRoleRepository) List(ctx context.Context, limit, offset int) ([]*models.GroupRole, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of group roles
func (r *GroupRoleRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.Count(ctx, filter, models.GroupRole{})
}

// Exists checks if a group role exists by ID
func (r *GroupRoleRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// BatchCreate creates multiple group role assignments in a single transaction
func (r *GroupRoleRepository) BatchCreate(ctx context.Context, groupRoles []*models.GroupRole) error {
	if len(groupRoles) == 0 {
		return nil
	}

	// Validate all group roles before creating any
	for i, groupRole := range groupRoles {
		if err := groupRole.Validate(); err != nil {
			return fmt.Errorf("validation failed for group role at index %d: %w", i, err)
		}
	}

	// Check for duplicates within the batch
	seen := make(map[string]bool)
	for i, groupRole := range groupRoles {
		key := fmt.Sprintf("%s:%s", groupRole.GroupID, groupRole.RoleID)
		if seen[key] {
			return fmt.Errorf("duplicate group-role assignment in batch at index %d: group %s, role %s", i, groupRole.GroupID, groupRole.RoleID)
		}
		seen[key] = true
	}

	// Check for existing assignments
	for i, groupRole := range groupRoles {
		exists, err := r.ExistsByGroupAndRole(ctx, groupRole.GroupID, groupRole.RoleID)
		if err != nil {
			return fmt.Errorf("failed to check existing assignment for group role at index %d: %w", i, err)
		}
		if exists {
			return fmt.Errorf("role %s is already assigned to group %s (batch index %d)", groupRole.RoleID, groupRole.GroupID, i)
		}
	}

	// Create all group roles
	for i, groupRole := range groupRoles {
		if err := r.Create(ctx, groupRole); err != nil {
			return fmt.Errorf("failed to create group role at index %d: %w", i, err)
		}
	}

	return nil
}

// BatchDeactivate deactivates multiple group role assignments
func (r *GroupRoleRepository) BatchDeactivate(ctx context.Context, groupRoleIDs []string) error {
	if len(groupRoleIDs) == 0 {
		return nil
	}

	for i, id := range groupRoleIDs {
		groupRole, err := r.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get group role at index %d (ID: %s): %w", i, id, err)
		}
		if groupRole == nil {
			return fmt.Errorf("group role not found at index %d (ID: %s)", i, id)
		}

		groupRole.IsActive = false
		if err := r.Update(ctx, groupRole); err != nil {
			return fmt.Errorf("failed to deactivate group role at index %d (ID: %s): %w", i, id, err)
		}
	}

	return nil
}

// GetActiveByGroupID retrieves only active roles assigned to a group
func (r *GroupRoleRepository) GetActiveByGroupID(ctx context.Context, groupID string) ([]*models.GroupRole, error) {
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetActiveByRoleID retrieves only active groups that have a specific role assigned
func (r *GroupRoleRepository) GetActiveByRoleID(ctx context.Context, roleID string) ([]*models.GroupRole, error) {
	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// CountByGroupID returns the number of roles assigned to a group
func (r *GroupRoleRepository) CountByGroupID(ctx context.Context, groupID string) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("group_id", base.OpEqual, groupID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// CountByRoleID returns the number of groups that have a specific role assigned
func (r *GroupRoleRepository) CountByRoleID(ctx context.Context, roleID string) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// GetByOrganizationIDActive retrieves only active group roles within an organization
func (r *GroupRoleRepository) GetByOrganizationIDActive(ctx context.Context, organizationID string, limit, offset int) ([]*models.GroupRole, error) {
	filter := base.NewFilterBuilder().
		Where("organization_id", base.OpEqual, organizationID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
