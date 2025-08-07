package users

import (
	"github.com/Kisanlink/aaa-service/entities/requests"
)

// CreateUserProfileRequest represents a request to create a user profile
type CreateUserProfileRequest struct {
	*requests.BaseRequest
	UserID        string  `json:"user_id" validate:"required"`
	Name          *string `json:"name" validate:"omitempty,min=2,max=255"`
	CareOf        *string `json:"care_of" validate:"omitempty,max=255"`
	DateOfBirth   *string `json:"date_of_birth" validate:"omitempty,len=10"`
	Photo         *string `json:"photo"`
	YearOfBirth   *string `json:"year_of_birth" validate:"omitempty,len=4"`
	Message       *string `json:"message"`
	AadhaarNumber *string `json:"aadhaar_number" validate:"omitempty,len=12"`
	EmailHash     *string `json:"email_hash" validate:"omitempty,email"`
	ShareCode     *string `json:"share_code" validate:"omitempty,max=50"`
	AddressID     *string `json:"address_id" validate:"omitempty"`
}

// NewCreateUserProfileRequest creates a new CreateUserProfileRequest instance
func NewCreateUserProfileRequest(
	userID string,
	name *string,
	careOf *string,
	dateOfBirth *string,
	photo *string,
	yearOfBirth *string,
	message *string,
	aadhaarNumber *string,
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
) *CreateUserProfileRequest {
	return &CreateUserProfileRequest{
		BaseRequest: requests.NewBaseRequest(
			protocol,
			operation,
			version,
			requestID,
			"CreateUserProfile",
			headers,
			body,
			context,
		),
		UserID:        userID,
		Name:          name,
		CareOf:        careOf,
		DateOfBirth:   dateOfBirth,
		Photo:         photo,
		YearOfBirth:   yearOfBirth,
		Message:       message,
		AadhaarNumber: aadhaarNumber,
		EmailHash:     emailHash,
		ShareCode:     shareCode,
		AddressID:     addressID,
	}
}

// Validate validates the CreateUserProfileRequest
func (r *CreateUserProfileRequest) Validate() error {
	if r.UserID == "" {
		return requests.NewValidationError("user_id", "User ID is required")
	}

	if r.Name != nil && (*r.Name == "" || len(*r.Name) < 2) {
		return requests.NewValidationError("name", "Name must be at least 2 characters long")
	}

	if r.DateOfBirth != nil && len(*r.DateOfBirth) != 10 {
		return requests.NewValidationError("date_of_birth", "Date of birth must be in YYYY-MM-DD format")
	}

	if r.YearOfBirth != nil && len(*r.YearOfBirth) != 4 {
		return requests.NewValidationError("year_of_birth", "Year of birth must be 4 digits")
	}

	if r.AadhaarNumber != nil && len(*r.AadhaarNumber) != 12 {
		return requests.NewValidationError("aadhaar_number", "Aadhaar number must be 12 digits")
	}

	return nil
}

// GetUserID returns the user ID
func (r *CreateUserProfileRequest) GetUserID() string {
	return r.UserID
}

// GetName returns the name
func (r *CreateUserProfileRequest) GetName() *string {
	return r.Name
}

// GetCareOf returns the care of
func (r *CreateUserProfileRequest) GetCareOf() *string {
	return r.CareOf
}

// GetDateOfBirth returns the date of birth
func (r *CreateUserProfileRequest) GetDateOfBirth() *string {
	return r.DateOfBirth
}

// GetPhoto returns the photo
func (r *CreateUserProfileRequest) GetPhoto() *string {
	return r.Photo
}

// GetYearOfBirth returns the year of birth
func (r *CreateUserProfileRequest) GetYearOfBirth() *string {
	return r.YearOfBirth
}

// GetMessage returns the message
func (r *CreateUserProfileRequest) GetMessage() *string {
	return r.Message
}

// GetAadhaarNumber returns the Aadhaar number
func (r *CreateUserProfileRequest) GetAadhaarNumber() *string {
	return r.AadhaarNumber
}

// GetEmailHash returns the email hash
func (r *CreateUserProfileRequest) GetEmailHash() *string {
	return r.EmailHash
}

// GetShareCode returns the share code
func (r *CreateUserProfileRequest) GetShareCode() *string {
	return r.ShareCode
}

// GetAddressID returns the address ID
func (r *CreateUserProfileRequest) GetAddressID() *string {
	return r.AddressID
}
