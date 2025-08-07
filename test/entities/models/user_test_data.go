package models

// Test data for TestNewUser
var NewUserTests = []struct {
	name     string
	username string
	email    string
}{
	{
		name:     "Valid user",
		username: "johndoe",
		email:    "john@example.com",
	},
	{
		name:     "User with special characters",
		username: "john_doe123",
		email:    "john.doe@example.com",
	},
}

// Test data for TestUserBeforeCreate
var UserBeforeCreateTests = []struct {
	name        string
	username    string
	email       string
	shouldError bool
}{
	{
		name:        "Valid user creation",
		username:    "johndoe",
		email:       "john@example.com",
		shouldError: false,
	},
	{
		name:        "User with empty username",
		username:    "",
		email:       "john@example.com",
		shouldError: true,
	},
	{
		name:        "User with empty password",
		username:    "johndoe",
		email:       "john@example.com",
		shouldError: true, // Password is required
	},
}

// Test data for TestUserBeforeUpdate
var UserBeforeUpdateTests = []struct {
	name        string
	username    string
	email       string
	shouldError bool
}{
	{
		name:        "Valid user update",
		username:    "johndoe",
		email:       "john@example.com",
		shouldError: false,
	},
	{
		name:        "User update with empty username",
		username:    "",
		email:       "john@example.com",
		shouldError: false, // Update doesn't validate username
	},
}

// Test data for TestUserValidation
var UserValidationTests = []struct {
	name        string
	username    string
	email       string
	phone       string
	status      string
	shouldError bool
}{
	{
		name:        "Valid user validation",
		username:    "johndoe",
		email:       "john@example.com",
		phone:       "+1234567890",
		status:      "active",
		shouldError: false,
	},
	{
		name:        "User with invalid phone",
		username:    "johndoe",
		email:       "john@example.com",
		phone:       "invalid-phone",
		status:      "active",
		shouldError: false, // Phone validation is lenient
	},
}
