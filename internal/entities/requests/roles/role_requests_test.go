//nolint:typecheck
package roles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssignRoleRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *AssignRoleRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &AssignRoleRequest{
				RoleID: "role-123",
			},
			wantErr: false,
		},
		{
			name: "empty role ID",
			request: &AssignRoleRequest{
				RoleID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssignRoleRequest_ValidateWithUserID(t *testing.T) {
	tests := []struct {
		name    string
		request *AssignRoleRequest
		userID  string
		wantErr bool
	}{
		{
			name: "valid request with user ID",
			request: &AssignRoleRequest{
				RoleID: "role-123",
			},
			userID:  "user-456",
			wantErr: false,
		},
		{
			name: "empty user ID",
			request: &AssignRoleRequest{
				RoleID: "role-123",
			},
			userID:  "",
			wantErr: true,
		},
		{
			name: "empty role ID",
			request: &AssignRoleRequest{
				RoleID: "",
			},
			userID:  "user-456",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.ValidateWithUserID(tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRemoveRoleRequest_ValidateWithIDs(t *testing.T) {
	tests := []struct {
		name    string
		request *RemoveRoleRequest
		userID  string
		roleID  string
		wantErr bool
	}{
		{
			name:    "valid request with IDs",
			request: &RemoveRoleRequest{},
			userID:  "user-456",
			roleID:  "role-123",
			wantErr: false,
		},
		{
			name:    "empty user ID",
			request: &RemoveRoleRequest{},
			userID:  "",
			roleID:  "role-123",
			wantErr: true,
		},
		{
			name:    "empty role ID",
			request: &RemoveRoleRequest{},
			userID:  "user-456",
			roleID:  "",
			wantErr: true,
		},
		{
			name:    "both empty",
			request: &RemoveRoleRequest{},
			userID:  "",
			roleID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.ValidateWithIDs(tt.userID, tt.roleID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssignRoleRequest_GetRoleID(t *testing.T) {
	request := &AssignRoleRequest{
		RoleID: "test-role-123",
	}

	assert.Equal(t, "test-role-123", request.GetRoleID())
}

func TestNewAssignRoleRequest(t *testing.T) {
	roleID := "role-123"
	protocol := "http"
	operation := "post"
	version := "v2"
	requestID := "req-456"
	headers := map[string][]string{"Content-Type": {"application/json"}}
	body := map[string]interface{}{"test": "data"}
	context := map[string]interface{}{"user": "admin"}

	request := NewAssignRoleRequest(roleID, protocol, operation, version, requestID, headers, body, context)

	assert.NotNil(t, request)
	assert.Equal(t, roleID, request.RoleID)
	assert.Equal(t, protocol, request.GetProtocol())
	assert.Equal(t, operation, request.GetOperation())
	assert.Equal(t, version, request.GetVersion())
	assert.Equal(t, requestID, request.GetRequestID())
	assert.Equal(t, "AssignRole", request.GetType())
}

func TestNewRemoveRoleRequest(t *testing.T) {
	protocol := "http"
	operation := "delete"
	version := "v2"
	requestID := "req-789"
	headers := map[string][]string{"Content-Type": {"application/json"}}
	body := map[string]interface{}{"test": "data"}
	context := map[string]interface{}{"user": "admin"}

	request := NewRemoveRoleRequest(protocol, operation, version, requestID, headers, body, context)

	assert.NotNil(t, request)
	assert.Equal(t, protocol, request.GetProtocol())
	assert.Equal(t, operation, request.GetOperation())
	assert.Equal(t, version, request.GetVersion())
	assert.Equal(t, requestID, request.GetRequestID())
	assert.Equal(t, "RemoveRole", request.GetType())
}
