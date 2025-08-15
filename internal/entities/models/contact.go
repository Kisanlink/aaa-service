package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

const (
	ContactTable     = "CONTACT"
	ContactTableSize = hash.Small
)

type Contact struct {
	*base.BaseModel
	UserID       string  `json:"user_id" gorm:"type:varchar(255);not null;index"`
	MobileNumber uint64  `json:"mobile_number" gorm:"type:bigint;not null"`
	CountryCode  *string `json:"country_code" gorm:"type:varchar(10);default:'+91'"`
	EmailHash    *string `json:"email_hash" gorm:"type:varchar(255)"`
	ShareCode    *string `json:"share_code" gorm:"type:varchar(50)"`
	AddressID    *string `json:"address_id" gorm:"type:varchar(255)"`

	// Relationships
	Address Address `json:"address" gorm:"foreignKey:AddressID;references:ID"`
}

func NewContact(userID string, mobileNumber uint64) *Contact {
	return &Contact{
		BaseModel:    base.NewBaseModel(ContactTable, ContactTableSize),
		UserID:       userID,
		MobileNumber: mobileNumber,
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

// Explicit method implementations to satisfy linter
func (c *Contact) GetID() string   { return c.BaseModel.GetID() }
func (c *Contact) SetID(id string) { c.BaseModel.SetID(id) }
