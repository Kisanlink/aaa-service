package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/Kisanlink/aaa-service/services"
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
		"/api/v2/users",
		"/api/v2/roles",
		"/api/v2/permissions",
		"/api/v2/admin",
	}

	for _, operation := range sensitiveOperations {
		if len(path) >= len(operation) && path[:len(operation)] == operation {
			return true
		}
	}

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
