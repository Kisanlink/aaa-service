package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	PrimaryBackend db.BackendType
	Postgres       PostgresConfig
	DynamoDB       DynamoDBConfig
}

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxConns     int
	IdleConns    int
	ReadReplicas []string
}

// DynamoDBConfig holds DynamoDB configuration
type DynamoDBConfig struct {
	Region string
	Table  string
}

// LoadDatabaseConfig loads database configuration from environment variables
func LoadDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		PrimaryBackend: db.BackendType(getEnv("DB_PRIMARY_BACKEND", "gorm")),
		Postgres: PostgresConfig{
			Host:         getEnv("DB_POSTGRES_HOST", "localhost"),
			Port:         getEnv("DB_POSTGRES_PORT", "5432"),
			User:         getEnv("DB_POSTGRES_USER", "postgres"),
			Password:     getEnv("DB_POSTGRES_PASSWORD", ""),
			DBName:       getEnv("DB_POSTGRES_DBNAME", "kisanlink"),
			SSLMode:      getEnv("DB_POSTGRES_SSLMODE", "disable"),
			MaxConns:     getEnvAsInt("DB_POSTGRES_MAX_CONNS", 10),
			IdleConns:    getEnvAsInt("DB_POSTGRES_IDLE_CONNS", 5),
			ReadReplicas: getEnvAsSlice("DB_POSTGRES_READ_REPLICAS", ","),
		},
		DynamoDB: DynamoDBConfig{
			Region: getEnv("DB_DYNAMO_REGION", "us-east-1"),
			Table:  getEnv("DB_DYNAMO_TABLE", ""),
		},
	}
}

// NewDatabaseManager creates a new database manager with the loaded configuration
func NewDatabaseManager(logger *zap.Logger) (*db.DatabaseManager, error) {
	logger.Info("Loading database configuration")
	config := LoadDatabaseConfig()

	dbConfig := &db.Config{
		PrimaryBackend:       config.PrimaryBackend,
		PostgresHost:         config.Postgres.Host,
		PostgresPort:         config.Postgres.Port,
		PostgresUser:         config.Postgres.User,
		PostgresPassword:     config.Postgres.Password,
		PostgresDBName:       config.Postgres.DBName,
		PostgresSSLMode:      config.Postgres.SSLMode,
		PostgresMaxConns:     config.Postgres.MaxConns,
		PostgresIdleConns:    config.Postgres.IdleConns,
		PostgresReadReplicas: config.Postgres.ReadReplicas,
		DynamoDBRegion:       config.DynamoDB.Region,
		DynamoDBTable:        config.DynamoDB.Table,
		LogLevel:             getEnv("DB_LOG_LEVEL", "info"),
	}

	dm := db.NewDatabaseManagerWithConfig(dbConfig)

	// Connect to the database
	if err := dm.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", sanitizeError(err.Error()))
	}

	// Run automigration for all models if enabled
	if getEnv("AAA_AUTO_MIGRATE", "false") == "true" {
		if err := runAutomigration(dm, logger); err != nil {
			return nil, fmt.Errorf("failed to run automigration: %s", sanitizeError(err.Error()))
		}
	} else {
		logger.Info("Skipping automigration; AAA_AUTO_MIGRATE is not true")
	}

	logger.Info("Database manager initialized successfully")
	return dm, nil
}

// runAutomigration runs automigration for all models
func runAutomigration(dm *db.DatabaseManager, logger *zap.Logger) error {
	logger.Info("Starting automigration")

	// Import models from aaa-service
	allModels := []interface{}{
		// Core identity models
		&models.User{},
		&models.UserProfile{},
		&models.Contact{},
		&models.Address{},
		&models.Service{},
		&models.Principal{},

		// Organization and groups
		&models.Organization{},
		&models.Group{},
		&models.GroupMembership{},
		&models.GroupInheritance{},
		&models.GroupRole{},

		// Roles and permissions
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.Action{},
		&models.RolePermission{},     // Role-Permission mapping
		&models.ResourcePermission{}, // Resource-Role-Action mapping
		&models.ServiceRoleMapping{}, // Service-Role mapping for audit trail

		// Resources
		&models.Resource{},

		// Bindings
		&models.Binding{},
		&models.BindingHistory{},

		// Column-level authorization
		&models.ColumnGroup{},
		&models.ColumnGroupMember{},
		&models.ColumnSet{},

		// Attributes for ABAC
		&models.Attribute{},
		&models.AttributeHistory{},

		// Audit and events
		&models.AuditLog{},
		&models.Event{},
		&models.EventCheckpoint{},

		// Password reset and SMS
		&models.PasswordResetToken{},
		&models.SMSDeliveryLog{},
	}

	logger.Info("Models to migrate", zap.Int("count", len(allModels)))

	// Get the primary backend manager (assuming it's PostgreSQL/GORM)
	primaryManager := dm.GetManager(db.BackendGorm)
	if primaryManager == nil {
		logger.Warn("PostgreSQL manager not available, skipping automigration")
		return nil
	}

	logger.Info("Found PostgreSQL manager, running automigration")

	// Run automigration
	if err := primaryManager.AutoMigrateModels(context.Background(), allModels...); err != nil {
		logger.Error("Automigration failed", zap.Error(err))
		return fmt.Errorf("automigration failed: %w", err)
	}

	logger.Info("Automigration completed successfully")
	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := parseInt(value); err == nil {
			return intValue
		} else {
			fmt.Printf("Warning: Invalid integer value for %s: %s, using default %d\n", key, value, defaultValue)
		}
	}
	return defaultValue
}

func getEnvAsSlice(key, separator string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, separator)
	}
	return []string{}
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

func sanitizeError(errMsg string) string {
	return strings.ReplaceAll(strings.ReplaceAll(errMsg, "\n", " "), "\r", " ")
}
