# AAA Service V2 - Implementation Plan

## Overview
Complete redesign of the AAA (Authentication, Authorization, and Audit) service to support:
- **Groups** that can own roles/policies and inherit from other groups
- **Reusable roles** with global and organization scope
- **RBAC** (Role-Based Access Control) using Zanzibar/SpiceDB
- **ABAC** (Attribute-Based Access Control) via caveats
- **Resource Registry** with hierarchy and column-level authorization
- **Dynamic actions** via service contracts
- **Consistency modes** (eventual vs strict)
- **Replayable audit** with blockchain-style hash chain

## Implementation Status

### ✅ Completed
- **M1.1 - Core Database Schema**
  - 13 new models created
  - Support for organizations, groups, principals, bindings, attributes, events
  - Column-level authorization with bitmap operations
  - Immutable event chain with SHA256 hashing

### 🚧 In Progress
- **M1.2 - gRPC + Gateway Setup**
  - Define gRPC services
  - Configure HTTP gateway
  - Seed static actions

### 📋 Planned
- M1.3 to M1.5 - Complete V1 Foundations
- V2 - ABAC & Column authorization
- V3 - Contracts & Consistency
- V4 - Versioning & Replay
- V5 - Admin UI

## Architecture Highlights

### Data Model
```
Organizations
    └── Groups (with inheritance)
         └── Members (users/services, time-bounded)
         └── Roles (owned by groups)
              └── Permissions
                   └── Bindings (with caveats)
```

### Authorization Flow
1. **Subject** (User/Service/Group) requests access
2. **Binding** links subject to role/permission on resource
3. **Caveats** apply constraints (time, attributes, columns)
4. **SpiceDB** evaluates with low latency
5. **Audit Event** records decision with hash chain

### Key Features

#### Groups & Inheritance
- Groups can own roles and policies
- Nested group membership with time bounds
- Inheritance follows time windows

#### Enhanced Roles
- **Scope**: GLOBAL or ORG-specific
- **Versioning**: Track changes, support rollback
- **Hierarchy**: Roles can inherit from parents

#### Flexible Bindings
- Connect any subject to any role/permission
- JSONB caveats for extensible constraints:
  ```json
  {
    "starts_at": "2024-01-01T00:00:00Z",
    "ends_at": "2024-12-31T23:59:59Z",
    "required_attributes": {
      "department": "engineering",
      "clearance_level": 3
    },
    "column_groups": ["pii_basic", "financial_summary"]
  }
  ```

#### Column-Level Authorization
- Define column groups (e.g., PII, financial, medical)
- Bitmap operations for efficient permission checks
- Scale to hundreds of columns per table

#### Audit Chain
- Immutable events with hash linking
- Blockchain-style integrity verification
- Periodic Merkle tree checkpoints
- Full replay capability

## Development Guidelines

### Adding New Features
1. Update models in `entities/models/`
2. Add to database migrations
3. Update SpiceDB schema if needed
4. Implement service layer
5. Add gRPC/HTTP endpoints
6. Write comprehensive tests

### Testing Strategy
- **Unit tests**: Model validation, business logic
- **Integration tests**: End-to-end flows with SpiceDB
- **Performance tests**: Check/CheckColumns at scale
- **Security tests**: Caveat evaluation, hash verification

### Performance Targets
- Check latency: < 15ms P95 at 2k QPS
- Column check: < 20ms for 100+ columns
- Tuple write: < 100ms with eventual consistency
- Strict mode: < 500ms with consistency guarantee

## Next Immediate Steps

### M1.2 - gRPC + Gateway (Current)
1. Define proto files for all services
2. Configure gRPC-gateway for HTTP/JSON
3. Implement service stubs
4. Create action seeding migration
5. Set up proto compilation pipeline

### M1.3 - Tuple Compiler
1. Implement role → relation mapping
2. Build caveat compiler
3. Create tuple writer service
4. Implement CreateBinding API
5. Add event emission

## Resources
- Design Document: See initial planning document
- Progress Tracking: `docs/V1_FOUNDATIONS_PROGRESS.md`
- SpiceDB Schema: `services/schema/spicedb_schema_v2.zed`
- Models: `entities/models/`

## Team Contacts
- **AAA TL**: Overall design and implementation
- **Platform TL**: SpiceDB infrastructure
- **Data TL**: Column-level authorization
- **Security TL**: Audit and consistency
