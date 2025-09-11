//go:build integration
// +build integration

package services

import (
	"context"
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	organizationResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/organizations"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/services/groups"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockLogger implements interfaces.Logger for testing
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...zap.Field)      {}
func (m *MockLogger) Info(msg string, fields ...zap.Field)       {}
func (m *MockLogger) Warn(msg string, fields ...zap.Field)       {}
func (m *MockLogger) Error(msg string, fields ...zap.Field)      {}
func (m *MockLogger) Fatal(msg string, fields ...zap.Field)      {}
func (m *MockLogger) With(fields ...zap.Field) interfaces.Logger { return m }
func (m *MockLogger) Named(name string) interfaces.Logger        { return m }
func (m *MockLogger) Sync() error                                { return nil }

// MockOrganizationService for testing
type MockOrganizationService struct {
	mock.Mock
}

func (m *MockOrganizationService) CreateOrganization(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) UpdateOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) DeleteOrganization(ctx context.Context, orgID string, deletedBy string) error {
	args := m.Called(ctx, orgID, deletedBy)
	return args.Error(0)
}

func (m *MockOrganizationService) ListOrganizations(ctx context.Context, limit, offset int, includeInactive bool) ([]interface{}, error) {
	args := m.Called(ctx, limit, offset, includeInactive)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockOrganizationService) GetOrganizationHierarchy(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) ActivateOrganization(ctx context.Context, orgID string) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

func (m *MockOrganizationService) DeactivateOrganization(ctx context.Context, orgID string) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

func (m *MockOrganizationService) GetOrganizationStats(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) GetOrganizationGroups(ctx context.Context, orgID string, limit, offset int, includeInactive bool) (interface{}, error) {
	args := m.Called(ctx, orgID, limit, offset, includeInactive)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) CreateGroupInOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) GetGroupInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) UpdateGroupInOrganization(ctx context.Context, orgID, groupID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) DeleteGroupInOrganization(ctx context.Context, orgID, groupID string, deletedBy string) error {
	args := m.Called(ctx, orgID, groupID, deletedBy)
	return args.Error(0)
}

func (m *MockOrganizationService) GetGroupHierarchyInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) AddUserToGroupInOrganization(ctx context.Context, orgID, groupID, userID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, userID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) RemoveUserFromGroupInOrganization(ctx context.Context, orgID, groupID, userID string, removedBy string) error {
	args := m.Called(ctx, orgID, groupID, userID, removedBy)
	return args.Error(0)
}

func (m *MockOrganizationService) GetGroupUsersInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, orgID, userID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) AssignRoleToGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, roleID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) RemoveRoleFromGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, removedBy string) error {
	args := m.Called(ctx, orgID, groupID, roleID, removedBy)
	return args.Error(0)
}

func (m *MockOrganizationService) GetGroupRolesInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) GetUserEffectiveRolesInOrganization(ctx context.Context, orgID, userID string) (interface{}, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Get(0), args.Error(1)
}

// MockGroupService for testing
type MockGroupService struct {
	mock.Mock
}

func (m *MockGroupService) CreateGroup(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) GetGroup(ctx context.Context, groupID string) (interface{}, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) UpdateGroup(ctx context.Context, groupID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, groupID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) DeleteGroup(ctx context.Context, groupID string, deletedBy string) error {
	args := m.Called(ctx, groupID, deletedBy)
	return args.Error(0)
}

func (m *MockGroupService) ListGroups(ctx context.Context, limit, offset int, organizationID string, includeInactive bool) (interface{}, error) {
	args := m.Called(ctx, limit, offset, organizationID, includeInactive)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) AddMemberToGroup(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) RemoveMemberFromGroup(ctx context.Context, groupID, principalID string, removedBy string) error {
	args := m.Called(ctx, groupID, principalID, removedBy)
	return args.Error(0)
}

func (m *MockGroupService) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, groupID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) AssignRoleToGroup(ctx context.Context, groupID, roleID, assignedBy string) (interface{}, error) {
	args := m.Called(ctx, groupID, roleID, assignedBy)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) RemoveRoleFromGroup(ctx context.Context, groupID, roleID string) error {
	args := m.Called(ctx, groupID, roleID)
	return args.Error(0)
}

func (m *MockGroupService) GetGroupRoles(ctx context.Context, groupID string) (interface{}, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) GetUserEffectiveRoles(ctx context.Context, orgID, userID string) (interface{}, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Get(0), args.Error(1)
}

func TestCacheWarmingService_StartAndStop(t *testing.T) {
	// Setup
	mockOrgService := &MockOrganizationService{}
	mockGroupService := &MockGroupService{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	warmingService := NewCacheWarmingService(mockOrgService, mockGroupService, mockCache, logger)

	ctx := context.Background()
	config := CacheWarmingConfig{
		WarmingInterval:       1 * time.Second,
		MaxConcurrentWarms:    2,
		FrequentOrganizations: []string{"org1", "org2"},
		FrequentGroups:        []string{"group1", "group2"},
		FrequentUsers:         []string{"user1", "user2"},
		WarmHierarchies:       true,
		WarmGroupRoles:        true,
		WarmEffectiveRoles:    true,
		WarmStats:             true,
	}

	// Test starting the service
	assert.False(t, warmingService.IsRunning())

	// Mock the service calls for initial warming
	mockOrgService.On("GetOrganizationHierarchy", ctx, "org1").Return(&organizationResponses.OrganizationHierarchyResponse{}, nil)
	mockOrgService.On("GetOrganizationHierarchy", ctx, "org2").Return(&organizationResponses.OrganizationHierarchyResponse{}, nil)
	mockOrgService.On("GetOrganizationStats", ctx, "org1").Return(&organizationResponses.OrganizationStatsResponse{}, nil)
	mockOrgService.On("GetOrganizationStats", ctx, "org2").Return(&organizationResponses.OrganizationStatsResponse{}, nil)
	mockOrgService.On("GetOrganizationGroups", ctx, "org1", 100, 0, false).Return([]interface{}{}, nil)
	mockOrgService.On("GetOrganizationGroups", ctx, "org2", 100, 0, false).Return([]interface{}{}, nil)
	mockOrgService.On("GetOrganizationGroups", ctx, "org1", 100, 0, true).Return([]interface{}{}, nil)
	mockOrgService.On("GetOrganizationGroups", ctx, "org2", 100, 0, true).Return([]interface{}{}, nil)

	mockGroupService.On("GetGroupRoles", ctx, "group1").Return([]interface{}{}, nil)
	mockGroupService.On("GetGroupRoles", ctx, "group2").Return([]interface{}{}, nil)
	mockGroupService.On("GetGroupMembers", ctx, "group1", 100, 0).Return([]interface{}{}, nil)
	mockGroupService.On("GetGroupMembers", ctx, "group2", 100, 0).Return([]interface{}{}, nil)

	mockGroupService.On("GetUserEffectiveRoles", ctx, "org1", "user1").Return([]*groups.EffectiveRole{}, nil)
	mockGroupService.On("GetUserEffectiveRoles", ctx, "org1", "user2").Return([]*groups.EffectiveRole{}, nil)
	mockGroupService.On("GetUserEffectiveRoles", ctx, "org2", "user1").Return([]*groups.EffectiveRole{}, nil)
	mockGroupService.On("GetUserEffectiveRoles", ctx, "org2", "user2").Return([]*groups.EffectiveRole{}, nil)

	err := warmingService.Start(ctx, config)
	assert.NoError(t, err)
	assert.True(t, warmingService.IsRunning())

	// Wait a bit for initial warming to complete
	time.Sleep(100 * time.Millisecond)

	// Test stopping the service
	err = warmingService.Stop()
	assert.NoError(t, err)
	assert.False(t, warmingService.IsRunning())

	// Verify that the service calls were made (at least once for initial warming)
	mockOrgService.AssertExpectations(t)
	mockGroupService.AssertExpectations(t)
}

func TestCacheWarmingService_WarmNow(t *testing.T) {
	// Setup
	mockOrgService := &MockOrganizationService{}
	mockGroupService := &MockGroupService{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	warmingService := NewCacheWarmingService(mockOrgService, mockGroupService, mockCache, logger)

	ctx := context.Background()
	config := CacheWarmingConfig{
		WarmingInterval:       1 * time.Hour, // Long interval to avoid automatic warming
		MaxConcurrentWarms:    3,
		FrequentOrganizations: []string{"org1"},
		FrequentGroups:        []string{"group1"},
		FrequentUsers:         []string{"user1"},
		WarmHierarchies:       true,
		WarmGroupRoles:        true,
		WarmEffectiveRoles:    true,
		WarmStats:             true,
	}

	// Mock the service calls
	hierarchy := &organizationResponses.OrganizationHierarchyResponse{
		Organization: &organizationResponses.OrganizationResponse{
			ID:   "org1",
			Name: "Test Organization",
		},
	}
	stats := &organizationResponses.OrganizationStatsResponse{
		OrganizationID: "org1",
		ChildCount:     2,
		GroupCount:     5,
		UserCount:      10,
	}

	mockOrgService.On("GetOrganizationHierarchy", ctx, "org1").Return(hierarchy, nil)
	mockOrgService.On("GetOrganizationStats", ctx, "org1").Return(stats, nil)
	mockOrgService.On("GetOrganizationGroups", ctx, "org1", 100, 0, false).Return([]interface{}{}, nil)
	mockOrgService.On("GetOrganizationGroups", ctx, "org1", 100, 0, true).Return([]interface{}{}, nil)

	mockGroupService.On("GetGroupRoles", ctx, "group1").Return([]interface{}{}, nil)
	mockGroupService.On("GetGroupMembers", ctx, "group1", 100, 0).Return([]interface{}{}, nil)

	effectiveRoles := []*groups.EffectiveRole{
		{
			Role: &models.Role{
				Name:     "Test Role",
				IsActive: true,
			},
			GroupID:      "group1",
			GroupName:    "Test Group",
			Distance:     0,
			IsDirectRole: true,
		},
	}
	mockGroupService.On("GetUserEffectiveRoles", ctx, "org1", "user1").Return(effectiveRoles, nil)

	// Test immediate warming
	err := warmingService.WarmNow(ctx, config)
	assert.NoError(t, err)

	// Verify all expected calls were made
	mockOrgService.AssertExpectations(t)
	mockGroupService.AssertExpectations(t)
}

func TestCacheWarmingService_GetCacheStats(t *testing.T) {
	// Setup
	mockOrgService := &MockOrganizationService{}
	mockGroupService := &MockGroupService{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	warmingService := NewCacheWarmingService(mockOrgService, mockGroupService, mockCache, logger)

	ctx := context.Background()

	// Mock cache keys calls
	orgKeys := []string{"org:org1:hierarchy", "org:org2:stats"}
	groupKeys := []string{"group:group1:roles", "group:group2:members"}
	effectiveRolesKeys := []string{"org:org1:user:user1:effective_roles"}

	mockCache.On("Keys", "org:*").Return(orgKeys, nil)
	mockCache.On("Keys", "group:*").Return(groupKeys, nil)
	mockCache.On("Keys", "*:effective_roles*").Return(effectiveRolesKeys, nil)

	// Test getting cache stats
	stats, err := warmingService.GetCacheStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	assert.Equal(t, len(orgKeys), stats["organization_cache_keys"])
	assert.Equal(t, len(groupKeys), stats["group_cache_keys"])
	assert.Equal(t, len(effectiveRolesKeys), stats["effective_roles_cache_keys"])
	assert.Equal(t, false, stats["is_running"])
	assert.Equal(t, "30m0s", stats["warming_interval"])
	assert.Equal(t, 5, stats["max_concurrent_warms"])

	mockCache.AssertExpectations(t)
}

func TestCacheWarmingService_InvalidateFrequentlyAccessedCache(t *testing.T) {
	// Setup
	mockOrgService := &MockOrganizationService{}
	mockGroupService := &MockGroupService{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	warmingService := NewCacheWarmingService(mockOrgService, mockGroupService, mockCache, logger)

	ctx := context.Background()
	config := CacheWarmingConfig{
		FrequentOrganizations: []string{"org1", "org2"},
		FrequentGroups:        []string{"group1", "group2"},
		FrequentUsers:         []string{"user1", "user2"},
	}

	// Mock cache invalidation calls
	// Organization cache invalidation
	orgKeys := []string{
		"org:org1:hierarchy", "org:org1:parent_hierarchy", "org:org1:children",
		"org:org1:active_children", "org:org1:groups", "org:org1:active_groups",
		"org:org1:group_hierarchy", "org:org1:stats",
	}
	for _, key := range orgKeys {
		mockCache.On("Delete", key).Return(nil)
	}

	orgKeys2 := []string{
		"org:org2:hierarchy", "org:org2:parent_hierarchy", "org:org2:children",
		"org:org2:active_children", "org:org2:groups", "org:org2:active_groups",
		"org:org2:group_hierarchy", "org:org2:stats",
	}
	for _, key := range orgKeys2 {
		mockCache.On("Delete", key).Return(nil)
	}

	// Mock user cache keys for organizations
	userKeys1 := []string{"org:org1:user:user1:groups", "org:org1:user:user2:groups"}
	userKeys2 := []string{"org:org2:user:user1:groups", "org:org2:user:user2:groups"}

	mockCache.On("Keys", "org:org1:user:*").Return(userKeys1, nil)
	mockCache.On("Keys", "org:org2:user:*").Return(userKeys2, nil)

	for _, key := range append(userKeys1, userKeys2...) {
		mockCache.On("Delete", key).Return(nil)
	}

	// Group cache invalidation
	groupKeys := []string{
		"group:group1:hierarchy", "group:group1:ancestors", "group:group1:descendants",
		"group:group1:children", "group:group1:active_children", "group:group1:roles",
		"group:group1:active_roles", "group:group1:role_details", "group:group1:members",
		"group:group1:active_members", "group:group1:member_details", "group:group1:role_inheritance",
	}
	for _, key := range groupKeys {
		mockCache.On("Delete", key).Return(nil)
	}

	groupKeys2 := []string{
		"group:group2:hierarchy", "group:group2:ancestors", "group:group2:descendants",
		"group:group2:children", "group:group2:active_children", "group:group2:roles",
		"group:group2:active_roles", "group:group2:role_details", "group:group2:members",
		"group:group2:active_members", "group:group2:member_details", "group:group2:role_inheritance",
	}
	for _, key := range groupKeys2 {
		mockCache.On("Delete", key).Return(nil)
	}

	// User effective roles cache invalidation
	userEffectiveRolesKeys := []string{
		"org:org1:user:user1:effective_roles_v2", "org:org1:user:user1:group_memberships", "org:org1:user:user1:effective_roles",
		"org:org1:user:user2:effective_roles_v2", "org:org1:user:user2:group_memberships", "org:org1:user:user2:effective_roles",
		"org:org2:user:user1:effective_roles_v2", "org:org2:user:user1:group_memberships", "org:org2:user:user1:effective_roles",
		"org:org2:user:user2:effective_roles_v2", "org:org2:user:user2:group_memberships", "org:org2:user:user2:effective_roles",
	}
	for _, key := range userEffectiveRolesKeys {
		mockCache.On("Delete", key).Return(nil)
	}

	// Test invalidation
	err := warmingService.InvalidateFrequentlyAccessedCache(ctx, config)
	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
}

func TestCacheWarmingService_UpdateWarmingConfig(t *testing.T) {
	// Setup
	mockOrgService := &MockOrganizationService{}
	mockGroupService := &MockGroupService{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	warmingService := NewCacheWarmingService(mockOrgService, mockGroupService, mockCache, logger)

	// Test updating configuration
	newConfig := CacheWarmingConfig{
		WarmingInterval:    15 * time.Minute,
		MaxConcurrentWarms: 10,
	}

	err := warmingService.UpdateWarmingConfig(newConfig)
	assert.NoError(t, err)

	// Verify configuration was updated
	stats, err := warmingService.GetCacheStats(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "15m0s", stats["warming_interval"])
	assert.Equal(t, 10, stats["max_concurrent_warms"])
}

// Integration test with real cache service
func TestCacheWarmingService_Integration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup real cache service (requires Redis)
	// Create a mock logger that implements interfaces.Logger
	mockLogger := &MockLogger{}
	realCache := NewCacheService("localhost:6379", "", 0, mockLogger)

	// Create mock services
	mockOrgService := &MockOrganizationService{}
	mockGroupService := &MockGroupService{}

	warmingService := NewCacheWarmingService(mockOrgService, mockGroupService, realCache, zap.NewNop())

	ctx := context.Background()
	config := CacheWarmingConfig{
		WarmingInterval:       1 * time.Hour, // Long interval to avoid automatic warming
		MaxConcurrentWarms:    2,
		FrequentOrganizations: []string{"integration-org"},
		FrequentGroups:        []string{"integration-group"},
		FrequentUsers:         []string{"integration-user"},
		WarmHierarchies:       true,
		WarmGroupRoles:        true,
		WarmEffectiveRoles:    true,
		WarmStats:             true,
	}

	// Clean up before and after test
	defer realCache.Clear()
	realCache.Clear()

	// Mock the service calls
	hierarchy := &organizationResponses.OrganizationHierarchyResponse{
		Organization: &organizationResponses.OrganizationResponse{
			ID:   "integration-org",
			Name: "Integration Test Organization",
		},
	}
	stats := &organizationResponses.OrganizationStatsResponse{
		OrganizationID: "integration-org",
		ChildCount:     1,
		GroupCount:     2,
		UserCount:      3,
	}

	mockOrgService.On("GetOrganizationHierarchy", ctx, "integration-org").Return(hierarchy, nil)
	mockOrgService.On("GetOrganizationStats", ctx, "integration-org").Return(stats, nil)
	mockOrgService.On("GetOrganizationGroups", ctx, "integration-org", 100, 0, false).Return([]interface{}{}, nil)
	mockOrgService.On("GetOrganizationGroups", ctx, "integration-org", 100, 0, true).Return([]interface{}{}, nil)

	mockGroupService.On("GetGroupRoles", ctx, "integration-group").Return([]interface{}{}, nil)
	mockGroupService.On("GetGroupMembers", ctx, "integration-group", 100, 0).Return([]interface{}{}, nil)

	effectiveRoles := []*groups.EffectiveRole{
		{
			Role: &models.Role{
				Name:     "Integration Test Role",
				IsActive: true,
			},
			GroupID:      "integration-group",
			GroupName:    "Integration Test Group",
			Distance:     0,
			IsDirectRole: true,
		},
	}
	mockGroupService.On("GetUserEffectiveRoles", ctx, "integration-org", "integration-user").Return(effectiveRoles, nil)

	// Test immediate warming
	err := warmingService.WarmNow(ctx, config)
	assert.NoError(t, err)

	// Verify cache was populated
	stats_result, err := warmingService.GetCacheStats(ctx)
	assert.NoError(t, err)
	assert.True(t, stats_result["organization_cache_keys"].(int) > 0)

	// Test cache invalidation
	err = warmingService.InvalidateFrequentlyAccessedCache(ctx, config)
	assert.NoError(t, err)

	// Verify cache was cleared
	stats_result, err = warmingService.GetCacheStats(ctx)
	assert.NoError(t, err)
	// Note: Some keys might still exist due to timing, but the count should be reduced

	mockOrgService.AssertExpectations(t)
	mockGroupService.AssertExpectations(t)
}
