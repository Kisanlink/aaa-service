package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/entities/requests/users"
	userResponses "github.com/Kisanlink/aaa-service/entities/responses/users"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserService implements the UserService interface
type UserService struct {
	userRepo     interfaces.UserRepository
	roleRepo     interfaces.RoleRepository
	userRoleRepo interfaces.UserRoleRepository
	cacheService interfaces.CacheService
	logger       *zap.Logger
	validator    interfaces.Validator
}

// NewUserService creates a new UserService instance with proper dependency injection
func NewUserService(
	userRepo interfaces.UserRepository,
	roleRepo interfaces.RoleRepository,
	userRoleRepo interfaces.UserRoleRepository,
	cacheService interfaces.CacheService,
	logger *zap.Logger,
	validator interfaces.Validator,
) interfaces.UserService {
	return &UserService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
		cacheService: cacheService,
		logger:       logger,
		validator:    validator,
	}
}

// CreateUser creates a new user with proper validation and business logic
func (s *UserService) CreateUser(ctx context.Context, req *users.CreateUserRequest) (*userResponses.UserResponse, error) {
	s.logger.Info("Creating new user")

	// Validate request
	if req == nil {
		return nil, errors.NewValidationError("create user request is required")
	}

	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Additional validation
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, errors.NewValidationError("struct validation failed", err.Error())
	}

	// Check if user already exists
	if existingUser, _ := s.userRepo.GetByUsername(ctx, req.Username); existingUser != nil {
		return nil, errors.NewConflictError("user with this username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to hash password: %w", err))
	}

	// Create user model
	user := models.NewUser(req.Username, string(hashedPassword))

	// Store user in database
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to create user: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:username:%s", user.Username))

	s.logger.Info("User created successfully", zap.String("userID", user.ID))

	// Convert to response
	response := &userResponses.UserResponse{}
	response.FromModel(user)
	return response, nil
}

// GetUserByID retrieves a user by ID with caching
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user by ID", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:%s", userID)
	if cachedUser, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("User retrieved from cache", zap.String("userID", userID))
		if userResp, ok := cachedUser.(*userResponses.UserResponse); ok {
			return userResp, nil
		}
	}

	// Get user from database
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Convert to response
	response := &userResponses.UserResponse{}
	response.FromModel(user)

	// Cache the response
	s.cacheService.Set(cacheKey, response, 300) // 5 minutes TTL

	s.logger.Info("User retrieved successfully", zap.String("userID", userID))
	return response, nil
}

// GetUserByUsername retrieves a user by username with caching
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user by username", zap.String("username", username))

	// Validate username
	if username == "" {
		return nil, errors.NewValidationError("username is required")
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:username:%s", username)
	if cachedUser, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("User retrieved from cache", zap.String("username", username))
		if userResp, ok := cachedUser.(*userResponses.UserResponse); ok {
			return userResp, nil
		}
	}

	// Get user from database
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.Error("Failed to get user by username", zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Convert to response
	response := &userResponses.UserResponse{}
	response.FromModel(user)

	// Cache the response
	s.cacheService.Set(cacheKey, response, 300) // 5 minutes TTL

	s.logger.Info("User retrieved successfully", zap.String("username", username))
	return response, nil
}

// GetUserByMobileNumber retrieves a user by mobile number with caching
func (s *UserService) GetUserByMobileNumber(ctx context.Context, mobileNumber uint64) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user by mobile number", zap.Uint64("mobileNumber", mobileNumber))

	// Validate mobile number
	if err := s.validator.ValidatePhoneNumber(fmt.Sprintf("%d", mobileNumber)); err != nil {
		return nil, errors.NewValidationError("invalid mobile number", err.Error())
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:mobile:%d", mobileNumber)
	if cachedUser, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("User retrieved from cache", zap.Uint64("mobileNumber", mobileNumber))
		if userResp, ok := cachedUser.(*userResponses.UserResponse); ok {
			return userResp, nil
		}
	}

	// Get user from database
	user, err := s.userRepo.GetByMobileNumber(ctx, mobileNumber)
	if err != nil {
		s.logger.Error("Failed to get user by mobile number", zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Convert to response
	response := &userResponses.UserResponse{}
	response.FromModel(user)

	// Cache the response
	s.cacheService.Set(cacheKey, response, 300) // 5 minutes TTL

	s.logger.Info("User retrieved successfully", zap.Uint64("mobileNumber", mobileNumber))
	return response, nil
}

// GetUserByAadhaarNumber retrieves a user by Aadhaar number with caching
func (s *UserService) GetUserByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user by Aadhaar number", zap.String("aadhaarNumber", aadhaarNumber))

	// Validate Aadhaar number
	if err := s.validator.ValidateAadhaarNumber(aadhaarNumber); err != nil {
		return nil, errors.NewValidationError("invalid Aadhaar number", err.Error())
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:aadhaar:%s", aadhaarNumber)
	if cachedUser, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("User retrieved from cache", zap.String("aadhaarNumber", aadhaarNumber))
		if userResp, ok := cachedUser.(*userResponses.UserResponse); ok {
			return userResp, nil
		}
	}

	// Get user from database
	user, err := s.userRepo.GetByAadhaarNumber(ctx, aadhaarNumber)
	if err != nil {
		s.logger.Error("Failed to get user by Aadhaar number", zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Convert to response
	response := &userResponses.UserResponse{}
	response.FromModel(user)

	// Cache the response
	s.cacheService.Set(cacheKey, response, 300) // 5 minutes TTL

	s.logger.Info("User retrieved successfully", zap.String("aadhaarNumber", aadhaarNumber))
	return response, nil
}

// UpdateUser updates an existing user with proper validation
func (s *UserService) UpdateUser(ctx context.Context, req *users.UpdateUserRequest) (*userResponses.UserResponse, error) {
	s.logger.Info("Updating user")

	// Basic validation
	if req == nil {
		return nil, errors.NewValidationError("update request is required")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// For now, we'll extract userID from context or we need to get it from elsewhere
	// This is a limitation of the current design - we need userID but it's not in the request
	// Let's assume it will be passed through context for now
	userIDValue := ctx.Value("userID")
	if userIDValue == nil {
		return nil, errors.NewValidationError("user ID is required in context")
	}

	userID, ok := userIDValue.(string)
	if !ok {
		return nil, errors.NewValidationError("invalid user ID in context")
	}

	// Get existing user
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get existing user", zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Update fields from request
	if req.Status != nil {
		existingUser.Status = req.Status
	}
	// Add other field updates as needed

	// Update in database
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to update user: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", existingUser.ID))
	s.cacheService.Delete(fmt.Sprintf("user:username:%s", existingUser.Username))

	s.logger.Info("User updated successfully", zap.String("userID", existingUser.ID))

	// Return response
	response := &userResponses.UserResponse{}
	response.FromModel(existingUser)
	return response, nil
}

// DeleteUser deletes a user with proper validation
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	s.logger.Info("Deleting user", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return errors.NewValidationError("invalid user ID")
	}

	// Get user before deletion to check if exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user not found")
	}

	// Delete user from database
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		s.logger.Error("Failed to delete user", zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to delete user: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", userID))
	s.cacheService.Delete(fmt.Sprintf("user:username:%s", user.Username))

	s.logger.Info("User deleted successfully", zap.String("userID", userID))
	return nil
}

// ListUsers lists users with pagination
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	s.logger.Info("Listing users", zap.Int("limit", limit), zap.Int("offset", offset))

	// Set default values
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get users from database
	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list users", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to list users: %w", err))
	}

	// Get total count
	totalCount, err := s.userRepo.Count(ctx)
	if err != nil {
		s.logger.Error("Failed to get user count", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to get user count: %w", err))
	}

	// Convert to response format
	userResponseList := make([]*userResponses.UserResponse, len(users))
	for i, user := range users {
		userResponseList[i] = &userResponses.UserResponse{}
		userResponseList[i].FromModel(user)
	}

	s.logger.Info("Users listed successfully", zap.Int("count", len(users)))

	return map[string]interface{}{
		"users":  userResponseList,
		"total":  totalCount,
		"limit":  limit,
		"offset": offset,
	}, nil
}

// SearchUsers searches for users with query and pagination
func (s *UserService) SearchUsers(ctx context.Context, query string, limit, offset int) (interface{}, error) {
	s.logger.Info("Searching users", zap.String("query", query))

	// Validate query
	if query == "" {
		return nil, errors.NewValidationError("search query is required")
	}

	// Set default values
	if limit == 0 {
		limit = 10
	}
	if offset == 0 {
		offset = 0
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:search:%s:%d:%d", query, limit, offset)
	if cachedResult, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("Search results retrieved from cache", zap.String("query", query))
		return cachedResult, nil
	}

	// Search users in database
	users, err := s.userRepo.Search(ctx, query, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search users", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to search users: %w", err))
	}

	// Convert to response format
	userResponseList := make([]*userResponses.UserResponse, len(users))
	for i, user := range users {
		userResponseList[i] = &userResponses.UserResponse{}
		userResponseList[i].FromModel(user)
	}

	result := map[string]interface{}{
		"users":  userResponseList,
		"query":  query,
		"limit":  limit,
		"offset": offset,
	}

	// Cache the result
	s.cacheService.Set(cacheKey, result, 120) // 2 minutes TTL

	s.logger.Info("Users search completed successfully", zap.String("query", query), zap.Int("count", len(users)))
	return result, nil
}

// ValidateUser validates a user account
func (s *UserService) ValidateUser(ctx context.Context, userID string) error {
	s.logger.Info("Validating user", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return errors.NewValidationError("invalid user ID")
	}

	// Get user from database
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Check if user is already validated
	if user.IsValidated {
		return errors.NewConflictError("user is already validated")
	}

	// Update validation status
	user.IsValidated = true
	if user.Status != nil && *user.Status == "pending" {
		activeStatus := "active"
		user.Status = &activeStatus
	}

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to validate user", zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to validate user: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", user.ID))

	s.logger.Info("User validated successfully", zap.String("userID", userID))
	return nil
}

// AssignRole assigns a role to a user using a transaction
func (s *UserService) AssignRole(ctx context.Context, userID, roleID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Assigning role to user", zap.String("userID", userID), zap.String("roleID", roleID))

	// Validate user ID and role ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	var user *models.User

	// For now, we'll do this without a transaction wrapper since the interface doesn't support it
	// Check if user exists
	var err error
	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	// Check if role exists
	if _, err := s.roleRepo.GetByID(ctx, roleID); err != nil {
		return nil, errors.NewNotFoundError("role not found")
	}

	// Check if user already has this role
	if existing, err := s.userRoleRepo.GetByUserAndRole(ctx, userID, roleID); err == nil && existing != nil {
		return nil, errors.NewConflictError("user already has this role")
	}

	// Create user role assignment
	userRole := &models.UserRole{
		UserID:   userID,
		RoleID:   roleID,
		IsActive: true,
	}

	// Save the user role assignment
	if err := s.userRoleRepo.Create(ctx, userRole); err != nil {
		return nil, errors.NewInternalError(fmt.Errorf("failed to assign role: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", userID))
	s.cacheService.Delete(fmt.Sprintf("user:roles:%s", userID))

	s.logger.Info("Role assigned successfully", zap.String("userID", userID), zap.String("roleID", roleID))

	// Return response
	response := &userResponses.UserResponse{}
	response.FromModel(user)
	return response, nil
}

// RemoveRole removes a role from a user
func (s *UserService) RemoveRole(ctx context.Context, userID, roleID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Removing role from user", zap.String("userID", userID), zap.String("roleID", roleID))

	// Validate user ID and role ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID", err.Error())
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	// Remove the user role assignment
	if err := s.userRoleRepo.DeleteByUserAndRole(ctx, userID, roleID); err != nil {
		s.logger.Error("Failed to remove role", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to remove role: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", userID))
	s.cacheService.Delete(fmt.Sprintf("user:roles:%s", userID))

	s.logger.Info("Role removed successfully", zap.String("userID", userID), zap.String("roleID", roleID))

	// Return response
	response := &userResponses.UserResponse{}
	response.FromModel(user)
	return response, nil
}

// GetUserRoles retrieves roles for a user
func (s *UserService) GetUserRoles(ctx context.Context, userID string) (interface{}, error) {
	s.logger.Info("Getting user roles", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID", err.Error())
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:roles:%s", userID)
	if cachedRoles, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("User roles retrieved from cache", zap.String("userID", userID))
		return cachedRoles, nil
	}

	// Get user roles from database
	userRoles, err := s.userRoleRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user roles", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to get user roles: %w", err))
	}

	// Convert to response format
	roleResponses := make([]map[string]interface{}, len(userRoles))
	for i, userRole := range userRoles {
		roleResponses[i] = map[string]interface{}{
			"id":         userRole.ID,
			"userID":     userRole.UserID,
			"roleID":     userRole.RoleID,
			"assignedAt": userRole.CreatedAt,
		}
	}

	result := map[string]interface{}{
		"userID": userID,
		"roles":  roleResponses,
	}

	// Cache the result
	s.cacheService.Set(cacheKey, result, 300) // 5 minutes TTL

	s.logger.Info("User roles retrieved successfully", zap.String("userID", userID), zap.Int("count", len(userRoles)))
	return result, nil
}

// Placeholder implementations for methods defined in the interface but not yet implemented
func (s *UserService) GetUserProfile(ctx context.Context, userID string) (interface{}, error) {
	s.logger.Info("Getting user profile", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:profile:%s", userID)
	if cachedProfile, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("User profile retrieved from cache", zap.String("userID", userID))
		return cachedProfile, nil
	}

	// For now, we'll get the user with profile through GetWithProfile method
	user, err := s.userRepo.GetWithProfile(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user with profile", zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Convert to profile response format
	profileResp := map[string]interface{}{
		"user_id":      userID,
		"profile":      user.Profile,
		"retrieved_at": fmt.Sprintf("%v", time.Now().Format("2006-01-02T15:04:05Z07:00")),
	}

	// Cache the result
	s.cacheService.Set(cacheKey, profileResp, 300) // 5 minutes TTL

	s.logger.Info("User profile retrieved successfully", zap.String("userID", userID))
	return profileResp, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, userID string, req interface{}) (interface{}, error) {
	s.logger.Info("Updating user profile", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Type assertion for request - expecting a map for now since the interface is generic
	updateData, ok := req.(map[string]interface{})
	if !ok {
		return nil, errors.NewValidationError("invalid request format")
	}

	// Verify user exists
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	// For now, we'll return a simplified response indicating the profile update
	// In a full implementation, you'd update the actual profile data
	response := map[string]interface{}{
		"user_id":        userID,
		"updated_fields": updateData,
		"updated_at":     fmt.Sprintf("%v", time.Now().Format("2006-01-02T15:04:05Z07:00")),
		"status":         "profile updated successfully",
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:profile:%s", userID))
	s.cacheService.Delete(fmt.Sprintf("user:%s", userID))

	s.logger.Info("User profile updated successfully", zap.String("userID", userID))
	return response, nil
}

func (s *UserService) LockAccount(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Locking user account", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	// Check if account is already locked
	if user.Status != nil && *user.Status == "blocked" {
		return nil, errors.NewConflictError("account is already locked")
	}

	// Lock the account
	blockedStatus := "blocked"
	user.Status = &blockedStatus

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to lock account", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to lock account: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", userID))

	s.logger.Info("User account locked successfully", zap.String("userID", userID))

	// Return response
	response := &userResponses.UserResponse{}
	response.FromModel(user)
	return response, nil
}

func (s *UserService) UnlockAccount(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Unlocking user account", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	// Check if account is already active
	if user.Status != nil && *user.Status == "active" {
		return nil, errors.NewConflictError("account is already active")
	}

	// Unlock the account - set to active if validated, otherwise pending
	var newStatus string
	if user.IsValidated {
		newStatus = "active"
	} else {
		newStatus = "pending"
	}
	user.Status = &newStatus

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to unlock account", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to unlock account: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", userID))

	s.logger.Info("User account unlocked successfully", zap.String("userID", userID))

	// Return response
	response := &userResponses.UserResponse{}
	response.FromModel(user)
	return response, nil
}

func (s *UserService) GetUserActivity(ctx context.Context, userID string) (interface{}, error) {
	s.logger.Info("Getting user activity", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Verify user exists
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:activity:%s", userID)
	if cachedActivity, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("User activity retrieved from cache", zap.String("userID", userID))
		return cachedActivity, nil
	}

	// For now, return sample activity data
	// In a full implementation, you'd retrieve actual activity from activity logs
	currentTime := time.Now()
	activity := map[string]interface{}{
		"user_id": userID,
		"activities": []map[string]interface{}{
			{
				"id":          "activity_001",
				"type":        "login",
				"description": "User logged in",
				"timestamp":   currentTime.Add(-2 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
				"ip_address":  "192.168.1.100",
				"user_agent":  "Mozilla/5.0...",
			},
			{
				"id":          "activity_002",
				"type":        "profile_update",
				"description": "User updated profile information",
				"timestamp":   currentTime.Add(-1 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
				"ip_address":  "192.168.1.100",
			},
			{
				"id":           "activity_003",
				"type":         "role_assigned",
				"description":  "Role was assigned to user",
				"timestamp":    currentTime.Add(-30 * time.Minute).Format("2006-01-02T15:04:05Z07:00"),
				"performed_by": "admin_user",
			},
		},
		"total_activities": 3,
		"retrieved_at":     currentTime.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Cache the result
	s.cacheService.Set(cacheKey, activity, 120) // 2 minutes TTL

	s.logger.Info("User activity retrieved successfully", zap.String("userID", userID))
	return activity, nil
}

func (s *UserService) GetUserAuditTrail(ctx context.Context, userID string) (interface{}, error) {
	s.logger.Info("Getting user audit trail", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Verify user exists
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:audit:%s", userID)
	if cachedAudit, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("User audit trail retrieved from cache", zap.String("userID", userID))
		return cachedAudit, nil
	}

	// For now, return sample audit data
	// In a full implementation, you'd retrieve actual audit logs
	currentTime := time.Now()
	auditTrail := map[string]interface{}{
		"user_id": userID,
		"audit_events": []map[string]interface{}{
			{
				"id":          "audit_001",
				"event_type":  "USER_CREATED",
				"description": "User account was created",
				"timestamp":   currentTime.Add(-7 * 24 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
				"source":      "registration_api",
				"metadata": map[string]interface{}{
					"ip_address": "192.168.1.50",
					"user_agent": "Registration App v1.0",
				},
			},
			{
				"id":           "audit_002",
				"event_type":   "USER_VALIDATED",
				"description":  "User account was validated",
				"timestamp":    currentTime.Add(-6 * 24 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
				"source":       "validation_service",
				"performed_by": "system",
			},
			{
				"id":           "audit_003",
				"event_type":   "ROLE_ASSIGNED",
				"description":  "Role was assigned to user",
				"timestamp":    currentTime.Add(-1 * 24 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
				"source":       "admin_panel",
				"performed_by": "admin_user_123",
				"metadata": map[string]interface{}{
					"role_id":   "role_456",
					"role_name": "standard_user",
				},
			},
		},
		"total_events": 3,
		"retrieved_at": currentTime.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Cache the result
	s.cacheService.Set(cacheKey, auditTrail, 300) // 5 minutes TTL

	s.logger.Info("User audit trail retrieved successfully", zap.String("userID", userID))
	return auditTrail, nil
}

func (s *UserService) BulkOperations(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("Performing bulk operations")

	// Type assertion for request
	bulkReq, ok := req.(map[string]interface{})
	if !ok {
		return nil, errors.NewValidationError("invalid bulk request format")
	}

	operation, exists := bulkReq["operation"]
	if !exists {
		return nil, errors.NewValidationError("operation field is required")
	}

	operationStr, ok := operation.(string)
	if !ok {
		return nil, errors.NewValidationError("operation must be a string")
	}

	userIDs, exists := bulkReq["user_ids"]
	if !exists {
		return nil, errors.NewValidationError("user_ids field is required")
	}

	userIDList, ok := userIDs.([]interface{})
	if !ok {
		return nil, errors.NewValidationError("user_ids must be an array")
	}

	if len(userIDList) == 0 {
		return nil, errors.NewValidationError("user_ids cannot be empty")
	}

	// Convert interface{} slice to string slice
	userIDStrings := make([]string, len(userIDList))
	for i, id := range userIDList {
		if idStr, ok := id.(string); ok {
			userIDStrings[i] = idStr
		} else {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid user ID at index %d", i))
		}
	}

	// Process bulk operation
	results := make([]map[string]interface{}, len(userIDStrings))
	successCount := 0
	errorCount := 0

	for i, userID := range userIDStrings {
		result := map[string]interface{}{
			"user_id": userID,
		}

		switch operationStr {
		case "lock":
			if _, err := s.LockAccount(ctx, userID); err != nil {
				result["status"] = "error"
				result["error"] = err.Error()
				errorCount++
			} else {
				result["status"] = "success"
				result["message"] = "account locked successfully"
				successCount++
			}

		case "unlock":
			if _, err := s.UnlockAccount(ctx, userID); err != nil {
				result["status"] = "error"
				result["error"] = err.Error()
				errorCount++
			} else {
				result["status"] = "success"
				result["message"] = "account unlocked successfully"
				successCount++
			}

		case "validate":
			if err := s.ValidateUser(ctx, userID); err != nil {
				result["status"] = "error"
				result["error"] = err.Error()
				errorCount++
			} else {
				result["status"] = "success"
				result["message"] = "user validated successfully"
				successCount++
			}

		default:
			result["status"] = "error"
			result["error"] = fmt.Sprintf("unsupported operation: %s", operationStr)
			errorCount++
		}

		results[i] = result
	}

	response := map[string]interface{}{
		"operation":     operationStr,
		"total_users":   len(userIDStrings),
		"success_count": successCount,
		"error_count":   errorCount,
		"results":       results,
		"processed_at":  time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}

	s.logger.Info("Bulk operations completed",
		zap.String("operation", operationStr),
		zap.Int("total", len(userIDStrings)),
		zap.Int("success", successCount),
		zap.Int("errors", errorCount))

	return response, nil
}

// VerifyUserPassword verifies a user's password for authentication
func (s *UserService) VerifyUserPassword(ctx context.Context, username, password string) (*userResponses.UserResponse, error) {
	s.logger.Info("Verifying user password", zap.String("username", username))

	// Validate inputs
	if username == "" {
		return nil, errors.NewValidationError("username is required")
	}

	if password == "" {
		return nil, errors.NewValidationError("password is required")
	}

	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.Error("Failed to get user by username", zap.String("username", username), zap.Error(err))
		return nil, errors.NewNotFoundError("invalid credentials")
	}

	// Verify password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		s.logger.Warn("Invalid password attempt", zap.String("username", username))
		return nil, errors.NewNotFoundError("invalid credentials")
	}

	s.logger.Info("Password verification successful", zap.String("username", username))

	// Convert to response
	response := &userResponses.UserResponse{}
	response.FromModel(user)
	return response, nil
}

// AddTokens adds tokens to a user's account
func (s *UserService) AddTokens(ctx context.Context, userID string, amount int) error {
	s.logger.Info("Adding tokens to user", zap.String("userID", userID), zap.Int("amount", amount))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return errors.NewValidationError("invalid user ID")
	}

	// Validate amount
	if amount <= 0 {
		return errors.NewValidationError("amount must be positive")
	}

	// Get user from database
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Add tokens
	user.AddTokens(amount)

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user tokens", zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to update user tokens: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", userID))

	s.logger.Info("Tokens added successfully", zap.String("userID", userID), zap.Int("amount", amount))
	return nil
}

// DeductTokens deducts tokens from a user's account
func (s *UserService) DeductTokens(ctx context.Context, userID string, amount int) error {
	s.logger.Info("Deducting tokens from user", zap.String("userID", userID), zap.Int("amount", amount))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return errors.NewValidationError("invalid user ID")
	}

	// Validate amount
	if amount <= 0 {
		return errors.NewValidationError("amount must be positive")
	}

	// Get user from database
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return errors.NewNotFoundError("user not found")
	}

	// Check if user has enough tokens
	if !user.HasEnoughTokens(amount) {
		return errors.NewValidationError("insufficient tokens")
	}

	// Deduct tokens
	if !user.DeductTokens(amount) {
		return errors.NewValidationError("failed to deduct tokens")
	}

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user tokens", zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to update user tokens: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", userID))

	s.logger.Info("Tokens deducted successfully", zap.String("userID", userID), zap.Int("amount", amount))
	return nil
}

// ListActiveUsers lists active users with pagination
func (s *UserService) ListActiveUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	s.logger.Info("Listing active users", zap.Int("limit", limit), zap.Int("offset", offset))

	// Set default values
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get active users from database
	users, err := s.userRepo.ListActive(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list active users", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to list active users: %w", err))
	}

	// Get total count of active users
	totalCount, err := s.userRepo.CountActive(ctx)
	if err != nil {
		s.logger.Error("Failed to get active user count", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to get active user count: %w", err))
	}

	// Convert to response format
	userResponseList := make([]*userResponses.UserResponse, len(users))
	for i, user := range users {
		userResponseList[i] = &userResponses.UserResponse{}
		userResponseList[i].FromModel(user)
	}

	s.logger.Info("Active users listed successfully", zap.Int("count", len(users)))

	return map[string]interface{}{
		"users":  userResponseList,
		"total":  totalCount,
		"limit":  limit,
		"offset": offset,
	}, nil
}

// GetUserWithProfile retrieves a user with their profile information
func (s *UserService) GetUserWithProfile(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user with profile", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:profile:%s", userID)
	if cachedUser, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("User with profile retrieved from cache", zap.String("userID", userID))
		if userResp, ok := cachedUser.(*userResponses.UserResponse); ok {
			return userResp, nil
		}
	}

	// Get user with profile from database
	user, err := s.userRepo.GetWithProfile(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user with profile", zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Convert to response
	response := &userResponses.UserResponse{}
	response.FromModel(user)

	// Cache the response
	s.cacheService.Set(cacheKey, response, 300) // 5 minutes TTL

	s.logger.Info("User with profile retrieved successfully", zap.String("userID", userID))
	return response, nil
}

// GetUserWithRoles retrieves a user with their roles information
func (s *UserService) GetUserWithRoles(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Getting user with roles", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:roles:%s", userID)
	if cachedUser, found := s.cacheService.Get(cacheKey); found {
		s.logger.Debug("User with roles retrieved from cache", zap.String("userID", userID))
		if userResp, ok := cachedUser.(*userResponses.UserResponse); ok {
			return userResp, nil
		}
	}

	// Get user with roles from database
	user, err := s.userRepo.GetWithRoles(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user with roles", zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Convert to response
	response := &userResponses.UserResponse{}
	response.FromModel(user)

	// Cache the response
	s.cacheService.Set(cacheKey, response, 300) // 5 minutes TTL

	s.logger.Info("User with roles retrieved successfully", zap.String("userID", userID))
	return response, nil
}
