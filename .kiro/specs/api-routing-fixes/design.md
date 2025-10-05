# Design Document

## Overview

The AAA service has a routing architecture where individual route files define endpoints but are not being registered in the main setup. This design addresses the systematic fix needed to ensure all API routes are properly registered and accessible, with proper service dependency injection.

## Architecture

### Current State Analysis

The current routing setup has these components:

- `internal/routes/setup.go` - Main route registration function
- `internal/routes/organization_routes.go` - Organization route definitions (not registered)
- `internal/routes/permission_routes.go` - Permission route definitions (partially registered)
- `cmd/server/main.go` - Server initialization with service setup

### Root Cause

1. **Missing Route Registration**: Organization routes are defined but `SetupOrganizationRoutes` is never called
2. **Missing Service Dependencies**: Organization routes require `OrganizationService` and `GroupService` which are not initialized
3. **Incomplete Permission Setup**: Permission routes are registered but may have dependency issues

## Components and Interfaces

### Service Dependencies

The organization handler requires:

```go
type Handler struct {
    orgService   interfaces.OrganizationService
    groupService interfaces.GroupService
    logger       *zap.Logger
    responder    interfaces.Responder
}
```

### Route Registration Flow

```
main.go
├── initializeServer()
├── setupRoutesAndDocs()
├── routes.SetupAAAWithAdmin()
└── routes.SetupAAA() // Missing organization routes
```

## Data Models

### Service Interfaces Required

1. **OrganizationService Interface**

   - CreateOrganization
   - GetOrganization
   - UpdateOrganization
   - DeleteOrganization
   - ListOrganizations
   - GetOrganizationHierarchy
   - ActivateOrganization
   - DeactivateOrganization
   - GetOrganizationStats

2. **GroupService Interface**
   - CreateGroup
   - GetGroup
   - ListGroups
   - AddMemberToGroup
   - RemoveMemberFromGroup

## Error Handling

### Service Initialization Errors

- Check if required services exist before route registration
- Provide clear error messages for missing dependencies
- Graceful degradation when optional services are unavailable

### Route Registration Errors

- Validate handler dependencies during setup
- Log missing service warnings
- Prevent server startup if critical routes cannot be registered

## Testing Strategy

### Integration Tests

1. Test all organization endpoints return proper responses (not 404)
2. Test permission endpoints are accessible
3. Test service dependency injection works correctly

### Unit Tests

1. Test route registration functions
2. Test service initialization
3. Test error handling for missing dependencies

## Implementation Plan

### Phase 1: Service Discovery and Creation

1. Identify existing organization and group services
2. Create missing service implementations if needed
3. Update service interfaces as required

### Phase 2: Route Registration Fix

1. Add organization route registration to main setup
2. Ensure proper service dependency injection
3. Add error handling for missing services

### Phase 3: Validation and Testing

1. Test all endpoints return proper responses
2. Verify authentication and authorization work
3. Add integration tests for the fixed routes

## Design Decisions

### Service Initialization Strategy

- **Decision**: Initialize services in main.go and pass to route setup
- **Rationale**: Centralized dependency management, easier testing
- **Alternative**: Service locator pattern (rejected due to complexity)

### Error Handling Approach

- **Decision**: Fail fast on missing critical services, warn on optional services
- **Rationale**: Clear feedback to developers, prevents runtime surprises
- **Alternative**: Runtime service discovery (rejected due to performance)

### Route Organization

- **Decision**: Keep existing file structure, fix registration only
- **Rationale**: Minimal changes, maintains existing patterns
- **Alternative**: Restructure all routes (rejected due to scope)
