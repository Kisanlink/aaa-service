package services

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/responses"
)

// ResponseValidator validates response structures for consistency
type ResponseValidator struct{}

// NewResponseValidator creates a new ResponseValidator instance
func NewResponseValidator() *ResponseValidator {
	return &ResponseValidator{}
}

// ValidateUserResponse validates a StandardUserResponse for consistency
func (rv *ResponseValidator) ValidateUserResponse(response interface{}) error {
	userResponse, ok := response.(*responses.StandardUserResponse)
	if !ok {
		return fmt.Errorf("expected *StandardUserResponse, got %T", response)
	}
	if response == nil {
		return fmt.Errorf("response cannot be nil")
	}

	// Validate required fields
	if userResponse.ID == "" {
		return fmt.Errorf("user ID is required")
	}

	if userResponse.PhoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}

	if userResponse.CountryCode == "" {
		return fmt.Errorf("country code is required")
	}

	// Validate field naming consistency (snake_case)
	if err := rv.validateFieldNaming(userResponse); err != nil {
		return fmt.Errorf("field naming validation failed: %w", err)
	}

	// Validate state consistency
	if err := rv.validateUserStateConsistency(userResponse); err != nil {
		return fmt.Errorf("state consistency validation failed: %w", err)
	}

	return nil
}

// ValidateRoleResponse validates a StandardRoleResponse for consistency
func (rv *ResponseValidator) ValidateRoleResponse(response interface{}) error {
	roleResponse, ok := response.(*responses.StandardRoleResponse)
	if !ok {
		return fmt.Errorf("expected *StandardRoleResponse, got %T", response)
	}
	if response == nil {
		return fmt.Errorf("response cannot be nil")
	}

	// Validate required fields
	if roleResponse.ID == "" {
		return fmt.Errorf("role ID is required")
	}

	if roleResponse.Name == "" {
		return fmt.Errorf("role name is required")
	}

	// Validate field naming consistency
	if err := rv.validateFieldNaming(roleResponse); err != nil {
		return fmt.Errorf("field naming validation failed: %w", err)
	}

	// Validate state consistency
	if err := rv.validateStateConsistency(roleResponse.IsActive, roleResponse.DeletedAt); err != nil {
		return fmt.Errorf("state consistency validation failed: %w", err)
	}

	return nil
}

// ValidateUserRoleResponse validates a StandardUserRoleResponse for consistency
func (rv *ResponseValidator) ValidateUserRoleResponse(response interface{}) error {
	userRoleResponse, ok := response.(*responses.StandardUserRoleResponse)
	if !ok {
		return fmt.Errorf("expected *StandardUserRoleResponse, got %T", response)
	}
	if response == nil {
		return fmt.Errorf("response cannot be nil")
	}

	// Validate required fields
	if userRoleResponse.ID == "" {
		return fmt.Errorf("user role ID is required")
	}

	if userRoleResponse.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if userRoleResponse.RoleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Validate field naming consistency
	if err := rv.validateFieldNaming(userRoleResponse); err != nil {
		return fmt.Errorf("field naming validation failed: %w", err)
	}

	// Validate state consistency
	if err := rv.validateStateConsistency(userRoleResponse.IsActive, userRoleResponse.DeletedAt); err != nil {
		return fmt.Errorf("state consistency validation failed: %w", err)
	}

	// Validate nested objects if present
	if userRoleResponse.User != nil {
		if err := rv.ValidateUserResponse(userRoleResponse.User); err != nil {
			return fmt.Errorf("nested user validation failed: %w", err)
		}
	}

	if userRoleResponse.Role != nil {
		if err := rv.ValidateRoleResponse(userRoleResponse.Role); err != nil {
			return fmt.Errorf("nested role validation failed: %w", err)
		}
	}

	return nil
}

// ValidateResponseConsistency validates that multiple responses have consistent structures
func (rv *ResponseValidator) ValidateResponseConsistency(responses []interface{}) error {
	if len(responses) < 2 {
		return nil // Nothing to compare
	}

	// Group responses by type
	responsesByType := make(map[string][]interface{})
	for _, response := range responses {
		responseType := reflect.TypeOf(response).String()
		responsesByType[responseType] = append(responsesByType[responseType], response)
	}

	// Validate consistency within each type
	for responseType, typeResponses := range responsesByType {
		if len(typeResponses) < 2 {
			continue
		}

		if err := rv.validateTypeConsistency(responseType, typeResponses); err != nil {
			return fmt.Errorf("consistency validation failed for type %s: %w", responseType, err)
		}
	}

	return nil
}

// ValidateNoSensitiveData ensures no sensitive data is present in responses
func (rv *ResponseValidator) ValidateNoSensitiveData(response interface{}) error {
	sensitiveFields := []string{
		"password", "Password", "PASSWORD",
		"mpin", "MPin", "MPIN", "m_pin",
		"secret", "Secret", "SECRET",
		"token", "Token", "TOKEN", // Except for public tokens like access tokens
		"key", "Key", "KEY",
	}

	responseValue := reflect.ValueOf(response)
	if responseValue.Kind() == reflect.Ptr {
		responseValue = responseValue.Elem()
	}

	if responseValue.Kind() != reflect.Struct {
		return nil
	}

	responseType := responseValue.Type()
	for i := 0; i < responseValue.NumField(); i++ {
		field := responseType.Field(i)
		fieldValue := responseValue.Field(i)

		// Check if field name contains sensitive keywords
		for _, sensitiveField := range sensitiveFields {
			if strings.Contains(field.Name, sensitiveField) {
				// Allow certain exceptions
				if field.Name == "HasMPin" || field.Name == "has_mpin" {
					continue // This is safe - just indicates presence
				}

				if !fieldValue.IsZero() {
					return fmt.Errorf("sensitive field %s contains data", field.Name)
				}
			}
		}

		// Recursively check nested structs
		if fieldValue.Kind() == reflect.Struct {
			if err := rv.ValidateNoSensitiveData(fieldValue.Interface()); err != nil {
				return err
			}
		} else if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			if err := rv.ValidateNoSensitiveData(fieldValue.Interface()); err != nil {
				return err
			}
		} else if fieldValue.Kind() == reflect.Slice {
			for j := 0; j < fieldValue.Len(); j++ {
				if err := rv.ValidateNoSensitiveData(fieldValue.Index(j).Interface()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Helper methods

// validateFieldNaming validates that field names follow snake_case convention in JSON tags
func (rv *ResponseValidator) validateFieldNaming(response interface{}) error {
	responseValue := reflect.ValueOf(response)
	if responseValue.Kind() == reflect.Ptr {
		responseValue = responseValue.Elem()
	}

	if responseValue.Kind() != reflect.Struct {
		return nil
	}

	responseType := responseValue.Type()
	for i := 0; i < responseValue.NumField(); i++ {
		field := responseType.Field(i)
		jsonTag := field.Tag.Get("json")

		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Extract field name from JSON tag (remove omitempty, etc.)
		jsonFieldName := strings.Split(jsonTag, ",")[0]

		// Validate snake_case naming
		if !rv.isSnakeCase(jsonFieldName) {
			return fmt.Errorf("field %s has non-snake_case JSON tag: %s", field.Name, jsonFieldName)
		}
	}

	return nil
}

// validateUserStateConsistency validates user-specific state consistency
func (rv *ResponseValidator) validateUserStateConsistency(response *responses.StandardUserResponse) error {
	// Validate is_active and deleted_at consistency
	if err := rv.validateStateConsistency(response.IsActive, response.DeletedAt); err != nil {
		return err
	}

	// Validate status and is_active consistency
	if response.Status != nil {
		if *response.Status == "active" && !response.IsActive {
			return fmt.Errorf("user status is 'active' but is_active is false")
		}
		if *response.Status != "active" && response.IsActive && response.DeletedAt == nil {
			// This might be acceptable in some cases, so just log it
			// return fmt.Errorf("user status is '%s' but is_active is true", *response.Status)
		}
	}

	return nil
}

// validateStateConsistency validates is_active and deleted_at field consistency
func (rv *ResponseValidator) validateStateConsistency(isActive bool, deletedAt *time.Time) error {
	if deletedAt != nil && isActive {
		return fmt.Errorf("entity is marked as active but has deleted_at timestamp")
	}

	// Note: An entity can be inactive without being deleted (e.g., suspended)
	// So we don't validate the reverse case

	return nil
}

// validateTypeConsistency validates that responses of the same type have consistent structures
func (rv *ResponseValidator) validateTypeConsistency(responseType string, responses []interface{}) error {
	if len(responses) < 2 {
		return nil
	}

	// Compare the first response with all others
	firstResponse := responses[0]
	firstValue := reflect.ValueOf(firstResponse)
	if firstValue.Kind() == reflect.Ptr {
		firstValue = firstValue.Elem()
	}

	for i := 1; i < len(responses); i++ {
		otherResponse := responses[i]
		otherValue := reflect.ValueOf(otherResponse)
		if otherValue.Kind() == reflect.Ptr {
			otherValue = otherValue.Elem()
		}

		// Compare types
		if firstValue.Type() != otherValue.Type() {
			return fmt.Errorf("response %d has different type than first response", i)
		}

		// Compare field presence (not values, just structure)
		if err := rv.compareStructureConsistency(firstValue, otherValue, i); err != nil {
			return err
		}
	}

	return nil
}

// compareStructureConsistency compares the structure of two responses
func (rv *ResponseValidator) compareStructureConsistency(first, other reflect.Value, index int) error {
	if first.Type() != other.Type() {
		return fmt.Errorf("response %d has different structure", index)
	}

	// For now, just ensure they have the same type
	// More detailed structure comparison can be added if needed

	return nil
}

// isSnakeCase checks if a string follows snake_case convention
func (rv *ResponseValidator) isSnakeCase(s string) bool {
	if s == "" {
		return true
	}

	// Allow single words in lowercase
	if !strings.Contains(s, "_") {
		return strings.ToLower(s) == s
	}

	// Check each part separated by underscores
	parts := strings.Split(s, "_")
	for _, part := range parts {
		if part == "" {
			return false // Empty part (consecutive underscores)
		}
		if strings.ToLower(part) != part {
			return false // Part contains uppercase letters
		}
	}

	return true
}
