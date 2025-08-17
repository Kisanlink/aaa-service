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
