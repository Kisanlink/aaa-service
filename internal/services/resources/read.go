package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"go.uber.org/zap"
)

// GetByID retrieves a resource by ID with caching
func (s *Service) GetByID(ctx context.Context, id string) (*models.Resource, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("resource:%s", id)
	if cached, found := s.cacheService.Get(cacheKey); found {
		if resource, ok := cached.(*models.Resource); ok {
			s.logger.Debug("resource retrieved from cache", zap.String("resource_id", id))
			return resource, nil
		}
	}

	// Fetch from database
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get resource by ID", zap.String("resource_id", id), zap.Error(err))
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	// Cache the result (TTL: 15 minutes)
	_ = s.cacheService.Set(cacheKey, resource, 900)

	s.logger.Debug("resource retrieved from database", zap.String("resource_id", id))
	return resource, nil
}

// GetByName retrieves a resource by name with caching
func (s *Service) GetByName(ctx context.Context, name string) (*models.Resource, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("resource:name:%s", name)
	if cached, found := s.cacheService.Get(cacheKey); found {
		if resource, ok := cached.(*models.Resource); ok {
			s.logger.Debug("resource retrieved from cache by name", zap.String("name", name))
			return resource, nil
		}
	}

	// Fetch from database
	resource, err := s.resourceRepo.GetByName(ctx, name)
	if err != nil {
		s.logger.Error("failed to get resource by name", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	// Cache the result (TTL: 15 minutes)
	_ = s.cacheService.Set(cacheKey, resource, 900)
	// Also cache by ID
	idCacheKey := fmt.Sprintf("resource:%s", resource.ID)
	_ = s.cacheService.Set(idCacheKey, resource, 900)

	s.logger.Debug("resource retrieved from database by name", zap.String("name", name))
	return resource, nil
}

// GetByType retrieves resources by type with pagination
func (s *Service) GetByType(ctx context.Context, resourceType string, limit, offset int) ([]*models.Resource, error) {
	// Validate pagination
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if limit > 100 {
		limit = 100
	}

	// Try cache first
	cacheKey := fmt.Sprintf("resources:type:%s:limit:%d:offset:%d", resourceType, limit, offset)
	if cached, found := s.cacheService.Get(cacheKey); found {
		if resources, ok := cached.([]*models.Resource); ok {
			s.logger.Debug("resources retrieved from cache by type",
				zap.String("type", resourceType),
				zap.Int("limit", limit),
				zap.Int("offset", offset))
			return resources, nil
		}
	}

	// Fetch from database
	resources, err := s.resourceRepo.GetByType(ctx, resourceType, limit, offset)
	if err != nil {
		s.logger.Error("failed to get resources by type",
			zap.String("type", resourceType),
			zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve resources: %w", err)
	}

	// Cache the result (TTL: 10 minutes)
	_ = s.cacheService.Set(cacheKey, resources, 600)

	s.logger.Debug("resources retrieved from database by type",
		zap.String("type", resourceType),
		zap.Int("count", len(resources)))

	return resources, nil
}

// List retrieves all resources with pagination
func (s *Service) List(ctx context.Context, limit, offset int) ([]*models.Resource, error) {
	// Validate pagination
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if limit > 100 {
		limit = 100
	}

	// Try cache first
	cacheKey := fmt.Sprintf("resources:list:limit:%d:offset:%d", limit, offset)
	if cached, found := s.cacheService.Get(cacheKey); found {
		if resources, ok := cached.([]*models.Resource); ok {
			s.logger.Debug("resources list retrieved from cache",
				zap.Int("limit", limit),
				zap.Int("offset", offset))
			return resources, nil
		}
	}

	// Fetch from database
	resources, err := s.resourceRepo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("failed to list resources", zap.Error(err))
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	// Cache the result (TTL: 5 minutes)
	_ = s.cacheService.Set(cacheKey, resources, 300)

	s.logger.Debug("resources list retrieved from database",
		zap.Int("count", len(resources)),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	return resources, nil
}

// Count returns the total number of resources
func (s *Service) Count(ctx context.Context) (int64, error) {
	// Try cache first
	cacheKey := "resources:count"
	if cached, found := s.cacheService.Get(cacheKey); found {
		if count, ok := cached.(int64); ok {
			s.logger.Debug("resource count retrieved from cache", zap.Int64("count", count))
			return count, nil
		}
	}

	// Fetch from database
	count, err := s.resourceRepo.Count(ctx)
	if err != nil {
		s.logger.Error("failed to count resources", zap.Error(err))
		return 0, fmt.Errorf("failed to count resources: %w", err)
	}

	// Cache the result (TTL: 5 minutes)
	_ = s.cacheService.Set(cacheKey, count, 300)

	s.logger.Debug("resource count retrieved from database", zap.Int64("count", count))
	return count, nil
}

// Exists checks if a resource exists by ID
func (s *Service) Exists(ctx context.Context, id string) (bool, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("resource:%s", id)
	if _, found := s.cacheService.Get(cacheKey); found {
		return true, nil
	}

	// Check database
	exists, err := s.resourceRepo.Exists(ctx, id)
	if err != nil {
		s.logger.Error("failed to check resource existence", zap.String("resource_id", id), zap.Error(err))
		return false, fmt.Errorf("failed to check resource existence: %w", err)
	}

	// If exists, try to cache it for future lookups
	if exists {
		resource, err := s.resourceRepo.GetByID(ctx, id)
		if err == nil {
			_ = s.cacheService.Set(cacheKey, resource, 900)
		}
	}

	return exists, nil
}

// GetChildren retrieves all child resources for a given parent
func (s *Service) GetChildren(ctx context.Context, parentID string) ([]*models.Resource, error) {
	// Validate parent exists
	exists, err := s.Exists(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check parent existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("parent resource not found: %s", parentID)
	}

	// Try cache first
	cacheKey := fmt.Sprintf("resource:%s:children", parentID)
	if cached, found := s.cacheService.Get(cacheKey); found {
		if children, ok := cached.([]*models.Resource); ok {
			s.logger.Debug("children retrieved from cache", zap.String("parent_id", parentID))
			return children, nil
		}
	}

	// Fetch from database
	children, err := s.resourceRepo.GetChildren(ctx, parentID)
	if err != nil {
		s.logger.Error("failed to get children",
			zap.String("parent_id", parentID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get children: %w", err)
	}

	// Cache the result (TTL: 10 minutes)
	_ = s.cacheService.Set(cacheKey, children, 600)

	s.logger.Debug("children retrieved from database",
		zap.String("parent_id", parentID),
		zap.Int("count", len(children)))

	return children, nil
}

// GetParent retrieves the parent resource
func (s *Service) GetParent(ctx context.Context, id string) (*models.Resource, error) {
	// Get the resource first
	resource, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	// Check if it has a parent
	if resource.ParentID == nil {
		return nil, fmt.Errorf("resource has no parent")
	}

	// Get the parent
	parent, err := s.GetByID(ctx, *resource.ParentID)
	if err != nil {
		s.logger.Error("failed to get parent resource",
			zap.String("resource_id", id),
			zap.String("parent_id", *resource.ParentID),
			zap.Error(err))
		return nil, fmt.Errorf("parent resource not found: %w", err)
	}

	return parent, nil
}

// HasChildren checks if a resource has any children
func (s *Service) HasChildren(ctx context.Context, id string) (bool, error) {
	count, err := s.resourceRepo.CountChildren(ctx, id)
	if err != nil {
		s.logger.Error("failed to count children",
			zap.String("resource_id", id),
			zap.Error(err))
		return false, fmt.Errorf("failed to count children: %w", err)
	}

	return count > 0, nil
}

// getCachedResourceOrFetch retrieves a resource from cache or database
func (s *Service) getCachedResourceOrFetch(ctx context.Context, id string) (*models.Resource, error) {
	cacheKey := fmt.Sprintf("resource:%s", id)
	if cached, found := s.cacheService.Get(cacheKey); found {
		if resource, ok := cached.(*models.Resource); ok {
			return resource, nil
		}
	}

	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cacheService.Set(cacheKey, resource, int(15*time.Minute.Seconds()))
	return resource, nil
}
