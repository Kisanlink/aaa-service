package model

type User struct {
	Base
	Username    string     `json:"username" gorm:"unique" validate:"required,email"`
	Password    string     `json:"password" validate:"required,min=8,max=128"`
	IsValidated bool       `json:"isValidate" validate:"default:false"`
	Roles       []UserRole `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Relation to UserRole
}

type Role struct {
	Base
	Name            string           `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description     string           `json:"description" gorm:"type:text;default:null"`
	RolePermissions []RolePermission `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Permission struct {
	Base
	Name              string             `gorm:"size:100;not null;unique"`
	Description       string             `gorm:"type:text"`
	PermissionOnRoles []PermissionOnRole `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type RolePermission struct {
	Base
	RoleID            string             `gorm:"type:uuid"`
	Role              *Role              `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PermissionOnRoles []PermissionOnRole `gorm:"foreignKey:UserRoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserRoles         []UserRole         `gorm:"foreignKey:RolePermissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
type PermissionOnRole struct {
	Base
	PermissionID string          `gorm:"type:uuid"`
	Permission   *Permission     `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserRoleID   string          `gorm:"type:uuid"`
	User         *RolePermission `gorm:"foreignKey:UserRoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
type UserRole struct {
	Base
	UserID           string          `gorm:"type:uuid"`
	User             User            `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RolePermissionID string          `gorm:"type:uuid"`
	RolePermID       *RolePermission `gorm:"foreignKey:RolePermissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
