package services

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// Test data for TestUserService_CreateUser
var UserServiceCreateUserTests = []struct {
	testName    string
	userName    string
	email       string
	phone       string
	status      string
	shouldError bool
}{
	{
		testName:    "Valid user creation",
		userName:    "John Doe",
		email:       "john@example.com",
		phone:       "+1234567890",
		status:      "active",
		shouldError: false,
	},
	{
		testName:    "User without phone",
		userName:    "Jane Smith",
		email:       "jane@example.com",
		phone:       "",
		status:      "active",
		shouldError: false,
	},
	{
		testName:    "User with invalid email",
		userName:    "Invalid User",
		email:       "invalid-email",
		phone:       "+1234567890",
		status:      "active",
		shouldError: true,
	},
	{
		testName:    "User with empty name",
		userName:    "",
		email:       "test@example.com",
		phone:       "+1234567890",
		status:      "active",
		shouldError: true,
	},
}

// Test data for TestUserService_GetUserByID
var UserServiceGetUserByIDTests = []struct {
	testName    string
	userID      string
	setupUser   bool
	userName    string
	shouldError bool
}{
	{
		testName:    "Valid user retrieval",
		userID:      "user123",
		setupUser:   true,
		userName:    "Test User",
		shouldError: false,
	},
	{
		testName:    "User not found",
		userID:      "nonexistent",
		setupUser:   false,
		userName:    "",
		shouldError: true,
	},
	{
		testName:    "Empty user ID",
		userID:      "",
		setupUser:   false,
		userName:    "",
		shouldError: true,
	},
}

// Test data for TestUserService_UpdateUser
var UserServiceUpdateUserTests = []struct {
	testName    string
	userID      string
	userName    string
	newUsername string
	newEmail    string
	shouldError bool
}{
	{
		testName:    "Valid user update",
		userID:      "user123",
		userName:    "Original User",
		newUsername: "Updated User",
		newEmail:    "updated@example.com",
		shouldError: false,
	},
	{
		testName:    "Update non-existent user",
		userID:      "nonexistent",
		userName:    "Original User",
		newUsername: "Updated User",
		newEmail:    "updated@example.com",
		shouldError: true,
	},
	{
		testName:    "Update with invalid email",
		userID:      "user123",
		userName:    "Original User",
		newUsername: "Updated User",
		newEmail:    "invalid-email",
		shouldError: true,
	},
}

// Test data for TestUserService_DeleteUser
var UserServiceDeleteUserTests = []struct {
	testName    string
	userID      string
	setupUser   bool
	userName    string
	shouldError bool
}{
	{
		testName:    "Valid user deletion",
		userID:      "user123",
		setupUser:   true,
		userName:    "Test User",
		shouldError: false,
	},
	{
		testName:    "Delete non-existent user",
		userID:      "nonexistent",
		setupUser:   false,
		userName:    "",
		shouldError: true,
	},
	{
		testName:    "Delete with empty ID",
		userID:      "",
		setupUser:   false,
		userName:    "",
		shouldError: true,
	},
}

// Test data for TestUserService_ListUsers
var UserServiceListUsersTests = []struct {
	testName      string
	testUsers     []TestUserData
	limit         int
	offset        int
	expectedCount int
	shouldError   bool
}{
	{
		testName: "List all users",
		testUsers: []TestUserData{
			{userName: "User 1", email: "user1@example.com", status: "active"},
			{userName: "User 2", email: "user2@example.com", status: "active"},
			{userName: "User 3", email: "user3@example.com", status: "inactive"},
		},
		limit:         10,
		offset:        0,
		expectedCount: 3,
		shouldError:   false,
	},
	{
		testName: "List with limit",
		testUsers: []TestUserData{
			{userName: "User 1", email: "user1@example.com", status: "active"},
			{userName: "User 2", email: "user2@example.com", status: "active"},
			{userName: "User 3", email: "user3@example.com", status: "active"},
		},
		limit:         2,
		offset:        0,
		expectedCount: 2,
		shouldError:   false,
	},
	{
		testName: "List with offset",
		testUsers: []TestUserData{
			{userName: "User 1", email: "user1@example.com", status: "active"},
			{userName: "User 2", email: "user2@example.com", status: "active"},
			{userName: "User 3", email: "user3@example.com", status: "active"},
		},
		limit:         2,
		offset:        1,
		expectedCount: 2,
		shouldError:   false,
	},
}

// TestUserData represents test user data structure
type TestUserData struct {
	userName string
	email    string
	status   string
}

// CreateTestUserData creates test user data
func CreateTestUserData(name, email, status string) TestUserData {
	return TestUserData{
		userName: name,
		email:    email,
		status:   status,
	}
}

// CreateTestFilters creates test filters for database operations
func CreateTestFilters(conditions map[string]interface{}, limit, offset int) *base.Filter {
	return &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{},
			Logic:      base.LogicAnd,
		},
		Limit:  limit,
		Offset: offset,
	}
}
