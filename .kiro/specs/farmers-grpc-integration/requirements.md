# Farmers Module gRPC Integration - Requirements

## Overview

Implement comprehensive gRPC services for AAA service to support the farmers-module integration. The farmers-module requires five critical services: OrganizationService, GroupService, RoleService, PermissionService, and CatalogService.

## Business Context

The farmers-module manages FPOs (Farmer Producer Organizations), farmers, farms, crop cycles, and KisanSathi (field agents). It requires robust authentication and authorization from the AAA service with:

- **Multi-tenant organization management** for FPOs and cooperatives
- **Hierarchical group management** for organizing users within organizations
- **Role-based access control** with roles like admin, fpo_manager, kisansathi, farmer
- **Permission management** for fine-grained access control
- **Catalog management** for role and permission definitions

## Current State Analysis

### ✅ Already Implemented

1. **Database Schema** - Complete schema exists in `migrations/001_create_all_tables.up.sql`:
   - `organizations` table with hierarchical support
   - `groups` table with organization and parent relationships
   - `group_memberships` table for user-group associations
   - `roles` table with scope (GLOBAL/ORG) support
   - `permissions` table with resource and action linkage
   - `actions` and `resources` tables for catalog
   - All necessary junction tables and indexes

2. **Proto Definitions** - Partial implementation exists:
   - ✅ `organization.proto` - Complete OrganizationService definition
   - ✅ `group.proto` - Complete GroupService definition
   - ✅ `catalog.proto` - Partial CatalogService (needs enhancement)
   - ❌ RoleService proto - Missing
   - ❌ PermissionService proto - Missing

3. **gRPC Infrastructure**:
   - ✅ gRPC server setup in `internal/grpc_server/grpc_server.go`
   - ✅ Basic handlers for auth and authorization
   - ❌ Organization, Group, Role, Permission, Catalog handlers - Need implementation

### ❌ Needs Implementation

1. **Proto Definitions**:
   - RoleService proto (assign, check, remove, list roles)
   - PermissionService proto (assign to groups, check, list)
   - Enhanced CatalogService proto (seed, manage roles/permissions)

2. **Service Layer** (internal/services/):
   - Organization service with CRUD operations
   - Group service with membership and inheritance management
   - Role assignment service
   - Permission assignment service
   - Catalog seeding service

3. **gRPC Handlers** (internal/grpc_server/):
   - organization_handler.go
   - group_handler.go
   - role_handler.go
   - permission_handler.go
   - catalog_handler.go

4. **Repositories** (if needed):
   - Most operations can use kisanlink-db directly
   - Complex queries may need repository helpers

## Functional Requirements

### FR1: Organization Management

**FR1.1** - Create organizations with type (FPO, COOPERATIVE, COMPANY, NGO)
**FR1.2** - Support hierarchical organizations (parent-child relationships)
**FR1.3** - Update organization details (name, description, status, CEO)
**FR1.4** - List organizations with filtering (type, status, search)
**FR1.5** - Soft delete organizations
**FR1.6** - Add/remove users to/from organizations with roles

### FR2: Group Management

**FR2.1** - Create groups within organizations
**FR2.2** - Support hierarchical groups (parent-child relationships)
**FR2.3** - Add/remove users to/from groups
**FR2.4** - List group members with inherited members from parent groups
**FR2.5** - Link/unlink groups for inheritance
**FR2.6** - Update and delete groups

### FR3: Role Management

**FR3.1** - Assign roles to users in organization context
**FR3.2** - Check if user has specific role in organization
**FR3.3** - Remove roles from users
**FR3.4** - List all roles for a user (across organizations)
**FR3.5** - List all users with specific role
**FR3.6** - Support role scopes (GLOBAL vs ORG)

### FR4: Permission Management

**FR4.1** - Assign permissions to groups
**FR4.2** - Check if group has specific permission
**FR4.3** - List all permissions for a group
**FR4.4** - Remove permissions from groups
**FR4.5** - Get user's effective permissions (from all groups and roles)
**FR4.6** - Support resource-level permissions (resource:action format)

### FR5: Catalog Management

**FR5.1** - Seed default roles (admin, fpo_manager, kisansathi, farmer, readonly)
**FR5.2** - Seed default permissions for farmers-module
**FR5.3** - Create custom roles with permissions
**FR5.4** - Create custom permissions (resource:action)
**FR5.5** - List available roles and permissions
**FR5.6** - Update and delete roles (non-system only)

## Non-Functional Requirements

### NFR1: Performance

- **Response Time**: 95th percentile < 100ms for permission checks
- **Throughput**: Support 1000+ concurrent requests
- **Caching**: Implement Redis caching for permission checks (5-minute TTL)
- **Pagination**: All list operations support pagination (default 50, max 500)

### NFR2: Security

- **Authentication**: All gRPC calls require valid JWT token
- **Authorization**: Enforce organization-level data isolation
- **Input Validation**: Validate all inputs at handler boundary
- **SQL Injection**: Use parameterized queries via GORM/kisanlink-db
- **Audit Logging**: Log all permission and role changes

### NFR3: Reliability

- **Error Handling**: Return appropriate gRPC status codes
- **Transactions**: Use database transactions for multi-table operations
- **Idempotency**: CreateOrganization, CreateGroup should be idempotent
- **Data Integrity**: Enforce foreign key constraints

### NFR4: Maintainability

- **Code Organization**: Follow clean architecture (handlers → services → repositories)
- **File Size**: Maximum 300 lines per file
- **Testing**: Unit tests for all service methods, integration tests for flows
- **Documentation**: Inline comments for complex logic

## Default Roles and Permissions

### Roles

1. **admin**
   - Full access to all resources
   - Permissions: farmers:*, farms:*, fpos:*, users:*, reports:*, admin:maintain

2. **fpo_manager**
   - Manage FPO and its members
   - Permissions: farmers:create/read/update/list, farms:create/read/update/list, fpos:read/update, reports:read

3. **kisansathi**
   - Field agent managing farmers
   - Permissions: farmers:create/read/update, farms:create/read/update

4. **farmer**
   - Self-service access
   - Permissions: farmers:read:self, farmers:update:self, farms:read:self, farms:update:self, farms:create:self

5. **readonly**
   - Read-only access
   - Permissions: farmers:read/list, farms:read/list, fpos:read/list, reports:read

### Permission Format

Format: `resource:action[:scope]`

Examples:
- `farmers:create` - Create any farmer
- `farmers:read:self` - Read own farmer record
- `farms:update` - Update any farm
- `reports:generate` - Generate reports

## Integration Points

### Farmers Module Usage Pattern

```go
// 1. Create FPO organization
orgResp := aaaClient.CreateOrganization(ctx, &aaa.CreateOrganizationRequest{
    Name: "Green Valley FPO",
    Type: "FPO",
    OwnerId: ceoUserID,
})

// 2. Assign FPO manager role to CEO
roleResp := aaaClient.AssignRole(ctx, &aaa.AssignRoleRequest{
    UserId: ceoUserID,
    OrgId: orgResp.OrgId,
    RoleName: "fpo_manager",
})

// 3. Create farmers and assign farmer role
farmerUser := aaaClient.CreateUser(...)
aaaClient.AssignRole(ctx, &aaa.AssignRoleRequest{
    UserId: farmerUser.Id,
    OrgId: orgResp.OrgId,
    RoleName: "farmer",
})

// 4. Check permissions before operations
hasPermission := aaaClient.CheckPermission(ctx, &aaa.CheckPermissionRequest{
    UserId: currentUser,
    Resource: "farmers",
    Action: "create",
    OrgId: currentOrg,
})
```

## Success Criteria

1. ✅ All five gRPC services fully implemented and tested
2. ✅ Database seeding creates default roles and permissions
3. ✅ Farmers module can create FPOs and assign roles
4. ✅ Permission checks work correctly with inheritance
5. ✅ All operations are audited
6. ✅ Performance meets NFR requirements (< 100ms p95)
7. ✅ Integration tests pass with farmers-module patterns
8. ✅ golangci-lint passes with no errors
9. ✅ Code coverage > 80% for service layer

## Out of Scope

- UI/frontend changes (farmers-module handles this)
- Migration from existing data (fresh deployment)
- Advanced ABAC features (future enhancement)
- Multi-region deployment (future enhancement)
- GraphQL API (only gRPC required)

## Dependencies

- kisanlink-db v0.1.9+ (database operations)
- golang-jwt/jwt/v4 (JWT parsing)
- google.golang.org/grpc (gRPC framework)
- google.golang.org/protobuf (proto support)

## Timeline Estimate

- Phase 1: Proto definitions (0.5 day)
- Phase 2: Service layer implementation (2 days)
- Phase 3: gRPC handlers (1.5 days)
- Phase 4: Testing and validation (1 day)
- Total: ~5 days for full implementation
