package user

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	userResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
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

	// Use the repository's Search method which has database-level pagination
	users, err := s.userRepo.Search(ctx, keyword, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search users", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Search completed", zap.Int("result_count", len(users)))

	// Convert to response format
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

// GetUserWithProfile retrieves user with profile information with caching
func (s *Service) GetUserWithProfile(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user with profile", zap.String("user_id", userID))

	if userID == "" {
		return nil, errors.NewValidationError("user ID cannot be empty")
	}

	// Use cached profile method
	return s.getCachedUserProfile(ctx, userID)
}

// GetUserWithRoles retrieves user with complete role information with caching
func (s *Service) GetUserWithRoles(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user with roles", zap.String("user_id", userID))

	if userID == "" {
		return nil, errors.NewValidationError("user ID cannot be empty")
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user_with_roles:%s", userID)
	if cachedResponse, exists := s.cacheService.Get(cacheKey); exists {
		if response, ok := cachedResponse.(*userResponses.UserResponse); ok {
			s.logger.Debug("User with roles found in cache", zap.String("user_id", userID))
			return response, nil
		}
	}

	// Get user from repository
	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to get user by ID", zap.String("user_id", userID), zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Get user roles with role details (with caching)
	userRoles, err := s.getCachedUserRoles(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user roles", zap.String("user_id", userID), zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format with roles
	response := &userResponses.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		CountryCode: user.CountryCode,
		IsValidated: user.IsValidated,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Tokens:      user.Tokens,
		HasMPin:     user.HasMPin(),
	}

	// Add roles to response
	roles := make([]userResponses.UserRoleDetail, len(userRoles))
	for i, userRole := range userRoles {
		roles[i] = userResponses.UserRoleDetail{
			ID:       userRole.ID,
			UserID:   userRole.UserID,
			RoleID:   userRole.RoleID,
			IsActive: userRole.IsActive,
			Role: userResponses.RoleDetail{
				ID:          userRole.Role.ID,
				Name:        userRole.Role.Name,
				Description: userRole.Role.Description,
				IsActive:    userRole.Role.IsActive,
				AssignedAt:  userRole.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
		}
	}
	response.Roles = roles

	// Cache the complete response for 15 minutes
	if err := s.cacheService.Set(cacheKey, response, 900); err != nil {
		s.logger.Warn("Failed to cache user with roles response", zap.Error(err))
	}

	s.logger.Info("User with roles retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("role_count", len(roles)))
	return response, nil
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

// SetMPin sets user's MPIN with secure hashing and validation (requires password verification)
func (s *Service) SetMPin(ctx context.Context, userID string, mPin string, currentPassword string) error {
	s.logger.Info("Setting MPIN", zap.String("user_id", userID))

	if userID == "" || mPin == "" {
		return errors.NewValidationError("user ID and MPIN are required")
	}

	if currentPassword == "" {
		return errors.NewValidationError("current password is required to set MPIN")
	}

	// Validate MPIN format (4-6 digits)
	if len(mPin) < 4 || len(mPin) > 6 {
		return errors.NewValidationError("MPIN must be 4-6 digits")
	}

	// Validate MPIN contains only digits
	for _, char := range mPin {
		if char < '0' || char > '9' {
			return errors.NewValidationError("MPIN must contain only digits")
		}
	}

	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to get user for MPIN setting", zap.String("user_id", userID), zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Security check: Ensure user is not deleted
	if user.DeletedAt != nil {
		s.logger.Warn("SetMPin attempt for deleted user", zap.String("user_id", userID))
		return errors.NewNotFoundError("user not found")
	}

	// Check if MPIN is already set
	if user.HasMPin() {
		return errors.NewConflictError("MPIN is already set. Use update-mpin endpoint to change it")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		s.logger.Warn("Invalid password during MPIN setup", zap.String("user_id", userID))
		return errors.NewUnauthorizedError("invalid password")
	}

	// Hash the MPIN with appropriate cost
	hashedMPin, err := bcrypt.GenerateFromPassword([]byte(mPin), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash MPIN", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	user.SetMPin(string(hashedMPin))
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to set MPIN", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.clearUserCache(userID)
	s.logger.Info("MPIN set successfully", zap.String("user_id", userID))
	return nil
}

// UpdateMPin updates user's existing MPIN with current MPIN verification
func (s *Service) UpdateMPin(ctx context.Context, userID, currentMPin, newMPin string) error {
	s.logger.Info("Updating MPIN", zap.String("user_id", userID))

	if userID == "" || currentMPin == "" || newMPin == "" {
		return errors.NewValidationError("user ID, current MPIN, and new MPIN are required")
	}

	// Validate new MPIN format (4-6 digits)
	if len(newMPin) < 4 || len(newMPin) > 6 {
		return errors.NewValidationError("new MPIN must be 4-6 digits")
	}

	// Validate new MPIN contains only digits
	for _, char := range newMPin {
		if char < '0' || char > '9' {
			return errors.NewValidationError("new MPIN must contain only digits")
		}
	}

	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to get user for MPIN update", zap.String("user_id", userID), zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Security check: Ensure user is not deleted
	if user.DeletedAt != nil {
		s.logger.Warn("UpdateMPin attempt for deleted user", zap.String("user_id", userID))
		return errors.NewNotFoundError("user not found")
	}

	// Check if MPIN is set
	if !user.HasMPin() {
		return errors.NewNotFoundError("MPIN not set for this user. Use set-mpin endpoint first")
	}

	// Verify current MPIN
	if err := bcrypt.CompareHashAndPassword([]byte(*user.MPin), []byte(currentMPin)); err != nil {
		s.logger.Warn("Invalid current MPIN during update", zap.String("user_id", userID))
		return errors.NewUnauthorizedError("invalid current MPIN")
	}

	// Hash the new MPIN
	hashedMPin, err := bcrypt.GenerateFromPassword([]byte(newMPin), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new MPIN", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	user.SetMPin(string(hashedMPin))
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to update MPIN", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.clearUserCache(userID)
	s.logger.Info("MPIN updated successfully", zap.String("user_id", userID))
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

	// Security check: Ensure user is not deleted
	if user.DeletedAt != nil {
		s.logger.Warn("VerifyMPin attempt for deleted user", zap.String("user_id", userID))
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

// VerifyUserCredentials verifies user credentials supporting both password and MPIN authentication
func (s *Service) VerifyUserCredentials(ctx context.Context, phone, countryCode string, password, mpin *string) (*userResponses.UserResponse, error) {
	s.logger.Info("Verifying user credentials", zap.String("phone", phone), zap.String("country", countryCode))

	if phone == "" || countryCode == "" {
		return nil, errors.NewValidationError("phone number and country code are required")
	}

	if password == nil && mpin == nil {
		return nil, errors.NewValidationError("either password or mpin must be provided")
	}

	// Get user by phone number
	user, err := s.userRepo.GetByPhoneNumber(ctx, phone, countryCode)
	if err != nil {
		s.logger.Error("Failed to get user by phone", zap.String("phone", phone), zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Prioritize password authentication if both are provided
	if password != nil && *password != "" {
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(*password))
		if err != nil {
			s.logger.Error("Password verification failed", zap.String("user_id", user.ID))
			return nil, errors.NewUnauthorizedError("invalid credentials")
		}
	} else if mpin != nil && *mpin != "" {
		if !user.HasMPin() {
			s.logger.Error("MPIN not set for user", zap.String("user_id", user.ID))
			return nil, errors.NewBadRequestError("mpin not set for user")
		}

		err = bcrypt.CompareHashAndPassword([]byte(*user.MPin), []byte(*mpin))
		if err != nil {
			s.logger.Error("MPIN verification failed", zap.String("user_id", user.ID))
			return nil, errors.NewUnauthorizedError("invalid mpin")
		}
	} else {
		return nil, errors.NewValidationError("no valid credentials provided")
	}

	// Warm cache for frequently accessed data after successful login
	go func() {
		if err := s.warmUserCache(context.Background(), user.ID); err != nil {
			s.logger.Warn("Failed to warm user cache after login", zap.String("user_id", user.ID), zap.Error(err))
		}
	}()

	// Get user with roles for complete response
	return s.GetUserWithRoles(ctx, user.ID)
}

// getCachedUserRoles retrieves user roles with caching
func (s *Service) getCachedUserRoles(ctx context.Context, userID string) ([]*models.UserRole, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("user_roles:%s", userID)
	if cachedRoles, exists := s.cacheService.Get(cacheKey); exists {
		if roles, ok := cachedRoles.([]*models.UserRole); ok {
			s.logger.Debug("User roles found in cache", zap.String("user_id", userID))
			return roles, nil
		}
	}

	// Get from repository if not in cache
	userRoles, err := s.userRoleRepo.GetActiveRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache the result for 15 minutes
	if err := s.cacheService.Set(cacheKey, userRoles, 900); err != nil {
		s.logger.Warn("Failed to cache user roles", zap.Error(err))
	}

	return userRoles, nil
}

// getCachedUserProfile retrieves user profile with caching
func (s *Service) getCachedUserProfile(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("user_profile:%s", userID)
	if cachedProfile, exists := s.cacheService.Get(cacheKey); exists {
		if profile, ok := cachedProfile.(*userResponses.UserResponse); ok {
			s.logger.Debug("User profile found in cache", zap.String("user_id", userID))
			return profile, nil
		}
	}

	// Get user with profile from repository
	user, err := s.userRepo.GetWithProfile(ctx, userID)
	if err != nil {
		return nil, err
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
		Tokens:      user.Tokens,
		HasMPin:     user.HasMPin(),
	}

	// Cache the result for 30 minutes
	if err := s.cacheService.Set(cacheKey, response, 1800); err != nil {
		s.logger.Warn("Failed to cache user profile", zap.Error(err))
	}

	return response, nil
}

// warmUserCache preloads frequently accessed user data into cache
func (s *Service) warmUserCache(ctx context.Context, userID string) error {
	s.logger.Debug("Warming user cache", zap.String("user_id", userID))

	// Warm user basic info cache
	_, err := s.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to warm user basic cache", zap.String("user_id", userID), zap.Error(err))
	}

	// Warm user roles cache
	_, err = s.getCachedUserRoles(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to warm user roles cache", zap.String("user_id", userID), zap.Error(err))
	}

	// Warm user profile cache
	_, err = s.getCachedUserProfile(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to warm user profile cache", zap.String("user_id", userID), zap.Error(err))
	}

	// Warm user with roles cache
	_, err = s.GetUserWithRoles(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to warm user with roles cache", zap.String("user_id", userID), zap.Error(err))
	}

	s.logger.Debug("User cache warming completed", zap.String("user_id", userID))
	return nil
}

// clearUserCache removes all user-related data from cache
func (s *Service) clearUserCache(userID string) {
	cacheKeys := []string{
		fmt.Sprintf("user:%s", userID),
		fmt.Sprintf("user_roles:%s", userID),
		fmt.Sprintf("user_profile:%s", userID),
		fmt.Sprintf("user_with_roles:%s", userID),
	}

	for _, key := range cacheKeys {
		if err := s.cacheService.Delete(key); err != nil {
			s.logger.Warn("Failed to delete cache key", zap.String("key", key), zap.Error(err))
		}
	}

	s.logger.Debug("User cache cleared", zap.String("user_id", userID))
}

// SoftDeleteUserWithCascade performs soft delete of user with cascade operations and cache invalidation
func (s *Service) SoftDeleteUserWithCascade(ctx context.Context, userID, deletedBy string) error {
	s.logger.Info("Soft deleting user with cascade", zap.String("user_id", userID), zap.String("deleted_by", deletedBy))

	if userID == "" {
		return errors.NewValidationError("user ID cannot be empty")
	}

	if deletedBy == "" {
		return errors.NewValidationError("deleted by cannot be empty")
	}

	// Check if user exists
	user := &models.User{}
	_, err := s.userRepo.GetByID(ctx, userID, user)
	if err != nil {
		s.logger.Error("Failed to get user for deletion", zap.String("user_id", userID), zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Deactivate all user role assignments first
	userRoles, err := s.userRoleRepo.GetActiveRolesByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user roles for cascade deletion", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	for _, userRole := range userRoles {
		if err := s.userRoleRepo.RemoveRole(ctx, userID, userRole.RoleID); err != nil {
			s.logger.Error("Failed to remove role during cascade deletion",
				zap.String("user_id", userID),
				zap.String("role_id", userRole.RoleID),
				zap.Error(err))
			// Continue with other roles even if one fails
		}
	}

	// Perform soft delete on user
	if err := s.userRepo.SoftDelete(ctx, userID, deletedBy); err != nil {
		s.logger.Error("Failed to soft delete user", zap.String("user_id", userID), zap.Error(err))
		return errors.NewInternalError(err)
	}

	// Clear all user-related cache entries
	s.clearUserCache(userID)

	s.logger.Info("User soft deleted successfully with cascade",
		zap.String("user_id", userID),
		zap.String("deleted_by", deletedBy),
		zap.Int("roles_removed", len(userRoles)))
	return nil
}

// invalidateUserRoleCache removes user role-related cache entries
func (s *Service) invalidateUserRoleCache(userID string) {
	cacheKeys := []string{
		fmt.Sprintf("user_roles:%s", userID),
		fmt.Sprintf("user_with_roles:%s", userID),
	}

	for _, key := range cacheKeys {
		if err := s.cacheService.Delete(key); err != nil {
			s.logger.Warn("Failed to delete role cache key", zap.String("key", key), zap.Error(err))
		}
	}

	s.logger.Debug("User role cache invalidated", zap.String("user_id", userID))
}
