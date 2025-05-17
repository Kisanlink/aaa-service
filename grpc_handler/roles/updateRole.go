package roles

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RoleServer) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Role ID is required")
	}
	updatedRole := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Source:      req.Source,
	}
	if req.Name != "" {
		updatedRole.Name = req.Name
	}
	if req.Description != "" {
		updatedRole.Description = req.Description
	}
	if req.Source != "" {
		updatedRole.Source = req.Source
	}
	if err := s.roleService.UpdateRole(req.Id, updatedRole); err != nil {
		return nil, err
	}
	updatedRoleModel, err := s.roleService.FindRoleByID(req.Id)
	if err != nil {
		return nil, err
	}
	pbRole := &pb.Role{
		Id:          updatedRoleModel.ID,
		Name:        updatedRoleModel.Name,
		Description: updatedRoleModel.Description,
		Source:      updatedRoleModel.Source,
	}
	return &pb.UpdateRoleResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role updated successfully",
		Data:          pbRole,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format

	}, nil
}
