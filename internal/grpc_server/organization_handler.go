package grpc_server

import (
	"context"

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
	orgService interfaces.OrganizationService
	logger     *zap.Logger
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

// Stub implementations for other methods
func (h *OrganizationHandler) AddUserToOrganization(ctx context.Context, req *pb.AddUserToOrganizationRequest) (*pb.AddUserToOrganizationResponse, error) {
	return &pb.AddUserToOrganizationResponse{
		StatusCode: 501,
		Message:    "Not implemented yet",
	}, status.Error(codes.Unimplemented, "method not implemented")
}

func (h *OrganizationHandler) RemoveUserFromOrganization(ctx context.Context, req *pb.RemoveUserFromOrganizationRequest) (*pb.RemoveUserFromOrganizationResponse, error) {
	return &pb.RemoveUserFromOrganizationResponse{
		StatusCode: 501,
		Message:    "Not implemented yet",
		Success:    false,
	}, status.Error(codes.Unimplemented, "method not implemented")
}

func (h *OrganizationHandler) ValidateOrganizationAccess(ctx context.Context, req *pb.ValidateOrganizationAccessRequest) (*pb.ValidateOrganizationAccessResponse, error) {
	return &pb.ValidateOrganizationAccessResponse{
		Allowed: false,
	}, status.Error(codes.Unimplemented, "method not implemented")
}

func (h *OrganizationHandler) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	return &pb.CreateRoleResponse{
		StatusCode: 501,
		Message:    "Not implemented yet",
	}, status.Error(codes.Unimplemented, "method not implemented")
}

func (h *OrganizationHandler) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	return &pb.ListRolesResponse{
		StatusCode: 501,
		Message:    "Not implemented yet",
	}, status.Error(codes.Unimplemented, "method not implemented")
}

func (h *OrganizationHandler) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	return &pb.UpdateRoleResponse{
		StatusCode: 501,
		Message:    "Not implemented yet",
	}, status.Error(codes.Unimplemented, "method not implemented")
}

func (h *OrganizationHandler) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	return &pb.DeleteRoleResponse{
		StatusCode: 501,
		Message:    "Not implemented yet",
		Success:    false,
	}, status.Error(codes.Unimplemented, "method not implemented")
}
