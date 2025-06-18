package model

import (
	"time"
)

type QueryResult struct {
	ID          string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-01T12:00:00Z"`
	Role        string    `json:"role" example:"admin"`
	Permissions string    `json:"permissions" example:"read,write,delete"`
}

type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Role   string `json:"role" binding:"required" example:"admin"`
}

type AadhaarOTPResponse struct {
	Timestamp     int64  `json:"timestamp" example:"1672531200"`
	TransactionID string `json:"transaction_id" example:"TXN123456789"`
	Entity        string `json:"entity" example:"aadhaar"`
	OtpMessage    string `json:"otp_message" example:"OTP sent successfully"`
	ReferenceID   string `json:"reference_id" example:"REF987654321"`
	StatusCode    int32  `json:"status_code" example:"200"`
}

type MinimalUser struct {
	ID           string              `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username     string              `json:"username" example:"johndoe"`
	MobileNumber uint64              `json:"mobile_number" example:"9876543210"`
	CountryCode  string              `json:"country_code" example:"+91"`
	IsValidated  bool                `json:"is_validated" example:"true"`
	CreatedAt    string              `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt    string              `json:"updated_at" example:"2023-01-02T10:30:00Z"`
	OtpResponse  *AadhaarOTPResponse `json:"otp_response,omitempty"`
}

type PasswordResetFlowRequest struct {
	Username    string `json:"username" binding:"required" example:"johndoe"`
	OTP         string `json:"otp,omitempty" example:"123456"`
	NewPassword string `json:"new_password,omitempty" example:"newSecurePassword123"`
}

type AddressRes struct {
	ID          string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	House       string `json:"house" example:"123"`
	Street      string `json:"street" example:"Main Street"`
	Landmark    string `json:"landmark" example:"Near Central Park"`
	PostOffice  string `json:"post_office" example:"Main Post Office"`
	Subdistrict string `json:"subdistrict" example:"Downtown"`
	District    string `json:"district" example:"Central District"`
	VTC         string `json:"vtc" example:"Metro City"`
	State       string `json:"state" example:"California"`
	Country     string `json:"country" example:"USA"`
	Pincode     string `json:"pincode" example:"123456"`
	FullAddress string `json:"full_address" example:"123 Main Street, Near Central Park, Metro City, California, USA - 123456"`
	CreatedAt   string `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt   string `json:"updated_at" example:"2023-01-02T10:30:00Z"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe"`
	Password string `json:"password" binding:"required" example:"securePassword123"`
}

type CreditUsageRequest struct {
	UserID          string  `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	TransactionType *string `json:"transaction_type" example:"debit"`
	Tokens          *int    `json:"tokens" example:"100"`
}

type CreateUserRequest struct {
	Username      string  `json:"username" example:"johndoe"`
	MobileNumber  uint64  `json:"mobile_number" validate:"required,numeric,len=10" example:"9876543210"`
	Password      string  `json:"password" example:"securePassword123"`
	AadhaarNumber *string `json:"aadhaar_number,omitempty" validate:"omitempty,numeric,len=12" example:"123456789012"`
	CountryCode   *string `json:"country_code,omitempty" example:"+91"`
}

type UpdateUserRequest struct {
	Username      string `json:"username" example:"johndoe"`
	AadhaarNumber string `json:"aadhaar_number" example:"123456789012"`
	EmailHash     string `json:"email_hash" example:"example@gmail.com"`
	MobileNumber  uint64 `json:"mobile_number" example:"9876543210"`
}

type CreateSchema struct {
	Resource  string   `json:"resource" example:"document"`
	Relations []string `json:"relations" example:"owner,reader"`
	Data      []Data   `json:"data"`
}

type Data struct {
	Action string   `json:"action" example:"read"`
	Roles  []string `json:"roles" example:"admin,user"`
}

type AssignRolePermission struct {
	ID          string       `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username    string       `json:"username" example:"johndoe"`
	IsValidated bool         `json:"is_validated" example:"true"`
	CreatedAt   string       `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt   string       `json:"updated_at" example:"2023-01-02T10:30:00Z"`
	Roles       []RoleDetail `json:"roles"`
}

type UserRes struct {
	ID            string       `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username      string       `json:"username" example:"johndoe"`
	Password      string       `json:"password" example:"hashedpassword"`
	IsValidated   bool         `json:"is_validated" example:"true"`
	CreatedAt     string       `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt     string       `json:"updated_at" example:"2023-01-02T10:30:00Z"`
	AadhaarNumber string       `json:"aadhaar_number" example:"123456789012"`
	Status        string       `json:"status" example:"active"`
	Name          string       `json:"name" example:"John Doe"`
	CareOf        string       `json:"care_of" example:"Father"`
	DateOfBirth   string       `json:"date_of_birth" example:"1990-01-01"`
	Photo         string       `json:"photo" example:"base64encodedimage"`
	EmailHash     string       `json:"email_hash" example:"a1b2c3d4e5f6"`
	ShareCode     string       `json:"share_code" example:"SHARE123"`
	YearOfBirth   string       `json:"year_of_birth" example:"1990"`
	Message       string       `json:"message" example:"Welcome"`
	MobileNumber  uint64       `json:"mobile_number" example:"9876543210"`
	CountryCode   string       `json:"country_code" example:"+91"`
	Address       *AddressRes  `json:"address"`
	Roles         []RoleDetail `json:"roles"`
}

type UserResponse struct {
	ID            string       `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username      string       `json:"username" example:"johndoe"`
	IsValidated   bool         `json:"is_validated" example:"true"`
	CreatedAt     string       `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt     string       `json:"updated_at" example:"2023-01-02T10:30:00Z"`
	AadhaarNumber string       `json:"aadhaar_number,omitempty" example:"123456789012"`
	Status        string       `json:"status,omitempty" example:"active"`
	Name          string       `json:"name,omitempty" example:"John Doe"`
	CareOf        string       `json:"care_of,omitempty" example:"Father"`
	DateOfBirth   string       `json:"date_of_birth,omitempty" example:"1990-01-01"`
	Photo         string       `json:"photo,omitempty" example:"base64encodedimage"`
	EmailHash     string       `json:"email_hash,omitempty" example:"a1b2c3d4e5f6"`
	ShareCode     string       `json:"share_code,omitempty" example:"SHARE123"`
	YearOfBirth   string       `json:"year_of_birth,omitempty" example:"1990"`
	Message       string       `json:"message,omitempty" example:"Welcome"`
	MobileNumber  *uint64      `json:"mobile_number,omitempty" example:"9876543210"`
	CountryCode   string       `json:"country_code,omitempty" example:"+91"`
	Address       *AddressRes  `json:"address,omitempty"`
	Roles         []RoleDetail `json:"roles"`
}

type CreateActionRequest struct {
	Name string `json:"name" binding:"required" example:"read"`
}

type CreateResourceRequest struct {
	Name string `json:"name" binding:"required" example:"document"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required" example:"admin"`
	Description string `json:"description" example:"Administrator role with full access"`
}

type CreatePermissionInput struct {
	Resource string   `json:"resource" binding:"required" example:"document"`
	Actions  []string `json:"actions" binding:"required" example:"read,write,delete"`
}

type RoleResponse struct {
	Roles []RoleDetail `json:"roles"`
}

type RoleDetail struct {
	RoleName    string              `json:"role_name" example:"admin"`
	Permissions []RolePermissionRes `json:"permissions"`
}

type RolePermissionRes struct {
	Resource string   `json:"resource" example:"document"`
	Actions  []string `json:"actions" example:"read,write,delete"`
}

type RolePermissionRequest struct {
	RoleID        string   `json:"roleId" binding:"required"`
	PermissionIDs []string `json:"permissionId" binding:"required"`
}

type CreatePermissionRequest struct {
	Resource string   `json:"resource"`
	Effect   string   `json:"effect"`
	Actions  []string `json:"actions"`
}

type GetRolePermissionResponse struct {
	ID         string      `json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Role       *Role       `json:"Role"`
	Permission *Permission `json:"Permission"`
}

type CheckPermissionRequest struct {
	Username     string `json:"username"`     // "alice"
	Action       string `json:"action"`       // "edit"
	ResourceType string `json:"resourceType"` // "db/farmers"
	ResourceID   string `json:"resourceID"`   // "123" (userid)
}
