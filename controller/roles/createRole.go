package roles

import (
	"context"
	"strings"
	"time"

	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/kisanlink/protobuf/pb-aaa"
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
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Role Name is required")
	}
	if err := s.RoleRepo.CheckIfRoleExists(ctx, req.Name); err != nil {
		return nil, err
	}

	newRole := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Source: req.Source,
	}

	if err := s.RoleRepo.CreateRole(ctx, &newRole); err != nil {
		return nil, err
	}
	
	roles, err := s.RoleRepo.FindAllRoles(ctx)
	if err != nil {
		log.Printf("Failed to fetch roles: %v", err)
		return nil, status.Error(codes.Internal, "Failed to retrieve roles")
	}
	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}
	permissions, err := s.PermissionRepo.FindAllPermissions(ctx)
	if err != nil {
		log.Printf("Failed to fetch permissions: %v", err)
		return nil, status.Error(codes.Internal, "Failed to retrieve permissions")
	}
	var permissionNames []string
	var allActions []string
	actionSet := make(map[string]struct{})
	
	for _, permission := range permissions {
		permissionNames = append(permissionNames, permission.Name)
		actionSet[permission.Action] = struct{}{}
	}
	
	for action := range actionSet {
		allActions = append(allActions, action)
	}
	
	for i, action := range allActions {
		allActions[i] = strings.ToLower(action)
	}
	defaultRoles := []string{"test role"}
	defaultPermissions := []string{"test permission"}
	defaultActions := []string{"test action"}

	if len(roleNames) == 0 {
		roleNames = defaultRoles
	}
	if len(permissionNames) == 0 {
		permissionNames = defaultPermissions
	}
	if len(allActions) == 0 {
		allActions = defaultActions
	}
	updated, err := client.UpdateSchema(roleNames, permissionNames, allActions)
	if err != nil {
		log.Printf("Failed to update schema: %v", err)
		return nil, status.Error(codes.Internal, "Failed to update schema")
	}
	log.Printf("Schema updated successfully: %+v", updated)
	pbRole := &pb.Role{
		Id:          newRole.ID,
		Name:        newRole.Name,
		Description: newRole.Description,
		Source: newRole.Source,
	}
	return &pb.CreateRoleResponse{
		StatusCode: http.StatusCreated,
		Success: true,
		Message:    "Role created successfully",
		Data:       pbRole,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format

	}, nil
}
