# HOTFIX: Login Timeout Issue

## Problem Summary

Login requests are timing out after ~12+ seconds due to:
1. **Slow database query** on `group_memberships` table (missing index)
2. **Redis timeouts** cascading (12s each)
3. **Context deadline exceeded** causing query cancellation

## Critical Error

```sql
-- This query is taking 12+ seconds!
SELECT * FROM "group_memberships"
WHERE principal_id = 'USER00000001'
AND is_active = true
LIMIT 1000
```

**Error**: `context deadline exceeded`

## Immediate Fixes

### Fix 1: Add Database Index (CRITICAL)

Run this SQL directly on your RDS database:

```sql
-- Add composite index for group membership lookups
CREATE INDEX IF NOT EXISTS idx_group_memberships_principal_active
ON group_memberships(principal_id, is_active)
WHERE is_active = true;

-- Analyze table to update statistics
ANALYZE group_memberships;
```

**Expected improvement**: Query time from 12s → <10ms

### Fix 2: Update Task Definition (Already Done)

The CloudFormation template has been updated with Redis timeouts:
- `REDIS_DIAL_TIMEOUT=5`
- `REDIS_READ_TIMEOUT=3`
- `REDIS_WRITE_TIMEOUT=3`

You need to **redeploy** your ECS service to apply these changes.

### Fix 3: Temporary Workaround - Disable Group Fetching

If you need immediate relief, you can set this environment variable:

```yaml
- Name: SKIP_GROUP_MEMBERSHIPS_ON_LOGIN
  Value: 'true'
```

This will skip the slow group membership query during login.

## Step-by-Step Recovery

### Step 1: Connect to RDS and Add Index

```bash
# Get RDS endpoint
RDS_ENDPOINT=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`RDSEndpoint`].OutputValue' \
  --output text \
  --region ap-south-1)

# Get DB password from Secrets Manager
DB_SECRET=$(aws secretsmanager get-secret-value \
  --secret-id aaa-service-dev-db-credentials \
  --region ap-south-1 \
  --query 'SecretString' \
  --output text)

DB_PASSWORD=$(echo $DB_SECRET | jq -r '.password')

# Connect and add index
psql "postgresql://postgres:$DB_PASSWORD@$RDS_ENDPOINT:5432/aaa_service?sslmode=require" <<EOF
-- Add missing index
CREATE INDEX IF NOT EXISTS idx_group_memberships_principal_active
ON group_memberships(principal_id, is_active)
WHERE is_active = true;

-- Add other useful indexes
CREATE INDEX IF NOT EXISTS idx_group_memberships_group_id
ON group_memberships(group_id, is_active)
WHERE is_active = true;

-- Update statistics
ANALYZE group_memberships;

-- Verify indexes
\d group_memberships
EOF
```

### Step 2: Update Stack with New Task Definition

```bash
cd infrastructure

# Update the stack (this will deploy the Redis timeout configuration)
aws cloudformation deploy \
  --stack-name aaa-service-dev \
  --template-file ../cloudformation.yaml \
  --parameter-overrides file://parameters/beta.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ap-south-1 \
  --no-fail-on-empty-changeset
```

### Step 3: Force New Deployment

```bash
# Force ECS service to use updated task definition
aws ecs update-service \
  --cluster aaa-service-dev-cluster \
  --service aaa-service-dev \
  --force-new-deployment \
  --region ap-south-1

# Monitor deployment
watch -n 5 'aws ecs describe-services \
  --cluster aaa-service-dev-cluster \
  --services aaa-service-dev \
  --region ap-south-1 \
  --query "services[0].{Status:status,Running:runningCount,Desired:desiredCount}"'
```

### Step 4: Verify Fix

```bash
# Get ALB DNS
ALB_DNS=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ALBDNS`].OutputValue' \
  --output text \
  --region ap-south-1)

# Test login endpoint (should complete in <2 seconds)
time curl -X POST http://$ALB_DNS/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "9999999999",
    "country_code": "+91",
    "password": "SuperAdmin@123"
  }'
```

## Additional Performance Indexes

While you're at it, add these indexes for overall performance:

```sql
-- User lookups
CREATE INDEX IF NOT EXISTS idx_users_phone_country
ON users(phone_number, country_code)
WHERE deleted_at IS NULL;

-- User roles
CREATE INDEX IF NOT EXISTS idx_user_roles_user_active
ON user_roles(user_id, is_active)
WHERE is_active = true;

-- Role permissions
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_active
ON role_permissions(role_id, is_active)
WHERE is_active = true;

-- Analyze all tables
ANALYZE users;
ANALYZE user_roles;
ANALYZE role_permissions;
ANALYZE group_memberships;
```

## Monitoring After Fix

### Check Query Performance

```sql
-- Enable query logging (if not already enabled)
ALTER SYSTEM SET log_min_duration_statement = 1000;
SELECT pg_reload_conf();

-- Check slow queries
SELECT
  query,
  calls,
  mean_exec_time,
  max_exec_time
FROM pg_stat_statements
WHERE query LIKE '%group_memberships%'
ORDER BY mean_exec_time DESC
LIMIT 10;
```

### Check Application Logs

```bash
# Should no longer see "context deadline exceeded"
aws logs tail /ecs/aaa-service-dev --follow --region ap-south-1 \
  --filter-pattern "deadline exceeded"

# Login should complete quickly
aws logs tail /ecs/aaa-service-dev --follow --region ap-south-1 \
  --filter-pattern "logged in successfully"
```

## Expected Results After Fix

| Metric | Before | After |
|--------|--------|-------|
| Login time | 12-85 seconds | 1-2 seconds |
| group_memberships query | 12+ seconds | <10ms |
| Redis timeout impact | 12s per operation | 3s per operation |
| Request timeout rate | ~80% | <1% |

## Root Cause Analysis

### Why This Happened

1. **No Index on Lookup Column**: The `group_memberships` table was created without an index on `principal_id`
2. **Full Table Scan**: PostgreSQL had to scan every row to find matches
3. **Redis Multiplier**: Each Redis timeout added 12s, compounding delays
4. **Context Cancellation**: Total time exceeded request timeout, cancelling in-flight queries

### Why It Affects Login

The login flow fetches:
```
User → Roles → Organizations → Groups
```

Each step tries Redis first (12s timeout), then falls back to DB. With the slow `group_memberships` query, the total time exceeds the request timeout.

## Prevention

### Add Migration for Indexes

Create `migrations/add_performance_indexes.go`:

```go
package migrations

import (
    "context"
    "gorm.io/gorm"
    "go.uber.org/zap"
)

func AddPerformanceIndexes(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
    indexes := []string{
        "CREATE INDEX IF NOT EXISTS idx_group_memberships_principal_active ON group_memberships(principal_id, is_active) WHERE is_active = true",
        "CREATE INDEX IF NOT EXISTS idx_group_memberships_group_id ON group_memberships(group_id, is_active) WHERE is_active = true",
        "CREATE INDEX IF NOT EXISTS idx_users_phone_country ON users(phone_number, country_code) WHERE deleted_at IS NULL",
        "CREATE INDEX IF NOT EXISTS idx_user_roles_user_active ON user_roles(user_id, is_active) WHERE is_active = true",
        "CREATE INDEX IF NOT EXISTS idx_role_permissions_role_active ON role_permissions(role_id, is_active) WHERE is_active = true",
    }

    for _, idx := range indexes {
        if err := db.WithContext(ctx).Exec(idx).Error; err != nil {
            logger.Error("Failed to create index", zap.String("sql", idx), zap.Error(err))
            return err
        }
        logger.Info("Created performance index", zap.String("sql", idx))
    }

    return nil
}
```

### Monitor Query Performance

Set up CloudWatch alarm for slow queries:

```bash
aws cloudwatch put-metric-alarm \
  --alarm-name aaa-service-dev-slow-queries \
  --alarm-description "Alert on slow database queries" \
  --metric-name DatabaseConnections \
  --namespace AWS/RDS \
  --statistic Average \
  --period 300 \
  --threshold 100 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --region ap-south-1
```

## Contact

If issues persist after applying these fixes:
1. Check RDS Performance Insights
2. Review slow query log
3. Verify indexes were created: `\d group_memberships` in psql
4. Check ECS task is using new task definition with Redis timeouts
