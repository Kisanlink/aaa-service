package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Kisanlink/aaa-service/v2/migrations"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize logger
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// Get database configuration
	dbHost := getEnv("DB_POSTGRES_HOST", "localhost")
	dbPort := getEnv("DB_POSTGRES_PORT", "5432")
	dbUser := getEnv("DB_POSTGRES_USER", "postgres")
	dbPassword := getEnv("DB_POSTGRES_PASSWORD", "")
	dbName := getEnv("DB_POSTGRES_DBNAME", "aaa_service")

	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database
	zapLogger.Info("Connecting to PostgreSQL database",
		zap.String("host", dbHost),
		zap.String("port", dbPort),
		zap.String("database", dbName))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Get underlying SQL DB for connection management
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Fatal("Failed to get SQL DB", zap.Error(err))
	}
	defer sqlDB.Close()

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx := context.Background()
	if err := sqlDB.PingContext(ctx); err != nil {
		zapLogger.Fatal("Failed to ping database", zap.Error(err))
	}

	zapLogger.Info("Successfully connected to database")

	// Parse command line arguments
	action := "migrate"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	switch action {
	case "migrate", "up":
		// Run migrations
		if err := migrations.TestAadhaarMigrations(ctx, db, zapLogger); err != nil {
			zapLogger.Fatal("Migration failed", zap.Error(err))
		}
		zapLogger.Info("✅ All migrations completed successfully!")

	case "rollback", "down":
		// Rollback migrations
		zapLogger.Warn("⚠️  WARNING: This will drop all Aadhaar-related tables and columns!")
		zapLogger.Warn("⚠️  Press Ctrl+C within 5 seconds to cancel...")
		time.Sleep(5 * time.Second)

		if err := migrations.TestAadhaarMigrationsRollback(ctx, db, zapLogger); err != nil {
			zapLogger.Fatal("Rollback failed", zap.Error(err))
		}
		zapLogger.Info("✅ All rollbacks completed successfully!")

	case "status", "summary":
		// Get migration status
		summary, err := migrations.GetMigrationSummary(ctx, db, zapLogger)
		if err != nil {
			zapLogger.Fatal("Failed to get migration summary", zap.Error(err))
		}

		zapLogger.Info("========================================")
		zapLogger.Info("Aadhaar Migration Status")
		zapLogger.Info("========================================")
		for key, value := range summary {
			zapLogger.Info(fmt.Sprintf("%s: %v", key, value))
		}

	case "validate":
		// Validate migrations
		zapLogger.Info("Validating Aadhaar migrations...")

		if err := migrations.ValidateAadhaarVerificationsTable(ctx, db, zapLogger); err != nil {
			zapLogger.Error("Validation failed for aadhaar_verifications", zap.Error(err))
		}

		if err := migrations.ValidateOTPAttemptsTable(ctx, db, zapLogger); err != nil {
			zapLogger.Error("Validation failed for otp_attempts", zap.Error(err))
		}

		if err := migrations.ValidateUserProfilesKYCFields(ctx, db, zapLogger); err != nil {
			zapLogger.Error("Validation failed for user_profiles KYC fields", zap.Error(err))
		}

		zapLogger.Info("✅ Validation completed!")

	default:
		printUsage()
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func printUsage() {
	fmt.Println("Aadhaar Integration Migrations")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println("Usage: go run scripts/run_aadhaar_migrations.go [action]")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  migrate, up      - Run all Aadhaar migrations (default)")
	fmt.Println("  rollback, down   - Rollback all Aadhaar migrations (DESTRUCTIVE)")
	fmt.Println("  status, summary  - Show current migration status")
	fmt.Println("  validate         - Validate existing migrations")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run scripts/run_aadhaar_migrations.go migrate")
	fmt.Println("  go run scripts/run_aadhaar_migrations.go status")
	fmt.Println("  go run scripts/run_aadhaar_migrations.go validate")
	fmt.Println("  go run scripts/run_aadhaar_migrations.go rollback")
	fmt.Println()
}
