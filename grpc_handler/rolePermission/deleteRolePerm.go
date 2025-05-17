package rolepermission

import (
	"context"
	"net/http"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConnectRolePermissionServer) DeleteRolePermission(ctx context.Context, req *pb.DeleteConnRolePermissionRequest) (*pb.DeleteConnRolePermissionResponse, error) {
	role, err := s.roleService.GetRoleByName(req.Role)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role with name %s not found", req.Role)
	}
	err = s.rolePermService.DeleteRolePermissionByRoleID(role.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete role-permission connections: %v", err)
	}
	return &pb.DeleteConnRolePermissionResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role with Permissions deleted successfully",
		DataTimeStamp: time.Now().Format(time.RFC3339Nano),
	}, nil
}
