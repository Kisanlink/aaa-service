package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AddOTPAttemptsTable creates the otp_attempts table with all required indexes.
// This table stores individual OTP verification attempts for audit and security tracking.
func AddOTPAttemptsTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Starting otp_attempts table creation migration")
	}

	// Step 1: Check if table already exists
	var tableExists bool
	checkTableSQL := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'otp_attempts'
		)`

	if err := db.WithContext(ctx).Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to check if otp_attempts table exists", zap.Error(err))
		}
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists {
		if logger != nil {
			logger.Info("otp_attempts table already exists, skipping creation")
		}
		return nil
	}

	// Step 2: Create the otp_attempts table
	createTableSQL := `
CREATE TABLE otp_attempts (
	id VARCHAR(255) PRIMARY KEY,
	aadhaar_verification_id VARCHAR(255) NOT NULL,
	attempt_number INT NOT NULL,
	otp_value VARCHAR(6),
	ip_address VARCHAR(45),
	user_agent TEXT,
	status VARCHAR(50),
	failed_reason VARCHAR(255),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (aadhaar_verification_id) REFERENCES aadhaar_verifications(id) ON DELETE CASCADE
)`

	if logger != nil {
		logger.Info("Creating otp_attempts table")
	}

	if err := db.WithContext(ctx).Exec(createTableSQL).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to create otp_attempts table", zap.Error(err))
		}
		return fmt.Errorf("failed to create otp_attempts table: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully created otp_attempts table")
	}

	// Step 3: Create indexes for performance optimization
	indexes := []struct {
		name        string
		description string
		sql         string
	}{
		{
			name:        "idx_otp_attempts_verification_id",
			description: "Optimize queries by aadhaar_verification_id (most common join pattern)",
			sql: `CREATE INDEX IF NOT EXISTS idx_otp_attempts_verification_id
				  ON otp_attempts(aadhaar_verification_id)`,
		},
		{
			name:        "idx_otp_attempts_status",
			description: "Optimize filtering by attempt status for analytics and monitoring",
			sql: `CREATE INDEX IF NOT EXISTS idx_otp_attempts_status
				  ON otp_attempts(status)`,
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

	// Step 4: Summary
	if logger != nil {
		logger.Info("otp_attempts table migration completed",
			zap.Int("total_indexes", len(indexes)),
			zap.Int("created", successCount),
			zap.Int("failed", failureCount))
	}

	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("failed to create all indexes: %d failures", failureCount)
	}

	return nil
}

// DropOTPAttemptsTable removes the otp_attempts table and all its indexes (rollback).
func DropOTPAttemptsTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Rolling back otp_attempts table migration")
	}

	// Step 1: Drop all indexes first
	indexNames := []string{
		"idx_otp_attempts_verification_id",
		"idx_otp_attempts_status",
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

	// Step 2: Drop the table
	dropTableSQL := "DROP TABLE IF EXISTS otp_attempts CASCADE"

	if logger != nil {
		logger.Info("Dropping otp_attempts table")
	}

	if err := db.WithContext(ctx).Exec(dropTableSQL).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to drop otp_attempts table", zap.Error(err))
		}
		return fmt.Errorf("failed to drop otp_attempts table: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully rolled back otp_attempts table migration")
	}

	return nil
}

// ValidateOTPAttemptsTable checks if the migration was successful.
func ValidateOTPAttemptsTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Validating otp_attempts table migration")
	}

	// Check if table exists
	var tableExists bool
	checkTableSQL := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'otp_attempts'
		)`

	if err := db.WithContext(ctx).Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !tableExists {
		return fmt.Errorf("otp_attempts table does not exist")
	}

	// Check if all required columns exist
	var columnCount int64
	checkColumnsSQL := `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'otp_attempts'
		  AND column_name IN (
			'id', 'aadhaar_verification_id', 'attempt_number', 'otp_value',
			'ip_address', 'user_agent', 'status', 'failed_reason', 'created_at'
		  )`

	if err := db.WithContext(ctx).Raw(checkColumnsSQL).Count(&columnCount).Error; err != nil {
		return fmt.Errorf("failed to check columns: %w", err)
	}

	expectedColumns := int64(9)
	if columnCount != expectedColumns {
		return fmt.Errorf("column count mismatch: expected %d, found %d", expectedColumns, columnCount)
	}

	// Check if foreign key constraint exists
	var fkExists bool
	checkFKSQL := `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.table_constraints
			WHERE table_name = 'otp_attempts'
			  AND constraint_type = 'FOREIGN KEY'
		)`

	if err := db.WithContext(ctx).Raw(checkFKSQL).Scan(&fkExists).Error; err != nil {
		return fmt.Errorf("failed to check foreign key constraint: %w", err)
	}

	if !fkExists {
		if logger != nil {
			logger.Warn("Foreign key constraint may be missing")
		}
	}

	// Check if indexes exist
	var indexCount int64
	checkIndexesSQL := `
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE tablename = 'otp_attempts'
		  AND indexname IN (
			'idx_otp_attempts_verification_id',
			'idx_otp_attempts_status'
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

	if logger != nil {
		logger.Info("otp_attempts table validation completed",
			zap.Int64("columns", columnCount),
			zap.Int64("indexes", indexCount),
			zap.Bool("foreign_key_exists", fkExists))
	}

	return nil
}
