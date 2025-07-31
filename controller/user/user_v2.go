package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/helper"
	pb "github.com/Kisanlink/aaa-service/proto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ServerV2 represents the v2 user service server
type ServerV2 struct {
	pb.UnimplementedUserServiceV2Server
	DB *gorm.DB
}

// NewUserServerV2 creates a new v2 user server instance
func NewUserServerV2(db *gorm.DB) *ServerV2 {
	return &ServerV2{DB: db}
}

// CheckPasswordHash verifies a password against its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Login handles user login with enhanced functionality
func (s *ServerV2) Login(ctx context.Context, req *pb.LoginRequestV2) (*pb.LoginResponseV2, error) {
	log.Printf("V2 Login request received for user: %s", req.Username)

	// Find user by username
	var user models.User
	if err := s.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return &pb.LoginResponseV2{
			StatusCode: 401,
			Message:    "Invalid credentials",
		}, nil
	}

	// Verify password
	if !CheckPasswordHash(req.Password, user.Password) {
		return &pb.LoginResponseV2{
			StatusCode: 401,
			Message:    "Invalid credentials",
		}, nil
	}

	// Get user roles
	var userRoles []models.UserRole
	if err := s.DB.Where("user_id = ?", user.ID).Find(&userRoles).Error; err != nil {
		log.Printf("Error fetching user roles: %v", err)
	}

	// Convert to protobuf user roles
	var pbUserRoles []*pb.UserRoleV2
	for _, ur := range userRoles {
		pbUserRole := &pb.UserRoleV2{
			Id:        ur.ID,
			UserId:    ur.UserID,
			RoleId:    ur.RoleID,
			CreatedAt: ur.CreatedAt.Format(time.RFC3339),
			UpdatedAt: ur.UpdatedAt.Format(time.RFC3339),
		}
		pbUserRoles = append(pbUserRoles, pbUserRole)
	}

	// Generate tokens
	accessToken, err := helper.GenerateAccessToken(user.ID, nil, user.Username, user.IsValidated)
	if err != nil {
		return &pb.LoginResponseV2{
			StatusCode: 500,
			Message:    "Failed to generate access token",
		}, nil
	}

	refreshToken, err := helper.GenerateRefreshToken(user.ID, nil, user.Username, user.IsValidated)
	if err != nil {
		return &pb.LoginResponseV2{
			StatusCode: 500,
			Message:    "Failed to generate refresh token",
		}, nil
	}

	// Create user response
	pbUser := &pb.UserV2{
		Id:          user.ID,
		Username:    user.Username,
		Email:       "", // User model doesn't have email field
		FullName:    "", // User model doesn't have full_name field
		IsValidated: user.IsValidated,
		Status:      "", // Will be set below
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		UserRoles:   pbUserRoles,
	}

	// Set status if available
	if user.Status != nil {
		pbUser.Status = *user.Status
	}

	return &pb.LoginResponseV2{
		StatusCode:   200,
		Message:      "Login successful",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User:         pbUser,
	}, nil
}

// Register handles user registration with enhanced functionality
func (s *ServerV2) Register(ctx context.Context, req *pb.RegisterRequestV2) (*pb.RegisterResponseV2, error) {
	log.Printf("V2 Register request received for user: %s", req.Username)

	// Check if user already exists
	var existingUser models.User
	if err := s.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return &pb.RegisterResponseV2{
			StatusCode: 409,
			Message:    "User already exists",
		}, nil
	}

	// Hash password using the existing function
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return &pb.RegisterResponseV2{
			StatusCode: 500,
			Message:    "Failed to hash password",
		}, nil
	}

	// Create user
	user := models.NewUser(req.Username, hashedPassword)
	user.IsValidated = false

	if err := s.DB.Create(&user).Error; err != nil {
		return &pb.RegisterResponseV2{
			StatusCode: 500,
			Message:    "Failed to create user",
		}, nil
	}

	// Assign default roles if provided
	if len(req.RoleIds) > 0 {
		for _, roleID := range req.RoleIds {
			userRole := models.NewUserRole(user.ID, roleID)
			s.DB.Create(&userRole)
		}
	}

	// Generate tokens
	accessToken, err := helper.GenerateAccessToken(user.ID, nil, user.Username, user.IsValidated)
	if err != nil {
		return &pb.RegisterResponseV2{
			StatusCode: 500,
			Message:    "Failed to generate access token",
		}, nil
	}

	refreshToken, err := helper.GenerateRefreshToken(user.ID, nil, user.Username, user.IsValidated)
	if err != nil {
		return &pb.RegisterResponseV2{
			StatusCode: 500,
			Message:    "Failed to generate refresh token",
		}, nil
	}

	// Create user response
	pbUser := &pb.UserV2{
		Id:          user.ID,
		Username:    user.Username,
		Email:       "", // User model doesn't have email field
		FullName:    "", // User model doesn't have full_name field
		IsValidated: user.IsValidated,
		Status:      "", // Will be set below
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
	}

	// Set status if available
	if user.Status != nil {
		pbUser.Status = *user.Status
	}

	return &pb.RegisterResponseV2{
		StatusCode:   201,
		Message:      "User registered successfully",
		User:         pbUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUser retrieves a user by ID with optional role and permission inclusion
func (s *ServerV2) GetUser(ctx context.Context, req *pb.GetUserRequestV2) (*pb.GetUserResponseV2, error) {
	log.Printf("V2 GetUser request received for user ID: %s", req.Id)

	var user models.User
	if err := s.DB.Where("id = ?", req.Id).First(&user).Error; err != nil {
		return &pb.GetUserResponseV2{
			StatusCode: 404,
			Message:    "User not found",
		}, nil
	}

	// Create base user response
	pbUser := &pb.UserV2{
		Id:          user.ID,
		Username:    user.Username,
		Email:       "", // User model doesn't have email field
		FullName:    "", // User model doesn't have full_name field
		IsValidated: user.IsValidated,
		Status:      "", // Will be set below
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
	}

	// Set status if available
	if user.Status != nil {
		pbUser.Status = *user.Status
	}

	// Include roles if requested
	if req.IncludeRoles {
		var userRoles []models.UserRole
		if err := s.DB.Where("user_id = ?", user.ID).Find(&userRoles).Error; err == nil {
			var pbUserRoles []*pb.UserRoleV2
			for _, ur := range userRoles {
				pbUserRole := &pb.UserRoleV2{
					Id:        ur.ID,
					UserId:    ur.UserID,
					RoleId:    ur.RoleID,
					CreatedAt: ur.CreatedAt.Format(time.RFC3339),
					UpdatedAt: ur.UpdatedAt.Format(time.RFC3339),
				}
				pbUserRoles = append(pbUserRoles, pbUserRole)
			}
			pbUser.UserRoles = pbUserRoles
		}
	}

	// Include permissions if requested
	if req.IncludePermissions {
		// TODO: Implement permission fetching logic
		pbUser.Permissions = []string{}
	}

	return &pb.GetUserResponseV2{
		StatusCode: 200,
		Message:    "User retrieved successfully",
		User:       pbUser,
	}, nil
}

// GetAllUsers retrieves all users with pagination and filtering
func (s *ServerV2) GetAllUsers(ctx context.Context, req *pb.GetAllUsersRequestV2) (*pb.GetAllUsersResponseV2, error) {
	log.Printf("V2 GetAllUsers request received")

	query := s.DB.Model(&models.User{})

	// Apply search filter
	if req.Search != "" {
		query = query.Where("username ILIKE ?", fmt.Sprintf("%%%s%%", req.Search))
	}

	// Apply status filter
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Apply role filter
	if len(req.RoleIds) > 0 {
		query = query.Joins("JOIN user_roles ON users.id = user_roles.user_id").
			Where("user_roles.role_id IN ?", req.RoleIds)
	}

	// Get total count
	var totalCount int64
	query.Count(&totalCount)

	// Apply pagination
	offset := int((req.Page - 1) * req.PerPage)
	query = query.Offset(offset).Limit(int(req.PerPage))

	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return &pb.GetAllUsersResponseV2{
			StatusCode: 500,
			Message:    "Failed to retrieve users",
		}, nil
	}

	// Convert to protobuf users
	var pbUsers []*pb.UserV2
	for _, user := range users {
		pbUser := &pb.UserV2{
			Id:          user.ID,
			Username:    user.Username,
			Email:       "", // User model doesn't have email field
			FullName:    "", // User model doesn't have full_name field
			IsValidated: user.IsValidated,
			Status:      "", // Will be set below
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		}

		// Set status if available
		if user.Status != nil {
			pbUser.Status = *user.Status
		}

		pbUsers = append(pbUsers, pbUser)
	}

	return &pb.GetAllUsersResponseV2{
		StatusCode: 200,
		Message:    "Users retrieved successfully",
		Users:      pbUsers,
		TotalCount: int32(totalCount),
		Page:       req.Page,
		PerPage:    req.PerPage,
	}, nil
}

// UpdateUser updates a user with enhanced functionality
func (s *ServerV2) UpdateUser(ctx context.Context, req *pb.UpdateUserRequestV2) (*pb.UpdateUserResponseV2, error) {
	log.Printf("V2 UpdateUser request received for user ID: %s", req.Id)

	var user models.User
	if err := s.DB.Where("id = ?", req.Id).First(&user).Error; err != nil {
		return &pb.UpdateUserResponseV2{
			StatusCode: 404,
			Message:    "User not found",
		}, nil
	}

	// Update fields if provided
	if req.Username != "" {
		user.Username = req.Username
	}
	user.IsValidated = req.IsValidated
	if req.Status != "" {
		user.Status = &req.Status
	}

	if err := s.DB.Save(&user).Error; err != nil {
		return &pb.UpdateUserResponseV2{
			StatusCode: 500,
			Message:    "Failed to update user",
		}, nil
	}

	// Update roles if provided
	if len(req.RoleIds) > 0 {
		// Delete existing roles
		s.DB.Where("user_id = ?", user.ID).Delete(&models.UserRole{})

		// Add new roles
		for _, roleID := range req.RoleIds {
			userRole := models.NewUserRole(user.ID, roleID)
			s.DB.Create(&userRole)
		}
	}

	// Create updated user response
	pbUser := &pb.UserV2{
		Id:          user.ID,
		Username:    user.Username,
		Email:       "", // User model doesn't have email field
		FullName:    "", // User model doesn't have full_name field
		IsValidated: user.IsValidated,
		Status:      "", // Will be set below
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
	}

	// Set status if available
	if user.Status != nil {
		pbUser.Status = *user.Status
	}

	return &pb.UpdateUserResponseV2{
		StatusCode: 200,
		Message:    "User updated successfully",
		User:       pbUser,
	}, nil
}

// DeleteUser deletes a user
func (s *ServerV2) DeleteUser(ctx context.Context, req *pb.DeleteUserRequestV2) (*pb.DeleteUserResponseV2, error) {
	log.Printf("V2 DeleteUser request received for user ID: %s", req.Id)

	// Delete user roles first
	if err := s.DB.Where("user_id = ?", req.Id).Delete(&models.UserRole{}).Error; err != nil {
		return &pb.DeleteUserResponseV2{
			StatusCode: 500,
			Message:    "Failed to delete user roles",
		}, nil
	}

	// Delete user
	if err := s.DB.Where("id = ?", req.Id).Delete(&models.User{}).Error; err != nil {
		return &pb.DeleteUserResponseV2{
			StatusCode: 500,
			Message:    "Failed to delete user",
		}, nil
	}

	return &pb.DeleteUserResponseV2{
		StatusCode: 200,
		Message:    "User deleted successfully",
	}, nil
}

// RefreshToken handles token refresh
func (s *ServerV2) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequestV2) (*pb.RefreshTokenResponseV2, error) {
	log.Printf("V2 RefreshToken request received")

	// TODO: Implement token refresh logic
	// For now, return a placeholder response
	return &pb.RefreshTokenResponseV2{
		StatusCode:   200,
		Message:      "Token refreshed successfully",
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		ExpiresIn:    3600,
	}, nil
}

// Logout handles user logout
func (s *ServerV2) Logout(ctx context.Context, req *pb.LogoutRequestV2) (*pb.LogoutResponseV2, error) {
	log.Printf("V2 Logout request received")

	// TODO: Implement token invalidation logic
	// For now, return a placeholder response
	return &pb.LogoutResponseV2{
		StatusCode: 200,
		Message:    "Logout successful",
	}, nil
}
