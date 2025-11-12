# Authorization Migration Guide

## Overview

This guide provides step-by-step instructions for migrating from the current broken authorization logic to the fixed implementation that properly handles service-based role seeding.

## Migration Option 1: Direct Replacement (Recommended)

Replace the existing `authorization.go` with the refactored implementation. This is the cleanest approach.

### Steps:

1. **Backup existing file:**
```bash
cp internal/grpc_server/authorization.go internal/grpc_server/authorization.go.backup
```

2. **Copy new files into place:**
```bash
# These files are already created in the grpc_server directory:
# - principal.go
# - service_authorization.go
# - user_authorization.go
# - authorization_refactored.go
```

3. **Replace authorization.go with new implementation:**
```go
// internal/grpc_server/authorization.go
package grpc_server

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/services"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthorizationChecker provides authorization checking for gRPC handlers
type AuthorizationChecker struct {
	authzService *services.AuthorizationService
	logger       *zap.Logger
}

// NewAuthorizationChecker creates a new authorization checker
func NewAuthorizationChecker(
	authzService *services.AuthorizationService,
	logger *zap.Logger,
) *AuthorizationChecker {
	return &AuthorizationChecker{
		authzService: authzService,
		logger:       logger,
	}
}

// CheckSeedPermission validates authorization for role seeding operations
//
// FIXED IMPLEMENTATION - Key Changes:
// 1. Services bypass catalog:seed permission check
// 2. Use service_name (not service_id) for ownership comparison
// 3. Clear separation between service and user authorization paths
//
// Authorization Rules:
// - Services: Can only seed their own roles (name must match)
// - Users: Need catalog:seed + admin:* for service-specific seeding
func (ac *AuthorizationChecker) CheckSeedPermission(
	ctx context.Context,
	targetServiceID string,
) error {
	// Extract comprehensive principal information
	principal, err := ac.extractEnhancedPrincipal(ctx)
	if err != nil {
		ac.logger.Error("Failed to extract principal from context",
			zap.Error(err),
			zap.String("target_service_id", targetServiceID))
		return status.Errorf(codes.Unauthenticated,
			"authentication required: %v", err)
	}

	// Log the authorization request
	ac.logger.Info("Seed permission check initiated",
		zap.String("principal_id", principal.ID),
		zap.String("principal_name", principal.Name),
		zap.String("principal_type", principal.Type),
		zap.String("target_service_id", targetServiceID))

	// CRITICAL FIX: Route based on principal type
	// Services bypass permission checks and use ownership validation only
	if principal.Type == "service" {
		return ac.checkServiceSeedAuthorization(ctx, principal, targetServiceID)
	}

	// Users follow traditional permission-based authorization
	return ac.checkUserSeedAuthorization(ctx, principal, targetServiceID)
}

// checkServiceOwnership validates service-specific seed authorization
// This method is kept for backward compatibility but now delegates to the new implementation
func (ac *AuthorizationChecker) checkServiceOwnership(
	ctx context.Context,
	principalID string,
	principalType string,
	targetServiceID string,
) error {
	if principalType == "service" {
		// Extract service name (the fix: use name, not ID)
		serviceName := ac.getContextValue(ctx, "service_name")

		principal := &Principal{
			ID:   principalID,
			Name: serviceName,
			Type: "service",
		}

		return ac.checkServiceSeedAuthorization(ctx, principal, targetServiceID)
	}

	// For users, check admin permissions
	principal := &Principal{
		ID:   principalID,
		Type: "user",
		Name: "",
	}

	return ac.checkUserAdminPermission(ctx, principal, targetServiceID)
}

// extractPrincipal maintains backward compatibility
// Delegates to the enhanced version
func (ac *AuthorizationChecker) extractPrincipal(
	ctx context.Context,
) (string, string, error) {
	principal, err := ac.extractEnhancedPrincipal(ctx)
	if err != nil {
		return "", "", err
	}
	return principal.ID, principal.Type, nil
}

// getContextValue safely extracts a string value from context
func (ac *AuthorizationChecker) getContextValue(
	ctx context.Context,
	key string,
) string {
	if val := ctx.Value(key); val != nil {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}
```

## Migration Option 2: Minimal Changes (Quick Fix)

If you need a minimal change to fix the immediate issue:

### Change 1: Fix the comparison (Line 111-112)
```go
// OLD (BROKEN):
serviceID := ac.getContextValue(ctx, "service_id")
if serviceID != targetServiceID {

// NEW (FIXED):
serviceName := ac.getContextValue(ctx, "service_name")
if serviceName != targetServiceID {
```

### Change 2: Skip permission check for services (Line 50-75)
```go
// Add this BEFORE the permission check:
if principalType == "service" {
    // Services don't need catalog:seed permission
    // Skip directly to ownership check
    if targetServiceID == "" || targetServiceID == "farmers-module" {
        return status.Errorf(codes.PermissionDenied,
            "services cannot seed default farmers-module roles")
    }
    return ac.checkServiceOwnership(ctx, principalID, principalType, targetServiceID)
}

// Continue with existing permission check for users only...
```

## Testing the Fix

### Test Case 1: Service Self-Seeding
```bash
# Service 'erp-module' seeding its own roles
grpcurl -H "x-api-key: ${ERP_API_KEY}" \
  -d '{"service_id": "erp-module", "force": false}' \
  localhost:50051 catalog.CatalogService/SeedRolesAndPermissions

# Expected: SUCCESS
```

### Test Case 2: Service Cross-Seeding (Should Fail)
```bash
# Service 'erp-module' trying to seed 'crm-module' roles
grpcurl -H "x-api-key: ${ERP_API_KEY}" \
  -d '{"service_id": "crm-module", "force": false}' \
  localhost:50051 catalog.CatalogService/SeedRolesAndPermissions

# Expected: PERMISSION_DENIED
```

### Test Case 3: User Seeding Default
```bash
# User with catalog:seed permission seeding farmers-module
curl -X POST http://localhost:8080/api/v1/catalog/seed \
  -H "Authorization: Bearer ${USER_TOKEN}" \
  -d '{"service_id": "", "force": false}'

# Expected: SUCCESS
```

### Test Case 4: Admin Seeding Any Service
```bash
# Admin user seeding specific service
curl -X POST http://localhost:8080/api/v1/catalog/seed \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -d '{"service_id": "erp-module", "force": false}'

# Expected: SUCCESS
```

## Verification Checklist

- [ ] Services can seed their own roles using API key only
- [ ] Services cannot seed other services' roles
- [ ] Services cannot seed farmers-module
- [ ] Users with catalog:seed can seed farmers-module
- [ ] Admins can seed any service's roles
- [ ] Proper audit logs are generated
- [ ] No breaking changes to existing user flows

## Rollback Procedure

If issues are encountered:

1. **Restore backup:**
```bash
cp internal/grpc_server/authorization.go.backup internal/grpc_server/authorization.go
```

2. **Remove new files (if using Option 1):**
```bash
rm internal/grpc_server/principal.go
rm internal/grpc_server/service_authorization.go
rm internal/grpc_server/user_authorization.go
```

3. **Restart service:**
```bash
make run
```

## Production Deployment

1. **Deploy to staging first**
2. **Run all test cases**
3. **Monitor logs for any authorization errors**
4. **Deploy to production during maintenance window**
5. **Keep old version ready for quick rollback**

## Monitoring

Watch for these log patterns after deployment:

### Success Patterns:
```
"Service authorized to seed its own roles" service_name=erp-module
"User authorized to seed default/farmers-module roles" user_id=USR001
```

### Error Patterns to Monitor:
```
"Service attempted cross-service role seeding"
"Service attempted to seed default/farmers-module roles"
"Failed to extract principal from context"
```

## Support

If you encounter issues:
1. Check the logs for detailed error messages
2. Verify the context values are being set correctly in middleware
3. Ensure service names match exactly (case-sensitive)
4. Review the comprehensive design document in `.kiro/specs/service-rbac-authorization-fix.md`
