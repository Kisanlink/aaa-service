package rolepermission

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

func (s *ConnectRolePermissionServer) UpdateRolePermission(ctx context.Context, req *pb.UpdateConnRolePermissionRequest) (*pb.UpdateConnRolePermissionResponse, error) {
	if req.Id == "" || len(req.RoleIds) == 0 || len(req.PermissionIds) == 0 {
		return &pb.UpdateConnRolePermissionResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "ID, role_ids, and permission_ids are required",
		}, nil
	}
	updates := map[string]interface{}{
		"role_id": req.RoleIds[0],
	}
	if err := s.DB.Table("role_permissions").Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		return &pb.UpdateConnRolePermissionResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to update RolePermission: %v", err),
		}, nil
	}

	if err := s.DB.Table("permission_on_roles").Where("user_role_id = ?", req.Id).Delete(&model.PermissionOnRole{}).Error; err != nil {
		return &pb.UpdateConnRolePermissionResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to delete existing PermissionOnRole entries: %v", err),
		}, nil
	}
	for _, permissionID := range req.PermissionIds {
		permissionOnRole := model.PermissionOnRole{
			PermissionID: permissionID,
			UserRoleID:   req.Id,
		}
		if err := s.DB.Table("permission_on_roles").Create(&permissionOnRole).Error; err != nil {
			return &pb.UpdateConnRolePermissionResponse{
				StatusCode: int32(codes.Internal),
				Message:    fmt.Sprintf("Failed to create PermissionOnRole: %v", err),
			}, nil
		}
	}
	var rolePermission model.RolePermission
	if err := s.DB.Table("role_permissions").
		Preload("PermissionOnRoles", func(db *gorm.DB) *gorm.DB {
			return db.Table("permission_on_roles")
		}).
		Preload("PermissionOnRoles.Permission", func(db *gorm.DB) *gorm.DB {
			return db.Table("permissions")
		}).
		Preload("Role", func(db *gorm.DB) *gorm.DB {
			return db.Table("roles")
		}).
		Where("id = ?", req.Id).First(&rolePermission).Error; err != nil {
		return &pb.UpdateConnRolePermissionResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to fetch updated RolePermission: %v", err),
		}, nil
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
	return &pb.UpdateConnRolePermissionResponse{
		StatusCode:         http.StatusOK,
		Message:            "RolePermission updated successfully",
		ConnRolePermission: pbRolePermission,
	}, nil
}
