package authorization

import (
	"context"
	"os"
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func TestServiceAuthorizer_Authorize(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name          string
		config        *config.ServiceAuthorizationConfig
		serviceID     string
		permission    string
		ctx           context.Context
		expectedError bool
		errorContains string
	}{
		{
			name: "authorization disabled with allow_all - should allow",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled:  false,
					Services: make(map[string]config.ServicePermission),
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:     "test-service",
			permission:    "catalog:seed_roles",
			ctx:           context.Background(),
			expectedError: false,
		},
		{
			name: "authorization disabled with deny_all - should deny",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled:  false,
					Services: make(map[string]config.ServicePermission),
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "deny_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:     "test-service",
			permission:    "catalog:seed_roles",
			ctx:           context.Background(),
			expectedError: true,
			errorContains: "deny_all policy",
		},
		{
			name: "empty service_id - should fail",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled:  true,
					Services: make(map[string]config.ServicePermission),
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:     "",
			permission:    "catalog:seed_roles",
			ctx:           context.Background(),
			expectedError: true,
			errorContains: "service_id is required",
		},
		{
			name: "invalid permission format - should fail",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled:  true,
					Services: make(map[string]config.ServicePermission),
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:     "test-service",
			permission:    "invalid-permission",
			ctx:           context.Background(),
			expectedError: true,
			errorContains: "invalid permission format",
		},
		{
			name: "service not in configuration - should deny",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled:  true,
					Services: make(map[string]config.ServicePermission),
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:     "unknown-service",
			permission:    "catalog:seed_roles",
			ctx:           context.Background(),
			expectedError: true,
			errorContains: "not authorized",
		},
		{
			name: "service configured with exact permission - should allow",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled: true,
					Services: map[string]config.ServicePermission{
						"farmers-module": {
							ServiceID:      "farmers-module",
							DisplayName:    "Farmers Module",
							APIKeyRequired: false,
							Permissions: []string{
								"catalog:seed_roles",
								"catalog:seed_permissions",
							},
						},
					},
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:     "farmers-module",
			permission:    "catalog:seed_roles",
			ctx:           context.Background(),
			expectedError: false,
		},
		{
			name: "service configured with wildcard permission - should allow",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled: true,
					Services: map[string]config.ServicePermission{
						"admin-service": {
							ServiceID:      "admin-service",
							DisplayName:    "Admin Service",
							APIKeyRequired: false,
							Permissions: []string{
								"catalog:*",
							},
						},
					},
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:     "admin-service",
			permission:    "catalog:seed_roles",
			ctx:           context.Background(),
			expectedError: false,
		},
		{
			name: "service configured with global wildcard - should allow any permission",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled: true,
					Services: map[string]config.ServicePermission{
						"super-service": {
							ServiceID:      "super-service",
							DisplayName:    "Super Service",
							APIKeyRequired: false,
							Permissions: []string{
								"*:*",
							},
						},
					},
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:     "super-service",
			permission:    "anything:everything",
			ctx:           context.Background(),
			expectedError: false,
		},
		{
			name: "service configured but missing required permission - should deny",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled: true,
					Services: map[string]config.ServicePermission{
						"limited-service": {
							ServiceID:      "limited-service",
							DisplayName:    "Limited Service",
							APIKeyRequired: false,
							Permissions: []string{
								"catalog:read",
							},
						},
					},
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:     "limited-service",
			permission:    "catalog:seed_roles",
			ctx:           context.Background(),
			expectedError: true,
			errorContains: "does not have permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizer := NewServiceAuthorizer(tt.config, logger)
			err := authorizer.Authorize(tt.ctx, tt.serviceID, tt.permission)

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

func TestServiceAuthorizer_APIKeyValidation(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name          string
		config        *config.ServiceAuthorizationConfig
		serviceID     string
		permission    string
		ctx           context.Context
		envVars       map[string]string
		expectedError bool
		errorContains string
	}{
		{
			name: "API key required but missing metadata - should deny",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled: true,
					Services: map[string]config.ServicePermission{
						"secure-service": {
							ServiceID:      "secure-service",
							DisplayName:    "Secure Service",
							APIKeyRequired: true,
							APIKey:         "test-key-123",
							Permissions:    []string{"catalog:seed_roles"},
						},
					},
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:     "secure-service",
			permission:    "catalog:seed_roles",
			ctx:           context.Background(),
			expectedError: true,
			errorContains: "no metadata",
		},
		{
			name: "API key required but header missing - should deny",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled: true,
					Services: map[string]config.ServicePermission{
						"secure-service": {
							ServiceID:      "secure-service",
							DisplayName:    "Secure Service",
							APIKeyRequired: true,
							APIKey:         "test-key-123",
							Permissions:    []string{"catalog:seed_roles"},
						},
					},
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:  "secure-service",
			permission: "catalog:seed_roles",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				"other-header", "value",
			)),
			expectedError: true,
			errorContains: "missing x-api-key",
		},
		{
			name: "API key required with invalid key - should deny",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled: true,
					Services: map[string]config.ServicePermission{
						"secure-service": {
							ServiceID:      "secure-service",
							DisplayName:    "Secure Service",
							APIKeyRequired: true,
							APIKey:         "correct-key-123",
							Permissions:    []string{"catalog:seed_roles"},
						},
					},
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:  "secure-service",
			permission: "catalog:seed_roles",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				"x-api-key", "wrong-key",
			)),
			expectedError: true,
			errorContains: "invalid API key",
		},
		{
			name: "API key required with correct key - should allow",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled: true,
					Services: map[string]config.ServicePermission{
						"secure-service": {
							ServiceID:      "secure-service",
							DisplayName:    "Secure Service",
							APIKeyRequired: true,
							APIKey:         "correct-key-123",
							Permissions:    []string{"catalog:seed_roles"},
						},
					},
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:  "secure-service",
			permission: "catalog:seed_roles",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				"x-api-key", "correct-key-123",
			)),
			expectedError: false,
		},
		{
			name: "API key from environment variable - should allow",
			config: &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled: true,
					Services: map[string]config.ServicePermission{
						"env-service": {
							ServiceID:      "env-service",
							DisplayName:    "Env Service",
							APIKeyRequired: true,
							// No APIKey in config, should read from env
							Permissions: []string{"catalog:seed_roles"},
						},
					},
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            "allow_all",
					LogUnauthorizedAttempts: true,
				},
			},
			serviceID:  "env-service",
			permission: "catalog:seed_roles",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				"x-api-key", "env-key-456",
			)),
			envVars: map[string]string{
				"AAA_SERVICE_API_KEY_ENV_SERVICE": "env-key-456",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			authorizer := NewServiceAuthorizer(tt.config, logger)
			err := authorizer.Authorize(tt.ctx, tt.serviceID, tt.permission)

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

func TestServiceAuthorizer_PermissionMatching(t *testing.T) {
	logger := zap.NewNop()

	cfg := &config.ServiceAuthorizationConfig{
		ServiceAuthorization: config.ServiceAuthSection{
			Enabled: true,
			Services: map[string]config.ServicePermission{
				"test-service": {
					ServiceID:      "test-service",
					DisplayName:    "Test Service",
					APIKeyRequired: false,
					Permissions: []string{
						"catalog:seed_roles",
						"catalog:seed_permissions",
						"users:*",
						"admin:read",
					},
				},
			},
		},
		DefaultBehavior: config.DefaultBehavior{
			WhenDisabled:            "allow_all",
			LogUnauthorizedAttempts: true,
		},
	}

	authorizer := NewServiceAuthorizer(cfg, logger)
	ctx := context.Background()

	tests := []struct {
		permission    string
		shouldSucceed bool
	}{
		{"catalog:seed_roles", true},       // Exact match
		{"catalog:seed_permissions", true}, // Exact match
		{"users:create", true},             // Wildcard match (users:*)
		{"users:delete", true},             // Wildcard match (users:*)
		{"admin:read", true},               // Exact match
		{"admin:write", false},             // No match
		{"catalog:delete", false},          // No wildcard for catalog
		{"other:action", false},            // No match
	}

	for _, tt := range tests {
		t.Run(tt.permission, func(t *testing.T) {
			err := authorizer.Authorize(ctx, "test-service", tt.permission)
			if tt.shouldSucceed {
				assert.NoError(t, err, "Expected permission %s to be granted", tt.permission)
			} else {
				assert.Error(t, err, "Expected permission %s to be denied", tt.permission)
			}
		})
	}
}

func TestIsValidPermissionFormat(t *testing.T) {
	tests := []struct {
		permission string
		valid      bool
	}{
		{"catalog:seed_roles", true},
		{"catalog:*", true},
		{"*:*", true},
		{"users:create", true},
		{"invalid", false},
		{"too:many:parts", false},
		{":action", false},
		{"resource:", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.permission, func(t *testing.T) {
			result := isValidPermissionFormat(tt.permission)
			assert.Equal(t, tt.valid, result, "Permission: %s", tt.permission)
		})
	}
}

func TestLoadServiceAuthorizationConfig(t *testing.T) {
	// Test that loading non-existent config returns default config
	os.Setenv("AAA_ENV", "test-non-existent")
	defer os.Unsetenv("AAA_ENV")

	cfg, err := config.LoadServiceAuthorizationConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.False(t, cfg.IsEnabled())
	assert.Equal(t, "allow_all", cfg.GetDefaultBehavior())
}
