package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewGroupRole(t *testing.T) {
	groupID := "test-group-id"
	roleID := "test-role-id"
	organizationID := "test-org-id"
	assignedBy := "test-user-id"

	groupRole := NewGroupRole(groupID, roleID, organizationID, assignedBy)

	assert.NotNil(t, groupRole)
	assert.NotNil(t, groupRole.BaseModel)
	assert.Equal(t, groupID, groupRole.GroupID)
	assert.Equal(t, roleID, groupRole.RoleID)
	assert.Equal(t, organizationID, groupRole.OrganizationID)
	assert.Equal(t, assignedBy, groupRole.AssignedBy)
	assert.True(t, groupRole.IsActive)
	assert.Nil(t, groupRole.StartsAt)
	assert.Nil(t, groupRole.EndsAt)
	assert.Nil(t, groupRole.Metadata)
}

func TestNewGroupRoleWithTimebound(t *testing.T) {
	groupID := "test-group-id"
	roleID := "test-role-id"
	organizationID := "test-org-id"
	assignedBy := "test-user-id"
	startsAt := time.Now()
	endsAt := startsAt.Add(24 * time.Hour)

	groupRole := NewGroupRoleWithTimebound(groupID, roleID, organizationID, assignedBy, &startsAt, &endsAt)

	assert.NotNil(t, groupRole)
	assert.Equal(t, groupID, groupRole.GroupID)
	assert.Equal(t, roleID, groupRole.RoleID)
	assert.Equal(t, organizationID, groupRole.OrganizationID)
	assert.Equal(t, assignedBy, groupRole.AssignedBy)
	assert.True(t, groupRole.IsActive)
	assert.NotNil(t, groupRole.StartsAt)
	assert.NotNil(t, groupRole.EndsAt)
	assert.Equal(t, startsAt, *groupRole.StartsAt)
	assert.Equal(t, endsAt, *groupRole.EndsAt)
}

func TestGroupRole_TableName(t *testing.T) {
	groupRole := &GroupRole{}
	assert.Equal(t, "group_roles", groupRole.TableName())
}

func TestGroupRole_GetTableIdentifier(t *testing.T) {
	groupRole := &GroupRole{}
	assert.Equal(t, "GRPR", groupRole.GetTableIdentifier())
}

func TestGroupRole_GetResourceType(t *testing.T) {
	groupRole := &GroupRole{}
	assert.Equal(t, ResourceTypeGroupRole, groupRole.GetResourceType())
}

func TestGroupRole_IsEffective(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *GroupRole
		testTime time.Time
		expected bool
	}{
		{
			name: "active with no time bounds",
			setup: func() *GroupRole {
				return &GroupRole{IsActive: true}
			},
			testTime: time.Now(),
			expected: true,
		},
		{
			name: "inactive",
			setup: func() *GroupRole {
				return &GroupRole{IsActive: false}
			},
			testTime: time.Now(),
			expected: false,
		},
		{
			name: "before start time",
			setup: func() *GroupRole {
				startsAt := time.Now().Add(1 * time.Hour)
				return &GroupRole{
					IsActive: true,
					StartsAt: &startsAt,
				}
			},
			testTime: time.Now(),
			expected: false,
		},
		{
			name: "after start time",
			setup: func() *GroupRole {
				startsAt := time.Now().Add(-1 * time.Hour)
				return &GroupRole{
					IsActive: true,
					StartsAt: &startsAt,
				}
			},
			testTime: time.Now(),
			expected: true,
		},
		{
			name: "after end time",
			setup: func() *GroupRole {
				endsAt := time.Now().Add(-1 * time.Hour)
				return &GroupRole{
					IsActive: true,
					EndsAt:   &endsAt,
				}
			},
			testTime: time.Now(),
			expected: false,
		},
		{
			name: "before end time",
			setup: func() *GroupRole {
				endsAt := time.Now().Add(1 * time.Hour)
				return &GroupRole{
					IsActive: true,
					EndsAt:   &endsAt,
				}
			},
			testTime: time.Now(),
			expected: true,
		},
		{
			name: "within time bounds",
			setup: func() *GroupRole {
				now := time.Now()
				startsAt := now.Add(-1 * time.Hour)
				endsAt := now.Add(1 * time.Hour)
				return &GroupRole{
					IsActive: true,
					StartsAt: &startsAt,
					EndsAt:   &endsAt,
				}
			},
			testTime: time.Now(),
			expected: true,
		},
		{
			name: "outside time bounds (before)",
			setup: func() *GroupRole {
				now := time.Now()
				startsAt := now.Add(1 * time.Hour)
				endsAt := now.Add(2 * time.Hour)
				return &GroupRole{
					IsActive: true,
					StartsAt: &startsAt,
					EndsAt:   &endsAt,
				}
			},
			testTime: time.Now(),
			expected: false,
		},
		{
			name: "outside time bounds (after)",
			setup: func() *GroupRole {
				now := time.Now()
				startsAt := now.Add(-2 * time.Hour)
				endsAt := now.Add(-1 * time.Hour)
				return &GroupRole{
					IsActive: true,
					StartsAt: &startsAt,
					EndsAt:   &endsAt,
				}
			},
			testTime: time.Now(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupRole := tt.setup()
			result := groupRole.IsEffective(tt.testTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupRole_IsCurrentlyEffective(t *testing.T) {
	// Test active group role with no time bounds
	groupRole := &GroupRole{IsActive: true}
	assert.True(t, groupRole.IsCurrentlyEffective())

	// Test inactive group role
	groupRole = &GroupRole{IsActive: false}
	assert.False(t, groupRole.IsCurrentlyEffective())

	// Test group role that starts in the future
	startsAt := time.Now().Add(1 * time.Hour)
	groupRole = &GroupRole{
		IsActive: true,
		StartsAt: &startsAt,
	}
	assert.False(t, groupRole.IsCurrentlyEffective())

	// Test group role that ended in the past
	endsAt := time.Now().Add(-1 * time.Hour)
	groupRole = &GroupRole{
		IsActive: true,
		EndsAt:   &endsAt,
	}
	assert.False(t, groupRole.IsCurrentlyEffective())
}

func TestGroupRole_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *GroupRole
		expectError bool
		errorType   error
	}{
		{
			name: "valid group role",
			setup: func() *GroupRole {
				return &GroupRole{
					GroupID:        "group-id",
					RoleID:         "role-id",
					OrganizationID: "org-id",
					AssignedBy:     "user-id",
				}
			},
			expectError: false,
		},
		{
			name: "missing group ID",
			setup: func() *GroupRole {
				return &GroupRole{
					RoleID:         "role-id",
					OrganizationID: "org-id",
					AssignedBy:     "user-id",
				}
			},
			expectError: true,
			errorType:   ErrInvalidGroupID,
		},
		{
			name: "missing role ID",
			setup: func() *GroupRole {
				return &GroupRole{
					GroupID:        "group-id",
					OrganizationID: "org-id",
					AssignedBy:     "user-id",
				}
			},
			expectError: true,
			errorType:   ErrInvalidRoleID,
		},
		{
			name: "missing organization ID",
			setup: func() *GroupRole {
				return &GroupRole{
					GroupID:    "group-id",
					RoleID:     "role-id",
					AssignedBy: "user-id",
				}
			},
			expectError: true,
			errorType:   ErrInvalidOrganizationID,
		},
		{
			name: "missing assigned by",
			setup: func() *GroupRole {
				return &GroupRole{
					GroupID:        "group-id",
					RoleID:         "role-id",
					OrganizationID: "org-id",
				}
			},
			expectError: true,
			errorType:   ErrInvalidAssignedBy,
		},
		{
			name: "invalid time range",
			setup: func() *GroupRole {
				startsAt := time.Now()
				endsAt := startsAt.Add(-1 * time.Hour) // ends before it starts
				return &GroupRole{
					GroupID:        "group-id",
					RoleID:         "role-id",
					OrganizationID: "org-id",
					AssignedBy:     "user-id",
					StartsAt:       &startsAt,
					EndsAt:         &endsAt,
				}
			},
			expectError: true,
			errorType:   ErrInvalidTimeRange,
		},
		{
			name: "valid time range",
			setup: func() *GroupRole {
				startsAt := time.Now()
				endsAt := startsAt.Add(1 * time.Hour)
				return &GroupRole{
					GroupID:        "group-id",
					RoleID:         "role-id",
					OrganizationID: "org-id",
					AssignedBy:     "user-id",
					StartsAt:       &startsAt,
					EndsAt:         &endsAt,
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupRole := tt.setup()
			err := groupRole.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGroupRole_GORMHooks(t *testing.T) {
	groupRole := NewGroupRole("group-id", "role-id", "org-id", "user-id")

	// Test BeforeCreate hook
	err := groupRole.BeforeCreate()
	assert.NoError(t, err)

	// Test BeforeUpdate hook
	err = groupRole.BeforeUpdate()
	assert.NoError(t, err)

	// Test BeforeDelete hook
	err = groupRole.BeforeDelete()
	assert.NoError(t, err)

	// Test BeforeSoftDelete hook
	err = groupRole.BeforeSoftDelete()
	assert.NoError(t, err)
}

func TestGroupRole_GetAndSetID(t *testing.T) {
	groupRole := NewGroupRole("group-id", "role-id", "org-id", "user-id")

	// Test that ID is generated
	assert.NotEmpty(t, groupRole.GetID())

	// Test SetID
	newID := "new-test-id"
	groupRole.SetID(newID)
	assert.Equal(t, newID, groupRole.GetID())
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "test_field",
		Message: "test message",
	}
	assert.Equal(t, "test message", err.Error())
}

func TestGroupRole_Relationships(t *testing.T) {
	groupRole := NewGroupRole("group-id", "role-id", "org-id", "user-id")

	// Test that relationship fields are properly set
	assert.Equal(t, "group-id", groupRole.GroupID)
	assert.Equal(t, "role-id", groupRole.RoleID)
	assert.Equal(t, "org-id", groupRole.OrganizationID)
	assert.Equal(t, "user-id", groupRole.AssignedBy)

	// Test that relationship pointers are nil initially
	assert.Nil(t, groupRole.Group)
	assert.Nil(t, groupRole.Role)
	assert.Nil(t, groupRole.Organization)
	assert.Nil(t, groupRole.Assigner)
}

func TestGroupRole_TableSize(t *testing.T) {
	groupRole := &GroupRole{}
	// Based on the pattern from other models, GroupRole should use hash.Small
	assert.Equal(t, "GRPR", groupRole.GetTableIdentifier())
}
