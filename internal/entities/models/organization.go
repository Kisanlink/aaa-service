package models

import (
	"fmt"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// Organization represents an organization in the AAA service
type Organization struct {
	*base.BaseModel
	Name           string  `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Type           string  `json:"type" gorm:"size:50;default:'individual'"` // "enterprise", "small_business", "individual"
	Description    string  `json:"description" gorm:"type:text"`
	ParentID       *string `json:"parent_id" gorm:"type:varchar(255);default:null"` // For org hierarchy
	IsActive       bool    `json:"is_active" gorm:"default:true"`
	Metadata       *string `json:"metadata" gorm:"type:jsonb"`                                                                                            // Additional org metadata
	Version        int     `json:"version" gorm:"column:version;default:1;not null"`                                                                      // Optimistic locking
	HierarchyDepth int     `json:"hierarchy_depth" gorm:"column:hierarchy_depth;default:0;not null;check:hierarchy_depth >= 0 AND hierarchy_depth <= 10"` // Depth in hierarchy (0=root, max=10)
	HierarchyPath  string  `json:"hierarchy_path" gorm:"column:hierarchy_path;type:text;index:idx_org_hierarchy_path"`                                    // Materialized path for efficient queries

	// Relationships
	Parent   *Organization  `json:"parent" gorm:"foreignKey:ParentID;references:ID"`
	Children []Organization `json:"children" gorm:"foreignKey:ParentID;references:ID"`
}

const (
	OrgTable     = "ORGN"
	OrgTableSize = hash.Medium

	// Organization types
	OrgTypeEnterprise        = "enterprise"
	OrgTypeSmallBusiness     = "small_business"
	OrgTypeIndividual        = "individual"
	OrgTypeFPO               = "fpo"                // Farmer Producer Organization
	OrgTypeCooperative       = "cooperative"        // Agricultural Cooperative
	OrgTypeAgribusiness      = "agribusiness"       // Agribusiness Company
	OrgTypeFarmersGroup      = "farmers_group"      // Informal Farmers Group
	OrgTypeSHG               = "shg"                // Self Help Group
	OrgTypeNGO               = "ngo"                // Non-Governmental Organization
	OrgTypeGovernment        = "government"         // Government Agency
	OrgTypeInputSupplier     = "input_supplier"     // Seeds, Fertilizers, Equipment Suppliers
	OrgTypeTrader            = "trader"             // Agricultural Traders/Aggregators
	OrgTypeProcessingUnit    = "processing_unit"    // Food Processing Units
	OrgTypeResearchInstitute = "research_institute" // Agricultural Research Organizations
)

// NewOrganization creates a new Organization instance
func NewOrganization(name, description, orgType string) *Organization {
	// Default to individual if type is empty
	if orgType == "" {
		orgType = OrgTypeIndividual
	}
	return &Organization{
		BaseModel:   base.NewBaseModel("ORGN", hash.Medium),
		Name:        name,
		Type:        orgType,
		Description: description,
		IsActive:    true,
	}
}

// ValidOrganizationType checks if the provided organization type is valid
func ValidOrganizationType(orgType string) bool {
	validTypes := map[string]bool{
		OrgTypeEnterprise:        true,
		OrgTypeSmallBusiness:     true,
		OrgTypeIndividual:        true,
		OrgTypeFPO:               true,
		OrgTypeCooperative:       true,
		OrgTypeAgribusiness:      true,
		OrgTypeFarmersGroup:      true,
		OrgTypeSHG:               true,
		OrgTypeNGO:               true,
		OrgTypeGovernment:        true,
		OrgTypeInputSupplier:     true,
		OrgTypeTrader:            true,
		OrgTypeProcessingUnit:    true,
		OrgTypeResearchInstitute: true,
	}
	return validTypes[orgType]
}

// BeforeCreate is called before creating a new organization
func (o *Organization) BeforeCreate() error {
	// Set hierarchy fields if not already set
	if o.ParentID == nil || *o.ParentID == "" {
		// Root organization
		o.HierarchyDepth = 0
		if o.BaseModel != nil && o.BaseModel.ID != "" {
			o.HierarchyPath = "/" + o.BaseModel.ID
		}
	}
	return o.BaseModel.BeforeCreate()
}

// GORM Hooks
func (o *Organization) BeforeCreateGORM(tx *gorm.DB) error {
	// Calculate hierarchy fields based on parent
	if o.ParentID != nil && *o.ParentID != "" {
		var parent Organization
		if err := tx.Where("id = ?", *o.ParentID).First(&parent).Error; err == nil {
			o.HierarchyDepth = parent.HierarchyDepth + 1
			if o.HierarchyDepth > 10 {
				return fmt.Errorf("maximum organization hierarchy depth (10) exceeded")
			}
			// Path will be set after ID is generated
		}
	} else {
		o.HierarchyDepth = 0
	}
	return o.BeforeCreate()
}

// AfterCreate updates the hierarchy path after the ID is generated
func (o *Organization) AfterCreate(tx *gorm.DB) error {
	if o.HierarchyPath == "" && o.BaseModel != nil && o.BaseModel.ID != "" {
		if o.ParentID != nil && *o.ParentID != "" {
			var parent Organization
			if err := tx.Where("id = ?", *o.ParentID).First(&parent).Error; err == nil {
				o.HierarchyPath = parent.HierarchyPath + "/" + o.BaseModel.ID
			} else {
				o.HierarchyPath = "/" + o.BaseModel.ID
			}
		} else {
			o.HierarchyPath = "/" + o.BaseModel.ID
		}
		// Update the path in the database
		return tx.Model(o).Update("hierarchy_path", o.HierarchyPath).Error
	}
	return nil
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
func (o *Organization) GetTableIdentifier() string   { return "ORGN" }
func (o *Organization) GetTableSize() hash.TableSize { return hash.Medium }

// TableName returns the GORM table name for this model
func (o *Organization) TableName() string { return "organizations" }

// Explicit method implementations to satisfy linter
func (o *Organization) GetID() string   { return o.BaseModel.GetID() }
func (o *Organization) SetID(id string) { o.BaseModel.SetID(id) }

// GetResourceType returns the PostgreSQL RBAC resource type for organizations
func (o *Organization) GetResourceType() string {
	return "aaa/organization"
}

// GetObjectID returns the object ID for this organization
func (o *Organization) GetObjectID() string {
	return o.GetID()
}
