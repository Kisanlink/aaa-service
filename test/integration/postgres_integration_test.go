//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func TestPostgres_Connection(t *testing.T) {
	cfg := &db.Config{
		PostgresHost:     getEnv("DB_POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("DB_POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("DB_POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("DB_POSTGRES_PASSWORD", "password"),
		PostgresDBName:   getEnv("DB_POSTGRES_DBNAME", "kisanlink_test"),
		PostgresSSLMode:  getEnv("DB_POSTGRES_SSLMODE", "disable"),
	}

	manager := db.NewDatabaseManagerWithConfig(cfg)
	if err := manager.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer manager.Close()

	if !manager.IsConnected(db.BackendGorm) {
		t.Fatal("Database manager should be connected")
	}
}

func TestPostgres_UserCRUD(t *testing.T) {
	cfg := &db.Config{
		PostgresHost:     getEnv("DB_POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("DB_POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("DB_POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("DB_POSTGRES_PASSWORD", "password"),
		PostgresDBName:   getEnv("DB_POSTGRES_DBNAME", "kisanlink_test"),
		PostgresSSLMode:  getEnv("DB_POSTGRES_SSLMODE", "disable"),
	}

	manager := db.NewDatabaseManagerWithConfig(cfg)
	if err := manager.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer manager.Close()

	postgresManager := manager.GetManager(db.BackendGorm)
	if postgresManager == nil {
		t.Fatal("Postgres manager is not available")
	}

	// Test Create
	user := models.NewUser("integration_test_user", "password123")
	if err := postgresManager.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test GetByID
	retrievedUser := &models.User{}
	if err := postgresManager.GetByID(context.Background(), user.ID, retrievedUser); err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if retrievedUser.Username != "integration_test_user" {
		t.Errorf("Expected username 'integration_test_user', got %s", retrievedUser.Username)
	}

	// Test Update
	retrievedUser.Username = "updated_integration_user"
	if err := postgresManager.Update(context.Background(), retrievedUser); err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Test Delete
	if err := postgresManager.Delete(context.Background(), user.ID); err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
}
