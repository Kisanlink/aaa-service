package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// Address represents a user's address
type Address struct {
	*base.BaseModel
	UserID      string  `json:"user_id" gorm:"type:varchar(255);not null;index"`
	House       *string `json:"house" gorm:"type:varchar(255)"`
	Street      *string `json:"street" gorm:"type:varchar(255)"`
	Landmark    *string `json:"landmark" gorm:"type:varchar(255)"`
	PostOffice  *string `json:"post_office" gorm:"type:varchar(255)"`
	Subdistrict *string `json:"subdistrict" gorm:"type:varchar(255)"`
	District    *string `json:"district" gorm:"type:varchar(255)"`
	VTC         *string `json:"vtc" gorm:"type:varchar(255)"` // Village/Town/City
	State       *string `json:"state" gorm:"type:varchar(255)"`
	Country     *string `json:"country" gorm:"type:varchar(255)"`
	Pincode     *string `json:"pincode" gorm:"type:varchar(10)"`
	FullAddress *string `json:"full_address" gorm:"type:text"`
}

const (
	AddressTable     = "ADDR"
	AddressTableSize = hash.Large
)

// NewAddress creates a new Address instance
func NewAddress() *Address {
	return &Address{
		BaseModel: base.NewBaseModel(AddressTable, AddressTableSize),
	}
}

// BeforeCreate is called before creating a new address
func (a *Address) BeforeCreate() error {
	return a.BaseModel.BeforeCreate()
}

// BeforeUpdate is called before updating an address
func (a *Address) BeforeUpdate() error {
	return a.BaseModel.BeforeUpdate()
}

// BeforeDelete is called before deleting an address
func (a *Address) BeforeDelete() error {
	return a.BaseModel.BeforeDelete()
}

// BeforeSoftDelete is called before soft deleting an address
func (a *Address) BeforeSoftDelete() error {
	return a.BaseModel.BeforeSoftDelete()
}

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (a *Address) BeforeCreateGORM(tx *gorm.DB) error {
	return a.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (a *Address) BeforeUpdateGORM(tx *gorm.DB) error {
	return a.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (a *Address) BeforeDeleteGORM(tx *gorm.DB) error {
	return a.BeforeDelete()
}

// GetTableIdentifier returns the table identifier for Address
func (a *Address) GetTableIdentifier() string {
	return AddressTable
}

// GetTableSize returns the table size for Address
func (a *Address) GetTableSize() hash.TableSize {
	return AddressTableSize
}

// TableName returns the GORM table name for this model
func (a *Address) TableName() string { return "addresses" }

// Explicit method implementations to satisfy linter
func (a *Address) GetID() string   { return a.BaseModel.GetID() }
func (a *Address) SetID(id string) { a.BaseModel.SetID(id) }

// BuildFullAddress builds the full address string from individual components
func (a *Address) BuildFullAddress() string {
	var parts []string

	if a.House != nil && *a.House != "" {
		parts = append(parts, *a.House)
	}

	if a.Street != nil && *a.Street != "" {
		parts = append(parts, *a.Street)
	}

	if a.Landmark != nil && *a.Landmark != "" {
		parts = append(parts, *a.Landmark)
	}

	if a.PostOffice != nil && *a.PostOffice != "" {
		parts = append(parts, *a.PostOffice)
	}

	if a.Subdistrict != nil && *a.Subdistrict != "" {
		parts = append(parts, *a.Subdistrict)
	}

	if a.District != nil && *a.District != "" {
		parts = append(parts, *a.District)
	}

	if a.VTC != nil && *a.VTC != "" {
		parts = append(parts, *a.VTC)
	}

	if a.State != nil && *a.State != "" {
		parts = append(parts, *a.State)
	}

	if a.Country != nil && *a.Country != "" {
		parts = append(parts, *a.Country)
	}

	if a.Pincode != nil && *a.Pincode != "" {
		parts = append(parts, *a.Pincode)
	}

	fullAddress := ""
	for i, part := range parts {
		if i > 0 {
			fullAddress += ", "
		}
		fullAddress += part
	}

	a.FullAddress = &fullAddress
	return fullAddress
}

// IsComplete checks if the address has all required fields
func (a *Address) IsComplete() bool {
	return a.House != nil && a.Street != nil && a.District != nil &&
		a.State != nil && a.Country != nil && a.Pincode != nil
}

// GetResourceType returns the PostgreSQL RBAC resource type for addresses
func (a *Address) GetResourceType() string {
	return "aaa/address"
}

// GetObjectID returns the PostgreSQL RBAC object ID for this address
func (a *Address) GetObjectID() string {
	return a.GetID()
}
