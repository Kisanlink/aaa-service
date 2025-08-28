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

// GetRoleByID retrieves a role by ID
func (s *RoleService) GetRoleByID(ctx context.Context, roleID string) (*models.Role, error) {
	s.logger.Info("Getting role by ID", zap.String("roleID", roleID))

	if roleID == "" {
		return nil, errors.NewValidationError("role ID cannot be empty")
	}

	role := &models.Role{}
	_, err := s.roleRepo.GetByID(ctx, roleID, role)
	if err != nil {
		s.logger.Error("Failed to get role by ID", zap.String("roleID", roleID), zap.Error(err))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

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
	existingRole := &models.Role{}
	_, err := s.roleRepo.GetByID(ctx, role.ID, existingRole)
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

	// Get role to delete
	role := &models.Role{}
	_, err := s.roleRepo.GetByID(ctx, roleID, role)
	if err != nil {
		s.logger.Error("Failed to get role for deletion", zap.String("roleID", roleID), zap.Error(err))
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Get user ID from context for audit trail
	userID := "system" // Default fallback
	if ctxUserID, ok := ctx.Value("user_id").(string); ok && ctxUserID != "" {
		userID = ctxUserID
	}

	// Soft delete role instead of hard delete
	if err := s.roleRepo.SoftDelete(ctx, roleID, userID); err != nil {
		s.logger.Error("Failed to soft delete role", zap.String("roleID", roleID), zap.Error(err))
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// Clear cache
	if err := s.cacheService.Delete(fmt.Sprintf("role:%s", roleID)); err != nil {
		s.logger.Error("Failed to delete role from cache", zap.Error(err))
	}

	s.logger.Info("Role soft deleted successfully", zap.String("roleID", roleID), zap.String("deletedBy", userID))
	return nil
}

// HardDeleteRole permanently deletes a role (admin only)
func (s *RoleService) HardDeleteRole(ctx context.Context, roleID string) error {
	s.logger.Info("Hard deleting role", zap.String("roleID", roleID))

	// Get role to delete
	role := &models.Role{}
	_, err := s.roleRepo.GetByID(ctx, roleID, role)
	if err != nil {
		s.logger.Error("Failed to get role for hard deletion", zap.String("roleID", roleID), zap.Error(err))
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Get user ID from context for audit trail
	userID := "system" // Default fallback
	if ctxUserID, ok := ctx.Value("user_id").(string); ok && ctxUserID != "" {
		userID = ctxUserID
	}

	// Hard delete role
	if err := s.roleRepo.Delete(ctx, roleID, role); err != nil {
		s.logger.Error("Failed to hard delete role", zap.String("roleID", roleID), zap.Error(err))
		return fmt.Errorf("failed to hard delete role: %w", err)
	}

	// Clear cache
	if err := s.cacheService.Delete(fmt.Sprintf("role:%s", roleID)); err != nil {
		s.logger.Error("Failed to delete role from cache", zap.Error(err))
	}

	s.logger.Info("Role hard deleted successfully", zap.String("roleID", roleID), zap.String("deletedBy", userID))
	return nil
}

// ListRoles lists roles with pagination
func (s *RoleService) ListRoles(ctx context.Context, limit, offset int) ([]*models.Role, error) {
	s.logger.Info("Listing roles", zap.Int("limit", limit), zap.Int("offset", offset))

	// Add debug logging to see what's happening
	s.logger.Debug("Calling role repository List method")

	roles, err := s.roleRepo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list roles", zap.Error(err))
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	s.logger.Info("Role listing completed", zap.Int("count", len(roles)))

	// Add debug logging to see what roles were returned
	for i, role := range roles {
		s.logger.Debug("Role found",
			zap.Int("index", i),
			zap.String("role_id", role.ID),
			zap.String("role_name", role.Name),
			zap.Bool("is_active", role.IsActive))
	}

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

// AssignRole assigns a role to a user with comprehensive validation and error handling
func (s *RoleService) AssignRole(ctx context.Context, userID, roleID string) error {
	s.logger.Info("Assigning role to user", zap.String("userID", userID), zap.String("roleID", roleID))

	// Validate role assignment
	if err := s.ValidateRoleAssignment(ctx, userID, roleID); err != nil {
		return err
	}

	// Use the enhanced repository method with transaction support
	if err := s.userRoleRepo.AssignRole(ctx, userID, roleID); err != nil {
		s.logger.Error("Failed to assign role to user", zap.String("userID", userID), zap.String("roleID", roleID), zap.Error(err))
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Clear cache
	if err := s.cacheService.Delete(fmt.Sprintf("user_roles:%s", userID)); err != nil {
		s.logger.Error("Failed to clear user roles cache", zap.Error(err))
	}

	s.logger.Info("Role assigned to user successfully", zap.String("userID", userID), zap.String("roleID", roleID))
	return nil
}

// AssignRoleToUser assigns a role to a user (legacy method for backward compatibility)
func (s *RoleService) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	return s.AssignRole(ctx, userID, roleID)
}

// RemoveRole removes a role from a user with proper constraint handling
func (s *RoleService) RemoveRole(ctx context.Context, userID, roleID string) error {
	s.logger.Info("Removing role from user", zap.String("userID", userID), zap.String("roleID", roleID))

	if userID == "" || roleID == "" {
		return errors.NewValidationError("user ID and role ID are required")
	}

	// Use the enhanced repository method with proper constraint handling
	if err := s.userRoleRepo.RemoveRole(ctx, userID, roleID); err != nil {
		s.logger.Error("Failed to remove role from user", zap.String("userID", userID), zap.String("roleID", roleID), zap.Error(err))
		return fmt.Errorf("failed to remove role: %w", err)
	}

	// Clear cache
	if err := s.cacheService.Delete(fmt.Sprintf("user_roles:%s", userID)); err != nil {
		s.logger.Error("Failed to clear user roles cache", zap.Error(err))
	}

	s.logger.Info("Role removed from user successfully", zap.String("userID", userID), zap.String("roleID", roleID))
	return nil
}

// RemoveRoleFromUser removes a role from a user (legacy method for backward compatibility)
func (s *RoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	return s.RemoveRole(ctx, userID, roleID)
}

// GetUserRoles retrieves all active roles for a user with complete role details
func (s *RoleService) GetUserRoles(ctx context.Context, userID string) ([]*models.UserRole, error) {
	s.logger.Info("Getting user roles with details", zap.String("userID", userID))

	if userID == "" {
		return nil, errors.NewValidationError("user ID cannot be empty")
	}

	// Use the enhanced repository method that preloads role details
	userRoles, err := s.userRoleRepo.GetActiveRolesByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user roles with details", zap.String("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	s.logger.Info("User roles with details retrieved", zap.String("userID", userID), zap.Int("count", len(userRoles)))
	return userRoles, nil
}

// ValidateRoleAssignment validates that both user and role exist and are active before assignment
func (s *RoleService) ValidateRoleAssignment(ctx context.Context, userID, roleID string) error {
	s.logger.Debug("Validating role assignment", zap.String("userID", userID), zap.String("roleID", roleID))

	if userID == "" || roleID == "" {
		return errors.NewValidationError("user ID and role ID are required")
	}

	// Check if role exists and is active
	role := &models.Role{}
	_, err := s.roleRepo.GetByID(ctx, roleID, role)
	if err != nil {
		s.logger.Error("Role not found for assignment", zap.String("roleID", roleID), zap.Error(err))
		return errors.NewNotFoundError("role not found")
	}

	if !role.IsActive {
		s.logger.Error("Role is not active", zap.String("roleID", roleID))
		return errors.NewValidationError("role is not active")
	}

	// Check if role is already assigned to user
	isAssigned, err := s.userRoleRepo.IsRoleAssigned(ctx, userID, roleID)
	if err != nil {
		s.logger.Error("Failed to check existing role assignment", zap.String("userID", userID), zap.String("roleID", roleID), zap.Error(err))
		return fmt.Errorf("failed to validate role assignment: %w", err)
	}

	if isAssigned {
		s.logger.Warn("Role already assigned to user", zap.String("userID", userID), zap.String("roleID", roleID))
		return errors.NewConflictError("role already assigned to user")
	}

	s.logger.Debug("Role assignment validation successful", zap.String("userID", userID), zap.String("roleID", roleID))
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
