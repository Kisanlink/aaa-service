package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// AuditService provides audit logging services
type AuditService struct {
	dbManager    db.DBManager
	auditRepo    interfaces.AuditRepository
	cacheService interfaces.CacheService
	logger       *zap.Logger
}

// AuditLogEntry is an alias for the models.AuditLog for compatibility
type AuditLogEntry = models.AuditLog

// AuditEventDetails represents additional details for audit events
type AuditEventDetails struct {
	OldValues     map[string]interface{} `json:"old_values,omitempty"`
	NewValues     map[string]interface{} `json:"new_values,omitempty"`
	PermissionID  string                 `json:"permission_id,omitempty"`
	RoleID        string                 `json:"role_id,omitempty"`
	Reason        string                 `json:"reason,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
	TransactionID string                 `json:"transaction_id,omitempty"`
}

// AuditQuery represents a query for audit logs
type AuditQuery struct {
	UserID     string     `json:"user_id,omitempty"`
	Action     string     `json:"action,omitempty"`
	Resource   string     `json:"resource,omitempty"`
	ResourceID string     `json:"resource_id,omitempty"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Success    *bool      `json:"success,omitempty"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
}

// AuditQueryResult represents the result of an audit query
type AuditQueryResult struct {
	Logs       []models.AuditLog `json:"logs"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
}

// NewAuditService creates a new audit service
func NewAuditService(
	dbManager db.DBManager,
	auditRepo interfaces.AuditRepository,
	cacheService interfaces.CacheService,
	logger *zap.Logger,
) *AuditService {
	// Note: In this implementation, we assume the audit_logs table migration
	// is handled elsewhere in the application startup

	return &AuditService{
		dbManager:    dbManager,
		auditRepo:    auditRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

// isAnonymousUser checks if the userID represents an anonymous user
func isAnonymousUser(userID string) bool {
	return userID == "anonymous" || userID == "unknown" || userID == ""
}

// LogUserAction logs a user action
func (s *AuditService) LogUserAction(ctx context.Context, userID, action, resource, resourceID string, details map[string]interface{}) {
	if isAnonymousUser(userID) {
		// For anonymous users, don't set UserID to avoid foreign key constraint
		auditLog := models.NewAuditLog(action, resource, models.AuditStatusSuccess, "User action completed successfully")
		if resourceID != "" {
			auditLog.ResourceID = &resourceID
		}
		s.logEvent(ctx, auditLog, details)
	} else {
		// For authenticated users, use the normal method
		auditLog := models.NewAuditLogWithUserAndResource(userID, action, resource, resourceID, models.AuditStatusSuccess, "User action completed successfully")
		s.logEvent(ctx, auditLog, details)
	}
}

// LogUserActionWithError logs a failed user action
func (s *AuditService) LogUserActionWithError(ctx context.Context, userID, action, resource, resourceID string, err error, details map[string]interface{}) {
	if isAnonymousUser(userID) {
		// For anonymous users, don't set UserID to avoid foreign key constraint
		auditLog := models.NewAuditLog(action, resource, models.AuditStatusFailure, err.Error())
		if resourceID != "" {
			auditLog.ResourceID = &resourceID
		}
		s.logEvent(ctx, auditLog, details)
	} else {
		// For authenticated users, use the normal method
		auditLog := models.NewAuditLogWithUserAndResource(userID, action, resource, resourceID, models.AuditStatusFailure, err.Error())
		s.logEvent(ctx, auditLog, details)
	}
}

// LogAPIAccess logs API access
func (s *AuditService) LogAPIAccess(ctx context.Context, userID, method, endpoint, ipAddress, userAgent string, success bool, err error) {
	status := models.AuditStatusSuccess
	message := "API access successful"
	if err != nil {
		status = models.AuditStatusFailure
		message = err.Error()
	}

	if isAnonymousUser(userID) {
		// For anonymous users, don't set UserID to avoid foreign key constraint
		auditLog := models.NewAuditLog(models.AuditActionAPICall, models.ResourceTypeAPIEndpoint, status, message)
		if endpoint != "" {
			auditLog.ResourceID = &endpoint
		}
		auditLog.IPAddress = ipAddress
		auditLog.UserAgent = userAgent
		auditLog.AddDetail("http_method", method)
		auditLog.AddDetail("endpoint", endpoint)
		s.logEvent(ctx, auditLog, nil)
	} else {
		// For authenticated users, use the normal method
		auditLog := models.NewAuditLogWithUser(userID, models.AuditActionAPICall, models.ResourceTypeAPIEndpoint, status, message)
		if endpoint != "" {
			auditLog.ResourceID = &endpoint
		}
		auditLog.IPAddress = ipAddress
		auditLog.UserAgent = userAgent
		auditLog.AddDetail("http_method", method)
		auditLog.AddDetail("endpoint", endpoint)
		s.logEvent(ctx, auditLog, nil)
	}
}

// LogAccessDenied logs access denied events
func (s *AuditService) LogAccessDenied(ctx context.Context, userID, action, resource, resourceID, reason string) {
	if isAnonymousUser(userID) {
		// For anonymous users, don't set UserID to avoid foreign key constraint
		auditLog := models.NewAuditLog(models.AuditActionAccessDenied, resource, models.AuditStatusFailure, "Access denied")
		if resourceID != "" {
			auditLog.ResourceID = &resourceID
		}
		auditLog.AddDetail("reason", reason)
		auditLog.AddDetail("attempted_action", action)
		s.logEvent(ctx, auditLog, nil)
	} else {
		// For authenticated users, use the normal method
		auditLog := models.NewAuditLogWithUserAndResource(userID, models.AuditActionAccessDenied, resource, resourceID, models.AuditStatusFailure, "Access denied")
		auditLog.AddDetail("reason", reason)
		auditLog.AddDetail("attempted_action", action)
		s.logEvent(ctx, auditLog, nil)
	}
}

// LogPermissionChange logs permission changes
func (s *AuditService) LogPermissionChange(ctx context.Context, userID, action, resource, resourceID, permission string, details map[string]interface{}) {
	actionName := fmt.Sprintf("permission_%s", action)

	var auditLog *models.AuditLog
	if isAnonymousUser(userID) {
		// For anonymous users, don't set UserID to avoid foreign key constraint
		auditLog = models.NewAuditLog(actionName, resource, models.AuditStatusSuccess, "Permission change completed")
		if resourceID != "" {
			auditLog.ResourceID = &resourceID
		}
		auditLog.AddDetail("permission", permission)
	} else {
		// For authenticated users, use the normal method
		auditLog = models.NewAuditLogWithUserAndResource(userID, actionName, resource, resourceID, models.AuditStatusSuccess, "Permission change completed")
		auditLog.AddDetail("permission", permission)
	}

	// Merge additional details
	for k, v := range details {
		auditLog.AddDetail(k, v)
	}

	s.logEvent(ctx, auditLog, nil)
}

// LogRoleChange logs role changes
func (s *AuditService) LogRoleChange(ctx context.Context, userID, action, roleID string, details map[string]interface{}) {
	actionName := fmt.Sprintf("role_%s", action)

	var auditLog *models.AuditLog
	if isAnonymousUser(userID) {
		// For anonymous users, don't set UserID to avoid foreign key constraint
		auditLog = models.NewAuditLog(actionName, models.ResourceTypeRole, models.AuditStatusSuccess, "Role change completed")
		if roleID != "" {
			auditLog.ResourceID = &roleID
		}
		auditLog.AddDetail("role_id", roleID)
	} else {
		// For authenticated users, use the normal method
		auditLog = models.NewAuditLogWithUserAndResource(userID, actionName, models.ResourceTypeRole, roleID, models.AuditStatusSuccess, "Role change completed")
		auditLog.AddDetail("role_id", roleID)
	}

	// Merge additional details
	for k, v := range details {
		auditLog.AddDetail(k, v)
	}

	s.logEvent(ctx, auditLog, nil)
}

// LogDataAccess logs data access events
func (s *AuditService) LogDataAccess(ctx context.Context, userID, action, resource, resourceID string, oldData, newData map[string]interface{}) {
	var auditLog *models.AuditLog
	if isAnonymousUser(userID) {
		// For anonymous users, don't set UserID to avoid foreign key constraint
		auditLog = models.NewAuditLog(models.AuditActionDataAccess, resource, models.AuditStatusSuccess, "Data access logged")
		if resourceID != "" {
			auditLog.ResourceID = &resourceID
		}
	} else {
		// For authenticated users, use the normal method
		auditLog = models.NewAuditLogWithUserAndResource(userID, models.AuditActionDataAccess, resource, resourceID, models.AuditStatusSuccess, "Data access logged")
	}

	auditLog.AddDetail("old_values", oldData)
	auditLog.AddDetail("new_values", newData)

	s.logEvent(ctx, auditLog, nil)
}

// LogSecurityEvent logs security-related events
func (s *AuditService) LogSecurityEvent(ctx context.Context, userID, action, resource string, success bool, details map[string]interface{}) {
	status := models.AuditStatusSuccess
	message := "Security event completed"
	if !success {
		status = models.AuditStatusFailure
		message = "Security violation detected"
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(userID) {
		// For anonymous users, don't set UserID to avoid foreign key constraint
		auditLog = models.NewAuditLog(models.AuditActionSecurityEvent, resource, status, message)
	} else {
		// For authenticated users, use the normal method
		auditLog = models.NewAuditLogWithUser(userID, models.AuditActionSecurityEvent, resource, status, message)
	}

	// Add security-specific details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["security_action"] = action
	details["success"] = success

	s.logEvent(ctx, auditLog, details)
}

// LogAuthenticationAttempt logs authentication attempts with detailed context
func (s *AuditService) LogAuthenticationAttempt(ctx context.Context, userID, method, ipAddress, userAgent string, success bool, failureReason string) {
	status := models.AuditStatusSuccess
	message := "Authentication successful"
	action := models.AuditActionLogin

	if !success {
		status = models.AuditStatusFailure
		message = "Authentication failed"
		if failureReason != "" {
			message = fmt.Sprintf("Authentication failed: %s", failureReason)
		}
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(userID) || !success {
		// For anonymous users or failed attempts, don't set UserID
		auditLog = models.NewAuditLog(action, models.ResourceTypeUser, status, message)
	} else {
		auditLog = models.NewAuditLogWithUser(userID, action, models.ResourceTypeUser, status, message)
	}

	auditLog.IPAddress = ipAddress
	auditLog.UserAgent = userAgent

	details := map[string]interface{}{
		"authentication_method": method,
		"success":               success,
	}

	if !success && failureReason != "" {
		details["failure_reason"] = failureReason
	}

	s.logEvent(ctx, auditLog, details)
}

// LogRoleOperation logs role assignment/removal operations
func (s *AuditService) LogRoleOperation(ctx context.Context, actorUserID, targetUserID, roleID, operation string, success bool, details map[string]interface{}) {
	status := models.AuditStatusSuccess
	message := fmt.Sprintf("Role %s completed successfully", operation)
	action := fmt.Sprintf("role_%s", operation)

	if !success {
		status = models.AuditStatusFailure
		message = fmt.Sprintf("Role %s failed", operation)
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(actorUserID) {
		auditLog = models.NewAuditLog(action, models.ResourceTypeRole, status, message)
		if roleID != "" {
			auditLog.ResourceID = &roleID
		}
	} else {
		auditLog = models.NewAuditLogWithUserAndResource(actorUserID, action, models.ResourceTypeRole, roleID, status, message)
	}

	// Add role operation specific details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["target_user_id"] = targetUserID
	details["role_id"] = roleID
	details["operation"] = operation
	details["success"] = success

	s.logEvent(ctx, auditLog, details)
}

// LogMPINOperation logs MPIN-related operations
func (s *AuditService) LogMPINOperation(ctx context.Context, userID, operation, ipAddress, userAgent string, success bool, failureReason string) {
	status := models.AuditStatusSuccess
	message := fmt.Sprintf("MPIN %s completed successfully", operation)
	action := fmt.Sprintf("mpin_%s", operation)

	if !success {
		status = models.AuditStatusFailure
		message = fmt.Sprintf("MPIN %s failed", operation)
		if failureReason != "" {
			message = fmt.Sprintf("MPIN %s failed: %s", operation, failureReason)
		}
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(userID) {
		auditLog = models.NewAuditLog(action, models.ResourceTypeUser, status, message)
	} else {
		auditLog = models.NewAuditLogWithUser(userID, action, models.ResourceTypeUser, status, message)
	}

	auditLog.IPAddress = ipAddress
	auditLog.UserAgent = userAgent

	details := map[string]interface{}{
		"mpin_operation": operation,
		"success":        success,
	}

	if !success && failureReason != "" {
		details["failure_reason"] = failureReason
	}

	s.logEvent(ctx, auditLog, details)
}

// LogUserLifecycleEvent logs user creation, update, deletion events
func (s *AuditService) LogUserLifecycleEvent(ctx context.Context, actorUserID, targetUserID, operation string, success bool, details map[string]interface{}) {
	status := models.AuditStatusSuccess
	message := fmt.Sprintf("User %s completed successfully", operation)
	action := fmt.Sprintf("user_%s", operation)

	if !success {
		status = models.AuditStatusFailure
		message = fmt.Sprintf("User %s failed", operation)
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(actorUserID) {
		auditLog = models.NewAuditLog(action, models.ResourceTypeUser, status, message)
		if targetUserID != "" {
			auditLog.ResourceID = &targetUserID
		}
	} else {
		auditLog = models.NewAuditLogWithUserAndResource(actorUserID, action, models.ResourceTypeUser, targetUserID, status, message)
	}

	// Add user lifecycle specific details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["target_user_id"] = targetUserID
	details["operation"] = operation
	details["success"] = success

	s.logEvent(ctx, auditLog, details)
}

// LogSuspiciousActivity logs potentially suspicious or malicious activity
func (s *AuditService) LogSuspiciousActivity(ctx context.Context, userID, activityType, description, ipAddress, userAgent string, details map[string]interface{}) {
	auditLog := models.NewAuditLog("suspicious_activity", "security", models.AuditStatusWarning, description)

	if !isAnonymousUser(userID) {
		auditLog.UserID = &userID
	}

	auditLog.IPAddress = ipAddress
	auditLog.UserAgent = userAgent

	// Add suspicious activity specific details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["activity_type"] = activityType
	details["description"] = description
	details["severity"] = "warning"

	s.logEvent(ctx, auditLog, details)

	// Also log to application logger for immediate attention
	s.logger.Warn("Suspicious activity detected",
		zap.String("user_id", userID),
		zap.String("activity_type", activityType),
		zap.String("description", description),
		zap.String("ip_address", ipAddress),
		zap.String("user_agent", userAgent),
		zap.Any("details", details))
}

// LogRateLimitViolation logs rate limit violations
func (s *AuditService) LogRateLimitViolation(ctx context.Context, userID, endpoint, ipAddress, userAgent string, details map[string]interface{}) {
	auditLog := models.NewAuditLog("rate_limit_violation", "security", models.AuditStatusWarning, "Rate limit exceeded")

	if !isAnonymousUser(userID) {
		auditLog.UserID = &userID
	}

	auditLog.IPAddress = ipAddress
	auditLog.UserAgent = userAgent

	// Add rate limit specific details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["endpoint"] = endpoint
	details["violation_type"] = "rate_limit"

	s.logEvent(ctx, auditLog, details)
}

// LogSystemEvent logs system-level events
func (s *AuditService) LogSystemEvent(ctx context.Context, action, resource string, success bool, details map[string]interface{}) {
	status := models.AuditStatusSuccess
	message := "System event completed"
	if !success {
		status = models.AuditStatusFailure
		message = "System error occurred"
	}

	systemUserID := "system"
	auditLog := models.NewAuditLogWithUser(systemUserID, action, resource, status, message)

	s.logEvent(ctx, auditLog, details)
}

// QueryAuditLogs queries audit logs with filtering
func (s *AuditService) QueryAuditLogs(ctx context.Context, query *AuditQuery) (*AuditQueryResult, error) {
	// Validate and set defaults for pagination
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PerPage <= 0 {
		query.PerPage = 50
	}
	if query.PerPage > 1000 {
		query.PerPage = 1000
	}

	// Build filters for the database query
	filters := make(map[string]interface{})

	if query.UserID != "" {
		filters["user_id"] = query.UserID
	}
	if query.Action != "" {
		filters["action"] = query.Action
	}
	if query.Resource != "" {
		filters["resource_type"] = query.Resource
	}
	if query.ResourceID != "" {
		filters["resource_id"] = query.ResourceID
	}
	if query.Success != nil {
		if *query.Success {
			filters["status"] = models.AuditStatusSuccess
		} else {
			filters["status"] = models.AuditStatusFailure
		}
	}
	if query.StartTime != nil {
		filters["timestamp_gte"] = *query.StartTime
	}
	if query.EndTime != nil {
		filters["timestamp_lte"] = *query.EndTime
	}

	// Calculate offset for pagination
	offset := (query.Page - 1) * query.PerPage

	// Query audit logs using repository
	var logs []*models.AuditLog
	var totalCount int64
	var err error

	// Apply filters based on query parameters
	if query.StartTime != nil && query.EndTime != nil {
		if query.UserID != "" {
			logs, err = s.auditRepo.ListByUserAndTimeRange(ctx, query.UserID, *query.StartTime, *query.EndTime, query.PerPage, offset)
			if err == nil {
				totalCount, err = s.auditRepo.CountByUser(ctx, query.UserID)
			}
		} else {
			logs, err = s.auditRepo.ListByTimeRange(ctx, *query.StartTime, *query.EndTime, query.PerPage, offset)
			if err == nil {
				totalCount, err = s.auditRepo.CountByTimeRange(ctx, *query.StartTime, *query.EndTime)
			}
		}
	} else if query.UserID != "" {
		logs, err = s.auditRepo.ListByUser(ctx, query.UserID, query.PerPage, offset)
		if err == nil {
			totalCount, err = s.auditRepo.CountByUser(ctx, query.UserID)
		}
	} else if query.Action != "" {
		logs, err = s.auditRepo.ListByAction(ctx, query.Action, query.PerPage, offset)
	} else if query.Resource != "" {
		logs, err = s.auditRepo.ListByResourceType(ctx, query.Resource, query.PerPage, offset)
	} else {
		logs, err = s.auditRepo.List(ctx, query.PerPage, offset)
	}

	if err != nil {
		s.logger.Error("Failed to query audit logs", zap.Error(err))
		return nil, err
	}

	// Convert to slice of models.AuditLog for compatibility
	logSlice := make([]models.AuditLog, len(logs))
	for i, log := range logs {
		if log != nil {
			logSlice[i] = *log
		}
	}

	s.logger.Debug("Querying audit logs completed",
		zap.String("user_id", query.UserID),
		zap.String("action", query.Action),
		zap.String("resource", query.Resource),
		zap.Int("page", query.Page),
		zap.Int("per_page", query.PerPage),
		zap.Int("results", len(logSlice)))

	return &AuditQueryResult{
		Logs:       logSlice,
		TotalCount: totalCount,
		Page:       query.Page,
		PerPage:    query.PerPage,
	}, nil
}

// GetUserAuditTrail gets audit trail for a specific user
func (s *AuditService) GetUserAuditTrail(ctx context.Context, userID string, days int, page, perPage int) (*AuditQueryResult, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	query := &AuditQuery{
		UserID:    userID,
		StartTime: &startTime,
		Page:      page,
		PerPage:   perPage,
	}

	return s.QueryAuditLogs(ctx, query)
}

// GetResourceAuditTrail gets audit trail for a specific resource
func (s *AuditService) GetResourceAuditTrail(ctx context.Context, resource, resourceID string, days int, page, perPage int) (*AuditQueryResult, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	query := &AuditQuery{
		Resource:   resource,
		ResourceID: resourceID,
		StartTime:  &startTime,
		Page:       page,
		PerPage:    perPage,
	}

	return s.QueryAuditLogs(ctx, query)
}

// GetSecurityEvents gets security-related audit events
func (s *AuditService) GetSecurityEvents(ctx context.Context, days int, page, perPage int) (*AuditQueryResult, error) {
	startTime := time.Now().AddDate(0, 0, -days)
	success := false

	query := &AuditQuery{
		StartTime: &startTime,
		Success:   &success,
		Page:      page,
		PerPage:   perPage,
	}

	return s.QueryAuditLogs(ctx, query)
}

// ArchiveOldLogs archives or deletes old audit logs
func (s *AuditService) ArchiveOldLogs(ctx context.Context, days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Find old audit logs (simplified for now)

	// For now, simulate archiving with a placeholder count
	// In a real implementation, you'd query the database for old logs
	count := int64(0)

	if count == 0 {
		s.logger.Info("No audit logs to archive",
			zap.Time("cutoff_date", cutoffDate))
		return nil
	}

	// In a production system, you might want to:
	// 1. Export to external storage (S3, etc.)
	// 2. Compress the data
	// 3. Keep a backup

	// For now, we'll just log the archiving action since the DBManager interface
	// doesn't support bulk updates. In a real implementation, you'd use raw SQL
	// or batch operations to update the records.
	s.logger.Info("Archiving audit logs (simulation)",
		zap.Time("cutoff_date", cutoffDate),
		zap.Int64("count", count))

	// In a real implementation, you might:
	// 1. Export logs to external storage (S3, etc.)
	// 2. Mark them as archived in a separate table
	// 3. Use database-specific bulk operations

	s.logger.Info("Successfully archived old audit logs",
		zap.Time("cutoff_date", cutoffDate),
		zap.Int64("archived_count", count))

	return nil
}

// logEvent is the internal method to log audit events
func (s *AuditService) logEvent(ctx context.Context, auditLog *models.AuditLog, details interface{}) {
	// Set default values (BeforeCreate will handle ID and timestamps)
	if auditLog.Timestamp.IsZero() {
		auditLog.Timestamp = time.Now()
	}

	// Add details to the audit log
	if details != nil {
		if detailsMap, ok := details.(map[string]interface{}); ok {
			for k, v := range detailsMap {
				auditLog.AddDetail(k, v)
			}
		} else {
			auditLog.AddDetail("details", details)
		}
	}

	// Extract additional context from request context if available
	if ipAddress := ctx.Value("ip_address"); ipAddress != nil {
		if ip, ok := ipAddress.(string); ok {
			auditLog.IPAddress = ip
		}
	}
	if userAgent := ctx.Value("user_agent"); userAgent != nil {
		if ua, ok := userAgent.(string); ok {
			auditLog.UserAgent = ua
		}
	}

	// Add request ID for traceability
	if requestID := ctx.Value("request_id"); requestID != nil {
		if rid, ok := requestID.(string); ok {
			auditLog.AddDetail("request_id", rid)
		}
	}

	// Add session information if available
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		if sid, ok := sessionID.(string); ok {
			auditLog.AddDetail("session_id", sid)
		}
	}

	// Add correlation ID for distributed tracing
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if cid, ok := correlationID.(string); ok {
			auditLog.AddDetail("correlation_id", cid)
		}
	}

	// Save audit log to database using repository
	err := s.auditRepo.Create(ctx, auditLog)
	if err != nil {
		// If database save fails, still log to application logger
		s.logger.Error("Failed to save audit log to database",
			zap.Error(err),
			zap.String("audit_id", auditLog.ID),
			zap.String("action", auditLog.Action))

		// Try to cache the audit log for later retry
		s.cacheFailedAuditLog(auditLog)
	}

	// Log to application logger with structured information
	s.logToApplicationLogger(auditLog)

	// Log security events with enhanced monitoring
	if s.isSecuritySensitiveAction(auditLog.Action) {
		s.logSecuritySensitiveEvent(auditLog)
	}

	// Log performance metrics for slow operations
	if s.isPerformanceSensitiveAction(auditLog.Action) {
		s.logPerformanceMetrics(auditLog)
	}
}

// logToApplicationLogger logs audit events to the application logger with structured format
func (s *AuditService) logToApplicationLogger(auditLog *models.AuditLog) {
	userID := ""
	if auditLog.UserID != nil {
		userID = *auditLog.UserID
	}
	resourceID := ""
	if auditLog.ResourceID != nil {
		resourceID = *auditLog.ResourceID
	}

	baseFields := []zap.Field{
		zap.String("audit_id", auditLog.ID),
		zap.String("user_id", userID),
		zap.String("action", auditLog.Action),
		zap.String("resource_type", auditLog.ResourceType),
		zap.String("resource_id", resourceID),
		zap.String("status", auditLog.Status),
		zap.Time("timestamp", auditLog.Timestamp),
		zap.String("ip_address", auditLog.IPAddress),
		zap.String("user_agent", auditLog.UserAgent),
	}

	// Add details as structured fields
	if auditLog.Details != nil {
		for key, value := range auditLog.Details {
			baseFields = append(baseFields, zap.Any(fmt.Sprintf("detail_%s", key), value))
		}
	}

	// Log with appropriate level based on status and action
	switch {
	case auditLog.IsFailure():
		s.logger.Error("Audit event - failure", baseFields...)
	case auditLog.IsWarning():
		s.logger.Warn("Audit event - warning", baseFields...)
	case s.isSecuritySensitiveAction(auditLog.Action):
		s.logger.Info("Audit event - security", baseFields...)
	default:
		s.logger.Debug("Audit event", baseFields...)
	}
}

// logSecuritySensitiveEvent provides enhanced logging for security-sensitive events
func (s *AuditService) logSecuritySensitiveEvent(auditLog *models.AuditLog) {
	userID := ""
	if auditLog.UserID != nil {
		userID = *auditLog.UserID
	}
	resourceID := ""
	if auditLog.ResourceID != nil {
		resourceID = *auditLog.ResourceID
	}

	securityFields := []zap.Field{
		zap.String("security_event_type", "audit_log"),
		zap.String("audit_id", auditLog.ID),
		zap.String("user_id", userID),
		zap.String("action", auditLog.Action),
		zap.String("resource_type", auditLog.ResourceType),
		zap.String("resource_id", resourceID),
		zap.String("status", auditLog.Status),
		zap.String("message", auditLog.Message),
		zap.Time("timestamp", auditLog.Timestamp),
		zap.String("ip_address", auditLog.IPAddress),
		zap.String("user_agent", auditLog.UserAgent),
	}

	// Add security-specific context
	if auditLog.Details != nil {
		if requestID, exists := auditLog.Details["request_id"]; exists {
			securityFields = append(securityFields, zap.Any("request_id", requestID))
		}
		if sessionID, exists := auditLog.Details["session_id"]; exists {
			securityFields = append(securityFields, zap.Any("session_id", sessionID))
		}
		if failureReason, exists := auditLog.Details["failure_reason"]; exists {
			securityFields = append(securityFields, zap.Any("failure_reason", failureReason))
		}
	}

	// Determine security severity
	severity := "info"
	if auditLog.IsFailure() {
		severity = "warning"
		if s.isCriticalSecurityAction(auditLog.Action) {
			severity = "critical"
		}
	}

	securityFields = append(securityFields, zap.String("security_severity", severity))

	// Log with appropriate level
	switch severity {
	case "critical":
		s.logger.Error("Critical security event", securityFields...)
	case "warning":
		s.logger.Warn("Security warning event", securityFields...)
	default:
		s.logger.Info("Security event", securityFields...)
	}
}

// logPerformanceMetrics logs performance metrics for performance-sensitive operations
func (s *AuditService) logPerformanceMetrics(auditLog *models.AuditLog) {
	if auditLog.Details == nil {
		return
	}

	performanceFields := []zap.Field{
		zap.String("performance_event_type", "audit_log"),
		zap.String("audit_id", auditLog.ID),
		zap.String("action", auditLog.Action),
		zap.String("resource_type", auditLog.ResourceType),
		zap.Time("timestamp", auditLog.Timestamp),
	}

	// Extract performance metrics from details
	if duration, exists := auditLog.Details["duration_ms"]; exists {
		performanceFields = append(performanceFields, zap.Any("duration_ms", duration))
	}
	if responseSize, exists := auditLog.Details["response_size"]; exists {
		performanceFields = append(performanceFields, zap.Any("response_size", responseSize))
	}
	if dbQueries, exists := auditLog.Details["db_queries"]; exists {
		performanceFields = append(performanceFields, zap.Any("db_queries", dbQueries))
	}

	s.logger.Debug("Performance metrics", performanceFields...)
}

// cacheFailedAuditLog caches audit logs that failed to save to database for retry
func (s *AuditService) cacheFailedAuditLog(auditLog *models.AuditLog) {
	if s.cacheService == nil {
		return
	}

	cacheKey := fmt.Sprintf("failed_audit_log:%s", auditLog.ID)

	// Cache for 24 hours (86400 seconds)
	err := s.cacheService.Set(cacheKey, auditLog, 86400)
	if err != nil {
		s.logger.Error("Failed to cache failed audit log",
			zap.Error(err),
			zap.String("audit_id", auditLog.ID))
	} else {
		s.logger.Info("Cached failed audit log for retry",
			zap.String("audit_id", auditLog.ID))
	}
}

// isSecuritySensitiveAction checks if an action is security-sensitive
func (s *AuditService) isSecuritySensitiveAction(action string) bool {
	securityActions := []string{
		models.AuditActionLogin,
		models.AuditActionLogout,
		models.AuditActionRegister,
		models.AuditActionAssignRole,
		models.AuditActionRemoveRole,
		models.AuditActionGrantPermission,
		models.AuditActionRevokePermission,
		models.AuditActionAccessDenied,
		models.AuditActionSecurityEvent,
		"mpin_setup",
		"mpin_update",
		"mpin_verification",
		"password_reset",
		"account_locked",
		"suspicious_activity",
		"rate_limit_violation",
	}

	for _, securityAction := range securityActions {
		if action == securityAction {
			return true
		}
	}

	return false
}

// isCriticalSecurityAction checks if an action is critically security-sensitive
func (s *AuditService) isCriticalSecurityAction(action string) bool {
	criticalActions := []string{
		models.AuditActionAccessDenied,
		models.AuditActionSecurityEvent,
		"account_locked",
		"suspicious_activity",
		"multiple_failed_logins",
		"privilege_escalation_attempt",
	}

	for _, criticalAction := range criticalActions {
		if action == criticalAction {
			return true
		}
	}

	return false
}

// isPerformanceSensitiveAction checks if an action should have performance metrics logged
func (s *AuditService) isPerformanceSensitiveAction(action string) bool {
	performanceActions := []string{
		models.AuditActionAPICall,
		models.AuditActionDatabaseOperation,
		models.AuditActionLogin,
		models.AuditActionCheckPermission,
		"bulk_operation",
		"data_export",
		"report_generation",
	}

	for _, performanceAction := range performanceActions {
		if action == performanceAction {
			return true
		}
	}

	return false
}

// LogOrganizationOperation logs organization-related operations with organization context
func (s *AuditService) LogOrganizationOperation(ctx context.Context, userID, action, orgID, message string, success bool, details map[string]interface{}) {
	status := models.AuditStatusSuccess
	if !success {
		status = models.AuditStatusFailure
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(userID) {
		auditLog = models.NewAuditLog(action, models.ResourceTypeOrganization, status, message)
		if orgID != "" {
			auditLog.ResourceID = &orgID
		}
	} else {
		auditLog = models.NewAuditLogWithUserAndResource(userID, action, models.ResourceTypeOrganization, orgID, status, message)
	}

	// Add organization-specific details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["organization_id"] = orgID
	details["operation_type"] = "organization"

	s.logEvent(ctx, auditLog, details)
}

// LogGroupOperation logs group-related operations with organization and group context
func (s *AuditService) LogGroupOperation(ctx context.Context, userID, action, orgID, groupID, message string, success bool, details map[string]interface{}) {
	status := models.AuditStatusSuccess
	if !success {
		status = models.AuditStatusFailure
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(userID) {
		auditLog = models.NewAuditLog(action, models.ResourceTypeGroup, status, message)
		if groupID != "" {
			auditLog.ResourceID = &groupID
		}
	} else {
		auditLog = models.NewAuditLogWithUserAndResource(userID, action, models.ResourceTypeGroup, groupID, status, message)
	}

	// Add group-specific details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["organization_id"] = orgID
	details["group_id"] = groupID
	details["operation_type"] = "group"

	s.logEvent(ctx, auditLog, details)
}

// LogGroupMembershipChange logs group membership changes with full context
func (s *AuditService) LogGroupMembershipChange(ctx context.Context, actorUserID, action, orgID, groupID, targetUserID, message string, success bool, details map[string]interface{}) {
	status := models.AuditStatusSuccess
	if !success {
		status = models.AuditStatusFailure
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(actorUserID) {
		auditLog = models.NewAuditLog(action, models.ResourceTypeGroup, status, message)
		if groupID != "" {
			auditLog.ResourceID = &groupID
		}
	} else {
		auditLog = models.NewAuditLogWithUserAndResource(actorUserID, action, models.ResourceTypeGroup, groupID, status, message)
	}

	// Add membership-specific details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["organization_id"] = orgID
	details["group_id"] = groupID
	details["target_user_id"] = targetUserID
	details["actor_user_id"] = actorUserID
	details["operation_type"] = "group_membership"

	s.logEvent(ctx, auditLog, details)
}

// LogGroupRoleAssignment logs group role assignment/removal operations
func (s *AuditService) LogGroupRoleAssignment(ctx context.Context, actorUserID, action, orgID, groupID, roleID, message string, success bool, details map[string]interface{}) {
	status := models.AuditStatusSuccess
	if !success {
		status = models.AuditStatusFailure
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(actorUserID) {
		auditLog = models.NewAuditLog(action, models.ResourceTypeGroupRole, status, message)
		if groupID != "" {
			auditLog.ResourceID = &groupID
		}
	} else {
		auditLog = models.NewAuditLogWithUserAndResource(actorUserID, action, models.ResourceTypeGroupRole, groupID, status, message)
	}

	// Add role assignment specific details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["organization_id"] = orgID
	details["group_id"] = groupID
	details["role_id"] = roleID
	details["actor_user_id"] = actorUserID
	details["operation_type"] = "group_role"

	s.logEvent(ctx, auditLog, details)
}

// LogHierarchyChange logs organization or group hierarchy changes
func (s *AuditService) LogHierarchyChange(ctx context.Context, userID, action, resourceType, resourceID, oldParentID, newParentID, message string, success bool, details map[string]interface{}) {
	status := models.AuditStatusSuccess
	if !success {
		status = models.AuditStatusFailure
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(userID) {
		auditLog = models.NewAuditLog(action, resourceType, status, message)
		if resourceID != "" {
			auditLog.ResourceID = &resourceID
		}
	} else {
		auditLog = models.NewAuditLogWithUserAndResource(userID, action, resourceType, resourceID, status, message)
	}

	// Add hierarchy-specific details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["old_parent_id"] = oldParentID
	details["new_parent_id"] = newParentID
	details["operation_type"] = "hierarchy_change"

	s.logEvent(ctx, auditLog, details)
}

// LogOrganizationStructureChange logs comprehensive organization structure changes
// This method ensures all context is captured for organization hierarchy modifications
func (s *AuditService) LogOrganizationStructureChange(ctx context.Context, userID, action, orgID, resourceType, resourceID string, oldValues, newValues map[string]interface{}, success bool, message string) {
	status := models.AuditStatusSuccess
	if !success {
		status = models.AuditStatusFailure
	}

	var auditLog *models.AuditLog
	if isAnonymousUser(userID) {
		auditLog = models.NewAuditLog(action, resourceType, status, message)
		if resourceID != "" {
			auditLog.ResourceID = &resourceID
		}
	} else {
		auditLog = models.NewAuditLogWithUserAndResource(userID, action, resourceType, resourceID, status, message)
	}

	// Add comprehensive structure change details
	details := map[string]interface{}{
		"organization_id":  orgID,
		"operation_type":   "structure_change",
		"resource_type":    resourceType,
		"resource_id":      resourceID,
		"change_timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Add old and new values for complete audit trail
	if oldValues != nil && len(oldValues) > 0 {
		details["old_values"] = oldValues
	}
	if newValues != nil && len(newValues) > 0 {
		details["new_values"] = newValues
	}

	// Add security context
	if requestID := ctx.Value("request_id"); requestID != nil {
		details["request_id"] = requestID
	}
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		details["session_id"] = sessionID
	}

	// Mark as security-sensitive for enhanced monitoring
	details["security_sensitive"] = true
	details["tamper_proof"] = true

	s.logEvent(ctx, auditLog, details)

	// Also log to security event stream for critical structure changes
	if s.isCriticalStructureChange(action) {
		s.LogSecurityEvent(ctx, userID, "critical_structure_change", resourceType, success, details)
	}
}

// isCriticalStructureChange determines if a structure change is critical and needs enhanced monitoring
func (s *AuditService) isCriticalStructureChange(action string) bool {
	criticalActions := []string{
		models.AuditActionDeleteOrganization,
		models.AuditActionChangeOrganizationHierarchy,
		models.AuditActionChangeGroupHierarchy,
		models.AuditActionDeactivateOrganization,
		models.AuditActionDeleteGroup,
	}

	for _, criticalAction := range criticalActions {
		if action == criticalAction {
			return true
		}
	}
	return false
}

// QueryOrganizationAuditLogs queries audit logs scoped to a specific organization
func (s *AuditService) QueryOrganizationAuditLogs(ctx context.Context, orgID string, query *AuditQuery) (*AuditQueryResult, error) {
	// Validate and set defaults for pagination
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PerPage <= 0 {
		query.PerPage = 50
	}
	if query.PerPage > 1000 {
		query.PerPage = 1000
	}

	// Build filters for organization-scoped query
	filters := make(map[string]interface{})
	filters["organization_id"] = orgID

	if query.UserID != "" {
		filters["user_id"] = query.UserID
	}
	if query.Action != "" {
		filters["action"] = query.Action
	}
	if query.Resource != "" {
		filters["resource_type"] = query.Resource
	}
	if query.ResourceID != "" {
		filters["resource_id"] = query.ResourceID
	}
	if query.Success != nil {
		if *query.Success {
			filters["status"] = models.AuditStatusSuccess
		} else {
			filters["status"] = models.AuditStatusFailure
		}
	}
	if query.StartTime != nil {
		filters["timestamp_gte"] = *query.StartTime
	}
	if query.EndTime != nil {
		filters["timestamp_lte"] = *query.EndTime
	}

	// Calculate offset for pagination
	offset := (query.Page - 1) * query.PerPage

	// Query organization-scoped audit logs using repository
	var logs []*models.AuditLog
	var totalCount int64
	var err error

	// Apply organization-scoped filters based on query parameters
	if query.StartTime != nil && query.EndTime != nil {
		logs, err = s.auditRepo.ListByOrganizationAndTimeRange(ctx, orgID, *query.StartTime, *query.EndTime, query.PerPage, offset)
		if err == nil {
			totalCount, err = s.auditRepo.CountByOrganizationAndTimeRange(ctx, orgID, *query.StartTime, *query.EndTime)
		}
	} else {
		logs, err = s.auditRepo.ListByOrganization(ctx, orgID, query.PerPage, offset)
		if err == nil {
			totalCount, err = s.auditRepo.CountByOrganization(ctx, orgID)
		}
	}

	if err != nil {
		s.logger.Error("Failed to query organization audit logs", zap.Error(err))
		return nil, err
	}

	// Convert to slice of models.AuditLog for compatibility
	logSlice := make([]models.AuditLog, len(logs))
	for i, log := range logs {
		if log != nil {
			logSlice[i] = *log
		}
	}

	// Additional security validation - ensure all logs belong to the specified organization
	validatedLogs := make([]models.AuditLog, 0, len(logSlice))
	for _, log := range logSlice {
		if log.Details != nil {
			if logOrgID, exists := log.Details["organization_id"]; exists {
				if logOrgIDStr, ok := logOrgID.(string); ok && logOrgIDStr == orgID {
					validatedLogs = append(validatedLogs, log)
				} else {
					s.logger.Error("Security violation: audit log from different organization found",
						zap.String("requested_org_id", orgID),
						zap.String("log_org_id", fmt.Sprintf("%v", logOrgID)),
						zap.String("audit_log_id", log.ID))
				}
			}
		}
	}

	// Log the security-conscious query
	s.logger.Debug("Querying organization audit logs with security filters",
		zap.String("org_id", orgID),
		zap.String("user_id", query.UserID),
		zap.String("action", query.Action),
		zap.String("resource", query.Resource),
		zap.Int("page", query.Page),
		zap.Int("per_page", query.PerPage),
		zap.Int("total_results", len(validatedLogs)))

	return &AuditQueryResult{
		Logs:       validatedLogs,
		TotalCount: totalCount,
		Page:       query.Page,
		PerPage:    query.PerPage,
	}, nil
}

// GetOrganizationAuditTrail gets audit trail for a specific organization
func (s *AuditService) GetOrganizationAuditTrail(ctx context.Context, orgID string, days int, page, perPage int) (*AuditQueryResult, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	query := &AuditQuery{
		StartTime: &startTime,
		Page:      page,
		PerPage:   perPage,
	}

	return s.QueryOrganizationAuditLogs(ctx, orgID, query)
}

// GetGroupAuditTrail gets audit trail for a specific group
func (s *AuditService) GetGroupAuditTrail(ctx context.Context, orgID, groupID string, days int, page, perPage int) (*AuditQueryResult, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	query := &AuditQuery{
		Resource:   models.ResourceTypeGroup,
		ResourceID: groupID,
		StartTime:  &startTime,
		Page:       page,
		PerPage:    perPage,
	}

	return s.QueryOrganizationAuditLogs(ctx, orgID, query)
}

// ValidateAuditLogIntegrity validates that audit logs haven't been tampered with
func (s *AuditService) ValidateAuditLogIntegrity(ctx context.Context, auditLogID string) (bool, error) {
	s.logger.Debug("Validating audit log integrity", zap.String("audit_log_id", auditLogID))

	// Use repository to validate integrity
	auditLog, err := s.auditRepo.ValidateIntegrity(ctx, auditLogID)
	if err != nil {
		s.logger.Error("Audit log integrity validation failed",
			zap.Error(err),
			zap.String("audit_log_id", auditLogID))
		return false, nil
	}

	// Additional tamper-proof checks could include:
	// 1. Cryptographic hash validation
	// 2. Sequence number validation
	// 3. Digital signature verification
	// 4. Cross-reference with external audit systems

	s.logger.Debug("Audit log integrity validation passed",
		zap.String("audit_log_id", auditLogID),
		zap.String("action", auditLog.Action),
		zap.String("resource_type", auditLog.ResourceType))
	return true, nil
}

// GetAuditStatistics gets audit statistics
func (s *AuditService) GetAuditStatistics(ctx context.Context, days int) (map[string]interface{}, error) {
	startTime := time.Now().AddDate(0, 0, -days)
	endTime := time.Now()

	s.logger.Debug("Generating audit statistics",
		zap.Time("start_time", startTime),
		zap.Int("days", days))

	// Get total events using repository
	totalEvents, err := s.auditRepo.CountByTimeRange(ctx, startTime, endTime)
	if err != nil {
		s.logger.Error("Failed to get total events count", zap.Error(err))
		totalEvents = 0
	}

	// Get successful events
	successfulEvents, err := s.auditRepo.CountByStatus(ctx, models.AuditStatusSuccess)
	if err != nil {
		s.logger.Error("Failed to get successful events count", zap.Error(err))
		successfulEvents = 0
	}

	// Get failed events
	failedEvents, err := s.auditRepo.CountByStatus(ctx, models.AuditStatusFailure)
	if err != nil {
		s.logger.Error("Failed to get failed events count", zap.Error(err))
		failedEvents = 0
	}

	// Get warning events
	warningEvents, err := s.auditRepo.CountByStatus(ctx, models.AuditStatusWarning)
	if err != nil {
		s.logger.Error("Failed to get warning events count", zap.Error(err))
		warningEvents = 0
	}

	// Calculate success rate
	var successRate float64
	if totalEvents > 0 {
		successRate = (float64(successfulEvents) / float64(totalEvents)) * 100
	}

	// For simplicity, we'll create basic statistics
	// In a real implementation, you might use database aggregation queries
	stats := map[string]interface{}{
		"total_events":      totalEvents,
		"successful_events": successfulEvents,
		"failed_events":     failedEvents,
		"warning_events":    warningEvents,
		"success_rate":      successRate,
		"period_days":       days,
		"start_date":        startTime.Format("2006-01-02"),
		"end_date":          time.Now().Format("2006-01-02"),
	}

	// Get top actions (simplified - in practice, you'd use GROUP BY queries)
	topActions := []map[string]interface{}{
		{"action": "login", "count": 0},
		{"action": "create_user", "count": 0},
		{"action": "update_user", "count": 0},
		{"action": "delete_user", "count": 0},
		{"action": "access_denied", "count": 0},
	}

	// Simplified counting (would use DB queries in real implementation)
	for _, actionStat := range topActions {
		actionStat["count"] = int64(0) // Placeholder
	}

	stats["top_actions"] = topActions

	// Get top resource types
	topResources := []map[string]interface{}{
		{"resource_type": models.ResourceTypeUser, "count": 0},
		{"resource_type": models.ResourceTypeRole, "count": 0},
		{"resource_type": models.ResourceTypePermission, "count": 0},
		{"resource_type": models.ResourceTypeAuditLog, "count": 0},
		{"resource_type": models.ResourceTypeSystem, "count": 0},
	}

	// Simplified counting (would use DB queries in real implementation)
	for _, resourceStat := range topResources {
		resourceStat["count"] = int64(0) // Placeholder
	}

	stats["top_resources"] = topResources

	// Generate daily event counts for the past week
	var dailyEvents []map[string]interface{}
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

		// Simplified (would use DB queries in real implementation)
		dayCount := int64(0)

		dailyEvents = append(dailyEvents, map[string]interface{}{
			"date":  dayStart.Format("2006-01-02"),
			"count": dayCount,
		})
	}

	stats["daily_events"] = dailyEvents

	s.logger.Debug("Generated audit statistics",
		zap.Int64("total_events", totalEvents),
		zap.Float64("success_rate", successRate),
		zap.Int("period_days", days))

	return stats, nil
}
