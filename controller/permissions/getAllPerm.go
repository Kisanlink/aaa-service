package permissions

import (
	"context"
	"net/http"

	"github.com/kisanlink/protobuf/pb-aaa"
)

func (s *PermissionServer) GetAllPermissions(ctx context.Context, req *pb.GetAllPermissionsRequest) (*pb.GetAllPermissionsResponse, error) {
	permissions, err := s.PermissionRepo.FindAllPermissions(ctx)
	if err != nil {
		return nil, err
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
		StatusCode:http.StatusOK,
		Success: true,
		Message:     "Permissions retrieved successfully",
		Permissions: pbPermissions,
	}, nil
}
