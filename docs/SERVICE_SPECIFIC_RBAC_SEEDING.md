# Service-Specific RBAC Seeding Guide

## Overview

The AAA service now supports **service-specific role and permission seeding**, allowing each microservice in the Kisanlink platform to define and seed its own custom RBAC configuration while maintaining complete isolation from other services.

## Key Features

- **Service Isolation**: Each service can have its own "admin" role without conflicts
- **Authorization Controls**: Two-tier permission model (basic seed + ownership validation)
- **Audit Trail**: Complete tracking of which service seeded which roles
- **Transaction Safety**: All-or-nothing seeding with automatic rollback on failures
- **Thread-Safe**: Concurrent seeding operations properly synchronized

## Architecture

### Authorization Model

| Caller Type | Permission Required | Can Seed For |
|-------------|---------------------|--------------|
| **Super Admin User** | `catalog:seed` + `admin:*` | Any service |
| **Regular User** | `catalog:seed` | Default (farmers-module) only |
| **Service (Self)** | `catalog:seed` | Its own service_id only |
| **Service (Other)** | ❌ Denied | Cannot seed other services |

## Usage Examples

### Example 1: Service Seeding Its Own Roles (gRPC)

```go
// ERP service seeding its own roles
conn, err := grpc.Dial("aaa-service:50051", grpc.WithInsecure())
client := pb.NewCatalogServiceClient(conn)

// Add API key for service authentication
md := metadata.New(map[string]string{
    "x-api-key": "your-erp-service-api-key",
})
ctx := metadata.NewOutgoingContext(context.Background(), md)

// Seed ERP-specific roles
resp, err := client.SeedRolesAndPermissions(ctx, &pb.SeedRolesAndPermissionsRequest{
    ServiceId: "erp-service",
    Force:     false, // Don't overwrite existing roles
})

// Response includes:
// - ActionsCreated: 5
// - ResourcesCreated: 8
// - PermissionsCreated: 40
// - RolesCreated: 4
// - CreatedRoles: ["sales_manager", "inventory_clerk", "accountant", "erp_admin"]
```

### Example 2: Admin User Seeding Any Service (HTTP)

```bash
# Admin user can seed roles for any service
curl -X POST https://aaa-service/api/v1/catalog/seed \
  -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": "traceability-service",
    "force": false
  }'

# Response:
# {
#   "status_code": 200,
#   "message": "Successfully seeded roles and permissions",
#   "actions_created": 7,
#   "resources_created": 5,
#   "permissions_created": 35,
#   "roles_created": 3,
#   "created_roles": ["auditor", "trace_admin", "trace_viewer"]
# }
```

### Example 3: Default Seeding (Farmers Module)

```bash
# Seed default farmers-module roles (backward compatible)
curl -X POST https://aaa-service/api/v1/catalog/seed \
  -H "Authorization: Bearer $USER_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": "",  # Empty defaults to farmers-module
    "force": false
  }'
```

### Example 4: Check Seed Status

```bash
# Get current seed status
curl -X GET https://aaa-service/api/v1/catalog/seed/status \
  -H "Authorization: Bearer $JWT_TOKEN"

# Response:
# {
#   "total_roles": 25,
#   "total_permissions": 150,
#   "total_actions": 9,
#   "total_resources": 12,
#   "registered_services": ["farmers-module", "erp-service", "traceability-service"]
# }
```

## Implementing a Custom Seed Provider

### Step 1: Define Your Roles and Permissions

```go
package myservice

import "github.com/Kisanlink/aaa-service/v2/internal/services/catalog"

type MyServiceSeedProvider struct {
    *catalog.BaseSeedProvider
}

func NewMyServiceSeedProvider() *MyServiceSeedProvider {
    return &MyServiceSeedProvider{
        BaseSeedProvider: catalog.NewBaseSeedProvider(
            "my-service",      // service_id
            "My Service",      // service_name
        ),
    }
}

func (p *MyServiceSeedProvider) GetResources() []catalog.ResourceDefinition {
    return []catalog.ResourceDefinition{
        {Name: "sales_order", Type: "business/order", Description: "Sales order resource"},
        {Name: "invoice", Type: "business/invoice", Description: "Invoice resource"},
        {Name: "customer", Type: "business/customer", Description: "Customer resource"},
    }
}

func (p *MyServiceSeedProvider) GetActions() []catalog.ActionDefinition {
    return []catalog.ActionDefinition{
        {Name: "create", Description: "Create new resource", Category: "write", IsStatic: true},
        {Name: "read", Description: "Read resource", Category: "read", IsStatic: true},
        {Name: "update", Description: "Update resource", Category: "write", IsStatic: true},
        {Name: "delete", Description: "Delete resource", Category: "write", IsStatic: true},
        {Name: "approve", Description: "Approve resource", Category: "approval", IsStatic: false},
    }
}

func (p *MyServiceSeedProvider) GetRoles() []catalog.RoleDefinition {
    return []catalog.RoleDefinition{
        {
            Name:        "sales_manager",
            Description: "Can manage all sales operations",
            Scope:       "GLOBAL",
            Permissions: []string{
                "sales_order:*",      // All actions on sales orders
                "invoice:create",      // Create invoices
                "invoice:read",        // Read invoices
                "customer:*",          // All actions on customers
            },
        },
        {
            Name:        "accountant",
            Description: "Can manage invoices and payments",
            Scope:       "GLOBAL",
            Permissions: []string{
                "invoice:*",           // All invoice actions
                "sales_order:read",    // Read-only sales orders
            },
        },
    }
}
```

### Step 2: Register Provider on Service Startup

```go
// In your service initialization
func initializeAAA() {
    // Connect to AAA service
    conn, _ := grpc.Dial("aaa-service:50051", grpc.WithInsecure())
    client := pb.NewCatalogServiceClient(conn)

    // Seed your service-specific roles
    ctx := context.Background()
    resp, err := client.SeedRolesAndPermissions(ctx, &pb.SeedRolesAndPermissionsRequest{
        ServiceId: "my-service",
        Force:     false,
    })

    if err != nil {
        log.Fatalf("Failed to seed RBAC: %v", err)
    }

    log.Printf("Successfully seeded %d roles: %v", resp.RolesCreated, resp.CreatedRoles)
}
```

## Error Handling

### Common Error Scenarios

```bash
# 1. Unauthorized service trying to seed another service
curl -X POST https://aaa-service/api/v1/catalog/seed \
  -H "x-api-key: $ERP_API_KEY" \
  -d '{"service_id": "farmers-module"}'
# Response: 403 Forbidden
# Message: "services can only seed their own roles"

# 2. Invalid service_id format
curl -X POST https://aaa-service/api/v1/catalog/seed \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"service_id": "my service with spaces"}'
# Response: 400 Bad Request
# Message: "invalid service_id format"

# 3. Missing catalog:seed permission
curl -X POST https://aaa-service/api/v1/catalog/seed \
  -H "Authorization: Bearer $REGULAR_USER_TOKEN" \
  -d '{"service_id": "erp-service"}'
# Response: 403 Forbidden
# Message: "insufficient permissions to seed roles"
```

## Best Practices

1. **Seed on Service Startup**: Include RBAC seeding in your service initialization
2. **Use Meaningful Role Names**: Prefix with service name to avoid confusion (e.g., `erp_sales_manager`)
3. **Document Permissions**: Maintain clear documentation of what each role can do
4. **Test Authorization**: Verify both success and failure cases for your roles
5. **Use Force Sparingly**: Only set `force=true` when you intentionally want to update existing roles
6. **Monitor Audit Logs**: Track who seeded what roles via `service_role_mappings` table

## Security Considerations

- **API Keys**: Store service API keys securely (environment variables, secrets manager)
- **JWT Tokens**: Use short-lived tokens for user operations
- **Permission Granularity**: Define fine-grained permissions rather than broad wildcards
- **Regular Audits**: Review `service_role_mappings` table to track role ownership
- **Least Privilege**: Only grant `admin:*` permission to trusted administrators

## Troubleshooting

### Issue: "Provider not found for service"
**Solution**: Register your provider before calling seed, or ensure service_id matches

### Issue: "Transaction rolled back"
**Solution**: Check logs for specific failure reason - likely constraint violation or invalid data

### Issue: "Duplicate role name"
**Solution**: Roles are unique per (service_id, role_name) - check if role already exists

### Issue: "Permission check failed"
**Solution**: Verify caller has `catalog:seed` permission and appropriate service ownership

## API Reference

### gRPC Endpoint

```protobuf
service CatalogService {
  rpc SeedRolesAndPermissions(SeedRolesAndPermissionsRequest) returns (SeedRolesAndPermissionsResponse);
}

message SeedRolesAndPermissionsRequest {
  string service_id = 1;  // Optional, defaults to "farmers-module"
  bool force = 2;          // Overwrite existing roles if true
}
```

### HTTP Endpoints

- **POST** `/api/v1/catalog/seed` - Seed roles and permissions
- **GET** `/api/v1/catalog/seed/status` - Get seeding status

## Support

For issues or questions:
1. Check logs: `kubectl logs -n aaa-service <pod-name>`
2. Verify permissions: Query `role_permissions` and `service_role_mappings` tables
3. Contact: AAA service maintainers
