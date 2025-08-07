package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// AuditService provides audit logging services
type AuditService struct {
	dbManager    db.DBManager
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
	cacheService interfaces.CacheService,
	logger *zap.Logger,
) *AuditService {
	// Note: In this implementation, we assume the audit_logs table migration
	// is handled elsewhere in the application startup

	return &AuditService{
		dbManager:    dbManager,
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

	// For now, return empty results as a placeholder
	// In a real implementation, you'd query the database with proper filtering
	logs := []models.AuditLog{}
	totalCount := int64(0)

	s.logger.Debug("Querying audit logs (simplified implementation)",
		zap.String("user_id", query.UserID),
		zap.String("action", query.Action),
		zap.String("resource", query.Resource),
		zap.Int("page", query.Page),
		zap.Int("per_page", query.PerPage))

	return &AuditQueryResult{
		Logs:       logs,
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

	// Save audit log to database
	err := s.dbManager.Create(ctx, auditLog)
	if err != nil {
		// If database save fails, still log to application logger
		s.logger.Error("Failed to save audit log to database",
			zap.Error(err),
			zap.String("audit_id", auditLog.ID),
			zap.String("action", auditLog.Action))
	}

	// Also log to application logger for real-time monitoring
	userID := ""
	if auditLog.UserID != nil {
		userID = *auditLog.UserID
	}
	resourceID := ""
	if auditLog.ResourceID != nil {
		resourceID = *auditLog.ResourceID
	}
	s.logger.Info("Audit event",
		zap.String("audit_id", auditLog.ID),
		zap.String("user_id", userID),
		zap.String("action", auditLog.Action),
		zap.String("resource_type", auditLog.ResourceType),
		zap.String("resource_id", resourceID),
		zap.String("status", auditLog.Status))

	// Also log to application logs for critical security events
	if auditLog.Action == models.AuditActionAccessDenied || auditLog.Action == "login_failed" || auditLog.IsFailure() {
		s.logger.Warn("Security event logged",
			zap.String("user_id", userID),
			zap.String("action", auditLog.Action),
			zap.String("resource_type", auditLog.ResourceType),
			zap.String("resource_id", resourceID),
			zap.String("status", auditLog.Status),
			zap.String("message", auditLog.Message))
	}
}

// GetAuditStatistics gets audit statistics
func (s *AuditService) GetAuditStatistics(ctx context.Context, days int) (map[string]interface{}, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	// For now, use simplified statistics (would be replaced with actual DB queries)
	totalEvents := int64(0)
	successfulEvents := int64(0)

	// In a real implementation, you'd query the database here
	s.logger.Debug("Generating audit statistics",
		zap.Time("start_time", startTime),
		zap.Int("days", days))

	// Calculate success rate
	var successRate float64
	if totalEvents > 0 {
		successRate = (float64(successfulEvents) / float64(totalEvents)) * 100
	}

	// Simplified for now - would use actual DB queries
	failedEvents := int64(0)
	warningEvents := int64(0)

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

// generateUniqueID generates a unique ID for audit logs
func generateUniqueID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("audit_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}
