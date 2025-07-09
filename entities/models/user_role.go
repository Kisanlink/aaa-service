package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// UserRole represents the relationship between users and roles
type UserRole struct {
	*base.BaseModel
	UserID   string `gorm:"type:varchar(255);not null"`
	RoleID   string `gorm:"type:varchar(255);not null"`
	IsActive bool   `json:"is_active" gorm:"default:true"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Role Role `json:"role" gorm:"foreignKey:RoleID;references:ID"`
}

// NewUserRole creates a new UserRole instance linking a user to a role
func NewUserRole(userID, roleID string) *UserRole {
	return &UserRole{
		BaseModel: base.NewBaseModel("ur", hash.Small),
		UserID:    userID,
		RoleID:    roleID,
		IsActive:  true,
	}
}

// BeforeCreate is called before creating a new user role
func (ur *UserRole) BeforeCreate() error {
	return ur.BaseModel.BeforeCreate()
}

// BeforeUpdate is called before updating a user role
func (ur *UserRole) BeforeUpdate() error {
	return ur.BaseModel.BeforeUpdate()
}

// BeforeDelete is called before deleting a user role
func (ur *UserRole) BeforeDelete() error {
	return ur.BaseModel.BeforeDelete()
}

// BeforeSoftDelete is called before soft deleting a user role
func (ur *UserRole) BeforeSoftDelete() error {
	return ur.BaseModel.BeforeSoftDelete()
}

// GetTableIdentifier returns the table identifier for UserRole
func (ur *UserRole) GetTableIdentifier() string {
	return "usr_rol"
}

// GetTableSize returns the table size for UserRole
func (ur *UserRole) GetTableSize() hash.TableSize { return hash.Small }
