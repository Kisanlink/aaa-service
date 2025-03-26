package rolepermission

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConnectRolePermissionServer) GetRolePermissionById(ctx context.Context, req *pb.GetConnRolePermissionByIdRequest) (*pb.GetConnRolePermissionByIdResponse, error) {
	rolePermission, err := s.RolePermissionRepo.GetRolePermissionByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "RolePermission with ID %s not found", req.Id)
	}
	var connPermission *pb.ConnPermission
	if rolePermission.Permission.ID != "" {
		connPermission = &pb.ConnPermission{
			Id:             rolePermission.Permission.ID,
			Name:           rolePermission.Permission.Name,
			Description:    rolePermission.Permission.Description,
			Action:         rolePermission.Permission.Action,
			Resource:         rolePermission.Permission.Resource,
			Source:         rolePermission.Permission.Source,
			ValidStartTime: rolePermission.Permission.ValidStartTime.String(),
			ValidEndTime:   rolePermission.Permission.ValidEndTime.String(),
			CreatedAt:      rolePermission.Permission.CreatedAt.String(),
			UpdatedAt:      rolePermission.Permission.UpdatedAt.String(),
		}
	} else {
		fmt.Println("Permission not found for RolePermission ID:", rolePermission.ID)
	}
	connRolePermission := &pb.RolePermissionConn{
		Id:           rolePermission.ID,
		CreatedAt:    rolePermission.CreatedAt.String(),
		UpdatedAt:    rolePermission.UpdatedAt.String(),
		RoleId:       rolePermission.RoleID,
		PermissionId: rolePermission.PermissionID,
		Role: &pb.ConnRole{
			Id:          rolePermission.Role.ID,
			Name:        rolePermission.Role.Name,
			Description: rolePermission.Role.Description,
			Source:      rolePermission.Role.Source,
			CreatedAt:   rolePermission.Role.CreatedAt.String(),
			UpdatedAt:   rolePermission.Role.UpdatedAt.String(),
		},
		Permission: connPermission,
		IsActive:   rolePermission.IsActive,
	}
	return &pb.GetConnRolePermissionByIdResponse{
		StatusCode: http.StatusOK,
		Success: true,
		Message:    "Role with Permissions fetched successfully",
		Data:       connRolePermission,
	}, nil
}