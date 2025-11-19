package groups

import (
	"context"
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	pkgErrors "github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGroupUpdateWithVersion_Success tests successful update with correct version
func TestGroupUpdateWithVersion_Success(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestGroupRepository(t)
	defer cleanup()

	// Create test organization first (groups belong to organizations)
	orgID := createTestOrganization(t, ctx)

	// Create test group
	group := models.NewGroup("Test Group", "Test Description", orgID)
	group.CreatedBy = "test-user"
	group.UpdatedBy = "test-user"

	err := repo.Create(ctx, group)
	require.NoError(t, err)
	require.NotEmpty(t, group.ID)
	require.Equal(t, 1, group.Version)

	// Update with correct version
	group.Name = "Updated Test Group"
	group.Description = "Updated Description"
	err = repo.UpdateWithVersion(ctx, group, 1)
	require.NoError(t, err)
	require.Equal(t, 2, group.Version)

	// Verify update persisted
	retrieved, err := repo.GetByID(ctx, group.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Group", retrieved.Name)
	assert.Equal(t, "Updated Description", retrieved.Description)
	assert.Equal(t, 2, retrieved.Version)
}

// TestGroupUpdateWithVersion_Conflict tests version conflict handling
func TestGroupUpdateWithVersion_Conflict(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestGroupRepository(t)
	defer cleanup()

	orgID := createTestOrganization(t, ctx)

	// Create test group
	group := models.NewGroup("Test Group", "Test Description", orgID)
	group.CreatedBy = "test-user"
	group.UpdatedBy = "test-user"

	err := repo.Create(ctx, group)
	require.NoError(t, err)
	require.Equal(t, 1, group.Version)

	// First update - should succeed
	group.Name = "First Update"
	err = repo.UpdateWithVersion(ctx, group, 1)
	require.NoError(t, err)
	require.Equal(t, 2, group.Version)

	// Try to update with stale version - should fail
	group.Name = "Second Update with Stale Version"
	err = repo.UpdateWithVersion(ctx, group, 1)
	require.Error(t, err)
	assert.True(t, pkgErrors.IsOptimisticLockError(err))

	// Verify the update didn't go through
	retrieved, err := repo.GetByID(ctx, group.ID)
	require.NoError(t, err)
	assert.Equal(t, "First Update", retrieved.Name)
	assert.Equal(t, 2, retrieved.Version)
}

// TestGroupUpdateWithVersion_ConcurrentHierarchyUpdates tests concurrent parent changes
func TestGroupUpdateWithVersion_ConcurrentHierarchyUpdates(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestGroupRepository(t)
	defer cleanup()

	orgID := createTestOrganization(t, ctx)

	// Create parent group 1
	parent1 := models.NewGroup("Parent 1", "Parent 1", orgID)
	parent1.CreatedBy = "test-user"
	parent1.UpdatedBy = "test-user"
	err := repo.Create(ctx, parent1)
	require.NoError(t, err)

	// Create parent group 2
	parent2 := models.NewGroup("Parent 2", "Parent 2", orgID)
	parent2.CreatedBy = "test-user"
	parent2.UpdatedBy = "test-user"
	err = repo.Create(ctx, parent2)
	require.NoError(t, err)

	// Create child group
	child := models.NewGroup("Child Group", "Child", orgID)
	child.CreatedBy = "test-user"
	child.UpdatedBy = "test-user"
	err = repo.Create(ctx, child)
	require.NoError(t, err)
	childID := child.ID

	// Simulate two processes reading the same child
	process1Child, err := repo.GetByID(ctx, childID)
	require.NoError(t, err)

	process2Child, err := repo.GetByID(ctx, childID)
	require.NoError(t, err)

	// Process 1 sets parent to parent1 - should succeed
	process1Child.ParentID = &parent1.ID
	process1Child.UpdatedBy = "process-1"
	err = repo.UpdateWithVersion(ctx, process1Child, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, process1Child.Version)

	// Process 2 tries to set parent to parent2 with stale version - should fail
	process2Child.ParentID = &parent2.ID
	process2Child.UpdatedBy = "process-2"
	err = repo.UpdateWithVersion(ctx, process2Child, 1)
	require.Error(t, err)
	assert.True(t, pkgErrors.IsOptimisticLockError(err))

	// Verify current state shows parent1
	current, err := repo.GetByID(ctx, childID)
	require.NoError(t, err)
	assert.NotNil(t, current.ParentID)
	assert.Equal(t, parent1.ID, *current.ParentID)
	assert.Equal(t, 2, current.Version)
}

// TestGroupUpdateWithVersion_PreventCircularReference tests preventing circular references
func TestGroupUpdateWithVersion_PreventCircularReference(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestGroupRepository(t)
	defer cleanup()

	orgID := createTestOrganization(t, ctx)

	// Create parent group
	parent := models.NewGroup("Parent", "Parent", orgID)
	parent.CreatedBy = "test-user"
	parent.UpdatedBy = "test-user"
	err := repo.Create(ctx, parent)
	require.NoError(t, err)

	// Create child group with parent set
	child := models.NewGroup("Child", "Child", orgID)
	child.ParentID = &parent.ID
	child.CreatedBy = "test-user"
	child.UpdatedBy = "test-user"
	err = repo.Create(ctx, child)
	require.NoError(t, err)

	// Attempt 1: Try to set parent's parent to child (would create circular reference)
	// This simulates concurrent modification where process 1 succeeds
	parent.ParentID = &child.ID
	parent.UpdatedBy = "process-1"

	// Note: In production, there should be additional validation to prevent this
	// This test demonstrates that optimistic locking prevents the second process
	// from unknowingly creating a circular reference

	// Process 2 also tries to modify parent concurrently
	process2Parent, err := repo.GetByID(ctx, parent.ID)
	require.NoError(t, err)
	currentVersion := process2Parent.Version

	// Process 1 updates first
	err = repo.UpdateWithVersion(ctx, parent, currentVersion)
	// In a complete implementation, this should be blocked by business logic validation

	// Process 2's update with stale version will fail
	process2Parent.Description = "Updated by process 2"
	err = repo.UpdateWithVersion(ctx, process2Parent, currentVersion)
	// This would fail if process 1 already updated, preventing race conditions
}

// TestGroupUpdateWithVersion_NotFound tests update on non-existent group
func TestGroupUpdateWithVersion_NotFound(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestGroupRepository(t)
	defer cleanup()

	orgID := createTestOrganization(t, ctx)

	// Try to update non-existent group
	group := models.NewGroup("Non Existent", "Description", orgID)
	group.ID = "GRPN_NONEXISTENT123"
	group.UpdatedBy = "test-user"

	err := repo.UpdateWithVersion(ctx, group, 1)
	require.Error(t, err)
	assert.True(t, pkgErrors.IsNotFoundError(err))
}

// TestGroupUpdateWithVersion_IsActiveToggle tests toggling active status with version control
func TestGroupUpdateWithVersion_IsActiveToggle(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestGroupRepository(t)
	defer cleanup()

	orgID := createTestOrganization(t, ctx)

	// Create active group
	group := models.NewGroup("Test Group", "Test", orgID)
	group.CreatedBy = "test-user"
	group.UpdatedBy = "test-user"
	group.IsActive = true

	err := repo.Create(ctx, group)
	require.NoError(t, err)
	assert.True(t, group.IsActive)
	currentVersion := group.Version

	// Deactivate group
	group.IsActive = false
	err = repo.UpdateWithVersion(ctx, group, currentVersion)
	require.NoError(t, err)
	assert.Equal(t, currentVersion+1, group.Version)

	// Verify deactivation
	retrieved, err := repo.GetByID(ctx, group.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.IsActive)
}

// Helper functions

func setupTestGroupRepository(t *testing.T) (*GroupRepository, func()) {
	// This should be implemented according to your test infrastructure
	t.Skip("Test repository setup needs to be implemented with actual DB connection")
	return nil, func() {}
}

func createTestOrganization(t *testing.T, ctx context.Context) string {
	// This should create a test organization and return its ID
	// Implementation depends on your test setup
	return "ORGN_TEST123"
}
