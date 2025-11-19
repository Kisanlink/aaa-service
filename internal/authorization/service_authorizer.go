package authorization

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

// ServiceAuthorizer validates service permissions based on configuration
type ServiceAuthorizer struct {
	config *config.ServiceAuthorizationConfig
	logger *zap.Logger
}

// NewServiceAuthorizer creates a new service authorizer
func NewServiceAuthorizer(config *config.ServiceAuthorizationConfig, logger *zap.Logger) *ServiceAuthorizer {
	return &ServiceAuthorizer{
		config: config,
		logger: logger,
	}
}

// Authorize validates that a service has permission to perform an action
// Returns nil if authorized, error otherwise
func (sa *ServiceAuthorizer) Authorize(ctx context.Context, serviceID string, permission string) error {
	// If authorization is disabled, check default behavior
	if !sa.config.IsEnabled() {
		if sa.config.GetDefaultBehavior() == "allow_all" {
			sa.logger.Debug("Service authorization disabled, allowing all requests",
				zap.String("service_id", serviceID),
				zap.String("permission", permission))
			return nil
		}

		// deny_all behavior
		if sa.config.ShouldLogUnauthorizedAttempts() {
			sa.logger.Warn("Service authorization disabled with deny_all policy",
				zap.String("service_id", serviceID),
				zap.String("permission", permission))
		}
		return fmt.Errorf("service authorization is disabled with deny_all policy")
	}

	// Validate service_id is not empty
	if serviceID == "" {
		sa.logger.Error("Empty service_id provided for authorization")
		return fmt.Errorf("service_id is required")
	}

	// Validate permission format
	if !isValidPermissionFormat(permission) {
		sa.logger.Error("Invalid permission format",
			zap.String("service_id", serviceID),
			zap.String("permission", permission))
		return fmt.Errorf("invalid permission format: %s, expected format: resource:action", permission)
	}

	// Get service configuration
	serviceConfig, exists := sa.config.GetServicePermission(serviceID)
	if !exists {
		if sa.config.ShouldLogUnauthorizedAttempts() {
			sa.logger.Warn("Service not authorized - not in configuration",
				zap.String("service_id", serviceID),
				zap.String("permission", permission))
		}
		return fmt.Errorf("service '%s' is not authorized", serviceID)
	}

	// Check if API key is required
	if serviceConfig.APIKeyRequired {
		if err := sa.validateAPIKey(ctx, serviceConfig); err != nil {
			if sa.config.ShouldLogUnauthorizedAttempts() {
				sa.logger.Warn("API key validation failed",
					zap.String("service_id", serviceID),
					zap.Error(err))
			}
			return err
		}
	}

	// Check if service has the required permission
	if !sa.hasPermission(serviceConfig, permission) {
		if sa.config.ShouldLogUnauthorizedAttempts() {
			sa.logger.Warn("Service does not have required permission",
				zap.String("service_id", serviceID),
				zap.String("permission", permission),
				zap.Strings("allowed_permissions", serviceConfig.Permissions))
		}
		return fmt.Errorf("service '%s' does not have permission '%s'", serviceID, permission)
	}

	// Authorization successful
	sa.logger.Debug("Service authorized successfully",
		zap.String("service_id", serviceID),
		zap.String("permission", permission))

	return nil
}

// hasPermission checks if a service has a specific permission
// Supports exact match and wildcard matching (e.g., catalog:*)
func (sa *ServiceAuthorizer) hasPermission(serviceConfig config.ServicePermission, requiredPermission string) bool {
	// Parse required permission
	parts := strings.Split(requiredPermission, ":")
	if len(parts) != 2 {
		return false
	}
	requiredResource := parts[0]
	_ = parts[1] // requiredAction - kept for future granular matching

	// Check each configured permission
	for _, configuredPermission := range serviceConfig.Permissions {
		// Parse configured permission
		configParts := strings.Split(configuredPermission, ":")
		if len(configParts) != 2 {
			continue
		}
		configResource := configParts[0]
		configAction := configParts[1]

		// Exact match
		if configuredPermission == requiredPermission {
			return true
		}

		// Wildcard match: catalog:* matches catalog:seed_roles
		if configResource == requiredResource && configAction == "*" {
			return true
		}

		// Global wildcard: *:* matches everything
		if configResource == "*" && configAction == "*" {
			return true
		}
	}

	return false
}

// validateAPIKey validates the API key from gRPC metadata
func (sa *ServiceAuthorizer) validateAPIKey(ctx context.Context, serviceConfig config.ServicePermission) error {
	// Extract API key from gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("no metadata in context")
	}

	// Check for x-api-key header
	apiKeys := md.Get("x-api-key")
	if len(apiKeys) == 0 {
		return fmt.Errorf("missing x-api-key header")
	}

	providedKey := apiKeys[0]
	if providedKey == "" {
		return fmt.Errorf("empty x-api-key header")
	}

	// TODO: Implement secure API key comparison (hashing)
	// For now, use plaintext comparison
	// In production, this should use bcrypt or similar
	expectedKey := serviceConfig.APIKey
	if expectedKey == "" {
		// Try to get API key from environment variable
		// Format: AAA_SERVICE_API_KEY_<SERVICE_ID_UPPERCASE>
		// Replace spaces and hyphens with underscores for valid env var names
		normalizedID := strings.ReplaceAll(serviceConfig.ServiceID, " ", "_")
		normalizedID = strings.ReplaceAll(normalizedID, "-", "_")
		envKey := fmt.Sprintf("AAA_SERVICE_API_KEY_%s", strings.ToUpper(normalizedID))
		expectedKey = getEnv(envKey, "")
	}

	if expectedKey == "" {
		sa.logger.Error("API key not configured for service",
			zap.String("service_id", serviceConfig.ServiceID))
		return fmt.Errorf("API key not configured for service '%s'", serviceConfig.ServiceID)
	}

	// Compare keys (TODO: use constant-time comparison)
	if providedKey != expectedKey {
		return fmt.Errorf("invalid API key")
	}

	return nil
}

// isValidPermissionFormat validates permission format: resource:action
func isValidPermissionFormat(permission string) bool {
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return false
	}

	resource := parts[0]
	action := parts[1]

	// Both parts must not be empty
	if resource == "" || action == "" {
		return false
	}

	return true
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
