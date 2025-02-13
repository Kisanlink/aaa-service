package rolepermission

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConnectRolePermissionServer) UpdateRolePermission(ctx context.Context, req *pb.UpdateConnRolePermissionRequest) (*pb.UpdateConnRolePermissionResponse, error) {
	// Validate input
	if req.Id == "" || len(req.RoleIds) == 0 || len(req.PermissionIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "ID, role_ids, and permission_ids are required")
	}

	// Prepare updates for RolePermission
	updates := map[string]interface{}{
		"role_id": req.RoleIds[0],
	}

	// Update the RolePermission entry
	if err := s.RolePermissionRepo.UpdateRolePermission(ctx, req.Id, updates); err != nil {
		return nil, err
	}

	// Delete existing PermissionOnRole entries
	if err := s.RolePermissionRepo.DeletePermissionOnRoleByUserRoleID(ctx, req.Id); err != nil {
		return nil, err
	}

	// Create new PermissionOnRole entries
	for _, permissionID := range req.PermissionIds {
		permissionOnRole := model.PermissionOnRole{
			PermissionID: permissionID,
			UserRoleID:   req.Id,
		}
		if err := s.RolePermissionRepo.CreatePermissionOnRole(ctx, &permissionOnRole); err != nil {
			return nil, err
		}
	}

	// Fetch the updated RolePermission
	updatedRolePermission, err := s.RolePermissionRepo.FindRolePermissionByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Convert the updated RolePermission to protobuf format
	var permissionOnRoles []*pb.ConnPermissionOnRole
	for _, por := range updatedRolePermission.PermissionOnRoles {
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
		Id:          updatedRolePermission.Role.ID,
		Name:        updatedRolePermission.Role.Name,
		Description: updatedRolePermission.Role.Description,
	}

	pbRolePermission := &pb.ConnRolePermission{
		Id:                updatedRolePermission.ID,
		CreatedAt:         updatedRolePermission.CreatedAt.String(),
		UpdatedAt:         updatedRolePermission.UpdatedAt.String(),
		Role:              pbRole,
		PermissionOnRoles: permissionOnRoles,
	}

	// Return success response
	return &pb.UpdateConnRolePermissionResponse{
		StatusCode:         int32(http.StatusOK),
		Message:            "RolePermission updated successfully",
		ConnRolePermission: pbRolePermission,
	}, nil
}
