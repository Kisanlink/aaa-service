package resources

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"go.uber.org/zap"
)

var (
	// Resource type pattern: lowercase alphanumeric with forward slashes
	resourceTypePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(/[a-z][a-z0-9]*)*$`)
)

// Create creates a new resource with validation
func (s *Service) Create(ctx context.Context, name, resourceType, description string) (*models.Resource, error) {
	// Validate resource name
	if err := s.validateResourceName(name); err != nil {
		s.logger.Error("invalid resource name", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("invalid resource name: %w", err)
	}

	// Validate resource type
	if err := s.validateResourceType(resourceType); err != nil {
		s.logger.Error("invalid resource type", zap.String("type", resourceType), zap.Error(err))
		return nil, fmt.Errorf("invalid resource type: %w", err)
	}

	// Check for duplicate name
	exists, err := s.checkResourceNameExists(ctx, name)
	if err != nil {
		s.logger.Error("failed to check resource name existence", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("failed to check resource name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("resource with name '%s' already exists", name)
	}

	// Create resource
	resource := models.NewResource(name, resourceType, description)

	if err := s.resourceRepo.Create(ctx, resource); err != nil {
		s.logger.Error("failed to create resource",
			zap.String("name", name),
			zap.String("type", resourceType),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCache()

	s.logger.Info("resource created successfully",
		zap.String("resource_id", resource.ID),
		zap.String("name", name),
		zap.String("type", resourceType))

	return resource, nil
}

// CreateWithParent creates a new resource with a parent resource
func (s *Service) CreateWithParent(ctx context.Context, name, resourceType, description, parentID string) (*models.Resource, error) {
	// Validate parent exists
	parentExists, err := s.resourceRepo.Exists(ctx, parentID)
	if err != nil {
		s.logger.Error("failed to check parent existence", zap.String("parent_id", parentID), zap.Error(err))
		return nil, fmt.Errorf("failed to check parent existence: %w", err)
	}
	if !parentExists {
		return nil, fmt.Errorf("parent resource not found: %s", parentID)
	}

	// Validate resource name and type
	if err := s.validateResourceName(name); err != nil {
		return nil, fmt.Errorf("invalid resource name: %w", err)
	}
	if err := s.validateResourceType(resourceType); err != nil {
		return nil, fmt.Errorf("invalid resource type: %w", err)
	}

	// Check for duplicate name
	exists, err := s.checkResourceNameExists(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check resource name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("resource with name '%s' already exists", name)
	}

	// Check for circular reference
	if err := s.validateNoCircularReference(ctx, name, parentID); err != nil {
		return nil, fmt.Errorf("circular reference detected: %w", err)
	}

	// Create resource with parent
	resource := models.NewResourceWithParent(name, resourceType, description, parentID)

	if err := s.resourceRepo.Create(ctx, resource); err != nil {
		s.logger.Error("failed to create resource with parent",
			zap.String("name", name),
			zap.String("parent_id", parentID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCache()
	s.invalidateHierarchyCache(parentID)

	s.logger.Info("resource with parent created successfully",
		zap.String("resource_id", resource.ID),
		zap.String("name", name),
		zap.String("parent_id", parentID))

	return resource, nil
}

// CreateWithOwner creates a new resource with an owner
func (s *Service) CreateWithOwner(ctx context.Context, name, resourceType, description, ownerID string) (*models.Resource, error) {
	// Validate resource name and type
	if err := s.validateResourceName(name); err != nil {
		return nil, fmt.Errorf("invalid resource name: %w", err)
	}
	if err := s.validateResourceType(resourceType); err != nil {
		return nil, fmt.Errorf("invalid resource type: %w", err)
	}

	// Check for duplicate name
	exists, err := s.checkResourceNameExists(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check resource name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("resource with name '%s' already exists", name)
	}

	// Create resource with owner
	resource := models.NewResourceWithOwner(name, resourceType, description, ownerID)

	if err := s.resourceRepo.Create(ctx, resource); err != nil {
		s.logger.Error("failed to create resource with owner",
			zap.String("name", name),
			zap.String("owner_id", ownerID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCache()

	s.logger.Info("resource with owner created successfully",
		zap.String("resource_id", resource.ID),
		zap.String("name", name),
		zap.String("owner_id", ownerID))

	return resource, nil
}

// CreateBatch creates multiple resources sequentially
func (s *Service) CreateBatch(ctx context.Context, resources []*models.Resource) error {
	if len(resources) == 0 {
		return fmt.Errorf("no resources to create")
	}

	// Validate all resources before creating
	for _, resource := range resources {
		if err := s.validateResourceName(resource.Name); err != nil {
			return fmt.Errorf("invalid resource name '%s': %w", resource.Name, err)
		}
		if err := s.validateResourceType(resource.Type); err != nil {
			return fmt.Errorf("invalid resource type '%s': %w", resource.Type, err)
		}
	}

	// Create all resources
	for _, resource := range resources {
		if err := s.resourceRepo.Create(ctx, resource); err != nil {
			s.logger.Error("failed to create resource in batch",
				zap.String("name", resource.Name),
				zap.Error(err))
			return fmt.Errorf("failed to create resource '%s': %w", resource.Name, err)
		}
	}

	// Invalidate cache
	s.invalidateResourceCache()

	s.logger.Info("batch resources created successfully", zap.Int("count", len(resources)))
	return nil
}

// validateResourceName validates resource name format
func (s *Service) validateResourceName(name string) error {
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("resource name too long (max 100 characters)")
	}
	if strings.TrimSpace(name) != name {
		return fmt.Errorf("resource name cannot have leading or trailing whitespace")
	}
	return nil
}

// validateResourceType validates resource type format
func (s *Service) validateResourceType(resourceType string) error {
	if resourceType == "" {
		return fmt.Errorf("resource type cannot be empty")
	}
	if !resourceTypePattern.MatchString(resourceType) {
		return fmt.Errorf("resource type must match pattern '^[a-z][a-z0-9]*(/[a-z][a-z0-9]*)*$'")
	}
	if len(resourceType) > 100 {
		return fmt.Errorf("resource type too long (max 100 characters)")
	}
	return nil
}

// checkResourceNameExists checks if a resource with the given name already exists
func (s *Service) checkResourceNameExists(ctx context.Context, name string) (bool, error) {
	_, err := s.resourceRepo.GetByName(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// validateNoCircularReference validates that setting a parent doesn't create a circular reference
func (s *Service) validateNoCircularReference(ctx context.Context, resourceID, parentID string) error {
	if resourceID == parentID {
		return fmt.Errorf("resource cannot be its own parent")
	}
	// Additional circular reference validation would require traversing the hierarchy
	// This is a simplified version
	return nil
}

// invalidateResourceCache invalidates resource-related cache entries
func (s *Service) invalidateResourceCache() {
	// Clear resource list cache
	_ = s.cacheService.Delete("resources:list:*")
	_ = s.cacheService.Delete("resources:count")
}

// invalidateHierarchyCache invalidates hierarchy cache for a resource
func (s *Service) invalidateHierarchyCache(resourceID string) {
	cacheKey := fmt.Sprintf("resource:%s:hierarchy", resourceID)
	_ = s.cacheService.Delete(cacheKey)
}
