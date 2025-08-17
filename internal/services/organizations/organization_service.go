package organizations

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/entities/requests/organizations"
	organizationResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/organizations"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	orgRepo "github.com/Kisanlink/aaa-service/internal/repositories/organizations"
	"github.com/Kisanlink/aaa-service/internal/repositories/users"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
)

// Service handles business logic for organization operations
type Service struct {
	orgRepo   *orgRepo.OrganizationRepository
	userRepo  *users.UserRepository
	validator interfaces.Validator
	logger    *zap.Logger
}

// NewOrganizationService creates a new organization service instance
func NewOrganizationService(
	orgRepo *orgRepo.OrganizationRepository,
	userRepo *users.UserRepository,
	validator interfaces.Validator,
	logger *zap.Logger,
) *Service {
	return &Service{
		orgRepo:   orgRepo,
		userRepo:  userRepo,
		validator: validator,
		logger:    logger,
	}
}

// CreateOrganization creates a new organization with proper validation and business logic
func (s *Service) CreateOrganization(ctx context.Context, req *organizations.CreateOrganizationRequest) (*organizationResponses.OrganizationResponse, error) {
	s.logger.Info("Creating new organization", zap.String("name", req.Name))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Organization creation validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid organization data", err.Error())
	}

	// Check if organization already exists by name
	existingOrg, err := s.orgRepo.GetByName(ctx, req.Name)
	if err == nil && existingOrg != nil {
		s.logger.Warn("Organization already exists with name", zap.String("name", req.Name))
		return nil, errors.NewConflictError("organization with this name already exists")
	}

	// Validate parent organization if specified
	if req.ParentID != nil && *req.ParentID != "" {
		parentOrg, err := s.orgRepo.GetByID(ctx, *req.ParentID)
		if err != nil || parentOrg == nil {
			s.logger.Warn("Parent organization not found", zap.String("parent_id", *req.ParentID))
			return nil, errors.NewNotFoundError("parent organization not found")
		}
		if !parentOrg.IsActive {
			s.logger.Warn("Parent organization is inactive", zap.String("parent_id", *req.ParentID))
			return nil, errors.NewValidationError("parent organization is inactive")
		}
	}

	// Create organization model
	org := models.NewOrganization(req.Name, req.Description)
	if req.ParentID != nil && *req.ParentID != "" {
		org.ParentID = req.ParentID
	}

	// Save organization to repository
	err = s.orgRepo.Create(ctx, org)
	if err != nil {
		s.logger.Error("Failed to create organization in repository", zap.Error(err))
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.NewConflictError("organization with this information already exists")
		}
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Organization created successfully",
		zap.String("org_id", org.ID),
		zap.String("name", org.Name))

	// Convert to response format
	response := &organizationResponses.OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Description: org.Description,
		ParentID:    org.ParentID,
		IsActive:    org.IsActive,
		CreatedAt:   &org.CreatedAt,
		UpdatedAt:   &org.UpdatedAt,
	}

	return response, nil
}

// GetOrganization retrieves an organization by ID
func (s *Service) GetOrganization(ctx context.Context, orgID string) (*organizationResponses.OrganizationResponse, error) {
	s.logger.Info("Retrieving organization", zap.String("org_id", orgID))

	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to retrieve organization", zap.Error(err))
		return nil, errors.NewNotFoundError("organization not found")
	}

	response := &organizationResponses.OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Description: org.Description,
		ParentID:    org.ParentID,
		IsActive:    org.IsActive,
		CreatedAt:   &org.CreatedAt,
		UpdatedAt:   &org.UpdatedAt,
	}

	return response, nil
}

// UpdateOrganization updates an existing organization
func (s *Service) UpdateOrganization(ctx context.Context, orgID string, req *organizations.UpdateOrganizationRequest) (*organizationResponses.OrganizationResponse, error) {
	s.logger.Info("Updating organization", zap.String("org_id", orgID))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Organization update validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid organization data", err.Error())
	}

	// Get existing organization
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found for update", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// Check if name is being changed and if new name already exists
	if req.Name != nil && *req.Name != org.Name {
		existingOrg, err := s.orgRepo.GetByName(ctx, *req.Name)
		if err == nil && existingOrg != nil && existingOrg.ID != orgID {
			s.logger.Warn("Organization name already taken", zap.String("name", *req.Name))
			return nil, errors.NewConflictError("organization name is already taken")
		}
	}

	// Validate parent organization if being changed
	if req.ParentID != nil && (org.ParentID == nil || *req.ParentID != *org.ParentID) {
		if *req.ParentID != "" {
			parentOrg, err := s.orgRepo.GetByID(ctx, *req.ParentID)
			if err != nil || parentOrg == nil {
				s.logger.Warn("Parent organization not found", zap.String("parent_id", *req.ParentID))
				return nil, errors.NewNotFoundError("parent organization not found")
			}
			if !parentOrg.IsActive {
				s.logger.Warn("Parent organization is inactive", zap.String("parent_id", *req.ParentID))
				return nil, errors.NewValidationError("parent organization is inactive")
			}
			// Check for circular references
			if err := s.checkCircularReference(ctx, orgID, *req.ParentID); err != nil {
				s.logger.Warn("Circular reference detected", zap.Error(err))
				return nil, errors.NewValidationError("circular reference detected in organization hierarchy")
			}
		}
	}

	// Update fields
	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.Description != nil {
		org.Description = *req.Description
	}
	if req.ParentID != nil {
		org.ParentID = req.ParentID
	}
	if req.IsActive != nil {
		org.IsActive = *req.IsActive
	}

	// Save changes
	err = s.orgRepo.Update(ctx, org)
	if err != nil {
		s.logger.Error("Failed to update organization", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Organization updated successfully", zap.String("org_id", orgID))

	response := &organizationResponses.OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Description: org.Description,
		ParentID:    org.ParentID,
		IsActive:    org.IsActive,
		CreatedAt:   &org.CreatedAt,
		UpdatedAt:   &org.UpdatedAt,
	}

	return response, nil
}

// DeleteOrganization deletes an organization
func (s *Service) DeleteOrganization(ctx context.Context, orgID string, deletedBy string) error {
	s.logger.Info("Deleting organization", zap.String("org_id", orgID))

	// Get organization to check if it has children
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found for deletion", zap.String("org_id", orgID))
		return errors.NewNotFoundError("organization not found")
	}

	// Check if organization has children
	children, err := s.orgRepo.GetChildren(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to check organization children", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if len(children) > 0 {
		s.logger.Warn("Cannot delete organization with children", zap.String("org_id", orgID))
		return errors.NewValidationError("cannot delete organization with child organizations")
	}

	// Check if organization has active groups
	hasGroups, err := s.orgRepo.HasActiveGroups(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to check organization groups", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if hasGroups {
		s.logger.Warn("Cannot delete organization with active groups", zap.String("org_id", orgID))
		return errors.NewValidationError("cannot delete organization with active groups")
	}

	// Soft delete the organization
	err = s.orgRepo.SoftDelete(ctx, orgID, deletedBy)
	if err != nil {
		s.logger.Error("Failed to delete organization", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Organization deleted successfully", zap.String("org_id", orgID))
	return nil
}

// ListOrganizations retrieves organizations with pagination and filtering
func (s *Service) ListOrganizations(ctx context.Context, limit, offset int, includeInactive bool) ([]*organizationResponses.OrganizationResponse, error) {
	s.logger.Info("Listing organizations", zap.Int("limit", limit), zap.Int("offset", offset))

	var orgs []*models.Organization
	var err error

	if includeInactive {
		orgs, err = s.orgRepo.List(ctx, limit, offset)
	} else {
		orgs, err = s.orgRepo.ListActive(ctx, limit, offset)
	}

	if err != nil {
		s.logger.Error("Failed to list organizations", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format
	responses := make([]*organizationResponses.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		responses[i] = &organizationResponses.OrganizationResponse{
			ID:          org.ID,
			Name:        org.Name,
			Description: org.Description,
			ParentID:    org.ParentID,
			IsActive:    org.IsActive,
			CreatedAt:   &org.CreatedAt,
			UpdatedAt:   &org.UpdatedAt,
		}
	}

	return responses, nil
}

// GetOrganizationHierarchy retrieves the complete hierarchy for an organization
func (s *Service) GetOrganizationHierarchy(ctx context.Context, orgID string) (*organizationResponses.OrganizationHierarchyResponse, error) {
	s.logger.Info("Retrieving organization hierarchy", zap.String("org_id", orgID))

	// Get the organization
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// Get parent hierarchy
	parents, err := s.orgRepo.GetParentHierarchy(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to get parent hierarchy", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Get children
	children, err := s.orgRepo.GetChildren(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to get children", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Build response
	response := &organizationResponses.OrganizationHierarchyResponse{
		Organization: &organizationResponses.OrganizationResponse{
			ID:          org.ID,
			Name:        org.Name,
			Description: org.Description,
			ParentID:    org.ParentID,
			IsActive:    org.IsActive,
			CreatedAt:   &org.CreatedAt,
			UpdatedAt:   &org.UpdatedAt,
		},
		Parents:  make([]*organizationResponses.OrganizationResponse, len(parents)),
		Children: make([]*organizationResponses.OrganizationResponse, len(children)),
	}

	// Convert parents
	for i, parent := range parents {
		response.Parents[i] = &organizationResponses.OrganizationResponse{
			ID:          parent.ID,
			Name:        parent.Name,
			Description: parent.Description,
			ParentID:    parent.ParentID,
			IsActive:    parent.IsActive,
			CreatedAt:   &parent.CreatedAt,
			UpdatedAt:   &parent.UpdatedAt,
		}
	}

	// Convert children
	for i, child := range children {
		response.Children[i] = &organizationResponses.OrganizationResponse{
			ID:          child.ID,
			Name:        child.Name,
			Description: child.Description,
			ParentID:    child.ParentID,
			IsActive:    child.IsActive,
			CreatedAt:   &child.CreatedAt,
			UpdatedAt:   &child.UpdatedAt,
		}
	}

	return response, nil
}

// ActivateOrganization activates an inactive organization
func (s *Service) ActivateOrganization(ctx context.Context, orgID string) error {
	s.logger.Info("Activating organization", zap.String("org_id", orgID))

	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found for activation", zap.String("org_id", orgID))
		return errors.NewNotFoundError("organization not found")
	}

	if org.IsActive {
		s.logger.Warn("Organization is already active", zap.String("org_id", orgID))
		return errors.NewValidationError("organization is already active")
	}

	// Check if parent organization is active
	if org.ParentID != nil {
		parentOrg, err := s.orgRepo.GetByID(ctx, *org.ParentID)
		if err != nil || parentOrg == nil || !parentOrg.IsActive {
			s.logger.Warn("Parent organization is inactive", zap.String("parent_id", *org.ParentID))
			return errors.NewValidationError("cannot activate organization with inactive parent")
		}
	}

	org.IsActive = true
	err = s.orgRepo.Update(ctx, org)
	if err != nil {
		s.logger.Error("Failed to activate organization", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Organization activated successfully", zap.String("org_id", orgID))
	return nil
}

// DeactivateOrganization deactivates an active organization
func (s *Service) DeactivateOrganization(ctx context.Context, orgID string) error {
	s.logger.Info("Deactivating organization", zap.String("org_id", orgID))

	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found for deactivation", zap.String("org_id", orgID))
		return errors.NewNotFoundError("organization not found")
	}

	if !org.IsActive {
		s.logger.Warn("Organization is already inactive", zap.String("org_id", orgID))
		return errors.NewValidationError("organization is already inactive")
	}

	// Check if organization has active children
	children, err := s.orgRepo.GetActiveChildren(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to check active children", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if len(children) > 0 {
		s.logger.Warn("Cannot deactivate organization with active children", zap.String("org_id", orgID))
		return errors.NewValidationError("cannot deactivate organization with active children")
	}

	org.IsActive = false
	err = s.orgRepo.Update(ctx, org)
	if err != nil {
		s.logger.Error("Failed to deactivate organization", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Organization deactivated successfully", zap.String("org_id", orgID))
	return nil
}

// checkCircularReference checks if setting a parent would create a circular reference
func (s *Service) checkCircularReference(ctx context.Context, orgID, newParentID string) error {
	// Start from the new parent and traverse up the hierarchy
	currentID := newParentID
	visited := make(map[string]bool)

	for currentID != "" {
		if currentID == orgID {
			return fmt.Errorf("circular reference detected: %s would be its own ancestor", orgID)
		}

		if visited[currentID] {
			return fmt.Errorf("circular reference detected in organization hierarchy")
		}

		visited[currentID] = true

		// Get the current organization's parent
		org, err := s.orgRepo.GetByID(ctx, currentID)
		if err != nil || org == nil {
			break
		}

		currentID = ""
		if org.ParentID != nil {
			currentID = *org.ParentID
		}
	}

	return nil
}

// GetOrganizationStats retrieves statistics about an organization
func (s *Service) GetOrganizationStats(ctx context.Context, orgID string) (*organizationResponses.OrganizationStatsResponse, error) {
	s.logger.Info("Retrieving organization stats", zap.String("org_id", orgID))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// Get various counts
	childCount, err := s.orgRepo.CountChildren(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to count children", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	groupCount, err := s.orgRepo.CountGroups(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to count groups", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	userCount, err := s.orgRepo.CountUsers(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to count users", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	response := &organizationResponses.OrganizationStatsResponse{
		OrganizationID: orgID,
		ChildCount:     childCount,
		GroupCount:     groupCount,
		UserCount:      userCount,
	}

	return response, nil
}
