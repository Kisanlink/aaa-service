package main

import (
	"context"
	"log"

	"github.com/Kisanlink/aaa-service/config"
	"github.com/Kisanlink/aaa-service/entities/models"
	"go.uber.org/zap"
)

func main() {
	// Set up logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Initialize database manager
	dbManager, err := config.NewDatabaseManager(logger)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}
	defer func() {
		if err := dbManager.Close(); err != nil {
			log.Printf("Failed to close database manager: %v", err)
		}
	}()

	logger.Info("Database manager initialized successfully")

	// Test creating a user to see if tables exist
	user := models.NewUser("1234567890", "+91", "testpassword")

	// Get the primary database manager
	primaryManager := dbManager.GetManager(dbManager.GetPostgresManager().GetBackendType())
	if primaryManager == nil {
		logger.Fatal("No database manager available")
	}

	// Try to create the user
	if err := primaryManager.Create(context.Background(), user); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
	} else {
		logger.Info("Successfully created user", zap.String("user_id", user.ID))
	}
}
