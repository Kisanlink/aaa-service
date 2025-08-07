package grpc_server

import (
	"context"

	"github.com/Kisanlink/aaa-service/services"
	"go.uber.org/zap"
)

// AuditHandler implements audit-related gRPC services
type AuditHandler struct {
	auditService *services.AuditService
	logger       *zap.Logger
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService, logger *zap.Logger) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		logger:       logger,
	}
}

// LogUserAction logs a user action
func (h *AuditHandler) LogUserAction(ctx context.Context, userID, action, resource, resourceID string, details map[string]interface{}) error {
	h.logger.Info("gRPC LogUserAction request",
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID))

	h.auditService.LogUserAction(ctx, userID, action, resource, resourceID, details)

	h.logger.Info("User action logged successfully",
		zap.String("user_id", userID),
		zap.String("action", action))

	return nil
}

// LogUserActionWithError logs a failed user action
func (h *AuditHandler) LogUserActionWithError(ctx context.Context, userID, action, resource, resourceID string, err error, details map[string]interface{}) error {
	h.logger.Info("gRPC LogUserActionWithError request",
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
		zap.Error(err))

	h.auditService.LogUserActionWithError(ctx, userID, action, resource, resourceID, err, details)

	h.logger.Info("User action error logged successfully",
		zap.String("user_id", userID),
		zap.String("action", action))

	return nil
}

// LogAPIAccess logs API access
func (h *AuditHandler) LogAPIAccess(ctx context.Context, userID, method, endpoint, ipAddress, userAgent string, success bool, err error) error {
	h.logger.Info("gRPC LogAPIAccess request",
		zap.String("user_id", userID),
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.Bool("success", success))

	h.auditService.LogAPIAccess(ctx, userID, method, endpoint, ipAddress, userAgent, success, err)

	h.logger.Info("API access logged successfully",
		zap.String("user_id", userID),
		zap.String("endpoint", endpoint))

	return nil
}

// LogAccessDenied logs access denied events
func (h *AuditHandler) LogAccessDenied(ctx context.Context, userID, action, resource, resourceID, reason string) error {
	h.logger.Info("gRPC LogAccessDenied request",
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
		zap.String("reason", reason))

	h.auditService.LogAccessDenied(ctx, userID, action, resource, resourceID, reason)

	h.logger.Info("Access denied logged successfully",
		zap.String("user_id", userID),
		zap.String("action", action))

	return nil
}

// LogPermissionChange logs permission changes
func (h *AuditHandler) LogPermissionChange(ctx context.Context, userID, action, resource, resourceID, permission string, details map[string]interface{}) error {
	h.logger.Info("gRPC LogPermissionChange request",
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
		zap.String("permission", permission))

	h.auditService.LogPermissionChange(ctx, userID, action, resource, resourceID, permission, details)

	h.logger.Info("Permission change logged successfully",
		zap.String("user_id", userID),
		zap.String("action", action))

	return nil
}

// LogRoleChange logs role changes
func (h *AuditHandler) LogRoleChange(ctx context.Context, userID, action, roleID string, details map[string]interface{}) error {
	h.logger.Info("gRPC LogRoleChange request",
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("role_id", roleID))

	h.auditService.LogRoleChange(ctx, userID, action, roleID, details)

	h.logger.Info("Role change logged successfully",
		zap.String("user_id", userID),
		zap.String("action", action))

	return nil
}

// QueryAuditLogs queries audit logs with filtering
func (h *AuditHandler) QueryAuditLogs(ctx context.Context, query *services.AuditQuery) (*services.AuditQueryResult, error) {
	h.logger.Info("gRPC QueryAuditLogs request",
		zap.String("user_id", query.UserID),
		zap.String("action", query.Action),
		zap.String("resource", query.Resource),
		zap.Int("page", query.Page),
		zap.Int("per_page", query.PerPage))

	result, err := h.auditService.QueryAuditLogs(ctx, query)
	if err != nil {
		h.logger.Error("Query audit logs failed", zap.Error(err))
		return nil, err
	}

	h.logger.Info("Audit logs queried successfully",
		zap.Int64("total_count", result.TotalCount),
		zap.Int("results_count", len(result.Logs)))

	return result, nil
}

// GetUserAuditTrail gets audit trail for a specific user
func (h *AuditHandler) GetUserAuditTrail(ctx context.Context, userID string, days, page, perPage int) (*services.AuditQueryResult, error) {
	h.logger.Info("gRPC GetUserAuditTrail request",
		zap.String("user_id", userID),
		zap.Int("days", days),
		zap.Int("page", page),
		zap.Int("per_page", perPage))

	result, err := h.auditService.GetUserAuditTrail(ctx, userID, days, page, perPage)
	if err != nil {
		h.logger.Error("Get user audit trail failed", zap.Error(err))
		return nil, err
	}

	h.logger.Info("User audit trail retrieved successfully",
		zap.String("user_id", userID),
		zap.Int64("total_count", result.TotalCount))

	return result, nil
}

// GetResourceAuditTrail gets audit trail for a specific resource
func (h *AuditHandler) GetResourceAuditTrail(ctx context.Context, resource, resourceID string, days, page, perPage int) (*services.AuditQueryResult, error) {
	h.logger.Info("gRPC GetResourceAuditTrail request",
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
		zap.Int("days", days),
		zap.Int("page", page),
		zap.Int("per_page", perPage))

	result, err := h.auditService.GetResourceAuditTrail(ctx, resource, resourceID, days, page, perPage)
	if err != nil {
		h.logger.Error("Get resource audit trail failed", zap.Error(err))
		return nil, err
	}

	h.logger.Info("Resource audit trail retrieved successfully",
		zap.String("resource", resource),
		zap.Int64("total_count", result.TotalCount))

	return result, nil
}

// GetSecurityEvents gets security-related audit events
func (h *AuditHandler) GetSecurityEvents(ctx context.Context, days, page, perPage int) (*services.AuditQueryResult, error) {
	h.logger.Info("gRPC GetSecurityEvents request",
		zap.Int("days", days),
		zap.Int("page", page),
		zap.Int("per_page", perPage))

	result, err := h.auditService.GetSecurityEvents(ctx, days, page, perPage)
	if err != nil {
		h.logger.Error("Get security events failed", zap.Error(err))
		return nil, err
	}

	h.logger.Info("Security events retrieved successfully",
		zap.Int64("total_count", result.TotalCount))

	return result, nil
}

// GetAuditStatistics gets audit statistics
func (h *AuditHandler) GetAuditStatistics(ctx context.Context, days int) (map[string]interface{}, error) {
	h.logger.Info("gRPC GetAuditStatistics request",
		zap.Int("days", days))

	stats, err := h.auditService.GetAuditStatistics(ctx, days)
	if err != nil {
		h.logger.Error("Get audit statistics failed", zap.Error(err))
		return nil, err
	}

	h.logger.Info("Audit statistics retrieved successfully")

	return stats, nil
}

// ArchiveOldLogs archives or deletes old audit logs
func (h *AuditHandler) ArchiveOldLogs(ctx context.Context, days int) error {
	h.logger.Info("gRPC ArchiveOldLogs request",
		zap.Int("days", days))

	err := h.auditService.ArchiveOldLogs(ctx, days)
	if err != nil {
		h.logger.Error("Archive old logs failed", zap.Error(err))
		return err
	}

	h.logger.Info("Old logs archived successfully")

	return nil
}
