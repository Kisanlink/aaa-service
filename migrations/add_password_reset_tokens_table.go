package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AddPasswordResetTokensTable creates the password_reset_tokens table with all required indexes.
func AddPasswordResetTokensTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Starting password_reset_tokens table creation migration")
	}

	// Step 1: Check if table already exists
	var tableExists bool
	checkTableSQL := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'password_reset_tokens'
		)`

	if err := db.WithContext(ctx).Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to check if password_reset_tokens table exists", zap.Error(err))
		}
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists {
		if logger != nil {
			logger.Info("password_reset_tokens table already exists, skipping creation")
		}
		return nil
	}

	// Step 2: Create the password_reset_tokens table
	createTableSQL := `
CREATE TABLE password_reset_tokens (
	id VARCHAR(255) PRIMARY KEY,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	created_by VARCHAR(255) DEFAULT '',
	updated_by VARCHAR(255) DEFAULT '',
	deleted_at TIMESTAMP,
	deleted_by VARCHAR(255),
	user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token VARCHAR(255) NOT NULL,
	expires_at TIMESTAMP NOT NULL,
	used BOOLEAN DEFAULT FALSE,
	used_at TIMESTAMP
)`

	if logger != nil {
		logger.Info("Creating password_reset_tokens table")
	}

	if err := db.WithContext(ctx).Exec(createTableSQL).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to create password_reset_tokens table", zap.Error(err))
		}
		return fmt.Errorf("failed to create password_reset_tokens table: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully created password_reset_tokens table")
	}

	// Step 3: Create indexes for performance optimization
	indexes := []struct {
		name        string
		description string
		sql         string
	}{
		{
			name:        "idx_password_reset_tokens_user_id",
			description: "Optimize queries by user_id",
			sql:         `CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id)`,
		},
		{
			name:        "idx_password_reset_tokens_token",
			description: "Unique index on token for fast lookup",
			sql:         `CREATE UNIQUE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token)`,
		},
		{
			name:        "idx_password_reset_tokens_expires_at",
			description: "Index on expires_at for cleanup queries",
			sql:         `CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at)`,
		},
		{
			name:        "idx_password_reset_tokens_used",
			description: "Index on used flag for filtering unused tokens",
			sql:         `CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_used ON password_reset_tokens(used) WHERE used = false`,
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
		logger.Info("password_reset_tokens table migration completed",
			zap.Int("total_indexes", len(indexes)),
			zap.Int("created", successCount),
			zap.Int("failed", failureCount))
	}

	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("failed to create all indexes: %d failures", failureCount)
	}

	return nil
}

// DropPasswordResetTokensTable removes the password_reset_tokens table (rollback).
func DropPasswordResetTokensTable(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("Rolling back password_reset_tokens table migration")
	}

	// Step 1: Drop all indexes first
	indexNames := []string{
		"idx_password_reset_tokens_user_id",
		"idx_password_reset_tokens_token",
		"idx_password_reset_tokens_expires_at",
		"idx_password_reset_tokens_used",
	}

	for _, indexName := range indexNames {
		dropIndexSQL := fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)
		if err := db.WithContext(ctx).Exec(dropIndexSQL).Error; err != nil {
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

	// Step 2: Drop the table
	dropTableSQL := "DROP TABLE IF EXISTS password_reset_tokens CASCADE"

	if logger != nil {
		logger.Info("Dropping password_reset_tokens table")
	}

	if err := db.WithContext(ctx).Exec(dropTableSQL).Error; err != nil {
		if logger != nil {
			logger.Error("Failed to drop password_reset_tokens table", zap.Error(err))
		}
		return fmt.Errorf("failed to drop password_reset_tokens table: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully rolled back password_reset_tokens table migration")
	}

	return nil
}
