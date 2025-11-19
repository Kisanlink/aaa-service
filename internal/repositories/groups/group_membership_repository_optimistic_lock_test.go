package groups

import (
	"context"
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	pkgErrors "github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMembershipUpdateWithVersion_Success tests successful membership update
func TestMembershipUpdateWithVersion_Success(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestMembershipRepository(t)
	defer cleanup()

	groupID, principalID := setupTestGroupAndPrincipal(t, ctx)

	// Create membership
	membership := models.NewGroupMembership(groupID, principalID, "user", "admin-user")
	membership.CreatedBy = "admin-user"
	membership.UpdatedBy = "admin-user"

	err := repo.Create(ctx, membership)
	require.NoError(t, err)
	require.Equal(t, 1, membership.Version)

	// Update membership with correct version
	futureTime := time.Now().Add(30 * 24 * time.Hour)
	membership.EndsAt = &futureTime
	err = repo.UpdateWithVersion(ctx, membership, 1)
	require.NoError(t, err)
	require.Equal(t, 2, membership.Version)

	// Verify update
	retrieved, err := repo.GetByID(ctx, membership.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.EndsAt)
	assert.Equal(t, 2, retrieved.Version)
}

// TestMembershipUpdateWithVersion_Conflict tests concurrent membership updates
func TestMembershipUpdateWithVersion_Conflict(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestMembershipRepository(t)
	defer cleanup()

	groupID, principalID := setupTestGroupAndPrincipal(t, ctx)

	// Create membership
	membership := models.NewGroupMembership(groupID, principalID, "user", "admin-user")
	membership.CreatedBy = "admin-user"
	membership.UpdatedBy = "admin-user"

	err := repo.Create(ctx, membership)
	require.NoError(t, err)
	membershipID := membership.ID

	// Two admins fetch the same membership
	admin1Membership, err := repo.GetByID(ctx, membershipID)
	require.NoError(t, err)

	admin2Membership, err := repo.GetByID(ctx, membershipID)
	require.NoError(t, err)

	// Admin 1 sets end date - should succeed
	endDate1 := time.Now().Add(7 * 24 * time.Hour)
	admin1Membership.EndsAt = &endDate1
	admin1Membership.UpdatedBy = "admin-1"
	err = repo.UpdateWithVersion(ctx, admin1Membership, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, admin1Membership.Version)

	// Admin 2 tries to deactivate with stale version - should fail
	admin2Membership.IsActive = false
	admin2Membership.UpdatedBy = "admin-2"
	err = repo.UpdateWithVersion(ctx, admin2Membership, 1)
	require.Error(t, err)
	assert.True(t, pkgErrors.IsOptimisticLockError(err))

	// Verify admin 1's changes persisted
	current, err := repo.GetByID(ctx, membershipID)
	require.NoError(t, err)
	assert.NotNil(t, current.EndsAt)
	assert.True(t, current.IsActive) // Admin 2's deactivation didn't go through
	assert.Equal(t, 2, current.Version)
}

// TestMembershipUpdateWithVersion_TimeBoundsConflict tests time-bound conflicts
func TestMembershipUpdateWithVersion_TimeBoundsConflict(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestMembershipRepository(t)
	defer cleanup()

	groupID, principalID := setupTestGroupAndPrincipal(t, ctx)

	// Create membership with start date
	startDate := time.Now().Add(-1 * time.Hour)
	membership := models.NewGroupMembership(groupID, principalID, "user", "admin-user")
	membership.StartsAt = &startDate
	membership.CreatedBy = "admin-user"
	membership.UpdatedBy = "admin-user"

	err := repo.Create(ctx, membership)
	require.NoError(t, err)
	membershipID := membership.ID
	currentVersion := membership.Version

	// Two processes attempt to set different end dates
	process1Membership, err := repo.GetByID(ctx, membershipID)
	require.NoError(t, err)

	process2Membership, err := repo.GetByID(ctx, membershipID)
	require.NoError(t, err)

	// Process 1 sets end date to 30 days
	endDate1 := time.Now().Add(30 * 24 * time.Hour)
	process1Membership.EndsAt = &endDate1
	process1Membership.UpdatedBy = "process-1"
	err = repo.UpdateWithVersion(ctx, process1Membership, currentVersion)
	require.NoError(t, err)

	// Process 2 tries to set end date to 60 days with stale version
	endDate2 := time.Now().Add(60 * 24 * time.Hour)
	process2Membership.EndsAt = &endDate2
	process2Membership.UpdatedBy = "process-2"
	err = repo.UpdateWithVersion(ctx, process2Membership, currentVersion)
	require.Error(t, err)
	assert.True(t, pkgErrors.IsOptimisticLockError(err))

	// Verify process 1's end date is set
	final, err := repo.GetByID(ctx, membershipID)
	require.NoError(t, err)
	assert.NotNil(t, final.EndsAt)
	assert.WithinDuration(t, endDate1, *final.EndsAt, 1*time.Second)
}

// TestMembershipUpdateWithVersion_ActivationToggle tests concurrent activation changes
func TestMembershipUpdateWithVersion_ActivationToggle(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestMembershipRepository(t)
	defer cleanup()

	groupID, principalID := setupTestGroupAndPrincipal(t, ctx)

	// Create active membership
	membership := models.NewGroupMembership(groupID, principalID, "user", "admin-user")
	membership.IsActive = true
	membership.CreatedBy = "admin-user"
	membership.UpdatedBy = "admin-user"

	err := repo.Create(ctx, membership)
	require.NoError(t, err)
	membershipID := membership.ID

	// Two admins fetch the membership
	admin1, err := repo.GetByID(ctx, membershipID)
	require.NoError(t, err)

	admin2, err := repo.GetByID(ctx, membershipID)
	require.NoError(t, err)

	// Admin 1 deactivates - should succeed
	admin1.IsActive = false
	admin1.UpdatedBy = "admin-1"
	err = repo.UpdateWithVersion(ctx, admin1, 1)
	require.NoError(t, err)

	// Admin 2 tries to extend end date while it's being deactivated - should fail
	futureDate := time.Now().Add(90 * 24 * time.Hour)
	admin2.EndsAt = &futureDate
	admin2.UpdatedBy = "admin-2"
	err = repo.UpdateWithVersion(ctx, admin2, 1)
	require.Error(t, err)
	assert.True(t, pkgErrors.IsOptimisticLockError(err))

	// Admin 2 must fetch latest and decide whether to reactivate or update end date
	latest, err := repo.GetByID(ctx, membershipID)
	require.NoError(t, err)
	assert.False(t, latest.IsActive)
	assert.Equal(t, 2, latest.Version)
}

// TestMembershipUpdateWithVersion_SequentialTimeUpdates tests sequential time-bound updates
func TestMembershipUpdateWithVersion_SequentialTimeUpdates(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestMembershipRepository(t)
	defer cleanup()

	groupID, principalID := setupTestGroupAndPrincipal(t, ctx)

	// Create membership
	membership := models.NewGroupMembership(groupID, principalID, "user", "admin-user")
	membership.CreatedBy = "admin-user"
	membership.UpdatedBy = "admin-user"

	err := repo.Create(ctx, membership)
	require.NoError(t, err)
	currentVersion := membership.Version

	// Update 1: Set start date
	startDate := time.Now()
	membership.StartsAt = &startDate
	err = repo.UpdateWithVersion(ctx, membership, currentVersion)
	require.NoError(t, err)
	currentVersion = membership.Version

	// Update 2: Set end date
	endDate := time.Now().Add(30 * 24 * time.Hour)
	membership.EndsAt = &endDate
	err = repo.UpdateWithVersion(ctx, membership, currentVersion)
	require.NoError(t, err)
	currentVersion = membership.Version

	// Update 3: Extend end date
	extendedEndDate := time.Now().Add(60 * 24 * time.Hour)
	membership.EndsAt = &extendedEndDate
	err = repo.UpdateWithVersion(ctx, membership, currentVersion)
	require.NoError(t, err)
	currentVersion = membership.Version

	// Verify final state
	assert.Equal(t, 4, currentVersion) // 1 initial + 3 updates
	retrieved, err := repo.GetByID(ctx, membership.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.StartsAt)
	assert.NotNil(t, retrieved.EndsAt)
	assert.WithinDuration(t, extendedEndDate, *retrieved.EndsAt, 1*time.Second)
	assert.Equal(t, 4, retrieved.Version)
}

// TestMembershipUpdateWithVersion_NotFound tests update on non-existent membership
func TestMembershipUpdateWithVersion_NotFound(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestMembershipRepository(t)
	defer cleanup()

	groupID, principalID := setupTestGroupAndPrincipal(t, ctx)

	// Try to update non-existent membership
	membership := models.NewGroupMembership(groupID, principalID, "user", "admin-user")
	membership.ID = "GRPM_NONEXISTENT123"
	membership.UpdatedBy = "admin-user"

	err := repo.UpdateWithVersion(ctx, membership, 1)
	require.Error(t, err)
	assert.True(t, pkgErrors.IsNotFoundError(err))
}

// Helper functions

func setupTestMembershipRepository(t *testing.T) (*GroupMembershipRepository, func()) {
	// This should be implemented according to your test infrastructure
	t.Skip("Test repository setup needs to be implemented with actual DB connection")
	return nil, func() {}
}

func setupTestGroupAndPrincipal(t *testing.T, ctx context.Context) (groupID, principalID string) {
	// This should create a test group and principal and return their IDs
	// Implementation depends on your test setup
	return "GRPN_TEST123", "USER_TEST123"
}
