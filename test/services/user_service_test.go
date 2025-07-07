package services

import (
	"context"
	"testing"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/entities/requests/users"
	"github.com/Kisanlink/aaa-service/entities/responses/users"
	"github.com/Kisanlink/aaa-service/repositories/users"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

func TestUserService_CreateUser(t *testing.T) {
	for _, tt := range UserServiceCreateUserTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database and repositories
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userRepo := users.NewUserRepository(dbManager)
			userService := NewUserService(userRepo)
			ctx := context.Background()

			// Create request
			req := users.NewCreateUserRequest(
				tt.name,
				tt.email,
				tt.phone,
				tt.status,
				"test-protocol",
				"test-operation",
				"v1",
				"test-request-id",
				map[string][]string{},
				nil,
				map[string]interface{}{},
			)

			// Test CreateUser
			response, err := userService.CreateUser(ctx, req)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				if response == nil {
					t.Fatal("Response is nil")
				}

				if !response.IsSuccess() {
					t.Error("Response should indicate success")
				}

				if response.GetType() != "UserCreated" {
					t.Errorf("Expected response type UserCreated, got %s", response.GetType())
				}

				// Verify response body
				userResponse, ok := response.GetBody().(*responses.UserResponse)
				if !ok {
					t.Fatal("Response body is not UserResponse")
				}

				if userResponse.Name != tt.name {
					t.Errorf("Expected name %s, got %s", tt.name, userResponse.Name)
				}

				if userResponse.Email != tt.email {
					t.Errorf("Expected email %s, got %s", tt.email, userResponse.Email)
				}

				if userResponse.Phone != tt.phone {
					t.Errorf("Expected phone %s, got %s", tt.phone, userResponse.Phone)
				}

				if userResponse.Status != tt.status {
					t.Errorf("Expected status %s, got %s", tt.status, userResponse.Status)
				}
			}
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	for _, tt := range UserServiceGetUserByIDTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database and repositories
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userRepo := users.NewUserRepository(dbManager)
			userService := NewUserService(userRepo)
			ctx := context.Background()

			// Create a test user first
			if tt.createTestUser {
				testUser := models.NewUser("Test User", "test@example.com")
				_, err := userRepo.Create(ctx, testUser)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
			}

			// Test GetUserByID
			response, err := userService.GetUserByID(ctx, tt.userID)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				if response == nil {
					t.Fatal("Response is nil")
				}

				if !response.IsSuccess() {
					t.Error("Response should indicate success")
				}

				if response.GetType() != "UserRetrieved" {
					t.Errorf("Expected response type UserRetrieved, got %s", response.GetType())
				}

				// Verify response body
				userResponse, ok := response.GetBody().(*responses.UserResponse)
				if !ok {
					t.Fatal("Response body is not UserResponse")
				}

				if userResponse.ID != tt.userID {
					t.Errorf("Expected user ID %s, got %s", tt.userID, userResponse.ID)
				}
			}
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	for _, tt := range UserServiceUpdateUserTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database and repositories
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userRepo := users.NewUserRepository(dbManager)
			userService := NewUserService(userRepo)
			ctx := context.Background()

			// Create a test user first
			testUser := models.NewUser("Original Name", "original@example.com")
			createdUser, err := userRepo.Create(ctx, testUser)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Create update request
			req := users.NewUpdateUserRequest(
				createdUser.ID,
				tt.newName,
				tt.newEmail,
				tt.newPhone,
				tt.newStatus,
				"test-protocol",
				"test-operation",
				"v1",
				"test-request-id",
				map[string][]string{},
				nil,
				map[string]interface{}{},
			)

			// Test UpdateUser
			response, err := userService.UpdateUser(ctx, req)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				if response == nil {
					t.Fatal("Response is nil")
				}

				if !response.IsSuccess() {
					t.Error("Response should indicate success")
				}

				if response.GetType() != "UserUpdated" {
					t.Errorf("Expected response type UserUpdated, got %s", response.GetType())
				}

				// Verify response body
				userResponse, ok := response.GetBody().(*responses.UserResponse)
				if !ok {
					t.Fatal("Response body is not UserResponse")
				}

				if userResponse.Name != tt.newName {
					t.Errorf("Expected name %s, got %s", tt.newName, userResponse.Name)
				}

				if userResponse.Email != tt.newEmail {
					t.Errorf("Expected email %s, got %s", tt.newEmail, userResponse.Email)
				}

				if userResponse.Phone != tt.newPhone {
					t.Errorf("Expected phone %s, got %s", tt.newPhone, userResponse.Phone)
				}

				if userResponse.Status != tt.newStatus {
					t.Errorf("Expected status %s, got %s", tt.newStatus, userResponse.Status)
				}
			}
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	for _, tt := range UserServiceDeleteUserTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database and repositories
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userRepo := users.NewUserRepository(dbManager)
			userService := NewUserService(userRepo)
			ctx := context.Background()

			// Create a test user first
			testUser := models.NewUser("Test User", "test@example.com")
			createdUser, err := userRepo.Create(ctx, testUser)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Test DeleteUser
			response, err := userService.DeleteUser(ctx, tt.userID)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				if response == nil {
					t.Fatal("Response is nil")
				}

				if !response.IsSuccess() {
					t.Error("Response should indicate success")
				}

				if response.GetType() != "UserDeleted" {
					t.Errorf("Expected response type UserDeleted, got %s", response.GetType())
				}
			}
		})
	}
}

func TestUserService_ListUsers(t *testing.T) {
	for _, tt := range UserServiceListUsersTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database and repositories
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userRepo := users.NewUserRepository(dbManager)
			userService := NewUserService(userRepo)
			ctx := context.Background()

			// Create test users
			for _, userData := range tt.testUsers {
				user := models.NewUser(userData.name, userData.email)
				user.Status = userData.status
				_, err := userRepo.Create(ctx, user)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
			}

			// Test ListUsers
			response, err := userService.ListUsers(ctx, tt.filters)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.shouldError {
				if response == nil {
					t.Fatal("Response is nil")
				}

				if !response.IsSuccess() {
					t.Error("Response should indicate success")
				}

				if response.GetType() != "UsersListed" {
					t.Errorf("Expected response type UsersListed, got %s", response.GetType())
				}

				// Verify response body
				usersResponse, ok := response.GetBody().(*responses.UsersListResponse)
				if !ok {
					t.Fatal("Response body is not UsersListResponse")
				}

				if len(usersResponse.Users) != tt.expectedCount {
					t.Errorf("Expected %d users, got %d", tt.expectedCount, len(usersResponse.Users))
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
