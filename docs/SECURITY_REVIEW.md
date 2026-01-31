# Kisanlink Multi-Tenant Security Review

## Executive Summary

This document provides a comprehensive security analysis of the Kisanlink platform. The architecture deploys some services centrally (aaa-service, farmers-module, admin-panel) and others per-FPO (fpo-erp, smartcard-ui).

### Deployment Architecture

```
CENTRALIZED SERVICES
+-----------------------------------------------------------+
|  +----------+  +-----------------+  +----------------+    |
|  | AAA      |  | Farmers-Module  |  | Admin-Panel    |    |
|  | Service  |  | (Farmer Data)   |  | (Management UI)|    |
|  +----------+  +-----------------+  +----------------+    |
+-----------------------------------------------------------+
                          |
                JWT + Organization Context
                          |
+-----------------------------------------------------------+
                    PER-FPO DEPLOYMENTS
|  +--------------------+       +--------------------+      |
|  | FPO-1 Cluster      |       | FPO-2 Cluster      | ...  |
|  | +----------------+ |       | +----------------+ |      |
|  | | fpo-erp        | |       | | fpo-erp        | |      |
|  | +----------------+ |       | +----------------+ |      |
|  | +----------------+ |       | +----------------+ |      |
|  | | smartcard-ui   | |       | | smartcard-ui   | |      |
|  | +----------------+ |       | +----------------+ |      |
|  | +----------------+ |       | +----------------+ |      |
|  | | PostgreSQL DB  | |       | | PostgreSQL DB  | |      |
|  | +----------------+ |       | +----------------+ |      |
|  +--------------------+       +--------------------+      |
+-----------------------------------------------------------+
```

---

## Security Controls Implemented

### 1. AAA-Service Security

**JWT Token Security:**
- HS256 signing with validated secrets
- Token type validation (access vs refresh)
- Issuer and audience validation
- Organization context embedded in claims

**Authorization:**
- gRPC-based permission checking
- Organization-scoped RBAC implementation
- Service-to-service API key authentication

### 2. FPO-ERP Deployment Validation

**Environment-Based Organization Binding:**
Each FPO-ERP deployment MUST set `EXPECTED_ORG_ID` environment variable:

```yaml
Environment Variables:
  EXPECTED_ORG_ID: "fpo-specific-org-id"  # Validates token org matches deployment
```

The `RequireDeploymentOrg()` middleware validates that:
1. The JWT token's organization_id matches the deployment's expected organization
2. Users cannot access FPO-2's ERP deployment with FPO-1's token

### 3. Cross-FPO Attack Prevention

**Attack Scenario Prevented:**
1. FPO-1 Admin obtains valid JWT token
2. Admin discovers FPO-2's ERP deployment URL
3. Admin sends request to FPO-2's ERP with FPO-1 token
4. **BLOCKED:** Middleware validates token org doesn't match deployment org

---

## Deployment Checklist

### Pre-Deployment Verification

- [ ] AAA-Service: JWT secrets configured per environment
- [ ] FPO-ERP: `EXPECTED_ORG_ID` environment variable set per deployment
- [ ] FPO-ERP: Deployment-level org validation middleware enabled
- [ ] All services: CORS configured to specific domains (not wildcard)
- [ ] All services: TLS/HTTPS enabled

### Per-FPO Deployment Configuration

Each FPO-ERP deployment MUST have:

```yaml
Environment Variables:
  EXPECTED_ORG_ID: "fpo-specific-org-id"       # Validates token org matches deployment
  DB_POSTGRES_HOST: "fpo-specific-rds"
  DB_POSTGRES_DBNAME: "erp_fpo_xxx"
  AWS_S3_BUCKET: "kisanlink-attachments-fpo-xxx"
  CORS_ALLOWED_ORIGINS: "https://fpo-xxx.erp.kisanlink.in"
  AAA_JWT_SECRET: "${AAA_JWT_SECRET}"          # Shared with AAA service
```

### Security Testing Before Go-Live

1. **Cross-FPO Access Test:**
   - Obtain token for FPO-1
   - Attempt to access FPO-2's ERP deployment
   - Should return 403 Forbidden

2. **Token Manipulation Test:**
   - Modify JWT claims (org_id)
   - Attempt API access
   - Should fail signature validation

3. **Data Isolation Test:**
   - Query products/sales via FPO-1 token
   - Verify only FPO-1 data returned

4. **CORS Test:**
   - From unauthorized origin, attempt API calls
   - Should be blocked by CORS

---

## Known Security Considerations

### Current Mitigations

| Risk | Mitigation |
|------|------------|
| Cross-FPO Data Access | Physical database isolation + deployment org validation |
| Token Theft | Short access token TTL, refresh token rotation |
| Unauthorized Access | JWT signature validation + RBAC |
| CSRF | SameSite cookies + token-based auth |

### Future Improvements (Roadmap)

1. **HttpOnly Cookies for Tokens** - Move from localStorage to secure cookies
2. **Rate Limiting** - Per-organization request throttling
3. **PostgreSQL Row-Level Security** - Database-level tenant isolation
4. **Comprehensive Audit Logging** - Security event tracking

---

## Architecture Security Notes

### Why Per-FPO Database Isolation?

1. **Data Sovereignty:** Each FPO owns their data completely
2. **Blast Radius:** Compromise of one database doesn't affect others
3. **Compliance:** Easier to meet data residency requirements
4. **Performance:** No multi-tenant query overhead

### Why Centralized AAA?

1. **Single Sign-On:** Users authenticate once for all services
2. **Consistent Authorization:** Unified permission model
3. **Audit Trail:** Centralized security logging
4. **Token Management:** Single point for token issuance/revocation

---

## Contact

For security concerns or vulnerability reports, contact the security team.

**Last Updated:** January 2026
**Review Frequency:** Quarterly
