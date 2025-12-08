package models

import (
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// Group represents a group in the AAA service that can own roles/policies
type Group struct {
	*base.BaseModel
	Name           string  `json:"name" gorm:"size:100;not null;index:idx_groups_name"`
	Description    string  `json:"description" gorm:"type:text"`
	OrganizationID string  `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_groups_org_active,priority:1"`
	ParentID       *string `json:"parent_id" gorm:"type:varchar(255);default:null;index:idx_groups_parent"` // For group hierarchy
	IsActive       bool    `json:"is_active" gorm:"default:true;index:idx_groups_org_active,priority:2;index:idx_groups_active"`
	Metadata       *string `json:"metadata" gorm:"type:jsonb"`                                                                                           // Additional group metadata
	Version        int     `json:"version" gorm:"column:version;default:1;not null"`                                                                     // Optimistic locking
	HierarchyDepth int     `json:"hierarchy_depth" gorm:"column:hierarchy_depth;default:0;not null;check:hierarchy_depth >= 0 AND hierarchy_depth <= 8"` // Depth in hierarchy (0=root, max=8)
	HierarchyPath  string  `json:"hierarchy_path" gorm:"column:hierarchy_path;type:text;index:idx_group_hierarchy_path"`                                 // Materialized path for efficient queries

	// Relationships
	Organization *Organization      `json:"organization" gorm:"foreignKey:OrganizationID;references:ID"`
	Parent       *Group             `json:"parent" gorm:"foreignKey:ParentID;references:ID"`
	Children     []Group            `json:"children" gorm:"foreignKey:ParentID;references:ID"`
	Memberships  []GroupMembership  `json:"memberships" gorm:"foreignKey:GroupID"`
	ChildGroups  []GroupInheritance `json:"child_groups" gorm:"foreignKey:ParentGroupID"`
	ParentGroups []GroupInheritance `json:"parent_groups" gorm:"foreignKey:ChildGroupID"`
}

// GroupMembership represents a user's membership in a group with time bounds
type GroupMembership struct {
	*base.BaseModel
	GroupID       string     `json:"group_id" gorm:"type:varchar(255);not null;index:idx_gm_group;index:idx_gm_principal_group,priority:2"`
	PrincipalID   string     `json:"principal_id" gorm:"type:varchar(255);not null;index:idx_gm_principal;index:idx_gm_principal_group,priority:1;index:idx_gm_principal_type_active,priority:1"` // User or Service principal
	PrincipalType string     `json:"principal_type" gorm:"size:50;not null;index:idx_gm_principal_type_active,priority:2"`                                                                        // "user" or "service"
	StartsAt      *time.Time `json:"starts_at" gorm:"type:timestamp"`                                                                                                                             // When membership becomes active
	EndsAt        *time.Time `json:"ends_at" gorm:"type:timestamp"`                                                                                                                               // When membership expires
	IsActive      bool       `json:"is_active" gorm:"default:true;index:idx_gm_active;index:idx_gm_principal_type_active,priority:3"`
	AddedByID     string     `json:"added_by_id" gorm:"type:varchar(255);not null"`    // Who added this member
	Metadata      *string    `json:"metadata" gorm:"type:jsonb"`                       // Additional membership metadata
	Version       int        `json:"version" gorm:"column:version;default:1;not null"` // Optimistic locking

	// Relationships
	Group   *Group `json:"group" gorm:"foreignKey:GroupID;references:ID"`
	AddedBy *User  `json:"added_by" gorm:"foreignKey:AddedByID;references:ID"`
}

// GroupInheritance represents group-to-group inheritance relationships
type GroupInheritance struct {
	*base.BaseModel
	ParentGroupID string     `json:"parent_group_id" gorm:"type:varchar(255);not null"`
	ChildGroupID  string     `json:"child_group_id" gorm:"type:varchar(255);not null"`
	StartsAt      *time.Time `json:"starts_at" gorm:"type:timestamp"`
	EndsAt        *time.Time `json:"ends_at" gorm:"type:timestamp"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`

	// Relationships
	ParentGroup *Group `json:"parent_group" gorm:"foreignKey:ParentGroupID;references:ID"`
	ChildGroup  *Group `json:"child_group" gorm:"foreignKey:ChildGroupID;references:ID"`
}

// Add unique constraints
func (g *Group) TableName() string {
	return "groups"
}

func (gm *GroupMembership) TableName() string {
	return "group_memberships"
}

func (gi *GroupInheritance) TableName() string {
	return "group_inheritance"
}

// NewGroup creates a new Group instance
func NewGroup(name, description, organizationID string) *Group {
	return &Group{
		BaseModel:      base.NewBaseModel("GRPN", hash.Medium),
		Name:           name,
		Description:    description,
		OrganizationID: organizationID,
		IsActive:       true,
	}
}

// NewGroupMembership creates a new GroupMembership instance
func NewGroupMembership(groupID, principalID, principalType, addedByID string) *GroupMembership {
	return &GroupMembership{
		BaseModel:     base.NewBaseModel("GRPM", hash.Medium),
		GroupID:       groupID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
		AddedByID:     addedByID,
		IsActive:      true,
	}
}

// NewGroupInheritance creates a new GroupInheritance instance
func NewGroupInheritance(parentGroupID, childGroupID string) *GroupInheritance {
	return &GroupInheritance{
		BaseModel:     base.NewBaseModel("GRPI", hash.Small),
		ParentGroupID: parentGroupID,
		ChildGroupID:  childGroupID,
		IsActive:      true,
	}
}

// BeforeCreate hooks
func (g *Group) BeforeCreate() error {
	// Set hierarchy fields if not already set
	if g.ParentID == nil || *g.ParentID == "" {
		// Root group
		g.HierarchyDepth = 0
		if g.BaseModel != nil && g.BaseModel.ID != "" {
			g.HierarchyPath = "/" + g.BaseModel.ID
		}
	}
	return g.BaseModel.BeforeCreate()
}

func (gm *GroupMembership) BeforeCreate() error {
	return gm.BaseModel.BeforeCreate()
}

func (gi *GroupInheritance) BeforeCreate() error {
	return gi.BaseModel.BeforeCreate()
}

// GORM Hooks
func (g *Group) BeforeCreateGORM(tx *gorm.DB) error {
	// Calculate hierarchy fields based on parent
	if g.ParentID != nil && *g.ParentID != "" {
		var parent Group
		if err := tx.Where("id = ?", *g.ParentID).First(&parent).Error; err == nil {
			g.HierarchyDepth = parent.HierarchyDepth + 1
			if g.HierarchyDepth > 8 {
				return fmt.Errorf("maximum group hierarchy depth (8) exceeded")
			}
			// Path will be set after ID is generated
		}
	} else {
		g.HierarchyDepth = 0
	}
	return g.BeforeCreate()
}

// AfterCreate updates the hierarchy path after the ID is generated
func (g *Group) AfterCreate(tx *gorm.DB) error {
	if g.HierarchyPath == "" && g.BaseModel != nil && g.BaseModel.ID != "" {
		if g.ParentID != nil && *g.ParentID != "" {
			var parent Group
			if err := tx.Where("id = ?", *g.ParentID).First(&parent).Error; err == nil {
				g.HierarchyPath = parent.HierarchyPath + "/" + g.BaseModel.ID
			} else {
				g.HierarchyPath = "/" + g.BaseModel.ID
			}
		} else {
			g.HierarchyPath = "/" + g.BaseModel.ID
		}
		// Update the path in the database
		return tx.Model(g).Update("hierarchy_path", g.HierarchyPath).Error
	}
	return nil
}

func (gm *GroupMembership) BeforeCreateGORM(tx *gorm.DB) error {
	return gm.BeforeCreate()
}

func (gi *GroupInheritance) BeforeCreateGORM(tx *gorm.DB) error {
	return gi.BeforeCreate()
}

// Helper methods
func (g *Group) GetTableIdentifier() string   { return "GRPN" }
func (g *Group) GetTableSize() hash.TableSize { return hash.Medium }

// Explicit method implementations to satisfy linter
func (g *Group) GetID() string   { return g.BaseModel.GetID() }
func (g *Group) SetID(id string) { g.BaseModel.SetID(id) }

func (gm *GroupMembership) GetTableIdentifier() string   { return "GRPM" }
func (gm *GroupMembership) GetTableSize() hash.TableSize { return hash.Medium }

func (gi *GroupInheritance) GetTableIdentifier() string   { return "GRPI" }
func (gi *GroupInheritance) GetTableSize() hash.TableSize { return hash.Small }

// IsEffective checks if the membership is currently effective based on time bounds
func (gm *GroupMembership) IsEffective(at time.Time) bool {
	if !gm.IsActive {
		return false
	}

	if gm.StartsAt != nil && at.Before(*gm.StartsAt) {
		return false
	}

	if gm.EndsAt != nil && at.After(*gm.EndsAt) {
		return false
	}

	return true
}

// IsEffective checks if the inheritance is currently effective based on time bounds
func (gi *GroupInheritance) IsEffective(at time.Time) bool {
	if !gi.IsActive {
		return false
	}

	if gi.StartsAt != nil && at.Before(*gi.StartsAt) {
		return false
	}

	if gi.EndsAt != nil && at.After(*gi.EndsAt) {
		return false
	}

	return true
}

// GetResourceType returns the resource type for groups
func (g *Group) GetResourceType() string {
	return ResourceTypeGroup
}

// GetObjectID returns the object ID for this group
func (g *Group) GetObjectID() string {
	return g.GetID()
}
