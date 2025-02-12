package rolepermission

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
)

func (s *ConnectRolePermissionServer) DeleteRolePermission(ctx context.Context, req *pb.DeleteConnRolePermissionRequest) (*pb.DeleteConnRolePermissionResponse, error) {
	if req.Id == "" {
		return &pb.DeleteConnRolePermissionResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "ID is required",
		}, nil
	}

	if err := s.DB.Table("permission_on_roles").Where("user_role_id = ?", req.Id).Delete(&model.PermissionOnRole{}).Error; err != nil {
		return &pb.DeleteConnRolePermissionResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to delete associated PermissionOnRole entries: %v", err),
		}, nil
	}

	if err := s.DB.Table("role_permissions").Where("id = ?", req.Id).Delete(&model.RolePermission{}).Error; err != nil {
		return &pb.DeleteConnRolePermissionResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to delete RolePermission: %v", err),
		}, nil
	}

	return &pb.DeleteConnRolePermissionResponse{
		StatusCode: http.StatusOK,
		Message:    "RolePermission deleted successfully",
	}, nil
}
