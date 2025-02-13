package user

import (
	"context"
	"time"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
)

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	users, err := s.UserRepo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	var pbUsers []*pb.User
	for _, user := range users {
		userRoles, err := s.UserRepo.FindUserRoles(ctx, user.ID)
		if err != nil {
			return nil, err
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
