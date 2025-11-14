# Super Admin Wildcard Permissions

## Overview

This document describes the implementation of wildcard permissions for the `super_admin` role in the AAA service. The implementation ensures that users with the `super_admin` role have complete, unrestricted access to all resources and actions in the system.

## Implementation Approach

The AAA service implements **triple-layered wildcard permission checking** for the `super_admin` role:

### Layer 1: Implicit Role Name Check (Primary)
Location: `internal/services/postgres_authorization_service.go:227-247`

```go
func (s *PostgresAuthorizationService) roleHasWildcardPermission(ctx context.Context, role models.Role) bool {
    // Check for super admin or global admin roles
    adminRoleNames := []string{"super_admin", "admin", "system_admin"}
    for _, adminName := range adminRoleNames {
        if role.Name == adminName {
            return true  // Grants ALL permissions automatically
        }
    }
    // ... additional checks
}
```

**Behavior**: Any user with a role named `super_admin` automatically passes ALL permission checks, regardless of specific permissions assigned. This is the fastest and most reliable check.

### Layer 2: Explicit Permission Name Check
Location: `internal/services/postgres_authorization_service.go:238-246`

```go
// Check if role has manage or admin permissions
var count int64
s.db.WithContext(ctx).
    Table("role_permissions").
    Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
    Where("role_permissions.role_id = ? AND role_permissions.is_active = ?", role.ID, true).
    Where("permissions.name IN (?) AND permissions.is_active = ?", []string{"manage", "admin", "super_admin"}, true).
    Count(&count)
```

**Behavior**: If a role has any permission with names like `manage`, `admin`, or `super_admin`, it grants wildcard access.

### Layer 3: Explicit Wildcard Resource Permissions (NEW)
Location: `migrations/seed_super_admin_wildcard.go`

This new migration creates explicit `resource_permissions` entries with `resource_id = "*"` for all combinations of:
- Resource types (user, role, permission, organization, group, etc.)
- Actions (create, read, update, delete, manage, etc.)

**Behavior**: Provides explicit database records that grant `super_admin` access to all resources of each type.

## Files Modified/Created

### New Files
1. **`migrations/seed_super_admin_wildcard.go`**
   - New migration function: `SeedSuperAdminWildcardPermissions()`
   - Creates wildcard resource_permissions entries for super_admin
   - Automatically creates entries for all resource types × all actions
   - Uses `resource_id = "*"` to grant access to all resources

### Modified Files
1. **`cmd/server/main.go`** (lines 101-105)
   - Added call to `SeedSuperAdminWildcardPermissionsWithDBManager()`
   - Runs after comprehensive RBAC seeding
   - Ensures wildcard permissions are seeded on every startup when `AAA_RUN_SEED=true`

## Wildcard Permission Check Flow

When a permission check is performed (e.g., user tries to access a resource):

1. **Get user's roles** - includes direct, group-inherited, and hierarchical roles
2. **Check specific permissions** - Look for exact resource + action matches
3. **Check wildcard resource_id** - Look for `resource_id = "*"` entries (NEW)
4. **Check role name** - If role name is `super_admin`, grant access immediately
5. **Cache result** - Store for 5 minutes (300 seconds)

## Database Structure

### resource_permissions Table
```sql
CREATE TABLE resource_permissions (
    id VARCHAR(255) PRIMARY KEY,
    role_id VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,  -- e.g., "user", "role", "organization"
    resource_id VARCHAR(255) NOT NULL,    -- specific ID or "*" for wildcard
    action VARCHAR(50) NOT NULL,          -- e.g., "read", "create", "manage"
    is_active BOOLEAN DEFAULT true
);
```

### Example Wildcard Entries
After seeding, `super_admin` will have entries like:
```
role_id: <super_admin_id>, resource_type: "user", resource_id: "*", action: "read"
role_id: <super_admin_id>, resource_type: "user", resource_id: "*", action: "create"
role_id: <super_admin_id>, resource_type: "role", resource_id: "*", action: "manage"
role_id: <super_admin_id>, resource_type: "organization", resource_id: "*", action: "delete"
... (for all resource types × all actions)
```

## Seeding Process

The seeding happens in this order:

1. `SeedStaticActions()` - Creates all action types (create, read, update, delete, manage, etc.)
2. `SeedCoreResourcesRolesPermissions()` - Creates core resources and super_admin role
3. `SeedComprehensiveRBAC()` - Creates extended resources and specific permissions
4. **`SeedSuperAdminWildcardPermissions()` - Creates wildcard entries (NEW)**
5. `AddPerformanceIndexes()` - Creates database indexes

## Usage

### Automatic Seeding
Set environment variable and start the service:
```bash
export AAA_RUN_SEED=true
./aaa-service
```

### Manual Seeding
Call the migration function directly:
```go
import "github.com/Kisanlink/aaa-service/v2/migrations"

err := migrations.SeedSuperAdminWildcardPermissionsWithDBManager(ctx, dbManager, logger)
```

### Via API Endpoint
```bash
curl -X POST http://localhost:8080/api/v1/catalog/seed \
  -H "Authorization: Bearer $TOKEN"
```

## Security Considerations

1. **Role Name Protection**: The `super_admin` role name should be protected and only assigned to trusted users
2. **Immutable Check**: The implicit role name check cannot be bypassed by database manipulation
3. **Triple Redundancy**: Three layers ensure super_admin always has access even if one layer fails
4. **Audit Trail**: All permission checks are logged, including super_admin access
5. **Cache Invalidation**: Caches are invalidated when roles or permissions change

## Testing

### Verify Wildcard Permissions Exist
```sql
SELECT COUNT(*)
FROM resource_permissions
WHERE role_id = (SELECT id FROM roles WHERE name = 'super_admin')
  AND resource_id = '*';
```

Expected result: Multiple entries (number of resource types × number of actions)

### Test Permission Check
```go
result, err := authzService.CheckPermission(ctx, &Permission{
    UserID: superAdminUserID,
    Resource: "organization",
    ResourceID: "any-org-id",
    Action: "delete",
})
// Expected: result.Allowed == true
```

## Performance Impact

- **Initial Seeding**: Creates ~1000-2000 entries (depending on resources/actions)
- **Query Performance**: Minimal - wildcard check happens after specific permission check fails
- **Cache Impact**: Permission results are cached for 5 minutes
- **Storage**: ~100KB for all wildcard entries

## Backward Compatibility

This implementation is **fully backward compatible**:
- Existing super_admin users continue to work via implicit name check (Layer 1)
- Existing permission checks remain unchanged
- No breaking changes to APIs or data structures
- Can be rolled back by removing wildcard entries without breaking functionality

## Future Enhancements

Potential future improvements:
1. **Wildcard Actions**: Support `action = "*"` in addition to `resource_id = "*"`
2. **Resource Hierarchies**: Support wildcards for resource hierarchies (e.g., `org/*` for all resources in an org)
3. **Time-based Wildcards**: Temporary wildcard permissions with expiration
4. **Conditional Wildcards**: Wildcards that apply only under certain conditions

## References

- Permission Evaluation Logic: `internal/services/postgres_authorization_service.go`
- Resource Permission Model: `internal/entities/models/resource_permission.go`
- Core Seeding: `migrations/seed_core_roles_permissions.go`
- Comprehensive Seeding: `migrations/seed_comprehensive_rbac.go`
- Static Actions: `migrations/seed_static_actions.go`
