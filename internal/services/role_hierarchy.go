package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"go.uber.org/zap"
)

// GetRoleHierarchy retrieves all roles with their hierarchy structure
func (s *RoleService) GetRoleHierarchy(ctx context.Context) ([]*models.Role, error) {
	s.logger.Info("Getting role hierarchy")

	// Get all roles
	roles, err := s.roleRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all roles", zap.Error(err))
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	// Build hierarchy by finding root roles (no parent)
	var rootRoles []*models.Role
	roleMap := make(map[string]*models.Role)

	// First pass: create map and identify root roles
	for _, role := range roles {
		roleMap[role.ID] = role
		if role.ParentID == nil {
			rootRoles = append(rootRoles, role)
		}
	}

	// Second pass: attach children to their parents
	for _, role := range roles {
		if role.ParentID != nil {
			if parent, exists := roleMap[*role.ParentID]; exists {
				parent.Children = append(parent.Children, *role)
			}
		}
	}

	s.logger.Info("Role hierarchy retrieved successfully",
		zap.Int("total_roles", len(roles)),
		zap.Int("root_roles", len(rootRoles)))

	return rootRoles, nil
}

// GetRoleWithChildren retrieves a role with all its children loaded
func (s *RoleService) GetRoleWithChildren(ctx context.Context, roleID string) (*models.Role, error) {
	s.logger.Info("Getting role with children", zap.String("roleID", roleID))

	// Get the role
	var role models.Role
	_, err := s.roleRepo.GetByID(ctx, roleID, &role)
	if err != nil {
		s.logger.Error("Failed to get role", zap.Error(err))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Get all children recursively
	children, err := s.getRoleChildren(ctx, roleID)
	if err != nil {
		s.logger.Error("Failed to get role children", zap.Error(err))
		return nil, fmt.Errorf("failed to get children: %w", err)
	}

	role.Children = children

	s.logger.Info("Role with children retrieved successfully",
		zap.String("roleID", roleID),
		zap.Int("children_count", len(children)))

	return &role, nil
}

// getRoleChildren is a helper function to recursively get all children of a role
func (s *RoleService) getRoleChildren(ctx context.Context, roleID string) ([]models.Role, error) {
	// Get direct children
	directChildren, err := s.roleRepo.GetChildRoles(ctx, roleID)
	if err != nil {
		return nil, err
	}

	var allChildren []models.Role
	for _, child := range directChildren {
		// Add the child
		allChildren = append(allChildren, *child)

		// Recursively get grandchildren
		grandchildren, err := s.getRoleChildren(ctx, child.ID)
		if err != nil {
			return nil, err
		}

		// Update the child's Children field
		if len(grandchildren) > 0 {
			child.Children = grandchildren
			// Update in the allChildren slice
			allChildren[len(allChildren)-1].Children = grandchildren
		}
	}

	return allChildren, nil
}

// AddChildRole establishes a parent-child relationship between two roles
func (s *RoleService) AddChildRole(ctx context.Context, parentRoleID, childRoleID string) error {
	s.logger.Info("Adding child role",
		zap.String("parentRoleID", parentRoleID),
		zap.String("childRoleID", childRoleID))

	// Validate IDs
	if parentRoleID == "" || childRoleID == "" {
		return fmt.Errorf("parent and child role IDs are required")
	}

	if parentRoleID == childRoleID {
		return fmt.Errorf("a role cannot be its own parent")
	}

	// Get both roles
	var parentRole models.Role
	_, err := s.roleRepo.GetByID(ctx, parentRoleID, &parentRole)
	if err != nil {
		s.logger.Error("Failed to get parent role", zap.Error(err))
		return fmt.Errorf("parent role not found: %w", err)
	}

	var childRole models.Role
	_, err = s.roleRepo.GetByID(ctx, childRoleID, &childRole)
	if err != nil {
		s.logger.Error("Failed to get child role", zap.Error(err))
		return fmt.Errorf("child role not found: %w", err)
	}

	// Check if child already has a parent
	if childRole.ParentID != nil {
		return fmt.Errorf("child role already has a parent role")
	}

	// Check for circular dependency
	if err := s.checkCircularDependency(ctx, parentRoleID, childRoleID); err != nil {
		return err
	}

	// Update child role's parent
	childRole.ParentID = &parentRoleID
	if err := s.roleRepo.Update(ctx, &childRole); err != nil {
		s.logger.Error("Failed to update child role", zap.Error(err))
		return fmt.Errorf("failed to establish parent-child relationship: %w", err)
	}

	// Clear cache for both roles
	s.cacheService.Delete(fmt.Sprintf("role:%s", parentRoleID))
	s.cacheService.Delete(fmt.Sprintf("role:%s", childRoleID))

	s.logger.Info("Child role added successfully",
		zap.String("parentRole", parentRole.Name),
		zap.String("childRole", childRole.Name))

	return nil
}

// RemoveChildRole removes a parent-child relationship between two roles
func (s *RoleService) RemoveChildRole(ctx context.Context, parentRoleID, childRoleID string) error {
	s.logger.Info("Removing child role",
		zap.String("parentRoleID", parentRoleID),
		zap.String("childRoleID", childRoleID))

	// Get child role
	var childRole models.Role
	_, err := s.roleRepo.GetByID(ctx, childRoleID, &childRole)
	if err != nil {
		s.logger.Error("Failed to get child role", zap.Error(err))
		return fmt.Errorf("child role not found: %w", err)
	}

	// Verify it's actually a child of the specified parent
	if childRole.ParentID == nil || *childRole.ParentID != parentRoleID {
		return fmt.Errorf("role is not a child of the specified parent")
	}

	// Remove parent relationship
	childRole.ParentID = nil
	if err := s.roleRepo.Update(ctx, &childRole); err != nil {
		s.logger.Error("Failed to update child role", zap.Error(err))
		return fmt.Errorf("failed to remove parent-child relationship: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("role:%s", parentRoleID))
	s.cacheService.Delete(fmt.Sprintf("role:%s", childRoleID))

	s.logger.Info("Child role removed successfully")
	return nil
}

// checkCircularDependency checks if adding a parent-child relationship would create a circular dependency
func (s *RoleService) checkCircularDependency(ctx context.Context, parentRoleID, childRoleID string) error {
	// Start from the proposed parent and walk up the tree
	currentID := parentRoleID

	for currentID != "" {
		if currentID == childRoleID {
			return fmt.Errorf("circular dependency detected: cannot make a role's ancestor its child")
		}

		// Get the current role
		var role models.Role
		_, err := s.roleRepo.GetByID(ctx, currentID, &role)
		if err != nil {
			return fmt.Errorf("failed to check circular dependency: %w", err)
		}

		// Move up to the parent
		if role.ParentID == nil {
			break
		}
		currentID = *role.ParentID
	}

	return nil
}
