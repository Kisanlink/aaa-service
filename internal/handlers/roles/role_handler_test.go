//nolint:typecheck
package roles

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests/roles"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRoleService is a mock implementation of the RoleService interface
type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) CreateRole(ctx context.Context, role *models.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleService) GetRoleByID(ctx context.Context, roleID string) (*models.Role, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleService) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	args := m.Called(ctx, name)
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
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleService) CountRoles(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoleService) SearchRoles(ctx context.Context, query string, limit, offset int) ([]*models.Role, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleService) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleService) GetUserRoles(ctx context.Context, userID string) ([]*models.UserRole, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.UserRole), args.Error(1)
}

func (m *MockRoleService) GetRoleHierarchy(ctx context.Context) ([]*models.Role, error) {
	return nil, nil
}

func (m *MockRoleService) AddChildRole(ctx context.Context, parentRoleID, childRoleID string) error {
	return nil
}

func (m *MockRoleService) RemoveChildRole(ctx context.Context, parentRoleID, childRoleID string) error {
	return nil
}

func (m *MockRoleService) GetRoleWithChildren(ctx context.Context, roleID string) (*models.Role, error) {
	return nil, nil
}

func (m *MockRoleService) ValidateRoleAssignment(ctx context.Context, userID, roleID string) error {
	return nil
}

func (m *MockRoleService) AssignRole(ctx context.Context, userID, roleID string) error {
	return nil
}

func (m *MockRoleService) RemoveRole(ctx context.Context, userID, roleID string) error {
	return nil
}

func (m *MockRoleService) HardDeleteRole(ctx context.Context, roleID string) error {
	return nil
}

// MockValidator is a mock implementation of the Validator interface
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

// MockResponder is a mock implementation of the Responder interface
type MockResponder struct {
	mock.Mock
}

func (m *MockResponder) SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	m.Called(c, statusCode, data)
	c.JSON(statusCode, data)
}

func (m *MockResponder) SendError(c *gin.Context, statusCode int, message string, err error) {
	m.Called(c, statusCode, message, err)
	c.JSON(statusCode, gin.H{"error": message})
}

func (m *MockResponder) SendValidationError(c *gin.Context, errors []string) {
	m.Called(c, errors)
	c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
}

func (m *MockResponder) SendInternalError(c *gin.Context, err error) {
	m.Called(c, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}

func (m *MockResponder) SendPaginatedResponse(c *gin.Context, data interface{}, total, limit, offset int) {
	m.Called(c, data, total, limit, offset)
	c.JSON(http.StatusOK, gin.H{"data": data, "pagination": gin.H{"total": total, "limit": limit, "offset": offset}})
}

// MockAuditService is a mock implementation of the AuditService
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogUserAction(ctx context.Context, userID, action, resource, resourceID string, details map[string]interface{}) {
	m.Called(ctx, userID, action, resource, resourceID, details)
}

func (m *MockAuditService) LogUserActionWithError(ctx context.Context, userID, action, resource, resourceID string, err error, details map[string]interface{}) {
	m.Called(ctx, userID, action, resource, resourceID, err, details)
}

func (m *MockAuditService) LogRoleChange(ctx context.Context, userID, action, roleID string, details map[string]interface{}) {
	m.Called(ctx, userID, action, roleID, details)
}

// Add other required methods to satisfy the interface
func (m *MockAuditService) LogAPIAccess(ctx context.Context, userID, method, endpoint, ipAddress, userAgent string, success bool, err error) {
	m.Called(ctx, userID, method, endpoint, ipAddress, userAgent, success, err)
}

func (m *MockAuditService) LogAccessDenied(ctx context.Context, userID, action, resource, resourceID, reason string) {
	m.Called(ctx, userID, action, resource, resourceID, reason)
}

func (m *MockAuditService) LogPermissionChange(ctx context.Context, userID, action, resource, resourceID, permission string, details map[string]interface{}) {
	m.Called(ctx, userID, action, resource, resourceID, permission, details)
}

func (m *MockAuditService) LogDataAccess(ctx context.Context, userID, action, resource, resourceID string, oldData, newData map[string]interface{}) {
	m.Called(ctx, userID, action, resource, resourceID, oldData, newData)
}

func (m *MockAuditService) LogSecurityEvent(ctx context.Context, userID, action, resource string, success bool, details map[string]interface{}) {
	m.Called(ctx, userID, action, resource, success, details)
}

func (m *MockAuditService) LogSystemEvent(ctx context.Context, action, resource string, success bool, details map[string]interface{}) {
	m.Called(ctx, action, resource, success, details)
}

func setupTestHandler() (*RoleHandler, *MockRoleService, *MockValidator, *MockResponder, *MockAuditService) {
	mockRoleService := &MockRoleService{}
	mockValidator := &MockValidator{}
	mockResponder := &MockResponder{}
	mockAuditService := &MockAuditService{}
	logger := zap.NewNop()

	handler := NewRoleHandler(mockRoleService, mockValidator, mockResponder, mockAuditService, logger)
	return handler, mockRoleService, mockValidator, mockResponder, mockAuditService
}

func TestRoleHandler_AssignRole_Success(t *testing.T) {
	handler, mockRoleService, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup test data
	userID := "user-123"
	roleID := "role-456"
	actorID := "admin-789"
	role := models.NewRole("Test Role", "Test role description", models.RoleScopeOrg)
	role.SetID(roleID)

	// Setup request
	assignRequest := roles.AssignRoleRequest{
		RoleID: roleID,
	}
	requestBody, _ := json.Marshal(assignRequest)

	// Setup expectations
	mockRoleService.On("AssignRoleToUser", mock.Anything, userID, roleID).Return(nil)
	mockRoleService.On("GetRoleByID", mock.Anything, roleID).Return(role, nil)
	mockAuditService.On("LogUserAction", mock.Anything, actorID, "assign_role", "user", userID, mock.Anything).Return()
	mockAuditService.On("LogRoleChange", mock.Anything, actorID, "assign", roleID, mock.Anything).Return()
	mockResponder.On("SendSuccess", mock.Anything, http.StatusOK, mock.AnythingOfType("*roles.AssignRoleResponse")).Return()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/users/"+userID+"/roles", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}
	// Set authenticated user context
	c.Set("user_id", actorID)

	// Execute
	handler.AssignRole(c)

	// Assertions
	mockRoleService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_AssignRole_Unauthenticated(t *testing.T) {
	handler, _, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup request
	assignRequest := roles.AssignRoleRequest{
		RoleID: "role-456",
	}
	requestBody, _ := json.Marshal(assignRequest)

	// Setup expectations
	mockAuditService.On("LogAccessDenied", mock.Anything, "anonymous", "assign_role", "user", "user-123", "not authenticated").Return()
	mockResponder.On("SendError", mock.Anything, http.StatusUnauthorized, "Authentication required", mock.AnythingOfType("*errors.errorString")).Return()

	// Setup Gin context without authentication
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/users/user-123/roles", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "user-123"}}
	// Don't set user_id to simulate unauthenticated request

	// Execute
	handler.AssignRole(c)

	// Assertions
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_AssignRole_InvalidUserID(t *testing.T) {
	handler, _, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup request with empty user ID
	assignRequest := roles.AssignRoleRequest{
		RoleID: "role-456",
	}
	requestBody, _ := json.Marshal(assignRequest)
	actorID := "admin-789"

	// Setup expectations
	mockAuditService.On("LogUserActionWithError", mock.Anything, actorID, "assign_role", "user", "", mock.Anything, mock.Anything).Return()
	mockResponder.On("SendValidationError", mock.Anything, []string{"user ID is required"}).Return()

	// Setup Gin context with empty user ID
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/users//roles", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: ""}}
	c.Set("user_id", actorID)

	// Execute
	handler.AssignRole(c)

	// Assertions
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_AssignRole_RoleNotFound(t *testing.T) {
	handler, mockRoleService, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup test data
	userID := "user-123"
	roleID := "role-456"
	actorID := "admin-789"

	// Setup request
	assignRequest := roles.AssignRoleRequest{
		RoleID: roleID,
	}
	requestBody, _ := json.Marshal(assignRequest)

	// Setup expectations
	mockRoleService.On("AssignRoleToUser", mock.Anything, userID, roleID).Return(errors.New("role not found"))
	mockAuditService.On("LogUserActionWithError", mock.Anything, actorID, "assign_role", "user", userID, mock.Anything, mock.Anything).Return()
	mockResponder.On("SendError", mock.Anything, http.StatusNotFound, "Role not found", mock.AnythingOfType("*errors.errorString")).Return()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/users/"+userID+"/roles", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}
	c.Set("user_id", actorID)

	// Execute
	handler.AssignRole(c)

	// Assertions
	mockRoleService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_AssignRole_RoleAlreadyAssigned(t *testing.T) {
	handler, mockRoleService, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup test data
	userID := "user-123"
	roleID := "role-456"
	actorID := "admin-789"

	// Setup request
	assignRequest := roles.AssignRoleRequest{
		RoleID: roleID,
	}
	requestBody, _ := json.Marshal(assignRequest)

	// Setup expectations
	mockRoleService.On("AssignRoleToUser", mock.Anything, userID, roleID).Return(errors.New("role already assigned to user"))
	mockAuditService.On("LogUserActionWithError", mock.Anything, actorID, "assign_role", "user", userID, mock.Anything, mock.Anything).Return()
	mockResponder.On("SendError", mock.Anything, http.StatusConflict, "Role already assigned to user", mock.AnythingOfType("*errors.errorString")).Return()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/users/"+userID+"/roles", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}
	c.Set("user_id", actorID)

	// Execute
	handler.AssignRole(c)

	// Assertions
	mockRoleService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_RemoveRole_Success(t *testing.T) {
	handler, mockRoleService, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup test data
	userID := "user-123"
	roleID := "role-456"
	actorID := "admin-789"

	// Setup expectations
	mockRoleService.On("GetRoleByID", mock.Anything, roleID).Return(&models.Role{}, errors.New("not found"))
	mockRoleService.On("RemoveRoleFromUser", mock.Anything, userID, roleID).Return(nil)
	mockAuditService.On("LogUserAction", mock.Anything, actorID, "remove_role", "user", userID, mock.Anything).Return()
	mockAuditService.On("LogRoleChange", mock.Anything, actorID, "remove", roleID, mock.Anything).Return()
	mockResponder.On("SendSuccess", mock.Anything, http.StatusOK, mock.AnythingOfType("*roles.RemoveRoleResponse")).Return()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/users/"+userID+"/roles/"+roleID, nil)
	c.Params = gin.Params{
		gin.Param{Key: "id", Value: userID},
		gin.Param{Key: "role_id", Value: roleID},
	}
	c.Set("user_id", actorID)

	// Execute
	handler.RemoveRole(c)

	// Assertions
	mockRoleService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_RemoveRole_Unauthenticated(t *testing.T) {
	handler, _, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup expectations
	mockAuditService.On("LogAccessDenied", mock.Anything, "anonymous", "remove_role", "user", "user-123", "not authenticated").Return()
	mockResponder.On("SendError", mock.Anything, http.StatusUnauthorized, "Authentication required", mock.AnythingOfType("*errors.errorString")).Return()

	// Setup Gin context without authentication
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/users/user-123/roles/role-456", nil)
	c.Params = gin.Params{
		gin.Param{Key: "id", Value: "user-123"},
		gin.Param{Key: "role_id", Value: "role-456"},
	}
	// Don't set user_id to simulate unauthenticated request

	// Execute
	handler.RemoveRole(c)

	// Assertions
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_RemoveRole_InvalidUserID(t *testing.T) {
	handler, _, _, mockResponder, mockAuditService := setupTestHandler()

	actorID := "admin-789"

	// Setup expectations
	mockAuditService.On("LogUserActionWithError", mock.Anything, actorID, "remove_role", "user", "", mock.Anything, mock.Anything).Return()
	mockResponder.On("SendValidationError", mock.Anything, []string{"user ID is required"}).Return()

	// Setup Gin context with empty user ID
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/users//roles/role-456", nil)
	c.Params = gin.Params{
		gin.Param{Key: "id", Value: ""},
		gin.Param{Key: "role_id", Value: "role-456"},
	}
	c.Set("user_id", actorID)

	// Execute
	handler.RemoveRole(c)

	// Assertions
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_RemoveRole_InvalidRoleID(t *testing.T) {
	handler, _, _, mockResponder, mockAuditService := setupTestHandler()

	actorID := "admin-789"

	// Setup expectations
	mockAuditService.On("LogUserActionWithError", mock.Anything, actorID, "remove_role", "user", "user-123", mock.Anything, mock.Anything).Return()
	mockResponder.On("SendValidationError", mock.Anything, []string{"role ID is required"}).Return()

	// Setup Gin context with empty role ID
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/users/user-123/roles/", nil)
	c.Params = gin.Params{
		gin.Param{Key: "id", Value: "user-123"},
		gin.Param{Key: "role_id", Value: ""},
	}
	c.Set("user_id", actorID)

	// Execute
	handler.RemoveRole(c)

	// Assertions
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_RemoveRole_AssignmentNotFound(t *testing.T) {
	handler, mockRoleService, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup test data
	userID := "user-123"
	roleID := "role-456"
	actorID := "admin-789"

	// Setup expectations
	mockRoleService.On("GetRoleByID", mock.Anything, roleID).Return(&models.Role{}, errors.New("not found"))
	mockRoleService.On("RemoveRoleFromUser", mock.Anything, userID, roleID).Return(errors.New("user role assignment not found"))
	mockAuditService.On("LogUserActionWithError", mock.Anything, actorID, "remove_role", "user", userID, mock.Anything, mock.Anything).Return()
	mockResponder.On("SendError", mock.Anything, http.StatusNotFound, "Role assignment not found", mock.AnythingOfType("*errors.errorString")).Return()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/users/"+userID+"/roles/"+roleID, nil)
	c.Params = gin.Params{
		gin.Param{Key: "id", Value: userID},
		gin.Param{Key: "role_id", Value: roleID},
	}
	c.Set("user_id", actorID)

	// Execute
	handler.RemoveRole(c)

	// Assertions
	mockRoleService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_AssignRole_InvalidJSON(t *testing.T) {
	handler, _, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup invalid JSON
	invalidJSON := `{"role_id": }`
	actorID := "admin-789"

	// Setup expectations
	mockAuditService.On("LogUserActionWithError", mock.Anything, actorID, "assign_role", "user", "user-123", mock.Anything, mock.Anything).Return()
	mockResponder.On("SendValidationError", mock.Anything, mock.AnythingOfType("[]string")).Return()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/users/user-123/roles", bytes.NewBufferString(invalidJSON))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "user-123"}}
	c.Set("user_id", actorID)

	// Execute
	handler.AssignRole(c)

	// Assertions
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_AssignRole_EmptyRoleID(t *testing.T) {
	handler, _, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup request with empty role ID
	assignRequest := roles.AssignRoleRequest{
		RoleID: "",
	}
	requestBody, _ := json.Marshal(assignRequest)
	actorID := "admin-789"

	// Setup expectations
	mockAuditService.On("LogUserActionWithError", mock.Anything, actorID, "assign_role", "user", "user-123", mock.Anything, mock.Anything).Return()
	mockResponder.On("SendValidationError", mock.Anything, mock.AnythingOfType("[]string")).Return()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/users/user-123/roles", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "user-123"}}
	c.Set("user_id", actorID)

	// Execute
	handler.AssignRole(c)

	// Assertions
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_AssignRole_ServiceError(t *testing.T) {
	handler, mockRoleService, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup test data
	userID := "user-123"
	roleID := "role-456"
	actorID := "admin-789"

	// Setup request
	assignRequest := roles.AssignRoleRequest{
		RoleID: roleID,
	}
	requestBody, _ := json.Marshal(assignRequest)

	// Setup expectations
	mockRoleService.On("AssignRoleToUser", mock.Anything, userID, roleID).Return(errors.New("database connection error"))
	mockAuditService.On("LogUserActionWithError", mock.Anything, actorID, "assign_role", "user", userID, mock.Anything, mock.Anything).Return()
	mockResponder.On("SendInternalError", mock.Anything, mock.AnythingOfType("*errors.errorString")).Return()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/users/"+userID+"/roles", bytes.NewBuffer(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: userID}}
	c.Set("user_id", actorID)

	// Execute
	handler.AssignRole(c)

	// Assertions
	mockRoleService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}

func TestRoleHandler_RemoveRole_ServiceError(t *testing.T) {
	handler, mockRoleService, _, mockResponder, mockAuditService := setupTestHandler()

	// Setup test data
	userID := "user-123"
	roleID := "role-456"
	actorID := "admin-789"

	// Setup expectations
	mockRoleService.On("GetRoleByID", mock.Anything, roleID).Return(&models.Role{}, errors.New("not found"))
	mockRoleService.On("RemoveRoleFromUser", mock.Anything, userID, roleID).Return(errors.New("database connection error"))
	mockAuditService.On("LogUserActionWithError", mock.Anything, actorID, "remove_role", "user", userID, mock.Anything, mock.Anything).Return()
	mockResponder.On("SendInternalError", mock.Anything, mock.AnythingOfType("*errors.errorString")).Return()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/users/"+userID+"/roles/"+roleID, nil)
	c.Params = gin.Params{
		gin.Param{Key: "id", Value: userID},
		gin.Param{Key: "role_id", Value: roleID},
	}
	c.Set("user_id", actorID)

	// Execute
	handler.RemoveRole(c)

	// Assertions
	mockRoleService.AssertExpectations(t)
	mockResponder.AssertExpectations(t)
}
