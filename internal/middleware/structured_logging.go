package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// StructuredLoggingConfig holds configuration for structured logging
type StructuredLoggingConfig struct {
	Logger               *zap.Logger
	LogRequestBody       bool
	LogResponseBody      bool
	LogHeaders           bool
	MaxBodySize          int64
	SkipPaths            []string
	SensitivePaths       []string
	EnablePerformanceLog bool
	EnableSecurityLog    bool
}

// NewStructuredLoggingConfig creates a new structured logging configuration
func NewStructuredLoggingConfig(logger *zap.Logger) *StructuredLoggingConfig {
	return &StructuredLoggingConfig{
		Logger:               logger,
		LogRequestBody:       false, // Default to false for security
		LogResponseBody:      false, // Default to false for performance
		LogHeaders:           true,
		MaxBodySize:          1024, // 1KB max body logging
		EnablePerformanceLog: true,
		EnableSecurityLog:    true,
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
		SensitivePaths: []string{
			"/api/v2/auth/login",
			"/api/v2/auth/register",
			"/api/v2/auth/set-mpin",
			"/api/v2/auth/update-mpin",
			"/api/v2/auth/forgot-password",
			"/api/v2/auth/reset-password",
		},
	}
}

// StructuredLoggingMiddleware provides comprehensive structured logging
func StructuredLoggingMiddleware(config *StructuredLoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for certain paths
		if shouldSkipLogging(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		start := time.Now()
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = generateRequestID()
			c.Set("request_id", requestID)
		}

		// Capture request information
		requestInfo := &RequestInfo{
			RequestID:     requestID,
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			RawQuery:      c.Request.URL.RawQuery,
			ClientIP:      c.ClientIP(),
			UserAgent:     c.GetHeader("User-Agent"),
			Referer:       c.GetHeader("Referer"),
			ContentType:   c.GetHeader("Content-Type"),
			ContentLength: c.Request.ContentLength,
			StartTime:     start,
		}

		// Add user context if available
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(string); ok {
				requestInfo.UserID = uid
			}
		}

		// Capture request headers if enabled
		if config.LogHeaders {
			requestInfo.Headers = captureHeaders(c.Request.Header, []string{
				"Authorization", // Exclude sensitive headers
				"Cookie",
				"X-API-Key",
			})
		}

		// Capture request body if enabled and appropriate
		var requestBody []byte
		if config.LogRequestBody && shouldLogBody(c.Request.Method, c.Request.URL.Path, config.SensitivePaths) {
			if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength <= config.MaxBodySize {
				requestBody, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
				requestInfo.RequestBodySize = len(requestBody)
			}
		}

		// Create response writer wrapper to capture response
		responseWriter := &responseWriterWrapper{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Log request start
		logRequestStart(config.Logger, requestInfo)

		// Process request
		c.Next()

		// Calculate duration and capture response information
		duration := time.Since(start)
		responseInfo := &ResponseInfo{
			StatusCode:   c.Writer.Status(),
			ResponseSize: responseWriter.body.Len(),
			Duration:     duration,
			HasErrors:    len(c.Errors) > 0,
		}

		// Capture response body if enabled
		if config.LogResponseBody && responseWriter.body.Len() > 0 && responseWriter.body.Len() <= int(config.MaxBodySize) {
			responseInfo.ResponseBodySize = responseWriter.body.Len()
		}

		// Log request completion
		logRequestCompletion(config.Logger, requestInfo, responseInfo)

		// Log performance metrics if enabled
		if config.EnablePerformanceLog {
			logPerformanceMetrics(config.Logger, requestInfo, responseInfo)
		}

		// Log security events if enabled and applicable
		if config.EnableSecurityLog && isSensitivePath(c.Request.URL.Path, config.SensitivePaths) {
			logSecurityEvent(config.Logger, requestInfo, responseInfo)
		}

		// Log errors if any occurred
		if len(c.Errors) > 0 {
			logRequestErrors(config.Logger, requestInfo, c.Errors)
		}
	}
}

// RequestInfo holds information about the incoming request
type RequestInfo struct {
	RequestID       string
	Method          string
	Path            string
	RawQuery        string
	ClientIP        string
	UserAgent       string
	Referer         string
	ContentType     string
	ContentLength   int64
	RequestBodySize int
	Headers         map[string]string
	UserID          string
	StartTime       time.Time
}

// ResponseInfo holds information about the outgoing response
type ResponseInfo struct {
	StatusCode       int
	ResponseSize     int
	ResponseBodySize int
	Duration         time.Duration
	HasErrors        bool
}

// responseWriterWrapper wraps gin.ResponseWriter to capture response body
type responseWriterWrapper struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// logRequestStart logs the start of a request
func logRequestStart(logger *zap.Logger, req *RequestInfo) {
	fields := []zap.Field{
		zap.String("request_id", req.RequestID),
		zap.String("method", req.Method),
		zap.String("path", req.Path),
		zap.String("client_ip", req.ClientIP),
		zap.String("user_agent", req.UserAgent),
		zap.Time("start_time", req.StartTime),
	}

	if req.UserID != "" {
		fields = append(fields, zap.String("user_id", req.UserID))
	}

	if req.RawQuery != "" {
		fields = append(fields, zap.String("query", req.RawQuery))
	}

	if req.ContentLength > 0 {
		fields = append(fields, zap.Int64("content_length", req.ContentLength))
	}

	if req.ContentType != "" {
		fields = append(fields, zap.String("content_type", req.ContentType))
	}

	if req.Referer != "" {
		fields = append(fields, zap.String("referer", req.Referer))
	}

	if len(req.Headers) > 0 {
		fields = append(fields, zap.Any("headers", req.Headers))
	}

	logger.Info("Request started", fields...)
}

// logRequestCompletion logs the completion of a request
func logRequestCompletion(logger *zap.Logger, req *RequestInfo, resp *ResponseInfo) {
	fields := []zap.Field{
		zap.String("request_id", req.RequestID),
		zap.String("method", req.Method),
		zap.String("path", req.Path),
		zap.String("client_ip", req.ClientIP),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", resp.Duration),
		zap.Int("response_size", resp.ResponseSize),
		zap.Bool("has_errors", resp.HasErrors),
	}

	if req.UserID != "" {
		fields = append(fields, zap.String("user_id", req.UserID))
	}

	// Log with appropriate level based on status code
	switch {
	case resp.StatusCode >= 500:
		logger.Error("Request completed with server error", fields...)
	case resp.StatusCode >= 400:
		logger.Warn("Request completed with client error", fields...)
	case resp.StatusCode >= 300:
		logger.Info("Request completed with redirect", fields...)
	default:
		logger.Info("Request completed successfully", fields...)
	}
}

// logPerformanceMetrics logs performance-related metrics
func logPerformanceMetrics(logger *zap.Logger, req *RequestInfo, resp *ResponseInfo) {
	fields := []zap.Field{
		zap.String("request_id", req.RequestID),
		zap.String("method", req.Method),
		zap.String("path", req.Path),
		zap.Duration("duration", resp.Duration),
		zap.Int("response_size", resp.ResponseSize),
		zap.Float64("duration_ms", float64(resp.Duration.Nanoseconds())/1e6),
	}

	if req.UserID != "" {
		fields = append(fields, zap.String("user_id", req.UserID))
	}

	// Log performance warnings for slow requests
	if resp.Duration > 5*time.Second {
		logger.Warn("Slow request detected", fields...)
	} else if resp.Duration > 1*time.Second {
		logger.Info("Request performance metrics", fields...)
	} else {
		logger.Debug("Request performance metrics", fields...)
	}
}

// logSecurityEvent logs security-related events
func logSecurityEvent(logger *zap.Logger, req *RequestInfo, resp *ResponseInfo) {
	fields := []zap.Field{
		zap.String("request_id", req.RequestID),
		zap.String("method", req.Method),
		zap.String("path", req.Path),
		zap.String("client_ip", req.ClientIP),
		zap.String("user_agent", req.UserAgent),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", resp.Duration),
	}

	if req.UserID != "" {
		fields = append(fields, zap.String("user_id", req.UserID))
	}

	// Determine security event type
	eventType := "security_access"
	if resp.StatusCode == 401 {
		eventType = "authentication_failure"
	} else if resp.StatusCode == 403 {
		eventType = "authorization_failure"
	} else if resp.StatusCode >= 400 {
		eventType = "security_error"
	}

	fields = append(fields, zap.String("security_event_type", eventType))

	if resp.StatusCode >= 400 {
		logger.Warn("Security event logged", fields...)
	} else {
		logger.Info("Security access logged", fields...)
	}
}

// logRequestErrors logs any errors that occurred during request processing
func logRequestErrors(logger *zap.Logger, req *RequestInfo, ginErrors []*gin.Error) {
	for _, ginErr := range ginErrors {
		fields := []zap.Field{
			zap.String("request_id", req.RequestID),
			zap.String("method", req.Method),
			zap.String("path", req.Path),
			zap.String("client_ip", req.ClientIP),
			zap.Error(ginErr.Err),
			zap.Uint64("error_type", uint64(ginErr.Type)),
		}

		if req.UserID != "" {
			fields = append(fields, zap.String("user_id", req.UserID))
		}

		logger.Error("Request error occurred", fields...)
	}
}

// Helper functions

// shouldSkipLogging checks if logging should be skipped for a path
func shouldSkipLogging(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// shouldLogBody checks if request body should be logged
func shouldLogBody(method, path string, sensitivePaths []string) bool {
	// Only log body for POST, PUT, PATCH methods
	if method != "POST" && method != "PUT" && method != "PATCH" {
		return false
	}

	// Don't log body for sensitive paths
	for _, sensitivePath := range sensitivePaths {
		if strings.HasPrefix(path, sensitivePath) {
			return false
		}
	}

	return true
}

// isSensitivePath checks if a path is considered sensitive
func isSensitivePath(path string, sensitivePaths []string) bool {
	for _, sensitivePath := range sensitivePaths {
		if strings.HasPrefix(path, sensitivePath) {
			return true
		}
	}
	return false
}

// captureHeaders captures HTTP headers while excluding sensitive ones
func captureHeaders(headers http.Header, excludeHeaders []string) map[string]string {
	result := make(map[string]string)

	for name, values := range headers {
		// Skip sensitive headers
		skip := false
		for _, exclude := range excludeHeaders {
			if strings.EqualFold(name, exclude) {
				skip = true
				break
			}
		}

		if !skip && len(values) > 0 {
			result[name] = values[0] // Only capture first value
		}
	}

	return result
}

// RequestLoggingMiddleware provides basic request logging (simpler version)
func RequestLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetString("request_id")

		// Process request
		c.Next()

		// Log request completion
		duration := time.Since(start)

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("status_code", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.Int("response_size", c.Writer.Size()),
		}

		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(string); ok {
				fields = append(fields, zap.String("user_id", uid))
			}
		}

		logger.Info("Request processed", fields...)
	}
}
