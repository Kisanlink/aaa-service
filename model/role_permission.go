package model

import "github.com/lib/pq"

type Permission struct {
	Base
	RoleID   string         `json:"roleId" gorm:"index"`
	Resource string         `json:"resource" gorm:"size:100;not null"`
	Actions  pq.StringArray `json:"actions" gorm:"type:text[]"`
	// Actions []string `json:"actions" gorm:"type:text[]"`
}
type Role struct {
	Base
	Name        string       `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description string       `json:"description" gorm:"type:text;default:null"`
	Source      string       `json:"source" gorm:"type:text;default:null"`
	Roles       []UserRole   `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Permissions []Permission `json:"permissions" gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Resource struct {
	Base
	Name string `json:"name" gorm:"size:100;not null"`
}

type Action struct {
	Base
	Name string `json:"name" gorm:"size:100;not null"`
}
