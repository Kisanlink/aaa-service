package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Validator implements the Validator interface
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new Validator instance
func NewValidator() interfaces.Validator {
	validate := validator.New()

	// Register custom validators
	validate.RegisterValidation("phone", validatePhoneNumber)
	validate.RegisterValidation("aadhaar", validateAadhaarNumber)

	return &Validator{
		validate: validate,
	}
}

// ValidateStruct validates a struct using the validator tags
func (v *Validator) ValidateStruct(s interface{}) error {
	return v.validate.Struct(s)
}

// ValidateUserID validates a user ID format
func (v *Validator) ValidateUserID(userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Check if it's a valid UUID-like format (alphanumeric with optional hyphens)
	matched, err := regexp.MatchString(`^[a-zA-Z0-9\-_]{1,50}$`, userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}
	if !matched {
		return fmt.Errorf("user ID must be alphanumeric with optional hyphens/underscores")
	}

	return nil
}

// ValidateEmail validates an email address format
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// Check for at least one digit
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;':"\\|,.<>?~]`).MatchString(password)

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidatePhoneNumber validates a phone number format
func (v *Validator) ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone number cannot be empty")
	}

	// Remove any non-digit characters for validation
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Indian phone number validation (10 digits, starting with 6-9)
	if len(digits) == 10 {
		if regexp.MustCompile(`^[6-9]\d{9}$`).MatchString(digits) {
			return nil
		}
		return fmt.Errorf("phone number must start with 6, 7, 8, or 9")
	}

	// International format with country code (10-15 digits)
	if len(digits) >= 10 && len(digits) <= 15 {
		return nil
	}

	return fmt.Errorf("phone number must be 10 digits (Indian) or 10-15 digits (international)")
}

// ValidateAadhaarNumber validates an Aadhaar number format
func (v *Validator) ValidateAadhaarNumber(aadhaar string) error {
	if aadhaar == "" {
		return fmt.Errorf("Aadhaar number cannot be empty")
	}

	// Remove any non-digit characters
	digits := regexp.MustCompile(`\D`).ReplaceAllString(aadhaar, "")

	// Aadhaar should be exactly 12 digits
	if len(digits) != 12 {
		return fmt.Errorf("Aadhaar number must be exactly 12 digits")
	}

	// Simple check for all same digits (invalid Aadhaar)
	firstDigit := digits[0]
	allSame := true
	for i := 1; i < len(digits); i++ {
		if digits[i] != firstDigit {
			allSame = false
			break
		}
	}

	if allSame {
		return fmt.Errorf("Aadhaar number cannot have all same digits")
	}

	return nil
}

// ParseListFilters parses query parameters into filter structure
func (v *Validator) ParseListFilters(c *gin.Context) (interface{}, error) {
	filters := make(map[string]interface{})

	// Parse pagination parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters["limit"] = limit
		} else {
			return nil, fmt.Errorf("invalid limit parameter: must be between 1 and 100")
		}
	} else {
		filters["limit"] = 10 // default
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters["offset"] = offset
		} else {
			return nil, fmt.Errorf("invalid offset parameter: must be >= 0")
		}
	} else {
		filters["offset"] = 0 // default
	}

	// Parse search query
	if query := c.Query("q"); query != "" {
		filters["search"] = strings.TrimSpace(query)
	}

	// Parse sort parameters
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters["sort_by"] = sortBy

		// Parse sort order
		sortOrder := c.Query("sort_order")
		if sortOrder == "" {
			sortOrder = "asc"
		}
		if sortOrder != "asc" && sortOrder != "desc" {
			return nil, fmt.Errorf("invalid sort_order: must be 'asc' or 'desc'")
		}
		filters["sort_order"] = sortOrder
	}

	// Parse filter by status
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	// Parse date range filters
	if createdAfter := c.Query("created_after"); createdAfter != "" {
		filters["created_after"] = createdAfter
	}
	if createdBefore := c.Query("created_before"); createdBefore != "" {
		filters["created_before"] = createdBefore
	}

	return filters, nil
}

// Custom validation functions

func validatePhoneNumber(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return false
	}

	// Remove any non-digit characters for validation
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Indian phone number validation (10 digits, starting with 6-9)
	if len(digits) == 10 {
		return regexp.MustCompile(`^[6-9]\d{9}$`).MatchString(digits)
	}

	// International format with country code (10-15 digits)
	if len(digits) >= 10 && len(digits) <= 15 {
		return true
	}

	return false
}

func validateAadhaarNumber(fl validator.FieldLevel) bool {
	aadhaar := fl.Field().String()
	if aadhaar == "" {
		return false
	}

	// Remove any non-digit characters
	digits := regexp.MustCompile(`\D`).ReplaceAllString(aadhaar, "")

	// Aadhaar should be exactly 12 digits
	if len(digits) != 12 {
		return false
	}

	// Simple check for all same digits (invalid Aadhaar)
	firstDigit := digits[0]
	allSame := true
	for i := 1; i < len(digits); i++ {
		if digits[i] != firstDigit {
			allSame = false
			break
		}
	}

	return !allSame
}
