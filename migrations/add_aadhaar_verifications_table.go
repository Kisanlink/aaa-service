package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AddAadhaarVerificationsTable creates the aadhaar_verifications table with all required indexes.
// This table stores Aadhaar OTP verification requests and KYC data.
func AddAadhaarVerificationsTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Starting aadhaar_verifications table creation migration")
	}

	// Step 1: Check if table already exists
	var tableExists bool
	checkTableSQL := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'aadhaar_verifications'
		)`

	if err := db.WithContext(ctx).Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to check if aadhaar_verifications table exists", zap.Error(err))
		}
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists {
		if logger != nil {
			logger.Info("aadhaar_verifications table already exists, skipping creation")
		}
		return nil
	}

	// Step 2: Create the aadhaar_verifications table
	createTableSQL := `
CREATE TABLE aadhaar_verifications (
	id VARCHAR(255) PRIMARY KEY,
	user_id VARCHAR(255) NOT NULL,
	aadhaar_number VARCHAR(12),
	transaction_id VARCHAR(255) UNIQUE,
	reference_id VARCHAR(255) UNIQUE,
	otp_requested_at TIMESTAMP,
	otp_verified_at TIMESTAMP,
	verification_status VARCHAR(50) DEFAULT 'PENDING',
	kyc_status VARCHAR(50) DEFAULT 'PENDING',
	photo_url TEXT,
	name VARCHAR(255),
	date_of_birth DATE,
	gender VARCHAR(20),
	full_address TEXT,
	address_json JSONB,
	attempts INT DEFAULT 0,
	last_attempt_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP,
	created_by VARCHAR(255),
	updated_by VARCHAR(255),
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)`

	if logger != nil {
		logger.Info("Creating aadhaar_verifications table")
	}

	if err := db.WithContext(ctx).Exec(createTableSQL).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to create aadhaar_verifications table", zap.Error(err))
		}
		return fmt.Errorf("failed to create aadhaar_verifications table: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully created aadhaar_verifications table")
	}

	// Step 3: Create indexes for performance optimization
	indexes := []struct {
		name        string
		description string
		sql         string
	}{
		{
			name:        "idx_aadhaar_verifications_user_id",
			description: "Optimize queries by user_id (most common access pattern)",
			sql: `CREATE INDEX IF NOT EXISTS idx_aadhaar_verifications_user_id
				  ON aadhaar_verifications(user_id)`,
		},
		{
			name:        "idx_aadhaar_verifications_transaction_id",
			description: "Optimize OTP generation lookups by transaction_id",
			sql: `CREATE INDEX IF NOT EXISTS idx_aadhaar_verifications_transaction_id
				  ON aadhaar_verifications(transaction_id)`,
		},
		{
			name:        "idx_aadhaar_verifications_reference_id",
			description: "Optimize OTP verification lookups by reference_id",
			sql: `CREATE INDEX IF NOT EXISTS idx_aadhaar_verifications_reference_id
				  ON aadhaar_verifications(reference_id)`,
		},
		{
			name:        "idx_aadhaar_verifications_status",
			description: "Optimize status filtering for verification and KYC workflows",
			sql: `CREATE INDEX IF NOT EXISTS idx_aadhaar_verifications_status
				  ON aadhaar_verifications(verification_status, kyc_status)`,
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
		logger.Info("aadhaar_verifications table migration completed",
			zap.Int("total_indexes", len(indexes)),
			zap.Int("created", successCount),
			zap.Int("failed", failureCount))
	}

	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("failed to create all indexes: %d failures", failureCount)
	}

	return nil
}

// DropAadhaarVerificationsTable removes the aadhaar_verifications table and all its indexes (rollback).
func DropAadhaarVerificationsTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Rolling back aadhaar_verifications table migration")
	}

	// Step 1: Drop all indexes first
	indexNames := []string{
		"idx_aadhaar_verifications_user_id",
		"idx_aadhaar_verifications_transaction_id",
		"idx_aadhaar_verifications_reference_id",
		"idx_aadhaar_verifications_status",
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
	dropTableSQL := "DROP TABLE IF EXISTS aadhaar_verifications CASCADE"

	if logger != nil {
		logger.Info("Dropping aadhaar_verifications table")
	}

	if err := db.WithContext(ctx).Exec(dropTableSQL).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to drop aadhaar_verifications table", zap.Error(err))
		}
		return fmt.Errorf("failed to drop aadhaar_verifications table: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully rolled back aadhaar_verifications table migration")
	}

	return nil
}

// ValidateAadhaarVerificationsTable checks if the migration was successful.
func ValidateAadhaarVerificationsTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Validating aadhaar_verifications table migration")
	}

	// Check if table exists
	var tableExists bool
	checkTableSQL := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'aadhaar_verifications'
		)`

	if err := db.WithContext(ctx).Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !tableExists {
		return fmt.Errorf("aadhaar_verifications table does not exist")
	}

	// Check if all required columns exist
	var columnCount int64
	checkColumnsSQL := `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'aadhaar_verifications'
		  AND column_name IN (
			'id', 'user_id', 'aadhaar_number', 'transaction_id', 'reference_id',
			'otp_requested_at', 'otp_verified_at', 'verification_status', 'kyc_status',
			'photo_url', 'name', 'date_of_birth', 'gender', 'full_address', 'address_json',
			'attempts', 'last_attempt_at', 'created_at', 'updated_at', 'deleted_at',
			'created_by', 'updated_by'
		  )`

	if err := db.WithContext(ctx).Raw(checkColumnsSQL).Count(&columnCount).Error; err != nil {
		return fmt.Errorf("failed to check columns: %w", err)
	}

	expectedColumns := int64(22)
	if columnCount != expectedColumns {
		return fmt.Errorf("column count mismatch: expected %d, found %d", expectedColumns, columnCount)
	}

	// Check if indexes exist
	var indexCount int64
	checkIndexesSQL := `
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE tablename = 'aadhaar_verifications'
		  AND indexname IN (
			'idx_aadhaar_verifications_user_id',
			'idx_aadhaar_verifications_transaction_id',
			'idx_aadhaar_verifications_reference_id',
			'idx_aadhaar_verifications_status'
		  )`

	if err := db.WithContext(ctx).Raw(checkIndexesSQL).Count(&indexCount).Error; err != nil {
		return fmt.Errorf("failed to check indexes: %w", err)
	}

	if indexCount < 4 {
		if logger != nil {
			logger.Warn("Some indexes may be missing",
				zap.Int64("found", indexCount),
				zap.Int("expected", 4))
		}
	}

	if logger != nil {
		logger.Info("aadhaar_verifications table validation completed",
			zap.Int64("columns", columnCount),
			zap.Int64("indexes", indexCount))
	}

	return nil
}
