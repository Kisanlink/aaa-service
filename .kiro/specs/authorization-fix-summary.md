# Authorization Fix Summary - Production-Ready Solution

## Quick Fix for Existing authorization.go

Here are the **exact changes** needed to fix the authorization issue in `/Users/kaushik/aaa-service/internal/grpc_server/authorization.go`:

### Change #1: Early Service Detection (Line 37-96)

**REPLACE** the entire `CheckSeedPermission` method with:

```go
func (ac *AuthorizationChecker) CheckSeedPermission(ctx context.Context, targetServiceID string) error {
	// Extract principal information from context
	principalID, principalType, err := ac.extractPrincipal(ctx)
	if err != nil {
		ac.logger.Error("Failed to extract principal from context", zap.Error(err))
		return status.Errorf(codes.Unauthenticated, "authentication required")
	}

	ac.logger.Debug("Checking seed permission",
		zap.String("principal_id", principalID),
		zap.String("principal_type", principalType),
		zap.String("target_service_id", targetServiceID))

	// CRITICAL FIX: Services bypass permission check
	if principalType == "service" {
		// Extract service name (FIX: use service_name, not service_id)
		serviceName := ac.getContextValue(ctx, "service_name")
		if serviceName == "" {
			ac.logger.Error("Service name missing in context",
				zap.String("service_id", principalID))
			return status.Errorf(codes.Unauthenticated,
				"service authentication incomplete: service_name missing")
		}

		// Services cannot seed default/farmers-module
		if targetServiceID == "" || targetServiceID == "farmers-module" {
			ac.logger.Warn("Service cannot seed default/farmers-module roles",
				zap.String("service_name", serviceName))
			return status.Errorf(codes.PermissionDenied,
				"service '%s' cannot seed default farmers-module roles", serviceName)
		}

		// Check ownership: service can only seed its own roles
		if serviceName != targetServiceID {
			ac.logger.Warn("Service attempting to seed another service's roles",
				zap.String("caller_service", serviceName),
				zap.String("target_service", targetServiceID))
			return status.Errorf(codes.PermissionDenied,
				"service '%s' cannot seed roles for service '%s'",
				serviceName, targetServiceID)
		}

		// Service is authorized to seed its own roles
		ac.logger.Info("Service authorized to seed own roles",
			zap.String("service_id", principalID),
			zap.String("service_name", serviceName))
		return nil
	}

	// For users, check catalog:seed permission
	permission := &services.Permission{
		UserID:     principalID,
		Resource:   "catalog",
		ResourceID: "catalog",
		Action:     "seed",
	}

	result, err := ac.authzService.CheckPermission(ctx, permission)
	if err != nil {
		ac.logger.Error("Permission check failed",
			zap.String("principal_id", principalID),
			zap.Error(err))
		return status.Errorf(codes.Internal, "authorization check failed: %v", err)
	}

	if !result.Allowed {
		ac.logger.Warn("Seed permission denied - insufficient permissions",
			zap.String("principal_id", principalID),
			zap.String("reason", result.Reason))
		return status.Errorf(codes.PermissionDenied,
			"insufficient permissions to seed roles: %s", result.Reason)
	}

	// For default/farmers-module, basic permission is sufficient
	if targetServiceID == "" || targetServiceID == "farmers-module" {
		ac.logger.Debug("Seed permission granted for default/farmers-module",
			zap.String("principal_id", principalID))
		return nil
	}

	// For service-specific seeding, users need admin permission
	return ac.checkServiceOwnership(ctx, principalID, principalType, targetServiceID)
}
```

### Change #2: Fix checkServiceOwnership Method (Line 102-156)

**REPLACE** lines 111-112 (the comparison fix):

```go
// OLD (BROKEN - Line 111-112):
serviceID := ac.getContextValue(ctx, "service_id")
if serviceID != targetServiceID {

// NEW (FIXED):
serviceName := ac.getContextValue(ctx, "service_name")
if serviceName != targetServiceID {
```

And update the error message at lines 113-118:

```go
// OLD:
ac.logger.Warn("Service attempting to seed another service's roles",
    zap.String("caller_service_id", serviceID),
    zap.String("target_service_id", targetServiceID))
return status.Errorf(codes.PermissionDenied,
    "services can only seed their own roles (attempted to seed %s, but caller is %s)",
    targetServiceID, serviceID)

// NEW:
ac.logger.Warn("Service attempting to seed another service's roles",
    zap.String("caller_service_name", serviceName),
    zap.String("target_service_id", targetServiceID))
return status.Errorf(codes.PermissionDenied,
    "services can only seed their own roles (attempted to seed %s, but caller is %s)",
    targetServiceID, serviceName)
```

### Change #3: Enhance Principal Extraction (Line 162-180)

Add validation for service_name:

```go
func (ac *AuthorizationChecker) extractPrincipal(ctx context.Context) (string, string, error) {
	// Check if this is a service principal
	principalType := ac.getContextValue(ctx, "principal_type")
	if principalType == "service" {
		serviceID := ac.getContextValue(ctx, "service_id")
		serviceName := ac.getContextValue(ctx, "service_name")

		if serviceID == "" {
			return "", "", fmt.Errorf("service principal_type set but service_id missing")
		}
		// NEW: Also validate service_name
		if serviceName == "" {
			return "", "", fmt.Errorf("service principal_type set but service_name missing")
		}
		return serviceID, "service", nil
	}

	// Check if this is a user principal
	userID := ac.getContextValue(ctx, "user_id")
	if userID == "" {
		return "", "", fmt.Errorf("no authenticated principal found in context")
	}

	return userID, "user", nil
}
```

## Complete Fixed File

For reference, here are the complete fixed methods in order:

1. **CheckSeedPermission** - Routes services to ownership check, users to permission check
2. **checkServiceOwnership** - Uses service_name instead of service_id for comparison
3. **extractPrincipal** - Validates both service_id and service_name

## Verification Commands

After applying the fix, verify with these commands:

```bash
# 1. Run tests
go test ./internal/grpc_server -v -run TestAuthorization

# 2. Test service self-seeding (should work)
grpcurl -H "x-api-key: YOUR_SERVICE_API_KEY" \
  -d '{"service_id": "your-service-name"}' \
  localhost:50051 catalog.CatalogService/SeedRolesAndPermissions

# 3. Check logs for correct behavior
grep "Service authorized to seed own roles" /var/log/aaa-service.log
```

## Key Points

1. **Services bypass permission checks** - They don't need `catalog:seed` permission
2. **Use service_name for comparison** - Not service_id (DB ID vs service name)
3. **Services can only seed their own roles** - Strict ownership validation
4. **Backward compatible** - User flows unchanged

## Files Created

The following files contain the complete refactored solution:

- `/Users/kaushik/aaa-service/.kiro/specs/service-rbac-authorization-fix.md` - Full architectural design
- `/Users/kaushik/aaa-service/.kiro/specs/authorization-migration-guide.md` - Migration instructions
- `/Users/kaushik/aaa-service/internal/grpc_server/principal.go` - Enhanced principal extraction
- `/Users/kaushik/aaa-service/internal/grpc_server/service_authorization.go` - Service-specific auth
- `/Users/kaushik/aaa-service/internal/grpc_server/user_authorization.go` - User-specific auth
- `/Users/kaushik/aaa-service/internal/grpc_server/authorization_refactored.go` - Complete refactor
- `/Users/kaushik/aaa-service/internal/grpc_server/authorization_test.go` - Comprehensive tests

## Next Steps

1. **Apply the fix** to `internal/grpc_server/authorization.go`
2. **Run the test suite** to validate the changes
3. **Test with actual services** using their API keys
4. **Monitor logs** for any authorization errors
5. **Deploy to staging** for integration testing
6. **Deploy to production** after validation
