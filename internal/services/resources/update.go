package resources

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"go.uber.org/zap"
)

// Update updates an existing resource
func (s *Service) Update(ctx context.Context, resource *models.Resource) error {
	if resource == nil {
		return fmt.Errorf("resource cannot be nil")
	}

	// Validate resource exists
	exists, err := s.resourceRepo.Exists(ctx, resource.ID)
	if err != nil {
		s.logger.Error("failed to check resource existence", zap.String("resource_id", resource.ID), zap.Error(err))
		return fmt.Errorf("failed to check resource existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("resource not found: %s", resource.ID)
	}

	// Validate resource name and type
	if err := s.validateResourceName(resource.Name); err != nil {
		return fmt.Errorf("invalid resource name: %w", err)
	}
	if err := s.validateResourceType(resource.Type); err != nil {
		return fmt.Errorf("invalid resource type: %w", err)
	}

	// Update in database
	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		s.logger.Error("failed to update resource",
			zap.String("resource_id", resource.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update resource: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(resource.ID)
	if resource.ParentID != nil {
		s.invalidateHierarchyCache(*resource.ParentID)
	}

	s.logger.Info("resource updated successfully",
		zap.String("resource_id", resource.ID),
		zap.String("name", resource.Name))

	return nil
}

// UpdateName updates only the name of a resource
func (s *Service) UpdateName(ctx context.Context, id, name string) error {
	// Validate name
	if err := s.validateResourceName(name); err != nil {
		return fmt.Errorf("invalid resource name: %w", err)
	}

	// Get existing resource
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get resource for name update", zap.String("resource_id", id), zap.Error(err))
		return fmt.Errorf("resource not found: %w", err)
	}

	// Check if name is different
	if resource.Name == name {
		return nil // No change needed
	}

	// Check for duplicate name
	exists, err := s.checkResourceNameExists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check resource name existence: %w", err)
	}
	if exists {
		return fmt.Errorf("resource with name '%s' already exists", name)
	}

	// Update name
	resource.Name = name
	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		s.logger.Error("failed to update resource name",
			zap.String("resource_id", id),
			zap.String("new_name", name),
			zap.Error(err))
		return fmt.Errorf("failed to update resource name: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(id)

	s.logger.Info("resource name updated successfully",
		zap.String("resource_id", id),
		zap.String("new_name", name))

	return nil
}

// UpdateDescription updates only the description of a resource
func (s *Service) UpdateDescription(ctx context.Context, id, description string) error {
	// Get existing resource
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get resource for description update", zap.String("resource_id", id), zap.Error(err))
		return fmt.Errorf("resource not found: %w", err)
	}

	// Check if description is different
	if resource.Description == description {
		return nil // No change needed
	}

	// Update description
	resource.Description = description
	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		s.logger.Error("failed to update resource description",
			zap.String("resource_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to update resource description: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(id)

	s.logger.Info("resource description updated successfully",
		zap.String("resource_id", id))

	return nil
}

// Activate activates a resource
func (s *Service) Activate(ctx context.Context, id string) error {
	// Get existing resource
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get resource for activation", zap.String("resource_id", id), zap.Error(err))
		return fmt.Errorf("resource not found: %w", err)
	}

	// Check if already active
	if resource.IsActive {
		return nil // Already active
	}

	// Activate
	resource.IsActive = true
	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		s.logger.Error("failed to activate resource",
			zap.String("resource_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to activate resource: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(id)

	s.logger.Info("resource activated successfully", zap.String("resource_id", id))
	return nil
}

// Deactivate deactivates a resource
func (s *Service) Deactivate(ctx context.Context, id string) error {
	// Get existing resource
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get resource for deactivation", zap.String("resource_id", id), zap.Error(err))
		return fmt.Errorf("resource not found: %w", err)
	}

	// Check if already inactive
	if !resource.IsActive {
		return nil // Already inactive
	}

	// Deactivate
	resource.IsActive = false
	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		s.logger.Error("failed to deactivate resource",
			zap.String("resource_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to deactivate resource: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(id)

	s.logger.Info("resource deactivated successfully", zap.String("resource_id", id))
	return nil
}

// SetParent sets or updates the parent of a resource
func (s *Service) SetParent(ctx context.Context, id, parentID string) error {
	// Validate parent exists
	if parentID != "" {
		exists, err := s.resourceRepo.Exists(ctx, parentID)
		if err != nil {
			return fmt.Errorf("failed to check parent existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("parent resource not found: %s", parentID)
		}

		// Validate no circular reference
		if err := s.validateNoCircularReference(ctx, id, parentID); err != nil {
			return fmt.Errorf("circular reference detected: %w", err)
		}
	}

	// Get existing resource
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get resource for parent update", zap.String("resource_id", id), zap.Error(err))
		return fmt.Errorf("resource not found: %w", err)
	}

	// Store old parent ID for cache invalidation
	oldParentID := resource.ParentID

	// Update parent
	if parentID == "" {
		resource.ParentID = nil
	} else {
		resource.ParentID = &parentID
	}

	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		s.logger.Error("failed to update resource parent",
			zap.String("resource_id", id),
			zap.String("parent_id", parentID),
			zap.Error(err))
		return fmt.Errorf("failed to update resource parent: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(id)
	if oldParentID != nil {
		s.invalidateHierarchyCache(*oldParentID)
	}
	if parentID != "" {
		s.invalidateHierarchyCache(parentID)
	}

	s.logger.Info("resource parent updated successfully",
		zap.String("resource_id", id),
		zap.String("parent_id", parentID))

	return nil
}

// SetOwner sets or updates the owner of a resource
func (s *Service) SetOwner(ctx context.Context, id, ownerID string) error {
	// Get existing resource
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get resource for owner update", zap.String("resource_id", id), zap.Error(err))
		return fmt.Errorf("resource not found: %w", err)
	}

	// Update owner
	if ownerID == "" {
		resource.OwnerID = nil
	} else {
		resource.OwnerID = &ownerID
	}

	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		s.logger.Error("failed to update resource owner",
			zap.String("resource_id", id),
			zap.String("owner_id", ownerID),
			zap.Error(err))
		return fmt.Errorf("failed to update resource owner: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(id)

	s.logger.Info("resource owner updated successfully",
		zap.String("resource_id", id),
		zap.String("owner_id", ownerID))

	return nil
}

// invalidateResourceCacheForUpdate invalidates cache entries for a resource update
func (s *Service) invalidateResourceCacheForUpdate(resourceID string) {
	// Delete specific resource cache
	_ = s.cacheService.Delete(fmt.Sprintf("resource:%s", resourceID))

	// Invalidate list caches
	s.invalidateResourceCache()
}
