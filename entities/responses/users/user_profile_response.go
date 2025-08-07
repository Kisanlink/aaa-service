package users

import (
	"time"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/entities/responses"
)

// UserProfileResponse represents a user profile response
type UserProfileResponse struct {
	responses.Response
	ID            string           `json:"id"`
	UserID        string           `json:"user_id"`
	Name          *string          `json:"name,omitempty"`
	CareOf        *string          `json:"care_of,omitempty"`
	DateOfBirth   *string          `json:"date_of_birth,omitempty"`
	Photo         *string          `json:"photo,omitempty"`
	YearOfBirth   *string          `json:"year_of_birth,omitempty"`
	Message       *string          `json:"message,omitempty"`
	AadhaarNumber *string          `json:"aadhaar_number,omitempty"`
	EmailHash     *string          `json:"email_hash,omitempty"`
	ShareCode     *string          `json:"share_code,omitempty"`
	AddressID     *string          `json:"address_id,omitempty"`
	Address       *AddressResponse `json:"address,omitempty"`
	CreatedAt     string           `json:"created_at"`
	UpdatedAt     string           `json:"updated_at"`
}

// NewUserProfileResponse creates a new UserProfileResponse from a UserProfile model
func NewUserProfileResponse(profile *models.UserProfile) *UserProfileResponse {
	response := &UserProfileResponse{
		ID:            profile.ID,
		UserID:        profile.UserID,
		Name:          profile.Name,
		CareOf:        profile.CareOf,
		DateOfBirth:   profile.DateOfBirth,
		Photo:         profile.Photo,
		YearOfBirth:   profile.YearOfBirth,
		Message:       profile.Message,
		AadhaarNumber: profile.AadhaarNumber,
		EmailHash:     profile.EmailHash,
		ShareCode:     profile.ShareCode,
		AddressID:     profile.AddressID,
		CreatedAt:     profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     profile.UpdatedAt.Format(time.RFC3339),
	}

	// Include address if available
	if profile.Address.ID != "" {
		addressResp := &AddressResponse{}
		addressResp.FromModel(&profile.Address)
		response.Address = addressResp
	}

	return response
}

// NewUserProfileResponseFromModel creates a new UserProfileResponse from a UserProfile model
func NewUserProfileResponseFromModel(profile *models.UserProfile) *UserProfileResponse {
	return NewUserProfileResponse(profile)
}

// GetID returns the profile ID
func (r *UserProfileResponse) GetID() string {
	return r.ID
}

// GetUserID returns the user ID
func (r *UserProfileResponse) GetUserID() string {
	return r.UserID
}

// GetName returns the name
func (r *UserProfileResponse) GetName() *string {
	return r.Name
}

// GetCareOf returns the care of
func (r *UserProfileResponse) GetCareOf() *string {
	return r.CareOf
}

// GetDateOfBirth returns the date of birth
func (r *UserProfileResponse) GetDateOfBirth() *string {
	return r.DateOfBirth
}

// GetPhoto returns the photo
func (r *UserProfileResponse) GetPhoto() *string {
	return r.Photo
}

// GetYearOfBirth returns the year of birth
func (r *UserProfileResponse) GetYearOfBirth() *string {
	return r.YearOfBirth
}

// GetMessage returns the message
func (r *UserProfileResponse) GetMessage() *string {
	return r.Message
}

// GetAadhaarNumber returns the Aadhaar number
func (r *UserProfileResponse) GetAadhaarNumber() *string {
	return r.AadhaarNumber
}

// GetEmailHash returns the email hash
func (r *UserProfileResponse) GetEmailHash() *string {
	return r.EmailHash
}

// GetShareCode returns the share code
func (r *UserProfileResponse) GetShareCode() *string {
	return r.ShareCode
}

// GetAddressID returns the address ID
func (r *UserProfileResponse) GetAddressID() *string {
	return r.AddressID
}

// GetAddress returns the address
func (r *UserProfileResponse) GetAddress() *AddressResponse {
	return r.Address
}

// GetCreatedAt returns the created at timestamp
func (r *UserProfileResponse) GetCreatedAt() string {
	return r.CreatedAt
}

// GetUpdatedAt returns the updated at timestamp
func (r *UserProfileResponse) GetUpdatedAt() string {
	return r.UpdatedAt
}
