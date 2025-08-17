package groups

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	groupRequests "github.com/Kisanlink/aaa-service/internal/entities/requests/groups"
	groupResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/groups"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/repositories/groups"
	"github.com/Kisanlink/aaa-service/internal/repositories/organizations"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
)

// Service handles business logic for group operations
type Service struct {
	groupRepo *groups.GroupRepository
	orgRepo   *organizations.OrganizationRepository
	validator interfaces.Validator
	logger    *zap.Logger
}

// NewGroupService creates a new group service instance
func NewGroupService(
	groupRepo *groups.GroupRepository,
	orgRepo *organizations.OrganizationRepository,
	validator interfaces.Validator,
	logger *zap.Logger,
) *Service {
	return &Service{
		groupRepo: groupRepo,
		orgRepo:   orgRepo,
		validator: validator,
		logger:    logger,
	}
}

// CreateGroup creates a new group with proper validation and business logic
func (s *Service) CreateGroup(ctx context.Context, req *groupRequests.CreateGroupRequest) (*groupResponses.GroupResponse, error) {
	s.logger.Info("Creating new group", zap.String("name", req.Name))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Group creation validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid group data", err.Error())
	}

	// Verify organization exists and is active
	org, err := s.orgRepo.GetByID(ctx, req.OrganizationID)
	if err != nil || org == nil {
		s.logger.Warn("Organization not found", zap.String("org_id", req.OrganizationID))
		return nil, errors.NewNotFoundError("organization not found")
	}
	if !org.IsActive {
		s.logger.Warn("Organization is inactive", zap.String("org_id", req.OrganizationID))
		return nil, errors.NewValidationError("cannot create group in inactive organization")
	}

	// Check if group name already exists in the organization
	existingGroup, err := s.groupRepo.GetByNameAndOrganization(ctx, req.Name, req.OrganizationID)
	if err == nil && existingGroup != nil {
		s.logger.Warn("Group name already exists in organization",
			zap.String("name", req.Name),
			zap.String("org_id", req.OrganizationID))
		return nil, errors.NewConflictError("group name already exists in this organization")
	}

	// Validate parent group if specified
	if req.ParentID != nil && *req.ParentID != "" {
		parentGroup, err := s.groupRepo.GetByID(ctx, *req.ParentID)
		if err != nil || parentGroup == nil {
			s.logger.Warn("Parent group not found", zap.String("parent_id", *req.ParentID))
			return nil, errors.NewNotFoundError("parent group not found")
		}
		if !parentGroup.IsActive {
			s.logger.Warn("Parent group is inactive", zap.String("parent_id", *req.ParentID))
			return nil, errors.NewValidationError("parent group is inactive")
		}
		if parentGroup.OrganizationID != req.OrganizationID {
			s.logger.Warn("Parent group belongs to different organization",
				zap.String("parent_id", *req.ParentID),
				zap.String("parent_org", parentGroup.OrganizationID),
				zap.String("req_org", req.OrganizationID))
			return nil, errors.NewValidationError("parent group must belong to the same organization")
		}
	}

	// Create group model
	group := models.NewGroup(req.Name, req.Description, req.OrganizationID)
	if req.ParentID != nil && *req.ParentID != "" {
		group.ParentID = req.ParentID
	}

	// Save group to repository
	err = s.groupRepo.Create(ctx, group)
	if err != nil {
		s.logger.Error("Failed to create group in repository", zap.Error(err))
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.NewConflictError("group with this information already exists")
		}
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Group created successfully",
		zap.String("group_id", group.ID),
		zap.String("name", group.Name),
		zap.String("org_id", group.OrganizationID))

	// Convert to response format
	response := &groupResponses.GroupResponse{
		ID:             group.ID,
		Name:           group.Name,
		Description:    group.Description,
		OrganizationID: group.OrganizationID,
		ParentID:       group.ParentID,
		IsActive:       group.IsActive,
		CreatedAt:      &group.CreatedAt,
		UpdatedAt:      &group.UpdatedAt,
	}

	return response, nil
}

// GetGroup retrieves a group by ID
func (s *Service) GetGroup(ctx context.Context, groupID string) (*groupResponses.GroupResponse, error) {
	s.logger.Info("Retrieving group", zap.String("group_id", groupID))

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		s.logger.Error("Failed to retrieve group", zap.Error(err))
		return nil, errors.NewNotFoundError("group not found")
	}

	response := &groupResponses.GroupResponse{
		ID:             group.ID,
		Name:           group.Name,
		Description:    group.Description,
		OrganizationID: group.OrganizationID,
		ParentID:       group.ParentID,
		IsActive:       group.IsActive,
		CreatedAt:      &group.CreatedAt,
		UpdatedAt:      &group.UpdatedAt,
	}

	return response, nil
}

// UpdateGroup updates an existing group
func (s *Service) UpdateGroup(ctx context.Context, groupID string, req *groupRequests.UpdateGroupRequest) (*groupResponses.GroupResponse, error) {
	s.logger.Info("Updating group", zap.String("group_id", groupID))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Group update validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid group data", err.Error())
	}

	// Get existing group
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		s.logger.Error("Group not found for update", zap.String("group_id", groupID))
		return nil, errors.NewNotFoundError("group not found")
	}

	// Check if name is being changed and if new name already exists in organization
	if req.Name != nil && *req.Name != group.Name {
		existingGroup, err := s.groupRepo.GetByNameAndOrganization(ctx, *req.Name, group.OrganizationID)
		if err == nil && existingGroup != nil && existingGroup.ID != groupID {
			s.logger.Warn("Group name already taken in organization", zap.String("name", *req.Name))
			return nil, errors.NewConflictError("group name already exists in this organization")
		}
	}

	// Validate parent group if being changed
	if req.ParentID != nil && (group.ParentID == nil || *req.ParentID != *group.ParentID) {
		if *req.ParentID != "" {
			parentGroup, err := s.groupRepo.GetByID(ctx, *req.ParentID)
			if err != nil || parentGroup == nil {
				s.logger.Warn("Parent group not found", zap.String("parent_id", *req.ParentID))
				return nil, errors.NewNotFoundError("parent group not found")
			}
			if !parentGroup.IsActive {
				s.logger.Warn("Parent group is inactive", zap.String("parent_id", *req.ParentID))
				return nil, errors.NewValidationError("parent group is inactive")
			}
			if parentGroup.OrganizationID != group.OrganizationID {
				s.logger.Warn("Parent group belongs to different organization", zap.String("parent_id", *req.ParentID))
				return nil, errors.NewValidationError("parent group must belong to the same organization")
			}
			// Check for circular references
			if err := s.checkCircularReference(ctx, groupID, *req.ParentID); err != nil {
				s.logger.Warn("Circular reference detected", zap.Error(err))
				return nil, errors.NewValidationError("circular reference detected in group hierarchy")
			}
		}
	}

	// Update fields
	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = *req.Description
	}
	if req.ParentID != nil {
		group.ParentID = req.ParentID
	}
	if req.IsActive != nil {
		group.IsActive = *req.IsActive
	}

	// Save changes
	err = s.groupRepo.Update(ctx, group)
	if err != nil {
		s.logger.Error("Failed to update group", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Group updated successfully", zap.String("group_id", groupID))

	response := &groupResponses.GroupResponse{
		ID:             group.ID,
		Name:           group.Name,
		Description:    group.Description,
		OrganizationID: group.OrganizationID,
		ParentID:       group.ParentID,
		IsActive:       group.IsActive,
		CreatedAt:      &group.CreatedAt,
		UpdatedAt:      &group.UpdatedAt,
	}

	return response, nil
}

// DeleteGroup deletes a group
func (s *Service) DeleteGroup(ctx context.Context, groupID string, deletedBy string) error {
	s.logger.Info("Deleting group", zap.String("group_id", groupID))

	// Get group to check if it has children
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		s.logger.Error("Group not found for deletion", zap.String("group_id", groupID))
		return errors.NewNotFoundError("group not found")
	}

	// Check if group has children
	children, err := s.groupRepo.GetChildren(ctx, groupID)
	if err != nil {
		s.logger.Error("Failed to check group children", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if len(children) > 0 {
		s.logger.Warn("Cannot delete group with children", zap.String("group_id", groupID))
		return errors.NewValidationError("cannot delete group with child groups")
	}

	// Check if group has active members
	hasMembers, err := s.groupRepo.HasActiveMembers(ctx, groupID)
	if err != nil {
		s.logger.Error("Failed to check group members", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if hasMembers {
		s.logger.Warn("Cannot delete group with active members", zap.String("group_id", groupID))
		return errors.NewValidationError("cannot delete group with active members")
	}

	// Soft delete the group
	err = s.groupRepo.SoftDelete(ctx, groupID, deletedBy)
	if err != nil {
		s.logger.Error("Failed to delete group", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Group deleted successfully", zap.String("group_id", groupID))
	return nil
}

// ListGroups retrieves groups with pagination and filtering
func (s *Service) ListGroups(ctx context.Context, limit, offset int, organizationID string, includeInactive bool) ([]*groupResponses.GroupResponse, error) {
	s.logger.Info("Listing groups", zap.Int("limit", limit), zap.Int("offset", offset))

	var groups []*models.Group
	var err error

	if organizationID != "" {
		groups, err = s.groupRepo.GetByOrganization(ctx, organizationID, limit, offset, includeInactive)
	} else {
		if includeInactive {
			groups, err = s.groupRepo.List(ctx, limit, offset)
		} else {
			groups, err = s.groupRepo.ListActive(ctx, limit, offset)
		}
	}

	if err != nil {
		s.logger.Error("Failed to list groups", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format
	responses := make([]*groupResponses.GroupResponse, len(groups))
	for i, group := range groups {
		responses[i] = &groupResponses.GroupResponse{
			ID:             group.ID,
			Name:           group.Name,
			Description:    group.Description,
			OrganizationID: group.OrganizationID,
			ParentID:       group.ParentID,
			IsActive:       group.IsActive,
			CreatedAt:      &group.CreatedAt,
			UpdatedAt:      &group.UpdatedAt,
		}
	}

	return responses, nil
}

// AddMemberToGroup adds a member to a group
func (s *Service) AddMemberToGroup(ctx context.Context, req *groupRequests.AddMemberRequest) (*groupResponses.GroupMembershipResponse, error) {
	s.logger.Info("Adding member to group",
		zap.String("group_id", req.GroupID),
		zap.String("principal_id", req.PrincipalID))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Add member validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid request data", err.Error())
	}

	// Verify group exists and is active
	group, err := s.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil || group == nil {
		s.logger.Warn("Group not found", zap.String("group_id", req.GroupID))
		return nil, errors.NewNotFoundError("group not found")
	}
	if !group.IsActive {
		s.logger.Warn("Group is inactive", zap.String("group_id", req.GroupID))
		return nil, errors.NewValidationError("cannot add member to inactive group")
	}

	// Check if member is already in the group
	existingMembership, err := s.groupRepo.GetMembership(ctx, req.GroupID, req.PrincipalID)
	if err == nil && existingMembership != nil {
		s.logger.Warn("Member already in group",
			zap.String("group_id", req.GroupID),
			zap.String("principal_id", req.PrincipalID))
		return nil, errors.NewConflictError("member is already in this group")
	}

	// Create membership
	membership := models.NewGroupMembership(req.GroupID, req.PrincipalID, req.PrincipalType, req.AddedByID)

	// Set time bounds if provided
	if req.StartsAt != nil {
		membership.StartsAt = req.StartsAt
	}
	if req.EndsAt != nil {
		membership.EndsAt = req.EndsAt
	}

	// Save membership
	err = s.groupRepo.CreateMembership(ctx, membership)
	if err != nil {
		s.logger.Error("Failed to create group membership", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Member added to group successfully",
		zap.String("group_id", req.GroupID),
		zap.String("principal_id", req.PrincipalID))

	response := &groupResponses.GroupMembershipResponse{
		ID:            membership.ID,
		GroupID:       membership.GroupID,
		PrincipalID:   membership.PrincipalID,
		PrincipalType: membership.PrincipalType,
		StartsAt:      membership.StartsAt,
		EndsAt:        membership.EndsAt,
		IsActive:      membership.IsActive,
		AddedByID:     membership.AddedByID,
		CreatedAt:     &membership.CreatedAt,
	}

	return response, nil
}

// RemoveMemberFromGroup removes a member from a group
func (s *Service) RemoveMemberFromGroup(ctx context.Context, groupID, principalID string, removedBy string) error {
	s.logger.Info("Removing member from group",
		zap.String("group_id", groupID),
		zap.String("principal_id", principalID))

	// Get existing membership
	membership, err := s.groupRepo.GetMembership(ctx, groupID, principalID)
	if err != nil || membership == nil {
		s.logger.Warn("Membership not found",
			zap.String("group_id", groupID),
			zap.String("principal_id", principalID))
		return errors.NewNotFoundError("membership not found")
	}

	// Deactivate membership
	membership.IsActive = false
	err = s.groupRepo.UpdateMembership(ctx, membership)
	if err != nil {
		s.logger.Error("Failed to remove member from group", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Member removed from group successfully",
		zap.String("group_id", groupID),
		zap.String("principal_id", principalID))

	return nil
}

// GetGroupMembers retrieves all members of a group
func (s *Service) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) ([]*groupResponses.GroupMembershipResponse, error) {
	s.logger.Info("Retrieving group members", zap.String("group_id", groupID))

	memberships, err := s.groupRepo.GetGroupMembers(ctx, groupID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve group members", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format
	responses := make([]*groupResponses.GroupMembershipResponse, len(memberships))
	for i, membership := range memberships {
		responses[i] = &groupResponses.GroupMembershipResponse{
			ID:            membership.ID,
			GroupID:       membership.GroupID,
			PrincipalID:   membership.PrincipalID,
			PrincipalType: membership.PrincipalType,
			StartsAt:      membership.StartsAt,
			EndsAt:        membership.EndsAt,
			IsActive:      membership.IsActive,
			AddedByID:     membership.AddedByID,
			CreatedAt:     &membership.CreatedAt,
		}
	}

	return responses, nil
}

// checkCircularReference checks if setting a parent would create a circular reference
func (s *Service) checkCircularReference(ctx context.Context, groupID, newParentID string) error {
	// Start from the new parent and traverse up the hierarchy
	currentID := newParentID
	visited := make(map[string]bool)

	for currentID != "" {
		if currentID == groupID {
			return fmt.Errorf("circular reference detected: %s would be its own ancestor", groupID)
		}

		if visited[currentID] {
			return fmt.Errorf("circular reference detected in group hierarchy")
		}

		visited[currentID] = true

		// Get the current group's parent
		group, err := s.groupRepo.GetByID(ctx, currentID)
		if err != nil || group == nil {
			break
		}

		currentID = ""
		if group.ParentID != nil {
			currentID = *group.ParentID
		}
	}

	return nil
}
