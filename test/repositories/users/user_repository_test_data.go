package users

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// Test data for TestUserRepository_Create
var UserRepositoryCreateTests = []struct {
	testName    string
	userName    string
	email       string
	phone       string
	status      string
	shouldError bool
}{
	{
		testName:    "Valid user creation",
		userName:    "john_doe",
		email:       "john@example.com",
		phone:       "+1234567890",
		status:      "active",
		shouldError: false,
	},
	{
		testName:    "User without phone",
		userName:    "jane_smith",
		email:       "jane@example.com",
		phone:       "",
		status:      "active",
		shouldError: false,
	},
	{
		testName:    "User with empty username",
		userName:    "",
		email:       "invalid@example.com",
		phone:       "+1234567890",
		status:      "active",
		shouldError: true,
	},
}

// Test data for TestUserRepository_GetByID
var UserRepositoryGetByIDTests = []struct {
	testName    string
	userID      string
	shouldError bool
}{
	{
		testName:    "Valid user ID",
		userID:      "user_123",
		shouldError: false,
	},
	{
		testName:    "Empty user ID",
		userID:      "",
		shouldError: true,
	},
	{
		testName:    "Non-existent user ID",
		userID:      "non_existent_id",
		shouldError: true,
	},
}

// Test data for TestUserRepository_GetByUsername
var UserRepositoryGetByUsernameTests = []struct {
	testName    string
	userName    string
	shouldError bool
}{
	{
		testName:    "Valid username",
		userName:    "john_doe",
		shouldError: false,
	},
	{
		testName:    "Empty username",
		userName:    "",
		shouldError: true,
	},
	{
		testName:    "Non-existent username",
		userName:    "non_existent_user",
		shouldError: true,
	},
}

// Test data for TestUserRepository_Update
var UserRepositoryUpdateTests = []struct {
	testName    string
	userID      string
	newStatus   string
	shouldError bool
}{
	{
		testName:    "Valid user update",
		userID:      "user_123",
		newStatus:   "active",
		shouldError: false,
	},
	{
		testName:    "Update non-existent user",
		userID:      "non_existent_id",
		newStatus:   "active",
		shouldError: true,
	},
}

// Test data for TestUserRepository_Delete
var UserRepositoryDeleteTests = []struct {
	testName    string
	userID      string
	shouldError bool
}{
	{
		testName:    "Valid user deletion",
		userID:      "user_123",
		shouldError: false,
	},
	{
		testName:    "Delete non-existent user",
		userID:      "non_existent_id",
		shouldError: true,
	},
	{
		testName:    "Delete with empty ID",
		userID:      "",
		shouldError: true,
	},
}

// Test data for TestUserRepository_List
var UserRepositoryListTests = []struct {
	testName      string
	limit         int
	offset        int
	expectedCount int
	shouldError   bool
}{
	{
		testName:      "Valid list with limit",
		limit:         10,
		offset:        0,
		expectedCount: 2,
		shouldError:   false,
	},
	{
		testName:      "List with offset",
		limit:         10,
		offset:        1,
		expectedCount: 1,
		shouldError:   false,
	},
	{
		testName:      "List with zero limit",
		limit:         0,
		offset:        0,
		expectedCount: 0,
		shouldError:   false,
	},
}

// Test data for TestUserRepository_Search
var UserRepositorySearchTests = []struct {
	testName      string
	searchTerm    string
	limit         int
	offset        int
	expectedCount int
	shouldError   bool
}{
	{
		testName:      "Search by name",
		searchTerm:    "John",
		limit:         10,
		offset:        0,
		expectedCount: 1,
		shouldError:   false,
	},
	{
		testName:      "Search by email",
		searchTerm:    "example.com",
		limit:         10,
		offset:        0,
		expectedCount: 2,
		shouldError:   false,
	},
	{
		testName:      "Search with no results",
		searchTerm:    "nonexistent",
		limit:         10,
		offset:        0,
		expectedCount: 0,
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
