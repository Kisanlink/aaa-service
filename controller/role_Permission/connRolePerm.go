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
	if len(req.RoleIds) == 0 || len(req.PermissionIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Both role_ids and permission_ids are required")
	}

	var connRolePermission pb.ConnRolePermission
	var permissionOnRoles []*pb.ConnPermissionOnRole

	for _, roleID := range req.RoleIds {
		role, err := s.RoleRepo.FindRoleByID(ctx, roleID)
		if err != nil {
			return nil, err
		}
		rolePermission := model.RolePermission{
			RoleID: role.ID,
		}

		if err := s.RolePermissionRepo.CreateRolePermission(ctx, &rolePermission); err != nil {
			return nil, err
		}
		for _, permissionID := range req.PermissionIds {
			permission, err := s.PermissionRepo.FindPermissionByID(ctx, permissionID)
			if err != nil {
				return nil, err
			}
			permissionOnRole := model.PermissionOnRole{
				PermissionID: permission.ID,
				UserRoleID:   rolePermission.ID,
			}

			if err := s.RolePermissionRepo.CreatePermissionOnRole(ctx, &permissionOnRole); err != nil {
				return nil, err
			}
			pbPermissionOnRole := &pb.ConnPermissionOnRole{
				Id:         permissionOnRole.ID,
				CreatedAt:  permissionOnRole.CreatedAt.String(),
				UpdatedAt:  permissionOnRole.UpdatedAt.String(),
				UserRoleId: permissionOnRole.UserRoleID,
				Permission: &pb.ConnPermission{
					Id:          permission.ID,
					Name:        permission.Name,
					Description: permission.Description,
				},
			}
			permissionOnRoles = append(permissionOnRoles, pbPermissionOnRole)
		}

		connRolePermission.Id = rolePermission.ID
		connRolePermission.CreatedAt = rolePermission.CreatedAt.String()
		connRolePermission.UpdatedAt = rolePermission.UpdatedAt.String()
		connRolePermission.Role = &pb.ConnRole{
			Id:          role.ID,
			Name:        role.Name,
			Description: role.Description,
		}
		connRolePermission.PermissionOnRoles = permissionOnRoles
	}

	return &pb.CreateConnRolePermissionResponse{
		StatusCode:         http.StatusCreated,
		Message:            "RolePermissions created successfully",
		ConnRolePermission: &connRolePermission,
	}, nil
}
