package permissions

import (
	"context"
	"net/http"
	"time"

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
			Source: permission.Source,
			Action: permission.Action,
			Resource: permission.Resource,
			ValidStartTime: permission.ValidStartTime.Format(time.RFC3339Nano),
			ValiedEndTime: permission.ValidEndTime.Format(time.RFC3339Nano),
		})
	}

	return &pb.GetAllPermissionsResponse{
		StatusCode:http.StatusOK,
		Success: true,
		Message:     "Permissions retrieved successfully",
		Permissions: pbPermissions,
	}, nil
}
