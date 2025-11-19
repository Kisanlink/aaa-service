# Farmers Module Authorization Fix

## Summary

Fixed the service authorization issue where farmers-module was being rejected when trying to seed roles.

## Problem

The farmers-module sends `service_id: "Farmers Module"` (with space) but the configuration was expecting `"farmers-module"` (hyphenated).

## Changes Made

### 1. Updated Configuration (`config/service_permissions.yaml`)

Added support for both service ID formats:
- `"Farmers Module"` - Current format sent by farmers-module
- `"farmers-module"` - Legacy format for backward compatibility

### 2. Fixed API Key Environment Variable Generation

Updated `internal/authorization/service_authorizer.go` to handle spaces in service IDs:
- Before: `AAA_SERVICE_API_KEY_FARMERS MODULE` (invalid - space in env var name)
- After: `AAA_SERVICE_API_KEY_FARMERS_MODULE` (valid)

The code now normalizes service IDs by replacing both spaces AND hyphens with underscores.

## Setup Instructions

### Step 1: Set the API Key Environment Variable

The AAA service expects the API key in an environment variable:

```bash
# Generate a secure API key
export AAA_SERVICE_API_KEY_FARMERS_MODULE="your-api-key-here"

# Or use this to generate a random key
export AAA_SERVICE_API_KEY_FARMERS_MODULE=$(openssl rand -base64 32)
```

**IMPORTANT:** The environment variable name is `AAA_SERVICE_API_KEY_FARMERS_MODULE`
- "Farmers Module" → normalized to "FARMERS_MODULE" (spaces → underscores)

### Step 2: Configure Farmers Module

The farmers-module must send the same API key in the `x-api-key` header when making gRPC calls.

From the logs, it appears the farmers-module is already doing this:
```
2025/11/19 14:29:58 AAA Client: Added x-api-key to request for method: /pb.CatalogService/SeedRolesAndPermissions
```

### Step 3: Restart AAA Service

```bash
# Restart to load the environment variable
systemctl restart aaa-service

# Or if running locally
go run ./cmd/server
```

### Step 4: Test the Integration

Try seeding roles from farmers-module:

```bash
# From farmers-module, call the seed endpoint
curl -X POST http://localhost:8080/admin/seed
```

You should see success logs in the AAA service:
```
Service 'Farmers Module' authorized successfully
Service 'Farmers Module' completed seeding roles
```

## Configuration Details

### Production Config (`config/service_permissions.yaml`)

```yaml
service_authorization:
  enabled: true
  services:
    "Farmers Module":
      service_id: "Farmers Module"
      api_key_required: true
      permissions:
        - "catalog:seed_roles"
        - "catalog:seed_permissions"
        - "catalog:register_resource"
        - "catalog:register_action"

    farmers-module:  # Legacy support
      service_id: "farmers-module"
      api_key_required: true
      permissions:
        - "catalog:seed_roles"
        - "catalog:seed_permissions"
        - "catalog:register_resource"
        - "catalog:register_action"
```

### Development Config (`config/service_permissions.dev.yaml`)

```yaml
service_authorization:
  enabled: false  # Disabled for local development
default_behavior:
  when_disabled: "allow_all"
```

## Environment Variables

| Service ID | Environment Variable Name |
|-----------|---------------------------|
| "Farmers Module" | `AAA_SERVICE_API_KEY_FARMERS_MODULE` |
| "farmers-module" | `AAA_SERVICE_API_KEY_FARMERS_MODULE` |

Both service IDs normalize to the same environment variable name.

## Troubleshooting

### Error: "missing x-api-key header"

**Cause:** The farmers-module is not sending the API key.

**Solution:** Ensure the farmers-module gRPC client is configured to send the `x-api-key` header.

### Error: "invalid API key"

**Cause:** The API key sent by farmers-module doesn't match the one configured in AAA service.

**Solution:**
1. Check the environment variable is set: `echo $AAA_SERVICE_API_KEY_FARMERS_MODULE`
2. Ensure farmers-module is using the same key
3. Restart both services to reload configuration

### Error: "service 'Farmers Module' is not authorized"

**Cause:** The service is not in the configuration file.

**Solution:** Verify `config/service_permissions.yaml` has the "Farmers Module" entry with required permissions.

### Error: "service 'Farmers Module' does not have permission 'catalog:seed_roles'"

**Cause:** The service is configured but missing the specific permission.

**Solution:** Add `catalog:seed_roles` to the permissions list in the config.

## Testing

All tests pass:
```bash
go test -v ./internal/authorization/...
go test -v ./internal/grpc_server/... -run Authorization
```

Build succeeds:
```bash
go build ./cmd/server
```

## Deployment Checklist

- [ ] Set environment variable `AAA_SERVICE_API_KEY_FARMERS_MODULE` on AAA service
- [ ] Configure farmers-module to use the same API key in `x-api-key` header
- [ ] Deploy updated AAA service configuration
- [ ] Restart AAA service
- [ ] Test farmers-module seeding endpoint
- [ ] Verify logs show successful authorization

## Files Modified

1. `config/service_permissions.yaml` - Added "Farmers Module" service configuration
2. `internal/authorization/service_authorizer.go` - Fixed environment variable name normalization

## Security Notes

- API keys are currently compared in plaintext
- TODO: Implement bcrypt hashing for production
- API keys should be stored in a secrets manager (AWS Secrets Manager, etc.)
- Rotate keys every 90 days

## Next Steps

Once verified working:
1. Consider consolidating to single service ID format (recommend "Farmers Module" to match actual usage)
2. Implement API key hashing for production
3. Set up API key rotation policy
4. Add monitoring/alerting for unauthorized access attempts
