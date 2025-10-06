# Farmers Module gRPC Integration - Tasks

## Task Breakdown

### Phase 1: Proto Definitions (0.5 day)

#### Task 1.1: Create RoleService Proto
**File**: `pkg/proto/role.proto`
**Estimate**: 1 hour

- [ ] Define message types for role operations
- [ ] Define RoleService with 5 RPCs
- [ ] Add proper imports and package declaration
- [ ] Generate Go code with `make proto`

**Deliverable**: Working role.proto with generated Go bindings

---

#### Task 1.2: Create PermissionService Proto
**File**: `pkg/proto/permission.proto`
**Estimate**: 1 hour

- [ ] Define message types for permission operations
- [ ] Define PermissionService with 5 RPCs
- [ ] Add proper imports and package declaration
- [ ] Generate Go code with `make proto`

**Deliverable**: Working permission.proto with generated Go bindings

---

#### Task 1.3: Enhance CatalogService Proto
**File**: `pkg/proto/catalog.proto`
**Estimate**: 1 hour

- [ ] Add SeedRolesAndPermissions RPC
- [ ] Ensure all CRUD operations for roles/permissions exist
- [ ] Update message types if needed
- [ ] Generate Go code with `make proto`

**Deliverable**: Enhanced catalog.proto with seeding support

---

#### Task 1.4: Update Proto Generation
**Files**: `Makefile`, `.proto` build config
**Estimate**: 30 minutes

- [ ] Ensure all protos are in proto generation pipeline
- [ ] Test proto generation: `make proto`
- [ ] Verify no compilation errors
- [ ] Commit generated pb.go files

**Deliverable**: All proto files generate without errors

---

### Phase 2: Service Layer Implementation (2 days)

#### Task 2.1: Implement OrganizationService
**Files**: `internal/services/organization/*.go`
**Estimate**: 4 hours

- [ ] Create service.go with struct and constructor
- [ ] Implement create.go for CreateOrganization
- [ ] Implement read.go for Get and List operations
- [ ] Implement update.go for UpdateOrganization
- [ ] Implement delete.go for DeleteOrganization
- [ ] Implement members.go for user operations

**Deliverable**: Full OrganizationService with 6 files, <300 lines each

---

#### Task 2.2: Implement GroupService
**Files**: `internal/services/group/*.go`
**Estimate**: 4 hours

- [ ] Create service.go with struct and constructor
- [ ] Implement create.go for CreateGroup
- [ ] Implement read.go for Get and List operations
- [ ] Implement members.go for membership operations
- [ ] Implement inheritance.go for group hierarchy
- [ ] Implement update.go for Update and Delete

**Deliverable**: Full GroupService with 6 files, <300 lines each

---

#### Task 2.3: Implement RoleService
**Files**: `internal/services/role/*.go`
**Estimate**: 3 hours

- [ ] Create service.go with struct and constructor
- [ ] Implement assign.go for AssignRole
- [ ] Implement check.go for role checking
- [ ] Implement remove.go for RemoveRole
- [ ] Implement list.go for listing operations

**Deliverable**: Full RoleService with 5 files, <300 lines each

---

#### Task 2.4: Implement PermissionService
**Files**: `internal/services/permission/*.go`
**Estimate**: 3 hours

- [ ] Create service.go with struct and constructor
- [ ] Implement assign.go for permission assignment
- [ ] Implement check.go for permission checking
- [ ] Implement effective.go for effective permissions
- [ ] Implement list.go for listing operations

**Deliverable**: Full PermissionService with 4-5 files, <300 lines each

---

#### Task 2.5: Implement CatalogService
**Files**: `internal/services/catalog/*.go`
**Estimate**: 2 hours

- [ ] Create service.go with struct and constructor
- [ ] Implement seed.go for seeding default data
- [ ] Implement roles.go for role CRUD
- [ ] Implement permissions.go for permission CRUD

**Deliverable**: Full CatalogService with 4 files, <300 lines each

---

### Phase 3: gRPC Handler Implementation (1.5 days)

#### Task 3.1: Implement OrganizationHandler
**File**: `internal/grpc_server/organization_handler.go`
**Estimate**: 2 hours

- [ ] Implement all 8 RPCs from OrganizationService
- [ ] Add request validation
- [ ] Add error mapping

**Deliverable**: organization_handler.go (<300 lines)

---

#### Task 3.2: Implement GroupHandler
**File**: `internal/grpc_server/group_handler.go`
**Estimate**: 2 hours

- [ ] Implement all 10 RPCs from GroupService
- [ ] Add request validation
- [ ] Add error mapping

**Deliverable**: group_handler.go (<300 lines)

---

#### Task 3.3: Implement RoleHandler
**File**: `internal/grpc_server/role_handler.go`
**Estimate**: 1.5 hours

- [ ] Implement all 5 RPCs from RoleService
- [ ] Add request validation
- [ ] Add error mapping

**Deliverable**: role_handler.go (<250 lines)

---

#### Task 3.4: Implement PermissionHandler
**File**: `internal/grpc_server/permission_handler.go`
**Estimate**: 1.5 hours

- [ ] Implement all 5 RPCs from PermissionService
- [ ] Add request validation
- [ ] Add error mapping

**Deliverable**: permission_handler.go (<250 lines)

---

#### Task 3.5: Implement CatalogHandler
**File**: `internal/grpc_server/catalog_handler.go`
**Estimate**: 1.5 hours

- [ ] Implement all RPCs from CatalogService
- [ ] Add admin authorization checks
- [ ] Add error mapping

**Deliverable**: catalog_handler.go (<300 lines)

---

#### Task 3.6: Update gRPC Server Registration
**File**: `internal/grpc_server/grpc_server.go`
**Estimate**: 30 minutes

- [ ] Register all 5 new services
- [ ] Update constructor

**Deliverable**: Updated grpc_server.go

---

### Phase 4: Testing and Validation (1 day)

#### Task 4.1: Write Unit Tests
**Estimate**: 3 hours

- [ ] Test OrganizationService
- [ ] Test GroupService
- [ ] Test RoleService
- [ ] Test PermissionService
- [ ] Test CatalogService

**Deliverable**: Test coverage > 80%

---

#### Task 4.2: Write Integration Tests
**Estimate**: 2 hours

- [ ] Test organization flow
- [ ] Test permission inheritance
- [ ] Test role assignment flow
- [ ] Test farmers-module simulation

**Deliverable**: Integration tests passing

---

#### Task 4.3: Performance Testing
**Estimate**: 1 hour

- [ ] Benchmark permission checks
- [ ] Benchmark role checks
- [ ] Test concurrency
- [ ] Verify cache performance

**Deliverable**: Performance benchmarks meeting NFRs

---

#### Task 4.4: Code Quality
**Estimate**: 1 hour

- [ ] Run `make lint`
- [ ] Run `make format`
- [ ] Fix all linting errors
- [ ] Verify file size limits

**Deliverable**: golangci-lint passes

---

## Task Summary

| Phase | Tasks | Estimate |
|-------|-------|----------|
| Phase 1: Proto Definitions | 4 | 0.5 day |
| Phase 2: Service Layer | 5 | 2 days |
| Phase 3: gRPC Handlers | 6 | 1.5 days |
| Phase 4: Testing | 4 | 1 day |
| **Total** | **19** | **5 days** |

## Success Criteria

- [ ] All 5 gRPC services working
- [ ] Proto files compile
- [ ] Default roles seeded
- [ ] Tests pass (>80% coverage)
- [ ] Performance meets NFRs
- [ ] Lint passes
- [ ] All files <300 lines
