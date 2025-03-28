package roles

import (
	"context"
	"net/http"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
)

func (s *RoleServer) GetAllRoles(ctx context.Context, req *pb.GetAllRolesRequest) (*pb.GetAllRolesResponse, error) {
	roles, err := s.RoleRepo.FindAllRoles(ctx)
	if err != nil {
		return nil, err
	}
	var pbRoles []*pb.Role
	for _, role := range roles {
		pbRoles = append(pbRoles, &pb.Role{
			Id:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Source: role.Source,
		})
	}
	return &pb.GetAllRolesResponse{
		StatusCode:http.StatusOK,
		Success: true,
		Message:    "Roles retrieved successfully",
		Data:      pbRoles,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format

	}, nil
}
