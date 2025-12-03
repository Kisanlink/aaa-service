//go:build integration
// +build integration

package organizations

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	groupRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/groups"
	groupResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/groups"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockOrganizationService is a mock implementation of the OrganizationService interface
type MockOrganizationService struct {
	mock.Mock
}

func (m *MockOrganizationService) CreateOrganization(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) UpdateOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) DeleteOrganization(ctx context.Context, orgID string, deletedBy string) error {
	args := m.Called(ctx, orgID, deletedBy)
	return args.Error(0)
}

func (m *MockOrganizationService) ListOrganizations(ctx context.Context, limit, offset int, includeInactive bool, orgType string) ([]interface{}, error) {
	args := m.Called(ctx, limit, offset, includeInactive, orgType)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockOrganizationService) CountOrganizations(ctx context.Context, includeInactive bool, orgType string) (int64, error) {
	args := m.Called(ctx, includeInactive, orgType)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrganizationService) GetOrganizationHierarchy(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) ActivateOrganization(ctx context.Context, orgID string) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

func (m *MockOrganizationService) DeactivateOrganization(ctx context.Context, orgID string) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

func (m *MockOrganizationService) GetOrganizationStats(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) GetOrganizationGroups(ctx context.Context, orgID string, limit, offset int, includeInactive bool) (interface{}, error) {
	args := m.Called(ctx, orgID, limit, offset, includeInactive)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) CreateGroupInOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) GetGroupInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) UpdateGroupInOrganization(ctx context.Context, orgID, groupID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) DeleteGroupInOrganization(ctx context.Context, orgID, groupID string, deletedBy string) error {
	args := m.Called(ctx, orgID, groupID, deletedBy)
	return args.Error(0)
}

func (m *MockOrganizationService) GetGroupHierarchyInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) AddUserToGroupInOrganization(ctx context.Context, orgID, groupID, userID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, userID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) RemoveUserFromGroupInOrganization(ctx context.Context, orgID, groupID, userID string, removedBy string) error {
	args := m.Called(ctx, orgID, groupID, userID, removedBy)
	return args.Error(0)
}

func (m *MockOrganizationService) GetGroupUsersInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, orgID, userID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) AssignRoleToGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, roleID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) RemoveRoleFromGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, removedBy string) error {
	args := m.Called(ctx, orgID, groupID, roleID, removedBy)
	return args.Error(0)
}

func (m *MockOrganizationService) GetGroupRolesInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockOrganizationService) GetUserEffectiveRolesInOrganization(ctx context.Context, orgID, userID string) (interface{}, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Get(0), args.Error(1)
}

// MockGroupService is a mock implementation of the GroupService interface
type MockGroupService struct {
	mock.Mock
}

func (m *MockGroupService) CreateGroup(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) GetGroup(ctx context.Context, groupID string) (interface{}, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) UpdateGroup(ctx context.Context, groupID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, groupID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) DeleteGroup(ctx context.Context, groupID string, deletedBy string) error {
	args := m.Called(ctx, groupID, deletedBy)
	return args.Error(0)
}

func (m *MockGroupService) ListGroups(ctx context.Context, limit, offset int, organizationID string, includeInactive bool) (interface{}, error) {
	args := m.Called(ctx, limit, offset, organizationID, includeInactive)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) AddMemberToGroup(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) RemoveMemberFromGroup(ctx context.Context, groupID, principalID string, removedBy string) error {
	args := m.Called(ctx, groupID, principalID, removedBy)
	return args.Error(0)
}

func (m *MockGroupService) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, groupID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) AssignRoleToGroup(ctx context.Context, groupID, roleID, assignedBy string) (interface{}, error) {
	args := m.Called(ctx, groupID, roleID, assignedBy)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) RemoveRoleFromGroup(ctx context.Context, groupID, roleID string) error {
	args := m.Called(ctx, groupID, roleID)
	return args.Error(0)
}

func (m *MockGroupService) GetGroupRoles(ctx context.Context, groupID string) (interface{}, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) GetUserEffectiveRoles(ctx context.Context, orgID, userID string) (interface{}, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) CountGroups(ctx context.Context, organizationID string, includeInactive bool) (int64, error) {
	args := m.Called(ctx, organizationID, includeInactive)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGroupService) CountGroupMembers(ctx context.Context, groupID string) (int64, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGroupService) GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, orgID, userID, limit, offset)
	return args.Get(0), args.Error(1)
}

// TestAssignRoleToGroupInOrganization_ValidationErrors tests validation scenarios
func TestAssignRoleToGroupInOrganization_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		orgID          string
		groupID        string
		requestBody    interface{}
		expectedStatus int
		setupAuth      bool
	}{
		{
			name:           "missing organization ID",
			orgID:          "",
			groupID:        "group-456",
			requestBody:    groupRequests.AssignRoleToGroupRequest{RoleID: "role-789"},
			expectedStatus: http.StatusBadRequest,
			setupAuth:      true,
		},
		{
			name:           "missing group ID",
			orgID:          "org-123",
			groupID:        "",
			requestBody:    groupRequests.AssignRoleToGroupRequest{RoleID: "role-789"},
			expectedStatus: http.StatusBadRequest,
			setupAuth:      true,
		},
		{
			name:           "missing user authentication",
			orgID:          "org-123",
			groupID:        "group-456",
			requestBody:    groupRequests.AssignRoleToGroupRequest{RoleID: "role-789"},
			expectedStatus: http.StatusUnauthorized,
			setupAuth:      false,
		},
		{
			name:           "invalid JSON body",
			orgID:          "org-123",
			groupID:        "group-456",
			requestBody:    "invalid-json",
			expectedStatus: http.StatusBadRequest,
			setupAuth:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock services
			mockOrgService := &MockOrganizationService{}
			mockGroupService := &MockGroupService{}
			logger := zap.NewNop()

			// Create a real responder instead of mocking it
			responder := &realResponder{}
			handler := NewOrganizationHandler(mockOrgService, mockGroupService, logger, responder)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request body
			var bodyBytes []byte
			if tt.requestBody != nil {
				if str, ok := tt.requestBody.(string); ok && str == "invalid-json" {
					bodyBytes = []byte("invalid-json")
				} else {
					bodyBytes, _ = json.Marshal(tt.requestBody)
				}
			}

			c.Request = httptest.NewRequest("POST", "/api/v1/organizations/"+tt.orgID+"/groups/"+tt.groupID+"/roles", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = []gin.Param{
				{Key: "orgId", Value: tt.orgID},
				{Key: "groupId", Value: tt.groupID},
			}

			// Setup authentication if required
			if tt.setupAuth {
				c.Set("user_id", "test-user-id")
			}

			// Execute
			handler.AssignRoleToGroupInOrganization(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRemoveRoleFromGroupInOrganization_ValidationErrors tests validation scenarios
func TestRemoveRoleFromGroupInOrganization_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		orgID          string
		groupID        string
		roleID         string
		expectedStatus int
		setupAuth      bool
	}{
		{
			name:           "missing organization ID",
			orgID:          "",
			groupID:        "group-456",
			roleID:         "role-789",
			expectedStatus: http.StatusBadRequest,
			setupAuth:      true,
		},
		{
			name:           "missing group ID",
			orgID:          "org-123",
			groupID:        "",
			roleID:         "role-789",
			expectedStatus: http.StatusBadRequest,
			setupAuth:      true,
		},
		{
			name:           "missing role ID",
			orgID:          "org-123",
			groupID:        "group-456",
			roleID:         "",
			expectedStatus: http.StatusBadRequest,
			setupAuth:      true,
		},
		{
			name:           "missing user authentication",
			orgID:          "org-123",
			groupID:        "group-456",
			roleID:         "role-789",
			expectedStatus: http.StatusUnauthorized,
			setupAuth:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock services
			mockOrgService := &MockOrganizationService{}
			mockGroupService := &MockGroupService{}
			logger := zap.NewNop()

			// Create a real responder instead of mocking it
			responder := &realResponder{}
			handler := NewOrganizationHandler(mockOrgService, mockGroupService, logger, responder)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("DELETE", "/api/v1/organizations/"+tt.orgID+"/groups/"+tt.groupID+"/roles/"+tt.roleID, nil)
			c.Params = []gin.Param{
				{Key: "orgId", Value: tt.orgID},
				{Key: "groupId", Value: tt.groupID},
				{Key: "roleId", Value: tt.roleID},
			}

			// Setup authentication if required
			if tt.setupAuth {
				c.Set("user_id", "test-user-id")
			}

			// Execute
			handler.RemoveRoleFromGroupInOrganization(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetGroupRolesInOrganization_ValidationErrors tests validation scenarios
func TestGetGroupRolesInOrganization_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		orgID          string
		groupID        string
		expectedStatus int
	}{
		{
			name:           "missing organization ID",
			orgID:          "",
			groupID:        "group-456",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing group ID",
			orgID:          "org-123",
			groupID:        "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock services
			mockOrgService := &MockOrganizationService{}
			mockGroupService := &MockGroupService{}
			logger := zap.NewNop()

			// Create a real responder instead of mocking it
			responder := &realResponder{}
			handler := NewOrganizationHandler(mockOrgService, mockGroupService, logger, responder)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/api/v1/organizations/"+tt.orgID+"/groups/"+tt.groupID+"/roles", nil)
			c.Params = []gin.Param{
				{Key: "orgId", Value: tt.orgID},
				{Key: "groupId", Value: tt.groupID},
			}

			// Execute
			handler.GetGroupRolesInOrganization(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestOrganizationBoundaryValidation tests that role-group assignments are validated within organization boundaries
func TestOrganizationBoundaryValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("group not in organization", func(t *testing.T) {
		// Create mock services
		mockOrgService := &MockOrganizationService{}
		mockGroupService := &MockGroupService{}
		logger := zap.NewNop()

		// Setup mocks
		mockOrgService.On("GetOrganization", mock.Anything, "org-123").Return(map[string]interface{}{
			"id":   "org-123",
			"name": "Test Organization",
		}, nil)

		mockGroupService.On("GetGroup", mock.Anything, "group-456").Return(map[string]interface{}{
			"id":              "group-456",
			"name":            "Test Group",
			"organization_id": "different-org-456", // Different organization
		}, nil)

		// Create a real responder
		responder := &realResponder{}
		handler := NewOrganizationHandler(mockOrgService, mockGroupService, logger, responder)

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := groupRequests.AssignRoleToGroupRequest{RoleID: "role-789"}
		bodyBytes, _ := json.Marshal(requestBody)

		c.Request = httptest.NewRequest("POST", "/api/v1/organizations/org-123/groups/group-456/roles", bytes.NewBuffer(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "orgId", Value: "org-123"},
			{Key: "groupId", Value: "group-456"},
		}
		c.Set("user_id", "test-user-id")

		// Execute
		handler.AssignRoleToGroupInOrganization(c)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockOrgService.AssertExpectations(t)
		mockGroupService.AssertExpectations(t)
	})
}

// TestSuccessfulRoleGroupOperations tests successful scenarios with proper mocking
func TestSuccessfulRoleGroupOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful role assignment", func(t *testing.T) {
		// Create mock services
		mockOrgService := &MockOrganizationService{}
		mockGroupService := &MockGroupService{}
		logger := zap.NewNop()

		// Setup mocks for successful flow
		mockOrgService.On("GetOrganization", mock.Anything, "org-123").Return(map[string]interface{}{
			"id":   "org-123",
			"name": "Test Organization",
		}, nil)

		mockGroupService.On("GetGroup", mock.Anything, "group-456").Return(map[string]interface{}{
			"id":              "group-456",
			"name":            "Test Group",
			"organization_id": "org-123",
		}, nil)

		mockGroupService.On("AssignRoleToGroup", mock.Anything, "group-456", "role-789", "test-user-id").Return(&groupResponses.GroupRoleResponse{
			GroupID:        "group-456",
			OrganizationID: "org-123",
			Role: groupResponses.RoleDetail{
				ID:   "role-789",
				Name: "Test Role",
			},
			AssignedBy: "test-user-id",
			IsActive:   true,
		}, nil)

		// Create a real responder
		responder := &realResponder{}
		handler := NewOrganizationHandler(mockOrgService, mockGroupService, logger, responder)

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := groupRequests.AssignRoleToGroupRequest{RoleID: "role-789"}
		bodyBytes, _ := json.Marshal(requestBody)

		c.Request = httptest.NewRequest("POST", "/api/v1/organizations/org-123/groups/group-456/roles", bytes.NewBuffer(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "orgId", Value: "org-123"},
			{Key: "groupId", Value: "group-456"},
		}
		c.Set("user_id", "test-user-id")

		// Execute
		handler.AssignRoleToGroupInOrganization(c)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		mockOrgService.AssertExpectations(t)
		mockGroupService.AssertExpectations(t)
	})

	t.Run("successful role removal", func(t *testing.T) {
		// Create mock services
		mockOrgService := &MockOrganizationService{}
		mockGroupService := &MockGroupService{}
		logger := zap.NewNop()

		// Setup mocks for successful flow
		mockOrgService.On("GetOrganization", mock.Anything, "org-123").Return(map[string]interface{}{
			"id":   "org-123",
			"name": "Test Organization",
		}, nil)

		mockGroupService.On("GetGroup", mock.Anything, "group-456").Return(map[string]interface{}{
			"id":              "group-456",
			"name":            "Test Group",
			"organization_id": "org-123",
		}, nil)

		mockGroupService.On("RemoveRoleFromGroup", mock.Anything, "group-456", "role-789").Return(nil)

		// Create a real responder
		responder := &realResponder{}
		handler := NewOrganizationHandler(mockOrgService, mockGroupService, logger, responder)

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("DELETE", "/api/v1/organizations/org-123/groups/group-456/roles/role-789", nil)
		c.Params = []gin.Param{
			{Key: "orgId", Value: "org-123"},
			{Key: "groupId", Value: "group-456"},
			{Key: "roleId", Value: "role-789"},
		}
		c.Set("user_id", "test-user-id")

		// Execute
		handler.RemoveRoleFromGroupInOrganization(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockOrgService.AssertExpectations(t)
		mockGroupService.AssertExpectations(t)
	})

	t.Run("successful get group roles", func(t *testing.T) {
		// Create mock services
		mockOrgService := &MockOrganizationService{}
		mockGroupService := &MockGroupService{}
		logger := zap.NewNop()

		// Setup mocks for successful flow
		mockOrgService.On("GetOrganization", mock.Anything, "org-123").Return(map[string]interface{}{
			"id":   "org-123",
			"name": "Test Organization",
		}, nil)

		mockGroupService.On("GetGroup", mock.Anything, "group-456").Return(map[string]interface{}{
			"id":              "group-456",
			"name":            "Test Group",
			"organization_id": "org-123",
		}, nil)

		roles := []interface{}{
			map[string]interface{}{
				"id":   "role-1",
				"name": "Admin Role",
			},
			map[string]interface{}{
				"id":   "role-2",
				"name": "User Role",
			},
		}
		mockGroupService.On("GetGroupRoles", mock.Anything, "group-456").Return(roles, nil)

		// Create a real responder
		responder := &realResponder{}
		handler := NewOrganizationHandler(mockOrgService, mockGroupService, logger, responder)

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/api/v1/organizations/org-123/groups/group-456/roles", nil)
		c.Params = []gin.Param{
			{Key: "orgId", Value: "org-123"},
			{Key: "groupId", Value: "group-456"},
		}

		// Execute
		handler.GetGroupRolesInOrganization(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockOrgService.AssertExpectations(t)
		mockGroupService.AssertExpectations(t)
	})
}

// realResponder is a simple implementation of the Responder interface for testing
type realResponder struct{}

func (r *realResponder) SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{"data": data})
}

func (r *realResponder) SendError(c *gin.Context, statusCode int, message string, err error) {
	c.JSON(statusCode, gin.H{"error": message})
}

func (r *realResponder) SendValidationError(c *gin.Context, errors []string) {
	c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
}

func (r *realResponder) SendInternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
