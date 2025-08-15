package models

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// RolePermission represents the many-to-many relationship between roles and permissions
// This is the join table for role_permissions
type RolePermission struct {
	// Base fields - made optional for migration compatibility
	ID        string     `json:"id" gorm:"primaryKey;type:varchar(255)"`
	CreatedAt *time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy string     `json:"created_by" gorm:"type:varchar(255)"`
	UpdatedBy string     `json:"updated_by" gorm:"type:varchar(255)"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
	DeletedBy string     `json:"deleted_by" gorm:"type:varchar(255)"`

	RoleID       string `json:"role_id" gorm:"type:varchar(255);not null"`
	PermissionID string `json:"permission_id" gorm:"type:varchar(255);not null"`
	IsActive     bool   `json:"is_active" gorm:"default:true"`

	// Relationships
	Role       *Role       `json:"role" gorm:"foreignKey:RoleID;references:ID"`
	Permission *Permission `json:"permission" gorm:"foreignKey:PermissionID;references:ID"`
}

// NewRolePermission creates a new RolePermission instance
func NewRolePermission(roleID, permissionID string) *RolePermission {
	id, _ := hash.GenerateRandomID("rlp", hash.Small)
	return &RolePermission{
		ID:           id,
		RoleID:       roleID,
		PermissionID: permissionID,
		IsActive:     true,
	}
}

// BeforeCreate is called before creating a new role permission
func (rp *RolePermission) BeforeCreate() error {
	if rp.ID == "" {
		id, _ := hash.GenerateRandomID("rlp", hash.Small)
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

// BeforeUpdate is called before updating a role permission
func (rp *RolePermission) BeforeUpdate() error {
	now := time.Now()
	rp.UpdatedAt = &now
	return nil
}

// BeforeDelete is called before deleting a role permission
func (rp *RolePermission) BeforeDelete() error {
	now := time.Now()
	rp.DeletedAt = &now
	return nil
}

// BeforeSoftDelete is called before soft deleting a role permission
func (rp *RolePermission) BeforeSoftDelete() error {
	now := time.Now()
	rp.DeletedAt = &now
	return nil
}

// GORM Hooks
func (rp *RolePermission) BeforeCreateGORM(tx *gorm.DB) error {
	return rp.BeforeCreate()
}

func (rp *RolePermission) BeforeUpdateGORM(tx *gorm.DB) error {
	return rp.BeforeUpdate()
}

func (rp *RolePermission) BeforeDeleteGORM(tx *gorm.DB) error {
	return rp.BeforeDelete()
}

// GetTableIdentifier returns the table identifier for RolePermission
func (rp *RolePermission) GetTableIdentifier() string {
	return "RLP"
}

// GetTableSize returns the table size for RolePermission
func (rp *RolePermission) GetTableSize() hash.TableSize {
	return hash.Small
}

// Explicit method implementations to satisfy linter
func (rp *RolePermission) GetID() string   { return rp.ID }
func (rp *RolePermission) SetID(id string) { rp.ID = id }
