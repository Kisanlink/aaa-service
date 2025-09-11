//nolint:typecheck
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/requests"
	userRequests "github.com/Kisanlink/aaa-service/internal/entities/requests/users"
	userResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) VerifyUserCredentials(ctx context.Context, phone, countryCode string, password, mpin *string) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, phone, countryCode, password, mpin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) CreateUser(ctx context.Context, req *userRequests.CreateUserRequest) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) SetMPin(ctx context.Context, userID, mPin, currentPassword string) error {
	args := m.Called(ctx, userID, mPin, currentPassword)
	return args.Error(0)
}

func (m *MockUserService) VerifyMPin(ctx context.Context, userID, mPin string) error {
	args := m.Called(ctx, userID, mPin)
	return args.Error(0)
}

// Add other required methods as no-ops for interface compliance
func (m *MockUserService) GetUserByUsername(ctx context.Context, username string) (*userResponses.UserResponse, error) {
	return nil, nil
}
func (m *MockUserService) GetUserByMobileNumber(ctx context.Context, mobileNumber uint64) (*userResponses.UserResponse, error) {
	return nil, nil
}
func (m *MockUserService) GetUserByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*userResponses.UserResponse, error) {
	return nil, nil
}
func (m *MockUserService) UpdateUser(ctx context.Context, req *userRequests.UpdateUserRequest) (*userResponses.UserResponse, error) {
	return nil, nil
}
func (m *MockUserService) DeleteUser(ctx context.Context, userID string) error {
	return nil
}
func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	return nil, nil
}
func (m *MockUserService) ListActiveUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	return nil, nil
}
func (m *MockUserService) SearchUsers(ctx context.Context, keyword string, limit, offset int) (interface{}, error) {
	return nil, nil
}
func (m *MockUserService) ValidateUser(ctx context.Context, userID string) error {
	return nil
}
func (m *MockUserService) DeductTokens(ctx context.Context, userID string, amount int) error {
	return nil
}
func (m *MockUserService) AddTokens(ctx context.Context, userID string, amount int) error {
	return nil
}
func (m *MockUserService) GetUserWithProfile(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	return nil, nil
}
func (m *MockUserService) GetUserWithRoles(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	return nil, nil
}
func (m *MockUserService) VerifyUserPassword(ctx context.Context, username, password string) (*userResponses.UserResponse, error) {
	return nil, nil
}
func (m *MockUserService) VerifyUserPasswordByPhone(ctx context.Context, phoneNumber, countryCode, password string) (*userResponses.UserResponse, error) {
	return nil, nil
}
func (m *MockUserService) UpdateMPin(ctx context.Context, userID, currentMPin, newMPin string) error {
	args := m.Called(ctx, userID, currentMPin, newMPin)
	return args.Error(0)
}
func (m *MockUserService) GetUserByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (*userResponses.UserResponse, error) {
	return nil, nil
}

func (m *MockUserService) SoftDeleteUserWithCascade(ctx context.Context, userID, deletedBy string) error {
	args := m.Called(ctx, userID, deletedBy)
	return args.Error(0)
}

// MockValidator is a mock implementation of Validator
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) ValidateStruct(s interface{}) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockValidator) ValidateUserID(userID string) error {
	return nil
}
func (m *MockValidator) ValidateEmail(email string) error {
	return nil
}
func (m *MockValidator) ValidatePassword(password string) error {
	return nil
}
func (m *MockValidator) ValidatePhoneNumber(phone string) error {
	return nil
}
func (m *MockValidator) ValidateAadhaarNumber(aadhaar string) error {
	return nil
}
func (m *MockValidator) ParseListFilters(c *gin.Context) (interface{}, error) {
	return nil, nil
}

// MockResponder is a mock implementation of Responder
type MockResponder struct {
	mock.Mock
}

func (m *MockResponder) SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	m.Called(c, statusCode, data)
}

func (m *MockResponder) SendError(c *gin.Context, statusCode int, message string, err error) {
	m.Called(c, statusCode, message, err)
}

func (m *MockResponder) SendValidationError(c *gin.Context, errors []string) {
	m.Called(c, errors)
}

func (m *MockResponder) SendInternalError(c *gin.Context, err error) {
	m.Called(c, err)
}

func setupTestHandler() (*AuthHandler, *MockUserService, *MockValidator, *MockResponder) {
	mockUserService := &MockUserService{}
	mockValidator := &MockValidator{}
	mockResponder := &MockResponder{}
	logger := zap.NewNop()

	handler := NewAuthHandler(mockUserService, mockValidator, mockResponder, logger)
	return handler, mockUserService, mockValidator, mockResponder
}

func TestLoginV2_PasswordAuthentication_Success(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	password := "testpassword123"
	loginReq := requests.LoginRequest{
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		Password:    &password,
	}

	userResponse := &userResponses.UserResponse{
		ID:          "user-123",
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		Username:    stringPtr("testuser"),
		IsValidated: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tokens:      100,
		HasMPin:     false,
		Roles: []userResponses.UserRoleDetail{
			{
				ID:       "ur-1",
				UserID:   "user-123",
				RoleID:   "role-1",
				IsActive: true,
				Role: userResponses.RoleDetail{
					ID:          "role-1",
					Name:        "user",
					Description: "Standard user role",
					IsActive:    true,
				},
			},
		},
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &loginReq).Return(nil)
	mockUserService.On("VerifyUserCredentials", mock.Anything, "1234567890", "+91", &password, (*string)(nil)).Return(userResponse, nil)
	mockResponder.On("SendSuccess", mock.Anything, http.StatusOK, mock.AnythingOfType("*responses.LoginResponse")).Return()

	// Create request
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/v2/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.LoginV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestLoginV2_MPinAuthentication_Success(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	mpin := "1234"
	loginReq := requests.LoginRequest{
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		MPin:        &mpin,
	}

	userResponse := &userResponses.UserResponse{
		ID:          "user-123",
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		Username:    stringPtr("testuser"),
		IsValidated: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tokens:      100,
		HasMPin:     true,
		Roles:       []userResponses.UserRoleDetail{},
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &loginReq).Return(nil)
	mockUserService.On("VerifyUserCredentials", mock.Anything, "1234567890", "+91", (*string)(nil), &mpin).Return(userResponse, nil)
	mockResponder.On("SendSuccess", mock.Anything, http.StatusOK, mock.AnythingOfType("*responses.LoginResponse")).Return()

	// Create request
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/v2/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.LoginV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestLoginV2_BothPasswordAndMPin_PrioritizesPassword(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data - both password and MPIN provided
	password := "testpassword123"
	mpin := "1234"
	loginReq := requests.LoginRequest{
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		Password:    &password,
		MPin:        &mpin,
	}

	userResponse := &userResponses.UserResponse{
		ID:          "user-123",
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		Username:    stringPtr("testuser"),
		IsValidated: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tokens:      100,
		HasMPin:     true,
		Roles:       []userResponses.UserRoleDetail{},
	}

	// Setup mocks - should call with password, not MPIN
	mockValidator.On("ValidateStruct", &loginReq).Return(nil)
	mockUserService.On("VerifyUserCredentials", mock.Anything, "1234567890", "+91", &password, &mpin).Return(userResponse, nil)
	mockResponder.On("SendSuccess", mock.Anything, http.StatusOK, mock.AnythingOfType("*responses.LoginResponse")).Return()

	// Create request
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/v2/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.LoginV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestLoginV2_InvalidCredentials_ReturnsUnauthorized(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	password := "wrongpassword"
	loginReq := requests.LoginRequest{
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		Password:    &password,
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &loginReq).Return(nil)
	mockUserService.On("VerifyUserCredentials", mock.Anything, "1234567890", "+91", &password, (*string)(nil)).Return(nil, errors.NewUnauthorizedError("invalid credentials"))
	mockResponder.On("SendError", mock.Anything, http.StatusUnauthorized, "Invalid credentials", mock.AnythingOfType("*errors.UnauthorizedError")).Return()

	// Create request
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/v2/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.LoginV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestLoginV2_UserNotFound_ReturnsUnauthorized(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	password := "testpassword123"
	loginReq := requests.LoginRequest{
		PhoneNumber: "9999999999",
		CountryCode: "+91",
		Password:    &password,
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &loginReq).Return(nil)
	mockUserService.On("VerifyUserCredentials", mock.Anything, "9999999999", "+91", &password, (*string)(nil)).Return(nil, errors.NewNotFoundError("user not found"))
	mockResponder.On("SendError", mock.Anything, http.StatusUnauthorized, "Invalid credentials", mock.AnythingOfType("*errors.NotFoundError")).Return()

	// Create request
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/v2/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.LoginV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestLoginV2_MPinNotSet_ReturnsBadRequest(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	mpin := "1234"
	loginReq := requests.LoginRequest{
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		MPin:        &mpin,
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &loginReq).Return(nil)
	mockUserService.On("VerifyUserCredentials", mock.Anything, "1234567890", "+91", (*string)(nil), &mpin).Return(nil, errors.NewBadRequestError("mpin not set for user"))
	mockResponder.On("SendError", mock.Anything, http.StatusBadRequest, "mpin not set for user", mock.AnythingOfType("*errors.BadRequestError")).Return()

	// Create request
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/v2/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.LoginV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestLoginV2_ValidationError_ReturnsBadRequest(t *testing.T) {
	handler, _, _, mockResponder := setupTestHandler()

	// Setup test data - missing both password and MPIN
	loginReq := requests.LoginRequest{
		PhoneNumber: "1234567890",
		CountryCode: "+91",
	}

	// Setup mocks
	mockResponder.On("SendValidationError", mock.Anything, []string{"either password or mpin is required"}).Return()

	// Create request
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/v2/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.LoginV2(c)

	// Verify
	mockResponder.AssertExpectations(t)
}

func TestConvertToAuthUserInfo(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	// Setup test data
	userResponse := &userResponses.UserResponse{
		ID:          "user-123",
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		Username:    stringPtr("testuser"),
		IsValidated: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tokens:      100,
		HasMPin:     true,
		Roles: []userResponses.UserRoleDetail{
			{
				ID:       "ur-1",
				UserID:   "user-123",
				RoleID:   "role-1",
				IsActive: true,
				Role: userResponses.RoleDetail{
					ID:          "role-1",
					Name:        "admin",
					Description: "Administrator role",
					IsActive:    true,
				},
			},
		},
	}

	// Execute
	result := handler.convertToAuthUserInfo(userResponse)

	// Verify
	assert.Equal(t, userResponse.ID, result.ID)
	assert.Equal(t, userResponse.PhoneNumber, result.PhoneNumber)
	assert.Equal(t, userResponse.CountryCode, result.CountryCode)
	assert.Equal(t, userResponse.Username, result.Username)
	assert.Equal(t, userResponse.IsValidated, result.IsValidated)
	assert.Equal(t, userResponse.CreatedAt, result.CreatedAt)
	assert.Equal(t, userResponse.UpdatedAt, result.UpdatedAt)
	assert.Equal(t, userResponse.Tokens, result.Tokens)
	assert.Equal(t, userResponse.HasMPin, result.HasMPin)
	assert.Len(t, result.Roles, 1)
	assert.Equal(t, "admin", result.Roles[0].Role.Name)
}

func TestSetMPinV2_Success(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	setMPinReq := requests.SetMPinRequest{
		MPin:     "1234",
		Password: "testpassword123",
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &setMPinReq).Return(nil)
	mockUserService.On("SetMPin", mock.Anything, "user-123", "1234", "testpassword123").Return(nil)
	mockResponder.On("SendSuccess", mock.Anything, http.StatusOK, mock.MatchedBy(func(data map[string]any) bool {
		return data["success"] == true && data["message"] == "mPin set successfully"
	})).Return()

	// Create request
	reqBody, _ := json.Marshal(setMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/set-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context with user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	// Execute
	handler.SetMPinV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestSetMPinV2_UserNotAuthenticated(t *testing.T) {
	handler, _, _, mockResponder := setupTestHandler()

	// Setup test data
	setMPinReq := requests.SetMPinRequest{
		MPin:     "1234",
		Password: "testpassword123",
	}

	// Setup mocks
	mockResponder.On("SendError", mock.Anything, http.StatusUnauthorized, "User not authenticated", nil).Return()

	// Create request
	reqBody, _ := json.Marshal(setMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/set-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context without user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.SetMPinV2(c)

	// Verify
	mockResponder.AssertExpectations(t)
}

func TestSetMPinV2_ValidationError(t *testing.T) {
	handler, _, _, mockResponder := setupTestHandler()

	// Setup test data - invalid MPIN (too short)
	setMPinReq := requests.SetMPinRequest{
		MPin:     "12", // Invalid - too short
		Password: "testpassword123",
	}

	// Setup mocks
	mockResponder.On("SendValidationError", mock.Anything, []string{"mPin must be 4 or 6 digits"}).Return()

	// Create request
	reqBody, _ := json.Marshal(setMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/set-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context with user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	// Execute
	handler.SetMPinV2(c)

	// Verify
	mockResponder.AssertExpectations(t)
}

func TestSetMPinV2_InvalidPassword(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	setMPinReq := requests.SetMPinRequest{
		MPin:     "1234",
		Password: "wrongpassword",
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &setMPinReq).Return(nil)
	mockUserService.On("SetMPin", mock.Anything, "user-123", "1234", "wrongpassword").Return(errors.NewUnauthorizedError("invalid password"))
	mockResponder.On("SendError", mock.Anything, http.StatusUnauthorized, "Invalid password", mock.AnythingOfType("*errors.UnauthorizedError")).Return()

	// Create request
	reqBody, _ := json.Marshal(setMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/set-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context with user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	// Execute
	handler.SetMPinV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestSetMPinV2_ServiceValidationError(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	setMPinReq := requests.SetMPinRequest{
		MPin:     "1234",
		Password: "testpassword123",
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &setMPinReq).Return(nil)
	mockUserService.On("SetMPin", mock.Anything, "user-123", "1234", "testpassword123").Return(errors.NewValidationError("mPin already exists"))
	mockResponder.On("SendValidationError", mock.Anything, []string{"mPin already exists"}).Return()

	// Create request
	reqBody, _ := json.Marshal(setMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/set-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context with user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	// Execute
	handler.SetMPinV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestUpdateMPinV2_Success(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	updateMPinReq := requests.UpdateMPinRequest{
		CurrentMPin: "1234",
		NewMPin:     "5678",
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &updateMPinReq).Return(nil)
	mockUserService.On("UpdateMPin", mock.Anything, "user-123", "1234", "5678").Return(nil)
	mockResponder.On("SendSuccess", mock.Anything, http.StatusOK, mock.MatchedBy(func(data map[string]any) bool {
		return data["success"] == true && data["message"] == "mPin updated successfully"
	})).Return()

	// Create request
	reqBody, _ := json.Marshal(updateMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/update-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context with user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	// Execute
	handler.UpdateMPinV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestUpdateMPinV2_UserNotAuthenticated(t *testing.T) {
	handler, _, _, mockResponder := setupTestHandler()

	// Setup test data
	updateMPinReq := requests.UpdateMPinRequest{
		CurrentMPin: "1234",
		NewMPin:     "5678",
	}

	// Setup mocks
	mockResponder.On("SendError", mock.Anything, http.StatusUnauthorized, "User not authenticated", nil).Return()

	// Create request
	reqBody, _ := json.Marshal(updateMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/update-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context without user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.UpdateMPinV2(c)

	// Verify
	mockResponder.AssertExpectations(t)
}

func TestUpdateMPinV2_ValidationError_SameMPin(t *testing.T) {
	handler, _, _, mockResponder := setupTestHandler()

	// Setup test data - same current and new MPIN
	updateMPinReq := requests.UpdateMPinRequest{
		CurrentMPin: "1234",
		NewMPin:     "1234", // Same as current
	}

	// Setup mocks
	mockResponder.On("SendValidationError", mock.Anything, []string{"new mPin must be different from current mPin"}).Return()

	// Create request
	reqBody, _ := json.Marshal(updateMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/update-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context with user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	// Execute
	handler.UpdateMPinV2(c)

	// Verify
	mockResponder.AssertExpectations(t)
}

func TestUpdateMPinV2_InvalidCurrentMPin(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	updateMPinReq := requests.UpdateMPinRequest{
		CurrentMPin: "9999", // Wrong current MPIN
		NewMPin:     "5678",
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &updateMPinReq).Return(nil)
	mockUserService.On("UpdateMPin", mock.Anything, "user-123", "9999", "5678").Return(errors.NewUnauthorizedError("invalid current mPin"))
	mockResponder.On("SendError", mock.Anything, http.StatusUnauthorized, "Invalid current mPin", mock.AnythingOfType("*errors.UnauthorizedError")).Return()

	// Create request
	reqBody, _ := json.Marshal(updateMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/update-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context with user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	// Execute
	handler.UpdateMPinV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestUpdateMPinV2_MPinNotSet(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	updateMPinReq := requests.UpdateMPinRequest{
		CurrentMPin: "1234",
		NewMPin:     "5678",
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &updateMPinReq).Return(nil)
	mockUserService.On("UpdateMPin", mock.Anything, "user-123", "1234", "5678").Return(errors.NewNotFoundError("mPin not set for user"))
	mockResponder.On("SendError", mock.Anything, http.StatusNotFound, "User not found or mPin not set", mock.AnythingOfType("*errors.NotFoundError")).Return()

	// Create request
	reqBody, _ := json.Marshal(updateMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/update-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context with user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	// Execute
	handler.UpdateMPinV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestUpdateMPinV2_ValidationError_InvalidFormat(t *testing.T) {
	handler, _, _, mockResponder := setupTestHandler()

	// Setup test data - invalid MPIN format
	updateMPinReq := requests.UpdateMPinRequest{
		CurrentMPin: "12", // Too short
		NewMPin:     "5678",
	}

	// Setup mocks
	mockResponder.On("SendValidationError", mock.Anything, []string{"current mPin must be 4 or 6 digits"}).Return()

	// Create request
	reqBody, _ := json.Marshal(updateMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/update-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context with user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	// Execute
	handler.UpdateMPinV2(c)

	// Verify
	mockResponder.AssertExpectations(t)
}

func TestUpdateMPinV2_InternalError(t *testing.T) {
	handler, mockUserService, mockValidator, mockResponder := setupTestHandler()

	// Setup test data
	updateMPinReq := requests.UpdateMPinRequest{
		CurrentMPin: "1234",
		NewMPin:     "5678",
	}

	// Setup mocks
	mockValidator.On("ValidateStruct", &updateMPinReq).Return(nil)
	mockUserService.On("UpdateMPin", mock.Anything, "user-123", "1234", "5678").Return(assert.AnError)
	mockResponder.On("SendInternalError", mock.Anything, assert.AnError).Return()

	// Create request
	reqBody, _ := json.Marshal(updateMPinReq)
	req := httptest.NewRequest("POST", "/v2/auth/update-mpin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup Gin context with user ID
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	// Execute
	handler.UpdateMPinV2(c)

	// Verify
	mockUserService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
