# Role Inheritance Engine Documentation

## Overview

The Role Inheritance Engine implements **bottom-up (upward) inheritance** for hierarchical group structures in the AAA service. This means that parent groups inherit roles from their child groups, allowing executives and managers to perform any action their subordinates can perform.

## Inheritance Flow

### Bottom-Up Inheritance Logic

The inheritance follows this pattern:

```
CEO Group (Parent)
├── Manager Group (Child)
│   └── Employee Group (Grandchild)
└── Director Group (Child)
    └── Senior Employee Group (Grandchild)
```

If a user is a member of the **CEO Group**, they will inherit:

1. **Direct roles** assigned to CEO Group (distance = 0, highest precedence)
2. **Inherited roles** from Manager Group (distance = 1)
3. **Inherited roles** from Employee Group (distance = 2)
4. **Inherited roles** from Director Group (distance = 1)
5. **Inherited roles** from Senior Employee Group (distance = 2)

### Algorithm Steps

1. **Start with user's direct group memberships**

   - Get all groups the user is directly a member of
   - These become the starting points for inheritance calculation

2. **For each direct group, calculate bottom-up roles**

   - Get roles directly assigned to the group (distance = 0)
   - Recursively traverse all child groups
   - Collect roles from each child group with increasing distance
   - Build inheritance paths showing the route from user's group to role source

3. **Apply conflict resolution**

   - When the same role exists at multiple levels, keep the most specific one
   - Most specific = shortest distance from user's direct group
   - Direct assignments (distance 0) always win over inherited roles

4. **Sort by precedence**
   - Primary sort: distance (ascending - most specific first)
   - Secondary sort: role name (for consistency)

## Key Concepts

### Distance and Precedence

- **Distance 0**: Direct role assignment to user's group (highest precedence)
- **Distance 1**: Role inherited from immediate child group
- **Distance 2**: Role inherited from grandchild group
- **Distance N**: Role inherited from N-levels down the hierarchy

### Inheritance Path

Each effective role includes an inheritance path showing how the role was obtained:

```go
type EffectiveRole struct {
    Role            *models.Role `json:"role"`
    GroupID         string       `json:"group_id"`         // Source group that has the role
    GroupName       string       `json:"group_name"`       // Name of source group
    InheritancePath []string     `json:"inheritance_path"` // Path from user's group to source
    Distance        int          `json:"distance"`         // Distance from user's direct group
    IsDirectRole    bool         `json:"is_direct_role"`   // True if distance = 0
}
```

Example inheritance path: `["ceo-group", "manager-group", "employee-group"]`

- User is in `ceo-group`
- Role comes from `employee-group`
- Path shows: CEO → Manager → Employee

### Conflict Resolution

When the same role exists at multiple levels:

```
CEO Group: [Admin Role (distance 0)]
├── Manager Group: [Admin Role (distance 1)]
└── Employee Group: [Admin Role (distance 2)]
```

Result: CEO gets Admin Role with distance 0 (direct assignment wins)

## Business Logic Rationale

### Why Bottom-Up Inheritance?

This inheritance model reflects real-world organizational hierarchies:

1. **Executive Oversight**: CEOs and managers need to perform any action their subordinates can perform
2. **Escalation Support**: Higher-level staff can step in and handle lower-level tasks
3. **Audit and Compliance**: Executives can access all systems their teams use for oversight
4. **Operational Flexibility**: Managers can cover for absent team members

### Example Scenarios

#### Scenario 1: Software Development Team

```
Tech Lead Group
├── Senior Developer Group: [Deploy to Staging, Code Review]
└── Junior Developer Group: [Submit Code, Run Tests]
```

**Result**: Tech Lead inherits all permissions (Deploy, Review, Submit, Test)

#### Scenario 2: Financial Department

```
CFO Group: [Approve Budget]
├── Finance Manager Group: [View Reports, Process Payments]
└── Accountant Group: [Enter Transactions, Generate Reports]
```

**Result**: CFO can approve budgets (direct) + view reports + process payments + enter transactions + generate reports (inherited)

## Implementation Details

### Core Methods

#### `CalculateEffectiveRoles(ctx, orgID, userID)`

Main entry point that calculates all effective roles for a user.

#### `calculateBottomUpRoles(ctx, group, currentDistance)`

Recursive method that:

1. Gets direct roles for the current group
2. Gets all child groups
3. Recursively calculates roles for each child
4. Merges child roles with conflict resolution
5. Updates inheritance paths

### Caching Strategy

The engine uses Redis caching with these keys:

- `org:{orgId}:user:{userId}:effective_roles` - Cached effective roles (5 min TTL)
- `org:{orgId}:user:{userId}:groups` - Cached user group memberships (5 min TTL)

Cache invalidation occurs when:

- User group memberships change
- Group role assignments change
- Group hierarchy changes

### Performance Considerations

1. **Recursive Depth**: Algorithm handles deep hierarchies efficiently
2. **Caching**: Aggressive caching reduces database queries
3. **Early Termination**: Inactive groups are skipped
4. **Conflict Resolution**: In-memory resolution avoids duplicate database calls

## Testing Coverage

The implementation includes comprehensive tests for:

### Basic Functionality

- Single-level inheritance (parent ← child)
- Multi-level inheritance (grandparent ← parent ← child)
- No group memberships (empty result)

### Edge Cases

- Role conflict resolution (same role at multiple levels)
- Inactive groups (should be ignored)
- Cache hit/miss scenarios
- Error handling for failed database calls

### Test Scenarios

#### Single-Level Inheritance Test

```
Parent Group (user member) ← Child Group
Parent Role (distance 0) + Child Role (distance 1)
```

#### Conflict Resolution Test

```
Parent Group (user member): Shared Role
Child Group: Same Shared Role
Result: Parent's role wins (distance 0 vs distance 1)
```

## Usage Examples

### Getting User's Effective Roles

```go
engine := NewRoleInheritanceEngine(groupRepo, groupRoleRepo, roleRepo, membershipRepo, cache, logger)
effectiveRoles, err := engine.CalculateEffectiveRoles(ctx, "org-123", "user-456")

for _, role := range effectiveRoles {
    fmt.Printf("Role: %s (distance: %d, path: %v)\n",
        role.Role.Name, role.Distance, role.InheritancePath)
}
```

### Cache Invalidation

```go
// Invalidate when user's group memberships change
err := engine.InvalidateUserRoleCache(ctx, "org-123", "user-456")

// Invalidate when group roles change (affects all users in org)
err := engine.InvalidateGroupRoleCache(ctx, "org-123", "group-789")
```

## Monitoring and Debugging

### Logging

The engine provides structured logging for:

- Role calculation start/completion
- Cache hits/misses
- Conflict resolution decisions
- Error conditions
- Performance metrics

### Debug Information

Each `EffectiveRole` contains debug information:

- Source group ID and name
- Complete inheritance path
- Distance from user's direct group
- Whether it's a direct or inherited role

This information helps administrators understand why a user has specific permissions and troubleshoot access issues.
