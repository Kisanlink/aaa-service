package grpc_server

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests/users"
	userResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/services"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CombinedUserHandler implements the UserService gRPC service with both auth and user methods
type CombinedUserHandler struct {
	pb.UnimplementedUserServiceServer
	authService *services.AuthService
	userService interfaces.UserService
	logger      *zap.Logger
}

// NewCombinedUserHandler creates a new combined handler
func NewCombinedUserHandler(authService *services.AuthService, userService interfaces.UserService, logger *zap.Logger) *CombinedUserHandler {
	return &CombinedUserHandler{
		authService: authService,
		userService: userService,
		logger:      logger,
	}
}

// Login authenticates a user and returns JWT tokens
func (h *CombinedUserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	h.logger.Info("gRPC Login request", zap.String("username", req.Username))

	// Convert gRPC request to service request
	loginReq := &services.UsernameLoginRequest{
		Username: req.Username,
		Password: req.Password,
		MFACode:  req.MfaCode,
	}

	// Call username-based authentication service
	response, err := h.authService.LoginWithUsername(ctx, loginReq)
	if err != nil {
		h.logger.Error("Login failed", zap.String("username", req.Username), zap.Error(err))
		return &pb.LoginResponse{
			StatusCode: 401,
			Message:    "Authentication failed",
		}, status.Error(codes.Unauthenticated, err.Error())
	}

	// Convert service response to gRPC response
	grpcResponse := &pb.LoginResponse{
		StatusCode:   200,
		Message:      "Login successful",
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int32(response.ExpiresIn),
		Permissions:  response.Permissions,
	}

	// Add user information if available
	if response.User.ID != "" {
		grpcResponse.User = &pb.User{
			Id:          response.User.ID,
			Username:    getStringValue(response.User.Username),
			PhoneNumber: response.User.PhoneNumber,
			CountryCode: response.User.CountryCode,
			IsValidated: response.User.IsValidated,
			CreatedAt:   timestamppb.New(response.User.CreatedAt).String(),
			UpdatedAt:   timestamppb.New(response.User.UpdatedAt).String(),
		}
	}

	return grpcResponse, nil
}

// Register creates a new user
func (h *CombinedUserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	h.logger.Info("gRPC Register request", zap.String("username", req.Username))

	// Validate request - phone_number, country_code, and password are required
	if req.PhoneNumber == "" {
		return &pb.RegisterResponse{
			StatusCode: 400,
			Message:    "Phone number is required",
		}, status.Error(codes.InvalidArgument, "phone number is required")
	}
	if req.CountryCode == "" {
		return &pb.RegisterResponse{
			StatusCode: 400,
			Message:    "Country code is required",
		}, status.Error(codes.InvalidArgument, "country code is required")
	}
	if req.Password == "" {
		return &pb.RegisterResponse{
			StatusCode: 400,
			Message:    "Password is required",
		}, status.Error(codes.InvalidArgument, "password is required")
	}

	// Create user request
	var username *string
	if req.Username != "" {
		username = &req.Username
	}
	createReq := &users.CreateUserRequest{
		Username:    username,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
		CountryCode: req.CountryCode,
	}

	// Create user via service
	userResponse, err := h.userService.CreateUser(ctx, createReq)
	if err != nil {
		h.logger.Error("Failed to create user", zap.String("username", req.Username), zap.Error(err))

		// Check if it's a conflict error
		if err.Error() == "user with this phone number already exists" ||
			err.Error() == "username is already taken" {
			return &pb.RegisterResponse{
				StatusCode: 409,
				Message:    "User already exists",
			}, status.Error(codes.AlreadyExists, err.Error())
		}

		return &pb.RegisterResponse{
			StatusCode: 500,
			Message:    "Internal server error",
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf user
	pbUser := &pb.User{
		Id:          userResponse.ID,
		Username:    getStringValue(userResponse.Username),
		PhoneNumber: userResponse.PhoneNumber,
		CountryCode: userResponse.CountryCode,
		IsValidated: userResponse.IsValidated,
		CreatedAt:   timestamppb.New(userResponse.CreatedAt).String(),
		UpdatedAt:   timestamppb.New(userResponse.UpdatedAt).String(),
	}

	return &pb.RegisterResponse{
		StatusCode: 201,
		Message:    "User created successfully",
		User:       pbUser,
	}, nil
}

// GetUser retrieves a user by ID
func (h *CombinedUserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	h.logger.Info("gRPC GetUser request", zap.String("user_id", req.Id))

	// Validate request
	if req.Id == "" {
		return &pb.GetUserResponse{
			StatusCode: 400,
			Message:    "User ID is required",
		}, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Get user from service
	userResponse, err := h.userService.GetUserByID(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get user", zap.String("user_id", req.Id), zap.Error(err))

		// Check if it's a not found error
		if err.Error() == "user not found" {
			return &pb.GetUserResponse{
				StatusCode: 404,
				Message:    "User not found",
			}, status.Error(codes.NotFound, "user not found")
		}

		return &pb.GetUserResponse{
			StatusCode: 500,
			Message:    "Internal server error",
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf user
	pbUser := &pb.User{
		Id:          userResponse.ID,
		Username:    getStringValue(userResponse.Username),
		PhoneNumber: userResponse.PhoneNumber,
		CountryCode: userResponse.CountryCode,
		IsValidated: userResponse.IsValidated,
		CreatedAt:   timestamppb.New(userResponse.CreatedAt).String(),
		UpdatedAt:   timestamppb.New(userResponse.UpdatedAt).String(),
	}

	return &pb.GetUserResponse{
		StatusCode: 200,
		Message:    "User retrieved successfully",
		User:       pbUser,
	}, nil
}

// GetAllUsers retrieves all users with pagination
func (h *CombinedUserHandler) GetAllUsers(ctx context.Context, req *pb.GetAllUsersRequest) (*pb.GetAllUsersResponse, error) {
	h.logger.Info("gRPC GetAllUsers request")

	// Set default pagination values
	page := int(req.Page)
	perPage := int(req.PerPage)
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}

	// Calculate offset
	offset := (page - 1) * perPage

	// Get users from service
	var usersInterface interface{}
	var err error

	if req.Status == "active" {
		usersInterface, err = h.userService.ListActiveUsers(ctx, perPage, offset)
	} else if req.Search != "" {
		usersInterface, err = h.userService.SearchUsers(ctx, req.Search, perPage, offset)
	} else {
		usersInterface, err = h.userService.ListUsers(ctx, perPage, offset)
	}

	if err != nil {
		h.logger.Error("Failed to get users", zap.Error(err))
		return &pb.GetAllUsersResponse{
			StatusCode: 500,
			Message:    "Internal server error",
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert interface{} to typed user response slice
	var pbUsers []*pb.User
	var totalCount int32

	// Type assert to the expected response type
	if userList, ok := usersInterface.([]*userResponses.UserResponse); ok {
		pbUsers = make([]*pb.User, len(userList))
		for i, userResp := range userList {
			pbUsers[i] = &pb.User{
				Id:          userResp.ID,
				Username:    getStringValue(userResp.Username),
				PhoneNumber: userResp.PhoneNumber,
				CountryCode: userResp.CountryCode,
				IsValidated: userResp.IsValidated,
				CreatedAt:   timestamppb.New(userResp.CreatedAt).String(),
				UpdatedAt:   timestamppb.New(userResp.UpdatedAt).String(),
			}
		}
		totalCount = int32(len(userList))
	} else {
		h.logger.Warn("Unexpected type from user service", zap.Any("type", usersInterface))
	}

	return &pb.GetAllUsersResponse{
		StatusCode: 200,
		Message:    "Users retrieved successfully",
		Users:      pbUsers,
		TotalCount: totalCount,
		Page:       int32(page),
		PerPage:    int32(perPage),
	}, nil
}

// GetUserByPhone retrieves a user by phone number
func (h *CombinedUserHandler) GetUserByPhone(ctx context.Context, req *pb.GetUserByPhoneRequest) (*pb.GetUserResponse, error) {
	h.logger.Info("gRPC GetUserByPhone request", zap.String("phone", req.PhoneNumber))

	// Validate request
	if req.PhoneNumber == "" {
		return &pb.GetUserResponse{
			StatusCode: 400,
			Message:    "Phone number is required",
		}, status.Error(codes.InvalidArgument, "phone number is required")
	}

	countryCode := req.CountryCode
	if countryCode == "" {
		countryCode = "+91" // Default to India
	}

	// Get user by phone number
	userResponse, err := h.userService.GetUserByPhoneNumber(ctx, req.PhoneNumber, countryCode)
	if err != nil {
		h.logger.Error("Failed to get user by phone", zap.String("phone", req.PhoneNumber), zap.Error(err))

		if err.Error() == "user not found" {
			return &pb.GetUserResponse{
				StatusCode: 404,
				Message:    "User not found",
			}, status.Error(codes.NotFound, "user not found")
		}

		return &pb.GetUserResponse{
			StatusCode: 500,
			Message:    "Internal server error",
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf user
	pbUser := &pb.User{
		Id:          userResponse.ID,
		Username:    getStringValue(userResponse.Username),
		PhoneNumber: userResponse.PhoneNumber,
		CountryCode: userResponse.CountryCode,
		IsValidated: userResponse.IsValidated,
		CreatedAt:   timestamppb.New(userResponse.CreatedAt).String(),
		UpdatedAt:   timestamppb.New(userResponse.UpdatedAt).String(),
	}

	return &pb.GetUserResponse{
		StatusCode: 200,
		Message:    "User retrieved successfully",
		User:       pbUser,
	}, nil
}

// VerifyUserPassword verifies a user's password
func (h *CombinedUserHandler) VerifyUserPassword(ctx context.Context, req *pb.VerifyPasswordRequest) (*pb.VerifyPasswordResponse, error) {
	h.logger.Info("gRPC VerifyUserPassword request", zap.String("username", req.Username))

	// Validate request
	if req.Username == "" || req.Password == "" {
		return &pb.VerifyPasswordResponse{
			StatusCode: 400,
			Message:    "Username and password are required",
			Valid:      false,
		}, status.Error(codes.InvalidArgument, "username and password are required")
	}

	// Verify password
	userResponse, err := h.userService.VerifyUserPassword(ctx, req.Username, req.Password)
	if err != nil {
		h.logger.Error("Password verification failed", zap.String("username", req.Username), zap.Error(err))

		return &pb.VerifyPasswordResponse{
			StatusCode: 401,
			Message:    "Invalid credentials",
			Valid:      false,
		}, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Convert to protobuf user
	pbUser := &pb.User{
		Id:          userResponse.ID,
		Username:    getStringValue(userResponse.Username),
		PhoneNumber: userResponse.PhoneNumber,
		CountryCode: userResponse.CountryCode,
		IsValidated: userResponse.IsValidated,
		CreatedAt:   timestamppb.New(userResponse.CreatedAt).String(),
		UpdatedAt:   timestamppb.New(userResponse.UpdatedAt).String(),
	}

	return &pb.VerifyPasswordResponse{
		StatusCode: 200,
		Message:    "Password verified successfully",
		Valid:      true,
		User:       pbUser,
	}, nil
}

// UpdateUser updates a user
func (h *CombinedUserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	h.logger.Info("gRPC UpdateUser request", zap.String("user_id", req.Id))

	// Validate request
	if req.Id == "" {
		return &pb.UpdateUserResponse{
			StatusCode: 400,
			Message:    "User ID is required",
		}, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Build update request
	updateReq := &users.UpdateUserRequest{
		UserID: req.Id,
	}

	// Map optional fields from protobuf to service request
	if req.Username != "" {
		updateReq.Name = &req.Username
	}
	if req.Email != "" {
		updateReq.EmailHash = &req.Email
	}
	if req.FullName != "" {
		updateReq.Name = &req.FullName
	}
	if req.Status != "" {
		updateReq.Status = &req.Status
	}

	// Set user ID in context for service layer
	type userIDKey struct{}
	ctxWithUser := context.WithValue(ctx, userIDKey{}, req.Id)

	// Update user via service
	userResponse, err := h.userService.UpdateUser(ctxWithUser, updateReq)
	if err != nil {
		h.logger.Error("Failed to update user", zap.String("user_id", req.Id), zap.Error(err))

		if err.Error() == "user not found" {
			return &pb.UpdateUserResponse{
				StatusCode: 404,
				Message:    "User not found",
			}, status.Error(codes.NotFound, "user not found")
		}

		return &pb.UpdateUserResponse{
			StatusCode: 500,
			Message:    "Internal server error",
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf user
	pbUser := &pb.User{
		Id:          userResponse.ID,
		Username:    getStringValue(userResponse.Username),
		PhoneNumber: userResponse.PhoneNumber,
		CountryCode: userResponse.CountryCode,
		IsValidated: userResponse.IsValidated,
		CreatedAt:   timestamppb.New(userResponse.CreatedAt).String(),
		UpdatedAt:   timestamppb.New(userResponse.UpdatedAt).String(),
	}

	return &pb.UpdateUserResponse{
		StatusCode: 200,
		Message:    "User updated successfully",
		User:       pbUser,
	}, nil
}

// DeleteUser deletes a user
func (h *CombinedUserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	h.logger.Info("gRPC DeleteUser request", zap.String("user_id", req.Id))

	// Validate request
	if req.Id == "" {
		return &pb.DeleteUserResponse{
			StatusCode: 400,
			Message:    "User ID is required",
		}, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Delete user via service
	err := h.userService.DeleteUser(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.String("user_id", req.Id), zap.Error(err))

		if err.Error() == "user not found" {
			return &pb.DeleteUserResponse{
				StatusCode: 404,
				Message:    "User not found",
			}, status.Error(codes.NotFound, "user not found")
		}

		return &pb.DeleteUserResponse{
			StatusCode: 500,
			Message:    "Internal server error",
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteUserResponse{
		StatusCode: 200,
		Message:    "User deleted successfully",
	}, nil
}

// RefreshToken refreshes an access token
func (h *CombinedUserHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	h.logger.Info("gRPC RefreshToken request")

	// Validate request
	if req.RefreshToken == "" {
		return &pb.RefreshTokenResponse{
			StatusCode: 400,
			Message:    "Refresh token is required",
		}, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	// Call auth service to refresh token
	// Note: The gRPC proto doesn't include mPin, so we pass empty string
	// This is a limitation that should be addressed in the proto definition
	response, err := h.authService.RefreshToken(ctx, req.RefreshToken, "")
	if err != nil {
		h.logger.Error("Token refresh failed", zap.Error(err))
		return &pb.RefreshTokenResponse{
			StatusCode: 401,
			Message:    "Token refresh failed",
		}, status.Error(codes.Unauthenticated, err.Error())
	}

	// Convert service response to gRPC response
	grpcResponse := &pb.RefreshTokenResponse{
		StatusCode:   200,
		Message:      "Token refreshed successfully",
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresIn:    int32(response.ExpiresIn),
	}

	h.logger.Info("Token refresh successful")
	return grpcResponse, nil
}

// Logout logs out a user
func (h *CombinedUserHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	h.logger.Info("gRPC Logout request")

	// Validate request
	if req.AccessToken == "" {
		return &pb.LogoutResponse{
			StatusCode: 400,
			Message:    "Access token is required",
		}, status.Error(codes.InvalidArgument, "access token is required")
	}

	// Extract user ID from token
	claims, err := h.authService.ValidateToken(req.AccessToken)
	if err != nil {
		h.logger.Error("Invalid token for logout", zap.Error(err))
		return &pb.LogoutResponse{
			StatusCode: 401,
			Message:    "Invalid token",
		}, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Extract user ID from claims
	userID := claims.UserID
	if userID == "" {
		h.logger.Error("Could not extract user ID from token")
		return &pb.LogoutResponse{
			StatusCode: 401,
			Message:    "Invalid token claims",
		}, status.Error(codes.Unauthenticated, "invalid token claims")
	}

	// Call authentication service logout
	err = h.authService.Logout(ctx, userID)
	if err != nil {
		h.logger.Error("Logout failed", zap.String("user_id", userID), zap.Error(err))
		return &pb.LogoutResponse{
			StatusCode: 500,
			Message:    "Logout failed",
		}, status.Error(codes.Internal, err.Error())
	}

	grpcResponse := &pb.LogoutResponse{
		StatusCode: 200,
		Message:    "Logout successful",
	}

	h.logger.Info("Logout successful", zap.String("user_id", userID))
	return grpcResponse, nil
}

// Helper function to safely get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
