package rolepermission

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConnectRolePermissionServer) GetRolePermissionByRoleName(ctx context.Context, req *pb.GetRolePermissionByRoleNameRequest) (*pb.GetRolePermissionByRoleNameResponse, error) {
	if req.Role == "" {
		return nil, status.Error(codes.InvalidArgument, "Role name is required")
	}

	role, err := s.roleService.GetRoleByName(req.Role)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role with name %s not found", req.Role)
	}

	rolePermissions, err := s.rolePermService.GetRolePermissionsByRoleID(role.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch role-permission connections: %v", err)
	}

	if len(rolePermissions) == 0 {
		return nil, status.Error(codes.NotFound, "No permissions found for this role")
	}

	response := &pb.ConnRolePermissionResponse{
		Id:        rolePermissions[0].ID,
		CreatedAt: rolePermissions[0].CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: rolePermissions[0].UpdatedAt.Format(time.RFC3339Nano),
		Role: &pb.ConnRole{
			Id:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Source:      role.Source,
			CreatedAt:   role.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   role.UpdatedAt.Format(time.RFC3339Nano),
		},
		Permissions: []*pb.ConnPermission{},
		IsActive:    rolePermissions[0].IsActive,
	}

	for _, rp := range rolePermissions {
		if !helper.IsZeroValued(rp.Permission) && rp.Permission.ID != "" {
			response.Permissions = append(response.Permissions, &pb.ConnPermission{
				Id:             rp.Permission.ID,
				Name:           rp.Permission.Name,
				Description:    rp.Permission.Description,
				Action:         rp.Permission.Action,
				Resource:       rp.Permission.Resource,
				Source:         rp.Permission.Source,
				ValidStartTime: rp.Permission.ValidStartTime.Format(time.RFC3339Nano),
				ValidEndTime:   rp.Permission.ValidEndTime.Format(time.RFC3339Nano),
				CreatedAt:      rp.Permission.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:      rp.Permission.UpdatedAt.Format(time.RFC3339Nano),
			})
		}
	}

	return &pb.GetRolePermissionByRoleNameResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role with Permissions fetched successfully",
		Data:          response,
		DataTimeStamp: time.Now().Format(time.RFC3339Nano),
	}, nil
}
