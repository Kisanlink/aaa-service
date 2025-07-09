package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

type UserProfile struct {
	*base.BaseModel
	UserID        string  `json:"user_id" gorm:"type:varchar(255);not null;uniqueIndex"`
	Name          *string `json:"name" gorm:"type:varchar(255)"`
	CareOf        *string `json:"care_of" gorm:"type:varchar(255)"`
	DateOfBirth   *string `json:"date_of_birth" gorm:"type:varchar(10)"`
	Photo         *string `json:"photo" gorm:"type:text"`
	YearOfBirth   *string `json:"year_of_birth" gorm:"type:varchar(4)"`
	Message       *string `json:"message" gorm:"type:text"`
	AadhaarNumber *string `json:"aadhaar_number" gorm:"type:varchar(12)"`
	EmailHash     *string `json:"email_hash" gorm:"type:varchar(255)"`
	ShareCode     *string `json:"share_code" gorm:"type:varchar(50)"`
	AddressID     *string `json:"address_id" gorm:"type:varchar(255)"`

	// Relationships
	Address Address `json:"address" gorm:"foreignKey:AddressID;references:ID"`
}

// NewUserProfile creates a new UserProfile instance
func NewUserProfile(userID string) *UserProfile {
	return &UserProfile{
		BaseModel: base.NewBaseModel("profile", hash.Small),
		UserID:    userID,
	}
}

func (p *UserProfile) BeforeCreate() error          { return p.BaseModel.BeforeCreate() }
func (p *UserProfile) BeforeUpdate() error          { return p.BaseModel.BeforeUpdate() }
func (p *UserProfile) BeforeDelete() error          { return p.BaseModel.BeforeDelete() }
func (p *UserProfile) BeforeSoftDelete() error      { return p.BaseModel.BeforeSoftDelete() }
func (p *UserProfile) GetTableIdentifier() string   { return "usr_prof" }
func (p *UserProfile) GetTableSize() hash.TableSize { return hash.Small }
