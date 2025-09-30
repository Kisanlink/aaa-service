package groups

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"go.uber.org/zap"
)

// ServiceAdapter adapts the concrete group service to the interface
type ServiceAdapter struct {
	service *Service
	logger  *zap.Logger
}

// NewServiceAdapter creates a new group service adapter
func NewServiceAdapter(service *Service, logger *zap.Logger) interfaces.GroupService {
	return &ServiceAdapter{
		service: service,
		logger:  logger,
	}
}

// CreateGroup adapts the concrete method to the interface
func (a *ServiceAdapter) CreateGroup(ctx context.Context, req interface{}) (interface{}, error) {
	return a.service.CreateGroup(ctx, req)
}

// GetGroup adapts the concrete method to the interface
func (a *ServiceAdapter) GetGroup(ctx context.Context, groupID string) (interface{}, error) {
	return a.service.GetGroup(ctx, groupID)
}

// UpdateGroup adapts the concrete method to the interface
func (a *ServiceAdapter) UpdateGroup(ctx context.Context, groupID string, req interface{}) (interface{}, error) {
	return a.service.UpdateGroup(ctx, groupID, req)
}

// DeleteGroup adapts the concrete method to the interface
func (a *ServiceAdapter) DeleteGroup(ctx context.Context, groupID string, deletedBy string) error {
	return a.service.DeleteGroup(ctx, groupID, deletedBy)
}

// ListGroups adapts the concrete method to the interface
func (a *ServiceAdapter) ListGroups(ctx context.Context, limit, offset int, organizationID string, includeInactive bool) (interface{}, error) {
	return a.service.ListGroups(ctx, limit, offset, organizationID, includeInactive)
}

// AddMemberToGroup adapts the concrete method to the interface
func (a *ServiceAdapter) AddMemberToGroup(ctx context.Context, req interface{}) (interface{}, error) {
	return a.service.AddMemberToGroup(ctx, req)
}

// RemoveMemberFromGroup adapts the concrete method to the interface
func (a *ServiceAdapter) RemoveMemberFromGroup(ctx context.Context, groupID, principalID string, removedBy string) error {
	return a.service.RemoveMemberFromGroup(ctx, groupID, principalID, removedBy)
}

// GetGroupMembers adapts the concrete method to the interface
func (a *ServiceAdapter) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) (interface{}, error) {
	// This method needs to be implemented in the concrete service
	a.logger.Warn("GetGroupMembers not fully implemented in concrete service")
	return nil, &NotImplementedError{Method: "GetGroupMembers"}
}

// AssignRoleToGroup adapts the concrete method to the interface
func (a *ServiceAdapter) AssignRoleToGroup(ctx context.Context, groupID, roleID, assignedBy string) (interface{}, error) {
	return a.service.AssignRoleToGroup(ctx, groupID, roleID, assignedBy)
}

// RemoveRoleFromGroup adapts the concrete method to the interface
func (a *ServiceAdapter) RemoveRoleFromGroup(ctx context.Context, groupID, roleID string) error {
	return a.service.RemoveRoleFromGroup(ctx, groupID, roleID)
}

// GetGroupRoles adapts the concrete method to the interface
func (a *ServiceAdapter) GetGroupRoles(ctx context.Context, groupID string) (interface{}, error) {
	return a.service.GetGroupRoles(ctx, groupID)
}

// GetUserEffectiveRoles adapts the concrete method to the interface
func (a *ServiceAdapter) GetUserEffectiveRoles(ctx context.Context, orgID, userID string) (interface{}, error) {
	// This method needs to be implemented in the concrete service
	a.logger.Warn("GetUserEffectiveRoles not fully implemented in concrete service")
	return nil, &NotImplementedError{Method: "GetUserEffectiveRoles"}
}

// Custom error types for the adapter
type NotImplementedError struct {
	Method string
}

func (e *NotImplementedError) Error() string {
	return "method not implemented: " + e.Method
}
