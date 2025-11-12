package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// ServiceRoleMapping tracks which service owns/seeded which roles for audit trail and cleanup
type ServiceRoleMapping struct {
	*base.BaseModel
	ServiceID   string  `json:"service_id" gorm:"size:255;not null;index:idx_service_role_mappings_service"`
	ServiceName string  `json:"service_name" gorm:"size:100;not null"`
	RoleID      string  `json:"role_id" gorm:"type:varchar(255);not null;index:idx_service_role_mappings_role"`
	Version     int     `json:"version" gorm:"default:1"`
	Metadata    *string `json:"metadata" gorm:"type:jsonb"`
	IsActive    bool    `json:"is_active" gorm:"default:true;index"`

	// Relationships
	Role *Role `json:"role" gorm:"foreignKey:RoleID;references:ID"`
}

// NewServiceRoleMapping creates a new ServiceRoleMapping instance
func NewServiceRoleMapping(serviceID, serviceName, roleID string) *ServiceRoleMapping {
	return &ServiceRoleMapping{
		BaseModel:   base.NewBaseModel("SRM", hash.Small),
		ServiceID:   serviceID,
		ServiceName: serviceName,
		RoleID:      roleID,
		Version:     1,
		IsActive:    true,
	}
}

// BeforeCreate is called before creating a new service role mapping
func (srm *ServiceRoleMapping) BeforeCreate() error {
	return srm.BaseModel.BeforeCreate()
}

// BeforeUpdate is called before updating a service role mapping
func (srm *ServiceRoleMapping) BeforeUpdate() error {
	return srm.BaseModel.BeforeUpdate()
}

// BeforeDelete is called before deleting a service role mapping
func (srm *ServiceRoleMapping) BeforeDelete() error {
	return srm.BaseModel.BeforeDelete()
}

// BeforeSoftDelete is called before soft deleting a service role mapping
func (srm *ServiceRoleMapping) BeforeSoftDelete() error {
	return srm.BaseModel.BeforeSoftDelete()
}

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (srm *ServiceRoleMapping) BeforeCreateGORM(tx *gorm.DB) error {
	return srm.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (srm *ServiceRoleMapping) BeforeUpdateGORM(tx *gorm.DB) error {
	return srm.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (srm *ServiceRoleMapping) BeforeDeleteGORM(tx *gorm.DB) error {
	return srm.BeforeDelete()
}

// Helper methods
func (srm *ServiceRoleMapping) GetTableIdentifier() string   { return "SRM" }
func (srm *ServiceRoleMapping) GetTableSize() hash.TableSize { return hash.Small }

// TableName returns the GORM table name for this model
func (srm *ServiceRoleMapping) TableName() string { return "service_role_mappings" }

// Explicit method implementations to satisfy linter
func (srm *ServiceRoleMapping) GetID() string   { return srm.BaseModel.GetID() }
func (srm *ServiceRoleMapping) SetID(id string) { srm.BaseModel.SetID(id) }

// Deactivate marks the mapping as inactive
func (srm *ServiceRoleMapping) Deactivate() {
	srm.IsActive = false
}

// Activate marks the mapping as active
func (srm *ServiceRoleMapping) Activate() {
	srm.IsActive = true
}

// IncrementVersion increments the version number
func (srm *ServiceRoleMapping) IncrementVersion() {
	srm.Version++
}
