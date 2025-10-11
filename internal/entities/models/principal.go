package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// PrincipalType defines the type of principal
type PrincipalType string

const (
	PrincipalTypeUser    PrincipalType = "user"
	PrincipalTypeService PrincipalType = "service"
)

// Principal represents a unified identity (user or service) in the system
type Principal struct {
	*base.BaseModel
	Type           PrincipalType `json:"type" gorm:"size:20;not null"`
	UserID         *string       `json:"user_id" gorm:"type:varchar(255);uniqueIndex"`    // When type is user
	ServiceID      *string       `json:"service_id" gorm:"type:varchar(255);uniqueIndex"` // When type is service
	Name           string        `json:"name" gorm:"size:100;not null"`
	OrganizationID *string       `json:"organization_id" gorm:"type:varchar(255);index"`
	IsActive       bool          `json:"is_active" gorm:"default:true"`
	Metadata       *string       `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	User         *User         `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Organization *Organization `json:"organization" gorm:"foreignKey:OrganizationID;references:ID"`
}

// Service represents a service account in the system
type Service struct {
	*base.BaseModel
	Name           string  `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description    string  `json:"description" gorm:"type:text"`
	OrganizationID string  `json:"organization_id" gorm:"type:varchar(255);not null"`
	APIKey         string  `json:"api_key" gorm:"size:255;not null;unique"` // Hashed
	IsActive       bool    `json:"is_active" gorm:"default:true"`
	Metadata       *string `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization *Organization `json:"organization" gorm:"foreignKey:OrganizationID;references:ID"`
	Principal    *Principal    `json:"principal" gorm:"foreignKey:ServiceID;references:ID"`
}

func (p *Principal) TableName() string {
	return "principals"
}

func (s *Service) TableName() string {
	return "services"
}

// NewPrincipal creates a new Principal instance
func NewPrincipal(principalType PrincipalType, name string) *Principal {
	return &Principal{
		BaseModel: base.NewBaseModel("PRN", hash.Medium),
		Type:      principalType,
		Name:      name,
		IsActive:  true,
	}
}

// NewUserPrincipal creates a new user-type Principal
func NewUserPrincipal(userID, name string) *Principal {
	principal := NewPrincipal(PrincipalTypeUser, name)
	principal.UserID = &userID
	return principal
}

// NewServicePrincipal creates a new service-type Principal
func NewServicePrincipal(serviceID, name string) *Principal {
	principal := NewPrincipal(PrincipalTypeService, name)
	principal.ServiceID = &serviceID
	return principal
}

// NewService creates a new Service instance
func NewService(name, description, organizationID, hashedAPIKey string) *Service {
	return &Service{
		BaseModel:      base.NewBaseModel("SVC", hash.Small),
		Name:           name,
		Description:    description,
		OrganizationID: organizationID,
		APIKey:         hashedAPIKey,
		IsActive:       true,
	}
}

// BeforeCreate hooks
func (p *Principal) BeforeCreate() error {
	if err := p.BaseModel.BeforeCreate(); err != nil {
		return err
	}

	// Validate that either UserID or ServiceID is set based on Type
	if p.Type == PrincipalTypeUser && p.UserID == nil {
		return gorm.ErrInvalidData
	}
	if p.Type == PrincipalTypeService && p.ServiceID == nil {
		return gorm.ErrInvalidData
	}

	return nil
}

func (s *Service) BeforeCreate() error {
	return s.BaseModel.BeforeCreate()
}

// GORM Hooks
func (p *Principal) BeforeCreateGORM(tx *gorm.DB) error {
	return p.BeforeCreate()
}

func (s *Service) BeforeCreateGORM(tx *gorm.DB) error {
	return s.BeforeCreate()
}

// Helper methods
func (p *Principal) GetTableIdentifier() string   { return "PRN" }
func (p *Principal) GetTableSize() hash.TableSize { return hash.Medium }

// Explicit method implementations to satisfy linter
func (p *Principal) GetID() string   { return p.BaseModel.GetID() }
func (p *Principal) SetID(id string) { p.BaseModel.SetID(id) }

func (s *Service) GetTableIdentifier() string   { return "SVC" }
func (s *Service) GetTableSize() hash.TableSize { return hash.Small }

// Explicit method implementations to satisfy linter
func (s *Service) GetID() string   { return s.BaseModel.GetID() }
func (s *Service) SetID(id string) { s.BaseModel.SetID(id) }

// GetSubjectType returns the subject type for PostgreSQL RBAC
func (p *Principal) GetSubjectType() string {
	return "aaa/principal"
}

// GetSubjectID returns the subject ID for PostgreSQL RBAC
func (p *Principal) GetSubjectID() string {
	return p.GetID()
}

// GetResourceType returns the resource type for services in PostgreSQL RBAC
func (s *Service) GetResourceType() string {
	return "aaa/service"
}

// GetObjectID returns the object ID for this service in PostgreSQL RBAC
func (s *Service) GetObjectID() string {
	return s.GetID()
}
