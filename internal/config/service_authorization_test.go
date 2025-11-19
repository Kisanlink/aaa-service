package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadServiceAuthorizationConfig(t *testing.T) {
	// Test loading with non-existent file (should return default config)
	t.Run("non-existent config returns defaults", func(t *testing.T) {
		os.Setenv("AAA_ENV", "test-nonexistent")
		defer os.Unsetenv("AAA_ENV")

		cfg, err := LoadServiceAuthorizationConfig()
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.False(t, cfg.IsEnabled())
		assert.Equal(t, "allow_all", cfg.GetDefaultBehavior())
		assert.True(t, cfg.ShouldLogUnauthorizedAttempts())
	})

	// Test loading development config
	t.Run("development environment uses dev config", func(t *testing.T) {
		// Create a temporary dev config
		tmpDir := t.TempDir()
		devConfigPath := filepath.Join(tmpDir, "service_permissions.dev.yaml")

		devContent := `
service_authorization:
  enabled: false

default_behavior:
  when_disabled: "allow_all"
  log_unauthorized_attempts: true
`
		err := os.WriteFile(devConfigPath, []byte(devContent), 0644)
		require.NoError(t, err)

		// Change to temp directory
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		os.Chdir(tmpDir)

		os.Setenv("AAA_ENV", "development")
		defer os.Unsetenv("AAA_ENV")

		cfg, err := LoadServiceAuthorizationConfig()
		require.NoError(t, err)
		assert.False(t, cfg.IsEnabled())
	})
}

func TestLoadServiceAuthorizationConfigFromPath(t *testing.T) {
	t.Run("valid config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test_config.yaml")

		content := `
service_authorization:
  enabled: true
  services:
    test-service:
      service_id: "test-service"
      display_name: "Test Service"
      description: "Test service description"
      api_key_required: true
      permissions:
        - "catalog:seed_roles"
        - "catalog:*"

default_behavior:
  when_disabled: "allow_all"
  log_unauthorized_attempts: true
`
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		cfg, err := LoadServiceAuthorizationConfigFromPath(configPath)
		require.NoError(t, err)
		assert.True(t, cfg.IsEnabled())

		service, exists := cfg.GetServicePermission("test-service")
		assert.True(t, exists)
		assert.Equal(t, "Test Service", service.DisplayName)
		assert.Equal(t, 2, len(service.Permissions))
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := LoadServiceAuthorizationConfigFromPath("/nonexistent/path/config.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")

		content := `
invalid yaml content
  this is not: valid: yaml:
    - broken
`
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		_, err = LoadServiceAuthorizationConfigFromPath(configPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        *ServiceAuthorizationConfig
		expectedError bool
		errorContains string
	}{
		{
			name: "valid config",
			config: &ServiceAuthorizationConfig{
				ServiceAuthorization: ServiceAuthSection{
					Enabled: true,
					Services: map[string]ServicePermission{
						"test-service": {
							ServiceID:   "test-service",
							DisplayName: "Test Service",
							Permissions: []string{"catalog:seed_roles"},
						},
					},
				},
				DefaultBehavior: DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			expectedError: false,
		},
		{
			name: "invalid default behavior",
			config: &ServiceAuthorizationConfig{
				ServiceAuthorization: ServiceAuthSection{
					Enabled:  false,
					Services: make(map[string]ServicePermission),
				},
				DefaultBehavior: DefaultBehavior{
					WhenDisabled:            "invalid",
					LogUnauthorizedAttempts: true,
				},
			},
			expectedError: true,
			errorContains: "must be 'allow_all' or 'deny_all'",
		},
		{
			name: "service_id mismatch",
			config: &ServiceAuthorizationConfig{
				ServiceAuthorization: ServiceAuthSection{
					Enabled: true,
					Services: map[string]ServicePermission{
						"key-id": {
							ServiceID:   "different-id",
							DisplayName: "Test",
							Permissions: []string{"catalog:*"},
						},
					},
				},
				DefaultBehavior: DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			expectedError: true,
			errorContains: "does not match map key",
		},
		{
			name: "missing display name",
			config: &ServiceAuthorizationConfig{
				ServiceAuthorization: ServiceAuthSection{
					Enabled: true,
					Services: map[string]ServicePermission{
						"test-service": {
							ServiceID:   "test-service",
							DisplayName: "",
							Permissions: []string{"catalog:*"},
						},
					},
				},
				DefaultBehavior: DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			expectedError: true,
			errorContains: "display_name is required",
		},
		{
			name: "invalid permission format",
			config: &ServiceAuthorizationConfig{
				ServiceAuthorization: ServiceAuthSection{
					Enabled: true,
					Services: map[string]ServicePermission{
						"test-service": {
							ServiceID:   "test-service",
							DisplayName: "Test",
							Permissions: []string{"invalid-permission"},
						},
					},
				},
				DefaultBehavior: DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			expectedError: true,
			errorContains: "invalid permission format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidPermission(t *testing.T) {
	tests := []struct {
		permission string
		valid      bool
	}{
		{"catalog:seed_roles", true},
		{"catalog:*", true},
		{"*:*", true},
		{"users:create", true},
		{"resource:action", true},
		{"invalid", false},
		{"too:many:parts", false},
		{":action", false},
		{"resource:", false},
		{"", false},
		{":", false},
	}

	for _, tt := range tests {
		t.Run(tt.permission, func(t *testing.T) {
			result := isValidPermission(tt.permission)
			assert.Equal(t, tt.valid, result, "Permission: %s", tt.permission)
		})
	}
}

func TestServiceAuthorizationConfig_Methods(t *testing.T) {
	cfg := &ServiceAuthorizationConfig{
		ServiceAuthorization: ServiceAuthSection{
			Enabled: true,
			Services: map[string]ServicePermission{
				"test-service": {
					ServiceID:   "test-service",
					DisplayName: "Test Service",
					Permissions: []string{"catalog:seed_roles"},
				},
			},
		},
		DefaultBehavior: DefaultBehavior{
			WhenDisabled:            "deny_all",
			LogUnauthorizedAttempts: false,
		},
	}

	t.Run("IsEnabled", func(t *testing.T) {
		assert.True(t, cfg.IsEnabled())

		cfg.ServiceAuthorization.Enabled = false
		assert.False(t, cfg.IsEnabled())
	})

	t.Run("GetServicePermission", func(t *testing.T) {
		cfg.ServiceAuthorization.Enabled = true

		service, exists := cfg.GetServicePermission("test-service")
		assert.True(t, exists)
		assert.Equal(t, "Test Service", service.DisplayName)

		_, exists = cfg.GetServicePermission("nonexistent")
		assert.False(t, exists)
	})

	t.Run("GetServicePermission when disabled", func(t *testing.T) {
		cfg.ServiceAuthorization.Enabled = false

		_, exists := cfg.GetServicePermission("test-service")
		assert.False(t, exists)
	})

	t.Run("GetDefaultBehavior", func(t *testing.T) {
		assert.Equal(t, "deny_all", cfg.GetDefaultBehavior())
	})

	t.Run("ShouldLogUnauthorizedAttempts", func(t *testing.T) {
		assert.False(t, cfg.ShouldLogUnauthorizedAttempts())

		cfg.DefaultBehavior.LogUnauthorizedAttempts = true
		assert.True(t, cfg.ShouldLogUnauthorizedAttempts())
	})
}

func TestServiceAuthorizationConfig_WildcardPermissions(t *testing.T) {
	cfg := &ServiceAuthorizationConfig{
		ServiceAuthorization: ServiceAuthSection{
			Enabled: true,
			Services: map[string]ServicePermission{
				"wildcard-service": {
					ServiceID:   "wildcard-service",
					DisplayName: "Wildcard Service",
					Permissions: []string{
						"catalog:*",
						"*:*",
					},
				},
			},
		},
		DefaultBehavior: DefaultBehavior{
			WhenDisabled:            "allow_all",
			LogUnauthorizedAttempts: true,
		},
	}

	err := validateConfig(cfg)
	assert.NoError(t, err, "Wildcard permissions should be valid")

	service, exists := cfg.GetServicePermission("wildcard-service")
	assert.True(t, exists)
	assert.Contains(t, service.Permissions, "catalog:*")
	assert.Contains(t, service.Permissions, "*:*")
}
