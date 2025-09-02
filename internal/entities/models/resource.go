package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// Resource represents a resource in the AAA service that corresponds to PostgreSQL RBAC resource definitions
type Resource struct {
	*base.BaseModel
	Name        string  `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Type        string  `json:"type" gorm:"size:100;not null;index"` // Resource type (e.g., "aaa/user", "aaa/role")
	Description string  `json:"description" gorm:"type:text"`
	IsActive    bool    `json:"is_active" gorm:"default:true"`
	ParentID    *string `json:"parent_id" gorm:"type:varchar(255);default:null"` // For resource hierarchy
	OwnerID     *string `json:"owner_id" gorm:"type:varchar(255);default:null"`  // Resource owner

	// Relationships
	Parent   *Resource  `json:"parent" gorm:"foreignKey:ParentID;references:ID"`
	Children []Resource `json:"children" gorm:"foreignKey:ParentID;references:ID"`
	Owner    *User      `json:"owner" gorm:"foreignKey:OwnerID;references:ID"`
}

// ResourceType constants matching PostgreSQL RBAC schema
const (
	ResourceTypeUser                 = "aaa/user"
	ResourceTypeUserProfile          = "aaa/user_profile"
	ResourceTypeContact              = "aaa/contact"
	ResourceTypeAddress              = "aaa/address"
	ResourceTypeRole                 = "aaa/role"
	ResourceTypePermission           = "aaa/permission"
	ResourceTypeResource             = "aaa/resource"
	ResourceTypeAction               = "aaa/action"
	ResourceTypeColumnPermission     = "aaa/column_permission"
	ResourceTypeColumn               = "aaa/column"
	ResourceTypeUserResource         = "aaa/user_resource"
	ResourceTypeRoleResource         = "aaa/role_resource"
	ResourceTypePermissionResource   = "aaa/permission_resource"
	ResourceTypeAuditLog             = "aaa/audit_log"
	ResourceTypeSystem               = "aaa/system"
	ResourceTypeTemporaryPermission  = "aaa/temporary_permission"
	ResourceTypeHierarchicalResource = "aaa/hierarchical_resource"
	ResourceTypeAPIEndpoint          = "aaa/api_endpoint"
	ResourceTypeDatabaseOperation    = "aaa/database_operation"
	ResourceTypeTable                = "aaa/table"
	ResourceTypeDatabase             = "aaa/database"
	ResourceTypeGroup                = "aaa/group"
	ResourceTypeGroupRole            = "aaa/group_role"
	ResourceTypeOrganization         = "aaa/organization"
)

// NewResource creates a new Resource instance
func NewResource(name, resourceType, description string) *Resource {
	return &Resource{
		BaseModel:   base.NewBaseModel("RES", hash.Small),
		Name:        name,
		Type:        resourceType,
		Description: description,
		IsActive:    true,
	}
}

// NewResourceWithParent creates a new Resource instance with a parent
func NewResourceWithParent(name, resourceType, description, parentID string) *Resource {
	resource := NewResource(name, resourceType, description)
	resource.ParentID = &parentID
	return resource
}

// NewResourceWithOwner creates a new Resource instance with an owner
func NewResourceWithOwner(name, resourceType, description, ownerID string) *Resource {
	resource := NewResource(name, resourceType, description)
	resource.OwnerID = &ownerID
	return resource
}

// BeforeCreate is called before creating a new resource
func (r *Resource) BeforeCreate() error {
	return r.BaseModel.BeforeCreate()
}

// BeforeUpdate is called before updating a resource
func (r *Resource) BeforeUpdate() error {
	return r.BaseModel.BeforeUpdate()
}

// BeforeDelete is called before deleting a resource
func (r *Resource) BeforeDelete() error {
	return r.BaseModel.BeforeDelete()
}

// BeforeSoftDelete is called before soft deleting a resource
func (r *Resource) BeforeSoftDelete() error {
	return r.BaseModel.BeforeSoftDelete()
}

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (r *Resource) BeforeCreateGORM(tx *gorm.DB) error {
	return r.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (r *Resource) BeforeUpdateGORM(tx *gorm.DB) error {
	return r.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (r *Resource) BeforeDeleteGORM(tx *gorm.DB) error {
	return r.BeforeDelete()
}

// Helper methods
func (r *Resource) GetTableIdentifier() string   { return "RES" }
func (r *Resource) GetTableSize() hash.TableSize { return hash.Medium }

// TableName returns the GORM table name for this model
func (r *Resource) TableName() string { return "resources" }

// Explicit method implementations to satisfy linter
func (r *Resource) GetID() string   { return r.BaseModel.GetID() }
func (r *Resource) SetID(id string) { r.BaseModel.SetID(id) }

// HasParent checks if the resource has a parent
func (r *Resource) HasParent() bool {
	return r.ParentID != nil
}

// HasOwner checks if the resource has an owner
func (r *Resource) HasOwner() bool {
	return r.OwnerID != nil
}

// IsResourceType checks if the resource is of a specific type
func (r *Resource) IsResourceType(resourceType string) bool {
	return r.Type == resourceType
}

// GetObjectType returns the PostgreSQL RBAC object type for this resource
func (r *Resource) GetObjectType() string {
	return r.Type
}

// GetResourceType returns the PostgreSQL RBAC resource type for resources
func (r *Resource) GetResourceType() string {
	return "aaa/resource"
}

// GetObjectID returns the PostgreSQL RBAC object ID for this resource
func (r *Resource) GetObjectID() string {
	return r.ID
}
