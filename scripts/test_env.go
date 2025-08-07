package main

import (
	"fmt"
	"os"

	"github.com/Kisanlink/aaa-service/config"
)

func main() {
	fmt.Println("Testing environment variable loading...")

	// Print all relevant environment variables
	envVars := []string{
		"DB_PRIMARY_BACKEND",
		"DB_POSTGRES_HOST",
		"DB_POSTGRES_PORT",
		"DB_POSTGRES_USER",
		"DB_POSTGRES_PASSWORD",
		"DB_POSTGRES_DBNAME",
		"DB_POSTGRES_SSLMODE",
		"DB_POSTGRES_MAX_CONNS",
		"DB_POSTGRES_IDLE_CONNS",
	}

	fmt.Println("Environment variables:")
	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		if value == "" {
			fmt.Printf("  %s: <not set>\n", envVar)
		} else {
			fmt.Printf("  %s: %s\n", envVar, value)
		}
	}

	fmt.Println("\nLoading database configuration...")
	config := config.LoadDatabaseConfig()

	fmt.Printf("Loaded config:\n")
	fmt.Printf("  PrimaryBackend: %s\n", config.PrimaryBackend)
	fmt.Printf("  Postgres.Host: %s\n", config.Postgres.Host)
	fmt.Printf("  Postgres.Port: %s\n", config.Postgres.Port)
	fmt.Printf("  Postgres.User: %s\n", config.Postgres.User)
	fmt.Printf("  Postgres.DBName: %s\n", config.Postgres.DBName)
	fmt.Printf("  Postgres.SSLMode: %s\n", config.Postgres.SSLMode)
}
