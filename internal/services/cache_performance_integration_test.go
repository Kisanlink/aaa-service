//go:build integration
// +build integration

package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	organizationResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/organizations"
	"github.com/Kisanlink/aaa-service/internal/services/groups"
	"github.com/Kisanlink/aaa-service/internal/services/organizations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockCacheServiceComplete implements all methods of interfaces.CacheService
type MockCacheServiceComplete struct {
	mock.Mock
}

func (m *MockCacheServiceComplete) Get(key string) (interface{}, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

func (m *MockCacheServiceComplete) Set(key string, value interface{}, ttl int) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheServiceComplete) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCacheServiceComplete) Exists(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockCacheServiceComplete) Clear() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheServiceComplete) Keys(pattern string) ([]string, error) {
	args := m.Called(pattern)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheServiceComplete) Expire(key string, ttl int) error {
	args := m.Called(key, ttl)
	return args.Error(0)
}

func (m *MockCacheServiceComplete) TTL(key string) (int, error) {
	args := m.Called(key)
	return args.Int(0), args.Error(1)
}

func (m *MockCacheServiceComplete) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestCachePerformanceOptimization tests the complete caching layer for performance optimization
func TestCachePerformanceOptimization(t *testing.T) {
	// Setup
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	// Create cache services
	orgCache := organizations.NewOrganizationCacheService(mockCache, logger)
	groupCache := groups.NewGroupCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "perf-test-org"
	groupID := "perf-test-group"
	userID := "perf-test-user"

	t.Run("Organization Hierarchy Caching Performance", func(t *testing.T) {
		// Test data
		hierarchy := &organizationResponses.OrganizationHierarchyResponse{
			Organization: &organizationResponses.OrganizationResponse{
				ID:   orgID,
				Name: "Performance Test Organization",
			},
		}

		// Mock cache operations
		hierarchyKey := fmt.Sprintf("org:%s:hierarchy", orgID)
		mockCache.On("Set", hierarchyKey, hierarchy, organizations.HierarchyCacheTTL).Return(nil)
		mockCache.On("Get", hierarchyKey).Return(hierarchy, true)

		// Test caching
		err := orgCache.CacheOrganizationHierarchy(ctx, orgID, hierarchy)
		assert.NoError(t, err)

		// Test retrieval (should be fast from cache)
		cached, found := orgCache.GetCachedOrganizationHierarchy(ctx, orgID)
		assert.True(t, found)
		assert.Equal(t, hierarchy, cached)

		mockCache.AssertExpectations(t)
	})

	t.Run("User Effective Roles Caching Performance", func(t *testing.T) {
		// Test data
		effectiveRoles := []*groups.EffectiveRole{
			{
				Role: &models.Role{
					Name:     "Performance Test Role",
					IsActive: true,
				},
				GroupID:      groupID,
				GroupName:    "Performance Test Group",
				Distance:     0,
				IsDirectRole: true,
			},
		}

		// Mock cache operations
		effectiveRolesKey := fmt.Sprintf("org:%s:user:%s:effective_roles_v2", orgID, userID)
		mockCache.On("Set", effectiveRolesKey, effectiveRoles, groups.EffectiveRolesCacheTTL).Return(nil)
		mockCache.On("Get", effectiveRolesKey).Return(effectiveRoles, true)

		// Test caching
		err := groupCache.CacheUserEffectiveRoles(ctx, orgID, userID, effectiveRoles)
		assert.NoError(t, err)

		// Test retrieval (should be fast from cache)
		cached, found := groupCache.GetCachedUserEffectiveRoles(ctx, orgID, userID)
		assert.True(t, found)
		assert.Equal(t, effectiveRoles, cached)

		mockCache.AssertExpectations(t)
	})

	t.Run("Group Roles Caching Performance", func(t *testing.T) {
		// Test data
		groupRoles := []interface{}{
			map[string]interface{}{
				"role_id":   "role-1",
				"role_name": "Test Role 1",
				"group_id":  groupID,
			},
		}

		// Mock cache operations
		groupRolesKey := fmt.Sprintf("group:%s:active_roles", groupID)
		mockCache.On("Set", groupRolesKey, groupRoles, groups.GroupRolesCacheTTL).Return(nil)
		mockCache.On("Get", groupRolesKey).Return(groupRoles, true)

		// Test caching
		err := groupCache.CacheGroupRoles(ctx, groupID, groupRoles, true)
		assert.NoError(t, err)

		// Test retrieval (should be fast from cache)
		cached, found := groupCache.GetCachedGroupRoles(ctx, groupID, true)
		assert.True(t, found)
		assert.Equal(t, groupRoles, cached)

		mockCache.AssertExpectations(t)
	})
}

// TestCacheInvalidationStrategies tests cache invalidation for hierarchy and role changes
func TestCacheInvalidationStrategies(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	orgCache := organizations.NewOrganizationCacheService(mockCache, logger)
	groupCache := groups.NewGroupCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "invalidation-test-org"
	groupID := "invalidation-test-group"
	roleID := "invalidation-test-role"

	t.Run("Organization Hierarchy Change Invalidation", func(t *testing.T) {
		// Mock cache invalidation for organization hierarchy changes
		expectedKeys := []string{
			fmt.Sprintf("org:%s:hierarchy", orgID),
			fmt.Sprintf("org:%s:parent_hierarchy", orgID),
			fmt.Sprintf("org:%s:children", orgID),
			fmt.Sprintf("org:%s:active_children", orgID),
			fmt.Sprintf("org:%s:groups", orgID),
			fmt.Sprintf("org:%s:active_groups", orgID),
			fmt.Sprintf("org:%s:group_hierarchy", orgID),
			fmt.Sprintf("org:%s:stats", orgID),
		}

		for _, key := range expectedKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		// Mock user cache invalidation
		userPattern := fmt.Sprintf("org:%s:user:*", orgID)
		userKeys := []string{
			fmt.Sprintf("org:%s:user:user1:groups", orgID),
			fmt.Sprintf("org:%s:user:user2:active_groups", orgID),
		}
		mockCache.On("Keys", userPattern).Return(userKeys, nil)

		for _, key := range userKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		// Test invalidation
		err := orgCache.InvalidateOrganizationCache(ctx, orgID)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("Role Assignment Change Invalidation", func(t *testing.T) {
		// Mock cache invalidation for role assignment changes
		groupPatterns := []string{
			fmt.Sprintf("group:%s:roles", groupID),
			fmt.Sprintf("group:%s:active_roles", groupID),
			fmt.Sprintf("group:%s:role_details", groupID),
			fmt.Sprintf("group:%s:role_inheritance", groupID),
		}

		for _, pattern := range groupPatterns {
			mockCache.On("Delete", pattern).Return(nil)
		}

		// Mock user effective roles invalidation
		userRolePattern := fmt.Sprintf("org:%s:user:*:effective_roles*", orgID)
		userRoleKeys := []string{
			fmt.Sprintf("org:%s:user:user1:effective_roles_v2", orgID),
			fmt.Sprintf("org:%s:user:user2:effective_roles", orgID),
		}
		mockCache.On("Keys", userRolePattern).Return(userRoleKeys, nil)

		for _, key := range userRoleKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		// Test invalidation
		err := groupCache.InvalidateRoleAssignmentCache(ctx, orgID, groupID, roleID)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("Group Hierarchy Change Invalidation", func(t *testing.T) {
		affectedGroupIDs := []string{"child-group-1", "child-group-2"}

		// Mock cache invalidation for main group
		mainGroupPatterns := []string{
			fmt.Sprintf("group:%s:hierarchy", groupID),
			fmt.Sprintf("group:%s:ancestors", groupID),
			fmt.Sprintf("group:%s:descendants", groupID),
			fmt.Sprintf("group:%s:children", groupID),
			fmt.Sprintf("group:%s:active_children", groupID),
			fmt.Sprintf("group:%s:roles", groupID),
			fmt.Sprintf("group:%s:active_roles", groupID),
			fmt.Sprintf("group:%s:role_details", groupID),
			fmt.Sprintf("group:%s:members", groupID),
			fmt.Sprintf("group:%s:active_members", groupID),
			fmt.Sprintf("group:%s:member_details", groupID),
			fmt.Sprintf("group:%s:role_inheritance", groupID),
		}

		for _, pattern := range mainGroupPatterns {
			mockCache.On("Delete", pattern).Return(nil)
		}

		// Mock cache invalidation for affected groups
		for _, affectedGroupID := range affectedGroupIDs {
			for _, pattern := range mainGroupPatterns {
				affectedKey := fmt.Sprintf(pattern, affectedGroupID)
				mockCache.On("Delete", affectedKey).Return(nil)
			}
		}

		// Test invalidation
		err := groupCache.InvalidateHierarchyCache(ctx, groupID, affectedGroupIDs)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})
}

// TestCacheWarmingIntegration tests cache warming for frequently accessed organization data
func TestCacheWarmingIntegration(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	mockOrgService := &MockOrganizationService{}
	mockGroupService := &MockGroupService{}
	logger := zap.NewNop()

	warmingService := NewCacheWarmingService(mockOrgService, mockGroupService, mockCache, logger)

	ctx := context.Background()
	config := CacheWarmingConfig{
		WarmingInterval:       1 * time.Hour, // Long interval to avoid automatic warming
		MaxConcurrentWarms:    3,
		FrequentOrganizations: []string{"warm-org-1", "warm-org-2"},
		FrequentGroups:        []string{"warm-group-1", "warm-group-2"},
		FrequentUsers:         []string{"warm-user-1", "warm-user-2"},
		WarmHierarchies:       true,
		WarmGroupRoles:        true,
		WarmEffectiveRoles:    true,
		WarmStats:             true,
	}

	t.Run("Cache Warming Performance", func(t *testing.T) {
		// Mock service calls for warming
		for _, orgID := range config.FrequentOrganizations {
			hierarchy := &organizationResponses.OrganizationHierarchyResponse{
				Organization: &organizationResponses.OrganizationResponse{
					ID:   orgID,
					Name: fmt.Sprintf("Warm Test Organization %s", orgID),
				},
			}
			stats := &organizationResponses.OrganizationStatsResponse{
				OrganizationID: orgID,
				ChildCount:     2,
				GroupCount:     5,
				UserCount:      10,
			}

			mockOrgService.On("GetOrganizationHierarchy", ctx, orgID).Return(hierarchy, nil)
			mockOrgService.On("GetOrganizationStats", ctx, orgID).Return(stats, nil)
			mockOrgService.On("GetOrganizationGroups", ctx, orgID, 100, 0, false).Return([]interface{}{}, nil)
			mockOrgService.On("GetOrganizationGroups", ctx, orgID, 100, 0, true).Return([]interface{}{}, nil)
		}

		for _, groupID := range config.FrequentGroups {
			mockGroupService.On("GetGroupRoles", ctx, groupID).Return([]interface{}{}, nil)
			mockGroupService.On("GetGroupMembers", ctx, groupID, 100, 0).Return([]interface{}{}, nil)
		}

		for _, orgID := range config.FrequentOrganizations {
			for _, userID := range config.FrequentUsers {
				effectiveRoles := []*groups.EffectiveRole{
					{
						Role: &models.Role{
							Name:     fmt.Sprintf("Warm Test Role for %s", userID),
							IsActive: true,
						},
						GroupID:      "warm-group-1",
						GroupName:    "Warm Test Group",
						Distance:     0,
						IsDirectRole: true,
					},
				}
				mockGroupService.On("GetUserEffectiveRoles", ctx, orgID, userID).Return(effectiveRoles, nil)
			}
		}

		// Measure warming performance
		startTime := time.Now()
		err := warmingService.WarmNow(ctx, config)
		warmingDuration := time.Since(startTime)

		assert.NoError(t, err)
		assert.Less(t, warmingDuration, 5*time.Second, "Cache warming should complete within 5 seconds")

		mockOrgService.AssertExpectations(t)
		mockGroupService.AssertExpectations(t)
	})

	t.Run("Cache Statistics Validation", func(t *testing.T) {
		// Mock cache keys for statistics
		orgKeys := []string{"org:warm-org-1:hierarchy", "org:warm-org-2:stats"}
		groupKeys := []string{"group:warm-group-1:roles", "group:warm-group-2:members"}
		effectiveRolesKeys := []string{"org:warm-org-1:user:warm-user-1:effective_roles_v2"}

		mockCache.On("Keys", "org:*").Return(orgKeys, nil)
		mockCache.On("Keys", "group:*").Return(groupKeys, nil)
		mockCache.On("Keys", "*:effective_roles*").Return(effectiveRolesKeys, nil)

		stats, err := warmingService.GetCacheStats(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Validate cache statistics
		assert.Equal(t, len(orgKeys), stats["organization_cache_keys"])
		assert.Equal(t, len(groupKeys), stats["group_cache_keys"])
		assert.Equal(t, len(effectiveRolesKeys), stats["effective_roles_cache_keys"])
		assert.Equal(t, false, stats["is_running"])
		assert.Equal(t, "1h0m0s", stats["warming_interval"])
		assert.Equal(t, 3, stats["max_concurrent_warms"])

		mockCache.AssertExpectations(t)
	})
}

// TestCacheTTLAndExpiration tests cache TTL values and expiration behavior
func TestCacheTTLAndExpiration(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	orgCache := organizations.NewOrganizationCacheService(mockCache, logger)
	groupCache := groups.NewGroupCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "ttl-test-org"
	groupID := "ttl-test-group"
	userID := "ttl-test-user"

	t.Run("Organization Cache TTL Values", func(t *testing.T) {
		hierarchy := &organizationResponses.OrganizationHierarchyResponse{
			Organization: &organizationResponses.OrganizationResponse{
				ID:   orgID,
				Name: "TTL Test Organization",
			},
		}

		// Test hierarchy cache TTL (30 minutes)
		hierarchyKey := fmt.Sprintf("org:%s:hierarchy", orgID)
		mockCache.On("Set", hierarchyKey, hierarchy, organizations.HierarchyCacheTTL).Return(nil)

		err := orgCache.CacheOrganizationHierarchy(ctx, orgID, hierarchy)
		assert.NoError(t, err)

		// Verify TTL is set correctly
		assert.Equal(t, 1800, organizations.HierarchyCacheTTL) // 30 minutes

		mockCache.AssertExpectations(t)
	})

	t.Run("Group Cache TTL Values", func(t *testing.T) {
		groupRoles := []interface{}{
			map[string]interface{}{
				"role_id":   "role-1",
				"role_name": "TTL Test Role",
			},
		}

		// Test group roles cache TTL (15 minutes)
		groupRolesKey := fmt.Sprintf("group:%s:active_roles", groupID)
		mockCache.On("Set", groupRolesKey, groupRoles, groups.GroupRolesCacheTTL).Return(nil)

		err := groupCache.CacheGroupRoles(ctx, groupID, groupRoles, true)
		assert.NoError(t, err)

		// Verify TTL is set correctly
		assert.Equal(t, 900, groups.GroupRolesCacheTTL) // 15 minutes

		mockCache.AssertExpectations(t)
	})

	t.Run("Effective Roles Cache TTL Values", func(t *testing.T) {
		effectiveRoles := []*groups.EffectiveRole{
			{
				Role: &models.Role{
					Name:     "TTL Test Role",
					IsActive: true,
				},
				GroupID:      groupID,
				GroupName:    "TTL Test Group",
				Distance:     0,
				IsDirectRole: true,
			},
		}

		// Test effective roles cache TTL (5 minutes)
		effectiveRolesKey := fmt.Sprintf("org:%s:user:%s:effective_roles_v2", orgID, userID)
		mockCache.On("Set", effectiveRolesKey, effectiveRoles, groups.EffectiveRolesCacheTTL).Return(nil)

		err := groupCache.CacheUserEffectiveRoles(ctx, orgID, userID, effectiveRoles)
		assert.NoError(t, err)

		// Verify TTL is set correctly (shorter TTL for frequently changing data)
		assert.Equal(t, 300, groups.EffectiveRolesCacheTTL) // 5 minutes

		mockCache.AssertExpectations(t)
	})
}

// TestCacheErrorHandling tests cache error handling and fallback behavior
func TestCacheErrorHandling(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	orgCache := organizations.NewOrganizationCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "error-test-org"

	t.Run("Cache Set Error Handling", func(t *testing.T) {
		hierarchy := &organizationResponses.OrganizationHierarchyResponse{
			Organization: &organizationResponses.OrganizationResponse{
				ID:   orgID,
				Name: "Error Test Organization",
			},
		}

		// Mock cache set error
		hierarchyKey := fmt.Sprintf("org:%s:hierarchy", orgID)
		mockCache.On("Set", hierarchyKey, hierarchy, organizations.HierarchyCacheTTL).Return(fmt.Errorf("cache set error"))

		// Should handle error gracefully
		err := orgCache.CacheOrganizationHierarchy(ctx, orgID, hierarchy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache set error")

		mockCache.AssertExpectations(t)
	})

	t.Run("Cache Get Error Handling", func(t *testing.T) {
		// Mock cache get returning invalid data type
		hierarchyKey := fmt.Sprintf("org:%s:hierarchy", orgID)
		mockCache.On("Get", hierarchyKey).Return("invalid-data", true)
		mockCache.On("Delete", hierarchyKey).Return(nil) // Should clean up invalid cache

		// Should handle invalid cache data gracefully
		cached, found := orgCache.GetCachedOrganizationHierarchy(ctx, orgID)
		assert.False(t, found)
		assert.Nil(t, cached)

		mockCache.AssertExpectations(t)
	})

	t.Run("Cache Delete Error Handling", func(t *testing.T) {
		// Mock cache delete error
		patterns := []string{
			fmt.Sprintf("org:%s:hierarchy", orgID),
			fmt.Sprintf("org:%s:parent_hierarchy", orgID),
		}

		for _, pattern := range patterns {
			mockCache.On("Delete", pattern).Return(fmt.Errorf("cache delete error"))
		}

		userPattern := fmt.Sprintf("org:%s:user:*", orgID)
		mockCache.On("Keys", userPattern).Return([]string{}, nil)

		// Should handle delete errors gracefully and continue
		err := orgCache.InvalidateOrganizationCache(ctx, orgID)
		assert.NoError(t, err) // Should not fail the operation

		mockCache.AssertExpectations(t)
	})
}
