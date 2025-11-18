package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AddUserProfilesKYCFields adds KYC-related columns to the user_profiles table.
// This migration supports Aadhaar verification tracking at the user profile level.
func AddUserProfilesKYCFields(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Starting user_profiles KYC fields migration")
	}

	// Step 1: Check if user_profiles table exists
	var tableExists bool
	checkTableSQL := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'user_profiles'
		)`

	if err := db.WithContext(ctx).Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to check if user_profiles table exists", zap.Error(err))
		}
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !tableExists {
		if logger != nil {
			logger.Error("user_profiles table does not exist, cannot add KYC fields")
		}
		return fmt.Errorf("user_profiles table does not exist")
	}

	// Step 2: Check if columns already exist
	columnsToAdd := []string{"aadhaar_verified", "aadhaar_verified_at", "kyc_status"}
	var existingColumnCount int64

	checkColumnsSQL := `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'user_profiles'
		  AND column_name IN ('aadhaar_verified', 'aadhaar_verified_at', 'kyc_status')`

	if err := db.WithContext(ctx).Raw(checkColumnsSQL).Count(&existingColumnCount).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to check existing columns", zap.Error(err))
		}
		return fmt.Errorf("failed to check existing columns: %w", err)
	}

	if existingColumnCount == int64(len(columnsToAdd)) {
		if logger != nil {
			logger.Info("KYC columns already exist in user_profiles table, skipping addition")
		}
		// Still create indexes if they don't exist
		return createUserProfilesKYCIndexes(ctx, db, logger)
	}

	// Step 3: Add KYC columns to user_profiles table
	alterTableSQL := `
ALTER TABLE user_profiles
ADD COLUMN IF NOT EXISTS aadhaar_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS aadhaar_verified_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS kyc_status VARCHAR(50) DEFAULT 'PENDING'`

	if logger != nil {
		logger.Info("Adding KYC fields to user_profiles table")
	}

	if err := db.WithContext(ctx).Exec(alterTableSQL).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to add KYC fields to user_profiles", zap.Error(err))
		}
		return fmt.Errorf("failed to add KYC fields to user_profiles: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully added KYC fields to user_profiles table")
	}

	// Step 4: Create indexes
	return createUserProfilesKYCIndexes(ctx, db, logger)
}

// createUserProfilesKYCIndexes creates indexes for KYC fields in user_profiles table.
func createUserProfilesKYCIndexes(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	indexes := []struct {
		name        string
		description string
		sql         string
	}{
		{
			name:        "idx_user_profiles_kyc_status",
			description: "Optimize filtering user profiles by KYC status",
			sql: `CREATE INDEX IF NOT EXISTS idx_user_profiles_kyc_status
				  ON user_profiles(kyc_status)`,
		},
		{
			name:        "idx_user_profiles_aadhaar_verified",
			description: "Optimize filtering verified vs unverified users",
			sql: `CREATE INDEX IF NOT EXISTS idx_user_profiles_aadhaar_verified
				  ON user_profiles(aadhaar_verified)`,
		},
	}

	successCount := 0
	failureCount := 0

	for _, idx := range indexes {
		if logger != nil {
			logger.Info("Creating index",
				zap.String("index", idx.name),
				zap.String("description", idx.description))
		}

		if err := db.WithContext(ctx).Exec(idx.sql).Error; err != nil {
			if logger != nil {
				logger.Error("Failed to create index",
					zap.String("index", idx.name),
					zap.Error(err))
			}
			failureCount++
			continue
		}

		successCount++
		if logger != nil {
			logger.Info("Successfully created index", zap.String("index", idx.name))
		}
	}

	// Step 5: Analyze table to update query planner statistics
	if logger != nil {
		logger.Info("Analyzing user_profiles table to update statistics")
	}

	analyzeSQL := "ANALYZE user_profiles"
	if err := db.WithContext(ctx).Exec(analyzeSQL).Error; err != nil {
		if logger != nil {
			logger.Warn("Failed to analyze user_profiles table", zap.Error(err))
		}
		// Don't fail - this is an optimization
	}

	// Step 6: Summary
	if logger != nil {
		logger.Info("user_profiles KYC fields migration completed",
			zap.Int("total_indexes", len(indexes)),
			zap.Int("created", successCount),
			zap.Int("failed", failureCount))
	}

	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("failed to create all indexes: %d failures", failureCount)
	}

	return nil
}

// DropUserProfilesKYCFields removes KYC columns from user_profiles table (rollback).
func DropUserProfilesKYCFields(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Rolling back user_profiles KYC fields migration")
	}

	// Step 1: Drop indexes first
	indexNames := []string{
		"idx_user_profiles_kyc_status",
		"idx_user_profiles_aadhaar_verified",
	}

	for _, indexName := range indexNames {
		dropIndexSQL := fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)
		if err := db.WithContext(ctx).Exec(dropIndexSQL).Error; err != nil {
			if logger != nil {
				logger.Warn("Failed to drop index",
					zap.String("index", indexName),
					zap.Error(err))
			}
			// Continue even if index drop fails
		} else {
			if logger != nil {
				logger.Info("Dropped index", zap.String("index", indexName))
			}
		}
	}

	// Step 2: Drop KYC columns
	alterTableSQL := `
ALTER TABLE user_profiles
DROP COLUMN IF EXISTS aadhaar_verified,
DROP COLUMN IF EXISTS aadhaar_verified_at,
DROP COLUMN IF EXISTS kyc_status`

	if logger != nil {
		logger.Info("Removing KYC fields from user_profiles table")
	}

	if err := db.WithContext(ctx).Exec(alterTableSQL).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to remove KYC fields from user_profiles", zap.Error(err))
		}
		return fmt.Errorf("failed to remove KYC fields from user_profiles: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully rolled back user_profiles KYC fields migration")
	}

	return nil
}

// ValidateUserProfilesKYCFields checks if the migration was successful.
func ValidateUserProfilesKYCFields(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Validating user_profiles KYC fields migration")
	}

	// Check if user_profiles table exists
	var tableExists bool
	checkTableSQL := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'user_profiles'
		)`

	if err := db.WithContext(ctx).Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !tableExists {
		return fmt.Errorf("user_profiles table does not exist")
	}

	// Check if KYC columns exist
	var columnCount int64
	checkColumnsSQL := `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'user_profiles'
		  AND column_name IN ('aadhaar_verified', 'aadhaar_verified_at', 'kyc_status')`

	if err := db.WithContext(ctx).Raw(checkColumnsSQL).Count(&columnCount).Error; err != nil {
		return fmt.Errorf("failed to check columns: %w", err)
	}

	expectedColumns := int64(3)
	if columnCount != expectedColumns {
		return fmt.Errorf("KYC column count mismatch: expected %d, found %d", expectedColumns, columnCount)
	}

	// Check if indexes exist
	var indexCount int64
	checkIndexesSQL := `
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE tablename = 'user_profiles'
		  AND indexname IN (
			'idx_user_profiles_kyc_status',
			'idx_user_profiles_aadhaar_verified'
		  )`

	if err := db.WithContext(ctx).Raw(checkIndexesSQL).Count(&indexCount).Error; err != nil {
		return fmt.Errorf("failed to check indexes: %w", err)
	}

	if indexCount < 2 {
		if logger != nil {
			logger.Warn("Some indexes may be missing",
				zap.Int64("found", indexCount),
				zap.Int("expected", 2))
		}
	}

	// Verify column defaults and data types
	type ColumnInfo struct {
		ColumnName    string
		DataType      string
		ColumnDefault *string
	}

	var columns []ColumnInfo
	checkDetailsSQL := `
		SELECT column_name, data_type, column_default
		FROM information_schema.columns
		WHERE table_name = 'user_profiles'
		  AND column_name IN ('aadhaar_verified', 'aadhaar_verified_at', 'kyc_status')
		ORDER BY column_name`

	if err := db.WithContext(ctx).Raw(checkDetailsSQL).Scan(&columns).Error; err != nil {
		return fmt.Errorf("failed to check column details: %w", err)
	}

	// Validate column properties
	for _, col := range columns {
		if logger != nil {
			defaultValue := "NULL"
			if col.ColumnDefault != nil {
				defaultValue = *col.ColumnDefault
			}
			logger.Info("KYC column details",
				zap.String("column", col.ColumnName),
				zap.String("type", col.DataType),
				zap.String("default", defaultValue))
		}
	}

	if logger != nil {
		logger.Info("user_profiles KYC fields validation completed",
			zap.Int64("columns", columnCount),
			zap.Int64("indexes", indexCount))
	}

	return nil
}
