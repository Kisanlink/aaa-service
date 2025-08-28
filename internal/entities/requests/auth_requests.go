package requests

import (
	"fmt"
	"regexp"
)

// LoginRequest represents a user login request supporting both password and MPIN authentication
type LoginRequest struct {
	PhoneNumber     string  `json:"phone_number" validate:"required"`
	CountryCode     string  `json:"country_code" validate:"required"`
	Password        *string `json:"password,omitempty" validate:"omitempty,min=8"`
	MPin            *string `json:"mpin,omitempty" validate:"omitempty,len=4|len=6"`
	MFACode         *string `json:"mfa_code,omitempty"`
	IncludeProfile  *bool   `json:"include_profile,omitempty"`
	IncludeRoles    *bool   `json:"include_roles,omitempty"`
	IncludeContacts *bool   `json:"include_contacts,omitempty"`
}

// Validate validates the LoginRequest
func (r *LoginRequest) Validate() error {
	if r.PhoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}
	if r.CountryCode == "" {
		return fmt.Errorf("country code is required")
	}

	// At least one authentication method must be provided
	if (r.Password == nil || *r.Password == "") && (r.MPin == nil || *r.MPin == "") {
		return fmt.Errorf("either password or mpin is required")
	}

	// Validate password if provided
	if r.Password != nil && *r.Password != "" {
		if len(*r.Password) < 8 {
			return fmt.Errorf("password must be at least 8 characters long")
		}
	}

	// Validate MPIN if provided
	if r.MPin != nil && *r.MPin != "" {
		if len(*r.MPin) != 4 && len(*r.MPin) != 6 {
			return fmt.Errorf("mpin must be 4 or 6 digits")
		}
		mPinRegex := regexp.MustCompile(`^\d+$`)
		if !mPinRegex.MatchString(*r.MPin) {
			return fmt.Errorf("mpin must contain only digits")
		}
	}

	return nil
}

// GetType returns the request type
func (r *LoginRequest) GetType() string {
	return "login"
}

// HasPassword checks if password is provided
func (r *LoginRequest) HasPassword() bool {
	return r.Password != nil && *r.Password != ""
}

// HasMPin checks if MPIN is provided
func (r *LoginRequest) HasMPin() bool {
	return r.MPin != nil && *r.MPin != ""
}

// GetPassword returns the password value or empty string if not provided
func (r *LoginRequest) GetPassword() string {
	if r.Password == nil {
		return ""
	}
	return *r.Password
}

// GetMPin returns the MPIN value or empty string if not provided
func (r *LoginRequest) GetMPin() string {
	if r.MPin == nil {
		return ""
	}
	return *r.MPin
}

// ShouldIncludeProfile returns true if profile should be included in response
func (r *LoginRequest) ShouldIncludeProfile() bool {
	return r.IncludeProfile != nil && *r.IncludeProfile
}

// ShouldIncludeRoles returns true if roles should be included in response
func (r *LoginRequest) ShouldIncludeRoles() bool {
	return r.IncludeRoles != nil && *r.IncludeRoles
}

// ShouldIncludeContacts returns true if contacts should be included in response
func (r *LoginRequest) ShouldIncludeContacts() bool {
	return r.IncludeContacts != nil && *r.IncludeContacts
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	PhoneNumber   string  `json:"phone_number" validate:"required"`
	CountryCode   string  `json:"country_code" validate:"required"`
	Password      string  `json:"password" validate:"required,min=8"`
	Username      *string `json:"username,omitempty" validate:"omitempty,username"`
	AadhaarNumber *string `json:"aadhaar_number,omitempty"`
	Name          *string `json:"name,omitempty"`
}

// Validate validates the RegisterRequest
func (r *RegisterRequest) Validate() error {
	// Validate phone number
	if r.PhoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}
	// Basic phone number validation (10 digits for Indian numbers)
	phoneRegex := regexp.MustCompile(`^\d{10}$`)
	if !phoneRegex.MatchString(r.PhoneNumber) {
		return fmt.Errorf("phone number must be 10 digits")
	}

	// Validate country code
	if r.CountryCode == "" {
		return fmt.Errorf("country code is required")
	}
	// Basic country code validation
	countryCodeRegex := regexp.MustCompile(`^\+\d{1,4}$`)
	if !countryCodeRegex.MatchString(r.CountryCode) {
		return fmt.Errorf("country code must start with + and contain 1-4 digits")
	}

	// Validate optional username if provided
	if r.Username != nil && *r.Username != "" {
		if len(*r.Username) < 3 || len(*r.Username) > 100 {
			return fmt.Errorf("username must be between 3 and 100 characters")
		}
		usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
		if !usernameRegex.MatchString(*r.Username) {
			return fmt.Errorf("username can only contain letters, numbers, and underscores")
		}
	}

	// Validate password
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(r.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	return nil
}

// GetType returns the request type
func (r *RegisterRequest) GetType() string {
	return "register"
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	MPin         string `json:"mpin" validate:"required,len=4|len=6"`
}

// Validate validates the RefreshTokenRequest
func (r *RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return fmt.Errorf("refresh token is required")
	}
	if r.MPin == "" {
		return fmt.Errorf("mPin is required")
	}
	// Validate mPin format (4 or 6 digits)
	if len(r.MPin) != 4 && len(r.MPin) != 6 {
		return fmt.Errorf("mPin must be 4 or 6 digits")
	}
	mPinRegex := regexp.MustCompile(`^\d+$`)
	if !mPinRegex.MatchString(r.MPin) {
		return fmt.Errorf("mPin must contain only digits")
	}
	return nil
}

// GetType returns the request type
func (r *RefreshTokenRequest) GetType() string {
	return "refresh_token"
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	PhoneNumber *string `json:"phone_number,omitempty"`
	CountryCode *string `json:"country_code,omitempty"`
	Username    *string `json:"username,omitempty"`
	Email       *string `json:"email,omitempty"`
}

// Validate validates the ForgotPasswordRequest
func (r *ForgotPasswordRequest) Validate() error {
	if r.PhoneNumber == nil && r.Username == nil && r.Email == nil {
		return fmt.Errorf("at least one of phone number, username, or email is required")
	}
	if r.PhoneNumber != nil && r.CountryCode == nil {
		return fmt.Errorf("country code is required when phone number is provided")
	}
	return nil
}

// GetType returns the request type
func (r *ForgotPasswordRequest) GetType() string {
	return "forgot_password"
}

// ResetPasswordRequest represents a reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// Validate validates the ResetPasswordRequest
func (r *ResetPasswordRequest) Validate() error {
	if r.Token == "" {
		return fmt.Errorf("reset token is required")
	}
	if r.NewPassword == "" {
		return fmt.Errorf("new password is required")
	}
	if len(r.NewPassword) < 8 {
		return fmt.Errorf("new password must be at least 8 characters long")
	}
	return nil
}

// GetType returns the request type
func (r *ResetPasswordRequest) GetType() string {
	return "reset_password"
}

// SetMPinRequest represents a request to set or update mPin
type SetMPinRequest struct {
	MPin     string `json:"mpin" validate:"required,len=4|len=6"`
	Password string `json:"password" validate:"required"`
}

// Validate validates the SetMPinRequest
func (r *SetMPinRequest) Validate() error {
	if r.MPin == "" {
		return fmt.Errorf("mPin is required")
	}
	if len(r.MPin) != 4 && len(r.MPin) != 6 {
		return fmt.Errorf("mPin must be 4 or 6 digits")
	}
	mPinRegex := regexp.MustCompile(`^\d+$`)
	if !mPinRegex.MatchString(r.MPin) {
		return fmt.Errorf("mPin must contain only digits")
	}
	if r.Password == "" {
		return fmt.Errorf("password is required for verification")
	}
	return nil
}

// GetType returns the request type
func (r *SetMPinRequest) GetType() string {
	return "set_mpin"
}

// UpdateMPinRequest represents a request to update existing mPin
type UpdateMPinRequest struct {
	CurrentMPin string `json:"current_mpin" validate:"required,len=4|len=6"`
	NewMPin     string `json:"new_mpin" validate:"required,len=4|len=6"`
}

// Validate validates the UpdateMPinRequest
func (r *UpdateMPinRequest) Validate() error {
	if r.CurrentMPin == "" {
		return fmt.Errorf("current mPin is required")
	}
	if len(r.CurrentMPin) != 4 && len(r.CurrentMPin) != 6 {
		return fmt.Errorf("current mPin must be 4 or 6 digits")
	}
	mPinRegex := regexp.MustCompile(`^\d+$`)
	if !mPinRegex.MatchString(r.CurrentMPin) {
		return fmt.Errorf("current mPin must contain only digits")
	}

	if r.NewMPin == "" {
		return fmt.Errorf("new mPin is required")
	}
	if len(r.NewMPin) != 4 && len(r.NewMPin) != 6 {
		return fmt.Errorf("new mPin must be 4 or 6 digits")
	}
	if !mPinRegex.MatchString(r.NewMPin) {
		return fmt.Errorf("new mPin must contain only digits")
	}

	if r.CurrentMPin == r.NewMPin {
		return fmt.Errorf("new mPin must be different from current mPin")
	}

	return nil
}

// GetType returns the request type
func (r *UpdateMPinRequest) GetType() string {
	return "update_mpin"
}
