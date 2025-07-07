package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	PrimaryBackend db.BackendType
	Postgres       PostgresConfig
	DynamoDB       DynamoDBConfig
	SpiceDB        SpiceDBConfig
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

// SpiceDBConfig holds SpiceDB configuration
type SpiceDBConfig struct {
	Endpoint string
	Token    string
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
		SpiceDB: SpiceDBConfig{
			Endpoint: getEnv("DB_SPICEDB_ENDPOINT", ""),
			Token:    getEnv("DB_SPICEDB_TOKEN", ""),
		},
	}
}

// NewDatabaseManager creates a new database manager with the loaded configuration
func NewDatabaseManager(logger *zap.Logger) (*db.DatabaseManager, error) {
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
		SpiceDBEndpoint:      config.SpiceDB.Endpoint,
		SpiceDBToken:         config.SpiceDB.Token,
		LogLevel:             getEnv("DB_LOG_LEVEL", "info"),
	}

	dm := db.NewDatabaseManagerWithConfig(dbConfig)

	// Connect to the database
	if err := dm.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return dm, nil
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
