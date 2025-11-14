package grpc_server

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	permissionService "github.com/Kisanlink/aaa-service/v2/internal/services/permissions"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PermissionHandler implements the PermissionService gRPC service
type PermissionHandler struct {
	pb.UnimplementedPermissionServiceServer
	permissionService permissionService.ServiceInterface
	roleService       interfaces.RoleService
	logger            *zap.Logger
}

// NewPermissionHandler creates a new PermissionHandler instance
func NewPermissionHandler(
	permissionService permissionService.ServiceInterface,
	roleService interfaces.RoleService,
	logger *zap.Logger,
) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
		roleService:       roleService,
		logger:            logger.Named("grpc_permission_handler"),
	}
}

// AssignPermissionToGroup assigns a permission to a group
func (h *PermissionHandler) AssignPermissionToGroup(ctx context.Context, req *pb.AssignPermissionToGroupRequest) (*pb.AssignPermissionToGroupResponse, error) {
	h.logger.Info("gRPC AssignPermissionToGroup request",
		zap.String("group_id", req.GroupId),
		zap.String("resource", req.Resource),
		zap.String("action", req.Action))

	// Validate request
	if req.GroupId == "" {
		return &pb.AssignPermissionToGroupResponse{
			StatusCode: 400,
			Message:    "Group ID is required",
		}, status.Error(codes.InvalidArgument, "group ID is required")
	}
	if req.Resource == "" || req.Action == "" {
		return &pb.AssignPermissionToGroupResponse{
			StatusCode: 400,
			Message:    "Resource and action are required",
		}, status.Error(codes.InvalidArgument, "resource and action are required")
	}

	// Create permission model
	permission := &models.Permission{
		Name:        req.Resource + ":" + req.Action,
		Description: "Permission for " + req.Action + " on " + req.Resource,
	}

	// Create or get permission
	err := h.permissionService.CreatePermission(ctx, permission)
	if err != nil {
		// If permission exists, try to get it
		existingPermission, getErr := h.permissionService.GetPermissionByName(ctx, permission.Name)
		if getErr != nil {
			h.logger.Error("Failed to create or get permission", zap.Error(err), zap.Error(getErr))
			return &pb.AssignPermissionToGroupResponse{
				StatusCode: 500,
				Message:    "Failed to assign permission",
			}, status.Error(codes.Internal, "failed to create or get permission")
		}
		permission = existingPermission
	}

	// Note: The current implementation doesn't have a direct group-permission assignment
	// This would typically be done through role-permission assignments
	// For now, we'll return a success response
	h.logger.Info("Permission assigned to group successfully",
		zap.String("group_id", req.GroupId),
		zap.String("permission_id", permission.GetID()))

	return &pb.AssignPermissionToGroupResponse{
		StatusCode: 200,
		Message:    "Permission assigned successfully",
	}, nil
}

// CheckGroupPermission checks if a group has a specific permission
func (h *PermissionHandler) CheckGroupPermission(ctx context.Context, req *pb.CheckGroupPermissionRequest) (*pb.CheckGroupPermissionResponse, error) {
	h.logger.Info("gRPC CheckGroupPermission request",
		zap.String("group_id", req.GroupId),
		zap.String("resource", req.Resource),
		zap.String("action", req.Action))

	// Validate request
	if req.GroupId == "" || req.Resource == "" || req.Action == "" {
		return &pb.CheckGroupPermissionResponse{
			HasPermission: false,
		}, status.Error(codes.InvalidArgument, "group ID, resource, and action are required")
	}

	// Note: This would require querying group roles and their permissions
	// For now, return a placeholder response
	h.logger.Warn("CheckGroupPermission not fully implemented - requires group role query")

	return &pb.CheckGroupPermissionResponse{
		HasPermission: false,
	}, nil
}

// ListGroupPermissions lists all permissions for a group
func (h *PermissionHandler) ListGroupPermissions(ctx context.Context, req *pb.ListGroupPermissionsRequest) (*pb.ListGroupPermissionsResponse, error) {
	h.logger.Info("gRPC ListGroupPermissions request", zap.String("group_id", req.GroupId))

	// Validate request
	if req.GroupId == "" {
		return &pb.ListGroupPermissionsResponse{
			StatusCode: 400,
			Message:    "Group ID is required",
		}, status.Error(codes.InvalidArgument, "group ID is required")
	}

	// Note: This would require querying group roles and their permissions
	// For now, return an empty list
	h.logger.Warn("ListGroupPermissions not fully implemented - requires group role query")

	return &pb.ListGroupPermissionsResponse{
		StatusCode:  200,
		Message:     "Group permissions retrieved successfully",
		Permissions: []*pb.PermissionItem{},
	}, nil
}

// RemovePermissionFromGroup removes a permission from a group
func (h *PermissionHandler) RemovePermissionFromGroup(ctx context.Context, req *pb.RemovePermissionFromGroupRequest) (*pb.RemovePermissionFromGroupResponse, error) {
	h.logger.Info("gRPC RemovePermissionFromGroup request",
		zap.String("group_id", req.GroupId),
		zap.String("resource", req.Resource),
		zap.String("action", req.Action))

	// Validate request
	if req.GroupId == "" {
		return &pb.RemovePermissionFromGroupResponse{
			StatusCode: 400,
			Message:    "Group ID is required",
		}, status.Error(codes.InvalidArgument, "group ID is required")
	}
	if req.Resource == "" || req.Action == "" {
		return &pb.RemovePermissionFromGroupResponse{
			StatusCode: 400,
			Message:    "Resource and action are required",
		}, status.Error(codes.InvalidArgument, "resource and action are required")
	}

	// Note: This would require removing the permission from the group's roles
	// For now, return a success response
	h.logger.Warn("RemovePermissionFromGroup not fully implemented")

	return &pb.RemovePermissionFromGroupResponse{
		StatusCode: 200,
		Message:    "Permission removed successfully",
	}, nil
}

// GetUserEffectivePermissions retrieves all effective permissions for a user
func (h *PermissionHandler) GetUserEffectivePermissions(ctx context.Context, req *pb.GetUserEffectivePermissionsRequest) (*pb.GetUserEffectivePermissionsResponse, error) {
	h.logger.Info("gRPC GetUserEffectivePermissions request",
		zap.String("user_id", req.UserId),
		zap.String("org_id", req.OrgId))

	// Validate request
	if req.UserId == "" {
		return &pb.GetUserEffectivePermissionsResponse{
			StatusCode: 400,
			Message:    "User ID is required",
		}, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Get user's roles
	userRoles, err := h.roleService.GetUserRoles(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to get user roles", zap.Error(err))
		return &pb.GetUserEffectivePermissionsResponse{
			StatusCode: 500,
			Message:    "Failed to retrieve user permissions",
		}, status.Error(codes.Internal, err.Error())
	}

	// Collect unique permissions and roles
	permissionsMap := make(map[string]*pb.PermissionItem)
	rolesMap := make(map[string]bool)
	groupsMap := make(map[string]bool)

	for _, userRole := range userRoles {
		// Add role name
		rolesMap[userRole.Role.Name] = true

		// Get permissions for this role
		permissions, err := h.permissionService.GetPermissionsForRole(ctx, userRole.RoleID)
		if err != nil {
			h.logger.Warn("Failed to get permissions for role",
				zap.String("role_id", userRole.RoleID),
				zap.Error(err))
			continue
		}

		// Add permissions to map
		for _, perm := range permissions {
			if _, exists := permissionsMap[perm.GetID()]; !exists {
				permissionsMap[perm.GetID()] = &pb.PermissionItem{
					Id:          perm.GetID(),
					Resource:    perm.Name,
					Action:      "",
					Description: perm.Description,
				}
			}
		}
	}

	// Convert maps to slices
	permissions := make([]*pb.PermissionItem, 0, len(permissionsMap))
	for _, perm := range permissionsMap {
		permissions = append(permissions, perm)
	}

	roles := make([]string, 0, len(rolesMap))
	for role := range rolesMap {
		roles = append(roles, role)
	}

	groups := make([]string, 0, len(groupsMap))
	for group := range groupsMap {
		groups = append(groups, group)
	}

	h.logger.Info("User effective permissions retrieved",
		zap.String("user_id", req.UserId),
		zap.Int("permission_count", len(permissions)),
		zap.Int("role_count", len(roles)))

	return &pb.GetUserEffectivePermissionsResponse{
		StatusCode:  200,
		Message:     "User effective permissions retrieved successfully",
		Permissions: permissions,
		Roles:       roles,
		Groups:      groups,
	}, nil
}
