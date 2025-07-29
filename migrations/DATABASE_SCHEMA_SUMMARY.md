# AAA Service Database Schema Summary

## Overview
This database schema supports a comprehensive Authentication, Authorization, and Accounting (AAA) service with user management, role-based access control, and address management capabilities.

## Files Generated
1. **`aaa_service_schema.sql`** - Complete schema with sample data and views
2. **`aaa_service_schema_clean.sql`** - Clean schema without sample data (recommended for drawSQL)
3. **Database ERD diagram** - Visual representation of table relationships

## How to Use with DrawSQL

### Option 1: Import SQL File
1. Go to [drawSQL.app](https://drawsql.app/)
2. Create a new project
3. Choose "Import from SQL"
4. Upload `aaa_service_schema_clean.sql`
5. DrawSQL will automatically generate the visual diagram

### Option 2: Manual Table Creation
Copy and paste the table creation statements from either SQL file into drawSQL's SQL editor.

## Database Tables

### Core Tables

#### 1. `users`
- **Purpose**: Core user authentication and account management
- **Key Fields**:
  - `username` (unique)
  - `password` (hashed)
  - `status` (pending/active/suspended/blocked)
  - `tokens` (token balance for paid operations)
- **Relationships**: One-to-one with user_profiles, one-to-many with contacts

#### 2. `addresses`
- **Purpose**: Shared address storage for users and contacts
- **Key Fields**: Complete Indian address structure including house, street, district, state, pincode
- **Special**: Can be referenced by both user_profiles and contacts

#### 3. `roles`
- **Purpose**: Define system roles with unique names
- **Key Fields**: `name` (unique), `description`
- **Relationships**: Many-to-many with users through user_roles, many-to-many with permissions

#### 4. `permissions`
- **Purpose**: Granular permissions for resources and actions
- **Key Fields**:
  - `resource` (what resource the permission applies to)
  - `effect` (allow/deny)
  - `actions` (PostgreSQL array of allowed actions)

### Relationship Tables

#### 5. `user_profiles`
- **Purpose**: Extended user information including personal details
- **Key Fields**:
  - `name`, `aadhaar_number`, `date_of_birth`
  - `user_id` (unique foreign key to users)
  - `address_id` (optional foreign key to addresses)

#### 6. `contacts`
- **Purpose**: Contact information for users
- **Key Fields**:
  - `mobile_number`, `country_code`, `email_hash`
  - `user_id` (foreign key to users)
  - `address_id` (optional foreign key to addresses)

#### 7. `user_roles`
- **Purpose**: Junction table linking users to roles
- **Key Fields**:
  - `user_id`, `role_id` (composite unique constraint)
  - `is_active` (enables/disables role assignment)

#### 8. `role_permissions`
- **Purpose**: Junction table linking roles to permissions
- **Key Fields**: `role_id`, `permission_id` (composite primary key)

## Key Design Features

### 1. Soft Deletes
All tables support soft deletes with:
- `deleted_at` timestamp
- `deleted_by` user tracking
- Indexes for performance

### 2. Audit Trail
Complete audit capabilities with:
- `created_at`, `updated_at` timestamps
- `created_by`, `updated_by` user tracking

### 3. Flexible Addressing
- Shared address model reduces redundancy
- Complete Indian address structure
- Support for Village/Town/City (VTC) designation

### 4. Role-Based Access Control (RBAC)
- Flexible many-to-many relationship between users and roles
- Granular permissions with resource-action mapping
- PostgreSQL arrays for efficient action storage

### 5. Performance Optimization
- Strategic indexing on frequently queried fields
- Unique constraints for business rules
- Foreign key constraints with CASCADE options

## Sample Queries

### Get User with Roles and Permissions
```sql
SELECT
    u.username,
    r.name as role_name,
    p.resource,
    p.actions
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE u.username = 'admin_user'
  AND ur.is_active = true
  AND u.deleted_at IS NULL;
```

### Get User Profile with Address
```sql
SELECT
    u.username,
    up.name,
    up.aadhaar_number,
    a.district,
    a.state,
    a.pincode
FROM users u
JOIN user_profiles up ON u.id = up.user_id
LEFT JOIN addresses a ON up.address_id = a.id
WHERE u.status = 'active'
  AND u.deleted_at IS NULL;
```

## Migration Considerations

### From Go Models to PostgreSQL
1. **String Pointers**: Go `*string` fields become nullable VARCHAR columns
2. **Enums**: Go string constants become VARCHAR with CHECK constraints
3. **Arrays**: Go slices become PostgreSQL arrays (TEXT[])
4. **Timestamps**: Go `time.Time` becomes PostgreSQL TIMESTAMP
5. **GORM Tags**: Translated to appropriate PostgreSQL constraints

### Indexes
All necessary indexes are included for:
- Primary keys (automatic)
- Foreign keys (for performance)
- Unique constraints (for business rules)
- Soft delete fields (for query performance)
- Commonly queried fields

## Security Considerations

1. **Password Storage**: Assumes bcrypt hashing (indicated by sample hash format)
2. **Soft Deletes**: Sensitive data remains in database but marked as deleted
3. **Email Privacy**: Email addresses are hashed, not stored in plain text
4. **Aadhaar Security**: Aadhaar numbers stored in user profiles (ensure compliance with local regulations)

## Scalability Notes

1. **ID Generation**: Uses custom hash-based ID generation from kisanlink-db package
2. **Table Sizes**: Models specify small/medium/large sizes for optimization
3. **Partitioning**: Consider partitioning audit tables by date for large systems
4. **Indexing**: Monitor query performance and add indexes as needed

## Next Steps

1. Import the schema into drawSQL for visualization
2. Review and adjust based on specific business requirements
3. Add any additional indexes based on query patterns
4. Consider data retention policies for soft-deleted records
5. Implement proper backup and recovery procedures
