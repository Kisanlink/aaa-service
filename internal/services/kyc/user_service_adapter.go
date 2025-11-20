package kyc

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/users"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
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

// GetProfile retrieves the user profile by user ID
func (a *UserServiceAdapter) GetProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	return a.userProfileRepo.GetByUserID(ctx, userID)
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

	// Get existing user - this is optional, we'll try to update it but profile is the priority
	// Initialize with BaseModel to allow GORM to scan into it
	canUpdateUser := true
	existingUser, err := a.userRepo.GetByID(ctx, userID, &models.User{
		BaseModel: &base.BaseModel{},
	})
	if err != nil {
		a.logger.Warn("Failed to get existing user for update, will skip user record update",
			zap.String("user_id", userID),
			zap.Error(err))
		canUpdateUser = false
	} else {
		// Defensive check: ensure user was actually populated
		// BaseModel is an embedded pointer, so we need to check if it's nil
		if existingUser == nil {
			a.logger.Warn("User object is nil after GetByID, will skip user record update",
				zap.String("user_id", userID))
			canUpdateUser = false
		} else if existingUser.BaseModel == nil {
			a.logger.Warn("User BaseModel is nil after GetByID, will skip user record update",
				zap.String("user_id", userID))
			canUpdateUser = false
		} else if existingUser.GetID() == "" {
			a.logger.Warn("User ID is empty after GetByID, will skip user record update",
				zap.String("user_id", userID))
			canUpdateUser = false
		} else {
			a.logger.Info("User record loaded successfully",
				zap.String("user_id", userID),
				zap.String("phone_number", existingUser.PhoneNumber),
				zap.Bool("is_validated", existingUser.IsValidated))
		}
	}

	// Get or create user profile
	existingProfile, err := a.userProfileRepo.GetByUserID(ctx, userID)
	profileExists := err == nil
	if err != nil {
		// Profile doesn't exist, create it using constructor
		a.logger.Info("User profile not found, creating new profile",
			zap.String("user_id", userID))
		existingProfile = models.NewUserProfile(userID)
	}

	// Apply updates to user and profile
	userUpdated := false
	profileUpdated := false

	a.logger.Info("Processing field updates",
		zap.String("user_id", userID),
		zap.Any("fields", func() []string {
			fields := make([]string, 0, len(updates))
			for k := range updates {
				fields = append(fields, k)
			}
			return fields
		}()))

	for field, value := range updates {
		switch field {
		case "is_validated":
			if v, ok := value.(bool); ok && canUpdateUser {
				existingUser.IsValidated = v
				userUpdated = true
				a.logger.Info("Setting user validation status",
					zap.String("user_id", userID),
					zap.Bool("is_validated", v))
			} else if !canUpdateUser {
				a.logger.Warn("Skipping user field update - user record not available",
					zap.String("field", field),
					zap.String("user_id", userID))
			}
		case "status":
			if v, ok := value.(string); ok && canUpdateUser {
				existingUser.Status = &v
				userUpdated = true
				a.logger.Info("Setting user status",
					zap.String("user_id", userID),
					zap.String("status", v))
			} else if !canUpdateUser {
				a.logger.Warn("Skipping user status update - user record not available",
					zap.String("field", field),
					zap.String("user_id", userID))
			}
		case "full_name", "name":
			if v, ok := value.(string); ok {
				existingProfile.Name = &v
				profileUpdated = true
				a.logger.Info("Setting profile name",
					zap.String("user_id", userID),
					zap.String("name", v))
			}
		case "aadhaar_verified":
			if v, ok := value.(bool); ok {
				existingProfile.AadhaarVerified = v
				profileUpdated = true
				a.logger.Info("Setting Aadhaar verification status",
					zap.String("user_id", userID),
					zap.Bool("verified", v))
			}
		case "aadhaar_verified_at":
			if v, ok := value.(time.Time); ok {
				existingProfile.AadhaarVerifiedAt = &v
				profileUpdated = true
				a.logger.Info("Setting Aadhaar verification timestamp",
					zap.String("user_id", userID),
					zap.Time("verified_at", v))
			}
		case "kyc_status":
			if v, ok := value.(string); ok {
				existingProfile.KYCStatus = v
				profileUpdated = true
				a.logger.Info("Setting KYC status",
					zap.String("user_id", userID),
					zap.String("kyc_status", v))
			}
		case "photo_url", "photo":
			if v, ok := value.(string); ok {
				existingProfile.Photo = &v
				profileUpdated = true
				a.logger.Info("Setting profile photo",
					zap.String("user_id", userID),
					zap.Int("photo_length", len(v)))
			}
		case "address_id":
			if v, ok := value.(string); ok {
				existingProfile.AddressID = &v
				profileUpdated = true
				a.logger.Info("Linking address to user profile",
					zap.String("user_id", userID),
					zap.String("address_id", v))
			}
		default:
			a.logger.Warn("Unknown field in updates",
				zap.String("field", field),
				zap.String("user_id", userID))
		}
	}

	// If profile was just created, ensure it's marked for creation
	if !profileExists && profileUpdated {
		a.logger.Info("New profile created with Aadhaar verification data",
			zap.String("user_id", userID),
			zap.String("name", func() string {
				if existingProfile.Name != nil {
					return *existingProfile.Name
				}
				return ""
			}()))
	}

	// Update user if needed and possible
	if userUpdated && canUpdateUser {
		a.logger.Info("Updating user record",
			zap.String("user_id", userID),
			zap.Bool("is_validated", existingUser.IsValidated))
		existingUser.UpdatedAt = time.Now()
		if err := a.userRepo.Update(ctx, existingUser); err != nil {
			a.logger.Error("Failed to update user in repository",
				zap.String("user_id", userID),
				zap.Error(err))
			// Don't return error - profile update is more critical
			a.logger.Warn("Continuing despite user update failure - profile is priority")
		} else {
			a.logger.Info("User record updated successfully",
				zap.String("user_id", userID))
		}
	} else if userUpdated && !canUpdateUser {
		a.logger.Warn("User record needs update but cannot be loaded - skipping user update",
			zap.String("user_id", userID))
	}

	// Update or create profile if needed
	if profileUpdated {
		existingProfile.UpdatedAt = time.Now()
		if !profileExists {
			// Create new profile (use profileExists flag, not ID check)
			a.logger.Info("Creating new user profile",
				zap.String("user_id", userID),
				zap.String("name", func() string {
					if existingProfile.Name != nil {
						return *existingProfile.Name
					}
					return ""
				}()),
				zap.Bool("aadhaar_verified", existingProfile.AadhaarVerified),
				zap.String("kyc_status", existingProfile.KYCStatus),
				zap.String("address_id", func() string {
					if existingProfile.AddressID != nil {
						return *existingProfile.AddressID
					}
					return "nil"
				}()))
			if err := a.userProfileRepo.Create(ctx, existingProfile); err != nil {
				a.logger.Error("Failed to create user profile in repository",
					zap.String("user_id", userID),
					zap.Error(err))
				return fmt.Errorf("failed to create user profile: %w", err)
			}
			a.logger.Info("User profile created successfully",
				zap.String("user_id", userID),
				zap.String("profile_id", existingProfile.ID))
		} else {
			// Update existing profile
			a.logger.Info("Updating existing user profile",
				zap.String("user_id", userID),
				zap.String("profile_id", existingProfile.ID),
				zap.String("name", func() string {
					if existingProfile.Name != nil {
						return *existingProfile.Name
					}
					return ""
				}()),
				zap.Bool("aadhaar_verified", existingProfile.AadhaarVerified),
				zap.String("kyc_status", existingProfile.KYCStatus))
			if err := a.userProfileRepo.Update(ctx, existingProfile); err != nil {
				a.logger.Error("Failed to update user profile in repository",
					zap.String("user_id", userID),
					zap.String("profile_id", existingProfile.ID),
					zap.Error(err))
				return fmt.Errorf("failed to update user profile: %w", err)
			}
			a.logger.Info("User profile updated successfully",
				zap.String("user_id", userID),
				zap.String("profile_id", existingProfile.ID))
		}
	} else {
		a.logger.Info("No profile updates needed",
			zap.String("user_id", userID))
	}

	a.logger.Info("User and profile processing completed",
		zap.String("user_id", userID),
		zap.Bool("user_updated", userUpdated && canUpdateUser),
		zap.Bool("profile_updated", profileUpdated),
		zap.Bool("profile_exists", profileExists))

	return nil
}
