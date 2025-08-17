package users

import (
	"github.com/Kisanlink/aaa-service/internal/entities/requests"
)

// UpdateContactRequest represents a request to update a contact
type UpdateContactRequest struct {
	*requests.BaseRequest
	ContactID    string  `json:"contact_id" validate:"required"`
	UserID       string  `json:"user_id" validate:"required"`
	MobileNumber *uint64 `json:"mobile_number" validate:"omitempty"`
	CountryCode  *string `json:"country_code" validate:"omitempty,max=10"`
	EmailHash    *string `json:"email_hash" validate:"omitempty,email"`
	ShareCode    *string `json:"share_code" validate:"omitempty,max=50"`
	AddressID    *string `json:"address_id" validate:"omitempty"`
}

// NewUpdateContactRequest creates a new UpdateContactRequest instance
func NewUpdateContactRequest(
	contactID string,
	userID string,
	mobileNumber *uint64,
	countryCode *string,
	emailHash *string,
	shareCode *string,
	addressID *string,
	protocol string,
	operation string,
	version string,
	requestID string,
	headers map[string][]string,
	body interface{},
	context map[string]interface{},
) *UpdateContactRequest {
	return &UpdateContactRequest{
		BaseRequest: requests.NewBaseRequest(
			protocol,
			operation,
			version,
			requestID,
			"UpdateContact",
			headers,
			body,
			context,
		),
		ContactID:    contactID,
		UserID:       userID,
		MobileNumber: mobileNumber,
		CountryCode:  countryCode,
		EmailHash:    emailHash,
		ShareCode:    shareCode,
		AddressID:    addressID,
	}
}

// Validate validates the UpdateContactRequest
func (r *UpdateContactRequest) Validate() error {
	if r.ContactID == "" {
		return requests.NewValidationError("contact_id", "Contact ID is required")
	}

	if r.UserID == "" {
		return requests.NewValidationError("user_id", "User ID is required")
	}

	if r.MobileNumber != nil && *r.MobileNumber == 0 {
		return requests.NewValidationError("mobile_number", "Mobile number cannot be zero")
	}

	// Validate mobile number format (10 digits for Indian numbers)
	if r.MobileNumber != nil {
		mobileStr := string(rune(*r.MobileNumber))
		if len(mobileStr) != 10 {
			return requests.NewValidationError("mobile_number", "Mobile number must be 10 digits")
		}
	}

	if r.CountryCode != nil && len(*r.CountryCode) > 10 {
		return requests.NewValidationError("country_code", "Country code must be at most 10 characters")
	}

	if r.ShareCode != nil && len(*r.ShareCode) > 50 {
		return requests.NewValidationError("share_code", "Share code must be at most 50 characters")
	}

	return nil
}

// GetContactID returns the contact ID
func (r *UpdateContactRequest) GetContactID() string {
	return r.ContactID
}

// GetUserID returns the user ID
func (r *UpdateContactRequest) GetUserID() string {
	return r.UserID
}

// GetMobileNumber returns the mobile number
func (r *UpdateContactRequest) GetMobileNumber() *uint64 {
	return r.MobileNumber
}

// GetCountryCode returns the country code
func (r *UpdateContactRequest) GetCountryCode() *string {
	return r.CountryCode
}

// GetEmailHash returns the email hash
func (r *UpdateContactRequest) GetEmailHash() *string {
	return r.EmailHash
}

// GetShareCode returns the share code
func (r *UpdateContactRequest) GetShareCode() *string {
	return r.ShareCode
}

// GetAddressID returns the address ID
func (r *UpdateContactRequest) GetAddressID() *string {
	return r.AddressID
}
