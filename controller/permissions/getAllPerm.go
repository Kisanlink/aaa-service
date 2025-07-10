package permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/model"
	pb "github.com/Kisanlink/aaa-service/proto"
	"google.golang.org/grpc/codes"
)

func (s *PermissionServer) GetAllPermissions(ctx context.Context, req *pb.GetAllPermissionsRequest) (*pb.GetAllPermissionsResponse, error) {
	var permissions []model.Permission
	result := s.DB.Table("permissions").Find(&permissions)
	if result.Error != nil {
		return &pb.GetAllPermissionsResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to retrieve permissions: %v", result.Error),
		}, nil
	}

	var pbPermissions []*pb.Permission
	for _, permission := range permissions {
		pbPermissions = append(pbPermissions, &pb.Permission{
			Id:          permission.ID,
			Name:        permission.Name,
			Description: permission.Description,
		})
	}

	return &pb.GetAllPermissionsResponse{
		StatusCode:  int32(codes.OK),
		Message:     "Permissions retrieved successfully",
		Permissions: pbPermissions,
	}, nil
}
