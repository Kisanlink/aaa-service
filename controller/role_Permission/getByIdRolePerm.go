package rolepermission

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConnectRolePermissionServer) GetRolePermissionById(ctx context.Context, req *pb.GetConnRolePermissionByIdRequest) (*pb.GetConnRolePermissionByIdResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	rolePermission, err := s.RolePermissionRepo.FindRolePermissionByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	var permissionOnRoles []*pb.ConnPermissionOnRole
	for _, por := range rolePermission.PermissionOnRoles {
		pbPermissionOnRole := &pb.ConnPermissionOnRole{
			Id:         por.ID,
			CreatedAt:  por.CreatedAt.String(),
			UpdatedAt:  por.UpdatedAt.String(),
			UserRoleId: por.UserRoleID,
			Permission: &pb.ConnPermission{
				Id:          por.Permission.ID,
				Name:        por.Permission.Name,
				Description: por.Permission.Description,
			},
		}
		permissionOnRoles = append(permissionOnRoles, pbPermissionOnRole)
	}

	pbRole := &pb.ConnRole{
		Id:          rolePermission.Role.ID,
		Name:        rolePermission.Role.Name,
		Description: rolePermission.Role.Description,
	}

	pbRolePermission := &pb.ConnRolePermission{
		Id:                rolePermission.ID,
		CreatedAt:         rolePermission.CreatedAt.String(),
		UpdatedAt:         rolePermission.UpdatedAt.String(),
		Role:              pbRole,
		PermissionOnRoles: permissionOnRoles,
	}
	return &pb.GetConnRolePermissionByIdResponse{
		StatusCode:         int32(http.StatusOK),
		Message:            "RolePermission fetched successfully",
		ConnRolePermission: pbRolePermission,
	}, nil
}
