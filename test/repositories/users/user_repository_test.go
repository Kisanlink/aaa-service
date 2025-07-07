package users

import (
	"context"
	"testing"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

func TestUserRepository_Create(t *testing.T) {
	for _, tt := range UserRepositoryCreateTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			repo := NewUserRepository(dbManager)
			ctx := context.Background()

			// Create user
			user := models.NewUser(tt.name, tt.email)
			user.Phone = tt.phone
			user.Status = tt.status

			createdUser, err := repo.Create(ctx, user)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				// Verify user was created
				if createdUser == nil {
					t.Fatal("Created user is nil")
				}

				if createdUser.ID == "" {
					t.Error("User ID should not be empty")
				}

				if createdUser.Name != tt.name {
					t.Errorf("Expected name %s, got %s", tt.name, createdUser.Name)
				}

				if createdUser.Email != tt.email {
					t.Errorf("Expected email %s, got %s", tt.email, createdUser.Email)
				}

				if createdUser.Phone != tt.phone {
					t.Errorf("Expected phone %s, got %s", tt.phone, createdUser.Phone)
				}

				if createdUser.Status != tt.status {
					t.Errorf("Expected status %s, got %s", tt.status, createdUser.Status)
				}

				// Verify timestamps
				if createdUser.CreatedAt.IsZero() {
					t.Error("CreatedAt should not be zero")
				}

				if createdUser.UpdatedAt.IsZero() {
					t.Error("UpdatedAt should not be zero")
				}
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	for _, tt := range UserRepositoryGetByIDTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			repo := NewUserRepository(dbManager)
			ctx := context.Background()

			// Create a test user first
			user := models.NewUser("Test User", "test@example.com")
			createdUser, err := repo.Create(ctx, user)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Test GetByID
			foundUser, err := repo.GetByID(ctx, tt.userID)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				if foundUser == nil {
					t.Fatal("Found user is nil")
				}

				if foundUser.ID != tt.userID {
					t.Errorf("Expected user ID %s, got %s", tt.userID, foundUser.ID)
				}
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	for _, tt := range UserRepositoryGetByEmailTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			repo := NewUserRepository(dbManager)
			ctx := context.Background()

			// Create a test user first
			user := models.NewUser("Test User", tt.email)
			_, err := repo.Create(ctx, user)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Test GetByEmail
			foundUser, err := repo.GetByEmail(ctx, tt.searchEmail)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				if foundUser == nil {
					t.Fatal("Found user is nil")
				}

				if foundUser.Email != tt.searchEmail {
					t.Errorf("Expected email %s, got %s", tt.searchEmail, foundUser.Email)
				}
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	for _, tt := range UserRepositoryUpdateTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			repo := NewUserRepository(dbManager)
			ctx := context.Background()

			// Create a test user first
			user := models.NewUser("Original Name", "original@example.com")
			createdUser, err := repo.Create(ctx, user)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Update user
			createdUser.Name = tt.newName
			createdUser.Email = tt.newEmail
			createdUser.Phone = tt.newPhone
			createdUser.Status = tt.newStatus

			updatedUser, err := repo.Update(ctx, createdUser)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				if updatedUser == nil {
					t.Fatal("Updated user is nil")
				}

				if updatedUser.Name != tt.newName {
					t.Errorf("Expected name %s, got %s", tt.newName, updatedUser.Name)
				}

				if updatedUser.Email != tt.newEmail {
					t.Errorf("Expected email %s, got %s", tt.newEmail, updatedUser.Email)
				}

				if updatedUser.Phone != tt.newPhone {
					t.Errorf("Expected phone %s, got %s", tt.newPhone, updatedUser.Phone)
				}

				if updatedUser.Status != tt.newStatus {
					t.Errorf("Expected status %s, got %s", tt.newStatus, updatedUser.Status)
				}

				// Verify UpdatedAt was changed
				if updatedUser.UpdatedAt.Equal(createdUser.UpdatedAt) {
					t.Error("UpdatedAt should be different after update")
				}
			}
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	for _, tt := range UserRepositoryDeleteTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			repo := NewUserRepository(dbManager)
			ctx := context.Background()

			// Create a test user first
			user := models.NewUser("Test User", "test@example.com")
			createdUser, err := repo.Create(ctx, user)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Test Delete
			err = repo.Delete(ctx, tt.userID)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				// Verify user is soft deleted
				foundUser, err := repo.GetByID(ctx, tt.userID)
				if err == nil && foundUser != nil && !foundUser.IsDeleted() {
					t.Error("User should be soft deleted")
				}
			}
		})
	}
}

func TestUserRepository_List(t *testing.T) {
	for _, tt := range UserRepositoryListTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			repo := NewUserRepository(dbManager)
			ctx := context.Background()

			// Create test users
			for _, userData := range tt.testUsers {
				user := models.NewUser(userData.name, userData.email)
				user.Status = userData.status
				_, err := repo.Create(ctx, user)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
			}

			// Test List
			users, err := repo.List(ctx, tt.filters)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				if users == nil {
					t.Fatal("Users list is nil")
				}

				if len(users) != tt.expectedCount {
					t.Errorf("Expected %d users, got %d", tt.expectedCount, len(users))
				}
			}
		})
	}
}

// Helper functions for test setup
func setupTestDatabase(t *testing.T) *db.Manager {
	// This would set up a test database (e.g., in-memory SQLite or test PostgreSQL)
	// For now, we'll return nil and skip tests that require database
	t.Skip("Database setup not implemented yet")
	return nil
}

func cleanupTestDatabase(t *testing.T, dbManager *db.Manager) {
	// Clean up test database
	if dbManager != nil {
		// Cleanup logic here
	}
}
