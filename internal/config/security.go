package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// Rate limiting configuration
	RateLimit               RateLimitConfig
	AuthenticationRateLimit AuthRateLimitConfig
	MPinRateLimit           MPinRateLimitConfig

	// Input validation configuration
	InputValidation InputValidationConfig

	// Security headers configuration
	SecurityHeaders SecurityHeadersConfig

	// CORS configuration
	CORS CORSConfig

	// Audit configuration
	Audit AuditConfig

	// Encryption configuration
	Encryption EncryptionConfig

	// Cookie configuration
	Cookie CookieConfig
}

// RateLimitConfig holds general rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	CleanupInterval   time.Duration
}

// AuthRateLimitConfig holds authentication-specific rate limiting configuration
type AuthRateLimitConfig struct {
	RequestsPerMinute  int
	BurstSize          int
	MaxFailedAttempts  int
	BlockDuration      time.Duration
	ExponentialBackoff bool
}

// MPinRateLimitConfig holds MPIN-specific rate limiting configuration
type MPinRateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	MaxFailedAttempts int
	BlockDuration     time.Duration
}

// InputValidationConfig holds input validation configuration
type InputValidationConfig struct {
	MaxRequestSize      int64
	MaxJSONDepth        int
	MaxJSONKeys         int
	EnableSanitization  bool
	AllowedContentTypes []string
}

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	EnableHSTS               bool
	HSTSMaxAge               int
	HSTSIncludeSubdomains    bool
	HSTSPreload              bool
	ContentSecurityPolicy    string
	ReferrerPolicy           string
	PermissionsPolicy        string
	EnableXSSProtection      bool
	EnableContentTypeNoSniff bool
	FrameOptions             string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// AuditConfig holds audit logging configuration
type AuditConfig struct {
	EnableSecurityAudit   bool
	EnableDataAccessAudit bool
	LogSuspiciousActivity bool
	RetentionDays         int
	SensitiveOperations   []string
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	PasswordHashCost       int
	MPinHashCost           int
	TokenSigningMethod     string
	TokenExpiration        time.Duration
	RefreshTokenExpiration time.Duration
}

// CookieConfig holds cookie configuration for cross-subdomain sharing
type CookieConfig struct {
	// Domain sets the cookie domain for cross-subdomain sharing
	// e.g., ".kisanlink.in" allows cookies to be shared across
	// aaa.kisanlink.in, farmers.kisanlink.in, etc.
	// Empty string means cookies are only valid for the exact host
	Domain string
}

// LoadSecurityConfig loads security configuration from environment variables
func LoadSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RateLimit: RateLimitConfig{
			RequestsPerMinute: getEnvInt("AAA_RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
			BurstSize:         getEnvInt("AAA_RATE_LIMIT_BURST_SIZE", 100),
			CleanupInterval:   getEnvDuration("AAA_RATE_LIMIT_CLEANUP_INTERVAL", time.Hour),
		},
		AuthenticationRateLimit: AuthRateLimitConfig{
			RequestsPerMinute:  getEnvInt("AAA_AUTH_RATE_LIMIT_REQUESTS_PER_MINUTE", 5),
			BurstSize:          getEnvInt("AAA_AUTH_RATE_LIMIT_BURST_SIZE", 3),
			MaxFailedAttempts:  getEnvInt("AAA_AUTH_MAX_FAILED_ATTEMPTS", 10),
			BlockDuration:      getEnvDuration("AAA_AUTH_BLOCK_DURATION", 15*time.Minute),
			ExponentialBackoff: getEnvBool("AAA_AUTH_EXPONENTIAL_BACKOFF", true),
		},
		MPinRateLimit: MPinRateLimitConfig{
			RequestsPerMinute: getEnvInt("AAA_MPIN_RATE_LIMIT_REQUESTS_PER_MINUTE", 3),
			BurstSize:         getEnvInt("AAA_MPIN_RATE_LIMIT_BURST_SIZE", 2),
			MaxFailedAttempts: getEnvInt("AAA_MPIN_MAX_FAILED_ATTEMPTS", 5),
			BlockDuration:     getEnvDuration("AAA_MPIN_BLOCK_DURATION", 15*time.Minute),
		},
		InputValidation: InputValidationConfig{
			MaxRequestSize:     getEnvInt64("AAA_MAX_REQUEST_SIZE", 10*1024*1024), // 10MB
			MaxJSONDepth:       getEnvInt("AAA_MAX_JSON_DEPTH", 5),
			MaxJSONKeys:        getEnvInt("AAA_MAX_JSON_KEYS", 20),
			EnableSanitization: getEnvBool("AAA_ENABLE_INPUT_SANITIZATION", true),
			AllowedContentTypes: getEnvStringSlice("AAA_ALLOWED_CONTENT_TYPES", []string{
				"application/json",
				"application/x-www-form-urlencoded",
				"multipart/form-data",
			}),
		},
		SecurityHeaders: SecurityHeadersConfig{
			EnableHSTS:               getEnvBool("AAA_ENABLE_HSTS", true),
			HSTSMaxAge:               getEnvInt("AAA_HSTS_MAX_AGE", 31536000), // 1 year
			HSTSIncludeSubdomains:    getEnvBool("AAA_HSTS_INCLUDE_SUBDOMAINS", true),
			HSTSPreload:              getEnvBool("AAA_HSTS_PRELOAD", false),
			ContentSecurityPolicy:    getEnvString("AAA_CSP", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://unpkg.com; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://unpkg.com; img-src 'self' data: https:; font-src 'self' https://cdn.jsdelivr.net https://unpkg.com; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"),
			ReferrerPolicy:           getEnvString("AAA_REFERRER_POLICY", "strict-origin-when-cross-origin"),
			PermissionsPolicy:        getEnvString("AAA_PERMISSIONS_POLICY", "camera=(), microphone=(), geolocation=()"),
			EnableXSSProtection:      getEnvBool("AAA_ENABLE_XSS_PROTECTION", true),
			EnableContentTypeNoSniff: getEnvBool("AAA_ENABLE_CONTENT_TYPE_NOSNIFF", true),
			FrameOptions:             getEnvString("AAA_FRAME_OPTIONS", "DENY"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvStringSlice("AAA_CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods: getEnvStringSlice("AAA_CORS_ALLOWED_METHODS", []string{
				"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
			}),
			AllowedHeaders: getEnvStringSlice("AAA_CORS_ALLOWED_HEADERS", []string{
				"Origin", "Content-Type", "Content-Length", "Accept-Encoding",
				"X-CSRF-Token", "Authorization", "X-Request-ID", "Accept",
				"Cache-Control", "X-Requested-With",
			}),
			ExposedHeaders: getEnvStringSlice("AAA_CORS_EXPOSED_HEADERS", []string{
				"Content-Length", "X-Request-ID", "X-Total-Count",
			}),
			AllowCredentials: getEnvBool("AAA_CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvInt("AAA_CORS_MAX_AGE", 86400), // 24 hours
		},
		Audit: AuditConfig{
			EnableSecurityAudit:   getEnvBool("AAA_ENABLE_SECURITY_AUDIT", true),
			EnableDataAccessAudit: getEnvBool("AAA_ENABLE_DATA_ACCESS_AUDIT", true),
			LogSuspiciousActivity: getEnvBool("AAA_LOG_SUSPICIOUS_ACTIVITY", true),
			RetentionDays:         getEnvInt("AAA_AUDIT_RETENTION_DAYS", 90),
			SensitiveOperations: getEnvStringSlice("AAA_SENSITIVE_OPERATIONS", []string{
				"/api/v1/auth/login",
				"/api/v1/auth/register",
				"/api/v1/auth/set-mpin",
				"/api/v1/auth/update-mpin",
				"/api/v1/users",
				"/api/v1/roles",
				"/api/v1/admin",
			}),
		},
		Encryption: EncryptionConfig{
			PasswordHashCost:       getEnvInt("AAA_PASSWORD_HASH_COST", 12),
			MPinHashCost:           getEnvInt("AAA_MPIN_HASH_COST", 12),
			TokenSigningMethod:     getEnvString("AAA_TOKEN_SIGNING_METHOD", "HS256"),
			TokenExpiration:        getEnvDuration("AAA_TOKEN_EXPIRATION", 15*time.Minute),
			RefreshTokenExpiration: getEnvDuration("AAA_REFRESH_TOKEN_EXPIRATION", 7*24*time.Hour),
		},
		Cookie: CookieConfig{
			// Domain for cross-subdomain sharing (e.g., ".kisanlink.in")
			// Empty string means cookies are only valid for the exact host
			Domain: getEnvString("AAA_COOKIE_DOMAIN", ""),
		},
	}
}

// Helper functions to get environment variables with defaults

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// Validate validates the security configuration
func (sc *SecurityConfig) Validate() error {
	// Add validation logic here if needed
	return nil
}

// IsProductionMode checks if the application is running in production mode
func IsProductionMode() bool {
	return getEnvString("GIN_MODE", "debug") == "release" ||
		getEnvString("AAA_ENVIRONMENT", "development") == "production"
}

// GetSecurityLevel returns the security level based on environment
func GetSecurityLevel() string {
	return getEnvString("AAA_SECURITY_LEVEL", "standard")
}
