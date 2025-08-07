package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIVersionConfig holds configuration for API versioning
type APIVersionConfig struct {
	DefaultVersion     string            `json:"default_version"`
	SupportedVersions  []string          `json:"supported_versions"`
	DeprecatedVersions map[string]string `json:"deprecated_versions"` // version -> deprecation message
}

// NewAPIVersionConfig creates a new API version configuration
func NewAPIVersionConfig() *APIVersionConfig {
	return &APIVersionConfig{
		DefaultVersion:    "v1",
		SupportedVersions: []string{"v1", "v2"},
		DeprecatedVersions: map[string]string{
			"v1": "API v1 is deprecated. Please migrate to v2 for enhanced features and better performance.",
		},
	}
}

// APIVersioning creates middleware for API version handling
func APIVersioning(config *APIVersionConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		version := extractAPIVersion(c, config.DefaultVersion)

		// Validate version
		if !isVersionSupported(version, config.SupportedVersions) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":              "Unsupported API version",
				"supported_versions": config.SupportedVersions,
				"requested_version":  version,
			})
			c.Abort()
			return
		}

		// Set version in context
		c.Set("api_version", version)
		c.Set("handler_version", version)

		// Add deprecation warning for deprecated versions
		if deprecationMsg, isDeprecated := config.DeprecatedVersions[version]; isDeprecated {
			c.Header("X-API-Deprecation-Warning", deprecationMsg)
			c.Header("X-API-Sunset-Date", "2024-12-31") // Example sunset date
		}

		// Add version info to response headers
		c.Header("X-API-Version", version)
		c.Header("X-Supported-Versions", strings.Join(config.SupportedVersions, ","))

		c.Next()
	}
}

// extractAPIVersion extracts API version from various sources
func extractAPIVersion(c *gin.Context, defaultVersion string) string {
	// Priority: Header -> Query Parameter -> URL Path -> Default

	// 1. Check API-Version header
	if version := c.GetHeader("API-Version"); version != "" {
		return normalizeVersion(version)
	}

	// 2. Check Accept header for version (e.g., application/vnd.api+json;version=2)
	if accept := c.GetHeader("Accept"); accept != "" {
		if version := extractVersionFromAcceptHeader(accept); version != "" {
			return normalizeVersion(version)
		}
	}

	// 3. Check version query parameter
	if version := c.Query("version"); version != "" {
		return normalizeVersion(version)
	}

	// 4. Check URL path for version (e.g., /api/v2/users)
	if version := extractVersionFromPath(c.Request.URL.Path); version != "" {
		return normalizeVersion(version)
	}

	// 5. Return default version
	return defaultVersion
}

// extractVersionFromAcceptHeader extracts version from Accept header
func extractVersionFromAcceptHeader(accept string) string {
	// Look for version=X in Accept header
	parts := strings.Split(accept, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "version=") {
			return strings.TrimPrefix(part, "version=")
		}
	}
	return ""
}

// extractVersionFromPath extracts version from URL path
func extractVersionFromPath(path string) string {
	// Look for /api/vX/ or /vX/ patterns
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "v") && len(part) >= 2 {
			// Check if it's a valid version format (v1, v2, etc.)
			version := part[1:]
			if isNumericVersion(version) {
				return part
			}
		}
	}
	return ""
}

// normalizeVersion ensures version is in correct format (v1, v2, etc.)
func normalizeVersion(version string) string {
	version = strings.ToLower(strings.TrimSpace(version))

	// If version doesn't start with 'v', add it
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return version
}

// isVersionSupported checks if version is in supported list
func isVersionSupported(version string, supportedVersions []string) bool {
	for _, supported := range supportedVersions {
		if version == supported {
			return true
		}
	}
	return false
}

// isNumericVersion checks if version part is numeric
func isNumericVersion(version string) bool {
	if len(version) == 0 {
		return false
	}

	for _, char := range version {
		if char < '0' || char > '9' {
			// Allow version like "1.1" or "2.0"
			if char != '.' {
				return false
			}
		}
	}
	return true
}

// GetAPIVersion retrieves API version from gin context
func GetAPIVersion(c *gin.Context) string {
	if version, exists := c.Get("api_version"); exists {
		if v, ok := version.(string); ok {
			return v
		}
	}
	return "v1" // default
}

// IsV2Request checks if the request is for API v2
func IsV2Request(c *gin.Context) bool {
	return GetAPIVersion(c) == "v2"
}

// IsV1Request checks if the request is for API v1 (legacy)
func IsV1Request(c *gin.Context) bool {
	return GetAPIVersion(c) == "v1"
}

// VersionHandler creates a handler that returns API version information
func VersionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		config := NewAPIVersionConfig()

		c.JSON(http.StatusOK, gin.H{
			"current_version":     GetAPIVersion(c),
			"supported_versions":  config.SupportedVersions,
			"deprecated_versions": config.DeprecatedVersions,
			"default_version":     config.DefaultVersion,
		})
	}
}
