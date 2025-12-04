package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AddSMSDeliveryLogsTable creates the sms_delivery_logs table with all required indexes.
// This table stores audit records for all SMS delivery attempts for security tracking.
func AddSMSDeliveryLogsTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Starting sms_delivery_logs table creation migration")
	}

	// Step 1: Check if table already exists
	var tableExists bool
	checkTableSQL := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'sms_delivery_logs'
		)`

	if err := db.WithContext(ctx).Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to check if sms_delivery_logs table exists", zap.Error(err))
		}
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists {
		if logger != nil {
			logger.Info("sms_delivery_logs table already exists, skipping creation")
		}
		return nil
	}

	// Step 2: Create the sms_delivery_logs table
	createTableSQL := `
CREATE TABLE sms_delivery_logs (
	id VARCHAR(255) PRIMARY KEY,
	user_id VARCHAR(255),
	phone_number_masked VARCHAR(20) NOT NULL,
	message_type VARCHAR(50) NOT NULL,
	sns_message_id VARCHAR(100),
	status VARCHAR(20) NOT NULL DEFAULT 'pending',
	failure_reason TEXT,
	sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	delivered_at TIMESTAMP,
	request_details JSONB,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
)`

	if logger != nil {
		logger.Info("Creating sms_delivery_logs table")
	}

	if err := db.WithContext(ctx).Exec(createTableSQL).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to create sms_delivery_logs table", zap.Error(err))
		}
		return fmt.Errorf("failed to create sms_delivery_logs table: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully created sms_delivery_logs table")
	}

	// Step 3: Create indexes for performance optimization
	indexes := []struct {
		name        string
		description string
		sql         string
	}{
		{
			name:        "idx_sms_delivery_logs_user_id",
			description: "Optimize queries by user_id for audit trail lookup",
			sql: `CREATE INDEX IF NOT EXISTS idx_sms_delivery_logs_user_id
				  ON sms_delivery_logs(user_id)`,
		},
		{
			name:        "idx_sms_delivery_logs_phone_masked",
			description: "Optimize rate limiting queries by masked phone number",
			sql: `CREATE INDEX IF NOT EXISTS idx_sms_delivery_logs_phone_masked
				  ON sms_delivery_logs(phone_number_masked)`,
		},
		{
			name:        "idx_sms_delivery_logs_status",
			description: "Optimize filtering by delivery status for monitoring",
			sql: `CREATE INDEX IF NOT EXISTS idx_sms_delivery_logs_status
				  ON sms_delivery_logs(status)`,
		},
		{
			name:        "idx_sms_delivery_logs_sent_at",
			description: "Optimize time-based queries for rate limiting and cleanup",
			sql: `CREATE INDEX IF NOT EXISTS idx_sms_delivery_logs_sent_at
				  ON sms_delivery_logs(sent_at)`,
		},
		{
			name:        "idx_sms_delivery_logs_message_type",
			description: "Optimize filtering by message type for analytics",
			sql: `CREATE INDEX IF NOT EXISTS idx_sms_delivery_logs_message_type
				  ON sms_delivery_logs(message_type)`,
		},
		{
			name:        "idx_sms_delivery_logs_rate_limit",
			description: "Composite index for rate limiting queries",
			sql: `CREATE INDEX IF NOT EXISTS idx_sms_delivery_logs_rate_limit
				  ON sms_delivery_logs(phone_number_masked, sent_at)`,
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
		logger.Info("sms_delivery_logs table migration completed",
			zap.Int("total_indexes", len(indexes)),
			zap.Int("created", successCount),
			zap.Int("failed", failureCount))
	}

	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("failed to create all indexes: %d failures", failureCount)
	}

	return nil
}

// DropSMSDeliveryLogsTable removes the sms_delivery_logs table and all its indexes (rollback).
func DropSMSDeliveryLogsTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Rolling back sms_delivery_logs table migration")
	}

	// Step 1: Drop all indexes first
	indexNames := []string{
		"idx_sms_delivery_logs_user_id",
		"idx_sms_delivery_logs_phone_masked",
		"idx_sms_delivery_logs_status",
		"idx_sms_delivery_logs_sent_at",
		"idx_sms_delivery_logs_message_type",
		"idx_sms_delivery_logs_rate_limit",
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
	dropTableSQL := "DROP TABLE IF EXISTS sms_delivery_logs CASCADE"

	if logger != nil {
		logger.Info("Dropping sms_delivery_logs table")
	}

	if err := db.WithContext(ctx).Exec(dropTableSQL).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to drop sms_delivery_logs table", zap.Error(err))
		}
		return fmt.Errorf("failed to drop sms_delivery_logs table: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully rolled back sms_delivery_logs table migration")
	}

	return nil
}

// ValidateSMSDeliveryLogsTable checks if the migration was successful.
func ValidateSMSDeliveryLogsTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Validating sms_delivery_logs table migration")
	}

	// Check if table exists
	var tableExists bool
	checkTableSQL := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'sms_delivery_logs'
		)`

	if err := db.WithContext(ctx).Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !tableExists {
		return fmt.Errorf("sms_delivery_logs table does not exist")
	}

	// Check if all required columns exist
	var columnCount int64
	checkColumnsSQL := `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'sms_delivery_logs'
		  AND column_name IN (
			'id', 'user_id', 'phone_number_masked', 'message_type',
			'sns_message_id', 'status', 'failure_reason', 'sent_at',
			'delivered_at', 'request_details', 'created_at', 'updated_at', 'deleted_at'
		  )`

	if err := db.WithContext(ctx).Raw(checkColumnsSQL).Scan(&columnCount).Error; err != nil {
		return fmt.Errorf("failed to check columns: %w", err)
	}

	expectedColumns := int64(13)
	if columnCount != expectedColumns {
		return fmt.Errorf("column count mismatch: expected %d, found %d", expectedColumns, columnCount)
	}

	// Check if indexes exist
	var indexCount int64
	checkIndexesSQL := `
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE tablename = 'sms_delivery_logs'
		  AND indexname IN (
			'idx_sms_delivery_logs_user_id',
			'idx_sms_delivery_logs_phone_masked',
			'idx_sms_delivery_logs_status',
			'idx_sms_delivery_logs_sent_at',
			'idx_sms_delivery_logs_message_type',
			'idx_sms_delivery_logs_rate_limit'
		  )`

	if err := db.WithContext(ctx).Raw(checkIndexesSQL).Scan(&indexCount).Error; err != nil {
		return fmt.Errorf("failed to check indexes: %w", err)
	}

	if indexCount < 6 {
		if logger != nil {
			logger.Warn("Some indexes may be missing",
				zap.Int64("found", indexCount),
				zap.Int("expected", 6))
		}
	}

	if logger != nil {
		logger.Info("sms_delivery_logs table validation completed",
			zap.Int64("columns", columnCount),
			zap.Int64("indexes", indexCount))
	}

	return nil
}
