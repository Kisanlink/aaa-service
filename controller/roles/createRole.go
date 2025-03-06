package roles

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"github.com/Kisanlink/aaa-service/repositories"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RoleServer struct {
	pb.UnimplementedRoleServiceServer
	RoleRepo *repositories.RoleRepository
	PermissionRepo *repositories.PermissionRepository

}

func NewRoleServer(roleRepo *repositories.RoleRepository,permissionRepo *repositories.PermissionRepository) *RoleServer {
	return &RoleServer{
		RoleRepo: roleRepo,
		PermissionRepo: permissionRepo,

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
	

	roles, err := s.RoleRepo.FindAllRoles(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to retrieve roles: %v", err))
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}
	permissions, err := s.PermissionRepo.FindAllPermissions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to retrieve permissions: %v", err))
	}

	var permissionNames []string
	for _, permission := range permissions {
		permissionNames = append(permissionNames, permission.Name)
	}
	defaultRoles := []string{"test role"}
	defaultPermissions := []string{"test permission"}
	
	if len(roleNames) == 0 {
		roleNames = defaultRoles
	}
	
	if len(permissionNames) == 0 {
		permissionNames = defaultPermissions
	}

	updated, err := client.UpdateSchema(roleNames,permissionNames)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error reading schema: %v", err))
	}
	log.Printf("Updated Response: %+v", updated)
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
