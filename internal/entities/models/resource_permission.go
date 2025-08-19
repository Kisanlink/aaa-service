package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// ResourcePermission represents a permission granted to a role for a specific resource and action
// This replaces PostgreSQL RBAC's permission checking
type ResourcePermission struct {
	*base.BaseModel

	RoleID       string `json:"role_id" gorm:"type:varchar(255);not null;index"`
	ResourceType string `json:"resource_type" gorm:"size:100;not null;index"` // e.g., "aaa/user", "aaa/role"
	ResourceID   string `json:"resource_id" gorm:"type:varchar(255);not null;index"`
	Action       string `json:"action" gorm:"size:50;not null;index"` // e.g., "read", "write", "delete"
	IsActive     bool   `json:"is_active" gorm:"default:true"`

	// Relationships
	Role     *Role     `json:"role" gorm:"foreignKey:RoleID;references:ID"`
	Resource *Resource `json:"resource" gorm:"foreignKey:ResourceID;references:ID"`
}

// NewResourcePermission creates a new ResourcePermission instance
func NewResourcePermission(resourceID, resourceType, roleID, actionID string) *ResourcePermission {
	return &ResourcePermission{
		BaseModel:    base.NewBaseModel("RSP", hash.Small),
		ResourceID:   resourceID,
		ResourceType: resourceType,
		RoleID:       roleID,
		Action:       actionID,
		IsActive:     true,
	}
}

// BeforeCreate is called before creating a new resource permission
func (rp *ResourcePermission) BeforeCreate() error {
	if rp.BaseModel == nil {
		rp.BaseModel = base.NewBaseModel("RSP", hash.Small)
	}
	return nil
}

// BeforeUpdate is called before updating a resource permission
func (rp *ResourcePermission) BeforeUpdate() error {
	// Base model handles this automatically
	return nil
}

// BeforeDelete is called before deleting a resource permission
func (rp *ResourcePermission) BeforeDelete() error {
	// Base model handles this automatically
	return nil
}

// BeforeSoftDelete is called before soft deleting a resource permission
func (rp *ResourcePermission) BeforeSoftDelete() error {
	// Base model handles this automatically
	return nil
}

// GORM Hooks
func (rp *ResourcePermission) BeforeCreateGORM(tx *gorm.DB) error {
	return rp.BeforeCreate()
}

func (rp *ResourcePermission) BeforeUpdateGORM(tx *gorm.DB) error {
	return rp.BeforeUpdate()
}

func (rp *ResourcePermission) BeforeDeleteGORM(tx *gorm.DB) error {
	return rp.BeforeDelete()
}

// GetTableIdentifier returns the table identifier for ResourcePermission
func (rp *ResourcePermission) GetTableIdentifier() string {
	return "RSP"
}

// GetTableSize returns the table size for ResourcePermission
func (rp *ResourcePermission) GetTableSize() hash.TableSize {
	return hash.Small
}

// TableName returns the GORM table name for this model
func (rp *ResourcePermission) TableName() string { return "resource_permissions" }

// GetResourceType returns the PostgreSQL RBAC resource type for resource permissions
func (rp *ResourcePermission) GetResourceType() string {
	return "aaa/resource_permission"
}

// GetObjectID returns the PostgreSQL RBAC object ID for this resource permission
func (rp *ResourcePermission) GetObjectID() string {
	return rp.GetID()
}

// Explicit method implementations to satisfy linter
func (rp *ResourcePermission) GetID() string   { return rp.BaseModel.GetID() }
func (rp *ResourcePermission) SetID(id string) { rp.BaseModel.SetID(id) }
