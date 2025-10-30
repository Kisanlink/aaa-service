package groups

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	groupRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/groups"
	groupResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/groups"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/groups"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/organizations"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/roles"
	"github.com/Kisanlink/aaa-service/v2/internal/services/user"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"go.uber.org/zap"
)

// Service handles business logic for group operations
type Service struct {
	groupRepo           *groups.GroupRepository
	groupRoleRepo       *groups.GroupRoleRepository
	groupMembershipRepo *groups.GroupMembershipRepository
	orgRepo             *organizations.OrganizationRepository
	roleRepo            *roles.RoleRepository
	validator           interfaces.Validator
	cache               interfaces.CacheService
	groupCache          *GroupCacheService
	auditService        interfaces.AuditService
	userService         interfaces.UserService // For invalidating user organizational cache
	logger              *zap.Logger
}

// NewGroupService creates a new group service instance
func NewGroupService(
	groupRepo *groups.GroupRepository,
	groupRoleRepo *groups.GroupRoleRepository,
	groupMembershipRepo *groups.GroupMembershipRepository,
	orgRepo *organizations.OrganizationRepository,
	roleRepo *roles.RoleRepository,
	validator interfaces.Validator,
	cache interfaces.CacheService,
	auditService interfaces.AuditService,
	logger *zap.Logger,
) *Service {
	groupCache := NewGroupCacheService(cache, logger)
	return &Service{
		groupRepo:           groupRepo,
		groupRoleRepo:       groupRoleRepo,
		groupMembershipRepo: groupMembershipRepo,
		orgRepo:             orgRepo,
		roleRepo:            roleRepo,
		validator:           validator,
		cache:               cache,
		groupCache:          groupCache,
		auditService:        auditService,
		logger:              logger,
	}
}

// SetUserService sets the user service for cache invalidation
// This is called after service construction to avoid circular dependencies
func (s *Service) SetUserService(userService interfaces.UserService) {
	s.userService = userService
	s.logger.Debug("User service injected into group service for cache invalidation")
}

// CreateGroup creates a new group with proper validation and business logic
func (s *Service) CreateGroup(ctx context.Context, req interface{}) (interface{}, error) {
	createReq, ok := req.(*groupRequests.CreateGroupRequest)
	if !ok {
		return nil, errors.NewValidationError("invalid request type for CreateGroup")
	}
	s.logger.Info("Creating new group", zap.String("name", createReq.Name))

	// Validate request
	if err := s.validator.ValidateStruct(createReq); err != nil {
		s.logger.Error("Group creation validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid group data", err.Error())
	}

	// Verify organization exists and is active
	org, err := s.orgRepo.GetByID(ctx, createReq.OrganizationID)
	if err != nil || org == nil {
		s.logger.Warn("Organization not found", zap.String("org_id", createReq.OrganizationID))
		return nil, errors.NewNotFoundError("organization not found")
	}
	if !org.IsActive {
		s.logger.Warn("Organization is inactive", zap.String("org_id", createReq.OrganizationID))
		return nil, errors.NewValidationError("cannot create group in inactive organization")
	}

	// Check if group name already exists in the organization
	existingGroup, err := s.groupRepo.GetByNameAndOrganization(ctx, createReq.Name, createReq.OrganizationID)
	if err == nil && existingGroup != nil {
		s.logger.Warn("Group name already exists in organization",
			zap.String("name", createReq.Name),
			zap.String("org_id", createReq.OrganizationID))
		return nil, errors.NewConflictError("group name already exists in this organization")
	}

	// Validate parent group if specified
	if createReq.ParentID != nil && *createReq.ParentID != "" {
		parentGroup, err := s.groupRepo.GetByID(ctx, *createReq.ParentID)
		if err != nil || parentGroup == nil {
			s.logger.Warn("Parent group not found", zap.String("parent_id", *createReq.ParentID))
			return nil, errors.NewNotFoundError("parent group not found")
		}
		if !parentGroup.IsActive {
			s.logger.Warn("Parent group is inactive", zap.String("parent_id", *createReq.ParentID))
			return nil, errors.NewValidationError("parent group is inactive")
		}
		if parentGroup.OrganizationID != createReq.OrganizationID {
			s.logger.Warn("Parent group belongs to different organization",
				zap.String("parent_id", *createReq.ParentID),
				zap.String("parent_org", parentGroup.OrganizationID),
				zap.String("req_org", createReq.OrganizationID))
			return nil, errors.NewValidationError("parent group must belong to the same organization")
		}
	}

	// Create group model
	group := models.NewGroup(createReq.Name, createReq.Description, createReq.OrganizationID)
	if createReq.ParentID != nil && *createReq.ParentID != "" {
		group.ParentID = createReq.ParentID
	}

	// Save group to repository
	err = s.groupRepo.Create(ctx, group)
	if err != nil {
		s.logger.Error("Failed to create group in repository", zap.Error(err))

		// Log audit event for failed creation
		auditDetails := map[string]interface{}{
			"group_name":  createReq.Name,
			"parent_id":   createReq.ParentID,
			"description": createReq.Description,
			"error":       err.Error(),
		}
		s.auditService.LogGroupOperation(ctx, "system", models.AuditActionCreateGroup, createReq.OrganizationID, "", "Failed to create group", false, auditDetails)

		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.NewConflictError("group with this information already exists")
		}
		return nil, errors.NewInternalError(err)
	}

	// Log successful group creation
	auditDetails := map[string]interface{}{
		"group_name":  group.Name,
		"parent_id":   group.ParentID,
		"description": group.Description,
		"is_active":   group.IsActive,
	}
	s.auditService.LogGroupOperation(ctx, "system", models.AuditActionCreateGroup, group.OrganizationID, group.ID, "Group created successfully", true, auditDetails)

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
func (s *Service) GetGroup(ctx context.Context, groupID string) (interface{}, error) {
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
func (s *Service) UpdateGroup(ctx context.Context, groupID string, req interface{}) (interface{}, error) {
	updateReq, ok := req.(*groupRequests.UpdateGroupRequest)
	if !ok {
		return nil, errors.NewValidationError("invalid request type for UpdateGroup")
	}
	s.logger.Info("Updating group", zap.String("group_id", groupID))

	// Validate request
	if err := s.validator.ValidateStruct(updateReq); err != nil {
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
	if updateReq.Name != nil && *updateReq.Name != group.Name {
		existingGroup, err := s.groupRepo.GetByNameAndOrganization(ctx, *updateReq.Name, group.OrganizationID)
		if err == nil && existingGroup != nil && existingGroup.ID != groupID {
			s.logger.Warn("Group name already taken in organization", zap.String("name", *updateReq.Name))
			return nil, errors.NewConflictError("group name already exists in this organization")
		}
	}

	// Validate parent group if being changed
	if updateReq.ParentID != nil && (group.ParentID == nil || *updateReq.ParentID != *group.ParentID) {
		if *updateReq.ParentID != "" {
			parentGroup, err := s.groupRepo.GetByID(ctx, *updateReq.ParentID)
			if err != nil || parentGroup == nil {
				s.logger.Warn("Parent group not found", zap.String("parent_id", *updateReq.ParentID))
				return nil, errors.NewNotFoundError("parent group not found")
			}
			if !parentGroup.IsActive {
				s.logger.Warn("Parent group is inactive", zap.String("parent_id", *updateReq.ParentID))
				return nil, errors.NewValidationError("parent group is inactive")
			}
			if parentGroup.OrganizationID != group.OrganizationID {
				s.logger.Warn("Parent group belongs to different organization", zap.String("parent_id", *updateReq.ParentID))
				return nil, errors.NewValidationError("parent group must belong to the same organization")
			}
			// Check for circular references
			if err := s.checkCircularReference(ctx, groupID, *updateReq.ParentID); err != nil {
				s.logger.Warn("Circular reference detected", zap.Error(err))
				return nil, errors.NewValidationError("circular reference detected in group hierarchy")
			}
		}
	}

	// Capture old values for audit logging
	oldValues := map[string]interface{}{
		"name":        group.Name,
		"description": group.Description,
		"parent_id":   group.ParentID,
		"is_active":   group.IsActive,
	}

	// Track hierarchy changes for special audit logging
	hierarchyChanged := false
	oldParentID := ""
	newParentID := ""
	if group.ParentID != nil {
		oldParentID = *group.ParentID
	}

	// Update fields
	if updateReq.Name != nil {
		group.Name = *updateReq.Name
	}
	if updateReq.Description != nil {
		group.Description = *updateReq.Description
	}
	if updateReq.ParentID != nil {
		if group.ParentID == nil || *updateReq.ParentID != *group.ParentID {
			hierarchyChanged = true
			if updateReq.ParentID != nil {
				newParentID = *updateReq.ParentID
			}
		}
		group.ParentID = updateReq.ParentID
	}
	if updateReq.IsActive != nil {
		group.IsActive = *updateReq.IsActive
	}

	// Capture new values for audit logging
	newValues := map[string]interface{}{
		"name":        group.Name,
		"description": group.Description,
		"parent_id":   group.ParentID,
		"is_active":   group.IsActive,
	}

	// Save changes
	err = s.groupRepo.Update(ctx, group)
	if err != nil {
		s.logger.Error("Failed to update group", zap.Error(err))

		// Log audit event for failed update
		auditDetails := map[string]interface{}{
			"old_values": oldValues,
			"new_values": newValues,
			"error":      err.Error(),
		}
		s.auditService.LogGroupOperation(ctx, "system", models.AuditActionUpdateGroup, group.OrganizationID, groupID, "Failed to update group", false, auditDetails)

		return nil, errors.NewInternalError(err)
	}

	// Log successful group update
	auditDetails := map[string]interface{}{
		"old_values": oldValues,
		"new_values": newValues,
	}
	s.auditService.LogGroupOperation(ctx, "system", models.AuditActionUpdateGroup, group.OrganizationID, groupID, "Group updated successfully", true, auditDetails)

	// Log hierarchy change separately if it occurred with comprehensive structure change logging
	if hierarchyChanged {
		s.auditService.LogHierarchyChange(ctx, "system", models.AuditActionChangeGroupHierarchy, models.ResourceTypeGroup, groupID, oldParentID, newParentID, "Group hierarchy changed", true, auditDetails)

		// Also log comprehensive structure change for enhanced audit trail
		hierarchyOldValues := map[string]interface{}{
			"parent_id": oldParentID,
		}
		hierarchyNewValues := map[string]interface{}{
			"parent_id": newParentID,
		}
		s.auditService.LogOrganizationStructureChange(ctx, "system", models.AuditActionChangeGroupHierarchy, group.OrganizationID, models.ResourceTypeGroup, groupID, hierarchyOldValues, hierarchyNewValues, true, "Group hierarchy structure changed")
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

		// Log audit event for failed deletion
		auditDetails := map[string]interface{}{
			"group_name":   group.Name,
			"deleted_by":   deletedBy,
			"had_children": len(children) > 0,
			"had_members":  hasMembers,
			"error":        err.Error(),
		}
		s.auditService.LogGroupOperation(ctx, deletedBy, models.AuditActionDeleteGroup, group.OrganizationID, groupID, "Failed to delete group", false, auditDetails)

		return errors.NewInternalError(err)
	}

	// Log successful group deletion
	auditDetails := map[string]interface{}{
		"group_name":   group.Name,
		"deleted_by":   deletedBy,
		"had_children": len(children) > 0,
		"had_members":  hasMembers,
	}
	s.auditService.LogGroupOperation(ctx, deletedBy, models.AuditActionDeleteGroup, group.OrganizationID, groupID, "Group deleted successfully", true, auditDetails)

	s.logger.Info("Group deleted successfully", zap.String("group_id", groupID))
	return nil
}

// ListGroups retrieves groups with pagination and filtering
func (s *Service) ListGroups(ctx context.Context, limit, offset int, organizationID string, includeInactive bool) (interface{}, error) {
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
func (s *Service) AddMemberToGroup(ctx context.Context, req interface{}) (interface{}, error) {
	addMemberReq, ok := req.(*groupRequests.AddMemberRequest)
	if !ok {
		return nil, errors.NewValidationError("invalid request type for AddMemberToGroup")
	}
	s.logger.Info("Adding member to group",
		zap.String("group_id", addMemberReq.GroupID),
		zap.String("principal_id", addMemberReq.PrincipalID))

	// Validate request
	if err := s.validator.ValidateStruct(addMemberReq); err != nil {
		s.logger.Error("Add member validation failed",
			zap.Error(err),
			zap.String("group_id", addMemberReq.GroupID),
			zap.String("principal_id", addMemberReq.PrincipalID),
			zap.String("principal_type", addMemberReq.PrincipalType),
			zap.String("added_by_id", addMemberReq.AddedByID),
			zap.Any("starts_at", addMemberReq.StartsAt),
			zap.Any("ends_at", addMemberReq.EndsAt))
		return nil, errors.NewValidationError("invalid request data", err.Error())
	}

	// Verify group exists and is active
	group, err := s.groupRepo.GetByID(ctx, addMemberReq.GroupID)
	if err != nil || group == nil {
		s.logger.Warn("Group not found", zap.String("group_id", addMemberReq.GroupID))
		return nil, errors.NewNotFoundError("group not found")
	}
	if !group.IsActive {
		s.logger.Warn("Group is inactive", zap.String("group_id", addMemberReq.GroupID))
		return nil, errors.NewValidationError("cannot add member to inactive group")
	}

	// Check if member is already in the group
	existingMembership, err := s.groupRepo.GetMembership(ctx, addMemberReq.GroupID, addMemberReq.PrincipalID)
	if err == nil && existingMembership != nil {
		s.logger.Warn("Member already in group",
			zap.String("group_id", addMemberReq.GroupID),
			zap.String("principal_id", addMemberReq.PrincipalID))
		return nil, errors.NewConflictError("member is already in this group")
	}

	// Create membership
	membership := models.NewGroupMembership(addMemberReq.GroupID, addMemberReq.PrincipalID, addMemberReq.PrincipalType, addMemberReq.AddedByID)

	// Set time bounds if provided
	if addMemberReq.StartsAt != nil {
		membership.StartsAt = addMemberReq.StartsAt
	}
	if addMemberReq.EndsAt != nil {
		membership.EndsAt = addMemberReq.EndsAt
	}

	// Save membership
	err = s.groupRepo.CreateMembership(ctx, membership)
	if err != nil {
		s.logger.Error("Failed to create group membership", zap.Error(err))

		// Log audit event for failed membership addition
		auditDetails := map[string]interface{}{
			"principal_type": addMemberReq.PrincipalType,
			"added_by":       addMemberReq.AddedByID,
			"starts_at":      addMemberReq.StartsAt,
			"ends_at":        addMemberReq.EndsAt,
			"error":          err.Error(),
		}
		s.auditService.LogGroupMembershipChange(ctx, addMemberReq.AddedByID, models.AuditActionAddGroupMember, group.OrganizationID, addMemberReq.GroupID, addMemberReq.PrincipalID, "Failed to add member to group", false, auditDetails)

		return nil, errors.NewInternalError(err)
	}

	// Invalidate user's organizational cache so they see the organization immediately
	if s.userService != nil {
		if userSvc, ok := s.userService.(*user.Service); ok {
			userSvc.InvalidateUserOrganizationalCache(addMemberReq.PrincipalID)
			s.logger.Debug("Invalidated user organizational cache after adding to group",
				zap.String("user_id", addMemberReq.PrincipalID),
				zap.String("group_id", addMemberReq.GroupID))
		}
	}

	// Log successful membership addition
	auditDetails := map[string]interface{}{
		"principal_type": membership.PrincipalType,
		"added_by":       membership.AddedByID,
		"starts_at":      membership.StartsAt,
		"ends_at":        membership.EndsAt,
		"is_active":      membership.IsActive,
	}
	s.auditService.LogGroupMembershipChange(ctx, addMemberReq.AddedByID, models.AuditActionAddGroupMember, group.OrganizationID, addMemberReq.GroupID, addMemberReq.PrincipalID, "Member added to group successfully", true, auditDetails)

	s.logger.Info("Member added to group successfully",
		zap.String("group_id", addMemberReq.GroupID),
		zap.String("principal_id", addMemberReq.PrincipalID))

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

	// Get group for organization context
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		s.logger.Error("Group not found for membership removal", zap.String("group_id", groupID))
		return errors.NewNotFoundError("group not found")
	}

	// Deactivate membership
	membership.IsActive = false
	err = s.groupRepo.UpdateMembership(ctx, membership)
	if err != nil {
		s.logger.Error("Failed to remove member from group", zap.Error(err))

		// Log audit event for failed membership removal
		auditDetails := map[string]interface{}{
			"principal_type": membership.PrincipalType,
			"removed_by":     removedBy,
			"was_active":     true,
			"error":          err.Error(),
		}
		s.auditService.LogGroupMembershipChange(ctx, removedBy, models.AuditActionRemoveGroupMember, group.OrganizationID, groupID, principalID, "Failed to remove member from group", false, auditDetails)

		return errors.NewInternalError(err)
	}

	// Invalidate user's organizational cache after removing from group
	if s.userService != nil {
		if userSvc, ok := s.userService.(*user.Service); ok {
			userSvc.InvalidateUserOrganizationalCache(principalID)
			s.logger.Debug("Invalidated user organizational cache after removing from group",
				zap.String("user_id", principalID),
				zap.String("group_id", groupID))
		}
	}

	// Log successful membership removal
	auditDetails := map[string]interface{}{
		"principal_type": membership.PrincipalType,
		"removed_by":     removedBy,
		"was_active":     true,
		"starts_at":      membership.StartsAt,
		"ends_at":        membership.EndsAt,
	}
	s.auditService.LogGroupMembershipChange(ctx, removedBy, models.AuditActionRemoveGroupMember, group.OrganizationID, groupID, principalID, "Member removed from group successfully", true, auditDetails)

	s.logger.Info("Member removed from group successfully",
		zap.String("group_id", groupID),
		zap.String("principal_id", principalID))

	return nil
}

// AssignRoleToGroup assigns a role to a group with organization-scoped validation
func (s *Service) AssignRoleToGroup(ctx context.Context, groupID, roleID, assignedBy string) (interface{}, error) {
	s.logger.Info("Assigning role to group",
		zap.String("group_id", groupID),
		zap.String("role_id", roleID),
		zap.String("assigned_by", assignedBy))

	// Validate input parameters
	if groupID == "" {
		return nil, errors.NewValidationError("group_id cannot be empty")
	}
	if roleID == "" {
		return nil, errors.NewValidationError("role_id cannot be empty")
	}
	if assignedBy == "" {
		return nil, errors.NewValidationError("assigned_by cannot be empty")
	}

	// Verify group exists and is active
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		s.logger.Warn("Group not found", zap.String("group_id", groupID))
		return nil, errors.NewNotFoundError("group not found")
	}
	if !group.IsActive {
		s.logger.Warn("Group is inactive", zap.String("group_id", groupID))
		return nil, errors.NewValidationError("cannot assign role to inactive group")
	}

	// Verify role exists and is active
	role := &models.Role{}
	role, err = s.roleRepo.GetByID(ctx, roleID, role)
	if err != nil || role == nil {
		s.logger.Warn("Role not found", zap.String("role_id", roleID))
		return nil, errors.NewNotFoundError("role not found")
	}
	if !role.IsActive {
		s.logger.Warn("Role is inactive", zap.String("role_id", roleID))
		return nil, errors.NewValidationError("cannot assign inactive role to group")
	}

	// Verify organization exists and is active
	org, err := s.orgRepo.GetByID(ctx, group.OrganizationID)
	if err != nil || org == nil {
		s.logger.Warn("Organization not found", zap.String("org_id", group.OrganizationID))
		return nil, errors.NewNotFoundError("organization not found")
	}
	if !org.IsActive {
		s.logger.Warn("Organization is inactive", zap.String("org_id", group.OrganizationID))
		return nil, errors.NewValidationError("cannot assign role in inactive organization")
	}

	// Check if role is already assigned to the group
	exists, err := s.groupRoleRepo.ExistsByGroupAndRole(ctx, groupID, roleID)
	if err != nil {
		s.logger.Error("Failed to check existing group role assignment", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}
	if exists {
		s.logger.Warn("Role already assigned to group",
			zap.String("group_id", groupID),
			zap.String("role_id", roleID))
		return nil, errors.NewConflictError("role is already assigned to this group")
	}

	// Create group role assignment
	groupRole := models.NewGroupRole(groupID, roleID, group.OrganizationID, assignedBy)

	// Save group role assignment
	err = s.groupRoleRepo.Create(ctx, groupRole)
	if err != nil {
		s.logger.Error("Failed to create group role assignment", zap.Error(err))

		// Log audit event for failed role assignment
		auditDetails := map[string]interface{}{
			"role_name":   role.Name,
			"group_name":  group.Name,
			"assigned_by": assignedBy,
			"is_active":   groupRole.IsActive,
			"error":       err.Error(),
		}
		s.auditService.LogGroupRoleAssignment(ctx, assignedBy, models.AuditActionAssignGroupRole, group.OrganizationID, groupID, roleID, "Failed to assign role to group", false, auditDetails)

		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.NewConflictError("role assignment already exists")
		}
		return nil, errors.NewInternalError(err)
	}

	// Log successful role assignment
	auditDetails := map[string]interface{}{
		"role_name":   role.Name,
		"group_name":  group.Name,
		"assigned_by": assignedBy,
		"is_active":   groupRole.IsActive,
		"starts_at":   groupRole.StartsAt,
		"ends_at":     groupRole.EndsAt,
	}
	s.auditService.LogGroupRoleAssignment(ctx, assignedBy, models.AuditActionAssignGroupRole, group.OrganizationID, groupID, roleID, "Role assigned to group successfully", true, auditDetails)

	// Invalidate relevant caches after successful role assignment
	s.groupCache.InvalidateRoleAssignmentCache(ctx, group.OrganizationID, groupID, roleID)

	s.logger.Info("Role assigned to group successfully",
		zap.String("group_id", groupID),
		zap.String("role_id", roleID),
		zap.String("org_id", group.OrganizationID))

	// Create response with role details
	response := groupResponses.NewGroupRoleResponse(groupRole, "Role assigned to group successfully")
	response.Role = groupResponses.RoleDetail{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsActive:    role.IsActive,
	}

	return response, nil
}

// RemoveRoleFromGroup removes a role from a group with organization-scoped validation
func (s *Service) RemoveRoleFromGroup(ctx context.Context, groupID, roleID string) error {
	s.logger.Info("Removing role from group",
		zap.String("group_id", groupID),
		zap.String("role_id", roleID))

	// Validate input parameters
	if groupID == "" {
		return errors.NewValidationError("group_id cannot be empty")
	}
	if roleID == "" {
		return errors.NewValidationError("role_id cannot be empty")
	}

	// Verify group exists
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		s.logger.Warn("Group not found", zap.String("group_id", groupID))
		return errors.NewNotFoundError("group not found")
	}

	// Get existing group role assignment
	groupRole, err := s.groupRoleRepo.GetByGroupAndRole(ctx, groupID, roleID)
	if err != nil {
		s.logger.Error("Failed to get group role assignment", zap.Error(err))
		return errors.NewInternalError(err)
	}
	if groupRole == nil {
		s.logger.Warn("Group role assignment not found",
			zap.String("group_id", groupID),
			zap.String("role_id", roleID))
		return errors.NewNotFoundError("role assignment not found")
	}

	// Get role details for audit logging
	role := &models.Role{}
	role, roleErr := s.roleRepo.GetByID(ctx, roleID, role)
	if roleErr != nil {
		s.logger.Warn("Failed to get role details for audit", zap.String("role_id", roleID))
	}

	// Deactivate the group role assignment
	err = s.groupRoleRepo.DeactivateByGroupAndRole(ctx, groupID, roleID)
	if err != nil {
		s.logger.Error("Failed to remove role from group", zap.Error(err))

		// Log audit event for failed role removal
		auditDetails := map[string]interface{}{
			"group_name": group.Name,
			"error":      err.Error(),
		}
		if role != nil {
			auditDetails["role_name"] = role.Name
		}
		s.auditService.LogGroupRoleAssignment(ctx, "system", models.AuditActionRemoveGroupRole, group.OrganizationID, groupID, roleID, "Failed to remove role from group", false, auditDetails)

		return errors.NewInternalError(err)
	}

	// Log successful role removal
	auditDetails := map[string]interface{}{
		"group_name":      group.Name,
		"was_active":      groupRole.IsActive,
		"assignment_date": groupRole.CreatedAt,
	}
	if role != nil {
		auditDetails["role_name"] = role.Name
	}
	s.auditService.LogGroupRoleAssignment(ctx, "system", models.AuditActionRemoveGroupRole, group.OrganizationID, groupID, roleID, "Role removed from group successfully", true, auditDetails)

	// Invalidate relevant caches after successful role removal
	s.groupCache.InvalidateRoleAssignmentCache(ctx, group.OrganizationID, groupID, roleID)

	s.logger.Info("Role removed from group successfully",
		zap.String("group_id", groupID),
		zap.String("role_id", roleID))

	return nil
}

// GetGroupRoles retrieves all roles assigned to a group with organization-scoped validation
func (s *Service) GetGroupRoles(ctx context.Context, groupID string) (interface{}, error) {
	s.logger.Info("Retrieving group roles", zap.String("group_id", groupID))

	// Validate input parameters
	if groupID == "" {
		return nil, errors.NewValidationError("group_id cannot be empty")
	}

	// Check cache first
	if cached, found := s.groupCache.GetCachedGroupRoles(ctx, groupID, true); found {
		s.logger.Debug("Returning cached group roles", zap.String("group_id", groupID))
		return cached, nil
	}

	// Verify group exists
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		s.logger.Warn("Group not found", zap.String("group_id", groupID))
		return nil, errors.NewNotFoundError("group not found")
	}

	// Get all roles assigned to the group
	groupRoles, err := s.groupRoleRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		s.logger.Error("Failed to retrieve group roles", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format and load role details
	responses := make([]*groupResponses.GroupRoleDetail, len(groupRoles))
	for i, groupRole := range groupRoles {
		// Load role details
		role := &models.Role{}
		role, err = s.roleRepo.GetByID(ctx, groupRole.RoleID, role)
		if err != nil {
			s.logger.Warn("Failed to load role details",
				zap.String("role_id", groupRole.RoleID),
				zap.Error(err))
			// Continue with empty role details rather than failing
		}

		detail := groupResponses.NewGroupRoleDetail(groupRole)
		if role != nil {
			detail.Role = groupResponses.NewRoleDetail(role)
		}

		responses[i] = &detail
	}

	// Cache the results
	s.groupCache.CacheGroupRoles(ctx, groupID, responses, true)

	s.logger.Info("Group roles retrieved successfully",
		zap.String("group_id", groupID),
		zap.Int("count", len(responses)))

	return responses, nil
}

// GetGroupMembers retrieves all members of a group
func (s *Service) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) (interface{}, error) {
	s.logger.Info("Retrieving group members", zap.String("group_id", groupID))

	// For small result sets, check cache first (only if limit <= 100 to avoid caching large datasets)
	cacheKey := fmt.Sprintf("%s_%d_%d", groupID, limit, offset)
	if limit <= 100 {
		if cached, found := s.groupCache.GetCachedGroupMembers(ctx, cacheKey, true); found {
			s.logger.Debug("Returning cached group members", zap.String("group_id", groupID))
			return cached, nil
		}
	}

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

	// Cache the results for small result sets
	if limit <= 100 {
		s.groupCache.CacheGroupMembers(ctx, cacheKey, responses, true)
	}

	return responses, nil
}

// GetUserEffectiveRoles calculates and returns all effective roles for a user in an organization
// using the role inheritance engine with upward inheritance
func (s *Service) GetUserEffectiveRoles(ctx context.Context, orgID, userID string) (interface{}, error) {
	s.logger.Info("Getting user effective roles",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	// Validate input parameters
	if orgID == "" {
		return nil, errors.NewValidationError("org_id cannot be empty")
	}
	if userID == "" {
		return nil, errors.NewValidationError("user_id cannot be empty")
	}

	// Verify organization exists and is active
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		s.logger.Warn("Organization not found", zap.String("org_id", orgID))
		return nil, errors.NewNotFoundError("organization not found")
	}
	if !org.IsActive {
		s.logger.Warn("Organization is inactive", zap.String("org_id", orgID))
		return nil, errors.NewValidationError("cannot get roles for inactive organization")
	}

	// Check enhanced cache first
	if cached, found := s.groupCache.GetCachedUserEffectiveRoles(ctx, orgID, userID); found {
		s.logger.Debug("Returning cached user effective roles",
			zap.String("org_id", orgID),
			zap.String("user_id", userID))
		return cached, nil
	}

	// Create role inheritance engine with caching support
	inheritanceEngine := NewRoleInheritanceEngineWithRepos(
		s.groupRepo,
		s.groupRoleRepo,
		s.roleRepo,
		s.groupMembershipRepo,
		s.cache,
		s.logger,
	)

	// Calculate effective roles using upward inheritance
	effectiveRoles, err := inheritanceEngine.CalculateEffectiveRoles(ctx, orgID, userID)
	if err != nil {
		s.logger.Error("Failed to calculate effective roles", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Cache the enhanced effective roles
	s.groupCache.CacheUserEffectiveRoles(ctx, orgID, userID, effectiveRoles)

	s.logger.Info("User effective roles calculated successfully",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.Int("role_count", len(effectiveRoles)))

	return effectiveRoles, nil
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
