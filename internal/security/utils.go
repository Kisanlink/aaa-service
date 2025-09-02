package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// PasswordStrength represents password strength levels
type PasswordStrength int

const (
	WeakPassword PasswordStrength = iota
	MediumPassword
	StrongPassword
	VeryStrongPassword
)

// SecurityUtils provides security-related utility functions
type SecurityUtils struct {
	passwordHashCost int
	mpinHashCost     int
}

// NewSecurityUtils creates a new SecurityUtils instance
func NewSecurityUtils(passwordHashCost, mpinHashCost int) *SecurityUtils {
	return &SecurityUtils{
		passwordHashCost: passwordHashCost,
		mpinHashCost:     mpinHashCost,
	}
}

// HashPassword securely hashes a password using bcrypt
func (su *SecurityUtils) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), su.passwordHashCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword verifies a password against its hash using constant-time comparison
func (su *SecurityUtils) VerifyPassword(password, hash string) error {
	if password == "" || hash == "" {
		return fmt.Errorf("password and hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed")
	}

	return nil
}

// HashMPin securely hashes an MPIN using bcrypt
func (su *SecurityUtils) HashMPin(mpin string) (string, error) {
	if mpin == "" {
		return "", fmt.Errorf("MPIN cannot be empty")
	}

	// Add salt to MPIN to prevent rainbow table attacks
	saltedMPin := fmt.Sprintf("mpin_%s_%d", mpin, time.Now().UnixNano()%1000)

	hash, err := bcrypt.GenerateFromPassword([]byte(saltedMPin), su.mpinHashCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash MPIN: %w", err)
	}

	return string(hash), nil
}

// VerifyMPin verifies an MPIN against its hash
func (su *SecurityUtils) VerifyMPin(mpin, hash string) error {
	if mpin == "" || hash == "" {
		return fmt.Errorf("MPIN and hash cannot be empty")
	}

	// Try different salt variations (since we add timestamp-based salt)
	for i := 0; i < 1000; i++ {
		saltedMPin := fmt.Sprintf("mpin_%s_%d", mpin, i)
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedMPin))
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("MPIN verification failed")
}

// GenerateSecureToken generates a cryptographically secure random token
func (su *SecurityUtils) GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("token length must be positive")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateSecureNumericCode generates a secure numeric code (for OTP, etc.)
func (su *SecurityUtils) GenerateSecureNumericCode(length int) (string, error) {
	if length <= 0 || length > 10 {
		return "", fmt.Errorf("code length must be between 1 and 10")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure code: %w", err)
	}

	code := ""
	for _, b := range bytes {
		code += fmt.Sprintf("%d", int(b)%10)
	}

	return code, nil
}

// AssessPasswordStrength assesses the strength of a password
func (su *SecurityUtils) AssessPasswordStrength(password string) PasswordStrength {
	if len(password) < 8 {
		return WeakPassword
	}

	score := 0

	// Length bonus
	if len(password) >= 12 {
		score += 2
	} else if len(password) >= 10 {
		score += 1
	}

	// Character variety
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`\d`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;':"\\|,.<>?~]`).MatchString(password) {
		score++
	}

	// Penalty for common patterns
	if su.hasCommonPatterns(password) {
		score -= 2
	}

	// Penalty for dictionary words (simplified check)
	if su.containsDictionaryWords(password) {
		score -= 1
	}

	switch {
	case score >= 6:
		return VeryStrongPassword
	case score >= 4:
		return StrongPassword
	case score >= 2:
		return MediumPassword
	default:
		return WeakPassword
	}
}

// ValidateMPinStrength validates MPIN strength
func (su *SecurityUtils) ValidateMPinStrength(mpin string) error {
	if len(mpin) != 4 && len(mpin) != 6 {
		return fmt.Errorf("MPIN must be 4 or 6 digits")
	}

	// Check if all digits
	if !regexp.MustCompile(`^\d+$`).MatchString(mpin) {
		return fmt.Errorf("MPIN must contain only digits")
	}

	// Check for weak patterns
	if su.isWeakMPin(mpin) {
		return fmt.Errorf("MPIN is too weak")
	}

	return nil
}

// SanitizeInput sanitizes user input to prevent injection attacks
func (su *SecurityUtils) SanitizeInput(input string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline, carriage return, and tab
	result := ""
	for _, r := range sanitized {
		if r >= 32 || r == '\n' || r == '\r' || r == '\t' {
			result += string(r)
		}
	}

	return strings.TrimSpace(result)
}

// IsValidEmail validates email format
func (su *SecurityUtils) IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidPhoneNumber validates phone number format
func (su *SecurityUtils) IsValidPhoneNumber(phone string) bool {
	// Remove non-digit characters
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Check length and format
	if len(digits) == 10 {
		// Indian mobile number
		return regexp.MustCompile(`^[6-9]\d{9}$`).MatchString(digits)
	} else if len(digits) >= 10 && len(digits) <= 15 {
		// International format
		return true
	}

	return false
}

// SecureCompare performs constant-time string comparison
func (su *SecurityUtils) SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// hasCommonPatterns checks for common weak password patterns
func (su *SecurityUtils) hasCommonPatterns(password string) bool {
	lowerPassword := strings.ToLower(password)

	// Common patterns
	patterns := []string{
		"123", "abc", "qwe", "asd", "zxc",
		"password", "admin", "user", "login",
	}

	for _, pattern := range patterns {
		if strings.Contains(lowerPassword, pattern) {
			return true
		}
	}

	// Sequential characters
	for i := 0; i < len(password)-2; i++ {
		if password[i]+1 == password[i+1] && password[i+1]+1 == password[i+2] {
			return true
		}
	}

	return false
}

// containsDictionaryWords checks for common dictionary words (simplified)
func (su *SecurityUtils) containsDictionaryWords(password string) bool {
	commonWords := []string{
		"password", "admin", "user", "login", "welcome",
		"hello", "world", "test", "demo", "sample",
	}

	lowerPassword := strings.ToLower(password)
	for _, word := range commonWords {
		if strings.Contains(lowerPassword, word) {
			return true
		}
	}

	return false
}

// isWeakMPin checks for weak MPIN patterns
func (su *SecurityUtils) isWeakMPin(mpin string) bool {
	// All same digits
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

	// Sequential patterns
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

	// Reverse sequential
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

	// Common weak patterns
	weakPatterns := []string{
		"1234", "4321", "0000", "1111", "2222", "3333",
		"4444", "5555", "6666", "7777", "8888", "9999",
		"123456", "654321", "000000", "111111",
	}

	for _, pattern := range weakPatterns {
		if mpin == pattern {
			return true
		}
	}

	return false
}

// GetPasswordStrengthString returns a human-readable password strength
func GetPasswordStrengthString(strength PasswordStrength) string {
	switch strength {
	case WeakPassword:
		return "Weak"
	case MediumPassword:
		return "Medium"
	case StrongPassword:
		return "Strong"
	case VeryStrongPassword:
		return "Very Strong"
	default:
		return "Unknown"
	}
}
