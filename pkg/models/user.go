package models

import "time"

// User represents a user in the system for external consumption
type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	PhoneNumber string    `json:"phone_number"`
	CountryCode string    `json:"country_code"`
	IsValidated bool      `json:"is_validated"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserWithRoles represents a user with their roles
type UserWithRoles struct {
	User
	Roles []Role `json:"roles"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=50"`
	PhoneNumber string `json:"phone_number" validate:"required,phone"`
	CountryCode string `json:"country_code" validate:"required,iso3166_1_alpha2"`
	Password    string `json:"password" validate:"required,min=8"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Username    string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	PhoneNumber string `json:"phone_number,omitempty" validate:"omitempty,phone"`
	CountryCode string `json:"country_code,omitempty" validate:"omitempty,iso3166_1_alpha2"`
}

// ChangePasswordRequest represents a request to change password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ValidateUserRequest represents a request to validate a user
type ValidateUserRequest struct {
	ValidationToken string `json:"validation_token" validate:"required"`
}
