package rolepermission

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConnectRolePermissionServer) DeleteRolePermission(ctx context.Context, req *pb.DeleteConnRolePermissionRequest) (*pb.DeleteConnRolePermissionResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	if err := s.RolePermissionRepo.DeletePermissionOnRoleByUserRoleID(ctx, req.Id); err != nil {
		return nil, err
	}
	if err := s.RolePermissionRepo.DeleteRolePermissionByID(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pb.DeleteConnRolePermissionResponse{
		StatusCode: int32(http.StatusOK),
		Message:    "RolePermission deleted successfully",
	}, nil
}
