package groups

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"go.uber.org/zap"
)

// GroupCacheService provides caching functionality for group-related operations
type GroupCacheService struct {
	cache  interfaces.CacheService
	logger *zap.Logger
}

// NewGroupCacheService creates a new group cache service
func NewGroupCacheService(cache interfaces.CacheService, logger *zap.Logger) *GroupCacheService {
	return &GroupCacheService{
		cache:  cache,
		logger: logger,
	}
}

// Cache key patterns for group data
const (
	// Group hierarchy caching
	GroupHierarchyKeyPattern   = "group:%s:hierarchy"
	GroupAncestorsPattern      = "group:%s:ancestors"
	GroupDescendantsPattern    = "group:%s:descendants"
	GroupChildrenPattern       = "group:%s:children"
	GroupActiveChildrenPattern = "group:%s:active_children"

	// Group roles caching
	GroupRolesPattern       = "group:%s:roles"
	GroupActiveRolesPattern = "group:%s:active_roles"
	GroupRoleDetailsPattern = "group:%s:role_details"

	// Group members caching
	GroupMembersPattern       = "group:%s:members"
	GroupActiveMembersPattern = "group:%s:active_members"
	GroupMemberDetailsPattern = "group:%s:member_details"

	// User effective roles caching (enhanced from role inheritance engine)
	UserEffectiveRolesPattern   = "org:%s:user:%s:effective_roles_v2"
	UserGroupMembershipsPattern = "org:%s:user:%s:group_memberships"

	// Group role inheritance paths
	GroupRoleInheritancePattern = "group:%s:role_inheritance"

	// Cache TTL values (in seconds)
	GroupHierarchyCacheTTL = 1800 // 30 minutes for hierarchy data
	GroupRolesCacheTTL     = 900  // 15 minutes for group roles
	GroupMembersCacheTTL   = 600  // 10 minutes for group members
	EffectiveRolesCacheTTL = 300  // 5 minutes for effective roles
	InheritanceCacheTTL    = 1200 // 20 minutes for inheritance paths
)

// CacheGroupHierarchy caches the complete group hierarchy
func (c *GroupCacheService) CacheGroupHierarchy(ctx context.Context, groupID string, hierarchy interface{}) error {
	key := fmt.Sprintf(GroupHierarchyKeyPattern, groupID)

	if err := c.cache.Set(key, hierarchy, GroupHierarchyCacheTTL); err != nil {
		c.logger.Warn("Failed to cache group hierarchy",
			zap.String("group_id", groupID),
			zap.String("cache_key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached group hierarchy",
		zap.String("group_id", groupID),
		zap.String("cache_key", key))

	return nil
}

// GetCachedGroupHierarchy retrieves cached group hierarchy
func (c *GroupCacheService) GetCachedGroupHierarchy(ctx context.Context, groupID string) (interface{}, bool) {
	key := fmt.Sprintf(GroupHierarchyKeyPattern, groupID)

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	c.logger.Debug("Retrieved cached group hierarchy",
		zap.String("group_id", groupID),
		zap.String("cache_key", key))

	return cached, true
}

// CacheGroupAncestors caches the ancestor chain for a group
func (c *GroupCacheService) CacheGroupAncestors(ctx context.Context, groupID string, ancestors []*models.Group) error {
	key := fmt.Sprintf(GroupAncestorsPattern, groupID)

	if err := c.cache.Set(key, ancestors, GroupHierarchyCacheTTL); err != nil {
		c.logger.Warn("Failed to cache group ancestors",
			zap.String("group_id", groupID),
			zap.String("cache_key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached group ancestors",
		zap.String("group_id", groupID),
		zap.String("cache_key", key),
		zap.Int("ancestor_count", len(ancestors)))

	return nil
}

// GetCachedGroupAncestors retrieves cached group ancestors
func (c *GroupCacheService) GetCachedGroupAncestors(ctx context.Context, groupID string) ([]*models.Group, bool) {
	key := fmt.Sprintf(GroupAncestorsPattern, groupID)

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	if ancestors, ok := cached.([]*models.Group); ok {
		c.logger.Debug("Retrieved cached group ancestors",
			zap.String("group_id", groupID),
			zap.String("cache_key", key),
			zap.Int("ancestor_count", len(ancestors)))
		return ancestors, true
	}

	c.logger.Warn("Invalid cached ancestors type",
		zap.String("group_id", groupID),
		zap.String("cache_key", key))

	// Remove invalid cache entry
	c.cache.Delete(key)
	return nil, false
}

// CacheGroupDescendants caches the descendant tree for a group
func (c *GroupCacheService) CacheGroupDescendants(ctx context.Context, groupID string, descendants []*models.Group) error {
	key := fmt.Sprintf(GroupDescendantsPattern, groupID)

	if err := c.cache.Set(key, descendants, GroupHierarchyCacheTTL); err != nil {
		c.logger.Warn("Failed to cache group descendants",
			zap.String("group_id", groupID),
			zap.String("cache_key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached group descendants",
		zap.String("group_id", groupID),
		zap.String("cache_key", key),
		zap.Int("descendant_count", len(descendants)))

	return nil
}

// GetCachedGroupDescendants retrieves cached group descendants
func (c *GroupCacheService) GetCachedGroupDescendants(ctx context.Context, groupID string) ([]*models.Group, bool) {
	key := fmt.Sprintf(GroupDescendantsPattern, groupID)

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	if descendants, ok := cached.([]*models.Group); ok {
		c.logger.Debug("Retrieved cached group descendants",
			zap.String("group_id", groupID),
			zap.String("cache_key", key),
			zap.Int("descendant_count", len(descendants)))
		return descendants, true
	}

	c.logger.Warn("Invalid cached descendants type",
		zap.String("group_id", groupID),
		zap.String("cache_key", key))

	// Remove invalid cache entry
	c.cache.Delete(key)
	return nil, false
}

// CacheGroupRoles caches roles assigned to a group
func (c *GroupCacheService) CacheGroupRoles(ctx context.Context, groupID string, roles interface{}, activeOnly bool) error {
	var key string
	if activeOnly {
		key = fmt.Sprintf(GroupActiveRolesPattern, groupID)
	} else {
		key = fmt.Sprintf(GroupRolesPattern, groupID)
	}

	if err := c.cache.Set(key, roles, GroupRolesCacheTTL); err != nil {
		c.logger.Warn("Failed to cache group roles",
			zap.String("group_id", groupID),
			zap.String("cache_key", key),
			zap.Bool("active_only", activeOnly),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached group roles",
		zap.String("group_id", groupID),
		zap.String("cache_key", key),
		zap.Bool("active_only", activeOnly))

	return nil
}

// GetCachedGroupRoles retrieves cached group roles
func (c *GroupCacheService) GetCachedGroupRoles(ctx context.Context, groupID string, activeOnly bool) (interface{}, bool) {
	var key string
	if activeOnly {
		key = fmt.Sprintf(GroupActiveRolesPattern, groupID)
	} else {
		key = fmt.Sprintf(GroupRolesPattern, groupID)
	}

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	c.logger.Debug("Retrieved cached group roles",
		zap.String("group_id", groupID),
		zap.String("cache_key", key),
		zap.Bool("active_only", activeOnly))

	return cached, true
}

// CacheGroupMembers caches members of a group
func (c *GroupCacheService) CacheGroupMembers(ctx context.Context, groupID string, members interface{}, activeOnly bool) error {
	var key string
	if activeOnly {
		key = fmt.Sprintf(GroupActiveMembersPattern, groupID)
	} else {
		key = fmt.Sprintf(GroupMembersPattern, groupID)
	}

	if err := c.cache.Set(key, members, GroupMembersCacheTTL); err != nil {
		c.logger.Warn("Failed to cache group members",
			zap.String("group_id", groupID),
			zap.String("cache_key", key),
			zap.Bool("active_only", activeOnly),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached group members",
		zap.String("group_id", groupID),
		zap.String("cache_key", key),
		zap.Bool("active_only", activeOnly))

	return nil
}

// GetCachedGroupMembers retrieves cached group members
func (c *GroupCacheService) GetCachedGroupMembers(ctx context.Context, groupID string, activeOnly bool) (interface{}, bool) {
	var key string
	if activeOnly {
		key = fmt.Sprintf(GroupActiveMembersPattern, groupID)
	} else {
		key = fmt.Sprintf(GroupMembersPattern, groupID)
	}

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	c.logger.Debug("Retrieved cached group members",
		zap.String("group_id", groupID),
		zap.String("cache_key", key),
		zap.Bool("active_only", activeOnly))

	return cached, true
}

// CacheUserEffectiveRoles caches enhanced effective roles for a user (v2 with more details)
func (c *GroupCacheService) CacheUserEffectiveRoles(ctx context.Context, orgID, userID string, effectiveRoles []*EffectiveRole) error {
	key := fmt.Sprintf(UserEffectiveRolesPattern, orgID, userID)

	if err := c.cache.Set(key, effectiveRoles, EffectiveRolesCacheTTL); err != nil {
		c.logger.Warn("Failed to cache user effective roles",
			zap.String("org_id", orgID),
			zap.String("user_id", userID),
			zap.String("cache_key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached user effective roles",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.String("cache_key", key),
		zap.Int("role_count", len(effectiveRoles)))

	return nil
}

// GetCachedUserEffectiveRoles retrieves cached user effective roles
func (c *GroupCacheService) GetCachedUserEffectiveRoles(ctx context.Context, orgID, userID string) ([]*EffectiveRole, bool) {
	key := fmt.Sprintf(UserEffectiveRolesPattern, orgID, userID)

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	if effectiveRoles, ok := cached.([]*EffectiveRole); ok {
		c.logger.Debug("Retrieved cached user effective roles",
			zap.String("org_id", orgID),
			zap.String("user_id", userID),
			zap.String("cache_key", key),
			zap.Int("role_count", len(effectiveRoles)))
		return effectiveRoles, true
	}

	c.logger.Warn("Invalid cached effective roles type",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.String("cache_key", key))

	// Remove invalid cache entry
	c.cache.Delete(key)
	return nil, false
}

// CacheUserGroupMemberships caches user's group memberships within an organization
func (c *GroupCacheService) CacheUserGroupMemberships(ctx context.Context, orgID, userID string, memberships []*models.GroupMembership) error {
	key := fmt.Sprintf(UserGroupMembershipsPattern, orgID, userID)

	if err := c.cache.Set(key, memberships, GroupMembersCacheTTL); err != nil {
		c.logger.Warn("Failed to cache user group memberships",
			zap.String("org_id", orgID),
			zap.String("user_id", userID),
			zap.String("cache_key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached user group memberships",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.String("cache_key", key),
		zap.Int("membership_count", len(memberships)))

	return nil
}

// GetCachedUserGroupMemberships retrieves cached user group memberships
func (c *GroupCacheService) GetCachedUserGroupMemberships(ctx context.Context, orgID, userID string) ([]*models.GroupMembership, bool) {
	key := fmt.Sprintf(UserGroupMembershipsPattern, orgID, userID)

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	if memberships, ok := cached.([]*models.GroupMembership); ok {
		c.logger.Debug("Retrieved cached user group memberships",
			zap.String("org_id", orgID),
			zap.String("user_id", userID),
			zap.String("cache_key", key),
			zap.Int("membership_count", len(memberships)))
		return memberships, true
	}

	c.logger.Warn("Invalid cached memberships type",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.String("cache_key", key))

	// Remove invalid cache entry
	c.cache.Delete(key)
	return nil, false
}

// CacheGroupRoleInheritance caches role inheritance paths for a group
func (c *GroupCacheService) CacheGroupRoleInheritance(ctx context.Context, groupID string, inheritance map[string]*EffectiveRole) error {
	key := fmt.Sprintf(GroupRoleInheritancePattern, groupID)

	if err := c.cache.Set(key, inheritance, InheritanceCacheTTL); err != nil {
		c.logger.Warn("Failed to cache group role inheritance",
			zap.String("group_id", groupID),
			zap.String("cache_key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached group role inheritance",
		zap.String("group_id", groupID),
		zap.String("cache_key", key),
		zap.Int("role_count", len(inheritance)))

	return nil
}

// GetCachedGroupRoleInheritance retrieves cached group role inheritance
func (c *GroupCacheService) GetCachedGroupRoleInheritance(ctx context.Context, groupID string) (map[string]*EffectiveRole, bool) {
	key := fmt.Sprintf(GroupRoleInheritancePattern, groupID)

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	if inheritance, ok := cached.(map[string]*EffectiveRole); ok {
		c.logger.Debug("Retrieved cached group role inheritance",
			zap.String("group_id", groupID),
			zap.String("cache_key", key),
			zap.Int("role_count", len(inheritance)))
		return inheritance, true
	}

	c.logger.Warn("Invalid cached inheritance type",
		zap.String("group_id", groupID),
		zap.String("cache_key", key))

	// Remove invalid cache entry
	c.cache.Delete(key)
	return nil, false
}

// InvalidateGroupCache invalidates all cache entries for a group
func (c *GroupCacheService) InvalidateGroupCache(ctx context.Context, groupID string) error {
	patterns := []string{
		fmt.Sprintf(GroupHierarchyKeyPattern, groupID),
		fmt.Sprintf(GroupAncestorsPattern, groupID),
		fmt.Sprintf(GroupDescendantsPattern, groupID),
		fmt.Sprintf(GroupChildrenPattern, groupID),
		fmt.Sprintf(GroupActiveChildrenPattern, groupID),
		fmt.Sprintf(GroupRolesPattern, groupID),
		fmt.Sprintf(GroupActiveRolesPattern, groupID),
		fmt.Sprintf(GroupRoleDetailsPattern, groupID),
		fmt.Sprintf(GroupMembersPattern, groupID),
		fmt.Sprintf(GroupActiveMembersPattern, groupID),
		fmt.Sprintf(GroupMemberDetailsPattern, groupID),
		fmt.Sprintf(GroupRoleInheritancePattern, groupID),
	}

	for _, pattern := range patterns {
		if err := c.cache.Delete(pattern); err != nil {
			c.logger.Warn("Failed to invalidate group cache key",
				zap.String("group_id", groupID),
				zap.String("cache_key", pattern),
				zap.Error(err))
		}
	}

	c.logger.Info("Invalidated group cache",
		zap.String("group_id", groupID),
		zap.Int("pattern_count", len(patterns)))

	return nil
}

// InvalidateUserEffectiveRolesCache invalidates effective roles cache for a user
func (c *GroupCacheService) InvalidateUserEffectiveRolesCache(ctx context.Context, orgID, userID string) error {
	patterns := []string{
		fmt.Sprintf(UserEffectiveRolesPattern, orgID, userID),
		fmt.Sprintf(UserGroupMembershipsPattern, orgID, userID),
		// Also invalidate the original effective roles cache key for backward compatibility
		fmt.Sprintf("org:%s:user:%s:effective_roles", orgID, userID),
	}

	for _, pattern := range patterns {
		if err := c.cache.Delete(pattern); err != nil {
			c.logger.Warn("Failed to invalidate user effective roles cache key",
				zap.String("org_id", orgID),
				zap.String("user_id", userID),
				zap.String("cache_key", pattern),
				zap.Error(err))
		}
	}

	c.logger.Debug("Invalidated user effective roles cache",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.Int("pattern_count", len(patterns)))

	return nil
}

// InvalidateHierarchyCache invalidates hierarchy-related cache for a group and its related groups
func (c *GroupCacheService) InvalidateHierarchyCache(ctx context.Context, groupID string, affectedGroupIDs []string) error {
	// Invalidate cache for the main group
	_ = c.InvalidateGroupCache(ctx, groupID)

	// Invalidate cache for all affected groups in the hierarchy
	for _, affectedGroupID := range affectedGroupIDs {
		if affectedGroupID != groupID {
			_ = c.InvalidateGroupCache(ctx, affectedGroupID)
		}
	}

	c.logger.Info("Invalidated hierarchy cache",
		zap.String("group_id", groupID),
		zap.Int("affected_groups", len(affectedGroupIDs)))

	return nil
}

// InvalidateRoleAssignmentCache invalidates cache when roles are assigned/removed from groups
func (c *GroupCacheService) InvalidateRoleAssignmentCache(ctx context.Context, orgID, groupID, roleID string) error {
	// Invalidate group-specific role caches
	groupPatterns := []string{
		fmt.Sprintf(GroupRolesPattern, groupID),
		fmt.Sprintf(GroupActiveRolesPattern, groupID),
		fmt.Sprintf(GroupRoleDetailsPattern, groupID),
		fmt.Sprintf(GroupRoleInheritancePattern, groupID),
	}

	for _, pattern := range groupPatterns {
		if err := c.cache.Delete(pattern); err != nil {
			c.logger.Warn("Failed to invalidate group role cache key",
				zap.String("org_id", orgID),
				zap.String("group_id", groupID),
				zap.String("role_id", roleID),
				zap.String("cache_key", pattern),
				zap.Error(err))
		}
	}

	// Invalidate all user effective roles in the organization since role assignments changed
	userRolePattern := fmt.Sprintf("org:%s:user:*:effective_roles*", orgID)
	keys, err := c.cache.Keys(userRolePattern)
	if err != nil {
		c.logger.Warn("Failed to get user effective roles keys for invalidation",
			zap.String("org_id", orgID),
			zap.String("pattern", userRolePattern),
			zap.Error(err))
	} else {
		for _, key := range keys {
			if err := c.cache.Delete(key); err != nil {
				c.logger.Warn("Failed to invalidate user effective roles cache key",
					zap.String("org_id", orgID),
					zap.String("cache_key", key),
					zap.Error(err))
			}
		}
	}

	c.logger.Info("Invalidated role assignment cache",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("role_id", roleID),
		zap.Int("group_patterns", len(groupPatterns)),
		zap.Int("user_keys", len(keys)))

	return nil
}

// WarmGroupCache pre-loads frequently accessed group data into cache
func (c *GroupCacheService) WarmGroupCache(ctx context.Context, groupID string, groupService interfaces.GroupService) error {
	c.logger.Info("Warming group cache", zap.String("group_id", groupID))

	// Warm group roles cache
	if roles, err := groupService.GetGroupRoles(ctx, groupID); err == nil {
		_ = c.CacheGroupRoles(ctx, groupID, roles, false)
	}

	// Warm group members cache
	if members, err := groupService.GetGroupMembers(ctx, groupID, 100, 0); err == nil {
		_ = c.CacheGroupMembers(ctx, groupID, members, false)
	}

	c.logger.Info("Group cache warming completed", zap.String("group_id", groupID))
	return nil
}

// ScheduleGroupCacheWarming sets up periodic cache warming for frequently accessed groups
func (c *GroupCacheService) ScheduleGroupCacheWarming(ctx context.Context, groupIDs []string, groupService interfaces.GroupService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Group cache warming stopped due to context cancellation")
				return
			case <-ticker.C:
				for _, groupID := range groupIDs {
					if err := c.WarmGroupCache(ctx, groupID, groupService); err != nil {
						c.logger.Warn("Failed to warm cache for group",
							zap.String("group_id", groupID),
							zap.Error(err))
					}
				}
			}
		}
	}()

	c.logger.Info("Group cache warming scheduled",
		zap.Int("group_count", len(groupIDs)),
		zap.Duration("interval", interval))
}
