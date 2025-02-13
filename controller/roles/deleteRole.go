package roles

import (
	"context"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RoleServer) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	_, err := s.RoleRepo.FindRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.RoleRepo.DeleteRole(ctx, id); err != nil {
		return nil, err
	}

	return &pb.DeleteRoleResponse{
		StatusCode: int32(codes.OK),
		Message:    "Role deleted successfully",
	}, nil
}
