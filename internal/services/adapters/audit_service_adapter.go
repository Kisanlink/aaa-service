package adapters

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/services"
)

// AuditServiceAdapter adapts the concrete audit service to the interface
type AuditServiceAdapter struct {
	service *services.AuditService
}

// NewAuditServiceAdapter creates a new audit service adapter
func NewAuditServiceAdapter(service *services.AuditService) interfaces.AuditService {
	return &AuditServiceAdapter{service: service}
}

// LogUserAction implements AuditService.LogUserAction
func (a *AuditServiceAdapter) LogUserAction(ctx context.Context, userID, action, resource, resourceID string, details map[string]interface{}) {
	a.service.LogUserAction(ctx, userID, action, resource, resourceID, details)
}

// LogUserActionWithError implements AuditService.LogUserActionWithError
func (a *AuditServiceAdapter) LogUserActionWithError(ctx context.Context, userID, action, resource, resourceID string, err error, details map[string]interface{}) {
	a.service.LogUserActionWithError(ctx, userID, action, resource, resourceID, err, details)
}

// LogAPIAccess implements AuditService.LogAPIAccess
func (a *AuditServiceAdapter) LogAPIAccess(ctx context.Context, userID, method, endpoint, ipAddress, userAgent string, success bool, err error) {
	a.service.LogAPIAccess(ctx, userID, method, endpoint, ipAddress, userAgent, success, err)
}

// LogAccessDenied implements AuditService.LogAccessDenied
func (a *AuditServiceAdapter) LogAccessDenied(ctx context.Context, userID, action, resource, resourceID, reason string) {
	a.service.LogAccessDenied(ctx, userID, action, resource, resourceID, reason)
}

// LogPermissionChange implements AuditService.LogPermissionChange
func (a *AuditServiceAdapter) LogPermissionChange(ctx context.Context, userID, action, resource, resourceID, permission string, details map[string]interface{}) {
	a.service.LogPermissionChange(ctx, userID, action, resource, resourceID, permission, details)
}

// LogRoleChange implements AuditService.LogRoleChange
func (a *AuditServiceAdapter) LogRoleChange(ctx context.Context, userID, action, roleID string, details map[string]interface{}) {
	a.service.LogRoleChange(ctx, userID, action, roleID, details)
}

// LogDataAccess implements AuditService.LogDataAccess
func (a *AuditServiceAdapter) LogDataAccess(ctx context.Context, userID, action, resource, resourceID string, oldData, newData map[string]interface{}) {
	a.service.LogDataAccess(ctx, userID, action, resource, resourceID, oldData, newData)
}

// LogSecurityEvent implements AuditService.LogSecurityEvent
func (a *AuditServiceAdapter) LogSecurityEvent(ctx context.Context, userID, action, resource string, success bool, details map[string]interface{}) {
	a.service.LogSecurityEvent(ctx, userID, action, resource, success, details)
}

// LogAuthenticationAttempt implements AuditService.LogAuthenticationAttempt
func (a *AuditServiceAdapter) LogAuthenticationAttempt(ctx context.Context, userID, method, ipAddress, userAgent string, success bool, failureReason string) {
	a.service.LogAuthenticationAttempt(ctx, userID, method, ipAddress, userAgent, success, failureReason)
}

// LogRoleOperation implements AuditService.LogRoleOperation
func (a *AuditServiceAdapter) LogRoleOperation(ctx context.Context, actorUserID, targetUserID, roleID, operation string, success bool, details map[string]interface{}) {
	a.service.LogRoleOperation(ctx, actorUserID, targetUserID, roleID, operation, success, details)
}

// LogMPINOperation implements AuditService.LogMPINOperation
func (a *AuditServiceAdapter) LogMPINOperation(ctx context.Context, userID, operation, ipAddress, userAgent string, success bool, failureReason string) {
	a.service.LogMPINOperation(ctx, userID, operation, ipAddress, userAgent, success, failureReason)
}

// LogUserLifecycleEvent implements AuditService.LogUserLifecycleEvent
func (a *AuditServiceAdapter) LogUserLifecycleEvent(ctx context.Context, actorUserID, targetUserID, operation string, success bool, details map[string]interface{}) {
	a.service.LogUserLifecycleEvent(ctx, actorUserID, targetUserID, operation, success, details)
}

// LogSuspiciousActivity implements AuditService.LogSuspiciousActivity
func (a *AuditServiceAdapter) LogSuspiciousActivity(ctx context.Context, userID, activityType, description, ipAddress, userAgent string, details map[string]interface{}) {
	a.service.LogSuspiciousActivity(ctx, userID, activityType, description, ipAddress, userAgent, details)
}

// LogRateLimitViolation implements AuditService.LogRateLimitViolation
func (a *AuditServiceAdapter) LogRateLimitViolation(ctx context.Context, userID, endpoint, ipAddress, userAgent string, details map[string]interface{}) {
	a.service.LogRateLimitViolation(ctx, userID, endpoint, ipAddress, userAgent, details)
}

// LogSystemEvent implements AuditService.LogSystemEvent
func (a *AuditServiceAdapter) LogSystemEvent(ctx context.Context, action, resource string, success bool, details map[string]interface{}) {
	a.service.LogSystemEvent(ctx, action, resource, success, details)
}

// LogOrganizationOperation implements AuditService.LogOrganizationOperation
func (a *AuditServiceAdapter) LogOrganizationOperation(ctx context.Context, userID, action, orgID, message string, success bool, details map[string]interface{}) {
	a.service.LogOrganizationOperation(ctx, userID, action, orgID, message, success, details)
}

// LogGroupOperation implements AuditService.LogGroupOperation
func (a *AuditServiceAdapter) LogGroupOperation(ctx context.Context, userID, action, orgID, groupID, message string, success bool, details map[string]interface{}) {
	a.service.LogGroupOperation(ctx, userID, action, orgID, groupID, message, success, details)
}

// LogGroupMembershipChange implements AuditService.LogGroupMembershipChange
func (a *AuditServiceAdapter) LogGroupMembershipChange(ctx context.Context, actorUserID, action, orgID, groupID, targetUserID, message string, success bool, details map[string]interface{}) {
	a.service.LogGroupMembershipChange(ctx, actorUserID, action, orgID, groupID, targetUserID, message, success, details)
}

// LogGroupRoleAssignment implements AuditService.LogGroupRoleAssignment
func (a *AuditServiceAdapter) LogGroupRoleAssignment(ctx context.Context, actorUserID, action, orgID, groupID, roleID, message string, success bool, details map[string]interface{}) {
	a.service.LogGroupRoleAssignment(ctx, actorUserID, action, orgID, groupID, roleID, message, success, details)
}

// LogHierarchyChange implements AuditService.LogHierarchyChange
func (a *AuditServiceAdapter) LogHierarchyChange(ctx context.Context, userID, action, resourceType, resourceID, oldParentID, newParentID, message string, success bool, details map[string]interface{}) {
	a.service.LogHierarchyChange(ctx, userID, action, resourceType, resourceID, oldParentID, newParentID, message, success, details)
}

// LogOrganizationStructureChange implements AuditService.LogOrganizationStructureChange
func (a *AuditServiceAdapter) LogOrganizationStructureChange(ctx context.Context, userID, action, orgID, resourceType, resourceID string, oldValues, newValues map[string]interface{}, success bool, message string) {
	a.service.LogOrganizationStructureChange(ctx, userID, action, orgID, resourceType, resourceID, oldValues, newValues, success, message)
}

// QueryAuditLogs implements AuditService.QueryAuditLogs
func (a *AuditServiceAdapter) QueryAuditLogs(ctx context.Context, query interface{}) (interface{}, error) {
	// For now, return a placeholder since the concrete service expects a specific type
	// TODO: Implement proper query type handling
	return nil, &NotImplementedError{Method: "QueryAuditLogs"}
}

// QueryOrganizationAuditLogs implements AuditService.QueryOrganizationAuditLogs
func (a *AuditServiceAdapter) QueryOrganizationAuditLogs(ctx context.Context, orgID string, query interface{}) (interface{}, error) {
	// For now, return a placeholder since the concrete service expects a specific type
	// TODO: Implement proper query type handling
	return nil, &NotImplementedError{Method: "QueryOrganizationAuditLogs"}
}

// GetUserAuditTrail implements AuditService.GetUserAuditTrail
func (a *AuditServiceAdapter) GetUserAuditTrail(ctx context.Context, userID string, days int, page, perPage int) (interface{}, error) {
	result, err := a.service.GetUserAuditTrail(ctx, userID, days, page, perPage)
	return result, err
}

// GetResourceAuditTrail implements AuditService.GetResourceAuditTrail
func (a *AuditServiceAdapter) GetResourceAuditTrail(ctx context.Context, resource, resourceID string, days int, page, perPage int) (interface{}, error) {
	result, err := a.service.GetResourceAuditTrail(ctx, resource, resourceID, days, page, perPage)
	return result, err
}

// GetOrganizationAuditTrail implements AuditService.GetOrganizationAuditTrail
func (a *AuditServiceAdapter) GetOrganizationAuditTrail(ctx context.Context, orgID string, days int, page, perPage int) (interface{}, error) {
	result, err := a.service.GetOrganizationAuditTrail(ctx, orgID, days, page, perPage)
	return result, err
}

// GetGroupAuditTrail implements AuditService.GetGroupAuditTrail (adapter method)
func (a *AuditServiceAdapter) GetGroupAuditTrail(ctx context.Context, orgID, groupID string, days int, page, perPage int) (interface{}, error) {
	// The concrete service returns *services.AuditQueryResult, but interface expects interface{}
	result, err := a.service.GetGroupAuditTrail(ctx, orgID, groupID, days, page, perPage)
	return result, err
}

// GetSecurityEvents implements AuditService.GetSecurityEvents
func (a *AuditServiceAdapter) GetSecurityEvents(ctx context.Context, days int, page, perPage int) (interface{}, error) {
	result, err := a.service.GetSecurityEvents(ctx, days, page, perPage)
	return result, err
}

// ValidateAuditLogIntegrity implements AuditService.ValidateAuditLogIntegrity
func (a *AuditServiceAdapter) ValidateAuditLogIntegrity(ctx context.Context, auditLogID string) (bool, error) {
	return a.service.ValidateAuditLogIntegrity(ctx, auditLogID)
}

// GetAuditStatistics implements AuditService.GetAuditStatistics
func (a *AuditServiceAdapter) GetAuditStatistics(ctx context.Context, days int) (map[string]interface{}, error) {
	return a.service.GetAuditStatistics(ctx, days)
}

// ArchiveOldLogs implements AuditService.ArchiveOldLogs
func (a *AuditServiceAdapter) ArchiveOldLogs(ctx context.Context, days int) error {
	return a.service.ArchiveOldLogs(ctx, days)
}

// NotImplementedError represents a method that is not yet implemented
type NotImplementedError struct {
	Method string
}

func (e *NotImplementedError) Error() string {
	return "method not implemented: " + e.Method
}
