package catalog

import "github.com/Kisanlink/aaa-service/v2/internal/entities/models"

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
