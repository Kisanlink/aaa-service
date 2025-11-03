package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PostgresAuthorizationService provides authorization services using PostgreSQL
type PostgresAuthorizationService struct {
	db           *gorm.DB
	cacheService interfaces.CacheService
	auditService *AuditService
	logger       *zap.Logger
}

// NewPostgresAuthorizationService creates a new PostgreSQL-based authorization service
func NewPostgresAuthorizationService(
	db *gorm.DB,
	cacheService interfaces.CacheService,
	auditService *AuditService,
	logger *zap.Logger,
) *PostgresAuthorizationService {
	return &PostgresAuthorizationService{
		db:           db,
		cacheService: cacheService,
		auditService: auditService,
		logger:       logger,
	}
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (s *PostgresAuthorizationService) CheckPermission(ctx context.Context, perm *Permission) (*PermissionResult, error) {
	// Create cache key for permission check
	cacheKey := fmt.Sprintf("permission:%s:%s:%s:%s", perm.UserID, perm.Resource, perm.ResourceID, perm.Action)

	// Try to get result from cache first
	if cachedResult, exists := s.cacheService.Get(cacheKey); exists {
		if result, ok := cachedResult.(*PermissionResult); ok {
			return result, nil
		}
	}

	// Check permission in database
	allowed, reason, err := s.checkPermissionInDB(ctx, perm)
	if err != nil {
		s.logger.Error("Failed to check permission",
			zap.String("user_id", perm.UserID),
			zap.String("resource", perm.Resource),
			zap.String("action", perm.Action),
			zap.Error(err))
		return nil, err
	}

	result := &PermissionResult{
		Allowed: allowed,
		Reason:  reason,
	}

	// Cache the result for 5 minutes (300 seconds)
	if err := s.cacheService.Set(cacheKey, result, 300); err != nil {
		s.logger.Warn("Failed to cache permission result", zap.String("key", cacheKey), zap.Error(err))
	}

	// Audit the permission check if denied
	if s.auditService != nil && !allowed {
		s.auditService.LogAccessDenied(ctx, perm.UserID, perm.Action, perm.Resource, perm.ResourceID, reason)
	}

	return result, nil
}

// checkPermissionInDB performs the actual permission check in the database
func (s *PostgresAuthorizationService) checkPermissionInDB(ctx context.Context, perm *Permission) (bool, string, error) {
	// Step 1: Get user's roles (including inherited from groups)
	userRoles, err := s.getUserRoles(ctx, perm.UserID)
	if err != nil {
		return false, "Failed to fetch user roles", err
	}

	if len(userRoles) == 0 {
		return false, "User has no roles", nil
	}

	// Step 2: Check if any role has the required permission for the resource
	for _, role := range userRoles {
		hasPermission, err := s.roleHasPermission(ctx, role.ID, perm.Resource, perm.ResourceID, perm.Action)
		if err != nil {
			s.logger.Warn("Failed to check role permission",
				zap.String("role_id", role.ID),
				zap.Error(err))
			continue
		}
		if hasPermission {
			return true, fmt.Sprintf("Permission granted through role: %s", role.Name), nil
		}
	}

	// Step 3: Check for wildcard permissions (e.g., admin roles)
	for _, role := range userRoles {
		if s.roleHasWildcardPermission(ctx, role) {
			return true, fmt.Sprintf("Permission granted through admin role: %s", role.Name), nil
		}
	}

	return false, "No matching permissions found", nil
}

// getUserRoles gets all roles for a user, including inherited ones
func (s *PostgresAuthorizationService) getUserRoles(ctx context.Context, userID string) ([]models.Role, error) {
	var roles []models.Role

	// Direct user roles
	err := s.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.is_active = ? AND roles.is_active = ?", userID, true, true).
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	// Get roles inherited from groups
	var groupRoles []models.Role
	err = s.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Joins("JOIN group_memberships ON user_roles.user_id = group_memberships.principal_id").
		Where("group_memberships.principal_id = ? AND group_memberships.principal_type = ? AND group_memberships.is_active = ?",
			userID, "user", true).
		Where("roles.is_active = ?", true).
		Find(&groupRoles).Error

	if err != nil {
		s.logger.Warn("Failed to fetch group roles", zap.Error(err))
	} else {
		roles = append(roles, groupRoles...)
	}

	// Include parent roles if hierarchical
	roles = s.includeParentRoles(ctx, roles)

	return roles, nil
}

// includeParentRoles adds parent roles for hierarchical role inheritance
func (s *PostgresAuthorizationService) includeParentRoles(ctx context.Context, roles []models.Role) []models.Role {
	roleMap := make(map[string]models.Role)
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	for _, role := range roles {
		if role.ParentID != nil && *role.ParentID != "" {
			var parentRole models.Role
			if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", *role.ParentID, true).First(&parentRole).Error; err == nil {
				if _, exists := roleMap[parentRole.ID]; !exists {
					roleMap[parentRole.ID] = parentRole
				}
			}
		}
	}

	result := make([]models.Role, 0, len(roleMap))
	for _, role := range roleMap {
		result = append(result, role)
	}
	return result
}

// roleHasPermission checks if a role has a specific permission for a resource
// SECURITY FIX: Validates that resource_type matches the permission name pattern to prevent
// authorization bypass where users with address_read could access other resource types.
// Permissions follow the naming convention: {resource_type}_{action}
// Examples: address_read, attachment_create, collaborator_update
func (s *PostgresAuthorizationService) roleHasPermission(ctx context.Context, roleID, resourceType, resourceID, action string) (bool, error) {
	var count int64

	// Check ResourcePermission table
	query := s.db.WithContext(ctx).
		Table("resource_permissions").
		Where("role_id = ? AND resource_type = ? AND is_active = ?", roleID, resourceType, true)

	// Check for specific resource or wildcard
	if resourceID != "" {
		query = query.Where("(resource_id = ? OR resource_id = '*')", resourceID)
	}

	// Check for specific action
	query = query.Where("resource_permissions.action = ?", action)

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// SECURITY FIX: Check RolePermission table for general permissions
	// BUT validate that the permission name matches {resource_type}_{action} pattern.
	// Previously this checked (actions.name = ? OR permissions.name = ?) with only the action,
	// which allowed ANY permission with that action name regardless of resource_type.
	// This caused users with "address_read" to be able to access "attachment" resources.
	expectedPermissionName := resourceType + "_" + action

	err := s.db.WithContext(ctx).
		Table("role_permissions").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ? AND role_permissions.is_active = ?", roleID, true).
		Where("permissions.is_active = ? AND permissions.name = ?", true, expectedPermissionName).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// roleHasWildcardPermission checks if a role has admin/wildcard permissions
func (s *PostgresAuthorizationService) roleHasWildcardPermission(ctx context.Context, role models.Role) bool {
	// Check for super admin or global admin roles
	adminRoleNames := []string{"super_admin", "admin", "system_admin"}
	for _, adminName := range adminRoleNames {
		if role.Name == adminName {
			return true
		}
	}

	// Check if role has manage or admin permissions
	var count int64
	s.db.WithContext(ctx).
		Table("role_permissions").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ? AND role_permissions.is_active = ?", role.ID, true).
		Where("permissions.name IN (?) AND permissions.is_active = ?", []string{"manage", "admin", "super_admin"}, true).
		Count(&count)

	return count > 0
}

// CheckBulkPermissions checks multiple permissions at once
func (s *PostgresAuthorizationService) CheckBulkPermissions(ctx context.Context, req *BulkPermissionRequest) (*BulkPermissionResult, error) {
	results := make(map[string]*PermissionResult)

	for _, perm := range req.Permissions {
		perm.UserID = req.UserID
		result, err := s.CheckPermission(ctx, &perm)
		if err != nil {
			results[fmt.Sprintf("%s:%s:%s", perm.Resource, perm.ResourceID, perm.Action)] = &PermissionResult{
				Allowed: false,
				Reason:  fmt.Sprintf("Error: %v", err),
			}
		} else {
			results[fmt.Sprintf("%s:%s:%s", perm.Resource, perm.ResourceID, perm.Action)] = result
		}
	}

	return &BulkPermissionResult{Results: results}, nil
}

// GrantPermission grants a permission to a role for a resource
func (s *PostgresAuthorizationService) GrantPermission(ctx context.Context, roleID, resourceType, resourceID, actionName string) error {
	// Find or create the action
	var action models.Action
	if err := s.db.WithContext(ctx).Where("name = ?", actionName).First(&action).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			action = *models.NewAction(actionName, fmt.Sprintf("Action: %s", actionName))
			if err := s.db.WithContext(ctx).Create(&action).Error; err != nil {
				return fmt.Errorf("failed to create action: %w", err)
			}
		} else {
			return fmt.Errorf("failed to find action: %w", err)
		}
	}

	// Create ResourcePermission
	resourcePerm := models.NewResourcePermission(resourceID, resourceType, roleID, actionName)
	if err := s.db.WithContext(ctx).Create(resourcePerm).Error; err != nil {
		return fmt.Errorf("failed to grant permission: %w", err)
	}

	// Invalidate cache for this role
	s.invalidateRoleCache(roleID)

	return nil
}

// RevokePermission revokes a permission from a role for a resource
func (s *PostgresAuthorizationService) RevokePermission(ctx context.Context, roleID, resourceType, resourceID, actionName string) error {
	// Find the action
	var action models.Action
	if err := s.db.WithContext(ctx).Where("name = ?", actionName).First(&action).Error; err != nil {
		return fmt.Errorf("action not found: %w", err)
	}

	// Soft delete the ResourcePermission
	err := s.db.WithContext(ctx).
		Model(&models.ResourcePermission{}).
		Where("role_id = ? AND resource_type = ? AND resource_id = ? AND action = ?",
			roleID, resourceType, resourceID, actionName).
		Update("is_active", false).Error

	if err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	// Invalidate cache for this role
	s.invalidateRoleCache(roleID)

	return nil
}

// AssignRoleToUser assigns a role to a user
func (s *PostgresAuthorizationService) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	userRole := models.NewUserRole(userID, roleID)
	if err := s.db.WithContext(ctx).Create(userRole).Error; err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Invalidate cache for this user
	s.invalidateUserCache(userID)

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (s *PostgresAuthorizationService) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	err := s.db.WithContext(ctx).
		Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Update("is_active", false).Error

	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	// Invalidate cache for this user
	s.invalidateUserCache(userID)

	return nil
}

// GetUserPermissions gets all permissions for a user
func (s *PostgresAuthorizationService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	roles, err := s.getUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissionMap := make(map[string]bool)

	for _, role := range roles {
		var permissions []models.Permission
		err := s.db.WithContext(ctx).
			Table("permissions").
			Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
			Where("role_permissions.role_id = ? AND role_permissions.is_active = ?", role.ID, true).
			Where("permissions.is_active = ?", true).
			Find(&permissions).Error

		if err != nil {
			s.logger.Warn("Failed to fetch role permissions",
				zap.String("role_id", role.ID),
				zap.Error(err))
			continue
		}

		for _, perm := range permissions {
			permissionMap[perm.Name] = true
		}
	}

	result := make([]string, 0, len(permissionMap))
	for perm := range permissionMap {
		result = append(result, perm)
	}

	return result, nil
}

// invalidateUserCache invalidates all cached permissions for a user
func (s *PostgresAuthorizationService) invalidateUserCache(userID string) {
	// Implementation depends on cache service capabilities
	// For now, we'll just log it
	s.logger.Info("Invalidating cache for user", zap.String("user_id", userID))
}

// invalidateRoleCache invalidates all cached permissions for a role
func (s *PostgresAuthorizationService) invalidateRoleCache(roleID string) {
	// Implementation depends on cache service capabilities
	// For now, we'll just log it
	s.logger.Info("Invalidating cache for role", zap.String("role_id", roleID))
}

// CreateRole creates a new role
func (s *PostgresAuthorizationService) CreateRole(ctx context.Context, name, description string, scope models.RoleScope) (*models.Role, error) {
	role := models.NewRole(name, description, scope)
	if err := s.db.WithContext(ctx).Create(role).Error; err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	return role, nil
}

// DeleteRole deletes a role
func (s *PostgresAuthorizationService) DeleteRole(ctx context.Context, roleID string) error {
	// Soft delete the role
	err := s.db.WithContext(ctx).
		Model(&models.Role{}).
		Where("id = ?", roleID).
		Update("is_active", false).Error

	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// Also deactivate all user-role associations
	s.db.WithContext(ctx).
		Model(&models.UserRole{}).
		Where("role_id = ?", roleID).
		Update("is_active", false)

	// Invalidate cache for this role
	s.invalidateRoleCache(roleID)

	return nil
}

// ListRoles lists all active roles
func (s *PostgresAuthorizationService) ListRoles(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	err := s.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&roles).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

// GetRole gets a role by ID
func (s *PostgresAuthorizationService) GetRole(ctx context.Context, roleID string) (*models.Role, error) {
	var role models.Role
	err := s.db.WithContext(ctx).
		Where("id = ? AND is_active = ?", roleID, true).
		First(&role).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}
