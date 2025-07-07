package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"github.com/lib/pq"
)

// Role represents a role in the AAA service
type Role struct {
	*base.BaseModel
	Name        string       `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description string       `json:"description" gorm:"type:text;default:null"`
	Permissions []Permission `gorm:"many2many:role_permissions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// NewRole creates a new Role instance
func NewRole(name string, description string) *Role {
	return &Role{
		BaseModel:   base.NewBaseModel("rol", hash.TableSizeSmall),
		Name:        name,
		Description: description,
	}
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
func (r *Role) GetTableSize() hash.TableSize {
	return hash.TableSizeSmall
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(resource, action string) bool {
	for _, permission := range r.Permissions {
		if permission.Resource == resource {
			for _, permAction := range permission.Actions {
				if permAction == action {
					return true
				}
			}
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

// Permission represents a permission in the AAA service
type Permission struct {
	*base.BaseModel
	Resource string         `json:"resource" gorm:"size:100;not null"`
	Effect   string         `json:"effect" gorm:"type:text"`
	Actions  pq.StringArray `json:"actions" gorm:"type:text[]"`
}

// NewPermission creates a new Permission instance
func NewPermission(resource string, effect string, actions []string) *Permission {
	return &Permission{
		BaseModel: base.NewBaseModel("perm", hash.TableSizeSmall),
		Resource:  resource,
		Effect:    effect,
		Actions:   actions,
	}
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
	return hash.TableSizeSmall
}

// HasAction checks if the permission has a specific action
func (p *Permission) HasAction(action string) bool {
	for _, permAction := range p.Actions {
		if permAction == action {
			return true
		}
	}
	return false
}

// AddAction adds an action to the permission
func (p *Permission) AddAction(action string) {
	if !p.HasAction(action) {
		p.Actions = append(p.Actions, action)
	}
}

// RemoveAction removes an action from the permission
func (p *Permission) RemoveAction(action string) {
	for i, permAction := range p.Actions {
		if permAction == action {
			p.Actions = append(p.Actions[:i], p.Actions[i+1:]...)
			break
		}
	}
}
