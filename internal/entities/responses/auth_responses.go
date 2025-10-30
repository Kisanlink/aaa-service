package responses

import (
	"time"
)

// UserInfo represents comprehensive user information in auth responses
// @Description Comprehensive user information returned in authentication responses
type UserInfo struct {
	ID          string           `json:"id" example:"USER00000001"`
	PhoneNumber string           `json:"phone_number" example:"9876543210"`
	CountryCode string           `json:"country_code" example:"+91"`
	Username    *string          `json:"username,omitempty" example:"ramesh_kumar"`
	IsValidated bool             `json:"is_validated" example:"true"`
	Status      *string          `json:"status,omitempty" example:"active"`
	CreatedAt   time.Time        `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   time.Time        `json:"updated_at" example:"2024-01-20T14:45:00Z"`
	Tokens      int              `json:"tokens" example:"2"`
	HasMPin     bool             `json:"has_mpin" example:"true"`
	Roles       []UserRoleDetail `json:"roles,omitempty"`
	Profile     *UserProfileInfo `json:"profile,omitempty"`
	Contacts    []ContactInfo    `json:"contacts,omitempty"`
}

// UserRoleDetail represents detailed user role information
type UserRoleDetail struct {
	ID       string     `json:"id"`
	UserID   string     `json:"user_id"`
	RoleID   string     `json:"role_id"`
	Role     RoleDetail `json:"role"`
	IsActive bool       `json:"is_active"`
}

// RoleDetail represents detailed role information
type RoleDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
	IsActive    bool   `json:"is_active"`
	Version     int    `json:"version"`
}

// Helper methods for UserInfo

// HasRoles checks if user has any roles
func (u *UserInfo) HasRoles() bool {
	return len(u.Roles) > 0
}

// HasProfile checks if user has profile information
func (u *UserInfo) HasProfile() bool {
	return u.Profile != nil
}

// HasContacts checks if user has contact information
func (u *UserInfo) HasContacts() bool {
	return len(u.Contacts) > 0
}

// GetActiveRoles returns only active roles
func (u *UserInfo) GetActiveRoles() []UserRoleDetail {
	var activeRoles []UserRoleDetail
	for _, role := range u.Roles {
		if role.IsActive && role.Role.IsActive {
			activeRoles = append(activeRoles, role)
		}
	}
	return activeRoles
}

// HasRole checks if user has a specific role by name
func (u *UserInfo) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.IsActive && role.Role.IsActive && role.Role.Name == roleName {
			return true
		}
	}
	return false
}

// GetRoleNames returns a list of active role names
func (u *UserInfo) GetRoleNames() []string {
	var roleNames []string
	for _, role := range u.Roles {
		if role.IsActive && role.Role.IsActive {
			roleNames = append(roleNames, role.Role.Name)
		}
	}
	return roleNames
}

// UserProfileInfo represents user profile information in responses
type UserProfileInfo struct {
	ID            string       `json:"id"`
	Name          *string      `json:"name,omitempty"`
	CareOf        *string      `json:"care_of,omitempty"`
	DateOfBirth   *string      `json:"date_of_birth,omitempty"`
	YearOfBirth   *string      `json:"year_of_birth,omitempty"`
	AadhaarNumber *string      `json:"aadhaar_number,omitempty"`
	EmailHash     *string      `json:"email_hash,omitempty"`
	ShareCode     *string      `json:"share_code,omitempty"`
	Address       *AddressInfo `json:"address,omitempty"`
}

// AddressInfo represents address information in responses
type AddressInfo struct {
	ID          string  `json:"id"`
	House       *string `json:"house,omitempty"`
	Street      *string `json:"street,omitempty"`
	Landmark    *string `json:"landmark,omitempty"`
	PostOffice  *string `json:"post_office,omitempty"`
	Subdistrict *string `json:"subdistrict,omitempty"`
	District    *string `json:"district,omitempty"`
	VTC         *string `json:"vtc,omitempty"` // Village/Town/City
	State       *string `json:"state,omitempty"`
	Country     *string `json:"country,omitempty"`
	Pincode     *string `json:"pincode,omitempty"`
	FullAddress *string `json:"full_address,omitempty"`
}

// ContactInfo represents contact information in responses
type ContactInfo struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Value       string  `json:"value"`
	IsPrimary   bool    `json:"is_primary"`
	IsVerified  bool    `json:"is_verified"`
	Description *string `json:"description,omitempty"`
}

// LoginResponse represents the response for a successful login
// @Description Successful login response with JWT tokens and user information
type LoginResponse struct {
	AccessToken  string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiVVNFUjAwMDAwMDAxIiwiZXhwIjoxNzA1NDM1MjAwfQ.abc123def456ghi789"`
	RefreshToken string    `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiVVNFUjAwMDAwMDAxIiwiZXhwIjoxNzA4MDI3MjAwfQ.xyz789uvw456rst123"`
	TokenType    string    `json:"token_type" example:"Bearer"`
	ExpiresIn    int64     `json:"expires_in" example:"3600"`
	User         *UserInfo `json:"user"`
	Message      string    `json:"message" example:"Login successful"`
}

// GetType returns the response type
func (r *LoginResponse) GetType() string {
	return "login"
}

// IsSuccess returns whether the response indicates success
func (r *LoginResponse) IsSuccess() bool {
	return r.AccessToken != ""
}

// RegisterResponse represents the response for a successful registration
type RegisterResponse struct {
	User    *UserInfo `json:"user"`
	Message string    `json:"message"`
}

// GetType returns the response type
func (r *RegisterResponse) GetType() string {
	return "register"
}

// IsSuccess returns whether the response indicates success
func (r *RegisterResponse) IsSuccess() bool {
	return r.User != nil && r.User.ID != ""
}

// RefreshTokenResponse represents the response for a token refresh
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Message      string `json:"message"`
}

// GetType returns the response type
func (r *RefreshTokenResponse) GetType() string {
	return "refresh_token"
}

// IsSuccess returns whether the response indicates success
func (r *RefreshTokenResponse) IsSuccess() bool {
	return r.AccessToken != ""
}

// ForgotPasswordResponse represents the response for a forgot password request
type ForgotPasswordResponse struct {
	Message string `json:"message"`
	SentTo  string `json:"sent_to,omitempty"`
}

// GetType returns the response type
func (r *ForgotPasswordResponse) GetType() string {
	return "forgot_password"
}

// IsSuccess returns whether the response indicates success
func (r *ForgotPasswordResponse) IsSuccess() bool {
	return r.Message != ""
}

// ResetPasswordResponse represents the response for a reset password request
type ResetPasswordResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// GetType returns the response type
func (r *ResetPasswordResponse) GetType() string {
	return "reset_password"
}

// IsSuccess returns whether the response indicates success
func (r *ResetPasswordResponse) IsSuccess() bool {
	return r.Success
}

// LogoutResponse represents the response for a logout request
type LogoutResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// GetType returns the response type
func (r *LogoutResponse) GetType() string {
	return "logout"
}

// IsSuccess returns whether the response indicates success
func (r *LogoutResponse) IsSuccess() bool {
	return r.Success
}

// TokenValidationResponse represents the response for token validation
type TokenValidationResponse struct {
	Valid   bool                   `json:"valid"`
	User    *UserInfo              `json:"user,omitempty"`
	Claims  map[string]interface{} `json:"claims,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// GetType returns the response type
func (r *TokenValidationResponse) GetType() string {
	return "token_validation"
}

// IsSuccess returns whether the response indicates success
func (r *TokenValidationResponse) IsSuccess() bool {
	return r.Valid
}

// SetMPinResponse represents the response for setting MPIN
type SetMPinResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// GetType returns the response type
func (r *SetMPinResponse) GetType() string {
	return "set_mpin"
}

// IsSuccess returns whether the response indicates success
func (r *SetMPinResponse) IsSuccess() bool {
	return r.Success
}

// UpdateMPinResponse represents the response for updating MPIN
type UpdateMPinResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// GetType returns the response type
func (r *UpdateMPinResponse) GetType() string {
	return "update_mpin"
}

// IsSuccess returns whether the response indicates success
func (r *UpdateMPinResponse) IsSuccess() bool {
	return r.Success
}

// AssignRoleResponse represents the response for role assignment
type AssignRoleResponse struct {
	Message string     `json:"message"`
	UserID  string     `json:"user_id"`
	Role    RoleDetail `json:"role"`
	Success bool       `json:"success"`
}

// GetType returns the response type
func (r *AssignRoleResponse) GetType() string {
	return "assign_role"
}

// IsSuccess returns whether the response indicates success
func (r *AssignRoleResponse) IsSuccess() bool {
	return r.Success
}

// RemoveRoleResponse represents the response for role removal
type RemoveRoleResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	RoleID  string `json:"role_id"`
	Success bool   `json:"success"`
}

// GetType returns the response type
func (r *RemoveRoleResponse) GetType() string {
	return "remove_role"
}

// IsSuccess returns whether the response indicates success
func (r *RemoveRoleResponse) IsSuccess() bool {
	return r.Success
}

// GetUserRolesResponse represents the response for getting user roles
type GetUserRolesResponse struct {
	UserID  string           `json:"user_id"`
	Roles   []UserRoleDetail `json:"roles"`
	Message string           `json:"message"`
	Success bool             `json:"success"`
}

// GetType returns the response type
func (r *GetUserRolesResponse) GetType() string {
	return "get_user_roles"
}

// IsSuccess returns whether the response indicates success
func (r *GetUserRolesResponse) IsSuccess() bool {
	return r.Success
}
