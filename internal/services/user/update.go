package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/entities/requests/users"
	userResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UpdateUser updates an existing user's information
func (s *Service) UpdateUser(ctx context.Context, req *users.UpdateUserRequest) (*userResponses.UserResponse, error) {
	userID := req.UserID
	s.logger.Info("Updating user", zap.String("user_id", userID))

	if userID == "" {
		return nil, errors.NewValidationError("user ID cannot be empty")
	}

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("User update validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid user data", err.Error())
	}

	// Get existing user
	existingUser := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, existingUser)
	if err != nil {
		s.logger.Error("Failed to get existing user for update", zap.String("user_id", userID), zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Update fields if provided
	if req.MobileNumber != nil {
		phoneStr := fmt.Sprintf("%d", *req.MobileNumber)
		countryCode := "+91" // default
		if req.CountryCode != nil {
			countryCode = *req.CountryCode
		}

		// Check if new phone number conflicts with existing users
		conflictUser, err := s.userRepo.GetByPhoneNumber(ctx, phoneStr, countryCode)
		if err == nil && conflictUser != nil && conflictUser.ID != userID {
			s.logger.Warn("Phone number already in use",
				zap.String("phone", phoneStr),
				zap.String("country", countryCode))
			return nil, errors.NewConflictError("phone number already in use")
		}
		existingUser.PhoneNumber = phoneStr
		existingUser.CountryCode = countryCode
		existingUser.IsValidated = false // Need to re-validate
	}

	existingUser.UpdatedAt = time.Now()

	// Update in repository
	err = s.userRepo.Update(ctx, existingUser)
	if err != nil {
		s.logger.Error("Failed to update user in repository",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Clear cache
	s.clearUserCache(userID)

	s.logger.Info("User updated successfully",
		zap.String("user_id", userID))

	// Convert to response format
	response := &userResponses.UserResponse{
		ID:          existingUser.ID,
		Username:    existingUser.Username,
		PhoneNumber: existingUser.PhoneNumber,
		CountryCode: existingUser.CountryCode,
		IsValidated: existingUser.IsValidated,
		CreatedAt:   existingUser.CreatedAt,
		UpdatedAt:   existingUser.UpdatedAt,
	}

	return response, nil
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	s.logger.Info("Changing user password", zap.String("user_id", userID))

	if userID == "" || oldPassword == "" || newPassword == "" {
		return errors.NewValidationError("user ID, old password, and new password are required")
	}

	// Get existing user
	existingUser := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, existingUser)
	if err != nil {
		s.logger.Error("Failed to get existing user for password change", zap.String("user_id", userID), zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(oldPassword)); err != nil {
		s.logger.Error("Invalid old password", zap.String("user_id", userID))
		return errors.NewUnauthorizedError("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new password", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Update password
	existingUser.Password = string(hashedPassword)
	existingUser.UpdatedAt = time.Now()

	// Update in repository
	err = s.userRepo.Update(ctx, existingUser)
	if err != nil {
		s.logger.Error("Failed to update user password in repository", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Clear cache
	s.clearUserCache(userID)

	s.logger.Info("User password changed successfully", zap.String("user_id", userID))
	return nil
}
