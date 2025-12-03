package organizations

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	groupRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/groups"
	orgRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/organizations"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// GetOrganizationGroups handles GET /organizations/:orgId/groups
//
//	@Summary		Get organization groups
//	@Description	Retrieve all groups within an organization with pagination
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId				path		string	true	"Organization ID"
//	@Param			limit				query		int		false	"Number of groups to return (default: 10, max: 100)"
//	@Param			offset				query		int		false	"Number of groups to skip (default: 0)"
//	@Param			include_inactive	query		bool	false	"Include inactive groups (default: false)"
//	@Success		200					{object}	organizations.OrganizationGroupListResponse
//	@Failure		400					{object}	responses.ErrorResponse
//	@Failure		404					{object}	responses.ErrorResponse
//	@Failure		500					{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups [get]
func (h *Handler) GetOrganizationGroups(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	includeInactiveStr := c.DefaultQuery("include_inactive", "false")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	includeInactive, err := strconv.ParseBool(includeInactiveStr)
	if err != nil {
		includeInactive = false
	}

	// Cap limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	// Verify organization exists
	_, err = h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Get groups for the organization
	groups, err := h.groupService.ListGroups(c.Request.Context(), limit, offset, orgID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to retrieve organization groups", zap.Error(err), zap.String("org_id", orgID))
		h.responder.SendInternalError(c, err)
		return
	}

	total, err := h.groupService.CountGroups(c.Request.Context(), orgID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to count organization groups", zap.Error(err), zap.String("org_id", orgID))
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendPaginatedResponse(c, groups, int(total), limit, offset)
}

// CreateGroupInOrganization handles POST /organizations/:orgId/groups
//
//	@Summary		Create group in organization
//	@Description	Create a new group within a specific organization
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Param			orgId	path		string											true	"Organization ID"
//	@Param			group	body		organizations.CreateOrganizationGroupRequest	true	"Group creation data"
//	@Success		201		{object}	organizations.OrganizationGroupResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		409		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups [post]
func (h *Handler) CreateGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	var req groupRequests.CreateGroupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for group creation", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Ensure the organization ID in the request matches the URL parameter
	req.OrganizationID = orgID

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Create group
	group, err := h.groupService.CreateGroup(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create group", zap.Error(err))

		switch {
		case errors.IsValidationError(err):
			h.responder.SendValidationError(c, []string{err.Error()})
		case errors.IsConflictError(err):
			h.responder.SendError(c, http.StatusConflict, "conflict", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("Group created successfully in organization",
		zap.String("org_id", orgID),
		zap.String("created_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusCreated, group)
}

// GetGroupInOrganization handles GET /organizations/:orgId/groups/:groupId
//
//	@Summary		Get group in organization
//	@Description	Retrieve a specific group within an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			groupId	path		string	true	"Group ID"
//	@Success		200		{object}	organizations.OrganizationGroupResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId} [get]
func (h *Handler) GetGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Get group
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Failed to retrieve group", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	// Note: This assumes the group response has an OrganizationID field
	// We need to check this by converting the interface{} response
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	h.responder.SendSuccess(c, http.StatusOK, group)
}

// UpdateGroupInOrganization handles PUT /organizations/:orgId/groups/:groupId
//
//	@Summary		Update group in organization
//	@Description	Update a specific group within an organization (super_admin only)
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Param			orgId	path		string										true	"Organization ID"
//	@Param			groupId	path		string										true	"Group ID"
//	@Param			request	body		organizations.UpdateOrganizationGroupRequest	true	"Group update data"
//	@Success		200		{object}	organizations.OrganizationGroupResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		401		{object}	responses.ErrorResponse
//	@Failure		403		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId} [put]
//	@Security		BearerAuth
func (h *Handler) UpdateGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	var req orgRequests.UpdateOrganizationGroupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for group update", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	existingGroup, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := existingGroup.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Convert to UpdateGroupRequest for the service layer
	updateGroupReq := &groupRequests.UpdateGroupRequest{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		IsActive:    req.IsActive,
	}

	// Update group
	updatedGroup, err := h.groupService.UpdateGroup(c.Request.Context(), groupID, updateGroupReq)
	if err != nil {
		h.logger.Error("Failed to update group",
			zap.Error(err),
			zap.String("group_id", groupID),
			zap.String("org_id", orgID))

		switch {
		case errors.IsValidationError(err):
			h.responder.SendValidationError(c, []string{err.Error()})
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		case errors.IsConflictError(err):
			h.responder.SendError(c, http.StatusConflict, "conflict", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("Group updated successfully in organization",
		zap.String("group_id", groupID),
		zap.String("org_id", orgID),
		zap.String("updated_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusOK, updatedGroup)
}

// DeleteGroupInOrganization handles DELETE /organizations/:orgId/groups/:groupId
//
//	@Summary		Delete group in organization
//	@Description	Delete a specific group within an organization (super_admin only)
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			groupId	path		string	true	"Group ID"
//	@Success		200		{object}	responses.SuccessResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		401		{object}	responses.ErrorResponse
//	@Failure		403		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId} [delete]
//	@Security		BearerAuth
func (h *Handler) DeleteGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	existingGroup, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := existingGroup.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Delete group
	err = h.groupService.DeleteGroup(c.Request.Context(), groupID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete group",
			zap.Error(err),
			zap.String("group_id", groupID),
			zap.String("org_id", orgID))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		case errors.IsConflictError(err):
			h.responder.SendError(c, http.StatusConflict, "cannot delete group with active members", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("Group deleted successfully from organization",
		zap.String("group_id", groupID),
		zap.String("org_id", orgID),
		zap.String("deleted_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusOK, "group deleted successfully")
}

// AddUserToGroupInOrganization handles POST /organizations/:orgId/groups/:groupId/users
//
//	@Summary		Add user to group in organization
//	@Description	Add a user to a specific group within an organization
//	@Tags			organizations
//	@Accept			json
//	@Produce		json
//	@Param			orgId	path		string									true	"Organization ID"
//	@Param			groupId	path		string									true	"Group ID"
//	@Param			request	body		organizations.AssignUserToGroupRequest	true	"User assignment data"
//	@Success		201		{object}	organizations.OrganizationGroupMemberResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		409		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId}/users [post]
func (h *Handler) AddUserToGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	var req orgRequests.AssignUserToGroupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for user-group assignment",
			zap.Error(err),
			zap.String("org_id", orgID),
			zap.String("group_id", groupID),
			zap.Any("request_body", req))

		// Try to extract detailed validation errors
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, fieldErr := range validationErrs {
				errorMessages = append(errorMessages, fmt.Sprintf("field '%s' failed validation '%s'", fieldErr.Field(), fieldErr.Tag()))
				h.logger.Error("Validation error detail",
					zap.String("field", fieldErr.Field()),
					zap.String("tag", fieldErr.Tag()),
					zap.Any("value", fieldErr.Value()),
					zap.String("param", fieldErr.Param()))
			}
			h.responder.SendValidationError(c, errorMessages)
			return
		}

		h.responder.SendError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Extract user ID from context (set by auth middleware)
	currentUserID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Convert to AddMemberRequest for the service layer
	addMemberReq := &groupRequests.AddMemberRequest{
		GroupID:       groupID,
		PrincipalID:   req.PrincipalID,
		PrincipalType: req.PrincipalType,
		AddedByID:     currentUserID.(string),
		StartsAt:      req.StartsAt,
		EndsAt:        req.EndsAt,
	}

	// Add member to group
	membership, err := h.groupService.AddMemberToGroup(c.Request.Context(), addMemberReq)
	if err != nil {
		h.logger.Error("Failed to add user to group", zap.Error(err))

		switch {
		case errors.IsValidationError(err):
			h.responder.SendValidationError(c, []string{err.Error()})
		case errors.IsConflictError(err):
			h.responder.SendError(c, http.StatusConflict, "user already in group", err)
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "user not found", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("User added to group successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("principal_id", req.PrincipalID),
		zap.String("added_by", currentUserID.(string)))

	h.responder.SendSuccess(c, http.StatusCreated, membership)
}

// RemoveUserFromGroupInOrganization handles DELETE /organizations/:orgId/groups/:groupId/users/:userId
//
//	@Summary		Remove user from group in organization
//	@Description	Remove a user from a specific group within an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			groupId	path		string	true	"Group ID"
//	@Param			userId	path		string	true	"User ID"
//	@Success		200		{object}	responses.SuccessResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId}/users/{userId} [delete]
func (h *Handler) RemoveUserFromGroupInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")
	principalID := c.Param("userId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	if principalID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "user ID is required", nil)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Verify organization exists
	_, err := h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Remove member from group
	err = h.groupService.RemoveMemberFromGroup(c.Request.Context(), groupID, principalID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to remove user from group", zap.Error(err))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "user not found in group", err)
		case errors.IsValidationError(err):
			h.responder.SendError(c, http.StatusBadRequest, "validation failed", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.logger.Info("User removed from group successfully",
		zap.String("org_id", orgID),
		zap.String("group_id", groupID),
		zap.String("principal_id", principalID),
		zap.String("removed_by", userID.(string)))

	h.responder.SendSuccess(c, http.StatusOK, "user removed from group successfully")
}

// GetGroupUsersInOrganization handles GET /organizations/:orgId/groups/:groupId/users
//
//	@Summary		Get group users in organization
//	@Description	Retrieve all users in a specific group within an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			groupId	path		string	true	"Group ID"
//	@Param			limit	query		int		false	"Number of users to return (default: 10, max: 100)"
//	@Param			offset	query		int		false	"Number of users to skip (default: 0)"
//	@Success		200		{object}	organizations.OrganizationGroupMembersResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/groups/{groupId}/users [get]
func (h *Handler) GetGroupUsersInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	groupID := c.Param("groupId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Cap limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	// Verify organization exists
	_, err = h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group exists and belongs to the organization
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Group not found", zap.Error(err), zap.String("group_id", groupID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "group not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Verify group belongs to the specified organization
	if groupResponse, ok := group.(map[string]interface{}); ok {
		if groupOrgID, exists := groupResponse["organization_id"]; exists {
			if groupOrgID != orgID {
				h.logger.Warn("Group does not belong to specified organization",
					zap.String("group_id", groupID),
					zap.String("group_org_id", fmt.Sprintf("%v", groupOrgID)),
					zap.String("requested_org_id", orgID))
				h.responder.SendError(c, http.StatusNotFound, "group not found in this organization", nil)
				return
			}
		}
	}

	// Get group members
	members, err := h.groupService.GetGroupMembers(c.Request.Context(), groupID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to retrieve group members", zap.Error(err), zap.String("group_id", groupID))
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, members)
}

// GetUserGroupsInOrganization handles GET /organizations/:orgId/users/:userId/groups
//
//	@Summary		Get user groups in organization
//	@Description	Retrieve all groups a user belongs to within an organization
//	@Tags			organizations
//	@Produce		json
//	@Param			orgId	path		string	true	"Organization ID"
//	@Param			userId	path		string	true	"User ID"
//	@Param			limit	query		int		false	"Number of groups to return (default: 10, max: 100)"
//	@Param			offset	query		int		false	"Number of groups to skip (default: 0)"
//	@Success		200		{object}	organizations.UserOrganizationGroupsResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/organizations/{orgId}/users/{userId}/groups [get]
func (h *Handler) GetUserGroupsInOrganization(c *gin.Context) {
	orgID := c.Param("id")
	principalID := c.Param("userId")

	if orgID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "organization ID is required", nil)
		return
	}

	if principalID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "user ID is required", nil)
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Cap limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	// Verify organization exists
	_, err = h.orgService.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Organization not found", zap.Error(err), zap.String("org_id", orgID))
		if errors.IsNotFoundError(err) {
			h.responder.SendError(c, http.StatusNotFound, "organization not found", err)
		} else {
			h.responder.SendInternalError(c, err)
		}
		return
	}

	// Get user's groups within the organization
	// Note: This would require extending the GroupService interface with a method like GetUserGroups
	// For now, we'll use a placeholder implementation that would need to be added to the service
	groups, err := h.getUserGroupsInOrganization(c.Request.Context(), orgID, principalID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to retrieve user groups", zap.Error(err),
			zap.String("org_id", orgID),
			zap.String("user_id", principalID))

		switch {
		case errors.IsNotFoundError(err):
			h.responder.SendError(c, http.StatusNotFound, "user not found", err)
		default:
			h.responder.SendInternalError(c, err)
		}
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, groups)
}

// getUserGroupsInOrganization is a helper method to get user's groups within an organization
func (h *Handler) getUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
	// Call organization service to get user's groups within the organization
	groups, err := h.orgService.GetUserGroupsInOrganization(ctx, orgID, userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user groups in organization",
			zap.String("org_id", orgID),
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	return groups, nil
}
