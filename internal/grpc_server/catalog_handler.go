package grpc_server

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/services/catalog"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"go.uber.org/zap"
)

// CatalogHandler implements the gRPC CatalogService interface
type CatalogHandler struct {
	pb.UnimplementedCatalogServiceServer
	catalogService *catalog.CatalogService
	logger         *zap.Logger
}

// NewCatalogHandler creates a new catalog handler
func NewCatalogHandler(catalogService *catalog.CatalogService, logger *zap.Logger) *CatalogHandler {
	return &CatalogHandler{
		catalogService: catalogService,
		logger:         logger,
	}
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

	// TODO: Implement RegisterAction
	return &pb.RegisterActionResponse{
		StatusCode: 501,
		Message:    "RegisterAction not yet implemented",
	}, nil
}

// ListActions lists all actions in the catalog
func (ch *CatalogHandler) ListActions(
	ctx context.Context,
	req *pb.ListActionsRequest,
) (*pb.ListActionsResponse, error) {
	ch.logger.Info("ListActions called")

	// TODO: Implement ListActions
	return &pb.ListActionsResponse{
		StatusCode: 501,
		Message:    "ListActions not yet implemented",
	}, nil
}

// RegisterResource registers a new resource in the catalog
func (ch *CatalogHandler) RegisterResource(
	ctx context.Context,
	req *pb.RegisterResourceRequest,
) (*pb.RegisterResourceResponse, error) {
	ch.logger.Info("RegisterResource called",
		zap.String("name", req.Name))

	// TODO: Implement RegisterResource
	return &pb.RegisterResourceResponse{
		StatusCode: 501,
		Message:    "RegisterResource not yet implemented",
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

	// TODO: Implement SetResourceParent
	return &pb.SetResourceParentResponse{
		StatusCode: 501,
		Message:    "SetResourceParent not yet implemented",
	}, nil
}

// ListResources lists all resources in the catalog
func (ch *CatalogHandler) ListResources(
	ctx context.Context,
	req *pb.ListResourcesRequest,
) (*pb.ListResourcesResponse, error) {
	ch.logger.Info("ListResources called")

	// TODO: Implement ListResources
	return &pb.ListResourcesResponse{
		StatusCode: 501,
		Message:    "ListResources not yet implemented",
	}, nil
}

// CreateRole creates a new role in the catalog
func (ch *CatalogHandler) CreateRole(
	ctx context.Context,
	req *pb.CreateRoleRequest,
) (*pb.CreateRoleResponse, error) {
	ch.logger.Info("CreateRole called",
		zap.String("name", req.Name))

	// TODO: Implement CreateRole
	return &pb.CreateRoleResponse{
		StatusCode: 501,
		Message:    "CreateRole not yet implemented",
	}, nil
}

// ListRoles lists all roles in the catalog
func (ch *CatalogHandler) ListRoles(
	ctx context.Context,
	req *pb.ListRolesRequest,
) (*pb.ListRolesResponse, error) {
	ch.logger.Info("ListRoles called")

	// TODO: Implement ListRoles
	return &pb.ListRolesResponse{
		StatusCode: 501,
		Message:    "ListRoles not yet implemented",
	}, nil
}

// CreatePermission creates a new permission in the catalog
func (ch *CatalogHandler) CreatePermission(
	ctx context.Context,
	req *pb.CreatePermissionRequest,
) (*pb.CreatePermissionResponse, error) {
	ch.logger.Info("CreatePermission called",
		zap.String("name", req.Name))

	// TODO: Implement CreatePermission
	return &pb.CreatePermissionResponse{
		StatusCode: 501,
		Message:    "CreatePermission not yet implemented",
	}, nil
}

// AttachPermissions attaches permissions to a role
func (ch *CatalogHandler) AttachPermissions(
	ctx context.Context,
	req *pb.AttachPermissionsRequest,
) (*pb.AttachPermissionsResponse, error) {
	ch.logger.Info("AttachPermissions called",
		zap.String("role_id", req.RoleId))

	// TODO: Implement AttachPermissions
	return &pb.AttachPermissionsResponse{
		StatusCode: 501,
		Message:    "AttachPermissions not yet implemented",
	}, nil
}

// ListPermissions lists all permissions in the catalog
func (ch *CatalogHandler) ListPermissions(
	ctx context.Context,
	req *pb.ListPermissionsRequest,
) (*pb.ListPermissionsResponse, error) {
	ch.logger.Info("ListPermissions called")

	// TODO: Implement ListPermissions
	return &pb.ListPermissionsResponse{
		StatusCode: 501,
		Message:    "ListPermissions not yet implemented",
	}, nil
}
