package permissions

import (
	"context"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
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
		StatusCode:  int32(codes.OK),
		Message:     "Permissions retrieved successfully",
		Permissions: pbPermissions,
	}, nil
}
