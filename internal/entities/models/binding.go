package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// BindingSubjectType defines what type of subject the binding applies to
type BindingSubjectType string

const (
	BindingSubjectUser    BindingSubjectType = "user"
	BindingSubjectGroup   BindingSubjectType = "group"
	BindingSubjectService BindingSubjectType = "service"
)

// BindingType defines whether this is a role or permission binding
type BindingType string

const (
	BindingTypeRole       BindingType = "role"
	BindingTypePermission BindingType = "permission"
)

// Caveat represents authorization constraints in JSONB format
type Caveat map[string]interface{}

// Scan implements the Scanner interface for database reads
func (c *Caveat) Scan(value interface{}) error {
	if value == nil {
		*c = make(Caveat)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		return errors.New("cannot scan caveat from database")
	}
}

// Value implements the Valuer interface for database writes
func (c Caveat) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Binding represents a subject-to-role/permission binding with optional caveats
type Binding struct {
	*base.BaseModel
	SubjectID      string             `json:"subject_id" gorm:"type:varchar(255);not null"` // User, Group, or Service ID
	SubjectType    BindingSubjectType `json:"subject_type" gorm:"size:20;not null"`         // user, group, service
	BindingType    BindingType        `json:"binding_type" gorm:"size:20;not null"`         // role or permission
	RoleID         *string            `json:"role_id" gorm:"type:varchar(255)"`             // When binding_type is role
	PermissionID   *string            `json:"permission_id" gorm:"type:varchar(255)"`       // When binding_type is permission
	ResourceType   string             `json:"resource_type" gorm:"size:100;not null"`       // e.g., "aaa/table", "aaa/resource"
	ResourceID     *string            `json:"resource_id" gorm:"type:varchar(255)"`         // Specific resource instance, null for type-level
	OrganizationID string             `json:"organization_id" gorm:"type:varchar(255);not null"`
	Caveat         *Caveat            `json:"caveat" gorm:"type:jsonb"` // JSON constraints
	Version        int                `json:"version" gorm:"default:1"`
	CreatedByID    string             `json:"created_by_id" gorm:"type:varchar(255);not null"`
	IsActive       bool               `json:"is_active" gorm:"default:true"`

	// Relationships
	Role         *Role         `json:"role" gorm:"foreignKey:RoleID;references:ID"`
	Permission   *Permission   `json:"permission" gorm:"foreignKey:PermissionID;references:ID"`
	Organization *Organization `json:"organization" gorm:"foreignKey:OrganizationID;references:ID"`
	CreatedBy    *User         `json:"created_by" gorm:"foreignKey:CreatedByID;references:ID"`
}

// BindingHistory tracks changes to bindings for audit and rollback
type BindingHistory struct {
	*base.BaseModel
	BindingID      string             `json:"binding_id" gorm:"type:varchar(255);not null"`
	SubjectID      string             `json:"subject_id" gorm:"type:varchar(255);not null"`
	SubjectType    BindingSubjectType `json:"subject_type" gorm:"size:20;not null"`
	BindingType    BindingType        `json:"binding_type" gorm:"size:20;not null"`
	RoleID         *string            `json:"role_id" gorm:"type:varchar(255)"`
	PermissionID   *string            `json:"permission_id" gorm:"type:varchar(255)"`
	ResourceType   string             `json:"resource_type" gorm:"size:100;not null"`
	ResourceID     *string            `json:"resource_id" gorm:"type:varchar(255)"`
	OrganizationID string             `json:"organization_id" gorm:"type:varchar(255);not null"`
	Caveat         *Caveat            `json:"caveat" gorm:"type:jsonb"`
	Version        int                `json:"version" gorm:"not null"`
	Action         string             `json:"action" gorm:"size:20;not null"` // CREATE, UPDATE, DELETE
	ChangedByID    string             `json:"changed_by_id" gorm:"type:varchar(255);not null"`
	ChangedAt      time.Time          `json:"changed_at" gorm:"not null"`

	// Relationships
	Binding   *Binding `json:"binding" gorm:"foreignKey:BindingID;references:ID"`
	ChangedBy *User    `json:"changed_by" gorm:"foreignKey:ChangedByID;references:ID"`
}

func (b *Binding) TableName() string {
	return "bindings"
}

func (bh *BindingHistory) TableName() string {
	return "binding_history"
}

// NewBinding creates a new Binding instance
func NewBinding(subjectID string, subjectType BindingSubjectType, bindingType BindingType,
	resourceType string, organizationID string, createdByID string) *Binding {
	return &Binding{
		BaseModel:      base.NewBaseModel("BND", hash.Medium),
		SubjectID:      subjectID,
		SubjectType:    subjectType,
		BindingType:    bindingType,
		ResourceType:   resourceType,
		OrganizationID: organizationID,
		CreatedByID:    createdByID,
		Version:        1,
		IsActive:       true,
	}
}

// NewRoleBinding creates a new role binding
func NewRoleBinding(subjectID string, subjectType BindingSubjectType, roleID string,
	resourceType string, organizationID string, createdByID string) *Binding {
	binding := NewBinding(subjectID, subjectType, BindingTypeRole, resourceType, organizationID, createdByID)
	binding.RoleID = &roleID
	return binding
}

// NewPermissionBinding creates a new permission binding
func NewPermissionBinding(subjectID string, subjectType BindingSubjectType, permissionID string,
	resourceType string, organizationID string, createdByID string) *Binding {
	binding := NewBinding(subjectID, subjectType, BindingTypePermission, resourceType, organizationID, createdByID)
	binding.PermissionID = &permissionID
	return binding
}

// BeforeCreate hooks
func (b *Binding) BeforeCreate() error {
	if err := b.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	// Validate that either RoleID or PermissionID is set based on BindingType
	if b.BindingType == BindingTypeRole && b.RoleID == nil {
		return errors.New("role_id is required for role bindings")
	}
	if b.BindingType == BindingTypePermission && b.PermissionID == nil {
		return errors.New("permission_id is required for permission bindings")
	}

	return nil
}

func (bh *BindingHistory) BeforeCreate() error {
	if err := bh.BaseModel.BeforeCreate(); err != nil {
		return err
	}
	if bh.ChangedAt.IsZero() {
		bh.ChangedAt = time.Now()
	}
	return nil
}

// GORM Hooks
func (b *Binding) BeforeCreateGORM(tx *gorm.DB) error {
	return b.BeforeCreate()
}

func (bh *BindingHistory) BeforeCreateGORM(tx *gorm.DB) error {
	return bh.BeforeCreate()
}

// Helper methods
func (b *Binding) GetTableIdentifier() string   { return "BND" }
func (b *Binding) GetTableSize() hash.TableSize { return hash.Medium }

// Explicit method implementations to satisfy linter
func (b *Binding) GetID() string   { return b.BaseModel.GetID() }
func (b *Binding) SetID(id string) { b.BaseModel.SetID(id) }

func (bh *BindingHistory) GetTableIdentifier() string   { return "BNH" }
func (bh *BindingHistory) GetTableSize() hash.TableSize { return hash.Medium }

// AddCaveat adds or updates a caveat constraint
func (b *Binding) AddCaveat(key string, value interface{}) {
	if b.Caveat == nil {
		caveat := make(Caveat)
		b.Caveat = &caveat
	}
	(*b.Caveat)[key] = value
}

// GetCaveat retrieves a caveat value by key
func (b *Binding) GetCaveat(key string) (interface{}, bool) {
	if b.Caveat == nil {
		return nil, false
	}
	val, exists := (*b.Caveat)[key]
	return val, exists
}

// HasTimeCaveat checks if the binding has time-based caveats
func (b *Binding) HasTimeCaveat() bool {
	if b.Caveat == nil {
		return false
	}
	_, hasStart := (*b.Caveat)["starts_at"]
	_, hasEnd := (*b.Caveat)["ends_at"]
	return hasStart || hasEnd
}

// HasAttributeCaveat checks if the binding has attribute-based caveats
func (b *Binding) HasAttributeCaveat() bool {
	if b.Caveat == nil {
		return false
	}
	_, hasAttrs := (*b.Caveat)["required_attributes"]
	return hasAttrs
}

// HasColumnCaveat checks if the binding has column-based caveats
func (b *Binding) HasColumnCaveat() bool {
	if b.Caveat == nil {
		return false
	}
	_, hasColumns := (*b.Caveat)["column_groups"]
	return hasColumns
}

// CreateHistoryRecord creates a history record for this binding
func (b *Binding) CreateHistoryRecord(action string, changedByID string) *BindingHistory {
	return &BindingHistory{
		BaseModel:      base.NewBaseModel("bnh", hash.Medium),
		BindingID:      b.GetID(),
		SubjectID:      b.SubjectID,
		SubjectType:    b.SubjectType,
		BindingType:    b.BindingType,
		RoleID:         b.RoleID,
		PermissionID:   b.PermissionID,
		ResourceType:   b.ResourceType,
		ResourceID:     b.ResourceID,
		OrganizationID: b.OrganizationID,
		Caveat:         b.Caveat,
		Version:        b.Version,
		Action:         action,
		ChangedByID:    changedByID,
		ChangedAt:      time.Now(),
	}
}

// NewBindingHistory creates a new BindingHistory instance
func NewBindingHistory(bindingID string, action string, changedByID string) *BindingHistory {
	return &BindingHistory{
		BaseModel:   base.NewBaseModel("BNH", hash.Medium),
		BindingID:   bindingID,
		Action:      action,
		ChangedByID: changedByID,
		ChangedAt:   time.Now(),
	}
}
