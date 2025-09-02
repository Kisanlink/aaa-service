package organizations

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	organizationResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/organizations"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"go.uber.org/zap"
)

// OrganizationCacheService provides caching functionality for organization-related operations
type OrganizationCacheService struct {
	cache  interfaces.CacheService
	logger *zap.Logger
}

// NewOrganizationCacheService creates a new organization cache service
func NewOrganizationCacheService(cache interfaces.CacheService, logger *zap.Logger) *OrganizationCacheService {
	return &OrganizationCacheService{
		cache:  cache,
		logger: logger,
	}
}

// Cache key patterns for organization data
const (
	// Organization hierarchy caching
	OrgHierarchyKeyPattern    = "org:%s:hierarchy"
	OrgParentHierarchyPattern = "org:%s:parent_hierarchy"
	OrgChildrenPattern        = "org:%s:children"
	OrgActiveChildrenPattern  = "org:%s:active_children"

	// Organization groups caching
	OrgGroupsPattern         = "org:%s:groups"
	OrgActiveGroupsPattern   = "org:%s:active_groups"
	OrgGroupHierarchyPattern = "org:%s:group_hierarchy"

	// User group memberships within organization
	OrgUserGroupsPattern       = "org:%s:user:%s:groups"
	OrgUserActiveGroupsPattern = "org:%s:user:%s:active_groups"

	// Group members within organization
	OrgGroupMembersPattern       = "org:%s:group:%s:members"
	OrgGroupActiveMembersPattern = "org:%s:group:%s:active_members"

	// Organization stats
	OrgStatsPattern = "org:%s:stats"

	// Cache TTL values (in seconds)
	HierarchyCacheTTL  = 1800 // 30 minutes for hierarchy data
	GroupsCacheTTL     = 900  // 15 minutes for group data
	UserGroupsCacheTTL = 600  // 10 minutes for user-group relationships
	StatsCacheTTL      = 300  // 5 minutes for stats
)

// CacheOrganizationHierarchy caches the complete organization hierarchy
func (c *OrganizationCacheService) CacheOrganizationHierarchy(ctx context.Context, orgID string, hierarchy *organizationResponses.OrganizationHierarchyResponse) error {
	key := fmt.Sprintf(OrgHierarchyKeyPattern, orgID)

	if err := c.cache.Set(key, hierarchy, HierarchyCacheTTL); err != nil {
		c.logger.Warn("Failed to cache organization hierarchy",
			zap.String("org_id", orgID),
			zap.String("cache_key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached organization hierarchy",
		zap.String("org_id", orgID),
		zap.String("cache_key", key))

	return nil
}

// GetCachedOrganizationHierarchy retrieves cached organization hierarchy
func (c *OrganizationCacheService) GetCachedOrganizationHierarchy(ctx context.Context, orgID string) (*organizationResponses.OrganizationHierarchyResponse, bool) {
	key := fmt.Sprintf(OrgHierarchyKeyPattern, orgID)

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	if hierarchy, ok := cached.(*organizationResponses.OrganizationHierarchyResponse); ok {
		c.logger.Debug("Retrieved cached organization hierarchy",
			zap.String("org_id", orgID),
			zap.String("cache_key", key))
		return hierarchy, true
	}

	c.logger.Warn("Invalid cached organization hierarchy type",
		zap.String("org_id", orgID),
		zap.String("cache_key", key))

	// Remove invalid cache entry
	c.cache.Delete(key)
	return nil, false
}

// CacheOrganizationParentHierarchy caches the parent hierarchy for an organization
func (c *OrganizationCacheService) CacheOrganizationParentHierarchy(ctx context.Context, orgID string, parents []*models.Organization) error {
	key := fmt.Sprintf(OrgParentHierarchyPattern, orgID)

	if err := c.cache.Set(key, parents, HierarchyCacheTTL); err != nil {
		c.logger.Warn("Failed to cache organization parent hierarchy",
			zap.String("org_id", orgID),
			zap.String("cache_key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached organization parent hierarchy",
		zap.String("org_id", orgID),
		zap.String("cache_key", key),
		zap.Int("parent_count", len(parents)))

	return nil
}

// GetCachedOrganizationParentHierarchy retrieves cached parent hierarchy
func (c *OrganizationCacheService) GetCachedOrganizationParentHierarchy(ctx context.Context, orgID string) ([]*models.Organization, bool) {
	key := fmt.Sprintf(OrgParentHierarchyPattern, orgID)

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	if parents, ok := cached.([]*models.Organization); ok {
		c.logger.Debug("Retrieved cached organization parent hierarchy",
			zap.String("org_id", orgID),
			zap.String("cache_key", key),
			zap.Int("parent_count", len(parents)))
		return parents, true
	}

	c.logger.Warn("Invalid cached parent hierarchy type",
		zap.String("org_id", orgID),
		zap.String("cache_key", key))

	// Remove invalid cache entry
	c.cache.Delete(key)
	return nil, false
}

// CacheOrganizationChildren caches the children of an organization
func (c *OrganizationCacheService) CacheOrganizationChildren(ctx context.Context, orgID string, children []*models.Organization, activeOnly bool) error {
	var key string
	if activeOnly {
		key = fmt.Sprintf(OrgActiveChildrenPattern, orgID)
	} else {
		key = fmt.Sprintf(OrgChildrenPattern, orgID)
	}

	if err := c.cache.Set(key, children, HierarchyCacheTTL); err != nil {
		c.logger.Warn("Failed to cache organization children",
			zap.String("org_id", orgID),
			zap.String("cache_key", key),
			zap.Bool("active_only", activeOnly),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached organization children",
		zap.String("org_id", orgID),
		zap.String("cache_key", key),
		zap.Bool("active_only", activeOnly),
		zap.Int("children_count", len(children)))

	return nil
}

// GetCachedOrganizationChildren retrieves cached organization children
func (c *OrganizationCacheService) GetCachedOrganizationChildren(ctx context.Context, orgID string, activeOnly bool) ([]*models.Organization, bool) {
	var key string
	if activeOnly {
		key = fmt.Sprintf(OrgActiveChildrenPattern, orgID)
	} else {
		key = fmt.Sprintf(OrgChildrenPattern, orgID)
	}

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	if children, ok := cached.([]*models.Organization); ok {
		c.logger.Debug("Retrieved cached organization children",
			zap.String("org_id", orgID),
			zap.String("cache_key", key),
			zap.Bool("active_only", activeOnly),
			zap.Int("children_count", len(children)))
		return children, true
	}

	c.logger.Warn("Invalid cached children type",
		zap.String("org_id", orgID),
		zap.String("cache_key", key))

	// Remove invalid cache entry
	c.cache.Delete(key)
	return nil, false
}

// CacheOrganizationGroups caches groups within an organization
func (c *OrganizationCacheService) CacheOrganizationGroups(ctx context.Context, orgID string, groups interface{}, activeOnly bool) error {
	var key string
	if activeOnly {
		key = fmt.Sprintf(OrgActiveGroupsPattern, orgID)
	} else {
		key = fmt.Sprintf(OrgGroupsPattern, orgID)
	}

	if err := c.cache.Set(key, groups, GroupsCacheTTL); err != nil {
		c.logger.Warn("Failed to cache organization groups",
			zap.String("org_id", orgID),
			zap.String("cache_key", key),
			zap.Bool("active_only", activeOnly),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached organization groups",
		zap.String("org_id", orgID),
		zap.String("cache_key", key),
		zap.Bool("active_only", activeOnly))

	return nil
}

// GetCachedOrganizationGroups retrieves cached organization groups
func (c *OrganizationCacheService) GetCachedOrganizationGroups(ctx context.Context, orgID string, activeOnly bool) (interface{}, bool) {
	var key string
	if activeOnly {
		key = fmt.Sprintf(OrgActiveGroupsPattern, orgID)
	} else {
		key = fmt.Sprintf(OrgGroupsPattern, orgID)
	}

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	c.logger.Debug("Retrieved cached organization groups",
		zap.String("org_id", orgID),
		zap.String("cache_key", key),
		zap.Bool("active_only", activeOnly))

	return cached, true
}

// CacheUserGroupsInOrganization caches user's groups within an organization
func (c *OrganizationCacheService) CacheUserGroupsInOrganization(ctx context.Context, orgID, userID string, groups interface{}, activeOnly bool) error {
	var key string
	if activeOnly {
		key = fmt.Sprintf(OrgUserActiveGroupsPattern, orgID, userID)
	} else {
		key = fmt.Sprintf(OrgUserGroupsPattern, orgID, userID)
	}

	if err := c.cache.Set(key, groups, UserGroupsCacheTTL); err != nil {
		c.logger.Warn("Failed to cache user groups in organization",
			zap.String("org_id", orgID),
			zap.String("user_id", userID),
			zap.String("cache_key", key),
			zap.Bool("active_only", activeOnly),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached user groups in organization",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.String("cache_key", key),
		zap.Bool("active_only", activeOnly))

	return nil
}

// GetCachedUserGroupsInOrganization retrieves cached user groups within an organization
func (c *OrganizationCacheService) GetCachedUserGroupsInOrganization(ctx context.Context, orgID, userID string, activeOnly bool) (interface{}, bool) {
	var key string
	if activeOnly {
		key = fmt.Sprintf(OrgUserActiveGroupsPattern, orgID, userID)
	} else {
		key = fmt.Sprintf(OrgUserGroupsPattern, orgID, userID)
	}

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	c.logger.Debug("Retrieved cached user groups in organization",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.String("cache_key", key),
		zap.Bool("active_only", activeOnly))

	return cached, true
}

// CacheOrganizationStats caches organization statistics
func (c *OrganizationCacheService) CacheOrganizationStats(ctx context.Context, orgID string, stats *organizationResponses.OrganizationStatsResponse) error {
	key := fmt.Sprintf(OrgStatsPattern, orgID)

	if err := c.cache.Set(key, stats, StatsCacheTTL); err != nil {
		c.logger.Warn("Failed to cache organization stats",
			zap.String("org_id", orgID),
			zap.String("cache_key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached organization stats",
		zap.String("org_id", orgID),
		zap.String("cache_key", key))

	return nil
}

// GetCachedOrganizationStats retrieves cached organization statistics
func (c *OrganizationCacheService) GetCachedOrganizationStats(ctx context.Context, orgID string) (*organizationResponses.OrganizationStatsResponse, bool) {
	key := fmt.Sprintf(OrgStatsPattern, orgID)

	cached, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	if stats, ok := cached.(*organizationResponses.OrganizationStatsResponse); ok {
		c.logger.Debug("Retrieved cached organization stats",
			zap.String("org_id", orgID),
			zap.String("cache_key", key))
		return stats, true
	}

	c.logger.Warn("Invalid cached stats type",
		zap.String("org_id", orgID),
		zap.String("cache_key", key))

	// Remove invalid cache entry
	c.cache.Delete(key)
	return nil, false
}

// InvalidateOrganizationCache invalidates all cache entries for an organization
func (c *OrganizationCacheService) InvalidateOrganizationCache(ctx context.Context, orgID string) error {
	patterns := []string{
		fmt.Sprintf(OrgHierarchyKeyPattern, orgID),
		fmt.Sprintf(OrgParentHierarchyPattern, orgID),
		fmt.Sprintf(OrgChildrenPattern, orgID),
		fmt.Sprintf(OrgActiveChildrenPattern, orgID),
		fmt.Sprintf(OrgGroupsPattern, orgID),
		fmt.Sprintf(OrgActiveGroupsPattern, orgID),
		fmt.Sprintf(OrgGroupHierarchyPattern, orgID),
		fmt.Sprintf(OrgStatsPattern, orgID),
	}

	for _, pattern := range patterns {
		if err := c.cache.Delete(pattern); err != nil {
			c.logger.Warn("Failed to invalidate cache key",
				zap.String("org_id", orgID),
				zap.String("cache_key", pattern),
				zap.Error(err))
		}
	}

	// Invalidate user-specific caches for this organization
	userGroupPattern := fmt.Sprintf("org:%s:user:*", orgID)
	keys, err := c.cache.Keys(userGroupPattern)
	if err != nil {
		c.logger.Warn("Failed to get user cache keys for invalidation",
			zap.String("org_id", orgID),
			zap.String("pattern", userGroupPattern),
			zap.Error(err))
	} else {
		for _, key := range keys {
			if err := c.cache.Delete(key); err != nil {
				c.logger.Warn("Failed to invalidate user cache key",
					zap.String("org_id", orgID),
					zap.String("cache_key", key),
					zap.Error(err))
			}
		}
	}

	c.logger.Info("Invalidated organization cache",
		zap.String("org_id", orgID),
		zap.Int("pattern_count", len(patterns)))

	return nil
}

// InvalidateUserGroupCache invalidates cache entries for a specific user in an organization
func (c *OrganizationCacheService) InvalidateUserGroupCache(ctx context.Context, orgID, userID string) error {
	patterns := []string{
		fmt.Sprintf(OrgUserGroupsPattern, orgID, userID),
		fmt.Sprintf(OrgUserActiveGroupsPattern, orgID, userID),
	}

	for _, pattern := range patterns {
		if err := c.cache.Delete(pattern); err != nil {
			c.logger.Warn("Failed to invalidate user group cache key",
				zap.String("org_id", orgID),
				zap.String("user_id", userID),
				zap.String("cache_key", pattern),
				zap.Error(err))
		}
	}

	c.logger.Debug("Invalidated user group cache",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.Int("pattern_count", len(patterns)))

	return nil
}

// InvalidateGroupCache invalidates cache entries for a specific group in an organization
func (c *OrganizationCacheService) InvalidateGroupCache(ctx context.Context, orgID, groupID string) error {
	patterns := []string{
		fmt.Sprintf(OrgGroupMembersPattern, orgID, groupID),
		fmt.Sprintf(OrgGroupActiveMembersPattern, orgID, groupID),
	}

	for _, pattern := range patterns {
		if err := c.cache.Delete(pattern); err != nil {
			c.logger.Warn("Failed to invalidate group cache key",
				zap.String("org_id", orgID),
				zap.String("group_id", groupID),
				zap.String("cache_key", pattern),
				zap.Error(err))
		}
	}

	// Also invalidate organization-level group caches
	orgGroupPatterns := []string{
		fmt.Sprintf(OrgGroupsPattern, orgID),
		fmt.Sprintf(OrgActiveGroupsPattern, orgID),
		fmt.Sprintf(OrgGroupHierarchyPattern, orgID),
	}

	for _, pattern := range orgGroupPatterns {
		if err := c.cache.Delete(pattern); err != nil {
			c.logger.Warn("Failed to invalidate organization group cache key",
				zap.String("org_id", orgID),
				zap.String("group_id", groupID),
				zap.String("cache_key", pattern),
				zap.Error(err))
		}
	}

	c.logger.Debug("Invalidated group cache",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.Int("pattern_count", len(patterns)+len(orgGroupPatterns)))

	return nil
}

// WarmOrganizationCache pre-loads frequently accessed organization data into cache
func (c *OrganizationCacheService) WarmOrganizationCache(ctx context.Context, orgID string, orgService interfaces.OrganizationService) error {
	c.logger.Info("Warming organization cache", zap.String("org_id", orgID))

	// Warm hierarchy cache
	if hierarchy, err := orgService.GetOrganizationHierarchy(ctx, orgID); err == nil {
		if hierarchyResp, ok := hierarchy.(*organizationResponses.OrganizationHierarchyResponse); ok {
			c.CacheOrganizationHierarchy(ctx, orgID, hierarchyResp)
		}
	}

	// Warm groups cache
	if groups, err := orgService.GetOrganizationGroups(ctx, orgID, 100, 0, false); err == nil {
		c.CacheOrganizationGroups(ctx, orgID, groups, false)
	}

	// Warm active groups cache
	if activeGroups, err := orgService.GetOrganizationGroups(ctx, orgID, 100, 0, true); err == nil {
		c.CacheOrganizationGroups(ctx, orgID, activeGroups, true)
	}

	// Warm stats cache
	if stats, err := orgService.GetOrganizationStats(ctx, orgID); err == nil {
		if statsResp, ok := stats.(*organizationResponses.OrganizationStatsResponse); ok {
			c.CacheOrganizationStats(ctx, orgID, statsResp)
		}
	}

	c.logger.Info("Organization cache warming completed", zap.String("org_id", orgID))
	return nil
}

// ScheduleCacheWarming sets up periodic cache warming for frequently accessed organizations
func (c *OrganizationCacheService) ScheduleCacheWarming(ctx context.Context, orgIDs []string, orgService interfaces.OrganizationService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Cache warming stopped due to context cancellation")
				return
			case <-ticker.C:
				for _, orgID := range orgIDs {
					if err := c.WarmOrganizationCache(ctx, orgID, orgService); err != nil {
						c.logger.Warn("Failed to warm cache for organization",
							zap.String("org_id", orgID),
							zap.Error(err))
					}
				}
			}
		}
	}()

	c.logger.Info("Cache warming scheduled",
		zap.Int("org_count", len(orgIDs)),
		zap.Duration("interval", interval))
}
