package organizations

import (
	"context"
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	pkgErrors "github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateWithVersion_Success tests successful update with correct version
func TestUpdateWithVersion_Success(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	// Create test organization
	org := models.NewOrganization("Test Org", "Test Description", models.OrgTypeEnterprise)
	org.CreatedBy = "test-user"
	org.UpdatedBy = "test-user"

	err := repo.Create(ctx, org)
	require.NoError(t, err)
	require.NotEmpty(t, org.ID)
	require.Equal(t, 1, org.Version) // Initial version should be 1

	// Update with correct version
	org.Name = "Updated Test Org"
	org.Description = "Updated Description"
	err = repo.UpdateWithVersion(ctx, org, 1)
	require.NoError(t, err)
	require.Equal(t, 2, org.Version) // Version should increment to 2

	// Verify update persisted
	retrieved, err := repo.GetByID(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Org", retrieved.Name)
	assert.Equal(t, "Updated Description", retrieved.Description)
	assert.Equal(t, 2, retrieved.Version)
}

// TestUpdateWithVersion_Conflict tests version conflict handling
func TestUpdateWithVersion_Conflict(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	// Create test organization
	org := models.NewOrganization("Test Org", "Test Description", models.OrgTypeEnterprise)
	org.CreatedBy = "test-user"
	org.UpdatedBy = "test-user"

	err := repo.Create(ctx, org)
	require.NoError(t, err)
	require.Equal(t, 1, org.Version)

	// First update - should succeed
	org.Name = "First Update"
	err = repo.UpdateWithVersion(ctx, org, 1)
	require.NoError(t, err)
	require.Equal(t, 2, org.Version)

	// Try to update with stale version - should fail
	org.Name = "Second Update with Stale Version"
	err = repo.UpdateWithVersion(ctx, org, 1) // Using old version
	require.Error(t, err)
	assert.True(t, pkgErrors.IsOptimisticLockError(err), "Expected OptimisticLockError")

	// Verify that the update didn't go through
	retrieved, err := repo.GetByID(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, "First Update", retrieved.Name, "Name should not have changed")
	assert.Equal(t, 2, retrieved.Version)
}

// TestUpdateWithVersion_ConcurrentUpdates simulates concurrent update scenario
func TestUpdateWithVersion_ConcurrentUpdates(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	// Create test organization
	org := models.NewOrganization("Test Org", "Test Description", models.OrgTypeEnterprise)
	org.CreatedBy = "test-user"
	org.UpdatedBy = "test-user"

	err := repo.Create(ctx, org)
	require.NoError(t, err)
	orgID := org.ID

	// Simulate two concurrent clients reading the same version
	client1Org, err := repo.GetByID(ctx, orgID)
	require.NoError(t, err)
	require.Equal(t, 1, client1Org.Version)

	client2Org, err := repo.GetByID(ctx, orgID)
	require.NoError(t, err)
	require.Equal(t, 1, client2Org.Version)

	// Client 1 updates first - should succeed
	client1Org.Name = "Client 1 Update"
	client1Org.UpdatedBy = "client-1"
	err = repo.UpdateWithVersion(ctx, client1Org, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, client1Org.Version)

	// Client 2 tries to update with stale version - should fail
	client2Org.Name = "Client 2 Update"
	client2Org.UpdatedBy = "client-2"
	err = repo.UpdateWithVersion(ctx, client2Org, 1)
	require.Error(t, err)
	assert.True(t, pkgErrors.IsOptimisticLockError(err))

	// Client 2 should retry with latest version
	latest, err := repo.GetByID(ctx, orgID)
	require.NoError(t, err)
	assert.Equal(t, 2, latest.Version)
	assert.Equal(t, "Client 1 Update", latest.Name)

	// Client 2 retry with correct version - should succeed
	latest.Name = "Client 2 Retry Update"
	latest.UpdatedBy = "client-2"
	err = repo.UpdateWithVersion(ctx, latest, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, latest.Version)

	// Final verification
	final, err := repo.GetByID(ctx, orgID)
	require.NoError(t, err)
	assert.Equal(t, "Client 2 Retry Update", final.Name)
	assert.Equal(t, 3, final.Version)
}

// TestUpdateWithVersion_NotFound tests update on non-existent organization
func TestUpdateWithVersion_NotFound(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	// Try to update non-existent organization
	org := models.NewOrganization("Non Existent", "Description", models.OrgTypeEnterprise)
	org.ID = "ORGN_NONEXISTENT123"
	org.UpdatedBy = "test-user"

	err := repo.UpdateWithVersion(ctx, org, 1)
	require.Error(t, err)
	assert.True(t, pkgErrors.IsNotFoundError(err))
}

// TestUpdateWithVersion_SequentialUpdates tests multiple sequential updates
func TestUpdateWithVersion_SequentialUpdates(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	// Create test organization
	org := models.NewOrganization("Test Org", "Test Description", models.OrgTypeEnterprise)
	org.CreatedBy = "test-user"
	org.UpdatedBy = "test-user"

	err := repo.Create(ctx, org)
	require.NoError(t, err)
	currentVersion := org.Version

	// Perform 5 sequential updates
	for i := 1; i <= 5; i++ {
		org.Name = "Update " + string(rune('0'+i))
		err = repo.UpdateWithVersion(ctx, org, currentVersion)
		require.NoError(t, err)
		currentVersion++
		assert.Equal(t, currentVersion, org.Version)
	}

	// Verify final state
	assert.Equal(t, 6, org.Version) // 1 initial + 5 updates
	assert.Equal(t, "Update 5", org.Name)

	// Verify persisted state
	retrieved, err := repo.GetByID(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, 6, retrieved.Version)
	assert.Equal(t, "Update 5", retrieved.Name)
}

// TestUpdateWithVersion_ParentHierarchyChange tests updating parent with version control
func TestUpdateWithVersion_ParentHierarchyChange(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	repo, cleanup := setupTestRepository(t)
	defer cleanup()

	// Create parent organization
	parent := models.NewOrganization("Parent Org", "Parent", models.OrgTypeEnterprise)
	parent.CreatedBy = "test-user"
	parent.UpdatedBy = "test-user"
	err := repo.Create(ctx, parent)
	require.NoError(t, err)

	// Create child organization
	child := models.NewOrganization("Child Org", "Child", models.OrgTypeSmallBusiness)
	child.CreatedBy = "test-user"
	child.UpdatedBy = "test-user"
	err = repo.Create(ctx, child)
	require.NoError(t, err)
	childVersion := child.Version

	// Update child to set parent with version control
	child.ParentID = &parent.ID
	err = repo.UpdateWithVersion(ctx, child, childVersion)
	require.NoError(t, err)
	assert.Equal(t, childVersion+1, child.Version)

	// Verify parent-child relationship
	retrieved, err := repo.GetByID(ctx, child.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.ParentID)
	assert.Equal(t, parent.ID, *retrieved.ParentID)
}

// setupTestRepository creates a test repository with a test database
// This is a helper function - actual implementation depends on your test setup
func setupTestRepository(t *testing.T) (*OrganizationRepository, func()) {
	// This should be implemented according to your test infrastructure
	// For now, returning nil - needs actual DB setup
	t.Skip("Test repository setup needs to be implemented with actual DB connection")
	return nil, func() {}
}
