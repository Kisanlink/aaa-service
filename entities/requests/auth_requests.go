package requests

import (
	"fmt"
	"regexp"
)

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Validate validates the LoginRequest
func (r *LoginRequest) Validate() error {
	if r.Username == "" {
		return fmt.Errorf("username is required")
	}
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

// GetType returns the request type
func (r *LoginRequest) GetType() string {
	return "login"
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username      string  `json:"username" validate:"required"`
	Password      string  `json:"password" validate:"required,min=8"`
	MobileNumber  uint64  `json:"mobile_number" validate:"required"`
	AadhaarNumber *string `json:"aadhaar_number,omitempty"`
	CountryCode   *string `json:"country_code,omitempty"`
	Name          *string `json:"name,omitempty"`
}

// Validate validates the RegisterRequest
func (r *RegisterRequest) Validate() error {
	// Validate username
	if r.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(r.Username) < 3 || len(r.Username) > 100 {
		return fmt.Errorf("username must be between 3 and 100 characters")
	}

	// Validate username format
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(r.Username) {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}

	// Validate password
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(r.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Validate mobile number
	if r.MobileNumber == 0 {
		return fmt.Errorf("mobile number is required")
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
}

// Validate validates the RefreshTokenRequest
func (r *RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return fmt.Errorf("refresh token is required")
	}
	return nil
}

// GetType returns the request type
func (r *RefreshTokenRequest) GetType() string {
	return "refresh_token"
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Mobile   *uint64 `json:"mobile,omitempty"`
}

// Validate validates the ForgotPasswordRequest
func (r *ForgotPasswordRequest) Validate() error {
	if r.Username == nil && r.Email == nil && r.Mobile == nil {
		return fmt.Errorf("at least one of username, email, or mobile is required")
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
