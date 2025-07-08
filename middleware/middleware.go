package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
			"request_id", param.Keys["request_id"],
		)
		return ""
	})
}

// CORS handles Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

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

// ErrorHandler handles panics and errors
func ErrorHandler(logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"request_id", c.GetString("request_id"),
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

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
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
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", duration,
			"request_id", c.GetString("request_id"),
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
