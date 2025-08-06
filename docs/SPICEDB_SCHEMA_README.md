# SpiceDB Schema for AAA Service

This document describes the SpiceDB schema designed for the AAA (Authentication, Authorization, and Accounting) service to support role-based access control at both resource_id and column level granularity.

## Overview

The SpiceDB schema provides fine-grained access control that maps to the existing aaa-service database relations while adding additional granularity for column-level permissions and hierarchical resource management.

## Core Entities

### 1. User Management (`aaa/user`)

**Maps to:** `users` table in aaa-service

**Relations:**
- `direct`: Direct user-to-user relationships
- `parent`: Hierarchical user relationships (e.g., admin → regular user)
- `role`: User's assigned roles
- `profile`: User profile information
- `contact`: User contact information

**Permissions:**
- `view`: View user information
- `edit`: Edit user information
- `delete`: Delete user
- `manage`: Full user management

### 2. User Profile (`aaa/user_profile`)

**Maps to:** `user_profiles` table

**Relations:**
- `user`: Associated user
- `address`: User's address

**Permissions:** Inherited from user permissions

### 3. Contact (`aaa/contact`)

**Maps to:** `contacts` table

**Relations:**
- `user`: Associated user
- `address`: Contact's address

**Permissions:** Inherited from user permissions

### 4. Address (`aaa/address`)

**Maps to:** `addresses` table

**Relations:**
- `owner`: User who owns the address
- `shared_with`: Users with shared access

**Permissions:**
- `view`: View address (owner + shared users)
- `edit`: Edit address (owner only)
- `delete`: Delete address (owner only)
- `manage`: Full address management (owner only)

## Role-Based Access Control

### 5. Role (`aaa/role`)

**Maps to:** `roles` table

**Relations:**
- `direct`: Users directly assigned to this role
- `parent`: Parent role for hierarchical role inheritance
- `permission`: Permissions assigned to this role

**Permissions:**
- `view`: View role information
- `edit`: Edit role
- `delete`: Delete role
- `manage`: Full role management
- `assign`: Assign users to this role

### 6. Permission (`aaa/permission`)

**Maps to:** `permissions` table

**Relations:**
- `role`: Role that has this permission
- `resource`: Resource this permission applies to
- `action`: Specific action allowed

**Permissions:**
- `view`: View permission details
- `edit`: Edit permission
- `delete`: Delete permission
- `manage`: Full permission management

### 7. Resource (`aaa/resource`)

**Maps to:** Various resource types in the system

**Relations:**
- `owner`: User who owns the resource
- `role`: Role with access to this resource
- `parent`: Parent resource for hierarchical access

**Permissions:**
- `view`: View resource
- `edit`: Edit resource
- `delete`: Delete resource
- `manage`: Full resource management
- `create`: Create new resources
- `read`: Read resource data
- `update`: Update resource data
- `delete_resource`: Delete the resource itself

## Column-Level Granular Access Control

### 8. Column Permission (`aaa/column_permission`)

**Purpose:** Provides granular access control at the database column level

**Relations:**
- `user`: User with column-level access
- `role`: Role with column-level access
- `resource`: Resource containing the column
- `column`: Specific column being controlled

**Permissions:**
- `view`: View column data
- `edit`: Edit column data
- `delete`: Delete column data
- `manage`: Full column management

### 9. Column (`aaa/column`)

**Purpose:** Represents individual database columns

**Relations:**
- `resource`: Resource containing this column
- `permission`: Permissions for this column

**Permissions:** Inherited from resource and column permissions

## Resource-Specific Definitions

### 10. User Resource (`aaa/user_resource`)

**Purpose:** Specific permissions for user-related operations

**Permissions:**
- Standard CRUD operations
- `read_profile`: Read user profile
- `update_profile`: Update user profile
- `read_contacts`: Read user contacts
- `update_contacts`: Update user contacts
- `read_addresses`: Read user addresses
- `update_addresses`: Update user addresses
- `manage_tokens`: Manage user tokens
- `validate_user`: Validate user account
- `suspend_user`: Suspend user account
- `block_user`: Block user account

### 11. Role Resource (`aaa/role_resource`)

**Purpose:** Specific permissions for role management

**Permissions:**
- Standard CRUD operations
- `assign_permissions`: Assign permissions to roles
- `remove_permissions`: Remove permissions from roles
- `assign_users`: Assign users to roles
- `remove_users`: Remove users from roles

### 12. Permission Resource (`aaa/permission_resource`)

**Purpose:** Specific permissions for permission management

**Permissions:**
- Standard CRUD operations
- `create_permission`: Create new permissions
- `assign_to_roles`: Assign permissions to roles
- `remove_from_roles`: Remove permissions from roles

## Audit and System Management

### 13. Audit Log (`aaa/audit_log`)

**Maps to:** `audit_logs` table

**Permissions:**
- `view`: View audit logs
- `create`: Create audit entries
- `manage`: Manage audit logs
- `read_all`: Read all audit logs
- `export`: Export audit logs

### 14. System (`aaa/system`)

**Purpose:** System-level administrative permissions

**Permissions:**
- `manage_users`: Manage all users
- `manage_roles`: Manage all roles
- `manage_permissions`: Manage all permissions
- `view_audit_logs`: View system audit logs
- `system_config`: Configure system settings
- `backup_restore`: Backup and restore operations

## Advanced Features

### 15. Temporary Permission (`aaa/temporary_permission`)

**Purpose:** Time-limited access control

**Relations:**
- `user`: User with temporary access
- `role`: Role with temporary access
- `resource`: Resource with temporary access
- `granted_by`: User who granted the permission
- `expires_at`: Expiration timestamp

**Permissions:**
- Standard CRUD operations
- `extend`: Extend temporary permission
- `revoke`: Revoke temporary permission

### 16. Hierarchical Resource (`aaa/hierarchical_resource`)

**Purpose:** Nested resource access control

**Relations:**
- `owner`: Resource owner
- `parent`: Parent resource
- `children`: Child resources
- `role`: Role with access

**Permissions:**
- Standard CRUD operations
- `inherit_from_parent`: Inherit permissions from parent
- `propagate_to_children`: Propagate permissions to children

## API and Database Access Control

### 17. API Endpoint (`aaa/api_endpoint`)

**Purpose:** HTTP method-level access control

**Permissions:**
- `get`: GET requests
- `post`: POST requests
- `put`: PUT requests
- `patch`: PATCH requests
- `delete`: DELETE requests
- `head`: HEAD requests
- `options`: OPTIONS requests

### 18. Database Operation (`aaa/database_operation`)

**Purpose:** Database operation-level access control

**Permissions:**
- `select`: SELECT operations
- `insert`: INSERT operations
- `update`: UPDATE operations
- `delete`: DELETE operations
- `create_table`: CREATE TABLE operations
- `drop_table`: DROP TABLE operations
- `alter_table`: ALTER TABLE operations
- `create_index`: CREATE INDEX operations
- `drop_index`: DROP INDEX operations

### 19. Table (`aaa/table`)

**Purpose:** Table-level access control

**Permissions:**
- Standard CRUD operations
- `read_all_rows`: Read all rows in table
- `read_own_rows`: Read only user's own rows
- `insert_rows`: Insert rows
- `update_rows`: Update rows
- `delete_rows`: Delete rows

### 20. Database (`aaa/database`)

**Purpose:** Database-level access control

**Permissions:**
- Standard CRUD operations
- `backup`: Database backup operations
- `restore`: Database restore operations
- `migrate`: Database migration operations

## Usage Examples

### Example 1: User Access Control

```zed
// Check if user can view their own profile
check aaa/user:user123#view@aaa/user_profile:profile456

// Check if admin can view any user profile
check aaa/user:admin#view@aaa/user_profile:profile789
```

### Example 2: Column-Level Access

```zed
// Check if user can view email column
check aaa/user:user123#view@aaa/column:email_column

// Check if user can edit password column
check aaa/user:user123#edit@aaa/column:password_column
```

### Example 3: Role-Based Access

```zed
// Check if role has permission to manage users
check aaa/role:admin_role#manage@aaa/user_resource:user_management

// Check if user can assign roles
check aaa/user:admin#assign@aaa/role:user_role
```

### Example 4: Resource-Specific Permissions

```zed
// Check if user can validate other users
check aaa/user:admin#validate_user@aaa/user_resource:user_validation

// Check if user can manage tokens
check aaa/user:admin#manage_tokens@aaa/user_resource:token_management
```

## Migration from Current System

The SpiceDB schema is designed to work alongside the existing aaa-service database schema. Key mapping points:

1. **Users**: `users` table → `aaa/user` definition
2. **Roles**: `roles` table → `aaa/role` definition
3. **Permissions**: `permissions` table → `aaa/permission` definition
4. **User Roles**: `user_roles` table → Relations in `aaa/user` and `aaa/role`
5. **Role Permissions**: `role_permissions` table → Relations in `aaa/role` and `aaa/permission`

## Benefits

1. **Granular Access Control**: Column-level permissions for sensitive data
2. **Hierarchical Permissions**: Support for nested resource access
3. **Temporary Access**: Time-limited permissions for temporary access
4. **Audit Trail**: Comprehensive audit logging capabilities
5. **API Security**: HTTP method-level access control
6. **Database Security**: Database operation-level access control
7. **Scalability**: Efficient permission checking for large-scale systems

## Implementation Notes

1. The schema supports both direct user permissions and role-based permissions
2. Column-level permissions provide fine-grained data access control
3. Hierarchical resources support organizational structures
4. Temporary permissions enable secure time-limited access
5. Audit logging ensures compliance and security monitoring
6. API and database operation permissions provide comprehensive security coverage
