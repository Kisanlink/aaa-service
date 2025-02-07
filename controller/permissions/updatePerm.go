package permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

func (s *PermissionServer) UpdatePermission(ctx context.Context, req *pb.UpdatePermissionRequest) (*pb.UpdatePermissionResponse, error) {
	permission := req.GetPermission()
	if permission == nil {
		return &pb.UpdatePermissionResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "Permission cannot be nil",
		}, nil
	}
	if permission.Id == "" {
		return &pb.UpdatePermissionResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "Permission ID is required",
		}, nil
	}

	existingPermission := model.Permission{}
	result := s.DB.Table("permissions").Where("id = ?", permission.Id).First(&existingPermission)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &pb.UpdatePermissionResponse{
				StatusCode: int32(codes.NotFound),
				Message:    fmt.Sprintf("Permission with ID %s not found", permission.Id),
			}, nil
		}
		return &pb.UpdatePermissionResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to query permission: %v", result.Error),
		}, nil
	}

	updatedPermission := model.Permission{
		Name:        permission.Name,
		Description: permission.Description,
	}
	if err := s.DB.Table("permissions").Model(&model.Permission{}).Where("id = ?", permission.Id).Updates(updatedPermission).Error; err != nil {
		return &pb.UpdatePermissionResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to update permission: %v", err),
		}, nil
	}

	pbPermission := &pb.Permission{
		Id:          existingPermission.ID.String(),
		Name:        updatedPermission.Name,
		Description: updatedPermission.Description,
	}

	return &pb.UpdatePermissionResponse{
		StatusCode: int32(codes.OK),
		Message:    "Permission updated successfully",
		Permission: pbPermission,
	}, nil
}
