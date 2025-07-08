package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/go-playground/validator/v10"
)

// Validator implements the Validator interface
type Validator struct {
	validate *validator.Validate
	logger   interfaces.Logger
}

// NewValidator creates a new Validator instance
func NewValidator(logger interfaces.Logger) interfaces.Validator {
	v := validator.New()

	// Register custom validations
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("mobile", validateMobile)
	v.RegisterValidation("aadhaar", validateAadhaar)
	v.RegisterValidation("pincode", validatePincode)

	return &Validator{
		validate: v,
		logger:   logger,
	}
}

// ValidateUserID validates user ID format
func (v *Validator) ValidateUserID(userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Check if it starts with "usr_" prefix
	if !strings.HasPrefix(userID, "usr_") {
		return fmt.Errorf("user ID must start with 'usr_' prefix")
	}

	// Check length (usr_ + 22 characters = 26 total)
	if len(userID) != 26 {
		return fmt.Errorf("user ID must be exactly 26 characters long")
	}

	// Check if it contains only alphanumeric characters and underscores
	matched, err := regexp.MatchString(`^usr_[a-zA-Z0-9_]+$`, userID)
	if err != nil {
		return fmt.Errorf("failed to validate user ID format: %w", err)
	}

	if !matched {
		return fmt.Errorf("user ID contains invalid characters")
	}

	return nil
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Use regex for email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	// Check length
	if len(email) > 254 {
		return fmt.Errorf("email cannot exceed 254 characters")
	}

	return nil
}

// ValidatePhone validates phone number format
func (v *Validator) ValidatePhone(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone number cannot be empty")
	}

	// Remove any non-digit characters
	phone = regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Check if it's a valid Indian mobile number (10 digits)
	if len(phone) != 10 {
		return fmt.Errorf("phone number must be exactly 10 digits")
	}

	// Check if it starts with valid Indian mobile prefixes
	validPrefixes := []string{"6", "7", "8", "9"}
	firstDigit := string(phone[0])
	isValidPrefix := false
	for _, prefix := range validPrefixes {
		if firstDigit == prefix {
			isValidPrefix = true
			break
		}
	}

	if !isValidPrefix {
		return fmt.Errorf("phone number must start with 6, 7, 8, or 9")
	}

	return nil
}

// ValidateUsername validates username format
func (v *Validator) ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Check length
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(username) > 50 {
		return fmt.Errorf("username cannot exceed 50 characters")
	}

	// Check if it contains only alphanumeric characters, underscores, and hyphens
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	if err != nil {
		return fmt.Errorf("failed to validate username format: %w", err)
	}

	if !matched {
		return fmt.Errorf("username can only contain letters, numbers, underscores, and hyphens")
	}

	// Check if it starts with a letter or number
	if !regexp.MustCompile(`^[a-zA-Z0-9]`).MatchString(username) {
		return fmt.Errorf("username must start with a letter or number")
	}

	// Check if it ends with a letter or number
	if !regexp.MustCompile(`[a-zA-Z0-9]$`).MatchString(username) {
		return fmt.Errorf("username must end with a letter or number")
	}

	return nil
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Check length
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return fmt.Errorf("password cannot exceed 128 characters")
	}

	// Check for at least one uppercase letter
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	if !regexp.MustCompile(`\d`).MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}

	// Check for at least one special character
	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateStruct validates a struct using tags
func (v *Validator) ValidateStruct(s interface{}) error {
	if s == nil {
		return fmt.Errorf("struct cannot be nil")
	}

	if err := v.validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errors []string
			for _, e := range validationErrors {
				errors = append(errors, formatValidationError(e))
			}
			return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
		}
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// Helper functions

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Check length
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	// Check format
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	if !matched {
		return false
	}

	// Check start and end
	if !regexp.MustCompile(`^[a-zA-Z0-9]`).MatchString(username) {
		return false
	}
	if !regexp.MustCompile(`[a-zA-Z0-9]$`).MatchString(username) {
		return false
	}

	return true
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Check length
	if len(password) < 8 || len(password) > 128 {
		return false
	}

	// Check requirements
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func validateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()

	// Remove non-digits
	mobile = regexp.MustCompile(`[^\d]`).ReplaceAllString(mobile, "")

	// Check length and prefix
	if len(mobile) != 10 {
		return false
	}

	validPrefixes := []string{"6", "7", "8", "9"}
	firstDigit := string(mobile[0])
	for _, prefix := range validPrefixes {
		if firstDigit == prefix {
			return true
		}
	}

	return false
}

func validateAadhaar(fl validator.FieldLevel) bool {
	aadhaar := fl.Field().String()

	// Remove non-digits
	aadhaar = regexp.MustCompile(`[^\d]`).ReplaceAllString(aadhaar, "")

	// Check length
	if len(aadhaar) != 12 {
		return false
	}

	// Check if it doesn't start with 0 or 1
	if strings.HasPrefix(aadhaar, "0") || strings.HasPrefix(aadhaar, "1") {
		return false
	}

	return true
}

func validatePincode(fl validator.FieldLevel) bool {
	pincode := fl.Field().String()

	// Remove non-digits
	pincode = regexp.MustCompile(`[^\d]`).ReplaceAllString(pincode, "")

	// Check length
	if len(pincode) != 6 {
		return false
	}

	return true
}

func formatValidationError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "username":
		return fmt.Sprintf("%s must be a valid username (3-50 characters, alphanumeric with underscores and hyphens)", field)
	case "password":
		return fmt.Sprintf("%s must be a strong password (8-128 characters, with uppercase, lowercase, digit, and special character)", field)
	case "mobile":
		return fmt.Sprintf("%s must be a valid 10-digit mobile number", field)
	case "aadhaar":
		return fmt.Sprintf("%s must be a valid 12-digit Aadhaar number", field)
	case "pincode":
		return fmt.Sprintf("%s must be a valid 6-digit pincode", field)
	default:
		return fmt.Sprintf("%s failed validation for tag %s", field, tag)
	}
}
