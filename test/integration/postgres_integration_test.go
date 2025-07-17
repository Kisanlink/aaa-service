//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

func TestPostgres_Connection(t *testing.T) {
	cfg := &db.Config{
		PostgresHost:     os.Getenv("DB_POSTGRES_HOST"),
		PostgresPort:     os.Getenv("DB_POSTGRES_PORT"),
		PostgresUser:     os.Getenv("DB_POSTGRES_USER"),
		PostgresPassword: os.Getenv("DB_POSTGRES_PASSWORD"),
		PostgresDBName:   os.Getenv("DB_POSTGRES_DBNAME"),
		PostgresSSLMode:  os.Getenv("DB_POSTGRES_SSLMODE"),
	}

	manager := db.NewDatabaseManagerWithConfig(cfg)
	if err := manager.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer manager.Close()

	if !manager.IsConnected() {
		t.Fatal("Database manager should be connected")
	}
}

func TestPostgres_UserCRUD(t *testing.T) {
	cfg := &db.Config{
		PostgresHost:     os.Getenv("DB_POSTGRES_HOST"),
		PostgresPort:     os.Getenv("DB_POSTGRES_PORT"),
		PostgresUser:     os.Getenv("DB_POSTGRES_USER"),
		PostgresPassword: os.Getenv("DB_POSTGRES_PASSWORD"),
		PostgresDBName:   os.Getenv("DB_POSTGRES_DBNAME"),
		PostgresSSLMode:  os.Getenv("DB_POSTGRES_SSLMODE"),
	}

	manager := db.NewDatabaseManagerWithConfig(cfg)
	if err := manager.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer manager.Close()

	// Test Create
	user := models.NewUser("integration_test_user", "password123")
	if err := manager.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test GetByID
	retrievedUser := &models.User{}
	if err := manager.GetByID(context.Background(), user.ID, retrievedUser); err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if retrievedUser.Username != "integration_test_user" {
		t.Errorf("Expected username 'integration_test_user', got %s", retrievedUser.Username)
	}

	// Test Update
	retrievedUser.Username = "updated_integration_user"
	if err := manager.Update(context.Background(), retrievedUser); err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Test Delete
	if err := manager.Delete(context.Background(), user.ID); err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
}
