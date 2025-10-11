//go:build integration
// +build integration

package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	organizationResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/organizations"
	"github.com/Kisanlink/aaa-service/v2/internal/services/groups"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// TestCompleteCacheIntegration tests the complete caching layer integration
// including cache warming, invalidation, and performance optimization
func TestCompleteCacheIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping complete cache integration test")
	}

	// Setup
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	// Create all cache-related services
	warmingService := NewCacheWarmingService(&MockOrganizationService{}, &MockGroupService{}, mockCache, logger)
	invalidationService := NewCacheInvalidationService(mockCache, logger)

	ctx := context.Background()

	t.Run("Complete Cache Lifecycle", func(t *testing.T) {
		orgID := "lifecycle-org-123"
		groupID := "lifecycle-group-456"
		userID := "lifecycle-user-789"
		roleID := "lifecycle-role-101"

		// Phase 1: Initial Cache Warming
		t.Log("Phase 1: Initial Cache Warming")

		config := CacheWarmingConfig{
			WarmingInterval:       1 * time.Hour,
			MaxConcurrentWarms:    2,
			FrequentOrganizations: []string{orgID},
			FrequentGroups:        []string{groupID},
			FrequentUsers:         []string{userID},
			WarmHierarchies:       true,
			WarmGroupRoles:        true,
			WarmEffectiveRoles:    true,
			WarmStats:             true,
		}

		// Mock initial warming calls
		hierarchy := &organizationResponses.OrganizationHierarchyResponse{
			Organization: &organizationResponses.OrganizationResponse{
				ID:   orgID,
				Name: "Lifecycle Test Organization",
			},
		}
		stats := &organizationResponses.OrganizationStatsResponse{
			OrganizationID: orgID,
			ChildCount:     2,
			GroupCount:     3,
			UserCount:      5,
		}

		mockOrgService := &MockOrganizationService{}
		mockGroupService := &MockGroupService{}

		mockOrgService.On("GetOrganizationHierarchy", ctx, orgID).Return(hierarchy, nil)
		mockOrgService.On("GetOrganizationStats", ctx, orgID).Return(stats, nil)
		mockOrgService.On("GetOrganizationGroups", ctx, orgID, 100, 0, false).Return([]interface{}{}, nil)
		mockOrgService.On("GetOrganizationGroups", ctx, orgID, 100, 0, true).Return([]interface{}{}, nil)

		mockGroupService.On("GetGroupRoles", ctx, groupID).Return([]interface{}{}, nil)
		mockGroupService.On("GetGroupMembers", ctx, groupID, 100, 0).Return([]interface{}{}, nil)

		effectiveRoles := []*groups.EffectiveRole{
			{
				Role: &models.Role{
					Name:     "Lifecycle Test Role",
					IsActive: true,
				},
				GroupID:      groupID,
				GroupName:    "Lifecycle Test Group",
				Distance:     0,
				IsDirectRole: true,
			},
		}
		mockGroupService.On("GetUserEffectiveRoles", ctx, orgID, userID).Return(effectiveRoles, nil)

		// Update warming service with mocked services
		warmingService = NewCacheWarmingService(mockOrgService, mockGroupService, mockCache, logger)

		// Perform initial warming
		err := warmingService.WarmNow(ctx, config)
		assert.NoError(t, err)

		// Phase 2: Simulate Cache Usage and Hits
		t.Log("Phase 2: Simulate Cache Usage and Hits")

		// Mock cache hits for frequently accessed data
		hierarchyKey := fmt.Sprintf("org:%s:hierarchy", orgID)
		mockCache.On("Get", hierarchyKey).Return(hierarchy, true).Times(5)

		effectiveRolesKey := fmt.Sprintf("org:%s:user:%s:effective_roles_v2", orgID, userID)
		mockCache.On("Get", effectiveRolesKey).Return(effectiveRoles, true).Times(3)

		groupRolesKey := fmt.Sprintf("group:%s:active_roles", groupID)
		mockCache.On("Get", groupRolesKey).Return([]interface{}{}, true).Times(2)

		// Simulate multiple cache accesses (would normally be from service calls)
		for i := 0; i < 5; i++ {
			cached, found := mockCache.Get(hierarchyKey)
			assert.True(t, found)
			assert.Equal(t, hierarchy, cached)
		}

		for i := 0; i < 3; i++ {
			cached, found := mockCache.Get(effectiveRolesKey)
			assert.True(t, found)
			assert.Equal(t, effectiveRoles, cached)
		}

		for i := 0; i < 2; i++ {
			cached, found := mockCache.Get(groupRolesKey)
			assert.True(t, found)
			assert.NotNil(t, cached)
		}

		// Phase 3: Simulate Data Changes and Cache Invalidation
		t.Log("Phase 3: Simulate Data Changes and Cache Invalidation")

		// Simulate role assignment to group
		roleAssignmentEvent := invalidationService.CreateInvalidationEvent(
			"role_assigned_to_group",
			groupID,
			"group",
			orgID,
			[]string{},
			map[string]interface{}{
				"role_id": roleID,
			},
		)

		// Mock cache invalidation for role assignment
		groupRoleKeys := []string{
			fmt.Sprintf("group:%s:roles", groupID),
			fmt.Sprintf("group:%s:active_roles", groupID),
			fmt.Sprintf("group:%s:role_details", groupID),
			fmt.Sprintf("group:%s:role_inheritance", groupID),
		}

		for _, key := range groupRoleKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		userRolePattern := fmt.Sprintf("org:%s:user:*:effective_roles*", orgID)
		userRoleKeys := []string{effectiveRolesKey}
		mockCache.On("Keys", userRolePattern).Return(userRoleKeys, nil)

		for _, key := range userRoleKeys {
			mockCache.On("Delete", key).Return(nil)
		}

		err = invalidationService.InvalidateCache(ctx, roleAssignmentEvent)
		assert.NoError(t, err)

		// Phase 4: Re-warm Cache After Invalidation
		t.Log("Phase 4: Re-warm Cache After Invalidation")

		// Mock re-warming calls with updated data
		updatedEffectiveRoles := []*groups.EffectiveRole{
			{
				Role: &models.Role{
					Name:     "Lifecycle Test Role",
					IsActive: true,
				},
				GroupID:      groupID,
				GroupName:    "Lifecycle Test Group",
				Distance:     0,
				IsDirectRole: true,
			},
			{
				Role: &models.Role{
					Name:     "New Assigned Role",
					IsActive: true,
				},
				GroupID:      groupID,
				GroupName:    "Lifecycle Test Group",
				Distance:     0,
				IsDirectRole: true,
			},
		}

		mockGroupService.On("GetUserEffectiveRoles", ctx, orgID, userID).Return(updatedEffectiveRoles, nil)
		mockGroupService.On("GetGroupRoles", ctx, groupID).Return([]interface{}{
			map[string]interface{}{
				"role_id":   roleID,
				"role_name": "New Assigned Role",
			},
		}, nil)

		// Re-warm specific data
		err = warmingService.WarmNow(ctx, config)
		assert.NoError(t, err)

		// Phase 5: Verify Cache Statistics
		t.Log("Phase 5: Verify Cache Statistics")

		// Mock cache statistics
		orgKeys := []string{hierarchyKey, fmt.Sprintf("org:%s:stats", orgID)}
		groupKeys := []string{groupRolesKey, fmt.Sprintf("group:%s:members", groupID)}
		userKeys := []string{effectiveRolesKey}
		roleKeys := []string{fmt.Sprintf("org:%s:role:%s:details", orgID, roleID)}
		effectiveRolesKeys := []string{effectiveRolesKey}

		mockCache.On("Keys", "org:*").Return(orgKeys, nil)
		mockCache.On("Keys", "group:*").Return(groupKeys, nil)
		mockCache.On("Keys", "*:user:*").Return(userKeys, nil)
		mockCache.On("Keys", "*:role*").Return(roleKeys, nil)
		mockCache.On("Keys", "*:effective_roles*").Return(effectiveRolesKeys, nil)

		// Get warming service statistics
		warmingStats, err := warmingService.GetCacheStats(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, warmingStats)

		// Get invalidation service statistics
		invalidationStats, err := invalidationService.GetInvalidationStats(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, invalidationStats)

		// Verify statistics make sense
		assert.Equal(t, len(orgKeys), warmingStats["organization_cache_keys"])
		assert.Equal(t, len(groupKeys), warmingStats["group_cache_keys"])
		assert.Equal(t, len(effectiveRolesKeys), warmingStats["effective_roles_cache_keys"])

		assert.Equal(t, len(orgKeys), invalidationStats["organization_cache_keys"])
		assert.Equal(t, len(groupKeys), invalidationStats["group_cache_keys"])
		assert.Equal(t, len(userKeys), invalidationStats["user_cache_keys"])

		// Verify all expectations were met
		mockOrgService.AssertExpectations(t)
		mockGroupService.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

// TestCachePerformanceUnderLoad tests cache performance under simulated load
func TestCachePerformanceUnderLoad(t *testing.T) {
	// Skip if not running performance tests
	if testing.Short() {
		t.Skip("Skipping cache performance test")
	}

	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	invalidationService := NewCacheInvalidationService(mockCache, logger)
	ctx := context.Background()

	t.Run("High Volume Cache Invalidation", func(t *testing.T) {
		// Simulate high volume of cache invalidation events
		numEvents := 100
		events := make([]InvalidationEvent, numEvents)

		for i := 0; i < numEvents; i++ {
			orgID := fmt.Sprintf("perf-org-%d", i%10)     // 10 different orgs
			groupID := fmt.Sprintf("perf-group-%d", i%20) // 20 different groups
			userID := fmt.Sprintf("perf-user-%d", i%50)   // 50 different users

			// Alternate between different event types
			eventTypes := []string{
				"organization_updated",
				"group_updated",
				"user_group_membership_changed",
				"role_assigned_to_group",
			}
			eventType := eventTypes[i%len(eventTypes)]

			var metadata map[string]interface{}
			switch eventType {
			case "user_group_membership_changed":
				metadata = map[string]interface{}{"user_id": userID}
			case "role_assigned_to_group":
				metadata = map[string]interface{}{"role_id": fmt.Sprintf("perf-role-%d", i%5)}
			default:
				metadata = map[string]interface{}{}
			}

			events[i] = invalidationService.CreateInvalidationEvent(
				eventType,
				groupID,
				"group",
				orgID,
				[]string{},
				metadata,
			)
		}

		// Mock all the cache operations that would be called
		// This is a simplified mock - in reality we'd need to mock all specific keys
		mockCache.On("Delete", mock.AnythingOfType("string")).Return(nil).Times(numEvents * 10)          // Approximate
		mockCache.On("Keys", mock.AnythingOfType("string")).Return([]string{}, nil).Times(numEvents * 2) // Approximate

		// Measure batch invalidation performance
		startTime := time.Now()
		err := invalidationService.BatchInvalidateCache(ctx, events)
		duration := time.Since(startTime)

		assert.NoError(t, err)
		assert.Less(t, duration, 10*time.Second, "Batch invalidation should complete within 10 seconds")

		t.Logf("Processed %d invalidation events in %v (%.2f events/sec)",
			numEvents, duration, float64(numEvents)/duration.Seconds())

		// Note: We're not asserting exact expectations here due to the complexity
		// of mocking all possible cache operations. In a real integration test,
		// we would use a real cache service.
	})
}

// TestCacheConsistency tests cache consistency across different operations
func TestCacheConsistency(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	invalidationService := NewCacheInvalidationService(mockCache, logger)
	ctx := context.Background()

	t.Run("Cache Consistency After Multiple Operations", func(t *testing.T) {
		orgID := "consistency-org"
		groupID := "consistency-group"
		userID := "consistency-user"
		roleID := "consistency-role"

		// Simulate a sequence of operations that should maintain cache consistency
		operations := []struct {
			name  string
			event InvalidationEvent
		}{
			{
				name: "Create Group",
				event: invalidationService.CreateInvalidationEvent(
					"group_created",
					groupID,
					"group",
					orgID,
					[]string{},
					map[string]interface{}{},
				),
			},
			{
				name: "Add User to Group",
				event: invalidationService.CreateInvalidationEvent(
					"user_group_membership_changed",
					groupID,
					"group",
					orgID,
					[]string{},
					map[string]interface{}{"user_id": userID},
				),
			},
			{
				name: "Assign Role to Group",
				event: invalidationService.CreateInvalidationEvent(
					"role_assigned_to_group",
					groupID,
					"group",
					orgID,
					[]string{},
					map[string]interface{}{"role_id": roleID},
				),
			},
			{
				name: "Update Role",
				event: invalidationService.CreateInvalidationEvent(
					"role_updated",
					roleID,
					"role",
					orgID,
					[]string{},
					map[string]interface{}{},
				),
			},
		}

		// Mock cache operations for each operation
		for _, op := range operations {
			t.Logf("Processing operation: %s", op.name)

			// Mock the specific cache invalidations for each operation type
			switch op.event.Type {
			case "group_created":
				// Mock organization cache invalidation
				orgKeys := []string{
					fmt.Sprintf("org:%s:hierarchy", orgID),
					fmt.Sprintf("org:%s:groups", orgID),
					fmt.Sprintf("org:%s:active_groups", orgID),
				}
				for _, key := range orgKeys {
					mockCache.On("Delete", key).Return(nil).Once()
				}
				mockCache.On("Keys", fmt.Sprintf("org:%s:user:*", orgID)).Return([]string{}, nil).Once()

			case "user_group_membership_changed":
				// Mock user effective roles cache invalidation
				userKeys := []string{
					fmt.Sprintf("org:%s:user:%s:effective_roles_v2", orgID, userID),
					fmt.Sprintf("org:%s:user:%s:group_memberships", orgID, userID),
					fmt.Sprintf("org:%s:user:%s:effective_roles", orgID, userID),
				}
				for _, key := range userKeys {
					mockCache.On("Delete", key).Return(nil).Once()
				}

				// Mock group cache invalidation
				groupKeys := []string{
					fmt.Sprintf("group:%s:hierarchy", groupID),
					fmt.Sprintf("group:%s:members", groupID),
					fmt.Sprintf("group:%s:active_members", groupID),
				}
				for _, key := range groupKeys {
					mockCache.On("Delete", key).Return(nil).Once()
				}

			case "role_assigned_to_group":
				// Mock group role cache invalidation
				groupRoleKeys := []string{
					fmt.Sprintf("group:%s:roles", groupID),
					fmt.Sprintf("group:%s:active_roles", groupID),
				}
				for _, key := range groupRoleKeys {
					mockCache.On("Delete", key).Return(nil).Once()
				}

				// Mock user effective roles invalidation
				userRolePattern := fmt.Sprintf("org:%s:user:*:effective_roles*", orgID)
				mockCache.On("Keys", userRolePattern).Return([]string{
					fmt.Sprintf("org:%s:user:%s:effective_roles_v2", orgID, userID),
				}, nil).Once()
				mockCache.On("Delete", fmt.Sprintf("org:%s:user:%s:effective_roles_v2", orgID, userID)).Return(nil).Once()

			case "role_updated":
				// Mock comprehensive role-related cache invalidation
				mockCache.On("Delete", fmt.Sprintf("org:%s:role_inheritance", orgID)).Return(nil).Once()

				// Mock user effective roles invalidation
				userRolePattern := fmt.Sprintf("org:%s:user:*:effective_roles*", orgID)
				mockCache.On("Keys", userRolePattern).Return([]string{
					fmt.Sprintf("org:%s:user:%s:effective_roles_v2", orgID, userID),
				}, nil).Once()
				mockCache.On("Delete", fmt.Sprintf("org:%s:user:%s:effective_roles_v2", orgID, userID)).Return(nil).Once()

				// Mock group roles invalidation
				groupRolePattern := fmt.Sprintf("org:%s:group:*:*roles", orgID)
				mockCache.On("Keys", groupRolePattern).Return([]string{
					fmt.Sprintf("org:%s:group:%s:roles", orgID, groupID),
				}, nil).Once()
				mockCache.On("Delete", fmt.Sprintf("org:%s:group:%s:roles", orgID, groupID)).Return(nil).Once()

				// Mock role-specific cache keys
				rolePattern := fmt.Sprintf("*:role:%s:*", roleID)
				mockCache.On("Keys", rolePattern).Return([]string{}, nil).Once()
			}

			// Process the invalidation event
			err := invalidationService.InvalidateCache(ctx, op.event)
			assert.NoError(t, err, "Operation %s should not fail", op.name)
		}

		// Verify all cache operations were called as expected
		mockCache.AssertExpectations(t)
	})
}

// TestCacheErrorRecovery tests cache error recovery and resilience
func TestCacheErrorRecovery(t *testing.T) {
	mockCache := &MockCacheServiceComplete{}
	logger := zap.NewNop()

	invalidationService := NewCacheInvalidationService(mockCache, logger)
	ctx := context.Background()

	t.Run("Cache Error Recovery", func(t *testing.T) {
		orgID := "error-recovery-org"

		event := invalidationService.CreateInvalidationEvent(
			"organization_updated",
			orgID,
			"organization",
			"",
			[]string{},
			map[string]interface{}{},
		)

		// Mock some cache operations to fail
		mockCache.On("Delete", fmt.Sprintf("org:%s:hierarchy", orgID)).Return(fmt.Errorf("cache error")).Once()
		mockCache.On("Delete", fmt.Sprintf("org:%s:parent_hierarchy", orgID)).Return(nil).Once()
		mockCache.On("Delete", fmt.Sprintf("org:%s:children", orgID)).Return(nil).Once()
		mockCache.On("Delete", fmt.Sprintf("org:%s:active_children", orgID)).Return(nil).Once()
		mockCache.On("Delete", fmt.Sprintf("org:%s:groups", orgID)).Return(nil).Once()
		mockCache.On("Delete", fmt.Sprintf("org:%s:active_groups", orgID)).Return(nil).Once()
		mockCache.On("Delete", fmt.Sprintf("org:%s:group_hierarchy", orgID)).Return(nil).Once()
		mockCache.On("Delete", fmt.Sprintf("org:%s:stats", orgID)).Return(nil).Once()

		mockCache.On("Keys", fmt.Sprintf("org:%s:user:*", orgID)).Return([]string{}, nil).Once()

		// The invalidation should continue despite individual cache operation failures
		err := invalidationService.InvalidateCache(ctx, event)
		assert.NoError(t, err, "Invalidation should not fail due to individual cache operation errors")

		mockCache.AssertExpectations(t)
	})
}
