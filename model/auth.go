package model

type User struct {
	Base
	Username    string     `json:"username" gorm:"unique" validate:"required,email"`
	Password    string     `json:"password" validate:"required,min=8,max=128"`
	IsValidated bool       `json:"isValidate" validate:"default:false"`
	Roles       []UserRole `gorm:"foreignKey:UserID"`
}

type Role struct {
	Base
	Name        string     `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description string     `json:"description" gorm:"type:text;default:null"`
	Users       []UserRole `gorm:"foreignKey:RoleID"`
}

type Permission struct {
	Base
	Name           string     `gorm:"size:100;not null;unique"`
	Description    string     `gorm:"type:text"`
	UserPermission []UserRole `gorm:"foreignKey:PermissionID"`
}

type UserRole struct {
	Base
	UserID       string     `gorm:"not null"`
	User         User       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RoleID       string     `gorm:"not null"`
	Role         Role       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PermissionID string     `gorm:"not null"`
	Permission   Permission `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
