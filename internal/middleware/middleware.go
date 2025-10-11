package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logger logs request details
func Logger(logger interfaces.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Any("request_id", param.Keys["request_id"]),
		)
		return ""
	})
}

// CORS handles Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed origins from environment variable
		allowedOrigins := os.Getenv("AAA_CORS_ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "*" // Default to allow all origins
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", allowedOrigins)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID, Accept, Cache-Control, X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID, X-Total-Count")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Handle actual requests
		c.Header("Access-Control-Allow-Origin", allowedOrigins)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID, Accept, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID, X-Total-Count")
		c.Header("Access-Control-Allow-Credentials", "true")

		c.Next()
	}
}

// RateLimit implements rate limiting using token bucket algorithm
func RateLimit() gin.HandlerFunc {
	// Create a rate limiter: 100 requests per minute per IP
	limiters := make(map[string]*rate.Limiter)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		limiter, exists := limiters[clientIP]
		if !exists {
			limiter = rate.NewLimiter(rate.Every(time.Minute/100), 100)
			limiters[clientIP] = limiter
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests from this IP",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SensitiveOperationRateLimit implements stricter rate limiting for sensitive operations
func SensitiveOperationRateLimit() gin.HandlerFunc {
	// Create a rate limiter: 10 requests per minute per IP for sensitive operations
	limiters := make(map[string]*rate.Limiter)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		limiter, exists := limiters[clientIP]
		if !exists {
			// 10 requests per minute with burst of 5
			limiter = rate.NewLimiter(rate.Every(time.Minute/10), 5)
			limiters[clientIP] = limiter
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded for sensitive operation",
				"message":     "Too many sensitive requests from this IP. Please try again later.",
				"retry_after": "60", // seconds
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthenticationRateLimit implements sophisticated rate limiting for authentication attempts
func AuthenticationRateLimit() gin.HandlerFunc {
	// Create rate limiters for different time windows
	ipLimiters := make(map[string]*rate.Limiter)
	failedAttempts := make(map[string]int)
	lastAttempt := make(map[string]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Clean up old entries (older than 1 hour)
		if len(lastAttempt) > 1000 { // Prevent memory leak
			for ip, lastTime := range lastAttempt {
				if now.Sub(lastTime) > time.Hour {
					delete(ipLimiters, ip)
					delete(failedAttempts, ip)
					delete(lastAttempt, ip)
				}
			}
		}

		// Check if IP is temporarily blocked due to too many failed attempts
		if attempts, exists := failedAttempts[clientIP]; exists {
			if attempts >= 10 { // Block after 10 failed attempts
				if lastTime, exists := lastAttempt[clientIP]; exists {
					// Block for exponential backoff: 2^(attempts-10) minutes, max 60 minutes
					blockDuration := time.Duration(1<<uint(min(attempts-10, 6))) * time.Minute
					if now.Sub(lastTime) < blockDuration {
						c.JSON(http.StatusTooManyRequests, gin.H{
							"error":       "IP temporarily blocked",
							"message":     "Too many failed authentication attempts. Please try again later.",
							"retry_after": fmt.Sprintf("%.0f", blockDuration.Seconds()),
						})
						c.Abort()
						return
					} else {
						// Reset after block period
						delete(failedAttempts, clientIP)
					}
				}
			}
		}

		// Apply rate limiting
		limiter, exists := ipLimiters[clientIP]
		if !exists {
			// 5 requests per minute with burst of 3
			limiter = rate.NewLimiter(rate.Every(time.Minute/5), 3)
			ipLimiters[clientIP] = limiter
		}

		if !limiter.Allow() {
			// Increment failed attempts counter
			failedAttempts[clientIP]++
			lastAttempt[clientIP] = now

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Authentication rate limit exceeded",
				"message":     "Too many authentication attempts. Please try again later.",
				"retry_after": "60",
			})
			c.Abort()
			return
		}

		// Store the limiter and timestamp for potential failure tracking
		c.Set("rate_limiter_ip", clientIP)
		c.Set("rate_limiter_timestamp", now)

		c.Next()

		// Track failed authentication attempts
		if c.Writer.Status() == http.StatusUnauthorized {
			failedAttempts[clientIP]++
			lastAttempt[clientIP] = now
		} else if c.Writer.Status() == http.StatusOK {
			// Reset failed attempts on successful authentication
			delete(failedAttempts, clientIP)
		}
	}
}

// MPinRateLimit implements specific rate limiting for MPIN operations
func MPinRateLimit() gin.HandlerFunc {
	// More restrictive rate limiting for MPIN operations
	limiters := make(map[string]*rate.Limiter)
	failedAttempts := make(map[string]int)
	lastAttempt := make(map[string]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Clean up old entries
		if len(lastAttempt) > 500 {
			for ip, lastTime := range lastAttempt {
				if now.Sub(lastTime) > time.Hour {
					delete(limiters, ip)
					delete(failedAttempts, ip)
					delete(lastAttempt, ip)
				}
			}
		}

		// Check for MPIN-specific blocking (more restrictive)
		if attempts, exists := failedAttempts[clientIP]; exists {
			if attempts >= 5 { // Block after 5 failed MPIN attempts
				if lastTime, exists := lastAttempt[clientIP]; exists {
					// Block for 15 minutes after 5 failed attempts
					blockDuration := 15 * time.Minute
					if now.Sub(lastTime) < blockDuration {
						c.JSON(http.StatusTooManyRequests, gin.H{
							"error":       "MPIN operations temporarily blocked",
							"message":     "Too many failed MPIN attempts. Please try again later.",
							"retry_after": fmt.Sprintf("%.0f", blockDuration.Seconds()),
						})
						c.Abort()
						return
					} else {
						delete(failedAttempts, clientIP)
					}
				}
			}
		}

		// Apply MPIN-specific rate limiting: 3 requests per minute
		limiter, exists := limiters[clientIP]
		if !exists {
			limiter = rate.NewLimiter(rate.Every(time.Minute/3), 2)
			limiters[clientIP] = limiter
		}

		if !limiter.Allow() {
			failedAttempts[clientIP]++
			lastAttempt[clientIP] = now

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "MPIN rate limit exceeded",
				"message":     "Too many MPIN operations. Please try again later.",
				"retry_after": "60",
			})
			c.Abort()
			return
		}

		c.Set("mpin_rate_limiter_ip", clientIP)
		c.Set("mpin_rate_limiter_timestamp", now)

		c.Next()

		// Track failed MPIN attempts
		if c.Writer.Status() == http.StatusUnauthorized || c.Writer.Status() == http.StatusBadRequest {
			failedAttempts[clientIP]++
			lastAttempt[clientIP] = now
		} else if c.Writer.Status() == http.StatusOK {
			delete(failedAttempts, clientIP)
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Timeout adds a timeout to requests
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan bool, 1)
		go func() {
			c.Next()
			done <- true
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "Request timeout",
				"message": fmt.Sprintf("Request timed out after %v", timeout),
			})
			c.Abort()
			return
		}
	}
}

// Auth handles authentication middleware
func Auth(authService interfaces.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Validate token and extract user info
		// This is a placeholder - implement actual token validation
		userID := "placeholder_user_id"

		// Set user info in context
		c.Set("user_id", userID)
		c.Next()
	}
}

// ValidateUserID validates user ID in URL parameters
func ValidateUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User ID is required",
			})
			c.Abort()
			return
		}

		// Add validation logic here if needed
		c.Next()
	}
}

// ValidateRoleID validates role ID in URL parameters
func ValidateRoleID() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.Param("roleId")
		if roleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Role ID is required",
			})
			c.Abort()
			return
		}

		// Add validation logic here if needed
		c.Next()
	}
}

// Pagination adds pagination parameters to context
func Pagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 10 // default limit
		offset := 0 // default offset

		if limitStr := c.Query("limit"); limitStr != "" {
			if parsed, err := parsePositiveInt(limitStr); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}

		if offsetStr := c.Query("offset"); offsetStr != "" {
			if parsed, err := parsePositiveInt(offsetStr); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		c.Set("limit", limit)
		c.Set("offset", offset)
		c.Next()
	}
}

// Search adds search parameters to context
func Search() gin.HandlerFunc {
	return func(c *gin.Context) {
		keyword := c.Query("q")
		if keyword != "" {
			c.Set("search_keyword", keyword)
		}
		c.Next()
	}
}

// PanicRecoveryHandler handles panics and errors
func PanicRecoveryHandler(logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("request_id", c.GetString("request_id")),
				)

				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"message": "An unexpected error occurred",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// SecurityHeaders adds comprehensive security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Force HTTPS (only in production)
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://unpkg.com; " +
			"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://unpkg.com; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' https://cdn.jsdelivr.net https://unpkg.com; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Header("Content-Security-Policy", csp)

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy (formerly Feature Policy)
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		// Remove server information
		c.Header("Server", "")

		// Prevent caching of sensitive responses
		if isSensitiveEndpoint(c.Request.URL.Path) {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

// isSensitiveEndpoint checks if an endpoint handles sensitive data
func isSensitiveEndpoint(path string) bool {
	sensitiveEndpoints := []string{
		"/api/v2/auth/login",
		"/api/v2/auth/register",
		"/api/v2/auth/refresh",
		"/api/v2/auth/set-mpin",
		"/api/v2/auth/update-mpin",
		"/api/v2/users",
		"/api/v2/roles",
		"/api/v2/admin",
	}

	for _, endpoint := range sensitiveEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}

	return false
}

// RequestSizeLimit limits the size of request bodies
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// ResponseCompression adds gzip compression to responses
func ResponseCompression() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is a placeholder - implement actual compression
		c.Next()
	}
}

// Metrics adds metrics collection
func Metrics(logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		logger.Info("Request metrics",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("request_id", c.GetString("request_id")),
		)
	}
}

// Helper function to parse positive integers
func parsePositiveInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0, err
	}
	if result < 0 {
		return 0, fmt.Errorf("negative number")
	}
	return result, nil
}
