# AAA Service Database Migrations

This directory contains database migrations for the AAA (Authentication, Authorization, and Accounting) service.

## Migration Files

### `001_create_all_tables.up.sql`
Comprehensive migration that creates all database tables for the AAA service. This migration includes:

#### Core Tables
- **organizations** - Organizations with hierarchical support
- **users** - User accounts with phone authentication and token system
- **user_profiles** - Extended user profile information
- **addresses** - Physical addresses for users and organizations
- **contacts** - Contact information including mobile numbers and email hashes

#### Authorization Tables
- **roles** - Roles with scope (global/organization) and hierarchical support
- **actions** - Actions that can be performed on resources
- **resources** - Resources that can be protected by permissions
- **permissions** - Permissions linking roles, resources, and actions
- **user_roles** - Many-to-many relationship between users and roles
- **role_permissions** - Many-to-many relationship between roles and permissions
- **resource_permissions** - Direct resource-level permissions for roles

#### Group and Organization Tables
- **groups** - Groups for organizing users and roles within organizations
- **group_memberships** - User and service memberships in groups with time bounds
- **group_inheritance** - Hierarchical relationships between groups

#### Principal and Service Tables
- **principals** - Unified identity representation for users and services
- **services** - Service accounts with API key authentication

#### Attribute and Binding Tables
- **attributes** - Key-value attributes for ABAC (Attribute-Based Access Control)
- **attribute_history** - Audit trail for attribute changes
- **bindings** - Subject-to-role/permission bindings with caveats
- **binding_history** - Audit trail for binding changes

#### Column Permission Tables
- **column_groups** - Named groups of columns for column-level permissions
- **column_group_members** - Columns belonging to column groups
- **column_sets** - Optimized bitmap representations of allowed columns

#### Audit and Event Tables
- **audit_logs** - Audit trail for all system activities
- **events** - Immutable event log for system state changes
- **event_checkpoints** - Periodic checkpoints of the event chain

### `001_create_all_tables.down.sql`
Rollback migration that drops all tables created in the up migration, handling foreign key constraints correctly.

## Features

### Base Model Support
All tables include the standard base model fields:
- `id` - Primary key (VARCHAR(255))
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp
- `created_by` - User who created the record
- `updated_by` - User who last updated the record
- `deleted_at` - Soft delete timestamp
- `deleted_by` - User who deleted the record

### Constraints and Validation
- **Foreign Key Constraints** - Proper referential integrity
- **Check Constraints** - Data validation (e.g., status values, scope values)
- **Unique Constraints** - Prevents duplicate data
- **Cascade Rules** - Appropriate delete behaviors

### Performance Optimization
- **Strategic Indexes** - On frequently queried columns
- **Composite Indexes** - For complex queries
- **Partial Indexes** - Where appropriate

### Audit and Compliance
- **Soft Delete Support** - Maintains data history
- **Audit Trails** - Comprehensive logging of changes
- **Event Sourcing** - Immutable event log with integrity verification

### Advanced Features
- **JSONB Support** - Flexible metadata storage
- **Hierarchical Data** - Self-referencing tables for trees
- **Time-bounded Relationships** - Expiring memberships and permissions
- **Column-level Security** - Bitmap-based column permissions
- **Caveat Support** - Conditional authorization rules

## Usage

### Running the Migration

```bash
# Apply the migration
psql -d your_database -f 001_create_all_tables.up.sql

# Rollback if needed
psql -d your_database -f 001_create_all_tables.down.sql
```

### Using with Migration Tools

#### Golang Migrate
```bash
# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migration
migrate -path migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" up

# Rollback
migrate -path migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" down
```

#### Flyway
```bash
# Add to flyway.conf
flyway.locations=filesystem:migrations
flyway.sqlMigrationPrefix=

# Run migration
flyway migrate
```

## Database Requirements

- **PostgreSQL 12+** (for JSONB support)
- **uuid-ossp extension** (automatically enabled in migration)

## Security Considerations

- **Password Hashing** - Store only hashed passwords
- **API Key Security** - Hash service API keys
- **Audit Logging** - Comprehensive activity tracking
- **Soft Deletes** - Maintain data integrity and history
- **Foreign Key Constraints** - Prevent orphaned records

## Performance Considerations

- **Index Strategy** - Balanced between query performance and write overhead
- **Partitioning Ready** - Table structure supports future partitioning
- **JSONB Indexes** - Can be added for specific metadata queries
- **Connection Pooling** - Optimize for concurrent access

## Monitoring and Maintenance

### Key Metrics to Monitor
- **Table Sizes** - Monitor growth of audit and event tables
- **Index Usage** - Ensure indexes are being used effectively
- **Constraint Performance** - Monitor foreign key and check constraint overhead
- **JSONB Performance** - Monitor complex JSON queries

### Maintenance Tasks
- **Regular Vacuuming** - For tables with frequent updates/deletes
- **Index Maintenance** - Rebuild indexes periodically
- **Partition Management** - For large audit/event tables
- **Archive Old Data** - Move historical data to archive tables

## Troubleshooting

### Common Issues

1. **Foreign Key Violations**
   - Ensure data is inserted in the correct order
   - Check for orphaned records before deletion

2. **Performance Issues**
   - Verify indexes are being used (EXPLAIN ANALYZE)
   - Consider adding composite indexes for complex queries

3. **Storage Issues**
   - Monitor JSONB column sizes
   - Consider compression for large text fields

### Debugging Queries

```sql
-- Check table sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Check constraint violations
SELECT conname, conrelid::regclass, confrelid::regclass
FROM pg_constraint
WHERE contype = 'f' AND convalidated = false;
```

## Future Enhancements

- **Partitioning** - For large audit and event tables
- **Materialized Views** - For complex authorization queries
- **Full-Text Search** - For audit log and event searching
- **Time-Series Optimization** - For event and audit data
- **Multi-Tenancy** - Enhanced organization isolation

## Support

For issues or questions regarding these migrations, please refer to:
- Database schema documentation
- API documentation
- Development team
