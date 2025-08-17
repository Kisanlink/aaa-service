package models

import "time"

// LoginRequest represents a user login request
type LoginRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,phone"`
	CountryCode string `json:"country_code" validate:"required,iso3166_1_alpha2"`
	Password    string `json:"password" validate:"required"`
	MPin        string `json:"mpin,omitempty" validate:"omitempty,len=4,numeric"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	User         User   `json:"user"`
	Message      string `json:"message"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=50"`
	PhoneNumber string `json:"phone_number" validate:"required,phone"`
	CountryCode string `json:"country_code" validate:"required,iso3166_1_alpha2"`
	Password    string `json:"password" validate:"required,min=8"`
}

// RegisterResponse represents a successful registration response
type RegisterResponse struct {
	User    User   `json:"user"`
	Message string `json:"message"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	MPin         string `json:"mpin" validate:"required,len=4,numeric"`
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Message      string `json:"message"`
}

// LogoutResponse represents a logout response
type LogoutResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	PhoneNumber string    `json:"phone_number"`
	CountryCode string    `json:"country_code"`
	Roles       []string  `json:"roles"`
	IssuedAt    time.Time `json:"iat"`
	ExpiresAt   time.Time `json:"exp"`
}

// MFARequest represents a multi-factor authentication request
type MFARequest struct {
	UserID string `json:"user_id" validate:"required"`
	Code   string `json:"code" validate:"required,len=6,numeric"`
}

// SetMPinRequest represents a request to set MPIN
type SetMPinRequest struct {
	Password string `json:"password" validate:"required"`
	MPin     string `json:"mpin" validate:"required,len=4,numeric"`
}
