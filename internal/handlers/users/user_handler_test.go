//nolint:typecheck
package users

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/entities/requests/users"
	userResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/users"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockUserService is a mock implementation of UserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req *users.CreateUserRequest) (*userResponses.UserResponse, error) {
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

func (m *MockUserService) UpdateUser(ctx context.Context, req *users.UpdateUserRequest) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserService) SoftDeleteUserWithCascade(ctx context.Context, userID, deletedBy string) error {
	args := m.Called(ctx, userID, deletedBy)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockUserService) SearchUsers(ctx context.Context, query string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockUserService) ValidateUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Add other required methods to satisfy the interface
func (m *MockUserService) GetUserByUsername(ctx context.Context, username string) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByMobileNumber(ctx context.Context, mobileNumber uint64) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, mobileNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, aadhaarNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) ListActiveUsers(ctx context.Context, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockUserService) DeductTokens(ctx context.Context, userID string, amount int) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

func (m *MockUserService) AddTokens(ctx context.Context, userID string, amount int) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

func (m *MockUserService) GetUserWithProfile(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserWithRoles(ctx context.Context, userID string) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) VerifyUserPassword(ctx context.Context, username, password string) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) VerifyUserPasswordByPhone(ctx context.Context, phoneNumber, countryCode, password string) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, phoneNumber, countryCode, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) SetMPin(ctx context.Context, userID, mPin, currentPassword string) error {
	args := m.Called(ctx, userID, mPin, currentPassword)
	return args.Error(0)
}

func (m *MockUserService) VerifyMPin(ctx context.Context, userID string, mPin string) error {
	args := m.Called(ctx, userID, mPin)
	return args.Error(0)
}

func (m *MockUserService) UpdateMPin(ctx context.Context, userID, currentMPin, newMPin string) error {
	args := m.Called(ctx, userID, currentMPin, newMPin)
	return args.Error(0)
}

func (m *MockUserService) VerifyUserCredentials(ctx context.Context, phone, countryCode string, password, mpin *string) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, phone, countryCode, password, mpin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (*userResponses.UserResponse, error) {
	args := m.Called(ctx, phoneNumber, countryCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userResponses.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserOrganizations(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return []map[string]interface{}{}, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockUserService) GetUserGroups(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return []map[string]interface{}{}, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// MockRoleService is a mock implementation of RoleService for testing
type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

// Add other required methods to satisfy the interface
func (m *MockRoleService) CreateRole(ctx context.Context, role *models.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleService) GetRoleByID(ctx context.Context, roleID string) (*models.Role, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleService) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleService) UpdateRole(ctx context.Context, role *models.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleService) DeleteRole(ctx context.Context, roleID string) error {
	args := m.Called(ctx, roleID)
	return args.Error(0)
}

func (m *MockRoleService) ListRoles(ctx context.Context, limit, offset int) ([]*models.Role, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleService) SearchRoles(ctx context.Context, query string, limit, offset int) ([]*models.Role, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleService) GetUserRoles(ctx context.Context, userID string) ([]*models.UserRole, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserRole), args.Error(1)
}

// MockValidator is a mock implementation of Validator for testing
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) ValidateStruct(s interface{}) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockValidator) ValidateUserID(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockValidator) ValidateEmail(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockValidator) ValidatePassword(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

func (m *MockValidator) ValidatePhoneNumber(phone string) error {
	args := m.Called(phone)
	return args.Error(0)
}

func (m *MockValidator) ValidateAadhaarNumber(aadhaar string) error {
	args := m.Called(aadhaar)
	return args.Error(0)
}

func (m *MockValidator) ParseListFilters(c *gin.Context) (interface{}, error) {
	args := m.Called(c)
	return args.Get(0), args.Error(1)
}

// MockResponder is a mock implementation of Responder for testing
type MockResponder struct {
	mock.Mock
}

func (m *MockResponder) SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	m.Called(c, statusCode, data)
	c.JSON(statusCode, gin.H{"success": true, "data": data})
}

func (m *MockResponder) SendError(c *gin.Context, statusCode int, message string, err error) {
	m.Called(c, statusCode, message, err)
	c.JSON(statusCode, gin.H{"success": false, "error": message})
}

func (m *MockResponder) SendValidationError(c *gin.Context, errors []string) {
	m.Called(c, errors)
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "errors": errors})
}

func (m *MockResponder) SendInternalError(c *gin.Context, err error) {
	m.Called(c, err)
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Internal server error"})
}

// Test helper function to create a test handler
func createTestHandler() (*UserHandler, *MockUserService, *MockRoleService, *MockValidator, *MockResponder) {
	mockUserService := &MockUserService{}
	mockRoleService := &MockRoleService{}
	mockValidator := &MockValidator{}
	mockResponder := &MockResponder{}
	logger := zap.NewNop()

	handler := NewUserHandler(
		mockUserService,
		mockRoleService,
		mockValidator,
		mockResponder,
		logger,
	)

	return handler, mockUserService, mockRoleService, mockValidator, mockResponder
}

// Test helper function to create a test context
func createTestContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	c.Request = httptest.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	return c, w
}

// TestDeleteUser_Success tests successful user deletion
func TestDeleteUser_Success(t *testing.T) {
	handler, mockUserService, _, mockValidator, mockResponder := createTestHandler()

	userID := "test-user-id"
	c, w := createTestContext("DELETE", "/users/"+userID, nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}

	// Set up mocks
	mockValidator.On("ValidateUserID", userID).Return(nil)
	mockUserService.On("SoftDeleteUserWithCascade", mock.Anything, userID, "system").Return(nil)
	mockResponder.On("SendSuccess", c, http.StatusOK, mock.MatchedBy(func(data interface{}) bool {
		response, ok := data.(map[string]interface{})
		return ok && response["message"] == "User deleted successfully" && response["user_id"] == userID
	}))

	// Execute
	handler.DeleteUser(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockValidator.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

// TestDeleteUser_EmptyUserID tests deletion with empty user ID
func TestDeleteUser_EmptyUserID(t *testing.T) {
	handler, _, _, _, mockResponder := createTestHandler()

	c, w := createTestContext("DELETE", "/users/", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: ""}}

	// Set up mocks
	mockResponder.On("SendValidationError", c, []string{"user ID is required"})

	// Execute
	handler.DeleteUser(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockResponder.AssertExpectations(t)
}

// TestDeleteUser_InvalidUserIDFormat tests deletion with invalid user ID format
func TestDeleteUser_InvalidUserIDFormat(t *testing.T) {
	handler, _, _, mockValidator, mockResponder := createTestHandler()

	userID := "invalid-id"
	c, w := createTestContext("DELETE", "/users/"+userID, nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}

	// Set up mocks
	mockValidator.On("ValidateUserID", userID).Return(errors.NewValidationError("invalid user ID format"))
	mockResponder.On("SendValidationError", c, []string{"invalid user ID format"})

	// Execute
	handler.DeleteUser(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockValidator.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

// TestDeleteUser_UserNotFound tests deletion of non-existent user
func TestDeleteUser_UserNotFound(t *testing.T) {
	handler, mockUserService, _, mockValidator, mockResponder := createTestHandler()

	userID := "non-existent-user"
	c, w := createTestContext("DELETE", "/users/"+userID, nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}

	// Set up mocks
	mockValidator.On("ValidateUserID", userID).Return(nil)
	mockUserService.On("SoftDeleteUserWithCascade", mock.Anything, userID, "system").Return(errors.NewNotFoundError("User not found"))
	mockResponder.On("SendError", c, http.StatusNotFound, "User not found", mock.AnythingOfType("*errors.NotFoundError"))

	// Execute
	handler.DeleteUser(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockValidator.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

// TestDeleteUser_ConflictError tests deletion with conflict (e.g., already deleted)
func TestDeleteUser_ConflictError(t *testing.T) {
	handler, mockUserService, _, mockValidator, mockResponder := createTestHandler()

	userID := "already-deleted-user"
	c, w := createTestContext("DELETE", "/users/"+userID, nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}

	// Set up mocks
	mockValidator.On("ValidateUserID", userID).Return(nil)
	mockUserService.On("SoftDeleteUserWithCascade", mock.Anything, userID, "system").Return(errors.NewConflictError("user is already deleted"))
	mockResponder.On("SendError", c, http.StatusConflict, "Cannot delete user due to constraints", mock.AnythingOfType("*errors.ConflictError"))

	// Execute
	handler.DeleteUser(c)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)
	mockValidator.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

// TestDeleteUser_ForbiddenError tests deletion with insufficient permissions
func TestDeleteUser_ForbiddenError(t *testing.T) {
	handler, mockUserService, _, mockValidator, mockResponder := createTestHandler()

	userID := "admin-user"
	c, w := createTestContext("DELETE", "/users/"+userID, nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}

	// Set up mocks
	mockValidator.On("ValidateUserID", userID).Return(nil)
	mockUserService.On("SoftDeleteUserWithCascade", mock.Anything, userID, "system").Return(errors.NewForbiddenError("cannot delete users with critical admin roles"))
	mockResponder.On("SendError", c, http.StatusForbidden, "Insufficient permissions to delete user", mock.AnythingOfType("*errors.ForbiddenError"))

	// Execute
	handler.DeleteUser(c)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockValidator.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

// TestDeleteUser_ValidationError tests deletion with validation error
func TestDeleteUser_ValidationError(t *testing.T) {
	handler, mockUserService, _, mockValidator, mockResponder := createTestHandler()

	userID := "test-user"
	c, w := createTestContext("DELETE", "/users/"+userID, nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}

	// Set up mocks
	mockValidator.On("ValidateUserID", userID).Return(nil)
	mockUserService.On("SoftDeleteUserWithCascade", mock.Anything, userID, "system").Return(errors.NewValidationError("users cannot delete themselves"))
	mockResponder.On("SendValidationError", c, []string{"users cannot delete themselves"})

	// Execute
	handler.DeleteUser(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockValidator.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

// TestDeleteUser_InternalError tests deletion with internal server error
func TestDeleteUser_InternalError(t *testing.T) {
	handler, mockUserService, _, mockValidator, mockResponder := createTestHandler()

	userID := "test-user"
	c, w := createTestContext("DELETE", "/users/"+userID, nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}

	// Set up mocks
	mockValidator.On("ValidateUserID", userID).Return(nil)
	mockUserService.On("SoftDeleteUserWithCascade", mock.Anything, userID, "system").Return(errors.NewInternalError(assert.AnError))
	mockResponder.On("SendInternalError", c, mock.AnythingOfType("*errors.InternalError"))

	// Execute
	handler.DeleteUser(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockValidator.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

// TestDeleteUser_WithActorFromContext tests deletion with actor ID from JWT context
func TestDeleteUser_WithActorFromContext(t *testing.T) {
	handler, mockUserService, _, mockValidator, mockResponder := createTestHandler()

	userID := "test-user-id"
	actorID := "admin-user-id"
	c, w := createTestContext("DELETE", "/users/"+userID, nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}

	// Set JWT claims in context
	c.Set("user_claims", map[string]interface{}{
		"user_id": actorID,
	})

	// Set up mocks
	mockValidator.On("ValidateUserID", userID).Return(nil)
	mockUserService.On("SoftDeleteUserWithCascade", mock.Anything, userID, actorID).Return(nil)
	mockResponder.On("SendSuccess", c, http.StatusOK, mock.MatchedBy(func(data interface{}) bool {
		response, ok := data.(map[string]interface{})
		return ok && response["deleted_by"] == actorID
	}))

	// Execute
	handler.DeleteUser(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockValidator.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}
