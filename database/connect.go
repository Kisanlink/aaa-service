package database

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Kisanlink/aaa-service/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var err error

func buildDatabaseURL() string {
	host := config.GetEnv("DB_HOST")
	port := config.GetEnv("DB_PORT")
	user := config.GetEnv("POSTGRESS_USER")
	password := config.GetEnv("POSTGRESS_PASS")
	dbname := config.GetEnv("DB_NAME")
	sslMode := config.GetEnv("DB_SSL_MODE") // "disable" or "require"

	if sslMode == "" {
		sslMode = "disable" // default to disable if not set
	}

	// Handle cases where any required field is missing
	required := []string{host, port, user, dbname}
	for _, val := range required {
		if val == "" {
			panic("One or more required database environment variables are not set")
		}
	}

	// Build the connection string
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		user,
		password,
		dbname,
		sslMode,
	)
}

func ConnectDB() {
	databaseURL := buildDatabaseURL()
	fmt.Println("Connecting to database...")

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error), // Only log errors
	}

	// Handle timeouts
	timeout := config.GetEnv("DB_CONNECT_TIMEOUT")
	if timeout != "" {
		if timeoutSec, err := strconv.Atoi(timeout); err == nil {
			dsn := databaseURL + fmt.Sprintf(" connect_timeout=%d", timeoutSec)
			var connectErr error
			DB, connectErr = gorm.Open(postgres.Open(dsn), gormConfig)
			if connectErr != nil {
				panic(fmt.Sprintf("Error connecting to database: %v", connectErr))
			}
		}
	} else {
		DB, err = gorm.Open(postgres.Open(databaseURL), gormConfig)
	}

	if err != nil {
		panic(fmt.Sprintf("Error connecting to database: %v", err))
	}
	fmt.Println("Successfully connected to database")

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		panic(fmt.Sprintf("Failed to get underlying sql.DB: %v", err))
	}

	// Set connection pool parameters from env or use defaults
	maxOpenConns := GetEnvAsInt("DB_MAX_OPEN_CONNS", 100)
	maxIdleConns := GetEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	maxLifetime := GetEnvAsInt("DB_CONN_MAX_LIFETIME", 0) // 0 means unlimited

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)

	fmt.Printf("DB Connection pool configured - MaxOpen: %d, MaxIdle: %d, MaxLifetime: %ds\n",
		maxOpenConns, maxIdleConns, maxLifetime)

}

func GetEnvAsInt(key string, defaultValue int) int {
	val := config.GetEnv(key)
	if val == "" {
		return defaultValue
	}
	if intVal, err := strconv.Atoi(val); err == nil {
		return intVal
	}
	return defaultValue
}
