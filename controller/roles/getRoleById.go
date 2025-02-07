package roles

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

func (s *RoleServer) GetRoleById(ctx context.Context, req *pb.GetRoleByIdRequest) (*pb.GetRoleByIdResponse, error) {
	id := req.GetId()
	if id == "" {
		return &pb.GetRoleByIdResponse{
			StatusCode: int32(codes.InvalidArgument),
			Message:    "ID is required",
		}, nil
	}

	role := model.Role{}
	result := s.DB.Table("roles").Where("id = ?", id).First(&role)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &pb.GetRoleByIdResponse{
				StatusCode: int32(codes.NotFound),
				Message:    fmt.Sprintf("Role with ID %s not found", id),
			}, nil
		}
		return &pb.GetRoleByIdResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to query role: %v", result.Error),
		}, nil
	}

	pbRole := &pb.Role{
		Id:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}

	return &pb.GetRoleByIdResponse{
		StatusCode: int32(codes.OK),
		Message:    "Role retrieved successfully",
		Role:       pbRole,
	}, nil
}
