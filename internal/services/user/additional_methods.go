package user

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	userResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// GetUserByUsername retrieves a user by username
func (s *Service) GetUserByUsername(ctx context.Context, username string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user by username", zap.String("username", username))

	if username == "" {
		return nil, errors.NewValidationError("username cannot be empty")
	}

	// Get user by username using repository
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.Error("Failed to get user by username", zap.String("username", username), zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	response := &userResponses.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		CountryCode: user.CountryCode,
		IsValidated: user.IsValidated,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return response, nil
}

// GetUserByMobileNumber retrieves a user by mobile number
func (s *Service) GetUserByMobileNumber(ctx context.Context, mobileNumber uint64) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user by mobile", zap.Uint64("mobile", mobileNumber))

	phoneStr := fmt.Sprintf("%d", mobileNumber)
	user, err := s.userRepo.GetByPhoneNumber(ctx, phoneStr, "+91") // default country code
	if err != nil {
		s.logger.Error("Failed to get user by mobile", zap.Uint64("mobile", mobileNumber), zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	response := &userResponses.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		CountryCode: user.CountryCode,
		IsValidated: user.IsValidated,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return response, nil
}

// GetUserByAadhaarNumber retrieves a user by Aadhaar number
func (s *Service) GetUserByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user by Aadhaar", zap.String("aadhaar", aadhaarNumber))
	// This would require aadhaar field in user model - stub for now
	return nil, errors.NewValidationError("Aadhaar lookup not implemented")
}

// ListActiveUsers retrieves only active users
func (s *Service) ListActiveUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	s.logger.Info("Listing active users", zap.Int("limit", limit), zap.Int("offset", offset))

	users, err := s.userRepo.ListActive(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list active users", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	responses := make([]*userResponses.UserResponse, len(users))
	for i, user := range users {
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

	return responses, nil
}

// SearchUsers searches for users by keyword
func (s *Service) SearchUsers(ctx context.Context, keyword string, limit, offset int) (interface{}, error) {
	s.logger.Info("Searching users", zap.String("keyword", keyword), zap.Int("limit", limit))

	users, err := s.userRepo.Search(ctx, keyword, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search users", zap.String("keyword", keyword), zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	responses := make([]*userResponses.UserResponse, len(users))
	for i, user := range users {
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

	return responses, nil
}

// ValidateUser validates a user account
func (s *Service) ValidateUser(ctx context.Context, userID string) error {
	s.logger.Info("Validating user", zap.String("user_id", userID))

	if userID == "" {
		return errors.NewValidationError("user ID cannot be empty")
	}

	// Get user to validate
	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to get user for validation", zap.String("user_id", userID), zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Mark user as validated
	user.IsValidated = true
	status := "active"
	user.Status = &status

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user validation status", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Clear cache
	s.clearUserCache(userID)

	s.logger.Info("User validated successfully", zap.String("user_id", userID))
	return nil
}

// DeductTokens deducts tokens from user's balance
func (s *Service) DeductTokens(ctx context.Context, userID string, amount int) error {
	s.logger.Info("Deducting tokens", zap.String("user_id", userID), zap.Int("amount", amount))

	if userID == "" {
		return errors.NewValidationError("user ID cannot be empty")
	}
	if amount <= 0 {
		return errors.NewValidationError("amount must be positive")
	}

	// Get user to deduct tokens from
	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to get user for token deduction", zap.String("user_id", userID), zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Check if user has enough tokens
	if !user.DeductTokens(amount) {
		return errors.NewValidationError("insufficient token balance")
	}

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user tokens", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Clear cache
	s.clearUserCache(userID)

	s.logger.Info("Tokens deducted successfully", zap.String("user_id", userID), zap.Int("amount", amount))
	return nil
}

// AddTokens adds tokens to user's balance
func (s *Service) AddTokens(ctx context.Context, userID string, amount int) error {
	s.logger.Info("Adding tokens", zap.String("user_id", userID), zap.Int("amount", amount))

	if userID == "" {
		return errors.NewValidationError("user ID cannot be empty")
	}
	if amount <= 0 {
		return errors.NewValidationError("amount must be positive")
	}

	// Get user to add tokens to
	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to get user for token addition", zap.String("user_id", userID), zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Add tokens
	user.AddTokens(amount)

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user tokens", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Clear cache
	s.clearUserCache(userID)

	s.logger.Info("Tokens added successfully", zap.String("user_id", userID), zap.Int("amount", amount))
	return nil
}

// GetUserWithProfile retrieves user with profile information
func (s *Service) GetUserWithProfile(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user with profile", zap.String("user_id", userID))
	// This would require profile joins - stub for now
	return s.GetUserByID(ctx, userID)
}

// GetUserWithRoles retrieves user with role information
func (s *Service) GetUserWithRoles(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user with roles", zap.String("user_id", userID))
	// This would require role joins - stub for now
	return s.GetUserByID(ctx, userID)
}

// VerifyUserPassword verifies user password by username
func (s *Service) VerifyUserPassword(ctx context.Context, username, password string) (*userResponses.UserResponse, error) {
	s.logger.Info("Verifying user password", zap.String("username", username))

	if username == "" || password == "" {
		return nil, errors.NewValidationError("username and password are required")
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	response := &userResponses.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		CountryCode: user.CountryCode,
		IsValidated: user.IsValidated,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return response, nil
}

// VerifyUserPasswordByPhone verifies user password by phone number
func (s *Service) VerifyUserPasswordByPhone(ctx context.Context, phoneNumber, countryCode, password string) (*userResponses.UserResponse, error) {
	s.logger.Info("Verifying user password by phone", zap.String("phone", phoneNumber))

	if phoneNumber == "" || password == "" {
		return nil, errors.NewValidationError("phone number and password are required")
	}

	user, err := s.userRepo.GetByPhoneNumber(ctx, phoneNumber, countryCode)
	if err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	response := &userResponses.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		CountryCode: user.CountryCode,
		IsValidated: user.IsValidated,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return response, nil
}

// SetMPin sets user's mPin
func (s *Service) SetMPin(ctx context.Context, userID string, mPin string) error {
	s.logger.Info("Setting mPin", zap.String("user_id", userID))

	if userID == "" || mPin == "" {
		return errors.NewValidationError("user ID and mPin are required")
	}

	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		return errors.NewNotFoundError("user not found")
	}

	// Hash the mPin
	hashedMPin, err := bcrypt.GenerateFromPassword([]byte(mPin), bcrypt.DefaultCost)
	if err != nil {
		return errors.NewInternalError(err)
	}

	user.SetMPin(string(hashedMPin))
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to set mPin", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.clearUserCache(userID)
	return nil
}

// VerifyMPin verifies user's mPin
func (s *Service) VerifyMPin(ctx context.Context, userID string, mPin string) error {
	s.logger.Info("Verifying mPin", zap.String("user_id", userID))

	if userID == "" || mPin == "" {
		return errors.NewValidationError("user ID and mPin are required")
	}

	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		return errors.NewNotFoundError("user not found")
	}

	if !user.HasMPin() {
		return errors.NewBadRequestError("mPin not set for user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.MPin), []byte(mPin))
	if err != nil {
		return errors.NewUnauthorizedError("invalid mPin")
	}

	return nil
}

// GetUserByPhoneNumber retrieves user by phone number
func (s *Service) GetUserByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user by phone", zap.String("phone", phoneNumber), zap.String("country", countryCode))

	if phoneNumber == "" {
		return nil, errors.NewValidationError("phone number cannot be empty")
	}

	user, err := s.userRepo.GetByPhoneNumber(ctx, phoneNumber, countryCode)
	if err != nil {
		s.logger.Error("Failed to get user by phone", zap.String("phone", phoneNumber), zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	response := &userResponses.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		CountryCode: user.CountryCode,
		IsValidated: user.IsValidated,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return response, nil
}

// clearUserCache removes user data from cache
func (s *Service) clearUserCache(userID string) {
	cacheKey := fmt.Sprintf("user:%s", userID)
	if err := s.cacheService.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete user cache", zap.Error(err))
	}
}
