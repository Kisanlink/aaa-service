package requests

// Swagger Documentation Request Models for AAA Service API
// These models are specifically designed for comprehensive API documentation

// RefreshTokenRequestSwagger represents refresh token request for Swagger docs
// @Description Request structure for token refresh
type RefreshTokenRequestSwagger struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	MPin         string `json:"mpin" binding:"required" example:"1234"`
}

// UpdateUserRequest represents user update request
// @Description Request structure for updating user information
type UpdateUserRequest struct {
	Username    string `json:"username,omitempty" example:"john_doe"`
	PhoneNumber string `json:"phone_number,omitempty" example:"+1234567890"`
	CountryCode string `json:"country_code,omitempty" example:"US"`
}

// CreateRoleRequest represents role creation request
// @Description Request structure for creating a new role
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required" example:"moderator"`
	Description string `json:"description,omitempty" example:"Moderator role with limited admin access"`
}

// UpdateRoleRequest represents role update request
// @Description Request structure for updating role information
type UpdateRoleRequest struct {
	Name        string `json:"name,omitempty" example:"moderator"`
	Description string `json:"description,omitempty" example:"Updated moderator role description"`
}

// CreatePermissionRequest represents permission creation request
// @Description Request structure for creating a new permission
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required" example:"users:write"`
	Description string `json:"description,omitempty" example:"Write access to user resources"`
	Resource    string `json:"resource" binding:"required" example:"user"`
	Action      string `json:"action" binding:"required" example:"write"`
}

// UpdatePermissionRequest represents permission update request
// @Description Request structure for updating permission information
type UpdatePermissionRequest struct {
	Name        string `json:"name,omitempty" example:"users:write"`
	Description string `json:"description,omitempty" example:"Updated write access to user resources"`
	Resource    string `json:"resource,omitempty" example:"user"`
	Action      string `json:"action,omitempty" example:"write"`
}

// GrantPermissionRequest represents permission grant request
// @Description Request structure for granting permission to user
type GrantPermissionRequest struct {
	UserID     string `json:"user_id" binding:"required" example:"USER123456789"`
	Resource   string `json:"resource" binding:"required" example:"user"`
	ResourceID string `json:"resource_id" binding:"required" example:"USER987654321"`
	Relation   string `json:"relation" binding:"required" example:"owner"`
}

// RevokePermissionRequest represents permission revoke request
// @Description Request structure for revoking permission from user
type RevokePermissionRequest struct {
	UserID     string `json:"user_id" binding:"required" example:"USER123456789"`
	Resource   string `json:"resource" binding:"required" example:"user"`
	ResourceID string `json:"resource_id" binding:"required" example:"USER987654321"`
	Relation   string `json:"relation" binding:"required" example:"owner"`
}

// AssignRoleRequest represents role assignment request
// @Description Request structure for assigning role to user
type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required" example:"USER123456789"`
	RoleID string `json:"role_id" binding:"required" example:"ROLE123456789"`
}

// RemoveRoleRequest represents role removal request
// @Description Request structure for removing role from user
type RemoveRoleRequest struct {
	UserID string `json:"user_id" binding:"required" example:"USER123456789"`
	RoleID string `json:"role_id" binding:"required" example:"ROLE123456789"`
}

// ArchiveLogsRequest represents log archiving request
// @Description Request structure for archiving old logs
type ArchiveLogsRequest struct {
	Days int `json:"days" binding:"required,min=1" example:"90"`
}

// BulkPermissionCheckRequest represents bulk permission check request
// @Description Request structure for checking multiple permissions
type BulkPermissionCheckRequest struct {
	UserID      string                `json:"user_id" binding:"required" example:"USER123456789"`
	Permissions []PermissionCheckItem `json:"permissions" binding:"required,min=1"`
}

// PermissionCheckItem represents individual permission check item
// @Description Individual permission check item structure
type PermissionCheckItem struct {
	Resource   string `json:"resource" binding:"required" example:"user"`
	Action     string `json:"action" binding:"required" example:"read"`
	ResourceID string `json:"resource_id,omitempty" example:"USER987654321"`
}

// ModuleRegistrationRequest represents module registration request
// @Description Request structure for registering a new module
type ModuleRegistrationRequest struct {
	ServiceName string                     `json:"service_name" binding:"required" example:"user-service"`
	Version     string                     `json:"version" binding:"required" example:"1.0.0"`
	Actions     []ModuleActionDefinition   `json:"actions,omitempty"`
	Resources   []ModuleResourceDefinition `json:"resources,omitempty"`
	Roles       []ModuleRoleDefinition     `json:"roles,omitempty"`
}

// ModuleActionDefinition represents action definition for module
// @Description Action definition structure for module registration
type ModuleActionDefinition struct {
	Name        string `json:"name" binding:"required" example:"create_user"`
	Description string `json:"description,omitempty" example:"Create a new user"`
	Resource    string `json:"resource" binding:"required" example:"user"`
}

// ModuleResourceDefinition represents resource definition for module
// @Description Resource definition structure for module registration
type ModuleResourceDefinition struct {
	Name        string   `json:"name" binding:"required" example:"user"`
	Description string   `json:"description,omitempty" example:"User resource"`
	Actions     []string `json:"actions,omitempty" example:"[\"create\", \"read\", \"update\", \"delete\"]"`
}

// ModuleRoleDefinition represents role definition for module
// @Description Role definition structure for module registration
type ModuleRoleDefinition struct {
	Name        string   `json:"name" binding:"required" example:"user_admin"`
	Description string   `json:"description,omitempty" example:"User administrator role"`
	Permissions []string `json:"permissions,omitempty" example:"[\"user:create\", \"user:read\", \"user:update\", \"user:delete\"]"`
}
