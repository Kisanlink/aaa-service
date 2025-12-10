package users

import (
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// UserResponse represents the user data sent in API responses
type UserResponse struct {
	ID            string            `json:"id"`
	PhoneNumber   string            `json:"phone_number"`
	CountryCode   string            `json:"country_code"`
	Username      *string           `json:"username,omitempty"`
	IsValidated   bool              `json:"is_validated"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	AadhaarNumber *string           `json:"aadhaar_number,omitempty"`
	Status        *string           `json:"status,omitempty"`
	Name          *string           `json:"name,omitempty"`
	CareOf        *string           `json:"care_of,omitempty"`
	DateOfBirth   *string           `json:"date_of_birth,omitempty"`
	Photo         *string           `json:"photo,omitempty"`
	EmailHash     *string           `json:"email_hash,omitempty"`
	ShareCode     *string           `json:"share_code,omitempty"`
	YearOfBirth   *string           `json:"year_of_birth,omitempty"`
	Message       *string           `json:"message,omitempty"`
	KYCStatus     string            `json:"kyc_status,omitempty"`
	Tokens        int               `json:"tokens"`
	Address       *AddressResponse  `json:"address,omitempty"`
	Contacts      []ContactResponse `json:"contacts,omitempty"`
	Roles         []UserRoleDetail  `json:"roles"`
	HasMPin       bool              `json:"has_mpin"`
}

// GetType returns the type of response
func (r *UserResponse) GetType() string {
	return "user"
}

// IsSuccess returns whether the response indicates success
func (r *UserResponse) IsSuccess() bool {
	return r.ID != ""
}

// FromModel converts a User model to UserResponse
// This method is optimized for list/search operations - it does NOT populate roles.
// For responses that need roles, use the GetUserWithRoles service method.
func (r *UserResponse) FromModel(user *models.User) {
	r.ID = user.ID
	r.PhoneNumber = user.PhoneNumber
	r.CountryCode = user.CountryCode
	r.Username = user.Username
	r.IsValidated = user.IsValidated
	r.CreatedAt = user.CreatedAt
	r.UpdatedAt = user.UpdatedAt
	r.Status = user.Status
	r.Tokens = user.Tokens
	r.HasMPin = user.HasMPin()

	// Map profile fields if profile exists
	// Check if profile is populated (GORM loads an empty struct if not found)
	if user.Profile.BaseModel != nil && user.Profile.ID != "" {
		r.Name = user.Profile.Name
		r.CareOf = user.Profile.CareOf
		r.Photo = user.Profile.Photo
		r.AadhaarNumber = user.Profile.AadhaarNumber
		r.EmailHash = user.Profile.EmailHash
		r.ShareCode = user.Profile.ShareCode
		r.YearOfBirth = user.Profile.YearOfBirth
		r.Message = user.Profile.Message
		r.DateOfBirth = user.Profile.DateOfBirth
		r.KYCStatus = user.Profile.KYCStatus

		// Map address if present
		if user.Profile.Address.BaseModel != nil && user.Profile.Address.ID != "" {
			addressResp := &AddressResponse{}
			addressResp.FromModel(&user.Profile.Address)
			r.Address = addressResp
		}
	}

	// Map contacts using existing ContactResponse
	r.Contacts = make([]ContactResponse, 0, len(user.Contacts))
	for _, contact := range user.Contacts {
		contactResp := ContactResponse{}
		contactResp.FromModel(&contact)
		r.Contacts = append(r.Contacts, contactResp)
	}

	// Roles are NOT populated - use GetUserWithRoles service method for role details
	r.Roles = []UserRoleDetail{}
}

// AddressResponse represents the address data in user responses
type AddressResponse struct {
	ID          string  `json:"id"`
	House       *string `json:"house,omitempty"`
	Street      *string `json:"street,omitempty"`
	Landmark    *string `json:"landmark,omitempty"`
	PostOffice  *string `json:"post_office,omitempty"`
	Subdistrict *string `json:"subdistrict,omitempty"`
	District    *string `json:"district,omitempty"`
	VTC         *string `json:"vtc,omitempty"`
	State       *string `json:"state,omitempty"`
	Country     *string `json:"country,omitempty"`
	Pincode     *string `json:"pincode,omitempty"`
	FullAddress *string `json:"full_address,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// FromModel converts an Address model to AddressResponse
func (r *AddressResponse) FromModel(address *models.Address) {
	r.ID = address.ID
	r.House = address.House
	r.Street = address.Street
	r.Landmark = address.Landmark
	r.PostOffice = address.PostOffice
	r.Subdistrict = address.Subdistrict
	r.District = address.District
	r.VTC = address.VTC
	r.State = address.State
	r.Country = address.Country
	r.Pincode = address.Pincode
	r.FullAddress = address.FullAddress
	r.CreatedAt = address.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	r.UpdatedAt = address.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
}

// RoleDetail represents role information in user responses
type RoleDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	AssignedAt  string `json:"assigned_at"`
}

// UserRoleDetail represents detailed user-role relationship information
type UserRoleDetail struct {
	ID       string     `json:"id"`
	UserID   string     `json:"user_id"`
	RoleID   string     `json:"role_id"`
	Role     RoleDetail `json:"role"`
	IsActive bool       `json:"is_active"`
}
