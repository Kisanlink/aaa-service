package grpc_server

import (
	"context"
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/authorization"
	"github.com/Kisanlink/aaa-service/v2/internal/config"
	"github.com/Kisanlink/aaa-service/v2/internal/services/catalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func TestCatalogHandler_SeedRolesAndPermissions_Authorization(t *testing.T) {
	logger := zap.NewNop()

	// Create authorization checker with test configuration
	testConfig := &config.ServiceAuthorizationConfig{
		ServiceAuthorization: config.ServiceAuthSection{
			Enabled: true,
			Services: map[string]config.ServicePermission{
				"farmers-module": {
					ServiceID:      "farmers-module",
					DisplayName:    "Farmers Module",
					APIKeyRequired: true,
					APIKey:         "test-farmers-key",
					Permissions: []string{
						"catalog:seed_roles",
						"catalog:seed_permissions",
					},
				},
				"unauthorized-service": {
					ServiceID:      "unauthorized-service",
					DisplayName:    "Unauthorized Service",
					APIKeyRequired: false,
					Permissions: []string{
						"catalog:read", // Does not have seed_roles
					},
				},
			},
		},
		DefaultBehavior: config.DefaultBehavior{
			WhenDisabled:            "allow_all",
			LogUnauthorizedAttempts: true,
		},
	}

	serviceAuthorizer := authorization.NewServiceAuthorizer(testConfig, logger)

	tests := []struct {
		name               string
		serviceID          string
		targetServiceID    string
		apiKey             string
		expectedStatusCode int32
		expectedError      bool
	}{
		{
			name:               "authorized service with valid API key",
			serviceID:          "farmers-module",
			targetServiceID:    "farmers-module",
			apiKey:             "test-farmers-key",
			expectedStatusCode: 200,
			expectedError:      false,
		},
		{
			name:               "authorized service with invalid API key",
			serviceID:          "farmers-module",
			targetServiceID:    "farmers-module",
			apiKey:             "wrong-key",
			expectedStatusCode: 403,
			expectedError:      true,
		},
		{
			name:               "service without required permission",
			serviceID:          "unauthorized-service",
			targetServiceID:    "unauthorized-service",
			apiKey:             "",
			expectedStatusCode: 403,
			expectedError:      true,
		},
		// Note: Service ownership validation is done in CheckSeedPermission, not in Authorize
		// This test case would pass Authorize but fail later in the catalog handler
		{
			name:               "authorized service can authorize for permission check",
			serviceID:          "farmers-module",
			targetServiceID:    "other-service", // ownership check happens elsewhere
			apiKey:             "test-farmers-key",
			expectedStatusCode: 200,
			expectedError:      false, // Authorize only checks permission, not ownership
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with service metadata
			ctx := context.Background()
			ctx = context.WithValue(ctx, "service_id", tt.serviceID)
			ctx = context.WithValue(ctx, "service_name", tt.serviceID)
			ctx = context.WithValue(ctx, "principal_type", "service")

			// Add API key to metadata if provided
			if tt.apiKey != "" {
				md := metadata.Pairs("x-api-key", tt.apiKey)
				ctx = metadata.NewIncomingContext(ctx, md)
			}

			// Test authorization directly
			err := serviceAuthorizer.Authorize(ctx, tt.serviceID, "catalog:seed_roles")

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceAuthorizer_Integration_Wildcard(t *testing.T) {
	logger := zap.NewNop()

	// Test wildcard permission matching
	cfg := &config.ServiceAuthorizationConfig{
		ServiceAuthorization: config.ServiceAuthSection{
			Enabled: true,
			Services: map[string]config.ServicePermission{
				"wildcard-service": {
					ServiceID:      "wildcard-service",
					DisplayName:    "Wildcard Service",
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
	}

	authorizer := authorization.NewServiceAuthorizer(cfg, logger)
	ctx := context.Background()

	// Test various catalog permissions
	permissions := []string{
		"catalog:seed_roles",
		"catalog:seed_permissions",
		"catalog:register_resource",
		"catalog:register_action",
		"catalog:delete",
		"catalog:update",
	}

	for _, perm := range permissions {
		t.Run(perm, func(t *testing.T) {
			err := authorizer.Authorize(ctx, "wildcard-service", perm)
			assert.NoError(t, err, "Wildcard service should have permission: %s", perm)
		})
	}

	// Test non-catalog permission should fail
	err := authorizer.Authorize(ctx, "wildcard-service", "users:create")
	assert.Error(t, err, "Wildcard service should not have non-catalog permission")
}

func TestServiceAuthorizer_Integration_Disabled(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name            string
		defaultBehavior string
		shouldAllow     bool
	}{
		{
			name:            "disabled with allow_all",
			defaultBehavior: "allow_all",
			shouldAllow:     true,
		},
		{
			name:            "disabled with deny_all",
			defaultBehavior: "deny_all",
			shouldAllow:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServiceAuthorizationConfig{
				ServiceAuthorization: config.ServiceAuthSection{
					Enabled:  false,
					Services: make(map[string]config.ServicePermission),
				},
				DefaultBehavior: config.DefaultBehavior{
					WhenDisabled:            tt.defaultBehavior,
					LogUnauthorizedAttempts: true,
				},
			}

			authorizer := authorization.NewServiceAuthorizer(cfg, logger)
			ctx := context.Background()

			err := authorizer.Authorize(ctx, "any-service", "catalog:seed_roles")

			if tt.shouldAllow {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateServiceID(t *testing.T) {
	tests := []struct {
		serviceID     string
		expectedError bool
	}{
		{"farmers-module", false},
		{"erp-module", false},
		{"test-service", false},
		{"", false}, // Empty is allowed for backward compatibility
	}

	for _, tt := range tests {
		t.Run(tt.serviceID, func(t *testing.T) {
			err := catalog.ValidateServiceID(tt.serviceID)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizationChecker_ServicePrincipalFlow(t *testing.T) {
	logger := zap.NewNop()

	// This test validates the complete authorization flow for service principals
	// from context extraction through configuration-based authorization

	cfg := &config.ServiceAuthorizationConfig{
		ServiceAuthorization: config.ServiceAuthSection{
			Enabled: true,
			Services: map[string]config.ServicePermission{
				"farmers-module": {
					ServiceID:      "farmers-module",
					DisplayName:    "Farmers Module",
					APIKeyRequired: true,
					APIKey:         "secure-key-123",
					Permissions: []string{
						"catalog:seed_roles",
					},
				},
			},
		},
		DefaultBehavior: config.DefaultBehavior{
			WhenDisabled:            "allow_all",
			LogUnauthorizedAttempts: true,
		},
	}

	authorizer := authorization.NewServiceAuthorizer(cfg, logger)

	// Test successful authorization
	t.Run("successful authorization with valid API key", func(t *testing.T) {
		ctx := context.Background()
		md := metadata.Pairs("x-api-key", "secure-key-123")
		ctx = metadata.NewIncomingContext(ctx, md)

		err := authorizer.Authorize(ctx, "farmers-module", "catalog:seed_roles")
		require.NoError(t, err)
	})

	// Test failed authorization - wrong API key
	t.Run("failed authorization with wrong API key", func(t *testing.T) {
		ctx := context.Background()
		md := metadata.Pairs("x-api-key", "wrong-key")
		ctx = metadata.NewIncomingContext(ctx, md)

		err := authorizer.Authorize(ctx, "farmers-module", "catalog:seed_roles")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid API key")
	})

	// Test failed authorization - missing permission
	t.Run("failed authorization with missing permission", func(t *testing.T) {
		ctx := context.Background()
		md := metadata.Pairs("x-api-key", "secure-key-123")
		ctx = metadata.NewIncomingContext(ctx, md)

		err := authorizer.Authorize(ctx, "farmers-module", "catalog:delete_roles")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not have permission")
	})
}
