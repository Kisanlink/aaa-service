package permissions

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

type PermissionServer struct {
	pb.UnimplementedPermissionServiceServer
	PermissionRepo *repositories.PermissionRepository
	RoleRepo *repositories.RoleRepository

}

func NewPermissionServer(permissionRepo *repositories.PermissionRepository,roleRepo *repositories.RoleRepository) *PermissionServer {
	return &PermissionServer{
		PermissionRepo: permissionRepo,
		RoleRepo: roleRepo,

	}
}

func (s *PermissionServer) CreatePermission(ctx context.Context, req *pb.CreatePermissionRequest) (*pb.CreatePermissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "Permission cannot be nil")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Permission Name is required")
	}
	if err := s.PermissionRepo.CheckIfPermissionExists(ctx, req.Name); err != nil {
		return nil, err
	}

	newPermission := model.Permission{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.PermissionRepo.CreatePermission(ctx, &newPermission); err != nil {
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
	pbPermission := &pb.Permission{
		Id:          newPermission.ID,
		Name:        newPermission.Name,
		Description: newPermission.Description,
	}
	return &pb.CreatePermissionResponse{
		StatusCode: http.StatusCreated,
		Message:    "Permission created successfully",
		Permission: pbPermission,
	}, nil
}
