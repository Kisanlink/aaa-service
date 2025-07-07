package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
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
		BaseModel:    base.NewBaseModel("contact", hash.TableSizeSmall),
		UserID:       userID,
		MobileNumber: mobileNumber,
	}
}

func (c *Contact) BeforeCreate() error          { return c.BaseModel.BeforeCreate() }
func (c *Contact) BeforeUpdate() error          { return c.BaseModel.BeforeUpdate() }
func (c *Contact) BeforeDelete() error          { return c.BaseModel.BeforeDelete() }
func (c *Contact) BeforeSoftDelete() error      { return c.BaseModel.BeforeSoftDelete() }
func (c *Contact) GetTableIdentifier() string   { return "contact" }
func (c *Contact) GetTableSize() hash.TableSize { return hash.TableSizeSmall }
