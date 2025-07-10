package roles

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/model"
	pb "github.com/Kisanlink/aaa-service/proto"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

func (s *RoleServer) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	id := req.GetId()
	if id == "" {
		return &pb.DeleteRoleResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "ID is required",
		}, nil
	}

	var role model.Role
	result := s.DB.Table("roles").Where("id = ?", id).First(&role)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &pb.DeleteRoleResponse{
				StatusCode: int32(codes.NotFound),
				Message:    fmt.Sprintf("Role with ID %s not found", id),
			}, nil
		}
		return &pb.DeleteRoleResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to query role: %v", result.Error),
		}, nil
	}

	if err := s.DB.Table("roles").Delete(&model.Role{}, "id = ?", id).Error; err != nil {
		return &pb.DeleteRoleResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to delete role: %v", err),
		}, nil
	}

	return &pb.DeleteRoleResponse{
		StatusCode: int32(codes.OK),
		Message:    "Role deleted successfully",
	}, nil
}
