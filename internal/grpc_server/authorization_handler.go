package grpc_server

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/services"
	"go.uber.org/zap"
)

// AuthorizationHandler implements authorization-related gRPC services
type AuthorizationHandler struct {
	authzService *services.AuthorizationService
	logger       *zap.Logger
}

// NewAuthorizationHandler creates a new authorization handler
func NewAuthorizationHandler(authzService *services.AuthorizationService, logger *zap.Logger) *AuthorizationHandler {
	return &AuthorizationHandler{
		authzService: authzService,
		logger:       logger,
	}
}

// CheckPermission checks if a user has permission to perform an action
func (h *AuthorizationHandler) CheckPermission(ctx context.Context, userID, resource, resourceID, action string) (bool, error) {
	h.logger.Info("gRPC CheckPermission request",
		zap.String("user_id", userID),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
		zap.String("action", action))

	permission := &services.Permission{
		UserID:     userID,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     action,
	}

	result, err := h.authzService.CheckPermission(ctx, permission)
	if err != nil {
		h.logger.Error("Permission check failed", zap.Error(err))
		return false, err
	}

	h.logger.Info("Permission check completed",
		zap.String("user_id", userID),
		zap.Bool("allowed", result.Allowed))

	return result.Allowed, nil
}

// CheckBulkPermissions checks multiple permissions for a user
func (h *AuthorizationHandler) CheckBulkPermissions(ctx context.Context, userID string, permissions []services.Permission) (map[string]bool, error) {
	h.logger.Info("gRPC CheckBulkPermissions request",
		zap.String("user_id", userID),
		zap.Int("permission_count", len(permissions)))

	request := &services.BulkPermissionRequest{
		UserID:      userID,
		Permissions: permissions,
	}

	result, err := h.authzService.CheckBulkPermissions(ctx, request)
	if err != nil {
		h.logger.Error("Bulk permission check failed", zap.Error(err))
		return nil, err
	}

	// Convert result to simple map
	resultMap := make(map[string]bool)
	for key, permResult := range result.Results {
		resultMap[key] = permResult.Allowed
	}

	h.logger.Info("Bulk permission check completed",
		zap.String("user_id", userID),
		zap.Int("results_count", len(resultMap)))

	return resultMap, nil
}

// GetUserPermissions retrieves all permissions for a user on a specific resource type
func (h *AuthorizationHandler) GetUserPermissions(ctx context.Context, userID, resourceType string) ([]string, error) {
	h.logger.Info("gRPC GetUserPermissions request",
		zap.String("user_id", userID),
		zap.String("resource_type", resourceType))

	permissions, err := h.authzService.GetUserPermissions(ctx, userID)
	if err != nil {
		h.logger.Error("Get user permissions failed", zap.Error(err))
		return nil, err
	}

	h.logger.Info("Get user permissions completed",
		zap.String("user_id", userID),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// GrantPermission grants a permission to a user
func (h *AuthorizationHandler) GrantPermission(ctx context.Context, userID, resource, resourceID, relation string) error {
	h.logger.Info("gRPC GrantPermission request",
		zap.String("user_id", userID),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
		zap.String("relation", relation))

	err := h.authzService.GrantPermission(ctx, userID, resource, resourceID, relation)
	if err != nil {
		h.logger.Error("Grant permission failed", zap.Error(err))
		return err
	}

	h.logger.Info("Permission granted successfully",
		zap.String("user_id", userID),
		zap.String("resource", resource))

	return nil
}

// RevokePermission revokes a permission from a user
func (h *AuthorizationHandler) RevokePermission(ctx context.Context, userID, resource, resourceID, relation string) error {
	h.logger.Info("gRPC RevokePermission request",
		zap.String("user_id", userID),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
		zap.String("relation", relation))

	err := h.authzService.RevokePermission(ctx, userID, resource, resourceID, relation)
	if err != nil {
		h.logger.Error("Revoke permission failed", zap.Error(err))
		return err
	}

	h.logger.Info("Permission revoked successfully",
		zap.String("user_id", userID),
		zap.String("resource", resource))

	return nil
}

// AssignRoleToUser assigns a role to a user
func (h *AuthorizationHandler) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	h.logger.Info("gRPC AssignRoleToUser request",
		zap.String("user_id", userID),
		zap.String("role_id", roleID))

	err := h.authzService.AssignRoleToUser(ctx, userID, roleID)
	if err != nil {
		h.logger.Error("Assign role failed", zap.Error(err))
		return err
	}

	h.logger.Info("Role assigned successfully",
		zap.String("user_id", userID),
		zap.String("role_id", roleID))

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (h *AuthorizationHandler) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	h.logger.Info("gRPC RemoveRoleFromUser request",
		zap.String("user_id", userID),
		zap.String("role_id", roleID))

	err := h.authzService.RemoveRoleFromUser(ctx, userID, roleID)
	if err != nil {
		h.logger.Error("Remove role failed", zap.Error(err))
		return err
	}

	h.logger.Info("Role removed successfully",
		zap.String("user_id", userID),
		zap.String("role_id", roleID))

	return nil
}

// ValidateAPIEndpointAccess validates access to API endpoints
func (h *AuthorizationHandler) ValidateAPIEndpointAccess(ctx context.Context, userID, method, endpoint string) (bool, error) {
	h.logger.Info("gRPC ValidateAPIEndpointAccess request",
		zap.String("user_id", userID),
		zap.String("method", method),
		zap.String("endpoint", endpoint))

	result, err := h.authzService.ValidateAPIEndpointAccess(ctx, userID, method, endpoint)
	if err != nil {
		h.logger.Error("API endpoint validation failed", zap.Error(err))
		return false, err
	}

	h.logger.Info("API endpoint validation completed",
		zap.String("user_id", userID),
		zap.Bool("allowed", result))

	return result, nil
}
