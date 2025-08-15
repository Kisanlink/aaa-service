package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
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
func (s *RoleService) CreateRole(ctx context.Context, role *models.Role) error {
	s.logger.Info("Creating new role")

	// Validate role
	if err := s.validateRole(role); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if role with same name already exists
	existingRole, err := s.roleRepo.GetByName(ctx, role.Name)
	if err == nil && existingRole != nil {
		s.logger.Error("Role with name already exists", zap.Error(err))
		return fmt.Errorf("role with name '%s' already exists", role.Name)
	}

	// Create role in database
	if err := s.roleRepo.Create(ctx, role); err != nil {
		s.logger.Error("Failed to create role", zap.Error(err))
		return fmt.Errorf("failed to create role: %w", err)
	}

	// Clear cache
	if err := s.cacheService.Delete(fmt.Sprintf("role:%s", role.ID)); err != nil {
		s.logger.Error("Failed to delete role from cache", zap.Error(err))
	}

	s.logger.Info("Role created successfully",
		zap.String("roleID", role.ID),
		zap.String("roleName", role.Name))
	return nil
}

// GetRoleByID retrieves a role by ID with caching
func (s *RoleService) GetRoleByID(ctx context.Context, roleID string) (*models.Role, error) {
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
	if err := s.cacheService.Set(cacheKey, role, 300); err != nil {
		s.logger.Error("Failed to cache role", zap.Error(err))
	} // Cache for 5 minutes

	return role, nil
}

// GetRoleByName retrieves a role by name
func (s *RoleService) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	s.logger.Info("Getting role by name", zap.String("name", name))

	role, err := s.roleRepo.GetByName(ctx, name)
	if err != nil {
		s.logger.Error("Failed to get role by name", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// UpdateRole updates an existing role
func (s *RoleService) UpdateRole(ctx context.Context, role *models.Role) error {
	s.logger.Info("Updating role")

	// Validate role
	if err := s.validateRole(role); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if role exists
	_, err := s.roleRepo.GetByID(ctx, role.ID)
	if err != nil {
		s.logger.Error("Failed to find role for update", zap.String("roleID", role.ID), zap.Error(err))
		return errors.NewNotFoundError("role not found")
	}

	// Update role in database
	if err := s.roleRepo.Update(ctx, role); err != nil {
		s.logger.Error("Failed to update role", zap.String("roleID", role.ID), zap.Error(err))
		return fmt.Errorf("failed to update role: %w", err)
	}

	// Clear cache
	if err := s.cacheService.Delete(fmt.Sprintf("role:%s", role.ID)); err != nil {
		s.logger.Error("Failed to delete role from cache", zap.Error(err))
	}

	s.logger.Info("Role updated successfully", zap.String("roleID", role.ID))
	return nil
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
	if err := s.cacheService.Delete(fmt.Sprintf("role:%s", roleID)); err != nil {
		s.logger.Error("Failed to delete role from cache", zap.Error(err))
	}

	s.logger.Info("Role deleted successfully", zap.String("roleID", roleID))
	return nil
}

// ListRoles lists roles with pagination
func (s *RoleService) ListRoles(ctx context.Context, limit, offset int) ([]*models.Role, error) {
	s.logger.Info("Listing roles", zap.Int("limit", limit), zap.Int("offset", offset))

	roles, err := s.roleRepo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list roles", zap.Error(err))
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	s.logger.Info("Role listing completed", zap.Int("count", len(roles)))
	return roles, nil
}

// SearchRoles searches roles by keyword
func (s *RoleService) SearchRoles(ctx context.Context, query string, limit, offset int) ([]*models.Role, error) {
	s.logger.Info("Searching roles", zap.String("query", query), zap.Int("limit", limit), zap.Int("offset", offset))

	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	roles, err := s.roleRepo.Search(ctx, query, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search roles", zap.Error(err))
		return nil, fmt.Errorf("failed to search roles: %w", err)
	}

	s.logger.Info("Role search completed", zap.Int("count", len(roles)))
	return roles, nil
}

// AssignRoleToUser assigns a role to a user
func (s *RoleService) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	s.logger.Info("Assigning role to user", zap.String("userID", userID), zap.String("roleID", roleID))

	// Check if role exists
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		s.logger.Error("Role not found", zap.String("roleID", roleID), zap.Error(err))
		return errors.NewNotFoundError("role not found")
	}

	// Check if assignment already exists
	exists, err := s.userRoleRepo.ExistsByUserAndRole(ctx, userID, roleID)
	if err != nil {
		s.logger.Error("Failed to check existing role assignment", zap.Error(err))
		return fmt.Errorf("failed to check role assignment: %w", err)
	}

	if exists {
		s.logger.Warn("Role already assigned to user", zap.String("userID", userID), zap.String("roleID", roleID))
		return fmt.Errorf("role already assigned to user")
	}

	// Create user-role assignment
	userRole := models.NewUserRole(userID, roleID)

	if err := s.userRoleRepo.Create(ctx, userRole); err != nil {
		s.logger.Error("Failed to assign role to user", zap.Error(err))
		return fmt.Errorf("failed to assign role: %w", err)
	}

	s.logger.Info("Role assigned to user successfully",
		zap.String("userID", userID),
		zap.String("roleID", roleID),
		zap.String("roleName", role.Name))
	return nil
}

// RemoveRoleFromUser removes a role from a user
func (s *RoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	s.logger.Info("Removing role from user", zap.String("userID", userID), zap.String("roleID", roleID))

	// Check if assignment exists
	userRole, err := s.userRoleRepo.GetByUserAndRole(ctx, userID, roleID)
	if err != nil {
		s.logger.Error("Role assignment not found", zap.Error(err))
		return errors.NewNotFoundError("role assignment not found")
	}

	// Delete the assignment
	if err := s.userRoleRepo.Delete(ctx, userRole.ID); err != nil {
		s.logger.Error("Failed to remove role from user", zap.Error(err))
		return fmt.Errorf("failed to remove role: %w", err)
	}

	s.logger.Info("Role removed from user successfully", zap.String("userID", userID), zap.String("roleID", roleID))
	return nil
}

// GetUserRoles retrieves all roles for a user
func (s *RoleService) GetUserRoles(ctx context.Context, userID string) ([]*models.UserRole, error) {
	s.logger.Info("Getting user roles", zap.String("userID", userID))

	userRoles, err := s.userRoleRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user roles", zap.String("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	s.logger.Info("User roles retrieved", zap.String("userID", userID), zap.Int("count", len(userRoles)))
	return userRoles, nil
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
