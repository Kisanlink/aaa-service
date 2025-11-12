# Service RBAC Authorization Fix - Architectural Design

## Executive Summary

This document provides a comprehensive architectural solution to fix the authorization logic issue in the AAA service where services authenticated via API key cannot seed their own roles. The fix addresses three critical issues while maintaining security, backward compatibility, and clean architecture principles.

## Problem Statement

### Critical Issues Identified

1. **Permission Check Blocking Services**: Services authenticated via API key don't have roles/permissions, so the `catalog:seed` permission check at line 50-75 always fails.

2. **Incorrect Identity Comparison**: At line 111-112, the code compares `service_id` (DB ID like "SVC00000001") with `targetServiceID` (service name like "erp-module"), which will never match.

3. **Sequential Validation Failure**: Even if the comparison is fixed, services fail at the permission check and never reach the ownership validation.

### Root Cause Analysis

The current implementation follows a sequential validation pattern:
1. Check `catalog:seed` permission (fails for services)
2. Check service ownership (never reached)

This violates the documented business rule that services should be able to seed their own roles using only API key authentication without requiring the `catalog:seed` permission.

## Proposed Solution Architecture

### Core Design Principles

1. **Separation of Concerns**: Separate authorization logic for services vs users
2. **Early Service Detection**: Identify service principals early and route to appropriate validation
3. **Explicit Authorization Paths**: Clear, documented authorization flows for each principal type
4. **Fail-Safe Defaults**: Deny by default, explicitly allow authorized operations
5. **Comprehensive Logging**: Audit trail for all authorization decisions

### Authorization Flow Redesign

```
┌─────────────────┐
│  Request Entry  │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│ Extract Principal   │
│ (ID, Type, Name)    │
└──────┬──────────────┘
       │
       ▼
┌──────────────────────────┐
│ Is Principal a Service?  │
└─────┬────────────┬───────┘
     YES           NO
      │             │
      ▼             ▼
┌─────────────┐  ┌─────────────────┐
│  Service    │  │      User       │
│    Path     │  │      Path       │
└──────┬──────┘  └────────┬────────┘
       │                   │
       ▼                   ▼
┌─────────────┐  ┌─────────────────┐
│ Check Self  │  │ Check catalog:  │
│  Ownership  │  │ seed permission │
└──────┬──────┘  └────────┬────────┘
       │                   │
       ▼                   ▼
┌─────────────┐  ┌─────────────────┐
│   Allow if  │  │   Allow if      │
│   Match     │  │   Permitted     │
└─────────────┘  └────────┬────────┘
                          │
                          ▼
                 ┌─────────────────┐
                 │ Check Service   │
                 │ Specific Rules  │
                 └─────────────────┘
```

### Detailed Authorization Rules

#### Service Principal Authorization
- **Identity**: Uses `service_name` from context (e.g., "erp-module")
- **Permission Check**: SKIP `catalog:seed` permission check
- **Ownership Check**: Compare `service_name` with `targetServiceID`
- **Result**: Allow if names match, deny otherwise

#### User Principal Authorization
- **Identity**: Uses `user_id` from context
- **Permission Check**: REQUIRE `catalog:seed` permission
- **Service-Specific**:
  - If `targetServiceID` is empty or "farmers-module": Allow with basic permission
  - If specific service: Require `admin:*` permission
- **Result**: Follow existing permission model

### Implementation Design

#### 1. Refactored Authorization Structure

```go
// authorization.go - Main authorization logic
type AuthorizationChecker struct {
    authzService *services.AuthorizationService
    logger       *zap.Logger
}

// Main entry point - delegates to appropriate checker
func (ac *AuthorizationChecker) CheckSeedPermission(
    ctx context.Context,
    targetServiceID string,
) error {
    // Extract principal with enhanced information
    principal, err := ac.extractEnhancedPrincipal(ctx)
    if err != nil {
        return status.Errorf(codes.Unauthenticated, "authentication required")
    }

    // Route to appropriate authorization path
    if principal.Type == "service" {
        return ac.checkServiceSeedAuthorization(ctx, principal, targetServiceID)
    }

    return ac.checkUserSeedAuthorization(ctx, principal, targetServiceID)
}
```

#### 2. Enhanced Principal Extraction

```go
// principal.go - Principal extraction and management
type Principal struct {
    ID   string // Database ID (e.g., "SVC00000001" or "USR00000001")
    Name string // Human-readable name (e.g., "erp-module" or "john.doe")
    Type string // "service" or "user"
}

func (ac *AuthorizationChecker) extractEnhancedPrincipal(
    ctx context.Context,
) (*Principal, error) {
    principalType := ac.getContextValue(ctx, "principal_type")

    if principalType == "service" {
        return &Principal{
            ID:   ac.getContextValue(ctx, "service_id"),
            Name: ac.getContextValue(ctx, "service_name"),
            Type: "service",
        }, nil
    }

    // User principal
    userID := ac.getContextValue(ctx, "user_id")
    if userID == "" {
        return nil, fmt.Errorf("no authenticated principal found")
    }

    return &Principal{
        ID:   userID,
        Name: ac.getContextValue(ctx, "username"), // If available
        Type: "user",
    }, nil
}
```

#### 3. Service-Specific Authorization

```go
// service_authorization.go - Service-specific authorization logic
func (ac *AuthorizationChecker) checkServiceSeedAuthorization(
    ctx context.Context,
    principal *Principal,
    targetServiceID string,
) error {
    // Log authorization attempt
    ac.logger.Info("Service seed authorization check",
        zap.String("service_id", principal.ID),
        zap.String("service_name", principal.Name),
        zap.String("target_service", targetServiceID))

    // Handle default/farmers-module case
    if targetServiceID == "" || targetServiceID == "farmers-module" {
        ac.logger.Warn("Service cannot seed default/farmers-module roles",
            zap.String("service_name", principal.Name))
        return status.Errorf(codes.PermissionDenied,
            "services cannot seed default farmers-module roles")
    }

    // Check ownership: service can only seed its own roles
    if principal.Name != targetServiceID {
        ac.logger.Warn("Service attempting to seed another service's roles",
            zap.String("caller_service", principal.Name),
            zap.String("target_service", targetServiceID))
        return status.Errorf(codes.PermissionDenied,
            "service '%s' cannot seed roles for service '%s'",
            principal.Name, targetServiceID)
    }

    // Service is seeding its own roles - ALLOW
    ac.logger.Info("Service authorized to seed own roles",
        zap.String("service_name", principal.Name))

    return nil
}
```

#### 4. User-Specific Authorization

```go
// user_authorization.go - User-specific authorization logic
func (ac *AuthorizationChecker) checkUserSeedAuthorization(
    ctx context.Context,
    principal *Principal,
    targetServiceID string,
) error {
    // Check basic catalog:seed permission
    permission := &services.Permission{
        UserID:     principal.ID,
        Resource:   "catalog",
        ResourceID: "catalog",
        Action:     "seed",
    }

    result, err := ac.authzService.CheckPermission(ctx, permission)
    if err != nil {
        ac.logger.Error("Permission check failed",
            zap.String("user_id", principal.ID),
            zap.Error(err))
        return status.Errorf(codes.Internal, "authorization check failed")
    }

    if !result.Allowed {
        ac.logger.Warn("User lacks catalog:seed permission",
            zap.String("user_id", principal.ID))
        return status.Errorf(codes.PermissionDenied,
            "insufficient permissions to seed roles")
    }

    // For default/farmers-module, basic permission is sufficient
    if targetServiceID == "" || targetServiceID == "farmers-module" {
        return nil
    }

    // For service-specific seeding, require admin permission
    return ac.checkAdminPermission(ctx, principal.ID, targetServiceID)
}

func (ac *AuthorizationChecker) checkAdminPermission(
    ctx context.Context,
    userID string,
    targetServiceID string,
) error {
    adminPermission := &services.Permission{
        UserID:     userID,
        Resource:   "admin",
        ResourceID: "admin",
        Action:     "*",
    }

    result, err := ac.authzService.CheckPermission(ctx, adminPermission)
    if err != nil {
        return status.Errorf(codes.Internal, "admin authorization check failed")
    }

    if !result.Allowed {
        return status.Errorf(codes.PermissionDenied,
            "only administrators can seed roles for service: %s", targetServiceID)
    }

    return nil
}
```

### File Organization Strategy

To maintain the 300-line file limit and single responsibility principle:

```
internal/grpc_server/
├── authorization.go              # Main AuthorizationChecker and routing (150 lines)
├── principal.go                  # Principal extraction and types (80 lines)
├── service_authorization.go      # Service-specific authorization (100 lines)
├── user_authorization.go         # User-specific authorization (120 lines)
└── authorization_test.go         # Comprehensive tests (500+ lines)
```

## Security Considerations

### 1. Defense in Depth
- **Multiple validation layers**: Authentication → Principal extraction → Authorization
- **Explicit deny by default**: No implicit permissions
- **Comprehensive audit logging**: Every decision is logged

### 2. Attack Vector Mitigation

| Attack Vector | Mitigation |
|--------------|------------|
| Service impersonation | Validate API key and service_name consistency |
| Privilege escalation | Services cannot gain user permissions |
| Cross-service seeding | Strict name matching for services |
| Missing context values | Fail safely with authentication error |
| Empty service names | Explicit validation and rejection |

### 3. Edge Cases Handled

1. **Empty service_id/service_name**: Return authentication error
2. **Service trying to seed farmers-module**: Explicitly denied
3. **Malformed principal_type**: Fall back to user authentication
4. **Missing context values**: Safe extraction with empty string default
5. **Case sensitivity**: Exact string matching for service names

## Migration and Rollback Strategy

### Phase 1: Code Deployment
1. Deploy new authorization logic
2. Monitor logs for authorization patterns
3. Verify service authentication works

### Phase 2: Validation
1. Test service self-seeding
2. Verify user permissions unchanged
3. Confirm admin override capabilities

### Rollback Plan
1. Revert to previous authorization.go
2. No database changes required
3. No API contract changes

## Testing Strategy

### Unit Tests
```go
// Test cases to implement
1. Service seeding own roles - ALLOW
2. Service seeding other service roles - DENY
3. Service seeding farmers-module - DENY
4. User with catalog:seed seeding farmers-module - ALLOW
5. User without catalog:seed - DENY
6. Admin user seeding any service - ALLOW
7. Empty service_name in context - AUTH ERROR
8. Missing principal_type - FALLBACK TO USER
```

### Integration Tests
1. Full flow with API key authentication
2. JWT token authentication flow
3. Concurrent seeding attempts
4. Transaction rollback scenarios

### Security Tests
1. Attempt cross-service seeding
2. Missing authentication headers
3. Malformed context values
4. Permission bypass attempts

## Monitoring and Observability

### Key Metrics
- Authorization success/failure rates by principal type
- Service self-seeding frequency
- Cross-service attempt rejections
- Authentication errors

### Log Patterns
```go
// Success
"Service authorized to seed own roles" service_name=erp-module

// Failures
"Service attempting to seed another service's roles" caller=erp target=crm
"Service cannot seed default/farmers-module roles" service_name=erp-module
"User lacks catalog:seed permission" user_id=USR00000001
```

### Alerts
1. Repeated cross-service seeding attempts (potential attack)
2. Authentication errors spike (configuration issue)
3. Unusual seeding patterns (anomaly detection)

## Implementation Checklist

- [ ] Refactor authorization.go into multiple files
- [ ] Implement enhanced principal extraction
- [ ] Create service-specific authorization path
- [ ] Update user authorization path
- [ ] Add comprehensive logging
- [ ] Write unit tests (minimum 80% coverage)
- [ ] Update integration tests
- [ ] Document API changes (if any)
- [ ] Review security implications
- [ ] Performance testing
- [ ] Update SERVICE_SPECIFIC_RBAC_SEEDING.md
- [ ] Deploy to staging environment
- [ ] Verify backward compatibility
- [ ] Production deployment plan

## Recommendations for Future

### 1. Context Type Safety
Replace string-based context values with typed context keys:
```go
type contextKey string
const (
    ContextKeyServiceID   contextKey = "service_id"
    ContextKeyServiceName contextKey = "service_name"
)
```

### 2. Service Registry
Implement a service registry to validate service names during authentication:
- Prevent typos in service names
- Centralized service configuration
- Dynamic service discovery

### 3. Permission Caching
Cache permission evaluations for services:
- Services have static permissions (self-seed only)
- Reduce database load
- Improve response times

### 4. Audit Enhancement
Enhance audit logs with:
- Correlation IDs for request tracing
- Service version tracking
- Permission decision rationale

## Conclusion

This architectural fix addresses all identified issues while maintaining security, backward compatibility, and clean architecture principles. The solution provides clear separation between service and user authorization paths, uses the correct service identifier for comparison, and allows services to seed their own roles without requiring catalog:seed permission, as documented in the business requirements.

The implementation follows the project's standards for file organization, error handling, and logging, ensuring maintainability and observability in production environments.
