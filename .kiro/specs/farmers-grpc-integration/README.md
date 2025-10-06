# Farmers Module gRPC Integration

## Overview

This specification defines the implementation of five critical gRPC services for AAA service to support farmers-module integration:

1. **OrganizationService** - Multi-tenant organization management for FPOs/cooperatives
2. **GroupService** - Hierarchical group management within organizations
3. **RoleService** - Role assignment and checking for users
4. **PermissionService** - Fine-grained permission management for groups
5. **CatalogService** - Role and permission catalog with seeding

## Why This Implementation?

The farmers-module manages FPOs (Farmer Producer Organizations), farmers, farms, and field agents (KisanSathi). It requires robust multi-tenant authentication and authorization with hierarchical permissions. This implementation provides:

- ✅ **Multi-tenancy** - Isolated FPO organizations
- ✅ **Hierarchical RBAC** - Roles inherit through groups
- ✅ **Fine-grained permissions** - Resource-level access control (farmers:create, farms:read, etc.)
- ✅ **Default roles** - admin, fpo_manager, kisansathi, farmer, readonly
- ✅ **High performance** - Caching for <100ms p95 latency
- ✅ **Audit logging** - Complete trail of all permission changes

## Current State

### ✅ Already Implemented
- Database schema (all tables exist in migration)
- Proto definitions for Organization and Group services
- gRPC server infrastructure
- kisanlink-db integration

### ❌ Needs Implementation
- RoleService and PermissionService proto definitions
- Service layer for all five services (business logic)
- gRPC handlers for all services
- Seeding of default roles and permissions
- Comprehensive tests

## Documents

1. **[requirements.md](./requirements.md)** - Detailed functional and non-functional requirements
2. **[design.md](./design.md)** - Architecture, data models, implementation patterns
3. **[tasks.md](./tasks.md)** - Phase-by-phase task breakdown with estimates

## Quick Start

### Phase 1: Proto Definitions (0.5 day)
Create role.proto, permission.proto, enhance catalog.proto

### Phase 2: Service Layer (2 days)
Implement business logic in `internal/services/`:
- organization/ (6 files)
- group/ (6 files)
- role/ (5 files)
- permission/ (4 files)
- catalog/ (4 files)

### Phase 3: gRPC Handlers (1.5 days)
Implement gRPC handlers in `internal/grpc_server/`:
- organization_handler.go
- group_handler.go
- role_handler.go
- permission_handler.go
- catalog_handler.go

### Phase 4: Testing (1 day)
- Unit tests (80% coverage)
- Integration tests
- Performance benchmarks
- Lint fixes

**Total Estimate**: 5 days

## Default Roles & Permissions

### Roles
- **admin** - Full access to all resources
- **fpo_manager** - Manage FPO and members
- **kisansathi** - Field agent managing farmers
- **farmer** - Self-service access
- **readonly** - Read-only access

### Permission Format
`resource:action[:scope]`

Examples:
- `farmers:create` - Create any farmer
- `farmers:read:self` - Read own record
- `farms:update` - Update any farm

## Integration Example

```go
// Farmers module creates FPO
orgResp := aaaClient.CreateOrganization(ctx, &aaa.CreateOrganizationRequest{
    Name: "Green Valley FPO",
    Type: "FPO",
    OwnerId: ceoUserID,
})

// Assign FPO manager role
aaaClient.AssignRole(ctx, &aaa.AssignRoleRequest{
    UserId: ceoUserID,
    OrgId: orgResp.OrgId,
    RoleName: "fpo_manager",
})

// Check permission before operation
hasPermission := aaaClient.CheckPermission(ctx, &aaa.CheckPermissionRequest{
    UserId: currentUser,
    Resource: "farmers",
    Action: "create",
})
```

## Success Criteria

- [ ] All 5 gRPC services implemented and tested
- [ ] Default roles and permissions seeded
- [ ] Permission checks work with inheritance
- [ ] Performance meets NFRs (p95 < 100ms)
- [ ] Integration tests pass
- [ ] golangci-lint passes
- [ ] Code coverage > 80%

## Architecture Principles

- **Clean Architecture** - Handlers → Services → Database
- **File Size Limit** - Maximum 300 lines per file
- **Repository Pattern** - Use kisanlink-db for all database ops
- **Single Responsibility** - Each file has one focused purpose
- **Caching** - Redis caching for performance-critical operations

## Technology Stack

- Go 1.24+
- gRPC + Protocol Buffers
- PostgreSQL 17+ (via kisanlink-db)
- Redis 7+ (caching)
- GORM (ORM)

## Next Steps

1. Review requirements.md for detailed functional requirements
2. Review design.md for architecture and implementation patterns
3. Follow tasks.md for step-by-step implementation
4. Start with Phase 1: Proto Definitions

## Questions?

Refer to:
- `.kiro/steering/product.md` - Business context
- `.kiro/steering/tech.md` - Development guidelines
- Existing proto files in `pkg/proto/` for reference
