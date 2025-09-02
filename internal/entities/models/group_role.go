package models

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// GroupRole represents the relationship between groups and roles within an organization
type GroupRole struct {
	*base.BaseModel
	GroupID        string     `json:"group_id" gorm:"type:varchar(255);not null;index:idx_group_roles_group_org,priority:1;index:idx_group_roles_group_role,priority:1"`
	RoleID         string     `json:"role_id" gorm:"type:varchar(255);not null;index:idx_group_roles_role_org,priority:1;index:idx_group_roles_group_role,priority:2"`
	OrganizationID string     `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_group_roles_group_org,priority:2;index:idx_group_roles_role_org,priority:2"`
	AssignedBy     string     `json:"assigned_by" gorm:"type:varchar(255);not null"`
	StartsAt       *time.Time `json:"starts_at" gorm:"type:timestamp"`
	EndsAt         *time.Time `json:"ends_at" gorm:"type:timestamp"`
	IsActive       bool       `json:"is_active" gorm:"default:true;index:idx_group_roles_active"`
	Metadata       *string    `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Group        *Group        `json:"group" gorm:"foreignKey:GroupID;references:ID"`
	Role         *Role         `json:"role" gorm:"foreignKey:RoleID;references:ID"`
	Organization *Organization `json:"organization" gorm:"foreignKey:OrganizationID;references:ID"`
	Assigner     *User         `json:"assigner" gorm:"foreignKey:AssignedBy;references:ID"`
}

// NewGroupRole creates a new GroupRole instance
func NewGroupRole(groupID, roleID, organizationID, assignedBy string) *GroupRole {
	return &GroupRole{
		BaseModel:      base.NewBaseModel("GRPR", hash.Small),
		GroupID:        groupID,
		RoleID:         roleID,
		OrganizationID: organizationID,
		AssignedBy:     assignedBy,
		IsActive:       true,
	}
}

// NewGroupRoleWithTimebound creates a new GroupRole instance with time bounds
func NewGroupRoleWithTimebound(groupID, roleID, organizationID, assignedBy string, startsAt, endsAt *time.Time) *GroupRole {
	groupRole := NewGroupRole(groupID, roleID, organizationID, assignedBy)
	groupRole.StartsAt = startsAt
	groupRole.EndsAt = endsAt
	return groupRole
}

// BeforeCreate is called before creating a new group role
func (gr *GroupRole) BeforeCreate() error {
	return gr.BaseModel.BeforeCreate()
}

// BeforeUpdate is called before updating a group role
func (gr *GroupRole) BeforeUpdate() error {
	return gr.BaseModel.BeforeUpdate()
}

// BeforeDelete is called before deleting a group role
func (gr *GroupRole) BeforeDelete() error {
	return gr.BaseModel.BeforeDelete()
}

// BeforeSoftDelete is called before soft deleting a group role
func (gr *GroupRole) BeforeSoftDelete() error {
	return gr.BaseModel.BeforeSoftDelete()
}

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (gr *GroupRole) BeforeCreateGORM(tx *gorm.DB) error {
	return gr.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (gr *GroupRole) BeforeUpdateGORM(tx *gorm.DB) error {
	return gr.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (gr *GroupRole) BeforeDeleteGORM(tx *gorm.DB) error {
	return gr.BeforeDelete()
}

// Helper methods
func (gr *GroupRole) GetTableIdentifier() string   { return "GRPR" }
func (gr *GroupRole) GetTableSize() hash.TableSize { return hash.Small }

// TableName returns the GORM table name for this model
func (gr *GroupRole) TableName() string { return "group_roles" }

// Explicit method implementations to satisfy linter
func (gr *GroupRole) GetID() string   { return gr.BaseModel.GetID() }
func (gr *GroupRole) SetID(id string) { gr.BaseModel.SetID(id) }

// IsEffective checks if the group role assignment is currently effective based on time bounds
func (gr *GroupRole) IsEffective(at time.Time) bool {
	if !gr.IsActive {
		return false
	}

	if gr.StartsAt != nil && at.Before(*gr.StartsAt) {
		return false
	}

	if gr.EndsAt != nil && at.After(*gr.EndsAt) {
		return false
	}

	return true
}

// IsCurrentlyEffective checks if the group role assignment is currently effective
func (gr *GroupRole) IsCurrentlyEffective() bool {
	return gr.IsEffective(time.Now())
}

// GetResourceType returns the PostgreSQL RBAC resource type for group roles
func (gr *GroupRole) GetResourceType() string {
	return ResourceTypeGroupRole
}

// GetObjectID returns the PostgreSQL RBAC object ID for this group role
func (gr *GroupRole) GetObjectID() string {
	return gr.GetID()
}

// Validate performs basic validation on the GroupRole
func (gr *GroupRole) Validate() error {
	if gr.GroupID == "" {
		return ErrInvalidGroupID
	}
	if gr.RoleID == "" {
		return ErrInvalidRoleID
	}
	if gr.OrganizationID == "" {
		return ErrInvalidOrganizationID
	}
	if gr.AssignedBy == "" {
		return ErrInvalidAssignedBy
	}
	if gr.StartsAt != nil && gr.EndsAt != nil && gr.StartsAt.After(*gr.EndsAt) {
		return ErrInvalidTimeRange
	}
	return nil
}

// Custom error types for GroupRole validation
var (
	ErrInvalidGroupID        = &ValidationError{Field: "group_id", Message: "group_id cannot be empty"}
	ErrInvalidRoleID         = &ValidationError{Field: "role_id", Message: "role_id cannot be empty"}
	ErrInvalidOrganizationID = &ValidationError{Field: "organization_id", Message: "organization_id cannot be empty"}
	ErrInvalidAssignedBy     = &ValidationError{Field: "assigned_by", Message: "assigned_by cannot be empty"}
	ErrInvalidTimeRange      = &ValidationError{Field: "time_range", Message: "starts_at cannot be after ends_at"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
