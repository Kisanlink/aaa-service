package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOrganizationAuditActionConstants verifies that all organization-related audit action constants are defined
func TestOrganizationAuditActionConstants(t *testing.T) {
	// Test organization action constants
	assert.Equal(t, "create_organization", AuditActionCreateOrganization)
	assert.Equal(t, "update_organization", AuditActionUpdateOrganization)
	assert.Equal(t, "delete_organization", AuditActionDeleteOrganization)
	assert.Equal(t, "activate_organization", AuditActionActivateOrganization)
	assert.Equal(t, "deactivate_organization", AuditActionDeactivateOrganization)

	// Test group action constants
	assert.Equal(t, "create_group", AuditActionCreateGroup)
	assert.Equal(t, "update_group", AuditActionUpdateGroup)
	assert.Equal(t, "delete_group", AuditActionDeleteGroup)
	assert.Equal(t, "add_group_member", AuditActionAddGroupMember)
	assert.Equal(t, "remove_group_member", AuditActionRemoveGroupMember)
	assert.Equal(t, "assign_group_role", AuditActionAssignGroupRole)
	assert.Equal(t, "remove_group_role", AuditActionRemoveGroupRole)

	// Test hierarchy action constants
	assert.Equal(t, "change_organization_hierarchy", AuditActionChangeOrganizationHierarchy)
	assert.Equal(t, "change_group_hierarchy", AuditActionChangeGroupHierarchy)
}

// TestOrganizationResourceTypeConstants verifies that organization-related resource type constants are defined
func TestOrganizationResourceTypeConstants(t *testing.T) {
	assert.Equal(t, "aaa/organization", ResourceTypeOrganization)
	assert.Equal(t, "aaa/group", ResourceTypeGroup)
	assert.Equal(t, "aaa/group_role", ResourceTypeGroupRole)
	assert.Equal(t, "aaa/audit_log", ResourceTypeAuditLog)
}

// TestAuditLogCreationForOrganizations verifies audit log creation for organization operations
func TestAuditLogCreationForOrganizations(t *testing.T) {
	// Test organization creation audit log
	auditLog := NewAuditLog(
		AuditActionCreateOrganization,
		ResourceTypeOrganization,
		AuditStatusSuccess,
		"Organization created successfully",
	)

	assert.NotEmpty(t, auditLog.BaseModel.ID)
	assert.Equal(t, AuditActionCreateOrganization, auditLog.Action)
	assert.Equal(t, ResourceTypeOrganization, auditLog.ResourceType)
	assert.Equal(t, AuditStatusSuccess, auditLog.Status)
	assert.Equal(t, "Organization created successfully", auditLog.Message)
	assert.NotNil(t, auditLog.Details)
	assert.False(t, auditLog.Timestamp.IsZero())

	// Test adding organization-specific details
	auditLog.AddDetail("organization_name", "Test Organization")
	auditLog.AddDetail("organization_id", "org123")
	auditLog.AddDetail("is_active", true)

	orgName, exists := auditLog.GetDetail("organization_name")
	assert.True(t, exists)
	assert.Equal(t, "Test Organization", orgName)

	orgID, exists := auditLog.GetDetail("organization_id")
	assert.True(t, exists)
	assert.Equal(t, "org123", orgID)

	isActive, exists := auditLog.GetDetail("is_active")
	assert.True(t, exists)
	assert.Equal(t, true, isActive)
}

// TestAuditLogCreationForGroups verifies audit log creation for group operations
func TestAuditLogCreationForGroups(t *testing.T) {
	// Test group creation audit log
	auditLog := NewAuditLogWithUserAndResource(
		"user123",
		AuditActionCreateGroup,
		ResourceTypeGroup,
		"group123",
		AuditStatusSuccess,
		"Group created successfully",
	)

	assert.NotNil(t, auditLog.UserID)
	assert.Equal(t, "user123", *auditLog.UserID)
	assert.NotNil(t, auditLog.ResourceID)
	assert.Equal(t, "group123", *auditLog.ResourceID)
	assert.Equal(t, AuditActionCreateGroup, auditLog.Action)
	assert.Equal(t, ResourceTypeGroup, auditLog.ResourceType)

	// Test adding group-specific details
	auditLog.AddDetail("group_name", "Test Group")
	auditLog.AddDetail("organization_id", "org123")
	auditLog.AddDetail("operation_type", "group")

	groupName, exists := auditLog.GetDetail("group_name")
	assert.True(t, exists)
	assert.Equal(t, "Test Group", groupName)

	orgID, exists := auditLog.GetDetail("organization_id")
	assert.True(t, exists)
	assert.Equal(t, "org123", orgID)

	opType, exists := auditLog.GetDetail("operation_type")
	assert.True(t, exists)
	assert.Equal(t, "group", opType)
}

// TestAuditLogCreationForGroupMembership verifies audit log creation for group membership operations
func TestAuditLogCreationForGroupMembership(t *testing.T) {
	// Test group membership change audit log
	auditLog := NewAuditLogWithUserAndResource(
		"admin123",
		AuditActionAddGroupMember,
		ResourceTypeGroup,
		"group123",
		AuditStatusSuccess,
		"Member added to group successfully",
	)

	assert.NotNil(t, auditLog.UserID)
	assert.Equal(t, "admin123", *auditLog.UserID)
	assert.Equal(t, AuditActionAddGroupMember, auditLog.Action)
	assert.Equal(t, ResourceTypeGroup, auditLog.ResourceType)

	// Test adding membership-specific details
	auditLog.AddDetail("organization_id", "org123")
	auditLog.AddDetail("group_id", "group123")
	auditLog.AddDetail("target_user_id", "user123")
	auditLog.AddDetail("actor_user_id", "admin123")
	auditLog.AddDetail("operation_type", "group_membership")

	targetUserID, exists := auditLog.GetDetail("target_user_id")
	assert.True(t, exists)
	assert.Equal(t, "user123", targetUserID)

	actorUserID, exists := auditLog.GetDetail("actor_user_id")
	assert.True(t, exists)
	assert.Equal(t, "admin123", actorUserID)

	opType, exists := auditLog.GetDetail("operation_type")
	assert.True(t, exists)
	assert.Equal(t, "group_membership", opType)
}

// TestAuditLogCreationForGroupRoles verifies audit log creation for group role operations
func TestAuditLogCreationForGroupRoles(t *testing.T) {
	// Test group role assignment audit log
	auditLog := NewAuditLogWithUserAndResource(
		"admin123",
		AuditActionAssignGroupRole,
		ResourceTypeGroupRole,
		"group123",
		AuditStatusSuccess,
		"Role assigned to group successfully",
	)

	assert.NotNil(t, auditLog.UserID)
	assert.Equal(t, "admin123", *auditLog.UserID)
	assert.Equal(t, AuditActionAssignGroupRole, auditLog.Action)
	assert.Equal(t, ResourceTypeGroupRole, auditLog.ResourceType)

	// Test adding role assignment specific details
	auditLog.AddDetail("organization_id", "org123")
	auditLog.AddDetail("group_id", "group123")
	auditLog.AddDetail("role_id", "role123")
	auditLog.AddDetail("actor_user_id", "admin123")
	auditLog.AddDetail("operation_type", "group_role")

	roleID, exists := auditLog.GetDetail("role_id")
	assert.True(t, exists)
	assert.Equal(t, "role123", roleID)

	opType, exists := auditLog.GetDetail("operation_type")
	assert.True(t, exists)
	assert.Equal(t, "group_role", opType)
}

// TestAuditLogCreationForHierarchyChanges verifies audit log creation for hierarchy change operations
func TestAuditLogCreationForHierarchyChanges(t *testing.T) {
	// Test organization hierarchy change audit log
	auditLog := NewAuditLogWithUserAndResource(
		"admin123",
		AuditActionChangeOrganizationHierarchy,
		ResourceTypeOrganization,
		"org123",
		AuditStatusSuccess,
		"Organization hierarchy changed",
	)

	assert.NotNil(t, auditLog.UserID)
	assert.Equal(t, "admin123", *auditLog.UserID)
	assert.Equal(t, AuditActionChangeOrganizationHierarchy, auditLog.Action)
	assert.Equal(t, ResourceTypeOrganization, auditLog.ResourceType)

	// Test adding hierarchy change specific details
	auditLog.AddDetail("old_parent_id", "parent1")
	auditLog.AddDetail("new_parent_id", "parent2")
	auditLog.AddDetail("operation_type", "hierarchy_change")

	oldParentID, exists := auditLog.GetDetail("old_parent_id")
	assert.True(t, exists)
	assert.Equal(t, "parent1", oldParentID)

	newParentID, exists := auditLog.GetDetail("new_parent_id")
	assert.True(t, exists)
	assert.Equal(t, "parent2", newParentID)

	opType, exists := auditLog.GetDetail("operation_type")
	assert.True(t, exists)
	assert.Equal(t, "hierarchy_change", opType)
}

// TestAuditLogFailureScenarios verifies audit log creation for failure scenarios
func TestAuditLogFailureScenarios(t *testing.T) {
	// Test failed organization deletion
	auditLog := NewAuditLogWithUserAndResource(
		"admin123",
		AuditActionDeleteOrganization,
		ResourceTypeOrganization,
		"org123",
		AuditStatusFailure,
		"Failed to delete organization",
	)

	assert.Equal(t, AuditStatusFailure, auditLog.Status)
	assert.True(t, auditLog.IsFailure())
	assert.False(t, auditLog.IsSuccess())

	// Add error details
	auditLog.SetErrorDetails("Organization has active children", "ORG_HAS_CHILDREN")
	auditLog.AddDetail("organization_name", "Test Organization")
	auditLog.AddDetail("had_children", true)
	auditLog.AddDetail("had_groups", false)

	errorMsg, exists := auditLog.GetDetail("error_message")
	assert.True(t, exists)
	assert.Equal(t, "Organization has active children", errorMsg)

	errorCode, exists := auditLog.GetDetail("error_code")
	assert.True(t, exists)
	assert.Equal(t, "ORG_HAS_CHILDREN", errorCode)

	hadChildren, exists := auditLog.GetDetail("had_children")
	assert.True(t, exists)
	assert.Equal(t, true, hadChildren)
}

// TestAuditLogAnonymousUser verifies audit log creation for anonymous users
func TestAuditLogAnonymousUser(t *testing.T) {
	// Test audit log creation without user ID (for anonymous operations)
	auditLog := NewAuditLog(
		AuditActionCreateOrganization,
		ResourceTypeOrganization,
		AuditStatusSuccess,
		"Organization created successfully",
	)

	// Verify UserID is nil for anonymous operations
	assert.Nil(t, auditLog.UserID)
	assert.Equal(t, AuditActionCreateOrganization, auditLog.Action)
	assert.Equal(t, ResourceTypeOrganization, auditLog.ResourceType)

	// Test setting resource ID separately
	orgID := "org123"
	auditLog.ResourceID = &orgID
	assert.NotNil(t, auditLog.ResourceID)
	assert.Equal(t, "org123", *auditLog.ResourceID)
}
