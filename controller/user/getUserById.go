package user

import (
	"context"
	"time"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	user, err := s.UserRepo.FindExistingUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
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
	return &pb.GetUserByIdResponse{
		StatusCode: int32(codes.OK),
		Message:    "User fetched successfully",
		User:       pbUser,
	}, nil
}
