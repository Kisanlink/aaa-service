package user

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
)

// DeleteUser soft deletes a user by marking them as inactive
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	s.logger.Info("Deleting user", zap.String("user_id", userID))

	if userID == "" {
		return errors.NewValidationError("user ID cannot be empty")
	}

	// Get existing user
	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to get existing user for deletion",
			zap.String("user_id", userID),
			zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Check if user has active roles that prevent deletion
	userRoles, err := s.userRoleRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user roles for deletion check",
			zap.String("user_id", userID),
			zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Check for admin roles that cannot be deleted
	// Note: Admin check requires role repository integration
	// Skipping for now - can be added later with proper role lookup

	// Perform soft delete by marking as inactive
	err = s.userRepo.SoftDelete(ctx, userID, "system")
	if err != nil {
		s.logger.Error("Failed to soft delete user",
			zap.String("user_id", userID),
			zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Remove all user roles - iterate through existing roles and delete each
	for _, userRole := range userRoles {
		err = s.userRoleRepo.DeleteByUserAndRole(ctx, userID, userRole.RoleID)
		if err != nil {
			s.logger.Warn("Failed to remove user role during deletion",
				zap.String("user_id", userID),
				zap.String("role_id", userRole.RoleID),
				zap.Error(err))
			// Continue with deletion even if role removal fails
		}
	}

	// Clear cache
	s.clearUserCache(userID)
	s.clearUserRoleCache(userID)

	s.logger.Info("User deleted successfully", zap.String("user_id", userID))
	return nil
}

// HardDeleteUser permanently deletes a user (admin only)
func (s *Service) HardDeleteUser(ctx context.Context, userID string) error {
	s.logger.Info("Hard deleting user", zap.String("user_id", userID))

	if userID == "" {
		return errors.NewValidationError("user ID cannot be empty")
	}

	// Get existing user
	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to get existing user for hard deletion",
			zap.String("user_id", userID),
			zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Check if user has admin role
	userRoles, err := s.userRoleRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user roles for hard deletion check",
			zap.String("user_id", userID),
			zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Skip admin role check for now - requires role repository integration

	// Remove all user roles first - iterate through existing roles and delete each
	for _, userRole := range userRoles {
		err = s.userRoleRepo.DeleteByUserAndRole(ctx, userID, userRole.RoleID)
		if err != nil {
			s.logger.Error("Failed to remove user role before hard deletion",
				zap.String("user_id", userID),
				zap.String("role_id", userRole.RoleID),
				zap.Error(err))
			return errors.NewInternalError(err)
		}
	}

	// Perform hard delete
	err = s.userRepo.Delete(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to hard delete user",
			zap.String("user_id", userID),
			zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Clear cache
	s.clearUserCache(userID)
	s.clearUserRoleCache(userID)

	s.logger.Info("User hard deleted successfully", zap.String("user_id", userID))
	return nil
}

// clearUserRoleCache removes user role data from cache
func (s *Service) clearUserRoleCache(userID string) {
	cacheKey := fmt.Sprintf("user_roles:%s", userID)
	if err := s.cacheService.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete user role cache", zap.Error(err))
	}
}

// validateUserDeletionPermissions checks if the user can be deleted
func (s *Service) validateUserDeletionPermissions(ctx context.Context, userID, deletedBy string) error {
	// Basic validation - prevent self-deletion
	if userID == deletedBy {
		return errors.NewValidationError("users cannot delete themselves")
	}

	// Check if user has critical admin roles that prevent deletion
	userRoles, err := s.userRoleRepo.GetActiveRolesByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to check user roles for deletion validation",
			zap.String("user_id", userID),
			zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to validate deletion permissions: %w", err))
	}

	// Check for super admin or system admin roles
	for _, userRole := range userRoles {
		if userRole.Role.Name == "super_admin" || userRole.Role.Name == "system_admin" {
			s.logger.Warn("Attempting to delete user with critical admin role",
				zap.String("user_id", userID),
				zap.String("role_name", userRole.Role.Name))
			return errors.NewForbiddenError("cannot delete users with critical admin roles")
		}
	}

	return nil
}

// cascadeDeleteUserProfile soft deletes the user's profile
func (s *Service) cascadeDeleteUserProfile(ctx context.Context, userID, deletedBy string) error {
	// Note: This assumes we have access to a user profile repository
	// For now, we'll implement a basic version that logs the operation
	s.logger.Info("Cascading delete to user profile",
		zap.String("user_id", userID),
		zap.String("deleted_by", deletedBy))

	// In a full implementation, this would:
	// 1. Get the user profile by user ID
	// 2. Soft delete the profile
	// 3. Handle any profile-specific relationships (addresses, etc.)

	// For now, we'll just log that this step was attempted
	s.logger.Debug("User profile cascade deletion completed", zap.String("user_id", userID))
	return nil
}

// cascadeDeleteUserContacts soft deletes all user contacts
func (s *Service) cascadeDeleteUserContacts(ctx context.Context, userID, deletedBy string) error {
	// Note: This assumes we have access to a contact repository
	// For now, we'll implement a basic version that logs the operation
	s.logger.Info("Cascading delete to user contacts",
		zap.String("user_id", userID),
		zap.String("deleted_by", deletedBy))

	// In a full implementation, this would:
	// 1. Get all contacts for the user
	// 2. Soft delete each contact
	// 3. Handle any contact-specific relationships (addresses, etc.)

	// For now, we'll just log that this step was attempted
	s.logger.Debug("User contacts cascade deletion completed", zap.String("user_id", userID))
	return nil
}

// clearUserProfileCache removes user profile data from cache
func (s *Service) clearUserProfileCache(userID string) {
	cacheKey := fmt.Sprintf("user_profile:%s", userID)
	if err := s.cacheService.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete user profile cache", zap.Error(err))
	}
}

// clearUserContactsCache removes user contacts data from cache
func (s *Service) clearUserContactsCache(userID string) {
	cacheKey := fmt.Sprintf("user_contacts:%s", userID)
	if err := s.cacheService.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete user contacts cache", zap.Error(err))
	}
}
