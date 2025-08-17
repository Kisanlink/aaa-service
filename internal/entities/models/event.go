package models

import (
	"crypto/sha256"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// EventKind defines the type of event
type EventKind string

const (
	// Organization events
	EventKindOrgCreated EventKind = "org.created"
	EventKindOrgUpdated EventKind = "org.updated"
	EventKindOrgDeleted EventKind = "org.deleted"

	// Group events
	EventKindGroupCreated        EventKind = "group.created"
	EventKindGroupUpdated        EventKind = "group.updated"
	EventKindGroupDeleted        EventKind = "group.deleted"
	EventKindGroupMemberAdded    EventKind = "group.member.added"
	EventKindGroupMemberRemoved  EventKind = "group.member.removed"
	EventKindGroupInheritanceSet EventKind = "group.inheritance.set"

	// Role events
	EventKindRoleCreated   EventKind = "role.created"
	EventKindRoleUpdated   EventKind = "role.updated"
	EventKindRoleDeleted   EventKind = "role.deleted"
	EventKindRoleVersioned EventKind = "role.versioned"

	// Permission events
	EventKindPermissionCreated EventKind = "permission.created"
	EventKindPermissionUpdated EventKind = "permission.updated"
	EventKindPermissionDeleted EventKind = "permission.deleted"

	// Binding events
	EventKindBindingCreated    EventKind = "binding.created"
	EventKindBindingUpdated    EventKind = "binding.updated"
	EventKindBindingDeleted    EventKind = "binding.deleted"
	EventKindBindingRolledBack EventKind = "binding.rolledback"

	// Resource events
	EventKindResourceCreated       EventKind = "resource.created"
	EventKindResourceUpdated       EventKind = "resource.updated"
	EventKindResourceDeleted       EventKind = "resource.deleted"
	EventKindResourceParentChanged EventKind = "resource.parent.changed"

	// Attribute events
	EventKindAttributeSet     EventKind = "attribute.set"
	EventKindAttributeDeleted EventKind = "attribute.deleted"

	// Contract events
	EventKindContractApplied  EventKind = "contract.applied"
	EventKindContractReverted EventKind = "contract.reverted"

	// Column events
	EventKindColumnGroupCreated EventKind = "columngroup.created"
	EventKindColumnGroupUpdated EventKind = "columngroup.updated"
	EventKindColumnGroupDeleted EventKind = "columngroup.deleted"
)

// EventPayload represents the JSON payload of an event
type EventPayload map[string]interface{}

// Scan implements the Scanner interface for database reads
func (ep *EventPayload) Scan(value interface{}) error {
	if value == nil {
		*ep = make(EventPayload)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ep)
	case string:
		return json.Unmarshal([]byte(v), ep)
	default:
		return errors.New("cannot scan event payload from database")
	}
}

// Value implements the Valuer interface for database writes
func (ep EventPayload) Value() (driver.Value, error) {
	if ep == nil {
		return nil, nil
	}
	return json.Marshal(ep)
}

// Event represents an immutable audit event in the system
type Event struct {
	*base.BaseModel
	OccurredAt     time.Time    `json:"occurred_at" gorm:"not null;index"`
	ActorID        string       `json:"actor_id" gorm:"type:varchar(255);not null"`
	ActorType      string       `json:"actor_type" gorm:"size:20;not null"` // user, service, system
	Kind           EventKind    `json:"kind" gorm:"size:50;not null;index"`
	ResourceType   string       `json:"resource_type" gorm:"size:100;not null;index"`
	ResourceID     string       `json:"resource_id" gorm:"type:varchar(255);not null;index"`
	OrganizationID *string      `json:"organization_id" gorm:"type:varchar(255);index"`
	Payload        EventPayload `json:"payload" gorm:"type:jsonb;not null"`
	PrevHash       *string      `json:"prev_hash" gorm:"type:varchar(64)"`            // SHA256 hex
	Hash           string       `json:"hash" gorm:"type:varchar(64);not null;unique"` // SHA256 hex
	SequenceNum    int64        `json:"sequence_num" gorm:"not null;unique"`

	// Metadata for debugging/tracing
	RequestID *string `json:"request_id" gorm:"type:varchar(255);index"`
	SourceIP  *string `json:"source_ip" gorm:"size:45"` // Support IPv6
	UserAgent *string `json:"user_agent" gorm:"type:text"`
}

func (e *Event) TableName() string {
	return "events"
}

// NewEvent creates a new Event instance
func NewEvent(actorID, actorType string, kind EventKind, resourceType, resourceID string, payload EventPayload) *Event {
	return &Event{
		BaseModel:    base.NewBaseModel("evt", hash.Large),
		OccurredAt:   time.Now(),
		ActorID:      actorID,
		ActorType:    actorType,
		Kind:         kind,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Payload:      payload,
	}
}

// BeforeCreate hooks
func (e *Event) BeforeCreate() error {
	if err := e.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	// OccurredAt must be set
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now()
	}

	// Payload cannot be nil
	if e.Payload == nil {
		e.Payload = make(EventPayload)
	}

	return nil
}

// GORM Hooks
func (e *Event) BeforeCreateGORM(tx *gorm.DB) error {
	return e.BeforeCreate()
}

// Helper methods
func (e *Event) GetTableIdentifier() string   { return "EVT" }
func (e *Event) GetTableSize() hash.TableSize { return hash.Medium }

// Explicit method implementations to satisfy linter
func (e *Event) GetID() string   { return e.BaseModel.GetID() }
func (e *Event) SetID(id string) { e.BaseModel.SetID(id) }

// ComputeHash calculates the SHA256 hash for this event
func (e *Event) ComputeHash() (string, error) {
	// Create a deterministic JSON representation
	hashData := map[string]interface{}{
		"id":            e.GetID(),
		"occurred_at":   e.OccurredAt.UTC().Format(time.RFC3339Nano),
		"actor_id":      e.ActorID,
		"actor_type":    e.ActorType,
		"kind":          string(e.Kind),
		"resource_type": e.ResourceType,
		"resource_id":   e.ResourceID,
		"payload":       e.Payload,
		"sequence_num":  e.SequenceNum,
	}

	if e.OrganizationID != nil {
		hashData["organization_id"] = *e.OrganizationID
	}

	if e.PrevHash != nil {
		hashData["prev_hash"] = *e.PrevHash
	}

	// Marshal to JSON with sorted keys
	jsonData, err := json.Marshal(hashData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Calculate SHA256
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash), nil
}

// SetHash computes and sets the hash for this event
func (e *Event) SetHash() error {
	hash, err := e.ComputeHash()
	if err != nil {
		return err
	}
	e.Hash = hash
	return nil
}

// VerifyHash verifies the hash integrity of this event
func (e *Event) VerifyHash() error {
	expectedHash, err := e.ComputeHash()
	if err != nil {
		return err
	}

	if e.Hash != expectedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, e.Hash)
	}

	return nil
}

// AddPayloadField adds a field to the event payload
func (e *Event) AddPayloadField(key string, value interface{}) {
	if e.Payload == nil {
		e.Payload = make(EventPayload)
	}
	e.Payload[key] = value
}

// GetPayloadField retrieves a field from the event payload
func (e *Event) GetPayloadField(key string) (interface{}, bool) {
	if e.Payload == nil {
		return nil, false
	}
	val, exists := e.Payload[key]
	return val, exists
}

// EventCheckpoint represents a periodic checkpoint of the event chain
type EventCheckpoint struct {
	*base.BaseModel
	CheckpointTime  time.Time `json:"checkpoint_time" gorm:"not null"`
	LastEventID     string    `json:"last_event_id" gorm:"type:varchar(255);not null"`
	LastSequenceNum int64     `json:"last_sequence_num" gorm:"not null"`
	LastEventHash   string    `json:"last_event_hash" gorm:"type:varchar(64);not null"`
	MerkleRoot      string    `json:"merkle_root" gorm:"type:varchar(64);not null"` // Merkle tree root of all events up to this point
	EventCount      int64     `json:"event_count" gorm:"not null"`
	CreatedByID     string    `json:"created_by_id" gorm:"type:varchar(255);not null"`

	// Relationships
	LastEvent *Event `json:"last_event" gorm:"foreignKey:LastEventID;references:ID"`
	CreatedBy *User  `json:"created_by" gorm:"foreignKey:CreatedByID;references:ID"`
}

func (ec *EventCheckpoint) TableName() string {
	return "event_checkpoints"
}

// NewEventCheckpoint creates a new EventCheckpoint instance
func NewEventCheckpoint(lastEventID string, lastSequenceNum int64, lastEventHash string, merkleRoot string, eventCount int64, createdByID string) *EventCheckpoint {
	return &EventCheckpoint{
		BaseModel:       base.NewBaseModel("EVENT", hash.Small),
		CheckpointTime:  time.Now(),
		LastEventID:     lastEventID,
		LastSequenceNum: lastSequenceNum,
		LastEventHash:   lastEventHash,
		MerkleRoot:      merkleRoot,
		EventCount:      eventCount,
		CreatedByID:     createdByID,
	}
}

// BeforeCreate hooks
func (ec *EventCheckpoint) BeforeCreate() error {
	if err := ec.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	if ec.CheckpointTime.IsZero() {
		ec.CheckpointTime = time.Now()
	}

	return nil
}

// GORM Hooks
func (ec *EventCheckpoint) BeforeCreateGORM(tx *gorm.DB) error {
	return ec.BeforeCreate()
}

// Helper methods
func (ec *EventCheckpoint) GetTableIdentifier() string   { return "evc" }
func (ec *EventCheckpoint) GetTableSize() hash.TableSize { return hash.Small }
