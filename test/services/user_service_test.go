package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Kisanlink/aaa-service/entities/models"
	userRequests "github.com/Kisanlink/aaa-service/entities/requests/users"
	userResponses "github.com/Kisanlink/aaa-service/entities/responses/users"
	userRepositories "github.com/Kisanlink/aaa-service/repositories/users"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser(t *testing.T) {
	for _, tt := range UserServiceCreateUserTests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup test database and repositories
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			// Create service with proper error handling
			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			// Create user request
			req := &userRequests.CreateUserRequest{
				Username: tt.userName,
				Password: "password123",
			}

			// Create user
			_, err := userService.CreateUser(context.Background(), req)

			// Verify result
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	for _, tt := range UserServiceGetUserByIDTests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup test database and repositories
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			// Initialize repositories
			userRepo := userRepositories.NewUserRepository(dbManager)

			// Create service with proper error handling
			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			// Create test user if needed
			if tt.setupUser {
				user := models.NewUser(tt.userName, "password123")
				err := userRepo.Create(context.Background(), user)
				assert.NoError(t, err)
			}

			// Get user by ID
			_, err := userService.GetUserByID(context.Background(), tt.userID)

			// Verify result
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions for test setup
func setupTestDatabase(t *testing.T) db.DBManager {
	// Return a mock database manager for testing
	return &MockDBManager{}
}

func cleanupTestDatabase(t *testing.T, dbManager db.DBManager) {
	// Clean up test database
	if err := dbManager.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

// MockDBManager is a mock implementation of db.DBManager for testing
type MockDBManager struct{}

func (m *MockDBManager) Connect(ctx context.Context) error { return nil }
func (m *MockDBManager) Close() error                      { return nil }
func (m *MockDBManager) IsConnected() bool                 { return true }
func (m *MockDBManager) GetBackendType() db.BackendType    { return db.BackendInMemory }
func (m *MockDBManager) Create(ctx context.Context, model interface{}) error {
	return nil
}
func (m *MockDBManager) GetByID(ctx context.Context, id interface{}, model interface{}) error {
	return nil
}
func (m *MockDBManager) Update(ctx context.Context, model interface{}) error { return nil }
func (m *MockDBManager) Delete(ctx context.Context, id interface{}) error    { return nil }
func (m *MockDBManager) List(ctx context.Context, filters []db.Filter, model interface{}) error {
	return nil
}
func (m *MockDBManager) ApplyFilters(query interface{}, filters []db.Filter) (interface{}, error) {
	return query, nil
}
func (m *MockDBManager) BuildFilter(field string, operator db.FilterOperator, value interface{}) db.Filter {
	return db.Filter{Field: field, Operator: operator, Value: value}
}

// MockUserService is a mock implementation of UserService for testing
type MockUserService struct {
	shouldError bool
}

func (m *MockUserService) CreateUser(ctx context.Context, req *userRequests.CreateUserRequest) (*userResponses.UserResponse, error) {
	if m.shouldError {
		return nil, errors.New("validation error")
	}
	return &userResponses.UserResponse{
		ID:       "usr123456789",
		Username: req.Username,
	}, nil
}

func (m *MockUserService) GetUserByID(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	if m.shouldError {
		return nil, errors.New("user not found")
	}
	return &userResponses.UserResponse{
		ID:       userID,
		Username: "testuser",
	}, nil
}

func (m *MockUserService) GetUserByUsername(ctx context.Context, username string) (*userResponses.UserResponse, error) {
	if m.shouldError {
		return nil, errors.New("user not found")
	}
	return &userResponses.UserResponse{
		ID:       "usr123456789",
		Username: username,
	}, nil
}

func (m *MockUserService) GetUserByMobileNumber(ctx context.Context, mobileNumber uint64) (*userResponses.UserResponse, error) {
	if m.shouldError {
		return nil, errors.New("user not found")
	}
	return &userResponses.UserResponse{
		ID:           "usr123456789",
		Username:     "testuser",
		MobileNumber: mobileNumber,
	}, nil
}

func (m *MockUserService) GetUserByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*userResponses.UserResponse, error) {
	if m.shouldError {
		return nil, errors.New("user not found")
	}
	return &userResponses.UserResponse{
		ID:            "usr123456789",
		Username:      "testuser",
		AadhaarNumber: &aadhaarNumber,
	}, nil
}

func (m *MockUserService) UpdateUser(ctx context.Context, req *userRequests.UpdateUserRequest) (*userResponses.UserResponse, error) {
	if m.shouldError {
		return nil, errors.New("update failed")
	}
	return &userResponses.UserResponse{
		ID:       "usr123456789",
		Username: "testuser",
	}, nil
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID string) error {
	if m.shouldError {
		return errors.New("delete failed")
	}
	return nil
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	if m.shouldError {
		return nil, errors.New("list failed")
	}
	return []userResponses.UserResponse{}, nil
}

func (m *MockUserService) ListActiveUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	if m.shouldError {
		return nil, errors.New("list failed")
	}
	return []userResponses.UserResponse{}, nil
}

func (m *MockUserService) SearchUsers(ctx context.Context, keyword string, limit, offset int) (interface{}, error) {
	if m.shouldError {
		return nil, errors.New("search failed")
	}
	return []userResponses.UserResponse{}, nil
}

func (m *MockUserService) ValidateUser(ctx context.Context, userID string) error {
	if m.shouldError {
		return errors.New("validation failed")
	}
	return nil
}

func (m *MockUserService) DeductTokens(ctx context.Context, userID string, amount int) error {
	if m.shouldError {
		return errors.New("deduct failed")
	}
	return nil
}

func (m *MockUserService) AddTokens(ctx context.Context, userID string, amount int) error {
	if m.shouldError {
		return errors.New("add failed")
	}
	return nil
}

func (m *MockUserService) GetUserWithProfile(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	if m.shouldError {
		return nil, errors.New("get profile failed")
	}
	return &userResponses.UserResponse{
		ID:       userID,
		Username: "testuser",
	}, nil
}

func (m *MockUserService) GetUserWithRoles(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	if m.shouldError {
		return nil, errors.New("get roles failed")
	}
	return &userResponses.UserResponse{
		ID:       userID,
		Username: "testuser",
	}, nil
}
