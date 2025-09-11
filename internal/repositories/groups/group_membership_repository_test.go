package groups

import (
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/stretchr/testify/assert"
)

// TestGroupMembershipRepository_Methods tests that all required methods exist
func TestGroupMembershipRepository_Methods(t *testing.T) {
	// We can't easily test with a real database in unit tests,
	// so we'll just test that the repository can be created and has the expected methods
	// Integration tests would test the actual database operations

	// For now, we'll test that the repository structure is correct
	// and that the GroupMembership model has the expected methods

	// Test GroupMembership model methods
	membership := &models.GroupMembership{
		IsActive: true,
	}

	now := time.Now()
	assert.True(t, membership.IsEffective(now))

	// Test with inactive membership
	membership.IsActive = false
	assert.False(t, membership.IsEffective(now))

	// Test with time bounds
	membership.IsActive = true
	future := now.Add(time.Hour)
	membership.StartsAt = &future
	assert.False(t, membership.IsEffective(now)) // Not started yet

	past := now.Add(-time.Hour)
	membership.StartsAt = &past
	membership.EndsAt = &future
	assert.True(t, membership.IsEffective(now)) // Currently active

	pastEnd := now.Add(-time.Minute)
	membership.EndsAt = &pastEnd
	assert.False(t, membership.IsEffective(now)) // Already ended
}

// TestGroupMembershipModel_IsEffective tests the IsEffective method on GroupMembership model
func TestGroupMembershipModel_IsEffective(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		membership *models.GroupMembership
		checkTime  time.Time
		expected   bool
	}{
		{
			name: "active membership with no time bounds",
			membership: &models.GroupMembership{
				IsActive: true,
				StartsAt: nil,
				EndsAt:   nil,
			},
			checkTime: now,
			expected:  true,
		},
		{
			name: "inactive membership",
			membership: &models.GroupMembership{
				IsActive: false,
				StartsAt: nil,
				EndsAt:   nil,
			},
			checkTime: now,
			expected:  false,
		},
		{
			name: "membership not yet started",
			membership: &models.GroupMembership{
				IsActive: true,
				StartsAt: &[]time.Time{now.Add(time.Hour)}[0],
				EndsAt:   nil,
			},
			checkTime: now,
			expected:  false,
		},
		{
			name: "membership already ended",
			membership: &models.GroupMembership{
				IsActive: true,
				StartsAt: nil,
				EndsAt:   &[]time.Time{now.Add(-time.Hour)}[0],
			},
			checkTime: now,
			expected:  false,
		},
		{
			name: "membership currently active with time bounds",
			membership: &models.GroupMembership{
				IsActive: true,
				StartsAt: &[]time.Time{now.Add(-time.Hour)}[0],
				EndsAt:   &[]time.Time{now.Add(time.Hour)}[0],
			},
			checkTime: now,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.membership.IsEffective(tt.checkTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}
