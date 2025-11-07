package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AddPerformanceIndexes creates all necessary database indexes for optimal query performance.
// This migration should be run after auto-migration completes to ensure tables exist.
func AddPerformanceIndexes(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Starting comprehensive index creation for AAA service")
	}

	// Define all indexes with proper naming and optimization
	indexes := []struct {
		name        string
		description string
		sql         string
	}{
		// ========================================
		// CRITICAL PERFORMANCE INDEXES
		// ========================================
		{
			name:        "idx_group_memberships_principal_active",
			description: "Group memberships lookup by principal (CRITICAL - fixes login timeout)",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_group_memberships_principal_active
				  ON group_memberships(principal_id, is_active)
				  WHERE is_active = true`,
		},
		{
			name:        "idx_group_memberships_group_active",
			description: "Group memberships lookup by group",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_group_memberships_group_active
				  ON group_memberships(group_id, is_active)
				  WHERE is_active = true`,
		},

		// ========================================
		// USER INDEXES
		// ========================================
		{
			name:        "idx_users_phone_country",
			description: "User lookup by phone and country (login authentication)",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_phone_country
				  ON users(phone_number, country_code)
				  WHERE deleted_at IS NULL`,
		},
		{
			name:        "idx_users_email",
			description: "User lookup by email",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email
				  ON users(email)
				  WHERE deleted_at IS NULL AND email IS NOT NULL`,
		},
		{
			name:        "idx_users_username",
			description: "User lookup by username",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_username
				  ON users(username)
				  WHERE deleted_at IS NULL AND username IS NOT NULL`,
		},
		{
			name:        "idx_users_status",
			description: "User filtering by status",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_status
				  ON users(status)
				  WHERE deleted_at IS NULL AND status IS NOT NULL`,
		},
		{
			name:        "idx_users_created_at",
			description: "User ordering and filtering by creation time",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_created_at
				  ON users(created_at DESC)
				  WHERE deleted_at IS NULL`,
		},

		// ========================================
		// USER ROLES INDEXES
		// ========================================
		{
			name:        "idx_user_roles_user_active",
			description: "User roles lookup by user (role checking)",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_roles_user_active
				  ON user_roles(user_id, is_active)
				  WHERE is_active = true`,
		},
		{
			name:        "idx_user_roles_role_active",
			description: "User roles lookup by role (reverse lookup)",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_roles_role_active
				  ON user_roles(role_id, is_active)
				  WHERE is_active = true`,
		},
		{
			name:        "idx_user_roles_user_role_active",
			description: "User roles unique combination lookup",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_roles_user_role_active
				  ON user_roles(user_id, role_id, is_active)
				  WHERE is_active = true`,
		},

		// ========================================
		// ROLES INDEXES
		// ========================================
		{
			name:        "idx_roles_name_active",
			description: "Role lookup by name",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_roles_name_active
				  ON roles(name, is_active)
				  WHERE is_active = true`,
		},
		{
			name:        "idx_roles_scope_active",
			description: "Role filtering by scope",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_roles_scope_active
				  ON roles(scope, is_active)
				  WHERE is_active = true`,
		},
		{
			name:        "idx_roles_org_active",
			description: "Role filtering by organization",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_roles_org_active
				  ON roles(organization_id, is_active)
				  WHERE is_active = true AND organization_id IS NOT NULL`,
		},

		// ========================================
		// ROLE PERMISSIONS INDEXES
		// ========================================
		{
			name:        "idx_role_permissions_role_active",
			description: "Role permissions lookup by role",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_role_permissions_role_active
				  ON role_permissions(role_id, is_active)
				  WHERE is_active = true`,
		},
		{
			name:        "idx_role_permissions_permission_active",
			description: "Role permissions lookup by permission (reverse lookup)",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_role_permissions_permission_active
				  ON role_permissions(permission_id, is_active)
				  WHERE is_active = true`,
		},
		{
			name:        "idx_role_permissions_role_perm_active",
			description: "Role permissions unique combination lookup",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_role_permissions_role_perm_active
				  ON role_permissions(role_id, permission_id, is_active)
				  WHERE is_active = true`,
		},

		// ========================================
		// PERMISSIONS INDEXES
		// ========================================
		{
			name:        "idx_permissions_name",
			description: "Permission lookup by name",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_permissions_name
				  ON permissions(name)`,
		},
		{
			name:        "idx_permissions_resource_action",
			description: "Permission lookup by resource and action",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_permissions_resource_action
				  ON permissions(resource_id, action_id)`,
		},
		{
			name:        "idx_permissions_resource",
			description: "Permission filtering by resource",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_permissions_resource
				  ON permissions(resource_id)`,
		},

		// ========================================
		// RESOURCES INDEXES
		// ========================================
		{
			name:        "idx_resources_name",
			description: "Resource lookup by name",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_resources_name
				  ON resources(name)`,
		},
		{
			name:        "idx_resources_type",
			description: "Resource filtering by type",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_resources_type
				  ON resources(type)`,
		},

		// ========================================
		// ACTIONS INDEXES
		// ========================================
		{
			name:        "idx_actions_name",
			description: "Action lookup by name",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_actions_name
				  ON actions(name)`,
		},

		// ========================================
		// USER PROFILES INDEXES
		// ========================================
		{
			name:        "idx_user_profiles_user",
			description: "User profile lookup by user (1-to-1 relationship)",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_profiles_user
				  ON user_profiles(user_id)`,
		},

		// ========================================
		// ORGANIZATIONS INDEXES
		// ========================================
		{
			name:        "idx_organizations_name",
			description: "Organization lookup by name",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_organizations_name
				  ON organizations(name)
				  WHERE deleted_at IS NULL`,
		},
		{
			name:        "idx_organizations_parent",
			description: "Organization hierarchy lookup",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_organizations_parent
				  ON organizations(parent_id)
				  WHERE deleted_at IS NULL AND parent_id IS NOT NULL`,
		},
		{
			name:        "idx_organizations_status",
			description: "Organization filtering by status",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_organizations_status
				  ON organizations(status)
				  WHERE deleted_at IS NULL`,
		},

		// ========================================
		// GROUPS INDEXES
		// ========================================
		{
			name:        "idx_groups_name",
			description: "Group lookup by name",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_groups_name
				  ON groups(name)
				  WHERE deleted_at IS NULL`,
		},
		{
			name:        "idx_groups_org",
			description: "Group filtering by organization",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_groups_org
				  ON groups(organization_id)
				  WHERE deleted_at IS NULL`,
		},
		{
			name:        "idx_groups_parent",
			description: "Group hierarchy lookup",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_groups_parent
				  ON groups(parent_id)
				  WHERE deleted_at IS NULL AND parent_id IS NOT NULL`,
		},

		// ========================================
		// AUDIT LOGS INDEXES
		// ========================================
		{
			name:        "idx_audit_logs_user_action",
			description: "Audit log lookup by user and action",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_action
				  ON audit_logs(user_id, action, created_at DESC)`,
		},
		{
			name:        "idx_audit_logs_resource",
			description: "Audit log lookup by resource",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_resource
				  ON audit_logs(resource_type, resource_id, created_at DESC)`,
		},
		{
			name:        "idx_audit_logs_timestamp",
			description: "Audit log ordering and filtering by time",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_timestamp
				  ON audit_logs(created_at DESC)`,
		},
		{
			name:        "idx_audit_logs_status",
			description: "Audit log filtering by status",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_status
				  ON audit_logs(status, created_at DESC)`,
		},
		{
			name:        "idx_audit_logs_ip_address",
			description: "Audit log security analysis by IP",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_ip_address
				  ON audit_logs(ip_address, created_at DESC)
				  WHERE ip_address IS NOT NULL`,
		},

		// ========================================
		// SESSIONS INDEXES (if table exists)
		// ========================================
		{
			name:        "idx_sessions_user_active",
			description: "Session lookup by user (active sessions)",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_user_active
				  ON sessions(user_id, is_active, expires_at DESC)
				  WHERE is_active = true`,
		},
		{
			name:        "idx_sessions_token",
			description: "Session lookup by token (authentication)",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_token
				  ON sessions(token)
				  WHERE is_active = true`,
		},
		{
			name:        "idx_sessions_expires_at",
			description: "Session cleanup by expiration",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_expires_at
				  ON sessions(expires_at)
				  WHERE is_active = true`,
		},

		// ========================================
		// REFRESH TOKENS INDEXES (if table exists)
		// ========================================
		{
			name:        "idx_refresh_tokens_token",
			description: "Refresh token lookup (authentication)",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_refresh_tokens_token
				  ON refresh_tokens(token)
				  WHERE is_revoked = false`,
		},
		{
			name:        "idx_refresh_tokens_user",
			description: "Refresh token lookup by user",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_refresh_tokens_user
				  ON refresh_tokens(user_id, is_revoked, expires_at DESC)
				  WHERE is_revoked = false`,
		},

		// ========================================
		// API KEYS INDEXES (if table exists)
		// ========================================
		{
			name:        "idx_api_keys_key_hash",
			description: "API key lookup (authentication)",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_api_keys_key_hash
				  ON api_keys(key_hash)
				  WHERE is_active = true`,
		},
		{
			name:        "idx_api_keys_user_active",
			description: "API key lookup by user",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_api_keys_user_active
				  ON api_keys(user_id, is_active, expires_at DESC)
				  WHERE is_active = true`,
		},
	}

	// Execute each index creation
	successCount := 0
	failureCount := 0
	skippedCount := 0

	for _, idx := range indexes {
		// Check if index already exists
		var exists bool
		checkSQL := fmt.Sprintf(`
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes
				WHERE indexname = '%s'
			)`, idx.name)

		if err := db.WithContext(ctx).Raw(checkSQL).Scan(&exists).Error; err != nil {
			if logger != nil {
				logger.Warn("Failed to check index existence",
					zap.String("index", idx.name),
					zap.Error(err))
			}
			// Continue anyway, CREATE INDEX IF NOT EXISTS will handle it
		}

		if exists {
			if logger != nil {
				logger.Debug("Index already exists, skipping",
					zap.String("index", idx.name))
			}
			skippedCount++
			continue
		}

		// Create the index
		if logger != nil {
			logger.Info("Creating index",
				zap.String("index", idx.name),
				zap.String("description", idx.description))
		}

		if err := db.WithContext(ctx).Exec(idx.sql).Error; err != nil {
			// Log error but continue with other indexes
			if logger != nil {
				logger.Error("Failed to create index",
					zap.String("index", idx.name),
					zap.String("description", idx.description),
					zap.String("sql", idx.sql),
					zap.Error(err))
			}
			failureCount++
			continue
		}

		successCount++
		if logger != nil {
			logger.Info("Successfully created index",
				zap.String("index", idx.name))
		}
	}

	// Analyze all tables to update statistics
	if logger != nil {
		logger.Info("Analyzing tables to update query planner statistics")
	}

	tables := []string{
		"users", "user_roles", "user_profiles",
		"roles", "role_permissions",
		"permissions", "resources", "actions",
		"organizations", "groups", "group_memberships",
		"audit_logs", "sessions", "refresh_tokens", "api_keys",
	}

	for _, table := range tables {
		analyzeSQL := fmt.Sprintf("ANALYZE %s", table)
		if err := db.WithContext(ctx).Exec(analyzeSQL).Error; err != nil {
			if logger != nil {
				logger.Warn("Failed to analyze table",
					zap.String("table", table),
					zap.Error(err))
			}
			// Continue with other tables
			continue
		}
	}

	// Summary
	if logger != nil {
		logger.Info("Index creation completed",
			zap.Int("total", len(indexes)),
			zap.Int("created", successCount),
			zap.Int("skipped", skippedCount),
			zap.Int("failed", failureCount))
	}

	// Return error only if all indexes failed
	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("failed to create any indexes: %d failures out of %d total", failureCount, len(indexes))
	}

	return nil
}

// DropPerformanceIndexes removes all performance indexes (for rollback if needed)
func DropPerformanceIndexes(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Dropping performance indexes")
	}

	indexNames := []string{
		"idx_group_memberships_principal_active",
		"idx_group_memberships_group_active",
		"idx_users_phone_country",
		"idx_users_email",
		"idx_users_username",
		"idx_users_status",
		"idx_users_created_at",
		"idx_user_roles_user_active",
		"idx_user_roles_role_active",
		"idx_user_roles_user_role_active",
		"idx_roles_name_active",
		"idx_roles_scope_active",
		"idx_roles_org_active",
		"idx_role_permissions_role_active",
		"idx_role_permissions_permission_active",
		"idx_role_permissions_role_perm_active",
		"idx_permissions_name",
		"idx_permissions_resource_action",
		"idx_permissions_resource",
		"idx_resources_name",
		"idx_resources_type",
		"idx_actions_name",
		"idx_user_profiles_user",
		"idx_organizations_name",
		"idx_organizations_parent",
		"idx_organizations_status",
		"idx_groups_name",
		"idx_groups_org",
		"idx_groups_parent",
		"idx_audit_logs_user_action",
		"idx_audit_logs_resource",
		"idx_audit_logs_timestamp",
		"idx_audit_logs_status",
		"idx_audit_logs_ip_address",
		"idx_sessions_user_active",
		"idx_sessions_token",
		"idx_sessions_expires_at",
		"idx_refresh_tokens_token",
		"idx_refresh_tokens_user",
		"idx_api_keys_key_hash",
		"idx_api_keys_user_active",
	}

	for _, indexName := range indexNames {
		dropSQL := fmt.Sprintf("DROP INDEX CONCURRENTLY IF EXISTS %s", indexName)
		if err := db.WithContext(ctx).Exec(dropSQL).Error; err != nil {
			if logger != nil {
				logger.Warn("Failed to drop index",
					zap.String("index", indexName),
					zap.Error(err))
			}
			// Continue with other indexes
			continue
		}

		if logger != nil {
			logger.Info("Dropped index", zap.String("index", indexName))
		}
	}

	return nil
}
