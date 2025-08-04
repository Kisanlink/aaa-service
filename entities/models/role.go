package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Role represents a role in the AAA service
// This aligns with SpiceDB schema supporting hierarchical roles
type Role struct {
	*base.BaseModel
	Name        string  `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description string  `json:"description" gorm:"type:text;default:null"`
	ParentID    *string `json:"parent_id" gorm:"type:varchar(255);default:null"` // For hierarchical roles
	IsActive    bool    `json:"is_active" gorm:"default:true"`

	// Relationships
	Parent      *Role        `json:"parent" gorm:"foreignKey:ParentID;references:ID"`
	Children    []Role       `json:"children" gorm:"foreignKey:ParentID;references:ID"`
	Permissions []Permission `gorm:"many2many:role_permissions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

const (
	RoleTable     = "ROLE"
	RoleTableSize = hash.Small
)

// NewRole creates a new Role instance with specified name and description
func NewRole(name, description string) *Role {
	return &Role{
		BaseModel:   base.NewBaseModel("role", hash.Small),
		Name:        name,
		Description: description,
		IsActive:    true,
	}
}

// NewRoleWithParent creates a new Role instance with a parent role
func NewRoleWithParent(name, description, parentID string) *Role {
	role := NewRole(name, description)
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

// GetTableIdentifier returns the table identifier for Role
func (r *Role) GetTableIdentifier() string {
	return "rol"
}

// GetTableSize returns the table size for Role
func (r *Role) GetTableSize() hash.TableSize { return hash.Small }

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

// GetSpiceDBResourceType returns the SpiceDB resource type for roles
func (r *Role) GetSpiceDBResourceType() string {
	return ResourceTypeRole
}

// GetSpiceDBObjectID returns the SpiceDB object ID for this role
func (r *Role) GetSpiceDBObjectID() string {
	return r.ID
}

// Permission represents a permission in the AAA service
// This model aligns with SpiceDB schema where permissions are relationships between roles, resources, and actions
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
		BaseModel:   base.NewBaseModel("perm", hash.Small),
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

// GetTableIdentifier returns the table identifier for Permission
func (p *Permission) GetTableIdentifier() string {
	return "perm"
}

// GetTableSize returns the table size for Permission
func (p *Permission) GetTableSize() hash.TableSize {
	return hash.Small
}
