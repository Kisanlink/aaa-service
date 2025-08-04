package roles

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/entities/models"
	pb "github.com/Kisanlink/aaa-service/proto"
	"google.golang.org/grpc/codes"
)

func (s *RoleServer) GetAllRoles(ctx context.Context, req *pb.GetAllRolesRequest) (*pb.GetAllRolesResponse, error) {
	var roles []models.Role
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
			Id:          role.ID,
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
