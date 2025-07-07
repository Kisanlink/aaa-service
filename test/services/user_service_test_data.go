package services

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// Test data for TestUserService_CreateUser
var UserServiceCreateUserTests = []struct {
	name        string
	name        string
	email       string
	phone       string
	status      string
	shouldError bool
}{
	{
		name:        "Valid user creation",
		name:        "John Doe",
		email:       "john@example.com",
		phone:       "+1234567890",
		status:      "active",
		shouldError: false,
	},
	{
		name:        "User without phone",
		name:        "Jane Smith",
		email:       "jane@example.com",
		phone:       "",
		status:      "active",
		shouldError: false,
	},
	{
		name:        "User with invalid email",
		name:        "Invalid User",
		email:       "invalid-email",
		phone:       "+1234567890",
		status:      "active",
		shouldError: true,
	},
	{
		name:        "User with empty name",
		name:        "",
		email:       "test@example.com",
		phone:       "+1234567890",
		status:      "active",
		shouldError: true,
	},
}

// Test data for TestUserService_GetUserByID
var UserServiceGetUserByIDTests = []struct {
	name           string
	userID         string
	createTestUser bool
	shouldError    bool
}{
	{
		name:           "Get existing user",
		userID:         "USER123456789", // This will be replaced with actual created user ID
		createTestUser: true,
		shouldError:    false,
	},
	{
		name:           "Get non-existent user",
		userID:         "USER999999999",
		createTestUser: false,
		shouldError:    true,
	},
	{
		name:           "Get with empty ID",
		userID:         "",
		createTestUser: false,
		shouldError:    true,
	},
}

// Test data for TestUserService_UpdateUser
var UserServiceUpdateUserTests = []struct {
	name        string
	newName     string
	newEmail    string
	newPhone    string
	newStatus   string
	shouldError bool
}{
	{
		name:        "Valid user update",
		newName:     "Updated Name",
		newEmail:    "updated@example.com",
		newPhone:    "+9876543210",
		newStatus:   "inactive",
		shouldError: false,
	},
	{
		name:        "Update with empty name",
		newName:     "",
		newEmail:    "updated@example.com",
		newPhone:    "+9876543210",
		newStatus:   "active",
		shouldError: true,
	},
	{
		name:        "Update with invalid email",
		newName:     "Updated Name",
		newEmail:    "invalid-email",
		newPhone:    "+9876543210",
		newStatus:   "active",
		shouldError: true,
	},
}

// Test data for TestUserService_DeleteUser
var UserServiceDeleteUserTests = []struct {
	name        string
	userID      string
	shouldError bool
}{
	{
		name:        "Delete existing user",
		userID:      "USER123456789", // This will be replaced with actual created user ID
		shouldError: false,
	},
	{
		name:        "Delete non-existent user",
		userID:      "USER999999999",
		shouldError: true,
	},
	{
		name:        "Delete with empty ID",
		userID:      "",
		shouldError: true,
	},
}

// Test data for TestUserService_ListUsers
type TestUserData struct {
	name   string
	email  string
	status string
}

var UserServiceListUsersTests = []struct {
	name          string
	testUsers     []TestUserData
	filters       *base.Filters
	expectedCount int
	shouldError   bool
}{
	{
		name: "List all users",
		testUsers: []TestUserData{
			{name: "User 1", email: "user1@example.com", status: "active"},
			{name: "User 2", email: "user2@example.com", status: "active"},
			{name: "User 3", email: "user3@example.com", status: "inactive"},
		},
		filters:       &base.Filters{},
		expectedCount: 3,
		shouldError:   false,
	},
	{
		name: "List active users only",
		testUsers: []TestUserData{
			{name: "User 1", email: "user1@example.com", status: "active"},
			{name: "User 2", email: "user2@example.com", status: "active"},
			{name: "User 3", email: "user3@example.com", status: "inactive"},
		},
		filters: &base.Filters{
			Conditions: map[string]interface{}{
				"status": "active",
			},
		},
		expectedCount: 2,
		shouldError:   false,
	},
	{
		name: "List with pagination",
		testUsers: []TestUserData{
			{name: "User 1", email: "user1@example.com", status: "active"},
			{name: "User 2", email: "user2@example.com", status: "active"},
			{name: "User 3", email: "user3@example.com", status: "active"},
		},
		filters: &base.Filters{
			Limit:  2,
			Offset: 0,
		},
		expectedCount: 2,
		shouldError:   false,
	},
}

// Helper function to create test user data
func CreateTestUserData(name, email, status string) TestUserData {
	return TestUserData{
		name:   name,
		email:  email,
		status: status,
	}
}

// Helper function to create test filters
func CreateTestFilters(conditions map[string]interface{}, limit, offset int) *base.Filters {
	return &base.Filters{
		Conditions: conditions,
		Limit:      limit,
		Offset:     offset,
	}
}
