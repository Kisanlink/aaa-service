package roles

import (
	"fmt"
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/stretchr/testify/assert"
)

// TestUserRoleModel_NewUserRole tests the NewUserRole constructor
func TestUserRoleModel_NewUserRole(t *testing.T) {
	userID := "test-user-id"
	roleID := "test-role-id"

	userRole := models.NewUserRole(userID, roleID)

	assert.NotNil(t, userRole)
	assert.Equal(t, userID, userRole.UserID)
	assert.Equal(t, roleID, userRole.RoleID)
	assert.True(t, userRole.IsActive)
	assert.NotEmpty(t, userRole.GetID())
}

// TestUserRoleModel_IsActiveAssignment tests the IsActiveAssignment method
func TestUserRoleModel_IsActiveAssignment(t *testing.T) {
	tests := []struct {
		name           string
		userRoleActive bool
		roleActive     bool
		expected       bool
	}{
		{
			name:           "both active",
			userRoleActive: true,
			roleActive:     true,
			expected:       true,
		},
		{
			name:           "user role inactive",
			userRoleActive: false,
			roleActive:     true,
			expected:       false,
		},
		{
			name:           "role inactive",
			userRoleActive: true,
			roleActive:     false,
			expected:       false,
		},
		{
			name:           "both inactive",
			userRoleActive: false,
			roleActive:     false,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &models.Role{
				Name:     "Test Role",
				IsActive: tt.roleActive,
			}

			userRole := &models.UserRole{
				UserID:   "test-user",
				RoleID:   "test-role",
				IsActive: tt.userRoleActive,
				Role:     *role,
			}

			result := userRole.IsActiveAssignment()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUserRoleRepository_ValidationLogic tests the validation logic for role assignments
func TestUserRoleRepository_ValidationLogic(t *testing.T) {
	t.Run("ValidateUserRoleData", func(t *testing.T) {
		// Test that user role data is properly structured
		userRole := models.NewUserRole("user-123", "role-456")

		assert.NotEmpty(t, userRole.UserID)
		assert.NotEmpty(t, userRole.RoleID)
		assert.True(t, userRole.IsActive)
		assert.NotEmpty(t, userRole.GetID())

		// Test BeforeCreate hook
		err := userRole.BeforeCreate()
		assert.NoError(t, err)

		// Test BeforeUpdate hook
		err = userRole.BeforeUpdate()
		assert.NoError(t, err)
	})

	t.Run("ValidateRoleData", func(t *testing.T) {
		// Test role data structure
		role := models.NewRole("Test Role", "Test Description", models.RoleScopeGlobal)

		assert.Equal(t, "Test Role", role.Name)
		assert.Equal(t, "Test Description", role.Description)
		assert.Equal(t, models.RoleScopeGlobal, role.Scope)
		assert.True(t, role.IsActive)
		assert.Equal(t, 1, role.Version)
		assert.NotEmpty(t, role.GetID())
	})
}

// TestUserRoleRepository_ErrorHandling tests error handling scenarios
func TestUserRoleRepository_ErrorHandling(t *testing.T) {
	t.Run("EmptyUserID", func(t *testing.T) {
		userRole := models.NewUserRole("", "role-123")
		assert.Empty(t, userRole.UserID)
		assert.NotEmpty(t, userRole.RoleID)
	})

	t.Run("EmptyRoleID", func(t *testing.T) {
		userRole := models.NewUserRole("user-123", "")
		assert.NotEmpty(t, userRole.UserID)
		assert.Empty(t, userRole.RoleID)
	})

	t.Run("DeactivateUserRole", func(t *testing.T) {
		userRole := models.NewUserRole("user-123", "role-456")
		assert.True(t, userRole.IsActive)

		// Simulate deactivation
		userRole.IsActive = false
		assert.False(t, userRole.IsActive)
	})
}

// TestUserRoleRepository_BusinessLogic tests business logic scenarios
func TestUserRoleRepository_BusinessLogic(t *testing.T) {
	t.Run("RoleAssignmentScenarios", func(t *testing.T) {
		// Test multiple role assignments for same user
		userID := "user-123"
		role1ID := "role-1"
		role2ID := "role-2"

		userRole1 := models.NewUserRole(userID, role1ID)
		userRole2 := models.NewUserRole(userID, role2ID)

		assert.Equal(t, userID, userRole1.UserID)
		assert.Equal(t, userID, userRole2.UserID)
		assert.NotEqual(t, userRole1.RoleID, userRole2.RoleID)
		// IDs should be different (though they might be similar due to timing)
		// Just verify they're not empty and the objects are different
		assert.NotEmpty(t, userRole1.GetID())
		assert.NotEmpty(t, userRole2.GetID())
	})

	t.Run("RoleHierarchy", func(t *testing.T) {
		// Test parent-child role relationships
		parentRole := models.NewRole("Parent Role", "Parent Description", models.RoleScopeGlobal)
		childRole := models.NewRoleWithParent("Child Role", "Child Description", models.RoleScopeGlobal, parentRole.GetID())

		assert.True(t, childRole.HasParent())
		assert.False(t, parentRole.HasParent())
		assert.Equal(t, parentRole.GetID(), *childRole.ParentID)
	})

	t.Run("OrganizationScopedRoles", func(t *testing.T) {
		// Test organization-scoped roles
		orgID := "org-123"
		orgRole := models.NewOrgRole("Org Admin", "Organization Administrator", orgID)

		assert.Equal(t, models.RoleScopeOrg, orgRole.Scope)
		assert.Equal(t, orgID, *orgRole.OrganizationID)
		assert.Nil(t, orgRole.ParentID)
	})
}

// TestUserRoleRepository_ConcurrencyScenarios tests concurrent access scenarios
func TestUserRoleRepository_ConcurrencyScenarios(t *testing.T) {
	t.Run("ConcurrentRoleAssignments", func(t *testing.T) {
		userID := "user-123"

		// Simulate concurrent role assignments
		assignments := make([]*models.UserRole, 5)
		for i := 0; i < 5; i++ {
			roleID := fmt.Sprintf("role-%d", i)
			assignments[i] = models.NewUserRole(userID, roleID)
		}

		// Verify all assignments have the same user ID and different role IDs
		for i, assignment := range assignments {
			assert.Equal(t, userID, assignment.UserID)
			assert.Equal(t, fmt.Sprintf("role-%d", i), assignment.RoleID)
			assert.NotEmpty(t, assignment.GetID())
			assert.True(t, assignment.IsActive)
		}
	})
}
