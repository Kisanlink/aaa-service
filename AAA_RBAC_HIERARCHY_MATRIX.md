# AAA Service RBAC Hierarchy Matrix

## Overview

This document provides a comprehensive role vs permissions vs resources vs actions matrix for the AAA service, showing the hierarchical relationships and inheritance patterns for effective permission management.

## Core Components

### 1. Subjects (Who)

- **Users**: Individual users with profiles and contacts
- **Groups**: Collections of users with hierarchical relationships
- **Services**: System services and applications

### 2. Resources (What)

- **Core Entities**: Users, Roles, Permissions, Organizations, Groups
- **System Resources**: Audit logs, configurations, API endpoints
- **Data Resources**: Tables, databases, columns

### 3. Actions (How)

- **CRUD Operations**: Create, Read, Update, Delete
- **Management Operations**: Assign, Revoke, Manage, Execute
- **System Operations**: Backup, Restore, Migrate, Configure

### 4. Roles (Authority Level)

- **Global Roles**: System-wide authority
- **Organization Roles**: Organization-scoped authority
- **Group Roles**: Group-scoped authority

## Hierarchy Tree Structure

```
AAA Service RBAC Hierarchy
│
├── GLOBAL SCOPE
│   ├── Super Admin
│   │   ├── System Management
│   │   ├── Organization Management
│   │   ├── Global User Management
│   │   └── Audit & Compliance
│   │
│   ├── System Admin
│   │   ├── System Configuration
│   │   ├── Database Operations
│   │   └── API Management
│   │
│   └── Auditor
│       ├── View Audit Logs
│       ├── Export Reports
│       └── Read-only Access
│
├── ORGANIZATION SCOPE
│   ├── Organization Admin
│   │   ├── Organization Management
│   │   ├── User Management (Org)
│   │   ├── Role Management (Org)
│   │   └── Group Management
│   │
│   ├── User Manager
│   │   ├── User CRUD Operations
│   │   ├── Profile Management
│   │   └── Contact Management
│   │
│   ├── Role Manager
│   │   ├── Role CRUD Operations
│   │   ├── Permission Assignment
│   │   └── User-Role Assignment
│   │
│   └── Group Manager
│       ├── Group CRUD Operations
│       ├── Membership Management
│       └── Group Hierarchy
│
└── GROUP SCOPE
    ├── Group Admin
    │   ├── Group Configuration
    │   ├── Member Management
    │   └── Sub-group Management
    │
    ├── Group Member
    │   ├── View Group Info
    │   ├── Update Own Profile
    │   └── Basic Operations
    │
    └── Group Viewer
        ├── Read-only Group Access
        └── View Member List
```

## Resource Type Hierarchy

### Core AAA Resources

```
aaa/
├── user
│   ├── user_profile
│   ├── contact
│   └── address
├── role
├── permission
├── resource
├── action
├── organization
├── group
├── audit_log
└── system
```

### Extended Resources

```
aaa/
├── api_endpoint
├── database_operation
├── table
├── database
├── column
├── temporary_permission
├── hierarchical_resource
├── user_resource
├── role_resource
└── permission_resource
```

## Action Categories & Hierarchy

### 1. Basic CRUD Actions

```
CRUD
├── create
├── read (view)
├── update (edit)
└── delete
```

### 2. Management Actions

```
Management
├── manage
├── assign
├── revoke
├── execute
└── validate
```

### 3. User-Specific Actions

```
User Actions
├── read_profile
├── update_profile
├── read_contacts
├── update_contacts
├── read_addresses
├── update_addresses
├── manage_tokens
├── suspend_user
└── block_user
```

### 4. Role-Specific Actions

```
Role Actions
├── assign_permissions
├── remove_permissions
├── assign_users
├── remove_users
└── inherit_from_parent
```

### 5. System Actions

```
System Actions
├── system_config
├── backup_restore
├── view_audit_logs
├── export
├── migrate
└── propagate_to_children
```

## Permission Inheritance Matrix

### Role Hierarchy Inheritance

```
Super Admin (Global)
├── Inherits: ALL permissions across ALL resources
├── Scope: Global (all organizations)
└── Special: Can create/delete organizations

Organization Admin (Org-scoped)
├── Inherits: ALL permissions within organization
├── Scope: Single organization
└── Cannot: Manage global roles or other organizations

Group Admin (Group-scoped)
├── Inherits: ALL permissions within group
├── Scope: Single group within organization
└── Cannot: Manage organization-level resources
```

### Resource-Based Inheritance

```
Organization Resource
├── Contains: Groups, Users, Roles (org-scoped)
├── Permissions cascade to: All contained resources
└── Access control: Organization membership required

Group Resource
├── Contains: Users (group members), Sub-groups
├── Permissions cascade to: Group members and sub-groups
└── Access control: Group membership or parent group access

User Resource
├── Contains: Profile, Contacts, Addresses
├── Permissions cascade to: All user-owned resources
└── Access control: User ownership or management permissions
```

## Role-Permission-Resource-Action Matrix

### Super Admin Role

| Resource Type    | Actions | Inheritance | Scope             |
| ---------------- | ------- | ----------- | ----------------- |
| aaa/user         | ALL     | Global      | All organizations |
| aaa/role         | ALL     | Global      | All organizations |
| aaa/organization | ALL     | Global      | System-wide       |
| aaa/system       | ALL     | Global      | System-wide       |
| aaa/audit_log    | ALL     | Global      | System-wide       |

### Organization Admin Role

| Resource Type    | Actions                              | Inheritance  | Scope            |
| ---------------- | ------------------------------------ | ------------ | ---------------- |
| aaa/user         | create, read, update, delete, manage | Organization | Within org       |
| aaa/role         | create, read, update, delete, assign | Organization | Org-scoped roles |
| aaa/group        | create, read, update, delete, manage | Organization | Within org       |
| aaa/organization | read, update                         | Self         | Own organization |

### User Manager Role

| Resource Type    | Actions                              | Inheritance  | Scope      |
| ---------------- | ------------------------------------ | ------------ | ---------- |
| aaa/user         | create, read, update, suspend, block | Organization | Within org |
| aaa/user_profile | read, update                         | Organization | Within org |
| aaa/contact      | read, update, create, delete         | Organization | Within org |
| aaa/address      | read, update, create, delete         | Organization | Within org |

### Group Admin Role

| Resource Type        | Actions                      | Inheritance | Scope         |
| -------------------- | ---------------------------- | ----------- | ------------- |
| aaa/group            | read, update, manage         | Group       | Own group     |
| aaa/user             | read, assign, remove         | Group       | Group members |
| aaa/group_membership | create, read, update, delete | Group       | Own group     |

## Permission Assignment Strategies

### 1. Direct Assignment

```
User → Role → Permissions
- Direct role assignment to users
- Explicit permission grants
- No inheritance involved
```

### 2. Group-Based Assignment

```
User → Group → Role → Permissions
- Users inherit roles through group membership
- Group-level role assignments
- Automatic permission inheritance
```

### 3. Hierarchical Assignment

```
User → Group → Parent Group → Role → Permissions
- Multi-level group hierarchy
- Permission inheritance through group tree
- Cascading role assignments
```

### 4. Resource-Scoped Assignment

```
User → Role (Resource-Scoped) → Permissions (Resource-Specific)
- Role limited to specific resource types
- Resource-level permission grants
- Fine-grained access control
```

## Best Practices for Permission Management

### 1. Principle of Least Privilege

- Start with minimal permissions
- Grant additional permissions as needed
- Regular permission audits

### 2. Role Hierarchy Design

- Use inheritance to reduce duplication
- Clear parent-child relationships
- Avoid circular dependencies

### 3. Resource Organization

- Group related resources logically
- Use consistent naming conventions
- Implement proper resource hierarchies

### 4. Permission Granularity

- Balance between too broad and too granular
- Use resource-action combinations effectively
- Consider operational efficiency

### 5. Audit and Compliance

- Track all permission changes
- Regular access reviews
- Automated compliance reporting

## Implementation Considerations

### Database Schema Alignment

- Roles table with hierarchy support (ParentID)
- Resource permissions with role-resource-action mapping
- Binding table for flexible subject-role/permission assignments

### Caching Strategy

- Cache frequently accessed permission checks
- Invalidate cache on role/permission changes
- Use Redis for distributed caching

### Performance Optimization

- Index on frequently queried permission combinations
- Batch permission checks where possible
- Optimize inheritance resolution queries

### Security Considerations

- Validate all permission assignments
- Prevent privilege escalation
- Secure permission inheritance chains
- Regular security audits

This matrix provides a comprehensive framework for understanding and implementing the AAA service's RBAC system with proper hierarchy and inheritance patterns.
