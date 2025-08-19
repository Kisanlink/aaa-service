package grpc_server

import (
	"context"
	"time"

	actionRequests "github.com/Kisanlink/aaa-service/internal/entities/requests/actions"
	"github.com/Kisanlink/aaa-service/internal/services"
	pb "github.com/Kisanlink/aaa-service/pkg/pb"
	"go.uber.org/zap"
)

// EnhancedRBACHandler handles enhanced RBAC gRPC requests
type EnhancedRBACHandler struct {
	logger        *zap.Logger
	actionService *services.ActionService
	roleService   *services.RoleService
	pb.UnimplementedEnhancedRBACServiceServer
}

// NewEnhancedRBACHandler creates a new EnhancedRBACHandler
func NewEnhancedRBACHandler(logger *zap.Logger, actionService *services.ActionService, roleService *services.RoleService) *EnhancedRBACHandler {
	return &EnhancedRBACHandler{
		logger:        logger,
		actionService: actionService,
		roleService:   roleService,
	}
}

// CreateAction creates a new action
func (h *EnhancedRBACHandler) CreateAction(ctx context.Context, req *pb.CreateActionRequest) (*pb.CreateActionResponse, error) {
	h.logger.Info("Creating action", zap.String("name", req.Name))

	// Convert proto request to service request
	actionReq := &actionRequests.CreateActionRequest{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		IsActive:    true,
	}

	// Call the action service
	actionResp, err := h.actionService.CreateAction(ctx, actionReq)
	if err != nil {
		h.logger.Error("Failed to create action", zap.Error(err))
		return &pb.CreateActionResponse{
			StatusCode: 500,
			Message:    "Failed to create action: " + err.Error(),
		}, nil
	}

	// Convert service response to proto response
	action := &pb.Action{
		Id:          actionResp.ID,
		Name:        actionResp.Name,
		Description: actionResp.Description,
		Category:    actionResp.Category,
		IsActive:    actionResp.IsActive,
		CreatedAt:   actionResp.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   actionResp.UpdatedAt.Format(time.RFC3339),
	}

	return &pb.CreateActionResponse{
		StatusCode: 200,
		Message:    "Action created successfully",
		Action:     action,
	}, nil
}

// CreateResource creates a new resource
func (h *EnhancedRBACHandler) CreateResource(ctx context.Context, req *pb.CreateResourceRequest) (*pb.CreateResourceResponse, error) {
	h.logger.Info("Creating resource", zap.String("name", req.Name))

	// TODO: Implement when resource service is available
	h.logger.Warn("Resource creation not yet implemented")

	resource := &pb.Resource{
		Id:          "temp_resource_id",
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
	}

	return &pb.CreateResourceResponse{
		Resource: resource,
	}, nil
}

// CreatePermission creates a new permission
func (h *EnhancedRBACHandler) CreatePermission(ctx context.Context, req *pb.CreatePermissionRequest) (*pb.CreatePermissionResponse, error) {
	h.logger.Info("Creating permission", zap.String("name", req.Name))

	// TODO: Implement when permission service is available
	h.logger.Warn("Permission creation not yet implemented")

	permission := &pb.Permission{
		Id:          "temp_permission_id",
		Name:        req.Name,
		Description: req.Description,
		ResourceId:  req.ResourceId,
		ActionId:    req.ActionId,
	}

	return &pb.CreatePermissionResponse{
		Permission: permission,
	}, nil
}

// CreateRole creates a new role
func (h *EnhancedRBACHandler) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	h.logger.Info("Creating role", zap.String("name", req.Name))

	// TODO: Implement when role service is available
	h.logger.Warn("Role creation not yet implemented")

	role := &pb.Role{
		Id:          "temp_role_id",
		Name:        req.Name,
		Description: req.Description,
	}

	return &pb.CreateRoleResponse{
		Role: role,
	}, nil
}

// AssignRoleToUser assigns a role to a user
func (h *EnhancedRBACHandler) AssignRoleToUser(ctx context.Context, req *pb.AssignRoleToUserRequest) (*pb.AssignRoleToUserResponse, error) {
	h.logger.Info("Assigning role to user", zap.String("user_id", req.UserId), zap.String("role_id", req.RoleId))

	// Create a mock UserRole for now
	// In a real implementation, this would call the user role service
	userRole := &pb.UserRole{
		Id:        "temp_user_role_id",
		UserId:    req.UserId,
		RoleId:    req.RoleId,
		IsActive:  true,
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	return &pb.AssignRoleToUserResponse{
		StatusCode: 200,
		Message:    "Role assigned successfully",
		UserRole:   userRole,
	}, nil
}
