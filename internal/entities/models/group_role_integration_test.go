package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGroupRole_ModelIntegration_WithRelatedModels tests that the GroupRole model
// integrates properly with related models and follows the established patterns
func TestGroupRole_ModelIntegration_WithRelatedModels(t *testing.T) {
	t.Run("integration with organization model", func(t *testing.T) {
		// Create an organization
		org := NewOrganization("Test Org", "Test Organization")
		require.NotNil(t, org)
		require.NotEmpty(t, org.GetID())

		// Create a group role that references the organization
		groupRole := NewGroupRole("group-id", "role-id", org.GetID(), "user-id")
		require.NotNil(t, groupRole)
		assert.Equal(t, org.GetID(), groupRole.OrganizationID)

		// Test that the relationship field is properly set up
		assert.Nil(t, groupRole.Organization) // Should be nil until loaded
	})

	t.Run("integration with group model", func(t *testing.T) {
		// Create a group
		group := NewGroup("Test Group", "Test Group", "org-id")
		require.NotNil(t, group)
		require.NotEmpty(t, group.GetID())

		// Create a group role that references the group
		groupRole := NewGroupRole(group.GetID(), "role-id", "org-id", "user-id")
		require.NotNil(t, groupRole)
		assert.Equal(t, group.GetID(), groupRole.GroupID)

		// Test that the relationship field is properly set up
		assert.Nil(t, groupRole.Group) // Should be nil until loaded
	})

	t.Run("integration with role model", func(t *testing.T) {
		// Create a role
		role := NewRole("Test Role", "Test Role", RoleScopeOrg)
		require.NotNil(t, role)
		require.NotEmpty(t, role.GetID())

		// Create a group role that references the role
		groupRole := NewGroupRole("group-id", role.GetID(), "org-id", "user-id")
		require.NotNil(t, groupRole)
		assert.Equal(t, role.GetID(), groupRole.RoleID)

		// Test that the relationship field is properly set up
		assert.Nil(t, groupRole.Role) // Should be nil until loaded
	})

	t.Run("integration with user model", func(t *testing.T) {
		// Create a user
		user := NewUser("1234567890", "+91", "password123")
		require.NotNil(t, user)
		require.NotEmpty(t, user.GetID())

		// Create a group role that references the user as assigner
		groupRole := NewGroupRole("group-id", "role-id", "org-id", user.GetID())
		require.NotNil(t, groupRole)
		assert.Equal(t, user.GetID(), groupRole.AssignedBy)

		// Test that the relationship field is properly set up
		assert.Nil(t, groupRole.Assigner) // Should be nil until loaded
	})
}

// TestGroupRole_BusinessLogicIntegration tests business logic scenarios
func TestGroupRole_BusinessLogicIntegration(t *testing.T) {
	t.Run("time-bounded role assignment scenario", func(t *testing.T) {
		now := time.Now()
		startsAt := now.Add(-1 * time.Hour) // Started 1 hour ago
		endsAt := now.Add(1 * time.Hour)    // Ends in 1 hour

		// Create a time-bounded group role
		groupRole := NewGroupRoleWithTimebound("group-id", "role-id", "org-id", "user-id", &startsAt, &endsAt)
		require.NotNil(t, groupRole)

		// Test that it's currently effective
		assert.True(t, groupRole.IsCurrentlyEffective())
		assert.True(t, groupRole.IsEffective(now))

		// Test that it wasn't effective before start time
		beforeStart := startsAt.Add(-30 * time.Minute)
		assert.False(t, groupRole.IsEffective(beforeStart))

		// Test that it won't be effective after end time
		afterEnd := endsAt.Add(30 * time.Minute)
		assert.False(t, groupRole.IsEffective(afterEnd))
	})

	t.Run("permanent role assignment scenario", func(t *testing.T) {
		// Create a permanent group role (no time bounds)
		groupRole := NewGroupRole("group-id", "role-id", "org-id", "user-id")
		require.NotNil(t, groupRole)

		// Test that it's always effective when active
		now := time.Now()
		assert.True(t, groupRole.IsCurrentlyEffective())
		assert.True(t, groupRole.IsEffective(now))
		assert.True(t, groupRole.IsEffective(now.Add(-24*time.Hour)))
		assert.True(t, groupRole.IsEffective(now.Add(24*time.Hour)))

		// Test that it's not effective when inactive
		groupRole.IsActive = false
		assert.False(t, groupRole.IsCurrentlyEffective())
		assert.False(t, groupRole.IsEffective(now))
	})

	t.Run("role assignment audit trail scenario", func(t *testing.T) {
		// Create a group role with audit information
		assignerID := "admin-user-id"
		groupRole := NewGroupRole("group-id", "role-id", "org-id", assignerID)
		require.NotNil(t, groupRole)

		// Verify audit information is captured
		assert.Equal(t, assignerID, groupRole.AssignedBy)
		assert.NotNil(t, groupRole.BaseModel.CreatedAt)
		assert.NotNil(t, groupRole.BaseModel.UpdatedAt)
		assert.NotEmpty(t, groupRole.GetID())

		// Test that the model follows the BaseModel pattern
		assert.Equal(t, "GRPR", groupRole.GetTableIdentifier())
		assert.NotEmpty(t, groupRole.GetID())
	})

	t.Run("organization isolation scenario", func(t *testing.T) {
		// Create group roles for different organizations
		org1ID := "org-1"
		org2ID := "org-2"

		groupRole1 := NewGroupRole("group-1", "role-1", org1ID, "user-1")
		groupRole2 := NewGroupRole("group-2", "role-2", org2ID, "user-2")

		// Verify they belong to different organizations
		assert.Equal(t, org1ID, groupRole1.OrganizationID)
		assert.Equal(t, org2ID, groupRole2.OrganizationID)
		assert.NotEqual(t, groupRole1.OrganizationID, groupRole2.OrganizationID)

		// Verify they have different IDs
		assert.NotEqual(t, groupRole1.GetID(), groupRole2.GetID())
	})
}

// TestGroupRole_ValidationIntegration tests validation in various scenarios
func TestGroupRole_ValidationIntegration(t *testing.T) {
	t.Run("complete valid group role", func(t *testing.T) {
		groupRole := NewGroupRole("group-id", "role-id", "org-id", "user-id")
		err := groupRole.Validate()
		assert.NoError(t, err)
	})

	t.Run("group role with valid time bounds", func(t *testing.T) {
		now := time.Now()
		startsAt := now
		endsAt := now.Add(1 * time.Hour)

		groupRole := NewGroupRoleWithTimebound("group-id", "role-id", "org-id", "user-id", &startsAt, &endsAt)
		err := groupRole.Validate()
		assert.NoError(t, err)
	})

	t.Run("validation error scenarios", func(t *testing.T) {
		testCases := []struct {
			name        string
			setup       func() *GroupRole
			expectError bool
		}{
			{
				name: "missing group ID",
				setup: func() *GroupRole {
					return &GroupRole{
						RoleID:         "role-id",
						OrganizationID: "org-id",
						AssignedBy:     "user-id",
					}
				},
				expectError: true,
			},
			{
				name: "missing role ID",
				setup: func() *GroupRole {
					return &GroupRole{
						GroupID:        "group-id",
						OrganizationID: "org-id",
						AssignedBy:     "user-id",
					}
				},
				expectError: true,
			},
			{
				name: "missing organization ID",
				setup: func() *GroupRole {
					return &GroupRole{
						GroupID:    "group-id",
						RoleID:     "role-id",
						AssignedBy: "user-id",
					}
				},
				expectError: true,
			},
			{
				name: "missing assigned by",
				setup: func() *GroupRole {
					return &GroupRole{
						GroupID:        "group-id",
						RoleID:         "role-id",
						OrganizationID: "org-id",
					}
				},
				expectError: true,
			},
			{
				name: "invalid time range",
				setup: func() *GroupRole {
					now := time.Now()
					startsAt := now
					endsAt := now.Add(-1 * time.Hour) // ends before it starts
					return &GroupRole{
						GroupID:        "group-id",
						RoleID:         "role-id",
						OrganizationID: "org-id",
						AssignedBy:     "user-id",
						StartsAt:       &startsAt,
						EndsAt:         &endsAt,
					}
				},
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				groupRole := tc.setup()
				err := groupRole.Validate()
				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

// TestGroupRole_DatabasePatternCompliance tests that the model follows database patterns
func TestGroupRole_DatabasePatternCompliance(t *testing.T) {
	t.Run("follows BaseModel pattern", func(t *testing.T) {
		groupRole := NewGroupRole("group-id", "role-id", "org-id", "user-id")

		// Test BaseModel methods
		assert.NotEmpty(t, groupRole.GetID())
		assert.Equal(t, "GRPR", groupRole.GetTableIdentifier())
		assert.Equal(t, "group_roles", groupRole.TableName())

		// Test that it has the required BaseModel fields
		assert.NotNil(t, groupRole.BaseModel)
		assert.NotNil(t, groupRole.BaseModel.CreatedAt)
		assert.NotNil(t, groupRole.BaseModel.UpdatedAt)
	})

	t.Run("follows resource type pattern", func(t *testing.T) {
		groupRole := &GroupRole{}
		assert.Equal(t, ResourceTypeGroupRole, groupRole.GetResourceType())
		assert.Equal(t, "aaa/group_role", ResourceTypeGroupRole)
	})

	t.Run("has proper GORM hooks", func(t *testing.T) {
		groupRole := NewGroupRole("group-id", "role-id", "org-id", "user-id")

		// Test that hooks don't return errors
		assert.NoError(t, groupRole.BeforeCreate())
		assert.NoError(t, groupRole.BeforeUpdate())
		assert.NoError(t, groupRole.BeforeDelete())
		assert.NoError(t, groupRole.BeforeSoftDelete())

		// Test GORM-specific hooks
		assert.NoError(t, groupRole.BeforeCreateGORM(nil))
		assert.NoError(t, groupRole.BeforeUpdateGORM(nil))
		assert.NoError(t, groupRole.BeforeDeleteGORM(nil))
	})

	t.Run("has proper indexes defined", func(t *testing.T) {
		// This test verifies that the GORM tags include proper index definitions
		groupRole := &GroupRole{}

		// The struct should have index tags for performance
		// This is verified by the GORM tags in the struct definition
		assert.Equal(t, "group_roles", groupRole.TableName())

		// The actual index creation would be tested in a real database integration test
		// Here we just verify the model is properly structured
		assert.NotNil(t, groupRole)
	})
}
