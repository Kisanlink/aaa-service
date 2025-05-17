package model

import "time"

type Role struct {
	Base
	Name            string           `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description     string           `json:"description" gorm:"type:text;default:null"`
	Source          string           `json:"source" gorm:"type:text;default:null"`
	RolePermissions []RolePermission `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Roles           []UserRole       `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Permission struct {
	Base
	Name        string `gorm:"size:100;not null;unique"`
	Description string `gorm:"type:text"`
	Action      string `json:"action" gorm:"type:text;default:null"`
	Resource    string `json:"resource" gorm:"type:text;default:null"`
	Source      string `json:"source" gorm:"type:text;default:null"`

	ValidStartTime  time.Time        `json:"valid_start_time" gorm:"column:valid_start_time"`
	ValidEndTime    time.Time        `json:"valid_end_time" gorm:"column:valid_end_time"`
	RolePermissions []RolePermission `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
type RolePermission struct {
	Base
	RoleID       string     `gorm:"type:uuid"`
	Role         Role       `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PermissionID string     `gorm:"type:uuid"`
	Permission   Permission `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	IsActive     bool       `json:"is_active" validate:"default:true"`
}
