package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// User represents a user in the AAA service
type User struct {
	*base.BaseModel
	Username    string  `json:"username" gorm:"unique;not null;size:100" validate:"required,username"`
	Password    string  `json:"password" gorm:"not null;size:255" validate:"required,min=8,max=128"`
	IsValidated bool    `json:"is_validated" gorm:"default:false"`
	// Status represents the current state of the user account
	// Possible values:
	// - "pending": Initial state when user is created but not validated
	// - "active": User is validated and can access all features
	// - "suspended": User access is temporarily suspended
	// - "blocked": User access is permanently blocked
	Status      *string `json:"status" gorm:"type:varchar(50);default:'pending'"`
	Tokens      int     `json:"tokens" gorm:"default:1000"`

	// Relationships
	Profile  UserProfile `json:"profile" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Contacts []Contact   `json:"contacts" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Roles    []UserRole  `json:"roles" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// NewUser creates a new User instance with default values
// - Sets username and password
// - Initializes with pending status
// - Sets default token balance to 1000
// - Sets validation status to false
func NewUser(username string, password string) *User {
	return &User{
		BaseModel:   base.NewBaseModel("usr", hash.TableSizeMedium),
		Username:    username,
		Password:    password,
		IsValidated: false,
		Tokens:      1000,
	}
}

// BeforeCreate is called before creating a new user
// Sets default status to "pending" if not already set
func (u *User) BeforeCreate() error {
	if err := u.BaseModel.BeforeCreate(); err != nil {
		return err
	}
	if u.Status == nil {
		status := "pending"
		u.Status = &status
	}
	return nil
}

func (u *User) BeforeUpdate() error     { return u.BaseModel.BeforeUpdate() }
func (u *User) BeforeDelete() error     { return u.BaseModel.BeforeDelete() }
func (u *User) BeforeSoftDelete() error { return u.BaseModel.BeforeSoftDelete() }

func (u *User) GetTableIdentifier() string   { return "usr" }
func (u *User) GetTableSize() hash.TableSize { return hash.TableSizeMedium }

// IsActive checks if the user account status is "active"
func (u *User) IsActive() bool                    { return u.Status != nil && *u.Status == "active" }

// HasEnoughTokens checks if user has sufficient tokens for an operation
func (u *User) HasEnoughTokens(required int) bool { return u.Tokens >= required }

// DeductTokens attempts to deduct tokens from user's balance
// Returns true if deduction was successful, false if insufficient balance
func (u *User) DeductTokens(amount int) bool {
	if u.HasEnoughTokens(amount) {
		u.Tokens -= amount
		return true
	}
	return false
}

// AddTokens increases user's token balance by the specified amount
func (u *User) AddTokens(amount int) { u.Tokens += amount }

// ValidateAadhaar marks the user as validated and sets status to "active"
// This is called after successful Aadhaar verification
func (u *User) ValidateAadhaar()     { u.IsValidated = true; status := "active"; u.Status = &status }
