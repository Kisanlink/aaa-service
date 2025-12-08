package models

import (
	"fmt"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// User represents a user in the AAA service
type User struct {
	*base.BaseModel
	PhoneNumber string  `json:"phone_number" gorm:"unique;not null;size:10;index:idx_users_phone_number;index:idx_users_phone_country_auth,priority:1;index:idx_users_phone_country_validated,priority:1;index:idx_users_search_phone" validate:"required,phone"`
	CountryCode string  `json:"country_code" gorm:"not null;size:10;default:'+91';index:idx_users_country_code;index:idx_users_phone_country_auth,priority:2;index:idx_users_phone_country_validated,priority:2" validate:"required"`
	Username    *string `json:"username" gorm:"unique;size:100;index:idx_users_username_search" validate:"omitempty,username"`
	Password    string  `json:"password" gorm:"not null;size:255" validate:"required,min=8,max=128"`
	MPin        *string `json:"mpin" gorm:"column:m_pin;size:255"`
	IsValidated bool    `json:"is_validated" gorm:"default:false;index:idx_users_is_validated;index:idx_users_phone_country_validated,priority:3"`
	// Status represents the current state of the user account
	// Possible values:
	// - "pending": Initial state when user is created but not validated
	// - "active": User is validated and can access all features
	// - "suspended": User access is temporarily suspended
	// - "blocked": User access is permanently blocked
	Status *string `json:"status" gorm:"type:varchar(50);default:'pending';index:idx_users_status"`
	Tokens int     `json:"tokens" gorm:"default:1000"`

	// Relationships
	Profile  UserProfile `json:"profile" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Contacts []Contact   `json:"contacts" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Roles    []UserRole  `json:"roles" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// NewUser creates a new User instance with default values
// - Sets phone number and password
// - Initializes with pending status
// - Sets default token balance to 1000
// - Sets validation status to false
func NewUser(phoneNumber string, countryCode string, password string) *User {
	return &User{
		BaseModel:   base.NewBaseModel("USER", hash.Medium),
		PhoneNumber: phoneNumber,
		CountryCode: countryCode,
		Password:    password,
		IsValidated: false,
		Tokens:      1000,
	}
}

// NewUserWithUsername creates a new User instance with username
func NewUserWithUsername(phoneNumber string, countryCode string, username string, password string) *User {
	user := NewUser(phoneNumber, countryCode, password)
	user.Username = &username
	return user
}

// BeforeCreate is called before creating a new user
// Sets default status to "pending" if not already set
// Validates required fields
func (u *User) BeforeCreate() error {
	if err := u.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	// Validate required fields
	if u.PhoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}
	if u.CountryCode == "" {
		return fmt.Errorf("country code is required")
	}
	if u.Password == "" {
		return fmt.Errorf("password is required")
	}

	// Set default status ONLY if not already set
	if u.Status == nil {
		status := "pending"
		u.Status = &status
	}

	return nil
}

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (u *User) BeforeCreateGORM(tx *gorm.DB) error {
	return u.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (u *User) BeforeUpdateGORM(tx *gorm.DB) error {
	return u.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (u *User) BeforeDeleteGORM(tx *gorm.DB) error {
	return u.BeforeDelete()
}

// AfterFind is called by GORM after loading a record from the database
// This initializes the embedded BaseModel pointer if it's nil
// GORM needs this because it can't populate fields into a nil embedded pointer
func (u *User) AfterFind(tx *gorm.DB) error {
	// Initialize BaseModel if it's nil
	// This happens when GORM loads a record using First() or Find()
	if u.BaseModel == nil {
		u.BaseModel = &base.BaseModel{}
	}
	return nil
}

func (u *User) BeforeUpdate() error     { return u.BaseModel.BeforeUpdate() }
func (u *User) BeforeDelete() error     { return u.BaseModel.BeforeDelete() }
func (u *User) BeforeSoftDelete() error { return u.BaseModel.BeforeSoftDelete() }

// Helper methods
// Note: GetTableIdentifier is used by the base model for ID generation
// GORM uses TableName() method for database operations
func (u *User) GetTableIdentifier() string   { return "USER" }
func (u *User) GetTableSize() hash.TableSize { return hash.Medium }

// Explicit method implementations to satisfy linter
func (u *User) GetID() string   { return u.BaseModel.GetID() }
func (u *User) SetID(id string) { u.BaseModel.SetID(id) }

// TableName specifies the table name for the User model
func (u *User) TableName() string {
	return "users"
}

// IsActive checks if the user account status is "active"
func (u *User) IsActive() bool { return u.Status != nil && *u.Status == "active" }

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
func (u *User) ValidateAadhaar() { u.IsValidated = true; status := "active"; u.Status = &status }

// GetResourceType returns the PostgreSQL RBAC resource type for users
func (u *User) GetResourceType() string {
	return "aaa/user"
}

// GetObjectID returns the PostgreSQL RBAC object ID for this user
func (u *User) GetObjectID() string {
	return u.ID
}

// SetMPin sets the user's mPin (hashed)
func (u *User) SetMPin(mPin string) {
	u.MPin = &mPin
}

// HasMPin checks if user has mPin set
func (u *User) HasMPin() bool {
	return u.MPin != nil && *u.MPin != ""
}

// GetFullPhoneNumber returns the full phone number with country code
func (u *User) GetFullPhoneNumber() string {
	return u.CountryCode + u.PhoneNumber
}
