package permissions

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"go.uber.org/zap"
)

// CreatePermission creates a new permission with validation
func (s *Service) CreatePermission(ctx context.Context, permission *models.Permission) error {
	if permission == nil {
		return fmt.Errorf("permission cannot be nil")
	}

	// Validate required fields
	if err := s.validatePermission(permission); err != nil {
		s.logger.Error("Permission validation failed", zap.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check for duplicate permission name
	existing, err := s.permissionRepo.GetByName(ctx, permission.Name)
	if err == nil && existing != nil {
		s.logger.Warn("Permission already exists",
			zap.String("name", permission.Name),
			zap.String("existing_id", existing.ID))
		return fmt.Errorf("permission with name '%s' already exists", permission.Name)
	}

	// Verify resource exists if provided
	if permission.ResourceID != nil && *permission.ResourceID != "" {
		if err := s.validateResource(ctx, *permission.ResourceID); err != nil {
			return fmt.Errorf("invalid resource: %w", err)
		}
	}

	// Verify action exists if provided
	if permission.ActionID != nil && *permission.ActionID != "" {
		if err := s.validateAction(ctx, *permission.ActionID); err != nil {
			return fmt.Errorf("invalid action: %w", err)
		}
	}

	// Create permission in database
	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		s.logger.Error("Failed to create permission",
			zap.String("name", permission.Name),
			zap.Error(err))
		return fmt.Errorf("failed to create permission: %w", err)
	}

	// Invalidate relevant caches
	s.invalidatePermissionRelatedCaches(ctx, permission)

	// Audit log
	if s.audit != nil {
		s.audit.LogPermissionChange(ctx, "", "create", "", permission.ID, permission.Name,
			map[string]interface{}{
				"name":        permission.Name,
				"description": permission.Description,
				"resource_id": permission.ResourceID,
				"action_id":   permission.ActionID,
			})
	}

	s.logger.Info("Permission created successfully",
		zap.String("permission_id", permission.ID),
		zap.String("name", permission.Name))

	return nil
}

// validatePermission validates permission fields
func (s *Service) validatePermission(permission *models.Permission) error {
	if permission.Name == "" {
		return fmt.Errorf("permission name is required")
	}

	// Validate name format: must be lowercase with underscores
	if !isValidPermissionName(permission.Name) {
		return fmt.Errorf("invalid permission name format: must be lowercase with underscores (e.g., manage_users)")
	}

	if len(permission.Name) < 3 {
		return fmt.Errorf("permission name must be at least 3 characters")
	}

	if len(permission.Name) > 100 {
		return fmt.Errorf("permission name must not exceed 100 characters")
	}

	// Validate description
	if len(permission.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}

	return nil
}

// isValidPermissionName checks if permission name follows the correct format
func isValidPermissionName(name string) bool {
	if name == "" {
		return false
	}

	// Must start with a lowercase letter
	if name[0] < 'a' || name[0] > 'z' {
		return false
	}

	// Can only contain lowercase letters, numbers, and underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	// Cannot have consecutive underscores
	if strings.Contains(name, "__") {
		return false
	}

	return true
}

// validateResource checks if a resource exists
func (s *Service) validateResource(ctx context.Context, resourceID string) error {
	// Note: This assumes we have access to resource repository
	// For now, we'll skip actual validation and log a warning
	s.logger.Debug("Resource validation skipped (implement when resource service is available)",
		zap.String("resource_id", resourceID))
	return nil
}

// validateAction checks if an action exists
func (s *Service) validateAction(ctx context.Context, actionID string) error {
	// Note: This assumes we have access to action repository
	// For now, we'll skip actual validation and log a warning
	s.logger.Debug("Action validation skipped (implement when action service is available)",
		zap.String("action_id", actionID))
	return nil
}

// invalidatePermissionRelatedCaches invalidates all caches related to a permission
func (s *Service) invalidatePermissionRelatedCaches(ctx context.Context, permission *models.Permission) {
	if s.cache == nil {
		return
	}

	// Invalidate permission-specific cache
	cacheKey := fmt.Sprintf("permission:%s", permission.ID)
	if err := s.cache.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate permission cache",
			zap.String("cache_key", cacheKey),
			zap.Error(err))
	}

	// Invalidate permission name cache
	nameCacheKey := fmt.Sprintf("permission:name:%s", permission.Name)
	if err := s.cache.Delete(nameCacheKey); err != nil {
		s.logger.Warn("Failed to invalidate permission name cache",
			zap.String("cache_key", nameCacheKey),
			zap.Error(err))
	}
}
