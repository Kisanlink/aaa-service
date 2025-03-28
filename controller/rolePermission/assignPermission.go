package rolepermission

import (
	"context"
	"log"
	"net/http"
	"time"

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
    if req.Role == "" || len(req.GetPermissions()) == 0 {
        return nil, status.Error(codes.InvalidArgument, "Both role_name and permission_names are required")
    }
	
    role, err := s.RoleRepo.GetRoleByName(ctx, req.Role)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "Role with name %s not found", req.Role)
    }
    roleID := role.ID
    permissionIDs := make([]string, 0)
    for _, permissionName := range req.GetPermissions() {
        permission, err := s.PermissionRepo.FindPermissionByName(ctx, permissionName)
        if err != nil {
            return nil, status.Errorf(codes.NotFound, "Permission with name %s not found", permissionName)
        }
        permissionIDs = append(permissionIDs, permission.ID)

		existing, err := s.RolePermissionRepo.GetRolePermissionByNames(ctx, req.Role, permissionName)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "Failed to check existing role-permission connection: %v", err)
        }
        if existing != nil {
            return nil, status.Errorf(
                codes.AlreadyExists, 
                "Permission '%s' is already assigned to role '%s'", 
                permissionName, 
                req.Role,
            )
        }
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
    
    roles, permissions, actions, usernames, err := s.userRepo.FindRoleUsersAndPermissionsByRoleId(ctx, roleID)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
    }
    log.Println(roles, permissions, actions, usernames)
    
    // Process each user one by one
    for _, username := range usernames {
        deleteResponse, err := client.DeleteUserRoleRelationship(
            username,
            roles,
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

    // Fetch role permissions with permission details
    fetchedRolePermissions, err := s.RolePermissionRepo.GetRolePermissionsByRoleID(ctx, role.ID)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "Failed to fetch role-permission connections: %v", err)
    }

    if len(fetchedRolePermissions) == 0 {
        return nil, status.Error(codes.NotFound, "No permissions found for this role")
    }

    // Convert to pointer slice
    var rolePermissionPtrs []*model.RolePermission
    for i := range fetchedRolePermissions {
        rolePermissionPtrs = append(rolePermissionPtrs, &fetchedRolePermissions[i])
    }

    response := &pb.ConnRolePermissionResponse{
        Id:         rolePermissionPtrs[0].ID,
        CreatedAt:  rolePermissionPtrs[0].CreatedAt.Format(time.RFC3339Nano),
        UpdatedAt:  rolePermissionPtrs[0].UpdatedAt.Format(time.RFC3339Nano),
        Role: &pb.ConnRole{
            Id:          role.ID,
            Name:        role.Name,
            Description: role.Description,
            Source:      role.Source,
            CreatedAt:   role.CreatedAt.Format(time.RFC3339Nano),
            UpdatedAt:   role.UpdatedAt.Format(time.RFC3339Nano),
        },
        Permissions: []*pb.ConnPermission{},
        IsActive:   rolePermissionPtrs[0].IsActive,
    }

    for _, rp := range rolePermissionPtrs {
        if !IsZeroValued(rp.Permission) && rp.Permission.ID != "" {
            response.Permissions = append(response.Permissions, &pb.ConnPermission{
                Id:             rp.Permission.ID,
                Name:           rp.Permission.Name,
                Description:    rp.Permission.Description,
                Action:         rp.Permission.Action,
                Resource:       rp.Permission.Resource,
                Source:         rp.Permission.Source,
                ValidStartTime: rp.Permission.ValidStartTime.Format(time.RFC3339Nano),
                ValidEndTime:   rp.Permission.ValidEndTime.Format(time.RFC3339Nano),
                CreatedAt:      rp.Permission.CreatedAt.Format(time.RFC3339Nano),
                UpdatedAt:      rp.Permission.UpdatedAt.Format(time.RFC3339Nano),
            })
        }
    }

    return &pb.CreateConnRolePermissionResponse{
        StatusCode:    http.StatusCreated,
        Success:       true,
        Message:       "Role with Permission created successfully",
        Data:          response,
        DataTimeStamp: time.Now().Format(time.RFC3339Nano),
    }, nil
}
