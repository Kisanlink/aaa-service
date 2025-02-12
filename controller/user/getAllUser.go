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

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	var users []model.User
	err := s.DB.Table("users").Find(&users).Error
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch users: %v", err))
	}

	var pbUsers []*pb.User
	for _, user := range users {
		var userRoles []model.UserRole
		err := s.DB.Table("user_roles").Where("user_id = ?", user.ID).Find(&userRoles).Error
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch roles for user %s: %v", user.ID, err))
		}

		var pbUserRoles []*pb.UserRole
		for _, role := range userRoles {
			pbUserRoles = append(pbUserRoles, &pb.UserRole{
				Id:               role.ID,
				UserId:           role.UserID,
				RolePermissionId: role.RolePermissionID,
				CreatedAt:        role.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:        role.UpdatedAt.Format(time.RFC3339Nano),
			})
		}

		pbUser := &pb.User{
			Id:          user.ID,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339Nano),
			Username:    user.Username,
			IsValidated: user.IsValidated,
			UserRoles:   pbUserRoles,
		}

		pbUsers = append(pbUsers, pbUser)
	}

	return &pb.GetUserResponse{
		StatusCode: int32(codes.OK),
		Message:    "Users fetched successfully",
		Users:      pbUsers,
	}, nil
}
