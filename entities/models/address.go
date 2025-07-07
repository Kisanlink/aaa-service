package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Address represents a user's address
type Address struct {
	*base.BaseModel
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

// NewAddress creates a new Address instance
func NewAddress() *Address {
	return &Address{
		BaseModel: base.NewBaseModel("addr", hash.TableSizeSmall),
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

// GetTableIdentifier returns the table identifier for Address
func (a *Address) GetTableIdentifier() string {
	return "addr"
}

// GetTableSize returns the table size for Address
func (a *Address) GetTableSize() hash.TableSize {
	return hash.TableSizeSmall
}

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
