package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
)

// RoleService implements the RoleService interface
type RoleService struct {
	roleRepo     interfaces.RoleRepository
	cacheService interfaces.CacheService
	logger       interfaces.Logger
	validator    interfaces.Validator
}

// NewRoleService creates a new RoleService instance
func NewRoleService(
	roleRepo interfaces.RoleRepository,
	cacheService interfaces.CacheService,
	logger interfaces.Logger,
	validator interfaces.Validator,
) interfaces.RoleService {
	return &RoleService{
		roleRepo:     roleRepo,
		cacheService: cacheService,
		logger:       logger,
		validator:    validator,
	}
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("Creating new role")

	// Type assertion for request
	createReq, ok := req.(*models.Role)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Validate role
	if err := s.validateRole(createReq); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if role with same name already exists
	existingRole, err := s.roleRepo.GetByName(ctx, createReq.Name)
	if err == nil && existingRole != nil {
		s.logger.Error("Role with name already exists", zap.Error(err))
		return nil, fmt.Errorf("role with name '%s' already exists", createReq.Name)
	}

	// Create role in database
	if err := s.roleRepo.Create(ctx, createReq); err != nil {
		s.logger.Error("Failed to create role", zap.Error(err))
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("role:%s", createReq.ID))

	s.logger.Info("Role created successfully",
		zap.String("roleID", createReq.ID),
		zap.String("roleName", createReq.Name))
	return createReq, nil
}

// GetRoleByID retrieves a role by ID with caching
func (s *RoleService) GetRoleByID(ctx context.Context, roleID string) (interface{}, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("role:%s", roleID)
	if cached, exists := s.cacheService.Get(cacheKey); exists {
		if role, ok := cached.(*models.Role); ok {
			s.logger.Debug("Role retrieved from cache", zap.String("roleID", roleID))
			return role, nil
		}
	}

	// Get from database
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		s.logger.Error("Failed to get role by ID", zap.String("roleID", roleID), zap.Error(err))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Cache the result
	s.cacheService.Set(cacheKey, role, 300) // Cache for 5 minutes

	return role, nil
}

// GetRoleByName retrieves a role by name
func (s *RoleService) GetRoleByName(ctx context.Context, name string) (interface{}, error) {
	s.logger.Info("Getting role by name", zap.String("name", name))

	role, err := s.roleRepo.GetByName(ctx, name)
	if err != nil {
		s.logger.Error("Failed to get role by name", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// UpdateRole updates an existing role
func (s *RoleService) UpdateRole(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("Updating role")

	// Type assertion for request
	updateReq, ok := req.(*models.Role)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Validate role
	if err := s.validateRole(updateReq); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if role exists
	_, err := s.roleRepo.GetByID(ctx, updateReq.ID)
	if err != nil {
		s.logger.Error("Failed to find role for update", zap.String("roleID", updateReq.ID), zap.Error(err))
		return nil, errors.NewNotFoundError("role not found")
	}

	// Update role in database
	if err := s.roleRepo.Update(ctx, updateReq); err != nil {
		s.logger.Error("Failed to update role", zap.String("roleID", updateReq.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("role:%s", updateReq.ID))

	s.logger.Info("Role updated successfully", zap.String("roleID", updateReq.ID))
	return updateReq, nil
}

// DeleteRole soft deletes a role
func (s *RoleService) DeleteRole(ctx context.Context, roleID string) error {
	s.logger.Info("Deleting role", zap.String("roleID", roleID))

	// Delete role
	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		s.logger.Error("Failed to delete role", zap.String("roleID", roleID), zap.Error(err))
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("role:%s", roleID))

	s.logger.Info("Role deleted successfully", zap.String("roleID", roleID))
	return nil
}

// ListRoles lists roles with filters
func (s *RoleService) ListRoles(ctx context.Context, filters interface{}) (interface{}, error) {
	s.logger.Info("Listing roles")

	// Default pagination
	limit, offset := 10, 0

	// Extract limit and offset from filters if available
	if filterMap, ok := filters.(map[string]interface{}); ok {
		if l, exists := filterMap["limit"]; exists {
			if limitInt, ok := l.(int); ok {
				limit = limitInt
			}
		}
		if o, exists := filterMap["offset"]; exists {
			if offsetInt, ok := o.(int); ok {
				offset = offsetInt
			}
		}
	}

	roles, err := s.roleRepo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list roles", zap.Error(err))
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	s.logger.Info("Role listing completed", zap.Int("count", len(roles)))
	return roles, nil
}

// SearchRoles searches roles by keyword
func (s *RoleService) SearchRoles(ctx context.Context, keyword string, limit, offset int) (interface{}, error) {
	s.logger.Info("Searching roles", zap.String("keyword", keyword), zap.Int("limit", limit), zap.Int("offset", offset))

	if strings.TrimSpace(keyword) == "" {
		return nil, fmt.Errorf("search keyword cannot be empty")
	}

	roles, err := s.roleRepo.Search(ctx, keyword, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search roles", zap.Error(err))
		return nil, fmt.Errorf("failed to search roles: %w", err)
	}

	s.logger.Info("Role search completed", zap.Int("count", len(roles)))
	return roles, nil
}

// GetActiveRoles retrieves active roles
func (s *RoleService) GetActiveRoles(ctx context.Context, limit, offset int) (interface{}, error) {
	s.logger.Info("Getting active roles", zap.Int("limit", limit), zap.Int("offset", offset))

	roles, err := s.roleRepo.GetActive(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get active roles", zap.Error(err))
		return nil, fmt.Errorf("failed to get active roles: %w", err)
	}

	s.logger.Info("Active roles retrieved", zap.Int("count", len(roles)))
	return roles, nil
}

// AssignPermission assigns a permission to a role
func (s *RoleService) AssignPermission(ctx context.Context, roleID, permissionID string) (interface{}, error) {
	s.logger.Info("Assigning permission to role", zap.String("roleID", roleID), zap.String("permissionID", permissionID))

	// Placeholder implementation - this would involve updating the role's permissions
	result := map[string]interface{}{
		"roleID":       roleID,
		"permissionID": permissionID,
		"assigned":     true,
	}

	s.logger.Info("Permission assigned to role successfully")
	return result, nil
}

// RemovePermission removes a permission from a role
func (s *RoleService) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	s.logger.Info("Removing permission from role", zap.String("roleID", roleID), zap.String("permissionID", permissionID))

	// Placeholder implementation - this would involve updating the role's permissions
	s.logger.Info("Permission removed from role successfully")
	return nil
}

// GetRolePermissions retrieves permissions for a role
func (s *RoleService) GetRolePermissions(ctx context.Context, roleID string) (interface{}, error) {
	s.logger.Info("Getting role permissions", zap.String("roleID", roleID))

	// Placeholder implementation - return empty permissions
	permissions := []interface{}{}

	s.logger.Info("Role permissions retrieved", zap.Int("count", len(permissions)))
	return permissions, nil
}

// GetRoleHierarchy retrieves role hierarchy
func (s *RoleService) GetRoleHierarchy(ctx context.Context) (interface{}, error) {
	s.logger.Info("Getting role hierarchy")

	// Placeholder implementation
	hierarchy := map[string]interface{}{
		"hierarchy": []interface{}{},
	}

	s.logger.Info("Role hierarchy retrieved")
	return hierarchy, nil
}

// AddChildRole adds a child role to a parent role
func (s *RoleService) AddChildRole(ctx context.Context, parentRoleID, childRoleID string) (interface{}, error) {
	s.logger.Info("Adding child role", zap.String("parentRoleID", parentRoleID), zap.String("childRoleID", childRoleID))

	// Placeholder implementation
	result := map[string]interface{}{
		"parentRoleID": parentRoleID,
		"childRoleID":  childRoleID,
		"added":        true,
	}

	s.logger.Info("Child role added successfully")
	return result, nil
}

// ValidateRoleHierarchy validates role hierarchy
func (s *RoleService) ValidateRoleHierarchy(ctx context.Context, roleID string) error {
	s.logger.Info("Validating role hierarchy", zap.String("roleID", roleID))

	// Placeholder implementation
	s.logger.Info("Role hierarchy validation completed")
	return nil
}

// Helper methods

func (s *RoleService) validateRole(role *models.Role) error {
	if role == nil {
		return fmt.Errorf("role cannot be nil")
	}

	if strings.TrimSpace(role.Name) == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	if len(role.Name) < 2 {
		return fmt.Errorf("role name must be at least 2 characters long")
	}

	if len(role.Name) > 50 {
		return fmt.Errorf("role name cannot exceed 50 characters")
	}

	return nil
}
