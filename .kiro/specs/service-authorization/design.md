# Service Authorization System - Design Document

## Overview

This document describes the configuration-based service authorization system for AAA Service's CatalogService operations. This system enables fine-grained control over which external services can perform catalog operations like seeding roles and permissions.

## Business Requirements

### Problem Statement

The farmers-module service needs to seed its own roles and permissions in the AAA service via the `CatalogService.SeedRolesAndPermissions` gRPC endpoint. However, there was no mechanism to:

1. Authorize external services to perform catalog operations
2. Validate API keys for service-to-service authentication
3. Configure service permissions without code changes
4. Support multiple environments (dev, staging, prod) with different policies

### Solution

Implement a YAML-based configuration system that:

- Defines which services can perform which catalog operations
- Supports permission-based access control (resource:action format)
- Validates API keys for service authentication
- Provides environment-specific configurations
- Maintains backward compatibility with existing functionality

## Architecture

### Components

#### 1. ServiceAuthorizationConfig (`internal/config/service_authorization.go`)

Configuration data structures that represent the service authorization settings:

```go
type ServiceAuthorizationConfig struct {
    ServiceAuthorization ServiceAuthSection
    DefaultBehavior      DefaultBehavior
}

type ServicePermission struct {
    ServiceID      string
    DisplayName    string
    Description    string
    APIKeyRequired bool
    APIKey         string
    Permissions    []string
}
```

**Responsibilities:**
- Load configuration from YAML files
- Validate configuration structure
- Provide environment-specific config loading

#### 2. ServiceAuthorizer (`internal/authorization/service_authorizer.go`)

Core authorization logic component:

```go
type ServiceAuthorizer struct {
    config *config.ServiceAuthorizationConfig
    logger *zap.Logger
}

func (sa *ServiceAuthorizer) Authorize(ctx context.Context, serviceID string, permission string) error
```

**Responsibilities:**
- Validate service permissions against configuration
- Support exact match and wildcard permissions (catalog:*, *:*)
- Validate API keys from gRPC metadata
- Log authorization attempts and failures

#### 3. AuthorizationChecker Enhancement (`internal/grpc_server/authorization.go`)

Integration point that connects ServiceAuthorizer with the existing authorization system:

**Responsibilities:**
- Initialize ServiceAuthorizer with configuration
- Route service principals to configuration-based authorization
- Route user principals to RBAC-based authorization
- Maintain backward compatibility

### Authorization Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ CatalogService.SeedRolesAndPermissions gRPC Request             │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ CatalogHandler.SeedRolesAndPermissions                          │
│  - Validates ServiceID format                                   │
│  - Calls AuthorizationChecker.CheckSeedPermission               │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ AuthorizationChecker.CheckSeedPermission                        │
│  - Extracts principal from context                              │
│  - Routes to appropriate authorization path                     │
└────────────┬────────────────────────────────┬───────────────────┘
             │                                │
    Service Principal                   User Principal
             │                                │
             ▼                                ▼
┌────────────────────────┐        ┌──────────────────────────┐
│ ServiceAuthorizer      │        │ RBAC Permission Check    │
│  - Check config        │        │  - catalog:seed          │
│  - Validate API key    │        │  - admin:* (for services)│
│  - Match permissions   │        └──────────────────────────┘
└────────────────────────┘
```

## Configuration Format

### Production Configuration (`config/service_permissions.yaml`)

```yaml
service_authorization:
  enabled: true
  services:
    farmers-module:
      service_id: "farmers-module"
      display_name: "Farmers Module Service"
      description: "Farmer management and agricultural operations service"
      api_key_required: true
      permissions:
        - "catalog:seed_roles"
        - "catalog:seed_permissions"
        - "catalog:register_resource"
        - "catalog:register_action"

default_behavior:
  when_disabled: "allow_all"
  log_unauthorized_attempts: true
```

### Development Configuration (`config/service_permissions.dev.yaml`)

```yaml
service_authorization:
  enabled: false

default_behavior:
  when_disabled: "allow_all"
  log_unauthorized_attempts: true
```

### Permission Format

Permissions use the format: `resource:action`

Examples:
- `catalog:seed_roles` - Exact permission
- `catalog:*` - All catalog operations
- `*:*` - Global wildcard (all operations)

## Security Considerations

### API Key Management

1. **Configuration File**: API keys can be stored in YAML (not recommended for production)
2. **Environment Variables**: Preferred method using pattern `AAA_SERVICE_API_KEY_<SERVICE_ID_UPPERCASE>`
   - Example: `AAA_SERVICE_API_KEY_FARMERS_MODULE=secret-key-123`

### Current Implementation

- **TODO**: API keys are currently compared in plaintext
- **Future**: Implement bcrypt hashing for API keys
- **TODO**: Implement constant-time comparison to prevent timing attacks

### API Key Validation Flow

1. Extract `x-api-key` from gRPC metadata
2. Look up expected key from config or environment variable
3. Compare keys (currently plaintext, should be hashed)
4. Return error if missing or invalid

## Authorization Rules

### For Service Principals

1. **Configuration Required**: Service must be defined in `service_permissions.yaml`
2. **Permission Required**: Service must have `catalog:seed_roles` or `catalog:*` permission
3. **API Key**: Must provide valid API key if `api_key_required: true`
4. **Ownership**: Service can only seed its own roles (service_name == targetServiceID)
5. **Restriction**: Cannot seed default/farmers-module roles (reserved for system)

### For User Principals

1. **RBAC Permission**: Must have `catalog:seed` permission
2. **Default Access**: Can seed default/farmers-module with basic permission
3. **Service-Specific**: Need `admin:*` permission to seed service-specific roles

## Environment Support

### Environment Selection

The system automatically selects configuration based on `AAA_ENV` environment variable:

- `AAA_ENV=development` or `AAA_ENV=dev` → Uses `config/service_permissions.dev.yaml`
- All other values → Uses `config/service_permissions.yaml`

### Fallback Behavior

If configuration file doesn't exist:
- System uses default configuration
- `enabled: false`
- `when_disabled: "allow_all"`
- No breaking changes to existing functionality

## Implementation Details

### Key Design Decisions

1. **Backward Compatibility**: System defaults to `allow_all` when disabled
2. **Environment-Specific**: Separate configs for dev/staging/prod
3. **Fail-Safe**: Missing config file doesn't break the service
4. **Logging**: All authorization attempts are logged with structured fields

### Wildcard Permission Matching

The `hasPermission` method supports multiple matching strategies:

1. **Exact Match**: `catalog:seed_roles` matches `catalog:seed_roles`
2. **Resource Wildcard**: `catalog:*` matches `catalog:seed_roles`
3. **Global Wildcard**: `*:*` matches any permission

### Error Handling

Authorization failures return gRPC status codes:
- `codes.Unauthenticated` (401): Missing or invalid authentication
- `codes.PermissionDenied` (403): Valid auth but insufficient permissions
- `codes.InvalidArgument` (400): Invalid service_id or permission format

## Testing Strategy

### Unit Tests (`internal/authorization/service_authorizer_test.go`)

Tests cover:
- Authorization enabled/disabled states
- Permission matching (exact, wildcard, global)
- API key validation
- Invalid input handling
- Configuration loading

### Integration Tests (`internal/grpc_server/catalog_service_authorization_test.go`)

Tests cover:
- End-to-end authorization flow
- Service principal vs user principal routing
- Wildcard permission behavior
- Environment-specific configuration
- Error propagation through gRPC layer

## Migration Path

### Phase 1: Development (Current)
- Authorization disabled by default
- Validate configuration loading
- Test with farmers-module

### Phase 2: Staging
- Enable authorization with `farmers-module` configured
- Validate API key authentication
- Monitor logs for unauthorized attempts

### Phase 3: Production
- Enable authorization for all services
- Enforce API key requirements
- Implement API key hashing (TODO)

## Future Enhancements

### Short Term
1. Implement bcrypt hashing for API keys
2. Add constant-time comparison for keys
3. Add rate limiting for failed authorization attempts

### Long Term
1. Support for JWT-based service authentication
2. OAuth2/OpenID Connect integration
3. Dynamic permission management via API
4. Audit trail for all authorization decisions
5. Support for permission inheritance/hierarchies

## Monitoring and Observability

### Logging

All authorization events are logged with structured fields:
- `service_id`: Identifier of the calling service
- `permission`: Permission being requested
- `result`: Authorized/denied
- `reason`: Why authorization failed (if applicable)

### Metrics (Recommended)

Implement Prometheus metrics:
- `aaa_service_authorization_total{service_id, permission, result}`
- `aaa_service_authorization_duration_seconds{service_id, permission}`
- `aaa_service_api_key_failures_total{service_id}`

## References

- [AAA Service Tech Stack](../../steering/tech.md)
- [Product Requirements](../../steering/product.md)
- [Catalog Service Architecture](../catalog-service-architecture.md)
