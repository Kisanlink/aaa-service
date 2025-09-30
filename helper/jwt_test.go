package helper

import (
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhancedJWTGeneration(t *testing.T) {
	// Create test user roles with organization context
	orgID := "org_123"
	groupID := "group_456"

	userRoles := []models.UserRole{
		{
			BaseModel: &base.BaseModel{},
			UserID:    "user_123",
			RoleID:    "role_admin",
			IsActive:  true,
			Role: models.Role{
				BaseModel:      &base.BaseModel{},
				Name:           "admin",
				Description:    "Administrator role",
				Scope:          models.RoleScopeOrg,
				OrganizationID: &orgID,
				GroupID:        &groupID,
				IsActive:       true,
				Permissions: []models.Permission{
					{
						BaseModel:   &base.BaseModel{},
						Name:        "user:read",
						Description: "Read user data",
						IsActive:    true,
					},
					{
						BaseModel:   &base.BaseModel{},
						Name:        "user:write",
						Description: "Write user data",
						IsActive:    true,
					},
				},
			},
		},
	}

	// Set IDs for the models
	userRoles[0].SetID("user_role_123")
	userRoles[0].Role.SetID("role_admin")
	userRoles[0].Role.Permissions[0].SetID("perm_read")
	userRoles[0].Role.Permissions[1].SetID("perm_write")

	t.Run("Generate Access Token with Enhanced Context", func(t *testing.T) {
		token, err := GenerateAccessToken("user_123", userRoles, "john_doe", true)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate the token
		tokenContext, err := ValidateTokenWithContext(token)
		require.NoError(t, err)

		// Verify basic claims
		assert.Equal(t, "user_123", tokenContext.UserID)
		assert.Equal(t, "access", tokenContext.TokenType)
		assert.Equal(t, "2.0", tokenContext.TokenVersion)
		assert.NotEmpty(t, tokenContext.SessionID)
		assert.NotEmpty(t, tokenContext.JTI)

		// Verify user context
		require.NotNil(t, tokenContext.UserContext)
		assert.Equal(t, "user_123", tokenContext.UserContext.ID)
		assert.Equal(t, "john_doe", *tokenContext.UserContext.Username)
		assert.True(t, tokenContext.UserContext.IsValidated)

		// Verify roles
		require.Len(t, tokenContext.UserContext.Roles, 1)
		role := tokenContext.UserContext.Roles[0]
		assert.Equal(t, "role_admin", role.ID)
		assert.Equal(t, "admin", role.Name)
		assert.Equal(t, "ORG", role.Scope)
		assert.Equal(t, &orgID, role.OrganizationID)
		assert.Equal(t, &groupID, role.GroupID)
		assert.True(t, role.IsActive)

		// Verify permissions
		assert.Contains(t, tokenContext.Permissions, "user:read")
		assert.Contains(t, tokenContext.Permissions, "user:write")

		// Verify scopes
		assert.Contains(t, tokenContext.Scopes, "role:admin")
		assert.Contains(t, tokenContext.Scopes, "org:org_123")
		assert.Contains(t, tokenContext.Scopes, "group:group_456")
		assert.Contains(t, tokenContext.Scopes, "scope:ORG")
	})

	t.Run("Generate Refresh Token", func(t *testing.T) {
		token, err := GenerateRefreshToken("user_123", userRoles, "john_doe", true)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate the token
		tokenContext, err := ValidateTokenWithContext(token)
		require.NoError(t, err)

		// Verify basic claims
		assert.Equal(t, "user_123", tokenContext.UserID)
		assert.Equal(t, "refresh", tokenContext.TokenType)
		assert.Equal(t, "2.0", tokenContext.TokenVersion)
		assert.NotEmpty(t, tokenContext.SessionID)
		assert.NotEmpty(t, tokenContext.JTI)

		// Refresh tokens should have minimal context
		// (UserContext might be nil or minimal for security)
	})

	t.Run("Backward Compatibility", func(t *testing.T) {
		token, err := GenerateAccessToken("user_123", userRoles, "john_doe", true)
		require.NoError(t, err)

		// Old validation method should still work
		userID, err := ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user_123", userID)
	})
}

func TestJWTUtilityFunctions(t *testing.T) {
	// Create test token context
	orgID := "org_123"
	groupID := "group_456"

	tokenContext := &TokenContext{
		UserID:       "user_123",
		TokenType:    "access",
		TokenVersion: "2.0",
		SessionID:    "session_123",
		JTI:          "jti_123",
		IssuedAt:     time.Now().Add(-time.Hour),
		ExpiresAt:    time.Now().Add(time.Hour),
		Permissions:  []string{"user:read", "user:write", "org:manage"},
		Scopes:       []string{"role:admin", "org:org_123", "group:group_456"},
		UserContext: &UserContext{
			ID:          "user_123",
			IsValidated: true,
			Roles: []RoleContext{
				{
					ID:             "role_admin",
					Name:           "admin",
					Scope:          "ORG",
					OrganizationID: &orgID,
					GroupID:        &groupID,
					IsActive:       true,
				},
			},
			Organizations: []OrganizationContext{
				{
					ID:   "org_123",
					Name: "Acme Corp",
				},
			},
			Groups: []GroupContext{
				{
					ID:             "group_456",
					Name:           "Engineering Team",
					OrganizationID: "org_123",
				},
			},
		},
	}

	t.Run("Permission Checks", func(t *testing.T) {
		assert.True(t, HasPermission(tokenContext, "user:read"))
		assert.True(t, HasPermission(tokenContext, "user:write"))
		assert.True(t, HasPermission(tokenContext, "org:manage"))
		assert.False(t, HasPermission(tokenContext, "admin:delete"))
	})

	t.Run("Scope Checks", func(t *testing.T) {
		assert.True(t, HasScope(tokenContext, "role:admin"))
		assert.True(t, HasScope(tokenContext, "org:org_123"))
		assert.True(t, HasScope(tokenContext, "group:group_456"))
		assert.False(t, HasScope(tokenContext, "role:user"))
	})

	t.Run("Organization Access", func(t *testing.T) {
		assert.True(t, HasOrganizationAccess(tokenContext, "org_123"))
		assert.False(t, HasOrganizationAccess(tokenContext, "org_456"))
	})

	t.Run("Group Access", func(t *testing.T) {
		assert.True(t, HasGroupAccess(tokenContext, "group_456"))
		assert.False(t, HasGroupAccess(tokenContext, "group_789"))
	})

	t.Run("Role Checks", func(t *testing.T) {
		assert.True(t, HasRole(tokenContext, "admin"))
		assert.False(t, HasRole(tokenContext, "user"))
		assert.True(t, HasActiveRole(tokenContext))
	})

	t.Run("Get User Organizations", func(t *testing.T) {
		orgs := GetUserOrganizations(tokenContext)
		assert.Contains(t, orgs, "org_123")
		assert.Len(t, orgs, 1)
	})

	t.Run("Get User Groups", func(t *testing.T) {
		groups := GetUserGroups(tokenContext)
		assert.Contains(t, groups, "group_456")
		assert.Len(t, groups, 1)
	})

	t.Run("Token Expiration", func(t *testing.T) {
		assert.False(t, IsTokenExpired(tokenContext))

		// Test with expired token
		expiredContext := *tokenContext
		expiredContext.ExpiresAt = time.Now().Add(-time.Hour)
		assert.True(t, IsTokenExpired(&expiredContext))
	})

	t.Run("Token Remaining Time", func(t *testing.T) {
		remaining := GetTokenRemainingTime(tokenContext)
		assert.True(t, remaining > 0)
		assert.True(t, remaining <= time.Hour)
	})
}

func TestTokenExtraction(t *testing.T) {
	// Create a test token
	userRoles := []models.UserRole{
		{
			BaseModel: &base.BaseModel{},
			UserID:    "user_123",
			RoleID:    "role_admin",
			IsActive:  true,
			Role: models.Role{
				BaseModel:   &base.BaseModel{},
				Name:        "admin",
				Description: "Administrator role",
				Scope:       models.RoleScopeGlobal,
				IsActive:    true,
			},
		},
	}

	userRoles[0].SetID("user_role_123")
	userRoles[0].Role.SetID("role_admin")

	token, err := GenerateAccessToken("user_123", userRoles, "john_doe", true)
	require.NoError(t, err)

	t.Run("Extract User ID", func(t *testing.T) {
		userID, err := ExtractUserIDFromToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user_123", userID)

		// Test with Bearer prefix
		userID, err = ExtractUserIDFromToken("Bearer " + token)
		require.NoError(t, err)
		assert.Equal(t, "user_123", userID)
	})

	t.Run("Get Token Type", func(t *testing.T) {
		tokenType, err := GetTokenType(token)
		require.NoError(t, err)
		assert.Equal(t, "access", tokenType)
	})
}
