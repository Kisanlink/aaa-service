package permissions

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/kisanlink/protobuf/pb-aaa"
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
		Name:           req.Name,
		Description:    req.Description,
		Action:         req.Action,
		Source: req.Source,
		Resource: req.Resource,
		ValidStartTime: time.Now(),
		ValidEndTime:   time.Now(),
	}
	if err := s.PermissionRepo.CreatePermission(ctx, &newPermission); err != nil {
		log.Printf("Failed to create permission in database: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create permission")
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
	pbPermission := &pb.Permission{
		Id:            newPermission.ID,
		Name:          newPermission.Name,
		Description:   newPermission.Description,
		Action:        newPermission.Action,
		Source: newPermission.Source,
		Resource: newPermission.Resource,
		ValidStartTime: newPermission.ValidStartTime.Format(time.RFC3339Nano),
		ValiedEndTime:  newPermission.ValidEndTime.Format(time.RFC3339Nano),
	}

	return &pb.CreatePermissionResponse{
		StatusCode: http.StatusCreated,
		Success: true,
		Message:    "Permission created successfully",
		Data: pbPermission,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format
		}, nil
}
