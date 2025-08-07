package users

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/repositories/users"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Create(t *testing.T) {
	for _, tt := range UserRepositoryCreateTests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup test database and user repository
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userRepo := users.NewUserRepository(dbManager)

			// Create test user
			user := models.NewUser(tt.userName, "+91", "password123")
			status := tt.status
			user.Status = &status

			// Create user
			err := userRepo.Create(context.Background(), user)

			// Verify result
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	for _, tt := range UserRepositoryGetByIDTests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup test database and user repository
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userRepo := users.NewUserRepository(dbManager)

			var userID string
			// Create test user if needed
			if tt.testName == "Valid user ID" {
				user := models.NewUser("testuser", "+91", "password123")
				err := userRepo.Create(context.Background(), user)
				assert.NoError(t, err)
				userID = user.ID // Use the actual created user ID
			} else {
				userID = tt.userID // Use the test data ID for invalid cases
			}

			// Get user by ID
			_, err := userRepo.GetByID(context.Background(), userID)

			// Verify result
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	// TODO: Implement mock GetByUsername method
	// Skipping this test for now as the mock doesn't implement GetByUsername
	t.Skip("Skipping GetByUsername test - mock implementation needed")
}

func TestUserRepository_Update(t *testing.T) {
	// Setup test database
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userRepo := users.NewUserRepository(dbManager)

	// Create test user
	user := models.NewUser("testuser", "+91", "password123")
	err := userRepo.Create(context.Background(), user)
	assert.NoError(t, err)

	// Update user
	newStatus := "active"
	user.Status = &newStatus
	err = userRepo.Update(context.Background(), user)
	assert.NoError(t, err)

	// Verify update
	updatedUser, err := userRepo.GetByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "active", *updatedUser.Status)
}

func TestUserRepository_Delete(t *testing.T) {
	// Setup test database
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userRepo := users.NewUserRepository(dbManager)

	// Create test user
	user := models.NewUser("testuser", "+91", "password123")
	err := userRepo.Create(context.Background(), user)
	assert.NoError(t, err)

	// Delete user
	err = userRepo.Delete(context.Background(), user.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = userRepo.GetByID(context.Background(), user.ID)
	assert.Error(t, err)
}

func TestUserRepository_List(t *testing.T) {
	// Setup test database
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userRepo := users.NewUserRepository(dbManager)

	// Create test users
	user1 := models.NewUser("user1", "+91", "password123")
	user2 := models.NewUser("user2", "+91", "password123")

	err := userRepo.Create(context.Background(), user1)
	assert.NoError(t, err)
	t.Logf("Created user1 with ID: %s", user1.ID)

	err = userRepo.Create(context.Background(), user2)
	assert.NoError(t, err)
	t.Logf("Created user2 with ID: %s", user2.ID)

	// List users
	usersList, err := userRepo.List(context.Background(), 10, 0)
	assert.NoError(t, err)
	t.Logf("Retrieved %d users", len(usersList))
	for i, user := range usersList {
		username := ""
		if user.Username != nil {
			username = *user.Username
		}
		t.Logf("User %d: ID=%s, Username=%s", i, user.ID, username)
	}
	assert.Len(t, usersList, 2)
}

func TestUserRepository_Count(t *testing.T) {
	// Setup test database
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userRepo := users.NewUserRepository(dbManager)

	// Create test users
	user1 := models.NewUser("user1", "+91", "password123")
	user2 := models.NewUser("user2", "+91", "password123")

	err := userRepo.Create(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Create(context.Background(), user2)
	assert.NoError(t, err)

	// Count users
	count, err := userRepo.Count(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestUserRepository_Exists(t *testing.T) {
	// Setup test database
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userRepo := users.NewUserRepository(dbManager)

	// Create test user
	user := models.NewUser("testuser", "+91", "password123")
	err := userRepo.Create(context.Background(), user)
	assert.NoError(t, err)

	// Check if user exists
	exists, err := userRepo.Exists(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check if non-existent user exists
	exists, err = userRepo.Exists(context.Background(), "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_ListActive(t *testing.T) {
	// Setup test database
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userRepo := users.NewUserRepository(dbManager)

	// Create test users
	user1 := models.NewUser("user1", "+91", "password123")
	user2 := models.NewUser("user2", "+91", "password123")

	// Create users first (they will have default "pending" status)
	err := userRepo.Create(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Create(context.Background(), user2)
	assert.NoError(t, err)

	// Now update user1 to active status
	activeStatus := "active"
	user1.Status = &activeStatus
	err = userRepo.Update(context.Background(), user1)
	assert.NoError(t, err)

	// List active users
	activeUsers, err := userRepo.ListActive(context.Background(), 10, 0)
	assert.NoError(t, err)
	t.Logf("Retrieved %d active users", len(activeUsers))
	for i, user := range activeUsers {
		status := ""
		if user.Status != nil {
			status = *user.Status
		}
		username := ""
		if user.Username != nil {
			username = *user.Username
		}
		t.Logf("Active User %d: ID=%s, Username=%s, Status=%s", i, user.ID, username, status)
	}
	assert.Len(t, activeUsers, 1)
	if len(activeUsers) > 0 {
		assert.Equal(t, "active", *activeUsers[0].Status)
	}
}

func TestUserRepository_CountActive(t *testing.T) {
	// Setup test database
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userRepo := users.NewUserRepository(dbManager)

	// Create test users
	user1 := models.NewUser("user1", "+91", "password123")
	user2 := models.NewUser("user2", "+91", "password123")

	// Create users first
	err := userRepo.Create(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Create(context.Background(), user2)
	assert.NoError(t, err)

	// Update user1 to active status
	activeStatus := "active"
	user1.Status = &activeStatus
	err = userRepo.Update(context.Background(), user1)
	assert.NoError(t, err)

	// Count active users
	count, err := userRepo.CountActive(context.Background())
	assert.NoError(t, err)
	t.Logf("CountActive returned: %d", count)
	assert.Equal(t, int64(1), count)
}

func TestUserRepository_Search(t *testing.T) {
	// Setup test database
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userRepo := users.NewUserRepository(dbManager)

	// Create test users
	user1 := models.NewUserWithUsername("john_doe", "+91", "john_doe", "password123")
	user2 := models.NewUserWithUsername("jane_smith", "+91", "jane_smith", "password123")

	err := userRepo.Create(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Create(context.Background(), user2)
	assert.NoError(t, err)

	// Search users
	users, err := userRepo.Search(context.Background(), "john", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	if len(users) > 0 && users[0].Username != nil {
		assert.Equal(t, "john_doe", *users[0].Username)
	} else {
		t.Error("Expected username to be set")
	}
}

// Helper functions for test setup
func setupTestDatabase(t *testing.T) db.DBManager {
	return &MockDBManager{
		users:  make(map[string]*models.User),
		nextID: 1,
	}
}

func cleanupTestDatabase(t *testing.T, dbManager db.DBManager) {
	// Clean up test database
	if err := dbManager.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

// MockDBManager is a mock implementation of db.DBManager for testing
type MockDBManager struct {
	users  map[string]*models.User
	nextID int
}

func (m *MockDBManager) Connect(ctx context.Context) error { return nil }
func (m *MockDBManager) Close() error                      { return nil }
func (m *MockDBManager) IsConnected() bool                 { return true }
func (m *MockDBManager) GetBackendType() db.BackendType    { return db.BackendInMemory }

func (m *MockDBManager) Create(ctx context.Context, model interface{}) error {
	if m.users == nil {
		m.users = make(map[string]*models.User)
	}

	switch v := model.(type) {
	case *models.User:
		// Ensure unique ID - if ID is empty or already exists, generate a new one
		if v.ID == "" {
			m.nextID++
			v.ID = fmt.Sprintf("USER%08d", m.nextID)
		}

		// If ID already exists, generate a new one
		for m.users[v.ID] != nil {
			m.nextID++
			v.ID = fmt.Sprintf("USER%08d", m.nextID)
		}

		m.users[v.ID] = v
		return nil
	default:
		return fmt.Errorf("unsupported model type")
	}
}

func (m *MockDBManager) GetByID(ctx context.Context, id interface{}, model interface{}) error {
	if m.users == nil {
		m.users = make(map[string]*models.User)
	}

	idStr, ok := id.(string)
	if !ok {
		return fmt.Errorf("invalid ID type")
	}

	user, exists := m.users[idStr]
	if !exists {
		return fmt.Errorf("user not found")
	}

	switch v := model.(type) {
	case *models.User:
		*v = *user
		return nil
	default:
		return fmt.Errorf("unsupported model type")
	}
}

func (m *MockDBManager) Update(ctx context.Context, model interface{}) error {
	if m.users == nil {
		m.users = make(map[string]*models.User)
	}

	switch v := model.(type) {
	case *models.User:
		if _, exists := m.users[v.ID]; !exists {
			return fmt.Errorf("user not found")
		}
		m.users[v.ID] = v
		return nil
	default:
		return fmt.Errorf("unsupported model type")
	}
}

func (m *MockDBManager) Delete(ctx context.Context, id interface{}) error {
	if m.users == nil {
		m.users = make(map[string]*models.User)
	}

	idStr, ok := id.(string)
	if !ok {
		return fmt.Errorf("invalid ID type")
	}

	if _, exists := m.users[idStr]; !exists {
		return fmt.Errorf("user not found")
	}

	delete(m.users, idStr)
	return nil
}

func (m *MockDBManager) List(ctx context.Context, filters []db.Filter, model interface{}) error {
	if m.users == nil {
		m.users = make(map[string]*models.User)
	}

	switch v := model.(type) {
	case *[]models.User:
		var result []models.User
		for _, user := range m.users {
			// Apply filters
			match := true
			for _, filter := range filters {
				switch filter.Field {
				case "status":
					if user.Status == nil || *user.Status != filter.Value {
						match = false
						break
					}
				case "username":
					if filter.Operator == db.FilterOpContains {
						keyword, ok := filter.Value.(string)
						if !ok {
							match = false
							break
						}
						if user.Username == nil || !strings.Contains(strings.ToLower(*user.Username), strings.ToLower(keyword)) {
							match = false
							break
						}
					} else if user.Username != filter.Value {
						match = false
						break
					}
				}
				// If no match found for this filter, break out of filter loop
				if !match {
					break
				}
			}
			if match {
				result = append(result, *user)
			}
		}
		*v = result
		return nil
	default:
		return fmt.Errorf("unsupported model type")
	}
}

func (m *MockDBManager) AutoMigrateModels(ctx context.Context, models ...interface{}) error {
	// Mock implementation - no-op for testing
	return nil
}

func (m *MockDBManager) ApplyFilters(query interface{}, filters []db.Filter) (interface{}, error) {
	// Mock implementation - return query as-is for testing
	return query, nil
}

func (m *MockDBManager) BuildFilter(field string, operator db.FilterOperator, value interface{}) db.Filter {
	return db.Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}
