package catalog

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// MaxServiceIDLength defines the maximum allowed length for a service ID
	MaxServiceIDLength = 255

	// MinServiceIDLength defines the minimum allowed length for a service ID (when non-empty)
	MinServiceIDLength = 2
)

var (
	// serviceIDPattern validates service ID format: lowercase alphanumeric with hyphens
	// Must start with letter, can contain letters, numbers, hyphens
	// Examples: "farmers-module", "erp-service", "traceability-svc"
	serviceIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)

	// dangerousPatterns detects potential SQL injection or malicious patterns
	dangerousPatterns = []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"exec", "execute", "select", "insert", "update", "delete",
		"drop", "create", "alter", "union", "script",
	}
)

// ValidateServiceID validates and sanitizes a service ID
// Returns error if the service ID is invalid
// Empty service ID is allowed (defaults to farmers-module for backward compatibility)
func ValidateServiceID(serviceID string) error {
	// Allow empty service ID for backward compatibility
	if serviceID == "" {
		return nil
	}

	// Check length constraints
	if len(serviceID) < MinServiceIDLength {
		return fmt.Errorf("service_id too short: minimum %d characters required", MinServiceIDLength)
	}

	if len(serviceID) > MaxServiceIDLength {
		return fmt.Errorf("service_id too long: maximum %d characters allowed", MaxServiceIDLength)
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(serviceID)
	if trimmed != serviceID {
		return fmt.Errorf("service_id contains leading or trailing whitespace")
	}

	// Check for dangerous patterns (SQL injection, XSS, etc.)
	lowerServiceID := strings.ToLower(serviceID)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerServiceID, pattern) {
			return fmt.Errorf("service_id contains forbidden pattern: %s", pattern)
		}
	}

	// Validate format using regex
	if !serviceIDPattern.MatchString(serviceID) {
		return fmt.Errorf("service_id has invalid format: must be lowercase alphanumeric with hyphens, starting with a letter (e.g., 'erp-service', 'farmers-module')")
	}

	// Check for consecutive hyphens
	if strings.Contains(serviceID, "--") {
		return fmt.Errorf("service_id cannot contain consecutive hyphens")
	}

	return nil
}

// SanitizeServiceID sanitizes a service ID by removing potentially dangerous characters
// This is a defensive fallback - validation should prevent invalid IDs from reaching this point
func SanitizeServiceID(serviceID string) string {
	// Trim whitespace
	serviceID = strings.TrimSpace(serviceID)

	// Convert to lowercase
	serviceID = strings.ToLower(serviceID)

	// Remove any character that isn't alphanumeric or hyphen
	var sanitized strings.Builder
	for _, char := range serviceID {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			sanitized.WriteRune(char)
		}
	}

	return sanitized.String()
}

// ValidateServiceName validates a service name
func ValidateServiceName(serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("service_name cannot be empty")
	}

	if len(serviceName) > MaxServiceIDLength {
		return fmt.Errorf("service_name too long: maximum %d characters allowed", MaxServiceIDLength)
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(serviceName)
	if trimmed != serviceName {
		return fmt.Errorf("service_name contains leading or trailing whitespace")
	}

	return nil
}
