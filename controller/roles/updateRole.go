package roles

import (
	"context"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RoleServer) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	role := req.GetRole()
	if role == nil {
		return nil, status.Error(codes.InvalidArgument, "Role cannot be nil")
	}
	if role.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Role ID is required")
	}
	updatedRole := make(map[string]interface{})
	if role.Name != "" {
		updatedRole["name"] = role.Name
	}
	if role.Description != "" {
		updatedRole["description"] = role.Description
	}
	if err := s.RoleRepo.UpdateRole(ctx, role.Id, updatedRole); err != nil {
		return nil, err
	}
	updatedRoleModel, err := s.RoleRepo.FindRoleByID(ctx, role.Id)
	if err != nil {
		return nil, err
	}
	pbRole := &pb.Role{
		Id:          updatedRoleModel.ID,
		Name:        updatedRoleModel.Name,
		Description: updatedRoleModel.Description,
	}
	return &pb.UpdateRoleResponse{
		StatusCode: int32(codes.OK),
		Message:    "Role updated successfully",
		Role:       pbRole,
	}, nil
}
