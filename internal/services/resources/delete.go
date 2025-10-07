package resources

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Delete performs a hard delete of a resource
func (s *Service) Delete(ctx context.Context, id string) error {
	// Validate resource exists
	exists, err := s.resourceRepo.Exists(ctx, id)
	if err != nil {
		s.logger.Error("failed to check resource existence", zap.String("resource_id", id), zap.Error(err))
		return fmt.Errorf("failed to check resource existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("resource not found: %s", id)
	}

	// Check if resource has children
	hasChildren, err := s.HasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check for children: %w", err)
	}
	if hasChildren {
		return fmt.Errorf("cannot delete resource with children, use DeleteCascade or remove children first")
	}

	// Delete from database
	if err := s.resourceRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete resource",
			zap.String("resource_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(id)

	s.logger.Info("resource deleted successfully", zap.String("resource_id", id))
	return nil
}

// SoftDelete performs a soft delete of a resource
func (s *Service) SoftDelete(ctx context.Context, id, deletedBy string) error {
	// Validate resource exists
	exists, err := s.resourceRepo.Exists(ctx, id)
	if err != nil {
		s.logger.Error("failed to check resource existence", zap.String("resource_id", id), zap.Error(err))
		return fmt.Errorf("failed to check resource existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("resource not found: %s", id)
	}

	// Check if resource has children
	hasChildren, err := s.HasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check for children: %w", err)
	}
	if hasChildren {
		return fmt.Errorf("cannot soft delete resource with children, use DeleteCascade or remove children first")
	}

	// Soft delete from database
	if err := s.resourceRepo.SoftDelete(ctx, id, deletedBy); err != nil {
		s.logger.Error("failed to soft delete resource",
			zap.String("resource_id", id),
			zap.String("deleted_by", deletedBy),
			zap.Error(err))
		return fmt.Errorf("failed to soft delete resource: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(id)

	s.logger.Info("resource soft deleted successfully",
		zap.String("resource_id", id),
		zap.String("deleted_by", deletedBy))

	return nil
}

// DeleteCascade deletes a resource and all its descendants
func (s *Service) DeleteCascade(ctx context.Context, id string) error {
	// Validate resource exists
	exists, err := s.resourceRepo.Exists(ctx, id)
	if err != nil {
		s.logger.Error("failed to check resource existence", zap.String("resource_id", id), zap.Error(err))
		return fmt.Errorf("failed to check resource existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("resource not found: %s", id)
	}

	// Get all descendants
	descendants, err := s.GetDescendants(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get descendants: %w", err)
	}

	// Delete descendants first (bottom-up)
	for i := len(descendants) - 1; i >= 0; i-- {
		descendant := descendants[i]
		if err := s.resourceRepo.Delete(ctx, descendant.ID); err != nil {
			s.logger.Error("failed to delete descendant",
				zap.String("descendant_id", descendant.ID),
				zap.Error(err))
			return fmt.Errorf("failed to delete descendant '%s': %w", descendant.ID, err)
		}
	}

	// Finally delete the root resource
	if err := s.resourceRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete root resource",
			zap.String("resource_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete root resource: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(id)
	for _, descendant := range descendants {
		s.invalidateResourceCacheForUpdate(descendant.ID)
	}

	s.logger.Info("resource cascade deleted successfully",
		zap.String("resource_id", id),
		zap.Int("descendants_deleted", len(descendants)))

	return nil
}

// Restore restores a soft-deleted resource
func (s *Service) Restore(ctx context.Context, id string) error {
	// Check if resource exists (including soft-deleted)
	exists, err := s.resourceRepo.ExistsWithDeleted(ctx, id)
	if err != nil {
		s.logger.Error("failed to check resource existence", zap.String("resource_id", id), zap.Error(err))
		return fmt.Errorf("failed to check resource existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("resource not found: %s", id)
	}

	// Restore from database
	if err := s.resourceRepo.Restore(ctx, id); err != nil {
		s.logger.Error("failed to restore resource",
			zap.String("resource_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to restore resource: %w", err)
	}

	// Invalidate cache
	s.invalidateResourceCacheForUpdate(id)

	s.logger.Info("resource restored successfully", zap.String("resource_id", id))
	return nil
}
