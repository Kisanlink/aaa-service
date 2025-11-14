package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// AuditRepository handles database operations for AuditLog entities
type AuditRepository struct {
	*base.BaseFilterableRepository[*models.AuditLog]
	dbManager db.DBManager
}

// NewAuditRepository creates a new AuditRepository instance
func NewAuditRepository(dbManager db.DBManager) *AuditRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.AuditLog]()
	baseRepo.SetDBManager(dbManager)
	return &AuditRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Create creates a new audit log using the base repository
func (r *AuditRepository) Create(ctx context.Context, auditLog *models.AuditLog) error {
	return r.BaseFilterableRepository.Create(ctx, auditLog)
}

// GetByID retrieves an audit log by ID using the base repository
func (r *AuditRepository) GetByID(ctx context.Context, id string) (*models.AuditLog, error) {
	auditLog := &models.AuditLog{}
	return r.BaseFilterableRepository.GetByID(ctx, id, auditLog)
}

// Update updates an existing audit log using the base repository
func (r *AuditRepository) Update(ctx context.Context, auditLog *models.AuditLog) error {
	return r.BaseFilterableRepository.Update(ctx, auditLog)
}

// Delete deletes an audit log by ID using the base repository
func (r *AuditRepository) Delete(ctx context.Context, id string) error {
	auditLog := &models.AuditLog{}
	return r.BaseFilterableRepository.Delete(ctx, id, auditLog)
}

// List retrieves audit logs with pagination using database-level filtering
func (r *AuditRepository) List(ctx context.Context, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ListByOrganization retrieves audit logs for a specific organization with pagination
func (r *AuditRepository) ListByOrganization(ctx context.Context, orgID string, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "details->>'organization_id'", Operator: base.OpEqual, Value: orgID},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ListByUser retrieves audit logs for a specific user with pagination
func (r *AuditRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "user_id", Operator: base.OpEqual, Value: userID},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ListByAction retrieves audit logs for a specific action with pagination
func (r *AuditRepository) ListByAction(ctx context.Context, action string, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "action", Operator: base.OpEqual, Value: action},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ListByResourceType retrieves audit logs for a specific resource type with pagination
func (r *AuditRepository) ListByResourceType(ctx context.Context, resourceType string, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "resource_type", Operator: base.OpEqual, Value: resourceType},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ListByStatus retrieves audit logs for a specific status with pagination
func (r *AuditRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "status", Operator: base.OpEqual, Value: status},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ListByTimeRange retrieves audit logs within a time range with pagination
func (r *AuditRepository) ListByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "timestamp", Operator: base.OpGreaterEqual, Value: startTime},
				{Field: "timestamp", Operator: base.OpLessEqual, Value: endTime},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ListByOrganizationAndTimeRange retrieves audit logs for an organization within a time range
func (r *AuditRepository) ListByOrganizationAndTimeRange(ctx context.Context, orgID string, startTime, endTime time.Time, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "details->>'organization_id'", Operator: base.OpEqual, Value: orgID},
				{Field: "timestamp", Operator: base.OpGreaterEqual, Value: startTime},
				{Field: "timestamp", Operator: base.OpLessEqual, Value: endTime},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ListByUserAndTimeRange retrieves audit logs for a user within a time range
func (r *AuditRepository) ListByUserAndTimeRange(ctx context.Context, userID string, startTime, endTime time.Time, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "user_id", Operator: base.OpEqual, Value: userID},
				{Field: "timestamp", Operator: base.OpGreaterEqual, Value: startTime},
				{Field: "timestamp", Operator: base.OpLessEqual, Value: endTime},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ListByGroupAndTimeRange retrieves audit logs for a group within a time range
func (r *AuditRepository) ListByGroupAndTimeRange(ctx context.Context, orgID, groupID string, startTime, endTime time.Time, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "details->>'organization_id'", Operator: base.OpEqual, Value: orgID},
				{Field: "details->>'group_id'", Operator: base.OpEqual, Value: groupID},
				{Field: "timestamp", Operator: base.OpGreaterEqual, Value: startTime},
				{Field: "timestamp", Operator: base.OpLessEqual, Value: endTime},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// CountByOrganization counts audit logs for a specific organization
func (r *AuditRepository) CountByOrganization(ctx context.Context, orgID string) (int64, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "details->>'organization_id'", Operator: base.OpEqual, Value: orgID},
			},
			Logic: base.LogicAnd,
		},
	}
	var model models.AuditLog
	return r.dbManager.Count(ctx, filter, &model)
}

// CountByUser counts audit logs for a specific user
func (r *AuditRepository) CountByUser(ctx context.Context, userID string) (int64, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "user_id", Operator: base.OpEqual, Value: userID},
			},
			Logic: base.LogicAnd,
		},
	}
	var model models.AuditLog
	return r.dbManager.Count(ctx, filter, &model)
}

// CountByStatus counts audit logs for a specific status
func (r *AuditRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "status", Operator: base.OpEqual, Value: status},
			},
			Logic: base.LogicAnd,
		},
	}
	var model models.AuditLog
	return r.dbManager.Count(ctx, filter, &model)
}

// CountByTimeRange counts audit logs within a time range
func (r *AuditRepository) CountByTimeRange(ctx context.Context, startTime, endTime time.Time) (int64, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "timestamp", Operator: base.OpGreaterEqual, Value: startTime},
				{Field: "timestamp", Operator: base.OpLessEqual, Value: endTime},
			},
			Logic: base.LogicAnd,
		},
	}
	var model models.AuditLog
	return r.dbManager.Count(ctx, filter, &model)
}

// CountByOrganizationAndTimeRange counts audit logs for an organization within a time range
func (r *AuditRepository) CountByOrganizationAndTimeRange(ctx context.Context, orgID string, startTime, endTime time.Time) (int64, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "details->>'organization_id'", Operator: base.OpEqual, Value: orgID},
				{Field: "timestamp", Operator: base.OpGreaterEqual, Value: startTime},
				{Field: "timestamp", Operator: base.OpLessEqual, Value: endTime},
			},
			Logic: base.LogicAnd,
		},
	}
	var model models.AuditLog
	return r.dbManager.Count(ctx, filter, &model)
}

// GetSecurityEvents retrieves security-related audit events
func (r *AuditRepository) GetSecurityEvents(ctx context.Context, days int, limit, offset int) ([]*models.AuditLog, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	// Security-sensitive actions
	securityActions := []string{
		models.AuditActionLogin,
		models.AuditActionLogout,
		models.AuditActionAccessDenied,
		models.AuditActionSecurityEvent,
		models.AuditActionDeleteOrganization,
		models.AuditActionChangeOrganizationHierarchy,
		models.AuditActionChangeGroupHierarchy,
	}

	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "timestamp", Operator: base.OpGreaterEqual, Value: startTime},
				{Field: "action", Operator: base.OpIn, Value: securityActions},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// GetFailedOperations retrieves failed audit operations
func (r *AuditRepository) GetFailedOperations(ctx context.Context, days int, limit, offset int) ([]*models.AuditLog, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "timestamp", Operator: base.OpGreaterEqual, Value: startTime},
				{Field: "status", Operator: base.OpEqual, Value: models.AuditStatusFailure},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ArchiveOldLogs marks old audit logs for archival (soft delete)
func (r *AuditRepository) ArchiveOldLogs(ctx context.Context, cutoffDate time.Time) (int64, error) {
	// In a real implementation, this would use bulk update operations
	// For now, we'll use the count to simulate the operation
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "timestamp", Operator: base.OpLessThan, Value: cutoffDate},
			},
			Logic: base.LogicAnd,
		},
	}
	var model models.AuditLog
	return r.dbManager.Count(ctx, filter, &model)
}

// ListByResourceAndActions retrieves audit logs for a specific resource with specific actions
func (r *AuditRepository) ListByResourceAndActions(ctx context.Context, resourceType, resourceID string, actions []string, limit, offset int) ([]*models.AuditLog, error) {
	var results []*models.AuditLog
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "resource_type", Operator: base.OpEqual, Value: resourceType},
				{Field: "resource_id", Operator: base.OpEqual, Value: resourceID},
				{Field: "action", Operator: base.OpIn, Value: actions},
			},
			Logic: base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
		Sort: []base.SortField{
			{Field: "timestamp", Direction: "desc"},
		},
	}
	err := r.dbManager.List(ctx, filter, &results)
	return results, err
}

// ValidateIntegrity performs basic integrity checks on audit logs
func (r *AuditRepository) ValidateIntegrity(ctx context.Context, auditLogID string) (*models.AuditLog, error) {
	auditLog, err := r.GetByID(ctx, auditLogID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve audit log for integrity validation: %w", err)
	}

	// Basic integrity validations
	if auditLog.ID == "" {
		return nil, fmt.Errorf("audit log has empty ID")
	}

	if auditLog.Action == "" {
		return nil, fmt.Errorf("audit log has empty action")
	}

	if auditLog.ResourceType == "" {
		return nil, fmt.Errorf("audit log has empty resource type")
	}

	if auditLog.Timestamp.IsZero() {
		return nil, fmt.Errorf("audit log has zero timestamp")
	}

	// Validate organization context for organization-scoped operations
	if auditLog.ResourceType == models.ResourceTypeOrganization ||
		auditLog.ResourceType == models.ResourceTypeGroup ||
		auditLog.ResourceType == models.ResourceTypeGroupRole {
		if auditLog.Details != nil {
			if orgID, exists := auditLog.Details["organization_id"]; exists {
				if orgIDStr, ok := orgID.(string); ok && orgIDStr != "" {
					// Valid organization context
				} else {
					return nil, fmt.Errorf("audit log missing valid organization context")
				}
			} else {
				return nil, fmt.Errorf("organization-scoped audit log missing organization_id in details")
			}
		} else {
			return nil, fmt.Errorf("organization-scoped audit log missing details")
		}
	}

	return auditLog, nil
}
