package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
)

// RoleService implements the RoleService interface
type RoleService struct {
	roleRepo     interfaces.RoleRepository
	userRoleRepo interfaces.UserRoleRepository
	cacheService interfaces.CacheService
	logger       interfaces.Logger
	validator    interfaces.Validator
}

// NewRoleService creates a new RoleService instance
func NewRoleService(
	roleRepo interfaces.RoleRepository,
	userRoleRepo interfaces.UserRoleRepository,
	cacheService interfaces.CacheService,
	logger interfaces.Logger,
	validator interfaces.Validator,
) interfaces.RoleService {
	return &RoleService{
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
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

	// Check if role already exists
	exists, err := s.roleRepo.ExistsByName(ctx, createReq.Name)
	if err != nil {
		s.logger.Error("Failed to check role existence", "error", err)
		return nil, fmt.Errorf("failed to check role existence: %w", err)
	}

	if exists {
		return nil, errors.NewConflictError("role already exists with this name")
	}

	// Create role in database
	if err := s.roleRepo.Create(ctx, createReq); err != nil {
		s.logger.Error("Failed to create role", "error", err)
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("role:%s", createReq.ID))
	s.cacheService.Delete("roles:list")

	s.logger.Info("Role created successfully", "roleID", createReq.ID, "roleName", createReq.Name)
	return createReq, nil
}

// GetRoleByID retrieves a role by ID with caching
func (s *RoleService) GetRoleByID(ctx context.Context, roleID string) (interface{}, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("role:%s", roleID)
	if cached, exists := s.cacheService.Get(cacheKey); exists {
		if role, ok := cached.(*models.Role); ok {
			s.logger.Debug("Role retrieved from cache", "roleID", roleID)
			return role, nil
		}
	}

	// Get from database
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		s.logger.Error("Failed to get role by ID", "roleID", roleID, "error", err)
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Cache the result
	s.cacheService.Set(cacheKey, role, 300) // Cache for 5 minutes

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
	exists, err := s.roleRepo.Exists(ctx, updateReq.ID)
	if err != nil {
		s.logger.Error("Failed to check role existence", "roleID", updateReq.ID, "error", err)
		return nil, fmt.Errorf("failed to check role existence: %w", err)
	}

	if !exists {
		return nil, errors.NewNotFoundError("role not found")
	}

	// Check if name is being changed and if it conflicts with existing role
	if updateReq.Name != "" {
		existingRole, err := s.roleRepo.GetByName(ctx, updateReq.Name)
		if err == nil && existingRole.ID != updateReq.ID {
			return nil, errors.NewConflictError("role name already exists")
		}
	}

	// Update role in database
	if err := s.roleRepo.Update(ctx, updateReq); err != nil {
		s.logger.Error("Failed to update role", "roleID", updateReq.ID, "error", err)
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("role:%s", updateReq.ID))
	s.cacheService.Delete("roles:list")

	s.logger.Info("Role updated successfully", "roleID", updateReq.ID)
	return updateReq, nil
}

// DeleteRole soft deletes a role
func (s *RoleService) DeleteRole(ctx context.Context, roleID string) (interface{}, error) {
	s.logger.Info("Deleting role", "roleID", roleID)

	// Check if role exists
	exists, err := s.roleRepo.Exists(ctx, roleID)
	if err != nil {
		s.logger.Error("Failed to check role existence", "roleID", roleID, "error", err)
		return nil, fmt.Errorf("failed to check role existence: %w", err)
	}

	if !exists {
		return nil, errors.NewNotFoundError("role not found")
	}

	// Check if role is assigned to any users
	userRoles, err := s.userRoleRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		s.logger.Error("Failed to check role assignments", "error", err)
		return nil, fmt.Errorf("failed to check role assignments: %w", err)
	}

	if len(userRoles) > 0 {
		return nil, errors.NewConflictError("cannot delete role that is assigned to users")
	}

	// Delete role
	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		s.logger.Error("Failed to delete role", "roleID", roleID, "error", err)
		return nil, fmt.Errorf("failed to delete role: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("role:%s", roleID))
	s.cacheService.Delete("roles:list")

	s.logger.Info("Role deleted successfully", "roleID", roleID)
	return map[string]string{"message": "Role deleted successfully"}, nil
}

// ListRoles retrieves a list of roles with pagination
func (s *RoleService) ListRoles(ctx context.Context, limit, offset int) (interface{}, error) {
	s.logger.Info("Listing roles", "limit", limit, "offset", offset)

	// Try to get from cache first
	cacheKey := fmt.Sprintf("roles:list:%d:%d", limit, offset)
	if cached, exists := s.cacheService.Get(cacheKey); exists {
		if roles, ok := cached.([]models.Role); ok {
			s.logger.Debug("Roles retrieved from cache", "count", len(roles))
			return roles, nil
		}
	}

	// Get from database
	roles, err := s.roleRepo.List(ctx, nil, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list roles", "error", err)
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	// Cache the result
	s.cacheService.Set(cacheKey, roles, 300) // Cache for 5 minutes

	s.logger.Info("Roles listed successfully", "count", len(roles))
	return roles, nil
}

// Helper methods

func (s *RoleService) validateRole(role *models.Role) error {
	if role == nil {
		return fmt.Errorf("role cannot be nil")
	}

	if strings.TrimSpace(role.Name) == "" {
		return fmt.Errorf("role name is required")
	}

	if len(role.Name) > 100 {
		return fmt.Errorf("role name cannot exceed 100 characters")
	}

	if role.Description != nil && len(*role.Description) > 500 {
		return fmt.Errorf("role description cannot exceed 500 characters")
	}

	return nil
}
