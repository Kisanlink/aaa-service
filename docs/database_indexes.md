# Database Performance Indexes

This document describes the database indexes added for performance optimization in the AAA service.

## Indexes Added

### User Roles Table (`user_roles`)

1. **idx_user_roles_user_id_is_active** - Composite index on `(user_id, is_active)`

   - **Purpose**: Optimizes queries that fetch active roles for a specific user
   - **Usage**: Login responses, role validation queries
   - **Query Pattern**: `SELECT * FROM user_roles WHERE user_id = ? AND is_active = true`

2. **idx_user_roles_user_role_assignment** - Composite index on `(user_id, role_id)`

   - **Purpose**: Optimizes role assignment checks and prevents duplicate assignments
   - **Usage**: Role assignment validation, checking if user has specific role
   - **Query Pattern**: `SELECT * FROM user_roles WHERE user_id = ? AND role_id = ?`

3. **idx_user_roles_is_active** - Index on `is_active`
   - **Purpose**: Optimizes queries that filter all active role assignments
   - **Usage**: Global role queries, administrative reports
   - **Query Pattern**: `SELECT * FROM user_roles WHERE is_active = true`

### Users Table (`users`)

1. **idx_users_phone_country_auth** - Composite index on `(phone_number, country_code)`

   - **Purpose**: Optimizes authentication queries using phone number and country code
   - **Usage**: Login authentication, user lookup
   - **Query Pattern**: `SELECT * FROM users WHERE phone_number = ? AND country_code = ?`

2. **idx_users_phone_country_validated** - Composite index on `(phone_number, country_code, is_validated)`
   - **Purpose**: Optimizes authentication queries that also check validation status
   - **Usage**: Login with validation check, user verification
   - **Query Pattern**: `SELECT * FROM users WHERE phone_number = ? AND country_code = ? AND is_validated = true`

## Existing Indexes (Verified)

The following indexes already exist and are maintained:

### Users Table

- `idx_users_phone_number` - Individual index on phone_number
- `idx_users_country_code` - Individual index on country_code
- `idx_users_status` - Index on user status
- `idx_users_is_validated` - Index on validation status

### User Roles Table

- `idx_user_roles_user_id` - Individual index on user_id
- `idx_user_roles_role_id` - Individual index on role_id

## Implementation Method

Indexes are implemented using GORM struct tags in the model definitions:

```go
// UserRole model with performance indexes
type UserRole struct {
    UserID   string `gorm:"index:idx_user_roles_user_id_is_active,priority:1"`
    RoleID   string `gorm:"index:idx_user_roles_user_role_assignment,priority:2"`
    IsActive bool   `gorm:"index:idx_user_roles_user_id_is_active,priority:2"`
}

// User model with authentication indexes
type User struct {
    PhoneNumber string `gorm:"index:idx_users_phone_country_auth,priority:1"`
    CountryCode string `gorm:"index:idx_users_phone_country_auth,priority:2"`
    IsValidated bool   `gorm:"index:idx_users_phone_country_validated,priority:3"`
}
```

## Performance Impact

These indexes will significantly improve:

1. **Login Performance**: Composite indexes on phone_number + country_code reduce authentication query time
2. **Role Query Performance**: User role lookups will be faster with user_id + is_active index
3. **Role Assignment Checks**: Duplicate role assignment prevention will be more efficient
4. **Administrative Queries**: Global role filtering will benefit from is_active index

## Auto-Migration

Indexes will be automatically created when GORM's auto-migration runs during application startup, ensuring consistency across all environments.
