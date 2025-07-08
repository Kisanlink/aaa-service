package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kisanlink/aaa-service/config"
	"github.com/Kisanlink/aaa-service/repositories/addresses"
	"github.com/Kisanlink/aaa-service/repositories/roles"
	"github.com/Kisanlink/aaa-service/repositories/users"
	"github.com/Kisanlink/aaa-service/server"
	"github.com/Kisanlink/aaa-service/services"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting AAA Service")

	// Initialize database manager
	dbManager, err := config.NewDatabaseManager(logger)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}
	defer dbManager.Close()

	// Initialize cache service
	cacheService, err := services.NewCacheService(
		"localhost:6379", // Redis address
		"",               // Redis password
		0,                // Redis database
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to initialize cache service", zap.Error(err))
	}
	defer cacheService.Close()

	// Initialize repositories
	userRepo := users.NewUserRepository(dbManager.GetPostgresManager())
	addressRepo := addresses.NewAddressRepository(dbManager.GetPostgresManager())
	roleRepo := roles.NewRoleRepository(dbManager.GetPostgresManager())
	userRoleRepo := roles.NewUserRoleRepository(dbManager.GetPostgresManager())

	// Initialize HTTP server
	httpServer, err := server.NewHTTPServer(
		logger,
		dbManager,
		cacheService,
		userRepo,
		addressRepo,
		roleRepo,
		userRoleRepo,
	)
	if err != nil {
		logger.Fatal("Failed to initialize HTTP server", zap.Error(err))
	}

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server on :8080")
		if err := httpServer.Start(); err != nil {
			logger.Error("HTTP server error", zap.Error(err))
			cancel()
		}
	}()

	// Wait for shutdown signal
	select {
	case <-sigChan:
		logger.Info("Received shutdown signal, starting graceful shutdown...")
	case <-ctx.Done():
		logger.Info("Context cancelled, starting graceful shutdown...")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30)
	defer shutdownCancel()

	logger.Info("Shutting down HTTP server...")
	if err := httpServer.Stop(shutdownCtx); err != nil {
		logger.Error("Error during HTTP server shutdown", zap.Error(err))
	}

	logger.Info("AAA Service shutdown complete")
}

// Example usage function for testing the refactored code
func runExample(ctx context.Context, userRepo *users.UserRepository, addressRepo *addresses.AddressRepository, roleRepo *roles.RoleRepository, userRoleRepo *roles.UserRoleRepository, logger *zap.Logger) error {
	logger.Info("Starting example usage of refactored AAA service")

	// Create a new user
	user := models.NewUser("testuser", "password123", 9876543210)
	user.Name = &[]string{"Test User"}[0]
	user.CountryCode = &[]string{"+91"}[0]

	if err := userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	logger.Info("Created user", zap.String("userID", user.ID), zap.String("username", user.Username))

	// Create a new address
	address := models.NewAddress()
	address.House = &[]string{"123"}[0]
	address.Street = &[]string{"Test Street"}[0]
	address.District = &[]string{"Test District"}[0]
	address.State = &[]string{"Test State"}[0]
	address.Country = &[]string{"India"}[0]
	address.Pincode = &[]string{"123456"}[0]
	address.BuildFullAddress()

	if err := addressRepo.Create(ctx, address); err != nil {
		return fmt.Errorf("failed to create address: %w", err)
	}
	logger.Info("Created address", zap.String("addressID", address.ID))

	// Update user with address
	user.AddressID = &address.ID
	if err := userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user with address: %w", err)
	}

	// Create a new role
	role := models.NewRole("user", "Basic user role")
	if err := roleRepo.Create(ctx, role); err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	logger.Info("Created role", zap.String("roleID", role.ID), zap.String("roleName", role.Name))

	// Assign role to user
	userRole := models.NewUserRole(user.ID, role.ID)
	if err := userRoleRepo.Create(ctx, userRole); err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}
	logger.Info("Assigned role to user", zap.String("userRoleID", userRole.ID))

	// Retrieve user with relationships
	retrievedUser, err := userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// Convert to response format
	userResponse := retrievedUser.ToResponse()
	logger.Info("Retrieved user",
		zap.String("userID", userResponse.ID),
		zap.String("username", userResponse.Username),
		zap.Bool("isValidated", userResponse.IsValidated),
		zap.Int("tokens", userResponse.Tokens),
	)

	// Search for users
	users, err := userRepo.ListActive(ctx, 10, 0)
	if err != nil {
		return fmt.Errorf("failed to list active users: %w", err)
	}
	logger.Info("Found active users", zap.Int("count", len(users)))

	// Search for addresses
	addresses, err := addressRepo.SearchByKeyword(ctx, "Test", 10, 0)
	if err != nil {
		return fmt.Errorf("failed to search addresses: %w", err)
	}
	logger.Info("Found addresses", zap.Int("count", len(addresses)))

	logger.Info("Example completed successfully")
	return nil
}
