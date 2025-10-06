package resources

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"go.uber.org/zap"
)

// GetHierarchy retrieves the complete hierarchical tree for a resource
func (s *Service) GetHierarchy(ctx context.Context, rootID string) (*ResourceTree, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("resource:%s:hierarchy", rootID)
	if cached, found := s.cacheService.Get(cacheKey); found {
		if tree, ok := cached.(*ResourceTree); ok {
			s.logger.Debug("resource hierarchy retrieved from cache", zap.String("root_id", rootID))
			return tree, nil
		}
	}

	// Validate root exists
	exists, err := s.resourceRepo.Exists(ctx, rootID)
	if err != nil {
		return nil, fmt.Errorf("failed to check root resource existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("root resource not found: %s", rootID)
	}

	// Build hierarchy tree
	tree, err := s.buildResourceTree(ctx, rootID)
	if err != nil {
		s.logger.Error("failed to build resource hierarchy",
			zap.String("root_id", rootID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to build hierarchy: %w", err)
	}

	// Cache the result (TTL: 15 minutes)
	_ = s.cacheService.Set(cacheKey, tree, 900)

	s.logger.Debug("resource hierarchy built successfully",
		zap.String("root_id", rootID))

	return tree, nil
}

// GetAncestors retrieves all ancestor resources (parent, grandparent, etc.)
func (s *Service) GetAncestors(ctx context.Context, id string) ([]*models.Resource, error) {
	// Get the resource
	resource, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	var ancestors []*models.Resource

	// Traverse up the hierarchy
	currentResource := resource
	for currentResource.ParentID != nil {
		parent, err := s.GetByID(ctx, *currentResource.ParentID)
		if err != nil {
			s.logger.Error("failed to get parent resource",
				zap.String("resource_id", currentResource.ID),
				zap.String("parent_id", *currentResource.ParentID),
				zap.Error(err))
			break
		}

		ancestors = append(ancestors, parent)
		currentResource = parent

		// Safety check to prevent infinite loops
		if len(ancestors) > 100 {
			s.logger.Warn("ancestor chain too long, possible circular reference",
				zap.String("resource_id", id))
			break
		}
	}

	s.logger.Debug("ancestors retrieved",
		zap.String("resource_id", id),
		zap.Int("count", len(ancestors)))

	return ancestors, nil
}

// GetDescendants retrieves all descendant resources (children, grandchildren, etc.)
func (s *Service) GetDescendants(ctx context.Context, id string) ([]*models.Resource, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("resource:%s:descendants", id)
	if cached, found := s.cacheService.Get(cacheKey); found {
		if descendants, ok := cached.([]*models.Resource); ok {
			s.logger.Debug("descendants retrieved from cache", zap.String("resource_id", id))
			return descendants, nil
		}
	}

	// Validate resource exists
	exists, err := s.resourceRepo.Exists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check resource existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", id)
	}

	// Build descendants list
	var descendants []*models.Resource
	if err := s.collectDescendants(ctx, id, &descendants); err != nil {
		s.logger.Error("failed to collect descendants",
			zap.String("resource_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to collect descendants: %w", err)
	}

	// Cache the result (TTL: 10 minutes)
	_ = s.cacheService.Set(cacheKey, descendants, 600)

	s.logger.Debug("descendants retrieved",
		zap.String("resource_id", id),
		zap.Int("count", len(descendants)))

	return descendants, nil
}

// ValidateHierarchy validates that a resource can be placed in a hierarchy
func (s *Service) ValidateHierarchy(ctx context.Context, id, parentID string) error {
	if id == parentID {
		return fmt.Errorf("resource cannot be its own parent")
	}

	// Check if parent exists
	exists, err := s.resourceRepo.Exists(ctx, parentID)
	if err != nil {
		return fmt.Errorf("failed to check parent existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("parent resource not found: %s", parentID)
	}

	// Check for circular reference by getting ancestors of parent
	ancestors, err := s.GetAncestors(ctx, parentID)
	if err != nil {
		return fmt.Errorf("failed to get parent ancestors: %w", err)
	}

	// Check if the resource to be child is in the parent's ancestors
	for _, ancestor := range ancestors {
		if ancestor.ID == id {
			return fmt.Errorf("circular reference detected: resource %s is an ancestor of parent %s", id, parentID)
		}
	}

	return nil
}

// buildResourceTree recursively builds a hierarchical tree
func (s *Service) buildResourceTree(ctx context.Context, resourceID string) (*ResourceTree, error) {
	// Get the resource
	resource, err := s.resourceRepo.GetByID(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	// Create tree node
	tree := &ResourceTree{
		Resource: resource,
		Children: make([]*ResourceTree, 0),
	}

	// Get children
	children, err := s.resourceRepo.GetChildren(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}

	// Recursively build subtrees
	for _, child := range children {
		childTree, err := s.buildResourceTree(ctx, child.ID)
		if err != nil {
			s.logger.Warn("failed to build child tree, skipping",
				zap.String("parent_id", resourceID),
				zap.String("child_id", child.ID),
				zap.Error(err))
			continue
		}
		tree.Children = append(tree.Children, childTree)
	}

	return tree, nil
}

// collectDescendants recursively collects all descendants
func (s *Service) collectDescendants(ctx context.Context, resourceID string, descendants *[]*models.Resource) error {
	// Get direct children
	children, err := s.resourceRepo.GetChildren(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to get children: %w", err)
	}

	// Add children to descendants
	*descendants = append(*descendants, children...)

	// Recursively collect descendants of each child
	for _, child := range children {
		if err := s.collectDescendants(ctx, child.ID, descendants); err != nil {
			return err
		}
	}

	return nil
}

// getResourceDepth calculates the depth of a resource in the hierarchy
func (s *Service) getResourceDepth(ctx context.Context, resourceID string) (int, error) {
	ancestors, err := s.GetAncestors(ctx, resourceID)
	if err != nil {
		return 0, err
	}
	return len(ancestors), nil
}

// isDescendantOf checks if a resource is a descendant of another
func (s *Service) isDescendantOf(ctx context.Context, resourceID, ancestorID string) (bool, error) {
	ancestors, err := s.GetAncestors(ctx, resourceID)
	if err != nil {
		return false, err
	}

	for _, ancestor := range ancestors {
		if ancestor.ID == ancestorID {
			return true, nil
		}
	}

	return false, nil
}

// getRootResources retrieves all resources that have no parent (root resources)
func (s *Service) getRootResources(ctx context.Context, limit, offset int) ([]*models.Resource, error) {
	// This would require a repository method to filter by null parent_id
	// For now, we'll use the basic list and filter in memory (not optimal for large datasets)
	allResources, err := s.resourceRepo.List(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	var roots []*models.Resource
	for _, resource := range allResources {
		if resource.ParentID == nil {
			roots = append(roots, resource)
		}
	}

	// Apply pagination in memory
	start := offset
	end := offset + limit
	if start > len(roots) {
		return []*models.Resource{}, nil
	}
	if end > len(roots) {
		end = len(roots)
	}

	return roots[start:end], nil
}
