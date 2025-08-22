package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthorizationService provides authorization services using PostgreSQL
type AuthorizationService struct {
	postgresAuth *PostgresAuthorizationService
	cacheService interfaces.CacheService
	auditService *AuditService
	logger       *zap.Logger
}

// AuthorizationServiceConfig contains configuration for AuthorizationService
type AuthorizationServiceConfig struct {
	DB *gorm.DB
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(
	cacheService interfaces.CacheService,
	auditService *AuditService,
	config *AuthorizationServiceConfig,
	logger *zap.Logger,
) (*AuthorizationService, error) {
	// Initialize PostgreSQL authorization service
	postgresAuth := NewPostgresAuthorizationService(
		config.DB,
		cacheService,
		auditService,
		logger,
	)

	return &AuthorizationService{
		postgresAuth: postgresAuth,
		cacheService: cacheService,
		auditService: auditService,
		logger:       logger,
	}, nil
}

// Permission represents a permission check request
type Permission struct {
	UserID     string `json:"user_id"`
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
	Action     string `json:"action"`
}

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Allowed          bool     `json:"allowed"`
	Reason           string   `json:"reason,omitempty"`
	Permissions      []string `json:"permissions,omitempty"`
	DecisionID       string   `json:"decision_id,omitempty"`
	ConsistencyToken string   `json:"consistency_token,omitempty"`
}

// BulkPermissionRequest represents a bulk permission check request
type BulkPermissionRequest struct {
	UserID      string       `json:"user_id"`
	Permissions []Permission `json:"permissions"`
}

// BulkPermissionResult represents the result of a bulk permission check
type BulkPermissionResult struct {
	Results map[string]*PermissionResult `json:"results"`
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (s *AuthorizationService) CheckPermission(ctx context.Context, perm *Permission) (*PermissionResult, error) {
	return s.postgresAuth.CheckPermission(ctx, perm)
}

// CheckBulkPermissions checks multiple permissions for a user
func (s *AuthorizationService) CheckBulkPermissions(ctx context.Context, req *BulkPermissionRequest) (*BulkPermissionResult, error) {
	return s.postgresAuth.CheckBulkPermissions(ctx, req)
}

// GetUserPermissions retrieves all permissions for a user
func (s *AuthorizationService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return s.postgresAuth.GetUserPermissions(ctx, userID)
}

// GrantPermission grants a permission to a role for a resource
func (s *AuthorizationService) GrantPermission(ctx context.Context, roleID, resourceType, resourceID, action string) error {
	return s.postgresAuth.GrantPermission(ctx, roleID, resourceType, resourceID, action)
}

// RevokePermission revokes a permission from a role for a resource
func (s *AuthorizationService) RevokePermission(ctx context.Context, roleID, resourceType, resourceID, action string) error {
	return s.postgresAuth.RevokePermission(ctx, roleID, resourceType, resourceID, action)
}

// AssignRoleToUser assigns a role to a user
func (s *AuthorizationService) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	return s.postgresAuth.AssignRoleToUser(ctx, userID, roleID)
}

// RemoveRoleFromUser removes a role from a user
func (s *AuthorizationService) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	return s.postgresAuth.RemoveRoleFromUser(ctx, userID, roleID)
}

// CreateRole creates a new role
func (s *AuthorizationService) CreateRole(ctx context.Context, name, description string, scope models.RoleScope) (*models.Role, error) {
	return s.postgresAuth.CreateRole(ctx, name, description, scope)
}

// DeleteRole deletes a role
func (s *AuthorizationService) DeleteRole(ctx context.Context, roleID string) error {
	return s.postgresAuth.DeleteRole(ctx, roleID)
}

// ListRoles lists all active roles
func (s *AuthorizationService) ListRoles(ctx context.Context) ([]models.Role, error) {
	return s.postgresAuth.ListRoles(ctx)
}

// GetRole gets a role by ID
func (s *AuthorizationService) GetRole(ctx context.Context, roleID string) (*models.Role, error) {
	return s.postgresAuth.GetRole(ctx, roleID)
}

// WriteSchema is no longer needed with PostgreSQL-based RBAC
func (s *AuthorizationService) WriteSchema(ctx context.Context, schema string) error {
	s.logger.Info("WriteSchema called but not needed for PostgreSQL RBAC")
	return nil
}

// CreateRelationship creates a relationship between entities (replaced by role/permission assignments)
func (s *AuthorizationService) CreateRelationship(ctx context.Context, resourceType, resourceID, relation, subjectType, subjectID string) error {
	// This is now handled through role and permission assignments
	s.logger.Info("CreateRelationship called - redirecting to role assignment",
		zap.String("resource_type", resourceType),
		zap.String("resource_id", resourceID),
		zap.String("relation", relation),
		zap.String("subject_type", subjectType),
		zap.String("subject_id", subjectID))

	// Map the relationship to appropriate role/permission assignment
	if subjectType == "user" && relation == "member" {
		// This would be assigning a user to a resource with member role
		// You'd need to find or create the appropriate role and assign it
		return nil
	}

	return fmt.Errorf("relationship type not supported: %s", relation)
}

// DeleteRelationship deletes a relationship between entities
func (s *AuthorizationService) DeleteRelationship(ctx context.Context, resourceType, resourceID, relation, subjectType, subjectID string) error {
	// This is now handled through role and permission revocations
	s.logger.Info("DeleteRelationship called - redirecting to role removal",
		zap.String("resource_type", resourceType),
		zap.String("resource_id", resourceID),
		zap.String("relation", relation),
		zap.String("subject_type", subjectType),
		zap.String("subject_id", subjectID))

	return nil
}

// ListRelationships lists relationships for a resource (replaced by listing role assignments)
func (s *AuthorizationService) ListRelationships(ctx context.Context, resourceType, resourceID string) ([]interface{}, error) {
	// This would now list role assignments for a resource
	s.logger.Info("ListRelationships called for resource",
		zap.String("resource_type", resourceType),
		zap.String("resource_id", resourceID))

	return []interface{}{}, nil
}

// CheckRelationship checks if a relationship exists
func (s *AuthorizationService) CheckRelationship(ctx context.Context, resourceType, resourceID, relation, subjectType, subjectID string) (bool, error) {
	// Convert to permission check
	perm := &Permission{
		UserID:     subjectID,
		Resource:   resourceType,
		ResourceID: resourceID,
		Action:     relation,
	}

	result, err := s.CheckPermission(ctx, perm)
	if err != nil {
		return false, err
	}

	return result.Allowed, nil
}

// LookupResources finds resources a user has access to
func (s *AuthorizationService) LookupResources(ctx context.Context, userID, resourceType, permission string) ([]string, error) {
	// This would query the database for resources the user has access to through their roles
	s.logger.Info("LookupResources called",
		zap.String("user_id", userID),
		zap.String("resource_type", resourceType),
		zap.String("permission", permission))

	// Implementation would involve querying resource_permissions table
	// joined with user_roles to find accessible resources
	return []string{}, nil
}

// LookupSubjects finds subjects that have access to a resource
func (s *AuthorizationService) LookupSubjects(ctx context.Context, resourceType, resourceID, permission string) ([]string, error) {
	// This would query the database for users who have access to the resource
	s.logger.Info("LookupSubjects called",
		zap.String("resource_type", resourceType),
		zap.String("resource_id", resourceID),
		zap.String("permission", permission))

	// Implementation would involve querying user_roles and resource_permissions
	// to find users with the specified permission
	return []string{}, nil
}

// ExpandPermissions expands all permissions for a user on a resource
func (s *AuthorizationService) ExpandPermissions(ctx context.Context, userID, resourceType, resourceID string) ([]string, error) {
	// Get all permissions the user has on this specific resource
	roles, err := s.postgresAuth.getUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissionMap := make(map[string]bool)

	for _, role := range roles {
		// Get all actions this role can perform on the resource
		actions := []string{"view", "edit", "delete", "manage", "create", "read", "update"}
		for _, action := range actions {
			hasPermission, _ := s.postgresAuth.roleHasPermission(ctx, role.ID, resourceType, resourceID, action)
			if hasPermission {
				permissionMap[action] = true
			}
		}
	}

	result := make([]string, 0, len(permissionMap))
	for perm := range permissionMap {
		result = append(result, perm)
	}

	return result, nil
}

// Helper methods for common operations

// IsAdmin checks if a user has admin privileges
func (s *AuthorizationService) IsAdmin(ctx context.Context, userID string) (bool, error) {
	perm := &Permission{
		UserID:   userID,
		Resource: "system",
		Action:   "admin",
	}

	result, err := s.CheckPermission(ctx, perm)
	if err != nil {
		return false, err
	}

	return result.Allowed, nil
}

// HasRole checks if a user has a specific role
func (s *AuthorizationService) HasRole(ctx context.Context, userID, roleName string) (bool, error) {
	roles, err := s.postgresAuth.getUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == roleName {
			return true, nil
		}
	}

	return false, nil
}

// GetUserRoles gets all roles for a user
func (s *AuthorizationService) GetUserRoles(ctx context.Context, userID string) ([]models.Role, error) {
	return s.postgresAuth.getUserRoles(ctx, userID)
}

// CreatePermission creates a new permission
func (s *AuthorizationService) CreatePermission(ctx context.Context, name, description string, resourceID, actionID *string) (*models.Permission, error) {
	permission := models.NewPermission(name, description)
	if resourceID != nil {
		permission.ResourceID = resourceID
	}
	if actionID != nil {
		permission.ActionID = actionID
	}

	if err := s.postgresAuth.db.WithContext(ctx).Create(permission).Error; err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

// AssignPermissionToRole assigns a permission to a role
func (s *AuthorizationService) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	rolePermission := models.NewRolePermission(roleID, permissionID)
	if err := s.postgresAuth.db.WithContext(ctx).Create(rolePermission).Error; err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	// Invalidate cache for this role
	s.postgresAuth.invalidateRoleCache(roleID)

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (s *AuthorizationService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	err := s.postgresAuth.db.WithContext(ctx).
		Model(&models.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Update("is_active", false).Error

	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	// Invalidate cache for this role
	s.postgresAuth.invalidateRoleCache(roleID)

	return nil
}

// ValidatePermissionString validates if a permission string is valid
func (s *AuthorizationService) ValidatePermissionString(permission string) bool {
	// Check if the permission follows the expected format
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return false
	}

	// Validate resource type and action
	validResourceTypes := []string{"user", "role", "permission", "organization", "group", "system"}
	validActions := []string{"view", "edit", "delete", "manage", "create", "read", "update", "admin"}

	resourceType := parts[0]
	action := parts[1]

	resourceValid := false
	for _, valid := range validResourceTypes {
		if resourceType == valid {
			resourceValid = true
			break
		}
	}

	actionValid := false
	for _, valid := range validActions {
		if action == valid {
			actionValid = true
			break
		}
	}

	return resourceValid && actionValid
}

// ValidateAPIEndpointAccess validates if a user has access to an API endpoint
func (s *AuthorizationService) ValidateAPIEndpointAccess(ctx context.Context, userID, method, endpoint string) (bool, error) {
	// Convert HTTP method to action
	action := strings.ToLower(method)

	// Extract resource from endpoint (simplified logic - you may need to enhance this)
	// For example: /api/v1/users/123 -> resource: user, resourceID: 123
	parts := strings.Split(strings.TrimPrefix(endpoint, "/"), "/")
	resource := "api_endpoint"
	resourceID := endpoint

	if len(parts) >= 3 {
		// Try to extract resource type from path
		resource = parts[2] // e.g., "users", "roles", etc.
		if len(parts) >= 4 {
			resourceID = parts[3]
		}
	}

	perm := &Permission{
		UserID:     userID,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     action,
	}

	result, err := s.CheckPermission(ctx, perm)
	if err != nil {
		return false, err
	}

	return result.Allowed, nil
}

// Close cleans up resources (no longer needs to close SpiceDB connection)
func (s *AuthorizationService) Close() error {
	s.logger.Info("Closing authorization service")
	return nil
}
