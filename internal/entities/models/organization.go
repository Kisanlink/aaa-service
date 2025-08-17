package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// Organization represents an organization in the AAA service
type Organization struct {
	*base.BaseModel
	Name        string  `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description string  `json:"description" gorm:"type:text"`
	ParentID    *string `json:"parent_id" gorm:"type:varchar(255);default:null"` // For org hierarchy
	IsActive    bool    `json:"is_active" gorm:"default:true"`
	Metadata    *string `json:"metadata" gorm:"type:jsonb"` // Additional org metadata

	// Relationships
	Parent   *Organization  `json:"parent" gorm:"foreignKey:ParentID;references:ID"`
	Children []Organization `json:"children" gorm:"foreignKey:ParentID;references:ID"`
}

const (
	OrgTable     = "ORG"
	OrgTableSize = hash.Medium
)

// NewOrganization creates a new Organization instance
func NewOrganization(name, description string) *Organization {
	return &Organization{
		BaseModel:   base.NewBaseModel("org", hash.Medium),
		Name:        name,
		Description: description,
		IsActive:    true,
	}
}

// BeforeCreate is called before creating a new organization
func (o *Organization) BeforeCreate() error {
	return o.BaseModel.BeforeCreate()
}

// GORM Hooks
func (o *Organization) BeforeCreateGORM(tx *gorm.DB) error {
	return o.BeforeCreate()
}

func (o *Organization) BeforeUpdateGORM(tx *gorm.DB) error {
	return o.BeforeUpdate()
}

func (o *Organization) BeforeDeleteGORM(tx *gorm.DB) error {
	return o.BeforeDelete()
}

func (o *Organization) BeforeUpdate() error     { return o.BaseModel.BeforeUpdate() }
func (o *Organization) BeforeDelete() error     { return o.BaseModel.BeforeDelete() }
func (o *Organization) BeforeSoftDelete() error { return o.BaseModel.BeforeSoftDelete() }

// Helper methods
func (o *Organization) GetTableIdentifier() string   { return "ORG" }
func (o *Organization) GetTableSize() hash.TableSize { return hash.Medium }

// Explicit method implementations to satisfy linter
func (o *Organization) GetID() string   { return o.BaseModel.GetID() }
func (o *Organization) SetID(id string) { o.BaseModel.SetID(id) }

// GetResourceType returns the resource type for organizations
func (o *Organization) GetResourceType() string {
	return ResourceTypeOrganization
}

// GetObjectID returns the object ID for this organization
func (o *Organization) GetObjectID() string {
	return o.GetID()
}
