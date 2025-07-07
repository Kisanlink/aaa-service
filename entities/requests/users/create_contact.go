package users

import (
	"github.com/Kisanlink/aaa-service/entities/requests"
)

// CreateContactRequest represents a request to create a contact
type CreateContactRequest struct {
	requests.Request
	UserID       string  `json:"user_id" validate:"required"`
	MobileNumber uint64  `json:"mobile_number" validate:"required"`
	CountryCode  *string `json:"country_code" validate:"omitempty,max=10"`
	EmailHash    *string `json:"email_hash" validate:"omitempty,email"`
	ShareCode    *string `json:"share_code" validate:"omitempty,max=50"`
	AddressID    *string `json:"address_id" validate:"omitempty"`
}

// NewCreateContactRequest creates a new CreateContactRequest instance
func NewCreateContactRequest(
	userID string,
	mobileNumber uint64,
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
) *CreateContactRequest {
	return &CreateContactRequest{
		Request: requests.Request{
			Protocol:  protocol,
			Operation: operation,
			Version:   version,
			RequestID: requestID,
			Headers:   headers,
			Body:      body,
			Context:   context,
		},
		UserID:       userID,
		MobileNumber: mobileNumber,
		CountryCode:  countryCode,
		EmailHash:    emailHash,
		ShareCode:    shareCode,
		AddressID:    addressID,
	}
}

// Validate validates the CreateContactRequest
func (r *CreateContactRequest) Validate() error {
	if r.UserID == "" {
		return requests.NewValidationError("user_id", "User ID is required")
	}

	if r.MobileNumber == 0 {
		return requests.NewValidationError("mobile_number", "Mobile number is required")
	}

	// Validate mobile number format (10 digits for Indian numbers)
	mobileStr := string(rune(r.MobileNumber))
	if len(mobileStr) != 10 {
		return requests.NewValidationError("mobile_number", "Mobile number must be 10 digits")
	}

	if r.CountryCode != nil && len(*r.CountryCode) > 10 {
		return requests.NewValidationError("country_code", "Country code must be at most 10 characters")
	}

	if r.ShareCode != nil && len(*r.ShareCode) > 50 {
		return requests.NewValidationError("share_code", "Share code must be at most 50 characters")
	}

	return nil
}

// GetUserID returns the user ID
func (r *CreateContactRequest) GetUserID() string {
	return r.UserID
}

// GetMobileNumber returns the mobile number
func (r *CreateContactRequest) GetMobileNumber() uint64 {
	return r.MobileNumber
}

// GetCountryCode returns the country code
func (r *CreateContactRequest) GetCountryCode() *string {
	return r.CountryCode
}

// GetEmailHash returns the email hash
func (r *CreateContactRequest) GetEmailHash() *string {
	return r.EmailHash
}

// GetShareCode returns the share code
func (r *CreateContactRequest) GetShareCode() *string {
	return r.ShareCode
}

// GetAddressID returns the address ID
func (r *CreateContactRequest) GetAddressID() *string {
	return r.AddressID
}
