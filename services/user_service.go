package services

import (
	"context"
	"fmt"

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

	// Check if user already exists
	if existingUser, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil && existingUser != nil {
		return nil, errors.NewConflictError("user already exists with this username")
	}

	// Create user model
	user := models.NewUser(req.Username, req.Password)

	// Hash password before saving
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Errorf("failed to hash password: %w", err))
	}
	user.Password = string(hashedPassword)

	// Create user in database
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to create user: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", user.ID))

	s.logger.Info("User created successfully", zap.String("userID", user.ID))

	// Return response
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
func (s *UserService) DeleteUser(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Deleting user", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Get user before deletion to return it in response
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	// Delete user from database
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		s.logger.Error("Failed to delete user", zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to delete user: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", userID))

	s.logger.Info("User deleted successfully", zap.String("userID", userID))

	// Return response
	response := &userResponses.UserResponse{}
	response.FromModel(user)
	return response, nil
}

// ListUsers lists users with filters and pagination
func (s *UserService) ListUsers(ctx context.Context, filters interface{}) (interface{}, error) {
	s.logger.Info("Listing users")

	// Type assertion for filters - for now, we'll use a simple struct with limit/offset
	type ListFilters struct {
		Limit  int
		Offset int
	}

	listFilters := &ListFilters{Limit: 10, Offset: 0}

	// If filters are provided, try to extract them
	if filters != nil {
		if f, ok := filters.(*ListFilters); ok {
			listFilters = f
		}
	}

	// Validate filters
	if listFilters.Limit <= 0 {
		listFilters.Limit = 10
	}
	if listFilters.Offset < 0 {
		listFilters.Offset = 0
	}

	// Get users from database
	users, err := s.userRepo.List(ctx, listFilters.Limit, listFilters.Offset)
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
		"limit":  listFilters.Limit,
		"offset": listFilters.Offset,
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
func (s *UserService) ValidateUser(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	s.logger.Info("Validating user", zap.String("userID", userID))

	// Validate user ID
	if err := s.validator.ValidateUserID(userID); err != nil {
		return nil, errors.NewValidationError("invalid user ID")
	}

	// Get user from database
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, errors.NewNotFoundError("user not found")
	}

	// Check if user is already validated
	if user.IsValidated {
		return nil, errors.NewConflictError("user is already validated")
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
		return nil, errors.NewInternalError(fmt.Errorf("failed to validate user: %w", err))
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("user:%s", user.ID))

	s.logger.Info("User validated successfully", zap.String("userID", userID))

	// Return response
	response := &userResponses.UserResponse{}
	response.FromModel(user)
	return response, nil
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
	return nil, errors.NewInternalError(fmt.Errorf("not implemented"))
}

func (s *UserService) UpdateUserProfile(ctx context.Context, userID string, req interface{}) (interface{}, error) {
	return nil, errors.NewInternalError(fmt.Errorf("not implemented"))
}

func (s *UserService) LockAccount(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	return nil, errors.NewInternalError(fmt.Errorf("not implemented"))
}

func (s *UserService) UnlockAccount(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	return nil, errors.NewInternalError(fmt.Errorf("not implemented"))
}

func (s *UserService) GetUserActivity(ctx context.Context, userID string) (interface{}, error) {
	return nil, errors.NewInternalError(fmt.Errorf("not implemented"))
}

func (s *UserService) GetUserAuditTrail(ctx context.Context, userID string) (interface{}, error) {
	return nil, errors.NewInternalError(fmt.Errorf("not implemented"))
}

func (s *UserService) BulkOperations(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, errors.NewInternalError(fmt.Errorf("not implemented"))
}
