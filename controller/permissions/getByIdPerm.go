package permissions

import (
	"context"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PermissionServer) GetPermissionById(ctx context.Context, req *pb.GetPermissionByIdRequest) (*pb.GetPermissionByIdResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	permission, err := s.PermissionRepo.FindPermissionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	pbPermission := &pb.Permission{
		Id:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
	}
	return &pb.GetPermissionByIdResponse{
		StatusCode: int32(codes.OK),
		Message:    "Permission retrieved successfully",
		Permission: pbPermission,
	}, nil
}
