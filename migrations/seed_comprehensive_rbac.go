package migrations

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// SeedComprehensiveRBAC seeds all RBAC resources and creates a comprehensive permission matrix
// This extends the core seed with all resource types defined in the system
func SeedComprehensiveRBAC(ctx context.Context, primary db.DBManager, logger *zap.Logger) error {
	if err := seedAllResources(ctx, primary, logger); err != nil {
		return fmt.Errorf("seed all resources: %w", err)
	}
	if err := seedComprehensivePermissions(ctx, primary, logger); err != nil {
		return fmt.Errorf("seed comprehensive permissions: %w", err)
	}
	return nil
}

// seedAllResources creates all resource types defined in the system
func seedAllResources(ctx context.Context, primary db.DBManager, logger *zap.Logger) error {
	// Comprehensive list of all resources matching resource.go constants
	allResources := []struct {
		Name string
		Type string
		Desc string
	}{
		// Core AAA resources (already seeded by seed_core_roles_permissions.go)
		{"user", models.ResourceTypeUser, "User accounts and authentication"},
		{"role", models.ResourceTypeRole, "Role definitions and assignments"},
		{"permission", models.ResourceTypePermission, "Permission definitions"},
		{"audit_log", models.ResourceTypeAuditLog, "Audit trail and logging"},
		{"system", models.ResourceTypeSystem, "System configuration and management"},
		{"api_endpoint", models.ResourceTypeAPIEndpoint, "API endpoint protection"},
		{"resource", models.ResourceTypeResource, "Generic AAA resource management"},

		// Extended AAA resources (new additions)
		{"organization", models.ResourceTypeOrganization, "Organization management and hierarchy"},
		{"group", models.ResourceTypeGroup, "Group management and memberships"},
		{"group_role", models.ResourceTypeGroupRole, "Group role assignments"},
		{"action", models.ResourceTypeAction, "Action definitions for RBAC"},

		// User-related resources
		{"user_profile", models.ResourceTypeUserProfile, "User profile information"},
		{"contact", models.ResourceTypeContact, "Contact information management"},
		{"address", models.ResourceTypeAddress, "Address management"},

		// Permission-related resources
		{"column_permission", models.ResourceTypeColumnPermission, "Column-level permissions"},
		{"column", models.ResourceTypeColumn, "Column definitions"},
		{"temporary_permission", models.ResourceTypeTemporaryPermission, "Time-bound permissions"},

		// Resource relationships
		{"user_resource", models.ResourceTypeUserResource, "User-specific resource access"},
		{"role_resource", models.ResourceTypeRoleResource, "Role-specific resource access"},
		{"permission_resource", models.ResourceTypePermissionResource, "Permission-resource mappings"},

		// Advanced resources
		{"hierarchical_resource", models.ResourceTypeHierarchicalResource, "Hierarchical resource structures"},

		// Database resources
		{"database", models.ResourceTypeDatabase, "Database instance management"},
		{"table", models.ResourceTypeTable, "Database table access"},
		{"database_operation", models.ResourceTypeDatabaseOperation, "Database operation control"},
	}

	for _, r := range allResources {
		var existing []models.Resource
		filters := []base.FilterCondition{{Field: "name", Operator: base.OpEqual, Value: r.Name}}
		filter := &base.Filter{
			Group: base.FilterGroup{
				Conditions: filters,
				Logic:      base.LogicAnd,
			},
		}
		if err := primary.List(ctx, filter, &existing); err != nil {
			return err
		}
		if len(existing) > 0 {
			if logger != nil {
				logger.Debug("Resource already exists", zap.String("name", r.Name))
			}
			continue
		}

		res := models.NewResource(r.Name, r.Type, r.Desc)
		if err := primary.Create(ctx, res); err != nil {
			return fmt.Errorf("create resource %s: %w", r.Name, err)
		}
		if logger != nil {
			logger.Info("Created resource", zap.String("name", r.Name), zap.String("type", r.Type))
		}
	}

	if logger != nil {
		logger.Info("Completed comprehensive resource seeding", zap.Int("total_resources", len(allResources)))
	}
	return nil
}

// seedComprehensivePermissions creates additional permissions for the new resources
// This complements the core permissions already created by seed_core_roles_permissions.go
func seedComprehensivePermissions(ctx context.Context, primary db.DBManager, logger *zap.Logger) error {
	// Load all resources
	var resources []models.Resource
	emptyFilter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{},
			Logic:      base.LogicAnd,
		},
	}
	if err := primary.List(ctx, emptyFilter, &resources); err != nil {
		return err
	}
	resByName := map[string]*models.Resource{}
	for i := range resources {
		resByName[resources[i].Name] = &resources[i]
	}

	// Build actions index
	actIdx, err := buildActionIndexDM(ctx, primary)
	if err != nil {
		return err
	}

	// Load all roles
	var roles []models.Role
	if err := primary.List(ctx, emptyFilter, &roles); err != nil {
		return err
	}
	rolesByName := map[string]*models.Role{}
	for i := range roles {
		rolesByName[roles[i].Name] = &roles[i]
	}

	// Extended permission matrices for new resources
	// Super admin gets comprehensive access to all resources
	superAdminExtensions := map[string][]string{
		"organization":         {"manage", "create", "read", "update", "delete"},
		"group":                {"manage", "create", "read", "update", "delete"},
		"group_role":           {"manage", "assign", "unassign", "read"},
		"action":               {"manage", "create", "read", "update", "delete"},
		"user_profile":         {"read", "update", "delete"},
		"contact":              {"read", "update", "delete"},
		"address":              {"read", "update", "delete"},
		"column_permission":    {"manage", "create", "read", "update", "delete"},
		"column":               {"read", "manage"},
		"temporary_permission": {"create", "read", "revoke"},
		"database":             {"manage", "backup", "restore"},
		"table":                {"read", "update", "truncate"},
		"database_operation":   {"execute", "manage"},
	}

	// Admin gets organization and group management
	adminExtensions := map[string][]string{
		"organization": {"read", "update"},
		"group":        {"create", "read", "update", "delete"},
		"group_role":   {"assign", "unassign", "read"},
		"user_profile": {"read", "update"},
		"contact":      {"read", "update"},
		"address":      {"read", "update"},
	}

	// AAA Admin gets full access to AAA-specific resources
	aaaAdminExtensions := map[string][]string{
		"action":               {"manage", "create", "read", "update", "delete"},
		"column_permission":    {"manage", "create", "read", "update", "delete"},
		"column":               {"read", "manage"},
		"temporary_permission": {"create", "read", "revoke"},
		"user_resource":        {"read", "manage"},
		"role_resource":        {"read", "manage"},
		"permission_resource":  {"read", "manage"},
	}

	// Module Admin gets organization and group access
	moduleAdminExtensions := map[string][]string{
		"organization": {"read"},
		"group":        {"read", "update"},
		"group_role":   {"read", "assign"},
	}

	// User gets self-service access to their own data
	userExtensions := map[string][]string{
		"user_profile": {"read", "update"},
		"contact":      {"read", "update"},
		"address":      {"read", "update"},
		"organization": {"read"},
		"group":        {"read"},
	}

	// Viewer gets read-only access
	viewerExtensions := map[string][]string{
		"organization": {"read"},
		"group":        {"read"},
		"user_profile": {"read"},
	}

	// Helper to create permission and attach to role
	upsertAndAttachPermission := func(role *models.Role, resourceName, actionName string) error {
		res := resByName[resourceName]
		act, ok := actIdx.byName[actionName]
		if res == nil || !ok {
			if logger != nil {
				logger.Debug("Skipping permission - resource or action not found",
					zap.String("resource", resourceName),
					zap.String("action", actionName))
			}
			return nil
		}

		permName := fmt.Sprintf("%s:%s", resourceName, actionName)
		var perms []models.Permission
		permFilter := &base.Filter{
			Group: base.FilterGroup{
				Conditions: []base.FilterCondition{{Field: "name", Operator: base.OpEqual, Value: permName}},
				Logic:      base.LogicAnd,
			},
		}
		if err := primary.List(ctx, permFilter, &perms); err != nil {
			return err
		}

		var perm models.Permission
		if len(perms) == 0 {
			// Create new permission
			newPerm := models.NewPermissionWithResourceAndAction(
				permName,
				fmt.Sprintf("%s on %s", actionName, resourceName),
				res.ID,
				act.ID,
			)
			if err := primary.Create(ctx, newPerm); err != nil {
				return fmt.Errorf("create permission %s: %w", permName, err)
			}
			perm = *newPerm
			if logger != nil {
				logger.Debug("Created permission", zap.String("permission", permName))
			}
		} else {
			perm = perms[0]
		}

		// Check if role-permission relationship exists
		var existingRPs []models.RolePermission
		rpFilter := &base.Filter{
			Group: base.FilterGroup{
				Conditions: []base.FilterCondition{
					{Field: "role_id", Operator: base.OpEqual, Value: role.ID},
					{Field: "permission_id", Operator: base.OpEqual, Value: perm.ID},
					{Field: "is_active", Operator: base.OpEqual, Value: true},
				},
				Logic: base.LogicAnd,
			},
		}
		if err := primary.List(ctx, rpFilter, &existingRPs); err != nil {
			return err
		}

		if len(existingRPs) == 0 {
			rp := models.NewRolePermission(role.ID, perm.ID)
			if err := primary.Create(ctx, rp); err != nil {
				return fmt.Errorf("create role-permission %s:%s: %w", role.Name, permName, err)
			}
			if logger != nil {
				logger.Debug("Assigned permission to role",
					zap.String("role", role.Name),
					zap.String("permission", permName))
			}
		}

		return nil
	}

	// Apply permission matrices to roles
	permissionSets := map[string]map[string][]string{
		"super_admin":  superAdminExtensions,
		"admin":        adminExtensions,
		"aaa_admin":    aaaAdminExtensions,
		"module_admin": moduleAdminExtensions,
		"user":         userExtensions,
		"viewer":       viewerExtensions,
	}

	for roleName, matrix := range permissionSets {
		role := rolesByName[roleName]
		if role == nil {
			if logger != nil {
				logger.Warn("Role not found, skipping permissions", zap.String("role", roleName))
			}
			continue
		}

		for resourceName, actions := range matrix {
			for _, action := range actions {
				if err := upsertAndAttachPermission(role, resourceName, action); err != nil {
					return fmt.Errorf("assign %s:%s to %s: %w", resourceName, action, roleName, err)
				}
			}
		}

		if logger != nil {
			logger.Info("Applied extended permissions to role", zap.String("role", roleName))
		}
	}

	if logger != nil {
		logger.Info("Completed comprehensive permission seeding")
	}
	return nil
}

// SeedComprehensiveRBACWithDBManager is a convenience wrapper using DatabaseManager
func SeedComprehensiveRBACWithDBManager(
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

	return SeedComprehensiveRBAC(ctx, primary, logger)
}
