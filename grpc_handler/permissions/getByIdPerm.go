package permissions

import (
	"context"
	"net/http"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PermissionServer) GetPermissionById(ctx context.Context, req *pb.GetPermissionByIdRequest) (*pb.GetPermissionByIdResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	permission, err := s.permissionService.FindPermissionByID(id)
	if err != nil {
		return nil, err
	}
	pbPermission := &pb.Permission{
		Id:             permission.ID,
		Name:           permission.Name,
		Description:    permission.Description,
		Source:         permission.Source,
		Action:         permission.Action,
		Resource:       permission.Resource,
		ValidStartTime: permission.ValidStartTime.Format(time.RFC3339Nano),
		ValiedEndTime:  permission.ValidEndTime.Format(time.RFC3339Nano),
	}
	return &pb.GetPermissionByIdResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Permission retrieved successfully",
		Data:          pbPermission,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format

	}, nil
}
