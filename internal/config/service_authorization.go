package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// ServiceAuthorizationConfig represents the service authorization configuration
type ServiceAuthorizationConfig struct {
	ServiceAuthorization ServiceAuthSection `yaml:"service_authorization"`
	DefaultBehavior      DefaultBehavior    `yaml:"default_behavior"`
}

// ServiceAuthSection contains the main authorization settings
type ServiceAuthSection struct {
	Enabled  bool                         `yaml:"enabled"`
	Services map[string]ServicePermission `yaml:"services,omitempty"`
}

// ServicePermission defines permissions for a specific service
type ServicePermission struct {
	ServiceID      string   `yaml:"service_id"`
	DisplayName    string   `yaml:"display_name"`
	Description    string   `yaml:"description"`
	APIKeyRequired bool     `yaml:"api_key_required"`
	APIKey         string   `yaml:"api_key,omitempty"`
	Permissions    []string `yaml:"permissions"`
}

// DefaultBehavior defines fallback behavior when authorization is disabled
type DefaultBehavior struct {
	WhenDisabled            string `yaml:"when_disabled"`
	LogUnauthorizedAttempts bool   `yaml:"log_unauthorized_attempts"`
}

// LoadServiceAuthorizationConfig loads the service authorization configuration from YAML file
// It supports environment-specific config files:
// - Development: config/service_permissions.dev.yaml
// - Production: config/service_permissions.yaml
func LoadServiceAuthorizationConfig() (*ServiceAuthorizationConfig, error) {
	// Determine which config file to load based on environment
	env := os.Getenv("AAA_ENV")
	configFile := "config/service_permissions.yaml"

	if env == "development" || env == "dev" {
		devConfigFile := "config/service_permissions.dev.yaml"
		if _, err := os.Stat(devConfigFile); err == nil {
			configFile = devConfigFile
		}
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Return default configuration if file doesn't exist
		return &ServiceAuthorizationConfig{
			ServiceAuthorization: ServiceAuthSection{
				Enabled:  false,
				Services: make(map[string]ServicePermission),
			},
			DefaultBehavior: DefaultBehavior{
				WhenDisabled:            "allow_all",
				LogUnauthorizedAttempts: true,
			},
		}, nil
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	// Parse YAML
	var config ServiceAuthorizationConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadServiceAuthorizationConfigFromPath loads config from a specific file path
func LoadServiceAuthorizationConfigFromPath(configPath string) (*ServiceAuthorizationConfig, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", absPath)
	}

	// Read config file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config ServiceAuthorizationConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig validates the service authorization configuration
func validateConfig(config *ServiceAuthorizationConfig) error {
	// Validate default behavior
	if config.DefaultBehavior.WhenDisabled != "allow_all" && config.DefaultBehavior.WhenDisabled != "deny_all" {
		return fmt.Errorf("default_behavior.when_disabled must be 'allow_all' or 'deny_all', got '%s'",
			config.DefaultBehavior.WhenDisabled)
	}

	// If authorization is enabled, validate services
	if config.ServiceAuthorization.Enabled {
		for serviceID, service := range config.ServiceAuthorization.Services {
			// Validate service_id matches map key
			if service.ServiceID != serviceID {
				return fmt.Errorf("service_id '%s' does not match map key '%s'", service.ServiceID, serviceID)
			}

			// Validate required fields
			if service.ServiceID == "" {
				return fmt.Errorf("service_id is required")
			}
			if service.DisplayName == "" {
				return fmt.Errorf("display_name is required for service '%s'", serviceID)
			}

			// Validate permissions format
			for _, perm := range service.Permissions {
				if !isValidPermission(perm) {
					return fmt.Errorf("invalid permission format '%s' for service '%s', expected format: resource:action or resource:*",
						perm, serviceID)
				}
			}

			// Warn if API key is required but not provided
			if service.APIKeyRequired && service.APIKey == "" {
				// This is not an error, as API keys might be provided via environment variables
				// but we should validate at runtime
			}
		}
	}

	return nil
}

// isValidPermission validates permission format: resource:action or resource:*
func isValidPermission(permission string) bool {
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return false
	}

	resource := parts[0]
	action := parts[1]

	// Resource must not be empty
	if resource == "" {
		return false
	}

	// Action must not be empty (can be * for wildcard)
	if action == "" {
		return false
	}

	return true
}

// GetServicePermission retrieves permission configuration for a specific service
func (c *ServiceAuthorizationConfig) GetServicePermission(serviceID string) (ServicePermission, bool) {
	if !c.ServiceAuthorization.Enabled {
		return ServicePermission{}, false
	}

	service, exists := c.ServiceAuthorization.Services[serviceID]
	return service, exists
}

// IsEnabled returns whether service authorization is enabled
func (c *ServiceAuthorizationConfig) IsEnabled() bool {
	return c.ServiceAuthorization.Enabled
}

// GetDefaultBehavior returns the default behavior when authorization is disabled
func (c *ServiceAuthorizationConfig) GetDefaultBehavior() string {
	return c.DefaultBehavior.WhenDisabled
}

// ShouldLogUnauthorizedAttempts returns whether unauthorized attempts should be logged
func (c *ServiceAuthorizationConfig) ShouldLogUnauthorizedAttempts() bool {
	return c.DefaultBehavior.LogUnauthorizedAttempts
}
