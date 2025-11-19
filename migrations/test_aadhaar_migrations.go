package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TestAadhaarMigrations runs all Aadhaar-related migrations in sequence and validates them.
// This is a test helper function to ensure migrations work correctly.
func TestAadhaarMigrations(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Info("========================================")
		logger.Info("Testing Aadhaar Integration Migrations")
		logger.Info("========================================")
	}

	// Step 1: Run all migrations in correct order
	if logger != nil {
		logger.Info("Step 1: Running migrations...")
	}

	// Migration 1: Create aadhaar_verifications table
	if err := AddAadhaarVerificationsTable(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to create aadhaar_verifications table: %w", err)
	}

	// Migration 2: Create otp_attempts table (depends on aadhaar_verifications)
	if err := AddOTPAttemptsTable(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to create otp_attempts table: %w", err)
	}

	// Migration 3: Update user_profiles table
	if err := AddUserProfilesKYCFields(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to add KYC fields to user_profiles: %w", err)
	}

	if logger != nil {
		logger.Info("✅ All migrations completed successfully")
	}

	// Step 2: Validate all migrations
	if logger != nil {
		logger.Info("")
		logger.Info("Step 2: Validating migrations...")
	}

	if err := ValidateAadhaarVerificationsTable(ctx, db, logger); err != nil {
		return fmt.Errorf("aadhaar_verifications table validation failed: %w", err)
	}

	if err := ValidateOTPAttemptsTable(ctx, db, logger); err != nil {
		return fmt.Errorf("otp_attempts table validation failed: %w", err)
	}

	if err := ValidateUserProfilesKYCFields(ctx, db, logger); err != nil {
		return fmt.Errorf("user_profiles KYC fields validation failed: %w", err)
	}

	if logger != nil {
		logger.Info("✅ All validations passed")
	}

	// Step 3: Test rollback capability (optional, commented out for safety)
	if logger != nil {
		logger.Info("")
		logger.Info("Step 3: Rollback test skipped (run manually if needed)")
		logger.Info("To test rollback, call: TestAadhaarMigrationsRollback()")
	}

	if logger != nil {
		logger.Info("")
		logger.Info("========================================")
		logger.Info("Migration Test Completed Successfully")
		logger.Info("========================================")
	}

	return nil
}

// TestAadhaarMigrationsRollback tests the rollback functionality of all Aadhaar migrations.
// WARNING: This will drop all Aadhaar-related tables and columns. Use only in test environments.
func TestAadhaarMigrationsRollback(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if logger != nil {
		logger.Warn("========================================")
		logger.Warn("Rolling Back Aadhaar Migrations (DESTRUCTIVE)")
		logger.Warn("========================================")
	}

	// Rollback in reverse order of creation
	// Step 1: Drop user_profiles KYC fields
	if err := DropUserProfilesKYCFields(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to drop user_profiles KYC fields: %w", err)
	}

	// Step 2: Drop otp_attempts table
	if err := DropOTPAttemptsTable(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to drop otp_attempts table: %w", err)
	}

	// Step 3: Drop aadhaar_verifications table
	if err := DropAadhaarVerificationsTable(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to drop aadhaar_verifications table: %w", err)
	}

	if logger != nil {
		logger.Info("✅ All rollbacks completed successfully")
		logger.Warn("========================================")
	}

	return nil
}

// GetMigrationSummary returns a summary of the Aadhaar migration status.
func GetMigrationSummary(ctx context.Context, db *gorm.DB, logger *zap.Logger) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Check aadhaar_verifications table
	var aadhaarTableExists bool
	if err := db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'aadhaar_verifications'
		)`).Scan(&aadhaarTableExists).Error; err != nil {
		return nil, fmt.Errorf("failed to check aadhaar_verifications: %w", err)
	}
	summary["aadhaar_verifications_table"] = aadhaarTableExists

	// Check otp_attempts table
	var otpTableExists bool
	if err := db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'otp_attempts'
		)`).Scan(&otpTableExists).Error; err != nil {
		return nil, fmt.Errorf("failed to check otp_attempts: %w", err)
	}
	summary["otp_attempts_table"] = otpTableExists

	// Check user_profiles KYC fields
	var kycColumnCount int64
	if err := db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'user_profiles'
		  AND column_name IN ('aadhaar_verified', 'aadhaar_verified_at', 'kyc_status')
	`).Count(&kycColumnCount).Error; err != nil {
		return nil, fmt.Errorf("failed to check KYC columns: %w", err)
	}
	summary["user_profiles_kyc_columns"] = kycColumnCount == 3

	// Count total indexes created
	var indexCount int64
	if err := db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE tablename IN ('aadhaar_verifications', 'otp_attempts', 'user_profiles')
		  AND (
			indexname LIKE 'idx_aadhaar_verifications_%' OR
			indexname LIKE 'idx_otp_attempts_%' OR
			indexname IN ('idx_user_profiles_kyc_status', 'idx_user_profiles_aadhaar_verified')
		  )
	`).Count(&indexCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count indexes: %w", err)
	}
	summary["total_indexes"] = indexCount

	// Overall status
	allMigrated := aadhaarTableExists && otpTableExists && kycColumnCount == 3
	summary["status"] = map[string]interface{}{
		"migrated": allMigrated,
		"message":  getMigrationStatusMessage(allMigrated),
	}

	if logger != nil {
		logger.Info("Migration Summary",
			zap.Bool("aadhaar_verifications", aadhaarTableExists),
			zap.Bool("otp_attempts", otpTableExists),
			zap.Bool("user_profiles_kyc", kycColumnCount == 3),
			zap.Int64("indexes", indexCount),
			zap.Bool("all_migrated", allMigrated))
	}

	return summary, nil
}

func getMigrationStatusMessage(allMigrated bool) string {
	if allMigrated {
		return "All Aadhaar integration migrations are applied successfully"
	}
	return "Some Aadhaar integration migrations are missing or incomplete"
}
