package utils

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/gin-gonic/gin"
)

// Validator provides validation utilities for requests
type Validator struct {
	emailRegex    *regexp.Regexp
	phoneRegex    *regexp.Regexp
	userIDRegex   *regexp.Regexp
	usernameRegex *regexp.Regexp
}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{
		emailRegex:    regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		phoneRegex:    regexp.MustCompile(`^\+?[1-9]\d{1,14}$`),
		userIDRegex:   regexp.MustCompile(`^USER[A-Z0-9]{8}$`),
		usernameRegex: regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`),
	}
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if !v.emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// ValidatePhone validates phone number format
func (v *Validator) ValidatePhone(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}
	if !v.phoneRegex.MatchString(phone) {
		return errors.New("invalid phone number format")
	}
	return nil
}

// ValidateUserID validates user ID format
func (v *Validator) ValidateUserID(userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	if !v.userIDRegex.MatchString(userID) {
		return errors.New("invalid user ID format")
	}
	return nil
}

// ValidateUsername validates username format
func (v *Validator) ValidateUsername(username string) error {
	if username == "" {
		return errors.New("username is required")
	}
	if !v.usernameRegex.MatchString(username) {
		return errors.New("username must be 3-20 characters long and contain only letters, numbers, and underscores")
	}
	return nil
}

// ValidateName validates name format
func (v *Validator) ValidateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}
	if len(name) < 2 || len(name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}
	if strings.TrimSpace(name) != name {
		return errors.New("name cannot start or end with whitespace")
	}
	return nil
}

// ValidateStatus validates user status
func (v *Validator) ValidateStatus(status string) error {
	if status == "" {
		return nil // Status is optional
	}
	validStatuses := []string{"active", "inactive", "pending", "suspended"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}
	return errors.New("invalid status value")
}

// ParseListFilters parses query parameters into base.Filters
func (v *Validator) ParseListFilters(c *gin.Context) (*base.Filters, error) {
	filters := &base.Filters{
		Conditions: make(map[string]interface{}),
	}

	// Parse pagination parameters
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return nil, errors.New("invalid page parameter")
		}
		filters.Offset = (page - 1) * filters.Limit
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return nil, errors.New("invalid limit parameter (must be between 1 and 100)")
		}
		filters.Limit = limit
	} else {
		filters.Limit = 10 // Default limit
	}

	// Parse filter parameters
	if status := c.Query("status"); status != "" {
		if err := v.ValidateStatus(status); err != nil {
			return nil, err
		}
		filters.Conditions["status"] = status
	}

	if search := c.Query("search"); search != "" {
		if len(search) < 2 {
			return nil, errors.New("search term must be at least 2 characters")
		}
		filters.Conditions["search"] = search
	}

	// Parse sorting parameters
	if sortBy := c.Query("sort_by"); sortBy != "" {
		validSortFields := []string{"name", "email", "created_at", "updated_at"}
		isValid := false
		for _, field := range validSortFields {
			if sortBy == field {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, errors.New("invalid sort_by parameter")
		}
		filters.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		if sortOrder != "asc" && sortOrder != "desc" {
			return nil, errors.New("invalid sort_order parameter (must be 'asc' or 'desc')")
		}
		filters.SortOrder = sortOrder
	} else {
		filters.SortOrder = "desc" // Default sort order
	}

	return filters, nil
}

// ValidatePagination validates pagination parameters
func (v *Validator) ValidatePagination(page, limit int) error {
	if page < 1 {
		return errors.New("page must be greater than 0")
	}
	if limit < 1 || limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}
	return nil
}

// ValidateRequiredField validates that a required field is not empty
func (v *Validator) ValidateRequiredField(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(fieldName + " is required")
	}
	return nil
}

// ValidateStringLength validates string length constraints
func (v *Validator) ValidateStringLength(value, fieldName string, min, max int) error {
	length := len(strings.TrimSpace(value))
	if length < min || length > max {
		return errors.New(fieldName + " must be between " + strconv.Itoa(min) + " and " + strconv.Itoa(max) + " characters")
	}
	return nil
}
