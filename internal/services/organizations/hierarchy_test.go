package organizations

import (
	"testing"

	groupResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/groups"
	organizationResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/organizations"
	"github.com/stretchr/testify/assert"
)

func TestGroupHierarchyNode_Structure(t *testing.T) {
	// Test that the GroupHierarchyNode structure is correctly defined
	parentGroupID := "group-parent"
	childGroupID := "group-child"
	orgID := "org-123"

	// Create test group responses
	parentGroupResponse := &groupResponses.GroupResponse{
		ID:             parentGroupID,
		Name:           "Parent Group",
		Description:    "Parent group description",
		OrganizationID: orgID,
		ParentID:       nil,
		IsActive:       true,
	}

	childGroupResponse := &groupResponses.GroupResponse{
		ID:             childGroupID,
		Name:           "Child Group",
		Description:    "Child group description",
		OrganizationID: orgID,
		ParentID:       &parentGroupID,
		IsActive:       true,
	}

	// Create test role details
	testRoleDetails := []*groupResponses.GroupRoleDetail{
		{
			ID:             "role-assignment-1",
			GroupID:        parentGroupID,
			RoleID:         "role-1",
			OrganizationID: orgID,
			IsActive:       true,
			Role: groupResponses.RoleDetail{
				ID:          "role-1",
				Name:        "Admin Role",
				Description: "Administrator role",
				IsActive:    true,
			},
		},
	}

	// Create hierarchy nodes
	childNode := &organizationResponses.GroupHierarchyNode{
		Group:    childGroupResponse,
		Roles:    []*groupResponses.GroupRoleDetail{},
		Children: []*organizationResponses.GroupHierarchyNode{},
	}

	parentNode := &organizationResponses.GroupHierarchyNode{
		Group:    parentGroupResponse,
		Roles:    testRoleDetails,
		Children: []*organizationResponses.GroupHierarchyNode{childNode},
	}

	// Test parent node
	assert.Equal(t, parentGroupID, parentNode.Group.ID)
	assert.Equal(t, "Parent Group", parentNode.Group.Name)
	assert.Len(t, parentNode.Roles, 1)
	assert.Equal(t, "Admin Role", parentNode.Roles[0].Role.Name)
	assert.Len(t, parentNode.Children, 1)

	// Test child node
	assert.Equal(t, childGroupID, parentNode.Children[0].Group.ID)
	assert.Equal(t, "Child Group", parentNode.Children[0].Group.Name)
	assert.Len(t, parentNode.Children[0].Roles, 0)
	assert.Len(t, parentNode.Children[0].Children, 0)
}

func TestOrganizationHierarchyResponse_WithGroups(t *testing.T) {
	// Test that the enhanced OrganizationHierarchyResponse includes groups
	orgID := "org-123"

	orgResponse := &organizationResponses.OrganizationResponse{
		ID:          orgID,
		Name:        "Test Organization",
		Description: "Test Description",
		IsActive:    true,
	}

	groupNode := &organizationResponses.GroupHierarchyNode{
		Group: &groupResponses.GroupResponse{
			ID:             "group-1",
			Name:           "Test Group",
			OrganizationID: orgID,
			IsActive:       true,
		},
		Roles:    []*groupResponses.GroupRoleDetail{},
		Children: []*organizationResponses.GroupHierarchyNode{},
	}

	hierarchyResponse := &organizationResponses.OrganizationHierarchyResponse{
		Organization: orgResponse,
		Parents:      []*organizationResponses.OrganizationResponse{},
		Children:     []*organizationResponses.OrganizationResponse{},
		Groups:       []*organizationResponses.GroupHierarchyNode{groupNode},
	}

	// Assertions
	assert.Equal(t, orgID, hierarchyResponse.Organization.ID)
	assert.Equal(t, "Test Organization", hierarchyResponse.Organization.Name)
	assert.Len(t, hierarchyResponse.Groups, 1)
	assert.Equal(t, "group-1", hierarchyResponse.Groups[0].Group.ID)
	assert.Equal(t, "Test Group", hierarchyResponse.Groups[0].Group.Name)
}

func TestGroupHierarchy_MultiLevel(t *testing.T) {
	// Test a 3-level hierarchy: Root -> Level1 -> Level2
	orgID := "org-123"

	// Create group responses
	rootGroup := &groupResponses.GroupResponse{
		ID:             "group-root",
		Name:           "Root Group",
		OrganizationID: orgID,
		ParentID:       nil,
		IsActive:       true,
	}

	level1Group := &groupResponses.GroupResponse{
		ID:             "group-level1",
		Name:           "Level 1 Group",
		OrganizationID: orgID,
		ParentID:       stringPtr("group-root"),
		IsActive:       true,
	}

	level2Group := &groupResponses.GroupResponse{
		ID:             "group-level2",
		Name:           "Level 2 Group",
		OrganizationID: orgID,
		ParentID:       stringPtr("group-level1"),
		IsActive:       true,
	}

	// Build hierarchy
	level2Node := &organizationResponses.GroupHierarchyNode{
		Group:    level2Group,
		Roles:    []*groupResponses.GroupRoleDetail{},
		Children: []*organizationResponses.GroupHierarchyNode{},
	}

	level1Node := &organizationResponses.GroupHierarchyNode{
		Group:    level1Group,
		Roles:    []*groupResponses.GroupRoleDetail{},
		Children: []*organizationResponses.GroupHierarchyNode{level2Node},
	}

	rootNode := &organizationResponses.GroupHierarchyNode{
		Group:    rootGroup,
		Roles:    []*groupResponses.GroupRoleDetail{},
		Children: []*organizationResponses.GroupHierarchyNode{level1Node},
	}

	// Test hierarchy structure
	assert.Equal(t, "group-root", rootNode.Group.ID)
	assert.Len(t, rootNode.Children, 1)

	assert.Equal(t, "group-level1", rootNode.Children[0].Group.ID)
	assert.Len(t, rootNode.Children[0].Children, 1)

	assert.Equal(t, "group-level2", rootNode.Children[0].Children[0].Group.ID)
	assert.Len(t, rootNode.Children[0].Children[0].Children, 0)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
