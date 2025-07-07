package models

// Test data for TestNewUser
var NewUserTests = []struct {
	name  string
	email string
}{
	{
		name:  "John Doe",
		email: "john@example.com",
	},
	{
		name:  "Jane Smith",
		email: "jane.smith@company.com",
	},
	{
		name:  "Test User",
		email: "test@test.org",
	},
}

// Test data for TestUserBeforeCreate
var UserBeforeCreateTests = []struct {
	name        string
	email       string
	shouldError bool
}{
	{
		name:        "Valid user",
		email:       "test@example.com",
		shouldError: false,
	},
	{
		name:        "Empty name",
		email:       "test@example.com",
		shouldError: true,
	},
	{
		name:        "Empty email",
		email:       "",
		shouldError: true,
	},
	{
		name:        "Invalid email format",
		email:       "invalid-email",
		shouldError: true,
	},
}

// Test data for TestUserBeforeUpdate
var UserBeforeUpdateTests = []struct {
	name        string
	email       string
	shouldError bool
}{
	{
		name:        "Valid user update",
		email:       "test@example.com",
		shouldError: false,
	},
	{
		name:        "Empty name update",
		email:       "test@example.com",
		shouldError: true,
	},
	{
		name:        "Empty email update",
		email:       "",
		shouldError: true,
	},
}

// Test data for TestUserValidation
var UserValidationTests = []struct {
	name        string
	email       string
	phone       string
	status      string
	shouldError bool
}{
	{
		name:        "Valid user with phone",
		email:       "test@example.com",
		phone:       "+1234567890",
		status:      "active",
		shouldError: false,
	},
	{
		name:        "Valid user without phone",
		email:       "test@example.com",
		phone:       "",
		status:      "active",
		shouldError: false,
	},
	{
		name:        "Invalid phone format",
		email:       "test@example.com",
		phone:       "invalid-phone",
		status:      "active",
		shouldError: true,
	},
	{
		name:        "Invalid status",
		email:       "test@example.com",
		phone:       "+1234567890",
		status:      "invalid-status",
		shouldError: true,
	},
	{
		name:        "Empty name",
		email:       "test@example.com",
		phone:       "+1234567890",
		status:      "active",
		shouldError: true,
	},
	{
		name:        "Empty email",
		email:       "",
		phone:       "+1234567890",
		status:      "active",
		shouldError: true,
	},
}

// Helper function to create test user with specific data
func CreateTestUser(name, email string) *User {
	return NewUser(name, email)
}

// Helper function to create test user with all fields
func CreateTestUserWithAllFields(name, email, phone, status string) *User {
	user := NewUser(name, email)
	user.Phone = phone
	user.Status = status
	return user
}

// Helper function to validate user fields
func ValidateUserFields(user *User, expectedName, expectedEmail string) bool {
	return user.Name == expectedName && user.Email == expectedEmail
}

// Helper function to validate user with all fields
func ValidateUserWithAllFields(user *User, expectedName, expectedEmail, expectedPhone, expectedStatus string) bool {
	return user.Name == expectedName &&
		user.Email == expectedEmail &&
		user.Phone == expectedPhone &&
		user.Status == expectedStatus
}
