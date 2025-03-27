package user

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/kisanlink/protobuf/pb-aaa"
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
		helper.LowerCaseSlice(roles), 
		helper.LowerCaseSlice(permissions),
		helper.LowerCaseSlice(actions),
	)
		if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete user relationship")

	}
	log.Printf("delete Relation  Response: %+v", updated)
	return &pb.DeleteUserResponse{
		StatusCode: http.StatusOK,
		Success: true,
		Message:    "User deleted successfully",
	}, nil
}
