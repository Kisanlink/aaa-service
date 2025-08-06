package users

import (
	"fmt"
	"regexp"
	"strconv"
)

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	PhoneNumber   string  `json:"phone_number" validate:"required"`
	CountryCode   string  `json:"country_code" validate:"required"`
	Password      string  `json:"password" validate:"required,min=8,max=128"`
	Username      *string `json:"username,omitempty" validate:"omitempty,username"`
	AadhaarNumber *string `json:"aadhaar_number,omitempty"`
	Name          *string `json:"name,omitempty"`
	CareOf        *string `json:"care_of,omitempty"`
	DateOfBirth   *string `json:"date_of_birth,omitempty"`
	YearOfBirth   *string `json:"year_of_birth,omitempty"`
}

// Validate validates the CreateUserRequest
func (r *CreateUserRequest) Validate() error {
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
	if len(r.Password) > 128 {
		return fmt.Errorf("password must be at most 128 characters long")
	}

	// Validate Aadhaar number if provided
	if r.AadhaarNumber != nil && *r.AadhaarNumber != "" {
		if len(*r.AadhaarNumber) != 12 {
			return fmt.Errorf("aadhaar number must be exactly 12 digits")
		}
		aadhaarRegex := regexp.MustCompile(`^[0-9]{12}$`)
		if !aadhaarRegex.MatchString(*r.AadhaarNumber) {
			return fmt.Errorf("aadhaar number must contain only digits")
		}
	}

	// Validate year of birth if provided
	if r.YearOfBirth != nil && *r.YearOfBirth != "" {
		year, err := strconv.Atoi(*r.YearOfBirth)
		if err != nil {
			return fmt.Errorf("year of birth must be a valid number")
		}
		if year < 1900 || year > 2024 {
			return fmt.Errorf("year of birth must be between 1900 and 2024")
		}
	}

	return nil
}

// GetType returns the type of request
func (r *CreateUserRequest) GetType() string {
	return "create_user"
}
