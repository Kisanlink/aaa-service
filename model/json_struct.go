package model

import "time"

type QueryResult struct {
	ID          string `json:"id"`
	CreatedAt   time.Time
	Role        string
	Permissions string
}
type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

type RoleResp struct {
	RoleName    string       `json:"role_name"`
	Permissions []Permission `json:"permissions"`
}

type AssignRolePermission struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	IsValidated    bool       `json:"is_validated"`
	CreatedAt      string     `json:"created_at"`
	UpdatedAt      string     `json:"updated_at"`
	RolePermission []RoleResp `json:"role_permissions"`
}

type AadhaarOTPResponse struct {
	Timestamp     int64  `json:"timestamp"`
	TransactionID string `json:"transaction_id"`
	Entity        string `json:"entity"`
	OtpMessage    string `json:"otp_message"`
	ReferenceID   string `json:"reference_id"`
	StatusCode    int32  `json:"status_code"`
}

type MinimalUser struct {
	ID           string              `json:"id"`
	Username     string              `json:"username"`
	MobileNumber uint64              `json:"mobile_number"`
	CountryCode  string              `json:"country_code"`
	IsValidated  bool                `json:"is_validated"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
	OtpResponse  *AadhaarOTPResponse `json:"otp_response,omitempty"`
}

type PasswordResetFlowRequest struct {
	Username    string `json:"username" binding:"required"`
	OTP         string `json:"otp,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

type AddressRes struct {
	ID          string `json:"id"`
	House       string `json:"house"`
	Street      string `json:"street"`
	Landmark    string `json:"landmark"`
	PostOffice  string `json:"post_office"`
	Subdistrict string `json:"subdistrict"`
	District    string `json:"district"`
	VTC         string `json:"vtc"`
	State       string `json:"state"`
	Country     string `json:"country"`
	Pincode     string `json:"pincode"`
	FullAddress string `json:"full_address"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type UserRes struct {
	ID             string      `json:"id"`
	Username       string      `json:"username"`
	Password       string      `json:"password"`
	IsValidated    bool        `json:"is_validated"`
	CreatedAt      string      `json:"created_at"`
	UpdatedAt      string      `json:"updated_at"`
	RolePermission []RoleResp  `json:"role_permissions"`
	AadhaarNumber  string      `json:"aadhaar_number"`
	Status         string      `json:"status"`
	Name           string      `json:"name"`
	CareOf         string      `json:"care_of"`
	DateOfBirth    string      `json:"date_of_birth"`
	Photo          string      `json:"photo"`
	EmailHash      string      `json:"email_hash"`
	ShareCode      string      `json:"share_code"`
	YearOfBirth    string      `json:"year_of_birth"`
	Message        string      `json:"message"`
	MobileNumber   uint64      `json:"mobile_number"`
	CountryCode    string      `json:"country_code"`
	Address        *AddressRes `json:"address"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RoleRes struct {
	RoleName    string       `json:"role_name"`
	Permissions []Permission `json:"permissions"`
}

type UserResponse struct {
	ID             string      `json:"id"`
	Username       string      `json:"username"`
	IsValidated    bool        `json:"is_validated"`
	CreatedAt      string      `json:"created_at"`
	UpdatedAt      string      `json:"updated_at"`
	RolePermission []RoleRes   `json:"role_permissions"`
	AadhaarNumber  string      `json:"aadhaar_number,omitempty"`
	Status         string      `json:"status,omitempty"`
	Name           string      `json:"name,omitempty"`
	CareOf         string      `json:"care_of,omitempty"`
	DateOfBirth    string      `json:"date_of_birth,omitempty"`
	Photo          string      `json:"photo,omitempty"`
	EmailHash      string      `json:"email_hash,omitempty"`
	ShareCode      string      `json:"share_code,omitempty"`
	YearOfBirth    string      `json:"year_of_birth,omitempty"`
	Message        string      `json:"message,omitempty"`
	MobileNumber   *uint64     `json:"mobile_number,omitempty"`
	CountryCode    string      `json:"country_code,omitempty"`
	Address        *AddressRes `json:"address,omitempty"`
}

type CreditUsageRequest struct {
	UserID          string  `json:"user_id" binding:"required"` // Using Base.ID
	TransactionType *string `json:"transaction_type"`           // "debit", "credit", or nil
	Tokens          *int    `json:"tokens"`                     // Required for transactions
}

type AssignPermissionRequest struct {
	Role        string   `json:"role" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

type ConnRolePermissionResponse struct {
	ID          string        `json:"id"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
	Role        *Role         `json:"role"`
	Permissions []*Permission `json:"permissions"`
	IsActive    bool          `json:"is_active"`
}

type AssignPermissionResponse struct {
	StatusCode    int                         `json:"status_code"`
	Success       bool                        `json:"success"`
	Message       string                      `json:"message"`
	Data          *ConnRolePermissionResponse `json:"data"`
	DataTimeStamp string                      `json:"data_time_stamp"`
}

type RolePermissionResponse struct {
	ID          string        `json:"id"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Role        *Role         `json:"role"`
	Permissions []*Permission `json:"permission"`
	IsActive    bool          `json:"is_active"`
}

type GetAllRolePermissionsResponse struct {
	StatusCode    int                      `json:"status_code"`
	Success       bool                     `json:"success"`
	Message       string                   `json:"message"`
	DataTimeStamp string                   `json:"data_time_stamp"`
	Data          []RolePermissionResponse `json:"data"`
}

type GetRolePermissionByRoleNameResponse struct {
	StatusCode    int                     `json:"status_code"`
	Success       bool                    `json:"success"`
	Message       string                  `json:"message"`
	DataTimeStamp string                  `json:"data_time_stamp"`
	Data          *RolePermissionResponse `json:"data"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Source      string `json:"source"`
	Resource    string `json:"resource"`
}
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

// Wrap response with timestamp
type RolePermissionWrapper struct {
	Data      *RolePermissionResponse `json:"data"`
	Timestamp string                  `json:"timestamp"`
}

// Wrap data with timestamp
type RolePermissionResponseWrapper struct {
	Data      []RolePermissionResponse `json:"data"`
	Timestamp string                   `json:"timestamp"`
}

type CreateUserRequest struct {
	Username      string  `json:"username"`
	MobileNumber  uint64  `json:"mobile_number" validate:"required,numeric,len=10"`
	Password      string  `json:"password"`
	AadhaarNumber *string `json:"aadhaar_number,omitempty" validate:"omitempty,numeric,len=12"`
	CountryCode   *string `json:"country_code,omitempty"`
}
