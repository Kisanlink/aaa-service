package roles

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"github.com/Kisanlink/aaa-service/repositories"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RoleServer struct {
	pb.UnimplementedRoleServiceServer
	RoleRepo *repositories.RoleRepository
}

func NewRoleServer(roleRepo *repositories.RoleRepository) *RoleServer {
	return &RoleServer{
		RoleRepo: roleRepo,
	}
}

func (s *RoleServer) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	// Validate input
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Role Name is required")
	}

	// Check if the role already exists
	if err := s.RoleRepo.CheckIfRoleExists(ctx, req.Name); err != nil {
		return nil, err
	}

	// Create a new role object
	newRole := model.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.RoleRepo.CreateRole(ctx, &newRole); err != nil {
		return nil, err
	}
	pbRole := &pb.Role{
		Id:          newRole.ID,
		Name:        newRole.Name,
		Description: newRole.Description,
	}
	return &pb.CreateRoleResponse{
		StatusCode: int32(http.StatusCreated),
		Message:    "Role created successfully",
		Role:       pbRole,
	}, nil
}
