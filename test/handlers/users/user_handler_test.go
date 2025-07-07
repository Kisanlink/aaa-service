package users

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kisanlink/aaa-service/entities/requests/users"
	"github.com/Kisanlink/aaa-service/entities/responses/users"
	"github.com/Kisanlink/aaa-service/handlers/users"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/Kisanlink/aaa-service/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req *requests.CreateUserRequest) (*responses.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*responses.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, userID string) (*responses.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*responses.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, req *requests.UpdateUserRequest) (*responses.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*responses.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID string) (*responses.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*responses.UserResponse), args.Error(1)
}

func (m *MockUserService) ListUsers(ctx context.Context, filters interface{}) (*responses.UsersListResponse, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*responses.UsersListResponse), args.Error(1)
}

func TestUserHandler_CreateUser(t *testing.T) {
	for _, tt := range UserHandlerCreateUserTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request body
			reqBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			// Setup mock service
			mockService := new(MockUserService)
			if tt.mockResponse != nil {
				mockService.On("CreateUser", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)
			}

			// Setup handler
			validator := utils.NewValidator()
			responder := utils.NewResponder(true)
			handler := handlers.NewUserHandler(mockService, validator, responder)

			// Execute
			handler.CreateUser(c)

			// Assert
			if tt.expectedStatus != w.Code {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify response body
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedSuccess != response["success"] {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, response["success"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUserByID(t *testing.T) {
	for _, tt := range UserHandlerGetUserByIDTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request
			c.Request = httptest.NewRequest("GET", "/users/"+tt.userID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.userID}}

			// Setup mock service
			mockService := new(MockUserService)
			if tt.mockResponse != nil {
				mockService.On("GetUserByID", mock.Anything, tt.userID).Return(tt.mockResponse, tt.mockError)
			}

			// Setup handler
			validator := utils.NewValidator()
			responder := utils.NewResponder(true)
			handler := handlers.NewUserHandler(mockService, validator, responder)

			// Execute
			handler.GetUserByID(c)

			// Assert
			if tt.expectedStatus != w.Code {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify response body
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedSuccess != response["success"] {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, response["success"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	for _, tt := range UserHandlerUpdateUserTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request body
			reqBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("PUT", "/users/"+tt.userID, bytes.NewBuffer(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.userID}}

			// Setup mock service
			mockService := new(MockUserService)
			if tt.mockResponse != nil {
				mockService.On("UpdateUser", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)
			}

			// Setup handler
			validator := utils.NewValidator()
			responder := utils.NewResponder(true)
			handler := handlers.NewUserHandler(mockService, validator, responder)

			// Execute
			handler.UpdateUser(c)

			// Assert
			if tt.expectedStatus != w.Code {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify response body
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedSuccess != response["success"] {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, response["success"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	for _, tt := range UserHandlerDeleteUserTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request
			c.Request = httptest.NewRequest("DELETE", "/users/"+tt.userID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.userID}}

			// Setup mock service
			mockService := new(MockUserService)
			if tt.mockResponse != nil {
				mockService.On("DeleteUser", mock.Anything, tt.userID).Return(tt.mockResponse, tt.mockError)
			}

			// Setup handler
			validator := utils.NewValidator()
			responder := utils.NewResponder(true)
			handler := handlers.NewUserHandler(mockService, validator, responder)

			// Execute
			handler.DeleteUser(c)

			// Assert
			if tt.expectedStatus != w.Code {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify response body
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedSuccess != response["success"] {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, response["success"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	for _, tt := range UserHandlerListUsersTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request with query parameters
			req := httptest.NewRequest("GET", "/users", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()
			c.Request = req

			// Setup mock service
			mockService := new(MockUserService)
			if tt.mockResponse != nil {
				mockService.On("ListUsers", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)
			}

			// Setup handler
			validator := utils.NewValidator()
			responder := utils.NewResponder(true)
			handler := handlers.NewUserHandler(mockService, validator, responder)

			// Execute
			handler.ListUsers(c)

			// Assert
			if tt.expectedStatus != w.Code {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify response body
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedSuccess != response["success"] {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, response["success"])
			}

			mockService.AssertExpectations(t)
		})
	}
}
