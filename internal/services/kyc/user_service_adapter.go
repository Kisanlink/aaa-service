package kyc

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/users"
	"go.uber.org/zap"
)

// UserServiceAdapter adapts the existing user service to provide a generic Update method
type UserServiceAdapter struct {
	userRepo        interfaces.UserRepository
	userProfileRepo *users.UserProfileRepository
	logger          *zap.Logger
}

// NewUserServiceAdapter creates a new user service adapter
func NewUserServiceAdapter(userRepo interfaces.UserRepository, userProfileRepo *users.UserProfileRepository, logger *zap.Logger) UserService {
	return &UserServiceAdapter{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		logger:          logger,
	}
}

// Update updates user and user profile fields using a map of field names to values
func (a *UserServiceAdapter) Update(ctx context.Context, userID string, updates map[string]interface{}) error {
	a.logger.Info("Updating user profile fields",
		zap.String("user_id", userID),
		zap.Int("field_count", len(updates)))

	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	// Get existing user
	existingUser := &models.User{}
	_, err := a.userRepo.GetByID(ctx, userID, existingUser)
	if err != nil {
		a.logger.Error("Failed to get existing user for update",
			zap.String("user_id", userID),
			zap.Error(err))
		return fmt.Errorf("user not found: %w", err)
	}

	// Get or create user profile
	existingProfile, err := a.userProfileRepo.GetByUserID(ctx, userID)
	if err != nil {
		// Profile doesn't exist, create it
		a.logger.Info("User profile not found, creating new profile",
			zap.String("user_id", userID))
		existingProfile = models.NewUserProfile(userID)
	}

	// Apply updates to user and profile
	userUpdated := false
	profileUpdated := false

	for field, value := range updates {
		switch field {
		case "is_validated":
			if v, ok := value.(bool); ok {
				existingUser.IsValidated = v
				userUpdated = true
			}
		case "full_name", "name":
			if v, ok := value.(string); ok {
				existingProfile.Name = &v
				profileUpdated = true
			}
		case "aadhaar_verified":
			if v, ok := value.(bool); ok {
				existingProfile.AadhaarVerified = v
				profileUpdated = true
			}
		case "aadhaar_verified_at":
			if v, ok := value.(time.Time); ok {
				existingProfile.AadhaarVerifiedAt = &v
				profileUpdated = true
			}
		case "kyc_status":
			if v, ok := value.(string); ok {
				existingProfile.KYCStatus = v
				profileUpdated = true
			}
		case "photo_url", "photo":
			if v, ok := value.(string); ok {
				existingProfile.Photo = &v
				profileUpdated = true
			}
		default:
			a.logger.Warn("Unknown field in updates",
				zap.String("field", field),
				zap.String("user_id", userID))
		}
	}

	// Update user if needed
	if userUpdated {
		existingUser.UpdatedAt = time.Now()
		if err := a.userRepo.Update(ctx, existingUser); err != nil {
			a.logger.Error("Failed to update user in repository",
				zap.String("user_id", userID),
				zap.Error(err))
			return fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Update or create profile if needed
	if profileUpdated {
		existingProfile.UpdatedAt = time.Now()
		if existingProfile.ID == "" {
			// Create new profile
			if err := a.userProfileRepo.Create(ctx, existingProfile); err != nil {
				a.logger.Error("Failed to create user profile in repository",
					zap.String("user_id", userID),
					zap.Error(err))
				return fmt.Errorf("failed to create user profile: %w", err)
			}
		} else {
			// Update existing profile
			if err := a.userProfileRepo.Update(ctx, existingProfile); err != nil {
				a.logger.Error("Failed to update user profile in repository",
					zap.String("user_id", userID),
					zap.Error(err))
				return fmt.Errorf("failed to update user profile: %w", err)
			}
		}
	}

	a.logger.Info("User and profile updated successfully",
		zap.String("user_id", userID),
		zap.Bool("user_updated", userUpdated),
		zap.Bool("profile_updated", profileUpdated))

	return nil
}
