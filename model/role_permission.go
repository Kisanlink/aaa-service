package model

type Permission struct {
	Base
	Resource string `json:"resource" gorm:"size:100;not null"`
	Effect   string `json:"effect" gorm:"type:text"`
	// Actions  pq.StringArray `json:"actions" gorm:"type:text[]"`
	Actions []string `json:"actions" gorm:"type:text[]"`
}
type Role struct {
	Base
	Name        string       `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description string       `json:"description" gorm:"type:text;default:null"`
	Permissions []Permission `gorm:"many2many:role_permissions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Resource struct {
	Base
	Name string `json:"name" gorm:"size:100;not null"`
}

type Action struct {
	Base
	Name string `json:"name" gorm:"size:100;not null"`
}

type RolePermission struct {
	Base
	RoleID       string `gorm:"primaryKey;index"`
	PermissionID string `gorm:"primaryKey;index"`
}
