# Service Authorization System - Deployment Guide

## Overview

This guide provides step-by-step instructions for deploying and configuring the service authorization system in different environments.

## Prerequisites

- AAA service version 2.0+
- Go 1.24+
- Access to deployment environment (dev/staging/prod)
- Service API keys for external services

## Quick Start

### 1. Generate API Key for farmers-module

```bash
# Generate a secure random API key
openssl rand -base64 32

# Example output: k7L9mN2pQ4rS6tU8vW0xY2zA4bC6dE8f
```

### 2. Configure Environment Variable

Add the API key to your environment:

```bash
# For farmers-module
export AAA_SERVICE_API_KEY_FARMERS_MODULE="k7L9mN2pQ4rS6tU8vW0xY2zA4bC6dE8f"
```

### 3. Enable Authorization (Production/Staging)

Edit `config/service_permissions.yaml`:

```yaml
service_authorization:
  enabled: true  # Set to true to enable
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

### 4. Restart AAA Service

```bash
# Restart the service to load new configuration
systemctl restart aaa-service

# Or if using Docker
docker-compose restart aaa-service
```

## Environment-Specific Configuration

### Development Environment

**Goal**: Allow unrestricted access for faster development

**Configuration** (`config/service_permissions.dev.yaml`):
```yaml
service_authorization:
  enabled: false

default_behavior:
  when_disabled: "allow_all"
  log_unauthorized_attempts: true
```

**Environment Variable**:
```bash
export AAA_ENV=development
```

**No API keys required** - All services can seed roles

### Staging Environment

**Goal**: Test authorization with production-like settings

**Configuration** (`config/service_permissions.yaml`):
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

**Environment Variables**:
```bash
export AAA_ENV=staging
export AAA_SERVICE_API_KEY_FARMERS_MODULE="staging-api-key-secure-random-string"
```

### Production Environment

**Goal**: Enforce strict authorization with API key validation

**Configuration** (`config/service_permissions.yaml`):
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
        - "catalog:register_resource"
        - "catalog:register_action"

default_behavior:
  when_disabled: "allow_all"
  log_unauthorized_attempts: true
```

**Environment Variables**:
```bash
export AAA_ENV=production
export AAA_SERVICE_API_KEY_FARMERS_MODULE="production-api-key-use-secrets-manager"
```

**Security Note**: In production, use AWS Secrets Manager or similar:
```bash
export AAA_SERVICE_API_KEY_FARMERS_MODULE=$(aws secretsmanager get-secret-value \
  --secret-id aaa-service/farmers-module/api-key \
  --query SecretString \
  --output text)
```

## Adding New Services

### Step 1: Generate API Key

```bash
# Generate secure API key
openssl rand -base64 32
```

### Step 2: Add to Configuration

Edit `config/service_permissions.yaml`:

```yaml
service_authorization:
  enabled: true
  services:
    # Existing services...

    # New service
    new-service:
      service_id: "new-service"
      display_name: "New Service Name"
      description: "Description of what this service does"
      api_key_required: true
      permissions:
        - "catalog:seed_roles"
        - "catalog:seed_permissions"
```

### Step 3: Set Environment Variable

```bash
export AAA_SERVICE_API_KEY_NEW_SERVICE="generated-api-key-here"
```

### Step 4: Test Configuration

```bash
# Check if service is configured
curl -X POST http://localhost:8080/api/v1/admin/config/validate

# Test authorization (from new-service)
grpcurl -d '{"service_id": "new-service", "force": false}' \
  -H "x-api-key: generated-api-key-here" \
  localhost:50051 catalog.CatalogService.SeedRolesAndPermissions
```

## Configuration Validation

### Validate YAML Syntax

```bash
# Install yamllint if not available
pip install yamllint

# Validate configuration
yamllint config/service_permissions.yaml
```

### Test Configuration Loading

Create a test script (`scripts/test_config.go`):

```go
package main

import (
	"fmt"
	"log"

	"github.com/Kisanlink/aaa-service/v2/internal/config"
)

func main() {
	cfg, err := config.LoadServiceAuthorizationConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Enabled: %v\n", cfg.IsEnabled())
	fmt.Printf("Default Behavior: %s\n", cfg.GetDefaultBehavior())
	fmt.Printf("Configured Services: %d\n", len(cfg.ServiceAuthorization.Services))

	for serviceID, service := range cfg.ServiceAuthorization.Services {
		fmt.Printf("\nService: %s\n", serviceID)
		fmt.Printf("  Display Name: %s\n", service.DisplayName)
		fmt.Printf("  API Key Required: %v\n", service.APIKeyRequired)
		fmt.Printf("  Permissions: %v\n", service.Permissions)
	}
}
```

Run the test:
```bash
go run scripts/test_config.go
```

## Troubleshooting

### Issue: Service Cannot Seed Roles

**Error**: `service 'farmers-module' is not authorized to seed roles`

**Solution**:
1. Check if service is configured in `config/service_permissions.yaml`
2. Verify `enabled: true` in configuration
3. Confirm service has `catalog:seed_roles` permission
4. Validate API key is correct

```bash
# Check configuration
cat config/service_permissions.yaml | grep -A 10 farmers-module

# Verify environment variable
echo $AAA_SERVICE_API_KEY_FARMERS_MODULE
```

### Issue: Invalid API Key

**Error**: `invalid API key`

**Solution**:
1. Verify API key environment variable is set
2. Check for whitespace or special characters
3. Ensure variable name matches pattern: `AAA_SERVICE_API_KEY_<SERVICE_ID_UPPERCASE>`

```bash
# Check if env var is set
env | grep AAA_SERVICE_API_KEY

# Variable name should be uppercase with underscores
# farmers-module → FARMERS_MODULE
export AAA_SERVICE_API_KEY_FARMERS_MODULE="your-api-key"
```

### Issue: Configuration Not Loading

**Error**: Service uses default configuration

**Solution**:
1. Check file path is correct: `config/service_permissions.yaml`
2. Verify file permissions are readable
3. Check for YAML syntax errors

```bash
# Check file exists
ls -la config/service_permissions.yaml

# Validate YAML
yamllint config/service_permissions.yaml

# Check file is readable
cat config/service_permissions.yaml
```

### Issue: Authorization Disabled in Production

**Error**: All requests are allowed when they shouldn't be

**Solution**:
1. Verify `enabled: true` in `config/service_permissions.yaml`
2. Check `AAA_ENV` is not set to `development`
3. Restart service after configuration change

```bash
# Check current environment
echo $AAA_ENV

# Should NOT be set to development in production
unset AAA_ENV  # or set to "production"

# Restart service
systemctl restart aaa-service
```

## Migration Checklist

### Pre-Deployment

- [ ] Generate API keys for all services
- [ ] Store API keys in secrets manager
- [ ] Update `config/service_permissions.yaml`
- [ ] Test configuration in development
- [ ] Validate YAML syntax
- [ ] Review permissions for each service

### Staging Deployment

- [ ] Deploy configuration to staging
- [ ] Set environment variables
- [ ] Restart AAA service
- [ ] Test farmers-module seed operation
- [ ] Monitor logs for authorization failures
- [ ] Validate API key authentication works

### Production Deployment

- [ ] Deploy configuration to production
- [ ] Set production API keys from secrets manager
- [ ] Restart AAA service with zero downtime
- [ ] Monitor authorization logs
- [ ] Test farmers-module integration
- [ ] Document any issues in runbook

### Post-Deployment

- [ ] Monitor authorization metrics
- [ ] Review logs for unauthorized attempts
- [ ] Update documentation
- [ ] Train team on new authorization system

## Monitoring

### Log Monitoring

Monitor logs for authorization events:

```bash
# Successful authorizations
journalctl -u aaa-service | grep "Service authorized to seed own roles"

# Failed authorizations
journalctl -u aaa-service | grep "Service authorization failed"

# API key failures
journalctl -u aaa-service | grep "invalid API key"
```

### Health Checks

Add authorization health check:

```bash
# Check if configuration loaded
curl http://localhost:8080/health

# Expected output should include authorization status
```

### Alerts

Set up alerts for:
1. Repeated API key failures (potential security issue)
2. Services missing from configuration
3. Configuration load failures

Example Prometheus alert:
```yaml
- alert: ServiceAuthorizationFailureHigh
  expr: rate(aaa_service_authorization_failures_total[5m]) > 10
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High rate of service authorization failures"
```

## Rollback Procedure

If issues occur after deployment:

### Quick Rollback (Disable Authorization)

```bash
# Edit configuration
sed -i 's/enabled: true/enabled: false/' config/service_permissions.yaml

# Restart service
systemctl restart aaa-service
```

### Full Rollback (Revert to Previous Version)

```bash
# Restore previous configuration
git checkout HEAD~1 config/service_permissions.yaml

# Restart service
systemctl restart aaa-service
```

## Security Best Practices

1. **Never commit API keys to git**
   - Use `.gitignore` for any files containing keys
   - Store keys in secrets manager

2. **Rotate API keys regularly**
   - Recommended: Every 90 days
   - Use automated rotation if possible

3. **Use different keys per environment**
   - Development keys != Staging keys != Production keys
   - Makes key compromise less severe

4. **Monitor for unauthorized access**
   - Set up alerts for failed authorization
   - Review logs regularly

5. **Implement rate limiting**
   - Prevent brute force API key attacks
   - TODO: Add rate limiting to service authorizer

## Support

For issues or questions:
- Create a ticket in project management system
- Check logs: `/var/log/aaa-service/`
- Review this guide and design document
- Contact AAA service team

## References

- [Design Document](./design.md)
- [Testing Documentation](./testing.md)
- [AAA Service Configuration](../../steering/tech.md)
