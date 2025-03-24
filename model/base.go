package model

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Base struct {
    ID        string    `gorm:"type:varchar(36);primaryKey"` // Use string for ID
    CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
    UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// BeforeCreate hook to generate a unique ID
func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
    if b.ID == "" { // Check if ID is empty
        b.ID = cuid.New() // Generate a unique ID using cuid
    }
    return
}
