package grpc_server

import (
	"context"

	pb "github.com/Kisanlink/aaa-service/proto"
	"github.com/Kisanlink/aaa-service/services"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthHandler implements the UserServiceV2 gRPC service for authentication
type AuthHandler struct {
	pb.UnimplementedUserServiceV2Server
	authService *services.AuthService
	logger      *zap.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Login authenticates a user and returns JWT tokens
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequestV2) (*pb.LoginResponseV2, error) {
	h.logger.Info("gRPC Login request", zap.String("username", req.Username))

	// Convert gRPC request to service request
	loginReq := &services.LoginRequest{
		Username: req.Username,
		Password: req.Password,
		MFACode:  req.MfaCode,
	}

	// Call authentication service
	response, err := h.authService.Login(ctx, loginReq)
	if err != nil {
		h.logger.Error("Login failed", zap.String("username", req.Username), zap.Error(err))
		return &pb.LoginResponseV2{
			StatusCode: 401,
			Message:    "Authentication failed",
		}, status.Error(codes.Unauthenticated, err.Error())
	}

	// Convert service response to gRPC response
	grpcResponse := &pb.LoginResponseV2{
		StatusCode:   200,
		Message:      "Login successful",
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		TokenType:    response.TokenType,
		ExpiresIn:    int32(response.ExpiresIn),
		User:         convertUserToGRPC(response),
		Permissions:  response.Permissions,
	}

	h.logger.Info("Login successful", zap.String("user_id", response.User.ID))
	return grpcResponse, nil
}

// Register creates a new user account
func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequestV2) (*pb.RegisterResponseV2, error) {
	h.logger.Info("gRPC Register request", zap.String("username", req.Username))

	// Convert gRPC request to service request
	registerReq := &services.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
		RoleIDs:  req.RoleIds,
	}

	// Call authentication service
	response, err := h.authService.Register(ctx, registerReq)
	if err != nil {
		h.logger.Error("Registration failed", zap.String("username", req.Username), zap.Error(err))
		return &pb.RegisterResponseV2{
			StatusCode: 400,
			Message:    "Registration failed",
		}, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert service response to gRPC response
	grpcResponse := &pb.RegisterResponseV2{
		StatusCode:   201,
		Message:      "Registration successful",
		User:         convertUserToGRPC(response),
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	}

	h.logger.Info("Registration successful", zap.String("user_id", response.User.ID))
	return grpcResponse, nil
}

// RefreshToken refreshes an access token
func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequestV2) (*pb.RefreshTokenResponseV2, error) {
	h.logger.Info("gRPC RefreshToken request")

	// Call authentication service
	response, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		h.logger.Error("Token refresh failed", zap.Error(err))
		return &pb.RefreshTokenResponseV2{
			StatusCode: 401,
			Message:    "Token refresh failed",
		}, status.Error(codes.Unauthenticated, err.Error())
	}

	// Convert service response to gRPC response
	grpcResponse := &pb.RefreshTokenResponseV2{
		StatusCode:   200,
		Message:      "Token refreshed successfully",
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresIn:    int32(response.ExpiresIn),
	}

	h.logger.Info("Token refresh successful")
	return grpcResponse, nil
}

// Logout invalidates user tokens
func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequestV2) (*pb.LogoutResponseV2, error) {
	h.logger.Info("gRPC Logout request")

	// Extract user ID from token
	claims, err := h.authService.ValidateToken(req.AccessToken)
	if err != nil {
		h.logger.Error("Invalid token for logout", zap.Error(err))
		return &pb.LogoutResponseV2{
			StatusCode: 401,
			Message:    "Invalid token",
		}, status.Error(codes.Unauthenticated, "Invalid token")
	}

	// Call authentication service
	err = h.authService.Logout(ctx, claims.UserID)
	if err != nil {
		h.logger.Error("Logout failed", zap.String("user_id", claims.UserID), zap.Error(err))
		return &pb.LogoutResponseV2{
			StatusCode: 500,
			Message:    "Logout failed",
		}, status.Error(codes.Internal, err.Error())
	}

	grpcResponse := &pb.LogoutResponseV2{
		StatusCode: 200,
		Message:    "Logout successful",
	}

	h.logger.Info("Logout successful", zap.String("user_id", claims.UserID))
	return grpcResponse, nil
}

// GetUser retrieves user information (placeholder implementation)
func (h *AuthHandler) GetUser(ctx context.Context, req *pb.GetUserRequestV2) (*pb.GetUserResponseV2, error) {
	h.logger.Info("gRPC GetUser request", zap.String("user_id", req.Id))

	// This would typically call a user service
	// For now, return a placeholder response
	return &pb.GetUserResponseV2{
		StatusCode: 501,
		Message:    "GetUser not implemented yet",
	}, status.Error(codes.Unimplemented, "GetUser not implemented yet")
}

// GetAllUsers retrieves all users (placeholder implementation)
func (h *AuthHandler) GetAllUsers(ctx context.Context, req *pb.GetAllUsersRequestV2) (*pb.GetAllUsersResponseV2, error) {
	h.logger.Info("gRPC GetAllUsers request")

	// This would typically call a user service
	// For now, return a placeholder response
	return &pb.GetAllUsersResponseV2{
		StatusCode: 501,
		Message:    "GetAllUsers not implemented yet",
	}, status.Error(codes.Unimplemented, "GetAllUsers not implemented yet")
}

// UpdateUser updates user information (placeholder implementation)
func (h *AuthHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequestV2) (*pb.UpdateUserResponseV2, error) {
	h.logger.Info("gRPC UpdateUser request", zap.String("user_id", req.Id))

	// This would typically call a user service
	// For now, return a placeholder response
	return &pb.UpdateUserResponseV2{
		StatusCode: 501,
		Message:    "UpdateUser not implemented yet",
	}, status.Error(codes.Unimplemented, "UpdateUser not implemented yet")
}

// DeleteUser deletes a user (placeholder implementation)
func (h *AuthHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequestV2) (*pb.DeleteUserResponseV2, error) {
	h.logger.Info("gRPC DeleteUser request", zap.String("user_id", req.Id))

	// This would typically call a user service
	// For now, return a placeholder response
	return &pb.DeleteUserResponseV2{
		StatusCode: 501,
		Message:    "DeleteUser not implemented yet",
	}, status.Error(codes.Unimplemented, "DeleteUser not implemented yet")
}

// Helper function to convert LoginResponse to gRPC UserV2
func convertUserToGRPC(response *services.LoginResponse) *pb.UserV2 {
	if response == nil {
		return nil
	}

	user := &response.User
	userRoles := make([]*pb.UserRoleV2, len(user.Roles))
	for i, role := range user.Roles {
		userRoles[i] = &pb.UserRoleV2{
			Id:     role.ID,
			UserId: role.UserID,
			RoleId: role.RoleID,
			// Add other fields as needed
		}
	}

	status := ""
	if user.Status != nil {
		status = *user.Status
	}

	return &pb.UserV2{
		Id:          user.ID,
		Username:    user.Username,
		IsValidated: user.IsValidated,
		Status:      status,
		UserRoles:   userRoles,
		Permissions: response.Permissions,
	}
}
