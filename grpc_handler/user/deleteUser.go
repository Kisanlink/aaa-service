package user

import (
	"context"
	"net/http"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	// existingUser, err := s.userService.FindExistingUserByID(id)
	// if err != nil {
	// 	return nil, err
	// }
	if err := s.userService.DeleteUserRoles(id); err != nil {
		return nil, err
	}
	if err := s.userService.DeleteUser(id); err != nil {
		return nil, err
	}

	return &pb.DeleteUserResponse{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "User deleted successfully",
	}, nil
}
