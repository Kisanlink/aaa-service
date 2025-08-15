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

// AttributeSubjectType defines what entity the attribute is attached to
type AttributeSubjectType string

const (
	AttributeSubjectPrincipal AttributeSubjectType = "principal"
	AttributeSubjectResource  AttributeSubjectType = "resource"
	AttributeSubjectOrg       AttributeSubjectType = "organization"
)

// AttributeValue represents a JSON value for an attribute
type AttributeValue map[string]interface{}

// Scan implements the Scanner interface for database reads
func (av *AttributeValue) Scan(value interface{}) error {
	if value == nil {
		*av = make(AttributeValue)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, av)
	case string:
		return json.Unmarshal([]byte(v), av)
	default:
		return errors.New("cannot scan attribute value from database")
	}
}

// Value implements the Valuer interface for database writes
func (av AttributeValue) Value() (driver.Value, error) {
	if av == nil {
		return nil, nil
	}
	return json.Marshal(av)
}

// Attribute represents a key-value attribute for ABAC
type Attribute struct {
	*base.BaseModel
	SubjectID      string               `json:"subject_id" gorm:"type:varchar(255);not null"`
	SubjectType    AttributeSubjectType `json:"subject_type" gorm:"size:20;not null"`
	Key            string               `json:"key" gorm:"size:100;not null"`
	Value          AttributeValue       `json:"value" gorm:"type:jsonb;not null"`
	OrganizationID *string              `json:"organization_id" gorm:"type:varchar(255)"`
	ExpiresAt      *time.Time           `json:"expires_at" gorm:"type:timestamp"`
	IsActive       bool                 `json:"is_active" gorm:"default:true"`
	SetByID        string               `json:"set_by_id" gorm:"type:varchar(255);not null"`
	Metadata       *string              `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization *Organization `json:"organization" gorm:"foreignKey:OrganizationID;references:ID"`
	SetBy        *User         `json:"set_by" gorm:"foreignKey:SetByID;references:ID"`
}

// AttributeHistory tracks changes to attributes for audit
type AttributeHistory struct {
	*base.BaseModel
	AttributeID    string               `json:"attribute_id" gorm:"type:varchar(255);not null"`
	SubjectID      string               `json:"subject_id" gorm:"type:varchar(255);not null"`
	SubjectType    AttributeSubjectType `json:"subject_type" gorm:"size:20;not null"`
	Key            string               `json:"key" gorm:"size:100;not null"`
	OldValue       *AttributeValue      `json:"old_value" gorm:"type:jsonb"`
	NewValue       *AttributeValue      `json:"new_value" gorm:"type:jsonb"`
	Action         string               `json:"action" gorm:"size:20;not null"` // SET, DELETE
	ChangedByID    string               `json:"changed_by_id" gorm:"type:varchar(255);not null"`
	ChangedAt      time.Time            `json:"changed_at" gorm:"not null"`
	OrganizationID *string              `json:"organization_id" gorm:"type:varchar(255)"`

	// Relationships
	Attribute *Attribute `json:"attribute" gorm:"foreignKey:AttributeID;references:ID"`
	ChangedBy *User      `json:"changed_by" gorm:"foreignKey:ChangedByID;references:ID"`
}

func (a *Attribute) TableName() string {
	return "attributes"
}

// GetID returns the ID of the attribute
func (a *Attribute) GetID() string {
	return a.BaseModel.ID
}

func (ah *AttributeHistory) TableName() string {
	return "attribute_history"
}

// NewAttribute creates a new Attribute instance
func NewAttribute(subjectID string, subjectType AttributeSubjectType, key string, value AttributeValue, setByID string) *Attribute {
	return &Attribute{
		BaseModel:   base.NewBaseModel("att", hash.Medium),
		SubjectID:   subjectID,
		SubjectType: subjectType,
		Key:         key,
		Value:       value,
		SetByID:     setByID,
		IsActive:    true,
	}
}

// NewAttributeHistory creates a new AttributeHistory instance
func NewAttributeHistory(attributeID, subjectID string, subjectType AttributeSubjectType,
	key string, action string, changedByID string) *AttributeHistory {
	return &AttributeHistory{
		BaseModel:   base.NewBaseModel("ath", hash.Medium),
		AttributeID: attributeID,
		SubjectID:   subjectID,
		SubjectType: subjectType,
		Key:         key,
		Action:      action,
		ChangedByID: changedByID,
		ChangedAt:   time.Now(),
	}
}

// BeforeCreate hooks
func (a *Attribute) BeforeCreate() error {
	if err := a.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	if a.Value == nil {
		a.Value = make(AttributeValue)
	}

	return nil
}

func (ah *AttributeHistory) BeforeCreate() error {
	if err := ah.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	if ah.ChangedAt.IsZero() {
		ah.ChangedAt = time.Now()
	}

	return nil
}

// GORM Hooks
func (a *Attribute) BeforeCreateGORM(tx *gorm.DB) error {
	return a.BeforeCreate()
}

func (ah *AttributeHistory) BeforeCreateGORM(tx *gorm.DB) error {
	return ah.BeforeCreate()
}

// Helper methods
func (a *Attribute) GetTableIdentifier() string   { return "att" }
func (a *Attribute) GetTableSize() hash.TableSize { return hash.Medium }

func (ah *AttributeHistory) GetTableIdentifier() string   { return "ath" }
func (ah *AttributeHistory) GetTableSize() hash.TableSize { return hash.Medium }

// IsExpired checks if the attribute has expired
func (a *Attribute) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// IsEffective checks if the attribute is currently effective
func (a *Attribute) IsEffective() bool {
	return a.IsActive && !a.IsExpired()
}

// SetStringValue sets a string value for the attribute
func (a *Attribute) SetStringValue(value string) {
	if a.Value == nil {
		a.Value = make(AttributeValue)
	}
	a.Value["value"] = value
	a.Value["type"] = "string"
}

// SetNumberValue sets a numeric value for the attribute
func (a *Attribute) SetNumberValue(value float64) {
	if a.Value == nil {
		a.Value = make(AttributeValue)
	}
	a.Value["value"] = value
	a.Value["type"] = "number"
}

// SetBoolValue sets a boolean value for the attribute
func (a *Attribute) SetBoolValue(value bool) {
	if a.Value == nil {
		a.Value = make(AttributeValue)
	}
	a.Value["value"] = value
	a.Value["type"] = "boolean"
}

// SetListValue sets a list value for the attribute
func (a *Attribute) SetListValue(value []interface{}) {
	if a.Value == nil {
		a.Value = make(AttributeValue)
	}
	a.Value["value"] = value
	a.Value["type"] = "list"
}

// SetMapValue sets a map value for the attribute
func (a *Attribute) SetMapValue(value map[string]interface{}) {
	if a.Value == nil {
		a.Value = make(AttributeValue)
	}
	a.Value["value"] = value
	a.Value["type"] = "map"
}

// GetStringValue retrieves the attribute value as a string
func (a *Attribute) GetStringValue() (string, bool) {
	if a.Value == nil {
		return "", false
	}
	if val, ok := a.Value["value"].(string); ok {
		return val, true
	}
	return "", false
}

// GetNumberValue retrieves the attribute value as a number
func (a *Attribute) GetNumberValue() (float64, bool) {
	if a.Value == nil {
		return 0, false
	}
	if val, ok := a.Value["value"].(float64); ok {
		return val, true
	}
	return 0, false
}

// GetBoolValue retrieves the attribute value as a boolean
func (a *Attribute) GetBoolValue() (bool, bool) {
	if a.Value == nil {
		return false, false
	}
	if val, ok := a.Value["value"].(bool); ok {
		return val, true
	}
	return false, false
}

// CreateHistoryRecord creates a history record for this attribute
func (a *Attribute) CreateHistoryRecord(action string, oldValue *AttributeValue, changedByID string) *AttributeHistory {
	// TODO: Fix BaseModel ID access pattern
	history := NewAttributeHistory(a.GetID(), a.SubjectID, a.SubjectType, a.Key, action, changedByID)
	history.OldValue = oldValue
	history.NewValue = &a.Value
	history.OrganizationID = a.OrganizationID
	return history
}

// AttributeRegistry represents a collection of attributes for efficient lookup
type AttributeRegistry struct {
	Attributes map[string]map[string]AttributeValue // subjectID -> key -> value
}

// NewAttributeRegistry creates a new AttributeRegistry
func NewAttributeRegistry() *AttributeRegistry {
	return &AttributeRegistry{
		Attributes: make(map[string]map[string]AttributeValue),
	}
}

// Set adds or updates an attribute in the registry
func (ar *AttributeRegistry) Set(subjectID, key string, value AttributeValue) {
	if ar.Attributes[subjectID] == nil {
		ar.Attributes[subjectID] = make(map[string]AttributeValue)
	}
	ar.Attributes[subjectID][key] = value
}

// Get retrieves an attribute from the registry
func (ar *AttributeRegistry) Get(subjectID, key string) (AttributeValue, bool) {
	if subjectAttrs, ok := ar.Attributes[subjectID]; ok {
		if value, exists := subjectAttrs[key]; exists {
			return value, true
		}
	}
	return nil, false
}

// GetAll retrieves all attributes for a subject
func (ar *AttributeRegistry) GetAll(subjectID string) (map[string]AttributeValue, bool) {
	attrs, ok := ar.Attributes[subjectID]
	return attrs, ok
}

// HasKey checks if a subject has a specific attribute key
func (ar *AttributeRegistry) HasKey(subjectID, key string) bool {
	_, exists := ar.Get(subjectID, key)
	return exists
}

// Delete removes an attribute from the registry
func (ar *AttributeRegistry) Delete(subjectID, key string) {
	if subjectAttrs, ok := ar.Attributes[subjectID]; ok {
		delete(subjectAttrs, key)
		if len(subjectAttrs) == 0 {
			delete(ar.Attributes, subjectID)
		}
	}
}
