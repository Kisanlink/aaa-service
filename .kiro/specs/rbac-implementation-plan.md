# Comprehensive RBAC Implementation Architecture Plan

## 1. Architecture Overview

### 1.1 High-Level Design Decisions

The AAA service implements a **dual-model permission system** to support both simple and complex authorization scenarios:

```
┌─────────────────────────────────────────────────┐
│                  User Request                    │
└─────────────────────┬───────────────────────────┘
                      │
                      ▼
        ┌─────────────────────────────┐
        │   Permission Evaluator      │
        │   (Unified Interface)       │
        └──────────┬──────────────────┘
                   │
        ┌──────────┴──────────┐
        ▼                      ▼
┌──────────────┐      ┌──────────────────┐
│ Model 1:     │      │ Model 2:         │
│ Role →       │      │ Role →           │
│ Permission → │      │ Resource+Action  │
│ Resource+    │      │ (Direct)         │
│ Action       │      └──────────────────┘
└──────────────┘
        │                      │
        └──────────┬───────────┘
                   ▼
        ┌──────────────────────┐
        │  PostgreSQL/Redis    │
        │  (Storage + Cache)   │
        └──────────────────────┘
```

### 1.2 Permission Models Strategy

#### Model 1: Role → Permission → (Resource + Action)
- **Use Case**: Named permissions with semantic meaning
- **Example**: "manage_users" permission includes create, read, update, delete on user resources
- **Tables**: `roles`, `permissions`, `role_permissions`

#### Model 2: Role → Resource + Action (Direct)
- **Use Case**: Fine-grained, resource-specific permissions
- **Example**: Role X can perform "read" action on Resource Y
- **Tables**: `roles`, `resource_permissions`

#### Unified Evaluation Strategy
```go
// Permission check flows through both models
func HasPermission(userID, resourceType, resourceID, action string) bool {
    // 1. Get user's effective roles (including inherited)
    // 2. Check Model 1: role_permissions → permissions
    // 3. Check Model 2: resource_permissions
    // 4. Apply hierarchical rules and inheritance
    // 5. Cache result with appropriate TTL
}
```

### 1.3 Caching Strategy

```
┌─────────────────┐
│  Redis Cache    │
├─────────────────┤
│ Key Patterns:   │
│ • user:{id}:roles (TTL: 5m)
│ • role:{id}:perms (TTL: 10m)
│ • perm:{user}:{resource}:{action} (TTL: 1m)
│ • resource:{id}:hierarchy (TTL: 15m)
└─────────────────┘
```

## 2. Component Design

### 2.1 Repository Layer

#### Action Repository
```
internal/repositories/actions/
├── action_repository.go      # Interface and struct (150 lines)
├── create.go                  # Creation operations (100 lines)
├── read.go                    # Read operations (150 lines)
├── update.go                  # Update operations (100 lines)
└── delete.go                  # Delete operations (80 lines)
```

**Interface Definition:**
```go
type ActionRepository interface {
    Create(ctx context.Context, action *models.Action) error
    GetByID(ctx context.Context, id string) (*models.Action, error)
    GetByName(ctx context.Context, name string) (*models.Action, error)
    List(ctx context.Context, filter *base.Filter) ([]*models.Action, error)
    Update(ctx context.Context, action *models.Action) error
    Delete(ctx context.Context, id string) error
    Count(ctx context.Context, filter *base.Filter) (int64, error)
}
```

#### Permission Repository
```
internal/repositories/permissions/
├── permission_repository.go   # Interface and struct (150 lines)
├── create.go                  # Creation operations (100 lines)
├── read.go                    # Read operations (200 lines)
├── update.go                  # Update operations (100 lines)
└── delete.go                  # Delete operations (80 lines)
```

#### RolePermission Repository
```
internal/repositories/role_permissions/
├── role_permission_repository.go  # Interface and struct (150 lines)
├── assign.go                       # Assignment operations (150 lines)
├── revoke.go                       # Revocation operations (100 lines)
└── query.go                        # Query operations (200 lines)
```

#### ResourcePermission Repository
```
internal/repositories/resource_permissions/
├── resource_permission_repository.go  # Interface and struct (150 lines)
├── assign.go                          # Assignment operations (150 lines)
├── revoke.go                          # Revocation operations (100 lines)
├── query.go                           # Query operations (200 lines)
└── evaluate.go                        # Permission evaluation (250 lines)
```

### 2.2 Service Layer

#### Resource Service
```
internal/services/resources/
├── service.go         # Service struct and constructor (100 lines)
├── create.go          # Resource creation with validation (150 lines)
├── read.go            # Resource retrieval and listing (200 lines)
├── update.go          # Resource updates (150 lines)
├── delete.go          # Resource deletion with cascade (100 lines)
└── hierarchy.go       # Hierarchy operations (200 lines)
```

#### Permission Service
```
internal/services/permissions/
├── service.go         # Service struct and constructor (100 lines)
├── create.go          # Permission creation (150 lines)
├── read.go            # Permission queries (200 lines)
├── update.go          # Permission updates (150 lines)
├── delete.go          # Permission deletion (100 lines)
├── evaluate.go        # Permission evaluation logic (250 lines)
└── cache.go           # Caching operations (150 lines)
```

#### Role Assignment Service
```
internal/services/role_assignments/
├── service.go         # Service struct and constructor (100 lines)
├── assign.go          # Assign permissions to roles (200 lines)
├── revoke.go          # Revoke permissions from roles (150 lines)
├── query.go           # Query role permissions (200 lines)
└── inheritance.go     # Handle role inheritance (250 lines)
```

### 2.3 Handler Layer

#### Resource Handler
```
internal/handlers/resources/
├── resource_handler.go    # Handler struct and routes (200 lines)
├── create.go              # POST /resources (100 lines)
├── read.go                # GET /resources/:id, GET /resources (150 lines)
├── update.go              # PUT /resources/:id (100 lines)
└── delete.go              # DELETE /resources/:id (80 lines)
```

#### Complete Permission Handler
```
internal/handlers/permissions/
├── permission_handler.go  # Handler struct (already exists, refactor)
├── create.go              # POST /permissions (100 lines)
├── read.go                # GET endpoints (150 lines)
├── update.go              # PUT /permissions/:id (100 lines)
├── delete.go              # DELETE /permissions/:id (80 lines)
├── evaluate.go            # POST /permissions/evaluate (150 lines)
└── assign.go              # Assignment endpoints (200 lines)
```

### 2.4 Request/Response DTOs

```
internal/entities/requests/permissions/
├── create_permission.go         # CreatePermissionRequest (50 lines)
├── update_permission.go         # UpdatePermissionRequest (50 lines)
├── assign_permission.go         # AssignPermissionRequest (50 lines)
├── evaluate_permission.go       # EvaluatePermissionRequest (50 lines)
└── query_permission.go          # QueryPermissionRequest (50 lines)

internal/entities/requests/resources/
├── create_resource.go           # CreateResourceRequest (50 lines)
├── update_resource.go           # UpdateResourceRequest (50 lines)
└── query_resource.go            # QueryResourceRequest (50 lines)

internal/entities/responses/permissions/
├── permission_response.go       # PermissionResponse (100 lines)
├── evaluation_response.go       # EvaluationResponse (50 lines)
└── assignment_response.go       # AssignmentResponse (50 lines)

internal/entities/responses/resources/
├── resource_response.go         # ResourceResponse (100 lines)
└── resource_list_response.go    # ResourceListResponse (50 lines)
```

## 3. Implementation Task Breakdown

### Phase 1: Foundation (3-4 days)
**Task 1.1: Implement Action Repository** ⏱️ 4 hours
- Files: 5 files (~580 lines)
- Dependencies: kisanlink-db, models.Action
- Testable: CRUD operations for actions

**Task 1.2: Implement Permission Repository** ⏱️ 4 hours
- Files: 5 files (~630 lines)
- Dependencies: kisanlink-db, models.Permission
- Testable: CRUD operations for permissions

**Task 1.3: Implement Role-Permission Repository** ⏱️ 4 hours
- Files: 4 files (~600 lines)
- Dependencies: kisanlink-db, models.RolePermission
- Testable: Assignment and query operations

**Task 1.4: Implement Resource-Permission Repository** ⏱️ 6 hours
- Files: 5 files (~850 lines)
- Dependencies: kisanlink-db, models.ResourcePermission
- Testable: Direct resource-action assignments

### Phase 2: Business Logic (4-5 days)
**Task 2.1: Implement Resource Service** ⏱️ 6 hours
- Files: 6 files (~900 lines)
- Dependencies: ResourceRepository, CacheService
- Testable: Resource management with hierarchy

**Task 2.2: Complete Action Service** ⏱️ 3 hours
- Files: Update existing + 3 new files (~400 lines)
- Dependencies: ActionRepository
- Testable: Action lifecycle management

**Task 2.3: Implement Permission Service** ⏱️ 8 hours
- Files: 7 files (~1100 lines)
- Dependencies: PermissionRepository, CacheService
- Testable: Permission evaluation with caching

**Task 2.4: Implement Role Assignment Service** ⏱️ 6 hours
- Files: 5 files (~900 lines)
- Dependencies: RolePermissionRepository, ResourcePermissionRepository
- Testable: Role-permission assignments with inheritance

### Phase 3: API Layer (3-4 days)
**Task 3.1: Implement Resource Handler** ⏱️ 4 hours
- Files: 5 files (~630 lines)
- Dependencies: ResourceService
- Testable: REST endpoints for resources

**Task 3.2: Complete Permission Handler** ⏱️ 6 hours
- Files: 7 files (~780 lines)
- Dependencies: PermissionService
- Testable: Full permission management endpoints

**Task 3.3: Implement Request/Response DTOs** ⏱️ 4 hours
- Files: 15 files (~1000 lines)
- Dependencies: None
- Testable: Validation and transformation

**Task 3.4: Wire up routes and dependencies** ⏱️ 2 hours
- Files: Update main.go, routes.go
- Dependencies: All handlers
- Testable: Integration tests

### Phase 4: Testing & Documentation (2-3 days)
**Task 4.1: Unit tests for repositories** ⏱️ 6 hours
**Task 4.2: Unit tests for services** ⏱️ 6 hours
**Task 4.3: Integration tests** ⏱️ 4 hours
**Task 4.4: API documentation** ⏱️ 2 hours
**Task 4.5: Seed data and migrations** ⏱️ 2 hours

## 4. API Design

### 4.1 REST Endpoints

#### Resource Endpoints
```
POST   /api/v1/resources              # Create resource
GET    /api/v1/resources              # List resources
GET    /api/v1/resources/:id          # Get resource
PUT    /api/v1/resources/:id          # Update resource
DELETE /api/v1/resources/:id          # Delete resource
GET    /api/v1/resources/:id/children # Get child resources
```

#### Permission Endpoints
```
POST   /api/v1/permissions            # Create permission
GET    /api/v1/permissions            # List permissions
GET    /api/v1/permissions/:id        # Get permission
PUT    /api/v1/permissions/:id        # Update permission
DELETE /api/v1/permissions/:id        # Delete permission
POST   /api/v1/permissions/evaluate   # Evaluate permission
```

#### Role-Permission Assignment Endpoints
```
POST   /api/v1/roles/:id/permissions         # Assign permissions to role
DELETE /api/v1/roles/:id/permissions/:permId # Revoke permission from role
GET    /api/v1/roles/:id/permissions         # List role permissions
POST   /api/v1/roles/:id/resources           # Assign resource+action to role
DELETE /api/v1/roles/:id/resources/:resId    # Revoke resource from role
GET    /api/v1/roles/:id/resources           # List role resources
```

#### User Permission Endpoints
```
GET    /api/v1/users/:id/permissions         # Get user's effective permissions
POST   /api/v1/users/:id/evaluate            # Check specific permission
GET    /api/v1/users/:id/roles/effective     # Get effective roles (with inheritance)
```

### 4.2 Request/Response Structures

#### Create Permission Request
```json
{
  "name": "manage_orders",
  "description": "Full access to order management",
  "resource_id": "RESOURCE_ID",
  "action_id": "ACTION_ID"
}
```

#### Assign Permission to Role Request
```json
{
  "permission_ids": ["PERM_ID_1", "PERM_ID_2"],
  "effective_from": "2024-01-01T00:00:00Z",
  "effective_until": "2024-12-31T23:59:59Z"
}
```

#### Assign Resource-Action to Role Request
```json
{
  "assignments": [
    {
      "resource_type": "order",
      "resource_id": "ORDER_123",
      "actions": ["read", "update"]
    }
  ]
}
```

#### Evaluate Permission Request
```json
{
  "user_id": "USER_ID",
  "resource_type": "order",
  "resource_id": "ORDER_123",
  "action": "update",
  "context": {
    "organization_id": "ORG_ID",
    "group_id": "GROUP_ID"
  }
}
```

#### Evaluation Response
```json
{
  "allowed": true,
  "reason": "Permission granted through role 'order_manager'",
  "effective_roles": ["order_manager", "staff"],
  "cache_hit": true,
  "evaluation_time_ms": 2
}
```

## 5. Data Flow

### 5.1 Permission Evaluation Flow

```
User Request → Handler → Service → Cache Check → Repository → Database
                                        ↓
                                   Cache Miss
                                        ↓
                                 Get User Roles
                                        ↓
                                Get Role Hierarchy
                                        ↓
                            Check Model 1 Permissions
                                        ↓
                            Check Model 2 Permissions
                                        ↓
                                 Apply Context
                                        ↓
                                  Cache Result
                                        ↓
                                 Return Decision
```

### 5.2 Role Inheritance Flow

```
User → Groups → Organizations
  ↓        ↓           ↓
Roles   Roles      Roles
  ↓        ↓           ↓
  └────────┴───────────┘
            ↓
     Effective Roles
            ↓
       Permissions
```

## 6. Testing Strategy

### 6.1 Unit Tests

#### Repository Tests
- CRUD operations for each entity
- Filter and pagination
- Transaction handling
- Error scenarios

#### Service Tests
- Business logic validation
- Cache interactions
- Permission evaluation logic
- Role inheritance

#### Handler Tests
- Request validation
- Response formatting
- Error handling
- Status codes

### 6.2 Integration Tests

```go
// Example: End-to-end permission flow
func TestPermissionEvaluation_E2E(t *testing.T) {
    // 1. Create organization and groups
    // 2. Create roles with hierarchy
    // 3. Create resources and actions
    // 4. Assign permissions (both models)
    // 5. Create user and assign to group
    // 6. Evaluate various permission scenarios
    // 7. Verify caching behavior
}
```

### 6.3 Business Logic Tests

For `@agent-business-logic-tester`:
1. **Hierarchical Permission Inheritance**
   - User inherits permissions from group
   - Group inherits from parent group
   - Organization-level permissions cascade

2. **Permission Model Interactions**
   - Model 1 overrides Model 2
   - Negative permissions (denials)
   - Time-based permissions

3. **Edge Cases**
   - Circular role references
   - Orphaned permissions
   - Cache invalidation scenarios

## 7. Security Considerations

### 7.1 Authorization Requirements

```go
// Who can manage permissions
const (
    ManagePermissions = "aaa.permissions.manage"
    AssignRoles       = "aaa.roles.assign"
    ManageResources   = "aaa.resources.manage"
)

// Required permissions for operations
operations := map[string]string{
    "CreatePermission": ManagePermissions,
    "AssignRole":       AssignRoles,
    "CreateResource":   ManageResources,
}
```

### 7.2 Input Validation

- Resource type format: `^[a-z]+(/[a-z]+)*$`
- Action names: `^[a-z_]+$`
- Permission names: `^[a-z][a-z0-9_]*$`
- SQL injection prevention via parameterized queries
- XSS prevention in descriptions

### 7.3 Audit Requirements

```go
type PermissionAuditLog struct {
    UserID       string
    Action       string // "grant", "revoke", "evaluate"
    ResourceType string
    ResourceID   string
    Permission   string
    Result       string // "allowed", "denied"
    Reason       string
    Timestamp    time.Time
    IPAddress    string
    UserAgent    string
}
```

### 7.4 Rate Limiting

- Permission evaluation: 100 req/sec per user
- Permission assignment: 10 req/min per admin
- Bulk operations: 1 req/min

## 8. Migration and Seeding

### 8.1 Default Permissions

```sql
-- Core system permissions
INSERT INTO permissions (name, description) VALUES
('aaa.users.create', 'Create new users'),
('aaa.users.read', 'View user information'),
('aaa.users.update', 'Update user information'),
('aaa.users.delete', 'Delete users'),
('aaa.roles.manage', 'Manage roles and permissions'),
('aaa.audit.view', 'View audit logs');
```

### 8.2 Default Roles

```sql
-- System roles
INSERT INTO roles (name, scope, description) VALUES
('super_admin', 'GLOBAL', 'Full system access'),
('org_admin', 'ORG', 'Organization administrator'),
('group_admin', 'ORG', 'Group administrator'),
('user', 'ORG', 'Basic user role');
```

### 8.3 Migration Strategy

1. **Add indexes for performance**:
```sql
CREATE INDEX idx_resource_permissions_evaluation
ON resource_permissions(role_id, resource_type, resource_id, action);

CREATE INDEX idx_role_permissions_lookup
ON role_permissions(role_id, permission_id);
```

2. **Add constraints**:
```sql
ALTER TABLE resource_permissions
ADD CONSTRAINT unique_role_resource_action
UNIQUE(role_id, resource_type, resource_id, action);
```

## 9. Performance Optimizations

### 9.1 Database Optimizations
- Composite indexes for permission lookups
- Materialized views for role hierarchies
- Partition tables by organization for multi-tenancy

### 9.2 Caching Strategy
- Multi-level cache (L1: in-memory, L2: Redis)
- Lazy loading with cache-aside pattern
- Batch cache invalidation on permission changes

### 9.3 Query Optimizations
```sql
-- Optimized permission check query
WITH RECURSIVE role_hierarchy AS (
    -- Get direct roles
    SELECT role_id FROM user_roles WHERE user_id = $1
    UNION
    -- Get inherited roles
    SELECT r.parent_id FROM roles r
    JOIN role_hierarchy rh ON r.id = rh.role_id
    WHERE r.parent_id IS NOT NULL
)
SELECT EXISTS (
    -- Check Model 1
    SELECT 1 FROM role_permissions rp
    JOIN permissions p ON rp.permission_id = p.id
    WHERE rp.role_id IN (SELECT role_id FROM role_hierarchy)
    AND p.resource_id = $2 AND p.action_id = $3
    UNION
    -- Check Model 2
    SELECT 1 FROM resource_permissions
    WHERE role_id IN (SELECT role_id FROM role_hierarchy)
    AND resource_type = $4 AND resource_id = $2 AND action = $3
);
```

## 10. Monitoring and Observability

### 10.1 Key Metrics
```go
// Prometheus metrics
permission_evaluations_total{result="allowed|denied"}
permission_evaluation_duration_seconds
cache_hit_rate{cache="permission|role|resource"}
permission_assignments_total{type="grant|revoke"}
```

### 10.2 Logging
```go
// Structured logging for permission operations
logger.Info("Permission evaluated",
    zap.String("user_id", userID),
    zap.String("resource", resourceType),
    zap.String("action", action),
    zap.Bool("allowed", result),
    zap.Duration("duration", elapsed))
```

### 10.3 Alerts
- Permission evaluation latency > 100ms
- Cache hit rate < 80%
- Failed permission assignments
- Unauthorized access attempts spike

## 11. Next Steps

1. **Immediate Actions**:
   - Review and approve this architecture plan
   - Set up development environment
   - Create feature branch from `farmers-grpc`

2. **Development Order**:
   - Start with Phase 1 (Foundation repositories)
   - Implement services with basic functionality
   - Add handlers and wire up routes
   - Comprehensive testing
   - Performance optimization

3. **Dependencies**:
   - Ensure kisanlink-db is properly configured
   - Redis cache service must be available
   - Audit service should be ready for integration

## Appendix: File Structure Summary

```
internal/
├── repositories/
│   ├── actions/               # ~580 lines
│   ├── permissions/           # ~630 lines
│   ├── role_permissions/      # ~600 lines
│   └── resource_permissions/  # ~850 lines
├── services/
│   ├── resources/             # ~900 lines
│   ├── actions/               # ~400 lines
│   ├── permissions/           # ~1100 lines
│   └── role_assignments/      # ~900 lines
├── handlers/
│   ├── resources/             # ~630 lines
│   └── permissions/           # ~780 lines
├── entities/
│   ├── requests/
│   │   ├── permissions/       # ~250 lines
│   │   └── resources/         # ~150 lines
│   └── responses/
│       ├── permissions/       # ~200 lines
│       └── resources/         # ~150 lines

Total: ~45 files, ~9,620 lines of code
Estimated Development Time: 12-16 days
```

This architecture provides a solid foundation for implementing a comprehensive RBAC system with support for hierarchical roles, dual permission models, and efficient evaluation with caching.
