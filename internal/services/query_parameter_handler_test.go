package services

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryParameterHandlerImpl_ParseTransformOptions(t *testing.T) {
	handler := NewQueryParameterHandler()

	tests := []struct {
		name     string
		query    string
		expected func(t *testing.T, options interface{})
	}{
		{
			name:  "Default options",
			query: "",
			expected: func(t *testing.T, options interface{}) {
				opts := options.(interfaces.TransformOptions)
				assert.False(t, opts.IncludeProfile)
				assert.False(t, opts.IncludeContacts)
				assert.False(t, opts.IncludeRole)
				assert.True(t, opts.ExcludeDeleted)
				assert.True(t, opts.MaskSensitiveData)
			},
		},
		{
			name:  "Include profile",
			query: "include_profile=true",
			expected: func(t *testing.T, options interface{}) {
				opts := options.(interfaces.TransformOptions)
				assert.True(t, opts.IncludeProfile)
			},
		},
		{
			name:  "Include multiple options",
			query: "include_profile=true&include_contacts=1&include_role=yes",
			expected: func(t *testing.T, options interface{}) {
				opts := options.(interfaces.TransformOptions)
				assert.True(t, opts.IncludeProfile)
				assert.True(t, opts.IncludeContacts)
				assert.True(t, opts.IncludeRole)
			},
		},
		{
			name:  "Exclude options",
			query: "exclude_deleted=false&exclude_inactive=true",
			expected: func(t *testing.T, options interface{}) {
				opts := options.(interfaces.TransformOptions)
				assert.False(t, opts.ExcludeDeleted)
				assert.True(t, opts.ExcludeInactive)
			},
		},
		{
			name:  "Legacy include parameter",
			query: "include=profile,contacts,role",
			expected: func(t *testing.T, options interface{}) {
				opts := options.(interfaces.TransformOptions)
				assert.True(t, opts.IncludeProfile)
				assert.True(t, opts.IncludeContacts)
				assert.True(t, opts.IncludeRole)
			},
		},
		{
			name:  "Legacy include all",
			query: "include=all",
			expected: func(t *testing.T, options interface{}) {
				opts := options.(interfaces.TransformOptions)
				assert.True(t, opts.IncludeProfile)
				assert.True(t, opts.IncludeContacts)
				assert.True(t, opts.IncludeRole)
				assert.True(t, opts.IncludeAddress)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test context with query parameters
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := &http.Request{
				URL: &url.URL{
					RawQuery: tt.query,
				},
			}
			c.Request = req

			options := handler.ParseTransformOptions(c)
			tt.expected(t, options)
		})
	}
}

func TestQueryParameterHandlerImpl_ValidateQueryParameters(t *testing.T) {
	handler := NewQueryParameterHandler()

	tests := []struct {
		name        string
		query       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid parameters",
			query:       "include_profile=true&limit=10&offset=0",
			expectError: false,
		},
		{
			name:        "Invalid parameter",
			query:       "invalid_param=true",
			expectError: true,
			errorMsg:    "invalid query parameter: invalid_param",
		},
		{
			name:        "Invalid boolean value",
			query:       "include_profile=maybe",
			expectError: true,
			errorMsg:    "invalid boolean value for parameter include_profile: maybe",
		},
		{
			name:        "Invalid numeric value",
			query:       "limit=abc",
			expectError: true,
			errorMsg:    "invalid numeric value for parameter limit: abc",
		},
		{
			name:        "Valid boolean variations",
			query:       "include_profile=1&include_contacts=yes&include_role=on&exclude_deleted=0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test context with query parameters
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := &http.Request{
				URL: &url.URL{
					RawQuery: tt.query,
				},
			}
			c.Request = req

			err := handler.ValidateQueryParameters(c)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQueryParameterHandlerImpl_GetPaginationParams(t *testing.T) {
	handler := &QueryParameterHandlerImpl{}

	tests := []struct {
		name           string
		query          string
		expectedLimit  int
		expectedOffset int
		expectError    bool
	}{
		{
			name:           "Default values",
			query:          "",
			expectedLimit:  20,
			expectedOffset: 0,
			expectError:    false,
		},
		{
			name:           "Custom limit and offset",
			query:          "limit=50&offset=100",
			expectedLimit:  50,
			expectedOffset: 100,
			expectError:    false,
		},
		{
			name:           "Page-based pagination",
			query:          "page=3&limit=10",
			expectedLimit:  10,
			expectedOffset: 20, // (page-1) * limit = (3-1) * 10 = 20
			expectError:    false,
		},
		{
			name:        "Invalid limit",
			query:       "limit=0",
			expectError: true,
		},
		{
			name:        "Limit too large",
			query:       "limit=2000",
			expectError: true,
		},
		{
			name:        "Negative offset",
			query:       "offset=-1",
			expectError: true,
		},
		{
			name:        "Invalid page",
			query:       "page=0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test context with query parameters
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := &http.Request{
				URL: &url.URL{
					RawQuery: tt.query,
				},
			}
			c.Request = req

			limit, offset, err := handler.GetPaginationParams(c)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLimit, limit)
				assert.Equal(t, tt.expectedOffset, offset)
			}
		})
	}
}

func TestQueryParameterHandlerImpl_GetSortParams(t *testing.T) {
	handler := &QueryParameterHandlerImpl{}

	tests := []struct {
		name           string
		query          string
		expectedSortBy string
		expectedOrder  string
	}{
		{
			name:           "Default values",
			query:          "",
			expectedSortBy: "created_at",
			expectedOrder:  "desc",
		},
		{
			name:           "Custom sort and order",
			query:          "sort=name&order=asc",
			expectedSortBy: "name",
			expectedOrder:  "asc",
		},
		{
			name:           "Invalid order defaults to desc",
			query:          "sort=name&order=invalid",
			expectedSortBy: "name",
			expectedOrder:  "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test context with query parameters
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := &http.Request{
				URL: &url.URL{
					RawQuery: tt.query,
				},
			}
			c.Request = req

			sortBy, order := handler.GetSortParams(c)
			assert.Equal(t, tt.expectedSortBy, sortBy)
			assert.Equal(t, tt.expectedOrder, order)
		})
	}
}

func TestQueryParameterHandlerImpl_GetSearchParam(t *testing.T) {
	handler := &QueryParameterHandlerImpl{}

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "No search parameter",
			query:    "",
			expected: "",
		},
		{
			name:     "Search parameter",
			query:    "search=test query",
			expected: "test query",
		},
		{
			name:     "Search parameter with spaces",
			query:    "search=  test query  ",
			expected: "test query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test context with query parameters
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := &http.Request{
				URL: &url.URL{
					RawQuery: tt.query,
				},
			}
			c.Request = req

			result := handler.GetSearchParam(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryParameterHandlerImpl_GetFilterParams(t *testing.T) {
	handler := &QueryParameterHandlerImpl{}

	tests := []struct {
		name     string
		query    string
		expected map[string]string
	}{
		{
			name:     "No filter parameters",
			query:    "",
			expected: map[string]string{},
		},
		{
			name:  "Single filter parameter",
			query: "status=active",
			expected: map[string]string{
				"status": "active",
			},
		},
		{
			name:  "Multiple filter parameters",
			query: "status=active&is_validated=true&role_id=123",
			expected: map[string]string{
				"status":       "active",
				"is_validated": "true",
				"role_id":      "123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test context with query parameters
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := &http.Request{
				URL: &url.URL{
					RawQuery: tt.query,
				},
			}
			c.Request = req

			result := handler.GetFilterParams(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryParameterHandlerImpl_ParseBoolParam(t *testing.T) {
	handler := &QueryParameterHandlerImpl{}

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true", "true", true},
		{"True", "True", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"YES", "YES", true},
		{"on", "on", true},
		{"ON", "ON", true},
		{"false", "false", false},
		{"False", "False", false},
		{"0", "0", false},
		{"no", "no", false},
		{"NO", "NO", false},
		{"off", "off", false},
		{"OFF", "OFF", false},
		{"invalid", "invalid", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.parseBoolParam(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryParameterHandlerImpl_IsValidBoolParam(t *testing.T) {
	handler := &QueryParameterHandlerImpl{}

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true", "true", true},
		{"false", "false", true},
		{"1", "1", true},
		{"0", "0", true},
		{"yes", "yes", true},
		{"no", "no", true},
		{"on", "on", true},
		{"off", "off", true},
		{"True", "True", true},
		{"FALSE", "FALSE", true},
		{"invalid", "invalid", false},
		{"maybe", "maybe", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isValidBoolParam(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
