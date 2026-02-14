package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests/users"
	userResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user with proper validation and business logic
func (s *Service) CreateUser(ctx context.Context, req *users.CreateUserRequest) (*userResponses.UserResponse, error) {
	s.logger.Info("Creating new user")

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("User creation validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid user data", err.Error())
	}

	// Check if user already exists by phone number
	existingUser, err := s.userRepo.GetByPhoneNumber(ctx, req.PhoneNumber, req.CountryCode)
	if err == nil && existingUser != nil {
		s.logger.Warn("User already exists with phone number",
			zap.String("phone", req.PhoneNumber),
			zap.String("country", req.CountryCode))
		return nil, errors.NewConflictError("user with this phone number already exists")
	}

	// Check if username is already taken (if username is provided)
	if req.Username != nil && *req.Username != "" {
		existingUserByUsername, err := s.userRepo.GetByUsername(ctx, *req.Username)
		if err == nil && existingUserByUsername != nil {
			s.logger.Warn("Username already taken",
				zap.String("username", *req.Username))
			return nil, errors.NewConflictError("username is already taken")
		}
		// If there was an error but it's not "user not found", log it
		if err != nil && !strings.Contains(err.Error(), "user not found") {
			s.logger.Error("Error checking username uniqueness", zap.Error(err))
			return nil, errors.NewInternalError(err)
		}
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Create user model using the appropriate constructor
	var user *models.User
	if req.Username != nil && *req.Username != "" {
		user = models.NewUserWithUsername(req.PhoneNumber, req.CountryCode, *req.Username, hashedPassword)
	} else {
		user = models.NewUser(req.PhoneNumber, req.CountryCode, hashedPassword)
	}

	// Set must_change_password flag if requested
	if req.MustChangePassword {
		user.MustChangePassword = true
	}

	// Save user to repository
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Error("Failed to create user in repository", zap.Error(err))
		// Check if it's a database constraint violation
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") ||
			strings.Contains(err.Error(), "already exists") {
			return nil, errors.NewConflictError("user with this information already exists")
		}
		return nil, errors.NewInternalError(err)
	}

	// Log user creation with safe username extraction
	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	s.logger.Info("User created successfully",
		zap.String("user_id", user.ID),
		zap.String("username", username))

	// Convert to response format
	response := &userResponses.UserResponse{
		ID:                 user.ID,
		Username:           user.Username,
		PhoneNumber:        user.PhoneNumber,
		CountryCode:        user.CountryCode,
		IsValidated:        user.IsValidated,
		MustChangePassword: user.MustChangePassword,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}

	return response, nil
}

// hashPassword creates a bcrypt hash of the password
func (s *Service) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// generateValidationToken creates a validation token for new users
func (s *Service) generateValidationToken() string {
	// Implementation would generate a secure random token
	// For now, return a placeholder
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}
