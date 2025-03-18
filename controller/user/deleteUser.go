package user

import (
	"context"
	"log"
	"strings"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	existingUser, err := s.UserRepo.FindExistingUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.UserRepo.DeleteUserRoles(ctx, id); err != nil {
		return nil, err
	}
	if err := s.UserRepo.DeleteUser(ctx, id); err != nil {
		return nil, err
	}
	roles, permissions,actions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, existingUser.ID)
    if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch user roles and permissions")

    }
	updated, err := client.DeleteUserRoleRelationship(
		strings.ToLower(existingUser.Username), 
		LowerCaseSlice(roles), 
		LowerCaseSlice(permissions),
		LowerCaseSlice(actions),
	)
		if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete user relationship")

	}
	log.Printf("delete Relation  Response: %+v", updated)
	return &pb.DeleteUserResponse{
		StatusCode: int32(codes.OK),
		Message:    "User deleted successfully",
	}, nil
}
