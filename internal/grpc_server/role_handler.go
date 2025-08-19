package grpc_server

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/services"
	pb "github.com/Kisanlink/aaa-service/pkg/proto"
	"go.uber.org/zap"
)

// RoleHandler handles role-related gRPC requests
type RoleHandler struct {
	roleService *services.RoleService
	logger      *zap.Logger
	pb.UnimplementedRoleServiceV2Server
}

// NewRoleHandler creates a new RoleHandler
func NewRoleHandler(roleService *services.RoleService, logger *zap.Logger) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		logger:      logger,
	}
}

// CreateRole creates a new role
func (h *RoleHandler) CreateRole(ctx context.Context, req *pb.CreateRoleRequestV2) (*pb.CreateRoleResponseV2, error) {
	h.logger.Info("Creating role", zap.String("name", req.Name))

	// Create role model
	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
		Scope:       models.RoleScopeGlobal,
	}

	err := h.roleService.CreateRole(ctx, role)
	if err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		return &pb.CreateRoleResponseV2{
			StatusCode: 500,
			Message:    "Failed to create role: " + err.Error(),
		}, nil
	}

	// Convert service response to proto response
	protoRole := &pb.RoleV2{
		Id:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Status:      "active",
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return &pb.CreateRoleResponseV2{
		StatusCode: 200,
		Message:    "Role created successfully",
		Role:       protoRole,
	}, nil
}

// GetRole retrieves a role by ID
func (h *RoleHandler) GetRole(ctx context.Context, req *pb.GetRoleRequestV2) (*pb.GetRoleResponseV2, error) {
	h.logger.Info("Getting role", zap.String("id", req.Id))

	role, err := h.roleService.GetRoleByID(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get role", zap.Error(err))
		return &pb.GetRoleResponseV2{
			StatusCode: 404,
			Message:    "Role not found: " + err.Error(),
		}, nil
	}

	protoRole := &pb.RoleV2{
		Id:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Status:      "active",
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return &pb.GetRoleResponseV2{
		StatusCode: 200,
		Message:    "Role retrieved successfully",
		Role:       protoRole,
	}, nil
}

// GetAllRoles retrieves all roles with pagination
func (h *RoleHandler) GetAllRoles(ctx context.Context, req *pb.GetAllRolesRequestV2) (*pb.GetAllRolesResponseV2, error) {
	h.logger.Info("Getting all roles", zap.Int32("page", req.Page), zap.Int32("per_page", req.PerPage))

	// Use ListRoles method with pagination
	limit := int(req.PerPage)
	offset := (int(req.Page) - 1) * limit

	roles, err := h.roleService.ListRoles(ctx, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get roles", zap.Error(err))
		return &pb.GetAllRolesResponseV2{
			StatusCode: 500,
			Message:    "Failed to get roles: " + err.Error(),
		}, nil
	}

	// Convert service response to proto response
	var protoRoles []*pb.RoleV2
	for _, role := range roles {
		status := "inactive"
		if role.IsActive {
			status = "active"
		}

		protoRole := &pb.RoleV2{
			Id:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Status:      status,
			CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		protoRoles = append(protoRoles, protoRole)
	}

	return &pb.GetAllRolesResponseV2{
		StatusCode: 200,
		Message:    "Roles retrieved successfully",
		Roles:      protoRoles,
		TotalCount: int32(len(roles)),
		Page:       req.Page,
		PerPage:    req.PerPage,
	}, nil
}

// UpdateRole updates an existing role
func (h *RoleHandler) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequestV2) (*pb.UpdateRoleResponseV2, error) {
	h.logger.Info("Updating role", zap.String("id", req.Id))

	// Get existing role first
	existingRole, err := h.roleService.GetRoleByID(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get role for update", zap.Error(err))
		return &pb.UpdateRoleResponseV2{
			StatusCode: 404,
			Message:    "Role not found: " + err.Error(),
		}, nil
	}

	// Update fields
	existingRole.Name = req.Name
	existingRole.Description = req.Description

	err = h.roleService.UpdateRole(ctx, existingRole)
	if err != nil {
		h.logger.Error("Failed to update role", zap.Error(err))
		return &pb.UpdateRoleResponseV2{
			StatusCode: 500,
			Message:    "Failed to update role: " + err.Error(),
		}, nil
	}

	status := "inactive"
	if existingRole.IsActive {
		status = "active"
	}

	protoRole := &pb.RoleV2{
		Id:          existingRole.ID,
		Name:        existingRole.Name,
		Description: existingRole.Description,
		Status:      status,
		CreatedAt:   existingRole.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   existingRole.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return &pb.UpdateRoleResponseV2{
		StatusCode: 200,
		Message:    "Role updated successfully",
		Role:       protoRole,
	}, nil
}

// DeleteRole deletes a role
func (h *RoleHandler) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequestV2) (*pb.DeleteRoleResponseV2, error) {
	h.logger.Info("Deleting role", zap.String("id", req.Id))

	err := h.roleService.DeleteRole(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to delete role", zap.Error(err))
		return &pb.DeleteRoleResponseV2{
			StatusCode: 500,
			Message:    "Failed to delete role: " + err.Error(),
		}, nil
	}

	return &pb.DeleteRoleResponseV2{
		StatusCode: 200,
		Message:    "Role deleted successfully",
	}, nil
}

// AssignPermissionToRole assigns a permission to a role
func (h *RoleHandler) AssignPermissionToRole(ctx context.Context, req *pb.AssignPermissionToRoleRequestV2) (*pb.AssignPermissionToRoleResponseV2, error) {
	h.logger.Info("Assigning permission to role", zap.String("role_id", req.RoleId), zap.String("permission_id", req.PermissionId))

	// TODO: Implement permission assignment when service method is available
	h.logger.Warn("Permission assignment not yet implemented")

	connection := &pb.RolePermissionConnectionV2{
		RoleId:       req.RoleId,
		PermissionId: req.PermissionId,
	}

	return &pb.AssignPermissionToRoleResponseV2{
		StatusCode: 200,
		Message:    "Permission assigned to role successfully",
		Connection: connection,
	}, nil
}

// RemovePermissionFromRole removes a permission from a role
func (h *RoleHandler) RemovePermissionFromRole(ctx context.Context, req *pb.RemovePermissionFromRoleRequestV2) (*pb.RemovePermissionFromRoleResponseV2, error) {
	h.logger.Info("Removing permission from role", zap.String("role_id", req.RoleId), zap.String("permission_id", req.PermissionId))

	// TODO: Implement permission removal when service method is available
	h.logger.Warn("Permission removal not yet implemented")

	return &pb.RemovePermissionFromRoleResponseV2{
		StatusCode: 200,
		Message:    "Permission removed from role successfully",
	}, nil
}

// GetRolePermissions retrieves all permissions for a role
func (h *RoleHandler) GetRolePermissions(ctx context.Context, req *pb.GetRolePermissionsRequestV2) (*pb.GetRolePermissionsResponseV2, error) {
	h.logger.Info("Getting role permissions", zap.String("role_id", req.RoleId))

	// TODO: Implement permission retrieval when service method is available
	h.logger.Warn("Permission retrieval not yet implemented")

	return &pb.GetRolePermissionsResponseV2{
		StatusCode:  200,
		Message:     "Role permissions retrieved successfully",
		Permissions: []*pb.PermissionV2{},
	}, nil
}
