package permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/model"
	pb "github.com/Kisanlink/aaa-service/proto"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

func (s *PermissionServer) GetPermissionById(ctx context.Context, req *pb.GetPermissionByIdRequest) (*pb.GetPermissionByIdResponse, error) {
	id := req.GetId()
	if id == "" {
		return &pb.GetPermissionByIdResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "ID is required",
		}, nil
	}

	// Fetch the permission from the database
	permission := model.Permission{}
	result := s.DB.Table("permissions").Where("id = ?", id).First(&permission)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &pb.GetPermissionByIdResponse{
				StatusCode: int32(codes.NotFound),
				Message:    fmt.Sprintf("Permission with ID %s not found", id),
			}, nil
		}
		return &pb.GetPermissionByIdResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to query permission: %v", result.Error),
		}, nil
	}

	// Populate the response object with the fetched permission data
	pbPermission := &pb.Permission{
		Id:          permission.ID, // Ensure ID is converted to a string
		Name:        permission.Name,
		Description: permission.Description,
	}

	// Return the response with the fully populated permission data
	return &pb.GetPermissionByIdResponse{
		StatusCode: int32(codes.OK),
		Message:    "Permission retrieved successfully",
		Permission: pbPermission,
	}, nil
}
