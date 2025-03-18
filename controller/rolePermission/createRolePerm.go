package rolepermission

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"github.com/Kisanlink/aaa-service/repositories"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConnectRolePermissionServer struct {
	pb.UnimplementedConnectRolePermissionServiceServer
	RolePermissionRepo *repositories.RolePermissionRepository
	RoleRepo           *repositories.RoleRepository
	PermissionRepo     *repositories.PermissionRepository
}

func NewConnectRolePermissionServer(
	rolePermissionRepo *repositories.RolePermissionRepository,
	roleRepo *repositories.RoleRepository,
	permissionRepo *repositories.PermissionRepository,
) *ConnectRolePermissionServer {
	return &ConnectRolePermissionServer{
		RolePermissionRepo: rolePermissionRepo,
		RoleRepo:           roleRepo,
		PermissionRepo:     permissionRepo,
	}
}

func (s *ConnectRolePermissionServer) CreateConnectRolePermission(ctx context.Context, req *pb.CreateConnRolePermissionRequest) (*pb.CreateConnRolePermissionResponse, error) {
	if len(req.GetRoles()) == 0 || len(req.GetPermissions()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Both role_names and permission_names are required")
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

	return &pb.CreateConnRolePermissionResponse{
		StatusCode: http.StatusCreated,
		Message:    "Role-Permission connections created successfully",
		Data:       connRolePermissions,
	}, nil
}