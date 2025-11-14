# Role Hierarchy and Assignment History Features

**Version:** v2.1.5
**Date:** 2025-11-14

## Overview

This document describes the implementation of three major features added to the AAA service:
1. Role Hierarchy Support
2. Group Role Retrieval in Organizations
3. Permission Assignment History Tracking

---

## 1. Role Hierarchy Support

### Description
Roles can now form parent-child hierarchical relationships, enabling role inheritance and organizational role structures.

### Implementation Location
- **Service:** `internal/services/role_hierarchy.go`
- **Repository:** `internal/repositories/roles/role_repository.go`
- **Handler:** `internal/handlers/roles/role_handler.go`

### Key Methods

#### `GetRoleHierarchy(ctx context.Context) ([]*models.Role, error)`
Retrieves all roles organized in a hierarchical tree structure.

**Algorithm:**
1. Fetches all roles from the database
2. First pass: Creates a map and identifies root roles (roles without parents)
3. Second pass: Attaches children to their respective parents
4. Returns array of root roles with nested children

**Returns:** Array of root role nodes with their children recursively attached

#### `AddChildRole(ctx context.Context, parentRoleID, childRoleID string) error`
Establishes a parent-child relationship between two roles.

**Validations:**
- Both role IDs must be non-empty
- Roles cannot be their own parent
- Child role must not already have a parent
- Circular dependency check prevents infinite loops

**Side Effects:**
- Updates child role's `ParentID` field
- Clears cache for both parent and child roles

#### `RemoveChildRole(ctx context.Context, parentRoleID, childRoleID string) error`
Removes a parent-child relationship.

**Validations:**
- Verifies the child role actually belongs to the specified parent
- Child role must exist

**Side Effects:**
- Sets child role's `ParentID` to null
- Clears cache for both roles

#### `GetRoleWithChildren(ctx context.Context, roleID string) (*models.Role, error)`
Retrieves a specific role with all its descendants recursively loaded.

### Circular Dependency Prevention

The `checkCircularDependency()` helper method walks up the parent chain from the proposed parent role. If it encounters the child role ID at any point in the chain, it rejects the operation to prevent circular references.

### HTTP Endpoints

#### `GET /api/v2/roles/hierarchy`
Returns the complete role hierarchy tree.

**Response:**
```json
{
  "hierarchy": [
    {
      "id": "role_id",
      "name": "Admin",
      "description": "Administrator role",
      "scope": "organization",
      "is_active": true,
      "children": [
        {
          "id": "child_role_id",
          "name": "Manager",
          "parent_id": "role_id",
          "children": []
        }
      ]
    }
  ],
  "count": 1
}
```

#### `POST /api/v2/roles/:id/children`
Adds a child role to a parent role.

**Request Body:**
```json
{
  "child_role_id": "child_role_id"
}
```

**Response:**
```json
{
  "message": "Child role added successfully",
  "parent_role_id": "parent_id",
  "child_role_id": "child_id"
}
```

### Database Changes

**New Repository Methods:**
- `GetChildRoles(ctx, parentRoleID)` - Returns direct children of a role
- `GetAll(ctx)` - Returns all roles (excluding soft-deleted)

---

## 2. Group Role Retrieval in Organizations

### Description
Enables retrieval of all roles assigned to a specific group within an organization context, with proper validation and organization scoping.

### Implementation Location
- **Service:** `internal/services/organizations/organization_service.go:1020-1092`
- **Repository:** `internal/repositories/groups/group_role_repository.go:110-120`

### Key Method

#### `GetGroupRolesInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error)`

**Validation:**
1. Organization ID and Group ID are required
2. Organization must exist
3. Group must exist
4. Group must belong to the specified organization

**Process:**
1. Validates input parameters and sets default pagination (limit: 50)
2. Verifies organization existence
3. Verifies group exists and validates it belongs to the organization
4. Delegates to `groupService.GetGroupRoles()` for actual role retrieval
5. Wraps response with organization context

**Returns:**
```json
{
  "organization_id": "org_id",
  "group_id": "group_id",
  "roles": [
    {
      "id": "group_role_id",
      "group_id": "group_id",
      "role_id": "role_id",
      "role": {
        "id": "role_id",
        "name": "Role Name",
        "description": "Role Description",
        "is_active": true
      },
      "assigned_by": "user_id",
      "is_active": true,
      "created_at": "2025-11-14T..."
    }
  ]
}
```

### HTTP Endpoint

#### `GET /organizations/:orgId/groups/:groupId/roles`
Retrieves all roles assigned to a group within an organization.

**Query Parameters:**
- `limit` (optional, default: 50) - Number of results to return
- `offset` (optional, default: 0) - Number of results to skip

---

## 3. Permission Assignment History Tracking

### Description
Tracks the complete history of permission assignments and revocations for roles by querying audit logs.

### Implementation Location
- **Service:** `internal/services/role_assignments/query.go:172-241`
- **Repository:** `internal/repositories/audit/audit_repository.go:363-383`

### Key Method

#### `GetPermissionAssignmentHistory(ctx context.Context, roleID string, limit, offset int) ([]*AssignmentHistory, error)`

**Data Source:** Audit logs with actions:
- `grant_permission` - Permission assignments
- `revoke_permission` - Permission revocations

**Process:**
1. Validates role ID is provided
2. Sets default pagination (limit: 50)
3. Queries audit logs filtered by:
   - Resource Type: `role`
   - Resource ID: `roleID`
   - Actions: `grant_permission`, `revoke_permission`
4. Sorts results by timestamp (descending)
5. Converts audit logs to `AssignmentHistory` format
6. Extracts metadata:
   - Permission ID from log details
   - Assigned by from user ID or details
   - Revocation timestamp and user for revoke actions

**Returns:** Array of `AssignmentHistory` objects:
```go
type AssignmentHistory struct {
    RoleID         string
    PermissionID   string
    AssignedAt     time.Time
    AssignedBy     string
    RevokedAt      *time.Time
    RevokedBy      *string
    AssignmentType string // "permission"
}
```

### New Repository Method

#### `ListByResourceAndActions(ctx, resourceType, resourceID string, actions []string, limit, offset int) ([]*models.AuditLog, error)`

Generic audit log query method that:
- Filters by resource type and ID
- Filters by multiple action types (using IN operator)
- Supports pagination
- Sorts by timestamp descending

**Use Cases:**
- Permission assignment history
- Role change history
- Group membership history
- Any resource-specific action tracking

---

## Database Schema Updates

### Role Model
The `Role` model already had the necessary fields:
```go
type Role struct {
    ParentID *string `json:"parent_id"` // For role hierarchy
    Children []Role  `json:"children"`  // Nested children
}
```

### Audit Log Usage
Leverages existing `audit_logs` table with:
- `resource_type` - Type of resource (e.g., "role")
- `resource_id` - ID of the resource
- `action` - Action performed (e.g., "grant_permission", "revoke_permission")
- `user_id` - User who performed the action
- `details` - JSONB field containing additional metadata
- `timestamp` - When the action occurred

---

## Interface Updates

### RoleService Interface
Added methods:
```go
GetRoleHierarchy(ctx context.Context) ([]*models.Role, error)
AddChildRole(ctx context.Context, parentRoleID, childRoleID string) error
RemoveChildRole(ctx context.Context, parentRoleID, childRoleID string) error
GetRoleWithChildren(ctx context.Context, roleID string) (*models.Role, error)
```

### RoleRepository Interface
Added methods:
```go
GetChildRoles(ctx context.Context, parentRoleID string) ([]*models.Role, error)
GetAll(ctx context.Context) ([]*models.Role, error)
```

### GroupRoleRepository Interface
Added method:
```go
GetByOrganizationAndGroupID(ctx context.Context, organizationID, groupID string, limit, offset int) ([]*models.GroupRole, error)
```

### AuditRepository Interface
Added method:
```go
ListByResourceAndActions(ctx context.Context, resourceType, resourceID string, actions []string, limit, offset int) ([]*models.AuditLog, error)
```

---

## Service Dependencies

### Role Assignment Service
Updated constructor to include audit repository:
```go
func NewService(
    roleRepo *roles.RoleRepository,
    rolePermissionRepo *role_permissions.RolePermissionRepository,
    resourcePermissionRepo *resource_permissions.ResourcePermissionRepository,
    permissionRepo *permissions.PermissionRepository,
    auditRepo interfaces.AuditRepository, // NEW
    cache interfaces.CacheService,
    audit interfaces.AuditService,
    logger interfaces.Logger,
) *Service
```

---

## Caching Strategy

### Role Hierarchy
- Cache cleared on `AddChildRole` and `RemoveChildRole`
- Cache keys: `role:{roleID}`
- Invalidation affects both parent and child roles

### Group Roles
- Uses existing group service caching mechanism
- Cached by group ID and query parameters

---

## Error Handling

All methods implement comprehensive error handling:
- Input validation errors
- Database operation errors
- Cache operation errors (logged but non-blocking)
- Not found errors for missing entities
- Business logic validation errors (e.g., circular dependencies)

---

## Logging

All operations include structured logging with:
- Operation name
- Relevant IDs (roleID, groupID, organizationID)
- Success/failure status
- Result counts
- Error details when applicable

---

## Testing Recommendations

### Unit Tests
1. Test role hierarchy circular dependency detection
2. Test role hierarchy tree building algorithm
3. Test audit log to history conversion
4. Test organization-group validation logic

### Integration Tests
1. Test complete role hierarchy creation and retrieval
2. Test group role retrieval across organization boundaries
3. Test assignment history with actual audit log creation

### Edge Cases
1. Empty role hierarchies
2. Deep role hierarchies (performance)
3. Roles with multiple children
4. Groups with no roles
5. Empty assignment history

---

## Performance Considerations

### Role Hierarchy
- Two-pass algorithm is O(n) where n is number of roles
- In-memory tree building - efficient for moderate role counts
- Consider pagination for very large role sets

### Assignment History
- Relies on audit log indexes on `resource_type`, `resource_id`, `action`
- Pagination supported for large histories
- Sorted by timestamp for chronological viewing

### Group Roles
- Delegates to existing cached group service
- Validation queries are single-record lookups (fast)

---

## Migration Notes

No database migrations required - all changes are backward compatible:
- Role hierarchy uses existing `parent_id` and `children` fields
- Group role retrieval uses existing tables and relationships
- Assignment history uses existing audit log infrastructure

---

## Future Enhancements

### Role Hierarchy
- [ ] Maximum hierarchy depth limit
- [ ] Bulk role hierarchy operations
- [ ] Role hierarchy visualization endpoint
- [ ] Inherited permission resolution

### Group Role Retrieval
- [ ] Filter by role scope or status
- [ ] Search roles by name within group
- [ ] Aggregate role statistics per group

### Assignment History
- [ ] Filter by date range
- [ ] Filter by assigner
- [ ] Export history to CSV/JSON
- [ ] Resource action assignment history (not just permissions)
- [ ] User role assignment history

---

## References

- Audit Log Model: `internal/entities/models/audit_log.go`
- Role Model: `internal/entities/models/role.go`
- Group Role Model: `internal/entities/models/group_role.go`
- Role Service Tests: `internal/services/role_service_test.go`
- Group Service Tests: `internal/services/groups/group_service_role_test.go`
