# Service Authorization System

## Quick Reference

This system provides configuration-based authorization for external services calling AAA Service's CatalogService operations.

## Key Files

```
config/
├── service_permissions.yaml       # Production/staging configuration
└── service_permissions.dev.yaml   # Development configuration

internal/
├── config/
│   └── service_authorization.go   # Configuration structures and loading
├── authorization/
│   ├── service_authorizer.go      # Core authorization logic
│   └── service_authorizer_test.go # Unit tests
└── grpc_server/
    ├── authorization.go           # Integration with existing auth
    └── catalog_service_authorization_test.go  # Integration tests

.kiro/specs/service-authorization/
├── README.md                      # This file
├── design.md                      # Architecture and design
└── deployment-guide.md            # Deployment instructions
```

## How It Works

1. **Service makes request** to `CatalogService.SeedRolesAndPermissions`
2. **CatalogHandler** validates service_id and calls `AuthorizationChecker`
3. **AuthorizationChecker** routes to `ServiceAuthorizer` for service principals
4. **ServiceAuthorizer** checks:
   - Is service in configuration?
   - Does it have required permission (catalog:seed_roles)?
   - Is API key valid (if required)?
   - Does service own the target roles (service_name == targetServiceID)?
5. **Return** authorization result to handler

## Permission Format

```
resource:action

Examples:
- catalog:seed_roles         (exact permission)
- catalog:*                  (all catalog operations)
- *:*                        (global wildcard)
```

## Configuration Example

```yaml
service_authorization:
  enabled: true
  services:
    farmers-module:
      service_id: "farmers-module"
      display_name: "Farmers Module Service"
      api_key_required: true
      permissions:
        - "catalog:seed_roles"
        - "catalog:seed_permissions"

default_behavior:
  when_disabled: "allow_all"
  log_unauthorized_attempts: true
```

## Environment Variables

```bash
# Select configuration file
export AAA_ENV=development  # Uses service_permissions.dev.yaml
export AAA_ENV=production   # Uses service_permissions.yaml

# API keys (one per service)
export AAA_SERVICE_API_KEY_FARMERS_MODULE="your-secure-api-key"
export AAA_SERVICE_API_KEY_ERP_MODULE="another-secure-key"
```

## Quick Setup

### 1. Generate API Key
```bash
openssl rand -base64 32
```

### 2. Configure Service
Edit `config/service_permissions.yaml`:
```yaml
services:
  your-service:
    service_id: "your-service"
    display_name: "Your Service Name"
    api_key_required: true
    permissions:
      - "catalog:seed_roles"
```

### 3. Set API Key
```bash
export AAA_SERVICE_API_KEY_YOUR_SERVICE="generated-key-from-step-1"
```

### 4. Restart Service
```bash
systemctl restart aaa-service
```

## Testing

### Unit Tests
```bash
# Test service authorizer
go test ./internal/authorization/... -v

# Test configuration loading
go test ./internal/config/... -v -run ServiceAuth
```

### Integration Tests
```bash
# Test full authorization flow
go test ./internal/grpc_server/... -v -run Authorization
```

### Manual Testing with grpcurl
```bash
# Test farmers-module seed operation
grpcurl -d '{
  "service_id": "farmers-module",
  "force": false
}' \
-H "x-api-key: your-api-key" \
localhost:50051 catalog.CatalogService.SeedRolesAndPermissions
```

## Common Operations

### Add New Service
1. Edit `config/service_permissions.yaml`
2. Add service configuration with permissions
3. Generate and set API key
4. Restart AAA service

### Disable Authorization (Emergency)
```bash
# Quick disable via config
sed -i 's/enabled: true/enabled: false/' config/service_permissions.yaml
systemctl restart aaa-service
```

### Rotate API Key
```bash
# Generate new key
NEW_KEY=$(openssl rand -base64 32)

# Update environment variable
export AAA_SERVICE_API_KEY_FARMERS_MODULE="$NEW_KEY"

# Restart service
systemctl restart aaa-service

# Update farmers-module with new key
```

## Troubleshooting

### Service Can't Seed Roles
```bash
# Check if service is configured
grep -A 5 "your-service:" config/service_permissions.yaml

# Verify API key is set
echo $AAA_SERVICE_API_KEY_YOUR_SERVICE

# Check logs for authorization failures
journalctl -u aaa-service | grep "authorization failed"
```

### Invalid API Key Error
```bash
# Verify variable name format: AAA_SERVICE_API_KEY_<SERVICE_ID_UPPERCASE>
# farmers-module → FARMERS_MODULE
env | grep AAA_SERVICE_API_KEY
```

### Configuration Not Loading
```bash
# Validate YAML syntax
yamllint config/service_permissions.yaml

# Check file permissions
ls -la config/service_permissions.yaml

# Verify environment
echo $AAA_ENV
```

## Security Notes

1. **Never commit API keys** - Use environment variables or secrets manager
2. **Rotate keys regularly** - Recommended every 90 days
3. **Use different keys per environment** - Dev ≠ Staging ≠ Production
4. **Monitor for failures** - Set up alerts for repeated authorization failures
5. **TODO: Implement key hashing** - Currently uses plaintext comparison

## Monitoring

### Logs to Watch
```bash
# Successful authorizations
grep "Service authorized to seed own roles" /var/log/aaa-service/

# Failed authorizations
grep "Service authorization failed" /var/log/aaa-service/

# API key failures
grep "invalid API key" /var/log/aaa-service/
```

### Metrics (Recommended)
- `aaa_service_authorization_total{service_id, permission, result}`
- `aaa_service_authorization_duration_seconds{service_id}`
- `aaa_service_api_key_failures_total{service_id}`

## Documentation Links

- **[Design Document](./design.md)** - Architecture and technical details
- **[Deployment Guide](./deployment-guide.md)** - Step-by-step deployment
- **[AAA Service Tech Stack](../../steering/tech.md)** - Overall tech architecture

## Support

For help:
1. Check logs: `journalctl -u aaa-service`
2. Review this README and deployment guide
3. Test configuration: `go run scripts/test_config.go`
4. Contact AAA service team

## Version History

- **v1.0** (2025-11-19): Initial implementation
  - Configuration-based authorization
  - API key validation
  - Wildcard permission support
  - Environment-specific configs
