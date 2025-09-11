package responses

import "time"

// Swagger Documentation Models for AAA Service API
// These models are specifically designed for comprehensive API documentation

// Standard Response Wrapper Models

// SuccessResponse represents a standardized success response wrapper
//
//	@Description	Standard success response structure
type SuccessResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message" example:"Operation completed successfully"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string      `json:"request_id,omitempty" example:"req-123456789"`
}

// ErrorResponseSwagger represents a standardized error response for Swagger
//
//	@Description	Standard error response structure
type ErrorResponseSwagger struct {
	Success   bool                   `json:"success" example:"false"`
	Error     string                 `json:"error" example:"VALIDATION_ERROR"`
	Message   string                 `json:"message" example:"Invalid input data"`
	Code      int                    `json:"code,omitempty" example:"400"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string                 `json:"request_id,omitempty" example:"req-123456789"`
}

// Authentication Response Models

// LoginSuccessResponse represents successful login response
//
//	@Description	Successful login response with tokens and user info
type LoginSuccessResponse struct {
	Success   bool              `json:"success" example:"true"`
	Message   string            `json:"message" example:"Login successful"`
	Data      LoginResponseData `json:"data"`
	Timestamp time.Time         `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string            `json:"request_id,omitempty" example:"req-123456789"`
}

// LoginResponseData represents the data section of login response
//
//	@Description	Login response data containing tokens and user info
type LoginResponseData struct {
	AccessToken  string   `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string   `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string   `json:"token_type" example:"Bearer"`
	ExpiresIn    int64    `json:"expires_in" example:"86400"`
	User         UserInfo `json:"user"`
}

// RegisterSuccessResponse represents successful registration response
//
//	@Description	Successful registration response with user info
type RegisterSuccessResponse struct {
	Success   bool                 `json:"success" example:"true"`
	Message   string               `json:"message" example:"User registered successfully"`
	Data      RegisterResponseData `json:"data"`
	Timestamp time.Time            `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string               `json:"request_id,omitempty" example:"req-123456789"`
}

// RegisterResponseData represents the data section of registration response
//
//	@Description	Registration response data containing user info
type RegisterResponseData struct {
	User UserInfoSwagger `json:"user"`
}

// RefreshTokenSuccessResponse represents successful token refresh response
//
//	@Description	Successful token refresh response
type RefreshTokenSuccessResponse struct {
	Success   bool                     `json:"success" example:"true"`
	Message   string                   `json:"message" example:"Token refreshed successfully"`
	Data      RefreshTokenResponseData `json:"data"`
	Timestamp time.Time                `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string                   `json:"request_id,omitempty" example:"req-123456789"`
}

// RefreshTokenResponseData represents the data section of refresh token response
//
//	@Description	Refresh token response data
type RefreshTokenResponseData struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in" example:"86400"`
}

// LogoutSuccessResponse represents successful logout response
//
//	@Description	Successful logout response
type LogoutSuccessResponse struct {
	Success   bool      `json:"success" example:"true"`
	Message   string    `json:"message" example:"Logout successful"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string    `json:"request_id,omitempty" example:"req-123456789"`
}

// MPIN Management Response Models

// SetMPinSuccessResponse represents successful MPIN set response
//
//	@Description	Successful MPIN set response
type SetMPinSuccessResponse struct {
	Success   bool      `json:"success" example:"true"`
	Message   string    `json:"message" example:"MPIN set successfully"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string    `json:"request_id,omitempty" example:"req-123456789"`
}

// UpdateMPinSuccessResponse represents successful MPIN update response
//
//	@Description	Successful MPIN update response
type UpdateMPinSuccessResponse struct {
	Success   bool      `json:"success" example:"true"`
	Message   string    `json:"message" example:"MPIN updated successfully"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string    `json:"request_id,omitempty" example:"req-123456789"`
}

// User Management Response Models

// UserInfoSwagger represents user information in responses for Swagger docs
//
//	@Description	Enhanced user information structure with roles and profile
type UserInfoSwagger struct {
	ID          string               `json:"id" example:"USER123456789"`
	Username    string               `json:"username" example:"john_doe"`
	PhoneNumber string               `json:"phone_number" example:"+1234567890"`
	CountryCode string               `json:"country_code" example:"US"`
	IsValidated bool                 `json:"is_validated" example:"true"`
	Status      string               `json:"status,omitempty" example:"active"`
	CreatedAt   time.Time            `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time            `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	Tokens      int                  `json:"tokens" example:"100"`
	HasMPin     bool                 `json:"has_mpin" example:"true"`
	Roles       []UserRoleDetailInfo `json:"roles,omitempty"`
	Profile     *UserProfileInfo     `json:"profile,omitempty"`
	Contacts    []ContactInfo        `json:"contacts,omitempty"`
}

// UserRoleDetailInfo represents user role detail information
//
//	@Description	User role detail information structure
type UserRoleDetailInfo struct {
	ID       string     `json:"id" example:"USERROLE123456789"`
	UserID   string     `json:"user_id" example:"USER123456789"`
	RoleID   string     `json:"role_id" example:"ROLE123456789"`
	Role     RoleDetail `json:"role"`
	IsActive bool       `json:"is_active" example:"true"`
}

// UserDetailResponse represents detailed user response
//
//	@Description	Detailed user response structure
type UserDetailResponse struct {
	Success   bool           `json:"success" example:"true"`
	Message   string         `json:"message" example:"User retrieved successfully"`
	Data      UserDetailData `json:"data"`
	Timestamp time.Time      `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string         `json:"request_id,omitempty" example:"req-123456789"`
}

// UserDetailData represents detailed user data
//
//	@Description	Detailed user data structure
type UserDetailData struct {
	User UserInfoSwagger `json:"user"`
}

// UsersListResponse represents list of users response
//
//	@Description	Users list response with pagination
type UsersListResponse struct {
	Success    bool           `json:"success" example:"true"`
	Message    string         `json:"message" example:"Users retrieved successfully"`
	Data       UsersListData  `json:"data"`
	Pagination PaginationInfo `json:"pagination,omitempty"`
	Timestamp  time.Time      `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID  string         `json:"request_id,omitempty" example:"req-123456789"`
}

// UsersListData represents users list data
//
//	@Description	Users list data structure
type UsersListData struct {
	Users []UserInfoSwagger `json:"users"`
}

// Role Management Response Models

// RoleInfo represents role information
//
//	@Description	Role information structure
type RoleInfo struct {
	ID          string    `json:"id" example:"ROLE123456789"`
	Name        string    `json:"name" example:"admin"`
	Description string    `json:"description" example:"Administrator role with full access"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// RoleDetailResponse represents detailed role response
//
//	@Description	Detailed role response structure
type RoleDetailResponse struct {
	Success   bool           `json:"success" example:"true"`
	Message   string         `json:"message" example:"Role retrieved successfully"`
	Data      RoleDetailData `json:"data"`
	Timestamp time.Time      `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string         `json:"request_id,omitempty" example:"req-123456789"`
}

// RoleDetailData represents detailed role data
//
//	@Description	Detailed role data structure
type RoleDetailData struct {
	Role RoleInfo `json:"role"`
}

// RolesListResponse represents list of roles response
//
//	@Description	Roles list response structure
type RolesListResponse struct {
	Success    bool           `json:"success" example:"true"`
	Message    string         `json:"message" example:"Roles retrieved successfully"`
	Data       RolesListData  `json:"data"`
	Pagination PaginationInfo `json:"pagination,omitempty"`
	Timestamp  time.Time      `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID  string         `json:"request_id,omitempty" example:"req-123456789"`
}

// RolesListData represents roles list data
//
//	@Description	Roles list data structure
type RolesListData struct {
	Roles []RoleInfo `json:"roles"`
}

// Permission Management Response Models

// PermissionInfo represents permission information
//
//	@Description	Permission information structure
type PermissionInfo struct {
	ID          string    `json:"id" example:"PERM123456789"`
	Name        string    `json:"name" example:"users:read"`
	Description string    `json:"description" example:"Read access to user resources"`
	Resource    string    `json:"resource" example:"user"`
	Action      string    `json:"action" example:"read"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// PermissionDetailResponse represents detailed permission response
//
//	@Description	Detailed permission response structure
type PermissionDetailResponse struct {
	Success   bool                 `json:"success" example:"true"`
	Message   string               `json:"message" example:"Permission retrieved successfully"`
	Data      PermissionDetailData `json:"data"`
	Timestamp time.Time            `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string               `json:"request_id,omitempty" example:"req-123456789"`
}

// PermissionDetailData represents detailed permission data
//
//	@Description	Detailed permission data structure
type PermissionDetailData struct {
	Permission PermissionInfo `json:"permission"`
}

// PermissionsListResponse represents list of permissions response
//
//	@Description	Permissions list response structure
type PermissionsListResponse struct {
	Success    bool                `json:"success" example:"true"`
	Message    string              `json:"message" example:"Permissions retrieved successfully"`
	Data       PermissionsListData `json:"data"`
	Pagination PaginationInfo      `json:"pagination,omitempty"`
	Timestamp  time.Time           `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID  string              `json:"request_id,omitempty" example:"req-123456789"`
}

// PermissionsListData represents permissions list data
//
//	@Description	Permissions list data structure
type PermissionsListData struct {
	Permissions []PermissionInfo `json:"permissions"`
}

// Authorization Response Models

// PermissionCheckResponse represents permission check response
//
//	@Description	Permission check response structure
type PermissionCheckResponse struct {
	Success   bool                `json:"success" example:"true"`
	Message   string              `json:"message" example:"Permission check completed"`
	Data      PermissionCheckData `json:"data"`
	Timestamp time.Time           `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string              `json:"request_id,omitempty" example:"req-123456789"`
}

// PermissionCheckData represents permission check result data
//
//	@Description	Permission check result data structure
type PermissionCheckData struct {
	Allowed    bool   `json:"allowed" example:"true"`
	UserID     string `json:"user_id" example:"USER123456789"`
	Resource   string `json:"resource" example:"user"`
	Action     string `json:"action" example:"read"`
	ResourceID string `json:"resource_id,omitempty" example:"USER987654321"`
}

// UserPermissionsResponse represents user permissions response
//
//	@Description	User permissions response structure
type UserPermissionsResponse struct {
	Success   bool                `json:"success" example:"true"`
	Message   string              `json:"message" example:"User permissions retrieved successfully"`
	Data      UserPermissionsData `json:"data"`
	Timestamp time.Time           `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string              `json:"request_id,omitempty" example:"req-123456789"`
}

// UserPermissionsData represents user permissions data
//
//	@Description	User permissions data structure
type UserPermissionsData struct {
	UserID      string           `json:"user_id" example:"USER123456789"`
	Permissions []PermissionInfo `json:"permissions"`
}

// Audit Response Models

// AuditLogEntry represents a single audit log entry
//
//	@Description	Audit log entry structure
type AuditLogEntry struct {
	ID         string                 `json:"id" example:"AUDIT123456789"`
	UserID     string                 `json:"user_id" example:"USER123456789"`
	Action     string                 `json:"action" example:"user.login"`
	Resource   string                 `json:"resource" example:"user"`
	ResourceID string                 `json:"resource_id,omitempty" example:"USER987654321"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	IPAddress  string                 `json:"ip_address" example:"192.168.1.1"`
	UserAgent  string                 `json:"user_agent" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"`
	Timestamp  time.Time              `json:"timestamp" example:"2024-01-01T00:00:00Z"`
}

// AuditLogsResponse represents audit logs response
//
//	@Description	Audit logs response structure
type AuditLogsResponse struct {
	Success    bool           `json:"success" example:"true"`
	Message    string         `json:"message" example:"Audit logs retrieved successfully"`
	Data       AuditLogsData  `json:"data"`
	Pagination PaginationInfo `json:"pagination,omitempty"`
	Timestamp  time.Time      `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID  string         `json:"request_id,omitempty" example:"req-123456789"`
}

// AuditLogsData represents audit logs data
//
//	@Description	Audit logs data structure
type AuditLogsData struct {
	Logs []AuditLogEntry `json:"logs"`
}

// AuditStatistics represents audit statistics
//
//	@Description	Audit statistics structure
type AuditStatistics struct {
	TotalLogs      int            `json:"total_logs" example:"1000"`
	LogsByAction   map[string]int `json:"logs_by_action"`
	LogsByResource map[string]int `json:"logs_by_resource"`
	LogsByUser     map[string]int `json:"logs_by_user"`
	SecurityEvents int            `json:"security_events" example:"5"`
	Period         string         `json:"period" example:"30 days"`
	GeneratedAt    time.Time      `json:"generated_at" example:"2024-01-01T00:00:00Z"`
}

// AuditStatisticsResponse represents audit statistics response
//
//	@Description	Audit statistics response structure
type AuditStatisticsResponse struct {
	Success   bool                `json:"success" example:"true"`
	Message   string              `json:"message" example:"Audit statistics retrieved successfully"`
	Data      AuditStatisticsData `json:"data"`
	Timestamp time.Time           `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string              `json:"request_id,omitempty" example:"req-123456789"`
}

// AuditStatisticsData represents audit statistics data
//
//	@Description	Audit statistics data structure
type AuditStatisticsData struct {
	Statistics AuditStatistics `json:"statistics"`
}

// Health Response Models

// HealthCheckResponse represents health check response
//
//	@Description	Health check response structure
type HealthCheckResponse struct {
	Success   bool            `json:"success" example:"true"`
	Message   string          `json:"message" example:"Service is healthy"`
	Data      HealthCheckData `json:"data"`
	Timestamp time.Time       `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string          `json:"request_id,omitempty" example:"req-123456789"`
}

// HealthCheckData represents health check data
//
//	@Description	Health check data structure
type HealthCheckData struct {
	Status  string `json:"status" example:"healthy"`
	Service string `json:"service" example:"aaa-service"`
	Version string `json:"version" example:"2.0"`
}

// Utility Response Model

// MessageResponse represents a simple message response
//
//	@Description	Simple message response structure
type MessageResponse struct {
	Success   bool      `json:"success" example:"true"`
	Message   string    `json:"message" example:"Operation completed successfully"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string    `json:"request_id,omitempty" example:"req-123456789"`
}

// Admin Response Models

// AdminActionResponse represents admin action response
//
//	@Description	Admin action response structure
type AdminActionResponse struct {
	Success   bool            `json:"success" example:"true"`
	Message   string          `json:"message" example:"Admin action completed successfully"`
	Data      AdminActionData `json:"data,omitempty"`
	Timestamp time.Time       `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string          `json:"request_id,omitempty" example:"req-123456789"`
}

// AdminActionData represents admin action data
//
//	@Description	Admin action data structure
type AdminActionData struct {
	Action     string                 `json:"action" example:"grant_permission"`
	Target     string                 `json:"target" example:"USER123456789"`
	Details    map[string]interface{} `json:"details,omitempty"`
	ExecutedBy string                 `json:"executed_by" example:"ADMIN123456789"`
}

// Module Response Models

// ModuleInfo represents module information
//
//	@Description	Module information structure
type ModuleInfo struct {
	ServiceName  string    `json:"service_name" example:"user-service"`
	Version      string    `json:"version" example:"1.0.0"`
	Status       string    `json:"status" example:"active"`
	Health       string    `json:"health" example:"healthy"`
	RegisteredAt time.Time `json:"registered_at" example:"2024-01-01T00:00:00Z"`
	LastSeen     time.Time `json:"last_seen" example:"2024-01-01T00:00:00Z"`
}

// ModuleListResponse represents module list response
//
//	@Description	Module list response structure
type ModuleListResponse struct {
	Success   bool           `json:"success" example:"true"`
	Message   string         `json:"message" example:"Modules retrieved successfully"`
	Data      ModuleListData `json:"data"`
	Timestamp time.Time      `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string         `json:"request_id,omitempty" example:"req-123456789"`
}

// ModuleListData represents module list data
//
//	@Description	Module list data structure
type ModuleListData struct {
	Modules []ModuleInfo `json:"modules"`
}

// ModuleDetailResponse represents module detail response
//
//	@Description	Module detail response structure
type ModuleDetailResponse struct {
	Success   bool             `json:"success" example:"true"`
	Message   string           `json:"message" example:"Module details retrieved successfully"`
	Data      ModuleDetailData `json:"data"`
	Timestamp time.Time        `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string           `json:"request_id,omitempty" example:"req-123456789"`
}

// ModuleDetailData represents module detail data
//
//	@Description	Module detail data structure
type ModuleDetailData struct {
	Module    ModuleInfo `json:"module"`
	ServiceID string     `json:"service_id" example:"svc_123456"`
	Actions   []string   `json:"actions,omitempty" example:"create,read,update,delete"`
	Roles     []string   `json:"roles,omitempty" example:"admin,user"`
	Resources []string   `json:"resources,omitempty" example:"user,role"`
}
