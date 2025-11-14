package grpc_server

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	actionRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/actions"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/services"
	"github.com/Kisanlink/aaa-service/v2/internal/services/catalog"
	permissionService "github.com/Kisanlink/aaa-service/v2/internal/services/permissions"
	resourceService "github.com/Kisanlink/aaa-service/v2/internal/services/resources"
	roleAssignmentService "github.com/Kisanlink/aaa-service/v2/internal/services/role_assignments"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CatalogHandler implements the gRPC CatalogService interface
type CatalogHandler struct {
	pb.UnimplementedCatalogServiceServer
	catalogService        *catalog.CatalogService
	actionService         *services.ActionService
	resourceService       resourceService.ResourceService
	roleService           interfaces.RoleService
	permissionService     permissionService.ServiceInterface
	roleAssignmentService roleAssignmentService.ServiceInterface
	authChecker           *AuthorizationChecker
	logger                *zap.Logger
}

// NewCatalogHandler creates a new catalog handler
func NewCatalogHandler(
	catalogService *catalog.CatalogService,
	authChecker *AuthorizationChecker,
	logger *zap.Logger,
) *CatalogHandler {
	return &CatalogHandler{
		catalogService: catalogService,
		authChecker:    authChecker,
		logger:         logger,
	}
}

// SetActionService sets the action service (dependency injection after construction)
func (ch *CatalogHandler) SetActionService(actionService *services.ActionService) {
	ch.actionService = actionService
}

// SetResourceService sets the resource service
func (ch *CatalogHandler) SetResourceService(resourceService resourceService.ResourceService) {
	ch.resourceService = resourceService
}

// SetRoleService sets the role service
func (ch *CatalogHandler) SetRoleService(roleService interfaces.RoleService) {
	ch.roleService = roleService
}

// SetPermissionService sets the permission service
func (ch *CatalogHandler) SetPermissionService(permissionService permissionService.ServiceInterface) {
	ch.permissionService = permissionService
}

// SetRoleAssignmentService sets the role assignment service
func (ch *CatalogHandler) SetRoleAssignmentService(roleAssignmentService roleAssignmentService.ServiceInterface) {
	ch.roleAssignmentService = roleAssignmentService
}

// Helper function to safely dereference string pointers
func getStringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// SeedRolesAndPermissions seeds the database with predefined roles and permissions
func (ch *CatalogHandler) SeedRolesAndPermissions(
	ctx context.Context,
	req *pb.SeedRolesAndPermissionsRequest,
) (*pb.SeedRolesAndPermissionsResponse, error) {
	ch.logger.Info("SeedRolesAndPermissions called",
		zap.Bool("force", req.Force),
		zap.String("organization_id", req.OrganizationId),
		zap.String("service_id", req.ServiceId))

	// Validate service_id before processing
	if err := catalog.ValidateServiceID(req.ServiceId); err != nil {
		ch.logger.Error("Invalid service_id",
			zap.String("service_id", req.ServiceId),
			zap.Error(err))

		return &pb.SeedRolesAndPermissionsResponse{
			StatusCode:         400,
			Message:            fmt.Sprintf("Invalid service_id: %v", err),
			RolesCreated:       0,
			PermissionsCreated: 0,
			ResourcesCreated:   0,
			ActionsCreated:     0,
			CreatedRoles:       []string{},
		}, fmt.Errorf("invalid service_id: %w", err)
	}

	// Check authorization: verify caller has permission to seed roles
	// This validates both basic seed permission and service ownership rules
	if ch.authChecker != nil {
		if err := ch.authChecker.CheckSeedPermission(ctx, req.ServiceId); err != nil {
			ch.logger.Warn("Seed operation authorization failed",
				zap.String("service_id", req.ServiceId),
				zap.Error(err))

			return &pb.SeedRolesAndPermissionsResponse{
				StatusCode:         403,
				Message:            fmt.Sprintf("Authorization failed: %v", err),
				RolesCreated:       0,
				PermissionsCreated: 0,
				ResourcesCreated:   0,
				ActionsCreated:     0,
				CreatedRoles:       []string{},
			}, err
		}

		ch.logger.Info("Seed operation authorized",
			zap.String("service_id", req.ServiceId),
			zap.Bool("force", req.Force))
	} else {
		ch.logger.Warn("Authorization checker not configured - skipping authorization check")
	}

	// Call the catalog service with service_id parameter
	// Empty service_id will default to farmers-module for backward compatibility
	result, err := ch.catalogService.SeedRolesAndPermissions(ctx, req.ServiceId, req.Force)
	if err != nil {
		ch.logger.Error("Failed to seed roles and permissions",
			zap.String("service_id", req.ServiceId),
			zap.Error(err))

		return &pb.SeedRolesAndPermissionsResponse{
			StatusCode:         500,
			Message:            fmt.Sprintf("Failed to seed roles and permissions: %v", err),
			RolesCreated:       0,
			PermissionsCreated: 0,
			ResourcesCreated:   0,
			ActionsCreated:     0,
			CreatedRoles:       []string{},
		}, err
	}

	// Build successful response
	response := &pb.SeedRolesAndPermissionsResponse{
		StatusCode:         200,
		Message:            "Successfully seeded roles and permissions",
		RolesCreated:       result.RolesCreated,
		PermissionsCreated: result.PermissionsCreated,
		ResourcesCreated:   result.ResourcesCreated,
		ActionsCreated:     result.ActionsCreated,
		CreatedRoles:       result.CreatedRoleNames,
	}

	ch.logger.Info("SeedRolesAndPermissions completed successfully",
		zap.Int32("roles_created", result.RolesCreated),
		zap.Int32("permissions_created", result.PermissionsCreated),
		zap.Int32("resources_created", result.ResourcesCreated),
		zap.Int32("actions_created", result.ActionsCreated),
		zap.Strings("created_roles", result.CreatedRoleNames))

	return response, nil
}

// RegisterAction registers a new action in the catalog
func (ch *CatalogHandler) RegisterAction(
	ctx context.Context,
	req *pb.RegisterActionRequest,
) (*pb.RegisterActionResponse, error) {
	ch.logger.Info("RegisterAction called",
		zap.String("name", req.Name))

	if ch.actionService == nil {
		return &pb.RegisterActionResponse{
			StatusCode: 503,
			Message:    "Action service not available",
		}, status.Error(codes.Unavailable, "action service not configured")
	}

	// Validate request
	if req.Name == "" {
		return &pb.RegisterActionResponse{
			StatusCode: 400,
			Message:    "Action name is required",
		}, status.Error(codes.InvalidArgument, "action name is required")
	}

	// Create action request
	serviceID := req.ServiceId
	createReq := &actionRequests.CreateActionRequest{
		Name:        req.Name,
		Description: req.Description,
		IsStatic:    req.IsStatic,
		ServiceID:   &serviceID,
		IsActive:    true,
	}

	// Create action via service
	actionResp, err := ch.actionService.CreateAction(ctx, createReq)
	if err != nil {
		ch.logger.Error("Failed to register action", zap.Error(err))
		return &pb.RegisterActionResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to register action: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf Action
	pbAction := &pb.Action{
		Id:          actionResp.ID,
		Name:        actionResp.Name,
		Description: actionResp.Description,
		IsStatic:    actionResp.IsStatic,
		ServiceId:   getStringOrEmpty(actionResp.ServiceID),
		Metadata:    req.Metadata, // Use request metadata
	}

	return &pb.RegisterActionResponse{
		StatusCode: 200,
		Message:    "Action registered successfully",
		Action:     pbAction,
	}, nil
}

// ListActions lists all actions in the catalog
func (ch *CatalogHandler) ListActions(
	ctx context.Context,
	req *pb.ListActionsRequest,
) (*pb.ListActionsResponse, error) {
	ch.logger.Info("ListActions called")

	if ch.actionService == nil {
		return &pb.ListActionsResponse{
			StatusCode: 503,
			Message:    "Action service not available",
		}, status.Error(codes.Unavailable, "action service not configured")
	}

	// Set default pagination
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// List actions via service
	actionsResp, err := ch.actionService.ListActions(ctx, pageSize, offset)
	if err != nil {
		ch.logger.Error("Failed to list actions", zap.Error(err))
		return &pb.ListActionsResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to list actions: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf format
	pbActions := make([]*pb.Action, len(actionsResp.Actions))
	for i, action := range actionsResp.Actions {
		pbActions[i] = &pb.Action{
			Id:          action.ID,
			Name:        action.Name,
			Description: action.Description,
			IsStatic:    action.IsStatic,
			ServiceId:   getStringOrEmpty(action.ServiceID),
			Metadata:    make(map[string]string), // Empty metadata for now
		}
	}

	return &pb.ListActionsResponse{
		StatusCode: 200,
		Message:    "Actions retrieved successfully",
		Actions:    pbActions,
		TotalCount: int32(actionsResp.Total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
	}, nil
}

// RegisterResource registers a new resource in the catalog
func (ch *CatalogHandler) RegisterResource(
	ctx context.Context,
	req *pb.RegisterResourceRequest,
) (*pb.RegisterResourceResponse, error) {
	ch.logger.Info("RegisterResource called",
		zap.String("name", req.Name))

	if ch.resourceService == nil {
		return &pb.RegisterResourceResponse{
			StatusCode: 503,
			Message:    "Resource service not available",
		}, status.Error(codes.Unavailable, "resource service not configured")
	}

	// Validate request
	if req.Name == "" {
		return &pb.RegisterResourceResponse{
			StatusCode: 400,
			Message:    "Resource name is required",
		}, status.Error(codes.InvalidArgument, "resource name is required")
	}

	// Create resource via service
	resource, err := ch.resourceService.Create(ctx, req.Name, req.Type, req.Description)
	if err != nil {
		ch.logger.Error("Failed to register resource", zap.Error(err))
		return &pb.RegisterResourceResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to register resource: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf Resource
	pbResource := &pb.Resource{
		Id:          resource.GetID(),
		Name:        resource.Name,
		Type:        resource.Type,
		Description: resource.Description,
		ParentId:    getStringOrEmpty(resource.ParentID),
		OwnerId:     getStringOrEmpty(resource.OwnerID),
	}

	return &pb.RegisterResourceResponse{
		StatusCode: 200,
		Message:    "Resource registered successfully",
		Resource:   pbResource,
	}, nil
}

// SetResourceParent sets the parent of a resource
func (ch *CatalogHandler) SetResourceParent(
	ctx context.Context,
	req *pb.SetResourceParentRequest,
) (*pb.SetResourceParentResponse, error) {
	ch.logger.Info("SetResourceParent called",
		zap.String("resource_id", req.ResourceId),
		zap.String("parent_id", req.ParentId))

	if ch.resourceService == nil {
		return &pb.SetResourceParentResponse{
			StatusCode: 503,
			Message:    "Resource service not available",
		}, status.Error(codes.Unavailable, "resource service not configured")
	}

	// Validate request
	if req.ResourceId == "" || req.ParentId == "" {
		return &pb.SetResourceParentResponse{
			StatusCode: 400,
			Message:    "Resource ID and parent ID are required",
		}, status.Error(codes.InvalidArgument, "resource ID and parent ID are required")
	}

	// Set parent via service
	err := ch.resourceService.SetParent(ctx, req.ResourceId, req.ParentId)
	if err != nil {
		ch.logger.Error("Failed to set resource parent", zap.Error(err))
		return &pb.SetResourceParentResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to set resource parent: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Get updated resource
	resource, err := ch.resourceService.GetByID(ctx, req.ResourceId)
	if err != nil {
		ch.logger.Warn("Failed to get updated resource", zap.Error(err))
		// Return success anyway as parent was set
		return &pb.SetResourceParentResponse{
			StatusCode: 200,
			Message:    "Resource parent set successfully",
		}, nil
	}

	// Convert to protobuf Resource
	pbResource := &pb.Resource{
		Id:          resource.GetID(),
		Name:        resource.Name,
		Type:        resource.Type,
		Description: resource.Description,
		ParentId:    getStringOrEmpty(resource.ParentID),
		OwnerId:     getStringOrEmpty(resource.OwnerID),
	}

	return &pb.SetResourceParentResponse{
		StatusCode: 200,
		Message:    "Resource parent set successfully",
		Resource:   pbResource,
	}, nil
}

// ListResources lists all resources in the catalog
func (ch *CatalogHandler) ListResources(
	ctx context.Context,
	req *pb.ListResourcesRequest,
) (*pb.ListResourcesResponse, error) {
	ch.logger.Info("ListResources called")

	if ch.resourceService == nil {
		return &pb.ListResourcesResponse{
			StatusCode: 503,
			Message:    "Resource service not available",
		}, status.Error(codes.Unavailable, "resource service not configured")
	}

	// Set default pagination
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// List resources via service
	resources, err := ch.resourceService.List(ctx, pageSize, offset)
	if err != nil {
		ch.logger.Error("Failed to list resources", zap.Error(err))
		return &pb.ListResourcesResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to list resources: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf format
	pbResources := make([]*pb.Resource, len(resources))
	for i, resource := range resources {
		pbResources[i] = &pb.Resource{
			Id:          resource.GetID(),
			Name:        resource.Name,
			Type:        resource.Type,
			Description: resource.Description,
			ParentId:    getStringOrEmpty(resource.ParentID),
			OwnerId:     getStringOrEmpty(resource.OwnerID),
		}
	}

	return &pb.ListResourcesResponse{
		StatusCode: 200,
		Message:    "Resources retrieved successfully",
		Resources:  pbResources,
		TotalCount: int32(len(resources)),
		Page:       int32(page),
		PageSize:   int32(pageSize),
	}, nil
}

// CreateRole creates a new role in the catalog
func (ch *CatalogHandler) CreateRole(
	ctx context.Context,
	req *pb.CreateRoleRequest,
) (*pb.CreateRoleResponse, error) {
	ch.logger.Info("CreateRole called",
		zap.String("name", req.Name))

	if ch.roleService == nil {
		return &pb.CreateRoleResponse{
			StatusCode: 503,
			Message:    "Role service not available",
		}, status.Error(codes.Unavailable, "role service not configured")
	}

	// Validate request
	if req.Name == "" {
		return &pb.CreateRoleResponse{
			StatusCode: 400,
			Message:    "Role name is required",
		}, status.Error(codes.InvalidArgument, "role name is required")
	}

	// Create role
	role := models.NewRole(req.Name, req.Description, models.RoleScopeOrg)
	err := ch.roleService.CreateRole(ctx, role)
	if err != nil {
		ch.logger.Error("Failed to create role", zap.Error(err))
		return &pb.CreateRoleResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to create role: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf CatalogRole
	pbRole := &pb.CatalogRole{
		Id:          role.GetID(),
		Name:        role.Name,
		Description: role.Description,
		Scope:       string(role.Scope),
		IsActive:    role.IsActive,
	}

	return &pb.CreateRoleResponse{
		StatusCode: 200,
		Message:    "Role created successfully",
		Role:       pbRole,
	}, nil
}

// ListRoles lists all roles in the catalog
func (ch *CatalogHandler) ListRoles(
	ctx context.Context,
	req *pb.ListRolesRequest,
) (*pb.ListRolesResponse, error) {
	ch.logger.Info("ListRoles called")

	if ch.roleService == nil {
		return &pb.ListRolesResponse{
			StatusCode: 503,
			Message:    "Role service not available",
		}, status.Error(codes.Unavailable, "role service not configured")
	}

	// Set default pagination
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// List roles via service
	roles, err := ch.roleService.ListRoles(ctx, pageSize, offset)
	if err != nil {
		ch.logger.Error("Failed to list roles", zap.Error(err))
		return &pb.ListRolesResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to list roles: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf format
	pbRoles := make([]*pb.CatalogRole, len(roles))
	for i, role := range roles {
		pbRoles[i] = &pb.CatalogRole{
			Id:          role.GetID(),
			Name:        role.Name,
			Description: role.Description,
			Scope:       string(role.Scope),
			IsActive:    role.IsActive,
		}
	}

	return &pb.ListRolesResponse{
		StatusCode: 200,
		Message:    "Roles retrieved successfully",
		Roles:      pbRoles,
		TotalCount: int32(len(roles)),
		Page:       int32(page),
		PageSize:   int32(pageSize),
	}, nil
}

// CreatePermission creates a new permission in the catalog
func (ch *CatalogHandler) CreatePermission(
	ctx context.Context,
	req *pb.CreatePermissionRequest,
) (*pb.CreatePermissionResponse, error) {
	ch.logger.Info("CreatePermission called",
		zap.String("name", req.Name))

	if ch.permissionService == nil {
		return &pb.CreatePermissionResponse{
			StatusCode: 503,
			Message:    "Permission service not available",
		}, status.Error(codes.Unavailable, "permission service not configured")
	}

	// Validate request
	if req.Name == "" {
		return &pb.CreatePermissionResponse{
			StatusCode: 400,
			Message:    "Permission name is required",
		}, status.Error(codes.InvalidArgument, "permission name is required")
	}

	// Create permission
	resourceID := req.ResourceId
	actionID := req.ActionId
	permission := &models.Permission{
		Name:        req.Name,
		Description: req.Description,
		ResourceID:  &resourceID,
		ActionID:    &actionID,
	}
	err := ch.permissionService.CreatePermission(ctx, permission)
	if err != nil {
		ch.logger.Error("Failed to create permission", zap.Error(err))
		return &pb.CreatePermissionResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to create permission: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf CatalogPermission
	pbPermission := &pb.CatalogPermission{
		Id:          permission.GetID(),
		Name:        permission.Name,
		Description: permission.Description,
		ResourceId:  getStringOrEmpty(permission.ResourceID),
		ActionId:    getStringOrEmpty(permission.ActionID),
	}

	return &pb.CreatePermissionResponse{
		StatusCode: 200,
		Message:    "Permission created successfully",
		Permission: pbPermission,
	}, nil
}

// AttachPermissions attaches permissions to a role
func (ch *CatalogHandler) AttachPermissions(
	ctx context.Context,
	req *pb.AttachPermissionsRequest,
) (*pb.AttachPermissionsResponse, error) {
	ch.logger.Info("AttachPermissions called",
		zap.String("role_id", req.RoleId))

	if ch.roleAssignmentService == nil {
		return &pb.AttachPermissionsResponse{
			StatusCode: 503,
			Message:    "Role assignment service not available",
		}, status.Error(codes.Unavailable, "role assignment service not configured")
	}

	// Validate request
	if req.RoleId == "" || len(req.PermissionIds) == 0 {
		return &pb.AttachPermissionsResponse{
			StatusCode: 400,
			Message:    "Role ID and permission IDs are required",
		}, status.Error(codes.InvalidArgument, "role ID and permission IDs are required")
	}

	// Attach permissions to role
	err := ch.roleAssignmentService.AssignPermissionsToRole(ctx, req.RoleId, req.PermissionIds, "system")
	if err != nil {
		ch.logger.Error("Failed to attach permissions", zap.Error(err))
		return &pb.AttachPermissionsResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to attach permissions: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.AttachPermissionsResponse{
		StatusCode: 200,
		Message:    "Permissions attached successfully",
	}, nil
}

// ListPermissions lists all permissions in the catalog
func (ch *CatalogHandler) ListPermissions(
	ctx context.Context,
	req *pb.ListPermissionsRequest,
) (*pb.ListPermissionsResponse, error) {
	ch.logger.Info("ListPermissions called")

	if ch.permissionService == nil {
		return &pb.ListPermissionsResponse{
			StatusCode: 503,
			Message:    "Permission service not available",
		}, status.Error(codes.Unavailable, "permission service not configured")
	}

	// Set default pagination
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// List permissions via service
	permissions, err := ch.permissionService.ListPermissions(ctx, &permissionService.PermissionFilter{
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		ch.logger.Error("Failed to list permissions", zap.Error(err))
		return &pb.ListPermissionsResponse{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to list permissions: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf format
	pbPermissions := make([]*pb.CatalogPermission, len(permissions))
	for i, permission := range permissions {
		pbPermissions[i] = &pb.CatalogPermission{
			Id:          permission.GetID(),
			Name:        permission.Name,
			Description: permission.Description,
			ResourceId:  getStringOrEmpty(permission.ResourceID),
			ActionId:    getStringOrEmpty(permission.ActionID),
		}
	}

	return &pb.ListPermissionsResponse{
		StatusCode:  200,
		Message:     "Permissions retrieved successfully",
		Permissions: pbPermissions,
		TotalCount:  int32(len(permissions)),
		Page:        int32(page),
		PageSize:    int32(pageSize),
	}, nil
}
