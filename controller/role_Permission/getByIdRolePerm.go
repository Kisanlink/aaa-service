package rolepermission

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	pb "github.com/Kisanlink/aaa-service/proto"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

func (s *ConnectRolePermissionServer) GetRolePermissionById(ctx context.Context, req *pb.GetConnRolePermissionByIdRequest) (*pb.GetConnRolePermissionByIdResponse, error) {
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
		if err == gorm.ErrRecordNotFound {
			return &pb.GetConnRolePermissionByIdResponse{
				StatusCode: int32(codes.NotFound),
				Message:    fmt.Sprintf("RolePermission with ID %s not found", req.Id),
			}, nil
		}
		return &pb.GetConnRolePermissionByIdResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to fetch RolePermission: %v", err),
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

	return &pb.GetConnRolePermissionByIdResponse{
		StatusCode:         http.StatusOK,
		Message:            "RolePermission fetched successfully",
		ConnRolePermission: pbRolePermission,
	}, nil
}
