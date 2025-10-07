package permissions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// InvalidateUserCache invalidates all cached data for a user
func (s *Service) InvalidateUserCache(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	if s.cache == nil {
		return nil
	}

	// Pattern to match all user-related cache keys
	patterns := []string{
		fmt.Sprintf("user:%s:*", userID),
		fmt.Sprintf("permission:%s:*", userID),
	}

	for _, pattern := range patterns {
		keys, err := s.cache.Keys(pattern)
		if err != nil {
			s.logger.Warn("Failed to get cache keys",
				zap.String("pattern", pattern),
				zap.Error(err))
			continue
		}

		for _, key := range keys {
			if err := s.cache.Delete(key); err != nil {
				s.logger.Warn("Failed to delete cache key",
					zap.String("key", key),
					zap.Error(err))
			}
		}
	}

	s.logger.Info("User cache invalidated",
		zap.String("user_id", userID))

	return nil
}

// InvalidateRoleCache invalidates all cached data for a role
func (s *Service) InvalidateRoleCache(ctx context.Context, roleID string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if s.cache == nil {
		return nil
	}

	// Pattern to match all role-related cache keys
	patterns := []string{
		fmt.Sprintf("role:%s:*", roleID),
	}

	for _, pattern := range patterns {
		keys, err := s.cache.Keys(pattern)
		if err != nil {
			s.logger.Warn("Failed to get cache keys",
				zap.String("pattern", pattern),
				zap.Error(err))
			continue
		}

		for _, key := range keys {
			if err := s.cache.Delete(key); err != nil {
				s.logger.Warn("Failed to delete cache key",
					zap.String("key", key),
					zap.Error(err))
			}
		}
	}

	s.logger.Info("Role cache invalidated",
		zap.String("role_id", roleID))

	return nil
}

// InvalidatePermissionCache invalidates all cached data for a permission
func (s *Service) InvalidatePermissionCache(ctx context.Context, permissionID string) error {
	if permissionID == "" {
		return fmt.Errorf("permission ID is required")
	}

	if s.cache == nil {
		return nil
	}

	// Direct permission cache keys
	cacheKeys := []string{
		fmt.Sprintf("permission:%s", permissionID),
	}

	for _, key := range cacheKeys {
		if err := s.cache.Delete(key); err != nil {
			s.logger.Warn("Failed to delete cache key",
				zap.String("key", key),
				zap.Error(err))
		}
	}

	// Get permission to find its name
	permission, err := s.permissionRepo.GetByID(ctx, permissionID)
	if err == nil && permission != nil {
		nameKey := fmt.Sprintf("permission:name:%s", permission.Name)
		if err := s.cache.Delete(nameKey); err != nil {
			s.logger.Warn("Failed to delete permission name cache",
				zap.String("key", nameKey),
				zap.Error(err))
		}
	}

	// Invalidate all role caches that use this permission
	s.invalidateAllRoleCaches(ctx, permissionID)

	s.logger.Info("Permission cache invalidated",
		zap.String("permission_id", permissionID))

	return nil
}

// InvalidateAllPermissionCaches invalidates all permission-related caches
func (s *Service) InvalidateAllPermissionCaches(ctx context.Context) error {
	if s.cache == nil {
		return nil
	}

	patterns := []string{
		"permission:*",
		"role:*:permissions",
		"user:*:effective_roles",
	}

	totalDeleted := 0
	for _, pattern := range patterns {
		keys, err := s.cache.Keys(pattern)
		if err != nil {
			s.logger.Warn("Failed to get cache keys",
				zap.String("pattern", pattern),
				zap.Error(err))
			continue
		}

		for _, key := range keys {
			if err := s.cache.Delete(key); err != nil {
				s.logger.Warn("Failed to delete cache key",
					zap.String("key", key),
					zap.Error(err))
			} else {
				totalDeleted++
			}
		}
	}

	s.logger.Info("All permission caches invalidated",
		zap.Int("total_deleted", totalDeleted))

	return nil
}

// InvalidateCacheByPattern invalidates all caches matching a pattern
func (s *Service) InvalidateCacheByPattern(ctx context.Context, pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern is required")
	}

	if s.cache == nil {
		return nil
	}

	keys, err := s.cache.Keys(pattern)
	if err != nil {
		s.logger.Warn("Failed to get cache keys",
			zap.String("pattern", pattern),
			zap.Error(err))
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	deletedCount := 0
	for _, key := range keys {
		if err := s.cache.Delete(key); err != nil {
			s.logger.Warn("Failed to delete cache key",
				zap.String("key", key),
				zap.Error(err))
		} else {
			deletedCount++
		}
	}

	s.logger.Info("Cache invalidated by pattern",
		zap.String("pattern", pattern),
		zap.Int("deleted_count", deletedCount))

	return nil
}

// WarmupCache pre-loads frequently accessed data into cache
func (s *Service) WarmupCache(ctx context.Context) error {
	if s.cache == nil {
		return nil
	}

	s.logger.Info("Starting cache warmup")

	// TODO: Implement cache warmup strategy
	// - Load all active permissions
	// - Load role-permission mappings for frequently used roles
	// - Pre-calculate common permission evaluations

	s.logger.Info("Cache warmup completed")

	return nil
}

// GetCacheStats returns statistics about cached data
func (s *Service) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	if s.cache == nil {
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	stats := map[string]interface{}{
		"enabled": true,
	}

	// Count permission caches
	permKeys, err := s.cache.Keys("permission:*")
	if err == nil {
		stats["permission_count"] = len(permKeys)
	}

	// Count role caches
	roleKeys, err := s.cache.Keys("role:*")
	if err == nil {
		stats["role_count"] = len(roleKeys)
	}

	// Count user caches
	userKeys, err := s.cache.Keys("user:*")
	if err == nil {
		stats["user_count"] = len(userKeys)
	}

	return stats, nil
}
