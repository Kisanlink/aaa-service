//go:build integration
// +build integration

package groups

import (
	"context"
	"testing"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// TestRoleInheritanceEngine_BottomUpInheritance_MultiLevel tests bottom-up inheritance with multiple levels
// This test verifies that parent groups inherit roles from all descendant groups in the hierarchy
func TestRoleInheritanceEngine_BottomUpInheritance_MultiLevel(t *testing.T) {
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

	// Create test hierarchy: CEO -> Manager -> Employee -> Intern
	ceoGroupID := "ceo-group-123"
	ceoGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "CEO Group",
		OrganizationID: orgID,
		IsActive:       true,
	}
	ceoGroup.BaseModel.ID = ceoGroupID

	managerGroupID := "manager-group-456"
	managerGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Manager Group",
		OrganizationID: orgID,
		ParentID:       &ceoGroupID,
		IsActive:       true,
	}
	managerGroup.BaseModel.ID = managerGroupID

	employeeGroupID := "employee-group-789"
	employeeGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Employee Group",
		OrganizationID: orgID,
		ParentID:       &managerGroupID,
		IsActive:       true,
	}
	employeeGroup.BaseModel.ID = employeeGroupID

	internGroupID := "intern-group-101"
	internGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Intern Group",
		OrganizationID: orgID,
		ParentID:       &employeeGroupID,
		IsActive:       true,
	}
	internGroup.BaseModel.ID = internGroupID

	// Create test roles for each level
	ceoRoleID := "ceo-role-111"
	ceoRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "CEO Role",
		Description: "Executive permissions",
		IsActive:    true,
	}
	ceoRole.BaseModel.ID = ceoRoleID

	managerRoleID := "manager-role-222"
	managerRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Manager Role",
		Description: "Management permissions",
		IsActive:    true,
	}
	managerRole.BaseModel.ID = managerRoleID

	employeeRoleID := "employee-role-333"
	employeeRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Employee Role",
		Description: "Standard employee permissions",
		IsActive:    true,
	}
	employeeRole.BaseModel.ID = employeeRoleID

	internRoleID := "intern-role-444"
	internRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Intern Role",
		Description: "Limited intern permissions",
		IsActive:    true,
	}
	internRole.BaseModel.ID = internRoleID

	// Create group roles
	ceoGroupRole := &models.GroupRole{
		GroupID:        ceoGroupID,
		RoleID:         ceoRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	managerGroupRole := &models.GroupRole{
		GroupID:        managerGroupID,
		RoleID:         managerRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	employeeGroupRole := &models.GroupRole{
		GroupID:        employeeGroupID,
		RoleID:         employeeRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	internGroupRole := &models.GroupRole{
		GroupID:        internGroupID,
		RoleID:         internRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	// Set up mocks
	// Cache misses
	mockCache.On("Get", "org:org-123:user:user-456:effective_roles").Return(nil, false)
	mockCache.On("Get", "org:org-123:user:user-456:groups").Return(nil, false)

	// User is member of CEO group only
	mockGroupMembershipRepo.On("GetUserDirectGroups", ctx, orgID, userID).Return([]*models.Group{ceoGroup}, nil)
	mockCache.On("Set", "org:org-123:user:user-456:groups", []*models.Group{ceoGroup}, 300).Return(nil)

	// Set up group hierarchy relationships
	mockGroupRepo.On("GetChildren", ctx, ceoGroupID).Return([]*models.Group{managerGroup}, nil)
	mockGroupRepo.On("GetChildren", ctx, managerGroupID).Return([]*models.Group{employeeGroup}, nil)
	mockGroupRepo.On("GetChildren", ctx, employeeGroupID).Return([]*models.Group{internGroup}, nil)
	mockGroupRepo.On("GetChildren", ctx, internGroupID).Return([]*models.Group{}, nil)

	// Group roles
	mockGroupRoleRepo.On("GetByGroupID", ctx, ceoGroupID).Return([]*models.GroupRole{ceoGroupRole}, nil)
	mockGroupRoleRepo.On("GetByGroupID", ctx, managerGroupID).Return([]*models.GroupRole{managerGroupRole}, nil)
	mockGroupRoleRepo.On("GetByGroupID", ctx, employeeGroupID).Return([]*models.GroupRole{employeeGroupRole}, nil)
	mockGroupRoleRepo.On("GetByGroupID", ctx, internGroupID).Return([]*models.GroupRole{internGroupRole}, nil)

	// Role details
	mockRoleRepo.On("GetByID", ctx, ceoRoleID, mock.AnythingOfType("*models.Role")).Return(ceoRole, nil)
	mockRoleRepo.On("GetByID", ctx, managerRoleID, mock.AnythingOfType("*models.Role")).Return(managerRole, nil)
	mockRoleRepo.On("GetByID", ctx, employeeRoleID, mock.AnythingOfType("*models.Role")).Return(employeeRole, nil)
	mockRoleRepo.On("GetByID", ctx, internRoleID, mock.AnythingOfType("*models.Role")).Return(internRole, nil)

	// Cache set for effective roles
	mockCache.On("Set", "org:org-123:user:user-456:effective_roles", mock.AnythingOfType("[]*groups.EffectiveRole"), 300).Return(nil)

	// Call the method
	effectiveRoles, err := engine.CalculateEffectiveRoles(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, effectiveRoles)
	assert.Len(t, effectiveRoles, 4) // Should have all four roles

	// Create a map for easier verification
	roleMap := make(map[string]*EffectiveRole)
	for _, role := range effectiveRoles {
		roleMap[role.Role.GetID()] = role
	}

	// Verify CEO role (direct assignment - distance 0)
	ceoEffectiveRole := roleMap[ceoRoleID]
	assert.NotNil(t, ceoEffectiveRole)
	assert.Equal(t, ceoRoleID, ceoEffectiveRole.Role.GetID())
	assert.Equal(t, ceoGroupID, ceoEffectiveRole.GroupID)
	assert.Equal(t, 0, ceoEffectiveRole.Distance)
	assert.True(t, ceoEffectiveRole.IsDirectRole)
	assert.Equal(t, []string{ceoGroupID}, ceoEffectiveRole.InheritancePath)

	// Verify Manager role (inherited from child - distance 1)
	managerEffectiveRole := roleMap[managerRoleID]
	assert.NotNil(t, managerEffectiveRole)
	assert.Equal(t, managerRoleID, managerEffectiveRole.Role.GetID())
	assert.Equal(t, managerGroupID, managerEffectiveRole.GroupID)
	assert.Equal(t, 1, managerEffectiveRole.Distance)
	assert.False(t, managerEffectiveRole.IsDirectRole)
	assert.Equal(t, []string{ceoGroupID, managerGroupID}, managerEffectiveRole.InheritancePath)

	// Verify Employee role (inherited from grandchild - distance 2)
	employeeEffectiveRole := roleMap[employeeRoleID]
	assert.NotNil(t, employeeEffectiveRole)
	assert.Equal(t, employeeRoleID, employeeEffectiveRole.Role.GetID())
	assert.Equal(t, employeeGroupID, employeeEffectiveRole.GroupID)
	assert.Equal(t, 2, employeeEffectiveRole.Distance)
	assert.False(t, employeeEffectiveRole.IsDirectRole)
	assert.Equal(t, []string{ceoGroupID, managerGroupID, employeeGroupID}, employeeEffectiveRole.InheritancePath)

	// Verify Intern role (inherited from great-grandchild - distance 3)
	internEffectiveRole := roleMap[internRoleID]
	assert.NotNil(t, internEffectiveRole)
	assert.Equal(t, internRoleID, internEffectiveRole.Role.GetID())
	assert.Equal(t, internGroupID, internEffectiveRole.GroupID)
	assert.Equal(t, 3, internEffectiveRole.Distance)
	assert.False(t, internEffectiveRole.IsDirectRole)
	assert.Equal(t, []string{ceoGroupID, managerGroupID, employeeGroupID, internGroupID}, internEffectiveRole.InheritancePath)

	// Verify roles are sorted by precedence (distance ascending)
	assert.Equal(t, 0, effectiveRoles[0].Distance) // CEO role first
	assert.Equal(t, 1, effectiveRoles[1].Distance) // Manager role second
	assert.Equal(t, 2, effectiveRoles[2].Distance) // Employee role third
	assert.Equal(t, 3, effectiveRoles[3].Distance) // Intern role last

	// Verify mocks were called
	mockCache.AssertExpectations(t)
	mockGroupMembershipRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockGroupRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// TestRoleInheritanceEngine_BottomUpInheritance_WideHierarchy tests inheritance with multiple child groups
// This test verifies that parent groups inherit roles from all their child groups, not just one branch
func TestRoleInheritanceEngine_BottomUpInheritance_WideHierarchy(t *testing.T) {
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

	// Create test hierarchy with multiple branches:
	// CEO Group
	// ├── Engineering Group
	// │   └── Backend Team
	// ├── Sales Group
	// │   └── Sales Team
	// └── Marketing Group
	//     └── Marketing Team

	ceoGroupID := "ceo-group-123"
	ceoGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "CEO Group",
		OrganizationID: orgID,
		IsActive:       true,
	}
	ceoGroup.BaseModel.ID = ceoGroupID

	engineeringGroupID := "engineering-group-456"
	engineeringGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Engineering Group",
		OrganizationID: orgID,
		ParentID:       &ceoGroupID,
		IsActive:       true,
	}
	engineeringGroup.BaseModel.ID = engineeringGroupID

	backendTeamID := "backend-team-789"
	backendTeam := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Backend Team",
		OrganizationID: orgID,
		ParentID:       &engineeringGroupID,
		IsActive:       true,
	}
	backendTeam.BaseModel.ID = backendTeamID

	salesGroupID := "sales-group-101"
	salesGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Sales Group",
		OrganizationID: orgID,
		ParentID:       &ceoGroupID,
		IsActive:       true,
	}
	salesGroup.BaseModel.ID = salesGroupID

	salesTeamID := "sales-team-202"
	salesTeam := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Sales Team",
		OrganizationID: orgID,
		ParentID:       &salesGroupID,
		IsActive:       true,
	}
	salesTeam.BaseModel.ID = salesTeamID

	marketingGroupID := "marketing-group-303"
	marketingGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Marketing Group",
		OrganizationID: orgID,
		ParentID:       &ceoGroupID,
		IsActive:       true,
	}
	marketingGroup.BaseModel.ID = marketingGroupID

	marketingTeamID := "marketing-team-404"
	marketingTeam := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Marketing Team",
		OrganizationID: orgID,
		ParentID:       &marketingGroupID,
		IsActive:       true,
	}
	marketingTeam.BaseModel.ID = marketingTeamID

	// Create roles for each group
	roles := map[string]*models.Role{
		"ceo-role":            {BaseModel: &base.BaseModel{}, Name: "CEO Role", IsActive: true},
		"engineering-role":    {BaseModel: &base.BaseModel{}, Name: "Engineering Role", IsActive: true},
		"backend-role":        {BaseModel: &base.BaseModel{}, Name: "Backend Role", IsActive: true},
		"sales-role":          {BaseModel: &base.BaseModel{}, Name: "Sales Role", IsActive: true},
		"sales-team-role":     {BaseModel: &base.BaseModel{}, Name: "Sales Team Role", IsActive: true},
		"marketing-role":      {BaseModel: &base.BaseModel{}, Name: "Marketing Role", IsActive: true},
		"marketing-team-role": {BaseModel: &base.BaseModel{}, Name: "Marketing Team Role", IsActive: true},
	}

	// Set IDs for roles
	for id, role := range roles {
		role.BaseModel.ID = id
	}

	// Create group roles
	groupRoles := map[string][]*models.GroupRole{
		ceoGroupID:         {{GroupID: ceoGroupID, RoleID: "ceo-role", OrganizationID: orgID, IsActive: true}},
		engineeringGroupID: {{GroupID: engineeringGroupID, RoleID: "engineering-role", OrganizationID: orgID, IsActive: true}},
		backendTeamID:      {{GroupID: backendTeamID, RoleID: "backend-role", OrganizationID: orgID, IsActive: true}},
		salesGroupID:       {{GroupID: salesGroupID, RoleID: "sales-role", OrganizationID: orgID, IsActive: true}},
		salesTeamID:        {{GroupID: salesTeamID, RoleID: "sales-team-role", OrganizationID: orgID, IsActive: true}},
		marketingGroupID:   {{GroupID: marketingGroupID, RoleID: "marketing-role", OrganizationID: orgID, IsActive: true}},
		marketingTeamID:    {{GroupID: marketingTeamID, RoleID: "marketing-team-role", OrganizationID: orgID, IsActive: true}},
	}

	// Set up mocks
	mockCache.On("Get", "org:org-123:user:user-456:effective_roles").Return(nil, false)
	mockCache.On("Get", "org:org-123:user:user-456:groups").Return(nil, false)

	// User is member of CEO group only
	mockGroupMembershipRepo.On("GetUserDirectGroups", ctx, orgID, userID).Return([]*models.Group{ceoGroup}, nil)
	mockCache.On("Set", "org:org-123:user:user-456:groups", []*models.Group{ceoGroup}, 300).Return(nil)

	// Set up group hierarchy relationships
	mockGroupRepo.On("GetChildren", ctx, ceoGroupID).Return([]*models.Group{engineeringGroup, salesGroup, marketingGroup}, nil)
	mockGroupRepo.On("GetChildren", ctx, engineeringGroupID).Return([]*models.Group{backendTeam}, nil)
	mockGroupRepo.On("GetChildren", ctx, backendTeamID).Return([]*models.Group{}, nil)
	mockGroupRepo.On("GetChildren", ctx, salesGroupID).Return([]*models.Group{salesTeam}, nil)
	mockGroupRepo.On("GetChildren", ctx, salesTeamID).Return([]*models.Group{}, nil)
	mockGroupRepo.On("GetChildren", ctx, marketingGroupID).Return([]*models.Group{marketingTeam}, nil)
	mockGroupRepo.On("GetChildren", ctx, marketingTeamID).Return([]*models.Group{}, nil)

	// Group roles
	for groupID, roles := range groupRoles {
		mockGroupRoleRepo.On("GetByGroupID", ctx, groupID).Return(roles, nil)
	}

	// Role details
	for roleID, role := range roles {
		mockRoleRepo.On("GetByID", ctx, roleID, mock.AnythingOfType("*models.Role")).Return(role, nil)
	}

	// Cache set for effective roles
	mockCache.On("Set", "org:org-123:user:user-456:effective_roles", mock.AnythingOfType("[]*groups.EffectiveRole"), 300).Return(nil)

	// Call the method
	effectiveRoles, err := engine.CalculateEffectiveRoles(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, effectiveRoles)
	assert.Len(t, effectiveRoles, 7) // Should have all seven roles

	// Create a map for easier verification
	roleMap := make(map[string]*EffectiveRole)
	for _, role := range effectiveRoles {
		roleMap[role.Role.GetID()] = role
	}

	// Verify CEO role (direct assignment - distance 0)
	ceoEffectiveRole := roleMap["ceo-role"]
	assert.NotNil(t, ceoEffectiveRole)
	assert.Equal(t, 0, ceoEffectiveRole.Distance)
	assert.True(t, ceoEffectiveRole.IsDirectRole)

	// Verify department roles (distance 1)
	engineeringEffectiveRole := roleMap["engineering-role"]
	assert.NotNil(t, engineeringEffectiveRole)
	assert.Equal(t, 1, engineeringEffectiveRole.Distance)
	assert.False(t, engineeringEffectiveRole.IsDirectRole)

	salesEffectiveRole := roleMap["sales-role"]
	assert.NotNil(t, salesEffectiveRole)
	assert.Equal(t, 1, salesEffectiveRole.Distance)
	assert.False(t, salesEffectiveRole.IsDirectRole)

	marketingEffectiveRole := roleMap["marketing-role"]
	assert.NotNil(t, marketingEffectiveRole)
	assert.Equal(t, 1, marketingEffectiveRole.Distance)
	assert.False(t, marketingEffectiveRole.IsDirectRole)

	// Verify team roles (distance 2)
	backendEffectiveRole := roleMap["backend-role"]
	assert.NotNil(t, backendEffectiveRole)
	assert.Equal(t, 2, backendEffectiveRole.Distance)
	assert.False(t, backendEffectiveRole.IsDirectRole)

	salesTeamEffectiveRole := roleMap["sales-team-role"]
	assert.NotNil(t, salesTeamEffectiveRole)
	assert.Equal(t, 2, salesTeamEffectiveRole.Distance)
	assert.False(t, salesTeamEffectiveRole.IsDirectRole)

	marketingTeamEffectiveRole := roleMap["marketing-team-role"]
	assert.NotNil(t, marketingTeamEffectiveRole)
	assert.Equal(t, 2, marketingTeamEffectiveRole.Distance)
	assert.False(t, marketingTeamEffectiveRole.IsDirectRole)

	// Verify inheritance paths are correct
	assert.Equal(t, []string{ceoGroupID, engineeringGroupID}, engineeringEffectiveRole.InheritancePath)
	assert.Equal(t, []string{ceoGroupID, engineeringGroupID, backendTeamID}, backendEffectiveRole.InheritancePath)
	assert.Equal(t, []string{ceoGroupID, salesGroupID}, salesEffectiveRole.InheritancePath)
	assert.Equal(t, []string{ceoGroupID, salesGroupID, salesTeamID}, salesTeamEffectiveRole.InheritancePath)

	// Verify mocks were called
	mockCache.AssertExpectations(t)
	mockGroupMembershipRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockGroupRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// TestRoleInheritanceEngine_BottomUpInheritance_InactiveGroups tests that inactive groups are properly skipped
func TestRoleInheritanceEngine_BottomUpInheritance_InactiveGroups(t *testing.T) {
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

	// Create test hierarchy: Parent -> Active Child, Inactive Child
	parentGroupID := "parent-group-123"
	parentGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Parent Group",
		OrganizationID: orgID,
		IsActive:       true,
	}
	parentGroup.BaseModel.ID = parentGroupID

	activeChildID := "active-child-456"
	activeChild := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Active Child",
		OrganizationID: orgID,
		ParentID:       &parentGroupID,
		IsActive:       true,
	}
	activeChild.BaseModel.ID = activeChildID

	inactiveChildID := "inactive-child-789"
	inactiveChild := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Inactive Child",
		OrganizationID: orgID,
		ParentID:       &parentGroupID,
		IsActive:       false, // This group is inactive
	}
	inactiveChild.BaseModel.ID = inactiveChildID

	// Create roles
	parentRoleID := "parent-role-111"
	parentRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Parent Role",
		Description: "Parent permissions",
		IsActive:    true,
	}
	parentRole.BaseModel.ID = parentRoleID

	activeChildRoleID := "active-child-role-222"
	activeChildRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Active Child Role",
		Description: "Active child permissions",
		IsActive:    true,
	}
	activeChildRole.BaseModel.ID = activeChildRoleID

	inactiveChildRoleID := "inactive-child-role-333"
	// Note: inactiveChildRole is intentionally not used since inactive groups should be skipped
	inactiveChildRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Inactive Child Role",
		Description: "Inactive child permissions",
		IsActive:    true,
	}
	inactiveChildRole.BaseModel.ID = inactiveChildRoleID
	_ = inactiveChildRole // Suppress unused variable warning

	// Create group roles
	parentGroupRole := &models.GroupRole{
		GroupID:        parentGroupID,
		RoleID:         parentRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	activeChildGroupRole := &models.GroupRole{
		GroupID:        activeChildID,
		RoleID:         activeChildRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	// Note: inactiveChildGroupRole is intentionally not used since inactive groups should be skipped
	_ = &models.GroupRole{
		GroupID:        inactiveChildID,
		RoleID:         inactiveChildRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	// Set up mocks
	mockCache.On("Get", "org:org-123:user:user-456:effective_roles").Return(nil, false)
	mockCache.On("Get", "org:org-123:user:user-456:groups").Return(nil, false)

	// User is member of parent group
	mockGroupMembershipRepo.On("GetUserDirectGroups", ctx, orgID, userID).Return([]*models.Group{parentGroup}, nil)
	mockCache.On("Set", "org:org-123:user:user-456:groups", []*models.Group{parentGroup}, 300).Return(nil)

	// Parent group has both active and inactive children
	mockGroupRepo.On("GetChildren", ctx, parentGroupID).Return([]*models.Group{activeChild, inactiveChild}, nil)

	// Active child has no children
	mockGroupRepo.On("GetChildren", ctx, activeChildID).Return([]*models.Group{}, nil)

	// Group roles
	mockGroupRoleRepo.On("GetByGroupID", ctx, parentGroupID).Return([]*models.GroupRole{parentGroupRole}, nil)
	mockGroupRoleRepo.On("GetByGroupID", ctx, activeChildID).Return([]*models.GroupRole{activeChildGroupRole}, nil)
	// Note: inactive child group roles should NOT be called

	// Role details
	mockRoleRepo.On("GetByID", ctx, parentRoleID, mock.AnythingOfType("*models.Role")).Return(parentRole, nil)
	mockRoleRepo.On("GetByID", ctx, activeChildRoleID, mock.AnythingOfType("*models.Role")).Return(activeChildRole, nil)
	// Note: inactive child role should NOT be called

	// Cache set for effective roles
	mockCache.On("Set", "org:org-123:user:user-456:effective_roles", mock.AnythingOfType("[]*groups.EffectiveRole"), 300).Return(nil)

	// Call the method
	effectiveRoles, err := engine.CalculateEffectiveRoles(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, effectiveRoles)
	assert.Len(t, effectiveRoles, 2) // Should have parent role + active child role only

	// Create a map for easier verification
	roleMap := make(map[string]*EffectiveRole)
	for _, role := range effectiveRoles {
		roleMap[role.Role.GetID()] = role
	}

	// Verify parent role is present
	parentEffectiveRole := roleMap[parentRoleID]
	assert.NotNil(t, parentEffectiveRole)
	assert.Equal(t, 0, parentEffectiveRole.Distance)
	assert.True(t, parentEffectiveRole.IsDirectRole)

	// Verify active child role is present
	activeChildEffectiveRole := roleMap[activeChildRoleID]
	assert.NotNil(t, activeChildEffectiveRole)
	assert.Equal(t, 1, activeChildEffectiveRole.Distance)
	assert.False(t, activeChildEffectiveRole.IsDirectRole)

	// Verify inactive child role is NOT present
	inactiveChildEffectiveRole := roleMap[inactiveChildRoleID]
	assert.Nil(t, inactiveChildEffectiveRole)

	// Verify mocks were called (inactive child should be skipped)
	mockCache.AssertExpectations(t)
	mockGroupMembershipRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockGroupRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// TestRoleInheritanceEngine_BottomUpInheritance_MultipleDirectGroups tests inheritance when user belongs to multiple groups
func TestRoleInheritanceEngine_BottomUpInheritance_MultipleDirectGroups(t *testing.T) {
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

	// Create test scenario: User belongs to both Engineering and Sales groups
	// Engineering Group -> Backend Team
	// Sales Group -> Sales Team

	engineeringGroupID := "engineering-group-123"
	engineeringGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Engineering Group",
		OrganizationID: orgID,
		IsActive:       true,
	}
	engineeringGroup.BaseModel.ID = engineeringGroupID

	backendTeamID := "backend-team-456"
	backendTeam := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Backend Team",
		OrganizationID: orgID,
		ParentID:       &engineeringGroupID,
		IsActive:       true,
	}
	backendTeam.BaseModel.ID = backendTeamID

	salesGroupID := "sales-group-789"
	salesGroup := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Sales Group",
		OrganizationID: orgID,
		IsActive:       true,
	}
	salesGroup.BaseModel.ID = salesGroupID

	salesTeamID := "sales-team-101"
	salesTeam := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Sales Team",
		OrganizationID: orgID,
		ParentID:       &salesGroupID,
		IsActive:       true,
	}
	salesTeam.BaseModel.ID = salesTeamID

	// Create roles
	engineeringRoleID := "engineering-role-111"
	engineeringRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Engineering Role",
		Description: "Engineering permissions",
		IsActive:    true,
	}
	engineeringRole.BaseModel.ID = engineeringRoleID

	backendRoleID := "backend-role-222"
	backendRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Backend Role",
		Description: "Backend permissions",
		IsActive:    true,
	}
	backendRole.BaseModel.ID = backendRoleID

	salesRoleID := "sales-role-333"
	salesRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Sales Role",
		Description: "Sales permissions",
		IsActive:    true,
	}
	salesRole.BaseModel.ID = salesRoleID

	salesTeamRoleID := "sales-team-role-444"
	salesTeamRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Sales Team Role",
		Description: "Sales team permissions",
		IsActive:    true,
	}
	salesTeamRole.BaseModel.ID = salesTeamRoleID

	// Create group roles
	engineeringGroupRole := &models.GroupRole{
		GroupID:        engineeringGroupID,
		RoleID:         engineeringRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	backendGroupRole := &models.GroupRole{
		GroupID:        backendTeamID,
		RoleID:         backendRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	salesGroupRole := &models.GroupRole{
		GroupID:        salesGroupID,
		RoleID:         salesRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	salesTeamGroupRole := &models.GroupRole{
		GroupID:        salesTeamID,
		RoleID:         salesTeamRoleID,
		OrganizationID: orgID,
		IsActive:       true,
	}

	// Set up mocks
	mockCache.On("Get", "org:org-123:user:user-456:effective_roles").Return(nil, false)
	mockCache.On("Get", "org:org-123:user:user-456:groups").Return(nil, false)

	// User is member of both engineering and sales groups
	mockGroupMembershipRepo.On("GetUserDirectGroups", ctx, orgID, userID).Return([]*models.Group{engineeringGroup, salesGroup}, nil)
	mockCache.On("Set", "org:org-123:user:user-456:groups", []*models.Group{engineeringGroup, salesGroup}, 300).Return(nil)

	// Set up group hierarchy relationships
	mockGroupRepo.On("GetChildren", ctx, engineeringGroupID).Return([]*models.Group{backendTeam}, nil)
	mockGroupRepo.On("GetChildren", ctx, backendTeamID).Return([]*models.Group{}, nil)
	mockGroupRepo.On("GetChildren", ctx, salesGroupID).Return([]*models.Group{salesTeam}, nil)
	mockGroupRepo.On("GetChildren", ctx, salesTeamID).Return([]*models.Group{}, nil)

	// Group roles
	mockGroupRoleRepo.On("GetByGroupID", ctx, engineeringGroupID).Return([]*models.GroupRole{engineeringGroupRole}, nil)
	mockGroupRoleRepo.On("GetByGroupID", ctx, backendTeamID).Return([]*models.GroupRole{backendGroupRole}, nil)
	mockGroupRoleRepo.On("GetByGroupID", ctx, salesGroupID).Return([]*models.GroupRole{salesGroupRole}, nil)
	mockGroupRoleRepo.On("GetByGroupID", ctx, salesTeamID).Return([]*models.GroupRole{salesTeamGroupRole}, nil)

	// Role details
	mockRoleRepo.On("GetByID", ctx, engineeringRoleID, mock.AnythingOfType("*models.Role")).Return(engineeringRole, nil)
	mockRoleRepo.On("GetByID", ctx, backendRoleID, mock.AnythingOfType("*models.Role")).Return(backendRole, nil)
	mockRoleRepo.On("GetByID", ctx, salesRoleID, mock.AnythingOfType("*models.Role")).Return(salesRole, nil)
	mockRoleRepo.On("GetByID", ctx, salesTeamRoleID, mock.AnythingOfType("*models.Role")).Return(salesTeamRole, nil)

	// Cache set for effective roles
	mockCache.On("Set", "org:org-123:user:user-456:effective_roles", mock.AnythingOfType("[]*groups.EffectiveRole"), 300).Return(nil)

	// Call the method
	effectiveRoles, err := engine.CalculateEffectiveRoles(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, effectiveRoles)
	assert.Len(t, effectiveRoles, 4) // Should have all four roles

	// Create a map for easier verification
	roleMap := make(map[string]*EffectiveRole)
	for _, role := range effectiveRoles {
		roleMap[role.Role.GetID()] = role
	}

	// Verify engineering roles
	engineeringEffectiveRole := roleMap[engineeringRoleID]
	assert.NotNil(t, engineeringEffectiveRole)
	assert.Equal(t, 0, engineeringEffectiveRole.Distance) // Direct assignment
	assert.True(t, engineeringEffectiveRole.IsDirectRole)

	backendEffectiveRole := roleMap[backendRoleID]
	assert.NotNil(t, backendEffectiveRole)
	assert.Equal(t, 1, backendEffectiveRole.Distance) // Inherited from child
	assert.False(t, backendEffectiveRole.IsDirectRole)

	// Verify sales roles
	salesEffectiveRole := roleMap[salesRoleID]
	assert.NotNil(t, salesEffectiveRole)
	assert.Equal(t, 0, salesEffectiveRole.Distance) // Direct assignment
	assert.True(t, salesEffectiveRole.IsDirectRole)

	salesTeamEffectiveRole := roleMap[salesTeamRoleID]
	assert.NotNil(t, salesTeamEffectiveRole)
	assert.Equal(t, 1, salesTeamEffectiveRole.Distance) // Inherited from child
	assert.False(t, salesTeamEffectiveRole.IsDirectRole)

	// Verify inheritance paths
	assert.Equal(t, []string{engineeringGroupID}, engineeringEffectiveRole.InheritancePath)
	assert.Equal(t, []string{engineeringGroupID, backendTeamID}, backendEffectiveRole.InheritancePath)
	assert.Equal(t, []string{salesGroupID}, salesEffectiveRole.InheritancePath)
	assert.Equal(t, []string{salesGroupID, salesTeamID}, salesTeamEffectiveRole.InheritancePath)

	// Verify mocks were called
	mockCache.AssertExpectations(t)
	mockGroupMembershipRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockGroupRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// TestRoleInheritanceEngine_BottomUpInheritance_ComplexConflictResolution tests complex role conflict scenarios
func TestRoleInheritanceEngine_BottomUpInheritance_ComplexConflictResolution(t *testing.T) {
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

	// Create complex hierarchy where user belongs to multiple groups with overlapping roles:
	// User belongs to: Group A and Group B
	// Group A -> Child A1 -> Grandchild A2
	// Group B -> Child B1
	// Same role exists at multiple levels with different distances

	groupAID := "group-a-123"
	groupA := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Group A",
		OrganizationID: orgID,
		IsActive:       true,
	}
	groupA.BaseModel.ID = groupAID

	childA1ID := "child-a1-456"
	childA1 := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Child A1",
		OrganizationID: orgID,
		ParentID:       &groupAID,
		IsActive:       true,
	}
	childA1.BaseModel.ID = childA1ID

	grandchildA2ID := "grandchild-a2-789"
	grandchildA2 := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Grandchild A2",
		OrganizationID: orgID,
		ParentID:       &childA1ID,
		IsActive:       true,
	}
	grandchildA2.BaseModel.ID = grandchildA2ID

	groupBID := "group-b-101"
	groupB := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Group B",
		OrganizationID: orgID,
		IsActive:       true,
	}
	groupB.BaseModel.ID = groupBID

	childB1ID := "child-b1-202"
	childB1 := &models.Group{
		BaseModel:      &base.BaseModel{},
		Name:           "Child B1",
		OrganizationID: orgID,
		ParentID:       &groupBID,
		IsActive:       true,
	}
	childB1.BaseModel.ID = childB1ID

	// Create shared role that exists at multiple levels
	sharedRoleID := "shared-role-999"
	sharedRole := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Shared Role",
		Description: "Role that exists at multiple levels",
		IsActive:    true,
	}
	sharedRole.BaseModel.ID = sharedRoleID

	// Create unique roles
	uniqueRoleAID := "unique-role-a-888"
	uniqueRoleA := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Unique Role A",
		Description: "Role unique to Group A hierarchy",
		IsActive:    true,
	}
	uniqueRoleA.BaseModel.ID = uniqueRoleAID

	uniqueRoleBID := "unique-role-b-777"
	uniqueRoleB := &models.Role{
		BaseModel:   &base.BaseModel{},
		Name:        "Unique Role B",
		Description: "Role unique to Group B hierarchy",
		IsActive:    true,
	}
	uniqueRoleB.BaseModel.ID = uniqueRoleBID

	// Create group roles with shared role at different levels
	groupRoles := map[string][]*models.GroupRole{
		groupAID:       {{GroupID: groupAID, RoleID: sharedRoleID, OrganizationID: orgID, IsActive: true}},       // Distance 0 from Group A
		childA1ID:      {{GroupID: childA1ID, RoleID: uniqueRoleAID, OrganizationID: orgID, IsActive: true}},     // Distance 1 from Group A
		grandchildA2ID: {{GroupID: grandchildA2ID, RoleID: sharedRoleID, OrganizationID: orgID, IsActive: true}}, // Distance 2 from Group A
		groupBID:       {{GroupID: groupBID, RoleID: uniqueRoleBID, OrganizationID: orgID, IsActive: true}},      // Distance 0 from Group B
		childB1ID:      {{GroupID: childB1ID, RoleID: sharedRoleID, OrganizationID: orgID, IsActive: true}},      // Distance 1 from Group B
	}

	// Set up mocks
	mockCache.On("Get", "org:org-123:user:user-456:effective_roles").Return(nil, false)
	mockCache.On("Get", "org:org-123:user:user-456:groups").Return(nil, false)

	// User is member of both Group A and Group B
	mockGroupMembershipRepo.On("GetUserDirectGroups", ctx, orgID, userID).Return([]*models.Group{groupA, groupB}, nil)
	mockCache.On("Set", "org:org-123:user:user-456:groups", []*models.Group{groupA, groupB}, 300).Return(nil)

	// Set up group hierarchy relationships
	mockGroupRepo.On("GetChildren", ctx, groupAID).Return([]*models.Group{childA1}, nil)
	mockGroupRepo.On("GetChildren", ctx, childA1ID).Return([]*models.Group{grandchildA2}, nil)
	mockGroupRepo.On("GetChildren", ctx, grandchildA2ID).Return([]*models.Group{}, nil)
	mockGroupRepo.On("GetChildren", ctx, groupBID).Return([]*models.Group{childB1}, nil)
	mockGroupRepo.On("GetChildren", ctx, childB1ID).Return([]*models.Group{}, nil)

	// Group roles
	for groupID, roles := range groupRoles {
		mockGroupRoleRepo.On("GetByGroupID", ctx, groupID).Return(roles, nil)
	}

	// Role details
	mockRoleRepo.On("GetByID", ctx, sharedRoleID, mock.AnythingOfType("*models.Role")).Return(sharedRole, nil)
	mockRoleRepo.On("GetByID", ctx, uniqueRoleAID, mock.AnythingOfType("*models.Role")).Return(uniqueRoleA, nil)
	mockRoleRepo.On("GetByID", ctx, uniqueRoleBID, mock.AnythingOfType("*models.Role")).Return(uniqueRoleB, nil)

	// Cache set for effective roles
	mockCache.On("Set", "org:org-123:user:user-456:effective_roles", mock.AnythingOfType("[]*groups.EffectiveRole"), 300).Return(nil)

	// Call the method
	effectiveRoles, err := engine.CalculateEffectiveRoles(ctx, orgID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, effectiveRoles)
	assert.Len(t, effectiveRoles, 3) // Should have 3 unique roles (shared role should be deduplicated)

	// Create a map for easier verification
	roleMap := make(map[string]*EffectiveRole)
	for _, role := range effectiveRoles {
		roleMap[role.Role.GetID()] = role
	}

	// Verify shared role conflict resolution (should keep the one with shortest distance)
	// Group A has shared role at distance 0, Group B has it at distance 1
	// Group A should win (distance 0 vs distance 1)
	sharedEffectiveRole := roleMap[sharedRoleID]
	assert.NotNil(t, sharedEffectiveRole)
	assert.Equal(t, 0, sharedEffectiveRole.Distance)       // Should be from Group A (distance 0)
	assert.Equal(t, groupAID, sharedEffectiveRole.GroupID) // Should be from Group A
	assert.True(t, sharedEffectiveRole.IsDirectRole)
	assert.Equal(t, []string{groupAID}, sharedEffectiveRole.InheritancePath)

	// Verify unique roles are present
	uniqueAEffectiveRole := roleMap[uniqueRoleAID]
	assert.NotNil(t, uniqueAEffectiveRole)
	assert.Equal(t, 1, uniqueAEffectiveRole.Distance) // Inherited from Child A1
	assert.False(t, uniqueAEffectiveRole.IsDirectRole)

	uniqueBEffectiveRole := roleMap[uniqueRoleBID]
	assert.NotNil(t, uniqueBEffectiveRole)
	assert.Equal(t, 0, uniqueBEffectiveRole.Distance) // Direct from Group B
	assert.True(t, uniqueBEffectiveRole.IsDirectRole)

	// Verify mocks were called
	mockCache.AssertExpectations(t)
	mockGroupMembershipRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockGroupRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}
