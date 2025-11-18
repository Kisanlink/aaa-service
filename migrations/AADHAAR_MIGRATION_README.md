# Aadhaar Integration Database Migrations

This document describes the database migrations for Aadhaar verification and KYC functionality.

## Overview

Three migration files have been created to support Aadhaar OTP verification and KYC workflows:

1. **add_aadhaar_verifications_table.go** - Core verification tracking table
2. **add_otp_attempts_table.go** - OTP attempt audit trail
3. **add_user_profiles_kyc_fields.go** - User profile KYC status tracking

## Migration Files

### 1. aadhaar_verifications Table

**File**: `migrations/add_aadhaar_verifications_table.go`

**Purpose**: Stores Aadhaar verification requests, OTP generation details, and verified KYC data.

**Table Structure**:
- Primary Key: `id` (VARCHAR 255)
- Foreign Key: `user_id` → `users(id)` ON DELETE CASCADE
- Unique Constraints: `transaction_id`, `reference_id`
- Total Columns: 22

**Key Columns**:
- `aadhaar_number` - Masked/encrypted Aadhaar number
- `transaction_id` - Sandbox API transaction identifier
- `reference_id` - OTP verification reference
- `verification_status` - PENDING, VERIFIED, FAILED
- `kyc_status` - PENDING, VERIFIED, REJECTED
- `address_json` - JSONB structure for parsed address
- `attempts` - Number of OTP attempts
- `photo_url` - S3 URL for Aadhaar photo

**Indexes Created** (4 total):
1. `idx_aadhaar_verifications_user_id` - User lookup optimization
2. `idx_aadhaar_verifications_transaction_id` - OTP generation lookup
3. `idx_aadhaar_verifications_reference_id` - OTP verification lookup
4. `idx_aadhaar_verifications_status` - Composite index on verification_status, kyc_status

**Functions**:
- `AddAadhaarVerificationsTable(ctx, db, logger)` - Creates table and indexes
- `DropAadhaarVerificationsTable(ctx, db, logger)` - Rollback (drops table)
- `ValidateAadhaarVerificationsTable(ctx, db, logger)` - Validates migration

---

### 2. otp_attempts Table

**File**: `migrations/add_otp_attempts_table.go`

**Purpose**: Audit trail for every OTP verification attempt (security and compliance).

**Table Structure**:
- Primary Key: `id` (VARCHAR 255)
- Foreign Key: `aadhaar_verification_id` → `aadhaar_verifications(id)` ON DELETE CASCADE
- Total Columns: 9

**Key Columns**:
- `aadhaar_verification_id` - Links to parent verification record
- `attempt_number` - Sequential attempt counter
- `otp_value` - Hashed OTP (never plain text)
- `ip_address` - Client IP (security audit)
- `user_agent` - Client user agent
- `status` - SUCCESS, FAILED, EXPIRED
- `failed_reason` - Error message on failure

**Indexes Created** (2 total):
1. `idx_otp_attempts_verification_id` - Join optimization
2. `idx_otp_attempts_status` - Analytics and monitoring

**Functions**:
- `AddOTPAttemptsTable(ctx, db, logger)` - Creates table and indexes
- `DropOTPAttemptsTable(ctx, db, logger)` - Rollback (drops table)
- `ValidateOTPAttemptsTable(ctx, db, logger)` - Validates migration

---

### 3. user_profiles KYC Fields

**File**: `migrations/add_user_profiles_kyc_fields.go`

**Purpose**: Adds KYC verification tracking to existing user_profiles table.

**Columns Added** (3 total):
- `aadhaar_verified` BOOLEAN DEFAULT FALSE
- `aadhaar_verified_at` TIMESTAMP
- `kyc_status` VARCHAR(50) DEFAULT 'PENDING'

**Indexes Created** (2 total):
1. `idx_user_profiles_kyc_status` - Filter by KYC status
2. `idx_user_profiles_aadhaar_verified` - Filter verified users

**Functions**:
- `AddUserProfilesKYCFields(ctx, db, logger)` - Adds columns and indexes
- `DropUserProfilesKYCFields(ctx, db, logger)` - Rollback (drops columns)
- `ValidateUserProfilesKYCFields(ctx, db, logger)` - Validates migration

---

## Running Migrations

### Method 1: Using Test Script (Recommended)

A migration runner script is provided for testing and validation:

```bash
# Run all migrations
go run scripts/run_aadhaar_migrations.go migrate

# Check migration status
go run scripts/run_aadhaar_migrations.go status

# Validate migrations
go run scripts/run_aadhaar_migrations.go validate

# Rollback (DESTRUCTIVE - use only in dev/test)
go run scripts/run_aadhaar_migrations.go rollback
```

**Output Example**:
```
INFO    Testing Aadhaar Integration Migrations
INFO    Step 1: Running migrations...
INFO    Creating aadhaar_verifications table
INFO    Successfully created aadhaar_verifications table
INFO    Creating index idx_aadhaar_verifications_user_id
INFO    ✅ All migrations completed successfully
```

### Method 2: Programmatic Execution

The migrations will be automatically executed on server startup if called from `cmd/server/main.go`:

```go
// Add to runSeedScripts() or equivalent initialization function
if err := migrations.TestAadhaarMigrations(ctx, gormDB, logger); err != nil {
    logger.Warn("Failed to run Aadhaar migrations", zap.Error(err))
}
```

---

## Validation

### Automatic Validation

Run the test helper to validate all migrations:

```bash
go run scripts/run_aadhaar_migrations.go validate
```

### Manual PostgreSQL Validation

```sql
-- Check tables exist
SELECT tablename FROM pg_tables WHERE tablename IN ('aadhaar_verifications', 'otp_attempts');

-- Check indexes exist
SELECT indexname FROM pg_indexes
WHERE tablename IN ('aadhaar_verifications', 'otp_attempts', 'user_profiles')
  AND indexname LIKE 'idx_%aadhaar%' OR indexname LIKE 'idx_%kyc%';

-- Check foreign key constraints
SELECT conname, conrelid::regclass, confrelid::regclass
FROM pg_constraint
WHERE contype = 'f' AND conrelid::regclass::text IN ('aadhaar_verifications', 'otp_attempts');

-- Verify column counts
SELECT
    'aadhaar_verifications' as table_name,
    COUNT(*) as column_count
FROM information_schema.columns
WHERE table_name = 'aadhaar_verifications'
UNION ALL
SELECT
    'otp_attempts' as table_name,
    COUNT(*) as column_count
FROM information_schema.columns
WHERE table_name = 'otp_attempts';

-- Check user_profiles KYC columns
SELECT column_name, data_type, column_default
FROM information_schema.columns
WHERE table_name = 'user_profiles'
  AND column_name IN ('aadhaar_verified', 'aadhaar_verified_at', 'kyc_status');
```

**Expected Results**:
- aadhaar_verifications: 22 columns, 4 indexes, 1 foreign key
- otp_attempts: 9 columns, 2 indexes, 1 foreign key
- user_profiles: 3 new KYC columns, 2 new indexes

---

## Migration Dependency Order

**CRITICAL**: Migrations must run in this exact order due to foreign key dependencies:

1. **First**: `AddAadhaarVerificationsTable()` - Creates parent table
2. **Second**: `AddOTPAttemptsTable()` - Depends on aadhaar_verifications FK
3. **Third**: `AddUserProfilesKYCFields()` - Independent, can run anytime

**Rollback Order** (reverse):

1. **First**: `DropUserProfilesKYCFields()`
2. **Second**: `DropOTPAttemptsTable()`
3. **Third**: `DropAadhaarVerificationsTable()`

---

## Security Considerations

### Data Protection

1. **Aadhaar Number Masking**:
   - Never log full Aadhaar numbers
   - Display only last 4 digits: `XXXX-XXXX-1234`
   - Store with encryption at application layer (not handled by migration)

2. **OTP Security**:
   - `otp_value` column stores hashed OTP, never plain text
   - Use bcrypt or similar before storing
   - OTP attempts tracked for rate limiting

3. **Photo URLs**:
   - S3 URLs should use signed URLs with expiration
   - Photos encrypted at rest (S3 SSE)

### OWASP ASVS Compliance

These migrations support ASVS Level 2 controls:

- **V2.1** (Password Security): OTP hashing in otp_attempts
- **V8.1** (Data Protection): Separate audit table for OTP attempts
- **V9.2** (Communications): Support for JSONB address storage
- **V13.1** (Malicious Code): Prepared statements via GORM prevent SQL injection

---

## Performance Optimization

### Index Strategy

All indexes use B-tree (PostgreSQL default) optimized for:

1. **Equality Lookups**: user_id, transaction_id, reference_id
2. **Composite Queries**: (verification_status, kyc_status)
3. **Range Scans**: status filtering for dashboards
4. **Join Optimization**: aadhaar_verification_id in otp_attempts

### Query Performance Targets

| Query Type | Index Used | Target |
|------------|-----------|--------|
| Get verification by user_id | idx_aadhaar_verifications_user_id | < 10ms |
| Get verification by reference_id | idx_aadhaar_verifications_reference_id | < 5ms |
| Get OTP attempts by verification_id | idx_otp_attempts_verification_id | < 10ms |
| Filter users by KYC status | idx_user_profiles_kyc_status | < 50ms |

### ANALYZE Tables

After migration, run `ANALYZE` to update query planner statistics:

```sql
ANALYZE aadhaar_verifications;
ANALYZE otp_attempts;
ANALYZE user_profiles;
```

This is automatically done by `AddUserProfilesKYCFields()` migration.

---

## Troubleshooting

### Migration Fails: Table Already Exists

**Symptom**: Error message "relation already exists"

**Solution**: Migrations check for existing tables and skip creation. If partial migration occurred, run validation:

```bash
go run scripts/run_aadhaar_migrations.go validate
```

### Foreign Key Constraint Error

**Symptom**: Cannot create otp_attempts table

**Cause**: aadhaar_verifications table not created first

**Solution**: Ensure migrations run in correct order (see Dependency Order section)

### Index Creation Fails

**Symptom**: Index creation error

**Cause**: Index may already exist from previous run

**Solution**: Migrations use `CREATE INDEX IF NOT EXISTS` - check logs for actual error

### Column Count Mismatch

**Symptom**: Validation reports incorrect column count

**Cause**: Partial migration or database drift

**Solution**:
1. Check actual table schema: `\d aadhaar_verifications`
2. Compare with migration SQL
3. Manually add missing columns or re-run migration

---

## Rollback Procedure

**WARNING**: Rollback is DESTRUCTIVE and will delete all Aadhaar verification data.

### When to Rollback

- Only in development/test environments
- Before re-running migrations with schema changes
- To clean up failed migration attempts

### How to Rollback

```bash
# Using script (recommended)
go run scripts/run_aadhaar_migrations.go rollback

# Programmatic rollback
TestAadhaarMigrationsRollback(ctx, db, logger)
```

### Manual Rollback (SQL)

If automated rollback fails:

```sql
-- Drop in reverse dependency order
DROP TABLE IF EXISTS otp_attempts CASCADE;
DROP TABLE IF EXISTS aadhaar_verifications CASCADE;

-- Remove KYC columns from user_profiles
ALTER TABLE user_profiles
DROP COLUMN IF EXISTS aadhaar_verified,
DROP COLUMN IF EXISTS aadhaar_verified_at,
DROP COLUMN IF EXISTS kyc_status;

-- Drop indexes
DROP INDEX IF EXISTS idx_aadhaar_verifications_user_id;
DROP INDEX IF EXISTS idx_aadhaar_verifications_transaction_id;
DROP INDEX IF EXISTS idx_aadhaar_verifications_reference_id;
DROP INDEX IF EXISTS idx_aadhaar_verifications_status;
DROP INDEX IF EXISTS idx_otp_attempts_verification_id;
DROP INDEX IF EXISTS idx_otp_attempts_status;
DROP INDEX IF EXISTS idx_user_profiles_kyc_status;
DROP INDEX IF EXISTS idx_user_profiles_aadhaar_verified;
```

---

## Integration with Server Startup

To run migrations on server startup, add to `cmd/server/main.go`:

```go
// After other migrations in runSeedScripts()
if err := migrations.TestAadhaarMigrations(ctx, gormDB, logger); err != nil {
    logger.Warn("Failed to run Aadhaar migrations", zap.Error(err))
    // Don't fail startup - migrations can be run manually
} else {
    logger.Info("✅ Aadhaar migrations completed successfully")
}
```

**Note**: Current implementation does NOT auto-run on startup. Run manually using script.

---

## Migration Summary

| Migration | Tables | Columns | Indexes | Foreign Keys |
|-----------|--------|---------|---------|--------------|
| aadhaar_verifications | 1 | 22 | 4 | 1 (to users) |
| otp_attempts | 1 | 9 | 2 | 1 (to aadhaar_verifications) |
| user_profiles KYC | 0 (ALTER) | 3 | 2 | 0 |
| **TOTAL** | **2 new** | **34** | **8** | **2** |

---

## Next Steps

After running migrations successfully:

1. ✅ Update `.kiro/implementation/aadhaar-integration-tracker.md` - Mark Task 1.1 COMPLETE
2. → Proceed to Task 1.2: Create Data Models (entities/models)
3. → Create repository interfaces and implementations
4. → Implement service layer (KYC service, Sandbox client)

---

## Support

For issues or questions:
- Check logs from migration runner script
- Review validation output
- Verify database connectivity and permissions
- Ensure PostgreSQL version >= 12 (JSONB support)

**Migration Status Check**:
```bash
go run scripts/run_aadhaar_migrations.go status
```

Expected output when successful:
```
INFO    Migration Summary
        aadhaar_verifications: true
        otp_attempts: true
        user_profiles_kyc: true
        total_indexes: 8
        all_migrated: true
```
