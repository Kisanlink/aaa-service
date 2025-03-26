package roles

import (
	"context"
	"net/http"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RoleServer) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Role ID is required")
	}
	updatedRole := make(map[string]interface{})
	if req.Name != "" {
		updatedRole["name"] = req.Name
	}
	if req.Description != "" {
		updatedRole["description"] = req.Description
	}
	if req.Source != "" {
		updatedRole["source"] = req.Source
	}
	if err := s.RoleRepo.UpdateRole(ctx, req.Id, updatedRole); err != nil {
		return nil, err
	}
	updatedRoleModel, err := s.RoleRepo.FindRoleByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	pbRole := &pb.Role{
		Id:          updatedRoleModel.ID,
		Name:        updatedRoleModel.Name,
		Description: updatedRoleModel.Description,
		Source: updatedRoleModel.Source,
	}
	return &pb.UpdateRoleResponse{
		StatusCode:http.StatusOK,
		Success: true,
		Message:    "Role updated successfully",
		Role:       pbRole,
	}, nil
}
