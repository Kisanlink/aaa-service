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

func (s *Server) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	var user model.User
	err := s.DB.Table("users").Where("id = ?", id).First(&user).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch user: %v", err))
	}

	pbUser := &pb.User{
		Id:          user.ID.String(),
		CreatedAt:   user.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339Nano),
		Username:    user.Username,
		IsValidated: user.IsValidated,
	}

	// Return the response
	return &pb.GetUserByIdResponse{
		User: pbUser,
	}, nil
}
