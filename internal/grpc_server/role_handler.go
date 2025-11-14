package grpc_server

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RoleHandler implements the RoleService gRPC service
type RoleHandler struct {
	pb.UnimplementedRoleServiceServer
	roleService interfaces.RoleService
	logger      *zap.Logger
}

// NewRoleHandler creates a new RoleHandler instance
func NewRoleHandler(roleService interfaces.RoleService, logger *zap.Logger) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		logger:      logger,
	}
}

// AssignRole assigns a role to a user in an organization context
func (h *RoleHandler) AssignRole(ctx context.Context, req *pb.AssignRoleRequest) (*pb.AssignRoleResponse, error) {
	h.logger.Info("gRPC AssignRole request",
		zap.String("user_id", req.UserId),
		zap.String("org_id", req.OrgId),
		zap.String("role_name", req.RoleName))

	// Validate request
	if req.UserId == "" {
		return &pb.AssignRoleResponse{
			StatusCode: 400,
			Message:    "User ID is required",
		}, status.Error(codes.InvalidArgument, "user ID is required")
	}
	if req.RoleName == "" {
		return &pb.AssignRoleResponse{
			StatusCode: 400,
			Message:    "Role name is required",
		}, status.Error(codes.InvalidArgument, "role name is required")
	}

	// Get role by name
	role, err := h.roleService.GetRoleByName(ctx, req.RoleName)
	if err != nil {
		h.logger.Error("Failed to get role by name", zap.String("role_name", req.RoleName), zap.Error(err))
		return &pb.AssignRoleResponse{
			StatusCode: 404,
			Message:    "Role not found",
		}, status.Error(codes.NotFound, "role not found")
	}

	// Assign role to user
	err = h.roleService.AssignRoleToUser(ctx, req.UserId, role.GetID())
	if err != nil {
		h.logger.Error("Failed to assign role", zap.Error(err))
		return &pb.AssignRoleResponse{
			StatusCode: 500,
			Message:    "Failed to assign role",
		}, status.Error(codes.Internal, err.Error())
	}

	h.logger.Info("Role assigned successfully",
		zap.String("user_id", req.UserId),
		zap.String("role_name", req.RoleName))

	return &pb.AssignRoleResponse{
		StatusCode: 200,
		Message:    "Role assigned successfully",
	}, nil
}

// CheckUserRole checks if a user has a specific role
func (h *RoleHandler) CheckUserRole(ctx context.Context, req *pb.CheckUserRoleRequest) (*pb.CheckUserRoleResponse, error) {
	h.logger.Info("gRPC CheckUserRole request",
		zap.String("user_id", req.UserId),
		zap.String("role_name", req.RoleName))

	// Validate request
	if req.UserId == "" || req.RoleName == "" {
		return &pb.CheckUserRoleResponse{
			HasRole: false,
		}, status.Error(codes.InvalidArgument, "user ID and role name are required")
	}

	// Get user roles
	userRoles, err := h.roleService.GetUserRoles(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to get user roles", zap.Error(err))
		return &pb.CheckUserRoleResponse{
			HasRole: false,
		}, status.Error(codes.Internal, err.Error())
	}

	// Check if user has the role
	hasRole := false
	for _, userRole := range userRoles {
		if userRole.Role.Name == req.RoleName {
			hasRole = true
			break
		}
	}

	return &pb.CheckUserRoleResponse{
		HasRole: hasRole,
		OrgId:   req.OrgId,
	}, nil
}

// RemoveRole removes a role from a user
func (h *RoleHandler) RemoveRole(ctx context.Context, req *pb.RemoveRoleRequest) (*pb.RemoveRoleResponse, error) {
	h.logger.Info("gRPC RemoveRole request",
		zap.String("user_id", req.UserId),
		zap.String("role_name", req.RoleName))

	// Validate request
	if req.UserId == "" {
		return &pb.RemoveRoleResponse{
			StatusCode: 400,
			Message:    "User ID is required",
		}, status.Error(codes.InvalidArgument, "user ID is required")
	}
	if req.RoleName == "" {
		return &pb.RemoveRoleResponse{
			StatusCode: 400,
			Message:    "Role name is required",
		}, status.Error(codes.InvalidArgument, "role name is required")
	}

	// Get role by name
	role, err := h.roleService.GetRoleByName(ctx, req.RoleName)
	if err != nil {
		h.logger.Error("Failed to get role by name", zap.String("role_name", req.RoleName), zap.Error(err))
		return &pb.RemoveRoleResponse{
			StatusCode: 404,
			Message:    "Role not found",
		}, status.Error(codes.NotFound, "role not found")
	}

	// Remove role from user
	err = h.roleService.RemoveRoleFromUser(ctx, req.UserId, role.GetID())
	if err != nil {
		h.logger.Error("Failed to remove role", zap.Error(err))
		return &pb.RemoveRoleResponse{
			StatusCode: 500,
			Message:    "Failed to remove role",
		}, status.Error(codes.Internal, err.Error())
	}

	h.logger.Info("Role removed successfully",
		zap.String("user_id", req.UserId),
		zap.String("role_name", req.RoleName))

	return &pb.RemoveRoleResponse{
		StatusCode: 200,
		Message:    "Role removed successfully",
	}, nil
}

// GetUserRoles retrieves all roles for a user
func (h *RoleHandler) GetUserRoles(ctx context.Context, req *pb.GetUserRolesRequest) (*pb.GetUserRolesResponse, error) {
	h.logger.Info("gRPC GetUserRoles request", zap.String("user_id", req.UserId))

	// Validate request
	if req.UserId == "" {
		return &pb.GetUserRolesResponse{
			StatusCode: 400,
			Message:    "User ID is required",
		}, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Get user roles from service
	userRoles, err := h.roleService.GetUserRoles(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to get user roles", zap.Error(err))
		return &pb.GetUserRolesResponse{
			StatusCode: 500,
			Message:    "Failed to retrieve user roles",
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf format
	pbRoles := make([]*pb.RoleAssignment, len(userRoles))
	for i, userRole := range userRoles {
		pbRoles[i] = &pb.RoleAssignment{
			RoleName:   userRole.Role.Name,
			OrgId:      req.OrgId, // Use org_id from request as context
			OrgName:    "",        // Not available in current model
			AssignedAt: timestamppb.New(userRole.GetCreatedAt()),
		}
	}

	return &pb.GetUserRolesResponse{
		StatusCode: 200,
		Message:    "User roles retrieved successfully",
		Roles:      pbRoles,
	}, nil
}

// ListUsersWithRole lists all users who have a specific role
func (h *RoleHandler) ListUsersWithRole(ctx context.Context, req *pb.ListUsersWithRoleRequest) (*pb.ListUsersWithRoleResponse, error) {
	h.logger.Info("gRPC ListUsersWithRole request",
		zap.String("role_name", req.RoleName),
		zap.String("org_id", req.OrgId))

	// Validate request
	if req.RoleName == "" {
		return &pb.ListUsersWithRoleResponse{
			StatusCode: 400,
			Message:    "Role name is required",
		}, status.Error(codes.InvalidArgument, "role name is required")
	}

	// Get role by name
	role, err := h.roleService.GetRoleByName(ctx, req.RoleName)
	if err != nil {
		h.logger.Error("Failed to get role by name", zap.String("role_name", req.RoleName), zap.Error(err))
		return &pb.ListUsersWithRoleResponse{
			StatusCode: 404,
			Message:    "Role not found",
		}, status.Error(codes.NotFound, "role not found")
	}

	// Note: The current RoleService interface doesn't have a method to list users by role
	// This would require querying the UserRoleRepository
	// For now, return a not implemented response
	h.logger.Warn("ListUsersWithRole not fully implemented - requires UserRoleRepository integration",
		zap.String("role_id", role.GetID()))

	return &pb.ListUsersWithRoleResponse{
		StatusCode: 501,
		Message:    "ListUsersWithRole not yet implemented",
		Users:      []*pb.UserSummary{},
		TotalCount: 0,
	}, status.Error(codes.Unimplemented, "list users with role not implemented")
}
