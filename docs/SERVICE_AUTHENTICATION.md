# Service-to-Service Authentication Guide

This document explains how to register services and use API key authentication for service-to-service gRPC calls.

## Overview

The AAA service supports two authentication methods for gRPC:
1. **JWT tokens** - For user authentication
2. **API keys** - For service-to-service authentication

When a service authenticates using an API key, the following context values are set:
- `service_id` - The unique ID of the authenticated service
- `service_name` - The name of the service
- `principal_type` - Set to "service"
- `user_id` - Set to service_id for backward compatibility with existing code

## How to Register a Service

### Step 1: Generate an API Key

First, generate a secure API key (you can use the provided endpoint or generate your own):

```bash
# Option 1: Use the AAA service endpoint
curl -X POST http://localhost:8080/api/v1/services/generate-api-key \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"

# Response:
{
  "api_key": "sk_live_abc123xyz..."
}

# Option 2: Generate manually (example)
openssl rand -base64 32
```

### Step 2: Register the Service

Register your service with the AAA service:

```bash
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "farmers-service",
    "description": "Service for farmer management",
    "organization_id": "ORGN00000001",
    "api_key": "sk_live_abc123xyz...",
    "metadata": "{\"version\":\"v1\"}"
  }'

# Response:
{
  "id": "SVC00000001",
  "name": "farmers-service",
  "description": "Service for farmer management",
  "organization_id": "ORGN00000001",
  "is_active": true,
  "created_at": "2025-10-06T10:30:00Z",
  "updated_at": "2025-10-06T10:30:00Z"
}
```

**Important Notes:**
- The `api_key` you provide will be hashed using SHA-256 before storage
- Keep the plain API key secure - it cannot be retrieved later
- The service must be active (`is_active: true`) to authenticate

## Using API Key Authentication in gRPC Clients

### Go Client Example (farmers-module)

```go
package main

import (
    "context"
    "log"

    pb "github.com/Kisanlink/aaa-service/pkg/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

func main() {
    // Connect to AAA service
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Create client
    client := pb.NewOrganizationServiceClient(conn)

    // Add API key to metadata
    md := metadata.New(map[string]string{
        "x-api-key": "sk_live_abc123xyz...", // Your service's API key
    })
    ctx := metadata.NewOutgoingContext(context.Background(), md)

    // Make gRPC call
    resp, err := client.GetOrganization(ctx, &pb.GetOrganizationRequest{
        Id: "ORGN00000001",
    })
    if err != nil {
        log.Fatalf("GetOrganization failed: %v", err)
    }

    log.Printf("Organization: %+v", resp.Organization)
}
```

### Using with Interceptors (Recommended for Production)

```go
package main

import (
    "context"

    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

// APIKeyInterceptor adds API key to all outgoing requests
func APIKeyInterceptor(apiKey string) grpc.UnaryClientInterceptor {
    return func(
        ctx context.Context,
        method string,
        req, reply interface{},
        cc *grpc.ClientConn,
        invoker grpc.UnaryInvoker,
        opts ...grpc.CallOption,
    ) error {
        // Add API key to metadata
        md := metadata.New(map[string]string{
            "x-api-key": apiKey,
        })
        ctx = metadata.NewOutgoingContext(ctx, md)

        // Call the RPC method
        return invoker(ctx, method, req, reply, cc, opts...)
    }
}

func main() {
    // Connect with API key interceptor
    conn, err := grpc.Dial(
        "localhost:50051",
        grpc.WithInsecure(),
        grpc.WithUnaryInterceptor(APIKeyInterceptor("sk_live_abc123xyz...")),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // All calls will automatically include the API key
    client := pb.NewOrganizationServiceClient(conn)
    resp, err := client.GetOrganization(context.Background(), &pb.GetOrganizationRequest{
        Id: "ORGN00000001",
    })
    // ...
}
```

## Where Context Values Are Used

The context values set during service authentication are used in:

### 1. Audit Logging (`internal/grpc_server/grpc_server.go:275-300`)

```go
func (s *GRPCServer) auditInterceptor(...) {
    // Extract user ID from context (works for both users and services)
    userID := "unknown"
    if userIDValue := ctx.Value("user_id"); userIDValue != nil {
        if uid, ok := userIDValue.(string); ok {
            userID = uid  // For services, this is service_id
        }
    }

    // Log audit event
    s.auditService.LogUserAction(ctx, userID, "grpc_call", "grpc", info.FullMethod, ...)
}
```

When a service makes a call, the audit log will show:
- `user_id` = service ID (e.g., "SVC00000001")
- `action` = "grpc_call"
- `resource` = method name

### 2. Authorization Checks

If you want to check if the caller is a service vs a user:

```go
// In your handler
func (h *OrganizationHandler) GetOrganization(ctx context.Context, req *pb.GetOrganizationRequest) (*pb.GetOrganizationResponse, error) {
    // Check if caller is a service
    if principalType, ok := ctx.Value("principal_type").(string); ok && principalType == "service" {
        serviceID := ctx.Value("service_id").(string)
        serviceName := ctx.Value("service_name").(string)
        h.logger.Info("Service accessing organization",
            zap.String("service_id", serviceID),
            zap.String("service_name", serviceName))
    }

    // Your business logic...
}
```

### 3. Future Use Cases

The context values can be used for:
- **Service-specific rate limiting** - Different limits for different services
- **Service-based permissions** - Grant specific services access to specific resources
- **Service tracking** - Track which services are using which endpoints
- **Billing/metering** - Track usage per service

## Service Management Endpoints

### List Services
```bash
GET /api/v1/services
```

### Get Service Details
```bash
GET /api/v1/services/:id
```

### Delete Service
```bash
DELETE /api/v1/services/:id
Authorization: Bearer YOUR_ADMIN_TOKEN
```

## Security Best Practices

1. **Store API keys securely**
   - Use environment variables or secret management systems
   - Never commit API keys to version control
   - Rotate keys periodically

2. **Use HTTPS/TLS in production**
   - Enable TLS for gRPC connections
   - API keys should never be sent over unencrypted connections

3. **Monitor service usage**
   - Check audit logs regularly
   - Set up alerts for unusual activity
   - Track which services are accessing which resources

4. **Principle of least privilege**
   - Create separate services for different purposes
   - Grant only necessary permissions
   - Deactivate unused services

## Troubleshooting

### "invalid API key" Error

**Causes:**
- API key doesn't match any registered service
- Service is inactive (`is_active: false`)
- API key was not hashed correctly during registration

**Solution:**
```bash
# Check if service exists and is active
curl http://localhost:8080/api/v1/services/SVC00000001

# Regenerate API key if needed
curl -X POST http://localhost:8080/api/v1/services/generate-api-key \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"

# Update service with new key (requires re-registration)
```

### "authorization token is not provided" Error

**Cause:**
- Missing `x-api-key` header in gRPC metadata
- Metadata not properly attached to context

**Solution:**
```go
// Make sure to create metadata and attach to context
md := metadata.New(map[string]string{
    "x-api-key": "your-api-key",
})
ctx := metadata.NewOutgoingContext(context.Background(), md)
```

### "service is inactive" Error

**Cause:**
- Service has been deactivated (`is_active: false`)

**Solution:**
- Contact AAA service administrator to reactivate the service
- Or create a new service registration

## Migration from Public Endpoints

If you previously used public endpoints (no authentication), migration steps:

1. Register your service with AAA
2. Update your gRPC client to include API key
3. Test in staging environment
4. Deploy to production

Example migration:

```diff
func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

+   // Add API key authentication
+   md := metadata.New(map[string]string{
+       "x-api-key": os.Getenv("AAA_SERVICE_API_KEY"),
+   })
+   ctx := metadata.NewOutgoingContext(context.Background(), md)
-   ctx := context.Background()

    resp, err := client.GetOrganization(ctx, &pb.GetOrganizationRequest{...})
}
```
