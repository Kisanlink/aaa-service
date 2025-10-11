package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/gin-gonic/gin"
	v10 "github.com/go-playground/validator/v10"
)

// Validator implements the Validator interface
type Validator struct {
	validate *v10.Validate
}

// NewValidator creates a new Validator instance
func NewValidator() interfaces.Validator {
	validate := v10.New()

	// Register custom validators
	if err := validate.RegisterValidation("phone", validatePhoneNumber); err != nil {
		panic(fmt.Sprintf("failed to register phone validation: %v", err))
	}
	if err := validate.RegisterValidation("aadhaar", validateAadhaarNumber); err != nil {
		panic(fmt.Sprintf("failed to register aadhaar validation: %v", err))
	}
	if err := validate.RegisterValidation("username", validateUsername); err != nil {
		panic(fmt.Sprintf("failed to register username validation: %v", err))
	}
	if err := validate.RegisterValidation("org_id", validateOrganizationID); err != nil {
		panic(fmt.Sprintf("failed to register org_id validation: %v", err))
	}
	if err := validate.RegisterValidation("group_id", validateGroupID); err != nil {
		panic(fmt.Sprintf("failed to register group_id validation: %v", err))
	}
	if err := validate.RegisterValidation("user_id", validateUserID); err != nil {
		panic(fmt.Sprintf("failed to register user_id validation: %v", err))
	}
	if err := validate.RegisterValidation("role_id", validateRoleID); err != nil {
		panic(fmt.Sprintf("failed to register role_id validation: %v", err))
	}

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

	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
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

	// Check for common weak patterns
	if v.isCommonWeakPassword(password) {
		return fmt.Errorf("password is too common or weak")
	}

	return nil
}

// ValidateMPin validates MPIN format and security requirements
func (v *Validator) ValidateMPin(mpin string) error {
	if mpin == "" {
		return fmt.Errorf("MPIN cannot be empty")
	}

	// MPIN must be 4 or 6 digits
	if len(mpin) != 4 && len(mpin) != 6 {
		return fmt.Errorf("MPIN must be 4 or 6 digits")
	}

	// Check if all characters are digits
	if !regexp.MustCompile(`^\d+$`).MatchString(mpin) {
		return fmt.Errorf("MPIN must contain only digits")
	}

	// Check for weak patterns
	if v.isWeakMPin(mpin) {
		return fmt.Errorf("MPIN is too weak - avoid sequential or repeated digits")
	}

	return nil
}

// SanitizeInput sanitizes user input to prevent injection attacks
func (v *Validator) SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline, carriage return, and tab
	sanitized := ""
	for _, r := range input {
		if r >= 32 || r == '\n' || r == '\r' || r == '\t' {
			sanitized += string(r)
		}
	}

	// Trim whitespace
	return strings.TrimSpace(sanitized)
}

// ValidateAndSanitizeString validates and sanitizes string input
func (v *Validator) ValidateAndSanitizeString(input string, fieldName string, minLen, maxLen int) (string, error) {
	if input == "" && minLen > 0 {
		return "", fmt.Errorf("%s cannot be empty", fieldName)
	}

	sanitized := v.SanitizeInput(input)

	if len(sanitized) < minLen {
		return "", fmt.Errorf("%s must be at least %d characters", fieldName, minLen)
	}

	if len(sanitized) > maxLen {
		return "", fmt.Errorf("%s must not exceed %d characters", fieldName, maxLen)
	}

	// Check for potential SQL injection patterns
	if v.containsSQLInjectionPatterns(sanitized) {
		return "", fmt.Errorf("%s contains invalid characters", fieldName)
	}

	return sanitized, nil
}

// isCommonWeakPassword checks for common weak password patterns
func (v *Validator) isCommonWeakPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
		"1234567890", "password1", "123456789", "welcome123",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if strings.Contains(lowerPassword, common) {
			return true
		}
	}

	// Check for keyboard patterns
	keyboardPatterns := []string{
		"qwertyuiop", "asdfghjkl", "zxcvbnm", "1234567890",
		"qwerty", "asdfgh", "zxcvbn", "123456", "abcdef",
	}

	for _, pattern := range keyboardPatterns {
		if strings.Contains(lowerPassword, pattern) {
			return true
		}
	}

	return false
}

// isWeakMPin checks for weak MPIN patterns
func (v *Validator) isWeakMPin(mpin string) bool {
	// Check for all same digits
	firstDigit := mpin[0]
	allSame := true
	for i := 1; i < len(mpin); i++ {
		if mpin[i] != firstDigit {
			allSame = false
			break
		}
	}
	if allSame {
		return true
	}

	// Check for sequential patterns (ascending)
	isSequential := true
	for i := 1; i < len(mpin); i++ {
		if int(mpin[i]) != int(mpin[i-1])+1 {
			isSequential = false
			break
		}
	}
	if isSequential {
		return true
	}

	// Check for sequential patterns (descending)
	isReverseSequential := true
	for i := 1; i < len(mpin); i++ {
		if int(mpin[i]) != int(mpin[i-1])-1 {
			isReverseSequential = false
			break
		}
	}
	if isReverseSequential {
		return true
	}

	// Check for common weak patterns
	weakPatterns := []string{"1234", "4321", "0000", "1111", "2222", "3333", "4444", "5555", "6666", "7777", "8888", "9999", "123456", "654321", "000000", "111111", "222222", "333333", "444444", "555555", "666666", "777777", "888888", "999999"}
	for _, pattern := range weakPatterns {
		if mpin == pattern {
			return true
		}
	}

	return false
}

// containsSQLInjectionPatterns checks for potential SQL injection patterns
func (v *Validator) containsSQLInjectionPatterns(input string) bool {
	lowerInput := strings.ToLower(input)

	// Common SQL injection patterns
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"union", "select", "insert", "update", "delete", "drop",
		"create", "alter", "exec", "execute", "script",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
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
		return fmt.Errorf("aadhaar number cannot be empty")
	}

	// Remove any non-digit characters
	digits := regexp.MustCompile(`\D`).ReplaceAllString(aadhaar, "")

	// Aadhaar should be exactly 12 digits
	if len(digits) != 12 {
		return fmt.Errorf("aadhaar number must be exactly 12 digits")
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
		return fmt.Errorf("aadhaar number cannot have all same digits")
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

func validatePhoneNumber(fl v10.FieldLevel) bool {
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

func validateAadhaarNumber(fl v10.FieldLevel) bool {
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

func validateUsername(fl v10.FieldLevel) bool {
	username := fl.Field().String()
	if username == "" {
		return true // Allow empty values (omitempty should handle this)
	}

	// Username must be between 3 and 100 characters
	if len(username) < 3 || len(username) > 100 {
		return false
	}

	// Username can only contain letters, numbers, and underscores
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return usernameRegex.MatchString(username)
}

func validateOrganizationID(fl v10.FieldLevel) bool {
	orgID := fl.Field().String()
	if orgID == "" {
		return true // Allow empty values (omitempty should handle this)
	}

	// Organization ID must be in format: ORGN followed by exactly 8 digits
	// Example: ORGN00000001, ORGN00000002, etc.
	orgIDRegex := regexp.MustCompile(`^ORGN\d{8}$`)
	return orgIDRegex.MatchString(orgID)
}

func validateGroupID(fl v10.FieldLevel) bool {
	groupID := fl.Field().String()
	if groupID == "" {
		return true // Allow empty values (omitempty should handle this)
	}

	// Group ID can be in two formats:
	// 1. GRP followed by exactly 8 digits (old format): GRP00000001
	// 2. GRPN followed by exactly 8 digits (new format): GRPN00000001
	// Accept both for backward compatibility during migration
	groupIDRegex := regexp.MustCompile(`^GRP[N]?\d{8}$`)
	return groupIDRegex.MatchString(groupID)
}

func validateUserID(fl v10.FieldLevel) bool {
	userID := fl.Field().String()
	if userID == "" {
		return true // Allow empty values (omitempty should handle this)
	}

	// User ID must be in format: USER followed by exactly 8 digits
	// Example: USER00000001, USER00000002, etc.
	userIDRegex := regexp.MustCompile(`^USER\d{8}$`)
	return userIDRegex.MatchString(userID)
}

func validateRoleID(fl v10.FieldLevel) bool {
	roleID := fl.Field().String()
	if roleID == "" {
		return true // Allow empty values (omitempty should handle this)
	}

	// Role ID must be in format: ROLE followed by exactly 8 digits
	// Example: ROLE00000001, ROLE00000002, etc.
	roleIDRegex := regexp.MustCompile(`^ROLE\d{8}$`)
	return roleIDRegex.MatchString(roleID)
}
