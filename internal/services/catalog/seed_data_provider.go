package catalog

import (
	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// SeedDataProvider provides embedded seed data for roles and permissions
type SeedDataProvider struct{}

// NewSeedDataProvider creates a new seed data provider
func NewSeedDataProvider() *SeedDataProvider {
	return &SeedDataProvider{}
}

// ResourceDefinition defines a resource to be seeded
type ResourceDefinition struct {
	Name        string
	Type        string
	Description string
}

// ActionDefinition defines an action to be seeded
type ActionDefinition struct {
	Name        string
	Description string
	Category    string
	IsStatic    bool
}

// RoleDefinition defines a role to be seeded
type RoleDefinition struct {
	Name        string
	Description string
	Scope       models.RoleScope
	Permissions []string // Format: "resource:action" or wildcard patterns
}

// GetResources returns all resources to be seeded
func (s *SeedDataProvider) GetResources() []ResourceDefinition {
	return []ResourceDefinition{
		{Name: "farmer", Type: "agriculture/farmer", Description: "Farmer resource"},
		{Name: "farm", Type: "agriculture/farm", Description: "Farm resource"},
		{Name: "cycle", Type: "agriculture/cycle", Description: "Crop cycle resource"},
		{Name: "activity", Type: "agriculture/activity", Description: "Farm activity resource"},
		{Name: "fpo", Type: "agriculture/fpo", Description: "Farmer Producer Organization resource"},
		{Name: "kisansathi", Type: "agriculture/kisansathi", Description: "Kisan Sathi resource"},
		{Name: "stage", Type: "agriculture/stage", Description: "Crop stage resource"},
		{Name: "variety", Type: "agriculture/variety", Description: "Crop variety resource"},
	}
}

// GetActions returns all actions to be seeded
func (s *SeedDataProvider) GetActions() []ActionDefinition {
	return []ActionDefinition{
		{Name: "create", Description: "Create a resource", Category: "general", IsStatic: true},
		{Name: "read", Description: "Read a resource", Category: "general", IsStatic: true},
		{Name: "update", Description: "Update a resource", Category: "general", IsStatic: true},
		{Name: "delete", Description: "Delete a resource", Category: "general", IsStatic: true},
		{Name: "list", Description: "List resources", Category: "general", IsStatic: true},
		{Name: "manage", Description: "Manage a resource", Category: "general", IsStatic: true},
		{Name: "start", Description: "Start an operation", Category: "general", IsStatic: true},
		{Name: "end", Description: "End an operation", Category: "general", IsStatic: true},
		{Name: "assign", Description: "Assign a resource", Category: "general", IsStatic: true},
	}
}

// GetRoles returns all roles to be seeded with their permissions
func (s *SeedDataProvider) GetRoles() []RoleDefinition {
	return []RoleDefinition{
		{
			Name:        "farmer",
			Description: "Basic farmer role with permissions to manage their own farms and cycles",
			Scope:       models.RoleScopeGlobal,
			Permissions: []string{
				"farmer:create", "farmer:read", "farmer:update",
				"farm:create", "farm:read", "farm:update", "farm:delete",
				"cycle:create", "cycle:read", "cycle:update", "cycle:end",
			},
		},
		{
			Name:        "kisansathi",
			Description: "Kisan Sathi role with farmer permissions plus ability to list and manage activities",
			Scope:       models.RoleScopeGlobal,
			Permissions: []string{
				// All farmer permissions
				"farmer:create", "farmer:read", "farmer:update", "farmer:list",
				"farm:create", "farm:read", "farm:update", "farm:delete", "farm:list",
				"cycle:create", "cycle:read", "cycle:update", "cycle:end", "cycle:list",
				// Additional activity permissions
				"activity:create", "activity:update", "activity:delete",
			},
		},
		{
			Name:        "CEO",
			Description: "CEO role with full FPO management and all kisansathi permissions",
			Scope:       models.RoleScopeGlobal,
			Permissions: []string{
				// All kisansathi permissions
				"farmer:create", "farmer:read", "farmer:update", "farmer:list", "farmer:manage",
				"farm:create", "farm:read", "farm:update", "farm:delete", "farm:list",
				"cycle:create", "cycle:read", "cycle:update", "cycle:end", "cycle:list",
				"activity:create", "activity:update", "activity:delete",
				// FPO management permissions
				"fpo:create", "fpo:read", "fpo:update", "fpo:manage",
				// Kisansathi management
				"kisansathi:assign", "kisansathi:manage",
			},
		},
		{
			Name:        "fpo_manager",
			Description: "FPO Manager role with kisansathi permissions plus FPO read access",
			Scope:       models.RoleScopeGlobal,
			Permissions: []string{
				// All kisansathi permissions
				"farmer:create", "farmer:read", "farmer:update", "farmer:list",
				"farm:create", "farm:read", "farm:update", "farm:delete", "farm:list",
				"cycle:create", "cycle:read", "cycle:update", "cycle:end", "cycle:list",
				"activity:create", "activity:update", "activity:delete",
				// FPO read access
				"fpo:read",
			},
		},
		{
			Name:        "admin",
			Description: "Administrator role with all permissions",
			Scope:       models.RoleScopeGlobal,
			Permissions: []string{
				"*:*", // All permissions wildcard
			},
		},
		{
			Name:        "readonly",
			Description: "Read-only role with read and list permissions on all resources",
			Scope:       models.RoleScopeGlobal,
			Permissions: []string{
				"*:read",
				"*:list",
			},
		},
	}
}
