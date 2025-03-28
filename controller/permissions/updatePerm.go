package permissions

import (
	"context"
	"net/http"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PermissionServer) UpdatePermission(ctx context.Context, req *pb.UpdatePermissionRequest) (*pb.UpdatePermissionResponse, error) {

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Permission ID is required")
	}
	updatedPermission := make(map[string]interface{})
	if req.Name != "" {
		updatedPermission["name"] = req.Name
	}
	if req.Description != "" {
		updatedPermission["description"] = req.Description
	}
	if err := s.PermissionRepo.UpdatePermission(ctx, req.Id, updatedPermission); err != nil {
		return nil, err
	}
	updatedPermissionModel, err := s.PermissionRepo.FindPermissionByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	pbPermission := &pb.Permission{
		Id:          updatedPermissionModel.ID,
		Name:        updatedPermissionModel.Name,
		Description: updatedPermissionModel.Description,
		Source: req.Source,
		Action: req.Action,
		Resource: req.Resource,
		// ValidStartTime: permission.ValidStartTime,
		// ValiedEndTime: permission.ValiedEndTime,
	}
	return &pb.UpdatePermissionResponse{
		StatusCode:http.StatusOK,
		Success: true,
		Message:    "Permission updated successfully",
		Data: pbPermission,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format

	}, nil
}
