package user

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AssignRole(ctx context.Context, req *pb.AssignRoleToUserRequest) (*pb.AssignRoleToUserResponse, error) {
	_, err := s.UserRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User with ID %s not found", req.UserId)
	}
	roleName := req.GetRole()
	role, err := s.RoleRepo.GetRoleByName(ctx, roleName)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role with name %s not found", roleName)
	}
	userRole := model.UserRole{
		UserID:   req.UserId,
		RoleID:   role.ID,
		IsActive: true,
	}

	if err := s.UserRepo.CreateUserRoles(ctx, userRole); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create user-role connection: %v", err)
	}
	updatedUser, err := s.UserRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch user details: %v", err)
	}
	roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, updatedUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}
	_, err = client.DeleteUserRoleRelationship(
		updatedUser.Username,
		helper.LowerCaseSlice(roles),
		helper.LowerCaseSlice(permissions),
		helper.LowerCaseSlice(actions),
	)
	if err != nil {
		log.Printf("Failed to delete relationships: %v", err)
	}
	_, err = client.CreateUserRoleRelationship(
		updatedUser.Username,
		helper.LowerCaseSlice(roles),
		helper.LowerCaseSlice(permissions),
		helper.LowerCaseSlice(actions),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create relation: %v", err)
	}
	roles, permissionsList, err := s.UserRepo.FindUsageRights(ctx, updatedUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}
	pbPermissions := make([]*pb.PermissionResponse, len(permissionsList))
	for i, perm := range permissionsList {
		pbPermissions[i] = &pb.PermissionResponse{
			Name:        perm.Name,
			Description: perm.Description,
			Action:      perm.Action,
			Source:      perm.Source,
			Resource:    perm.Resource,
		}
	}
	userRoleResponse := &pb.UserRoleResponse{
		Roles:       roles,
		Permissions: pbPermissions,
	}
	createdAt := ""
	if !updatedUser.CreatedAt.IsZero() {
		createdAt = updatedUser.CreatedAt.Format(time.RFC3339)
	}

	updatedAt := ""
	if !updatedUser.UpdatedAt.IsZero() {
		updatedAt = updatedUser.UpdatedAt.Format(time.RFC3339)
	}

	connUser := &pb.AssignRolePermission{
		Id:          updatedUser.ID,
		Username:    updatedUser.Username,
		Password:    "",
		IsValidated: updatedUser.IsValidated,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		UsageRight:  userRoleResponse,
	}
	return &pb.AssignRoleToUserResponse{
		StatusCode:    http.StatusOK,
		Success:      true,
		Message:      "Role assigned to user successfully",
		Data:         connUser,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}, nil
}
