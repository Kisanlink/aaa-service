package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/entities/models"
	pb "github.com/Kisanlink/aaa-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	// Validate input
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	// Fetch the user from the database
	var user models.User
	err := s.DB.Table("users").Where("id = ?", id).First(&user).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch user: %v", err))
	}

	// Fetch associated roles for the user
	var userRoles []models.UserRole
	err = s.DB.Table("user_roles").Where("user_id = ?", user.ID).Find(&userRoles).Error
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch roles for user %s: %v", user.ID, err))
	}

	// Convert UserRole models to protobuf UserRole messages
	var pbUserRoles []*pb.UserRole
	for _, role := range userRoles {
		pbUserRoles = append(pbUserRoles, &pb.UserRole{
			Id:               role.ID,
			UserId:           role.UserID,
			RolePermissionId: role.RoleID,
			CreatedAt:        role.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:        role.UpdatedAt.Format(time.RFC3339Nano),
		})
	}

	// Prepare the protobuf User message
	pbUser := &pb.User{
		Id:          user.ID,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339Nano),
		Username:    user.Username,
		IsValidated: user.IsValidated,
		UserRoles:   pbUserRoles, // Include associated roles
	}

	// Return the response
	return &pb.GetUserByIdResponse{
		StatusCode: int32(codes.OK),
		Message:    "User fetched successfully",
		User:       pbUser,
	}, nil
}
