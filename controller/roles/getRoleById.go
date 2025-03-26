package roles

import (
	"context"
	"net/http"

	"github.com/kisanlink/protobuf/pb-aaa"
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
		Source: role.Source,
	}
	return &pb.GetRoleByIdResponse{
		StatusCode:http.StatusOK,
		Success: true,
		Message:    "Role retrieved successfully",
		Role:       pbRole,
	}, nil
}
