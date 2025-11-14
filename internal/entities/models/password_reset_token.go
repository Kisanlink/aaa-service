package models

import (
	"time"

	db "github.com/Kisanlink/kisanlink-db/pkg/base"
)

// PasswordResetToken represents a token for password reset
type PasswordResetToken struct {
	db.BaseModel
	UserID    string     `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	Used      bool       `gorm:"default:false" json:"used"`
	UsedAt    *time.Time `gorm:"default:null" json:"used_at,omitempty"`
}

// TableName specifies the table name for PasswordResetToken
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

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
