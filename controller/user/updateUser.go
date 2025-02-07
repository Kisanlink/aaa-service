package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	var existingUser model.User
	err := s.DB.Table("users").Where("id = ?", id).First(&existingUser).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch user: %v", err))
	}
	if req.Username != "" {
		existingUser.Username = req.Username
	}
	// if req.IsValidated != nil {
	// 	existingUser.IsValidated = req.GetIsValidated()
	// }

	if err := s.DB.Table("users").Save(&existingUser).Error; err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to update user: %v", err))
	}

	pbUser := &pb.User{
		Id:          existingUser.ID,
		CreatedAt:   existingUser.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   existingUser.UpdatedAt.Format(time.RFC3339Nano),
		Username:    existingUser.Username,
		IsValidated: existingUser.IsValidated,
	}

	return &pb.UpdateUserResponse{
		User: pbUser,
	}, nil
}
