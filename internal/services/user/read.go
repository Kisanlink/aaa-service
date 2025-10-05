package user

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	userResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
)

// GetUserByID retrieves a user by their unique identifier (only active users)
func (s *Service) GetUserByID(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user by ID", zap.String("user_id", userID))

	if userID == "" {
		return nil, errors.NewValidationError("user ID cannot be empty")
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:%s", userID)
	if cachedUser, exists := s.cacheService.Get(cacheKey); exists {
		if user, ok := cachedUser.(*userResponses.UserResponse); ok {
			s.logger.Debug("User found in cache", zap.String("user_id", userID))
			return user, nil
		}
	}

	// Get user from repository
	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to get user by ID", zap.String("user_id", userID), zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Security check: Ensure user is not deleted
	if user.DeletedAt != nil {
		s.logger.Warn("Attempt to access deleted user", zap.String("user_id", userID))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Convert to response format
	response := &userResponses.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		CountryCode: user.CountryCode,
		IsValidated: user.IsValidated,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	// Cache the result
	if err := s.cacheService.Set(cacheKey, response, 300); err != nil {
		s.logger.Warn("Failed to cache user response", zap.Error(err))
	}

	s.logger.Info("User retrieved successfully", zap.String("user_id", userID))
	return response, nil
}

// ListUsers retrieves a paginated list of active (non-deleted) users
func (s *Service) ListUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	s.logger.Info("Listing users",
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get only active (non-deleted) users using repository method
	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list users",
			zap.Int("limit", limit),
			zap.Int("offset", offset),
			zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Filter out deleted users at service level as additional security
	activeUsers := make([]*models.User, 0, len(users))
	for _, user := range users {
		if user.DeletedAt == nil {
			activeUsers = append(activeUsers, user)
		}
	}

	// Convert to response format
	responses := make([]*userResponses.UserResponse, len(activeUsers))
	for i, user := range activeUsers {
		responses[i] = &userResponses.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			PhoneNumber: user.PhoneNumber,
			CountryCode: user.CountryCode,
			IsValidated: user.IsValidated,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}
	}

	s.logger.Info("Users retrieved successfully",
		zap.Int("count", len(responses)))
	return responses, nil
}
