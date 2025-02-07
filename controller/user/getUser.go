package user

import (
	"context"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetUsers(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	var users []model.User
	if err := s.DB.Find(&users).Error; err != nil {
		return nil, status.Error(codes.Internal, "Failed to fetch users")
	}
	pbUsers := []*pb.User{}
	for _, user := range users {
		pbUser := &pb.User{
			Id:          user.ID.String(),
			CreatedAt:   user.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339Nano),
			Username:    user.Username,
			IsValidated: user.IsValidated,
		}
		pbUsers = append(pbUsers, pbUser)
	}

	return &pb.GetUserResponse{
		Users: pbUsers,
	}, nil
}
