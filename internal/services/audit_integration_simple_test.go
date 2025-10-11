package services

import (
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/stretchr/testify/assert"
)

// TestAuditLoggingBasicFunctionality tests basic audit logging functionality
func TestAuditLoggingBasicFunctionality(t *testing.T) {
	t.Run("AuditLogCreation", func(t *testing.T) {
		// Test creating audit logs with different scenarios

		// Test organization creation audit log
		orgAuditLog := models.NewAuditLogWithUserAndResource(
			"user123",
			models.AuditActionCreateOrganization,
			models.ResourceTypeOrganization,
			"org123",
			models.AuditStatusSuccess,
			"Organization created successfully",
		)

		assert.NotNil(t, orgAuditLog)
		assert.Equal(t, "user123", *orgAuditLog.UserID)
		assert.Equal(t, models.AuditActionCreateOrganization, orgAuditLog.Action)
		assert.Equal(t, models.ResourceTypeOrganization, orgAuditLog.ResourceType)
		assert.Equal(t, "org123", *orgAuditLog.ResourceID)
		assert.Equal(t, models.AuditStatusSuccess, orgAuditLog.Status)
		assert.False(t, orgAuditLog.Timestamp.IsZero())

		// Add organization-specific details
		orgAuditLog.AddDetail("organization_name", "Test Organization")
		orgAuditLog.AddDetail("organization_id", "org123")
		orgAuditLog.AddDetail("operation_type", "organization")

		orgName, exists := orgAuditLog.GetDetail("organization_name")
		assert.True(t, exists)
		assert.Equal(t, "Test Organization", orgName)

		orgID, exists := orgAuditLog.GetDetail("organization_id")
		assert.True(t, exists)
		assert.Equal(t, "org123", orgID)
	})

	t.Run("GroupMembershipAuditLog", func(t *testing.T) {
		// Test group membership change audit log
		membershipAuditLog := models.NewAuditLogWithUserAndResource(
			"admin123",
			models.AuditActionAddGroupMember,
			models.ResourceTypeGroup,
			"group123",
			models.AuditStatusSuccess,
			"Member added to group successfully",
		)

		assert.NotNil(t, membershipAuditLog)
		assert.Equal(t, "admin123", *membershipAuditLog.UserID)
		assert.Equal(t, models.AuditActionAddGroupMember, membershipAuditLog.Action)
		assert.Equal(t, models.ResourceTypeGroup, membershipAuditLog.ResourceType)

		// Add membership-specific details
		membershipAuditLog.AddDetail("organization_id", "org123")
		membershipAuditLog.AddDetail("group_id", "group123")
		membershipAuditLog.AddDetail("target_user_id", "user123")
		membershipAuditLog.AddDetail("actor_user_id", "admin123")
		membershipAuditLog.AddDetail("operation_type", "group_membership")

		targetUserID, exists := membershipAuditLog.GetDetail("target_user_id")
		assert.True(t, exists)
		assert.Equal(t, "user123", targetUserID)

		opType, exists := membershipAuditLog.GetDetail("operation_type")
		assert.True(t, exists)
		assert.Equal(t, "group_membership", opType)
	})

	t.Run("HierarchyChangeAuditLog", func(t *testing.T) {
		// Test hierarchy change audit log
		hierarchyAuditLog := models.NewAuditLogWithUserAndResource(
			"admin123",
			models.AuditActionChangeOrganizationHierarchy,
			models.ResourceTypeOrganization,
			"org123",
			models.AuditStatusSuccess,
			"Organization hierarchy changed",
		)

		assert.NotNil(t, hierarchyAuditLog)
		assert.Equal(t, models.AuditActionChangeOrganizationHierarchy, hierarchyAuditLog.Action)

		// Add hierarchy-specific details
		hierarchyAuditLog.AddDetail("old_parent_id", "parent1")
		hierarchyAuditLog.AddDetail("new_parent_id", "parent2")
		hierarchyAuditLog.AddDetail("operation_type", "hierarchy_change")

		oldParentID, exists := hierarchyAuditLog.GetDetail("old_parent_id")
		assert.True(t, exists)
		assert.Equal(t, "parent1", oldParentID)

		newParentID, exists := hierarchyAuditLog.GetDetail("new_parent_id")
		assert.True(t, exists)
		assert.Equal(t, "parent2", newParentID)
	})

	t.Run("FailureScenarioAuditLog", func(t *testing.T) {
		// Test failed operation audit log
		failureAuditLog := models.NewAuditLogWithUserAndResource(
			"admin123",
			models.AuditActionDeleteOrganization,
			models.ResourceTypeOrganization,
			"org123",
			models.AuditStatusFailure,
			"Failed to delete organization",
		)

		assert.NotNil(t, failureAuditLog)
		assert.Equal(t, models.AuditStatusFailure, failureAuditLog.Status)
		assert.True(t, failureAuditLog.IsFailure())
		assert.False(t, failureAuditLog.IsSuccess())

		// Add error details
		failureAuditLog.SetErrorDetails("Organization has active children", "ORG_HAS_CHILDREN")
		failureAuditLog.AddDetail("organization_name", "Test Organization")
		failureAuditLog.AddDetail("had_children", true)

		errorMsg, exists := failureAuditLog.GetDetail("error_message")
		assert.True(t, exists)
		assert.Equal(t, "Organization has active children", errorMsg)

		errorCode, exists := failureAuditLog.GetDetail("error_code")
		assert.True(t, exists)
		assert.Equal(t, "ORG_HAS_CHILDREN", errorCode)
	})

	t.Run("AnonymousUserAuditLog", func(t *testing.T) {
		// Test audit log for anonymous operations
		anonymousAuditLog := models.NewAuditLog(
			models.AuditActionCreateOrganization,
			models.ResourceTypeOrganization,
			models.AuditStatusSuccess,
			"Organization created successfully",
		)

		assert.NotNil(t, anonymousAuditLog)
		assert.Nil(t, anonymousAuditLog.UserID) // No user ID for anonymous operations
		assert.Equal(t, models.AuditActionCreateOrganization, anonymousAuditLog.Action)

		// Set resource ID separately
		orgID := "org123"
		anonymousAuditLog.ResourceID = &orgID
		assert.NotNil(t, anonymousAuditLog.ResourceID)
		assert.Equal(t, "org123", *anonymousAuditLog.ResourceID)
	})
}

// TestAuditLogConstants tests that all required constants are defined
func TestAuditLogConstants(t *testing.T) {
	t.Run("StatusConstants", func(t *testing.T) {
		assert.Equal(t, "success", models.AuditStatusSuccess)
		assert.Equal(t, "failure", models.AuditStatusFailure)
		assert.Equal(t, "warning", models.AuditStatusWarning)
	})

	t.Run("ActionConstants", func(t *testing.T) {
		// Organization actions
		assert.Equal(t, "create_organization", models.AuditActionCreateOrganization)
		assert.Equal(t, "update_organization", models.AuditActionUpdateOrganization)
		assert.Equal(t, "delete_organization", models.AuditActionDeleteOrganization)
		assert.Equal(t, "activate_organization", models.AuditActionActivateOrganization)
		assert.Equal(t, "deactivate_organization", models.AuditActionDeactivateOrganization)

		// Group actions
		assert.Equal(t, "create_group", models.AuditActionCreateGroup)
		assert.Equal(t, "update_group", models.AuditActionUpdateGroup)
		assert.Equal(t, "delete_group", models.AuditActionDeleteGroup)
		assert.Equal(t, "add_group_member", models.AuditActionAddGroupMember)
		assert.Equal(t, "remove_group_member", models.AuditActionRemoveGroupMember)
		assert.Equal(t, "assign_group_role", models.AuditActionAssignGroupRole)
		assert.Equal(t, "remove_group_role", models.AuditActionRemoveGroupRole)

		// Hierarchy actions
		assert.Equal(t, "change_organization_hierarchy", models.AuditActionChangeOrganizationHierarchy)
		assert.Equal(t, "change_group_hierarchy", models.AuditActionChangeGroupHierarchy)
	})

	t.Run("ResourceTypeConstants", func(t *testing.T) {
		assert.Equal(t, "aaa/organization", models.ResourceTypeOrganization)
		assert.Equal(t, "aaa/group", models.ResourceTypeGroup)
		assert.Equal(t, "aaa/group_role", models.ResourceTypeGroupRole)
		assert.Equal(t, "aaa/audit_log", models.ResourceTypeAuditLog)
		assert.Equal(t, "aaa/user", models.ResourceTypeUser)
		assert.Equal(t, "aaa/role", models.ResourceTypeRole)
	})
}

// TestAuditLogMethods tests audit log helper methods
func TestAuditLogMethods(t *testing.T) {
	t.Run("StatusCheckMethods", func(t *testing.T) {
		// Test success status
		successLog := models.NewAuditLog(
			models.AuditActionCreateOrganization,
			models.ResourceTypeOrganization,
			models.AuditStatusSuccess,
			"Success message",
		)
		assert.True(t, successLog.IsSuccess())
		assert.False(t, successLog.IsFailure())
		assert.False(t, successLog.IsWarning())

		// Test failure status
		failureLog := models.NewAuditLog(
			models.AuditActionCreateOrganization,
			models.ResourceTypeOrganization,
			models.AuditStatusFailure,
			"Failure message",
		)
		assert.False(t, failureLog.IsSuccess())
		assert.True(t, failureLog.IsFailure())
		assert.False(t, failureLog.IsWarning())

		// Test warning status
		warningLog := models.NewAuditLog(
			models.AuditActionCreateOrganization,
			models.ResourceTypeOrganization,
			models.AuditStatusWarning,
			"Warning message",
		)
		assert.False(t, warningLog.IsSuccess())
		assert.False(t, warningLog.IsFailure())
		assert.True(t, warningLog.IsWarning())
	})

	t.Run("DetailMethods", func(t *testing.T) {
		auditLog := models.NewAuditLog(
			models.AuditActionCreateOrganization,
			models.ResourceTypeOrganization,
			models.AuditStatusSuccess,
			"Test message",
		)

		// Test adding and getting details
		auditLog.AddDetail("test_key", "test_value")
		auditLog.AddDetail("numeric_key", 123)
		auditLog.AddDetail("boolean_key", true)

		value, exists := auditLog.GetDetail("test_key")
		assert.True(t, exists)
		assert.Equal(t, "test_value", value)

		numValue, exists := auditLog.GetDetail("numeric_key")
		assert.True(t, exists)
		assert.Equal(t, 123, numValue)

		boolValue, exists := auditLog.GetDetail("boolean_key")
		assert.True(t, exists)
		assert.Equal(t, true, boolValue)

		// Test non-existent key
		_, exists = auditLog.GetDetail("non_existent_key")
		assert.False(t, exists)
	})

	t.Run("RequestDetailsMethods", func(t *testing.T) {
		auditLog := models.NewAuditLog(
			models.AuditActionCreateOrganization,
			models.ResourceTypeOrganization,
			models.AuditStatusSuccess,
			"Test message",
		)

		// Test setting request details
		auditLog.SetRequestDetails("POST", "/api/v1/organizations", "192.168.1.1", "Mozilla/5.0")

		assert.Equal(t, "192.168.1.1", auditLog.IPAddress)
		assert.Equal(t, "Mozilla/5.0", auditLog.UserAgent)

		method, exists := auditLog.GetDetail("http_method")
		assert.True(t, exists)
		assert.Equal(t, "POST", method)

		path, exists := auditLog.GetDetail("http_path")
		assert.True(t, exists)
		assert.Equal(t, "/api/v1/organizations", path)
	})

	t.Run("ErrorDetailsMethods", func(t *testing.T) {
		auditLog := models.NewAuditLog(
			models.AuditActionCreateOrganization,
			models.ResourceTypeOrganization,
			models.AuditStatusFailure,
			"Test failure",
		)

		// Test setting error details
		auditLog.SetErrorDetails("Organization already exists", "ORG_EXISTS")

		errorMsg, exists := auditLog.GetDetail("error_message")
		assert.True(t, exists)
		assert.Equal(t, "Organization already exists", errorMsg)

		errorCode, exists := auditLog.GetDetail("error_code")
		assert.True(t, exists)
		assert.Equal(t, "ORG_EXISTS", errorCode)
	})
}

// TestAuditLogTimestamps tests timestamp handling
func TestAuditLogTimestamps(t *testing.T) {
	t.Run("AutomaticTimestamp", func(t *testing.T) {
		before := time.Now()
		auditLog := models.NewAuditLog(
			models.AuditActionCreateOrganization,
			models.ResourceTypeOrganization,
			models.AuditStatusSuccess,
			"Test message",
		)
		after := time.Now()

		assert.False(t, auditLog.Timestamp.IsZero())
		assert.True(t, auditLog.Timestamp.After(before) || auditLog.Timestamp.Equal(before))
		assert.True(t, auditLog.Timestamp.Before(after) || auditLog.Timestamp.Equal(after))
	})

	t.Run("BeforeCreateHook", func(t *testing.T) {
		auditLog := models.NewAuditLog(
			models.AuditActionCreateOrganization,
			models.ResourceTypeOrganization,
			models.AuditStatusSuccess,
			"Test message",
		)

		// Clear timestamp to test BeforeCreate hook
		auditLog.Timestamp = time.Time{}
		assert.True(t, auditLog.Timestamp.IsZero())

		// Call BeforeCreate hook
		err := auditLog.BeforeCreate()
		assert.NoError(t, err)
		assert.False(t, auditLog.Timestamp.IsZero())
	})
}

// TestOrganizationScopedAuditLogs tests organization-scoped audit logging
func TestOrganizationScopedAuditLogs(t *testing.T) {
	t.Run("OrganizationContextValidation", func(t *testing.T) {
		// Create audit log with organization context
		auditLog := models.NewAuditLogWithUserAndResource(
			"admin123",
			models.AuditActionCreateGroup,
			models.ResourceTypeGroup,
			"group123",
			models.AuditStatusSuccess,
			"Group created in organization",
		)

		// Add organization context
		auditLog.AddDetail("organization_id", "org123")
		auditLog.AddDetail("operation_type", "group")

		// Verify organization context
		orgID, exists := auditLog.GetDetail("organization_id")
		assert.True(t, exists)
		assert.Equal(t, "org123", orgID)

		opType, exists := auditLog.GetDetail("operation_type")
		assert.True(t, exists)
		assert.Equal(t, "group", opType)
	})

	t.Run("TamperProofMarking", func(t *testing.T) {
		// Create audit log with tamper-proof marking
		auditLog := models.NewAuditLogWithUserAndResource(
			"admin123",
			models.AuditActionChangeOrganizationHierarchy,
			models.ResourceTypeOrganization,
			"org123",
			models.AuditStatusSuccess,
			"Critical organization structure change",
		)

		// Add tamper-proof and security-sensitive markings
		auditLog.AddDetail("security_sensitive", true)
		auditLog.AddDetail("tamper_proof", true)
		auditLog.AddDetail("change_timestamp", time.Now().UTC().Format(time.RFC3339))

		securitySensitive, exists := auditLog.GetDetail("security_sensitive")
		assert.True(t, exists)
		assert.Equal(t, true, securitySensitive)

		tamperProof, exists := auditLog.GetDetail("tamper_proof")
		assert.True(t, exists)
		assert.Equal(t, true, tamperProof)

		changeTimestamp, exists := auditLog.GetDetail("change_timestamp")
		assert.True(t, exists)
		assert.NotEmpty(t, changeTimestamp)
	})
}
