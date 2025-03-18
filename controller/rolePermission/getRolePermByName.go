package rolepermission

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConnectRolePermissionServer) GetRolePermissionByRoleName(ctx context.Context, req *pb.GetRolePermissionByRoleNameRequest) (*pb.GetRolePermissionByRoleNameResponse, error) {
	role, err := s.RoleRepo.GetRoleByName(ctx, req.Role)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role with name %s not found", req.Role)
	}
	rolePermissions, err := s.RolePermissionRepo.GetRolePermissionsByRoleID(ctx, role.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch role-permission connections: %v", err)
	}
	connRolePermission := &pb.ConnRolePermission{
		Id:         rolePermissions[0].ID,
		CreatedAt:  rolePermissions[0].CreatedAt.String(),
		UpdatedAt:  rolePermissions[0].UpdatedAt.String(),
		RoleId:     role.ID,
		Role: &pb.ConnRole{
			Id:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Source:      role.Source,
			CreatedAt:   role.CreatedAt.String(),
			UpdatedAt:   role.UpdatedAt.String(),
		},
		Permission: []*pb.ConnPermission{},
		IsActive:   rolePermissions[0].IsActive,
	}
	for _, rp := range rolePermissions {
		if !IsZeroValued(rp.Permission) && rp.Permission.ID != "" {
			connRolePermission.Permission = append(connRolePermission.Permission, &pb.ConnPermission{
				Id:             rp.Permission.ID,
				Name:           rp.Permission.Name,
				Description:    rp.Permission.Description,
				Action:         rp.Permission.Action,
				ValidStartTime: rp.Permission.ValidStartTime.String(),
				ValidEndTime:   rp.Permission.ValidEndTime.String(),
				CreatedAt:      rp.Permission.CreatedAt.String(),
				UpdatedAt:      rp.Permission.UpdatedAt.String(),
			})
		}
	}
	return &pb.GetRolePermissionByRoleNameResponse{
		StatusCode: http.StatusOK,
		Message:    "Role-Permission connections fetched successfully",
		Data:       connRolePermission,
	}, nil
}