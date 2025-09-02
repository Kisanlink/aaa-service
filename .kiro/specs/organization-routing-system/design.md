# Design Document

## Overview

The Organization Routing System addresses the current 404 errors for organization-related endpoints by enhancing the existing organization and group management infrastructure. The design focuses on proper route wiring, extending existing hierarchical group structures with bi-directional role inheritance, and ensuring complete tenant isolation.

The system will build upon existing AAA service artifacts:

- Existing `Organization` and `Group` models with kisanlink-db BaseModel
- Current organization and group handlers and services
- Established route structure in `internal/routes/`
- Existing interfaces and middleware components

## Architecture

### High-Level Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Routes   │────│   Handlers      │────│   Services      │
│                 │    │                 │    │                 │
│ Organization    │    │ Organization    │    │ Organization    │
│ Group           │    │ Group           │    │ Group           │
│ User-Group      │    │ User-Group      │    │ User-Group      │
│ Role-Group      │    │ Role-Group      │    │ Role-Group      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Repositories   │────│   Database      │    │   Cache Layer   │
│                 │    │                 │    │                 │
│ Organization    │    │ PostgreSQL      │    │ Redis           │
│ Group           │    │ Tables:         │    │                 │
│ GroupMembership │    │ - organizations │    │ Role Cache      │
│ GroupRole       │    │ - groups        │    │ Hierarchy Cache │
│                 │    │ - group_members │    │ Permission Cache│
│                 │    │ - group_roles   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Enhanced Route Structure

Building on existing routes in `internal/routes/organization_routes.go` and `internal/routes/group_routes.go`:

```
/api/v1/organizations (existing routes enhanced)
├── GET    /                           # List organizations (existing)
├── POST   /                           # Create organization (existing)
├── GET    /:orgId                     # Get organization details (existing)
├── PUT    /:orgId                     # Update organization (existing)
├── DELETE /:orgId                     # Delete organization (existing)
├── GET    /:orgId/hierarchy           # Get organization hierarchy (existing)
├── GET    /:orgId/groups              # List organization groups (NEW)
├── POST   /:orgId/groups              # Create group in organization (NEW)
├── GET    /:orgId/groups/:groupId     # Get group details (NEW)
├── PUT    /:orgId/groups/:groupId     # Update group (NEW)
├── DELETE /:orgId/groups/:groupId     # Delete group (NEW)
├── GET    /:orgId/groups/:groupId/hierarchy    # Get group hierarchy (NEW)
├── POST   /:orgId/groups/:groupId/users       # Add user to group (NEW)
├── DELETE /:orgId/groups/:groupId/users/:userId # Remove user from group (NEW)
├── GET    /:orgId/groups/:groupId/users       # List group users (NEW)
├── POST   /:orgId/groups/:groupId/roles       # Assign role to group (NEW)
├── DELETE /:orgId/groups/:groupId/roles/:roleId # Remove role from group (NEW)
├── GET    /:orgId/groups/:groupId/roles       # List group roles (NEW)
├── GET    /:orgId/users/:userId/groups        # Get user's groups (NEW)
├── GET    /:orgId/users/:userId/effective-roles # Get user's effective roles (NEW)
└── GET    /:orgId/hierarchy                   # Get complete org hierarchy (existing)
```

## Components and Interfaces

### Database Models

The system will reuse and extend existing models:

#### Existing Organization Model (internal/entities/models/organization.go)

- Already has `*base.BaseModel` with kisanlink-db integration
- Contains `Name`, `Description`, `ParentID` for hierarchy
- Has `IsActive` status field
- Includes `Metadata` JSONB field for extensibility
- Supports parent-child relationships

#### Existing Group Model (internal/entities/models/group.go)

- Already has `*base.BaseModel` with kisanlink-db integration
- Contains `OrganizationID`, `ParentID` for hierarchy
- Has existing `GroupMembership` and `GroupInheritance` models
- Supports time-bounded memberships with `StartsAt`/`EndsAt`
- Includes `PrincipalType` for user/service distinction

#### New GroupRole Junction Model (to be added)

```go
type GroupRole struct {
    *base.BaseModel
    GroupID        string     `json:"group_id" gorm:"type:varchar(255);not null"`
    RoleID         string     `json:"role_id" gorm:"type:varchar(255);not null"`
    OrganizationID string     `json:"organization_id" gorm:"type:varchar(255);not null"`
    AssignedBy     string     `json:"assigned_by" gorm:"type:varchar(255);not null"`
    StartsAt       *time.Time `json:"starts_at" gorm:"type:timestamp"`
    EndsAt         *time.Time `json:"ends_at" gorm:"type:timestamp"`
    IsActive       bool       `json:"is_active" gorm:"default:true"`
    Metadata       *string    `json:"metadata" gorm:"type:jsonb"`

    // Relationships
    Group        *Group        `json:"group" gorm:"foreignKey:GroupID;references:ID"`
    Role         *Role         `json:"role" gorm:"foreignKey:RoleID;references:ID"`
    Organization *Organization `json:"organization" gorm:"foreignKey:OrganizationID;references:ID"`
    Assigner     *User         `json:"assigner" gorm:"foreignKey:AssignedBy;references:ID"`
}
```

### Service Layer Architecture

#### Existing Organization Service (internal/interfaces/interfaces.go)

The existing `OrganizationService` interface will be extended with new methods:

- Current methods: `CreateOrganization`, `GetOrganization`, `UpdateOrganization`, etc.
- New methods to add: Group management within organization context

#### Existing Group Service (internal/interfaces/interfaces.go)

The existing `GroupService` interface will be extended with:

- Organization-scoped group operations
- Role-group assignment methods
- Bi-directional inheritance calculation methods

#### New Methods to Add to Existing Services

```go
// Extensions to OrganizationService interface
type OrganizationService interface {
    // ... existing methods ...

    // New group management methods
    GetOrganizationGroups(ctx context.Context, orgID string, pagination interface{}) (interface{}, error)
    GetOrganizationHierarchy(ctx context.Context, orgID string) (interface{}, error)
}

// Extensions to GroupService interface
type GroupService interface {
    // ... existing methods ...

    // New role-group methods
    AssignRoleToGroup(ctx context.Context, groupID, roleID, assignedBy string) (interface{}, error)
    RemoveRoleFromGroup(ctx context.Context, groupID, roleID string) error
    GetGroupRoles(ctx context.Context, groupID string) (interface{}, error)
    GetUserEffectiveRoles(ctx context.Context, orgID, userID string) (interface{}, error)
}
```

### Role Inheritance Engine

#### Bi-directional Inheritance Algorithm

```go
type RoleInheritanceEngine struct {
    groupRepo interfaces.GroupRepository
    roleRepo  interfaces.RoleRepository
    cache     interfaces.CacheService
}

func (r *RoleInheritanceEngine) CalculateEffectiveRoles(ctx context.Context, orgID, userID string) ([]Role, error) {
    // 1. Get user's direct group memberships using existing GroupMembership model
    directGroups := r.getUserDirectGroups(ctx, orgID, userID)

    // 2. Get all ancestor groups (upward inheritance) using existing Group.Parent relationships
    ancestorGroups := r.getAncestorGroups(ctx, directGroups)

    // 3. Get all descendant groups (downward inheritance) using existing Group.Children relationships
    descendantGroups := r.getDescendantGroups(ctx, directGroups)

    // 4. Collect roles from all groups using new GroupRole model
    allGroups := append(directGroups, ancestorGroups...)
    allGroups = append(allGroups, descendantGroups...)

    // 5. Get roles for all groups
    roles := r.getRolesForGroups(ctx, allGroups)

    // 6. Apply conflict resolution (most specific wins)
    effectiveRoles := r.resolveRoleConflicts(roles, directGroups)

    return effectiveRoles, nil
}
```

## Data Models

### Database Schema Extensions

The system will add minimal new tables while leveraging existing ones:

```sql
-- New GroupRole table (extends existing schema)
CREATE TABLE group_roles (
    id VARCHAR(255) PRIMARY KEY,
    group_id VARCHAR(255) NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    role_id VARCHAR(255) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    organization_id VARCHAR(255) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    assigned_by VARCHAR(255) NOT NULL REFERENCES users(id),
    starts_at TIMESTAMP,
    ends_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    UNIQUE(group_id, role_id, organization_id),
    INDEX idx_group_roles_group_org (group_id, organization_id),
    INDEX idx_group_roles_role_org (role_id, organization_id)
);
```

### Caching Strategy

#### Cache Keys Structure (extending existing cache patterns)

```
org:{orgId}:hierarchy                    # Complete organization hierarchy
org:{orgId}:groups                       # All groups in organization
org:{orgId}:user:{userId}:groups         # User's direct groups
org:{orgId}:user:{userId}:effective_roles # User's effective roles
org:{orgId}:group:{groupId}:ancestors    # Group's ancestor chain
org:{orgId}:group:{groupId}:descendants  # Group's descendant chain
org:{orgId}:group:{groupId}:roles        # Group's assigned roles
```

#### Cache Invalidation Strategy

- Organization changes: Invalidate `org:{orgId}:*`
- Group hierarchy changes: Invalidate hierarchy and affected user caches
- User-group changes: Invalidate user-specific caches
- Role-group changes: Invalidate role-related caches

## Error Handling

### Error Types (extending existing error handling)

```go
const (
    ErrOrgNotFound          = "ORG_NOT_FOUND"
    ErrGroupNotFound        = "GROUP_NOT_FOUND"
    ErrCircularHierarchy    = "CIRCULAR_HIERARCHY"
    ErrInvalidParentGroup   = "INVALID_PARENT_GROUP"
    ErrUserNotInOrg         = "USER_NOT_IN_ORG"
    ErrGroupNotInOrg        = "GROUP_NOT_IN_ORG"
    ErrRoleNotFound         = "ROLE_NOT_FOUND"
    ErrDuplicateAssignment  = "DUPLICATE_ASSIGNMENT"
    ErrInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
)
```

### Validation Rules

- Organization codes must be unique globally (existing)
- Group codes must be unique within organization (existing)
- Parent groups must exist in same organization
- Circular hierarchies are prevented
- Users and groups must belong to same organization for assignments
- Role assignments require appropriate permissions

## Testing Strategy

### Unit Tests

- Service layer methods with mocked repositories
- Role inheritance algorithm with various hierarchy scenarios
- Validation logic for all business rules
- Error handling for edge cases

### Integration Tests

- Database operations with real PostgreSQL
- Cache operations with Redis
- Complete request-response cycles
- Multi-tenant isolation verification

### End-to-End Tests

- Full API workflows for organization management
- Complex hierarchy creation and modification
- Role inheritance verification across multiple levels
- Performance testing with large hierarchies

### Test Data Scenarios

- Simple flat organization structure
- Deep hierarchical organization (5+ levels)
- Wide organization (many groups at same level)
- Complex role inheritance patterns
- Multi-tenant scenarios with isolation verification

## Implementation Approach

### Phase 1: Route Wiring

1. Extend existing `internal/routes/organization_routes.go` with new group-related routes
2. Add organization-scoped endpoints to existing handlers
3. Ensure proper middleware integration

### Phase 2: Model Extensions

1. Add `GroupRole` model to `internal/entities/models/`
2. Create migration for new `group_roles` table
3. Update existing models with new relationships

### Phase 3: Service Enhancements

1. Extend existing organization and group services
2. Implement role inheritance engine
3. Add caching layer for performance

### Phase 4: Handler Updates

1. Add new handler methods for organization-scoped group operations
2. Implement role-group assignment endpoints
3. Add effective roles calculation endpoints

This approach minimizes code duplication and builds upon the solid foundation already established in the AAA service.
