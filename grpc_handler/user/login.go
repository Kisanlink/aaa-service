package user

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/kisanlink/protobuf/pb-aaa"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validate inputs
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "Username is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}

	if !helper.IsValidUsername(req.Username) {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"username '%s' contains invalid characters",
			req.Username,
		)
	}

	// Validate user existence
	existingUser, err := s.userService.FindUserByUsername(req.Username)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid password")
	}

	// Set auth headers
	if err := helper.SetAuthHeadersWithTokens(
		ctx,
		existingUser.ID,
		existingUser.Username,
		existingUser.IsValidated,
	); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set auth headers: %v", err)
	}

	// Get updated user details with roles and permissions
	rolesResponse, err := s.userService.GetUserRolesWithPermissions(existingUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch user roles: %v", err)
	}
	// Build the response
	response := &pb.LoginResponse{
		StatusCode:    http.StatusOK,
		Message:       "Login successful",
		Success:       true,
		DataTimeStamp: time.Now().Format(time.RFC3339),
		Data: &pb.AssignRolePermission{
			Id:          existingUser.ID,
			Username:    existingUser.Username,
			Password:    "", // Explicitly empty for security
			IsValidated: existingUser.IsValidated,
			CreatedAt:   existingUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   existingUser.UpdatedAt.Format(time.RFC3339Nano),
			Roles:       convertRoleResponseToPB(rolesResponse),
		},
	}

	return response, nil
}
