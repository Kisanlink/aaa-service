package models

import (
	"fmt"
	"time"

	db "github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// PasswordResetToken represents a token for password reset
type PasswordResetToken struct {
	db.BaseModel
	UserID    string     `gorm:"type:varchar(255);not null;index" json:"user_id"`
	Token     string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	Used      bool       `gorm:"default:false" json:"used"`
	UsedAt    *time.Time `gorm:"default:null" json:"used_at,omitempty"`
}

// TableName specifies the table name for PasswordResetToken
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// BeforeCreate is the GORM hook that generates ID if not set
func (p *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if err := p.BaseModel.BeforeCreate(); err != nil {
		return err
	}
	// Generate ID with PRTOK prefix if not already set
	if p.GetID() == "" {
		p.SetID(fmt.Sprintf("PRTOK%d", time.Now().UnixNano()))
	}
	return nil
}

// SetID sets the ID
func (p *PasswordResetToken) SetID(id string) { p.BaseModel.SetID(id) }

// GetID returns the ID
func (p *PasswordResetToken) GetID() string { return p.BaseModel.GetID() }

// IsExpired checks if the token has expired
func (p *PasswordResetToken) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// IsValid checks if the token is valid (not used and not expired)
func (p *PasswordResetToken) IsValid() bool {
	return !p.Used && !p.IsExpired()
}

// MarkAsUsed marks the token as used
func (p *PasswordResetToken) MarkAsUsed() {
	p.Used = true
	now := time.Now()
	p.UsedAt = &now
}
