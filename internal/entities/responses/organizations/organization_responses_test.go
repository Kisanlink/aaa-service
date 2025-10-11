package organizations

import (
	"testing"
	"time"

	groupResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/groups"
	roleResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/roles"
	userResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/users"
	"github.com/stretchr/testify/assert"
)

func TestOrganizationGroupResponse_Structure(t *testing.T) {
	now := time.Now()
	response := OrganizationGroupResponse{
		ID:             "group-123",
		Name:           "Test Group",
		Description:    "Test group description",
		OrganizationID: "org-456",
		ParentID:       stringPtr("parent-789"),
		IsActive:       true,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	assert.Equal(t, "group-123", response.ID)
	assert.Equal(t, "Test Group", response.Name)
	assert.Equal(t, "Test group description", response.Description)
	assert.Equal(t, "org-456", response.OrganizationID)
	assert.Equal(t, "parent-789", *response.ParentID)
	assert.True(t, response.IsActive)
	assert.NotNil(t, response.CreatedAt)
	assert.NotNil(t, response.UpdatedAt)
}

func TestOrganizationGroupListResponse_Structure(t *testing.T) {
	now := time.Now()
	groups := []*OrganizationGroupResponse{
		{
			ID:             "group-1",
			Name:           "Group 1",
			Description:    "First group",
			OrganizationID: "org-123",
			IsActive:       true,
			CreatedAt:      &now,
			UpdatedAt:      &now,
		},
		{
			ID:             "group-2",
			Name:           "Group 2",
			Description:    "Second group",
			OrganizationID: "org-123",
			IsActive:       true,
			CreatedAt:      &now,
			UpdatedAt:      &now,
		},
	}

	response := OrganizationGroupListResponse{
		OrganizationID: "org-123",
		Groups:         groups,
		TotalCount:     2,
		Page:           1,
		PageSize:       10,
	}

	assert.Equal(t, "org-123", response.OrganizationID)
	assert.Len(t, response.Groups, 2)
	assert.Equal(t, int64(2), response.TotalCount)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
}

func TestOrganizationGroupMemberResponse_Structure(t *testing.T) {
	now := time.Now()
	user := &userResponses.UserResponse{
		ID:          "user-123",
		PhoneNumber: "1234567890",
		CountryCode: "+1",
		IsValidated: true,
	}
	addedBy := &userResponses.UserResponse{
		ID:          "admin-456",
		PhoneNumber: "0987654321",
		CountryCode: "+1",
		IsValidated: true,
	}

	response := OrganizationGroupMemberResponse{
		ID:            "membership-789",
		GroupID:       "group-123",
		User:          user,
		PrincipalType: "user",
		StartsAt:      &now,
		EndsAt:        nil,
		IsActive:      true,
		AddedBy:       addedBy,
		CreatedAt:     &now,
	}

	assert.Equal(t, "membership-789", response.ID)
	assert.Equal(t, "group-123", response.GroupID)
	assert.Equal(t, "user", response.PrincipalType)
	assert.True(t, response.IsActive)
	assert.NotNil(t, response.User)
	assert.NotNil(t, response.AddedBy)
	assert.Equal(t, "user-123", response.User.ID)
	assert.Equal(t, "admin-456", response.AddedBy.ID)
}

func TestOrganizationGroupRoleResponse_Structure(t *testing.T) {
	now := time.Now()
	role := &groupResponses.RoleDetail{
		ID:          "role-123",
		Name:        "Test Role",
		Description: "Test role description",
		IsActive:    true,
	}
	assignedBy := &userResponses.UserResponse{
		ID:          "admin-456",
		PhoneNumber: "0987654321",
		CountryCode: "+1",
		IsValidated: true,
	}

	response := OrganizationGroupRoleResponse{
		ID:             "group-role-789",
		GroupID:        "group-123",
		RoleID:         "role-123",
		OrganizationID: "org-456",
		Role:           role,
		AssignedBy:     assignedBy,
		StartsAt:       &now,
		EndsAt:         nil,
		IsActive:       true,
		CreatedAt:      &now,
	}

	assert.Equal(t, "group-role-789", response.ID)
	assert.Equal(t, "group-123", response.GroupID)
	assert.Equal(t, "role-123", response.RoleID)
	assert.Equal(t, "org-456", response.OrganizationID)
	assert.True(t, response.IsActive)
	assert.NotNil(t, response.Role)
	assert.NotNil(t, response.AssignedBy)
	assert.Equal(t, "role-123", response.Role.ID)
	assert.Equal(t, "admin-456", response.AssignedBy.ID)
}

func TestOrganizationHierarchyGroupNode_Structure(t *testing.T) {
	now := time.Now()
	group := &OrganizationGroupResponse{
		ID:             "group-123",
		Name:           "Parent Group",
		Description:    "Parent group description",
		OrganizationID: "org-456",
		IsActive:       true,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	roles := []*groupResponses.GroupRoleDetail{
		{
			ID:             "group-role-1",
			GroupID:        "group-123",
			RoleID:         "role-1",
			OrganizationID: "org-456",
			IsActive:       true,
		},
	}

	members := []*OrganizationGroupMemberResponse{
		{
			ID:            "membership-1",
			GroupID:       "group-123",
			PrincipalType: "user",
			IsActive:      true,
		},
	}

	childNode := &OrganizationHierarchyGroupNode{
		Group: &OrganizationGroupResponse{
			ID:             "child-group-789",
			Name:           "Child Group",
			OrganizationID: "org-456",
			ParentID:       stringPtr("group-123"),
			IsActive:       true,
		},
		Roles:    []*groupResponses.GroupRoleDetail{},
		Members:  []*OrganizationGroupMemberResponse{},
		Children: []*OrganizationHierarchyGroupNode{},
	}

	node := OrganizationHierarchyGroupNode{
		Group:    group,
		Roles:    roles,
		Members:  members,
		Children: []*OrganizationHierarchyGroupNode{childNode},
	}

	assert.NotNil(t, node.Group)
	assert.Len(t, node.Roles, 1)
	assert.Len(t, node.Members, 1)
	assert.Len(t, node.Children, 1)
	assert.Equal(t, "group-123", node.Group.ID)
	assert.Equal(t, "child-group-789", node.Children[0].Group.ID)
}

func TestUserEffectiveRolesResponse_Structure(t *testing.T) {
	role := &roleResponses.RoleResponse{
		ID:          "role-123",
		Name:        "Test Role",
		Description: "Test role description",
		IsActive:    true,
	}

	sourceGroup := &OrganizationGroupResponse{
		ID:             "group-456",
		Name:           "Source Group",
		OrganizationID: "org-789",
		IsActive:       true,
	}

	effectiveRole := &EffectiveRoleResponse{
		Role:        role,
		Source:      "group_inherited",
		SourceGroup: sourceGroup,
		IsActive:    true,
	}

	response := UserEffectiveRolesResponse{
		OrganizationID: "org-789",
		UserID:         "user-123",
		EffectiveRoles: []*EffectiveRoleResponse{effectiveRole},
		TotalCount:     1,
	}

	assert.Equal(t, "org-789", response.OrganizationID)
	assert.Equal(t, "user-123", response.UserID)
	assert.Len(t, response.EffectiveRoles, 1)
	assert.Equal(t, int64(1), response.TotalCount)
	assert.Equal(t, "group_inherited", response.EffectiveRoles[0].Source)
	assert.NotNil(t, response.EffectiveRoles[0].SourceGroup)
	assert.Equal(t, "group-456", response.EffectiveRoles[0].SourceGroup.ID)
}

func TestUserOrganizationGroupsResponse_Structure(t *testing.T) {
	groupMembership := &UserGroupMembershipResponse{
		GroupID:       "group-123",
		GroupName:     "Test Group",
		GroupPath:     "/org/department/team",
		PrincipalType: "user",
		IsActive:      true,
		IsDirect:      true,
		Source:        "direct",
	}

	response := UserOrganizationGroupsResponse{
		OrganizationID: "org-456",
		UserID:         "user-789",
		Groups:         []*UserGroupMembershipResponse{groupMembership},
		TotalCount:     1,
	}

	assert.Equal(t, "org-456", response.OrganizationID)
	assert.Equal(t, "user-789", response.UserID)
	assert.Len(t, response.Groups, 1)
	assert.Equal(t, int64(1), response.TotalCount)
	assert.Equal(t, "group-123", response.Groups[0].GroupID)
	assert.Equal(t, "direct", response.Groups[0].Source)
	assert.True(t, response.Groups[0].IsDirect)
}

func TestOrganizationCompleteHierarchyResponse_Structure(t *testing.T) {
	now := time.Now()
	org := &OrganizationResponse{
		ID:          "org-123",
		Name:        "Test Organization",
		Description: "Test organization description",
		IsActive:    true,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	stats := &OrganizationStatsResponse{
		OrganizationID: "org-123",
		ChildCount:     2,
		GroupCount:     5,
		UserCount:      25,
	}

	groupNode := &OrganizationHierarchyGroupNode{
		Group: &OrganizationGroupResponse{
			ID:             "group-456",
			Name:           "Root Group",
			OrganizationID: "org-123",
			IsActive:       true,
		},
		Roles:    []*groupResponses.GroupRoleDetail{},
		Members:  []*OrganizationGroupMemberResponse{},
		Children: []*OrganizationHierarchyGroupNode{},
	}

	response := OrganizationCompleteHierarchyResponse{
		Organization: org,
		Parents:      []*OrganizationResponse{},
		Children:     []*OrganizationResponse{},
		Groups:       []*OrganizationHierarchyGroupNode{groupNode},
		Stats:        stats,
	}

	assert.NotNil(t, response.Organization)
	assert.NotNil(t, response.Stats)
	assert.Len(t, response.Groups, 1)
	assert.Equal(t, "org-123", response.Organization.ID)
	assert.Equal(t, int64(25), response.Stats.UserCount)
	assert.Equal(t, "group-456", response.Groups[0].Group.ID)
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
