package catalog

import (
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestExpandWildcardPatterns tests the wildcard expansion logic
func TestExpandWildcardPatterns(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	pm := &PermissionManager{
		logger: logger,
	}

	// Create test resources
	resources := []*models.Resource{
		{BaseModel: base.NewBaseModel("RES", hash.Medium), Name: "farmer"},
		{BaseModel: base.NewBaseModel("RES", hash.Medium), Name: "farm"},
		{BaseModel: base.NewBaseModel("RES", hash.Medium), Name: "cycle"},
	}

	// Create test actions
	actions := []*models.Action{
		{BaseModel: base.NewBaseModel("ACT", hash.Small), Name: "create"},
		{BaseModel: base.NewBaseModel("ACT", hash.Small), Name: "read"},
		{BaseModel: base.NewBaseModel("ACT", hash.Small), Name: "update"},
		{BaseModel: base.NewBaseModel("ACT", hash.Small), Name: "delete"},
	}

	tests := []struct {
		name            string
		patterns        []string
		expectedCount   int
		checkContains   []string
		checkNotContain []string
	}{
		{
			name:          "Exact match pattern",
			patterns:      []string{"farmer:create", "farm:read"},
			expectedCount: 2,
			checkContains: []string{"farmer:create", "farm:read"},
		},
		{
			name:          "Wildcard all actions for one resource",
			patterns:      []string{"farmer:*"},
			expectedCount: 4, // farmer with all 4 actions
			checkContains: []string{"farmer:create", "farmer:read", "farmer:update", "farmer:delete"},
		},
		{
			name:          "Wildcard all resources for one action",
			patterns:      []string{"*:read"},
			expectedCount: 3, // 3 resources with read action
			checkContains: []string{"farmer:read", "farm:read", "cycle:read"},
		},
		{
			name:          "Wildcard all resources and all actions",
			patterns:      []string{"*:*"},
			expectedCount: 12, // 3 resources × 4 actions
			checkContains: []string{"farmer:create", "farm:update", "cycle:delete"},
		},
		{
			name:            "Multiple patterns with overlap",
			patterns:        []string{"farmer:create", "farmer:*", "farm:read"},
			expectedCount:   5, // farmer:* (4) + farm:read (1), with deduplication
			checkContains:   []string{"farmer:create", "farmer:read", "farm:read"},
			checkNotContain: []string{"farm:create", "cycle:read"},
		},
		{
			name:          "Readonly pattern",
			patterns:      []string{"*:read", "*:list"},
			expectedCount: 3, // Only read action exists in test data (list doesn't exist)
			checkContains: []string{"farmer:read", "farm:read", "cycle:read"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expanded := pm.expandWildcardPatterns(tt.patterns, resources, actions)

			// Check count
			assert.Equal(t, tt.expectedCount, len(expanded), "Expected %d expanded permissions, got %d", tt.expectedCount, len(expanded))

			// Build a map of expanded permissions for easy checking
			expandedMap := make(map[string]bool)
			for _, perm := range expanded {
				key := perm.Resource.Name + ":" + perm.Action.Name
				expandedMap[key] = true
			}

			// Check that expected permissions are present
			for _, expected := range tt.checkContains {
				assert.True(t, expandedMap[expected], "Expected permission %s to be present", expected)
			}

			// Check that unexpected permissions are not present
			for _, unexpected := range tt.checkNotContain {
				assert.False(t, expandedMap[unexpected], "Expected permission %s to NOT be present", unexpected)
			}
		})
	}
}

// TestMatchesPattern tests the pattern matching logic
func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name        string
		permName    string
		pattern     string
		shouldMatch bool
	}{
		{
			name:        "Exact match",
			permName:    "farmer:create",
			pattern:     "farmer:create",
			shouldMatch: true,
		},
		{
			name:        "Wildcard all actions",
			permName:    "farmer:create",
			pattern:     "farmer:*",
			shouldMatch: true,
		},
		{
			name:        "Wildcard all resources",
			permName:    "farmer:create",
			pattern:     "*:create",
			shouldMatch: true,
		},
		{
			name:        "Wildcard all",
			permName:    "farmer:create",
			pattern:     "*:*",
			shouldMatch: true,
		},
		{
			name:        "No match - different resource",
			permName:    "farmer:create",
			pattern:     "farm:create",
			shouldMatch: false,
		},
		{
			name:        "No match - different action",
			permName:    "farmer:create",
			pattern:     "farmer:read",
			shouldMatch: false,
		},
		{
			name:        "Partial wildcard match - action",
			permName:    "farm:update",
			pattern:     "farm:*",
			shouldMatch: true,
		},
		{
			name:        "Partial wildcard match - resource",
			permName:    "cycle:delete",
			pattern:     "*:delete",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPattern(tt.permName, tt.pattern)
			assert.Equal(t, tt.shouldMatch, result, "Pattern %s should%s match %s",
				tt.pattern,
				map[bool]string{true: "", false: " not"}[tt.shouldMatch],
				tt.permName)
		})
	}
}

// TestSeedDataProviderRoles tests that seed data is properly structured
func TestSeedDataProviderRoles(t *testing.T) {
	provider := NewDefaultSeedProvider()

	// Get roles
	roles := provider.GetRoles()

	// Verify we have the expected roles
	expectedRoles := []string{"farmer", "kisansathi", "CEO", "fpo_manager", "admin", "readonly"}
	assert.Equal(t, len(expectedRoles), len(roles), "Expected %d roles", len(expectedRoles))

	roleMap := make(map[string]RoleDefinition)
	for _, role := range roles {
		roleMap[role.Name] = role
	}

	// Verify each expected role exists
	for _, expectedRole := range expectedRoles {
		role, exists := roleMap[expectedRole]
		assert.True(t, exists, "Role %s should exist", expectedRole)
		assert.NotEmpty(t, role.Description, "Role %s should have a description", expectedRole)
		assert.Equal(t, models.RoleScopeGlobal, role.Scope, "Role %s should have GLOBAL scope", expectedRole)
		assert.NotEmpty(t, role.Permissions, "Role %s should have permissions", expectedRole)
	}

	// Verify admin has wildcard permission
	admin := roleMap["admin"]
	assert.Contains(t, admin.Permissions, "*:*", "Admin role should have *:* permission")

	// Verify readonly has read and list permissions
	readonly := roleMap["readonly"]
	assert.Contains(t, readonly.Permissions, "*:read", "Readonly role should have *:read permission")
	assert.Contains(t, readonly.Permissions, "*:list", "Readonly role should have *:list permission")
}

// TestSeedDataProviderActions tests action definitions
func TestSeedDataProviderActions(t *testing.T) {
	provider := NewDefaultSeedProvider()
	actions := provider.GetActions()

	// Verify expected actions exist
	expectedActions := []string{"create", "read", "update", "delete", "list", "manage", "start", "end", "assign"}
	assert.Equal(t, len(expectedActions), len(actions), "Expected %d actions", len(expectedActions))

	actionMap := make(map[string]ActionDefinition)
	for _, action := range actions {
		actionMap[action.Name] = action
	}

	for _, expectedAction := range expectedActions {
		action, exists := actionMap[expectedAction]
		assert.True(t, exists, "Action %s should exist", expectedAction)
		assert.True(t, action.IsStatic, "Action %s should be static", expectedAction)
		assert.NotEmpty(t, action.Description, "Action %s should have a description", expectedAction)
		assert.Equal(t, "general", action.Category, "Action %s should have general category", expectedAction)
	}
}

// TestSeedDataProviderResources tests resource definitions
func TestSeedDataProviderResources(t *testing.T) {
	provider := NewDefaultSeedProvider()
	resources := provider.GetResources()

	// Verify expected resources exist
	expectedResources := []string{"farmer", "farm", "cycle", "activity", "fpo", "kisansathi", "stage", "variety"}
	assert.Equal(t, len(expectedResources), len(resources), "Expected %d resources", len(expectedResources))

	resourceMap := make(map[string]ResourceDefinition)
	for _, resource := range resources {
		resourceMap[resource.Name] = resource
	}

	for _, expectedResource := range expectedResources {
		resource, exists := resourceMap[expectedResource]
		assert.True(t, exists, "Resource %s should exist", expectedResource)
		assert.NotEmpty(t, resource.Description, "Resource %s should have a description", expectedResource)
		assert.NotEmpty(t, resource.Type, "Resource %s should have a type", expectedResource)
	}
}
