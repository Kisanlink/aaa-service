package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AddHierarchyFields adds hierarchy tracking fields and performance indexes for organizations and groups.
// This migration supports efficient hierarchy traversal and role inheritance.
func AddHierarchyFields(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Starting hierarchy fields migration for AAA service")
	}

	// Step 1: Add hierarchy fields through auto-migration (already handled by model changes)
	// The fields are already added to models with proper GORM tags

	// Step 2: Create performance indexes for hierarchy optimization
	hierarchyIndexes := []struct {
		name        string
		description string
		sql         string
	}{
		// ========================================
		// ORGANIZATION HIERARCHY INDEXES
		// ========================================
		{
			name:        "idx_org_parent_active",
			description: "Optimize finding active child organizations",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_parent_active
				  ON organizations(parent_id, is_active)
				  WHERE deleted_at IS NULL`,
		},
		{
			name:        "idx_org_hierarchy_depth",
			description: "Optimize queries filtering by organization depth",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_hierarchy_depth
				  ON organizations(hierarchy_depth)
				  WHERE deleted_at IS NULL`,
		},
		{
			name:        "idx_org_hierarchy_composite",
			description: "Optimize complex hierarchy traversal queries",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_hierarchy_composite
				  ON organizations(parent_id, hierarchy_depth, is_active)
				  WHERE deleted_at IS NULL`,
		},

		// ========================================
		// GROUP HIERARCHY INDEXES
		// ========================================
		{
			name:        "idx_group_parent_org",
			description: "Optimize group hierarchy within organizations",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_group_parent_org
				  ON groups(parent_id, organization_id)
				  WHERE deleted_at IS NULL`,
		},
		{
			name:        "idx_group_hierarchy_depth",
			description: "Optimize queries filtering by group depth",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_group_hierarchy_depth
				  ON groups(hierarchy_depth)
				  WHERE deleted_at IS NULL`,
		},
		{
			name:        "idx_group_hierarchy_composite",
			description: "Optimize complex group hierarchy queries",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_group_hierarchy_composite
				  ON groups(organization_id, parent_id, hierarchy_depth, is_active)
				  WHERE deleted_at IS NULL`,
		},

		// ========================================
		// GROUP MEMBERSHIP VERSION INDEX
		// ========================================
		{
			name:        "idx_group_membership_version",
			description: "Optimize version-based change queries",
			sql: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_group_membership_version
				  ON group_memberships(version, updated_at)
				  WHERE deleted_at IS NULL`,
		},
	}

	// Execute each index creation
	successCount := 0
	failureCount := 0
	skippedCount := 0

	for _, idx := range hierarchyIndexes {
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
			logger.Info("Creating hierarchy index",
				zap.String("index", idx.name),
				zap.String("description", idx.description))
		}

		if err := db.WithContext(ctx).Exec(idx.sql).Error; err != nil {
			if logger != nil {
				logger.Error("Failed to create hierarchy index",
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
			logger.Info("Successfully created hierarchy index",
				zap.String("index", idx.name))
		}
	}

	// Step 3: Create functions for hierarchy maintenance
	hierarchyFunctions := []struct {
		name        string
		description string
		sql         string
	}{
		{
			name:        "calculate_hierarchy_depth",
			description: "Function to calculate hierarchy depth for existing records",
			sql: `
CREATE OR REPLACE FUNCTION calculate_hierarchy_depth(
    table_name TEXT,
    record_id VARCHAR(255),
    parent_column TEXT DEFAULT 'parent_id',
    max_depth INTEGER DEFAULT 10
) RETURNS INTEGER AS $$
DECLARE
    depth INTEGER := 0;
    current_parent VARCHAR(255);
BEGIN
    IF table_name = 'organizations' THEN
        SELECT parent_id INTO current_parent
        FROM organizations
        WHERE id = record_id AND deleted_at IS NULL;
    ELSIF table_name = 'groups' THEN
        SELECT parent_id INTO current_parent
        FROM groups
        WHERE id = record_id AND deleted_at IS NULL;
    ELSE
        RETURN 0;
    END IF;

    WHILE current_parent IS NOT NULL AND depth < max_depth LOOP
        depth := depth + 1;

        IF table_name = 'organizations' THEN
            SELECT parent_id INTO current_parent
            FROM organizations
            WHERE id = current_parent AND deleted_at IS NULL;
        ELSIF table_name = 'groups' THEN
            SELECT parent_id INTO current_parent
            FROM groups
            WHERE id = current_parent AND deleted_at IS NULL;
        END IF;
    END LOOP;

    RETURN depth;
END;
$$ LANGUAGE plpgsql;`,
		},
		{
			name:        "build_hierarchy_path",
			description: "Function to build hierarchy path for existing records",
			sql: `
CREATE OR REPLACE FUNCTION build_hierarchy_path(
    table_name TEXT,
    record_id VARCHAR(255)
) RETURNS TEXT AS $$
DECLARE
    path_array VARCHAR(255)[] := ARRAY[]::VARCHAR(255)[];
    current_id VARCHAR(255) := record_id;
    current_parent VARCHAR(255);
    max_iterations INTEGER := 20;
    iterations INTEGER := 0;
BEGIN
    -- Build path from bottom to top
    WHILE current_id IS NOT NULL AND iterations < max_iterations LOOP
        path_array := array_prepend(current_id, path_array);

        IF table_name = 'organizations' THEN
            SELECT parent_id INTO current_parent
            FROM organizations
            WHERE id = current_id AND deleted_at IS NULL;
        ELSIF table_name = 'groups' THEN
            SELECT parent_id INTO current_parent
            FROM groups
            WHERE id = current_id AND deleted_at IS NULL;
        ELSE
            EXIT;
        END IF;

        current_id := current_parent;
        iterations := iterations + 1;
    END LOOP;

    -- Return path as /id1/id2/id3
    IF array_length(path_array, 1) > 0 THEN
        RETURN '/' || array_to_string(path_array, '/');
    ELSE
        RETURN NULL;
    END IF;
END;
$$ LANGUAGE plpgsql;`,
		},
	}

	// Create functions
	for _, fn := range hierarchyFunctions {
		if logger != nil {
			logger.Info("Creating function",
				zap.String("function", fn.name),
				zap.String("description", fn.description))
		}

		if err := db.WithContext(ctx).Exec(fn.sql).Error; err != nil {
			if logger != nil {
				logger.Warn("Failed to create function (may already exist)",
					zap.String("function", fn.name),
					zap.Error(err))
			}
			// Continue - function might already exist
		}
	}

	// Step 4: Populate hierarchy fields for existing records
	if logger != nil {
		logger.Info("Populating hierarchy fields for existing records")
	}

	// Update organizations with calculated hierarchy values
	updateOrgSQL := `
UPDATE organizations
SET
    hierarchy_depth = COALESCE(calculate_hierarchy_depth('organizations', id), 0),
    hierarchy_path = COALESCE(build_hierarchy_path('organizations', id), '/' || id)
WHERE deleted_at IS NULL
  AND (hierarchy_depth = 0 OR hierarchy_path IS NULL OR hierarchy_path = '')`

	if err := db.WithContext(ctx).Exec(updateOrgSQL).Error; err != nil {
		if logger != nil {
			logger.Warn("Failed to update organization hierarchy fields",
				zap.Error(err))
		}
	} else {
		if logger != nil {
			logger.Info("Successfully updated organization hierarchy fields")
		}
	}

	// Update groups with calculated hierarchy values
	updateGroupSQL := `
UPDATE groups
SET
    hierarchy_depth = COALESCE(calculate_hierarchy_depth('groups', id), 0),
    hierarchy_path = COALESCE(build_hierarchy_path('groups', id), '/' || id)
WHERE deleted_at IS NULL
  AND (hierarchy_depth = 0 OR hierarchy_path IS NULL OR hierarchy_path = '')`

	if err := db.WithContext(ctx).Exec(updateGroupSQL).Error; err != nil {
		if logger != nil {
			logger.Warn("Failed to update group hierarchy fields",
				zap.Error(err))
		}
	} else {
		if logger != nil {
			logger.Info("Successfully updated group hierarchy fields")
		}
	}

	// Step 5: Analyze tables to update statistics
	if logger != nil {
		logger.Info("Analyzing tables to update query planner statistics")
	}

	tables := []string{"organizations", "groups", "group_memberships"}
	for _, table := range tables {
		analyzeSQL := fmt.Sprintf("ANALYZE %s", table)
		if err := db.WithContext(ctx).Exec(analyzeSQL).Error; err != nil {
			if logger != nil {
				logger.Warn("Failed to analyze table",
					zap.String("table", table),
					zap.Error(err))
			}
		}
	}

	// Summary
	if logger != nil {
		logger.Info("Hierarchy fields migration completed",
			zap.Int("total_indexes", len(hierarchyIndexes)),
			zap.Int("created", successCount),
			zap.Int("skipped", skippedCount),
			zap.Int("failed", failureCount))
	}

	// Return error only if critical indexes failed
	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("failed to create hierarchy indexes: %d failures out of %d total", failureCount, len(hierarchyIndexes))
	}

	return nil
}

// DropHierarchyFields removes hierarchy fields and indexes (for rollback if needed)
func DropHierarchyFields(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Dropping hierarchy fields and indexes")
	}

	// Drop functions
	dropFunctions := []string{
		"DROP FUNCTION IF EXISTS calculate_hierarchy_depth(TEXT, VARCHAR(255), TEXT, INTEGER)",
		"DROP FUNCTION IF EXISTS build_hierarchy_path(TEXT, VARCHAR(255))",
	}

	for _, sql := range dropFunctions {
		if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
			if logger != nil {
				logger.Warn("Failed to drop function", zap.Error(err))
			}
		}
	}

	// Drop indexes
	indexNames := []string{
		"idx_org_parent_active",
		"idx_org_hierarchy_depth",
		"idx_org_hierarchy_composite",
		"idx_group_parent_org",
		"idx_group_hierarchy_depth",
		"idx_group_hierarchy_composite",
		"idx_group_membership_version",
		"idx_org_hierarchy_path",
		"idx_group_hierarchy_path",
	}

	for _, indexName := range indexNames {
		dropSQL := fmt.Sprintf("DROP INDEX CONCURRENTLY IF EXISTS %s", indexName)
		if err := db.WithContext(ctx).Exec(dropSQL).Error; err != nil {
			if logger != nil {
				logger.Warn("Failed to drop index",
					zap.String("index", indexName),
					zap.Error(err))
			}
		} else {
			if logger != nil {
				logger.Info("Dropped index", zap.String("index", indexName))
			}
		}
	}

	// Note: Column removal should be done through model changes and auto-migration
	// We don't drop columns here to maintain data integrity

	return nil
}

// ValidateHierarchyMigration checks if the hierarchy migration was successful
func ValidateHierarchyMigration(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Validating hierarchy migration")
	}

	// Check if columns exist
	var columnCount int64
	checkColumnsSQL := `
SELECT COUNT(*)
FROM information_schema.columns
WHERE table_name IN ('organizations', 'groups')
  AND column_name IN ('hierarchy_depth', 'hierarchy_path')
`
	if err := db.WithContext(ctx).Raw(checkColumnsSQL).Count(&columnCount).Error; err != nil {
		return fmt.Errorf("failed to check hierarchy columns: %w", err)
	}

	expectedColumns := int64(4) // 2 columns x 2 tables
	if columnCount < expectedColumns {
		return fmt.Errorf("hierarchy columns missing: expected %d, found %d", expectedColumns, columnCount)
	}

	// Check if indexes exist
	var indexCount int64
	checkIndexesSQL := `
SELECT COUNT(*)
FROM pg_indexes
WHERE indexname LIKE 'idx_%hierarchy%'
   OR indexname LIKE 'idx_org_parent_active'
   OR indexname LIKE 'idx_group_parent_org'
   OR indexname LIKE 'idx_group_membership_version'
`
	if err := db.WithContext(ctx).Raw(checkIndexesSQL).Count(&indexCount).Error; err != nil {
		return fmt.Errorf("failed to check hierarchy indexes: %w", err)
	}

	if indexCount < 5 { // Minimum expected indexes
		if logger != nil {
			logger.Warn("Some hierarchy indexes may be missing",
				zap.Int64("found", indexCount),
				zap.Int("minimum_expected", 5))
		}
	}

	// Check if functions exist
	var functionCount int64
	checkFunctionsSQL := `
SELECT COUNT(*)
FROM pg_proc
WHERE proname IN ('calculate_hierarchy_depth', 'build_hierarchy_path')
`
	if err := db.WithContext(ctx).Raw(checkFunctionsSQL).Count(&functionCount).Error; err != nil {
		return fmt.Errorf("failed to check hierarchy functions: %w", err)
	}

	if functionCount < 2 {
		if logger != nil {
			logger.Warn("Hierarchy functions may be missing",
				zap.Int64("found", functionCount),
				zap.Int("expected", 2))
		}
	}

	if logger != nil {
		logger.Info("Hierarchy migration validation completed",
			zap.Int64("columns", columnCount),
			zap.Int64("indexes", indexCount),
			zap.Int64("functions", functionCount))
	}

	return nil
}
