package users

import (
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/responses"
)

// ContactResponse represents a contact response
type ContactResponse struct {
	responses.Response
	ID           string           `json:"id"`
	UserID       string           `json:"user_id"`
	MobileNumber uint64           `json:"mobile_number"`
	CountryCode  *string          `json:"country_code,omitempty"`
	EmailHash    *string          `json:"email_hash,omitempty"`
	ShareCode    *string          `json:"share_code,omitempty"`
	AddressID    *string          `json:"address_id,omitempty"`
	Address      *AddressResponse `json:"address,omitempty"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
}

// NewContactResponse creates a new ContactResponse from a Contact model
func NewContactResponse(contact *models.Contact) *ContactResponse {
	response := &ContactResponse{
		ID:           contact.ID,
		UserID:       contact.UserID,
		MobileNumber: contact.MobileNumber,
		CountryCode:  contact.CountryCode,
		EmailHash:    contact.EmailHash,
		ShareCode:    contact.ShareCode,
		AddressID:    contact.AddressID,
		CreatedAt:    contact.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    contact.UpdatedAt.Format(time.RFC3339),
	}

	// Include address if available
	if contact.Address.ID != "" {
		addressResp := &AddressResponse{}
		addressResp.FromModel(&contact.Address)
		response.Address = addressResp
	}

	return response
}

// NewContactResponseFromModel creates a new ContactResponse from a Contact model
func NewContactResponseFromModel(contact *models.Contact) *ContactResponse {
	return NewContactResponse(contact)
}

// FromModel converts a Contact model to ContactResponse
func (r *ContactResponse) FromModel(contact *models.Contact) {
	r.ID = contact.ID
	r.UserID = contact.UserID
	r.MobileNumber = contact.MobileNumber
	r.CountryCode = contact.CountryCode
	r.EmailHash = contact.EmailHash
	r.ShareCode = contact.ShareCode
	r.AddressID = contact.AddressID
	r.CreatedAt = contact.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	r.UpdatedAt = contact.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")

	// Convert address if present
	if contact.Address.ID != "" {
		addressResp := &AddressResponse{}
		addressResp.FromModel(&contact.Address)
		r.Address = addressResp
	}
}

// GetID returns the contact ID
func (r *ContactResponse) GetID() string {
	return r.ID
}

// GetUserID returns the user ID
func (r *ContactResponse) GetUserID() string {
	return r.UserID
}

// GetMobileNumber returns the mobile number
func (r *ContactResponse) GetMobileNumber() uint64 {
	return r.MobileNumber
}

// GetCountryCode returns the country code
func (r *ContactResponse) GetCountryCode() *string {
	return r.CountryCode
}

// GetEmailHash returns the email hash
func (r *ContactResponse) GetEmailHash() *string {
	return r.EmailHash
}

// GetShareCode returns the share code
func (r *ContactResponse) GetShareCode() *string {
	return r.ShareCode
}

// GetAddressID returns the address ID
func (r *ContactResponse) GetAddressID() *string {
	return r.AddressID
}

// GetAddress returns the address
func (r *ContactResponse) GetAddress() *AddressResponse {
	return r.Address
}

// GetCreatedAt returns the created at timestamp
func (r *ContactResponse) GetCreatedAt() string {
	return r.CreatedAt
}

// GetUpdatedAt returns the updated at timestamp
func (r *ContactResponse) GetUpdatedAt() string {
	return r.UpdatedAt
}
