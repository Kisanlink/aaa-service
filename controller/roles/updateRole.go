package roles

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/model"
	pb "github.com/Kisanlink/aaa-service/proto"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

func (s *RoleServer) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	role := req.GetRole()
	if role == nil {
		return &pb.UpdateRoleResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "Role cannot be nil",
		}, nil
	}

	if role.Id == "" {
		return &pb.UpdateRoleResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "Role ID is required",
		}, nil
	}

	existingRole := model.Role{}
	result := s.DB.Table("roles").Where("id = ?", role.Id).First(&existingRole)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &pb.UpdateRoleResponse{
				StatusCode: int32(codes.NotFound),
				Message:    fmt.Sprintf("Role with ID %s not found", role.Id),
			}, nil
		}
		return &pb.UpdateRoleResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to query role: %v", result.Error),
		}, nil
	}

	updatedRole := model.Role{
		Name:        role.Name,
		Description: role.Description,
	}
	if err := s.DB.Table("roles").Model(&model.Role{}).Where("id = ?", role.Id).Updates(updatedRole).Error; err != nil {
		return &pb.UpdateRoleResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to update role: %v", err),
		}, nil
	}

	pbRole := &pb.Role{
		Id:          existingRole.ID,
		Name:        updatedRole.Name,
		Description: updatedRole.Description,
	}

	return &pb.UpdateRoleResponse{
		StatusCode: int32(codes.OK),
		Message:    "Role updated successfully",
		Role:       pbRole,
	}, nil
}
