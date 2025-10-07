# AAA Service Seed Data Documentation

This document describes all the seed data created for the AAA service RBAC system.

## Overview

The AAA service uses three separate seed files to initialize the RBAC system:

1. **seed_static_actions.go** - Creates 38 built-in actions
2. **seed_core_roles_permissions.go** - Creates core resources, roles, permissions, and default users
3. **seed_comprehensive_rbac.go** - Extends with all additional resources and comprehensive permissions

## Execution Order

Seeds run automatically on service startup in this order:

```
1. SeedStaticActions
   ↓
2. SeedCoreResourcesRolesPermissions
   ↓
3. SeedComprehensiveRBAC
```

## Seed Data Reference

### 1. Actions (38 total)

Created by `seed_static_actions.go`

#### CRUD Operations (7)
- `create` - Create a new resource
- `read` - Read/view a resource
- `view` - View a resource (alias for read)
- `update` - Update/modify a resource
- `edit` - Edit a resource (alias for update)
- `delete` - Delete a resource
- `list` - List multiple resources

#### Administrative Actions (6)
- `manage` - Full management access to a resource
- `admin` - Administrative access to a resource
- `assign` - Assign roles or permissions
- `unassign` - Remove roles or permissions
- `grant` - Grant access to a resource
- `revoke` - Revoke access to a resource

#### Ownership Actions (3)
- `own` - Own a resource
- `transfer` - Transfer ownership of a resource
- `share` - Share a resource with others

#### Data Operations (4)
- `export` - Export data from a resource
- `import` - Import data into a resource
- `backup` - Backup a resource
- `restore` - Restore a resource from backup

#### API/Service Actions (3)
- `execute` - Execute an operation
- `invoke` - Invoke a service or function
- `call` - Call an API endpoint

#### Database Actions (5)
- `select` - Select data from database
- `insert` - Insert data into database
- `update_rows` - Update rows in database
- `delete_rows` - Delete rows from database
- `truncate` - Truncate a database table

#### Audit/Monitoring Actions (3)
- `audit` - View audit logs
- `monitor` - Monitor resource activity
- `inspect` - Inspect resource details

#### Workflow Actions (4)
- `approve` - Approve a request or change
- `reject` - Reject a request or change
- `submit` - Submit for approval
- `cancel` - Cancel an operation

#### Special Actions (3)
- `impersonate` - Impersonate another user (sensitive)
- `bypass` - Bypass normal restrictions (sensitive)
- `override` - Override settings or decisions (sensitive)

### 2. Resources (24 total)

#### Core AAA Resources (7)
Created by `seed_core_roles_permissions.go`:

- `user` (aaa/user) - User accounts and authentication
- `role` (aaa/role) - Role definitions and assignments
- `permission` (aaa/permission) - Permission definitions
- `audit_log` (aaa/audit_log) - Audit trail and logging
- `system` (aaa/system) - System configuration and management
- `api_endpoint` (aaa/api_endpoint) - API endpoint protection
- `resource` (aaa/resource) - Generic AAA resource management

#### Extended Resources (17)
Created by `seed_comprehensive_rbac.go`:

**Organization & Group Resources:**
- `organization` (aaa/organization) - Organization management and hierarchy
- `group` (aaa/group) - Group management and memberships
- `group_role` (aaa/group_role) - Group role assignments

**RBAC Resources:**
- `action` (aaa/action) - Action definitions for RBAC

**User-Related Resources:**
- `user_profile` (aaa/user_profile) - User profile information
- `contact` (aaa/contact) - Contact information management
- `address` (aaa/address) - Address management

**Permission Resources:**
- `column_permission` (aaa/column_permission) - Column-level permissions
- `column` (aaa/column) - Column definitions
- `temporary_permission` (aaa/temporary_permission) - Time-bound permissions

**Resource Relationships:**
- `user_resource` (aaa/user_resource) - User-specific resource access
- `role_resource` (aaa/role_resource) - Role-specific resource access
- `permission_resource` (aaa/permission_resource) - Permission-resource mappings

**Advanced Resources:**
- `hierarchical_resource` (aaa/hierarchical_resource) - Hierarchical resource structures

**Database Resources:**
- `database` (aaa/database) - Database instance management
- `table` (aaa/table) - Database table access
- `database_operation` (aaa/database_operation) - Database operation control

### 3. Roles (6 total)

Created by `seed_core_roles_permissions.go`:

#### super_admin
- **Scope:** GLOBAL
- **Description:** Super Administrator with global access
- **Permissions:** ALL permissions on ALL resources
- **Default User:** superadmin (phone: +91-9999999999, password: SuperAdmin@123)

#### admin
- **Scope:** ORG
- **Description:** Administrator with organization-level access
- **Core Permissions:**
  - user: read, update
  - role: read, assign
  - permission: read
  - audit_log: read
- **Extended Permissions:**
  - organization: read, update
  - group: create, read, update, delete
  - group_role: assign, unassign, read
  - user_profile: read, update
  - contact: read, update
  - address: read, update
- **Default User:** admin (phone: +91-8888888888, password: Admin@123)

#### user
- **Scope:** ORG
- **Description:** Regular user with basic access
- **Permissions:**
  - user: read, update
  - resource: read
  - user_profile: read, update
  - contact: read, update
  - address: read, update
  - organization: read
  - group: read

#### viewer
- **Scope:** ORG
- **Description:** Read-only access user
- **Permissions:**
  - user: read
  - resource: read
  - organization: read
  - group: read
  - user_profile: read

#### aaa_admin
- **Scope:** GLOBAL
- **Description:** AAA service administrator
- **Core Permissions:**
  - user: manage, create, read, update, delete
  - role: manage, create, read, update, delete
  - permission: manage, create, read, update, delete
  - audit_log: read, export
  - system: manage
  - api_endpoint: call
  - resource: manage, read, update
- **Extended Permissions:**
  - action: manage, create, read, update, delete
  - column_permission: manage, create, read, update, delete
  - column: read, manage
  - temporary_permission: create, read, revoke
  - user_resource: read, manage
  - role_resource: read, manage
  - permission_resource: read, manage

#### module_admin
- **Scope:** ORG
- **Description:** Module administrator for service management
- **Core Permissions:**
  - user: read, update
  - role: read, assign
  - permission: read
  - resource: read, update
- **Extended Permissions:**
  - organization: read
  - group: read, update
  - group_role: read, assign

### 4. Default Users (2 total)

Created by `seed_core_roles_permissions.go`:

#### Super Administrator
- **Username:** superadmin
- **Phone:** +91-9999999999
- **Password:** SuperAdmin@123
- **Role:** super_admin
- **Status:** active
- **Email Validated:** true

#### Administrator
- **Username:** admin
- **Phone:** +91-8888888888
- **Password:** Admin@123
- **Role:** admin
- **Status:** active
- **Email Validated:** true

## Permission Naming Convention

Permissions follow the pattern: `{resource}:{action}`

Examples:
- `user:create` - Permission to create users
- `role:assign` - Permission to assign roles
- `audit_log:read` - Permission to read audit logs
- `organization:manage` - Permission to manage organizations

## Permission Matrix Summary

| Role | User Mgmt | Role Mgmt | Org Mgmt | Group Mgmt | System | Database |
|------|-----------|-----------|----------|------------|--------|----------|
| super_admin | Full | Full | Full | Full | Full | Full |
| admin | Read/Update | Read/Assign | Read/Update | Full | - | - |
| user | Self-service | - | Read | Read | - | - |
| viewer | Read | - | Read | Read | - | - |
| aaa_admin | Full | Full | - | - | Full | - |
| module_admin | Read/Update | Read/Assign | Read | Read/Update | - | - |

## Idempotency

All seed functions are idempotent:
- Re-running seeds will not create duplicates
- Existing resources/roles/permissions are skipped
- Safe to run on every service startup

## Database Schema

Seeds create data in these tables:
- `actions` - Action definitions
- `resources` - Resource definitions
- `roles` - Role definitions
- `permissions` - Permission definitions (resource + action)
- `role_permissions` - Role to permission assignments
- `users` - Default system users
- `user_roles` - User to role assignments

## Security Notes

1. **Change Default Passwords:** The seeded users have well-known passwords. Change them immediately in production.
2. **Static Actions:** Actions marked as `is_static=true` cannot be deleted through the API
3. **Global Roles:** Roles with `GLOBAL` scope are not tied to any organization
4. **Soft Deletes:** All tables support soft deletes to maintain audit trail

## Testing Seeds

To test the seeds in development:

```bash
# Start the service (seeds run automatically)
go run cmd/server/main.go

# Or run seeds manually in tests
func TestSeeds(t *testing.T) {
    ctx := context.Background()
    err := migrations.SeedStaticActionsWithDBManager(ctx, dbManager, logger)
    assert.NoError(t, err)

    err = migrations.SeedCoreResourcesRolesPermissionsWithDBManager(ctx, dbManager, logger)
    assert.NoError(t, err)

    err = migrations.SeedComprehensiveRBACWithDBManager(ctx, dbManager, logger)
    assert.NoError(t, err)
}
```

## Extending Seeds

To add new resources or permissions:

1. **Add Resource Type Constant** in `internal/entities/models/resource.go`
2. **Add to allResources** in `seed_comprehensive_rbac.go`
3. **Add Permission Matrix** for each role in `seedComprehensivePermissions()`
4. **Restart Service** to apply new seeds

## Troubleshooting

### Seed Failures

If seeding fails:
1. Check database connectivity
2. Verify migrations have run (`001_create_all_tables.up.sql`)
3. Check logs for specific error messages
4. Ensure database user has CREATE/INSERT permissions

### Duplicate Key Errors

Seeds are idempotent, but if you encounter duplicates:
1. Check `name` uniqueness constraints
2. Verify filter conditions in seed code
3. May indicate partial seed failure - check logs

### Missing Permissions

If expected permissions are missing:
1. Verify action exists in static actions seed
2. Check resource exists in resources seed
3. Ensure role exists
4. Check permission matrix in seed code

## Production Considerations

1. **Run seeds before application starts** accepting traffic
2. **Monitor seed duration** - should complete in < 10 seconds
3. **Log seed results** for audit trail
4. **Backup database** before running seeds in production
5. **Test seeds** in staging environment first
6. **Change default passwords** immediately after first deployment
7. **Consider seed versioning** for future schema changes

## Seed Execution Logs

Successful seed execution produces logs like:

```
INFO 🌱 Starting database seeding...
INFO Created static action action=create
INFO Created static action action=read
...
INFO Successfully seeded static actions count=38
INFO Created core resource name=user
...
INFO Created default role name=super_admin
INFO Seeded core role permissions for all default roles
INFO Created resource name=organization type=aaa/organization
...
INFO Completed comprehensive resource seeding total_resources=24
INFO Applied extended permissions to role role=super_admin
INFO Completed comprehensive permission seeding
```
