package contacts

import (
	"time"
)

// ContactResponse represents a contact in API responses
type ContactResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Type        string     `json:"type"`
	Value       string     `json:"value"`
	Description *string    `json:"description,omitempty"`
	IsPrimary   bool       `json:"is_primary"`
	IsActive    bool       `json:"is_active"`
	IsVerified  bool       `json:"is_verified"`
	VerifiedAt  *string    `json:"verified_at,omitempty"`
	VerifiedBy  *string    `json:"verified_by,omitempty"`
	CountryCode *string    `json:"country_code,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// ContactListResponse represents a paginated list of contacts
type ContactListResponse struct {
	Contacts []*ContactResponse `json:"contacts"`
	Total    int64              `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
}
