package users

import (
	"errors"

	"github.com/Kisanlink/aaa-service/entities/responses/users"
)

// Test data for TestUserHandler_CreateUser
var UserHandlerCreateUserTests = []struct {
	name            string
	requestBody     map[string]interface{}
	mockResponse    *users.UserResponse
	mockError       error
	expectedStatus  int
	expectedSuccess bool
}{
	{
		name: "Valid user creation",
		requestBody: map[string]interface{}{
			"username": "johndoe",
			"password": "password123",
		},
		mockResponse: &users.UserResponse{
			ID:       "usr123456789",
			Username: stringPtr("johndoe"),
			Name:     stringPtr("John Doe"),
		},
		mockError:       nil,
		expectedStatus:  201,
		expectedSuccess: true,
	},
	{
		name: "Invalid request body",
		requestBody: map[string]interface{}{
			"username": "",
			"password": "short",
		},
		mockResponse:    nil,
		mockError:       nil,
		expectedStatus:  400,
		expectedSuccess: false,
	},
	{
		name: "Service error",
		requestBody: map[string]interface{}{
			"username": "johndoe",
			"password": "password123",
		},
		mockResponse:    nil,
		mockError:       errors.New("database error"),
		expectedStatus:  500,
		expectedSuccess: false,
	},
}

// Test data for TestUserHandler_GetUserByID
var UserHandlerGetUserByIDTests = []struct {
	name            string
	userID          string
	mockResponse    *users.UserResponse
	mockError       error
	expectedStatus  int
	expectedSuccess bool
}{
	{
		name:   "Valid user ID",
		userID: "usr123456789",
		mockResponse: &users.UserResponse{
			ID:       "usr123456789",
			Username: stringPtr("johndoe"),
			Name:     stringPtr("John Doe"),
		},
		mockError:       nil,
		expectedStatus:  200,
		expectedSuccess: true,
	},
	{
		name:            "Invalid user ID format",
		userID:          "invalid-id",
		mockResponse:    nil,
		mockError:       nil,
		expectedStatus:  400,
		expectedSuccess: false,
	},
	{
		name:            "Empty user ID",
		userID:          "",
		mockResponse:    nil,
		mockError:       nil,
		expectedStatus:  400,
		expectedSuccess: false,
	},
	{
		name:            "User not found",
		userID:          "usr999999999",
		mockResponse:    nil,
		mockError:       errors.New("user not found"),
		expectedStatus:  500,
		expectedSuccess: false,
	},
}

// Test data for TestUserHandler_UpdateUser
var UserHandlerUpdateUserTests = []struct {
	name            string
	userID          string
	requestBody     map[string]interface{}
	mockResponse    *users.UserResponse
	mockError       error
	expectedStatus  int
	expectedSuccess bool
}{
	{
		name:   "Valid user update",
		userID: "usr123456789",
		requestBody: map[string]interface{}{
			"username": "updateduser",
			"name":     "Updated Name",
		},
		mockResponse: &users.UserResponse{
			ID:       "usr123456789",
			Username: stringPtr("updateduser"),
			Name:     stringPtr("Updated Name"),
		},
		mockError:       nil,
		expectedStatus:  200,
		expectedSuccess: true,
	},
	{
		name:   "Invalid request body",
		userID: "usr123456789",
		requestBody: map[string]interface{}{
			"username": "",
			"password": "short",
		},
		mockResponse:    nil,
		mockError:       nil,
		expectedStatus:  400,
		expectedSuccess: false,
	},
	{
		name:   "Empty user ID",
		userID: "",
		requestBody: map[string]interface{}{
			"username": "updateduser",
			"name":     "Updated Name",
		},
		mockResponse:    nil,
		mockError:       nil,
		expectedStatus:  400,
		expectedSuccess: false,
	},
}

// Test data for TestUserHandler_DeleteUser
var UserHandlerDeleteUserTests = []struct {
	name            string
	userID          string
	mockResponse    *users.UserResponse
	mockError       error
	expectedStatus  int
	expectedSuccess bool
}{
	{
		name:   "Valid user deletion",
		userID: "usr123456789",
		mockResponse: &users.UserResponse{
			ID:       "usr123456789",
			Username: stringPtr("johndoe"),
			Name:     stringPtr("John Doe"),
			Status:   stringPtr("deleted"),
		},
		mockError:       nil,
		expectedStatus:  200,
		expectedSuccess: true,
	},
	{
		name:            "Invalid user ID format",
		userID:          "invalid-id",
		mockResponse:    nil,
		mockError:       nil,
		expectedStatus:  400,
		expectedSuccess: false,
	},
	{
		name:            "Empty user ID",
		userID:          "",
		mockResponse:    nil,
		mockError:       nil,
		expectedStatus:  400,
		expectedSuccess: false,
	},
	{
		name:            "User not found",
		userID:          "usr999999999",
		mockResponse:    nil,
		mockError:       errors.New("user not found"),
		expectedStatus:  500,
		expectedSuccess: false,
	},
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
