package users

import (
	"fmt"
	"strconv"
)

// UpdateUserRequest represents the request to update an existing user
type UpdateUserRequest struct {
	UserID       string  `json:"user_id" validate:"required"`
	Status       *string `json:"status,omitempty"`
	Name         *string `json:"name,omitempty"`
	CareOf       *string `json:"care_of,omitempty"`
	DateOfBirth  *string `json:"date_of_birth,omitempty"`
	Photo        *string `json:"photo,omitempty"`
	EmailHash    *string `json:"email_hash,omitempty"`
	ShareCode    *string `json:"share_code,omitempty"`
	YearOfBirth  *string `json:"year_of_birth,omitempty"`
	Message      *string `json:"message,omitempty"`
	MobileNumber *uint64 `json:"mobile_number,omitempty"`
	CountryCode  *string `json:"country_code,omitempty"`
	Tokens       *int    `json:"tokens,omitempty"`
}

// Validate validates the UpdateUserRequest
func (r *UpdateUserRequest) Validate() error {
	// Validate status if provided
	if r.Status != nil && *r.Status != "" {
		validStatuses := []string{"pending", "active", "suspended", "deleted"}
		isValid := false
		for _, status := range validStatuses {
			if *r.Status == status {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("status must be one of: %v", validStatuses)
		}
	}

	// Validate mobile number if provided
	if r.MobileNumber != nil {
		mobileStr := strconv.FormatUint(*r.MobileNumber, 10)
		if len(mobileStr) != 10 {
			return fmt.Errorf("mobile number must be exactly 10 digits")
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

	// Validate tokens if provided
	if r.Tokens != nil && *r.Tokens < 0 {
		return fmt.Errorf("tokens cannot be negative")
	}

	// Validate country code if provided
	if r.CountryCode != nil && *r.CountryCode != "" {
		if len(*r.CountryCode) < 1 || len(*r.CountryCode) > 10 {
			return fmt.Errorf("country code must be between 1 and 10 characters")
		}
	}

	return nil
}

// GetType returns the type of request
func (r *UpdateUserRequest) GetType() string {
	return "update_user"
}
