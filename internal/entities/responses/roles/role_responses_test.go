//nolint:typecheck
package roles

import (
	"testing"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"github.com/stretchr/testify/assert"
)

func TestNewAssignRoleResponse(t *testing.T) {
	// Create a test role
	role := &models.Role{
		BaseModel:   base.NewBaseModel("ROLE", hash.Medium),
		Name:        "Admin",
		Description: "Administrator role",
		IsActive:    true,
	}
	role.SetID("role-123")

	userID := "user-456"
	message := "Role assigned successfully"

	response := NewAssignRoleResponse(userID, role, message)

	assert.NotNil(t, response)
	assert.Equal(t, message, response.Message)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, "role-123", response.Role.ID)
	assert.Equal(t, "Admin", response.Role.Name)
	assert.Equal(t, "Administrator role", response.Role.Description)
	assert.True(t, response.Role.IsActive)
	assert.Equal(t, "AssignRoleResponse", response.GetType())
	assert.True(t, response.IsSuccess())
}

func TestNewRoleDetail(t *testing.T) {
	role := &models.Role{
		BaseModel:   base.NewBaseModel("ROLE", hash.Medium),
		Name:        "Manager",
		Description: "Manager role",
		IsActive:    false,
	}
	role.SetID("role-789")

	detail := NewRoleDetail(role)

	assert.Equal(t, "role-789", detail.ID)
	assert.Equal(t, "Manager", detail.Name)
	assert.Equal(t, "Manager role", detail.Description)
	assert.False(t, detail.IsActive)
}

func TestNewUserRoleDetail(t *testing.T) {
	// Create test role
	role := models.Role{
		BaseModel:   base.NewBaseModel("ROLE", hash.Medium),
		Name:        "Editor",
		Description: "Editor role",
		IsActive:    true,
	}
	role.SetID("role-456")

	// Create test user role
	userRole := &models.UserRole{
		BaseModel: base.NewBaseModel("USERROLE", hash.Small),
		UserID:    "user-123",
		RoleID:    "role-456",
		IsActive:  true,
		Role:      role,
	}
	userRole.SetID("userrole-789")

	detail := NewUserRoleDetail(userRole)

	assert.Equal(t, "userrole-789", detail.ID)
	assert.Equal(t, "user-123", detail.UserID)
	assert.Equal(t, "role-456", detail.RoleID)
	assert.True(t, detail.IsActive)
	assert.Equal(t, "role-456", detail.Role.ID)
	assert.Equal(t, "Editor", detail.Role.Name)
	assert.Equal(t, "Editor role", detail.Role.Description)
	assert.True(t, detail.Role.IsActive)
}

func TestNewRemoveRoleResponse(t *testing.T) {
	userID := "user-123"
	roleID := "role-456"
	message := "Role removed successfully"

	response := NewRemoveRoleResponse(userID, roleID, message)

	assert.NotNil(t, response)
	assert.Equal(t, message, response.Message)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, roleID, response.RoleID)
	assert.Equal(t, "RemoveRoleResponse", response.GetType())
	assert.True(t, response.IsSuccess())
}

func TestRoleErrorResponses(t *testing.T) {
	t.Run("NewRoleNotFoundError", func(t *testing.T) {
		roleID := "role-123"
		err := NewRoleNotFoundError(roleID)

		assert.Equal(t, RoleErrorTypeNotFound, err.ErrorType)
		assert.Equal(t, roleID, err.RoleID)
		assert.Equal(t, 404, err.ErrorResponse.Code)
		assert.Contains(t, err.ErrorResponse.Message, roleID)
		assert.False(t, err.IsSuccess())
	})

	t.Run("NewUserNotFoundError", func(t *testing.T) {
		userID := "user-456"
		err := NewUserNotFoundError(userID)

		assert.Equal(t, RoleErrorTypeUserNotFound, err.ErrorType)
		assert.Equal(t, userID, err.UserID)
		assert.Equal(t, 404, err.ErrorResponse.Code)
		assert.Contains(t, err.ErrorResponse.Message, userID)
	})

	t.Run("NewDuplicateRoleAssignmentError", func(t *testing.T) {
		userID := "user-123"
		roleID := "role-456"
		err := NewDuplicateRoleAssignmentError(userID, roleID)

		assert.Equal(t, RoleErrorTypeDuplicateAssignment, err.ErrorType)
		assert.Equal(t, userID, err.UserID)
		assert.Equal(t, roleID, err.RoleID)
		assert.Equal(t, 409, err.ErrorResponse.Code)
		assert.Contains(t, err.ErrorResponse.Message, userID)
		assert.Contains(t, err.ErrorResponse.Message, roleID)
	})

	t.Run("NewRoleAssignmentNotFoundError", func(t *testing.T) {
		userID := "user-789"
		roleID := "role-101"
		err := NewRoleAssignmentNotFoundError(userID, roleID)

		assert.Equal(t, RoleErrorTypeAssignmentNotFound, err.ErrorType)
		assert.Equal(t, userID, err.UserID)
		assert.Equal(t, roleID, err.RoleID)
		assert.Equal(t, 404, err.ErrorResponse.Code)
	})

	t.Run("NewInactiveRoleError", func(t *testing.T) {
		roleID := "role-inactive"
		err := NewInactiveRoleError(roleID)

		assert.Equal(t, RoleErrorTypeInactiveRole, err.ErrorType)
		assert.Equal(t, roleID, err.RoleID)
		assert.Equal(t, 400, err.ErrorResponse.Code)
		assert.Contains(t, err.ErrorResponse.Message, "inactive")
	})

	t.Run("NewRoleValidationError", func(t *testing.T) {
		field := "role_id"
		message := "Invalid format"
		err := NewRoleValidationError(field, message)

		assert.Equal(t, RoleErrorTypeValidationFailed, err.ErrorType)
		assert.Equal(t, 400, err.ErrorResponse.Code)
		assert.Contains(t, err.ErrorResponse.Message, field)
		assert.Contains(t, err.ErrorResponse.Message, message)
		assert.NotNil(t, err.Details)
		assert.Equal(t, field, err.Details["field"])
		assert.Equal(t, message, err.Details["validation_message"])
	})

	t.Run("NewRolePermissionDeniedError", func(t *testing.T) {
		operation := "assign_role"
		err := NewRolePermissionDeniedError(operation)

		assert.Equal(t, RoleErrorTypePermissionDenied, err.ErrorType)
		assert.Equal(t, 403, err.ErrorResponse.Code)
		assert.Contains(t, err.ErrorResponse.Message, operation)
	})
}

func TestRoleErrorResponse_WithMethods(t *testing.T) {
	err := NewRoleErrorResponse(RoleErrorTypeNotFound, "Test error", 404)

	// Test WithUserID
	err = err.WithUserID("user-123")
	assert.Equal(t, "user-123", err.UserID)

	// Test WithRoleID
	err = err.WithRoleID("role-456")
	assert.Equal(t, "role-456", err.RoleID)

	// Test WithDetails
	details := map[string]interface{}{
		"additional_info": "test",
		"code":            "ERR001",
	}
	err = err.WithDetails(details)
	assert.Equal(t, "test", err.Details["additional_info"])
	assert.Equal(t, "ERR001", err.Details["code"])

	// Test WithRequestID
	err = err.WithRequestID("req-789")
	assert.Equal(t, "req-789", err.ErrorResponse.RequestID)
}

func TestAssignRoleResponse_InterfaceMethods(t *testing.T) {
	role := &models.Role{
		BaseModel:   base.NewBaseModel("ROLE", hash.Medium),
		Name:        "Test Role",
		Description: "Test Description",
		IsActive:    true,
	}
	role.SetID("role-123")

	response := NewAssignRoleResponse("user-456", role, "Success")

	assert.Equal(t, "AssignRoleResponse", response.GetType())
	assert.True(t, response.IsSuccess())
	assert.Equal(t, "http", response.GetProtocol())
	assert.Equal(t, "post", response.GetOperation())
	assert.Equal(t, "v2", response.GetVersion())
	assert.Equal(t, "", response.GetResponseID())
	assert.Nil(t, response.GetHeaders())
	assert.Equal(t, response, response.GetBody())
	assert.Nil(t, response.GetContext())
	assert.Nil(t, response.ToProto())
	assert.Contains(t, response.String(), "user-456")
	assert.Contains(t, response.String(), "Test Role")
}
