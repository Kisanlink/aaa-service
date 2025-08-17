package actions

import (
	"time"
)

// ActionResponse represents an action in API responses
type ActionResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	IsStatic    bool       `json:"is_static"`
	ServiceID   *string    `json:"service_id,omitempty"`
	Metadata    *string    `json:"metadata,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// ActionListResponse represents a paginated list of actions
type ActionListResponse struct {
	Actions []*ActionResponse `json:"actions"`
	Total   int64             `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
}
