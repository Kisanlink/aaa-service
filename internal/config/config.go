package config

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

var (
	appConfig     *AppConfig
	appConfigOnce sync.Once
)

// AppConfig holds application-wide configuration
type AppConfig struct {
	Environment string // development, staging, production
}

// GetAppConfig returns the singleton AppConfig instance
func GetAppConfig() *AppConfig {
	appConfigOnce.Do(func() {
		// Load .env file if it exists (ignored in production with real env vars)
		_ = godotenv.Load()
		appConfig = &AppConfig{
			Environment: strings.ToLower(getenv("APP_ENV", "")),
		}
	})
	return appConfig
}

// IsProduction returns true if running in production or staging
func (c *AppConfig) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod" || c.Environment == "staging"
}

// IsDevelopment returns true if running in development environment
func (c *AppConfig) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev" || c.Environment == "local" || c.Environment == ""
}

// JWTConfig holds JWT signing and verification configuration
type JWTConfig struct {
	Secret   string        `mapstructure:"secret"`
	Issuer   string        `mapstructure:"issuer"`
	Audience string        `mapstructure:"audience"`
	TTL      time.Duration `mapstructure:"ttl"`
	Leeway   time.Duration `mapstructure:"leeway"`
}

// LoadJWTConfigFromEnv loads JWT configuration from environment variables.
// Variables:
//
//	AAA_JWT_SECRET (required)
//	AAA_JWT_ISSUER
//	AAA_JWT_AUDIENCE
//	AAA_JWT_TTL (e.g., "24h")
//	AAA_JWT_LEEWAY (optional; default 2m)
func LoadJWTConfigFromEnv() *JWTConfig {
	ttl := parseDurationWithDefault(getenv("AAA_JWT_TTL", "24h"), 24*time.Hour)
	leeway := parseDurationWithDefault(getenv("AAA_JWT_LEEWAY", ""), 0)
	cfg := &JWTConfig{
		Secret:   getenv("AAA_JWT_SECRET", ""),
		Issuer:   getenv("AAA_JWT_ISSUER", "aaa-service"),
		Audience: getenv("AAA_JWT_AUDIENCE", ""),
		TTL:      ttl,
		Leeway:   leeway,
	}
	if cfg.Leeway == 0 {
		cfg.Leeway = 2 * time.Minute
	}
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseDurationWithDefault(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return def
}
