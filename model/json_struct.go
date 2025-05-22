package model

import (
	"time"
)

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

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreditUsageRequest struct {
	UserID          string  `json:"user_id" binding:"required"` // Using Base.ID
	TransactionType *string `json:"transaction_type"`           // "debit", "credit", or nil
	Tokens          *int    `json:"tokens"`                     // Required for transactions
}

type CreateUserRequest struct {
	Username      string  `json:"username"`
	MobileNumber  uint64  `json:"mobile_number" validate:"required,numeric,len=10"`
	Password      string  `json:"password"`
	AadhaarNumber *string `json:"aadhaar_number,omitempty" validate:"omitempty,numeric,len=12"`
	CountryCode   *string `json:"country_code,omitempty"`
}

type UpdateUserRequest struct {
	Username     string `json:"username"`
	Status       string `json:"status"`
	Name         string `json:"name"`
	CareOf       string `json:"care_of"`
	DateOfBirth  string `json:"date_of_birth"`
	Photo        string `json:"photo"`
	EmailHash    string `json:"email_hash"`
	YearOfBirth  string `json:"year_of_birth"`
	Message      string `json:"message"`
	MobileNumber uint64 `json:"mobile_number"`
}

// type Data struct {
// 	Actions  []string `json:"actions"`
// 	RoleName string   `json:"role_name"`
// }

// type CreateSchema struct {
// 	Resource string `json:"resource"`
// 	Data     []Data `json:"data"`
// }

type CreateSchema struct {
	Resource  string   `json:"resource"`
	Relations []string `json:"relations"`
	Data      []Data   `json:"data"`
}

type Data struct {
	Action string   `json:"action"`
	Roles  []string `json:"roles"`
}
type AssignRolePermission struct {
	ID          string       `json:"id"`
	Username    string       `json:"username"`
	IsValidated bool         `json:"is_validated"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
	Roles       []RoleDetail `json:"roles"`
}

type UserRes struct {
	ID            string       `json:"id"`
	Username      string       `json:"username"`
	Password      string       `json:"password"`
	IsValidated   bool         `json:"is_validated"`
	CreatedAt     string       `json:"created_at"`
	UpdatedAt     string       `json:"updated_at"`
	AadhaarNumber string       `json:"aadhaar_number"`
	Status        string       `json:"status"`
	Name          string       `json:"name"`
	CareOf        string       `json:"care_of"`
	DateOfBirth   string       `json:"date_of_birth"`
	Photo         string       `json:"photo"`
	EmailHash     string       `json:"email_hash"`
	ShareCode     string       `json:"share_code"`
	YearOfBirth   string       `json:"year_of_birth"`
	Message       string       `json:"message"`
	MobileNumber  uint64       `json:"mobile_number"`
	CountryCode   string       `json:"country_code"`
	Address       *AddressRes  `json:"address"`
	Roles         []RoleDetail `json:"roles"`
}

type UserResponse struct {
	ID            string       `json:"id"`
	Username      string       `json:"username"`
	IsValidated   bool         `json:"is_validated"`
	CreatedAt     string       `json:"created_at"`
	UpdatedAt     string       `json:"updated_at"`
	AadhaarNumber string       `json:"aadhaar_number,omitempty"`
	Status        string       `json:"status,omitempty"`
	Name          string       `json:"name,omitempty"`
	CareOf        string       `json:"care_of,omitempty"`
	DateOfBirth   string       `json:"date_of_birth,omitempty"`
	Photo         string       `json:"photo,omitempty"`
	EmailHash     string       `json:"email_hash,omitempty"`
	ShareCode     string       `json:"share_code,omitempty"`
	YearOfBirth   string       `json:"year_of_birth,omitempty"`
	Message       string       `json:"message,omitempty"`
	MobileNumber  *uint64      `json:"mobile_number,omitempty"`
	CountryCode   string       `json:"country_code,omitempty"`
	Address       *AddressRes  `json:"address,omitempty"`
	Roles         []RoleDetail `json:"roles"`
}

type CreateActionRequest struct {
	Name string `json:"name" binding:"required"` // Action name
}
type CreateResourceRequest struct {
	Name string `json:"name" binding:"required"` // Action name
}

type CreateRoleRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	Source      string                  `json:"source"`
	Permissions []CreatePermissionInput `json:"permissions"`
}

type CreatePermissionInput struct {
	Resource string   `json:"resource" binding:"required"`
	Actions  []string `json:"actions" binding:"required"`
}

type RoleResponse struct {
	Roles []RoleDetail `json:"roles"`
}
type RoleDetail struct {
	RoleName    string           `json:"role_name"`
	Permissions []RolePermission `json:"permissions"`
}

type RolePermission struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}
