package rolepermission

import (
	"context"
	"net/http"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


func (s *ConnectRolePermissionServer) DeleteRolePermission(ctx context.Context, req *pb.DeleteConnRolePermissionRequest) (*pb.DeleteConnRolePermissionResponse, error) {
	role, err := s.RoleRepo.GetRoleByName(ctx, req.Role)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role with name %s not found", req.Role)
	}
	 err=s.RolePermissionRepo.DeleteRolePermissionByRoleID(ctx,role.ID)
	 if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete role-permission connections: %v", err)
	}
	return &pb.DeleteConnRolePermissionResponse{
		StatusCode: http.StatusOK,
		Message:    "Role-Permission connections deleted successfully",
	}, nil
}