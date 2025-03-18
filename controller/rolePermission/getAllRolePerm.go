package rolepermission

import (
	"context"
	"net/http"
	"reflect"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConnectRolePermissionServer) GetAllRolePermission(ctx context.Context, req *pb.GetConnRolePermissionallRequest) (*pb.GetConnRolePermissionallResponse, error) {
	rolePermissions, err := s.RolePermissionRepo.GetAllRolePermissions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch role-permission connections: %v", err)
	}
	rolePermissionMap := make(map[string]*pb.ConnRolePermission)
	for _, rp := range rolePermissions {
		if IsZeroValued(rp.Role) || IsZeroValued(rp.Permission) || rp.Permission.ID == "" {
			continue
		}
		if _, exists := rolePermissionMap[rp.RoleID]; !exists {
			rolePermissionMap[rp.RoleID] = &pb.ConnRolePermission{
				Id:         rp.ID,
				CreatedAt:  rp.CreatedAt.String(),
				UpdatedAt:  rp.UpdatedAt.String(),
				RoleId:     rp.RoleID,
				Role: &pb.ConnRole{
					Id:          rp.Role.ID,
					Name:        rp.Role.Name,
					Description: rp.Role.Description,
					Source:      rp.Role.Source,
					CreatedAt:   rp.Role.CreatedAt.String(),
					UpdatedAt:   rp.Role.UpdatedAt.String(),
				},
				Permission: []*pb.ConnPermission{},
				IsActive:   rp.IsActive,
			}
		}
		rolePermissionMap[rp.RoleID].Permission = append(rolePermissionMap[rp.RoleID].Permission, &pb.ConnPermission{
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
	var connRolePermissions []*pb.ConnRolePermission
	for _, rp := range rolePermissionMap {
		connRolePermissions = append(connRolePermissions, rp)
	}
	return &pb.GetConnRolePermissionallResponse{
		StatusCode: http.StatusOK,
		Message:    "Role-Permission connections fetched successfully",
		Data:       connRolePermissions,
	}, nil
}
func IsZeroValued[T any](v T) bool {
	return reflect.ValueOf(v).IsZero()
}

