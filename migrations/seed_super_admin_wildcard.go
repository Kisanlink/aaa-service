package migrations

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// SeedSuperAdminWildcardPermissions ensures super_admin has explicit wildcard
// resource_permissions entries for all resource types and actions.
// This guarantees that super_admin has ALL permissions in the system.
func SeedSuperAdminWildcardPermissions(ctx context.Context, primary db.DBManager, logger *zap.Logger) error {
	// Find super_admin role
	var roles []models.Role
	roleFilter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "name", Operator: base.OpEqual, Value: "super_admin"},
			},
			Logic: base.LogicAnd,
		},
	}
	if err := primary.List(ctx, roleFilter, &roles); err != nil {
		return fmt.Errorf("failed to find super_admin role: %w", err)
	}
	if len(roles) == 0 {
		return fmt.Errorf("super_admin role not found - run core seeding first")
	}
	superAdminRole := &roles[0]

	// Get all resource types
	var resources []models.Resource
	emptyFilter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{},
			Logic:      base.LogicAnd,
		},
	}
	if err := primary.List(ctx, emptyFilter, &resources); err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	// Get all actions
	var actions []models.Action
	if err := primary.List(ctx, emptyFilter, &actions); err != nil {
		return fmt.Errorf("failed to list actions: %w", err)
	}

	if logger != nil {
		logger.Info("Seeding wildcard permissions for super_admin",
			zap.String("role_id", superAdminRole.ID),
			zap.Int("resources", len(resources)),
			zap.Int("actions", len(actions)))
	}

	// Create wildcard resource_permissions for each resource type and action
	// Using resource_id = '*' to grant access to all resources of that type
	createdCount := 0
	skippedCount := 0

	for _, resource := range resources {
		for _, action := range actions {
			// Check if wildcard permission already exists
			var existingPerms []models.ResourcePermission
			permFilter := &base.Filter{
				Group: base.FilterGroup{
					Conditions: []base.FilterCondition{
						{Field: "role_id", Operator: base.OpEqual, Value: superAdminRole.ID},
						{Field: "resource_type", Operator: base.OpEqual, Value: resource.Type},
						{Field: "resource_id", Operator: base.OpEqual, Value: "*"},
						{Field: "action", Operator: base.OpEqual, Value: action.Name},
						{Field: "is_active", Operator: base.OpEqual, Value: true},
					},
					Logic: base.LogicAnd,
				},
			}
			if err := primary.List(ctx, permFilter, &existingPerms); err != nil {
				logger.Warn("Failed to check existing permission",
					zap.String("resource", resource.Name),
					zap.String("action", action.Name),
					zap.Error(err))
				continue
			}

			if len(existingPerms) > 0 {
				skippedCount++
				continue
			}

			// Create new wildcard resource permission
			wildcardPerm := models.NewResourcePermission("*", resource.Type, superAdminRole.ID, action.Name)
			if err := primary.Create(ctx, wildcardPerm); err != nil {
				// Log but continue - some combinations might fail due to constraints
				if logger != nil {
					logger.Debug("Skipped permission creation",
						zap.String("resource", resource.Name),
						zap.String("action", action.Name),
						zap.Error(err))
				}
				skippedCount++
				continue
			}

			createdCount++
			if logger != nil && createdCount%50 == 0 {
				logger.Debug("Progress update",
					zap.Int("created", createdCount),
					zap.Int("skipped", skippedCount))
			}
		}
	}

	if logger != nil {
		logger.Info("Completed wildcard permission seeding for super_admin",
			zap.Int("created", createdCount),
			zap.Int("skipped", skippedCount),
			zap.Int("total_combinations", len(resources)*len(actions)))
	}

	return nil
}

// SeedSuperAdminWildcardPermissionsWithDBManager is a convenience wrapper using DatabaseManager
func SeedSuperAdminWildcardPermissionsWithDBManager(
	ctx context.Context,
	dm *db.DatabaseManager,
	logger *zap.Logger,
) error {
	if dm == nil {
		return fmt.Errorf("database manager is nil")
	}

	primary := dm.GetManager(db.BackendGorm)
	if primary == nil {
		return fmt.Errorf("primary DB manager (gorm) not available")
	}

	return SeedSuperAdminWildcardPermissions(ctx, primary, logger)
}
