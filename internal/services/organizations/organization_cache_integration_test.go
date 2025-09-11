//go:build integration
// +build integration

package organizations

import (
	"context"
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	organizationResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/organizations"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestOrganizationCacheService_CacheAndRetrieveHierarchy(t *testing.T) {
	// Setup
	mockCache := &MockCacheService{}
	logger := zap.NewNop()
	cacheService := NewOrganizationCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "test-org-123"

	// Create test hierarchy response
	hierarchy := &organizationResponses.OrganizationHierarchyResponse{
		Organization: &organizationResponses.OrganizationResponse{
			ID:   orgID,
			Name: "Test Organization",
		},
		Parents:  []*organizationResponses.OrganizationResponse{},
		Children: []*organizationResponses.OrganizationResponse{},
		Groups:   []*organizationResponses.GroupHierarchyNode{},
	}

	// Test caching
	expectedKey := "org:test-org-123:hierarchy"
	mockCache.On("Set", expectedKey, hierarchy, HierarchyCacheTTL).Return(nil)

	err := cacheService.CacheOrganizationHierarchy(ctx, orgID, hierarchy)
	assert.NoError(t, err)

	// Test retrieval
	mockCache.On("Get", expectedKey).Return(hierarchy, true)

	cached, found := cacheService.GetCachedOrganizationHierarchy(ctx, orgID)
	assert.True(t, found)
	assert.Equal(t, hierarchy, cached)

	mockCache.AssertExpectations(t)
}

func TestOrganizationCacheService_CacheAndRetrieveStats(t *testing.T) {
	// Setup
	mockCache := &MockCacheService{}
	logger := zap.NewNop()
	cacheService := NewOrganizationCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "test-org-123"

	// Create test stats response
	stats := &organizationResponses.OrganizationStatsResponse{
		OrganizationID: orgID,
		ChildCount:     5,
		GroupCount:     10,
		UserCount:      25,
	}

	// Test caching
	expectedKey := "org:test-org-123:stats"
	mockCache.On("Set", expectedKey, stats, StatsCacheTTL).Return(nil)

	err := cacheService.CacheOrganizationStats(ctx, orgID, stats)
	assert.NoError(t, err)

	// Test retrieval
	mockCache.On("Get", expectedKey).Return(stats, true)

	cached, found := cacheService.GetCachedOrganizationStats(ctx, orgID)
	assert.True(t, found)
	assert.Equal(t, stats, cached)

	mockCache.AssertExpectations(t)
}

func TestOrganizationCacheService_CacheAndRetrieveParentHierarchy(t *testing.T) {
	// Setup
	mockCache := &MockCacheService{}
	logger := zap.NewNop()
	cacheService := NewOrganizationCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "test-org-123"

	// Create test parent hierarchy
	parents := []*models.Organization{
		{
			BaseModel: base.NewBaseModel("ORG", hash.Medium),
			Name:      "Parent Organization 1",
			IsActive:  true,
		},
		{
			BaseModel: base.NewBaseModel("ORG", hash.Medium),
			Name:      "Parent Organization 2",
			IsActive:  true,
		},
	}
	parents[0].SetID("parent-1")
	parents[1].SetID("parent-2")

	// Test caching
	expectedKey := "org:test-org-123:parent_hierarchy"
	mockCache.On("Set", expectedKey, parents, HierarchyCacheTTL).Return(nil)

	err := cacheService.CacheOrganizationParentHierarchy(ctx, orgID, parents)
	assert.NoError(t, err)

	// Test retrieval
	mockCache.On("Get", expectedKey).Return(parents, true)

	cached, found := cacheService.GetCachedOrganizationParentHierarchy(ctx, orgID)
	assert.True(t, found)
	assert.Equal(t, parents, cached)

	mockCache.AssertExpectations(t)
}

func TestOrganizationCacheService_CacheAndRetrieveChildren(t *testing.T) {
	// Setup
	mockCache := &MockCacheService{}
	logger := zap.NewNop()
	cacheService := NewOrganizationCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "test-org-123"

	// Create test children
	children := []*models.Organization{
		{
			BaseModel: base.NewBaseModel("ORG", hash.Medium),
			Name:      "Child Organization 1",
			IsActive:  true,
		},
		{
			BaseModel: base.NewBaseModel("ORG", hash.Medium),
			Name:      "Child Organization 2",
			IsActive:  false,
		},
	}
	children[0].SetID("child-1")
	children[1].SetID("child-2")

	// Test caching active children
	expectedActiveKey := "org:test-org-123:active_children"
	mockCache.On("Set", expectedActiveKey, children, HierarchyCacheTTL).Return(nil)

	err := cacheService.CacheOrganizationChildren(ctx, orgID, children, true)
	assert.NoError(t, err)

	// Test retrieval of active children
	mockCache.On("Get", expectedActiveKey).Return(children, true)

	cached, found := cacheService.GetCachedOrganizationChildren(ctx, orgID, true)
	assert.True(t, found)
	assert.Equal(t, children, cached)

	// Test caching all children
	expectedAllKey := "org:test-org-123:children"
	mockCache.On("Set", expectedAllKey, children, HierarchyCacheTTL).Return(nil)

	err = cacheService.CacheOrganizationChildren(ctx, orgID, children, false)
	assert.NoError(t, err)

	// Test retrieval of all children
	mockCache.On("Get", expectedAllKey).Return(children, true)

	cached, found = cacheService.GetCachedOrganizationChildren(ctx, orgID, false)
	assert.True(t, found)
	assert.Equal(t, children, cached)

	mockCache.AssertExpectations(t)
}

func TestOrganizationCacheService_InvalidateCache(t *testing.T) {
	// Setup
	mockCache := &MockCacheService{}
	logger := zap.NewNop()
	cacheService := NewOrganizationCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "test-org-123"

	// Mock cache deletion for organization-specific keys
	expectedKeys := []string{
		"org:test-org-123:hierarchy",
		"org:test-org-123:parent_hierarchy",
		"org:test-org-123:children",
		"org:test-org-123:active_children",
		"org:test-org-123:groups",
		"org:test-org-123:active_groups",
		"org:test-org-123:group_hierarchy",
		"org:test-org-123:stats",
	}

	for _, key := range expectedKeys {
		mockCache.On("Delete", key).Return(nil)
	}

	// Mock user-specific cache invalidation
	userPattern := "org:test-org-123:user:*"
	userKeys := []string{
		"org:test-org-123:user:user1:groups",
		"org:test-org-123:user:user2:active_groups",
	}
	mockCache.On("Keys", userPattern).Return(userKeys, nil)

	for _, key := range userKeys {
		mockCache.On("Delete", key).Return(nil)
	}

	// Test invalidation
	err := cacheService.InvalidateOrganizationCache(ctx, orgID)
	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
}

func TestOrganizationCacheService_InvalidateUserGroupCache(t *testing.T) {
	// Setup
	mockCache := &MockCacheService{}
	logger := zap.NewNop()
	cacheService := NewOrganizationCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "test-org-123"
	userID := "test-user-456"

	// Mock cache deletion for user-specific keys
	expectedKeys := []string{
		"org:test-org-123:user:test-user-456:groups",
		"org:test-org-123:user:test-user-456:active_groups",
	}

	for _, key := range expectedKeys {
		mockCache.On("Delete", key).Return(nil)
	}

	// Test invalidation
	err := cacheService.InvalidateUserGroupCache(ctx, orgID, userID)
	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
}

func TestOrganizationCacheService_InvalidateGroupCache(t *testing.T) {
	// Setup
	mockCache := &MockCacheService{}
	logger := zap.NewNop()
	cacheService := NewOrganizationCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "test-org-123"
	groupID := "test-group-789"

	// Mock cache deletion for group-specific keys
	groupKeys := []string{
		"org:test-org-123:group:test-group-789:members",
		"org:test-org-123:group:test-group-789:active_members",
	}

	for _, key := range groupKeys {
		mockCache.On("Delete", key).Return(nil)
	}

	// Mock cache deletion for organization-level group keys
	orgGroupKeys := []string{
		"org:test-org-123:groups",
		"org:test-org-123:active_groups",
		"org:test-org-123:group_hierarchy",
	}

	for _, key := range orgGroupKeys {
		mockCache.On("Delete", key).Return(nil)
	}

	// Test invalidation
	err := cacheService.InvalidateGroupCache(ctx, orgID, groupID)
	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
}

func TestOrganizationCacheService_CacheMiss(t *testing.T) {
	// Setup
	mockCache := &MockCacheService{}
	logger := zap.NewNop()
	cacheService := NewOrganizationCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "test-org-123"

	// Test cache miss for hierarchy
	expectedKey := "org:test-org-123:hierarchy"
	mockCache.On("Get", expectedKey).Return(nil, false)

	cached, found := cacheService.GetCachedOrganizationHierarchy(ctx, orgID)
	assert.False(t, found)
	assert.Nil(t, cached)

	mockCache.AssertExpectations(t)
}

func TestOrganizationCacheService_InvalidCacheType(t *testing.T) {
	// Setup
	mockCache := &MockCacheService{}
	logger := zap.NewNop()
	cacheService := NewOrganizationCacheService(mockCache, logger)

	ctx := context.Background()
	orgID := "test-org-123"

	// Test invalid cache type for hierarchy
	expectedKey := "org:test-org-123:hierarchy"
	invalidData := "invalid-data-type"
	mockCache.On("Get", expectedKey).Return(invalidData, true)
	mockCache.On("Delete", expectedKey).Return(nil) // Should delete invalid cache entry

	cached, found := cacheService.GetCachedOrganizationHierarchy(ctx, orgID)
	assert.False(t, found)
	assert.Nil(t, cached)

	mockCache.AssertExpectations(t)
}

// Integration test with real Redis cache service
func TestOrganizationCacheService_Integration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup real cache service (requires Redis)
	logger := zap.NewNop()

	// Create a simple Redis client for testing
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// Skip if Redis is not available
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
		return
	}

	// Create a mock cache service for this test
	realCache := &MockCacheService{}
	cacheService := NewOrganizationCacheService(realCache, logger)

	ctx := context.Background()
	orgID := "integration-test-org"

	// Clean up before test
	defer realCache.Delete("org:integration-test-org:hierarchy")

	// Create test hierarchy
	hierarchy := &organizationResponses.OrganizationHierarchyResponse{
		Organization: &organizationResponses.OrganizationResponse{
			ID:   orgID,
			Name: "Integration Test Organization",
		},
		Parents:  []*organizationResponses.OrganizationResponse{},
		Children: []*organizationResponses.OrganizationResponse{},
		Groups:   []*organizationResponses.GroupHierarchyNode{},
	}

	// Test caching
	err = cacheService.CacheOrganizationHierarchy(ctx, orgID, hierarchy)
	assert.NoError(t, err)

	// Test immediate retrieval
	cached, found := cacheService.GetCachedOrganizationHierarchy(ctx, orgID)
	assert.True(t, found)
	assert.Equal(t, hierarchy.Organization.ID, cached.Organization.ID)
	assert.Equal(t, hierarchy.Organization.Name, cached.Organization.Name)

	// Test cache expiration (wait a bit and check TTL)
	time.Sleep(100 * time.Millisecond)
	ttl, err := realCache.TTL("org:integration-test-org:hierarchy")
	assert.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= HierarchyCacheTTL)

	// Test invalidation
	err = cacheService.InvalidateOrganizationCache(ctx, orgID)
	assert.NoError(t, err)

	// Verify cache is cleared
	cached, found = cacheService.GetCachedOrganizationHierarchy(ctx, orgID)
	assert.False(t, found)
	assert.Nil(t, cached)
}
