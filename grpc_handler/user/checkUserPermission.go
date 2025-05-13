package user

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CheckUserPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	createdUser, err := s.userService.GetUserByID(req.Principal)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user details")
	}

	roles, permissions, _, err := s.userService.FindUserRolesAndPermissions(createdUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions")
	}

	results, err := client.CheckUserPermissions(
		createdUser.Username,
		helper.LowerCaseSlice(roles),
		helper.LowerCaseSlice(permissions),
		helper.LowerCaseSlice(req.Actions),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check permissions")
	}
	permissionMap := make(map[string]bool)
	actionMap := make(map[string]bool)

	for permission, hasPermission := range results["role_permissions"] {
		permissionMap[permission] = hasPermission
	}
	for action, hasPermission := range results["user_actions"] {
		if !hasPermission {
			return nil, status.Errorf(codes.PermissionDenied, "User doesn't have this action: %s", action)
		}
		actionMap[action] = hasPermission
	}

	response := &pb.CheckPermissionResponse{
		StatusCode:  http.StatusOK,
		Success:     true,
		Message:     "Permissions checked successfully",
		Permissions: permissionMap,
		Actions:     actionMap,
	}

	return response, nil
}
