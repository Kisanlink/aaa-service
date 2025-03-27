package rolepermission

import (
	"context"
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConnectRolePermissionServer struct {
	pb.UnimplementedConnectRolePermissionServiceServer
	RolePermissionRepo *repositories.RolePermissionRepository
	RoleRepo           *repositories.RoleRepository
	PermissionRepo     *repositories.PermissionRepository
	userRepo *repositories.UserRepository
}

func NewConnectRolePermissionServer(
	rolePermissionRepo *repositories.RolePermissionRepository,
	roleRepo *repositories.RoleRepository,
	permissionRepo *repositories.PermissionRepository,
	userRepo *repositories.UserRepository,
) *ConnectRolePermissionServer {
	return &ConnectRolePermissionServer{
		RolePermissionRepo: rolePermissionRepo,
		RoleRepo:           roleRepo,
		userRepo:           userRepo,
		PermissionRepo:     permissionRepo,
	}
}

func (s *ConnectRolePermissionServer) AssignPermission(ctx context.Context, req *pb.CreateConnRolePermissionRequest) (*pb.CreateConnRolePermissionResponse, error) {
	if req.Roles == "" || len(req.GetPermissions()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Both role_name and permission_names are required")
	}
	role, err := s.RoleRepo.GetRoleByName(ctx, req.Roles)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role with name %s not found", req.Roles)
	}
	roleID := role.ID
	permissionIDs := make([]string, 0)
	for _, permissionName := range req.GetPermissions() {
		permission, err := s.PermissionRepo.FindPermissionByName(ctx, permissionName)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "Permission with name %s not found", permissionName)
		}
		permissionIDs = append(permissionIDs, permission.ID)
	}

	var rolePermissions []*model.RolePermission
	for _, permissionID := range permissionIDs {
		rolePermission := &model.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
			IsActive:     true,
		}
		rolePermissions = append(rolePermissions, rolePermission)
	}

	if err := s.RolePermissionRepo.CreateRolePermissions(ctx, rolePermissions); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create role-permission connections: %v", err)
	}

	var connRolePermissions []*pb.ConnRolePermission
	for _, rp := range rolePermissions {
		connRolePermission := &pb.ConnRolePermission{
			Id:           rp.ID,
			CreatedAt:    rp.CreatedAt.String(),
			UpdatedAt:    rp.UpdatedAt.String(),
			RoleId:       rp.RoleID,
			PermissionId: rp.PermissionID,
			IsActive:     rp.IsActive,
		}
		connRolePermissions = append(connRolePermissions, connRolePermission)
	}
	roles, permissions, actions, usernames, err := s.userRepo.FindRoleUsersAndPermissionsByRoleId(ctx, roleID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}
	log.Println(roles,permissions,actions,usernames)
	// Process each user one by one
	for _, username := range usernames {
		deleteResponse, err := client.DeleteUserRoleRelationship(
			username,
			 (roles),
			helper.LowerCaseSlice(permissions),
			helper.LowerCaseSlice(actions),
		)
		if err != nil {
			log.Printf("Failed to delete relationships for user %s: %v", username, err)
			continue 
		}
		log.Printf("User roles and permissions deleted successfully for %s: %s", username, deleteResponse)
			createResponse, err := client.CreateUserRoleRelationship(
			username,
			helper.LowerCaseSlice(roles),
			helper.LowerCaseSlice(permissions),
			helper.LowerCaseSlice(actions),
		)
		if err != nil {
			log.Printf("Failed to create relationships for user %s: %v", username, err)
			continue
		}
		log.Printf("Relationships created successfully for %s: %v", username, createResponse)
	}
	return &pb.CreateConnRolePermissionResponse{
		StatusCode: http.StatusCreated,
		Success: true,
		Message:    "Role with Permission created successfully",
		Data:       connRolePermissions,
	}, nil
}