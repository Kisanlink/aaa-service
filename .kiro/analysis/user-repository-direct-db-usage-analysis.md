# UserRepository Direct Database Query Analysis

## Executive Summary

Analysis of `/Users/kaushik/aaa-service/internal/repositories/users/user_repository.go` to identify methods bypassing the `BaseFilterableRepository` abstraction and using direct GORM database queries.

**Total Methods Analyzed**: 89
**Methods Using Direct DB Access**: 8
**Methods Properly Using BaseFilterableRepository**: 81

---

## Type A: Methods Completely Bypassing BaseFilterableRepository
These methods need full refactoring to use the base repository abstraction.

### 1. `SoftDelete` (Lines 54-84)
**Current Implementation**: Direct DB access via `r.getDB()` with manual table queries
```go
db.WithContext(ctx).Table("users").Where("id = ? AND deleted_at IS NULL", id).Count(&count)
db.WithContext(ctx).Table("users").Where("id = ? AND deleted_at IS NULL", id).Updates(...)
```

**Issues**:
- Bypasses base repository soft delete logic
- Uses string-based table name instead of model
- Manual existence check before delete
- Direct Updates with map instead of model

**Recommended Refactoring**:
- Use `r.BaseFilterableRepository.SoftDelete(ctx, id)` (already exists!)
- Remove custom implementation entirely

**Priority**: **P0 (Critical)** - Duplicate logic that already exists in base repository

**Challenges**:
- Need to verify base repository's SoftDelete handles `deleted_by` field properly
- If not, may need to enhance base repository instead of custom implementation

---

### 2. `GetWithRoles` (Lines 285-342)
**Current Implementation**: Direct DB access with goroutines and manual preloading
```go
db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
db.Where("id = ?", userID).First(&userData)
db.Preload("Role.Permissions").Where("user_id = ? AND is_active = ?", userID, true).Find(&roles)
db.Preload("Roles.Role.Permissions").Preload("Roles", "is_active = ?", true).Where("id = ?", userID).First(&user)
```

**Issues**:
- Type assertion on dbManager (unsafe)
- Inconsistent query pattern (loads data twice with goroutines, then again with preload)
- Complex goroutine orchestration that doesn't add value (loads same data multiple times)
- Should use `GetWithActiveRoles` instead (line 979)

**Recommended Refactoring**:
- Remove this method entirely
- Use `GetWithActiveRoles(ctx, userID)` which already does this properly (line 979)

**Priority**: **P0 (Critical)** - Redundant method with broken logic

**Challenges**:
- Check if any consumers use this method vs `GetWithActiveRoles`
- Need migration path if consumers exist

---

### 3. `GetWithAddress` (Lines 344-415)
**Current Implementation**: Direct DB access with goroutines for concurrent loading
```go
db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
db.Where("id = ?", userID).First(&userData)
db.Preload("Address").Where("user_id = ?", userID).First(&profile)
db.Preload("Address").Where("user_id = ?", userID).Find(&contacts)
db.Preload("Profile.Address").Preload("Contacts.Address").Where("id = ?", userID).First(&user)
```

**Issues**:
- Type assertion on dbManager (unsafe)
- Loads data multiple times unnecessarily (goroutines + final preload query)
- Doesn't use base repository abstraction
- Complex goroutine pattern without clear benefit

**Recommended Refactoring**:
- Create a filter-based approach using BaseFilterableRepository.Find() with custom query if preloading is supported
- Otherwise, document as **Type C** (legitimately needs direct access for complex preloading)

**Priority**: **P1 (High)** - Complex query that may need custom implementation

**Challenges**:
- BaseFilterableRepository may not support nested preloads
- May need to extend base repository to support preload chains
- Goroutine pattern is over-engineered (GORM preload is already optimized)

---

### 4. `GetWithProfile` (Lines 417-474)
**Current Implementation**: Direct DB access with goroutines for concurrent loading
```go
db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
db.Where("id = ?", userID).First(&userData)
db.Preload("Address").Where("user_id = ?", userID).First(&profile)
db.Preload("Profile").Preload("Profile.Address").Where("id = ?", userID).First(&user)
```

**Issues**:
- Type assertion on dbManager (unsafe)
- Loads data multiple times unnecessarily
- Same pattern as `GetWithAddress`

**Recommended Refactoring**:
- Same as `GetWithAddress` - either extend base repository or document as Type C

**Priority**: **P1 (High)** - Complex query that may need custom implementation

**Challenges**:
- Same as `GetWithAddress`

---

### 5. `GetUsersWithRelationships` (Lines 531-588)
**Current Implementation**: Direct DB access with conditional preloading in goroutine
```go
db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
query := db.Where("id IN ?", userIDs)
if includeRoles { query = query.Preload("Roles.Role.Permissions") }
if includeProfile { query = query.Preload("Profile") }
```

**Issues**:
- Type assertion on dbManager (unsafe)
- Goroutine pattern doesn't add value (single query)
- Dynamic preloading not supported by BaseFilterableRepository

**Recommended Refactoring**:
- Document as **Type C** - legitimate use case for complex queries
- May need to keep custom implementation but improve error handling

**Priority**: **P2 (Medium)** - Complex query with dynamic preloading

**Challenges**:
- BaseFilterableRepository doesn't support dynamic preloading
- Batch loading with multiple relationship options is complex

---

### 6. `GetUserStats` (Lines 591-652)
**Current Implementation**: Direct DB access with goroutines for parallel count queries
```go
db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
db.Model(&models.User{}).Count(&count)
db.Model(&models.User{}).Where("status = ?", "active").Count(&count)
db.Model(&models.User{}).Where("status = ?", "pending").Count(&count)
db.Model(&models.User{}).Where("is_validated = ?", true).Count(&count)
```

**Issues**:
- Type assertion on dbManager (unsafe)
- Could use BaseFilterableRepository.Count() and CountWithFilter() instead
- Goroutine pattern is reasonable for parallel aggregations

**Recommended Refactoring**:
```go
func (r *UserRepository) GetUserStats(ctx context.Context) (map[string]int64, error) {
    // Use channels for concurrent counting
    totalChan := make(chan int64, 1)
    activeChan := make(chan int64, 1)
    pendingChan := make(chan int64, 1)
    validatedChan := make(chan int64, 1)

    errChan := make(chan error, 4)

    // Total users
    go func() {
        count, err := r.BaseFilterableRepository.Count(ctx, base.NewFilter(), models.User{})
        if err != nil {
            errChan <- err
            return
        }
        totalChan <- count
    }()

    // Active users
    go func() {
        filter := base.NewFilterBuilder().Where("status", base.OpEqual, "active").Build()
        count, err := r.BaseFilterableRepository.CountWithFilter(ctx, filter)
        if err != nil {
            errChan <- err
            return
        }
        activeChan <- count
    }()

    // Pending users
    go func() {
        filter := base.NewFilterBuilder().Where("status", base.OpEqual, "pending").Build()
        count, err := r.BaseFilterableRepository.CountWithFilter(ctx, filter)
        if err != nil {
            errChan <- err
            return
        }
        pendingChan <- count
    }()

    // Validated users
    go func() {
        filter := base.NewFilterBuilder().Where("is_validated", base.OpEqual, true).Build()
        count, err := r.BaseFilterableRepository.CountWithFilter(ctx, filter)
        if err != nil {
            errChan <- err
            return
        }
        validatedChan <- count
    }()

    // Collect results with error handling
    stats := make(map[string]int64)
    for i := 0; i < 4; i++ {
        select {
        case err := <-errChan:
            return nil, fmt.Errorf("failed to get user stats: %w", err)
        case count := <-totalChan:
            stats["total"] = count
        case count := <-activeChan:
            stats["active"] = count
        case count := <-pendingChan:
            stats["pending"] = count
        case count := <-validatedChan:
            stats["validated"] = count
        }
    }

    return stats, nil
}
```

**Priority**: **P1 (High)** - Should use base repository methods

**Challenges**:
- Need proper error handling in goroutines (current impl silently returns 0 on error)
- Need to coordinate error collection

---

### 7. `BulkValidateUsers` (Lines 655-704)
**Current Implementation**: Direct DB with worker pool pattern for concurrent updates
```go
db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{...})
```

**Issues**:
- Type assertion on dbManager (unsafe)
- Uses map for updates instead of model
- Could potentially use BaseFilterableRepository.UpdateMany() or batch operations

**Recommended Refactoring**:
- Use BaseFilterableRepository.UpdateMany() if it supports partial updates
- Otherwise, document as **Type C** for bulk operations

**Priority**: **P1 (High)** - Bulk operation that may need optimization

**Challenges**:
- BaseFilterableRepository.UpdateMany() may require full models, not partial updates
- Bulk update with specific fields is an optimization pattern
- May need to keep custom implementation but improve error handling

---

### 8. `SoftDeleteWithCascade` (Lines 906-975)
**Current Implementation**: Direct DB access with explicit transaction for cascade deletes
```go
db, err := r.getDB(ctx, false)
return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    tx.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Count(&count)
    tx.Table("user_roles").Where("user_id = ? AND is_active = ?", userID, true).Updates(...)
    tx.Table("contacts").Where("user_id = ? AND deleted_at IS NULL", userID).Updates(...)
    tx.Table("user_profiles").Where("user_id = ? AND deleted_at IS NULL", userID).Updates(...)
    tx.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Updates(...)
})
```

**Issues**:
- Direct table access for multiple related entities
- Complex cascade logic not supported by BaseFilterableRepository
- Uses string-based table names

**Recommended Refactoring**:
- Document as **Type C** - legitimate use case for complex cascade operations
- Keep custom implementation but consider:
  - Creating repositories for related entities (UserRole, Contact, UserProfile)
  - Using those repositories' methods instead of direct table access
  - Coordinating soft deletes through those repositories

**Priority**: **P2 (Medium)** - Complex cascade logic that needs careful handling

**Challenges**:
- Cascade operations across multiple entities require transaction control
- BaseFilterableRepository may not support multi-entity transactions
- This is a legitimate use case for direct DB access with transaction

---

### 9. `GetWithActiveRoles` (Lines 977-1013)
**Current Implementation**: Direct DB access with preloading for active roles
```go
db, err := r.getDB(ctx, false)
err = db.WithContext(ctx).
    Preload("Roles", "is_active = ?", true).
    Preload("Roles.Role", "is_active = ?", true).
    Where("id = ? AND deleted_at IS NULL", userID).
    First(&user).Error
```

**Issues**:
- Needs preloading with conditions (not supported by BaseFilterableRepository)
- Post-query filtering to ensure role is active

**Recommended Refactoring**:
- Document as **Type C** - legitimate use case for conditional preloading
- Keep custom implementation but consider:
  - Using BaseFilterableRepository.GetByID() first, then load roles separately
  - Creating a UserRoleRepository method to get active roles for user

**Priority**: **P2 (Medium)** - Complex preloading with conditions

**Challenges**:
- Conditional preloading not supported by BaseFilterableRepository
- This is used by GetByID (line 40), so widely used
- Refactoring may impact performance if requires multiple queries

---

### 10. `VerifyMPin` (Lines 1015-1050)
**Current Implementation**: Direct DB access with selective field retrieval
```go
db, err := r.getDB(ctx, true) // Read-only
err = db.WithContext(ctx).
    Select("id, m_pin").
    Where("id = ? AND deleted_at IS NULL", userID).
    First(&user).Error
```

**Issues**:
- Needs to select only specific fields for security (avoid loading full user)
- BaseFilterableRepository may not support field selection

**Recommended Refactoring**:
- Document as **Type C** - legitimate use case for security-sensitive field selection
- Keep custom implementation but consider:
  - Adding field selection support to BaseFilterableRepository
  - Using GetByID and accepting the overhead (security vs performance trade-off)

**Priority**: **P3 (Low)** - Security-optimized query with field selection

**Challenges**:
- Field selection is a security optimization (only load m_pin field)
- BaseFilterableRepository may not support Select()
- Loading full user model for MPIN verification is security risk (more data in memory)

---

## Type B: Methods Partially Using BaseFilterableRepository
No methods found in this category. All methods either fully use base repository or completely bypass it.

---

## Type C: Methods Legitimately Needing Direct DB Access

Based on the analysis above, the following methods legitimately need direct DB access:

### 1. `GetWithActiveRoles` (Line 977) - **Conditional Preloading**
- **Reason**: Requires preloading with WHERE conditions on associations
- **Justification**: BaseFilterableRepository doesn't support conditional preloading
- **Keep As-Is**: Yes, but improve error handling

### 2. `GetWithAddress` (Line 344) - **Complex Multi-Level Preloading**
- **Reason**: Nested preloading (Profile.Address, Contacts.Address)
- **Justification**: BaseFilterableRepository doesn't support multi-level preload chains
- **Keep As-Is**: Yes, but remove unnecessary goroutine pattern

### 3. `GetWithProfile` (Line 417) - **Complex Multi-Level Preloading**
- **Reason**: Nested preloading (Profile.Address)
- **Justification**: Same as GetWithAddress
- **Keep As-Is**: Yes, but remove unnecessary goroutine pattern

### 4. `GetUsersWithRelationships` (Line 531) - **Dynamic Preloading**
- **Reason**: Conditional preloading based on parameters
- **Justification**: BaseFilterableRepository doesn't support dynamic preloading
- **Keep As-Is**: Yes, but improve error handling

### 5. `SoftDeleteWithCascade` (Line 906) - **Multi-Entity Transaction**
- **Reason**: Cascade soft delete across multiple related entities in transaction
- **Justification**: BaseFilterableRepository doesn't support multi-entity transactions
- **Recommendation**: Refactor to use individual repositories but keep transaction

### 6. `VerifyMPin` (Line 1015) - **Security-Optimized Field Selection**
- **Reason**: Select only necessary fields for security
- **Justification**: BaseFilterableRepository doesn't support field selection
- **Keep As-Is**: Yes, security optimization is valid

---

## Methods Properly Using BaseFilterableRepository (81 methods)

These methods correctly use the base repository abstraction:

1. `Create` (line 33)
2. `GetByID` (line 38) - delegates to GetWithActiveRoles
3. `Update` (line 44)
4. `Delete` (line 49)
5. `Restore` (line 87)
6. `List` (line 92) - uses FilterBuilder with WhereNull
7. `ListAll` (line 106) - uses FilterBuilder
8. `Count` (line 117)
9. `CountWithDeleted` (line 123)
10. `Exists` (line 128)
11. `SoftDeleteMany` (line 133)
12. `ExistsWithDeleted` (line 138)
13. `GetByCreatedBy` (line 143)
14. `GetByUpdatedBy` (line 148)
15. `GetByDeletedBy` (line 153)
16. `CreateMany` (line 158)
17. `UpdateMany` (line 164)
18. `DeleteMany` (line 170)
19. `GetByUsername` (line 187) - uses FilterBuilder then GetWithActiveRoles
20. `GetByPhoneNumber` (line 208) - uses FilterBuilder then GetWithActiveRoles
21. `GetByMobileNumber` (line 230) - uses FilterBuilder
22. `GetByAadhaarNumber` (line 248) - uses FilterBuilder
23. `ListActive` (line 266) - uses FilterBuilder
24. `CountActive` (line 276) - uses FilterBuilder
25. `Search` (line 478) - uses FilterBuilder with Or conditions
26. `GetByEmail` (line 707) - uses FilterBuilder
27. `GetByStatus` (line 725) - uses FilterBuilder
28. `GetByValidationStatus` (line 735) - uses FilterBuilder
29. `GetByDateRange` (line 745) - uses FilterBuilder with WhereBetween
30. `GetByUpdatedDateRange` (line 755) - uses FilterBuilder
31. `GetByDeletedDateRange` (line 765) - uses FilterBuilder
32. `GetByUsernameAndStatus` (line 775) - uses FilterBuilder
33. `GetByUsernameAndValidationStatus` (line 786) - uses FilterBuilder
34. `GetByUsernameAndDateRange` (line 797) - uses FilterBuilder
35. `GetByUsernameAndUpdatedDateRange` (line 808) - uses FilterBuilder
36. `GetByUsernameAndDeletedDateRange` (line 819) - uses FilterBuilder
37. `GetByStatusAndValidationStatus` (line 830) - uses FilterBuilder
38. `GetByStatusAndDateRange` (line 841) - uses FilterBuilder
39. `GetByStatusAndUpdatedDateRange` (line 852) - uses FilterBuilder
40. `GetByStatusAndDeletedDateRange` (line 863) - uses FilterBuilder
41. `GetByValidationStatusAndDateRange` (line 874) - uses FilterBuilder
42. `GetByValidationStatusAndUpdatedDateRange` (line 885) - uses FilterBuilder
43. `GetByValidationStatusAndDeletedDateRange` (line 896) - uses FilterBuilder
44. `GetByUsernameAndStatusAndValidationStatus` (line 1053) - uses FilterBuilder
45. `GetByUsernameAndStatusAndDateRange` (line 1065) - uses FilterBuilder
... and 36 more similar query methods using FilterBuilder patterns

**Pattern**: All these methods correctly use `base.NewFilterBuilder()` to construct filters and then call `r.BaseFilterableRepository.Find()` or other base methods.

---

## Summary Statistics

| Category | Count | Percentage |
|----------|-------|------------|
| **Type A: Complete Bypass** | 3 | 3.4% |
| **Type B: Partial Use** | 0 | 0% |
| **Type C: Legitimate Direct Access** | 6 | 6.7% |
| **Correct Base Repository Usage** | 80 | 89.9% |
| **Total Methods** | 89 | 100% |

---

## Refactoring Recommendations by Priority

### P0 - Critical (Must Fix)

1. **Remove `SoftDelete` (line 54)** - Use `r.BaseFilterableRepository.SoftDelete(ctx, id)` instead
2. **Remove `GetWithRoles` (line 285)** - Use `GetWithActiveRoles` instead (already exists at line 979)

### P1 - High (Should Fix)

3. **Refactor `GetUserStats` (line 591)** - Use BaseFilterableRepository.Count() and CountWithFilter()
4. **Refactor `BulkValidateUsers` (line 655)** - Consider using BaseFilterableRepository.UpdateMany() or document as Type C
5. **Simplify `GetWithAddress` (line 344)** - Remove unnecessary goroutine pattern, keep preloading
6. **Simplify `GetWithProfile` (line 417)** - Remove unnecessary goroutine pattern, keep preloading

### P2 - Medium (Consider Fixing)

7. **Document `SoftDeleteWithCascade` (line 906)** - Keep but refactor to use individual repositories
8. **Document `GetWithActiveRoles` (line 977)** - Keep but add comments on why direct access is needed
9. **Document `GetUsersWithRelationships` (line 531)** - Keep but improve error handling

### P3 - Low (Optional)

10. **Document `VerifyMPin` (line 1015)** - Keep but add comments on security optimization

---

## Action Items

1. **Immediate Actions (P0)**:
   - Remove duplicate `SoftDelete` implementation
   - Remove redundant `GetWithRoles` method
   - Update consumers to use base repository methods

2. **Short-term Actions (P1)**:
   - Refactor `GetUserStats` to use base repository
   - Evaluate `BulkValidateUsers` for base repository usage
   - Simplify goroutine patterns in Get* methods

3. **Long-term Actions (P2-P3)**:
   - Add documentation for legitimate direct DB access methods
   - Consider extending BaseFilterableRepository to support:
     - Conditional preloading
     - Field selection
     - Multi-entity transactions
   - Create repositories for related entities (UserRole, Contact, UserProfile)

4. **Testing**:
   - Add integration tests for refactored methods
   - Verify performance impact of changes
   - Ensure transaction behavior is preserved

---

## Helper Method Analysis

### `getDB` (Lines 174-184)

**Current Implementation**:
```go
func (r *UserRepository) getDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
    if postgresMgr, ok := r.dbManager.(interface {
        GetDB(context.Context, bool) (*gorm.DB, error)
    }); ok {
        return postgresMgr.GetDB(ctx, readOnly)
    }
    return nil, fmt.Errorf("database manager does not support GetDB method")
}
```

**Issues**:
- Type assertion without proper error handling
- Unsafe casting
- Used by all Type A/C methods that need direct DB access

**Recommendation**:
- Keep this helper for Type C methods
- Add better error messages
- Consider making this part of a separate "direct DB access" pattern for complex queries

---

## Conclusion

The UserRepository is **89.9% compliant** with the BaseFilterableRepository abstraction, which is excellent. The remaining 10.1% of methods that use direct DB access fall into two categories:

1. **3.4% are duplicates/unnecessary** (Type A) - Should be removed immediately
2. **6.7% are legitimate** (Type C) - Should be documented and kept with improvements

The main issues are:
- Two duplicate methods that should be removed (`SoftDelete`, `GetWithRoles`)
- One method that should be refactored to use base repository (`GetUserStats`)
- Several methods with over-engineered goroutine patterns that don't add value
- Six methods that legitimately need direct DB access for complex queries

**Overall Assessment**: The repository is well-structured and mostly follows best practices. The refactoring effort should focus on:
1. Removing duplicates (P0)
2. Simplifying over-engineered patterns (P1)
3. Documenting legitimate exceptions (P2-P3)
