//go:build integration
// +build integration

package groups

import (
	"context"
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockGroupRepository is a mock implementation of GroupRepositoryInterface
type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) GetChildren(ctx context.Context, parentID string) ([]*models.Group, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]*models.Group), args.Error(1)
}

// MockGroupRoleRepository is a mock implementation of GroupRoleRepositoryInterface
type MockGroupRoleRepository struct {
	mock.Mock
}

func (m *MockGroupRoleRepository) GetByGroupID(ctx context.Context, groupID string) ([]*models.GroupRole, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]*models.GroupRole), args.Error(1)
}

// MockRoleRepository is a mock implementation of RoleRepositoryInterface
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id string, role *models.Role) (*models.Role, error) {
	args := m.Called(ctx, id, role)
	return args.Get(0).(*models.Role), args.Error(1)
}

// MockGroupMembershipRepository is a mock implementation of GroupMembershipRepositoryInterface
type MockGroupMembershipRepository struct {
	mock.Mock
}

func (m *MockGroupMembershipRepository) GetUserDirectGroups(ctx context.Context, orgID, userID string) ([]*models.Group, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Get(0).([]*models.Group), args.Error(1)
}

// MockCacheService is a mock implementation of interfaces.CacheService
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(key string) (interface{}, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

func (m *MockCacheService) Set(key string, value interface{}, ttl int) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheService) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCacheService) Exists(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockCacheService) Clear() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheService) Keys(pattern string) ([]string, error) {
	args := m.Called(pattern)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheService) Expire(key string, ttl int) error {
	args := m.Called(key, ttl)
	return args.Error(0)
}

func (m *MockCacheService) TTL(key string) (int, error) {
	args := m.Called(key)
	return args.Int(0), args.Error(1)
}

func (m *MockCacheService) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewRoleInheritanceEngine(t *testing.T) {
	mockGroupRepo := &MockGroupRepository{}
	mockGroupRoleRepo := &MockGroupRoleRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockGroupMembershipRepo := &MockGroupMembershipRepository{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	engine := NewRoleInheritanceEngine(
		mockGroupRepo,
		mockGroupRoleRepo,
		mockRoleRepo,
		mockGroupMembershipRepo,
		mockCache,
		logger,
	)

	assert.NotNil(t, engine)
	assert.Equal(t, mockGroupRepo, engine.groupRepo)
	assert.Equal(t, mockGroupRoleRepo, engine.groupRoleRepo)
	assert.Equal(t, mockRoleRepo, engine.roleRepo)
	assert.Equal(t, mockGroupMembershipRepo, engine.groupMembershipRepo)
	assert.Equal(t, mockCache, engine.cache)
	assert.Equal(t, logger, engine.logger)
}

func TestRoleInheritanceEngine_getUserDirectGroups(t *testing.T) {
	mockGroupRepo := &MockGroupRepository{}
	mockGroupRoleRepo := &MockGroupRoleRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockGroupMembershipRepo := &MockGroupMembershipRepository{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	engine := NewRoleInheritanceEngine(
		mockGroupRepo,
		mockGroupRoleRepo,
		mockRoleRepo,
		mockGroupMembershipRepo,
		mockCache,
		logger,
	)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"

	// Mock expected groups
	expectedGroups := []*models.Group{
		{
			Name:           "Test Group 1",
			OrganizationID: orgID,
			IsActive:       true,
		},
		{
			Name:           "Test Group 2",
			OrganizationID: orgID,
			IsActive:       true,
		},
	}

	// Set up mocks
	mockCache.On("Get", "org:org-123:user:user-456:groups").Return(nil, false)
	mockGroupMembershipRepo.On("GetUserDirectGroups", ctx, orgID, userID).Return(expectedGroups, nil)
	mockCache.On("Set", "org:org-123:user:user-456:groups", expectedGroups, 300).Return(nil)

	// Call the method
	groups, err := engine.getUserDirectGroups(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedGroups, groups)
	assert.Len(t, groups, 2)

	// Verify mocks were called
	mockCache.AssertExpectations(t)
	mockGroupMembershipRepo.AssertExpectations(t)
}

func TestRoleInheritanceEngine_getUserDirectGroups_FromCache(t *testing.T) {
	mockGroupRepo := &MockGroupRepository{}
	mockGroupRoleRepo := &MockGroupRoleRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockGroupMembershipRepo := &MockGroupMembershipRepository{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	engine := NewRoleInheritanceEngine(
		mockGroupRepo,
		mockGroupRoleRepo,
		mockRoleRepo,
		mockGroupMembershipRepo,
		mockCache,
		logger,
	)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"

	// Mock cached groups
	cachedGroups := []*models.Group{
		{
			Name:           "Cached Group",
			OrganizationID: orgID,
			IsActive:       true,
		},
	}

	// Set up mocks - cache hit
	mockCache.On("Get", "org:org-123:user:user-456:groups").Return(cachedGroups, true)

	// Call the method
	groups, err := engine.getUserDirectGroups(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, cachedGroups, groups)
	assert.Len(t, groups, 1)

	// Verify mocks were called (repository should NOT be called)
	mockCache.AssertExpectations(t)
	mockGroupMembershipRepo.AssertNotCalled(t, "GetUserDirectGroups")
}

func TestRoleInheritanceEngine_CalculateEffectiveRoles_NoGroups(t *testing.T) {
	mockGroupRepo := &MockGroupRepository{}
	mockGroupRoleRepo := &MockGroupRoleRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockGroupMembershipRepo := &MockGroupMembershipRepository{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	engine := NewRoleInheritanceEngine(
		mockGroupRepo,
		mockGroupRoleRepo,
		mockRoleRepo,
		mockGroupMembershipRepo,
		mockCache,
		logger,
	)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"

	// Set up mocks - no cached effective roles, no groups
	mockCache.On("Get", "org:org-123:user:user-456:effective_roles").Return(nil, false)
	mockCache.On("Get", "org:org-123:user:user-456:groups").Return(nil, false)
	mockGroupMembershipRepo.On("GetUserDirectGroups", ctx, orgID, userID).Return([]*models.Group{}, nil)
	mockCache.On("Set", "org:org-123:user:user-456:groups", []*models.Group{}, 300).Return(nil)
	mockCache.On("Set", "org:org-123:user:user-456:effective_roles", mock.Anything, 300).Return(nil)

	// Call the method
	effectiveRoles, err := engine.CalculateEffectiveRoles(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, effectiveRoles)
	assert.Len(t, effectiveRoles, 0)

	// Verify mocks were called
	mockCache.AssertExpectations(t)
	mockGroupMembershipRepo.AssertExpectations(t)
}

func TestRoleInheritanceEngine_InvalidateUserRoleCache(t *testing.T) {
	mockGroupRepo := &MockGroupRepository{}
	mockGroupRoleRepo := &MockGroupRoleRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockGroupMembershipRepo := &MockGroupMembershipRepository{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	engine := NewRoleInheritanceEngine(
		mockGroupRepo,
		mockGroupRoleRepo,
		mockRoleRepo,
		mockGroupMembershipRepo,
		mockCache,
		logger,
	)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"

	// Set up mocks
	mockCache.On("Delete", "org:org-123:user:user-456:effective_roles").Return(nil)

	// Call the method
	err := engine.InvalidateUserRoleCache(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)

	// Verify mocks were called
	mockCache.AssertExpectations(t)
}

func TestRoleInheritanceEngine_InvalidateGroupRoleCache(t *testing.T) {
	mockGroupRepo := &MockGroupRepository{}
	mockGroupRoleRepo := &MockGroupRoleRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockGroupMembershipRepo := &MockGroupMembershipRepository{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	engine := NewRoleInheritanceEngine(
		mockGroupRepo,
		mockGroupRoleRepo,
		mockRoleRepo,
		mockGroupMembershipRepo,
		mockCache,
		logger,
	)

	ctx := context.Background()
	orgID := "org-123"
	groupID := "group-789"

	// Set up mocks
	mockCache.On("Keys", "org:org-123:user:*:effective_roles").Return([]string{
		"org:org-123:user:user-456:effective_roles",
		"org:org-123:user:user-789:effective_roles",
	}, nil)
	mockCache.On("Delete", "org:org-123:user:user-456:effective_roles").Return(nil)
	mockCache.On("Delete", "org:org-123:user:user-789:effective_roles").Return(nil)

	// Call the method
	err := engine.InvalidateGroupRoleCache(ctx, orgID, groupID)

	// Assertions
	assert.NoError(t, err)

	// Verify mocks were called
	mockCache.AssertExpectations(t)
}

func TestEffectiveRole_Structure(t *testing.T) {
	role := &models.Role{
		Name:        "Test Role",
		Description: "Test Description",
		IsActive:    true,
	}

	effectiveRole := &EffectiveRole{
		Role:            role,
		GroupID:         "group-123",
		GroupName:       "Test Group",
		InheritancePath: []string{"group-123", "child-group-456"},
		Distance:        1,
		IsDirectRole:    false,
	}

	assert.Equal(t, role, effectiveRole.Role)
	assert.Equal(t, "group-123", effectiveRole.GroupID)
	assert.Equal(t, "Test Group", effectiveRole.GroupName)
	assert.Equal(t, []string{"group-123", "child-group-456"}, effectiveRole.InheritancePath)
	assert.Equal(t, 1, effectiveRole.Distance)
	assert.False(t, effectiveRole.IsDirectRole)
}
