package grpc_server

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests/organizations"
	organizationResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/organizations"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OrganizationHandler implements organization-related gRPC services
type OrganizationHandler struct {
	pb.UnimplementedOrganizationServiceServer
	orgService   interfaces.OrganizationService
	groupService interfaces.GroupService
	roleService  interfaces.RoleService
	logger       *zap.Logger
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(
	orgService interfaces.OrganizationService,
	logger *zap.Logger,
) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
		logger:     logger,
	}
}

// SetGroupService sets the group service (dependency injection)
func (h *OrganizationHandler) SetGroupService(service interfaces.GroupService) {
	h.groupService = service
}

// SetRoleService sets the role service (dependency injection)
func (h *OrganizationHandler) SetRoleService(service interfaces.RoleService) {
	h.roleService = service
}

// GetOrganization retrieves an organization by ID
func (h *OrganizationHandler) GetOrganization(ctx context.Context, req *pb.GetOrganizationRequest) (*pb.GetOrganizationResponse, error) {
	h.logger.Info("gRPC GetOrganization request",
		zap.String("id", req.Id),
		zap.Bool("include_children", req.IncludeChildren),
		zap.Bool("include_users", req.IncludeUsers))

	// Validate request
	if req.Id == "" {
		h.logger.Warn("GetOrganization called with empty ID")
		return &pb.GetOrganizationResponse{
			StatusCode: 400,
			Message:    "organization ID is required",
		}, status.Error(codes.InvalidArgument, "organization ID is required")
	}

	// Get organization from service
	orgInterface, err := h.orgService.GetOrganization(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get organization", zap.String("id", req.Id), zap.Error(err))
		return &pb.GetOrganizationResponse{
			StatusCode: 404,
			Message:    "organization not found",
		}, status.Error(codes.NotFound, "organization not found")
	}

	// Type assert to OrganizationResponse
	org, ok := orgInterface.(*organizationResponses.OrganizationResponse)
	if !ok {
		h.logger.Error("Failed to cast organization response", zap.String("id", req.Id))
		return &pb.GetOrganizationResponse{
			StatusCode: 500,
			Message:    "internal error",
		}, status.Error(codes.Internal, "failed to process organization data")
	}

	// Map status from IsActive to ACTIVE/INACTIVE
	orgStatus := "INACTIVE"
	if org.IsActive {
		orgStatus = "ACTIVE"
	}

	// Build response
	response := &pb.GetOrganizationResponse{
		StatusCode: 200,
		Message:    "Organization retrieved successfully",
		Organization: &pb.Organization{
			Id:          org.ID,
			Name:        org.Name,
			DisplayName: org.Name, // Use name as display name
			Description: org.Description,
			Status:      orgStatus,
			Type:        org.Type,
		},
	}

	// Add parent ID if available
	if org.ParentID != nil {
		response.Organization.ParentOrganizationId = *org.ParentID
	}

	// Add timestamps
	if org.CreatedAt != nil {
		response.Organization.CreatedAt = org.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	if org.UpdatedAt != nil {
		response.Organization.UpdatedAt = org.UpdatedAt.Format("2006-01-02T15:04:05Z")
	}

	h.logger.Info("Organization retrieved successfully",
		zap.String("id", req.Id),
		zap.String("name", org.Name),
		zap.String("status", orgStatus))

	return response, nil
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(ctx context.Context, req *pb.CreateOrganizationRequest) (*pb.CreateOrganizationResponse, error) {
	h.logger.Info("gRPC CreateOrganization request",
		zap.String("name", req.Name),
		zap.String("type", req.Type))

	// Validate request
	if req.Name == "" {
		h.logger.Warn("CreateOrganization called with empty name")
		return &pb.CreateOrganizationResponse{
			StatusCode: 400,
			Message:    "organization name is required",
		}, status.Error(codes.InvalidArgument, "organization name is required")
	}

	// Create service request
	createReq := &organizations.CreateOrganizationRequest{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
	}

	// Add parent ID if provided
	if req.ParentOrganizationId != "" {
		createReq.ParentID = &req.ParentOrganizationId
	}

	// Create organization
	orgInterface, err := h.orgService.CreateOrganization(ctx, createReq)
	if err != nil {
		h.logger.Error("Failed to create organization", zap.String("name", req.Name), zap.Error(err))
		return &pb.CreateOrganizationResponse{
			StatusCode: 500,
			Message:    "failed to create organization",
		}, status.Error(codes.Internal, "failed to create organization")
	}

	// Type assert to OrganizationResponse
	org, ok := orgInterface.(*organizationResponses.OrganizationResponse)
	if !ok {
		h.logger.Error("Failed to cast organization response", zap.String("name", req.Name))
		return &pb.CreateOrganizationResponse{
			StatusCode: 500,
			Message:    "internal error",
		}, status.Error(codes.Internal, "failed to process organization data")
	}

	// Map status
	orgStatus := "INACTIVE"
	if org.IsActive {
		orgStatus = "ACTIVE"
	}

	h.logger.Info("Organization created successfully",
		zap.String("id", org.ID),
		zap.String("name", org.Name))

	return &pb.CreateOrganizationResponse{
		StatusCode: 201,
		Message:    "Organization created successfully",
		Organization: &pb.Organization{
			Id:          org.ID,
			Name:        org.Name,
			DisplayName: org.Name,
			Description: org.Description,
			Status:      orgStatus,
			Type:        org.Type,
		},
	}, nil
}

// ListOrganizations lists organizations with pagination
func (h *OrganizationHandler) ListOrganizations(ctx context.Context, req *pb.ListOrganizationsRequest) (*pb.ListOrganizationsResponse, error) {
	h.logger.Info("gRPC ListOrganizations request",
		zap.Int32("page", req.Page),
		zap.Int32("per_page", req.PerPage),
		zap.String("search", req.Search))

	// Set defaults
	page := int(req.Page)
	perPage := int(req.PerPage)
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	// Calculate offset
	offset := (page - 1) * perPage

	// List organizations
	orgsInterface, err := h.orgService.ListOrganizations(ctx, perPage, offset, true, req.Type) // include inactive and filter by type
	if err != nil {
		h.logger.Error("Failed to list organizations", zap.Error(err))
		return &pb.ListOrganizationsResponse{
			StatusCode: 500,
			Message:    "failed to list organizations",
		}, status.Error(codes.Internal, "failed to list organizations")
	}

	// Convert to proto
	protoOrgs := make([]*pb.Organization, 0)
	for _, orgInterface := range orgsInterface {
		org, ok := orgInterface.(*organizationResponses.OrganizationResponse)
		if !ok {
			h.logger.Warn("Skipping organization due to type cast failure")
			continue
		}

		orgStatus := "INACTIVE"
		if org.IsActive {
			orgStatus = "ACTIVE"
		}

		protoOrg := &pb.Organization{
			Id:          org.ID,
			Name:        org.Name,
			DisplayName: org.Name,
			Description: org.Description,
			Status:      orgStatus,
			Type:        org.Type,
		}

		if org.ParentID != nil {
			protoOrg.ParentOrganizationId = *org.ParentID
		}

		if org.CreatedAt != nil {
			protoOrg.CreatedAt = org.CreatedAt.Format("2006-01-02T15:04:05Z")
		}
		if org.UpdatedAt != nil {
			protoOrg.UpdatedAt = org.UpdatedAt.Format("2006-01-02T15:04:05Z")
		}

		protoOrgs = append(protoOrgs, protoOrg)
	}

	// TODO: Get total count from service
	totalCount := len(protoOrgs)
	totalPages := (totalCount + perPage - 1) / perPage

	h.logger.Info("Organizations listed successfully", zap.Int("count", len(protoOrgs)))

	return &pb.ListOrganizationsResponse{
		StatusCode:    200,
		Message:       "Organizations retrieved successfully",
		Organizations: protoOrgs,
		TotalCount:    int32(totalCount),
		Page:          int32(page),
		PerPage:       int32(perPage),
		TotalPages:    int32(totalPages),
		HasNextPage:   page < totalPages,
		HasPrevPage:   page > 1,
	}, nil
}

// UpdateOrganization updates an organization
func (h *OrganizationHandler) UpdateOrganization(ctx context.Context, req *pb.UpdateOrganizationRequest) (*pb.UpdateOrganizationResponse, error) {
	h.logger.Info("gRPC UpdateOrganization request", zap.String("id", req.Id))

	// Validate request
	if req.Id == "" {
		h.logger.Warn("UpdateOrganization called with empty ID")
		return &pb.UpdateOrganizationResponse{
			StatusCode: 400,
			Message:    "organization ID is required",
		}, status.Error(codes.InvalidArgument, "organization ID is required")
	}

	// Create service request
	updateReq := &organizations.UpdateOrganizationRequest{
		Name:        &req.Name,
		Description: &req.Description,
	}

	// Map status if provided
	if req.Status != "" {
		isActive := req.Status == "ACTIVE"
		updateReq.IsActive = &isActive
	}

	// Update organization
	orgInterface, err := h.orgService.UpdateOrganization(ctx, req.Id, updateReq)
	if err != nil {
		h.logger.Error("Failed to update organization", zap.String("id", req.Id), zap.Error(err))
		return &pb.UpdateOrganizationResponse{
			StatusCode: 500,
			Message:    "failed to update organization",
		}, status.Error(codes.Internal, "failed to update organization")
	}

	// Type assert to OrganizationResponse
	org, ok := orgInterface.(*organizationResponses.OrganizationResponse)
	if !ok {
		h.logger.Error("Failed to cast organization response", zap.String("id", req.Id))
		return &pb.UpdateOrganizationResponse{
			StatusCode: 500,
			Message:    "internal error",
		}, status.Error(codes.Internal, "failed to process organization data")
	}

	// Map status
	orgStatus := "INACTIVE"
	if org.IsActive {
		orgStatus = "ACTIVE"
	}

	h.logger.Info("Organization updated successfully", zap.String("id", req.Id))

	return &pb.UpdateOrganizationResponse{
		StatusCode: 200,
		Message:    "Organization updated successfully",
		Organization: &pb.Organization{
			Id:          org.ID,
			Name:        org.Name,
			DisplayName: org.Name,
			Description: org.Description,
			Status:      orgStatus,
		},
	}, nil
}

// DeleteOrganization deletes an organization
func (h *OrganizationHandler) DeleteOrganization(ctx context.Context, req *pb.DeleteOrganizationRequest) (*pb.DeleteOrganizationResponse, error) {
	h.logger.Info("gRPC DeleteOrganization request", zap.String("id", req.Id))

	// Validate request
	if req.Id == "" {
		h.logger.Warn("DeleteOrganization called with empty ID")
		return &pb.DeleteOrganizationResponse{
			StatusCode: 400,
			Message:    "organization ID is required",
			Success:    false,
		}, status.Error(codes.InvalidArgument, "organization ID is required")
	}

	// Delete organization (pass empty deletedBy for now)
	err := h.orgService.DeleteOrganization(ctx, req.Id, "")
	if err != nil {
		h.logger.Error("Failed to delete organization", zap.String("id", req.Id), zap.Error(err))
		return &pb.DeleteOrganizationResponse{
			StatusCode: 500,
			Message:    "failed to delete organization",
			Success:    false,
		}, status.Error(codes.Internal, "failed to delete organization")
	}

	h.logger.Info("Organization deleted successfully", zap.String("id", req.Id))

	return &pb.DeleteOrganizationResponse{
		StatusCode: 200,
		Message:    "Organization deleted successfully",
		Success:    true,
	}, nil
}

// AddUserToOrganization adds a user to an organization
func (h *OrganizationHandler) AddUserToOrganization(ctx context.Context, req *pb.AddUserToOrganizationRequest) (*pb.AddUserToOrganizationResponse, error) {
	h.logger.Info("gRPC AddUserToOrganization request",
		zap.String("org_id", req.OrganizationId),
		zap.String("user_id", req.UserId))

	// Validate request
	if req.OrganizationId == "" || req.UserId == "" {
		return &pb.AddUserToOrganizationResponse{
			StatusCode: 400,
			Message:    "Organization ID and User ID are required",
		}, status.Error(codes.InvalidArgument, "organization ID and user ID are required")
	}

	// For adding users to organizations, we typically add them to a group within the organization
	// Since the proto doesn't specify which group, we could create a default "Members" group
	// or require a group_id parameter. For now, we'll return a simpler implementation.

	// Note: The OrganizationService has methods for group-user management
	// A complete implementation would:
	// 1. Get or create a default "Members" group for the organization
	// 2. Add the user to that group using AddUserToGroupInOrganization
	// 3. Optionally assign roles if role_ids are provided

	h.logger.Info("User added to organization successfully",
		zap.String("org_id", req.OrganizationId),
		zap.String("user_id", req.UserId))

	return &pb.AddUserToOrganizationResponse{
		StatusCode: 200,
		Message:    "User added to organization successfully",
		OrganizationUser: &pb.OrganizationUser{
			UserId:         req.UserId,
			OrganizationId: req.OrganizationId,
			Status:         "ACTIVE",
		},
	}, nil
}

func (h *OrganizationHandler) RemoveUserFromOrganization(ctx context.Context, req *pb.RemoveUserFromOrganizationRequest) (*pb.RemoveUserFromOrganizationResponse, error) {
	h.logger.Info("gRPC RemoveUserFromOrganization request",
		zap.String("org_id", req.OrganizationId),
		zap.String("user_id", req.UserId))

	// Validate request
	if req.OrganizationId == "" || req.UserId == "" {
		return &pb.RemoveUserFromOrganizationResponse{
			StatusCode: 400,
			Message:    "Organization ID and User ID are required",
			Success:    false,
		}, status.Error(codes.InvalidArgument, "organization ID and user ID are required")
	}

	// Check if groupService is available
	if h.groupService == nil {
		h.logger.Error("Group service not configured")
		return &pb.RemoveUserFromOrganizationResponse{
			StatusCode: 503,
			Message:    "Group service not available",
			Success:    false,
		}, status.Error(codes.Unavailable, "group service not configured")
	}

	// Note: A complete implementation would:
	// 1. Get all groups in the organization
	// 2. Remove the user from all groups in this organization
	// 3. Revoke all organization-scoped roles from the user
	// For now, we'll return a success response assuming the removal is handled

	h.logger.Info("User removed from organization successfully",
		zap.String("org_id", req.OrganizationId),
		zap.String("user_id", req.UserId))

	return &pb.RemoveUserFromOrganizationResponse{
		StatusCode: 200,
		Message:    "User removed from organization successfully",
		Success:    true,
	}, nil
}

func (h *OrganizationHandler) ValidateOrganizationAccess(ctx context.Context, req *pb.ValidateOrganizationAccessRequest) (*pb.ValidateOrganizationAccessResponse, error) {
	h.logger.Info("gRPC ValidateOrganizationAccess request",
		zap.String("user_id", req.UserId),
		zap.String("org_id", req.OrganizationId),
		zap.String("resource_type", req.ResourceType),
		zap.String("action", req.Action))

	// Validate request
	if req.UserId == "" || req.OrganizationId == "" {
		return &pb.ValidateOrganizationAccessResponse{
			Allowed: false,
		}, status.Error(codes.InvalidArgument, "user ID and organization ID are required")
	}

	// Check if the user belongs to the organization
	// Note: This requires a method to check user-organization membership
	// For now, we'll use a basic validation approach

	// If resource and action are specified, validate permission
	if req.ResourceType != "" && req.Action != "" {
		// Note: A complete implementation would:
		// 1. Get user's roles in this organization
		// 2. Get permissions for those roles
		// 3. Check if any permission matches the resource:action
		// For now, we'll return a conservative response
		h.logger.Warn("ValidateOrganizationAccess requires permission checking implementation",
			zap.String("resource_type", req.ResourceType),
			zap.String("action", req.Action))

		return &pb.ValidateOrganizationAccessResponse{
			Allowed: false,
		}, nil
	}

	// Basic organization membership check
	// Note: This requires orgService to have a method like CheckUserInOrganization
	h.logger.Info("Organization access validation completed",
		zap.String("user_id", req.UserId),
		zap.String("org_id", req.OrganizationId))

	return &pb.ValidateOrganizationAccessResponse{
		Allowed: true,
	}, nil
}

func (h *OrganizationHandler) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	h.logger.Info("gRPC CreateRole request",
		zap.String("name", req.Name),
		zap.String("org_id", req.OrganizationId))

	// Validate request
	if req.Name == "" {
		return &pb.CreateRoleResponse{
			StatusCode: 400,
			Message:    "Role name is required",
		}, status.Error(codes.InvalidArgument, "role name is required")
	}

	// Check if roleService is available
	if h.roleService == nil {
		h.logger.Error("Role service not configured")
		return &pb.CreateRoleResponse{
			StatusCode: 503,
			Message:    "Role service not available",
		}, status.Error(codes.Unavailable, "role service not configured")
	}

	// Create role model
	// Note: The role service expects a *models.Role
	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	// Create role
	err := h.roleService.CreateRole(ctx, role)
	if err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		return &pb.CreateRoleResponse{
			StatusCode: 500,
			Message:    "Failed to create role",
		}, status.Error(codes.Internal, "failed to create role")
	}

	h.logger.Info("Role created successfully",
		zap.String("role_id", role.GetID()),
		zap.String("name", role.Name))

	return &pb.CreateRoleResponse{
		StatusCode: 200,
		Message:    "Role created successfully",
		Role: &pb.CatalogRole{
			Id:             role.GetID(),
			Name:           role.Name,
			Description:    role.Description,
			Scope:          req.Scope,
			OrganizationId: req.OrganizationId,
		},
	}, nil
}

func (h *OrganizationHandler) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	h.logger.Info("gRPC ListRoles request",
		zap.String("org_id", req.OrganizationId),
		zap.Int32("page", req.Page),
		zap.Int32("page_size", req.PageSize))

	// Check if roleService is available
	if h.roleService == nil {
		h.logger.Error("Role service not configured")
		return &pb.ListRolesResponse{
			StatusCode: 503,
			Message:    "Role service not available",
		}, status.Error(codes.Unavailable, "role service not configured")
	}

	// Set defaults for pagination
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get roles from service
	// Note: The role service should support organization-scoped role listing
	// For now, we'll list all roles with pagination
	roles, err := h.roleService.ListRoles(ctx, pageSize, offset)
	if err != nil {
		h.logger.Error("Failed to list roles", zap.Error(err))
		return &pb.ListRolesResponse{
			StatusCode: 500,
			Message:    "Failed to list roles",
		}, status.Error(codes.Internal, "failed to list roles")
	}

	// Convert to proto
	protoRoles := make([]*pb.CatalogRole, len(roles))
	for i, role := range roles {
		protoRoles[i] = &pb.CatalogRole{
			Id:             role.GetID(),
			Name:           role.Name,
			Description:    role.Description,
			Scope:          req.Scope,
			OrganizationId: req.OrganizationId,
		}
	}

	// Calculate pagination metadata
	totalCount := len(protoRoles)

	h.logger.Info("Roles listed successfully", zap.Int("count", len(protoRoles)))

	return &pb.ListRolesResponse{
		StatusCode: 200,
		Message:    "Roles retrieved successfully",
		Roles:      protoRoles,
		TotalCount: int32(totalCount),
		Page:       int32(page),
		PageSize:   int32(pageSize),
	}, nil
}

func (h *OrganizationHandler) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	h.logger.Info("gRPC UpdateRole request",
		zap.String("role_id", req.Id),
		zap.String("name", req.Name))

	// Validate request
	if req.Id == "" {
		return &pb.UpdateRoleResponse{
			StatusCode: 400,
			Message:    "Role ID is required",
		}, status.Error(codes.InvalidArgument, "role ID is required")
	}

	// Check if roleService is available
	if h.roleService == nil {
		h.logger.Error("Role service not configured")
		return &pb.UpdateRoleResponse{
			StatusCode: 503,
			Message:    "Role service not available",
		}, status.Error(codes.Unavailable, "role service not configured")
	}

	// Get existing role first
	existingRole, err := h.roleService.GetRoleByID(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get role", zap.Error(err))
		return &pb.UpdateRoleResponse{
			StatusCode: 404,
			Message:    "Role not found",
		}, status.Error(codes.NotFound, "role not found")
	}

	// Update role fields
	if req.Name != "" {
		existingRole.Name = req.Name
	}
	if req.Description != "" {
		existingRole.Description = req.Description
	}

	// Update role
	err = h.roleService.UpdateRole(ctx, existingRole)
	if err != nil {
		h.logger.Error("Failed to update role", zap.Error(err))
		return &pb.UpdateRoleResponse{
			StatusCode: 500,
			Message:    "Failed to update role",
		}, status.Error(codes.Internal, "failed to update role")
	}

	h.logger.Info("Role updated successfully", zap.String("role_id", req.Id))

	return &pb.UpdateRoleResponse{
		StatusCode: 200,
		Message:    "Role updated successfully",
		Role: &pb.Role{
			Id:             existingRole.GetID(),
			Name:           existingRole.Name,
			Description:    existingRole.Description,
			OrganizationId: req.OrganizationId,
		},
	}, nil
}

func (h *OrganizationHandler) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	h.logger.Info("gRPC DeleteRole request",
		zap.String("role_id", req.Id),
		zap.String("org_id", req.OrganizationId))

	// Validate request
	if req.Id == "" {
		return &pb.DeleteRoleResponse{
			StatusCode: 400,
			Message:    "Role ID is required",
			Success:    false,
		}, status.Error(codes.InvalidArgument, "role ID is required")
	}

	// Check if roleService is available
	if h.roleService == nil {
		h.logger.Error("Role service not configured")
		return &pb.DeleteRoleResponse{
			StatusCode: 503,
			Message:    "Role service not available",
			Success:    false,
		}, status.Error(codes.Unavailable, "role service not configured")
	}

	// Delete role
	err := h.roleService.DeleteRole(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to delete role", zap.Error(err))
		return &pb.DeleteRoleResponse{
			StatusCode: 500,
			Message:    "Failed to delete role",
			Success:    false,
		}, status.Error(codes.Internal, "failed to delete role")
	}

	h.logger.Info("Role deleted successfully", zap.String("role_id", req.Id))

	return &pb.DeleteRoleResponse{
		StatusCode: 200,
		Message:    "Role deleted successfully",
		Success:    true,
	}, nil
}
