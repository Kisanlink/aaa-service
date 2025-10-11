package permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"go.uber.org/zap"
)

// GetPermissionByID retrieves a permission by ID with caching
func (s *Service) GetPermissionByID(ctx context.Context, id string) (*models.Permission, error) {
	if id == "" {
		return nil, fmt.Errorf("permission ID is required")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("permission:%s", id)
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			if permission, ok := cached.(*models.Permission); ok {
				s.logger.Debug("Permission cache hit", zap.String("id", id))
				return permission, nil
			}
		}
	}

	// Fetch from database
	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get permission by ID",
			zap.String("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	// Cache the result
	if s.cache != nil {
		if err := s.cache.Set(cacheKey, permission, 600); err != nil { // 10 minutes TTL
			s.logger.Warn("Failed to cache permission", zap.Error(err))
		}
	}

	return permission, nil
}

// GetPermissionByName retrieves a permission by name with caching
func (s *Service) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	if name == "" {
		return nil, fmt.Errorf("permission name is required")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("permission:name:%s", name)
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			if permission, ok := cached.(*models.Permission); ok {
				s.logger.Debug("Permission name cache hit", zap.String("name", name))
				return permission, nil
			}
		}
	}

	// Fetch from database
	permission, err := s.permissionRepo.GetByName(ctx, name)
	if err != nil {
		s.logger.Error("Failed to get permission by name",
			zap.String("name", name),
			zap.Error(err))
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	// Cache the result
	if s.cache != nil {
		if err := s.cache.Set(cacheKey, permission, 600); err != nil { // 10 minutes TTL
			s.logger.Warn("Failed to cache permission", zap.Error(err))
		}
	}

	return permission, nil
}

// ListPermissions retrieves permissions with filtering and pagination
func (s *Service) ListPermissions(ctx context.Context, filter *PermissionFilter) ([]*models.Permission, error) {
	if filter == nil {
		filter = &PermissionFilter{
			Limit:  50,
			Offset: 0,
		}
	}

	// Apply defaults
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}

	// Build database filter
	dbFilter := base.NewFilterBuilder()

	if filter.ResourceID != nil && *filter.ResourceID != "" {
		dbFilter.Where("resource_id", base.OpEqual, *filter.ResourceID)
	}

	if filter.ActionID != nil && *filter.ActionID != "" {
		dbFilter.Where("action_id", base.OpEqual, *filter.ActionID)
	}

	if filter.IsActive != nil {
		dbFilter.Where("is_active", base.OpEqual, *filter.IsActive)
	}

	dbFilter.Limit(filter.Limit, filter.Offset)

	// Fetch from database
	permissions, err := s.permissionRepo.FindByFilter(ctx, dbFilter.Build())
	if err != nil {
		s.logger.Error("Failed to list permissions", zap.Error(err))
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, nil
}

// GetPermissionsForRole retrieves all permissions assigned to a role
func (s *Service) GetPermissionsForRole(ctx context.Context, roleID string) ([]*models.Permission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("role:%s:permissions", roleID)
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			if permissions, ok := cached.([]*models.Permission); ok {
				s.logger.Debug("Role permissions cache hit", zap.String("role_id", roleID))
				return permissions, nil
			}
		}
	}

	// Get role-permission assignments
	rolePerms, err := s.rolePermissionRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		s.logger.Error("Failed to get role permissions",
			zap.String("role_id", roleID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	// Extract permissions
	permissions := make([]*models.Permission, 0, len(rolePerms))
	for _, rp := range rolePerms {
		if rp.Permission != nil {
			permissions = append(permissions, rp.Permission)
		} else if rp.PermissionID != "" {
			// Lazy load permission if not eager loaded
			perm, err := s.GetPermissionByID(ctx, rp.PermissionID)
			if err == nil {
				permissions = append(permissions, perm)
			}
		}
	}

	// Cache the result
	if s.cache != nil {
		if err := s.cache.Set(cacheKey, permissions, 600); err != nil { // 10 minutes TTL
			s.logger.Warn("Failed to cache role permissions", zap.Error(err))
		}
	}

	return permissions, nil
}

// GetPermissionsForResource retrieves all permissions for a specific resource
func (s *Service) GetPermissionsForResource(ctx context.Context, resourceID string) ([]*models.Permission, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("resource ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("resource_id", base.OpEqual, resourceID).
		Where("is_active", base.OpEqual, true).
		Build()

	permissions, err := s.permissionRepo.FindByFilter(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get permissions for resource",
			zap.String("resource_id", resourceID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions for resource: %w", err)
	}

	return permissions, nil
}

// GetPermissionsForAction retrieves all permissions for a specific action
func (s *Service) GetPermissionsForAction(ctx context.Context, actionID string) ([]*models.Permission, error) {
	if actionID == "" {
		return nil, fmt.Errorf("action ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("action_id", base.OpEqual, actionID).
		Where("is_active", base.OpEqual, true).
		Build()

	permissions, err := s.permissionRepo.FindByFilter(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get permissions for action",
			zap.String("action_id", actionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions for action: %w", err)
	}

	return permissions, nil
}
