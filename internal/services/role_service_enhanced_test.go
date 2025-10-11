package services

import (
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/stretchr/testify/assert"
)

// TestRoleServiceEnhancedMethodsExist tests that the enhanced methods exist with correct signatures
func TestRoleServiceEnhancedMethodsExist(t *testing.T) {
	t.Run("Enhanced methods exist", func(t *testing.T) {
		// This test verifies that the methods exist by checking if they compile
		// We don't actually call them to avoid nil pointer issues

		var service *RoleService

		// These should compile if the methods exist with correct signatures
		_ = func() error {
			return service.ValidateRoleAssignment(nil, "", "")
		}

		_ = func() error {
			return service.AssignRole(nil, "", "")
		}

		_ = func() error {
			return service.RemoveRole(nil, "", "")
		}

		_ = func() ([]*models.UserRole, error) {
			return service.GetUserRoles(nil, "")
		}

		// If we get here, all methods exist with correct signatures
		assert.True(t, true, "All enhanced methods exist with correct signatures")
	})
}

// TestRoleServiceLegacyMethodsExist tests that legacy methods still exist
func TestRoleServiceLegacyMethodsExist(t *testing.T) {
	t.Run("Legacy methods exist", func(t *testing.T) {
		var service *RoleService

		// These should compile if the legacy methods still exist
		_ = func() error {
			return service.AssignRoleToUser(nil, "", "")
		}

		_ = func() error {
			return service.RemoveRoleFromUser(nil, "", "")
		}

		assert.True(t, true, "All legacy methods exist with correct signatures")
	})
}
