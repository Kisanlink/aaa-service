package responses

import (
	"time"
)

// StandardUserResponse represents the standardized user response structure
// This can be used with any user data model that provides the required fields
type StandardUserResponse struct {
	ID          string     `json:"id"`
	PhoneNumber string     `json:"phone_number"`
	CountryCode string     `json:"country_code"`
	Username    *string    `json:"username,omitempty"`
	IsValidated bool       `json:"is_validated"`
	IsActive    bool       `json:"is_active"`
	Status      *string    `json:"status,omitempty"`
	Tokens      int        `json:"tokens"`
	HasMPin     bool       `json:"has_mpin"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`

	// Optional nested objects (controlled by TransformOptions)
	Profile  *StandardUserProfileResponse `json:"profile,omitempty"`
	Contacts []*StandardContactResponse   `json:"contacts,omitempty"`
	Address  *StandardAddressResponse     `json:"address,omitempty"`
	Roles    []*StandardUserRoleResponse  `json:"roles,omitempty"`
}

// StandardRoleResponse represents the standardized role response structure
// This can be used with any role data model that provides the required fields
type StandardRoleResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Scope          string     `json:"scope"`
	IsActive       bool       `json:"is_active"`
	Version        int        `json:"version"`
	OrganizationID *string    `json:"organization_id,omitempty"`
	GroupID        *string    `json:"group_id,omitempty"`
	ParentID       *string    `json:"parent_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`

	// Optional nested objects
	Permissions []*StandardPermissionResponse `json:"permissions,omitempty"`
	Children    []*StandardRoleResponse       `json:"children,omitempty"`
}

// StandardUserRoleResponse represents the standardized user-role relationship response
// This can be used with any user-role data model that provides the required fields
type StandardUserRoleResponse struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	RoleID     string     `json:"role_id"`
	IsActive   bool       `json:"is_active"`
	AssignedAt time.Time  `json:"assigned_at"`
	AssignedBy *string    `json:"assigned_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`

	// Optional nested objects (controlled by TransformOptions)
	User *StandardUserResponse `json:"user,omitempty"`
	Role *StandardRoleResponse `json:"role,omitempty"`
}

// StandardUserProfileResponse represents the standardized user profile response
// This can be used with any user profile data model that provides the required fields
type StandardUserProfileResponse struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Name          *string    `json:"name,omitempty"`
	CareOf        *string    `json:"care_of,omitempty"`
	DateOfBirth   *string    `json:"date_of_birth,omitempty"`
	YearOfBirth   *string    `json:"year_of_birth,omitempty"`
	Photo         *string    `json:"photo,omitempty"`
	EmailHash     *string    `json:"email_hash,omitempty"`
	ShareCode     *string    `json:"share_code,omitempty"`
	AadhaarNumber *string    `json:"aadhaar_number,omitempty"` // Masked for security
	Message       *string    `json:"message,omitempty"`
	AddressID     *string    `json:"address_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// StandardContactResponse represents the standardized contact response
// This can be used with any contact data model that provides the required fields
type StandardContactResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Type        string     `json:"type"`
	Value       string     `json:"value"`
	CountryCode *string    `json:"country_code,omitempty"`
	IsPrimary   bool       `json:"is_primary"`
	IsVerified  bool       `json:"is_verified"`
	IsActive    bool       `json:"is_active"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`
	VerifiedBy  *string    `json:"verified_by,omitempty"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// StandardAddressResponse represents the standardized address response
// This can be used with any address data model that provides the required fields
type StandardAddressResponse struct {
	ID          string     `json:"id"`
	UserID      *string    `json:"user_id,omitempty"`
	House       *string    `json:"house,omitempty"`
	Street      *string    `json:"street,omitempty"`
	Landmark    *string    `json:"landmark,omitempty"`
	PostOffice  *string    `json:"post_office,omitempty"`
	Subdistrict *string    `json:"subdistrict,omitempty"`
	District    *string    `json:"district,omitempty"`
	VTC         *string    `json:"vtc,omitempty"`
	State       *string    `json:"state,omitempty"`
	Country     *string    `json:"country,omitempty"`
	Pincode     *string    `json:"pincode,omitempty"`
	FullAddress *string    `json:"full_address,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// StandardPermissionResponse represents the standardized permission response
// This can be used with any permission data model that provides the required fields
type StandardPermissionResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ResourceID  *string    `json:"resource_id,omitempty"`
	ActionID    *string    `json:"action_id,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// StandardOrganizationResponse represents the standardized organization response
// This can be used with any organization data model that provides the required fields
type StandardOrganizationResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Type        *string    `json:"type,omitempty"`
	IsActive    bool       `json:"is_active"`
	ParentID    *string    `json:"parent_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`

	// Optional nested objects
	Children []*StandardOrganizationResponse `json:"children,omitempty"`
}

// StandardGroupResponse represents the standardized group response
// This can be used with any group data model that provides the required fields
type StandardGroupResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	Type           *string    `json:"type,omitempty"`
	IsActive       bool       `json:"is_active"`
	OrganizationID *string    `json:"organization_id,omitempty"`
	ParentID       *string    `json:"parent_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`

	// Optional nested objects
	Organization *StandardOrganizationResponse `json:"organization,omitempty"`
	Children     []*StandardGroupResponse      `json:"children,omitempty"`
}

// StandardActionResponse represents the standardized action response
// This can be used with any action data model that provides the required fields
type StandardActionResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// StandardResourceResponse represents the standardized resource response
// This can be used with any resource data model that provides the required fields
type StandardResourceResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Type        *string    `json:"type,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// StandardAuditLogResponse represents the standardized audit log response
// This can be used with any audit log data model that provides the required fields
type StandardAuditLogResponse struct {
	ID         string                 `json:"id"`
	UserID     *string                `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID *string                `json:"resource_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IPAddress  *string                `json:"ip_address,omitempty"`
	UserAgent  *string                `json:"user_agent,omitempty"`
	Success    bool                   `json:"success"`
	ErrorMsg   *string                `json:"error_message,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`

	// Optional nested objects
	User *StandardUserResponse `json:"user,omitempty"`
}

// StandardErrorResponse represents the standardized error response structure
type StandardErrorResponse struct {
	Error     string                 `json:"error"`
	Message   string                 `json:"message"`
	Code      string                 `json:"code"`
	RequestID *string                `json:"request_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Success   bool                   `json:"success"`
}

// StandardSuccessResponse represents the standardized success response structure
type StandardSuccessResponse struct {
	Message   string                 `json:"message"`
	Data      interface{}            `json:"data,omitempty"`
	RequestID *string                `json:"request_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Success   bool                   `json:"success"`
}

// StandardPaginatedResponse represents the standardized paginated response structure
type StandardPaginatedResponse struct {
	Data       interface{}            `json:"data"`
	Pagination PaginationInfo         `json:"pagination"`
	RequestID  *string                `json:"request_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Success    bool                   `json:"success"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}
