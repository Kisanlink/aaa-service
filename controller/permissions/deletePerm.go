package permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

func (s *PermissionServer) DeletePermission(ctx context.Context, req *pb.DeletePermissionRequest) (*pb.DeletePermissionResponse, error) {
	id := req.GetId()
	if id == "" {
		return &pb.DeletePermissionResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "ID is required",
		}, nil
	}

	permission := model.Permission{}
	result := s.DB.Table("permissions").Where("id = ?", id).First(&permission)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &pb.DeletePermissionResponse{
				StatusCode: int32(codes.NotFound),
				Message:    fmt.Sprintf("Permission with ID %s not found", id),
			}, nil
		}
		return &pb.DeletePermissionResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to query permission: %v", result.Error),
		}, nil
	}

	if err := s.DB.Table("permissions").Delete(&model.Permission{}, "id = ?", id).Error; err != nil {
		return &pb.DeletePermissionResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to delete permission: %v", err),
		}, nil
	}

	return &pb.DeletePermissionResponse{
		StatusCode: int32(codes.OK),
		Message:    "Permission deleted successfully",
	}, nil
}
