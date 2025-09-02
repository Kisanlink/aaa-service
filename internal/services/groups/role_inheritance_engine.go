package groups

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/repositories/groups"
	"github.com/Kisanlink/aaa-service/internal/repositories/roles"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
)

// GroupRepositoryInterface defines the interface for group repository operations
type GroupRepositoryInterface interface {
	GetChildren(ctx context.Context, parentID string) ([]*models.Group, error)
}

// GroupRoleRepositoryInterface defines the interface for group role repository operations
type GroupRoleRepositoryInterface interface {
	GetByGroupID(ctx context.Context, groupID string) ([]*models.GroupRole, error)
}

// RoleRepositoryInterface defines the interface for role repository operations
type RoleRepositoryInterface interface {
	GetByID(ctx context.Context, id string, role *models.Role) (*models.Role, error)
}

// RoleInheritanceEngine handles role inheritance calculations for hierarchical groups
// Implements upward inheritance: parent groups inherit roles from their child groups
type RoleInheritanceEngine struct {
	groupRepo     GroupRepositoryInterface
	groupRoleRepo GroupRoleRepositoryInterface
	roleRepo      RoleRepositoryInterface
	cache         interfaces.CacheService
	logger        *zap.Logger
}

// NewRoleInheritanceEngine creates a new role inheritance engine
func NewRoleInheritanceEngine(
	groupRepo GroupRepositoryInterface,
	groupRoleRepo GroupRoleRepositoryInterface,
	roleRepo RoleRepositoryInterface,
	cache interfaces.CacheService,
	logger *zap.Logger,
) *RoleInheritanceEngine {
	return &RoleInheritanceEngine{
		groupRepo:     groupRepo,
		groupRoleRepo: groupRoleRepo,
		roleRepo:      roleRepo,
		cache:         cache,
		logger:        logger,
	}
}

// NewRoleInheritanceEngineWithRepos creates a new role inheritance engine with concrete repository types
func NewRoleInheritanceEngineWithRepos(
	groupRepo *groups.GroupRepository,
	groupRoleRepo *groups.GroupRoleRepository,
	roleRepo *roles.RoleRepository,
	cache interfaces.CacheService,
	logger *zap.Logger,
) *RoleInheritanceEngine {
	return &RoleInheritanceEngine{
		groupRepo:     groupRepo,
		groupRoleRepo: groupRoleRepo,
		roleRepo:      roleRepo,
		cache:         cache,
		logger:        logger,
	}
}

// EffectiveRole represents a role with its inheritance path and precedence
type EffectiveRole struct {
	Role            *models.Role `json:"role"`
	GroupID         string       `json:"group_id"`
	GroupName       string       `json:"group_name"`
	InheritancePath []string     `json:"inheritance_path"` // Path from user's direct group to role source
	Distance        int          `json:"distance"`         // Distance from user's direct group (0 = direct, 1 = child, 2 = grandchild, etc.)
	IsDirectRole    bool         `json:"is_direct_role"`   // True if role is directly assigned to user's group
}

// CalculateEffectiveRoles calculates all effective roles for a user in an organization
// using upward inheritance (parent groups inherit from child groups)
func (r *RoleInheritanceEngine) CalculateEffectiveRoles(ctx context.Context, orgID, userID string) ([]*EffectiveRole, error) {
	r.logger.Info("Calculating effective roles for user",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	// Check cache first
	cacheKey := fmt.Sprintf("org:%s:user:%s:effective_roles", orgID, userID)
	if cached, found := r.cache.Get(cacheKey); found {
		if effectiveRoles, ok := cached.([]*EffectiveRole); ok {
			r.logger.Debug("Returning cached effective roles", zap.String("user_id", userID))
			return effectiveRoles, nil
		}
	}

	// 1. Get user's direct group memberships
	directGroups, err := r.getUserDirectGroups(ctx, orgID, userID)
	if err != nil {
		r.logger.Error("Failed to get user's direct groups", zap.Error(err))
		return nil, err
	}

	if len(directGroups) == 0 {
		r.logger.Info("User has no group memberships", zap.String("user_id", userID))
		return []*EffectiveRole{}, nil
	}

	// 2. For each direct group, collect roles through upward inheritance
	allEffectiveRoles := make(map[string]*EffectiveRole) // roleID -> EffectiveRole

	for _, directGroup := range directGroups {
		groupRoles, err := r.calculateRolesForGroupHierarchy(ctx, directGroup, 0)
		if err != nil {
			r.logger.Error("Failed to calculate roles for group hierarchy",
				zap.String("group_id", directGroup.ID),
				zap.Error(err))
			continue
		}

		// Merge roles with conflict resolution (most specific wins)
		for roleID, effectiveRole := range groupRoles {
			if existing, exists := allEffectiveRoles[roleID]; exists {
				// Keep the role with the shortest distance (most specific)
				if effectiveRole.Distance < existing.Distance {
					allEffectiveRoles[roleID] = effectiveRole
				}
			} else {
				allEffectiveRoles[roleID] = effectiveRole
			}
		}
	}

	// 3. Convert map to slice and sort by precedence
	effectiveRoles := make([]*EffectiveRole, 0, len(allEffectiveRoles))
	for _, role := range allEffectiveRoles {
		effectiveRoles = append(effectiveRoles, role)
	}

	// Sort by distance (most specific first), then by role name for consistency
	r.sortEffectiveRolesByPrecedence(effectiveRoles)

	// Cache the result
	err = r.cache.Set(cacheKey, effectiveRoles, 300) // Cache for 5 minutes
	if err != nil {
		r.logger.Warn("Failed to cache effective roles", zap.Error(err))
	}

	r.logger.Info("Calculated effective roles for user",
		zap.String("user_id", userID),
		zap.Int("role_count", len(effectiveRoles)))

	return effectiveRoles, nil
}

// calculateRolesForGroupHierarchy calculates roles for a group and all its descendant groups
// using upward inheritance (this group inherits roles from its children)
func (r *RoleInheritanceEngine) calculateRolesForGroupHierarchy(ctx context.Context, group *models.Group, currentDistance int) (map[string]*EffectiveRole, error) {
	roles := make(map[string]*EffectiveRole)

	// 1. Get direct roles assigned to this group
	directRoles, err := r.getDirectGroupRoles(ctx, group.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct roles for group %s: %w", group.ID, err)
	}

	// Add direct roles
	for _, groupRole := range directRoles {
		role := &models.Role{}
		role, err = r.roleRepo.GetByID(ctx, groupRole.RoleID, role)
		if err != nil {
			r.logger.Warn("Failed to load role details",
				zap.String("role_id", groupRole.RoleID),
				zap.Error(err))
			continue
		}

		if role != nil && role.IsActive {
			effectiveRole := &EffectiveRole{
				Role:            role,
				GroupID:         group.ID,
				GroupName:       group.Name,
				InheritancePath: []string{group.ID},
				Distance:        currentDistance,
				IsDirectRole:    currentDistance == 0,
			}
			roles[role.ID] = effectiveRole
		}
	}

	// 2. Get child groups and inherit their roles (upward inheritance)
	childGroups, err := r.groupRepo.GetChildren(ctx, group.ID)
	if err != nil {
		r.logger.Warn("Failed to get child groups",
			zap.String("group_id", group.ID),
			zap.Error(err))
		return roles, nil // Continue with direct roles even if children fail
	}

	// Recursively collect roles from child groups
	for _, childGroup := range childGroups {
		if !childGroup.IsActive {
			continue
		}

		childRoles, err := r.calculateRolesForGroupHierarchy(ctx, childGroup, currentDistance+1)
		if err != nil {
			r.logger.Warn("Failed to calculate roles for child group",
				zap.String("child_group_id", childGroup.ID),
				zap.Error(err))
			continue
		}

		// Merge child roles with conflict resolution
		for roleID, childRole := range childRoles {
			if existing, exists := roles[roleID]; exists {
				// Keep the role with the shortest distance (most specific)
				if childRole.Distance < existing.Distance {
					// Update inheritance path to include current group
					childRole.InheritancePath = append([]string{group.ID}, childRole.InheritancePath...)
					roles[roleID] = childRole
				}
			} else {
				// Update inheritance path to include current group
				childRole.InheritancePath = append([]string{group.ID}, childRole.InheritancePath...)
				roles[roleID] = childRole
			}
		}
	}

	return roles, nil
}

// getUserDirectGroups gets all groups that a user is directly a member of
func (r *RoleInheritanceEngine) getUserDirectGroups(ctx context.Context, orgID, userID string) ([]*models.Group, error) {
	r.logger.Debug("Getting user's direct groups",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	// TODO: Implement when GroupMembershipRepository is available
	// This would typically query the group_memberships table and join with groups table
	// to get active memberships for the user within the specified organization
	//
	// Expected query would be something like:
	// SELECT g.* FROM groups g
	// JOIN group_memberships gm ON g.id = gm.group_id
	// WHERE gm.principal_id = ? AND g.organization_id = ?
	// AND gm.is_active = true AND g.is_active = true
	// AND (gm.starts_at IS NULL OR gm.starts_at <= NOW())
	// AND (gm.ends_at IS NULL OR gm.ends_at > NOW())

	// For now, return empty slice to prevent errors
	// This means effective roles calculation will work but won't find any roles
	// until group membership functionality is fully implemented
	r.logger.Warn("Group membership functionality not yet implemented, returning empty groups",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	return []*models.Group{}, nil
}

// getDirectGroupRoles gets roles directly assigned to a group (not inherited)
func (r *RoleInheritanceEngine) getDirectGroupRoles(ctx context.Context, groupID string) ([]*models.GroupRole, error) {
	groupRoles, err := r.groupRoleRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Filter for currently effective roles
	now := time.Now()
	effectiveRoles := make([]*models.GroupRole, 0, len(groupRoles))
	for _, groupRole := range groupRoles {
		if groupRole.IsEffective(now) {
			effectiveRoles = append(effectiveRoles, groupRole)
		}
	}

	return effectiveRoles, nil
}

// sortEffectiveRolesByPrecedence sorts effective roles by precedence
// Most specific roles (shortest distance) come first
func (r *RoleInheritanceEngine) sortEffectiveRolesByPrecedence(roles []*EffectiveRole) {
	// Sort by distance (ascending), then by role name for consistency
	for i := 0; i < len(roles)-1; i++ {
		for j := i + 1; j < len(roles); j++ {
			// Primary sort: distance (most specific first)
			if roles[i].Distance > roles[j].Distance {
				roles[i], roles[j] = roles[j], roles[i]
			} else if roles[i].Distance == roles[j].Distance {
				// Secondary sort: role name for consistency
				if roles[i].Role.Name > roles[j].Role.Name {
					roles[i], roles[j] = roles[j], roles[i]
				}
			}
		}
	}
}

// InvalidateUserRoleCache invalidates the cached effective roles for a user
func (r *RoleInheritanceEngine) InvalidateUserRoleCache(ctx context.Context, orgID, userID string) error {
	cacheKey := fmt.Sprintf("org:%s:user:%s:effective_roles", orgID, userID)
	return r.cache.Delete(cacheKey)
}

// InvalidateGroupRoleCache invalidates cached effective roles for all users in a group hierarchy
func (r *RoleInheritanceEngine) InvalidateGroupRoleCache(ctx context.Context, orgID, groupID string) error {
	// This would need to invalidate cache for all users affected by changes to this group
	// For now, we'll use a simple pattern-based invalidation
	pattern := fmt.Sprintf("org:%s:user:*:effective_roles", orgID)
	keys, err := r.cache.Keys(pattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := r.cache.Delete(key); err != nil {
			r.logger.Warn("Failed to invalidate cache key", zap.String("key", key), zap.Error(err))
		}
	}

	return nil
}

// GetRoleInheritancePath returns the inheritance path for a specific role for a user
func (r *RoleInheritanceEngine) GetRoleInheritancePath(ctx context.Context, orgID, userID, roleID string) (*EffectiveRole, error) {
	effectiveRoles, err := r.CalculateEffectiveRoles(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	for _, effectiveRole := range effectiveRoles {
		if effectiveRole.Role.ID == roleID {
			return effectiveRole, nil
		}
	}

	return nil, errors.NewNotFoundError("role not found in user's effective roles")
}
