package models

import (
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

const (
	ContactTable     = "CONTACT"
	ContactTableSize = hash.Small

	// Contact types
	ContactTypeMobile   = "mobile"
	ContactTypeEmail    = "email"
	ContactTypePhone    = "phone"
	ContactTypeFax      = "fax"
	ContactTypeSkype    = "skype"
	ContactTypeWhatsApp = "whatsapp"
	ContactTypeTelegram = "telegram"
	ContactTypeOther    = "other"
)

type Contact struct {
	*base.BaseModel
	UserID      string  `json:"user_id" gorm:"type:varchar(255);not null;index"`
	Type        string  `json:"type" gorm:"type:varchar(50);not null;default:'mobile'"` // mobile, email, phone, fax, etc.
	Value       string  `json:"value" gorm:"type:varchar(255);not null"`                // The actual contact value
	Description *string `json:"description" gorm:"type:text"`                           // Optional description
	IsPrimary   bool    `json:"is_primary" gorm:"default:false"`                        // Whether this is the primary contact
	IsActive    bool    `json:"is_active" gorm:"default:true"`                          // Whether this contact is active
	IsVerified  bool    `json:"is_verified" gorm:"default:false"`                       // Whether this contact has been verified
	VerifiedAt  *string `json:"verified_at" gorm:"type:varchar(50)"`                    // When the contact was verified
	VerifiedBy  *string `json:"verified_by" gorm:"type:varchar(255)"`                   // Who verified the contact

	// Legacy fields for backward compatibility
	MobileNumber uint64  `json:"mobile_number" gorm:"type:bigint"`                   // Legacy mobile number field
	CountryCode  *string `json:"country_code" gorm:"type:varchar(10);default:'+91'"` // Legacy country code
	EmailHash    *string `json:"email_hash" gorm:"type:varchar(255)"`                // Legacy email hash
	ShareCode    *string `json:"share_code" gorm:"type:varchar(50)"`                 // Legacy share code
	AddressID    *string `json:"address_id" gorm:"type:varchar(255)"`                // Legacy address ID

	// Relationships
	Address Address `json:"address" gorm:"foreignKey:AddressID;references:ID"`
}

// NewContact creates a new contact with basic information
func NewContact(userID string, contactType string, value string) *Contact {
	return &Contact{
		BaseModel:  base.NewBaseModel(ContactTable, ContactTableSize),
		UserID:     userID,
		Type:       contactType,
		Value:      value,
		IsPrimary:  false,
		IsActive:   true,
		IsVerified: false,
	}
}

// NewMobileContact creates a new mobile contact (legacy support)
func NewMobileContact(userID string, mobileNumber uint64) *Contact {
	return &Contact{
		BaseModel:    base.NewBaseModel(ContactTable, ContactTableSize),
		UserID:       userID,
		Type:         ContactTypeMobile,
		Value:        fmt.Sprintf("%d", mobileNumber),
		MobileNumber: mobileNumber,
		IsPrimary:    false,
		IsActive:     true,
		IsVerified:   false,
	}
}

// NewEmailContact creates a new email contact
func NewEmailContact(userID string, email string) *Contact {
	return &Contact{
		BaseModel:  base.NewBaseModel(ContactTable, ContactTableSize),
		UserID:     userID,
		Type:       ContactTypeEmail,
		Value:      email,
		IsPrimary:  false,
		IsActive:   true,
		IsVerified: false,
	}
}

// NewPhoneContact creates a new phone contact
func NewPhoneContact(userID string, phoneNumber string) *Contact {
	return &Contact{
		BaseModel:  base.NewBaseModel(ContactTable, ContactTableSize),
		UserID:     userID,
		Type:       ContactTypePhone,
		Value:      phoneNumber,
		IsPrimary:  false,
		IsActive:   true,
		IsVerified: false,
	}
}

func (c *Contact) BeforeCreate() error     { return c.BaseModel.BeforeCreate() }
func (c *Contact) BeforeUpdate() error     { return c.BaseModel.BeforeUpdate() }
func (c *Contact) BeforeDelete() error     { return c.BaseModel.BeforeDelete() }
func (c *Contact) BeforeSoftDelete() error { return c.BaseModel.BeforeSoftDelete() }

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (c *Contact) BeforeCreateGORM(tx *gorm.DB) error {
	return c.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (c *Contact) BeforeUpdateGORM(tx *gorm.DB) error {
	return c.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (c *Contact) BeforeDeleteGORM(tx *gorm.DB) error {
	return c.BeforeDelete()
}

func (c *Contact) GetTableIdentifier() string   { return ContactTable }
func (c *Contact) GetTableSize() hash.TableSize { return ContactTableSize }

// TableName returns the GORM table name for this model
func (c *Contact) TableName() string { return "contacts" }

// Explicit method implementations to satisfy linter
func (c *Contact) GetID() string   { return c.BaseModel.GetID() }
func (c *Contact) SetID(id string) { c.BaseModel.SetID(id) }

// SetAsPrimary sets this contact as the primary contact for the user
func (c *Contact) SetAsPrimary() {
	c.IsPrimary = true
}

// SetAsSecondary sets this contact as a secondary contact
func (c *Contact) SetAsSecondary() {
	c.IsPrimary = false
}

// Verify marks this contact as verified
func (c *Contact) Verify(verifiedBy string) {
	c.IsVerified = true
	c.VerifiedBy = &verifiedBy
	now := time.Now().Format(time.RFC3339)
	c.VerifiedAt = &now
}

// Unverify marks this contact as unverified
func (c *Contact) Unverify() {
	c.IsVerified = false
	c.VerifiedBy = nil
	c.VerifiedAt = nil
}

// Activate marks this contact as active
func (c *Contact) Activate() {
	c.IsActive = true
}

// Deactivate marks this contact as inactive
func (c *Contact) Deactivate() {
	c.IsActive = false
}

// IsMobile checks if this is a mobile contact
func (c *Contact) IsMobile() bool {
	return c.Type == "mobile"
}

// IsEmail checks if this is an email contact
func (c *Contact) IsEmail() bool {
	return c.Type == "email"
}

// IsPhone checks if this is a phone contact
func (c *Contact) IsPhone() bool {
	return c.Type == "phone"
}

// GetResourceType returns the PostgreSQL RBAC resource type for contacts
func (c *Contact) GetResourceType() string {
	return "aaa/contact"
}

// GetObjectID returns the PostgreSQL RBAC object ID for this contact
func (c *Contact) GetObjectID() string {
	return c.GetID()
}

// GetDisplayValue returns a formatted display value for the contact
func (c *Contact) GetDisplayValue() string {
	switch c.Type {
	case ContactTypeMobile:
		if c.CountryCode != nil {
			return fmt.Sprintf("%s%s", *c.CountryCode, c.Value)
		}
		return c.Value
	case ContactTypeEmail:
		return c.Value
	case ContactTypePhone:
		if c.CountryCode != nil {
			return fmt.Sprintf("%s%s", *c.CountryCode, c.Value)
		}
		return c.Value
	default:
		return c.Value
	}
}

// Validate checks if the contact data is valid
func (c *Contact) Validate() error {
	if c.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if c.Type == "" {
		return fmt.Errorf("contact type is required")
	}

	if c.Value == "" {
		return fmt.Errorf("contact value is required")
	}

	// Validate based on type
	switch c.Type {
	case ContactTypeMobile:
		if len(c.Value) < 10 || len(c.Value) > 15 {
			return fmt.Errorf("mobile number must be between 10 and 15 digits")
		}
	case ContactTypeEmail:
		if !isValidEmail(c.Value) {
			return fmt.Errorf("invalid email format")
		}
	case ContactTypePhone:
		if len(c.Value) < 7 || len(c.Value) > 15 {
			return fmt.Errorf("phone number must be between 7 and 15 digits")
		}
	}

	return nil
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	// Basic email validation - you might want to use a more robust library
	if len(email) < 5 || len(email) > 254 {
		return false
	}

	atIndex := -1
	dotIndex := -1

	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				return false // Multiple @ symbols
			}
			atIndex = i
		} else if char == '.' {
			dotIndex = i
		}
	}

	return atIndex > 0 && dotIndex > atIndex && dotIndex < len(email)-1
}

// IsValidType checks if the contact type is valid
func (c *Contact) IsValidType() bool {
	validTypes := []string{
		ContactTypeMobile,
		ContactTypeEmail,
		ContactTypePhone,
		ContactTypeFax,
		ContactTypeSkype,
		ContactTypeWhatsApp,
		ContactTypeTelegram,
		ContactTypeOther,
	}

	for _, validType := range validTypes {
		if c.Type == validType {
			return true
		}
	}

	return false
}
