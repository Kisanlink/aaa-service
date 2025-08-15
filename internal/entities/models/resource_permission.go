package models

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// ResourcePermission represents a permission granted to a role for a specific resource and action
// This replaces PostgreSQL RBAC's permission checking
type ResourcePermission struct {
	// Base fields - made optional for migration compatibility
	ID        string     `json:"id" gorm:"primaryKey;type:varchar(255)"`
	CreatedAt *time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy string     `json:"created_by" gorm:"type:varchar(255)"`
	UpdatedBy string     `json:"updated_by" gorm:"type:varchar(255)"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
	DeletedBy string     `json:"deleted_by" gorm:"type:varchar(255)"`

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
	id, _ := hash.GenerateRandomID("rsp", hash.Small)
	return &ResourcePermission{
		ID:           id,
		ResourceID:   resourceID,
		ResourceType: resourceType,
		RoleID:       roleID,
		Action:       actionID,
		IsActive:     true,
	}
}

// BeforeCreate is called before creating a new resource permission
func (rp *ResourcePermission) BeforeCreate() error {
	if rp.ID == "" {
		id, _ := hash.GenerateRandomID("rsp", hash.Small)
		rp.ID = id
	}
	now := time.Now()
	if rp.CreatedAt == nil {
		rp.CreatedAt = &now
	}
	if rp.UpdatedAt == nil {
		rp.UpdatedAt = &now
	}
	return nil
}

// BeforeUpdate is called before updating a resource permission
func (rp *ResourcePermission) BeforeUpdate() error {
	now := time.Now()
	rp.UpdatedAt = &now
	return nil
}

// BeforeDelete is called before deleting a resource permission
func (rp *ResourcePermission) BeforeDelete() error {
	now := time.Now()
	rp.DeletedAt = &now
	return nil
}

// BeforeSoftDelete is called before soft deleting a resource permission
func (rp *ResourcePermission) BeforeSoftDelete() error {
	now := time.Now()
	rp.DeletedAt = &now
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

// Explicit method implementations to satisfy linter
func (rp *ResourcePermission) GetID() string   { return rp.ID }
func (rp *ResourcePermission) SetID(id string) { rp.ID = id }
