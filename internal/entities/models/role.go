package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// RoleScope defines the scope of a role
type RoleScope string

const (
	RoleScopeGlobal RoleScope = "GLOBAL" // Role applies across all organizations
	RoleScopeOrg    RoleScope = "ORG"    // Role is scoped to a specific organization
)

// Role represents a role in the AAA service with scope and versioning support
// This aligns with PostgreSQL RBAC schema supporting hierarchical roles
type Role struct {
	*base.BaseModel
	Name        string    `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description string    `json:"description" gorm:"type:text"`
	Scope       RoleScope `json:"scope" gorm:"size:20;not null;index"` // global, org, group
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Version     int       `json:"version" gorm:"default:1"`   // For role versioning
	Metadata    *string   `json:"metadata" gorm:"type:jsonb"` // Additional role metadata

	// Relationships
	OrganizationID *string      `json:"organization_id" gorm:"type:varchar(255);index"`
	GroupID        *string      `json:"group_id" gorm:"type:varchar(255);index"`
	ParentID       *string      `json:"parent_id" gorm:"type:varchar(255);index"` // For role hierarchy
	Children       []Role       `json:"children" gorm:"foreignKey:ParentID;references:ID"`
	Users          []UserRole   `json:"users" gorm:"foreignKey:RoleID;references:ID"`
	Permissions    []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

const (
	RoleTable     = "ROLE"
	RoleTableSize = hash.Small
)

// NewRole creates a new Role instance with specified name and description
func NewRole(name, description string, scope RoleScope) *Role {
	return &Role{
		BaseModel:   base.NewBaseModel("ROLE", hash.Medium),
		Name:        name,
		Description: description,
		Scope:       scope,
		Version:     1,
		IsActive:    true,
	}
}

// NewOrgRole creates a new organization-scoped Role instance
func NewOrgRole(name, description string, organizationID string) *Role {
	role := NewRole(name, description, RoleScopeOrg)
	role.OrganizationID = &organizationID
	return role
}

// NewGlobalRole creates a new globally-scoped Role instance
func NewGlobalRole(name, description string) *Role {
	return NewRole(name, description, RoleScopeGlobal)
}

// NewRoleWithParent creates a new Role instance with a parent role
func NewRoleWithParent(name, description string, scope RoleScope, parentID string) *Role {
	role := NewRole(name, description, scope)
	role.ParentID = &parentID
	return role
}

// BeforeCreate is called before creating a new role
func (r *Role) BeforeCreate() error {
	return r.BaseModel.BeforeCreate()
}

// BeforeUpdate is called before updating a role
func (r *Role) BeforeUpdate() error {
	return r.BaseModel.BeforeUpdate()
}

// BeforeDelete is called before deleting a role
func (r *Role) BeforeDelete() error {
	return r.BaseModel.BeforeDelete()
}

// BeforeSoftDelete is called before soft deleting a role
func (r *Role) BeforeSoftDelete() error {
	return r.BaseModel.BeforeSoftDelete()
}

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (r *Role) BeforeCreateGORM(tx *gorm.DB) error {
	return r.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (r *Role) BeforeUpdateGORM(tx *gorm.DB) error {
	return r.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (r *Role) BeforeDeleteGORM(tx *gorm.DB) error {
	return r.BeforeDelete()
}

// Helper methods
func (r *Role) GetTableIdentifier() string   { return "ROLE" }
func (r *Role) GetTableSize() hash.TableSize { return hash.Medium }

// TableName returns the GORM table name for this model
func (r *Role) TableName() string { return "roles" }

// Explicit method implementations to satisfy linter
func (r *Role) GetID() string   { return r.BaseModel.GetID() }
func (r *Role) SetID(id string) { r.BaseModel.SetID(id) }

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(permissionName string) bool {
	for _, permission := range r.Permissions {
		if permission.Name == permissionName {
			return true
		}
	}
	return false
}

// AddPermission adds a permission to the role
func (r *Role) AddPermission(permission *Permission) {
	r.Permissions = append(r.Permissions, *permission)
}

// RemovePermission removes a permission from the role
func (r *Role) RemovePermission(permissionID string) {
	for i, permission := range r.Permissions {
		if permission.ID == permissionID {
			r.Permissions = append(r.Permissions[:i], r.Permissions[i+1:]...)
			break
		}
	}
}

// HasParent checks if the role has a parent role
func (r *Role) HasParent() bool {
	return r.ParentID != nil
}

// GetResourceType returns the PostgreSQL RBAC resource type for roles
func (r *Role) GetResourceType() string {
	return "aaa/role"
}

// GetObjectID returns the PostgreSQL RBAC object ID for this role
func (r *Role) GetObjectID() string {
	return r.ID
}

// Permission represents a permission in the AAA service
// This model aligns with PostgreSQL RBAC schema where permissions are relationships between roles, resources, and actions
type Permission struct {
	*base.BaseModel
	Name        string  `json:"name" gorm:"size:100;not null;unique"`
	Description string  `json:"description" gorm:"type:text"`
	ResourceID  *string `json:"resource_id" gorm:"type:varchar(255);default:null"` // Which resource this permission applies to
	ActionID    *string `json:"action_id" gorm:"type:varchar(255);default:null"`   // Which action this permission allows
	IsActive    bool    `json:"is_active" gorm:"default:true"`

	// Relationships
	Resource *Resource `json:"resource" gorm:"foreignKey:ResourceID;references:ID"`
	Action   *Action   `json:"action" gorm:"foreignKey:ActionID;references:ID"`
}

// NewPermission creates a new Permission instance
func NewPermission(name, description string) *Permission {
	return &Permission{
		BaseModel:   base.NewBaseModel("PERM", hash.Small),
		Name:        name,
		Description: description,
		IsActive:    true,
	}
}

// NewPermissionWithResource creates a new Permission instance with a resource
func NewPermissionWithResource(name, description, resourceID string) *Permission {
	permission := NewPermission(name, description)
	permission.ResourceID = &resourceID
	return permission
}

// NewPermissionWithAction creates a new Permission instance with an action
func NewPermissionWithAction(name, description, actionID string) *Permission {
	permission := NewPermission(name, description)
	permission.ActionID = &actionID
	return permission
}

// NewPermissionWithResourceAndAction creates a new Permission instance with both resource and action
func NewPermissionWithResourceAndAction(name, description, resourceID, actionID string) *Permission {
	permission := NewPermission(name, description)
	permission.ResourceID = &resourceID
	permission.ActionID = &actionID
	return permission
}

// BeforeCreate is called before creating a new permission
func (p *Permission) BeforeCreate() error {
	return p.BaseModel.BeforeCreate()
}

// BeforeUpdate is called before updating a permission
func (p *Permission) BeforeUpdate() error {
	return p.BaseModel.BeforeUpdate()
}

// BeforeDelete is called before deleting a permission
func (p *Permission) BeforeDelete() error {
	return p.BaseModel.BeforeDelete()
}

// BeforeSoftDelete is called before soft deleting a permission
func (p *Permission) BeforeSoftDelete() error {
	return p.BaseModel.BeforeSoftDelete()
}

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (p *Permission) BeforeCreateGORM(tx *gorm.DB) error {
	return p.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (p *Permission) BeforeUpdateGORM(tx *gorm.DB) error {
	return p.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (p *Permission) BeforeDeleteGORM(tx *gorm.DB) error {
	return p.BeforeDelete()
}

// GetTableIdentifier returns the table identifier for Permission
func (p *Permission) GetTableIdentifier() string {
	return "PERM"
}

// GetTableSize returns the table size for Permission
func (p *Permission) GetTableSize() hash.TableSize {
	return hash.Medium
}

// TableName returns the GORM table name for this model
func (p *Permission) TableName() string { return "permissions" }

// GetResourceType returns the PostgreSQL RBAC resource type for permissions
func (p *Permission) GetResourceType() string {
	return "aaa/permission"
}

// GetObjectID returns the PostgreSQL RBAC object ID for this permission
func (p *Permission) GetObjectID() string {
	return p.GetID()
}

// Explicit method implementations to satisfy linter
func (p *Permission) GetID() string   { return p.BaseModel.GetID() }
func (p *Permission) SetID(id string) { p.BaseModel.SetID(id) }
