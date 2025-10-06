# Farmers Module gRPC Integration - Design Document

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      Farmers Module (Client)                    │
├─────────────────────────────────────────────────────────────────┤
│  - Farmer CRUD        - Farm Management                         │
│  - FPO Management     - KisanSathi Assignment                   │
└────────────────────────┬────────────────────────────────────────┘
                         │ gRPC (protobuf)
┌────────────────────────▼────────────────────────────────────────┐
│                    AAA Service (gRPC Layer)                     │
├─────────────────────────────────────────────────────────────────┤
│  OrganizationHandler  │  GroupHandler    │  RoleHandler         │
│  PermissionHandler    │  CatalogHandler                         │
└────────────────────────┬────────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────────┐
│                      Service Layer                              │
├─────────────────────────────────────────────────────────────────┤
│  OrganizationService  │  GroupService    │  RoleService         │
│  PermissionService    │  CatalogService                         │
└────────────────────────┬────────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────────┐
│                  Database Layer (kisanlink-db)                  │
├─────────────────────────────────────────────────────────────────┤
│  PostgreSQL 17+ with GORM abstraction                           │
└─────────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. gRPC Handler Layer (internal/grpc_server/)

**Purpose**: Request validation, authentication, response formatting

**Files** (max 300 lines each):
- `organization_handler.go` - OrganizationService gRPC implementation
- `group_handler.go` - GroupService gRPC implementation
- `role_handler.go` - RoleService gRPC implementation
- `permission_handler.go` - PermissionService gRPC implementation
- `catalog_handler.go` - CatalogService gRPC implementation

**Responsibilities**:
1. Parse and validate gRPC requests
2. Extract user context from JWT metadata
3. Call service layer methods
4. Handle errors and convert to gRPC status codes
5. Format responses

**Pattern**:
```go
func (h *OrganizationHandler) CreateOrganization(ctx context.Context, req *pb.CreateOrganizationRequest) (*pb.CreateOrganizationResponse, error) {
    // 1. Extract user context
    userID, err := extractUserFromContext(ctx)

    // 2. Validate request
    if err := validateCreateOrgRequest(req); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }

    // 3. Call service
    org, err := h.orgService.CreateOrganization(ctx, req)
    if err != nil {
        return nil, mapServiceError(err)
    }

    // 4. Return response
    return &pb.CreateOrganizationResponse{
        StatusCode: 200,
        Message: "Organization created successfully",
        Organization: org,
    }, nil
}
```

### 2. Service Layer (internal/services/)

**Purpose**: Business logic, transactions, caching, audit logging

**Files**:
```
internal/services/organization/
├── service.go           # Service struct and constructor
├── create.go           # CreateOrganization logic
├── read.go             # GetOrganization, ListOrganizations
├── update.go           # UpdateOrganization
├── delete.go           # DeleteOrganization
└── members.go          # AddUser, RemoveUser

internal/services/group/
├── service.go          # Service struct and constructor
├── create.go           # CreateGroup
├── read.go             # GetGroup, ListGroups
├── members.go          # AddMember, RemoveMember, ListMembers
├── inheritance.go      # LinkGroups, UnlinkGroups
└── update.go           # UpdateGroup, DeleteGroup

internal/services/role/
├── service.go          # Service struct and constructor
├── assign.go           # AssignRole
├── check.go            # CheckUserRole, GetUserRoles
├── remove.go           # RemoveRole
└── list.go             # ListUsersWithRole

internal/services/permission/
├── service.go          # Service struct and constructor
├── assign.go           # AssignPermissionToGroup
├── check.go            # CheckGroupPermission
├── effective.go        # GetUserEffectivePermissions
└── list.go             # ListGroupPermissions

internal/services/catalog/
├── service.go          # Service struct and constructor
├── seed.go             # SeedRolesAndPermissions
├── roles.go            # CreateRole, ListRoles, UpdateRole, DeleteRole
└── permissions.go      # CreatePermission, ListPermissions
```

**Responsibilities**:
1. Implement business logic
2. Manage database transactions
3. Implement caching with Redis
4. Create audit logs
5. Handle complex queries

**Pattern**:
```go
type OrganizationService struct {
    db    *gorm.DB
    cache CacheService
    audit AuditService
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, req *pb.CreateOrganizationRequest) (*entities.Organization, error) {
    // Start transaction
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // Create organization
    org := &entities.Organization{
        ID: generateID(),
        Name: req.Name,
        // ... other fields
    }

    if err := tx.Create(org).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // Audit log
    s.audit.Log(ctx, "organization.created", org.ID, userID)

    // Commit
    if err := tx.Commit().Error; err != nil {
        return nil, err
    }

    // Invalidate cache
    s.cache.Delete("org:" + org.ID)

    return org, nil
}
```

### 3. Database Layer (kisanlink-db)

**Usage**: Direct GORM operations via kisanlink-db manager

**Pattern**:
```go
// Use kisanlink-db for database operations
db := dbManager.GetDB()

// Simple queries
var org entities.Organization
err := db.Where("id = ?", orgID).First(&org).Error

// Complex queries with joins
var groups []entities.Group
err := db.Preload("Memberships").
    Where("organization_id = ?", orgID).
    Find(&groups).Error

// Transactions
tx := db.Begin()
tx.Create(&org)
tx.Create(&userOrg)
tx.Commit()
```

## Data Models (internal/entities/)

### Organization
```go
type Organization struct {
    ID              string         `gorm:"primaryKey"`
    CreatedAt       time.Time
    UpdatedAt       time.Time
    CreatedBy       string
    UpdatedBy       string
    DeletedAt       gorm.DeletedAt
    Name            string         `gorm:"uniqueIndex;not null"`
    Description     string
    ParentID        *string
    IsActive        bool           `gorm:"default:true"`
    Metadata        datatypes.JSON
    Parent          *Organization  `gorm:"foreignKey:ParentID"`
    Children        []Organization `gorm:"foreignKey:ParentID"`
}
```

### Group
```go
type Group struct {
    ID             string         `gorm:"primaryKey"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      gorm.DeletedAt
    Name           string         `gorm:"not null"`
    Description    string
    OrganizationID string         `gorm:"not null"`
    ParentID       *string
    IsActive       bool           `gorm:"default:true"`
    Metadata       datatypes.JSON
    Organization   Organization   `gorm:"foreignKey:OrganizationID"`
    Parent         *Group         `gorm:"foreignKey:ParentID"`
    Memberships    []GroupMembership
}
```

### Role
```go
type Role struct {
    ID             string         `gorm:"primaryKey"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
    Name           string         `gorm:"uniqueIndex;not null"`
    Description    string
    Scope          string         `gorm:"not null"` // GLOBAL or ORG
    IsActive       bool           `gorm:"default:true"`
    Version        int            `gorm:"default:1"`
    Metadata       datatypes.JSON
    OrganizationID *string
    ParentID       *string
    Permissions    []Permission   `gorm:"many2many:role_permissions"`
}
```

### Permission
```go
type Permission struct {
    ID          string    `gorm:"primaryKey"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Name        string    `gorm:"uniqueIndex;not null"`
    Description string
    ResourceID  *string
    ActionID    *string
    IsActive    bool      `gorm:"default:true"`
    Resource    *Resource `gorm:"foreignKey:ResourceID"`
    Action      *Action   `gorm:"foreignKey:ActionID"`
}
```

## Proto Definitions

### RoleService Proto (pkg/proto/role.proto)

```protobuf
syntax = "proto3";
package pb;
option go_package = "github.com/Kisanlink/aaa-service/pkg/proto;pb";

import "google/protobuf/timestamp.proto";

// Assign role to user
message AssignRoleRequest {
    string user_id = 1;
    string org_id = 2;
    string role_name = 3;
}

message AssignRoleResponse {
    int32 status_code = 1;
    string message = 2;
}

// Check if user has role
message CheckUserRoleRequest {
    string user_id = 1;
    string role_name = 2;
    string org_id = 3; // optional
}

message CheckUserRoleResponse {
    bool has_role = 1;
    string org_id = 2;
}

// Remove role from user
message RemoveRoleRequest {
    string user_id = 1;
    string org_id = 2;
    string role_name = 3;
}

message RemoveRoleResponse {
    int32 status_code = 1;
    string message = 2;
}

// Get user's roles
message GetUserRolesRequest {
    string user_id = 1;
    string org_id = 2; // optional filter
}

message UserRole {
    string role_name = 1;
    string org_id = 2;
    string org_name = 3;
    google.protobuf.Timestamp assigned_at = 4;
}

message GetUserRolesResponse {
    int32 status_code = 1;
    string message = 2;
    repeated UserRole roles = 3;
}

// List users with role
message ListUsersWithRoleRequest {
    string role_name = 1;
    string org_id = 2;
    int32 page = 3;
    int32 page_size = 4;
}

message UserSummary {
    string id = 1;
    string username = 2;
    string phone_number = 3;
}

message ListUsersWithRoleResponse {
    int32 status_code = 1;
    string message = 2;
    repeated UserSummary users = 3;
    int32 total_count = 4;
}

service RoleService {
    rpc AssignRole(AssignRoleRequest) returns (AssignRoleResponse);
    rpc CheckUserRole(CheckUserRoleRequest) returns (CheckUserRoleResponse);
    rpc RemoveRole(RemoveRoleRequest) returns (RemoveRoleResponse);
    rpc GetUserRoles(GetUserRolesRequest) returns (GetUserRolesResponse);
    rpc ListUsersWithRole(ListUsersWithRoleRequest) returns (ListUsersWithRoleResponse);
}
```

### PermissionService Proto (pkg/proto/permission.proto)

```protobuf
syntax = "proto3";
package pb;
option go_package = "github.com/Kisanlink/aaa-service/pkg/proto;pb";

message AssignPermissionToGroupRequest {
    string group_id = 1;
    string resource = 2;
    string action = 3;
}

message AssignPermissionToGroupResponse {
    int32 status_code = 1;
    string message = 2;
}

message CheckGroupPermissionRequest {
    string group_id = 1;
    string resource = 2;
    string action = 3;
}

message CheckGroupPermissionResponse {
    bool has_permission = 1;
}

message ListGroupPermissionsRequest {
    string group_id = 1;
}

message PermissionItem {
    string id = 1;
    string resource = 2;
    string action = 3;
    string description = 4;
}

message ListGroupPermissionsResponse {
    int32 status_code = 1;
    string message = 2;
    repeated PermissionItem permissions = 3;
}

message RemovePermissionFromGroupRequest {
    string group_id = 1;
    string resource = 2;
    string action = 3;
}

message RemovePermissionFromGroupResponse {
    int32 status_code = 1;
    string message = 2;
}

message GetUserEffectivePermissionsRequest {
    string user_id = 1;
    string org_id = 2; // optional
}

message GetUserEffectivePermissionsResponse {
    int32 status_code = 1;
    string message = 2;
    repeated PermissionItem permissions = 3;
    repeated string roles = 4;
    repeated string groups = 5;
}

service PermissionService {
    rpc AssignPermissionToGroup(AssignPermissionToGroupRequest) returns (AssignPermissionToGroupResponse);
    rpc CheckGroupPermission(CheckGroupPermissionRequest) returns (CheckGroupPermissionResponse);
    rpc ListGroupPermissions(ListGroupPermissionsRequest) returns (ListGroupPermissionsResponse);
    rpc RemovePermissionFromGroup(RemovePermissionFromGroupRequest) returns (RemovePermissionFromGroupResponse);
    rpc GetUserEffectivePermissions(GetUserEffectivePermissionsRequest) returns (GetUserEffectivePermissionsResponse);
}
```

## Error Handling

### Error Mapping
```go
func mapServiceError(err error) error {
    switch {
    case errors.Is(err, gorm.ErrRecordNotFound):
        return status.Error(codes.NotFound, "resource not found")
    case errors.Is(err, gorm.ErrDuplicatedKey):
        return status.Error(codes.AlreadyExists, "resource already exists")
    case strings.Contains(err.Error(), "foreign key"):
        return status.Error(codes.FailedPrecondition, "related resource not found")
    case strings.Contains(err.Error(), "validation"):
        return status.Error(codes.InvalidArgument, err.Error())
    default:
        return status.Error(codes.Internal, "internal server error")
    }
}
```

## Caching Strategy

### Cache Keys
- `org:{org_id}` - Organization details (TTL: 30min)
- `group:{group_id}` - Group details (TTL: 10min)
- `user_roles:{user_id}` - User's roles (TTL: 10min)
- `group_perms:{group_id}` - Group permissions (TTL: 5min)
- `user_eff_perms:{user_id}:{org_id}` - Effective permissions (TTL: 5min)

### Cache Invalidation
- On organization update: delete `org:{org_id}`
- On role assignment: delete `user_roles:{user_id}`, `user_eff_perms:{user_id}:*`
- On permission assignment: delete `group_perms:{group_id}`, `user_eff_perms:*`

## Audit Logging

### Audit Events
```go
type AuditEvent struct {
    Timestamp    time.Time
    UserID       string
    Action       string  // "org.created", "role.assigned", etc.
    ResourceType string  // "organization", "role", "permission"
    ResourceID   string
    OrgID        string
    Details      map[string]interface{}
}
```

### Events to Audit
- organization.created, organization.updated, organization.deleted
- group.created, group.updated, group.deleted
- group.member_added, group.member_removed
- role.assigned, role.removed
- permission.assigned, permission.removed

## Security Considerations

### Authentication
```go
func extractUserFromContext(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return "", status.Error(codes.Unauthenticated, "missing metadata")
    }

    token := md.Get("authorization")
    if len(token) == 0 {
        return "", status.Error(codes.Unauthenticated, "missing token")
    }

    userID, err := validateJWT(token[0])
    if err != nil {
        return "", status.Error(codes.Unauthenticated, "invalid token")
    }

    return userID, nil
}
```

### Authorization
```go
func (s *OrganizationService) checkOrgAccess(userID, orgID string) error {
    // Check if user is member of organization
    var count int64
    err := s.db.Model(&entities.GroupMembership{}).
        Joins("JOIN groups ON groups.id = group_memberships.group_id").
        Where("group_memberships.principal_id = ?", userID).
        Where("groups.organization_id = ?", orgID).
        Where("group_memberships.is_active = ?", true).
        Count(&count).Error

    if err != nil {
        return err
    }

    if count == 0 {
        return errors.New("access denied to organization")
    }

    return nil
}
```

## Testing Strategy

### Unit Tests
- Test each service method independently
- Mock database and cache dependencies
- Test error scenarios

### Integration Tests
- Test full flow: create org → assign role → check permission
- Test group inheritance
- Test effective permissions calculation
- Test transaction rollback

### Performance Tests
- Benchmark permission checks
- Test with 1000+ concurrent requests
- Verify cache hit rates

## Implementation Order

1. **Phase 1: Proto Definitions** (0.5 day)
   - Create role.proto
   - Create permission.proto
   - Enhance catalog.proto
   - Generate Go code

2. **Phase 2: Service Layer** (2 days)
   - Implement OrganizationService
   - Implement GroupService
   - Implement RoleService
   - Implement PermissionService
   - Implement CatalogService

3. **Phase 3: gRPC Handlers** (1.5 days)
   - Implement organization_handler.go
   - Implement group_handler.go
   - Implement role_handler.go
   - Implement permission_handler.go
   - Implement catalog_handler.go
   - Update grpc_server.go registration

4. **Phase 4: Testing** (1 day)
   - Write unit tests
   - Write integration tests
   - Performance testing
   - Fix issues

## File Organization Summary

```
pkg/proto/
├── role.proto                 # NEW
├── permission.proto           # NEW
├── organization.proto         # EXISTS
├── group.proto                # EXISTS
└── catalog.proto              # ENHANCE

internal/grpc_server/
├── organization_handler.go    # NEW
├── group_handler.go           # NEW
├── role_handler.go            # NEW
├── permission_handler.go      # NEW
├── catalog_handler.go         # NEW
└── grpc_server.go             # UPDATE (registration)

internal/services/
├── organization/              # NEW
│   ├── service.go
│   ├── create.go
│   ├── read.go
│   ├── update.go
│   ├── delete.go
│   └── members.go
├── group/                     # NEW
│   ├── service.go
│   ├── create.go
│   ├── read.go
│   ├── members.go
│   ├── inheritance.go
│   └── update.go
├── role/                      # NEW
│   ├── service.go
│   ├── assign.go
│   ├── check.go
│   ├── remove.go
│   └── list.go
├── permission/                # NEW
│   ├── service.go
│   ├── assign.go
│   ├── check.go
│   ├── effective.go
│   └── list.go
└── catalog/                   # NEW
    ├── service.go
    ├── seed.go
    ├── roles.go
    └── permissions.go
```

## Next Steps

After reviewing this design, proceed to:
1. Create proto definitions
2. Implement service layer
3. Implement gRPC handlers
4. Add comprehensive tests
5. Integration testing with farmers-module
