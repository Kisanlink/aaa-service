package routes

import "time"

// SwaggerUser represents a user for Swagger documentation (simplified version)
type SwaggerUser struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	IsValidated bool       `json:"is_validated"`
	Status      *string    `json:"status,omitempty"`
	Tokens      int        `json:"tokens"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// SwaggerLoginResponse represents a login response for Swagger documentation
type SwaggerLoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int64       `json:"expires_in"`
	User         SwaggerUser `json:"user"`
	Permissions  []string    `json:"permissions"`
}

// SwaggerRegisterRequest represents a registration request for Swagger documentation
type SwaggerRegisterRequest struct {
	Username string   `json:"username" validate:"required,min=3,max=50"`
	Email    string   `json:"email" validate:"required,email"`
	FullName string   `json:"full_name" validate:"required,min=1,max=100"`
	Password string   `json:"password" validate:"required,min=8"`
	RoleIDs  []string `json:"role_ids,omitempty"`
}

// SwaggerLoginRequest represents a login request for Swagger documentation
type SwaggerLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

// SwaggerRefreshTokenRequest represents a refresh token request for Swagger documentation
type SwaggerRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// SwaggerAuditLog represents an audit log for Swagger documentation (simplified version)
type SwaggerAuditLog struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Status       string                 `json:"status"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details"`
	Timestamp    time.Time              `json:"timestamp"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// SwaggerAuditQueryResult represents audit query results for Swagger documentation
type SwaggerAuditQueryResult struct {
	Logs       []SwaggerAuditLog `json:"logs"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

// SwaggerPermission represents a permission for Swagger documentation
type SwaggerPermission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	IsActive    bool   `json:"is_active"`
}

// SwaggerPermissionResult represents a permission check result for Swagger documentation
type SwaggerPermissionResult struct {
	Allowed    bool   `json:"allowed"`
	Permission string `json:"permission"`
	Resource   string `json:"resource"`
	UserID     string `json:"user_id"`
	Reason     string `json:"reason,omitempty"`
}
