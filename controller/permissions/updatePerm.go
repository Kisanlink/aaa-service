package permissions

import (
	"context"
	"net/http"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PermissionServer) UpdatePermission(ctx context.Context, req *pb.UpdatePermissionRequest) (*pb.UpdatePermissionResponse, error) {
	permission := req.GetPermission()
	if permission == nil {
		return nil, status.Error(codes.InvalidArgument, "Permission cannot be nil")
	}
	if permission.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Permission ID is required")
	}
	updatedPermission := make(map[string]interface{})
	if permission.Name != "" {
		updatedPermission["name"] = permission.Name
	}
	if permission.Description != "" {
		updatedPermission["description"] = permission.Description
	}
	if err := s.PermissionRepo.UpdatePermission(ctx, permission.Id, updatedPermission); err != nil {
		return nil, err
	}
	updatedPermissionModel, err := s.PermissionRepo.FindPermissionByID(ctx, permission.Id)
	if err != nil {
		return nil, err
	}
	pbPermission := &pb.Permission{
		Id:          updatedPermissionModel.ID,
		Name:        updatedPermissionModel.Name,
		Description: updatedPermissionModel.Description,
		Source: permission.Source,
		Action: permission.Action,
		Resource: permission.Resource,
		ValidStartTime: permission.ValidStartTime,
		ValiedEndTime: permission.ValiedEndTime,
	}
	return &pb.UpdatePermissionResponse{
		StatusCode:http.StatusOK,
		Success: true,
		Message:    "Permission updated successfully",
		Permission: pbPermission,
	}, nil
}
