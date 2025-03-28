package permissions

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PermissionServer) DeletePermission(ctx context.Context, req *pb.DeletePermissionRequest) (*pb.DeletePermissionResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	_, err := s.PermissionRepo.FindPermissionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.PermissionRepo.DeletePermission(ctx, id); err != nil {
		return nil, err
	}
	return &pb.DeletePermissionResponse{
		StatusCode:http.StatusOK,
		Success: true,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format
		Message:    fmt.Sprintf("Permission with ID %s deleted successfully", id),
	}, nil
}
