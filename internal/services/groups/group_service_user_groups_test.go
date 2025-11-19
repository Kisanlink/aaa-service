package groups

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestService_GetUserGroupsInOrganization_Validation tests input validation
func TestService_GetUserGroupsInOrganization_Validation(t *testing.T) {
	logger := zap.NewNop()

	// Create a service with minimal dependencies for validation testing
	service := &Service{
		logger: logger,
	}

	tests := []struct {
		name          string
		orgID         string
		userID        string
		limit         int
		offset        int
		expectedError string
	}{
		{
			name:          "empty org ID",
			orgID:         "",
			userID:        "user-123",
			limit:         10,
			offset:        0,
			expectedError: "org_id cannot be empty",
		},
		{
			name:          "empty user ID",
			orgID:         "org-456",
			userID:        "",
			limit:         10,
			offset:        0,
			expectedError: "user_id cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := service.GetUserGroupsInOrganization(context.Background(), tt.orgID, tt.userID, tt.limit, tt.offset)

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			assert.Nil(t, result)
		})
	}
}
