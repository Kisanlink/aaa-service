package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// RolePermission represents the many-to-many relationship between roles and permissions
// This is the join table for role_permissions
type RolePermission struct {
	*base.BaseModel

	RoleID       string `json:"role_id" gorm:"type:varchar(255);not null"`
	PermissionID string `json:"permission_id" gorm:"type:varchar(255);not null"`
	IsActive     bool   `json:"is_active" gorm:"default:true"`

	// Relationships
	Role       *Role       `json:"role" gorm:"foreignKey:RoleID;references:ID"`
	Permission *Permission `json:"permission" gorm:"foreignKey:PermissionID;references:ID"`
}

// NewRolePermission creates a new RolePermission instance
func NewRolePermission(roleID, permissionID string) *RolePermission {
	return &RolePermission{
		BaseModel:    base.NewBaseModel("ROLPERM", hash.Small),
		RoleID:       roleID,
		PermissionID: permissionID,
		IsActive:     true,
	}
}

// GetTableIdentifier returns the table identifier for RolePermission
func (rp *RolePermission) GetTableIdentifier() string {
	return "ROLPERM"
}

// GetTableSize returns the table size for RolePermission
func (rp *RolePermission) GetTableSize() hash.TableSize {
	return hash.Small
}
