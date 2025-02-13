// package user

// import (
// 	"context"
// 	"fmt"

// 	"github.com/Kisanlink/aaa-service/model"
// 	"github.com/Kisanlink/aaa-service/pb"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
// 	id := req.GetId()
// 	if id == "" {
// 		return nil, status.Error(codes.InvalidArgument, "ID is required")
// 	}

// 	var existingUser model.User
// 	err := s.DB.Table("users").Where("id = ?", id).First(&existingUser).Error
// 	if err != nil {
// 		if err.Error() == "record not found" {
// 			return nil, status.Error(codes.NotFound, "User not found")
// 		}
// 		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch user: %v", err))
// 	}
// 	if err := s.DB.Table("user_roles").Where("user_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
// 		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to delete user roles: %v", err))
// 	}

// 	if err := s.DB.Table("users").Delete(&model.User{}, "id = ?", id).Error; err != nil {
// 		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to delete user: %v", err))
// 	}

//		return &pb.DeleteUserResponse{
//			StatusCode: int32(codes.OK),
//			Message:    "User deleted successfully",
//		}, nil
//	}
package user

import (
	"context"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	_, err := s.UserRepo.FindExistingUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.UserRepo.DeleteUserRoles(ctx, id); err != nil {
		return nil, err
	}
	if err := s.UserRepo.DeleteUser(ctx, id); err != nil {
		return nil, err
	}
	return &pb.DeleteUserResponse{
		StatusCode: int32(codes.OK),
		Message:    "User deleted successfully",
	}, nil
}
