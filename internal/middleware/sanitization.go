package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InputSanitizationMiddleware provides comprehensive input sanitization
type InputSanitizationMiddleware struct {
	logger *zap.Logger
}

// NewInputSanitizationMiddleware creates a new input sanitization middleware
func NewInputSanitizationMiddleware(logger *zap.Logger) *InputSanitizationMiddleware {
	return &InputSanitizationMiddleware{
		logger: logger,
	}
}

// SanitizeInput sanitizes all input data in requests
func (m *InputSanitizationMiddleware) SanitizeInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		m.sanitizeQueryParams(c)

		// Sanitize path parameters
		m.sanitizePathParams(c)

		// Sanitize headers (selective)
		m.sanitizeHeaders(c)

		// Sanitize request body for POST/PUT/PATCH requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if err := m.sanitizeRequestBody(c); err != nil {
				m.logger.Warn("Failed to sanitize request body", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid request format",
					"message": "Request contains invalid data",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// sanitizeQueryParams sanitizes URL query parameters
func (m *InputSanitizationMiddleware) sanitizeQueryParams(c *gin.Context) {
	query := c.Request.URL.Query()
	modified := false

	for _, values := range query {
		for i, value := range values {
			sanitized := m.sanitizeString(value)
			if sanitized != value {
				values[i] = sanitized
				modified = true
			}
		}
	}

	if modified {
		c.Request.URL.RawQuery = query.Encode()
	}
}

// sanitizePathParams sanitizes URL path parameters
func (m *InputSanitizationMiddleware) sanitizePathParams(c *gin.Context) {
	for _, param := range c.Params {
		sanitized := m.sanitizeString(param.Value)
		if sanitized != param.Value {
			// Update the parameter value
			for i, p := range c.Params {
				if p.Key == param.Key {
					c.Params[i].Value = sanitized
					break
				}
			}
		}
	}
}

// sanitizeHeaders sanitizes specific headers that might contain user input
func (m *InputSanitizationMiddleware) sanitizeHeaders(c *gin.Context) {
	headersToSanitize := []string{
		"User-Agent",
		"Referer",
		"X-Forwarded-For",
		"X-Real-IP",
	}

	for _, headerName := range headersToSanitize {
		if value := c.GetHeader(headerName); value != "" {
			sanitized := m.sanitizeString(value)
			if sanitized != value {
				c.Request.Header.Set(headerName, sanitized)
			}
		}
	}
}

// sanitizeRequestBody sanitizes JSON request body
func (m *InputSanitizationMiddleware) sanitizeRequestBody(c *gin.Context) error {
	if c.Request.Body == nil {
		return nil
	}

	// Read the body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}

	// Close the original body
	_ = c.Request.Body.Close()

	// If body is empty, restore and continue
	if len(bodyBytes) == 0 {
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return nil
	}

	// Try to parse as JSON
	var jsonData interface{}
	if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
		// If not JSON, treat as plain text and sanitize
		sanitized := m.sanitizeString(string(bodyBytes))
		c.Request.Body = io.NopCloser(strings.NewReader(sanitized))
		return nil
	}

	// Sanitize JSON data recursively
	sanitizedData := m.sanitizeJSONValue(jsonData)

	// Marshal back to JSON
	sanitizedBytes, err := json.Marshal(sanitizedData)
	if err != nil {
		return err
	}

	// Replace the request body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(sanitizedBytes))
	c.Request.ContentLength = int64(len(sanitizedBytes))

	return nil
}

// sanitizeJSONValue recursively sanitizes JSON values
func (m *InputSanitizationMiddleware) sanitizeJSONValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return m.sanitizeString(v)
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, val := range v {
			sanitizedKey := m.sanitizeString(key)
			sanitized[sanitizedKey] = m.sanitizeJSONValue(val)
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, val := range v {
			sanitized[i] = m.sanitizeJSONValue(val)
		}
		return sanitized
	default:
		// For numbers, booleans, null - return as is
		return value
	}
}

// sanitizeString performs comprehensive string sanitization
func (m *InputSanitizationMiddleware) sanitizeString(input string) string {
	if input == "" {
		return input
	}

	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Remove other control characters except allowed ones
	sanitized = m.removeControlCharacters(sanitized)

	// Remove potential XSS patterns
	sanitized = m.removeXSSPatterns(sanitized)

	// Remove potential SQL injection patterns (basic)
	sanitized = m.removeSQLInjectionPatterns(sanitized)

	// Normalize whitespace
	sanitized = m.normalizeWhitespace(sanitized)

	// Limit length to prevent DoS
	if len(sanitized) > 10000 {
		sanitized = sanitized[:10000]
	}

	return sanitized
}

// removeControlCharacters removes dangerous control characters
func (m *InputSanitizationMiddleware) removeControlCharacters(input string) string {
	// Allow: tab (9), newline (10), carriage return (13), and printable characters (32-126)
	var result strings.Builder
	for _, r := range input {
		if r == '\t' || r == '\n' || r == '\r' || (r >= 32 && r <= 126) || r > 126 {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// removeXSSPatterns removes common XSS attack patterns
func (m *InputSanitizationMiddleware) removeXSSPatterns(input string) string {
	// Common XSS patterns to remove/neutralize
	xssPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`),
		regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`),
		regexp.MustCompile(`(?i)<embed[^>]*>`),
		regexp.MustCompile(`(?i)<link[^>]*>`),
		regexp.MustCompile(`(?i)<meta[^>]*>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)vbscript:`),
		regexp.MustCompile(`(?i)data:`),
		regexp.MustCompile(`(?i)on\w+\s*=`), // Event handlers like onclick, onload, etc.
	}

	result := input
	for _, pattern := range xssPatterns {
		result = pattern.ReplaceAllString(result, "")
	}

	return result
}

// removeSQLInjectionPatterns removes basic SQL injection patterns
func (m *InputSanitizationMiddleware) removeSQLInjectionPatterns(input string) string {
	// Basic SQL injection patterns - be careful not to break legitimate content
	sqlPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\s|^)(union\s+select)`),
		regexp.MustCompile(`(?i)(\s|^)(drop\s+table)`),
		regexp.MustCompile(`(?i)(\s|^)(delete\s+from)`),
		regexp.MustCompile(`(?i)(\s|^)(insert\s+into)`),
		regexp.MustCompile(`(?i)(\s|^)(update\s+\w+\s+set)`),
		regexp.MustCompile(`(?i)(\s|^)(exec\s*\()`),
		regexp.MustCompile(`(?i)(\s|^)(execute\s*\()`),
		regexp.MustCompile(`--\s*$`),    // SQL comments at end of line
		regexp.MustCompile(`/\*.*?\*/`), // SQL block comments
	}

	result := input
	for _, pattern := range sqlPatterns {
		result = pattern.ReplaceAllString(result, "")
	}

	return result
}

// normalizeWhitespace normalizes whitespace characters
func (m *InputSanitizationMiddleware) normalizeWhitespace(input string) string {
	// Replace multiple consecutive whitespace with single space
	whitespacePattern := regexp.MustCompile(`\s+`)
	normalized := whitespacePattern.ReplaceAllString(input, " ")

	// Trim leading and trailing whitespace
	return strings.TrimSpace(normalized)
}

// ValidateContentType validates that the content type is allowed
func ValidateContentType(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Missing Content-Type header",
				"message": "Content-Type header is required",
			})
			c.Abort()
			return
		}

		// Check if content type is allowed
		allowed := false
		for _, allowedType := range allowedTypes {
			if strings.HasPrefix(contentType, allowedType) {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error":   "Unsupported Content-Type",
				"message": "Content-Type not supported",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateJSONStructure validates that JSON has expected structure
func ValidateJSONStructure(maxDepth int, maxKeys int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			c.Next()
			return
		}

		if c.Request.Body == nil {
			c.Next()
			return
		}

		// Read body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"message": "Could not read request body",
			})
			c.Abort()
			return
		}

		// Restore body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if len(bodyBytes) == 0 {
			c.Next()
			return
		}

		// Parse JSON
		var jsonData interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON format",
				"message": "Request body contains invalid JSON",
			})
			c.Abort()
			return
		}

		// Validate structure
		if err := validateJSONStructure(jsonData, maxDepth, maxKeys, 0); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON structure",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateJSONStructure recursively validates JSON structure
func validateJSONStructure(data interface{}, maxDepth, maxKeys, currentDepth int) error {
	if currentDepth > maxDepth {
		return fmt.Errorf("JSON structure too deep (max depth: %d)", maxDepth)
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if len(v) > maxKeys {
			return fmt.Errorf("too many keys in JSON object (max: %d)", maxKeys)
		}
		for _, value := range v {
			if err := validateJSONStructure(value, maxDepth, maxKeys, currentDepth+1); err != nil {
				return err
			}
		}
	case []interface{}:
		if len(v) > maxKeys {
			return fmt.Errorf("array too large (max elements: %d)", maxKeys)
		}
		for _, value := range v {
			if err := validateJSONStructure(value, maxDepth, maxKeys, currentDepth+1); err != nil {
				return err
			}
		}
	}

	return nil
}
