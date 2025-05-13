package rolepermission

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConnectRolePermissionServer) GetRolePermissionById(ctx context.Context, req *pb.GetConnRolePermissionByIdRequest) (*pb.GetConnRolePermissionByIdResponse, error) {
	rolePermission, err := s.rolePermService.GetRolePermissionByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "RolePermission with ID %s not found", req.Id)
	}
	rolePermissions, err := s.rolePermService.GetRolePermissionsByRoleID(rolePermission.RoleID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch role permissions: %v", err)
	}
	var rolePermissionPtrs []*model.RolePermission
	for i := range rolePermissions {
		rolePermissionPtrs = append(rolePermissionPtrs, &rolePermissions[i])
	}
	connRolePermissionResponse, err := s.buildConnRolePermissionResponse(ctx, &rolePermission.Role, rolePermissionPtrs)
	if err != nil {
		return nil, err
	}
	return &pb.GetConnRolePermissionByIdResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role with Permissions fetched successfully",
		Data:          connRolePermissionResponse,
		DataTimeStamp: time.Now().Format(time.RFC3339Nano),
	}, nil
}

func (s *ConnectRolePermissionServer) buildConnRolePermissionResponse(
	_ context.Context,
	role *model.Role,
	rolePermissions []*model.RolePermission,
) (*pb.ConnRolePermissionResponse, error) {
	if len(rolePermissions) == 0 {
		return nil, status.Error(codes.NotFound, "No permissions found for this role")
	}

	response := &pb.ConnRolePermissionResponse{
		Id:        rolePermissions[0].ID,
		CreatedAt: rolePermissions[0].CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: rolePermissions[0].UpdatedAt.Format(time.RFC3339Nano),
		IsActive:  rolePermissions[0].IsActive,
		Role: &pb.ConnRole{
			Id:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Source:      role.Source,
			CreatedAt:   role.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   role.UpdatedAt.Format(time.RFC3339Nano),
		},
		Permissions: make([]*pb.ConnPermission, 0),
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

	return response, nil
}
