package model

type User struct {
	Base
	Username    string `json:"username" gorm:"unique" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8,max=128"`
	IsValidated bool   `json:"isValidate" validate:"default:false"`
	Roles       []Role `json:"roles" gorm:"many2many:user_roles;"`
}

type Role struct {
	Base
	Name        string       `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description string       `json:"description" gorm:"type:text;default:null"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	Users       []User       `json:"users" gorm:"many2many:user_roles;"`
}

type Permission struct {
	Base
	Name        string `gorm:"size:100;not null;unique"`
	Description string `gorm:"type:text"`
	Roles       []Role `json:"roles" gorm:"many2many:role_permissions;"`
}

type UserRole struct {
	UserID string `json:"userId" gorm:"primaryKey;size:36"`
	RoleID string `json:"roleId" gorm:"primaryKey;size:36"`
}
type RolePermission struct {
	Base
	RoleID       string `json:"roleId" gorm:"primaryKey;size:36"`
	PermissionID string `json:"permissionId" gorm:"primaryKey;size:36"`
}
