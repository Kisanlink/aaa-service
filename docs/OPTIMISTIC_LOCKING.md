# Optimistic Locking Implementation

## Overview

This document describes the optimistic locking implementation in the AAA service to prevent race conditions during concurrent hierarchy updates in Organizations, Groups, and Group Memberships.

## Problem Statement

Without optimistic locking, concurrent updates to hierarchical structures can lead to:

1. **Circular References**: Process A moves Group X under Group Y while Process B moves Group Y under Group X
2. **Invalid Hierarchy States**: Concurrent parent changes creating inconsistent relationships
3. **Lost Updates**: One update overwrites another without the second process knowing
4. **Data Integrity Issues**: Violations of business rules due to concurrent modifications

## Solution: Optimistic Locking with Version Counter

### Design Pattern

We implement optimistic locking using a version counter that:

1. **Increments** on every update
2. **Validates** version matches before allowing updates
3. **Returns conflict error** (HTTP 409) when version mismatch detected
4. **Requires client retry** with latest version

### Architecture

```
┌─────────────────┐
│  Client Request │
│  (with version) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Repository    │
│ UpdateWithVer() │──► Check: current_version == expected_version?
└────────┬────────┘          │
         │                    ├─► Yes: UPDATE ... SET version = version + 1
         │                    │
         │                    └─► No:  Return OptimisticLockError
         │
         ▼
┌─────────────────┐
│   Database      │
│   (Atomic CAS)  │
└─────────────────┘
```

## Implementation Details

### 1. Database Schema

Added `version` column to three tables:

```sql
-- organizations
ALTER TABLE organizations ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- groups
ALTER TABLE groups ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- group_memberships
ALTER TABLE group_memberships ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
```

**Indexes for Performance:**
```sql
CREATE INDEX idx_organizations_id_version ON organizations(id, version);
CREATE INDEX idx_groups_id_version ON groups(id, version);
CREATE INDEX idx_group_memberships_id_version ON group_memberships(id, version);
```

### 2. Model Changes

Each model now includes a `Version` field:

```go
type Organization struct {
    *base.BaseModel
    Name        string
    // ... other fields
    Version     int `json:"version" gorm:"column:version;default:1;not null"`
}

type Group struct {
    *base.BaseModel
    Name           string
    // ... other fields
    Version        int `json:"version" gorm:"column:version;default:1;not null"`
}

type GroupMembership struct {
    *base.BaseModel
    GroupID       string
    // ... other fields
    Version       int `json:"version" gorm:"column:version;default:1;not null"`
}
```

### 3. Repository Pattern

#### UpdateWithVersion Method

```go
func (r *OrganizationRepository) UpdateWithVersion(
    ctx context.Context,
    org *models.Organization,
    expectedVersion int,
) error {
    // 1. Get current version
    var current models.Organization
    if err := db.Where("id = ?", org.ID).First(&current).Error; err != nil {
        return handleError(err)
    }

    // 2. Check version match
    if current.Version != expectedVersion {
        return pkgErrors.NewOptimisticLockError(
            "organization", org.ID, expectedVersion, current.Version,
        )
    }

    // 3. Atomic update with version check
    result := db.Model(&models.Organization{}).
        Where("id = ? AND version = ?", org.ID, expectedVersion).
        Updates(map[string]interface{}{
            "name":       org.Name,
            // ... other fields
            "version":    gorm.Expr("version + 1"),
            "updated_at": gorm.Expr("NOW()"),
        })

    // 4. Verify exactly one row affected
    if result.RowsAffected == 0 {
        // Re-fetch for accurate error reporting
        db.Where("id = ?", org.ID).First(&current)
        return pkgErrors.NewOptimisticLockError(
            "organization", org.ID, expectedVersion, current.Version,
        )
    }

    // 5. Update in-memory version
    org.Version = expectedVersion + 1
    return nil
}
```

### 4. Error Handling

#### OptimisticLockError Type

```go
type OptimisticLockError struct {
    message string
    details map[string]interface{}
}

func NewOptimisticLockError(
    resourceType, resourceID string,
    expectedVersion, actualVersion int,
) *OptimisticLockError {
    return &OptimisticLockError{
        message: "Resource has been modified by another process. Please retry with the latest version.",
        details: map[string]interface{}{
            "resource_type":     resourceType,
            "resource_id":       resourceID,
            "expected_version":  expectedVersion,
            "current_version":   actualVersion,
            "retry_recommended": true,
        },
    }
}
```

#### HTTP Response

When an optimistic lock conflict occurs, the API returns HTTP 409 Conflict:

```json
{
  "error": "OPTIMISTIC_LOCK_ERROR",
  "message": "Resource has been modified by another process. Please retry with the latest version.",
  "details": {
    "resource_type": "organization",
    "resource_id": "ORGN_ABC123",
    "expected_version": 2,
    "current_version": 3,
    "retry_recommended": true
  },
  "timestamp": "2025-11-17T10:30:00Z"
}
```

## Usage Examples

### Client-Side Retry Pattern

```go
func updateOrganizationWithRetry(
    ctx context.Context,
    orgID string,
    updateFunc func(*models.Organization),
    maxRetries int,
) error {
    for attempt := 0; attempt < maxRetries; attempt++ {
        // 1. Fetch latest version
        org, err := orgRepo.GetByID(ctx, orgID)
        if err != nil {
            return err
        }

        currentVersion := org.Version

        // 2. Apply changes
        updateFunc(org)

        // 3. Attempt update with version check
        err = orgRepo.UpdateWithVersion(ctx, org, currentVersion)

        if err == nil {
            return nil // Success
        }

        // 4. Check if it's a version conflict
        if pkgErrors.IsOptimisticLockError(err) {
            // Retry with exponential backoff
            time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
            continue
        }

        // 5. Other error - don't retry
        return err
    }

    return errors.New("max retries exceeded")
}
```

### Service Layer Example

```go
func (s *OrganizationService) UpdateOrganization(
    ctx context.Context,
    orgID string,
    req UpdateOrganizationRequest,
) (*models.Organization, error) {
    // Fetch current organization
    org, err := s.repo.GetByID(ctx, orgID)
    if err != nil {
        return nil, err
    }

    // Apply updates
    org.Name = req.Name
    org.Description = req.Description
    org.UpdatedBy = getUserIDFromContext(ctx)

    // Update with version check
    // Client must provide current version in request
    err = s.repo.UpdateWithVersion(ctx, org, req.Version)
    if err != nil {
        if pkgErrors.IsOptimisticLockError(err) {
            // Return 409 Conflict - client should retry
            return nil, err
        }
        return nil, fmt.Errorf("failed to update organization: %w", err)
    }

    return org, nil
}
```

### API Handler Example

```go
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
    var req UpdateOrganizationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }

    orgID := c.Param("id")
    org, err := h.service.UpdateOrganization(c.Request.Context(), orgID, req)

    if err != nil {
        if pkgErrors.IsOptimisticLockError(err) {
            // Return conflict with retry information
            c.JSON(409, gin.H{
                "error": "OPTIMISTIC_LOCK_ERROR",
                "message": err.Error(),
                "details": err.(*pkgErrors.OptimisticLockError).Details(),
            })
            return
        }

        c.JSON(500, gin.H{"error": "Internal server error"})
        return
    }

    c.JSON(200, org)
}
```

### Request/Response Example

**Update Request:**
```json
{
  "name": "Updated Organization Name",
  "description": "Updated description",
  "version": 2
}
```

**Success Response (200):**
```json
{
  "id": "ORGN_ABC123",
  "name": "Updated Organization Name",
  "description": "Updated description",
  "version": 3,
  "updated_at": "2025-11-17T10:30:00Z"
}
```

**Conflict Response (409):**
```json
{
  "error": "OPTIMISTIC_LOCK_ERROR",
  "message": "Resource has been modified by another process. Please retry with the latest version.",
  "details": {
    "resource_type": "organization",
    "resource_id": "ORGN_ABC123",
    "expected_version": 2,
    "current_version": 3,
    "retry_recommended": true
  }
}
```

## Concurrent Update Scenarios

### Scenario 1: Simple Concurrent Update

```
Time    Process A                   Process B
────────────────────────────────────────────────
T0      GET /org/123 (v1)
T1                                  GET /org/123 (v1)
T2      PUT /org/123 (v1) ✓
        → version becomes 2
T3                                  PUT /org/123 (v1) ✗
                                    → 409 Conflict
T4                                  GET /org/123 (v2)
T5                                  PUT /org/123 (v2) ✓
                                    → version becomes 3
```

### Scenario 2: Hierarchy Change Conflict

```
Time    Admin A                     Admin B
────────────────────────────────────────────────
T0      GET /group/child (v1)
        parent_id = null
T1                                  GET /group/child (v1)
                                    parent_id = null
T2      PUT /group/child (v1)
        Set parent_id = "parent1" ✓
        → version becomes 2
T3                                  PUT /group/child (v1)
                                    Set parent_id = "parent2" ✗
                                    → 409 Conflict (prevents corruption)
T4                                  GET /group/child (v2)
                                    parent_id = "parent1"
T5                                  Decide: keep parent1 or change to parent2
T6                                  PUT /group/child (v2)
                                    Set parent_id = "parent2" ✓
                                    → version becomes 3
```

## Testing

### Unit Tests

Tests are provided for each repository:

1. **organization_repository_optimistic_lock_test.go**
   - Success with correct version
   - Conflict with stale version
   - Concurrent updates
   - Sequential updates
   - Not found scenarios

2. **group_repository_optimistic_lock_test.go**
   - Hierarchy change conflicts
   - Concurrent parent updates
   - Circular reference prevention

3. **group_membership_repository_optimistic_lock_test.go**
   - Time-bound conflicts
   - Activation toggle conflicts
   - Sequential time updates

### Integration Testing

For integration tests, use the following pattern:

```go
func TestConcurrentOrganizationUpdates(t *testing.T) {
    // Setup
    org := createTestOrganization(t)

    // Simulate concurrent updates
    var wg sync.WaitGroup
    errors := make(chan error, 2)

    // Process 1
    wg.Add(1)
    go func() {
        defer wg.Done()
        org1, _ := repo.GetByID(ctx, org.ID)
        org1.Name = "Update 1"
        errors <- repo.UpdateWithVersion(ctx, org1, org1.Version)
    }()

    // Process 2
    wg.Add(1)
    go func() {
        defer wg.Done()
        org2, _ := repo.GetByID(ctx, org.ID)
        org2.Name = "Update 2"
        errors <- repo.UpdateWithVersion(ctx, org2, org2.Version)
    }()

    wg.Wait()
    close(errors)

    // Verify one succeeded and one got conflict
    successCount := 0
    conflictCount := 0
    for err := range errors {
        if err == nil {
            successCount++
        } else if pkgErrors.IsOptimisticLockError(err) {
            conflictCount++
        }
    }

    assert.Equal(t, 1, successCount)
    assert.Equal(t, 1, conflictCount)
}
```

## Performance Considerations

### Database Impact

1. **Additional Index**: The `(id, version)` composite index adds minimal overhead
2. **Version Check**: Single additional WHERE clause - negligible performance impact
3. **Atomic Operation**: Database handles concurrency through row-level locking

### Optimization Tips

1. **Read Latest Before Update**: Always fetch latest version before attempting update
2. **Batch Operations**: For bulk updates, consider application-level coordination
3. **Retry Strategy**: Use exponential backoff to avoid thundering herd
4. **Monitor Conflicts**: Track conflict rate to identify hotspots

## Migration Path

### For Existing Data

The migration script sets default version to 1 for all existing records:

```sql
ALTER TABLE organizations ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
```

### Backward Compatibility

- Existing `Update()` methods remain unchanged
- New `UpdateWithVersion()` methods added alongside
- Services can gradually adopt optimistic locking
- No breaking changes to existing API contracts

## Best Practices

### DO ✓

- Always include `version` in update request payload
- Return updated `version` in response
- Handle 409 Conflict with retry logic
- Use exponential backoff for retries
- Log version conflicts for monitoring
- Include current version in conflict error response

### DON'T ✗

- Don't ignore version conflicts
- Don't retry infinitely
- Don't update without version check in critical operations
- Don't expose raw database errors to clients
- Don't use optimistic locking for read-heavy operations
- Don't mix optimistic and pessimistic locking

## Monitoring and Observability

### Metrics to Track

```go
// Prometheus metrics
var (
    optimisticLockConflicts = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "optimistic_lock_conflicts_total",
            Help: "Total number of optimistic lock conflicts",
        },
        []string{"resource_type"},
    )

    optimisticLockRetries = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "optimistic_lock_retries",
            Help: "Number of retries before success",
            Buckets: prometheus.LinearBuckets(0, 1, 5),
        },
        []string{"resource_type"},
    )
)
```

### Logging

```go
logger.Info("Optimistic lock conflict detected",
    zap.String("resource_type", "organization"),
    zap.String("resource_id", orgID),
    zap.Int("expected_version", expectedVersion),
    zap.Int("current_version", currentVersion),
    zap.String("user_id", userID),
)
```

## Troubleshooting

### High Conflict Rate

**Symptoms**: Many 409 responses, slow update performance

**Causes**:
- Multiple services updating same resource
- Chatty client behavior (fetch-update loops)
- Missing retry logic causing user retries

**Solutions**:
- Implement proper retry with backoff
- Batch updates where possible
- Consider pessimistic locking for high-contention resources

### Version Drift

**Symptoms**: Version numbers grow very large

**Causes**: Normal behavior under heavy update load

**Solutions**:
- This is expected and not a problem
- Version counter can handle very large numbers (INTEGER type)
- No action needed

### Deadlocks

**Symptoms**: Updates hang or timeout

**Causes**: Mixing optimistic and pessimistic locking

**Solutions**:
- Use consistent locking strategy
- Ensure transactions complete quickly
- Avoid long-running transactions with version checks

## References

- [Optimistic Locking Pattern](https://en.wikipedia.org/wiki/Optimistic_concurrency_control)
- [PostgreSQL Row-Level Locking](https://www.postgresql.org/docs/current/explicit-locking.html)
- [GORM Updates](https://gorm.io/docs/update.html)
- Martin Fowler - [Optimistic Offline Lock](https://martinfowler.com/eaaCatalog/optimisticOfflineLock.html)
