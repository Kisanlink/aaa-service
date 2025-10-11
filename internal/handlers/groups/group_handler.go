package groups

import (
	"net/http"
	"strconv"

	groupRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/groups"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for group operations
type Handler struct {
	groupService interfaces.GroupService
	logger       *zap.Logger
	responder    interfaces.Responder
}

// NewGroupHandler creates a new group handler instance
func NewGroupHandler(
	groupService interfaces.GroupService,
	logger *zap.Logger,
	responder interfaces.Responder,
) *Handler {
	return &Handler{
		groupService: groupService,
		logger:       logger,
		responder:    responder,
	}
}

// CreateGroup handles POST /groups
func (h *Handler) CreateGroup(c *gin.Context) {
	var req groupRequests.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind create group request", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request format", err)
		return
	}

	response, err := h.groupService.CreateGroup(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create group", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusCreated, response)
}

// GetGroup handles GET /groups/:id
func (h *Handler) GetGroup(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	response, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		h.logger.Error("Failed to get group", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// UpdateGroup handles PUT /groups/:id
func (h *Handler) UpdateGroup(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	var req groupRequests.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind update group request", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request format", err)
		return
	}

	response, err := h.groupService.UpdateGroup(c.Request.Context(), groupID, &req)
	if err != nil {
		h.logger.Error("Failed to update group", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// DeleteGroup handles DELETE /groups/:id
func (h *Handler) DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.responder.SendError(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	err := h.groupService.DeleteGroup(c.Request.Context(), groupID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete group", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, "group deleted successfully")
}

// ListGroups handles GET /groups
func (h *Handler) ListGroups(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	organizationID := c.Query("organization_id")
	includeInactiveStr := c.DefaultQuery("include_inactive", "false")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	includeInactive := includeInactiveStr == "true"

	response, err := h.groupService.ListGroups(c.Request.Context(), limit, offset, organizationID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to list groups", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// AddMemberToGroup handles POST /groups/:id/members
func (h *Handler) AddMemberToGroup(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID is required", nil)
		return
	}

	var req groupRequests.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind add member request", zap.Error(err))
		h.responder.SendError(c, http.StatusBadRequest, "invalid request format", err)
		return
	}

	// Set the group ID from the URL parameter
	req.GroupID = groupID

	response, err := h.groupService.AddMemberToGroup(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to add member to group", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusCreated, response)
}

// RemoveMemberFromGroup handles DELETE /groups/:id/members/:principal_id
func (h *Handler) RemoveMemberFromGroup(c *gin.Context) {
	groupID := c.Param("id")
	principalID := c.Param("principal_id")

	if groupID == "" || principalID == "" {
		h.responder.SendError(c, http.StatusBadRequest, "group ID and principal ID are required", nil)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.responder.SendError(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	err := h.groupService.RemoveMemberFromGroup(c.Request.Context(), groupID, principalID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to remove member from group", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, "member removed from group successfully")
}

// GetGroupMembers handles GET /groups/:id/members
func (h *Handler) GetGroupMembers(c *gin.Context) {
	groupID := c.Param("id")
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

	response, err := h.groupService.GetGroupMembers(c.Request.Context(), groupID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get group members", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// handleServiceError converts service errors to appropriate HTTP responses
func (h *Handler) handleServiceError(c *gin.Context, err error) {
	// This would need to be implemented based on your error types
	// For now, we'll return a generic internal server error
	h.responder.SendError(c, http.StatusInternalServerError, "internal server error", err)
}
