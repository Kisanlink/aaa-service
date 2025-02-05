package model

import (
	"time"
)

type User struct {
	ID           string    `gorm:"primaryKey;size:36"`
	Username     string    `gorm:"size:100;not null;unique"`
	PasswordHash string    `gorm:"size:255;not null"`
	IsValidated  bool      `gorm:"default:false"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type Role struct {
	ID          string `gorm:"primaryKey;size:36"`
	Name        string `gorm:"size:50;not null;unique"`
	Description string `gorm:"type:text"`
}

type Permission struct {
	ID          string `gorm:"primaryKey;size:36"`
	Name        string `gorm:"size:100;not null;unique"`
	Description string `gorm:"type:text"`
}

type RolePermission struct {
	RoleID       string `gorm:"primaryKey;size:36"`
	PermissionID string `gorm:"primaryKey;size:36"`
}

func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

func (User) TableName() string {
	return "users"
}
