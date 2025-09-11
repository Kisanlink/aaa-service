//go:build integration
// +build integration

package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCacheInvalidationService_Integration(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	invalidationService := NewCacheInvalidationService(mockCache, logger)
	ctx := context.Background()

	t.Run("Organization Created Event", func(t *testing.T) {
		// Test organization creation invalidation
		event := invalidationService.CreateInvalidationEvent(
			"organization_created",
			"new-org-123",
			"organization",
			"",
			[]string{},
			map[string]interface{}{
				"parent_id": "parent-org-456",
			},
		)

		// Mock parent organization cache invalidation
		parentOrgKeys := []string{
			"org:parent-org-456:hierarchy",
			"org:parent-org-456:parent_hierarchy",
			"org:parent-org-456:children",
			"org:parent-org-456:active_children",
			"org:parent-org-456:groups",
			"org:parent-org-456:active_groups",
			"org:parent-org-456:group_hierarchy",
			"org:parent-org-456:stats",
		}

		for _, key := range parentOrgKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		userPattern := "org:parent-org-456:user:*"
		mockCache.On("Keys", userPattern).Return([]string{}, nil)

		err := invalidationService.InvalidateCache(ctx, event)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("Organization Hierarchy Changed Event", func(t *testing.T) {
		event := invalidationService.CreateInvalidationEvent(
			"organization_hierarchy_changed",
			"main-org-123",
			"organization",
			"",
			[]string{"affected-org-1", "affected-org-2"},
			map[string]interface{}{},
		)

		// Mock hierarchy-related cache invalidation for main org
		mainOrgKeys := []string{
			"org:main-org-123:hierarchy_tree",
			"org:main-org-123:hierarchy",
			"org:main-org-123:parent_hierarchy",
			"org:main-org-123:children",
			"org:main-org-123:active_children",
		}

		for _, key := range mainOrgKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		// Mock hierarchy-related cache invalidation for affected orgs
		for _, affectedOrgID := range event.AffectedIDs {
			for _, pattern := range mainOrgKeys {
				affectedKey := pattern
				if affectedOrgID != "main-org-123" {
					affectedKey = "org:" + affectedOrgID + ":" + pattern[len("org:main-org-123:"):]
				}
				mockCache.On("Delete", affectedKey).Return(nil)
			}
		}

		err := invalidationService.InvalidateCache(ctx, event)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("Role Assigned to Group Event", func(t *testing.T) {
		event := invalidationService.CreateInvalidationEvent(
			"role_assigned_to_group",
			"group-123",
			"group",
			"org-456",
			[]string{},
			map[string]interface{}{
				"role_id": "role-789",
			},
		)

		// Mock group role cache invalidation
		groupRoleKeys := []string{
			"group:group-123:roles",
			"group:group-123:active_roles",
			"group:group-123:role_details",
			"group:group-123:role_inheritance",
		}

		for _, key := range groupRoleKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		// Mock user effective roles cache invalidation
		userRolePattern := "org:org-456:user:*:effective_roles*"
		userRoleKeys := []string{
			"org:org-456:user:user1:effective_roles_v2",
			"org:org-456:user:user2:effective_roles",
		}
		mockCache.On("Keys", userRolePattern).Return(userRoleKeys, nil)

		for _, key := range userRoleKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		err := invalidationService.InvalidateCache(ctx, event)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("User Group Membership Changed Event", func(t *testing.T) {
		event := invalidationService.CreateInvalidationEvent(
			"user_group_membership_changed",
			"group-123",
			"group",
			"org-456",
			[]string{},
			map[string]interface{}{
				"user_id": "user-789",
			},
		)

		// Mock user effective roles cache invalidation
		userEffectiveRolesKeys := []string{
			"org:org-456:user:user-789:effective_roles_v2",
			"org:org-456:user:user-789:group_memberships",
			"org:org-456:user:user-789:effective_roles",
		}

		for _, key := range userEffectiveRolesKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		// Mock group cache invalidation
		groupKeys := []string{
			"group:group-123:hierarchy",
			"group:group-123:ancestors",
			"group:group-123:descendants",
			"group:group-123:children",
			"group:group-123:active_children",
			"group:group-123:roles",
			"group:group-123:active_roles",
			"group:group-123:role_details",
			"group:group-123:members",
			"group:group-123:active_members",
			"group:group-123:member_details",
			"group:group-123:role_inheritance",
		}

		for _, key := range groupKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		err := invalidationService.InvalidateCache(ctx, event)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("Role Updated Event", func(t *testing.T) {
		event := invalidationService.CreateInvalidationEvent(
			"role_updated",
			"role-123",
			"role",
			"org-456",
			[]string{},
			map[string]interface{}{},
		)

		// Mock organization role-related cache invalidation
		orgRoleKeys := []string{
			"org:org-456:role_inheritance",
		}

		for _, key := range orgRoleKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		// Mock user effective roles invalidation
		userRolePattern := "org:org-456:user:*:effective_roles*"
		userRoleKeys := []string{
			"org:org-456:user:user1:effective_roles_v2",
			"org:org-456:user:user2:effective_roles",
		}
		mockCache.On("Keys", userRolePattern).Return(userRoleKeys, nil)

		for _, key := range userRoleKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		// Mock group roles invalidation
		groupRolePattern := "org:org-456:group:*:*roles"
		groupRoleKeys := []string{
			"org:org-456:group:group1:roles",
			"org:org-456:group:group2:active_roles",
		}
		mockCache.On("Keys", groupRolePattern).Return(groupRoleKeys, nil)

		for _, key := range groupRoleKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		// Mock role-specific cache keys
		rolePattern := "*:role:role-123:*"
		roleKeys := []string{
			"group:group1:role:role-123:details",
			"org:org-456:role:role-123:assignments",
		}
		mockCache.On("Keys", rolePattern).Return(roleKeys, nil)

		for _, key := range roleKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		err := invalidationService.InvalidateCache(ctx, event)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})
}

func TestCacheInvalidationService_BatchInvalidation(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	invalidationService := NewCacheInvalidationService(mockCache, logger)
	ctx := context.Background()

	t.Run("Batch Invalidation Success", func(t *testing.T) {
		events := []InvalidationEvent{
			invalidationService.CreateInvalidationEvent(
				"organization_updated",
				"org-1",
				"organization",
				"",
				[]string{},
				map[string]interface{}{},
			),
			invalidationService.CreateInvalidationEvent(
				"group_updated",
				"group-1",
				"group",
				"org-1",
				[]string{},
				map[string]interface{}{},
			),
		}

		// Mock cache invalidation for first event (organization_updated)
		org1Keys := []string{
			"org:org-1:hierarchy",
			"org:org-1:parent_hierarchy",
			"org:org-1:children",
			"org:org-1:active_children",
			"org:org-1:groups",
			"org:org-1:active_groups",
			"org:org-1:group_hierarchy",
			"org:org-1:stats",
		}

		for _, key := range org1Keys {
			mockCache.On("Delete", key).Return(nil)
		}

		userPattern1 := "org:org-1:user:*"
		mockCache.On("Keys", userPattern1).Return([]string{}, nil)

		// Mock cache invalidation for second event (group_updated)
		group1Keys := []string{
			"group:group-1:hierarchy",
			"group:group-1:ancestors",
			"group:group-1:descendants",
			"group:group-1:children",
			"group:group-1:active_children",
			"group:group-1:roles",
			"group:group-1:active_roles",
			"group:group-1:role_details",
			"group:group-1:members",
			"group:group-1:active_members",
			"group:group-1:member_details",
			"group:group-1:role_inheritance",
		}

		for _, key := range group1Keys {
			mockCache.On("Delete", key).Return(nil)
		}

		// Organization cache invalidation for group update
		for _, key := range org1Keys {
			mockCache.On("Delete", key).Return(nil)
		}

		mockCache.On("Keys", userPattern1).Return([]string{}, nil)

		err := invalidationService.BatchInvalidateCache(ctx, events)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("Batch Invalidation Partial Failure", func(t *testing.T) {
		events := []InvalidationEvent{
			invalidationService.CreateInvalidationEvent(
				"invalid_event_type",
				"resource-1",
				"unknown",
				"",
				[]string{},
				map[string]interface{}{},
			),
			invalidationService.CreateInvalidationEvent(
				"organization_updated",
				"org-2",
				"organization",
				"",
				[]string{},
				map[string]interface{}{},
			),
		}

		// Mock cache invalidation for second event only (first will fail)
		org2Keys := []string{
			"org:org-2:hierarchy",
			"org:org-2:parent_hierarchy",
			"org:org-2:children",
			"org:org-2:active_children",
			"org:org-2:groups",
			"org:org-2:active_groups",
			"org:org-2:group_hierarchy",
			"org:org-2:stats",
		}

		for _, key := range org2Keys {
			mockCache.On("Delete", key).Return(nil)
		}

		userPattern2 := "org:org-2:user:*"
		mockCache.On("Keys", userPattern2).Return([]string{}, nil)

		err := invalidationService.BatchInvalidateCache(ctx, events)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch invalidation completed with 1 errors")

		mockCache.AssertExpectations(t)
	})
}

func TestCacheInvalidationService_EventValidation(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	invalidationService := NewCacheInvalidationService(mockCache, logger)

	t.Run("Valid Event", func(t *testing.T) {
		event := invalidationService.CreateInvalidationEvent(
			"organization_updated",
			"org-123",
			"organization",
			"",
			[]string{},
			map[string]interface{}{},
		)

		err := invalidationService.ValidateInvalidationEvent(event)
		assert.NoError(t, err)
	})

	t.Run("Missing Event Type", func(t *testing.T) {
		event := InvalidationEvent{
			Type:         "",
			ResourceID:   "org-123",
			ResourceType: "organization",
		}

		err := invalidationService.ValidateInvalidationEvent(event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event type is required")
	})

	t.Run("Missing Resource ID", func(t *testing.T) {
		event := InvalidationEvent{
			Type:         "organization_updated",
			ResourceID:   "",
			ResourceType: "organization",
		}

		err := invalidationService.ValidateInvalidationEvent(event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource ID is required")
	})

	t.Run("Missing Resource Type", func(t *testing.T) {
		event := InvalidationEvent{
			Type:         "organization_updated",
			ResourceID:   "org-123",
			ResourceType: "",
		}

		err := invalidationService.ValidateInvalidationEvent(event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource type is required")
	})

	t.Run("Unsupported Event Type", func(t *testing.T) {
		event := InvalidationEvent{
			Type:         "unsupported_event",
			ResourceID:   "org-123",
			ResourceType: "organization",
		}

		err := invalidationService.ValidateInvalidationEvent(event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported event type")
	})
}

func TestCacheInvalidationService_Statistics(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	invalidationService := NewCacheInvalidationService(mockCache, logger)
	ctx := context.Background()

	t.Run("Get Invalidation Statistics", func(t *testing.T) {
		// Mock cache key counts
		orgKeys := []string{"org:org1:hierarchy", "org:org2:stats"}
		groupKeys := []string{"group:group1:roles", "group:group2:members"}
		userKeys := []string{"org:org1:user:user1:groups", "org:org2:user:user2:effective_roles"}
		roleKeys := []string{"org:org1:role:role1:details", "group:group1:role:role2:assignments"}
		effectiveRolesKeys := []string{"org:org1:user:user1:effective_roles_v2"}

		mockCache.On("Keys", "org:*").Return(orgKeys, nil)
		mockCache.On("Keys", "group:*").Return(groupKeys, nil)
		mockCache.On("Keys", "*:user:*").Return(userKeys, nil)
		mockCache.On("Keys", "*:role*").Return(roleKeys, nil)
		mockCache.On("Keys", "*:effective_roles*").Return(effectiveRolesKeys, nil)

		stats, err := invalidationService.GetInvalidationStats(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Verify statistics
		assert.Equal(t, len(orgKeys), stats["organization_cache_keys"])
		assert.Equal(t, len(groupKeys), stats["group_cache_keys"])
		assert.Equal(t, len(userKeys), stats["user_cache_keys"])
		assert.Equal(t, len(roleKeys), stats["role_cache_keys"])
		assert.Equal(t, len(effectiveRolesKeys), stats["effective_roles_cache_keys"])

		// Verify strategy counts
		strategies := invalidationService.GetInvalidationStrategies()
		assert.Equal(t, len(strategies), stats["available_strategies"])

		userAffectingCount := 0
		for _, strategy := range strategies {
			if strategy.AffectsUsers {
				userAffectingCount++
			}
		}
		assert.Equal(t, userAffectingCount, stats["user_affecting_strategies"])

		mockCache.AssertExpectations(t)
	})
}

func TestCacheInvalidationService_StrategyInformation(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	invalidationService := NewCacheInvalidationService(mockCache, logger)

	t.Run("Get All Invalidation Strategies", func(t *testing.T) {
		strategies := invalidationService.GetInvalidationStrategies()

		// Verify we have all expected strategies
		expectedStrategies := []string{
			"organization_created",
			"organization_updated",
			"organization_deleted",
			"organization_hierarchy_changed",
			"group_created",
			"group_updated",
			"group_deleted",
			"group_hierarchy_changed",
			"user_group_membership_changed",
			"role_assigned_to_group",
			"role_removed_from_group",
			"role_updated",
		}

		assert.Equal(t, len(expectedStrategies), len(strategies))

		strategyTypes := make(map[string]bool)
		for _, strategy := range strategies {
			strategyTypes[strategy.EventType] = true
			assert.NotEmpty(t, strategy.Description)
			assert.NotNil(t, strategy.Handler)
		}

		for _, expectedType := range expectedStrategies {
			assert.True(t, strategyTypes[expectedType], "Missing strategy: %s", expectedType)
		}
	})

	t.Run("Verify User-Affecting Strategies", func(t *testing.T) {
		strategies := invalidationService.GetInvalidationStrategies()

		userAffectingStrategies := []string{
			"organization_deleted",
			"group_deleted",
			"group_hierarchy_changed",
			"user_group_membership_changed",
			"role_assigned_to_group",
			"role_removed_from_group",
			"role_updated",
		}

		for _, strategy := range strategies {
			if contains(userAffectingStrategies, strategy.EventType) {
				assert.True(t, strategy.AffectsUsers,
					"Strategy %s should affect users", strategy.EventType)
			} else {
				assert.False(t, strategy.AffectsUsers,
					"Strategy %s should not affect users", strategy.EventType)
			}
		}
	})
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
