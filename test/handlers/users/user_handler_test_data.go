package users

import (
	"errors"

	"github.com/Kisanlink/aaa-service/entities/responses/users"
)

// Test data for TestUserHandler_CreateUser
var UserHandlerCreateUserTests = []struct {
	name            string
	requestBody     map[string]interface{}
	mockResponse    *responses.UserResponse
	mockError       error
	expectedStatus  int
	expectedSuccess bool
}{
	{
		name: "Valid user creation",
		requestBody: map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"phone": "+1234567890",
		},
		mockResponse: &responses.UserResponse{
			ID:    "USER123456789",
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "+1234567890",
		},
		mockError:       nil,
		expectedStatus:  201,
		expectedSuccess: true,
	},
	{
		name: "Invalid request body",
		requestBody: map[string]interface{}{
			"name":  "",
			"email": "invalid-email",
		},
		mockResponse:    nil,
		mockError:       nil,
		expectedStatus:  400,
		expectedSuccess: false,
	},
	{
		name: "Service error",
		requestBody: map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
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
	mockResponse    *responses.UserResponse
	mockError       error
	expectedStatus  int
	expectedSuccess bool
}{
	{
		name:   "Valid user ID",
		userID: "USER123456789",
		mockResponse: &responses.UserResponse{
			ID:    "USER123456789",
			Name:  "John Doe",
			Email: "john@example.com",
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
		userID:          "USER999999999",
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
	mockResponse    *responses.UserResponse
	mockError       error
	expectedStatus  int
	expectedSuccess bool
}{
	{
		name:   "Valid user update",
		userID: "USER123456789",
		requestBody: map[string]interface{}{
			"name":  "Updated Name",
			"email": "updated@example.com",
			"phone": "+9876543210",
		},
		mockResponse: &responses.UserResponse{
			ID:    "USER123456789",
			Name:  "Updated Name",
			Email: "updated@example.com",
			Phone: "+9876543210",
		},
		mockError:       nil,
		expectedStatus:  200,
		expectedSuccess: true,
	},
	{
		name:   "Invalid request body",
		userID: "USER123456789",
		requestBody: map[string]interface{}{
			"name":  "",
			"email": "invalid-email",
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
			"name":  "Updated Name",
			"email": "updated@example.com",
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
	mockResponse    *responses.UserResponse
	mockError       error
	expectedStatus  int
	expectedSuccess bool
}{
	{
		name:   "Valid user deletion",
		userID: "USER123456789",
		mockResponse: &responses.UserResponse{
			ID:     "USER123456789",
			Name:   "John Doe",
			Email:  "john@example.com",
			Status: "deleted",
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
		userID:          "USER999999999",
		mockResponse:    nil,
		mockError:       errors.New("user not found"),
		expectedStatus:  500,
		expectedSuccess: false,
	},
}

// Test data for TestUserHandler_ListUsers
var UserHandlerListUsersTests = []struct {
	name            string
	queryParams     map[string]string
	mockResponse    *responses.UsersListResponse
	mockError       error
	expectedStatus  int
	expectedSuccess bool
}{
	{
		name:        "List all users",
		queryParams: map[string]string{},
		mockResponse: &responses.UsersListResponse{
			Users: []responses.UserResponse{
				{ID: "USER123456789", Name: "John Doe", Email: "john@example.com"},
				{ID: "USER987654321", Name: "Jane Smith", Email: "jane@example.com"},
			},
			Total: 2,
		},
		mockError:       nil,
		expectedStatus:  200,
		expectedSuccess: true,
	},
	{
		name: "List with pagination",
		queryParams: map[string]string{
			"page":  "1",
			"limit": "10",
		},
		mockResponse: &responses.UsersListResponse{
			Users: []responses.UserResponse{
				{ID: "USER123456789", Name: "John Doe", Email: "john@example.com"},
			},
			Total: 1,
		},
		mockError:       nil,
		expectedStatus:  200,
		expectedSuccess: true,
	},
	{
		name: "List with status filter",
		queryParams: map[string]string{
			"status": "active",
		},
		mockResponse: &responses.UsersListResponse{
			Users: []responses.UserResponse{
				{ID: "USER123456789", Name: "John Doe", Email: "john@example.com", Status: "active"},
			},
			Total: 1,
		},
		mockError:       nil,
		expectedStatus:  200,
		expectedSuccess: true,
	},
	{
		name: "Invalid pagination parameters",
		queryParams: map[string]string{
			"page":  "0",
			"limit": "101",
		},
		mockResponse:    nil,
		mockError:       nil,
		expectedStatus:  400,
		expectedSuccess: false,
	},
	{
		name:            "Service error",
		queryParams:     map[string]string{},
		mockResponse:    nil,
		mockError:       errors.New("database error"),
		expectedStatus:  500,
		expectedSuccess: false,
	},
}
