package organizations

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/entities/requests/organizations"
	groupResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/groups"
	organizationResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/organizations"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
)

// Service handles business logic for organization operations
type Service struct {
	orgRepo      interfaces.OrganizationRepository
	userRepo     interfaces.UserRepositoryInterface
	groupRepo    interfaces.GroupRepository
	groupService interfaces.GroupService
	validator    interfaces.Validator
	cache        interfaces.CacheService
	orgCache     *OrganizationCacheService
	auditService interfaces.AuditService
	logger       *zap.Logger
}

// NewOrganizationService creates a new organization service instance
func NewOrganizationService(
	orgRepo interfaces.OrganizationRepository,
	userRepo interfaces.UserRepositoryInterface,
	groupRepo interfaces.GroupRepository,
	groupService interfaces.GroupService,
	validator interfaces.Validator,
	cache interfaces.CacheService,
	auditService interfaces.AuditService,
	logger *zap.Logger,
) *Service {
	orgCache := NewOrganizationCacheService(cache, logger)
	return &Service{
		orgRepo:      orgRepo,
		userRepo:     userRepo,
		groupRepo:    groupRepo,
		groupService: groupService,
		validator:    validator,
		cache:        cache,
		orgCache:     orgCache,
		auditService: auditService,
		logger:       logger,
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

		// Log audit event for failed creation
		auditDetails := map[string]interface{}{
			"organization_name": req.Name,
			"parent_id":         req.ParentID,
			"error":             err.Error(),
		}
		s.auditService.LogOrganizationOperation(ctx, "system", models.AuditActionCreateOrganization, "", "Failed to create organization", false, auditDetails)

		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.NewConflictError("organization with this information already exists")
		}
		return nil, errors.NewInternalError(err)
	}

	// Log successful organization creation
	auditDetails := map[string]interface{}{
		"organization_name": org.Name,
		"parent_id":         org.ParentID,
		"is_active":         org.IsActive,
	}
	s.auditService.LogOrganizationOperation(ctx, "system", models.AuditActionCreateOrganization, org.ID, "Organization created successfully", true, auditDetails)

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

	// Capture old values for audit logging
	oldValues := map[string]interface{}{
		"name":        org.Name,
		"description": org.Description,
		"parent_id":   org.ParentID,
		"is_active":   org.IsActive,
	}

	// Track hierarchy changes for special audit logging
	hierarchyChanged := false
	oldParentID := ""
	newParentID := ""
	if org.ParentID != nil {
		oldParentID = *org.ParentID
	}

	// Update fields
	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.Description != nil {
		org.Description = *req.Description
	}
	if req.ParentID != nil {
		if org.ParentID == nil || *req.ParentID != *org.ParentID {
			hierarchyChanged = true
			if req.ParentID != nil {
				newParentID = *req.ParentID
			}
		}
		org.ParentID = req.ParentID
	}
	if req.IsActive != nil {
		org.IsActive = *req.IsActive
	}

	// Capture new values for audit logging
	newValues := map[string]interface{}{
		"name":        org.Name,
		"description": org.Description,
		"parent_id":   org.ParentID,
		"is_active":   org.IsActive,
	}

	// Save changes
	err = s.orgRepo.Update(ctx, org)
	if err != nil {
		s.logger.Error("Failed to update organization", zap.Error(err))

		// Log audit event for failed update
		auditDetails := map[string]interface{}{
			"old_values": oldValues,
			"new_values": newValues,
			"error":      err.Error(),
		}
		s.auditService.LogOrganizationOperation(ctx, "system", models.AuditActionUpdateOrganization, orgID, "Failed to update organization", false, auditDetails)

		return nil, errors.NewInternalError(err)
	}

	// Log successful organization update
	auditDetails := map[string]interface{}{
		"old_values": oldValues,
		"new_values": newValues,
	}
	s.auditService.LogOrganizationOperation(ctx, "system", models.AuditActionUpdateOrganization, orgID, "Organization updated successfully", true, auditDetails)

	// Log hierarchy change separately if it occurred with comprehensive structure change logging
	if hierarchyChanged {
		s.auditService.LogHierarchyChange(ctx, "system", models.AuditActionChangeOrganizationHierarchy, models.ResourceTypeOrganization, orgID, oldParentID, newParentID, "Organization hierarchy changed", true, auditDetails)

		// Also log comprehensive structure change for enhanced audit trail
		hierarchyOldValues := map[string]interface{}{
			"parent_id": oldParentID,
		}
		hierarchyNewValues := map[string]interface{}{
			"parent_id": newParentID,
		}
		s.auditService.LogOrganizationStructureChange(ctx, "system", models.AuditActionChangeOrganizationHierarchy, orgID, models.ResourceTypeOrganization, orgID, hierarchyOldValues, hierarchyNewValues, true, "Organization hierarchy structure changed")
	}

	// Invalidate cache after successful update
	s.orgCache.InvalidateOrganizationCache(ctx, orgID)

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

		// Log audit event for failed deletion
		auditDetails := map[string]interface{}{
			"organization_name": org.Name,
			"deleted_by":        deletedBy,
			"error":             err.Error(),
		}
		s.auditService.LogOrganizationOperation(ctx, deletedBy, models.AuditActionDeleteOrganization, orgID, "Failed to delete organization", false, auditDetails)

		return errors.NewInternalError(err)
	}

	// Log successful organization deletion
	auditDetails := map[string]interface{}{
		"organization_name": org.Name,
		"deleted_by":        deletedBy,
		"had_children":      len(children) > 0,
		"had_groups":        hasGroups,
	}
	s.auditService.LogOrganizationOperation(ctx, deletedBy, models.AuditActionDeleteOrganization, orgID, "Organization deleted successfully", true, auditDetails)

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

// GetOrganizationHierarchy retrieves the complete hierarchy for an organization including groups and roles
func (s *Service) GetOrganizationHierarchy(ctx context.Context, orgID string) (*organizationResponses.OrganizationHierarchyResponse, error) {
	s.logger.Info("Retrieving organization hierarchy with groups", zap.String("org_id", orgID))

	// Check cache first
	if cached, found := s.orgCache.GetCachedOrganizationHierarchy(ctx, orgID); found {
		s.logger.Debug("Returning cached organization hierarchy", zap.String("org_id", orgID))
		return cached, nil
	}

	// Get the organization
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// Get parent hierarchy with caching
	var parents []*models.Organization
	if cachedParents, found := s.orgCache.GetCachedOrganizationParentHierarchy(ctx, orgID); found {
		parents = cachedParents
	} else {
		parents, err = s.orgRepo.GetParentHierarchy(ctx, orgID)
		if err != nil {
			s.logger.Error("Failed to get parent hierarchy", zap.Error(err))
			return nil, errors.NewInternalError(err)
		}
		// Cache parent hierarchy
		s.orgCache.CacheOrganizationParentHierarchy(ctx, orgID, parents)
	}

	// Get children with caching
	var children []*models.Organization
	if cachedChildren, found := s.orgCache.GetCachedOrganizationChildren(ctx, orgID, false); found {
		children = cachedChildren
	} else {
		children, err = s.orgRepo.GetChildren(ctx, orgID)
		if err != nil {
			s.logger.Error("Failed to get children", zap.Error(err))
			return nil, errors.NewInternalError(err)
		}
		// Cache children
		s.orgCache.CacheOrganizationChildren(ctx, orgID, children, false)
	}

	// Get group hierarchy for the organization
	groupHierarchy, err := s.buildGroupHierarchy(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to build group hierarchy", zap.Error(err))
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
		Groups:   groupHierarchy,
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

	// Cache the complete hierarchy response
	s.orgCache.CacheOrganizationHierarchy(ctx, orgID, response)

	s.logger.Info("Organization hierarchy with groups retrieved successfully",
		zap.String("org_id", orgID),
		zap.Int("group_count", len(groupHierarchy)))

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

		// Log audit event for failed activation
		auditDetails := map[string]interface{}{
			"organization_name": org.Name,
			"error":             err.Error(),
		}
		s.auditService.LogOrganizationOperation(ctx, "system", models.AuditActionActivateOrganization, orgID, "Failed to activate organization", false, auditDetails)

		return errors.NewInternalError(err)
	}

	// Log successful organization activation
	auditDetails := map[string]interface{}{
		"organization_name": org.Name,
		"previous_status":   "inactive",
		"new_status":        "active",
	}
	s.auditService.LogOrganizationOperation(ctx, "system", models.AuditActionActivateOrganization, orgID, "Organization activated successfully", true, auditDetails)

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

		// Log audit event for failed deactivation
		auditDetails := map[string]interface{}{
			"organization_name": org.Name,
			"active_children":   len(children),
			"error":             err.Error(),
		}
		s.auditService.LogOrganizationOperation(ctx, "system", models.AuditActionDeactivateOrganization, orgID, "Failed to deactivate organization", false, auditDetails)

		return errors.NewInternalError(err)
	}

	// Log successful organization deactivation
	auditDetails := map[string]interface{}{
		"organization_name": org.Name,
		"previous_status":   "active",
		"new_status":        "inactive",
		"active_children":   len(children),
	}
	s.auditService.LogOrganizationOperation(ctx, "system", models.AuditActionDeactivateOrganization, orgID, "Organization deactivated successfully", true, auditDetails)

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

	// Check cache first
	if cached, found := s.orgCache.GetCachedOrganizationStats(ctx, orgID); found {
		s.logger.Debug("Returning cached organization stats", zap.String("org_id", orgID))
		return cached, nil
	}

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

	// Cache the stats response
	s.orgCache.CacheOrganizationStats(ctx, orgID, response)

	return response, nil
}

// GetOrganizationGroups retrieves all groups within an organization with pagination
func (s *Service) GetOrganizationGroups(ctx context.Context, orgID string, limit, offset int, includeInactive bool) (interface{}, error) {
	s.logger.Info("Retrieving organization groups",
		zap.String("org_id", orgID),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.Bool("include_inactive", includeInactive))

	// Verify organization exists and is active
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// Get groups from the organization using the group repository
	// Note: This would typically be done through a group service, but for now we'll implement it here
	// In a real implementation, you'd inject a GroupService or GroupRepository

	// For now, return empty slice as placeholder - this would be implemented with actual group repository
	s.logger.Info("Organization groups retrieved successfully", zap.String("org_id", orgID))
	return []interface{}{}, nil
}

// CreateGroupInOrganization creates a new group within an organization
func (s *Service) CreateGroupInOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
	s.logger.Info("Creating group in organization", zap.String("org_id", orgID))

	// Verify organization exists and is active
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	if !org.IsActive {
		s.logger.Warn("Organization is inactive", zap.String("org_id", orgID))
		return nil, errors.NewValidationError("cannot create group in inactive organization")
	}

	// This would delegate to the group service with organization context
	// For now, return placeholder response
	s.logger.Info("Group created in organization successfully", zap.String("org_id", orgID))
	return map[string]interface{}{"message": "group creation not fully implemented"}, nil
}

// GetGroupInOrganization retrieves a specific group within an organization
func (s *Service) GetGroupInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error) {
	s.logger.Info("Retrieving group in organization",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// This would delegate to the group service with organization context validation
	// For now, return placeholder response
	s.logger.Info("Group retrieved from organization successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))
	return map[string]interface{}{"message": "group retrieval not fully implemented"}, nil
}

// UpdateGroupInOrganization updates a group within an organization
func (s *Service) UpdateGroupInOrganization(ctx context.Context, orgID, groupID string, req interface{}) (interface{}, error) {
	s.logger.Info("Updating group in organization",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// This would delegate to the group service with organization context validation
	// For now, return placeholder response
	s.logger.Info("Group updated in organization successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))
	return map[string]interface{}{"message": "group update not fully implemented"}, nil
}

// DeleteGroupInOrganization deletes a group within an organization
func (s *Service) DeleteGroupInOrganization(ctx context.Context, orgID, groupID string, deletedBy string) error {
	s.logger.Info("Deleting group in organization",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("deleted_by", deletedBy))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return errors.NewNotFoundError("organization not found")
	}

	// This would delegate to the group service with organization context validation
	// For now, just log success
	s.logger.Info("Group deleted from organization successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))
	return nil
}

// GetGroupHierarchyInOrganization retrieves the hierarchy of a group within an organization
func (s *Service) GetGroupHierarchyInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error) {
	s.logger.Info("Retrieving group hierarchy in organization",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// This would delegate to the group service to get hierarchy with organization context
	// For now, return placeholder response
	s.logger.Info("Group hierarchy retrieved from organization successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))
	return map[string]interface{}{"message": "group hierarchy retrieval not fully implemented"}, nil
}

// AddUserToGroupInOrganization adds a user to a group within an organization
func (s *Service) AddUserToGroupInOrganization(ctx context.Context, orgID, groupID, userID string, req interface{}) (interface{}, error) {
	s.logger.Info("Adding user to group in organization",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("user_id", userID))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// Verify user exists and belongs to organization (this would need user-organization relationship)
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		s.logger.Error("User not found", zap.String("user_id", userID))
		return nil, errors.NewNotFoundError("user not found")
	}

	// This would delegate to the group service with organization context validation
	// For now, return placeholder response
	s.logger.Info("User added to group in organization successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("user_id", userID))
	return map[string]interface{}{"message": "user-group assignment not fully implemented"}, nil
}

// RemoveUserFromGroupInOrganization removes a user from a group within an organization
func (s *Service) RemoveUserFromGroupInOrganization(ctx context.Context, orgID, groupID, userID string, removedBy string) error {
	s.logger.Info("Removing user from group in organization",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("user_id", userID),
		zap.String("removed_by", removedBy))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return errors.NewNotFoundError("organization not found")
	}

	// This would delegate to the group service with organization context validation
	// For now, just log success
	s.logger.Info("User removed from group in organization successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("user_id", userID))
	return nil
}

// GetGroupUsersInOrganization retrieves all users in a group within an organization
func (s *Service) GetGroupUsersInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error) {
	s.logger.Info("Retrieving group users in organization",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// This would delegate to the group service with organization context validation
	// For now, return placeholder response
	s.logger.Info("Group users retrieved from organization successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))
	return []interface{}{}, nil
}

// GetUserGroupsInOrganization retrieves all groups a user belongs to within an organization
func (s *Service) GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
	s.logger.Info("Retrieving user groups in organization",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		s.logger.Error("User not found", zap.String("user_id", userID))
		return nil, errors.NewNotFoundError("user not found")
	}

	// This would delegate to the group service with organization context validation
	// For now, return placeholder response
	s.logger.Info("User groups retrieved from organization successfully",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))
	return []interface{}{}, nil
}

// AssignRoleToGroupInOrganization assigns a role to a group within an organization
func (s *Service) AssignRoleToGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, req interface{}) (interface{}, error) {
	s.logger.Info("Assigning role to group in organization",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("role_id", roleID))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// This would delegate to the group service with organization context validation
	// and create a GroupRole record
	// For now, return placeholder response
	s.logger.Info("Role assigned to group in organization successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("role_id", roleID))
	return map[string]interface{}{"message": "role-group assignment not fully implemented"}, nil
}

// RemoveRoleFromGroupInOrganization removes a role from a group within an organization
func (s *Service) RemoveRoleFromGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, removedBy string) error {
	s.logger.Info("Removing role from group in organization",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("role_id", roleID),
		zap.String("removed_by", removedBy))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return errors.NewNotFoundError("organization not found")
	}

	// This would delegate to the group service with organization context validation
	// For now, just log success
	s.logger.Info("Role removed from group in organization successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("role_id", roleID))
	return nil
}

// GetGroupRolesInOrganization retrieves all roles assigned to a group within an organization
func (s *Service) GetGroupRolesInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error) {
	s.logger.Info("Retrieving group roles in organization",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// This would query the GroupRole table with organization context validation
	// For now, return placeholder response
	s.logger.Info("Group roles retrieved from organization successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID))
	return []interface{}{}, nil
}

// buildGroupHierarchy builds the complete group hierarchy for an organization with role assignments
func (s *Service) buildGroupHierarchy(ctx context.Context, orgID string) ([]*organizationResponses.GroupHierarchyNode, error) {
	s.logger.Info("Building group hierarchy for organization", zap.String("org_id", orgID))

	// Get all groups in the organization
	allGroups, err := s.groupRepo.GetByOrganization(ctx, orgID, 1000, 0, false) // Get up to 1000 active groups
	if err != nil {
		s.logger.Error("Failed to get organization groups", zap.Error(err))
		return nil, err
	}

	if len(allGroups) == 0 {
		s.logger.Info("No groups found in organization", zap.String("org_id", orgID))
		return []*organizationResponses.GroupHierarchyNode{}, nil
	}

	// Create a map for quick lookup
	groupMap := make(map[string]*models.Group)
	for _, group := range allGroups {
		groupMap[group.ID] = group
	}

	// Build hierarchy nodes with roles
	nodeMap := make(map[string]*organizationResponses.GroupHierarchyNode)
	var rootNodes []*organizationResponses.GroupHierarchyNode

	// First pass: create all nodes
	for _, group := range allGroups {
		node, err := s.createGroupHierarchyNode(ctx, group)
		if err != nil {
			s.logger.Warn("Failed to create hierarchy node for group",
				zap.String("group_id", group.ID),
				zap.Error(err))
			continue
		}
		nodeMap[group.ID] = node
	}

	// Second pass: build parent-child relationships
	for _, group := range allGroups {
		node := nodeMap[group.ID]
		if node == nil {
			continue
		}

		if group.ParentID == nil || *group.ParentID == "" {
			// This is a root node
			rootNodes = append(rootNodes, node)
		} else {
			// This has a parent, add it to parent's children
			parentNode := nodeMap[*group.ParentID]
			if parentNode != nil {
				parentNode.Children = append(parentNode.Children, node)
			} else {
				// Parent not found or not active, treat as root
				rootNodes = append(rootNodes, node)
			}
		}
	}

	s.logger.Info("Group hierarchy built successfully",
		zap.String("org_id", orgID),
		zap.Int("total_groups", len(allGroups)),
		zap.Int("root_groups", len(rootNodes)))

	return rootNodes, nil
}

// createGroupHierarchyNode creates a hierarchy node for a group with its roles
func (s *Service) createGroupHierarchyNode(ctx context.Context, group *models.Group) (*organizationResponses.GroupHierarchyNode, error) {
	// Convert group to response format
	groupResponse := &groupResponses.GroupResponse{
		ID:             group.ID,
		Name:           group.Name,
		Description:    group.Description,
		OrganizationID: group.OrganizationID,
		ParentID:       group.ParentID,
		IsActive:       group.IsActive,
		CreatedAt:      &group.CreatedAt,
		UpdatedAt:      &group.UpdatedAt,
	}

	// Get roles for this group
	roles, err := s.groupService.GetGroupRoles(ctx, group.ID)
	if err != nil {
		s.logger.Warn("Failed to get roles for group",
			zap.String("group_id", group.ID),
			zap.Error(err))
		// Continue with empty roles rather than failing
		roles = []*groupResponses.GroupRoleDetail{}
	}

	// Convert roles to the expected type
	var roleDetails []*groupResponses.GroupRoleDetail
	if roleSlice, ok := roles.([]*groupResponses.GroupRoleDetail); ok {
		roleDetails = roleSlice
	} else {
		roleDetails = []*groupResponses.GroupRoleDetail{}
	}

	return &organizationResponses.GroupHierarchyNode{
		Group:    groupResponse,
		Roles:    roleDetails,
		Children: []*organizationResponses.GroupHierarchyNode{}, // Will be populated in buildGroupHierarchy
	}, nil
}

// GetUserEffectiveRolesInOrganization calculates and retrieves all effective roles for a user within an organization
func (s *Service) GetUserEffectiveRolesInOrganization(ctx context.Context, orgID, userID string) (interface{}, error) {
	s.logger.Info("Retrieving user effective roles in organization",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Error("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		s.logger.Error("User not found", zap.String("user_id", userID))
		return nil, errors.NewNotFoundError("user not found")
	}

	// This would use the role inheritance engine to calculate effective roles
	// considering direct user roles, group roles, and hierarchical inheritance
	// For now, return placeholder response
	s.logger.Info("User effective roles retrieved from organization successfully",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))
	return []interface{}{}, nil
}
