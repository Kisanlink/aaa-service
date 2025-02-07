package roles

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
)

func (s *RoleServer) GetAllRoles(ctx context.Context, req *pb.GetAllRolesRequest) (*pb.GetAllRolesResponse, error) {
	var roles []model.Role
	result := s.DB.Table("roles").Find(&roles)
	if result.Error != nil {
		return &pb.GetAllRolesResponse{
			StatusCode: int32(codes.Internal),
			Message:    fmt.Sprintf("Failed to retrieve roles: %v", result.Error),
		}, nil
	}

	var pbRoles []*pb.Role
	for _, role := range roles {
		pbRoles = append(pbRoles, &pb.Role{
			Id:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
		})
	}

	return &pb.GetAllRolesResponse{
		StatusCode: http.StatusOK,
		Message:    "Roles retrieved successfully",
		Roles:      pbRoles,
	}, nil
}
