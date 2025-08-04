package rolepermission

import (
	"context"

	pb "github.com/Kisanlink/aaa-service/proto"
	"google.golang.org/grpc/codes"
)

func (s *ConnectRolePermissionServer) GetRolePermissionById(ctx context.Context, req *pb.GetConnRolePermissionByIdRequest) (*pb.GetConnRolePermissionByIdResponse, error) {
	// TODO: Implement this functionality after the model refactoring is complete
	// This requires RolePermission and PermissionOnRole models to be properly defined
	return &pb.GetConnRolePermissionByIdResponse{
		StatusCode: int32(codes.Unimplemented),
		Message:    "This functionality is temporarily disabled during model refactoring",
	}, nil
}
