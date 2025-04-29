package model

import (
	"time"

	"github.com/google/uuid"
	// "github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        string    `gorm:"type:varchar(36);primaryKey"` // Use string for ID
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

//	func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
//	    if b.ID == "" {
//	        b.ID = cuid.New()
//	    }
//	    return
//	}
func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return
}
