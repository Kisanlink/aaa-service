package roles

import (
	"context"
	"net/http"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RoleServer) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	_, err := s.roleService.FindRoleByID(id)
	if err != nil {
		return nil, err
	}
	if err := s.roleService.DeleteRole(id); err != nil {
		return nil, err
	}

	return &pb.DeleteRoleResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format
		Message:       "Role deleted successfully",
	}, nil
}
