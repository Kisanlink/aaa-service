# Role Inheritance - Quick Reference

## The Gap

Users get **direct roles ONLY** in their JWT tokens. Group role inheritance engine exists but isn't used during token generation.

## Key Files

### 1. Current Token Flow (Direct Roles Only)
```
auth_handler.go:LoginV2()
  ↓
user_service:VerifyUserCredentials()
  ↓
user_service:GetUserWithRoles()
  ↓
user_role_repo:GetActiveRolesByUserID() ← ONLY DIRECT ROLES
  ↓
GenerateAccessTokenWithContext()
  ↓
Token with direct roles only
```

### 2. Unused Role Inheritance Engine
```
role_inheritance_engine.go
  - CalculateEffectiveRoles() ← PERFECT BUT NOT CALLED DURING AUTH
  - Distance-based precedence
  - 5-minute cache
  - ~650 lines of production-ready code
  - 3 comprehensive test files
```

## File Locations Summary

| Component | File | Status |
|-----------|------|--------|
| **GroupRole Model** | `internal/entities/models/group_role.go` | Implemented ✓ |
| **GroupRole Repo** | `internal/repositories/groups/group_role_repository.go` | Implemented ✓ |
| **RoleInheritance Engine** | `internal/services/groups/role_inheritance_engine.go` | Implemented ✓ |
| **User Service** | `internal/services/user/additional_methods.go` | Missing integration ✗ |
| **Auth Handler** | `internal/handlers/auth/auth_handler.go` | Uses direct roles only ✗ |
| **User Interface** | `internal/interfaces/interfaces.go` | Missing effective roles method ✗ |

## What's Implemented

### Group Role Assignment (Line 123-126, interfaces.go)
```go
GroupService interface {
    AssignRoleToGroup(ctx context.Context, groupID, roleID, assignedBy string) (interface{}, error)
    RemoveRoleFromGroup(ctx context.Context, groupID, roleID string) error
    GetGroupRoles(ctx context.Context, groupID string) (interface{}, error)
    GetUserEffectiveRoles(ctx context.Context, orgID, userID string) (interface{}, error)
}
```

### RoleInheritanceEngine API
```go
// Get all effective roles (direct + inherited) for a user
CalculateEffectiveRoles(ctx context.Context, orgID, userID string) ([]*EffectiveRole, error)

// Verify inheritance is working correctly
VerifyBottomUpInheritance(ctx context.Context, orgID, userID string) (*InheritanceVerificationResult, error)

// Invalidate cache after role changes
InvalidateUserRoleCache(ctx context.Context, orgID, userID string) error
InvalidateGroupRoleCache(ctx context.Context, orgID, groupID string) error
```

### EffectiveRole Structure
```go
type EffectiveRole struct {
    Role            *models.Role
    GroupID         string       // Source group
    GroupName       string
    InheritancePath []string     // Path from user's group to role source
    Distance        int          // 0=direct, 1=child, 2=grandchild, etc.
    IsDirectRole    bool         // Direct assignment vs inherited
}
```

## What's NOT Implemented

1. **RoleInheritanceEngine injection into UserService** ✗
2. **Calling CalculateEffectiveRoles during token generation** ✗
3. **Merging direct + inherited roles for tokens** ✗
4. **GetUserEffectiveRoles in UserService interface** ✗

## Inheritance Model

**Type:** Bottom-up (Upward) ONLY
- Parent groups inherit roles from child groups
- Child groups do NOT inherit from parents
- Good for: Executives need all subordinate permissions

```
CEO Group (Parent)
├── Manager Group (Child) → its roles go UP to CEO
│   └── Employee Group (Grandchild) → its roles go UP through Manager to CEO
└── Director Group (Child) → its roles go UP to CEO

If user is in CEO Group:
- CEO direct roles: distance 0 (highest precedence)
- Manager roles: distance 1 (inherited from child)
- Employee roles: distance 2 (inherited from grandchild)
- Director roles: distance 1 (inherited from child)
```

## Next Steps

### To implement role inheritance in tokens:

1. **Modify UserService** (Service + Interface)
   - Inject RoleInheritanceEngine
   - Update GetUserWithRoles() to call CalculateEffectiveRoles()
   - Merge direct + inherited roles
   - Return combined effective roles

2. **Modify Auth Handler** (LoginV2 and RefreshTokenV2)
   - No changes needed if UserService returns effective roles
   - Or explicitly call inheritance engine there

3. **Update Token Generation**
   - If effective roles passed, they'll be included automatically

## Performance Notes

- RoleInheritanceEngine has 5-minute cache
- Cache invalidated on: group role changes, group hierarchy changes
- Pattern-based invalidation supported
- Recursive traversal with early termination for inactive groups
- In-memory conflict resolution (efficient)

## Testing

3 comprehensive test files:
- `role_inheritance_engine_test.go`
- `role_inheritance_bottom_up_test.go`
- `role_inheritance_comprehensive_test.go`

All tests are marked as `comprehensive` and well-documented.
