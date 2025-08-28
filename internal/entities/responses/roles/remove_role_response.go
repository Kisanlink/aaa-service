package roles

import (
	"fmt"
)

// RemoveRoleResponse represents the response for role removal operations
type RemoveRoleResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	RoleID  string `json:"role_id"`
}

// NewRemoveRoleResponse creates a new RemoveRoleResponse
func NewRemoveRoleResponse(userID, roleID, message string) *RemoveRoleResponse {
	return &RemoveRoleResponse{
		Message: message,
		UserID:  userID,
		RoleID:  roleID,
	}
}

// GetType returns the response type
func (r *RemoveRoleResponse) GetType() string {
	return "RemoveRoleResponse"
}

// IsSuccess returns whether the response indicates success
func (r *RemoveRoleResponse) IsSuccess() bool {
	return true
}

// GetProtocol returns the transport protocol
func (r *RemoveRoleResponse) GetProtocol() string {
	return "http"
}

// GetOperation returns the operation
func (r *RemoveRoleResponse) GetOperation() string {
	return "delete"
}

// GetVersion returns the API version
func (r *RemoveRoleResponse) GetVersion() string {
	return "v2"
}

// GetResponseID returns the response ID
func (r *RemoveRoleResponse) GetResponseID() string {
	return ""
}

// GetHeaders returns response headers
func (r *RemoveRoleResponse) GetHeaders() map[string][]string {
	return nil
}

// GetBody returns the response body
func (r *RemoveRoleResponse) GetBody() interface{} {
	return r
}

// GetContext returns response context
func (r *RemoveRoleResponse) GetContext() map[string]interface{} {
	return nil
}

// ToProto converts to protocol buffer format
func (r *RemoveRoleResponse) ToProto() interface{} {
	return nil
}

// String returns a string representation
func (r *RemoveRoleResponse) String() string {
	return fmt.Sprintf("RemoveRoleResponse{UserID: %s, RoleID: %s, Message: %s}",
		r.UserID, r.RoleID, r.Message)
}
