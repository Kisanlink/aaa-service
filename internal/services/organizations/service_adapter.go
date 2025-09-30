package organizations

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/entities/requests/organizations"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"go.uber.org/zap"
)

// ServiceAdapter adapts the concrete organization service to the interface
type ServiceAdapter struct {
	service *Service
	logger  *zap.Logger
}

// NewServiceAdapter creates a new service adapter
func NewServiceAdapter(service *Service, logger *zap.Logger) interfaces.OrganizationService {
	return &ServiceAdapter{
		service: service,
		logger:  logger,
	}
}

// CreateOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) CreateOrganization(ctx context.Context, req interface{}) (interface{}, error) {
	createReq, ok := req.(*organizations.CreateOrganizationRequest)
	if !ok {
		a.logger.Error("Invalid request type for CreateOrganization")
		return nil, &InvalidRequestTypeError{Expected: "*organizations.CreateOrganizationRequest"}
	}
	return a.service.CreateOrganization(ctx, createReq)
}

// GetOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
	return a.service.GetOrganization(ctx, orgID)
}

// UpdateOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) UpdateOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
	updateReq, ok := req.(*organizations.UpdateOrganizationRequest)
	if !ok {
		a.logger.Error("Invalid request type for UpdateOrganization")
		return nil, &InvalidRequestTypeError{Expected: "*organizations.UpdateOrganizationRequest"}
	}
	return a.service.UpdateOrganization(ctx, orgID, updateReq)
}

// DeleteOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) DeleteOrganization(ctx context.Context, orgID string, deletedBy string) error {
	return a.service.DeleteOrganization(ctx, orgID, deletedBy)
}

// ListOrganizations adapts the concrete method to the interface
func (a *ServiceAdapter) ListOrganizations(ctx context.Context, limit, offset int, includeInactive bool) ([]interface{}, error) {
	orgs, err := a.service.ListOrganizations(ctx, limit, offset, includeInactive)
	if err != nil {
		return nil, err
	}

	// Convert to []interface{}
	result := make([]interface{}, len(orgs))
	for i, org := range orgs {
		result[i] = org
	}
	return result, nil
}

// GetOrganizationHierarchy adapts the concrete method to the interface
func (a *ServiceAdapter) GetOrganizationHierarchy(ctx context.Context, orgID string) (interface{}, error) {
	return a.service.GetOrganizationHierarchy(ctx, orgID)
}

// ActivateOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) ActivateOrganization(ctx context.Context, orgID string) error {
	return a.service.ActivateOrganization(ctx, orgID)
}

// DeactivateOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) DeactivateOrganization(ctx context.Context, orgID string) error {
	return a.service.DeactivateOrganization(ctx, orgID)
}

// GetOrganizationStats adapts the concrete method to the interface
func (a *ServiceAdapter) GetOrganizationStats(ctx context.Context, orgID string) (interface{}, error) {
	return a.service.GetOrganizationStats(ctx, orgID)
}

// GetOrganizationGroups adapts the concrete method to the interface
func (a *ServiceAdapter) GetOrganizationGroups(ctx context.Context, orgID string, limit, offset int, includeInactive bool) (interface{}, error) {
	return a.service.GetOrganizationGroups(ctx, orgID, limit, offset, includeInactive)
}

// CreateGroupInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) CreateGroupInOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
	return a.service.CreateGroupInOrganization(ctx, orgID, req)
}

// GetGroupInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) GetGroupInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error) {
	return a.service.GetGroupInOrganization(ctx, orgID, groupID)
}

// UpdateGroupInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) UpdateGroupInOrganization(ctx context.Context, orgID, groupID string, req interface{}) (interface{}, error) {
	return a.service.UpdateGroupInOrganization(ctx, orgID, groupID, req)
}

// DeleteGroupInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) DeleteGroupInOrganization(ctx context.Context, orgID, groupID string, deletedBy string) error {
	return a.service.DeleteGroupInOrganization(ctx, orgID, groupID, deletedBy)
}

// GetGroupHierarchyInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) GetGroupHierarchyInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error) {
	return a.service.GetGroupHierarchyInOrganization(ctx, orgID, groupID)
}

// AddUserToGroupInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) AddUserToGroupInOrganization(ctx context.Context, orgID, groupID, userID string, req interface{}) (interface{}, error) {
	// This method is not implemented in the concrete service yet
	a.logger.Warn("AddUserToGroupInOrganization not implemented in concrete service")
	return nil, &NotImplementedError{Method: "AddUserToGroupInOrganization"}
}

// RemoveUserFromGroupInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) RemoveUserFromGroupInOrganization(ctx context.Context, orgID, groupID, userID string, removedBy string) error {
	// This method is not implemented in the concrete service yet
	a.logger.Warn("RemoveUserFromGroupInOrganization not implemented in concrete service")
	return &NotImplementedError{Method: "RemoveUserFromGroupInOrganization"}
}

// GetGroupUsersInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) GetGroupUsersInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error) {
	// This method is not implemented in the concrete service yet
	a.logger.Warn("GetGroupUsersInOrganization not implemented in concrete service")
	return nil, &NotImplementedError{Method: "GetGroupUsersInOrganization"}
}

// GetUserGroupsInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
	// This method is not implemented in the concrete service yet
	a.logger.Warn("GetUserGroupsInOrganization not implemented in concrete service")
	return nil, &NotImplementedError{Method: "GetUserGroupsInOrganization"}
}

// AssignRoleToGroupInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) AssignRoleToGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, req interface{}) (interface{}, error) {
	// This method is not implemented in the concrete service yet
	a.logger.Warn("AssignRoleToGroupInOrganization not implemented in concrete service")
	return nil, &NotImplementedError{Method: "AssignRoleToGroupInOrganization"}
}

// RemoveRoleFromGroupInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) RemoveRoleFromGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, removedBy string) error {
	// This method is not implemented in the concrete service yet
	a.logger.Warn("RemoveRoleFromGroupInOrganization not implemented in concrete service")
	return &NotImplementedError{Method: "RemoveRoleFromGroupInOrganization"}
}

// GetGroupRolesInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) GetGroupRolesInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error) {
	// This method is not implemented in the concrete service yet
	a.logger.Warn("GetGroupRolesInOrganization not implemented in concrete service")
	return nil, &NotImplementedError{Method: "GetGroupRolesInOrganization"}
}

// GetUserEffectiveRolesInOrganization adapts the concrete method to the interface
func (a *ServiceAdapter) GetUserEffectiveRolesInOrganization(ctx context.Context, orgID, userID string) (interface{}, error) {
	// This method is not implemented in the concrete service yet
	a.logger.Warn("GetUserEffectiveRolesInOrganization not implemented in concrete service")
	return nil, &NotImplementedError{Method: "GetUserEffectiveRolesInOrganization"}
}

// Custom error types for the adapter
type InvalidRequestTypeError struct {
	Expected string
}

func (e *InvalidRequestTypeError) Error() string {
	return "invalid request type, expected: " + e.Expected
}

type NotImplementedError struct {
	Method string
}

func (e *NotImplementedError) Error() string {
	return "method not implemented: " + e.Method
}
