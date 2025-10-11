//go:build integration
// +build integration

package groups

import (
	"context"
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// TestRoleInheritanceEngine_BottomUpInheritance_SingleLevel tests bottom-up inheritance with one level
func TestRoleInheritanceEngine_BottomUpInheritance_SingleLevel(t *testing.T) {
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

	// Create test groups: parent -> child
	parentGroupID := "parent-group-123"
	parentGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Parent Group",
		OrganizationID: orgID,
		IsActive:       true,
	}
	parentGroup.BaseModel.ID = parentGroupID

	childGroupID := "child-group-456"
	childGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Child Group",
		OrganizationID: orgID,
		ParentID:       &parentGroupID,
		IsActive:       true,
	}
	childGroup.BaseModel.ID = childGroupID

	// Create test roles
	parentRoleID := "parent-role-789"
	parentRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Parent Role",
		Description: "Role assigned to parent group",
		IsActive:    true,
	}
	parentRole.BaseModel.ID = parentRoleID

	childRoleID := "child-role-101"
	childRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Child Role",
		Description: "Role assigned to child group",
		IsActive:    true,
	}
	childRole.BaseModel.ID = childRoleID

	// Create group roles
	parentGroupRole := &models.GroupRole{
		GroupID:        parentGroupID,
		RoleID:         parentRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	childGroupRole := &models.GroupRole{
		GroupID:        childGroupID,
		RoleID:         childRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	// Set up mocks
	// Cache misses
	mockCache.On("Get", "org:org-123:user:user-456:effective_roles").Return(nil, false)
	mockCache.On("Get", "org:org-123:user:user-456:groups").Return(nil, false)

	// User is member of parent group only
	mockGroupMembershipRepo.On("GetUserDirectGroups", ctx, orgID, userID).Return([]*models.Group{parentGroup}, nil)
	mockCache.On("Set", "org:org-123:user:user-456:groups", []*models.Group{parentGroup}, 300).Return(nil)

	// Parent group has child group
	mockGroupRepo.On("GetChildren", ctx, parentGroupID).Return([]*models.Group{childGroup}, nil)

	// Child group has no children
	mockGroupRepo.On("GetChildren", ctx, childGroupID).Return([]*models.Group{}, nil)

	// Group roles
	mockGroupRoleRepo.On("GetByGroupID", ctx, parentGroupID).Return([]*models.GroupRole{parentGroupRole}, nil)
	mockGroupRoleRepo.On("GetByGroupID", ctx, childGroupID).Return([]*models.GroupRole{childGroupRole}, nil)

	// Role details
	mockRoleRepo.On("GetByID", ctx, parentRoleID, mock.AnythingOfType("*models.Role")).Return(parentRole, nil)
	mockRoleRepo.On("GetByID", ctx, childRoleID, mock.AnythingOfType("*models.Role")).Return(childRole, nil)

	// Cache set for effective roles
	mockCache.On("Set", "org:org-123:user:user-456:effective_roles", mock.AnythingOfType("[]*groups.EffectiveRole"), 300).Return(nil)

	// Call the method
	effectiveRoles, err := engine.CalculateEffectiveRoles(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, effectiveRoles)
	assert.Len(t, effectiveRoles, 2) // Should have both parent and child roles

	// Find roles by ID
	var parentEffectiveRole, childEffectiveRole *EffectiveRole
	for _, role := range effectiveRoles {
		if role.Role.GetID() == parentRoleID {
			parentEffectiveRole = role
		} else if role.Role.GetID() == childRoleID {
			childEffectiveRole = role
		}
	}

	// Verify parent role (direct assignment)
	assert.NotNil(t, parentEffectiveRole)
	assert.Equal(t, parentRoleID, parentEffectiveRole.Role.GetID())
	assert.Equal(t, parentGroupID, parentEffectiveRole.GroupID)
	assert.Equal(t, 0, parentEffectiveRole.Distance) // Direct assignment
	assert.True(t, parentEffectiveRole.IsDirectRole)
	assert.Equal(t, []string{parentGroupID}, parentEffectiveRole.InheritancePath)

	// Verify child role (inherited from child group)
	assert.NotNil(t, childEffectiveRole)
	assert.Equal(t, childRoleID, childEffectiveRole.Role.GetID())
	assert.Equal(t, childGroupID, childEffectiveRole.GroupID)
	assert.Equal(t, 1, childEffectiveRole.Distance) // Inherited from child
	assert.False(t, childEffectiveRole.IsDirectRole)
	assert.Equal(t, []string{parentGroupID, childGroupID}, childEffectiveRole.InheritancePath)

	// Verify mocks were called
	mockCache.AssertExpectations(t)
	mockGroupMembershipRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockGroupRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// TestRoleInheritanceEngine_BottomUpInheritance_ConflictResolution tests role conflict resolution
func TestRoleInheritanceEngine_BottomUpInheritance_ConflictResolution(t *testing.T) {
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

	// Create test groups: parent -> child
	parentGroupID := "parent-group-123"
	parentGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Parent Group",
		OrganizationID: orgID,
		IsActive:       true,
	}
	parentGroup.BaseModel.ID = parentGroupID

	childGroupID := "child-group-456"
	childGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Child Group",
		OrganizationID: orgID,
		ParentID:       &parentGroupID,
		IsActive:       true,
	}
	childGroup.BaseModel.ID = childGroupID

	// Create test role (same role assigned to both groups)
	sharedRoleID := "shared-role-789"
	sharedRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Shared Role",
		Description: "Role assigned to both parent and child groups",
		IsActive:    true,
	}
	sharedRole.BaseModel.ID = sharedRoleID

	// Create group roles
	parentGroupRole := &models.GroupRole{
		GroupID:        parentGroupID,
		RoleID:         sharedRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	childGroupRole := &models.GroupRole{
		GroupID:        childGroupID,
		RoleID:         sharedRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	// Set up mocks
	// Cache misses
	mockCache.On("Get", "org:org-123:user:user-456:effective_roles").Return(nil, false)
	mockCache.On("Get", "org:org-123:user:user-456:groups").Return(nil, false)

	// User is member of parent group only
	mockGroupMembershipRepo.On("GetUserDirectGroups", ctx, orgID, userID).Return([]*models.Group{parentGroup}, nil)
	mockCache.On("Set", "org:org-123:user:user-456:groups", []*models.Group{parentGroup}, 300).Return(nil)

	// Parent group has child group
	mockGroupRepo.On("GetChildren", ctx, parentGroupID).Return([]*models.Group{childGroup}, nil)

	// Child group has no children
	mockGroupRepo.On("GetChildren", ctx, childGroupID).Return([]*models.Group{}, nil)

	// Group roles (both groups have the same role)
	mockGroupRoleRepo.On("GetByGroupID", ctx, parentGroupID).Return([]*models.GroupRole{parentGroupRole}, nil)
	mockGroupRoleRepo.On("GetByGroupID", ctx, childGroupID).Return([]*models.GroupRole{childGroupRole}, nil)

	// Role details (called twice, once for each group)
	mockRoleRepo.On("GetByID", ctx, sharedRoleID, mock.AnythingOfType("*models.Role")).Return(sharedRole, nil)

	// Cache set for effective roles
	mockCache.On("Set", "org:org-123:user:user-456:effective_roles", mock.AnythingOfType("[]*groups.EffectiveRole"), 300).Return(nil)

	// Call the method
	effectiveRoles, err := engine.CalculateEffectiveRoles(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, effectiveRoles)
	assert.Len(t, effectiveRoles, 1) // Should have only one role (conflict resolved)

	// Verify the role (should be the direct assignment, not inherited)
	effectiveRole := effectiveRoles[0]
	assert.Equal(t, sharedRoleID, effectiveRole.Role.GetID())
	assert.Equal(t, parentGroupID, effectiveRole.GroupID) // Should be from parent (distance 0)
	assert.Equal(t, 0, effectiveRole.Distance)            // Direct assignment wins
	assert.True(t, effectiveRole.IsDirectRole)
	assert.Equal(t, []string{parentGroupID}, effectiveRole.InheritancePath)

	// Verify mocks were called
	mockCache.AssertExpectations(t)
	mockGroupMembershipRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockGroupRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}
