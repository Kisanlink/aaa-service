package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditMiddleware provides HTTP request auditing middleware
type AuditMiddleware struct {
	auditService *services.AuditService
	logger       *zap.Logger
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(auditService *services.AuditService, logger *zap.Logger) *AuditMiddleware {
	return &AuditMiddleware{
		auditService: auditService,
		logger:       logger,
	}
}

// HTTPAuditMiddleware logs HTTP requests for audit purposes
func (m *AuditMiddleware) HTTPAuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit for health checks and docs
		if m.isSkippedEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()

		// Get user ID from context (set by auth middleware)
		userID := "anonymous"
		if uid, exists := c.Get("user_id"); exists {
			if userIDStr, ok := uid.(string); ok {
				userID = userIDStr
			}
		}

		// Capture request body if needed (for sensitive operations)
		var requestBody []byte
		if m.shouldCaptureRequestBody(c.Request.Method, c.Request.URL.Path) {
			if c.Request.Body != nil {
				requestBody, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}
		}

		// Create a custom response writer to capture response
		responseWriter := &auditResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Prepare audit details
		details := map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"query":       c.Request.URL.RawQuery,
			"ip_address":  c.ClientIP(),
			"user_agent":  c.GetHeader("User-Agent"),
			"duration_ms": duration.Milliseconds(),
			"status_code": c.Writer.Status(),
		}

		// Add request body for sensitive operations (without passwords)
		if len(requestBody) > 0 && len(requestBody) < 1024 { // Limit size
			details["request_size"] = len(requestBody)
			// Note: Be careful not to log sensitive data like passwords
		}

		// Add response size
		if responseWriter.body.Len() > 0 {
			details["response_size"] = responseWriter.body.Len()
		}

		// Determine success
		statusCode := c.Writer.Status()
		success := statusCode >= 200 && statusCode < 400

		// Avoid double logging for unauthenticated login attempts
		if statusCode == http.StatusUnauthorized && len(c.Request.URL.Path) >= len("/api/v2/auth/login") && c.Request.URL.Path[:len("/api/v2/auth/login")] == "/api/v2/auth/login" {
			if !success {
				err := fmt.Errorf("HTTP %d", statusCode)
				m.auditService.LogUserActionWithError(c.Request.Context(), userID, "http_request", "api", c.Request.URL.Path, err, details)
			} else {
				m.auditService.LogUserAction(c.Request.Context(), userID, "http_request", "api", c.Request.URL.Path, details)
			}
			return
		}

		// Log to audit service
		if success {
			m.auditService.LogUserAction(c.Request.Context(), userID, "http_request", "api", c.Request.URL.Path, details)
		} else {
			err := fmt.Errorf("HTTP %d", statusCode)
			m.auditService.LogUserActionWithError(c.Request.Context(), userID, "http_request", "api", c.Request.URL.Path, err, details)
		}

		// Log detailed audit for sensitive operations
		if m.isSensitiveOperation(c.Request.Method, c.Request.URL.Path) {
			m.auditService.LogSecurityEvent(c.Request.Context(), userID, "sensitive_operation", c.Request.URL.Path, success, details)
		}
	}
}

// DataAccessAuditMiddleware logs data access for CRUD operations
func (m *AuditMiddleware) DataAccessAuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only audit data modification operations
		if !m.isDataModificationOperation(c.Request.Method) {
			c.Next()
			return
		}

		userID := "anonymous"
		if uid, exists := c.Get("user_id"); exists {
			if userIDStr, ok := uid.(string); ok {
				userID = userIDStr
			}
		}

		// Extract resource information from URL
		resource, resourceID := m.extractResourceInfo(c.Request.URL.Path)

		// Process request
		c.Next()

		// Determine action based on HTTP method
		action := m.mapHTTPMethodToAction(c.Request.Method)

		// Prepare details
		details := map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"ip_address":  c.ClientIP(),
			"status_code": c.Writer.Status(),
		}

		// Log data access
		statusCode := c.Writer.Status()
		success := statusCode >= 200 && statusCode < 400

		if success {
			m.auditService.LogDataAccess(c.Request.Context(), userID, action, resource, resourceID, nil, details)
		} else {
			err := fmt.Errorf("HTTP %d", statusCode)
			m.auditService.LogUserActionWithError(c.Request.Context(), userID, action, resource, resourceID, err, details)
		}
	}
}

// auditResponseWriter wraps gin.ResponseWriter to capture response body
type auditResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *auditResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// isSkippedEndpoint checks if an endpoint should be skipped for audit
func (m *AuditMiddleware) isSkippedEndpoint(path string) bool {
	skippedEndpoints := []string{
		"/health",
		"/docs",
		"/favicon.ico",
		"/metrics",
	}

	for _, endpoint := range skippedEndpoints {
		if path == endpoint || (len(path) > len(endpoint) && path[:len(endpoint)] == endpoint) {
			return true
		}
	}

	return false
}

// shouldCaptureRequestBody determines if request body should be captured
func (m *AuditMiddleware) shouldCaptureRequestBody(method, path string) bool {
	// Capture body for sensitive operations
	sensitiveEndpoints := []string{
		"/api/v2/auth/login",
		"/api/v2/auth/register",
		"/api/v2/users",
		"/api/v2/roles",
		"/api/v2/permissions",
	}

	if method == "POST" || method == "PUT" || method == "PATCH" {
		for _, endpoint := range sensitiveEndpoints {
			if len(path) >= len(endpoint) && path[:len(endpoint)] == endpoint {
				return true
			}
		}
	}

	return false
}

// isSensitiveOperation checks if an operation is sensitive and requires detailed auditing
func (m *AuditMiddleware) isSensitiveOperation(method, path string) bool {
	sensitiveOperations := []string{
		"/api/v2/auth/login",
		"/api/v2/auth/register",
		"/api/v2/auth/set-mpin",
		"/api/v2/auth/update-mpin",
		"/api/v2/auth/forgot-password",
		"/api/v2/auth/reset-password",
		"/api/v2/users",
		"/api/v2/roles",
		"/api/v2/permissions",
		"/api/v2/admin",
	}

	for _, operation := range sensitiveOperations {
		if strings.HasPrefix(path, operation) {
			return true
		}
	}

	return false
}

// SecurityAuditMiddleware provides enhanced security auditing
func (m *AuditMiddleware) SecurityAuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit for non-sensitive operations
		if !m.isSensitiveOperation(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()
		userID := "anonymous"
		if uid, exists := c.Get("user_id"); exists {
			if userIDStr, ok := uid.(string); ok {
				userID = userIDStr
			}
		}

		// Capture security-relevant request details
		securityDetails := map[string]interface{}{
			"method":          c.Request.Method,
			"path":            c.Request.URL.Path,
			"ip_address":      c.ClientIP(),
			"user_agent":      c.GetHeader("User-Agent"),
			"x_forwarded_for": c.GetHeader("X-Forwarded-For"),
			"x_real_ip":       c.GetHeader("X-Real-IP"),
			"referer":         c.GetHeader("Referer"),
			"content_type":    c.GetHeader("Content-Type"),
			"content_length":  c.Request.ContentLength,
			"timestamp":       start.UTC().Format(time.RFC3339),
			"request_id":      c.GetString("request_id"),
		}

		// Add rate limiting information if available
		if rateLimiterIP, exists := c.Get("rate_limiter_ip"); exists {
			securityDetails["rate_limiter_ip"] = rateLimiterIP
		}
		if mpinRateLimiterIP, exists := c.Get("mpin_rate_limiter_ip"); exists {
			securityDetails["mpin_rate_limiter_ip"] = mpinRateLimiterIP
		}

		// Process request
		c.Next()

		// Calculate duration and get final status
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		success := statusCode >= 200 && statusCode < 400

		// Update security details with response information
		securityDetails["duration_ms"] = duration.Milliseconds()
		securityDetails["status_code"] = statusCode
		securityDetails["response_size"] = c.Writer.Size()

		// Determine security event type
		eventType := m.getSecurityEventType(c.Request.Method, c.Request.URL.Path, statusCode)

		// Log security event with appropriate level
		if success {
			m.auditService.LogSecurityEvent(c.Request.Context(), userID, eventType, c.Request.URL.Path, true, securityDetails)
		} else {
			// Log failed security events with higher priority
			m.auditService.LogSecurityEvent(c.Request.Context(), userID, eventType+"_failed", c.Request.URL.Path, false, securityDetails)

			// Additional logging for authentication failures
			if statusCode == http.StatusUnauthorized && strings.Contains(c.Request.URL.Path, "/auth/") {
				m.auditService.LogSecurityEvent(c.Request.Context(), userID, "authentication_failure", c.Request.URL.Path, false, securityDetails)
			}

			// Additional logging for authorization failures
			if statusCode == http.StatusForbidden {
				m.auditService.LogSecurityEvent(c.Request.Context(), userID, "authorization_failure", c.Request.URL.Path, false, securityDetails)
			}
		}

		// Log suspicious patterns
		m.detectAndLogSuspiciousActivity(c, userID, securityDetails)
	}
}

// getSecurityEventType determines the type of security event based on the request
func (m *AuditMiddleware) getSecurityEventType(method, path string, statusCode int) string {
	switch {
	case strings.Contains(path, "/auth/login"):
		return "authentication_attempt"
	case strings.Contains(path, "/auth/register"):
		return "user_registration"
	case strings.Contains(path, "/auth/set-mpin"):
		return "mpin_setup"
	case strings.Contains(path, "/auth/update-mpin"):
		return "mpin_update"
	case strings.Contains(path, "/auth/forgot-password"):
		return "password_reset_request"
	case strings.Contains(path, "/auth/reset-password"):
		return "password_reset"
	case strings.Contains(path, "/users") && method == "POST":
		return "user_creation"
	case strings.Contains(path, "/users") && method == "DELETE":
		return "user_deletion"
	case strings.Contains(path, "/roles") && method == "POST":
		return "role_assignment"
	case strings.Contains(path, "/roles") && method == "DELETE":
		return "role_removal"
	case strings.Contains(path, "/admin"):
		return "admin_operation"
	default:
		return "sensitive_operation"
	}
}

// detectAndLogSuspiciousActivity detects and logs potentially suspicious activity patterns
func (m *AuditMiddleware) detectAndLogSuspiciousActivity(c *gin.Context, userID string, details map[string]interface{}) {
	suspiciousPatterns := []string{}

	// Check for suspicious user agents
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		suspiciousPatterns = append(suspiciousPatterns, "missing_user_agent")
	} else if m.isSuspiciousUserAgent(userAgent) {
		suspiciousPatterns = append(suspiciousPatterns, "suspicious_user_agent")
	}

	// Check for suspicious IP patterns
	clientIP := c.ClientIP()
	if m.isSuspiciousIP(clientIP) {
		suspiciousPatterns = append(suspiciousPatterns, "suspicious_ip")
	}

	// Check for rapid requests (if rate limiter was triggered)
	if _, exists := c.Get("rate_limiter_ip"); exists {
		suspiciousPatterns = append(suspiciousPatterns, "rate_limit_triggered")
	}

	// Check for unusual request patterns
	if m.hasUnusualRequestPattern(c) {
		suspiciousPatterns = append(suspiciousPatterns, "unusual_request_pattern")
	}

	// Log suspicious activity if any patterns detected
	if len(suspiciousPatterns) > 0 {
		suspiciousDetails := make(map[string]interface{})
		for k, v := range details {
			suspiciousDetails[k] = v
		}
		suspiciousDetails["suspicious_patterns"] = suspiciousPatterns

		m.auditService.LogSecurityEvent(c.Request.Context(), userID, "suspicious_activity", c.Request.URL.Path, false, suspiciousDetails)
	}
}

// isSuspiciousUserAgent checks if a user agent string is suspicious
func (m *AuditMiddleware) isSuspiciousUserAgent(userAgent string) bool {
	suspiciousAgents := []string{
		"curl", "wget", "python", "bot", "crawler", "scanner",
		"sqlmap", "nikto", "nmap", "masscan", "zap",
	}

	lowerAgent := strings.ToLower(userAgent)
	for _, suspicious := range suspiciousAgents {
		if strings.Contains(lowerAgent, suspicious) {
			return true
		}
	}

	return false
}

// isSuspiciousIP checks if an IP address is suspicious (placeholder implementation)
func (m *AuditMiddleware) isSuspiciousIP(ip string) bool {
	// This is a placeholder - in production, you would check against:
	// - Known malicious IP lists
	// - Tor exit nodes
	// - VPN/proxy services (if not allowed)
	// - Geographic restrictions

	// For now, just check for localhost/private IPs in production
	if gin.Mode() == gin.ReleaseMode {
		if strings.HasPrefix(ip, "127.") || strings.HasPrefix(ip, "10.") ||
			strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "172.") {
			return false // Allow private IPs in production for now
		}
	}

	return false
}

// hasUnusualRequestPattern checks for unusual request patterns
func (m *AuditMiddleware) hasUnusualRequestPattern(c *gin.Context) bool {
	// Check for unusual header combinations
	headers := c.Request.Header

	// Missing common headers that legitimate clients usually send
	if headers.Get("Accept") == "" && headers.Get("Accept-Language") == "" {
		return true
	}

	// Unusual content length for the endpoint
	if c.Request.ContentLength > 1024*1024 { // 1MB
		return true
	}

	// Multiple authentication attempts from same IP (would need session tracking)
	// This is a placeholder - implement based on your session tracking needs

	return false
}

// isDataModificationOperation checks if the HTTP method modifies data
func (m *AuditMiddleware) isDataModificationOperation(method string) bool {
	modificationMethods := []string{"POST", "PUT", "PATCH", "DELETE"}
	for _, modMethod := range modificationMethods {
		if method == modMethod {
			return true
		}
	}
	return false
}

// extractResourceInfo extracts resource type and ID from URL path
func (m *AuditMiddleware) extractResourceInfo(path string) (string, string) {
	// Simple extraction logic - can be enhanced based on your URL structure
	// Example: /api/v2/users/123 -> resource: "users", resourceID: "123"

	parts := make([]string, 0)
	for _, part := range []string{"api", "v2", "v1"} {
		if len(path) > len(part)+1 && path[:len(part)+1] == "/"+part {
			path = path[len(part)+1:]
			break
		}
	}

	// Remove leading slash
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// Split path into parts
	for _, part := range []string{} {
		if part != "" {
			parts = append(parts, part)
		}
	}

	if len(parts) == 0 {
		return "unknown", "unknown"
	}

	resource := parts[0]
	resourceID := resource // Default to resource type

	if len(parts) > 1 {
		resourceID = parts[1]
	}

	return resource, resourceID
}

// mapHTTPMethodToAction maps HTTP methods to audit actions
func (m *AuditMiddleware) mapHTTPMethodToAction(method string) string {
	actionMap := map[string]string{
		"GET":    "read",
		"POST":   "create",
		"PUT":    "update",
		"PATCH":  "update",
		"DELETE": "delete",
	}

	if action, exists := actionMap[method]; exists {
		return action
	}

	return "access"
}
