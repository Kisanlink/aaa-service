package roles

import (
	"context"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
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
		})
	}
	return &pb.GetAllRolesResponse{
		StatusCode: int32(codes.OK),
		Message:    "Roles retrieved successfully",
		Roles:      pbRoles,
	}, nil
}
