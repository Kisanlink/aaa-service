package user

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AssignRole(ctx context.Context, req *pb.AssignRoleToUserRequest) (*pb.AssignRoleToUserResponse, error) {
	// Validate user exists
	_, err := s.userService.GetUserByID(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User with ID %s not found", req.UserId)
	}

	// Validate role exists
	roleName := req.GetRole()
	role, err := s.roleService.GetRoleByName(roleName)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role with name %s not found", roleName)
	}

	// Create user-role relationship
	userRole := model.UserRole{
		UserID:   req.UserId,
		RoleID:   role.ID,
		IsActive: true,
	}

	if err := s.userService.CreateUserRoles(userRole); err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.AlreadyExists {
			return nil, status.Errorf(codes.AlreadyExists, "user already has role '%s' assigned", roleName)
		}
		return nil, status.Errorf(codes.Internal, "failed to assign role to user: %v", err)
	}
	// Get updated user details
	updatedUser, err := s.userService.GetUserByID(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch user details: %v", err)
	}

	// Build response
	return &pb.AssignRoleToUserResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role assigned to user successfully",
		DataTimeStamp: time.Now().Format(time.RFC3339),
		Data: &pb.AssignRolePermission{
			Id:          updatedUser.ID,
			Username:    updatedUser.Username,
			Password:    "", // Explicitly empty for security
			IsValidated: updatedUser.IsValidated,
		},
	}, nil
}
