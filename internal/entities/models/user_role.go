package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
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
		BaseModel: base.NewBaseModel("USRL", hash.Small),
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

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (ur *UserRole) BeforeCreateGORM(tx *gorm.DB) error {
	return ur.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (ur *UserRole) BeforeUpdateGORM(tx *gorm.DB) error {
	return ur.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (ur *UserRole) BeforeDeleteGORM(tx *gorm.DB) error {
	return ur.BeforeDelete()
}

// GetTableIdentifier returns the table identifier for UserRole
func (ur *UserRole) GetTableIdentifier() string {
	return "USR_ROL"
}

// GetTableSize returns the table size for UserRole
func (ur *UserRole) GetTableSize() hash.TableSize { return hash.Small }

// Explicit method implementations to satisfy linter
func (ur *UserRole) GetID() string   { return ur.BaseModel.GetID() }
func (ur *UserRole) SetID(id string) { ur.BaseModel.SetID(id) }
