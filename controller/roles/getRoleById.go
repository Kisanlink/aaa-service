package roles

import (
	"context"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RoleServer) GetRoleById(ctx context.Context, req *pb.GetRoleByIdRequest) (*pb.GetRoleByIdResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	role, err := s.RoleRepo.FindRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	pbRole := &pb.Role{
		Id:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}
	return &pb.GetRoleByIdResponse{
		StatusCode: int32(codes.OK),
		Message:    "Role retrieved successfully",
		Role:       pbRole,
	}, nil
}
