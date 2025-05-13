package permissions

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PermissionServer) UpdatePermission(ctx context.Context, req *pb.UpdatePermissionRequest) (*pb.UpdatePermissionResponse, error) {

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Permission ID is required")
	}
	updatedPermission := model.Permission{
		Name:        req.Name,
		Description: req.Description,
		Source:      req.Source,
		Action:      req.Action,
		Resource:    req.Resource,
	}
	if req.Name != "" {
		updatedPermission.Name = req.Name
	}
	if req.Description != "" {
		updatedPermission.Description = req.Description
	}
	if req.Action != "" {
		updatedPermission.Action = req.Action
	}
	if req.Source != "" {
		updatedPermission.Source = req.Source
	}
	if req.Resource != "" {
		updatedPermission.Resource = req.Resource
	}
	if err := s.permissionService.UpdatePermission(req.Id, updatedPermission); err != nil {
		return nil, err
	}
	updatedPermissionModel, err := s.permissionService.FindPermissionByID(req.Id)
	if err != nil {
		return nil, err
	}
	pbPermission := &pb.Permission{
		Id:          updatedPermissionModel.ID,
		Name:        updatedPermissionModel.Name,
		Description: updatedPermissionModel.Description,
		Source:      req.Source,
		Action:      req.Action,
		Resource:    req.Resource,
		// ValidStartTime: permission.ValidStartTime,
		// ValiedEndTime: permission.ValiedEndTime,
	}
	return &pb.UpdatePermissionResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Permission updated successfully",
		Data:          pbPermission,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format

	}, nil
}
