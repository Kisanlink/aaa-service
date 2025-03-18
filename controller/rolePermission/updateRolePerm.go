package rolepermission

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


func (s *ConnectRolePermissionServer) UpdateRolePermission(ctx context.Context, req *pb.UpdateConnRolePermissionRequest) (*pb.UpdateConnRolePermissionResponse, error) {
	_, err := s.RolePermissionRepo.GetRolePermissionByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "RolePermission with ID %s not found", req.Id)
	}

	roleIDs := make([]string, 0)
	for _, roleName := range req.GetRoles() {
		role, err := s.RoleRepo.GetRoleByName(ctx, roleName)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "Role with name %s not found", roleName)
		}
		roleIDs = append(roleIDs, role.ID)
	}
	permissionIDs := make([]string, 0)
	for _, permissionName := range req.GetPermissions() {
		permission, err := s.PermissionRepo.FindPermissionByName(ctx, permissionName)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "Permission with name %s not found", permissionName)
		}
		permissionIDs = append(permissionIDs, permission.ID)
	}
	if err := s.RolePermissionRepo.DeleteRolePermissionByID(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete existing role-permission connections: %v", err)
	}
	var rolePermissions []*model.RolePermission
	for _, roleID := range roleIDs {
		for _, permissionID := range permissionIDs {
			rolePermission := &model.RolePermission{
				RoleID:       roleID,
				PermissionID: permissionID,
				IsActive:     true,
			}
			rolePermissions = append(rolePermissions, rolePermission)
		}
	}

	if err := s.RolePermissionRepo.CreateRolePermissions(ctx, rolePermissions); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create role-permission connections: %v", err)
	}
	var connRolePermissions []*pb.ConnRolePermission
	for _, rp := range rolePermissions {
		connRolePermission := &pb.ConnRolePermission{
			Id:           rp.ID,
			CreatedAt:    rp.CreatedAt.String(),
			UpdatedAt:    rp.UpdatedAt.String(),
			RoleId:       rp.RoleID,
			PermissionId: rp.PermissionID,
			IsActive:     rp.IsActive,
		}
		connRolePermissions = append(connRolePermissions, connRolePermission)
	}
	return &pb.UpdateConnRolePermissionResponse{
		StatusCode: http.StatusOK,
		Message:    "Role-Permission connections updated successfully",
		Data:       connRolePermissions,
	}, nil
}