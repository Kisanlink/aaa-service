package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/gin-gonic/gin"
)

// QueryParameterHandlerImpl implements the QueryParameterHandler interface
type QueryParameterHandlerImpl struct{}

// NewQueryParameterHandler creates a new QueryParameterHandler instance
func NewQueryParameterHandler() interfaces.QueryParameterHandler {
	return &QueryParameterHandlerImpl{}
}

// ParseTransformOptions parses query parameters into TransformOptions
func (qph *QueryParameterHandlerImpl) ParseTransformOptions(c *gin.Context) interfaces.TransformOptions {
	options := qph.GetDefaultOptions()

	// Parse include flags
	if includeProfile := c.Query("include_profile"); includeProfile != "" {
		options.IncludeProfile = qph.parseBoolParam(includeProfile)
	}

	if includeContacts := c.Query("include_contacts"); includeContacts != "" {
		options.IncludeContacts = qph.parseBoolParam(includeContacts)
	}

	if includeRole := c.Query("include_role"); includeRole != "" {
		options.IncludeRole = qph.parseBoolParam(includeRole)
	}

	if includeUser := c.Query("include_user"); includeUser != "" {
		options.IncludeUser = qph.parseBoolParam(includeUser)
	}

	if includeAddress := c.Query("include_address"); includeAddress != "" {
		options.IncludeAddress = qph.parseBoolParam(includeAddress)
	}

	// Parse exclusion flags
	if excludeDeleted := c.Query("exclude_deleted"); excludeDeleted != "" {
		options.ExcludeDeleted = qph.parseBoolParam(excludeDeleted)
	}

	if excludeInactive := c.Query("exclude_inactive"); excludeInactive != "" {
		options.ExcludeInactive = qph.parseBoolParam(excludeInactive)
	}

	if onlyActiveRoles := c.Query("only_active_roles"); onlyActiveRoles != "" {
		options.OnlyActiveRoles = qph.parseBoolParam(onlyActiveRoles)
	}

	// Parse field control flags
	if maskSensitive := c.Query("mask_sensitive"); maskSensitive != "" {
		options.MaskSensitiveData = qph.parseBoolParam(maskSensitive)
	}

	if includeTimestamps := c.Query("include_timestamps"); includeTimestamps != "" {
		options.IncludeTimestamps = qph.parseBoolParam(includeTimestamps)
	}

	// Handle legacy include parameter for backward compatibility
	if include := c.Query("include"); include != "" {
		qph.parseLegacyIncludeParam(include, &options)
	}

	return options
}

// ValidateQueryParameters validates that query parameters are valid
func (qph *QueryParameterHandlerImpl) ValidateQueryParameters(c *gin.Context) error {
	allowedParams := map[string]bool{
		"include_profile":    true,
		"include_contacts":   true,
		"include_role":       true,
		"include_user":       true,
		"include_address":    true,
		"exclude_deleted":    true,
		"exclude_inactive":   true,
		"only_active_roles":  true,
		"mask_sensitive":     true,
		"include_timestamps": true,
		"include":            true, // Legacy parameter
		"limit":              true,
		"offset":             true,
		"page":               true,
		"sort":               true,
		"order":              true,
		"search":             true,
		"filter":             true,
	}

	for param := range c.Request.URL.Query() {
		if !allowedParams[param] {
			return fmt.Errorf("invalid query parameter: %s", param)
		}
	}

	// Validate boolean parameters
	boolParams := []string{
		"include_profile", "include_contacts", "include_role", "include_user", "include_address",
		"exclude_deleted", "exclude_inactive", "only_active_roles", "mask_sensitive", "include_timestamps",
	}

	for _, param := range boolParams {
		if value := c.Query(param); value != "" {
			if !qph.isValidBoolParam(value) {
				return fmt.Errorf("invalid boolean value for parameter %s: %s (expected: true, false, 1, 0)", param, value)
			}
		}
	}

	// Validate numeric parameters
	numericParams := []string{"limit", "offset", "page"}
	for _, param := range numericParams {
		if value := c.Query(param); value != "" {
			if _, err := strconv.Atoi(value); err != nil {
				return fmt.Errorf("invalid numeric value for parameter %s: %s", param, value)
			}
		}
	}

	return nil
}

// GetDefaultOptions returns the default transformation options
func (qph *QueryParameterHandlerImpl) GetDefaultOptions() interfaces.TransformOptions {
	return interfaces.TransformOptions{
		// Default to including basic nested objects
		IncludeProfile:  false,
		IncludeContacts: false,
		IncludeRole:     false,
		IncludeUser:     false,
		IncludeAddress:  false,

		// Default exclusion settings
		ExcludeDeleted:  true,  // By default, exclude soft-deleted records
		ExcludeInactive: false, // Include inactive records by default
		OnlyActiveRoles: false, // Include all roles by default

		// Default field control
		MaskSensitiveData: true, // Always mask sensitive data by default
		IncludeTimestamps: true, // Include timestamps by default
	}
}

// Helper methods

// parseBoolParam parses a string parameter to boolean
func (qph *QueryParameterHandlerImpl) parseBoolParam(value string) bool {
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return false // Default to false for invalid values
	}
}

// isValidBoolParam checks if a string is a valid boolean parameter
func (qph *QueryParameterHandlerImpl) isValidBoolParam(value string) bool {
	validValues := []string{"true", "false", "1", "0", "yes", "no", "on", "off"}
	lowerValue := strings.ToLower(value)

	for _, valid := range validValues {
		if lowerValue == valid {
			return true
		}
	}

	return false
}

// parseLegacyIncludeParam parses the legacy "include" parameter for backward compatibility
func (qph *QueryParameterHandlerImpl) parseLegacyIncludeParam(include string, options *interfaces.TransformOptions) {
	includes := strings.Split(include, ",")

	for _, item := range includes {
		item = strings.TrimSpace(strings.ToLower(item))

		switch item {
		case "profile":
			options.IncludeProfile = true
		case "contacts":
			options.IncludeContacts = true
		case "roles", "role":
			options.IncludeRole = true
		case "user":
			options.IncludeUser = true
		case "address":
			options.IncludeAddress = true
		case "all":
			options.IncludeProfile = true
			options.IncludeContacts = true
			options.IncludeRole = true
			options.IncludeAddress = true
		}
	}
}

// GetPaginationParams extracts pagination parameters from query string
func (qph *QueryParameterHandlerImpl) GetPaginationParams(c *gin.Context) (limit, offset int, err error) {
	// Default values
	limit = 20
	offset = 0

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, parseErr := strconv.Atoi(limitStr); parseErr != nil {
			return 0, 0, fmt.Errorf("invalid limit parameter: %s", limitStr)
		} else if parsedLimit < 1 {
			return 0, 0, fmt.Errorf("limit must be greater than 0")
		} else if parsedLimit > 1000 {
			return 0, 0, fmt.Errorf("limit cannot exceed 1000")
		} else {
			limit = parsedLimit
		}
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, parseErr := strconv.Atoi(offsetStr); parseErr != nil {
			return 0, 0, fmt.Errorf("invalid offset parameter: %s", offsetStr)
		} else if parsedOffset < 0 {
			return 0, 0, fmt.Errorf("offset cannot be negative")
		} else {
			offset = parsedOffset
		}
	}

	// Handle page-based pagination (alternative to offset)
	if pageStr := c.Query("page"); pageStr != "" {
		if page, parseErr := strconv.Atoi(pageStr); parseErr != nil {
			return 0, 0, fmt.Errorf("invalid page parameter: %s", pageStr)
		} else if page < 1 {
			return 0, 0, fmt.Errorf("page must be greater than 0")
		} else {
			offset = (page - 1) * limit
		}
	}

	return limit, offset, nil
}

// GetSortParams extracts sorting parameters from query string
func (qph *QueryParameterHandlerImpl) GetSortParams(c *gin.Context) (sortBy, order string) {
	sortBy = c.DefaultQuery("sort", "created_at")
	order = c.DefaultQuery("order", "desc")

	// Validate order parameter
	if order != "asc" && order != "desc" {
		order = "desc" // Default to desc for invalid values
	}

	return sortBy, order
}

// GetSearchParam extracts search parameter from query string
func (qph *QueryParameterHandlerImpl) GetSearchParam(c *gin.Context) string {
	return strings.TrimSpace(c.Query("search"))
}

// GetFilterParams extracts filter parameters from query string
func (qph *QueryParameterHandlerImpl) GetFilterParams(c *gin.Context) map[string]string {
	filters := make(map[string]string)

	// Common filter parameters
	filterParams := []string{
		"status", "is_active", "is_validated", "role_id", "organization_id", "group_id",
		"created_after", "created_before", "updated_after", "updated_before",
	}

	for _, param := range filterParams {
		if value := c.Query(param); value != "" {
			filters[param] = value
		}
	}

	return filters
}
